package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/contrib/websocket" // Use Fiber WebSocket
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MessageHandler interface for messages related operations via WebSocket
type MessageHandler interface {
	HandleWebSocketMessage(c *websocket.Conn) // Method to handle WebSocket messages
}

// messageHandlerImpl implements MessageHandler
type messageHandlerImpl struct {
	repo repository.MessageRepository
	// Use a map to manage active WebSocket connections by channel ID
	activeConnections map[string]map[*websocket.Conn]bool
	mu                sync.RWMutex // Mutex for thread-safe access to activeConnections
}

// NewMessageHandler creates a new instance of MessageHandler.
func NewMessageHandler(repo repository.MessageRepository) MessageHandler {
	return &messageHandlerImpl{
		repo:              repo,
		activeConnections: make(map[string]map[*websocket.Conn]bool),
	}
}

// HandleWebSocketMessage handles WebSocket connections and messages.
func (h *messageHandlerImpl) HandleWebSocketMessage(c *websocket.Conn) {
	defer func() {
		// Remove connection from active connections when it closes
		channelID := c.Params("channelId")
		h.mu.Lock()
		delete(h.activeConnections[channelID], c)
		if len(h.activeConnections[channelID]) == 0 {
			delete(h.activeConnections, channelID)
		}
		h.mu.Unlock()
		log.Printf("Client disconnected from channel %s: %s\n", channelID, c.LocalAddr().String())
		c.Close()
	}()

	channelID := c.Params("channelId")
	if channelID == "" {
		log.Println("Error: Channel ID not provided in WebSocket URL.")
		// Optionally send an error message back to the client before closing
		c.WriteMessage(websocket.TextMessage, []byte("Error: Channel ID required."))
		return
	}

	// Add the new connection to our active connections map
	h.mu.Lock()
	if h.activeConnections[channelID] == nil {
		h.activeConnections[channelID] = make(map[*websocket.Conn]bool)
	}
	h.activeConnections[channelID][c] = true
	h.mu.Unlock()

	log.Printf("Client connected to channel %s: %s\n", channelID, c.LocalAddr().String())

	// Send a confirmation to the client that they joined the channel
	c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Joined channel: %s", channelID)))

	// Loop to read messages from the client
	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			// Log the error and break the loop to close the connection
			log.Println("read error:", err)
			break
		}

		if mt != websocket.TextMessage {
			log.Println("Received non-text message type, skipping.")
			continue
		}

		// Handle different message types based on a 'type' field in the JSON payload
		var wsMessage struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(msg, &wsMessage); err != nil {
			logAndEmitErrorWS(c, "Invalid message format", err)
			continue
		}

		switch wsMessage.Type {
		case "client_message":
			h.handleCreateMessage(c, channelID, wsMessage.Payload)
		case "get_message_history":
			h.handleGetMessageHistory(c, channelID, wsMessage.Payload)
		case "update_message":
			h.handleUpdateMessage(c, channelID, wsMessage.Payload)
		case "delete_message":
			h.handleDeleteMessage(c, channelID, wsMessage.Payload)
		case "add_reaction":
			h.handleAddReaction(c, channelID, wsMessage.Payload)
		case "remove_reaction":
			h.handleRemoveReaction(c, channelID, wsMessage.Payload)
		default:
			logAndEmitErrorWS(c, "Unknown message type", nil)
		}
	}
}

// Helper to broadcast messages to all connections in a specific channel
func (h *messageHandlerImpl) broadcastToChannel(channelID string, messageType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if connections, ok := h.activeConnections[channelID]; ok {
		jsonMsg, err := json.Marshal(map[string]interface{}{
			"type":    messageType,
			"payload": data,
		})
		if err != nil {
			log.Printf("Error marshalling broadcast message: %v\n", err)
			return
		}

		for conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
				log.Printf("Error writing to websocket for channel %s: %v\n", channelID, err)
				// Remove broken connection
				func(brokenConn *websocket.Conn) {
					h.mu.Lock()
					delete(h.activeConnections[channelID], brokenConn)
					if len(h.activeConnections[channelID]) == 0 {
						delete(h.activeConnections, channelID)
					}
					h.mu.Unlock()
				}(conn)
			}
		}
	}
}

// --- Individual handler functions for different WebSocket message types ---

func (h *messageHandlerImpl) handleCreateMessage(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	var req dto.MessageCreateRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		logAndEmitErrorWS(c, "Invalid message payload", err)
		return
	}

	channelID, err := primitive.ObjectIDFromHex(channelIDStr)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid channel ID", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid user ID", err)
		return
	}

	var mediaMetadata *model.MessageMediaMetadata
	if req.MediaMetadata != nil {
		mediaMetadata = &model.MessageMediaMetadata{
			Filename: req.MediaMetadata.Filename,
			Size:     req.MediaMetadata.Size,
			Width:    req.MediaMetadata.Width,
			Height:   req.MediaMetadata.Height,
		}
	}

	newMessage := &model.Message{
		ChannelID:     channelID,
		UserID:        userID,
		Content:       req.Content,
		MessageType:   req.MessageType,
		MediaPath:     req.MediaPath,
		MediaMetadata: mediaMetadata,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.CreateMessage(ctx, newMessage); err != nil {
		utils.LogError(err, "Failed to save message to database")
		logAndEmitErrorWS(c, "Failed to create message", err)
		return
	}

	resp := dto.MessageResponse{
		ID:            newMessage.ID.Hex(),
		ChannelID:     newMessage.ChannelID.Hex(),
		UserID:        newMessage.UserID.Hex(),
		Content:       newMessage.Content,
		MessageType:   newMessage.MessageType,
		MediaPath:     newMessage.MediaPath,
		MediaMetadata: convertMediaMetadataToDTO(newMessage.MediaMetadata),
		CreatedAt:     newMessage.CreatedAt,
		UpdatedAt:     newMessage.UpdatedAt,
		IsPinned:      newMessage.IsPinned,
		Reactions:     convertMessageReactionsToDTO(newMessage.Reactions),
	}

	// Emit success back to the sender
	if err := c.WriteJSON(map[string]interface{}{"type": "message_created", "payload": resp}); err != nil {
		log.Printf("Error emitting message_created: %v\n", err)
	}
	// Broadcast to all other clients in the channel room
	h.broadcastToChannel(channelIDStr, "new_message", resp)
}

func (h *messageHandlerImpl) handleGetMessageHistory(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	channelID, err := primitive.ObjectIDFromHex(channelIDStr)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid channel ID", err)
		return
	}

	var req struct {
		Limit int64 `json:"limit,omitempty"`
		Skip  int64 `json:"skip,omitempty"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		logAndEmitErrorWS(c, "Invalid get_message_history payload", err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Skip == 0 {
		req.Skip = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages, err := h.repo.GetMessagesByChannelID(ctx, channelID, req.Limit, req.Skip)
	if err != nil {
		utils.LogError(err, "Failed to fetch message history")
		logAndEmitErrorWS(c, "Failed to retrieve message history", err)
		return
	}

	respMessages := []dto.MessageResponse{}
	for _, msg := range messages {
		respMessages = append(respMessages, dto.MessageResponse{
			ID:            msg.ID.Hex(),
			ChannelID:     msg.ChannelID.Hex(),
			UserID:        msg.UserID.Hex(),
			Content:       msg.Content,
			MessageType:   msg.MessageType,
			MediaPath:     msg.MediaPath,
			MediaMetadata: convertMediaMetadataToDTO(msg.MediaMetadata),
			CreatedAt:     msg.CreatedAt,
			UpdatedAt:     msg.UpdatedAt,
			IsPinned:      msg.IsPinned,
			Reactions:     convertMessageReactionsToDTO(msg.Reactions),
		})
	}
	if err := c.WriteJSON(map[string]interface{}{"type": "message_history", "payload": respMessages}); err != nil {
		log.Printf("Error emitting message_history: %v\n", err)
	}
}

func (h *messageHandlerImpl) handleUpdateMessage(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	var updatePayload struct {
		ID string `json:"id"`
		dto.MessageUpdateRequest
	}
	if err := json.Unmarshal(payload, &updatePayload); err != nil {
		logAndEmitErrorWS(c, "Invalid update_message payload", err)
		return
	}
	messageIDStr := updatePayload.ID
	req := updatePayload.MessageUpdateRequest

	msgObjectID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid message ID for update", err)
		return
	}

	updateMap := bson.M{}
	setMap := bson.M{}

	if req.Content != "" {
		setMap["content"] = req.Content
	}
	if req.MessageType != "" {
		setMap["messageType"] = req.MessageType
	}
	if req.MediaPath != "" {
		setMap["mediaPath"] = req.MediaPath
	}
	if req.MediaMetadata != nil {
		setMap["mediaMetadata"] = model.MessageMediaMetadata{
			Filename: req.MediaMetadata.Filename,
			Size:     req.MediaMetadata.Size,
			Width:    req.MediaMetadata.Width,
			Height:   req.MediaMetadata.Height,
		}
	}
	if req.IsPinned != nil {
		setMap["isPinned"] = *req.IsPinned
	}

	if len(setMap) > 0 {
		updateMap["$set"] = setMap
		// Removed duplicate line: updateMap["$set"] = setMap
	}

	if len(updateMap) == 0 {
		logAndEmitErrorWS(c, "No data to update", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updatedMessage, err := h.repo.UpdateMessage(ctx, msgObjectID, updateMap)
	if err != nil {
		utils.LogError(err, "Failed to update message in database")
		logAndEmitErrorWS(c, "Failed to update message", err)
		return
	}

	if updatedMessage == nil {
		logAndEmitErrorWS(c, "Message not found for update", nil)
		return
	}

	resp := dto.MessageResponse{
		ID:            updatedMessage.ID.Hex(),
		ChannelID:     updatedMessage.ChannelID.Hex(),
		UserID:        updatedMessage.UserID.Hex(),
		Content:       updatedMessage.Content,
		MessageType:   updatedMessage.MessageType,
		MediaPath:     updatedMessage.MediaPath,
		MediaMetadata: convertMediaMetadataToDTO(updatedMessage.MediaMetadata),
		CreatedAt:     updatedMessage.CreatedAt,
		UpdatedAt:     updatedMessage.UpdatedAt,
		IsPinned:      updatedMessage.IsPinned,
		Reactions:     convertMessageReactionsToDTO(updatedMessage.Reactions),
	}
	if err := c.WriteJSON(map[string]interface{}{"type": "message_updated", "payload": resp}); err != nil {
		log.Printf("Error emitting message_updated: %v\n", err)
	}
	h.broadcastToChannel(channelIDStr, "message_updated", resp)
}

func (h *messageHandlerImpl) handleDeleteMessage(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		logAndEmitErrorWS(c, "Invalid delete_message payload", err)
		return
	}

	msgObjectID, err := primitive.ObjectIDFromHex(req.ID)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid message ID for delete", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.repo.DeleteMessage(ctx, msgObjectID); err != nil {
		if err == mongo.ErrNoDocuments {
			logAndEmitErrorWS(c, "Message not found for deletion", nil)
		} else {
			utils.LogError(err, "Failed to delete message from database")
			logAndEmitErrorWS(c, "Failed to delete message", err)
		}
		return
	}

	if err := c.WriteJSON(map[string]interface{}{"type": "message_deleted", "payload": map[string]string{"id": req.ID}}); err != nil {
		log.Printf("Error emitting message_deleted: %v\n", err)
	}
	h.broadcastToChannel(channelIDStr, "message_deleted", map[string]string{"id": req.ID})
}

func (h *messageHandlerImpl) handleAddReaction(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	var reactionPayload struct {
		MessageID string `json:"messageId"`
		dto.MessageReactionAddRequest
	}
	if err := json.Unmarshal(payload, &reactionPayload); err != nil {
		logAndEmitErrorWS(c, "Invalid add_reaction payload", err)
		return
	}
	messageIDStr := reactionPayload.MessageID
	req := reactionPayload.MessageReactionAddRequest

	msgObjectID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid message ID for reaction", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid user ID for reaction", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updatedMessage, err := h.repo.AddMessageReaction(ctx, msgObjectID, userID, req.Emoji)
	if err != nil {
		utils.LogError(err, "Failed to add reaction to message")
		logAndEmitErrorWS(c, "Failed to add reaction", err)
		return
	}

	if updatedMessage == nil {
		logAndEmitErrorWS(c, "Message not found to add reaction", nil)
		return
	}

	resp := dto.MessageResponse{
		ID:            updatedMessage.ID.Hex(),
		ChannelID:     updatedMessage.ChannelID.Hex(),
		UserID:        updatedMessage.UserID.Hex(),
		Content:       updatedMessage.Content,
		MessageType:   updatedMessage.MessageType,
		MediaPath:     updatedMessage.MediaPath,
		MediaMetadata: convertMediaMetadataToDTO(updatedMessage.MediaMetadata),
		CreatedAt:     updatedMessage.CreatedAt,
		UpdatedAt:     updatedMessage.UpdatedAt,
		IsPinned:      updatedMessage.IsPinned,
		Reactions:     convertMessageReactionsToDTO(updatedMessage.Reactions),
	}
	if err := c.WriteJSON(map[string]interface{}{"type": "reaction_added", "payload": resp}); err != nil {
		log.Printf("Error emitting reaction_added: %v\n", err)
	}
	h.broadcastToChannel(channelIDStr, "reaction_added", resp)
}

func (h *messageHandlerImpl) handleRemoveReaction(c *websocket.Conn, channelIDStr string, payload json.RawMessage) {
	var reactionPayload struct {
		MessageID string `json:"messageId"`
		dto.MessageReactionRemoveRequest
	}
	if err := json.Unmarshal(payload, &reactionPayload); err != nil {
		logAndEmitErrorWS(c, "Invalid remove_reaction payload", err)
		return
	}
	messageIDStr := reactionPayload.MessageID
	req := reactionPayload.MessageReactionRemoveRequest

	msgObjectID, err := primitive.ObjectIDFromHex(messageIDStr)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid message ID for reaction removal", err)
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		logAndEmitErrorWS(c, "Invalid user ID for reaction removal", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updatedMessage, err := h.repo.RemoveMessageReaction(ctx, msgObjectID, userID, req.Emoji)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logAndEmitErrorWS(c, "Message or reaction not found for removal", nil)
		} else {
			utils.LogError(err, "Failed to remove reaction from message")
			logAndEmitErrorWS(c, "Failed to remove reaction", err)
		}
		return
	}

	if updatedMessage == nil {
		logAndEmitErrorWS(c, "Message not found to remove reaction", nil)
		return
	}

	resp := dto.MessageResponse{
		ID:            updatedMessage.ID.Hex(),
		ChannelID:     updatedMessage.ChannelID.Hex(),
		UserID:        updatedMessage.UserID.Hex(),
		Content:       updatedMessage.Content,
		MessageType:   updatedMessage.MessageType,
		MediaPath:     updatedMessage.MediaPath,
		MediaMetadata: convertMediaMetadataToDTO(updatedMessage.MediaMetadata),
		CreatedAt:     updatedMessage.CreatedAt,
		UpdatedAt:     updatedMessage.UpdatedAt,
		IsPinned:      updatedMessage.IsPinned,
		Reactions:     convertMessageReactionsToDTO(updatedMessage.Reactions),
	}
	if err := c.WriteJSON(map[string]interface{}{"type": "reaction_removed", "payload": resp}); err != nil {
		log.Printf("Error emitting reaction_removed: %v\n", err)
	}
	h.broadcastToChannel(channelIDStr, "reaction_removed", resp)
}

// Helper function to log and emit errors over WebSocket
func logAndEmitErrorWS(conn *websocket.Conn, message string, err error) {
	logMsg := message
	if err != nil {
		logMsg = fmt.Sprintf("%s: %v", message, err)
		utils.LogError(err, message)
	}
	log.Println(logMsg)
	if err := conn.WriteJSON(map[string]interface{}{"type": "error", "payload": logMsg}); err != nil {
		log.Printf("Error emitting error message over WS: %v\n", err)
	}
}

// convertMediaMetadataToDTO remains the same
func convertMediaMetadataToDTO(metadata *model.MessageMediaMetadata) *dto.MessageMediaMetadataRequest {
	if metadata == nil {
		return nil
	}
	return &dto.MessageMediaMetadataRequest{
		Filename: metadata.Filename,
		Size:     metadata.Size,
		Width:    metadata.Width,
		Height:   metadata.Height,
	}
}

// convertMessageReactionsToDTO remains the same
func convertMessageReactionsToDTO(reactions []model.MessageReaction) []dto.MessageReactionRequest {
	dtoReactions := make([]dto.MessageReactionRequest, len(reactions))
	for i, r := range reactions {
		var userID string
		if len(r.UserIDs) > 0 {
			userID = r.UserIDs[0].Hex()
		}
		dtoReactions[i] = dto.MessageReactionRequest{
			Emoji:  r.Emoji,
			UserID: userID,
		}
	}
	return dtoReactions
}

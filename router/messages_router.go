package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	"github.com/gofiber/contrib/websocket" // Import Fiber WebSocket
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupMessageRoutes mendaftarkan rute WebSocket untuk entitas Message.
func SetupMessageRoutes(api fiber.Router, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler untuk Message
	messageRepo := repository.NewMessageRepository(dbClient)
	messageHandler := handler.NewMessageHandler(messageRepo)

	// WebSocket route for messages
	// The path should be unique and not conflict with REST API endpoints.
	// It's common to use /ws or /socket for WebSocket routes.
	api.Get("/ws/messages/:channelId", websocket.New(messageHandler.HandleWebSocketMessage))

	// Fiber routing for RESTful API (if still present)
	// For WebSocket, all communication will be through the /ws/messages/:channelId endpoint.
}

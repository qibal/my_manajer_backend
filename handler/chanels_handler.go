package handler

import (
	"context"
	"fmt"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/service"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChannelHandler adalah interface untuk handler Channel.
type ChannelHandler interface {
	CreateChannel(c *fiber.Ctx) error
	GetChannelByID(c *fiber.Ctx) error
	GetAllChannels(c *fiber.Ctx) error
	UpdateChannel(c *fiber.Ctx) error
	DeleteChannel(c *fiber.Ctx) error
	GetChannelsByBusinessID(c *fiber.Ctx) error // Tambahan
}

// channelHandlerImpl adalah implementasi dari ChannelHandler.
type channelHandlerImpl struct {
	repo               repository.ChannelRepository
	activityLogService service.ActivityLogService
}

// NewChannelHandler membuat instance baru dari ChannelHandler.
func NewChannelHandler(repo repository.ChannelRepository, activityLogService service.ActivityLogService) ChannelHandler {
	return &channelHandlerImpl{
		repo:               repo,
		activityLogService: activityLogService,
	}
}

// @Summary Create a new channel
// @Description Creates a new channel with the provided details.
// @Tags Channels
// @Accept json
// @Produce json
// @Param channel body dto.ChannelCreateRequest true "Channel object to be created"
// @Success 201 {object} utils.APIResponse{data=dto.ChannelResponse} "Successfully created channel"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels [post]
func (h *channelHandlerImpl) CreateChannel(c *fiber.Ctx) error {
	var req dto.ChannelCreateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk membuat channel")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	businessID, err := primitive.ObjectIDFromHex(req.BusinessID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Business ID tidak valid", err.Error())
	}

	var categoryID *primitive.ObjectID // Use pointer
	if req.CategoryID != "" {
		oid, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Category ID tidak valid", err.Error())
		}
		categoryID = &oid // Assign address of ObjectID
	} else {
		categoryID = nil // Explicitly set to nil if empty
	}

	channel := &model.Channel{
		BusinessID: businessID,
		Name:       req.Name,
		Type:       req.Type,
		CategoryID: categoryID, // Assign pointer
		Order:      req.Order,
		CreatedAt:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.repo.CreateChannel(ctx, channel); err != nil {
		utils.LogError(err, "Gagal membuat channel di repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat channel", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Created channel '%s' in business %s", channel.Name, req.BusinessID), c.Method(), c.Path(), fiber.StatusCreated, c.IP())

	resp := dto.ChannelResponse{
		ID:         channel.ID.Hex(),
		BusinessID: channel.BusinessID.Hex(),
		Name:       channel.Name,
		Type:       channel.Type,
		Order:      channel.Order,
		CreatedAt:  channel.CreatedAt,
	}
	if channel.CategoryID != nil {
		resp.CategoryID = channel.CategoryID.Hex() // Convert pointer to hex string
	} else {
		resp.CategoryID = "" // Ensure it's an empty string if nil
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Channel berhasil dibuat", resp)
}

// @Summary Get a channel by ID
// @Description Retrieves a single channel by its ID.
// @Tags Channels
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} utils.APIResponse{data=dto.ChannelResponse} "Successfully retrieved channel"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels/{id} [get]
func (h *channelHandlerImpl) GetChannelByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak boleh kosong", nil)
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	channel, err := h.repo.GetChannelByID(ctx, objectID)
	if err != nil {
		utils.LogError(err, "Gagal mengambil channel berdasarkan ID dari repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil channel", err.Error())
	}

	if channel == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Channel tidak ditemukan", nil)
	}

	resp := dto.ChannelResponse{
		ID:         channel.ID.Hex(),
		BusinessID: channel.BusinessID.Hex(),
		Name:       channel.Name,
		Type:       channel.Type,
		Order:      channel.Order,
		CreatedAt:  channel.CreatedAt,
	}
	if channel.CategoryID != nil {
		resp.CategoryID = channel.CategoryID.Hex()
	} else {
		resp.CategoryID = ""
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Channel berhasil diambil", resp)
}

// @Summary Get all channels
// @Description Retrieves a list of all channels.
// @Tags Channels
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]dto.ChannelResponse} "Successfully retrieved all channels"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels [get]
func (h *channelHandlerImpl) GetAllChannels(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	channels, err := h.repo.GetAllChannels(ctx)
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua channel dari repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil semua channel", err.Error())
	}

	var resp []dto.ChannelResponse
	for _, ch := range channels {
		channelResp := dto.ChannelResponse{
			ID:         ch.ID.Hex(),
			BusinessID: ch.BusinessID.Hex(),
			Name:       ch.Name,
			Type:       ch.Type,
			Order:      ch.Order,
			CreatedAt:  ch.CreatedAt,
		}
		if ch.CategoryID != nil {
			channelResp.CategoryID = ch.CategoryID.Hex()
		} else {
			channelResp.CategoryID = ""
		}
		resp = append(resp, channelResp)
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Daftar channel berhasil diambil", resp)
}

// @Summary Update a channel by ID
// @Description Updates an existing channel with the provided details.
// @Tags Channels
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param channel body dto.ChannelUpdateRequest true "Channel object to be updated"
// @Success 200 {object} utils.APIResponse{data=dto.ChannelResponse} "Successfully updated channel"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels/{id} [put]
func (h *channelHandlerImpl) UpdateChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak boleh kosong", nil)
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", err.Error())
	}

	var req dto.ChannelUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk update channel")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	updateMap := bson.M{}
	setMap := bson.M{}

	if req.Name != "" {
		setMap["name"] = req.Name
	}
	if req.Type != "" {
		setMap["type"] = req.Type
	}
	if req.CategoryID != "" {
		categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Category ID tidak valid", err.Error())
		}
		setMap["categoryId"] = &categoryID // Assign address of ObjectID
	} else if c.Method() == fiber.MethodPut { // For PUT requests, if CategoryID is explicitly empty, remove it.
		// If CategoryID is an empty string in PUT request, explicitly unset it in MongoDB
		updateMap["$unset"] = bson.M{"categoryId": ""}
	} else {
		// For PATCH or if not present in PUT, do nothing.
		// If req.CategoryID is an empty string and it's not a PUT, don't update.
	}
	if req.Order != 0 {
		setMap["order"] = req.Order
	}
	// Unread dihapus

	if len(setMap) > 0 {
		updateMap["$set"] = setMap
	}

	if len(updateMap) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Tidak ada data untuk diperbarui", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedChannel, err := h.repo.UpdateChannel(ctx, objectID, updateMap)
	if err != nil {
		utils.LogError(err, "Gagal memperbarui channel di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui channel", err.Error())
	}

	if updatedChannel == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Channel tidak ditemukan untuk diperbarui", nil)
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Updated channel: %s (ID: %s)", updatedChannel.Name, id), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	resp := dto.ChannelResponse{
		ID:         updatedChannel.ID.Hex(),
		BusinessID: updatedChannel.BusinessID.Hex(),
		Name:       updatedChannel.Name,
		Type:       updatedChannel.Type,
		Order:      updatedChannel.Order,
		CreatedAt:  updatedChannel.CreatedAt,
	}
	if updatedChannel.CategoryID != nil {
		resp.CategoryID = updatedChannel.CategoryID.Hex()
	} else {
		resp.CategoryID = ""
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Channel berhasil diperbarui", resp)
}

// @Summary Delete a channel by ID
// @Description Deletes a channel by its ID.
// @Tags Channels
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} utils.APIResponse "Successfully deleted channel"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels/{id} [delete]
func (h *channelHandlerImpl) DeleteChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak boleh kosong", nil)
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID tidak valid", err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Get channel before deleting to log its name
	channelToDelete, err := h.repo.GetChannelByID(ctx, objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Channel tidak ditemukan", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memeriksa channel untuk dihapus", err.Error())
	}
	if channelToDelete == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Channel tidak ditemukan", nil)
	}

	if err := h.repo.DeleteChannel(ctx, objectID); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Channel dengan ID %s tidak ditemukan untuk dihapus (handler)", id)
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Channel tidak ditemukan", nil)
		}
		utils.LogError(err, "Gagal menghapus channel di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus channel", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Deleted channel: %s (ID: %s)", channelToDelete.Name, id), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Channel berhasil dihapus", nil)
}

// @Summary Get all channels by businessId
// @Description Retrieves a list of all channels for a specific business.
// @Tags Channels
// @Produce json
// @Param businessId path string true "Business ID"
// @Success 200 {object} utils.APIResponse{data=[]dto.ChannelResponse} "Successfully retrieved all channels by businessId"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid businessId"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channels/business/{businessId} [get]
func (h *channelHandlerImpl) GetChannelsByBusinessID(c *fiber.Ctx) error {
	businessId := c.Params("businessId")
	if businessId == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Business ID tidak boleh kosong", nil)
	}

	objectID, err := primitive.ObjectIDFromHex(businessId)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Business ID tidak valid", err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	channels, err := h.repo.GetChannelsByBusinessID(ctx, objectID)
	if err != nil {
		utils.LogError(err, "Gagal mengambil channel by businessId dari repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil channel by businessId", err.Error())
	}

	var resp []dto.ChannelResponse
	for _, ch := range channels {
		channelResp := dto.ChannelResponse{
			ID:         ch.ID.Hex(),
			BusinessID: ch.BusinessID.Hex(),
			Name:       ch.Name,
			Type:       ch.Type,
			Order:      ch.Order,
			CreatedAt:  ch.CreatedAt,
		}
		if ch.CategoryID != nil {
			channelResp.CategoryID = ch.CategoryID.Hex()
		} else {
			channelResp.CategoryID = ""
		}
		resp = append(resp, channelResp)
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Daftar channel by businessId berhasil diambil", resp)
}

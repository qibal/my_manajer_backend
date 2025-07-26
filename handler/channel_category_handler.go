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

// ChannelCategoryHandler adalah interface untuk handler ChannelCategory.
type ChannelCategoryHandler interface {
	CreateChannelCategory(c *fiber.Ctx) error
	GetChannelCategoryByID(c *fiber.Ctx) error
	GetAllChannelCategories(c *fiber.Ctx) error
	UpdateChannelCategory(c *fiber.Ctx) error
	DeleteChannelCategory(c *fiber.Ctx) error
	GetChannelCategoriesByBusinessID(c *fiber.Ctx) error // Tambahan
}

// channelCategoryHandlerImpl adalah implementasi dari ChannelCategoryHandler.
type channelCategoryHandlerImpl struct {
	repo               repository.ChannelCategoryRepository
	activityLogService service.ActivityLogService
}

// NewChannelCategoryHandler membuat instance baru dari ChannelCategoryHandler.
func NewChannelCategoryHandler(repo repository.ChannelCategoryRepository, activityLogService service.ActivityLogService) ChannelCategoryHandler {
	return &channelCategoryHandlerImpl{
		repo:               repo,
		activityLogService: activityLogService,
	}
}

// @Summary Create a new channel category
// @Description Creates a new channel category with the provided details.
// @Tags Channel Categories
// @Accept json
// @Produce json
// @Param category body dto.ChannelCategoryCreateRequest true "Channel category object to be created"
// @Success 201 {object} utils.APIResponse{data=dto.ChannelCategoryResponse} "Successfully created channel category"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories [post]
func (h *channelCategoryHandlerImpl) CreateChannelCategory(c *fiber.Ctx) error {
	var req dto.ChannelCategoryCreateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk membuat kategori channel")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	// TODO: Implementasi validasi DTO menggunakan library validator.

	businessID, err := primitive.ObjectIDFromHex(req.BusinessID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Business ID tidak valid", err.Error())
	}

	category := &model.ChannelCategory{
		BusinessID: businessID, // Tambahkan BusinessID
		Name:       req.Name,
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.repo.CreateChannelCategory(ctx, category); err != nil {
		utils.LogError(err, "Gagal membuat kategori channel di repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat kategori channel", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Created channel category '%s' in business %s", category.Name, req.BusinessID), c.Method(), c.Path(), fiber.StatusCreated, c.IP())

	resp := dto.ChannelCategoryResponse{
		ID:         category.ID.Hex(),
		BusinessID: category.BusinessID.Hex(), // Tambahkan BusinessID ke response
		Name:       category.Name,
		CreatedAt:  category.CreatedAt,
		UpdatedAt:  category.UpdatedAt,
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Kategori channel berhasil dibuat", resp)
}

// @Summary Get a channel category by ID
// @Description Retrieves a single channel category by its ID.
// @Tags Channel Categories
// @Produce json
// @Param id path string true "Channel Category ID"
// @Success 200 {object} utils.APIResponse{data=dto.ChannelCategoryResponse} "Successfully retrieved channel category"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel category not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories/{id} [get]
func (h *channelCategoryHandlerImpl) GetChannelCategoryByID(c *fiber.Ctx) error {
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

	category, err := h.repo.GetChannelCategoryByID(ctx, objectID)
	if err != nil {
		utils.LogError(err, "Gagal mengambil kategori channel berdasarkan ID dari repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil kategori channel", err.Error())
	}

	if category == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Kategori channel tidak ditemukan", nil)
	}

	resp := dto.ChannelCategoryResponse{
		ID:         category.ID.Hex(),
		BusinessID: category.BusinessID.Hex(), // Tambahkan BusinessID ke response
		Name:       category.Name,
		CreatedAt:  category.CreatedAt,
		UpdatedAt:  category.UpdatedAt,
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Kategori channel berhasil diambil", resp)
}

// @Summary Get all channel categories
// @Description Retrieves a list of all channel categories.
// @Tags Channel Categories
// @Produce json
// @Success 200 {object} utils.APIResponse{data=[]dto.ChannelCategoryResponse} "Successfully retrieved all channel categories"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories [get]
func (h *channelCategoryHandlerImpl) GetAllChannelCategories(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	categories, err := h.repo.GetAllChannelCategories(ctx)
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua kategori channel dari repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil semua kategori channel", err.Error())
	}

	var resp []dto.ChannelCategoryResponse
	for _, cat := range categories {
		resp = append(resp, dto.ChannelCategoryResponse{
			ID:         cat.ID.Hex(),
			BusinessID: cat.BusinessID.Hex(), // Tambahkan BusinessID ke response
			Name:       cat.Name,
			CreatedAt:  cat.CreatedAt,
			UpdatedAt:  cat.UpdatedAt,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Daftar kategori channel berhasil diambil", resp)
}

// @Summary Get all channel categories by businessId
// @Description Retrieves a list of all channel categories for a specific business.
// @Tags Channel Categories
// @Produce json
// @Param businessId path string true "Business ID"
// @Success 200 {object} utils.APIResponse{data=[]dto.ChannelCategoryResponse} "Successfully retrieved all channel categories by businessId"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid businessId"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories/business/{businessId} [get]
func (h *channelCategoryHandlerImpl) GetChannelCategoriesByBusinessID(c *fiber.Ctx) error {
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

	categories, err := h.repo.GetChannelCategoriesByBusinessID(ctx, objectID)
	if err != nil {
		utils.LogError(err, "Gagal mengambil kategori channel by businessId dari repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil kategori channel by businessId", err.Error())
	}

	var resp []dto.ChannelCategoryResponse
	for _, cat := range categories {
		resp = append(resp, dto.ChannelCategoryResponse{
			ID:         cat.ID.Hex(),
			BusinessID: cat.BusinessID.Hex(),
			Name:       cat.Name,
			CreatedAt:  cat.CreatedAt,
			UpdatedAt:  cat.UpdatedAt,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Daftar kategori channel by businessId berhasil diambil", resp)
}

// @Summary Update a channel category by ID
// @Description Updates an existing channel category with the provided details.
// @Tags Channel Categories
// @Accept json
// @Produce json
// @Param id path string true "Channel Category ID"
// @Param category body dto.ChannelCategoryUpdateRequest true "Channel category object to be updated"
// @Success 200 {object} utils.APIResponse{data=dto.ChannelCategoryResponse} "Successfully updated channel category"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel category not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories/{id} [put]
func (h *channelCategoryHandlerImpl) UpdateChannelCategory(c *fiber.Ctx) error {
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

	var req dto.ChannelCategoryUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk update kategori channel")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	// TODO: Implementasi validasi DTO menggunakan library validator.

	updateMap := bson.M{}
	setMap := bson.M{}

	if req.Name != "" {
		setMap["name"] = req.Name
	}

	if len(setMap) > 0 {
		updateMap["$set"] = setMap
	}

	if len(updateMap) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Tidak ada data untuk diperbarui", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedCategory, err := h.repo.UpdateChannelCategory(ctx, objectID, updateMap)
	if err != nil {
		utils.LogError(err, "Gagal memperbarui kategori channel di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui kategori channel", err.Error())
	}

	if updatedCategory == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Kategori channel tidak ditemukan untuk diperbarui", nil)
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Updated channel category: %s (ID: %s)", updatedCategory.Name, id), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	resp := dto.ChannelCategoryResponse{
		ID:        updatedCategory.ID.Hex(),
		Name:      updatedCategory.Name,
		CreatedAt: updatedCategory.CreatedAt,
		UpdatedAt: updatedCategory.UpdatedAt,
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Kategori channel berhasil diperbarui", resp)
}

// @Summary Delete a channel category by ID
// @Description Deletes a channel category by its ID.
// @Tags Channel Categories
// @Produce json
// @Param id path string true "Channel Category ID"
// @Success 200 {object} utils.APIResponse "Successfully deleted channel category"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Channel category not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /channel-categories/{id} [delete]
func (h *channelCategoryHandlerImpl) DeleteChannelCategory(c *fiber.Ctx) error {
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

	categoryToDelete, err := h.repo.GetChannelCategoryByID(ctx, objectID)
	if err != nil || categoryToDelete == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Kategori channel tidak ditemukan", err)
	}

	if err := h.repo.DeleteChannelCategory(ctx, objectID); err != nil {
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Kategori channel dengan ID %s tidak ditemukan untuk dihapus (handler)", id)
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Kategori channel tidak ditemukan", nil)
		}
		utils.LogError(err, "Gagal menghapus kategori channel di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus kategori channel", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Deleted channel category: %s (ID: %s)", categoryToDelete.Name, id), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Kategori channel berhasil dihapus", nil)
}

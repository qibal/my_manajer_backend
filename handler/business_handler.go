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

// BusinessHandler adalah interface untuk handler Business.
type BusinessHandler interface {
	CreateBusiness(c *fiber.Ctx) error
	GetBusinessByID(c *fiber.Ctx) error
	GetAllBusinesses(c *fiber.Ctx) error
	UpdateBusiness(c *fiber.Ctx) error
	DeleteBusiness(c *fiber.Ctx) error
}

// businessHandlerImpl adalah implementasi dari BusinessHandler.
type businessHandlerImpl struct {
	repo               repository.BusinessRepository
	activityLogService service.ActivityLogService
}

// NewBusinessHandler membuat instance baru dari BusinessHandler.
func NewBusinessHandler(repo repository.BusinessRepository, activityLogService service.ActivityLogService) BusinessHandler {
	return &businessHandlerImpl{
		repo:               repo,
		activityLogService: activityLogService,
	}
}

// @Summary Create a new business
// @Description Creates a new business with the provided details.
// @Tags Businesses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param business body dto.BusinessCreateRequest true "Business object to be created"
// @Success 201 {object} utils.APIResponse{data=dto.BusinessResponse} "Successfully created business"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /businesses [post]
func (h *businessHandlerImpl) CreateBusiness(c *fiber.Ctx) error {
	var req dto.BusinessCreateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk membuat bisnis")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	ownerID, ok := c.Locals("userID").(string)
	if !ok || ownerID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	// TODO: Implementasi validasi DTO menggunakan library validator (misal: go-playground/validator).
	// Saat ini, validasi hanya berdasarkan tag `validate` di struct DTO.

	business := &model.Business{
		// ID:        utils.GenerateRandomID(), // Akan dibuat di utils/constants.go
		Name:      req.Name,
		OwnerID:   ownerID,
		Avatar:    req.Avatar,
		Settings:  model.BusinessSettings{}, // Inisialisasi pengaturan default
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.repo.CreateBusiness(ctx, business); err != nil {
		utils.LogError(err, "Gagal membuat bisnis di repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat bisnis", err.Error())
	}

	// Log activity
	go h.activityLogService.LogActivity(context.Background(), ownerID, fmt.Sprintf("Created business: %s", business.Name), c.Method(), c.Path(), fiber.StatusCreated, c.IP())

	// Mengkonversi model.Business ke dto.BusinessResponse untuk respons
	resp := dto.BusinessResponse{
		ID:        business.ID.Hex(),
		Name:      business.Name,
		OwnerID:   business.OwnerID,
		CreatedAt: business.CreatedAt,
		UpdatedAt: business.UpdatedAt,
		Settings: dto.BusinessSettings{
			Theme:         business.Settings.Theme,
			Notifications: business.Settings.Notifications,
		},
		Avatar: business.Avatar,
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Bisnis berhasil dibuat", resp)
}

// @Summary Get a business by ID
// @Description Retrieves a single business by its ID.
// @Tags Businesses
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Business ID"
// @Success 200 {object} utils.APIResponse{data=dto.BusinessResponse} "Successfully retrieved business"
// @Failure 404 {object} utils.APIResponse "Not Found - Business not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /businesses/{id} [get]
func (h *businessHandlerImpl) GetBusinessByID(c *fiber.Ctx) error {
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

	business, err := h.repo.GetBusinessByID(ctx, objectID)
	if err != nil {
		utils.LogError(err, "Gagal mengambil bisnis berdasarkan ID dari repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil bisnis", err.Error())
	}

	if business == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Bisnis tidak ditemukan", nil)
	}

	// Mengkonversi model.Business ke dto.BusinessResponse untuk respons
	resp := dto.BusinessResponse{
		ID:        business.ID.Hex(),
		Name:      business.Name,
		OwnerID:   business.OwnerID,
		CreatedAt: business.CreatedAt,
		UpdatedAt: business.UpdatedAt,
		Settings: dto.BusinessSettings{
			Theme:         business.Settings.Theme,
			Notifications: business.Settings.Notifications,
		},
		Avatar: business.Avatar,
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Bisnis berhasil diambil", resp)
}

// @Summary Get all businesses
// @Description Retrieves a list of all businesses.
// @Tags Businesses
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.APIResponse{data=[]dto.BusinessResponse} "Successfully retrieved all businesses"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /businesses [get]
func (h *businessHandlerImpl) GetAllBusinesses(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	businesses, err := h.repo.GetAllBusinesses(ctx)
	if err != nil {
		utils.LogError(err, "Gagal mengambil semua bisnis dari repository")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil semua bisnis", err.Error())
	}

	// Mengkonversi []model.Business ke []dto.BusinessResponse
	var resp []dto.BusinessResponse
	for _, b := range businesses {
		resp = append(resp, dto.BusinessResponse{
			ID:        b.ID.Hex(),
			Name:      b.Name,
			OwnerID:   b.OwnerID,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
			Settings: dto.BusinessSettings{
				Theme:         b.Settings.Theme,
				Notifications: b.Settings.Notifications,
			},
			Avatar: b.Avatar,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Daftar bisnis berhasil diambil", resp)
}

// @Summary Update a business by ID
// @Description Updates an existing business with the provided details.
// @Tags Businesses
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Business ID"
// @Param business body dto.BusinessUpdateRequest true "Business object to be updated"
// @Success 200 {object} utils.APIResponse{data=dto.BusinessResponse} "Successfully updated business"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 404 {object} utils.APIResponse "Not Found - Business not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /businesses/{id} [put]
func (h *businessHandlerImpl) UpdateBusiness(c *fiber.Ctx) error {
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

	var req dto.BusinessUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		utils.LogError(err, "Gagal parse body request untuk update bisnis")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Input tidak valid", "Format body request salah")
	}

	// TODO: Implementasi validasi DTO menggunakan library validator.

	// Buat update map untuk MongoDB. Hanya field yang ada di DTO yang akan diupdate.
	updateMap := bson.M{}
	setMap := bson.M{}

	if req.Name != "" {
		setMap["name"] = req.Name
	}
	if req.Avatar != "" {
		setMap["avatar"] = req.Avatar
	}

	// Jika ada field yang akan diset, tambahkan $set operator.
	if len(setMap) > 0 {
		updateMap["$set"] = setMap
	}

	// Jika tidak ada field yang diupdate, kembalikan BadRequest
	if len(updateMap) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Tidak ada data untuk diperbarui", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedBusiness, err := h.repo.UpdateBusiness(ctx, objectID, updateMap)
	if err != nil {
		utils.LogError(err, "Gagal memperbarui bisnis di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui bisnis", err.Error())
	}

	if updatedBusiness == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Bisnis tidak ditemukan untuk diperbarui", nil)
	}

	// Log activity
	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Updated business: %s", updatedBusiness.Name), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	// Mengkonversi model.Business ke dto.BusinessResponse untuk respons
	resp := dto.BusinessResponse{
		ID:        updatedBusiness.ID.Hex(),
		Name:      updatedBusiness.Name,
		OwnerID:   updatedBusiness.OwnerID,
		CreatedAt: updatedBusiness.CreatedAt,
		UpdatedAt: updatedBusiness.UpdatedAt,
		Settings: dto.BusinessSettings{
			Theme:         updatedBusiness.Settings.Theme,
			Notifications: updatedBusiness.Settings.Notifications,
		},
		Avatar: updatedBusiness.Avatar,
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Bisnis berhasil diperbarui", resp)
}

// @Summary Delete a business by ID
// @Description Deletes a business by its ID.
// @Tags Businesses
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Business ID"
// @Success 200 {object} utils.APIResponse "Successfully deleted business"
// @Failure 404 {object} utils.APIResponse "Not Found - Business not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /businesses/{id} [delete]
func (h *businessHandlerImpl) DeleteBusiness(c *fiber.Ctx) error {
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

	// Get business before deleting to log its name
	businessToDelete, err := h.repo.GetBusinessByID(ctx, objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Bisnis tidak ditemukan", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memeriksa bisnis untuk dihapus", err.Error())
	}
	if businessToDelete == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Bisnis tidak ditemukan", nil)
	}

	if err := h.repo.DeleteBusiness(ctx, objectID); err != nil {
		// Khusus untuk error mongo.ErrNoDocuments dari repository DeleteBusiness
		if err == mongo.ErrNoDocuments {
			utils.LogWarning("Bisnis dengan ID %s tidak ditemukan untuk dihapus (handler)", id)
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Bisnis tidak ditemukan", nil)
		}
		utils.LogError(err, "Gagal menghapus bisnis di repository: %s", id)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus bisnis", err.Error())
	}

	// Log activity
	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Deleted business: %s", businessToDelete.Name), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Bisnis berhasil dihapus", nil)
}

/*
Cara Penggunaan (dalam router/business_router.go):

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"backend_my_manajer/config"
	"backend_my_manajer/repository"
	"backend_my_manajer/handler"
)

func SetupBusinessRoutes(app *fiber.App, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler
	businessRepo := repository.NewBusinessRepository(dbClient)
	businessHandler := handler.NewBusinessHandler(businessRepo)

	// Grup route untuk bisnis
	businessRoutes := app.Group("/businesses")

	// Contoh endpoint:
	businessRoutes.Post("/", businessHandler.CreateBusiness)
	businessRoutes.Get("/:id", businessHandler.GetBusinessByID)
	businessRoutes.Get("/", businessHandler.GetAllBusinesses)
	businessRoutes.Put("/:id", businessHandler.UpdateBusiness)
	businessRoutes.Delete("/:id", businessHandler.DeleteBusiness)

	// Contoh penggunaan middleware pada route tertentu:
	// businessRoutes.Post("/", middleware.AuthMiddleware(), businessHandler.CreateBusiness)
	// Asumsi AuthMiddleware() adalah fungsi middleware yang sudah Anda buat di paket middleware.
}
*/

package handler

import (
	"context"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SuperAdminHandler menangani logika terkait super admin.
type SuperAdminHandler interface {
	CreateSuperAdmin(c *fiber.Ctx) error
}

type superAdminHandlerImpl struct {
	userRepo repository.UserRepository
}

// NewSuperAdminHandler membuat instance baru dari SuperAdminHandler.
func NewSuperAdminHandler(userRepo repository.UserRepository) SuperAdminHandler {
	return &superAdminHandlerImpl{userRepo: userRepo}
}

// CreateSuperAdmin creates the initial super admin account.
// @Summary Create Super Admin
// @Description Creates the very first super admin account. This can only be done once.
// @Tags SuperAdmin
// @Accept json
// @Produce json
// @Param superadmin body dto.CreateSuperAdminRequest true "Super Admin Creation Details"
// @Success 201 {object} utils.APIResponse{data=dto.CreateSuperAdminResponse} "Super admin created successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 403 {object} utils.APIResponse "Forbidden - Super admin already exists"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /superadmin/create [post]
func (h *superAdminHandlerImpl) CreateSuperAdmin(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	exists, err := h.userRepo.IsSuperAdminExists(ctx)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error checking for super admin", err.Error())
	}
	if exists {
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Super admin already exists", nil)
	}

	var req dto.CreateSuperAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if req.Username == "" || req.Password == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Username and password are required", nil)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	superAdmin := &model.User{
		ID:           primitive.NewObjectID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Avatar:       req.Avatar,
		Status:       "active",
		IsActive:     true,
		Roles: map[string][]string{
			"system": {"super_admin"}, // Peran super admin
		},
		CreatedAt:   time.Now(),
		BusinessIDs: []string{},
	}

	if err := h.userRepo.CreateUser(ctx, superAdmin); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create super admin", err.Error())
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Super admin created successfully", dto.CreateSuperAdminResponse{
		ID:       superAdmin.ID.Hex(),
		Username: superAdmin.Username,
		Email:    superAdmin.Email,
	})
}

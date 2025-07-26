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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoleHandler struct {
	repo               *repository.RoleRepository
	activityLogService service.ActivityLogService
}

func NewRoleHandler(repo *repository.RoleRepository, activityLogService service.ActivityLogService) *RoleHandler {
	return &RoleHandler{
		repo:               repo,
		activityLogService: activityLogService,
	}
}

// CreateRole handles the creation of a new role.
// @Summary Create a new role
// @Description Create a new role with specified name, business ID, and permissions.
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param role body dto.CreateRoleRequest true "Role creation request"
// @Success 201 {object} utils.APIResponse{data=dto.RoleResponse} "Role created successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var req dto.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	// Validasi input manual (mengikuti pola business_handler.go)
	if req.Name == "" || req.BusinessID == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Nama dan BusinessID diperlukan", nil)
	}

	businessObjectID, err := primitive.ObjectIDFromHex(req.BusinessID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Business ID tidak valid", err.Error())
	}

	role := &model.Role{
		BusinessID:  businessObjectID,
		Name:        req.Name,
		Permissions: model.RolePermissions(req.Permissions),
	}

	err = h.repo.CreateRole(ctx, role)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat role", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Created role '%s' in business %s", role.Name, req.BusinessID), c.Method(), c.Path(), fiber.StatusCreated, c.IP())

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "Role berhasil dibuat", dto.RoleResponse{
		ID:          role.ID.Hex(),
		BusinessID:  role.BusinessID.Hex(),
		Name:        role.Name,
		Permissions: role.Permissions,
	})
}

// GetRoleByID retrieves a role by its ID.
// @Summary Get role by ID
// @Description Get role details by its unique ID.
// @Tags Roles
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID"
// @Success 200 {object} utils.APIResponse{data=dto.RoleResponse} "Role retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Role not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRoleByID(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roleIDStr := c.Params("id")
	if roleIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Role ID diperlukan", nil)
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID role tidak valid", err.Error())
	}

	role, err := h.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil role", err.Error())
	}
	if role == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Role tidak ditemukan", nil)
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Role berhasil diambil", dto.RoleResponse{
		ID:          role.ID.Hex(),
		BusinessID:  role.BusinessID.Hex(),
		Name:        role.Name,
		Permissions: role.Permissions,
	})
}

// GetAllRoles retrieves all roles.
// @Summary Get all roles
// @Description Get a list of all roles.
// @Tags Roles
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.APIResponse{data=[]dto.RoleResponse} "Roles retrieved successfully"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /roles [get]
func (h *RoleHandler) GetAllRoles(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roles, err := h.repo.GetAllRoles(ctx)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil roles", err.Error())
	}

	var roleResponses []dto.RoleResponse
	for _, role := range roles {
		roleResponses = append(roleResponses, dto.RoleResponse{
			ID:          role.ID.Hex(),
			BusinessID:  role.BusinessID.Hex(),
			Name:        role.Name,
			Permissions: role.Permissions,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Roles berhasil diambil", roleResponses)
}

// UpdateRole updates an existing role.
// @Summary Update a role
// @Description Update an existing role's name and permissions by ID.
// @Tags Roles
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID"
// @Param role body dto.UpdateRoleRequest true "Role update request"
// @Success 200 {object} utils.APIResponse "Role updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input or ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Role not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roleIDStr := c.Params("id")
	if roleIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Role ID diperlukan", nil)
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID role tidak valid", err.Error())
	}

	var req dto.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi input manual
	if req.Name == "" && req.Permissions == nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Tidak ada data untuk diperbarui", nil)
	}

	existingRole, err := h.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil role", err.Error())
	}
	if existingRole == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Role tidak ditemukan", nil)
	}

	// Update fields
	if req.Name != "" {
		existingRole.Name = req.Name
	}
	if req.Permissions != nil {
		existingRole.Permissions = req.Permissions
	}

	err = h.repo.UpdateRole(ctx, roleID, existingRole)
	if err != nil {
		if err == mongo.ErrNoDocuments { // Jika UpdateOne tidak menemukan dokumen
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Role tidak ditemukan", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui role", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Updated role: %s (ID: %s)", existingRole.Name, roleIDStr), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Role berhasil diperbarui", nil)
}

// DeleteRole deletes a role by its ID.
// @Summary Delete a role
// @Description Delete a role by its unique ID.
// @Tags Roles
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Role ID"
// @Success 204 {object} nil "Role deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 404 {object} utils.APIResponse "Not Found - Role not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	roleIDStr := c.Params("id")
	if roleIDStr == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Role ID diperlukan", nil)
	}

	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User ID tidak ditemukan di token", nil)
	}

	roleID, err := primitive.ObjectIDFromHex(roleIDStr)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "ID role tidak valid", err.Error())
	}

	existingRole, err := h.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil role", err.Error())
	}
	if existingRole == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Role tidak ditemukan", nil)
	}

	err = h.repo.DeleteRole(ctx, roleID)
	if err != nil {
		if err == mongo.ErrNoDocuments { // Jika DeleteOne tidak menemukan dokumen
			return utils.SendErrorResponse(c, fiber.StatusNotFound, "Role tidak ditemukan", nil)
		}
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus role", err.Error())
	}

	go h.activityLogService.LogActivity(context.Background(), userID, fmt.Sprintf("Deleted role: %s (ID: %s)", existingRole.Name, roleIDStr), c.Method(), c.Path(), fiber.StatusOK, c.IP())

	return c.SendStatus(fiber.StatusNoContent) // 204 No Content for successful deletion
}

package handler

import (
	"context"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserHandler menangani logika terkait pengguna.
type UserHandler interface {
	RegisterUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type userHandlerImpl struct {
	userRepo repository.UserRepository
}

// NewUserHandler membuat instance baru dari UserHandler.
func NewUserHandler(userRepo repository.UserRepository) UserHandler {
	return &userHandlerImpl{userRepo: userRepo}
}

// RegisterUser registers a new user (admin only).
// @Summary Register a new user
// @Description Creates a new user account. This endpoint is protected and requires admin privileges.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body dto.RegisterUserRequest true "User Registration Details"
// @Success 201 {object} utils.APIResponse{data=dto.UserResponse} "User registered successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Admin access required"
// @Failure 409 {object} utils.APIResponse "Conflict - User already exists"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /users/register [post]
func (h *userHandlerImpl) RegisterUser(c *fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	// Validasi dasar
	if req.Username == "" || req.Password == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Username and password are required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Cek apakah username sudah ada
	existingUserByUsername, err := h.userRepo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error checking for existing username", err.Error())
	}
	if existingUserByUsername != nil {
		return utils.SendErrorResponse(c, fiber.StatusConflict, "User with this username already exists", nil)
	}

	// Cek apakah email sudah ada, jika email disediakan
	if req.Email != "" {
		existingUserByEmail, err := h.userRepo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error checking for existing email", err.Error())
		}
		if existingUserByEmail != nil {
			return utils.SendErrorResponse(c, fiber.StatusConflict, "User with this email already exists", nil)
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to hash password", err.Error())
	}

	newUser := &model.User{
		ID:           primitive.NewObjectID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Avatar:       req.Avatar,
		Status:       "offline",
		IsActive:     true,
		Roles:        make(map[string][]string), // Default roles
		CreatedAt:    time.Now(),
		BusinessIDs:  make([]string, 0),
	}

	if err := h.userRepo.CreateUser(ctx, newUser); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user", err.Error())
	}

	return utils.SendSuccessResponse(c, fiber.StatusCreated, "User registered successfully", dto.UserResponse{
		ID:       newUser.ID.Hex(),
		Username: newUser.Username,
		Email:    newUser.Email,
		Avatar:   newUser.Avatar,
		IsActive: newUser.IsActive,
	})
}

// GetAllUsers retrieves all users (admin only).
// @Summary Get all users
// @Description Retrieves a list of all users. Requires admin privileges.
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} utils.APIResponse{data=[]dto.UserResponse} "Successfully retrieved all users"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /users [get]
func (h *userHandlerImpl) GetAllUsers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userRepo.GetAllUsers(ctx)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get users", err.Error())
	}

	var userResponses []dto.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, dto.UserResponse{
			ID:          user.ID.Hex(),
			Username:    user.Username,
			Email:       user.Email,
			Avatar:      user.Avatar,
			Status:      user.Status,
			IsActive:    user.IsActive,
			Roles:       user.Roles,
			CreatedAt:   user.CreatedAt,
			BusinessIDs: user.BusinessIDs,
		})
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Users retrieved successfully", userResponses)
}

// UpdateUser updates a user's details (admin only).
// @Summary Update a user
// @Description Updates a user's information by their ID. Requires admin privileges.
// @Tags Users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User Update Details"
// @Success 200 {object} utils.APIResponse{data=dto.UserResponse} "User updated successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 404 {object} utils.APIResponse "User not found"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /users/{id} [put]
func (h *userHandlerImpl) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	updateData := bson.M{}
	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}
	if req.IsActive != nil {
		updateData["isActive"] = *req.IsActive
	}
	if req.Roles != nil {
		updateData["roles"] = req.Roles
	}
	if req.BusinessIDs != nil {
		updateData["businessIds"] = req.BusinessIDs
	}

	if len(updateData) == 0 {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "No update data provided", nil)
	}

	updateData["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	updatedUser, err := h.userRepo.UpdateUser(ctx, objectID, updateData)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user", err.Error())
	}
	if updatedUser == nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "User updated successfully", dto.UserResponse{
		ID:          updatedUser.ID.Hex(),
		Username:    updatedUser.Username,
		Email:       updatedUser.Email,
		Avatar:      updatedUser.Avatar,
		Status:      updatedUser.Status,
		IsActive:    updatedUser.IsActive,
		Roles:       updatedUser.Roles,
		CreatedAt:   updatedUser.CreatedAt,
		BusinessIDs: updatedUser.BusinessIDs,
	})
}

// DeleteUser deletes a user (admin only).
// @Summary Delete a user
// @Description Deletes a user by their ID. Requires admin privileges.
// @Tags Users
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} utils.APIResponse "User deleted successfully"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid ID"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /users/{id} [delete]
func (h *userHandlerImpl) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", err.Error())
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := h.userRepo.DeleteUser(ctx, objectID); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete user", err.Error())
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "User deleted successfully", nil)
}

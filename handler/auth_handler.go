package handler

import (
	"context"
	"time"

	"backend_my_manajer/dto"
	"backend_my_manajer/model"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler menangani logika terkait autentikasi.
type AuthHandler interface {
	Login(c *fiber.Ctx) error
}

type authHandlerImpl struct {
	userRepo repository.UserRepository
}

// NewAuthHandler membuat instance baru dari AuthHandler.
func NewAuthHandler(userRepo repository.UserRepository) AuthHandler {
	return &authHandlerImpl{userRepo: userRepo}
}

// Login handles the user login request.
// @Summary User login
// @Description Authenticates a user and returns a JWT token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login Credentials"
// @Success 200 {object} utils.APIResponse{data=dto.LoginResponse} "Login successful"
// @Failure 400 {object} utils.APIResponse "Bad Request - Invalid input"
// @Failure 401 {object} utils.APIResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} utils.APIResponse "Internal Server Error"
// @Router /auth/login [post]
func (h *authHandlerImpl) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if (req.Email == "" && req.Username == "") || req.Password == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Email/Username and password are required", nil)
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var user *model.User
	var err error

	if req.Email != "" {
		user, err = h.userRepo.FindUserByEmail(ctx, req.Email)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error finding user by email", err.Error())
		}
	} else if req.Username != "" {
		user, err = h.userRepo.FindUserByUsername(ctx, req.Username)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Error finding user by username", err.Error())
		}
	}

	if user == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials", nil)
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials", nil)
	}

	token, err := utils.GenerateJWTToken(user.ID.Hex(), user.Email, user.Roles)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate token", err.Error())
	}

	return utils.SendSuccessResponse(c, fiber.StatusOK, "Login successful", dto.LoginResponse{
		Token:       token,
		ID:          user.ID.Hex(),
		Username:    user.Username,
		Email:       user.Email,
		Avatar:      user.Avatar,
		IsActive:    user.IsActive,
		Roles:       user.Roles,
		CreatedAt:   user.CreatedAt,
		BusinessIDs: user.BusinessIDs,
	})
}

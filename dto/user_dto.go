package dto

import "time"

// LoginRequest defines the structure for a login request.
// Can use either email or username.
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse defines the successful login response.
type LoginResponse struct {
	Token       string              `json:"token"`
	ID          string              `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	Avatar      string              `json:"avatar"`
	IsActive    bool                `json:"isActive"`
	Roles       map[string][]string `json:"roles"`
	CreatedAt   time.Time           `json:"createdAt"`
	BusinessIDs []string            `json:"businessIds"`
}

// RegisterUserRequest defines the structure for creating a new user by an admin.
type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Avatar   string `json:"avatar,omitempty"`
}

// UpdateUserRequest defines the structure for updating a user by an admin.
type UpdateUserRequest struct {
	Username    string              `json:"username,omitempty"`
	Email       string              `json:"email,omitempty" validate:"omitempty,email"`
	Avatar      string              `json:"avatar,omitempty"`
	IsActive    *bool               `json:"isActive,omitempty"`
	Roles       map[string][]string `json:"roles,omitempty"`
	BusinessIDs []string            `json:"businessIds,omitempty"`
}

// UserResponse defines the standard user data returned by the API.
type UserResponse struct {
	ID          string              `json:"id"`
	Username    string              `json:"username"`
	Email       string              `json:"email"`
	Avatar      string              `json:"avatar"`
	Status      string              `json:"status"`
	IsActive    bool                `json:"isActive"`
	Roles       map[string][]string `json:"roles"`
	CreatedAt   time.Time           `json:"createdAt"`
	BusinessIDs []string            `json:"businessIds"`
}

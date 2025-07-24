package utils

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse adalah struktur standar untuk respons API.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SendSuccessResponse mengirimkan respons sukses standar.
func SendSuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendErrorResponse mengirimkan respons error standar.
func SendErrorResponse(c *fiber.Ctx, statusCode int, message string, err interface{}) error {
	return c.Status(statusCode).JSON(APIResponse{
		Success: false,
		Message: message,
		Error:   err,
	})
}

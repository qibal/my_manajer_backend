package handler

import (
	"backend_my_manajer/dto"
	"backend_my_manajer/repository"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
)

// ActivityLogHandler menangani logika bisnis yang terkait dengan log aktivitas.
type ActivityLogHandler struct {
	repo repository.ActivityLogRepository
}

// NewActivityLogHandler membuat instance baru dari ActivityLogHandler.
func NewActivityLogHandler(repo repository.ActivityLogRepository) *ActivityLogHandler {
	return &ActivityLogHandler{
		repo: repo,
	}
}

// GetAllActivityLogs menangani permintaan untuk mendapatkan semua log aktivitas.
func (h *ActivityLogHandler) GetAllActivityLogs(c *fiber.Ctx) error {
	logs, err := h.repo.GetAllActivityLogs(c.Context())
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil log aktivitas", err.Error())
	}

	// Konversi model ke DTO untuk respons
	var logResponses []dto.ActivityLogResponse
	for _, log := range logs {
		logResponses = append(logResponses, dto.ActivityLogResponse{
			ID:         log.ID.Hex(),
			UserID:     log.UserID,
			Action:     log.Action,
			Method:     log.Method,
			Endpoint:   log.Endpoint,
			StatusCode: log.StatusCode,
			IPAddress:  log.IPAddress,
			CreatedAt:  log.CreatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(utils.APIResponse{
		Success: true,
		Message: "Log aktivitas berhasil diambil",
		Data:    logResponses,
	})
}

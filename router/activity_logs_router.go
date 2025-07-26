package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupActivityLogRoutes mendaftarkan rute untuk log aktivitas.
func SetupActivityLogRoutes(router fiber.Router, dbClient *mongo.Client) {
	activityLogRepo := repository.NewActivityLogRepository(dbClient)
	activityLogHandler := handler.NewActivityLogHandler(activityLogRepo)

	// Endpoint GET untuk mengambil semua log, dilindungi oleh middleware admin.
	activityLogsRoutes := router.Group("/activity-logs", middleware.AdminAuthMiddleware())
	activityLogsRoutes.Get("/", activityLogHandler.GetAllActivityLogs)
}

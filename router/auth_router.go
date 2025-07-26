package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"
	"backend_my_manajer/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupAuthRoutes mendaftarkan rute untuk autentikasi.
func SetupAuthRoutes(router fiber.Router, dbClient *mongo.Client) {
	userRepo := repository.NewUserRepository(dbClient)
	activityLogRepo := repository.NewActivityLogRepository(dbClient)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	authHandler := handler.NewAuthHandler(userRepo, activityLogService)

	authRoutes := router.Group("/auth")
	authRoutes.Post("/login", authHandler.Login)
}

package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupAuthRoutes mendaftarkan rute untuk autentikasi.
func SetupAuthRoutes(router fiber.Router, dbClient *mongo.Client) {
	userRepo := repository.NewUserRepository(dbClient)
	authHandler := handler.NewAuthHandler(userRepo)

	authRoutes := router.Group("/auth")
	authRoutes.Post("/login", authHandler.Login)
}

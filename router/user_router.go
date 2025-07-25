package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupUserRoutes mendaftarkan rute untuk pengguna.
func SetupUserRoutes(router fiber.Router, dbClient *mongo.Client) {
	userRepo := repository.NewUserRepository(dbClient)
	userHandler := handler.NewUserHandler(userRepo)

	userRoutes := router.Group("/users")
	// Semua rute di bawah ini memerlukan autentikasi admin
	userRoutes.Use(middleware.AdminAuthMiddleware())

	userRoutes.Post("/register", userHandler.RegisterUser)
	userRoutes.Get("/", userHandler.GetAllUsers)
	userRoutes.Put("/:id", userHandler.UpdateUser)
	userRoutes.Delete("/:id", userHandler.DeleteUser)
}

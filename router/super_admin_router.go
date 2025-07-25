package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupSuperAdminRoutes mendaftarkan rute untuk super admin.
func SetupSuperAdminRoutes(router fiber.Router, dbClient *mongo.Client) {
	userRepo := repository.NewUserRepository(dbClient)
	superAdminHandler := handler.NewSuperAdminHandler(userRepo)

	superAdminRoutes := router.Group("/superadmin")
	superAdminRoutes.Post("/create", superAdminHandler.CreateSuperAdmin)
}

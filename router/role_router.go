package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware"
	"backend_my_manajer/repository"
	"backend_my_manajer/service"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoleRoutes mengkonfigurasi rute API untuk entitas Role.
func SetupRoleRoutes(api fiber.Router, dbClient *mongo.Client) {
	repo := repository.NewRoleRepository(dbClient)
	activityLogRepo := repository.NewActivityLogRepository(dbClient)
	activityLogService := service.NewActivityLogService(activityLogRepo)
	handler := handler.NewRoleHandler(repo, activityLogService)

	// Menerapkan middleware autentikasi ke semua rute role
	roles := api.Group("/roles", middleware.AuthMiddleware())
	roles.Post("/", handler.CreateRole)
	roles.Get("/", handler.GetAllRoles)
	roles.Get("/:id", handler.GetRoleByID)
	roles.Put("/:id", handler.UpdateRole)
	roles.Delete("/:id", handler.DeleteRole)
}

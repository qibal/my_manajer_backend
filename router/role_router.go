package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoleRoutes mengkonfigurasi rute API untuk entitas Role.
func SetupRoleRoutes(api fiber.Router, dbClient *mongo.Client) {
	repo := repository.NewRoleRepository(dbClient)
	handler := handler.NewRoleHandler(repo)

	roles := api.Group("/roles")
	roles.Post("/", handler.CreateRole)      // POST /api/v1/roles
	roles.Get("/", handler.GetAllRoles)      // GET /api/v1/roles
	roles.Get("/:id", handler.GetRoleByID)   // GET /api/v1/roles/:id
	roles.Put("/:id", handler.UpdateRole)    // PUT /api/v1/roles/:id
	roles.Delete("/:id", handler.DeleteRole) // DELETE /api/v1/roles/:id
}

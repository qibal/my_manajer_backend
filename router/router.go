package router

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes mendaftarkan semua rute aplikasi.
func SetupRoutes(app *fiber.App, dbClient *mongo.Client) {
	// Membuat grup route utama, misalnya /api/v1
	api := app.Group("/api/v1")

	// Mendaftarkan rute untuk setiap entitas
	SetupAuthRoutes(api, dbClient)
	SetupUserRoutes(api, dbClient)
	SetupSuperAdminRoutes(api, dbClient)
	SetupBusinessRoutes(api, dbClient)
	SetupChannelRoutes(api, dbClient)
	SetupChannelCategoryRoutes(api, dbClient)
	SetupMessageRoutes(api, dbClient)
	SetupRoleRoutes(api, dbClient)
	// Tambahkan setup route lain di sini jika ada
}

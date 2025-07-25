package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware" // Mengimpor middleware
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupChannelCategoryRoutes mendaftarkan semua rute API untuk entitas ChannelCategory.
// router adalah instance Fiber.Router (bisa *fiber.App atau grup rute), dan dbClient adalah koneksi MongoDB.
func SetupChannelCategoryRoutes(router fiber.Router, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler untuk ChannelCategory
	channelCategoryRepo := repository.NewChannelCategoryRepository(dbClient)
	channelCategoryHandler := handler.NewChannelCategoryHandler(channelCategoryRepo)

	// Menerapkan middleware autentikasi ke semua rute kategori channel
	channelCategoryRoutes := router.Group("/channel-categories", middleware.AuthMiddleware())

	// Mendefinisikan endpoint CRUD untuk ChannelCategory
	channelCategoryRoutes.Post("/", channelCategoryHandler.CreateChannelCategory)
	channelCategoryRoutes.Get("/:id", channelCategoryHandler.GetChannelCategoryByID)
	channelCategoryRoutes.Get("/", channelCategoryHandler.GetAllChannelCategories)
	channelCategoryRoutes.Put("/:id", channelCategoryHandler.UpdateChannelCategory)
	channelCategoryRoutes.Delete("/:id", channelCategoryHandler.DeleteChannelCategory)
	channelCategoryRoutes.Get("/business/:businessId", channelCategoryHandler.GetChannelCategoriesByBusinessID) // Tambahan
}

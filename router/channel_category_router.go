package router

import (
	"backend_my_manajer/handler"
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

	// Membuat grup route untuk semua endpoint yang berkaitan dengan ChannelCategory
	// Semua rute di dalam grup ini akan memiliki prefix "/channel-categories"
	channelCategoryRoutes := router.Group("/channel-categories")

	// Mendefinisikan endpoint CRUD untuk ChannelCategory
	channelCategoryRoutes.Post("/", channelCategoryHandler.CreateChannelCategory)
	channelCategoryRoutes.Get("/:id", channelCategoryHandler.GetChannelCategoryByID)
	channelCategoryRoutes.Get("/", channelCategoryHandler.GetAllChannelCategories)
	channelCategoryRoutes.Put("/:id", channelCategoryHandler.UpdateChannelCategory)
	channelCategoryRoutes.Delete("/:id", channelCategoryHandler.DeleteChannelCategory)
	channelCategoryRoutes.Get("/business/:businessId", channelCategoryHandler.GetChannelCategoriesByBusinessID) // Tambahan
}

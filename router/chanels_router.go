package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupChannelRoutes mendaftarkan semua rute API untuk entitas Channel.
// router adalah instance Fiber.Router (bisa *fiber.App atau grup rute), dan dbClient adalah koneksi MongoDB.
func SetupChannelRoutes(router fiber.Router, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler untuk Channel
	channelRepo := repository.NewChannelRepository(dbClient)
	channelHandler := handler.NewChannelHandler(channelRepo)

	// Membuat grup route untuk semua endpoint yang berkaitan dengan Channel
	// Semua rute di dalam grup ini akan memiliki prefix "/channels"
	// Prefix ini akan digabungkan dengan prefix dari router induk (misalnya /api/v1/)
	channelRoutes := router.Group("/channels")

	// Mendefinisikan endpoint CRUD untuk Channel
	channelRoutes.Post("/", channelHandler.CreateChannel)
	channelRoutes.Get("/:id", channelHandler.GetChannelByID)
	channelRoutes.Get("/", channelHandler.GetAllChannels)
	channelRoutes.Put("/:id", channelHandler.UpdateChannel)
	channelRoutes.Delete("/:id", channelHandler.DeleteChannel)
	// Endpoint tambahan: Get channels by businessId
	channelRoutes.Get("/business/:businessId", channelHandler.GetChannelsByBusinessID)

	/*
		Cara Penggunaan Middleware:

		Untuk menambahkan middleware ke route tertentu, Anda bisa menambahkannya sebagai argumen sebelum handler:

		// Contoh: Menerapkan AuthMiddleware hanya untuk route POST (CreateChannel)
		// channelRoutes.Post("/", middleware.AuthMiddleware(), channelHandler.CreateChannel)

		// Contoh: Menerapkan middleware untuk semua route dalam grup Channel
		// channelRoutes := router.Group("/channels", middleware.AuthMiddleware())
		// Ini akan menerapkan AuthMiddleware ke semua rute di atas (POST, GET, PUT, DELETE).

		// Pastikan paket 'middleware' diimpor (uncomment baris di atas):
		// import "backend_my_manajer/middleware"
	*/
}

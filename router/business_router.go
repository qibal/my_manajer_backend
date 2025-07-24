package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/repository"

	// "backend_my_manajer/middleware" // Uncomment jika ingin menggunakan middleware

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupBusinessRoutes mendaftarkan semua rute API untuk entitas Business.
// router adalah instance Fiber.Router (bisa *fiber.App atau grup rute), dan dbClient adalah koneksi MongoDB.
func SetupBusinessRoutes(router fiber.Router, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler untuk Business
	businessRepo := repository.NewBusinessRepository(dbClient)
	businessHandler := handler.NewBusinessHandler(businessRepo)

	// Membuat grup route untuk semua endpoint yang berkaitan dengan Business
	// Semua rute di dalam grup ini akan memiliki prefix "/businesses"
	// Prefix ini akan digabungkan dengan prefix dari router induk (misalnya /api/v1/)
	businessRoutes := router.Group("/businesses")

	// Mendefinisikan endpoint CRUD untuk Business
	// POST /api/v1/businesses
	businessRoutes.Post("/", businessHandler.CreateBusiness)

	// GET /api/v1/businesses/:id
	businessRoutes.Get("/:id", businessHandler.GetBusinessByID)

	// GET /api/v1/businesses
	businessRoutes.Get("/", businessHandler.GetAllBusinesses)

	// PUT /api/v1/businesses/:id
	businessRoutes.Put("/:id", businessHandler.UpdateBusiness)

	// DELETE /api/v1/businesses/:id
	businessRoutes.Delete("/:id", businessHandler.DeleteBusiness)

	/*
		Cara Penggunaan Middleware:

		Untuk menambahkan middleware ke route tertentu, Anda bisa menambahkannya sebagai argumen sebelum handler:

		// Contoh: Menerapkan AuthMiddleware hanya untuk route POST (CreateBusiness)
		// businessRoutes.Post("/", middleware.AuthMiddleware(), businessHandler.CreateBusiness)

		// Contoh: Menerapkan middleware untuk semua route dalam grup Business
		// businessRoutes := router.Group("/businesses", middleware.AuthMiddleware())
		// Ini akan menerapkan AuthMiddleware ke semua rute di atas (POST, GET, PUT, DELETE).

		// Pastikan paket 'middleware' diimpor (uncomment baris di atas):
		// import "backend_my_manajer/middleware"
	*/
}

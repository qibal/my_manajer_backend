package main

import (
	"log"
	"os"

	"backend_my_manajer/config"
	"backend_my_manajer/router"
	"backend_my_manajer/utils"

	_ "backend_my_manajer/docs" // Import generated docs
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger" // Import a swagger
)

// @title My Manajer API
// @version 1.0
// @description This is a sample server for My Manajer API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description "Type 'Bearer' followed by a space and JWT token."
func main() {
	// Muat file .env
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}
	utils.LogInfo("Variabel lingkungan berhasil dimuat dari .env")

	// Inisialisasi koneksi database
	dbClient := config.ConnectDB()
	defer func() {
		if err := dbClient.Disconnect(nil); err != nil {
			log.Fatal(err)
		}
	}()

	// Inisialisasi Fiber app
	app := fiber.New()

	// Menggunakan middleware CORS dengan konfigurasi
	app.Use(cors.New(config.CorsConfig()))

	// Setup rute API
	router.SetupRoutes(app, dbClient)

	// Setup rute untuk Swagger
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}

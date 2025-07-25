package main

import (
	"context"
	"log"
	"os"

	"backend_my_manajer/config"
	"backend_my_manajer/router"
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	_ "backend_my_manajer/docs" // Import generated docs

	fiberSwagger "github.com/swaggo/fiber-swagger"
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
	err := godotenv.Load()
	if err != nil {
		utils.LogError(err, "Error loading .env file")
	}

	dbClient := config.ConnectDB()
	// Pastikan koneksi database ditutup saat aplikasi berhenti
	defer func() {
		if dbClient != nil {
			if err := dbClient.Disconnect(context.TODO()); err != nil {
				utils.LogError(err, "Failed to disconnect database")
			}
			utils.LogInfo("Database connection closed.")
		}
	}()

	app := fiber.New()

	// Gunakan CorsConfig dari paket config
	app.Use(cors.New(config.CorsConfig()))

	router.SetupRoutes(app, dbClient)

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	utils.LogInfo("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

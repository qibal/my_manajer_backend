package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware"
	"backend_my_manajer/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupDatabaseRoutes mendaftarkan rute untuk entitas Database.
func SetupDatabaseRoutes(router fiber.Router, dbClient *mongo.Client) {
	dbRepo := repository.NewDatabaseRepository(dbClient)
	userRepo := repository.NewUserRepository(dbClient) // Digunakan untuk otorisasi di handler
	dbHandler := handler.NewDatabaseHandler(dbRepo, userRepo)

	dbRoutes := router.Group("/databases")

	// Middleware autentikasi untuk semua rute database
	dbRoutes.Use(middleware.AuthMiddleware())

	// Rute CRUD untuk Database
	dbRoutes.Post("/", dbHandler.CreateDatabase)
	dbRoutes.Get("/:id", dbHandler.GetDatabaseByID)
	dbRoutes.Put("/:id", dbHandler.UpdateDatabase)
	dbRoutes.Delete("/:id", dbHandler.DeleteDatabase)

	// Rute untuk mendapatkan database berdasarkan ChannelID
	dbRoutes.Get("/channel/:channelId", dbHandler.GetDatabasesByChannelID)

	// Rute CRUD untuk baris data dalam database
	dbRoutes.Post("/:id/rows", dbHandler.AddRowToDatabase)
	dbRoutes.Put("/:id/rows/:rowId", dbHandler.UpdateRowInDatabase)
	dbRoutes.Delete("/:id/rows/:rowId", dbHandler.DeleteRowFromDatabase)
	dbRoutes.Get("/:id/rows/:rowId", dbHandler.GetRowInDatabase) // Menambahkan rute GET untuk baris tunggal
	dbRoutes.Get("/:id/rows", dbHandler.GetRowsByDatabaseID)     // Menambahkan rute GET untuk semua baris dalam database

	// Rute CRUD untuk kolom dalam database
	dbRoutes.Put("/:id/columns/:columnId", dbHandler.UpdateColumnInDatabase)
	dbRoutes.Delete("/:id/columns/:columnId", dbHandler.DeleteColumnFromDatabase)
	dbRoutes.Get("/:id/columns/:columnId", dbHandler.GetColumnInDatabase) // Menambahkan rute GET untuk kolom

	// Rute CRUD untuk select options dalam kolom
	dbRoutes.Post("/:id/columns/:columnId/options", dbHandler.AddSelectOptionToColumn)
	dbRoutes.Put("/:id/columns/:columnId/options/:optionId", dbHandler.UpdateSelectOptionInColumn)
	dbRoutes.Delete("/:id/columns/:columnId/options/:optionId", dbHandler.DeleteSelectOptionFromColumn)
	dbRoutes.Get("/:id/columns/:columnId/options/:optionId", dbHandler.GetSelectOptionInColumn) // Menambahkan rute GET untuk select option

	// Rute CRUD untuk rows dalam database
}

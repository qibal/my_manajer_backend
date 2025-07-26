package router

import (
	"backend_my_manajer/handler"
	"backend_my_manajer/middleware" // Pastikan middleware diimpor
	"backend_my_manajer/repository"

	"github.com/gofiber/contrib/websocket" // Import Fiber WebSocket
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupMessageRoutes mendaftarkan rute WebSocket untuk entitas Message.
func SetupMessageRoutes(api fiber.Router, dbClient *mongo.Client) {
	// Inisialisasi repository dan handler untuk Message
	messageRepo := repository.NewMessageRepository(dbClient)
	messageHandler := handler.NewMessageHandler(messageRepo)

	// Grup untuk WebSocket dengan middleware autentikasi
	wsGroup := api.Group("/ws", middleware.WebSocketAuthMiddleware())

	// Terapkan middleware ke rute spesifik
	wsGroup.Get("/messages/:channelId", websocket.New(messageHandler.HandleWebSocketMessage))
}

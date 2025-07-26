package middleware

import (
	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
)

// WebSocketAuthMiddleware adalah middleware untuk memvalidasi token JWT dari query param untuk koneksi WebSocket.
// Ini dirancang untuk digunakan SEBELUM websocket.New, dan akan menetapkan hasil autentikasi di c.Locals.
// Ini TIDAK akan mengirim respons error HTTP secara langsung, agar tidak mengganggu handshake WebSocket.
func WebSocketAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.Locals("authFailed", true)
			c.Locals("authError", "Token autentikasi diperlukan")
			return c.Next() // Tetap panggil Next untuk membiarkan handshake WebSocket
		}

		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			c.Locals("authFailed", true)
			c.Locals("authError", "Token tidak valid: "+err.Error())
			return c.Next() // Tetap panggil Next
		}

		c.Locals("authFailed", false)
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRoles", claims.Roles)

		return c.Next() // Lanjutkan ke handler berikutnya (websocket.New)
	}
}

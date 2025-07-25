package middleware

import (
	"strings"

	"backend_my_manajer/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware adalah middleware untuk memvalidasi token JWT.
// Ini akan memastikan bahwa setiap request yang melalui middleware ini memiliki token JWT yang valid.
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token autentikasi diperlukan", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid (Bearer token diperlukan)", nil)
		}

		tokenString := parts[1]

		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid", err.Error())
		}

		// Menyimpan klaim pengguna ke dalam context Fiber
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRoles", claims.Roles)

		// Melanjutkan ke handler berikutnya
		return c.Next()
	}
}

// AdminAuthMiddleware adalah middleware untuk memvalidasi token JWT dan memeriksa peran admin.
func AdminAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Duplikasi logika dari AuthMiddleware untuk validasi token dasar
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token autentikasi diperlukan", nil)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Format token tidak valid (Bearer token diperlukan)", nil)
		}

		tokenString := parts[1]

		claims, err := utils.ValidateJWTToken(tokenString)
		if err != nil {
			return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Token tidak valid", err.Error())
		}

		// Menyimpan klaim pengguna ke dalam context Fiber
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)
		c.Locals("userRoles", claims.Roles)

		// Ambil peran pengguna dari locals yang disimpan
		userRoles, ok := c.Locals("userRoles").(map[string][]string)
		if !ok {
			// Ini terjadi jika token tidak memiliki data peran yang valid.
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "Informasi peran tidak tersedia di dalam token", nil)
		}

		// Periksa apakah pengguna memiliki peran 'super_admin'
		hasAdminRole := false
		for _, roles := range userRoles {
			for _, role := range roles {
				if role == "super_admin" { // MEMPERBAIKI: Memeriksa nama peran yang benar
					hasAdminRole = true
					break
				}
			}
			if hasAdminRole {
				break
			}
		}

		if !hasAdminRole {
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "Akses ditolak: Hanya admin yang diizinkan", nil)
		}

		return c.Next()
	}
}

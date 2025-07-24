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
		// Pertama, jalankan AuthMiddleware dasar
		err := AuthMiddleware()(c)
		if err != nil {
			return err // Mengembalikan error dari AuthMiddleware (misal: token tidak ada/tidak valid)
		}

		// Ambil peran pengguna dari locals yang disimpan oleh AuthMiddleware
		userRoles, ok := c.Locals("userRoles").(map[string][]string)
		if !ok {
			return utils.SendErrorResponse(c, fiber.StatusForbidden, "Informasi peran tidak tersedia", nil)
		}

		// Periksa apakah pengguna memiliki peran 'role_admin' di bisnis manapun
		hasAdminRole := false
		for _, roles := range userRoles {
			for _, role := range roles {
				if role == "role_admin" {
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

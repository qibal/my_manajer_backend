package config

import (
	"os"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CorsConfig menyediakan konfigurasi CORS default untuk aplikasi Fiber.
// Ini membaca daftar origin yang diizinkan dari variabel lingkungan CORS_ALLOW_ORIGINS.
// Jika CORS_ALLOW_ORIGINS kosong, maka akan mengizinkan semua origin (*).
func CorsConfig() cors.Config {
	allowedOriginsEnv := os.Getenv("CORS_ALLOW_ORIGINS")
	var allowedOrigins string
	if allowedOriginsEnv == "" {
		allowedOrigins = "*"
	} else {
		// Mengizinkan multiple origins dipisahkan koma
		allowedOrigins = allowedOriginsEnv
	}

	return cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           300,
		ExposeHeaders:    "Content-Length",
	}
}

/*
// CorsConfigWithOrigins dihapus karena CorsConfig() sudah menangani origin dari .env
func CorsConfigWithOrigins(origins []string) cors.Config {
	return cors.Config{
		AllowOrigins:     strings.Join(origins, ","),
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           300,
		ExposeHeaders:    "Content-Length",
	}
}
*/

/*
Cara Penggunaan (dalam main.go):

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"backend_my_manajer/config"
)

func main() {
	app := fiber.New()

	// Menggunakan konfigurasi CORS dari config.CorsConfig()
	app.Use(cors.New(config.CorsConfig()))

	// ... definisikan route lainnya ...

	// log.Fatal(app.Listen(":3000"))
}
*/

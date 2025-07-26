package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTSecretKey adalah kunci rahasia untuk menandatangani JWT.
// Seharusnya dimuat dari variabel lingkungan atau konfigurasi yang aman.
var JWTSecretKey []byte

// getJWTSecretKey mengambil kunci rahasia JWT dari variabel lingkungan.
// Jika tidak ditemukan, akan menggunakan kunci default dan mencatat peringatan.
func GetJWTSecretKey() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET_KEY tidak ditemukan di environment. Aplikasi tidak bisa berjalan tanpa secret ini")
	}
	return []byte(secret), nil
}

// Claims adalah struktur kustom yang akan digunakan untuk JWT.
type Claims struct {
	UserID string              `json:"user_id"`
	Email  string              `json:"email"`
	Roles  map[string][]string `json:"roles"` // Map businessId to array of role IDs
	jwt.RegisteredClaims
}

// GenerateJWTToken menghasilkan token JWT baru untuk pengguna.
func GenerateJWTToken(userID, email string, roles map[string][]string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token berlaku 24 jam

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "my-manajer-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		LogError(err, "Gagal menandatangani token JWT")
		return "", err
	}

	return tokenString, nil
}

// ValidateJWTToken memvalidasi token JWT dan mengembalikan klaim jika valid.
func ValidateJWTToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTSecretKey, nil
	})

	if err != nil {
		LogError(err, "Gagal mengurai atau memvalidasi token JWT")
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token JWT tidak valid")
	}

	return claims, nil
}

// GetUserIDFromToken mengambil ID pengguna dari token JWT dalam konteks Fiber.
func GetUserIDFromToken(c *fiber.Ctx) (string, error) {
	user := c.Locals("user").(*jwt.Token)
	if user == nil {
		return "", fmt.Errorf("token pengguna tidak ditemukan dalam konteks")
	}

	claims, ok := user.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", fmt.Errorf("klaim token tidak valid")
	}

	return claims.Subject, nil
}

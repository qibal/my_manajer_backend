package utils

import (
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengenkripsi password menggunakan bcrypt.
// Mengembalikan hash password atau error jika terjadi masalah.
// Password yang dihasilkan dari fungsi ini sangat aman untuk disimpan di database.
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost adalah nilai cost default yang direkomendasikan.
	// Nilai cost yang lebih tinggi akan membuat hashing lebih lambat tapi lebih aman.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		LogError(err, "Gagal melakukan hashing password")
		return "", fmt.Errorf("gagal mengenkripsi password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash membandingkan password plain-text dengan hash password yang sudah ada.
// Mengembalikan true jika cocok, false jika tidak, dan error jika ada masalah teknis.
func CheckPasswordHash(password, hash string) bool {
	// bcrypt.CompareHashAndPassword akan mengembalikan nil jika password cocok.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Jika err bukan nil, berarti password tidak cocok atau ada masalah teknis.
		// Kita bisa log error ini, tapi tidak perlu mengembalikan error ke pemanggil
		// karena ini hanya perbandingan.
		LogError(err, "Gagal membandingkan password hash")
		return false
	}
	return true
}

// ValidatePasswordStrength memeriksa kekuatan password berdasarkan kriteria sederhana.
// Mengembalikan true jika password memenuhi kriteria, dan pesan error jika tidak.
// Kriteria:
// - Minimal 8 karakter
// - Mengandung setidaknya satu huruf besar
// - Mengandung setidaknya satu huruf kecil
// - Mengandung setidaknya satu angka
// - Mengandung setidaknya satu karakter spesial
func ValidatePasswordStrength(password string) (bool, string) {
	minLength := 8
	if len(password) < minLength {
		return false, fmt.Sprintf("Password minimal %d karakter", minLength)
	}

	// Regex untuk memeriksa kriteria
	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	// hasSpecialChar := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};'\\|,.<>/?"]`).MatchString(password)

	if !hasUppercase {
		return false, "Password harus mengandung setidaknya satu huruf besar"
	}
	if !hasLowercase {
		return false, "Password harus mengandung setidaknya satu huruf kecil"
	}
	if !hasDigit {
		return false, "Password harus mengandung setidaknya satu angka"
	}


	return true, "Password kuat"
}

/*
Cara Penggunaan:

// Contoh penggunaan HashPassword
// hashedPassword, err := HashPassword("MyStrongP@ssw0rd")
// if err != nil {
//     log.Fatal(err)
// }
// fmt.Println("Hashed Password:", hashedPassword)

// Contoh penggunaan CheckPasswordHash
// isMatch := CheckPasswordHash("MyStrongP@ssw0rd", hashedPassword)
// fmt.Println("Password Match:", isMatch)

// Contoh penggunaan ValidatePasswordStrength
// isValid, message := ValidatePasswordStrength("WeakPass1")
// fmt.Printf("Password Valid: %t, Message: %s\n", isValid, message)

// isValid, message := ValidatePasswordStrength("MyStrongP@ssw0rd")
// fmt.Printf("Password Valid: %t, Message: %s\n", isValid, message)
*/

package utils

import (
	"log"
	"os"
)

// Logger adalah instance logger global yang akan digunakan untuk mencatat pesan.
var Logger *log.Logger

func init() {
	// Inisialisasi Logger saat paket diimpor.
	// Logger akan menulis ke standar output (terminal).
	// log.Ldate|log.Ltime akan menambahkan tanggal dan waktu ke setiap pesan log.
	// log.Lshortfile akan menambahkan nama file dan nomor baris dari mana log dipanggil.
	Logger = log.New(os.Stdout, "APP_LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo mencatat pesan informasi ke terminal.
// Gunakan ini untuk pesan umum atau status yang berhasil.
func LogInfo(message string, args ...interface{}) {
	Logger.Printf("INFO: "+message+"\n", args...)
}

// LogWarning mencatat pesan peringatan ke terminal.
// Gunakan ini untuk kondisi yang tidak fatal tetapi mungkin memerlukan perhatian.
func LogWarning(message string, args ...interface{}) {
	Logger.Printf("WARNING: "+message+"\n", args...)
}

// LogError mencatat pesan kesalahan ke terminal.
// Gunakan ini untuk kesalahan yang mengganggu fungsionalitas, tetapi tidak menyebabkan crash aplikasi.
func LogError(err error, message string, args ...interface{}) {
	if err != nil {
		Logger.Printf("ERROR: "+message+" (Error: %v)\n", append(args, err)...)
	} else {
		Logger.Printf("ERROR: "+message+"\n", args...)
	}
}

// LogFatal mencatat pesan kesalahan fatal dan kemudian menghentikan aplikasi.
// Gunakan ini untuk kesalahan yang tidak dapat dipulihkan yang mengharuskan aplikasi berhenti.
func LogFatal(err error, message string, args ...interface{}) {
	Logger.Fatalf("FATAL: "+message+" (Error: %v)\n", append(args, err)...)
}

/*
Cara Penggunaan:

// Untuk mencatat pesan informasi:
// utils.LogInfo("Server dimulai pada port %d", 8080)

// Untuk mencatat pesan peringatan:
// utils.LogWarning("Autentikasi gagal untuk pengguna %s", "john_doe")

// Untuk mencatat pesan kesalahan:
// err := errors.New("gagal mengambil data")
// utils.LogError(err, "Terjadi kesalahan saat memproses request")

// Untuk mencatat kesalahan fatal (ini akan menghentikan aplikasi):
// err := errors.New("koneksi database gagal")
// utils.LogFatal(err, "Aplikasi tidak dapat berjalan tanpa koneksi database")
*/

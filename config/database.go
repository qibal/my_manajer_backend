package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"backend_my_manajer/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// globalMongoClient adalah instance tunggal (singleton) dari koneksi MongoDB.
var globalMongoClient *mongo.Client

// DatabaseConfig menyimpan nama database dan daftar koleksi yang digunakan.
type DatabaseConfig struct {
	DatabaseName string
	Collections  map[string]string
}

// DBConfig adalah konfigurasi database global.
var DBConfig = DatabaseConfig{
	DatabaseName: "my_manager_db", // Nama database Anda. Sesuaikan jika perlu.
	Collections: map[string]string{
		"Businesses":        "businesses",         // Nama koleksi untuk entitas Business
		"Users":             "users",              // Contoh untuk koleksi Users
		"Channels":          "channels",           // Contoh untuk koleksi Channels
		"ChannelCategories": "channel_categories", // Koleksi baru untuk kategori channel
		"Messages":          "messages",           // Koleksi baru untuk pesan
		"Databases":         "databases",          // Menambahkan koleksi Database di sini
		"Roles":             "roles",              // Tambahkan koleksi Roles di sini
		"ActivityLogs":      "activity_logs",      // Koleksi untuk log aktivitas
		// Tambahkan koleksi lain di sini sesuai kebutuhan Anda
	},
}

// ConnectDB membuat dan mengembalikan koneksi ke database MongoDB.
// Fungsi ini akan memastikan hanya ada satu koneksi (singleton).
func ConnectDB() *mongo.Client {
	// Jika koneksi sudah ada, kembalikan yang sudah ada (singleton).
	if globalMongoClient != nil {
		return globalMongoClient
	}

	// Ambil URI MongoDB dari variabel lingkungan. Jika tidak ada, gunakan default.
	// PENTING: Untuk membaca dari file .env, Anda perlu menginstal github.com/joho/godotenv
	// dan memuatnya sebelum memanggil fungsi ini (misal di main.go).
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		utils.LogWarning("Variabel lingkungan MONGO_URI tidak ditemukan. Menggunakan default: %s", mongoURI)
	}

	// Menyiapkan opsi koneksi MongoDB.
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Membuat konteks dengan timeout untuk proses koneksi.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Pastikan konteks dibatalkan setelah selesai.

	// Mencoba untuk terhubung ke MongoDB.
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		// Jika gagal terhubung, log error fatal dan hentikan aplikasi.
		utils.LogFatal(err, "Gagal terhubung ke MongoDB")
	}

	// Memeriksa koneksi ke MongoDB dengan melakukan ping.
	err = client.Ping(ctx, nil)
	if err != nil {
		// Jika ping gagal, log error fatal dan hentikan aplikasi.
		utils.LogFatal(err, "Gagal melakukan ping ke MongoDB")
	}

	globalMongoClient = client // Simpan instance client sebagai singleton
	utils.LogInfo("Koneksi ke MongoDB berhasil terjalin!")
	return globalMongoClient
}

// GetCollection membantu mendapatkan instance koleksi dari database.
// Gunakan DBConfig.DatabaseName dan DBConfig.Collections untuk mendapatkan nama yang benar.
func GetCollection(client *mongo.Client, collectionKey string) *mongo.Collection {
	collectionName, ok := DBConfig.Collections[collectionKey]
	if !ok {
		utils.LogFatal(fmt.Errorf("koleksi dengan kunci %s tidak ditemukan di konfigurasi database", collectionKey), "Invalid collection key")
	}
	return client.Database(DBConfig.DatabaseName).Collection(collectionName)
}

/*
Cara Penggunaan (dalam main.go):

import (
	"backend_my_manajer/config"
	"backend_my_manajer/repository"
	"context"
	"log"
	"time"
	"github.com/joho/godotenv" // Untuk membaca .env
)

func main() {
	// Memuat file .env (jika ada)
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Inisialisasi koneksi database
	dbClient := config.ConnectDB()
	defer func() {
		if err := dbClient.Disconnect(nil); err != nil { // Mengubah context.TODO() menjadi nil
			log.Fatal(err)
		}
	}()

	// Sekarang dbClient bisa diteruskan ke repository atau service
	// Contoh mendapatkan koleksi bisnis:
	// businessCollection := config.GetCollection(dbClient, "Businesses")
	// businessRepo := repository.NewBusinessRepository(dbClient, businessCollection)

	// ... kode aplikasi lainnya ...
}

Cara Penggunaan (dalam repository):

// import (
// 	"backend_my_manajer/config"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// func NewBusinessRepository(dbClient *mongo.Client) BusinessRepository {
// 	collection := config.GetCollection(dbClient, "Businesses")
// 	return &businessRepositoryImpl{
// 		collection: collection,
// 	}
// }
*/

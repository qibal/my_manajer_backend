package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User merepresentasikan struktur dokumen pengguna di database.
type User struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"` // Mengubah tipe ID
	BusinessIDs  []string            `json:"businessIds" bson:"businessIds"`
	Username     string              `json:"username" bson:"username"`
	Email        string              `json:"email" bson:"email"`
	PasswordHash string              `json:"passwordHash" bson:"passwordHash"` // Menambahkan field ini
	Avatar       string              `json:"avatar" bson:"avatar"`
	Status       string              `json:"status" bson:"status"`
	IsActive     bool                `json:"isActive" bson:"isActive"` // Menambahkan field ini
	Roles        map[string][]string `json:"roles" bson:"roles"`       // Map businessId to array of role IDs
	CreatedAt    time.Time           `json:"createdAt" bson:"createdAt"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi User
// user := model.User{
//     ID:          "user_123",
//     BusinessIDs: []string{"business_abc"},
//     Username:    "john_doe",
//     Email:       "john@example.com",
//     Avatar:      "https://example.com/avatar/john.png",
//     Status:      "online",
//     Roles:       map[string][]string{"business_abc": {"role_member"}},
//     CreatedAt:   time.Now(),
// }
*/

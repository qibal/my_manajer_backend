package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BusinessSettings merepresentasikan pengaturan bisnis.
type BusinessSettings struct {
	Theme         string `json:"theme" bson:"theme"`
	Notifications string `json:"notifications" bson:"notifications"`
}

// Business merepresentasikan struktur dokumen bisnis di database.
type Business struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `json:"name" bson:"name"`
	OwnerID   string             `json:"ownerId" bson:"ownerId"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	Settings  BusinessSettings   `json:"settings" bson:"settings"`
	Avatar    string             `json:"avatar" bson:"avatar"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Business
// business := model.Business{
//     ID:        "business_123",
//     Name:      "My Company",
//     OwnerID:   "user_abc",
//     CreatedAt: time.Now(),
//     UpdatedAt: time.Now(),
//     Settings:  model.BusinessSettings{
//         Theme:        "dark",
//         Notifications: "all",
//     },
//     Avatar: "/path/to/avatar.png",
// }
*/

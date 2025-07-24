package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DatabaseColumn merepresentasikan kolom dalam sebuah database.
type DatabaseColumn struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `json:"name"`
	Type    string             `json:"type"`              // e.g., "date", "text", "select", "boolean"
	Options []string           `json:"options,omitempty"` // Only for type "select"
}

// DatabaseRowValue merepresentasikan nilai-nilai dalam satu baris data.
// Menggunakan map[string]interface{} karena tipe data bisa bervariasi.
type DatabaseRowValue map[string]interface{}

// DatabaseRow merepresentasikan satu baris data dalam database.
type DatabaseRow struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Values DatabaseRowValue   `json:"values"`
}

// DatabaseData merepresentasikan struktur data internal dari database.
type DatabaseData struct {
	Columns []DatabaseColumn `json:"columns"`
	Rows    []DatabaseRow    `json:"rows"`
}

// Database merepresentasikan struktur dokumen database di database.
type Database struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChannelID    primitive.ObjectID `bson:"channelId,omitempty" json:"channelId"`
	AuthorID     primitive.ObjectID `bson:"authorId,omitempty" json:"authorId"`
	Title        string             `json:"title"`
	DatabaseData DatabaseData       `json:"databaseData"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Database
// db := model.Database{
//     ID:        "db_123",
//     ChannelID: "channel_abc",
//     AuthorID:  "user_xyz",
//     Title:     "Daftar Pengguna",
//     DatabaseData: model.DatabaseData{
//         Columns: []model.DatabaseColumn{
//             {ID: "col_1", Name: "Nama", Type: "text"},
//             {ID: "col_2", Name: "Umur", Type: "number"},
//         },
//         Rows: []model.DatabaseRow{
//             {ID: "row_1", Values: map[string]interface{}{"col_1": "Alice", "col_2": 30}},
//             {ID: "row_2", Values: map[string]interface{}{"col_1": "Bob", "col_2": 25}},
//         },
//     },
//     CreatedAt: time.Now(),
//     UpdatedAt: time.Now(),
// }
*/

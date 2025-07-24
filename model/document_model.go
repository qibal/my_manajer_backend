package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DocumentContentNode merepresentasikan node konten dalam dokumen.
// Ini adalah struktur rekursif yang bisa berisi heading, paragraph, bulletList, dll.
// Menggunakan interface{} untuk 'content' karena bisa berupa array dari node lain atau text.
type DocumentContentNode struct {
	Type    string                   `json:"type"`
	Attrs   map[string]interface{}   `json:"attrs,omitempty"`   // Atribut opsional (misal: level untuk heading)
	Content []DocumentContentSubNode `json:"content,omitempty"` // Bisa berisi array of sub-nodes atau teks
}

// DocumentContentSubNode merepresentasikan sub-node dari DocumentContentNode.
// Digunakan karena 'content' di DocumentContentNode bisa berisi kombinasi object atau text.
type DocumentContentSubNode struct {
	Type    string                   `json:"type"`
	Text    string                   `json:"text,omitempty"`
	Content []DocumentContentSubNode `json:"content,omitempty"` // Rekursif untuk nested lists
}

// Document merepresentasikan struktur dokumen di database.
type Document struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ChannelID primitive.ObjectID  `bson:"channelId,omitempty" json:"channelId"`
	AuthorID  primitive.ObjectID  `bson:"authorId,omitempty" json:"authorId"`
	Title     string              `json:"title"`
	Content   DocumentContentNode `json:"content"` // Root node of the document content
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Document
// doc := model.Document{
//     ID:        "doc_123",
//     ChannelID: "channel_abc",
//     AuthorID:  "user_xyz",
//     Title:     "Laporan Proyek",
//     Content: model.DocumentContentNode{
//         Type: "doc",
//         Content: []model.DocumentContentSubNode{
//             {
//                 Type: "heading",
//                 Attrs: map[string]interface{}{"level": 1},
//                 Content: []model.DocumentContentSubNode{{Type: "text", Text: "Judul Laporan"}},
//             },
//             {
//                 Type: "paragraph",
//                 Content: []model.DocumentContentSubNode{{Type: "text", Text: "Ini adalah isi paragraf."}},
//             },
//         },
//     },
//     CreatedAt: time.Now(),
//     UpdatedAt: time.Now(),
// }
*/

package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DrawingPoint merepresentasikan koordinat titik dalam gambar.
type DrawingPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// DrawingShape merepresentasikan bentuk dalam gambar.
// Menggunakan interface{} untuk 'points' karena bisa array of points atau tidak ada (tergantung type).
type DrawingShape struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type   string             `json:"type"` // e.g., "rectangle", "arrow"
	X      int                `json:"x,omitempty"`
	Y      int                `json:"y,omitempty"`
	Width  int                `json:"width,omitempty"`
	Height int                `json:"height,omitempty"`
	Fill   string             `json:"fill,omitempty"`
	Stroke string             `json:"stroke,omitempty"`
	Text   string             `json:"text,omitempty"`
	Points []DrawingPoint     `json:"points,omitempty"` // Untuk type "arrow"
}

// DrawingData merepresentasikan data internal dari gambar.
type DrawingData struct {
	Version string         `json:"version"`
	Shapes  []DrawingShape `json:"shapes"`
}

// Drawing merepresentasikan struktur dokumen gambar di database.
type Drawing struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChannelID   primitive.ObjectID `bson:"channelId,omitempty" json:"channelId"`
	AuthorID    primitive.ObjectID `bson:"authorId,omitempty" json:"authorId"`
	Title       string             `json:"title"`
	DrawingData DrawingData        `json:"drawingData"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

/*
Cara Penggunaan:

// Contoh inisialisasi Drawing
// drawing := model.Drawing{
//     ID:        "drawing_123",
//     ChannelID: "channel_abc",
//     AuthorID:  "user_xyz",
//     Title:     "Flowchart Login",
//     DrawingData: model.DrawingData{
//         Version: "1.0",
//         Shapes: []model.DrawingShape{
//             {
//                 ID:   "shape_1",
//                 Type: "rectangle",
//                 X:    50, Y: 50, Width: 100, Height: 50,
//                 Fill:  "#FF0000",
//                 Text: "Start",
//             },
//         },
//     },
//     CreatedAt: time.Now(),
//     UpdatedAt: time.Now(),
// }
*/

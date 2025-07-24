package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Channel merepresentasikan struktur dokumen channel di database.
type Channel struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	BusinessID primitive.ObjectID  `bson:"businessId,omitempty" json:"businessId"`
	Name       string              `json:"name" bson:"name"`
	Type       string              `json:"type" bson:"type"`                       // e.g., "messages", "documents", "drawings", "databases", "reports"
	CategoryID *primitive.ObjectID `bson:"categoryId,omitempty" json:"categoryId"` // Changed from Category to CategoryID, now a pointer
	Order      int                 `json:"order" bson:"order"`
	CreatedAt  time.Time           `json:"createdAt" bson:"createdAt"`
	// Unread removed
}

/*
// Contoh inisialisasi Channel
// channel := model.Channel{
//     ID:         "channel_123",
//     BusinessID: "business_abc",
//     Name:       "General Chat",
//     Type:       "messages",
//     Category:   "Text Channels",
//     Order:      1,
//     CreatedAt:  time.Now(),
//     Unread:     5,
// }
*/

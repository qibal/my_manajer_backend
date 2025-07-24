package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageMediaMetadata merepresentasikan metadata media yang dilampirkan ke pesan.
type MessageMediaMetadata struct {
	Filename string `json:"filename" bson:"filename,omitempty"`
	Size     int64  `json:"size" bson:"size,omitempty"`
	Width    int    `json:"width" bson:"width,omitempty"`
	Height   int    `json:"height" bson:"height,omitempty"`
}

// MessageReaction merepresentasikan reaksi terhadap pesan.
type MessageReaction struct {
	Emoji   string               `json:"emoji" bson:"emoji"`
	UserIDs []primitive.ObjectID `json:"userIds" bson:"userIds"`
}

// Message merepresentasikan struktur dokumen pesan di database.
type Message struct {
	ID            primitive.ObjectID    `bson:"_id,omitempty" json:"id"`
	ChannelID     primitive.ObjectID    `bson:"channelId" json:"channelId"`
	UserID        primitive.ObjectID    `bson:"userId" json:"userId"`
	Content       string                `json:"content" bson:"content"`
	MessageType   string                `json:"messageType" bson:"messageType"` // e.g., "text", "image", "file"
	MediaPath     string                `json:"mediaPath" bson:"mediaPath,omitempty"`
	MediaMetadata *MessageMediaMetadata `json:"mediaMetadata" bson:"mediaMetadata,omitempty"`
	CreatedAt     time.Time             `json:"createdAt" bson:"createdAt"`
	UpdatedAt     *time.Time            `json:"updatedAt" bson:"updatedAt,omitempty"`
	IsPinned      bool                  `json:"isPinned" bson:"isPinned"`
	Reactions     []MessageReaction     `json:"reactions" bson:"reactions"`
}

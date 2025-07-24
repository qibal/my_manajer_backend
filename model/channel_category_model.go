package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChannelCategory merepresentasikan struktur dokumen kategori channel di database.
type ChannelCategory struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	BusinessID primitive.ObjectID `bson:"businessId,omitempty" json:"businessId"`
	Name       string             `json:"name" bson:"name" validate:"required,min=3,max=100"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
}

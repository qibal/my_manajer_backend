package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityLog merepresentasikan satu entri log aktivitas dalam database.
type ActivityLog struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID     string             `bson:"user_id" json:"user_id"`
	Action     string             `bson:"action" json:"action"`
	Method     string             `bson:"method" json:"method"`
	Endpoint   string             `bson:"endpoint" json:"endpoint"`
	StatusCode int                `bson:"status_code" json:"status_code"`
	IPAddress  string             `bson:"ip_address" json:"ip_address"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}

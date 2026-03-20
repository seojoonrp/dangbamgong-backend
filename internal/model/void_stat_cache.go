package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidStatCache struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetDay        string             `bson:"target_day" json:"targetDay"`
	Bucket           string             `bson:"bucket" json:"bucket"`
	Count            int                `bson:"count" json:"count"`
	TotalDurationSec int64              `bson:"total_duration_sec,omitempty" json:"totalDurationSec,omitempty"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updatedAt"`
}

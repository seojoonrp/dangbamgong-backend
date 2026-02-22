package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidSession struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	StartedAt   time.Time          `bson:"started_at" json:"started_at"`
	EndedAt     time.Time          `bson:"ended_at" json:"ended_at"`
	DurationSec int64              `bson:"duration_sec" json:"duration_sec"`
	TargetDay   string             `bson:"target_day" json:"target_day"`
	Activities  []string           `bson:"activities" json:"activities"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

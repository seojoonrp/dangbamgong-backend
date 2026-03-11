package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidSession struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
	StartedAt   time.Time          `bson:"started_at" json:"startedAt"`
	EndedAt     time.Time          `bson:"ended_at" json:"endedAt"`
	DurationSec int64              `bson:"duration_sec" json:"durationSec"`
	TargetDay   string             `bson:"target_day" json:"targetDay"`
	Activities  []string           `bson:"activities" json:"activities"`
	CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
}

type VoidUserStats struct {
	TotalDurationSec int64 `bson:"total_duration_sec"`
	SessionCount     int   `bson:"session_count"`
	MaxDurationSec   int64 `bson:"max_duration_sec"`
}

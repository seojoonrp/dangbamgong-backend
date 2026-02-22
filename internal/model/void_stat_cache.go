package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VoidStatCache struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetDay string             `bson:"target_day" json:"target_day"`
	Bucket    string             `bson:"bucket" json:"bucket"`
	Count     int                `bson:"count" json:"count"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

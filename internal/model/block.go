package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Block struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"userId"`
	BlockedID primitive.ObjectID `bson:"blocked_id" json:"blockedId"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}

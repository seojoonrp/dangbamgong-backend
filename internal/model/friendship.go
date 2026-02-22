package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Friendship struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	FriendID  primitive.ObjectID `bson:"friend_id" json:"friend_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

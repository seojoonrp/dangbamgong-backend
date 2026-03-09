package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendRequestStatus string

const (
	FriendRequestPending  FriendRequestStatus = "PENDING"
	FriendRequestAccepted FriendRequestStatus = "ACCEPTED"
	FriendRequestRejected FriendRequestStatus = "REJECTED"
)

type FriendRequest struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	SenderID   primitive.ObjectID  `bson:"sender_id" json:"senderId"`
	ReceiverID primitive.ObjectID  `bson:"receiver_id" json:"receiverId"`
	Status     FriendRequestStatus `bson:"status" json:"status"`
	CreatedAt  time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt  time.Time           `bson:"updated_at" json:"updatedAt"`
}

package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotifVoidReminder  NotificationType = "VOID_REMINDER"
	NotifFriendRequest NotificationType = "FRIEND_REQUEST"
	NotifFriendAccept  NotificationType = "FRIEND_ACCEPT"
	NotifFriendNudge   NotificationType = "FRIEND_NUDGE"
)

type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Type      NotificationType   `bson:"type"`
	Title     string             `bson:"title"`
	Body      string             `bson:"body"`
	Data      map[string]string  `bson:"data,omitempty"`
	IsRead    bool               `bson:"is_read"`
	CreatedAt time.Time          `bson:"created_at"`
}

type DeviceToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Token     string             `bson:"token"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

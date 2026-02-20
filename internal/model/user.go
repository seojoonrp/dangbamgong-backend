package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SocialProvider string

const (
	ProviderGoogle SocialProvider = "GOOGLE"
	ProviderKakao  SocialProvider = "KAKAO"
	ProviderApple  SocialProvider = "APPLE"
	ProviderTest   SocialProvider = "TEST"
)

type NotificationSettings struct {
	VoidReminder  bool `bson:"void_reminder"   json:"void_reminder"`
	ReminderHours int  `bson:"reminder_hours"  json:"reminder_hours"`
	FriendNudge   bool `bson:"friend_nudge"    json:"friend_nudge"`
}

type User struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	SocialProvider       SocialProvider       `bson:"social_provider" json:"social_provider"`
	SocialID             string               `bson:"social_id" json:"social_id"`
	Nickname             string               `bson:"nickname" json:"nickname,omitempty"`
	Tag                  string               `bson:"tag" json:"tag"`
	IsInVoid             bool                 `bson:"is_in_void" json:"is_in_void"`
	CurrentVoidStartedAt *time.Time           `bson:"current_void_started_at,omitempty" json:"current_void_started_at"`
	NotificationSettings NotificationSettings `bson:"notification_settings" json:"notification_settings"`
	AppleRefreshToken    string               `bson:"apple_refresh_token,omitempty" json:"-"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
}

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
	VoidReminder  bool `bson:"void_reminder"   json:"voidReminder"`
	ReminderHours int  `bson:"reminder_hours"  json:"reminderHours"`
	FriendNudge   bool `bson:"friend_nudge"    json:"friendNudge"`
}

type User struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	SocialProvider       SocialProvider       `bson:"social_provider" json:"socialProvider"`
	SocialID             string               `bson:"social_id" json:"socialId"`
	Nickname             string               `bson:"nickname" json:"nickname,omitempty"`
	Tag                  string               `bson:"tag" json:"tag"`
	IsInVoid             bool                 `bson:"is_in_void" json:"isInVoid"`
	CurrentVoidStartedAt *time.Time           `bson:"current_void_started_at,omitempty" json:"currentVoidStartedAt"`
	LastVoidEndedAt      *time.Time           `bson:"last_void_ended_at,omitempty" json:"lastVoidEndedAt"`
	NotificationSettings NotificationSettings `bson:"notification_settings" json:"notificationSettings"`
	AppleRefreshToken    string               `bson:"apple_refresh_token,omitempty" json:"-"`
	CreatedAt            time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updatedAt"`
}

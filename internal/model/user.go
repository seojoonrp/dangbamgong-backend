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

type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"                  json:"id"`
	SocialProvider    SocialProvider     `bson:"social_provider"                json:"social_provider"`
	SocialID          string             `bson:"social_id"                      json:"social_id"`
	Nickname          string             `bson:"nickname"                       json:"nickname,omitempty"`
	Tag               string             `bson:"tag"                            json:"tag"`
	AppleRefreshToken string             `bson:"apple_refresh_token,omitempty"  json:"-"`
	CreatedAt         time.Time          `bson:"created_at"                     json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at"                     json:"updated_at"`
}

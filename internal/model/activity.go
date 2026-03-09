// internal/model/activity.go
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Activity struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"userId"`
	Name       string             `bson:"name" json:"name"`
	UsageCount int                `bson:"usage_count" json:"usageCount"`
	LastUsedAt *time.Time         `bson:"last_used_at,omitempty" json:"lastUsedAt"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}

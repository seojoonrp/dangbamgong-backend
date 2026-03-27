package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := map[string][]mongo.IndexModel{
		"users": {
			{
				Keys:    bson.D{{Key: "social_provider", Value: 1}, {Key: "social_id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{Keys: bson.D{{Key: "tag", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "is_in_void", Value: 1}}},
		},
		"activities": {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "last_used_at", Value: -1}}},
		},
		"blocks": {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "blocked_id", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "blocked_id", Value: 1}}},
		},
		"device_tokens": {
			{Keys: bson.D{{Key: "token", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "user_id", Value: 1}}},
		},
		"friendships": {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "friend_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"friend_requests": {
			{Keys: bson.D{{Key: "sender_id", Value: 1}, {Key: "receiver_id", Value: 1}, {Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "receiver_id", Value: 1}, {Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
		},
		"notifications": {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_read", Value: 1}}},
		},
		"void_sessions": {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "target_day", Value: 1}}},
			{Keys: bson.D{{Key: "target_day", Value: 1}}},
		},
		"void_stats_cache": {
			{Keys: bson.D{{Key: "target_day", Value: 1}, {Key: "bucket", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
	}

	for coll, models := range indexes {
		_, err := db.Collection(coll).Indexes().CreateMany(ctx, models)
		if err != nil {
			log.Printf("Failed to create indexes for %s: %v", coll, err)
		}
	}

	log.Println("MongoDB indexes ensured")
}

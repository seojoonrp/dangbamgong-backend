package repository

import (
	"context"
	"time"

	"dangbamgong-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, userID primitive.ObjectID, token string) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.DeviceToken, error)
	DeleteByUserAndToken(ctx context.Context, userID primitive.ObjectID, token string) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type deviceTokenRepository struct {
	coll *mongo.Collection
}

func NewDeviceTokenRepository(db *mongo.Database) DeviceTokenRepository {
	return &deviceTokenRepository{coll: db.Collection("device_tokens")}
}

func (r *deviceTokenRepository) Upsert(ctx context.Context, userID primitive.ObjectID, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"token": token},
		bson.M{
			"$set": bson.M{
				"user_id":    userID,
				"updated_at": now,
			},
			"$setOnInsert": bson.M{
				"created_at": now,
			},
		},
		opts,
	)
	return err
}

func (r *deviceTokenRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []model.DeviceToken
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *deviceTokenRepository) DeleteByUserAndToken(ctx context.Context, userID primitive.ObjectID, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"user_id": userID, "token": token})
	return err
}

func (r *deviceTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

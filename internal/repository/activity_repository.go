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

type ActivityRepository interface {
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Activity, error)
	FindByUserIDAndName(ctx context.Context, userID primitive.ObjectID, name string) (*model.Activity, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Activity, error)
	Create(ctx context.Context, activity *model.Activity) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
	IncrementUsage(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error
}

type activityRepository struct {
	coll *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) ActivityRepository {
	return &activityRepository{coll: db.Collection("activities")}
}

func (r *activityRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Activity, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// last_used_at desc, usage_count desc
	opts := options.Find().SetSort(bson.D{
		{Key: "last_used_at", Value: -1},
		{Key: "usage_count", Value: -1},
	})

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []model.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *activityRepository) FindByUserIDAndName(ctx context.Context, userID primitive.ObjectID, name string) (*model.Activity, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var activity model.Activity
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "name": name}).Decode(&activity)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &activity, err
}

func (r *activityRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Activity, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var activity model.Activity
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&activity)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &activity, err
}

func (r *activityRepository) Create(ctx context.Context, activity *model.Activity) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, activity)
	if err != nil {
		return err
	}
	activity.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *activityRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *activityRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

func (r *activityRepository) IncrementUsage(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$inc": bson.M{"usage_count": 1},
		"$set": bson.M{"last_used_at": usedAt},
	})
	return err
}

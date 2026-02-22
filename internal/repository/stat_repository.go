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

type UserDuration struct {
	UserID      primitive.ObjectID `bson:"_id"`
	TotalDurSec int64              `bson:"total_dur_sec"`
}

type StatRepository interface {
	CountCurrentVoid(ctx context.Context) (int, error)
	CountDistinctUsersForDay(ctx context.Context, targetDay string) (int, error)
	GetBucketCache(ctx context.Context, targetDay string) ([]model.VoidStatCache, error)
	UpsertBucketCache(ctx context.Context, caches []model.VoidStatCache) error
	GetUserDurations(ctx context.Context, targetDay string) ([]UserDuration, error)
}

type statRepository struct {
	usersColl    *mongo.Collection
	sessionsColl *mongo.Collection
	cacheColl    *mongo.Collection
}

func NewStatRepository(db *mongo.Database) StatRepository {
	return &statRepository{
		usersColl:    db.Collection("users"),
		sessionsColl: db.Collection("void_sessions"),
		cacheColl:    db.Collection("void_stats_cache"),
	}
}

func (r *statRepository) CountCurrentVoid(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := r.usersColl.CountDocuments(ctx, bson.M{"is_in_void": true})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *statRepository) CountDistinctUsersForDay(ctx context.Context, targetDay string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.sessionsColl.Distinct(ctx, "user_id", bson.M{"target_day": targetDay})
	if err != nil {
		return 0, err
	}
	return len(result), nil
}

func (r *statRepository) GetBucketCache(ctx context.Context, targetDay string) ([]model.VoidStatCache, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "bucket", Value: 1}})
	cursor, err := r.cacheColl.Find(ctx, bson.M{"target_day": targetDay}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var caches []model.VoidStatCache
	if err := cursor.All(ctx, &caches); err != nil {
		return nil, err
	}
	return caches, nil
}

func (r *statRepository) UpsertBucketCache(ctx context.Context, caches []model.VoidStatCache) error {
	if len(caches) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	models := make([]mongo.WriteModel, len(caches))
	for i, c := range caches {
		models[i] = mongo.NewUpdateOneModel().
			SetFilter(bson.M{"target_day": c.TargetDay, "bucket": c.Bucket}).
			SetUpdate(bson.M{"$set": bson.M{
				"count":      c.Count,
				"updated_at": c.UpdatedAt,
			}}).
			SetUpsert(true)
	}

	_, err := r.cacheColl.BulkWrite(ctx, models)
	return err
}

func (r *statRepository) GetUserDurations(ctx context.Context, targetDay string) ([]UserDuration, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"target_day": targetDay}}},
		{{Key: "$group", Value: bson.M{
			"_id":           "$user_id",
			"total_dur_sec": bson.M{"$sum": "$duration_sec"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total_dur_sec", Value: -1}}}},
	}

	cursor, err := r.sessionsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []UserDuration
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

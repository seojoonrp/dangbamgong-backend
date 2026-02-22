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

type UserRepository interface {
	FindBySocial(ctx context.Context, provider model.SocialProvider, socialID string) (*model.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	UpdateNickname(ctx context.Context, id primitive.ObjectID, nickname string) error
	UpdateSettings(ctx context.Context, id primitive.ObjectID, settings model.NotificationSettings) error
	SetVoidState(ctx context.Context, id primitive.ObjectID, isInVoid bool, startedAt *time.Time) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	SearchByTagPrefix(ctx context.Context, prefix string, excludeIDs []primitive.ObjectID, limit int) ([]model.User, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.User, error)
}

type userRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{coll: db.Collection("users")}
}

func (r *userRepository) FindBySocial(ctx context.Context, provider model.SocialProvider, socialID string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.coll.FindOne(ctx, bson.M{
		"social_provider": provider,
		"social_id":       socialID,
	}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *userRepository) UpdateNickname(ctx context.Context, id primitive.ObjectID, nickname string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"nickname": nickname, "updated_at": time.Now()},
	})
	return err
}

func (r *userRepository) UpdateSettings(ctx context.Context, id primitive.ObjectID, settings model.NotificationSettings) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"notification_settings": settings, "updated_at": time.Now()},
	})
	return err
}

func (r *userRepository) SetVoidState(ctx context.Context, id primitive.ObjectID, isInVoid bool, startedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{
			"is_in_void":               isInVoid,
			"current_void_started_at":  startedAt,
			"updated_at":               time.Now(),
		},
	})
	return err
}

func (r *userRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *userRepository) SearchByTagPrefix(ctx context.Context, prefix string, excludeIDs []primitive.ObjectID, limit int) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"tag": bson.M{"$regex": "^" + prefix, "$options": "i"},
	}
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

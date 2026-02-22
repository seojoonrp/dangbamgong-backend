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

type FriendshipRepository interface {
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Friendship, error)
	FindOne(ctx context.Context, userID, friendID primitive.ObjectID) (*model.Friendship, error)
	Create(ctx context.Context, friendship *model.Friendship) error
	DeleteByUserPair(ctx context.Context, userA, userB primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type friendshipRepository struct {
	coll *mongo.Collection
}

func NewFriendshipRepository(db *mongo.Database) FriendshipRepository {
	return &friendshipRepository{coll: db.Collection("friendships")}
}

func (r *friendshipRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Friendship, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friendships []model.Friendship
	if err := cursor.All(ctx, &friendships); err != nil {
		return nil, err
	}
	return friendships, nil
}

func (r *friendshipRepository) FindOne(ctx context.Context, userID, friendID primitive.ObjectID) (*model.Friendship, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var friendship model.Friendship
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "friend_id": friendID}).Decode(&friendship)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &friendship, err
}

func (r *friendshipRepository) Create(ctx context.Context, friendship *model.Friendship) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, friendship)
	if err != nil {
		return err
	}
	friendship.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *friendshipRepository) DeleteByUserPair(ctx context.Context, userA, userB primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"$or": bson.A{
		bson.M{"user_id": userA, "friend_id": userB},
		bson.M{"user_id": userB, "friend_id": userA},
	}})
	return err
}

func (r *friendshipRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"$or": bson.A{
		bson.M{"user_id": userID},
		bson.M{"friend_id": userID},
	}})
	return err
}

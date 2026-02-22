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

type FriendRequestRepository interface {
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.FriendRequest, error)
	FindPending(ctx context.Context, senderID, receiverID primitive.ObjectID) (*model.FriendRequest, error)
	FindByReceiverID(ctx context.Context, receiverID primitive.ObjectID, status model.FriendRequestStatus) ([]model.FriendRequest, error)
	FindBySenderID(ctx context.Context, senderID primitive.ObjectID) ([]model.FriendRequest, error)
	Create(ctx context.Context, req *model.FriendRequest) error
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.FriendRequestStatus) error
	DeleteByUserPair(ctx context.Context, userA, userB primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type friendRequestRepository struct {
	coll *mongo.Collection
}

func NewFriendRequestRepository(db *mongo.Database) FriendRequestRepository {
	return &friendRequestRepository{coll: db.Collection("friend_requests")}
}

func (r *friendRequestRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var req model.FriendRequest
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&req)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &req, err
}

func (r *friendRequestRepository) FindPending(ctx context.Context, senderID, receiverID primitive.ObjectID) (*model.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var req model.FriendRequest
	err := r.coll.FindOne(ctx, bson.M{
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"status":      model.FriendRequestPending,
	}).Decode(&req)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &req, err
}

func (r *friendRequestRepository) FindByReceiverID(ctx context.Context, receiverID primitive.ObjectID, status model.FriendRequestStatus) ([]model.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{
		"receiver_id": receiverID,
		"status":      status,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.FriendRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *friendRequestRepository) FindBySenderID(ctx context.Context, senderID primitive.ObjectID) ([]model.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"sender_id": senderID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []model.FriendRequest
	if err := cursor.All(ctx, &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func (r *friendRequestRepository) Create(ctx context.Context, req *model.FriendRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, req)
	if err != nil {
		return err
	}
	req.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *friendRequestRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.FriendRequestStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.UpdateByID(ctx, id, bson.M{
		"$set": bson.M{"status": status, "updated_at": time.Now()},
	})
	return err
}

func (r *friendRequestRepository) DeleteByUserPair(ctx context.Context, userA, userB primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"$or": bson.A{
		bson.M{"sender_id": userA, "receiver_id": userB},
		bson.M{"sender_id": userB, "receiver_id": userA},
	}})
	return err
}

func (r *friendRequestRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"$or": bson.A{
		bson.M{"sender_id": userID},
		bson.M{"receiver_id": userID},
	}})
	return err
}

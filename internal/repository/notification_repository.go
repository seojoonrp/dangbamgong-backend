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

type NotificationRepository interface {
	Create(ctx context.Context, notif *model.Notification) error
	FindByUserID(ctx context.Context, userID primitive.ObjectID, limit int, offset int) ([]model.Notification, error)
	MarkAsRead(ctx context.Context, notifID primitive.ObjectID, userID primitive.ObjectID) (int64, error)
	CountUnread(ctx context.Context, userID primitive.ObjectID) (int, error)
}

type notificationRepository struct {
	coll *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) NotificationRepository {
	return &notificationRepository{coll: db.Collection("notifications")}
}

func (r *notificationRepository) Create(ctx context.Context, notif *model.Notification) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, notif)
	if err != nil {
		return err
	}
	notif.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID, limit int, offset int) ([]model.Notification, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []model.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notifID primitive.ObjectID, userID primitive.ObjectID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": notifID, "user_id": userID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID primitive.ObjectID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := r.coll.CountDocuments(ctx, bson.M{"user_id": userID, "is_read": false})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

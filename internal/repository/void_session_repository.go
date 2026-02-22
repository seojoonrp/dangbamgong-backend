package repository

import (
	"context"
	"time"

	"dangbamgong-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VoidSessionRepository interface {
	Create(ctx context.Context, session *model.VoidSession) error
	FindByUserIDAndTargetDay(ctx context.Context, userID primitive.ObjectID, targetDay string) ([]model.VoidSession, error)
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type voidSessionRepository struct {
	coll *mongo.Collection
}

func NewVoidSessionRepository(db *mongo.Database) VoidSessionRepository {
	return &voidSessionRepository{coll: db.Collection("void_sessions")}
}

func (r *voidSessionRepository) Create(ctx context.Context, session *model.VoidSession) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, session)
	if err != nil {
		return err
	}
	session.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *voidSessionRepository) FindByUserIDAndTargetDay(ctx context.Context, userID primitive.ObjectID, targetDay string) ([]model.VoidSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID, "target_day": targetDay})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []model.VoidSession
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *voidSessionRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

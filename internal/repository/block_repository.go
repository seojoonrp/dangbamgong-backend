package repository

import (
	"context"
	"time"

	"dangbamgong-backend/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BlockRepository interface {
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Block, error)
	FindOne(ctx context.Context, userID, blockedID primitive.ObjectID) (*model.Block, error)
	Create(ctx context.Context, block *model.Block) error
	Delete(ctx context.Context, userID, blockedID primitive.ObjectID) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
	FindByBlockedID(ctx context.Context, blockedID primitive.ObjectID) ([]model.Block, error)
}

type blockRepository struct {
	coll *mongo.Collection
}

func NewBlockRepository(db *mongo.Database) BlockRepository {
	return &blockRepository{coll: db.Collection("blocks")}
}

func (r *blockRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]model.Block, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blocks []model.Block
	if err := cursor.All(ctx, &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

func (r *blockRepository) FindOne(ctx context.Context, userID, blockedID primitive.ObjectID) (*model.Block, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var block model.Block
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "blocked_id": blockedID}).Decode(&block)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &block, err
}

func (r *blockRepository) Create(ctx context.Context, block *model.Block) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.coll.InsertOne(ctx, block)
	if err != nil {
		return err
	}
	block.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *blockRepository) Delete(ctx context.Context, userID, blockedID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteOne(ctx, bson.M{"user_id": userID, "blocked_id": blockedID})
	return err
}

func (r *blockRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.coll.DeleteMany(ctx, bson.M{"$or": bson.A{
		bson.M{"user_id": userID},
		bson.M{"blocked_id": userID},
	}})
	return err
}

func (r *blockRepository) FindByBlockedID(ctx context.Context, blockedID primitive.ObjectID) ([]model.Block, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.coll.Find(ctx, bson.M{"blocked_id": blockedID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blocks []model.Block
	if err := cursor.All(ctx, &blocks); err != nil {
		return nil, err
	}
	return blocks, nil
}

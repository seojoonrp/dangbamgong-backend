package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type HealthRepository interface {
	Ping() error
}

type healthRepository struct {
	db *mongo.Database
}

func NewHealthRepository(db *mongo.Database) HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return r.db.Client().Ping(ctx, nil)
}

package database

import (
	"context"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	dbURI  = os.Getenv("DB_URI")
	dbName = os.Getenv("DB_NAME")
)

func New() *mongo.Database {
	if dbURI == "" {
		log.Fatal("DB_URI environment variable is required")
	}

	client, err := mongo.Connect(
		context.Background(),
		options.Client().ApplyURI(dbURI),
	)
	if err != nil {
		log.Fatal(err)
	}

	if dbName == "" {
		dbName = "dangbamgong"
	}

	return client.Database(dbName)
}

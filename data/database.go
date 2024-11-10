package data

import (
	"context"
	"log"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	once        sync.Once
)

func InitMongo() {
	var err error
	mongoURI := os.Getenv("MONGO_URL")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Mongo connection error:", err)
	}
}

// GetMongoClient provides access to the mongoClient instance.
func GetMongoClient() *mongo.Client {
	return mongoClient
}

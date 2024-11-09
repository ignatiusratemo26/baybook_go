package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initMongo() {
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

func main() {
	initMongo()

	r := mux.NewRouter()

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}), // Adjusted origin
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Accept", "X-Requested-With", "Origin"}),
		handlers.AllowCredentials(),
	)
	log.Println("Server running on:4000")
	http.ListenAndServe(":4000", cors(r))
}

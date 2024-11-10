package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"baybook_go/data"
	"baybook_go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetPlaces(w http.ResponseWriter, r *http.Request) {
	placeCollection := data.GetMongoClient().Database("baybookDB").Collection("places")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := placeCollection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var places []models.Place
	if err = cursor.All(ctx, &places); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(places)
}

func UserPlacesHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	placesCollection := data.GetMongoClient().Database("baybookDB").Collection("places")

	cursor, err := placesCollection.Find(context.TODO(), bson.M{"owner": userID})
	if err != nil {
		http.Error(w, "Error fetching places", http.StatusInternalServerError)
		return
	}
	var places []models.Place
	cursor.All(context.TODO(), &places)
	json.NewEncoder(w).Encode(places)
}

// places routes
func CreatePlaceHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	var place models.Place
	json.NewDecoder(r.Body).Decode(&place)
	place.Owner = userID

	// Ensuring the place ID is not set or is set to a new unique value
	place.ID = primitive.NewObjectID()

	placesCollection := data.GetMongoClient().Database("baybookDB").Collection("places")
	res, err := placesCollection.InsertOne(context.TODO(), place)
	if err != nil {
		http.Error(w, "Place creation failed", http.StatusInternalServerError)
		return
	}
	place.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(place)
}

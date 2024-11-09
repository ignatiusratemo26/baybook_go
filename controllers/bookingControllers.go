package controllers

import (
	"baybook_go/data"
	"baybook_go/models"
	"context"
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// booking routes
func CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	var booking models.Booking
	json.NewDecoder(r.Body).Decode(&booking)
	booking.UserID = userID

	bookingsCollection := data.GetMongoClient().Database("baybookDB").Collection("bookings")
	res, err := bookingsCollection.InsertOne(context.TODO(), booking)
	if err != nil {
		http.Error(w, "Booking creation failed", http.StatusInternalServerError)
		return
	}
	booking.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(booking)
}

func UserBookingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	bookingsCollection := data.GetMongoClient().Database("baybookDB").Collection("bookings")

	cursor, err := bookingsCollection.Find(context.TODO(), bson.M{"user": userID})
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}
	var bookings []models.Booking
	cursor.All(context.TODO(), &bookings)
	json.NewEncoder(w).Encode(bookings)
}

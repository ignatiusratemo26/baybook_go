package controllers

import (
	"baybook_go/data"
	"baybook_go/models"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateBookingHandler creates a new booking for a specific place by its ID
func CreateBookingHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the placeID from the URL path
	vars := mux.Vars(r)
	placeID, err := primitive.ObjectIDFromHex(vars["placeID"])
	if err != nil {
		http.Error(w, "Invalid place ID", http.StatusBadRequest)
		return
	}

	// Log the incoming request body for debugging
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	log.Printf("Request body: %s", string(bodyBytes))

	var bookingPayload models.BookingPayload
	var booking models.Booking

	err = json.NewDecoder(r.Body).Decode(&bookingPayload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %s", err), http.StatusBadRequest)
		return
	}
	log.Printf("Request body: %v", r.Body)

	// Parse dates
	check_in, err := time.Parse("2006-01-02", bookingPayload.CheckIn)
	if err != nil {
		http.Error(w, "Invalid check-in date format", http.StatusBadRequest)
		return
	}
	check_out, err := time.Parse("2006-01-02", bookingPayload.CheckOut)
	if err != nil {
		http.Error(w, "Invalid check-out date format", http.StatusBadRequest)
		return
	}

	booking.CheckIn = check_in.UTC()
	booking.CheckOut = check_out.UTC()

	// Set the UserID and PlaceID fields
	booking.ID = primitive.NewObjectID()
	booking.User = userID
	booking.Place = placeID
	booking.CheckIn = check_in
	booking.CheckOut = check_out

	// Insert the booking into the database
	bookingsCollection := data.GetMongoClient().Database("baybookDB").Collection("bookings")
	res, err := bookingsCollection.InsertOne(context.TODO(), booking)
	if err != nil {
		http.Error(w, "Booking creation failed", http.StatusInternalServerError)
		return
	}

	// Return the booking with the inserted ID
	booking.ID = res.InsertedID.(primitive.ObjectID)
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

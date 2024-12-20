package controllers

import (
	"baybook_go/data"
	"baybook_go/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetBookingById(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	bookingID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return

	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var booking models.Booking
	bookingCollection := data.GetMongoClient().Database("baybookDB").Collection("bookings")

	err = bookingCollection.FindOne(ctx, bson.M{"_id": bookingID}).Decode(&booking)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			http.Error(w, "No bookings found", http.StatusNotFound)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

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
	booking.NumberOfGuests = bookingPayload.NumberOfGuests
	booking.Name = bookingPayload.Name
	booking.Phone = bookingPayload.Phone
	booking.Price = bookingPayload.Price

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
	placesCollection := data.GetMongoClient().Database("baybookDB").Collection("places")

	cursor, err := bookingsCollection.Find(context.TODO(), bson.M{"user": userID})
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var bookings []models.Booking
	if err = cursor.All(context.TODO(), &bookings); err != nil {
		http.Error(w, "Error decoding bookings", http.StatusInternalServerError)
		return
	}

	// Prepare the enriched bookings
	var enrichedBookings []models.EnrichedBooking
	for _, booking := range bookings {
		var place models.Place
		err := placesCollection.FindOne(context.TODO(), bson.M{"_id": booking.Place}).Decode(&place)
		if err != nil {
			http.Error(w, "Error fetching place details", http.StatusInternalServerError)
			return
		}

		// Create the enriched booking
		enrichedBooking := models.EnrichedBooking{
			ID:             booking.ID,
			CheckIn:        booking.CheckIn,
			CheckOut:       booking.CheckOut,
			NumberOfGuests: booking.NumberOfGuests,
			Name:           booking.Name,
			Phone:          booking.Phone,
			Price:          booking.Price,
			User:           booking.User,
			Place:          place, // Full place object
		}

		enrichedBookings = append(enrichedBookings, enrichedBooking)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedBookings)
}

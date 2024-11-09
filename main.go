package main

import (
	"baybook_go/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("  ")

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

// auth and user routes
func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate email
	if user.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Ensuring the user ID is not set or is set to a new unique value
	user.ID = primitive.NewObjectID()

	usersCollection := mongoClient.Database("baybookDB").Collection("users")

	// Check if email already exists
	var existingUser models.User
	err = usersCollection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Email already in use", http.StatusUnprocessableEntity)
		return
	} else if err != mongo.ErrNoDocuments {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Insert the new user
	res, err := usersCollection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		http.Error(w, "User creation failed", http.StatusUnprocessableEntity)
		return
	}
	user.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(user)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&credentials)

	usersCollection := mongoClient.Database("baybookDB").Collection("users")

	var user models.User
	err := usersCollection.FindOne(context.TODO(), bson.M{"email": credentials.Email}).Decode(&user)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := generateToken(user.ID)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	})
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	})
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// jwt utility functions
func generateToken(userID primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{
		"id":  userID.Hex(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// getting user from jwt token
func getUserFromToken(r *http.Request) (primitive.ObjectID, error) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return primitive.NilObjectID, err
	}
	token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := primitive.ObjectIDFromHex(claims["id"].(string))
		return userID, nil
	}
	return primitive.NilObjectID, err
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	usersCollection := mongoClient.Database("baybookDB").Collection("users")
	var user models.User
	usersCollection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	json.NewEncoder(w).Encode(user)
}

// places routes
func createPlaceHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	var place models.Place
	json.NewDecoder(r.Body).Decode(&place)
	place.Owner = userID

	// Ensuring the place ID is not set or is set to a new unique value
	place.ID = primitive.NewObjectID()

	placesCollection := mongoClient.Database("baybookDB").Collection("places")
	res, err := placesCollection.InsertOne(context.TODO(), place)
	if err != nil {
		http.Error(w, "Place creation failed", http.StatusInternalServerError)
		return
	}
	place.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(place)
}

func userPlacesHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	placesCollection := mongoClient.Database("baybookDB").Collection("places")

	cursor, err := placesCollection.Find(context.TODO(), bson.M{"owner": userID})
	if err != nil {
		http.Error(w, "Error fetching places", http.StatusInternalServerError)
		return
	}
	var places []models.Place
	cursor.All(context.TODO(), &places)
	json.NewEncoder(w).Encode(places)
}

// booking routes
func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	var booking models.Booking
	json.NewDecoder(r.Body).Decode(&booking)
	booking.UserID = userID

	bookingsCollection := mongoClient.Database("baybookDB").Collection("bookings")
	res, err := bookingsCollection.InsertOne(context.TODO(), booking)
	if err != nil {
		http.Error(w, "Booking creation failed", http.StatusInternalServerError)
		return
	}
	booking.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(booking)
}

func userBookingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromToken(r)
	bookingsCollection := mongoClient.Database("baybookDB").Collection("bookings")

	cursor, err := bookingsCollection.Find(context.TODO(), bson.M{"user": userID})
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}
	var bookings []models.Booking
	cursor.All(context.TODO(), &bookings)
	json.NewEncoder(w).Encode(bookings)
}

func main() {
	initMongo()

	r := mux.NewRouter()
	r.HandleFunc("/api/register", registerHandler).Methods("POST")
	r.HandleFunc("/api/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/logout", logoutHandler).Methods("POST")
	r.HandleFunc("/api/profile", profileHandler).Methods("GET")
	r.HandleFunc("/api/places", createPlaceHandler).Methods("POST")
	r.HandleFunc("/api/user-places", userPlacesHandler).Methods("GET")
	r.HandleFunc("/api/bookings", createBookingHandler).Methods("POST")
	r.HandleFunc("/api/user-bookings", userBookingsHandler).Methods("GET")

	cors := handlers.CORS(handlers.AllowedOrigins([]string{"http://127.0.0.1:5173"}), handlers.AllowCredentials())
	log.Println("Server running on:4000")
	http.ListenAndServe(":4000", cors(r))
}

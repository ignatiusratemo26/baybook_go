package controllers

import (
	"baybook_go/data"
	"baybook_go/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("  ")

// auth and user routes
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	usersCollection := data.GetMongoClient().Database("baybookDB").Collection("users")

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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&credentials)

	usersCollection := data.GetMongoClient().Database("baybookDB").Collection("users")

	var user models.User
	err := usersCollection.FindOne(context.TODO(), bson.M{"email": credentials.Email}).Decode(&user)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, _ := GenerateToken(user.ID)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	})
	json.NewEncoder(w).Encode(user)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	})
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// jwt utility functions
func GenerateToken(userID primitive.ObjectID) (string, error) {
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

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	usersCollection := data.GetMongoClient().Database("baybookDB").Collection("users")
	var user models.User
	usersCollection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&user)
	json.NewEncoder(w).Encode(user)
}

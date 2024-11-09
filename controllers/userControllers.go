package controllers

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

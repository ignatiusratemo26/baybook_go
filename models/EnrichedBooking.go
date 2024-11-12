package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnrichedBooking struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CheckIn        time.Time          `bson:"checkIn" json:"checkIn"`
	CheckOut       time.Time          `bson:"checkOut" json:"checkOut"`
	NumberOfGuests int                `bson:"numberOfGuests" json:"numberOfGuests"`
	Name           string             `bson:"name" json:"name"`
	Phone          string             `bson:"phone" json:"phone"`
	Price          int                `bson:"price" json:"price"`
	User           primitive.ObjectID `bson:"user" json:"user"`
	Place          Place              `json:"place"` // Full place object instead of ObjectID
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	User           primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Place          primitive.ObjectID `bson:"place,omitempty" json:"place,omitempty"`
	CheckIn        time.Time          `bson:"checkIn" json:"checkIn"`
	CheckOut       time.Time          `bson:"checkOut" json:"checkOut"`
	NumberOfGuests int                `bson:"number_of_guests" json:"numberOfGuests"`
	Name           string             `bson:"name" json:"name"`
	Phone          string             `bson:"phone" json:"phone"`
	Price          int                `bson:"price" json:"price"`
}

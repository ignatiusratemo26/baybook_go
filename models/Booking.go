package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID             primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	PlaceID        primitive.ObjectID `bson:"place" json:"place"`
	UserID         primitive.ObjectID `bson:"user" json:"user"`
	CheckIn        time.Time          `bson:"checkIn" json:"checkIn"`
	CheckOut       time.Time          `bson:"checkOut" json:"checkOut"`
	NumberOfGuests int                `bson:"numberOfGuests" json:"numberOfGuests"`
}

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Place struct {
	ID          primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Address     string             `bson:"address" json:"address"`
	Owner       primitive.ObjectID `bson:"owner" json:"owner"`
	Description string             `bson:"description" json:"description"`
	Photos      []string           `bson:"photos" json:"photos"`
	Price       float64            `bson:"price" json:"price"`
	Perks       []string           `bson:"perks" json:"perks"`
	ExtraInfo   string             `bson:"extraInfo" json:"extraInfo"`
	CheckIn     int                `bson:"checkIn" json:"checkIn"`
	CheckOut    int                `bson:"checkOut" json:"checkOut"`
	MaxGuests   int                `bson:"maxGuests" json:"maxGuests"`
}

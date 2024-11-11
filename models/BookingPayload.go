package models

type BookingPayload struct {
	CheckIn        string `json:"checkIn"`  // String to receive date as YYYY-MM-DD
	CheckOut       string `json:"checkOut"` // String to receive date as YYYY-MM-DD
	NumberOfGuests int    `json:"numberOfGuests"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Price          int    `json:"price"`
}

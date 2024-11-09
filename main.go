package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

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

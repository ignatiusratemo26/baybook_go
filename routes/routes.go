package routes

import (
	"baybook_go/controllers"

	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/places", controllers.GetPlaces).Methods("GET")
	r.HandleFunc("/api/places/{id}", controllers.GetPlaceByID).Methods("GET")
	r.HandleFunc("/api/places", controllers.CreatePlaceHandler).Methods("POST")
	r.HandleFunc("/api/user-places", controllers.UserPlacesHandler).Methods("GET")
	r.HandleFunc("/api/register", controllers.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", controllers.LoginHandler).Methods("POST")
	r.HandleFunc("/api/logout", controllers.LogoutHandler).Methods("POST")
	r.HandleFunc("/api/profile", controllers.ProfileHandler).Methods("GET")
	r.HandleFunc("/api/places/{placeID}/bookings", controllers.CreateBookingHandler).Methods("POST")
	r.HandleFunc("/api/bookings", controllers.UserBookingsHandler).Methods("GET")
	r.HandleFunc("/api/bookings/{id}", controllers.GetBookingById).Methods("GET")

	return r
}

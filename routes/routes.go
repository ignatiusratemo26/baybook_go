package routes

import (
	"baybook_go/controllers"

	"github.com/gorilla/mux"
)

func RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	// ...existing code...
	r.HandleFunc("/api/places", controllers.GetPlaces).Methods("GET")
	r.HandleFunc("/api/places", controllers.CreatePlaceHandler).Methods("POST")
	r.HandleFunc("/api/user-places", controllers.UserPlacesHandler).Methods("GET")

	// ...existing code...
	return r
}

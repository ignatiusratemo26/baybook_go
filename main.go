package main

import (
	"log"
	"net/http"

	"baybook_go/data"
	"baybook_go/routes"

	"github.com/gorilla/handlers"
)

func main() {
	data.InitMongo()

	r := routes.RegisterRoutes()

	cors := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}), // Adjusted origin
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Accept", "X-Requested-With", "Origin"}),
		handlers.AllowCredentials(),
	)

	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	err := http.ListenAndServe(":4000", cors(r))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server running on:4000")

}

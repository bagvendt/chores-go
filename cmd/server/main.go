package main

import (
	"log"
	"net/http"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/handlers"
)

func main() {
	// Initialize the database first
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set up the server routes
	http.HandleFunc("/", handlers.MainHandler)
	http.HandleFunc("/routines/", handlers.RoutinesHandler)
	http.HandleFunc("/blueprints", handlers.BlueprintsHandler)
	http.HandleFunc("/blueprints/", handlers.BlueprintsHandler)

	log.Println("Server is starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

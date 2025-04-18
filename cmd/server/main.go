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

	// Set up the server
	http.HandleFunc("/", handlers.MainHandler)

	log.Println("Server is starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

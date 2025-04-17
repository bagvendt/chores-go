package main

import (
	"log"
	"net/http"

	"chores/internal/database"
)

func main() {
	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// TODO: Setup routes
	// TODO: Setup templates

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
} 
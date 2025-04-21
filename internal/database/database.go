package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	var err error

	// Get database URL from environment variable or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "chores.db"
	}
	log.Printf("Using database URL: %s", dbURL)

	DB, err = sql.Open("sqlite3", dbURL)
	if err != nil {
		return err
	}

	// Run migrations
	if err := RunMigrations(); err != nil {
		return err
	}

	return nil
}

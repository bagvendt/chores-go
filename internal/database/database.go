package database

import (
	"database/sql"
	"log"
	"os"

	"github.com/bagvendt/chores/internal/utils/auth"
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

// UpdateUsersWithHashedPasswords updates all users with bcrypt-hashed passwords
// This should be called when migrating from plaintext passwords
func UpdateUsersWithHashedPasswords() error {
	// Get all users
	rows, err := DB.Query(`
		SELECT id, password FROM users
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Process each user
	for rows.Next() {
		var id int64
		var password string
		if err := rows.Scan(&id, &password); err != nil {
			return err
		}

		// Only hash if the password doesn't look like a bcrypt hash
		// bcrypt hashes start with $2a$, $2b$, or $2y$
		if len(password) < 7 || (password[0:4] != "$2a$" && password[0:4] != "$2b$" && password[0:4] != "$2y$") {
			// Hash the password
			hashedPassword, err := auth.HashPassword(password)
			if err != nil {
				return err
			}

			// Update the user's password
			_, err = DB.Exec(`
				UPDATE users SET password = ? WHERE id = ?
			`, hashedPassword, id)
			if err != nil {
				return err
			}

			log.Printf("Updated user ID %d with hashed password", id)
		}
	}

	return rows.Err()
}

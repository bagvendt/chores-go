package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("sqlite3", "chores.db")
	if err != nil {
		return err
	}

	// Run migrations
	if err := RunMigrations(); err != nil {
		return err
	}

	return nil
} 
package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// MigrationManager handles the application of database migrations
type MigrationManager struct {
	db *sql.DB
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{db: db}
}

// EnsureMigrationsTable creates the migrations table if it doesn't exist
func (m *MigrationManager) EnsureMigrationsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			migration_id INTEGER UNIQUE NOT NULL
		)
	`)
	return err
}

// GetAppliedMigrations returns a map of already applied migration IDs
func (m *MigrationManager) GetAppliedMigrations() (map[int]bool, error) {
	applied := make(map[int]bool)

	// First ensure the migrations table exists
	if err := m.EnsureMigrationsTable(); err != nil {
		return nil, fmt.Errorf("error creating migrations table: %w", err)
	}

	rows, err := m.db.Query("SELECT migration_id FROM migrations ORDER BY migration_id")
	if err != nil {
		return nil, fmt.Errorf("error querying migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("error scanning migration row: %w", err)
		}
		applied[id] = true
	}

	return applied, nil
}

// RunMigrations applies any pending migrations
func (m *MigrationManager) RunMigrations() error {
	// Get applied migrations
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Get migration files
	files, err := m.GetMigrationFiles()
	if err != nil {
		return err
	}

	// Sort numerically by migration ID
	sort.Slice(files, func(i, j int) bool {
		return files[i].ID < files[j].ID
	})

	// Apply each migration in order
	for _, file := range files {
		// Skip if already applied
		if applied[file.ID] {
			log.Printf("Migration %d already applied, skipping", file.ID)
			continue
		}

		log.Printf("Applying migration %d: %s", file.ID, file.Path)
		
		// Apply migration within a transaction
		if err := m.ApplyMigration(file); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", file.ID, err)
		}
		
		log.Printf("Migration %d applied successfully", file.ID)
	}

	return nil
}

// MigrationFile represents a SQL migration file
type MigrationFile struct {
	ID   int
	Path string
	SQL  string
}

// GetMigrationFiles finds all SQL migration files
func (m *MigrationManager) GetMigrationFiles() ([]MigrationFile, error) {
	var files []MigrationFile

	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process SQL files
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Extract migration ID from filename
		filename := filepath.Base(path)
		idPart := strings.Split(filename, "_")[0]
		id, err := strconv.Atoi(idPart)
		if err != nil {
			log.Printf("Warning: Skipping file with invalid migration ID format: %s", path)
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", path, err)
		}

		files = append(files, MigrationFile{
			ID:   id,
			Path: path,
			SQL:  string(content),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking migrations directory: %w", err)
	}

	return files, nil
}

// ApplyMigration applies a single migration within a transaction
func (m *MigrationManager) ApplyMigration(file MigrationFile) error {
	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	// Ensure we either commit or rollback
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Execute migration SQL
	if _, err = tx.Exec(file.SQL); err != nil {
		return fmt.Errorf("error executing migration SQL: %w", err)
	}

	// Record migration as applied
	// We do this even if the SQL already inserts into migrations 
	// as a safety measure to ensure it's recorded
	_, err = tx.Exec("INSERT OR IGNORE INTO migrations (migration_id) VALUES (?)", file.ID)
	if err != nil {
		return fmt.Errorf("error recording migration: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// RunMigrations runs all pending migrations in order
func RunMigrations() error {
	manager := NewMigrationManager(DB)
	return manager.RunMigrations()
} 
package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	ID   int
	SQL  string
	Path string
}

// RunMigrations runs all pending migrations in order
func RunMigrations() error {
	// Get list of migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("error getting migration files: %w", err)
	}

	// Get list of applied migrations
	applied, err := getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %w", err)
	}

	// Filter out already applied migrations
	pending := getPendingMigrations(migrations, applied)

	// Apply pending migrations
	for _, migration := range pending {
		if err := applyMigration(migration); err != nil {
			return fmt.Errorf("error applying migration %d: %w", migration.ID, err)
		}
	}

	return nil
}

func getMigrationFiles() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir("migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Extract migration ID from filename
		filename := filepath.Base(path)
		idStr := strings.TrimSuffix(filename, ".sql")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid migration filename: %s", filename)
		}

		// Read migration SQL
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			ID:   id,
			SQL:  string(sql),
			Path: path,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	return migrations, nil
}

func getAppliedMigrations() (map[int]struct{}, error) {
	applied := make(map[int]struct{})

	rows, err := DB.Query("SELECT migration_id FROM migrations")
	if err != nil {
		// If the migrations table doesn't exist yet, return empty map
		if strings.Contains(err.Error(), "no such table") {
			return applied, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		applied[id] = struct{}{}
	}

	return applied, nil
}

func getPendingMigrations(all []Migration, applied map[int]struct{}) []Migration {
	var pending []Migration
	for _, m := range all {
		if _, exists := applied[m.ID]; !exists {
			pending = append(pending, m)
		}
	}
	return pending
}

func applyMigration(migration Migration) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		tx.Rollback()
		return err
	}

	// Record migration as applied
	if _, err := tx.Exec("INSERT INTO migrations (migration_id) VALUES (?)", migration.ID); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
} 
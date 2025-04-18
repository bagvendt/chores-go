package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, string, func()) {
	// Create temp directory for test migrations
	tempDir, err := os.MkdirTemp("", "migration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create temp DB
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, tempDir, cleanup
}

// createTestMigration creates a migration file for testing
func createTestMigration(t *testing.T, dir, name, content string) {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write migration file: %v", err)
	}
}

func TestEnsureMigrationsTable(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	manager := NewMigrationManager(db)
	err := manager.EnsureMigrationsTable()
	if err != nil {
		t.Fatalf("Failed to ensure migrations table: %v", err)
	}

	// Verify table exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query for migrations table: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected migrations table to exist, but it doesn't")
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	manager := NewMigrationManager(db)
	err := manager.EnsureMigrationsTable()
	if err != nil {
		t.Fatalf("Failed to ensure migrations table: %v", err)
	}

	// Insert test migrations
	_, err = db.Exec("INSERT INTO migrations (migration_id) VALUES (1), (2), (5)")
	if err != nil {
		t.Fatalf("Failed to insert test migrations: %v", err)
	}

	// Get applied migrations
	applied, err := manager.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Verify applied migrations
	expected := map[int]bool{1: true, 2: true, 5: true}
	if len(applied) != len(expected) {
		t.Errorf("Expected %d applied migrations, got %d", len(expected), len(applied))
	}

	for id := range expected {
		if !applied[id] {
			t.Errorf("Expected migration %d to be applied, but it wasn't", id)
		}
	}
}

func TestApplyMigration(t *testing.T) {
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	manager := NewMigrationManager(db)
	err := manager.EnsureMigrationsTable()
	if err != nil {
		t.Fatalf("Failed to ensure migrations table: %v", err)
	}

	// Create test migration
	testMigration := MigrationFile{
		ID:   1,
		Path: "1_test.sql",
		SQL:  "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);",
	}

	// Apply migration
	err = manager.ApplyMigration(testMigration)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Verify migration was applied (table exists)
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query for test table: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected test_table to exist, but it doesn't")
	}

	// Verify migration was recorded
	var migrationID int
	err = db.QueryRow("SELECT migration_id FROM migrations WHERE migration_id = 1").Scan(&migrationID)
	if err != nil {
		t.Fatalf("Failed to query for migration record: %v", err)
	}

	if migrationID != 1 {
		t.Errorf("Expected migration ID 1 to be recorded, got %d", migrationID)
	}
}

func TestRunMigrationsInOrder(t *testing.T) {
	db, tempDir, cleanup := setupTestDB(t)
	defer cleanup()

	// Create migrations directory within tempDir
	migrationsDir := filepath.Join(tempDir, "migrations")
	if err := os.Mkdir(migrationsDir, 0755); err != nil {
		t.Fatalf("Failed to create migrations dir: %v", err)
	}

	// Create test migrations
	createTestMigration(t, migrationsDir, "1_first.sql", `
		CREATE TABLE first_table (id INTEGER PRIMARY KEY);
		INSERT INTO first_table (id) VALUES (1);
	`)
	createTestMigration(t, migrationsDir, "2_second.sql", `
		CREATE TABLE second_table (id INTEGER PRIMARY KEY, first_id INTEGER,
		FOREIGN KEY (first_id) REFERENCES first_table(id));
		INSERT INTO second_table (id, first_id) VALUES (1, 1);
	`)
	createTestMigration(t, migrationsDir, "3_third.sql", `
		CREATE TABLE third_table (id INTEGER PRIMARY KEY);
	`)

	// Create migration manager
	manager := NewMigrationManager(db)

	// Override the migrations directory for testing
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run migrations
	if err := manager.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify all tables were created in order
	tables := []string{"first_table", "second_table", "third_table"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query for table %s: %v", table, err)
		}

		if count != 1 {
			t.Errorf("Expected table %s to exist, but it doesn't", table)
		}
	}

	// Verify data was inserted correctly by the first two migrations
	var firstCount, secondCount int
	err = db.QueryRow("SELECT COUNT(*) FROM first_table").Scan(&firstCount)
	if err != nil {
		t.Fatalf("Failed to query first_table: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM second_table").Scan(&secondCount)
	if err != nil {
		t.Fatalf("Failed to query second_table: %v", err)
	}

	if firstCount != 1 || secondCount != 1 {
		t.Errorf("Expected 1 row in each table, got %d in first_table and %d in second_table",
			firstCount, secondCount)
	}

	// Verify all migrations were recorded
	var appliedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&appliedCount)
	if err != nil {
		t.Fatalf("Failed to query migrations table: %v", err)
	}

	if appliedCount != 3 {
		t.Errorf("Expected 3 recorded migrations, got %d", appliedCount)
	}
}

func TestMigrationFailureTransactional(t *testing.T) {
	db, tempDir, cleanup := setupTestDB(t)
	defer cleanup()

	// Create migrations directory within tempDir
	migrationsDir := filepath.Join(tempDir, "migrations")
	if err := os.Mkdir(migrationsDir, 0755); err != nil {
		t.Fatalf("Failed to create migrations dir: %v", err)
	}

	// Create a successful first migration
	createTestMigration(t, migrationsDir, "1_good.sql", `
		CREATE TABLE good_table (id INTEGER PRIMARY KEY);
		INSERT INTO good_table (id) VALUES (1);
	`)

	// Create a failing second migration (invalid SQL)
	createTestMigration(t, migrationsDir, "2_bad.sql", `
		CREATE TABLE bad_table (id INTEGER PRIMARY KEY);
		-- This next statement will fail because the column name is missing
		INSERT INTO bad_table () VALUES (1);
	`)

	// Create a third migration that should never be applied
	createTestMigration(t, migrationsDir, "3_never.sql", `
		CREATE TABLE never_table (id INTEGER PRIMARY KEY);
	`)

	// Create migration manager
	manager := NewMigrationManager(db)

	// Override the migrations directory for testing
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run migrations - should fail on the second one
	err = manager.RunMigrations()
	if err == nil {
		t.Fatalf("Expected RunMigrations to fail, but it succeeded")
	}

	// Verify only the first table was created
	var goodCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='good_table'").Scan(&goodCount)
	if err != nil {
		t.Fatalf("Failed to query for good_table: %v", err)
	}
	if goodCount != 1 {
		t.Errorf("Expected good_table to exist, but it doesn't")
	}

	// Verify the second table doesn't exist due to rollback
	var badCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='bad_table'").Scan(&badCount)
	if err != nil {
		t.Fatalf("Failed to query for bad_table: %v", err)
	}
	if badCount != 0 {
		t.Errorf("Expected bad_table not to exist, but it does")
	}

	// Verify the third table doesn't exist because migration was stopped
	var neverCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='never_table'").Scan(&neverCount)
	if err != nil {
		t.Fatalf("Failed to query for never_table: %v", err)
	}
	if neverCount != 0 {
		t.Errorf("Expected never_table not to exist, but it does")
	}

	// Verify only the first migration was recorded
	var appliedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&appliedCount)
	if err != nil {
		t.Fatalf("Failed to query migrations table: %v", err)
	}
	if appliedCount != 1 {
		t.Errorf("Expected 1 recorded migration, got %d", appliedCount)
	}
}

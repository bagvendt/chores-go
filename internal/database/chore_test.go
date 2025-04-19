package database

import (
	"testing"

	"github.com/bagvendt/chores/internal/models"
)

func TestChoreOperations(t *testing.T) {
	// Setup test database
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Replace the global DB with our test DB
	originalDB := DB
	DB = db
	defer func() { DB = originalDB }()

	// Create the chores table for testing
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS chores (
			id INTEGER PRIMARY KEY,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			name TEXT NOT NULL,
			default_points INTEGER NOT NULL CHECK (default_points > 0),
			image TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create chores table: %v", err)
	}

	// Test CreateChore
	testChore := &models.Chore{
		Name:          "Test Chore",
		DefaultPoints: 10,
		Image:         "test.jpg",
	}

	err = CreateChore(db, testChore)
	if err != nil {
		t.Fatalf("Failed to create chore: %v", err)
	}

	if testChore.ID <= 0 {
		t.Errorf("Expected chore to have an ID, got %d", testChore.ID)
	}

	// Test GetChore
	chore, err := GetChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to get chore: %v", err)
	}

	if chore.Name != testChore.Name {
		t.Errorf("Expected chore name %s, got %s", testChore.Name, chore.Name)
	}

	if chore.DefaultPoints != testChore.DefaultPoints {
		t.Errorf("Expected default points %d, got %d", testChore.DefaultPoints, chore.DefaultPoints)
	}

	if chore.Image != testChore.Image {
		t.Errorf("Expected image %s, got %s", testChore.Image, chore.Image)
	}

	// Test UpdateChore
	updatedName := "Updated Chore"
	testChore.Name = updatedName
	testChore.DefaultPoints = 15

	err = UpdateChore(db, testChore)
	if err != nil {
		t.Fatalf("Failed to update chore: %v", err)
	}

	// Verify update
	updatedChore, err := GetChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to get updated chore: %v", err)
	}

	if updatedChore.Name != updatedName {
		t.Errorf("Expected updated name %s, got %s", updatedName, updatedChore.Name)
	}

	if updatedChore.DefaultPoints != 15 {
		t.Errorf("Expected updated points 15, got %d", updatedChore.DefaultPoints)
	}

	// Test GetChores
	// Add another chore
	secondChore := &models.Chore{
		Name:          "Second Chore",
		DefaultPoints: 5,
	}

	err = CreateChore(db, secondChore)
	if err != nil {
		t.Fatalf("Failed to create second chore: %v", err)
	}

	// Get all chores
	chores, err := GetChores(db)
	if err != nil {
		t.Fatalf("Failed to get chores: %v", err)
	}

	if len(chores) != 2 {
		t.Errorf("Expected 2 chores, got %d", len(chores))
	}

	// Test DeleteChore
	err = DeleteChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to delete chore: %v", err)
	}

	// Verify deletion
	deletedChores, err := GetChores(db)
	if err != nil {
		t.Fatalf("Failed to get chores after deletion: %v", err)
	}

	if len(deletedChores) != 1 {
		t.Errorf("Expected 1 chore after deletion, got %d", len(deletedChores))
	}

	if deletedChores[0].ID != secondChore.ID {
		t.Errorf("Expected remaining chore to be the second chore, got %d", deletedChores[0].ID)
	}
}

func TestChoreNullableImage(t *testing.T) {
	// Setup test database
	db, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Replace the global DB with our test DB
	originalDB := DB
	DB = db
	defer func() { DB = originalDB }()

	// Create the chores table for testing
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS chores (
			id INTEGER PRIMARY KEY,
			created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			name TEXT NOT NULL,
			default_points INTEGER NOT NULL CHECK (default_points > 0),
			image TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create chores table: %v", err)
	}

	// Test CreateChore with no image
	testChore := &models.Chore{
		Name:          "No Image Chore",
		DefaultPoints: 10,
		// No image
	}

	err = CreateChore(db, testChore)
	if err != nil {
		t.Fatalf("Failed to create chore: %v", err)
	}

	// Test GetChore
	chore, err := GetChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to get chore: %v", err)
	}

	if chore.Image != "" {
		t.Errorf("Expected empty image, got %s", chore.Image)
	}

	// Test update from no image to image
	testChore.Image = "new_image.jpg"
	err = UpdateChore(db, testChore)
	if err != nil {
		t.Fatalf("Failed to update chore with image: %v", err)
	}

	// Verify update
	updatedChore, err := GetChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to get updated chore: %v", err)
	}

	if updatedChore.Image != "new_image.jpg" {
		t.Errorf("Expected image new_image.jpg, got %s", updatedChore.Image)
	}

	// Test update from image to no image
	testChore.Image = ""
	err = UpdateChore(db, testChore)
	if err != nil {
		t.Fatalf("Failed to update chore removing image: %v", err)
	}

	// Verify update
	updatedChore, err = GetChore(db, testChore.ID)
	if err != nil {
		t.Fatalf("Failed to get updated chore: %v", err)
	}

	if updatedChore.Image != "" {
		t.Errorf("Expected empty image after removal, got %s", updatedChore.Image)
	}
}

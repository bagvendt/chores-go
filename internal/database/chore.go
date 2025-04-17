package database

import (
	"chores/internal/models"
	"database/sql"
	"time"
)

// GetChores returns all chores from the database
func GetChores() ([]models.Chore, error) {
	rows, err := DB.Query(`
		SELECT id, created, modified, name, default_points, image
		FROM chores
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chores []models.Chore
	for rows.Next() {
		var chore models.Chore
		var createdStr, modifiedStr string
		var image sql.NullString

		if err := rows.Scan(
			&chore.ID,
			&createdStr,
			&modifiedStr,
			&chore.Name,
			&chore.DefaultPoints,
			&image,
		); err != nil {
			return nil, err
		}

		chore.Created, _ = time.Parse(time.RFC3339, createdStr)
		chore.Modified, _ = time.Parse(time.RFC3339, modifiedStr)
		if image.Valid {
			chore.Image = image.String
		}

		chores = append(chores, chore)
	}

	return chores, nil
}

// GetChore returns a chore by ID
func GetChore(id int64) (*models.Chore, error) {
	var chore models.Chore
	var createdStr, modifiedStr string
	var image sql.NullString

	err := DB.QueryRow(`
		SELECT id, created, modified, name, default_points, image
		FROM chores
		WHERE id = ?
	`, id).Scan(
		&chore.ID,
		&createdStr,
		&modifiedStr,
		&chore.Name,
		&chore.DefaultPoints,
		&image,
	)
	if err != nil {
		return nil, err
	}

	chore.Created, _ = time.Parse(time.RFC3339, createdStr)
	chore.Modified, _ = time.Parse(time.RFC3339, modifiedStr)
	if image.Valid {
		chore.Image = image.String
	}

	return &chore, nil
}

// CreateChore creates a new chore in the database
func CreateChore(chore *models.Chore) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := DB.Exec(`
		INSERT INTO chores (created, modified, name, default_points, image)
		VALUES (?, ?, ?, ?, ?)
	`,
		now,
		now,
		chore.Name,
		chore.DefaultPoints,
		sql.NullString{String: chore.Image, Valid: chore.Image != ""},
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	chore.ID = id
	chore.Created, _ = time.Parse(time.RFC3339, now)
	chore.Modified = chore.Created

	return nil
}

// UpdateChore updates an existing chore in the database
func UpdateChore(chore *models.Chore) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := DB.Exec(`
		UPDATE chores
		SET modified = ?, name = ?, default_points = ?, image = ?
		WHERE id = ?
	`,
		now,
		chore.Name,
		chore.DefaultPoints,
		sql.NullString{String: chore.Image, Valid: chore.Image != ""},
		chore.ID,
	)
	if err != nil {
		return err
	}

	chore.Modified, _ = time.Parse(time.RFC3339, now)
	return nil
}

// DeleteChore deletes a chore from the database
func DeleteChore(id int64) error {
	_, err := DB.Exec("DELETE FROM chores WHERE id = ?", id)
	return err
} 
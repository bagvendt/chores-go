package database

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/models"
)

// GetChores returns all chores from the database
func GetChores(db *sql.DB) ([]models.Chore, error) {
	rows, err := db.Query(`
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
func GetChore(db *sql.DB, id int64) (*models.Chore, error) {
	var chore models.Chore
	var createdStr, modifiedStr string
	var image sql.NullString

	err := db.QueryRow(`
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
func CreateChore(db *sql.DB, chore *models.Chore) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(`
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
func UpdateChore(db *sql.DB, chore *models.Chore) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`
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
func DeleteChore(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM chores WHERE id = ?", id)
	return err
}

// UpsertChoreRoutine creates or updates a chore routine
// It takes a routine ID, chore ID, and a completed flag
// If the record exists, it updates the completion status
// If it doesn't exist, it creates a new record
func UpsertChoreRoutine(db *sql.DB, routineID int64, choreID int64, completed bool, userID int64) (*models.ChoreRoutine, error) {
	// First check if the record exists
	var choreRoutine models.ChoreRoutine
	var createdStr, modifiedStr string
	var completedAtStr sql.NullString
	var completedBy sql.NullInt64

	err := db.QueryRow(`
		SELECT id, created, modified, completed_at, completed_by, points_awarded, routine_id, chore_id
		FROM chore_routines
		WHERE routine_id = ? AND chore_id = ?
	`, routineID, choreID).Scan(
		&choreRoutine.ID,
		&createdStr,
		&modifiedStr,
		&completedAtStr,
		&completedBy,
		&choreRoutine.PointsAwarded,
		&choreRoutine.RoutineID,
		&choreRoutine.ChoreID,
	)

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// If record doesn't exist, create it
	if err == sql.ErrNoRows {
		// Get the default points from the chore
		var defaultPoints int
		err := db.QueryRow("SELECT default_points FROM chores WHERE id = ?", choreID).Scan(&defaultPoints)
		if err != nil {
			return nil, err
		}

		// Set completedAt and completedBy based on the completed flag
		var completedAtParam, completedByParam interface{}
		if completed {
			completedAtParam = nowStr
			completedByParam = userID
		} else {
			completedAtParam = nil
			completedByParam = nil
		}

		result, err := db.Exec(`
			INSERT INTO chore_routines (
				created, modified, completed_at, completed_by, 
				points_awarded, routine_id, chore_id
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
			nowStr,
			nowStr,
			completedAtParam,
			completedByParam,
			defaultPoints,
			routineID,
			choreID,
		)
		if err != nil {
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		choreRoutine = models.ChoreRoutine{
			ID:            id,
			Created:       now,
			Modified:      now,
			PointsAwarded: defaultPoints,
			RoutineID:     routineID,
			ChoreID:       choreID,
		}

		if completed {
			choreRoutine.CompletedAt = &now
			choreRoutine.CompletedByID = &userID
		}

		return &choreRoutine, nil
	} else if err != nil {
		return nil, err
	}

	// Record exists, update it
	choreRoutine.Created, _ = time.Parse(time.RFC3339, createdStr)
	choreRoutine.Modified = now

	if completedAtStr.Valid {
		completedAt, _ := time.Parse(time.RFC3339, completedAtStr.String)
		choreRoutine.CompletedAt = &completedAt
	}

	if completedBy.Valid {
		choreRoutine.CompletedByID = &completedBy.Int64
	}

	// If completed status is changing, update the record
	var isCurrentlyCompleted bool = choreRoutine.CompletedAt != nil
	if completed != isCurrentlyCompleted {
		var completedAtParam, completedByParam interface{}
		if completed {
			completedAtParam = nowStr
			completedByParam = userID
		} else {
			completedAtParam = nil
			completedByParam = nil
		}

		_, err := db.Exec(`
			UPDATE chore_routines
			SET modified = ?, completed_at = ?, completed_by = ?
			WHERE id = ?
		`,
			nowStr,
			completedAtParam,
			completedByParam,
			choreRoutine.ID,
		)
		if err != nil {
			return nil, err
		}

		// Update the model
		if completed {
			choreRoutine.CompletedAt = &now
			choreRoutine.CompletedByID = &userID
		} else {
			choreRoutine.CompletedAt = nil
			choreRoutine.CompletedByID = nil
		}
	}

	return &choreRoutine, nil
}

package database

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/models"
)

// GetRoutines retrieves all routines from the database for a specific user
func GetRoutines(db *sql.DB, userID int64) ([]models.Routine, error) {
	rows, err := db.Query(`
		SELECT r.id, r.created, r.modified, r.owner_id, r.routine_blueprint_id, -- Added r.routine_blueprint_id
		       u.name as owner_name, 
		       rb.image as image_url
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		LEFT JOIN routine_blueprints rb ON r.routine_blueprint_id = rb.id
		WHERE r.owner_id = ?
		ORDER BY r.created DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routines []models.Routine
	for rows.Next() {
		var r models.Routine
		var owner models.User
		var created, modified string
		var imageUrl sql.NullString
		err := rows.Scan(
			&r.ID,
			&created,
			&modified,
			&r.OwnerID,
			&r.RoutineBlueprintID, // Scan the new field
			&owner.Name,
			&imageUrl,
		)
		if err != nil {
			return nil, err
		}
		r.Created, _ = time.Parse(time.RFC3339, created)
		r.Modified, _ = time.Parse(time.RFC3339, modified)
		r.Owner = &owner
		if imageUrl.Valid {
			r.ImageUrl = imageUrl.String
		} else {
			r.ImageUrl = ""
		}
		routines = append(routines, r)
	}
	return routines, nil
}

// GetRoutine retrieves a single routine by ID
func GetRoutine(db *sql.DB, id int64) (*models.Routine, error) {
	var r models.Routine
	var owner models.User
	var created, modified string
	var imageUrl sql.NullString
	err := db.QueryRow(`
		SELECT r.id, r.created, r.modified, r.owner_id, r.routine_blueprint_id, -- Added r.routine_blueprint_id
		       u.name as owner_name, 
		       rb.image as image_url
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		LEFT JOIN routine_blueprints rb ON r.routine_blueprint_id = rb.id
		WHERE r.id = ?
	`, id).Scan(
		&r.ID,
		&created,
		&modified,
		&r.OwnerID,
		&r.RoutineBlueprintID, // Scan the new field
		&owner.Name,
		&imageUrl,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Created, _ = time.Parse(time.RFC3339, created)
	r.Modified, _ = time.Parse(time.RFC3339, modified)
	r.Owner = &owner
	if imageUrl.Valid {
		r.ImageUrl = imageUrl.String
	} else {
		r.ImageUrl = ""
	}
	return &r, nil
}

// CreateRoutine creates a new routine
func CreateRoutine(db *sql.DB, routine *models.Routine) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Prepare the routine_blueprint_id for the query
	var blueprintID interface{}
	if routine.RoutineBlueprintID.Valid {
		blueprintID = routine.RoutineBlueprintID.Int64
	} else {
		blueprintID = nil
	}

	result, err := db.Exec(`
		INSERT INTO routines (created, modified, owner_id, routine_blueprint_id)
		VALUES (?, ?, ?, ?)
	`,
		now,
		now,
		routine.OwnerID,
		blueprintID,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	routine.ID = id
	routine.Created, _ = time.Parse(time.RFC3339, now)
	routine.Modified = routine.Created

	return nil
}

// GetChoreCountsForRoutine counts the total number of chores and completed chores for a routine
func GetChoreCountsForRoutine(db *sql.DB, routineID int64) (total int, completed int, err error) {
	// Query to count total chores and completed chores for the routine
	row := db.QueryRow(`
		SELECT 
			COUNT(*), 
			COUNT(CASE WHEN completed_at IS NOT NULL THEN 1 END)
		FROM chore_routines
		WHERE routine_id = ?
	`, routineID)

	// Scan the results
	err = row.Scan(&total, &completed)
	return total, completed, err
}

// GetRelevantRoutines retrieves routines created before today for a specific user
func GetRelevantRoutines(db *sql.DB, userID int64, today time.Time) ([]models.Routine, error) {
	// Convert today to UTC and strip time part to get start of day
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	rows, err := db.Query(`
		SELECT r.id, r.created, r.modified, r.owner_id, r.routine_blueprint_id,
		       NULL as owner_name, 
		       rb.image as image_url
		FROM routines r
		LEFT JOIN routine_blueprints rb ON r.routine_blueprint_id = rb.id
		WHERE r.owner_id = ? AND r.created < ?
		ORDER BY r.created DESC
	`, userID, startOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routines []models.Routine
	for rows.Next() {
		var r models.Routine
		var owner models.User
		var created, modified string
		var imageUrl sql.NullString
		var ownerName sql.NullString
		err := rows.Scan(
			&r.ID,
			&created,
			&modified,
			&r.OwnerID,
			&r.RoutineBlueprintID,
			&ownerName,
			&imageUrl,
		)
		if err != nil {
			return nil, err
		}
		r.Created, _ = time.Parse(time.RFC3339, created)
		r.Modified, _ = time.Parse(time.RFC3339, modified)

		// Create a basic owner with just the ID
		owner.ID = r.OwnerID
		if ownerName.Valid {
			owner.Name = ownerName.String
		}
		r.Owner = &owner

		if imageUrl.Valid {
			r.ImageUrl = imageUrl.String
		} else {
			r.ImageUrl = ""
		}
		routines = append(routines, r)
	}
	return routines, nil
}

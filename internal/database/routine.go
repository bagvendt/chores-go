package database

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/models"
)

// GetRoutines retrieves all routines from the database
func GetRoutines() ([]models.Routine, error) {
	rows, err := DB.Query(`
		SELECT r.id, r.created, r.modified, r.name, r.to_be_completed_by, r.owner_id,
		       u.name as owner_name
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		ORDER BY r.created DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routines []models.Routine
	for rows.Next() {
		var r models.Routine
		var owner models.User
		var created, modified string
		err := rows.Scan(&r.ID, &created, &modified, &r.Name, &r.ToBeCompletedBy, &r.OwnerID, &owner.Name)
		if err != nil {
			return nil, err
		}
		r.Created, _ = time.Parse(time.RFC3339, created)
		r.Modified, _ = time.Parse(time.RFC3339, modified)
		r.Owner = &owner
		routines = append(routines, r)
	}
	return routines, nil
}

// GetRoutine retrieves a single routine by ID
func GetRoutine(id int64) (*models.Routine, error) {
	var r models.Routine
	var owner models.User
	var created, modified string
	err := DB.QueryRow(`
		SELECT r.id, r.created, r.modified, r.name, r.to_be_completed_by, r.owner_id,
		       u.name as owner_name
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		WHERE r.id = ?
	`, id).Scan(&r.ID, &created, &modified, &r.Name, &r.ToBeCompletedBy, &r.OwnerID, &owner.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Created, _ = time.Parse(time.RFC3339, created)
	r.Modified, _ = time.Parse(time.RFC3339, modified)
	r.Owner = &owner
	return &r, nil
}

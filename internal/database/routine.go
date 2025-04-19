package database

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/models"
)

// GetRoutines retrieves all routines from the database for a specific user
func GetRoutines(db *sql.DB, userID int64) ([]models.Routine, error) {
	rows, err := db.Query(`
		SELECT r.id, r.created, r.modified, r.owner_id,
		       u.name as owner_name, 
		       rb.image as image_url -- Get image from routine_blueprints
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		LEFT JOIN routine_blueprints rb ON r.routine_blueprint_id = rb.id -- Join routine_blueprints
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
		SELECT r.id, r.created, r.modified, r.owner_id,
		       u.name as owner_name, 
		       rb.image as image_url -- Get image from routine_blueprints
		FROM routines r
		LEFT JOIN users u ON r.owner_id = u.id
		LEFT JOIN routine_blueprints rb ON r.routine_blueprint_id = rb.id -- Join routine_blueprints
		WHERE r.id = ?
	`, id).Scan(&r.ID, &created, &modified, &r.OwnerID, &owner.Name, &imageUrl)
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

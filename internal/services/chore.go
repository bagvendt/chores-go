package services

import (
	"database/sql"
	"time"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
)

// ChoreService handles business logic related to chores
type ChoreService struct {
	db *sql.DB
}

// NewChoreService creates a new instance of ChoreService
func NewChoreService(db *sql.DB) *ChoreService {
	return &ChoreService{
		db: db,
	}
}

// GetChoresForRoutine retrieves all chores for a given routine, including both
// concrete chore_routines that have been created and synthetic ones based on
// blueprint chores that haven't been created yet
func (s *ChoreService) GetChoresForRoutine(routineID int64) ([]models.ChoreRoutine, error) {
	// Get the routine's blueprint ID directly (since it's not in the Routine model)
	var blueprintID sql.NullInt64
	err := s.db.QueryRow(`
		SELECT routine_blueprint_id 
		FROM routines 
		WHERE id = ?`, routineID).Scan(&blueprintID)
	if err != nil {
		return nil, err
	}

	// Get all existing chore_routines for this routine
	existingChoreRoutines, err := s.getExistingChoreRoutines(routineID)
	if err != nil {
		return nil, err
	}

	// If the routine is not based on a blueprint, just return the existing chore_routines
	if !blueprintID.Valid || blueprintID.Int64 == 0 {
		return existingChoreRoutines, nil
	}

	// Get all blueprint chores for the routine's blueprint
	blueprintChores, err := database.GetBlueprintChores(s.db, blueprintID.Int64)
	if err != nil {
		return nil, err
	}

	// Create a map of existing chore_routines by ChoreID for quick lookup
	existingChoresByID := make(map[int64]bool)
	for _, cr := range existingChoreRoutines {
		existingChoresByID[cr.ChoreID] = true
	}

	// Create synthetic chore_routines for blueprint chores that don't have
	// corresponding chore_routines yet
	var result []models.ChoreRoutine
	result = append(result, existingChoreRoutines...)

	for _, bc := range blueprintChores {
		// Skip if this blueprint chore already has a chore_routine
		if existingChoresByID[bc.ChoreID] {
			continue
		}

		// Create a synthetic chore_routine
		syntheticCR := models.ChoreRoutine{
			RoutineID:     routineID,
			ChoreID:       bc.ChoreID,
			PointsAwarded: bc.Chore.DefaultPoints,
			// Set the Chore field for convenience
			Chore: bc.Chore,
		}

		result = append(result, syntheticCR)
	}

	return result, nil
}

// getExistingChoreRoutines retrieves all existing chore_routines for a given routine
func (s *ChoreService) getExistingChoreRoutines(routineID int64) ([]models.ChoreRoutine, error) {
	rows, err := s.db.Query(`
		SELECT 
			cr.id, cr.created, cr.modified, cr.completed_at, cr.completed_by, 
			cr.points_awarded, cr.routine_id, cr.chore_id,
			c.id, c.name, c.default_points, c.image
		FROM chore_routines cr
		JOIN chores c ON cr.chore_id = c.id
		WHERE cr.routine_id = ?
		ORDER BY cr.id
	`, routineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choreRoutines []models.ChoreRoutine
	for rows.Next() {
		var cr models.ChoreRoutine
		var createdStr, modifiedStr string
		var completedAtStr sql.NullString
		var completedByID sql.NullInt64
		var chore models.Chore
		var image sql.NullString

		if err := rows.Scan(
			&cr.ID,
			&createdStr,
			&modifiedStr,
			&completedAtStr,
			&completedByID,
			&cr.PointsAwarded,
			&cr.RoutineID,
			&cr.ChoreID,
			&chore.ID,
			&chore.Name,
			&chore.DefaultPoints,
			&image,
		); err != nil {
			return nil, err
		}

		// Parse timestamps
		cr.Created, _ = time.Parse(time.RFC3339, createdStr)
		cr.Modified, _ = time.Parse(time.RFC3339, modifiedStr)

		// Handle nullable fields
		if completedAtStr.Valid {
			completedAt, _ := time.Parse(time.RFC3339, completedAtStr.String)
			cr.CompletedAt = &completedAt
		}

		if completedByID.Valid {
			id := completedByID.Int64
			cr.CompletedByID = &id
		}

		if image.Valid {
			chore.Image = image.String
		}

		// Set the joined chore
		cr.Chore = &chore

		choreRoutines = append(choreRoutines, cr)
	}

	return choreRoutines, nil
}

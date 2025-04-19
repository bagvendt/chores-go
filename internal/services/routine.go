package services

import (
	"database/sql"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
)

// RoutineService handles business logic related to routines
type RoutineService struct {
	db *sql.DB
}

// NewRoutineService creates a new instance of RoutineService
func NewRoutineService(db *sql.DB) *RoutineService {
	return &RoutineService{
		db: db,
	}
}

// GetRoutinesToDisplay returns all routines that should be displayed to the user,
// combining both database-stored and virtual routines
func (s *RoutineService) GetRoutinesToDisplay(userID int64) ([]models.DisplayableRoutine, error) {
	// Get routines from database
	dbRoutines, err := database.GetRoutines(s.db, userID)
	if err != nil {
		return nil, err
	}

	// Generate virtual routines from applicable blueprints
	virtualRoutines, err := s.generateRoutinesFromBlueprints(userID)
	if err != nil {
		return nil, err
	}

	// Combine both types
	var allRoutines []models.DisplayableRoutine

	// Convert database routines to DisplayableRoutine
	for _, routine := range dbRoutines {
		displayable := models.DisplayableRoutine{
			ID:         routine.ID,
			ImageUrl:   routine.ImageUrl,
			OwnerID:    routine.OwnerID,
			Owner:      routine.Owner,
			SourceType: models.DatabaseSource,

			// For database routines, we have creation timestamps
			Created:  &routine.Created,
			Modified: &routine.Modified,

			// Store original routine
			FromRoutine: &routine,

			// Get chore counts (would need to be populated elsewhere)
			ChoreCount:      0, // This should be populated from chore_routines
			CompletedChores: 0, // This should be populated from chore_routines
		}

		allRoutines = append(allRoutines, displayable)
	}

	// Add virtual routines to the list
	allRoutines = append(allRoutines, virtualRoutines...)

	return allRoutines, nil
}

// generateRoutinesFromBlueprints creates virtual routines from blueprint templates
// that are applicable for today based on their recurrence settings
func (s *RoutineService) generateRoutinesFromBlueprints(userID int64) ([]models.DisplayableRoutine, error) {
	// Get all blueprints
	blueprints, err := database.GetBlueprints(s.db)
	if err != nil {
		return nil, err
	}

	// Uncomment if needed for day-specific logic
	// today := time.Now().Weekday()

	var virtualRoutines []models.DisplayableRoutine

	for _, blueprint := range blueprints {
		// Check if blueprint is applicable today
		isApplicable := false

		switch blueprint.Recurrence {
		case models.Daily:
			// Daily routines are always applicable
			isApplicable = true
		case models.Weekly:
			// For weekly routines, we might implement more logic based on day of week
			// For now, assume all weekly routines are applicable on any day
			isApplicable = true
		default:
			// No recurrence or unknown, not automatically applicable
			isApplicable = false
		}

		if isApplicable {
			// Get the blueprint chores
			blueprintChores, err := database.GetBlueprintChores(s.db, blueprint.ID)
			if err != nil {
				return nil, err
			}

			// Create a virtual routine from this blueprint
			blueprintID := blueprint.ID

			displayable := models.DisplayableRoutine{
				// Use negative ID to avoid conflicts with database routines
				ID:              -blueprintID,
				Name:            blueprint.Name,
				ToBeCompletedBy: blueprint.ToBeCompletedBy,
				OwnerID:         userID, // Assign to the current user
				SourceType:      models.BlueprintSource,
				BlueprintID:     &blueprintID,

				// Store original blueprint
				FromBlueprint: &blueprint,

				// Chore counts
				ChoreCount:      len(blueprintChores),
				CompletedChores: 0, // None completed by default
			}

			// Set an image URL if available from the chores
			if len(blueprintChores) > 0 && blueprintChores[0].Image != "" {
				displayable.ImageUrl = blueprintChores[0].Image
			} else if len(blueprintChores) > 0 && blueprintChores[0].Chore != nil {
				displayable.ImageUrl = blueprintChores[0].Chore.Image
			}

			virtualRoutines = append(virtualRoutines, displayable)
		}
	}

	return virtualRoutines, nil
}

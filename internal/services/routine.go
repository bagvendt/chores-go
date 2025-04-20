package services

import (
	"database/sql"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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

// GetRelevantRoutines returns a mix of concrete routines that are still active
// and virtual routines generated from blueprints applicable today.
func (s *RoutineService) GetRelevantRoutines(userID int64) ([]models.DisplayableRoutine, error) {
	now := time.Now() // Get the current time
	today := now.Weekday()

	// 1. Get concrete routines from the database for the user
	dbRoutines, err := database.GetRoutines(s.db, userID)
	if err != nil {
		return nil, err
	}

	// 2. Get all blueprints
	blueprints, err := database.GetBlueprints(s.db)
	if err != nil {
		return nil, err
	}
	// Create a map for easy lookup
	blueprintMap := make(map[int64]models.RoutineBlueprint)
	for _, bp := range blueprints {
		blueprintMap[bp.ID] = bp
	}

	var relevantRoutines []models.DisplayableRoutine
	processedBlueprintIDs := make(map[int64]bool) // Track blueprints already added via concrete routines

	// 3. Process concrete routines
	for _, routine := range dbRoutines {
		// Only consider routines linked to a blueprint
		if !routine.RoutineBlueprintID.Valid {
			continue
		}

		blueprintID := routine.RoutineBlueprintID.Int64
		blueprint, exists := blueprintMap[blueprintID]
		if !exists {
			log.Printf("Warning: Routine %d links to non-existent blueprint %d", routine.ID, blueprintID)
			continue
		}

		// TODO: Add more sophisticated relevance check for concrete routines
		// (e.g., check deadline based on blueprint.ToBeCompletedBy and routine.Created,
		// check if already completed today if AllowMultipleInstancesPerDay is false)
		// For now, include all concrete routines linked to a blueprint.

		// Fetch chore counts (Placeholder - needs implementation)
		choreCount, completedChores := s.getChoreCountsForRoutine(routine.ID)

		displayable := models.DisplayableRoutine{
			ID:              routine.ID,
			Name:            blueprint.Name,            // Name from blueprint
			ToBeCompletedBy: blueprint.ToBeCompletedBy, // Deadline from blueprint
			ImageUrl:        routine.ImageUrl,          // Image might be specific to routine instance or fallback to blueprint
			OwnerID:         routine.OwnerID,
			Owner:           routine.Owner,
			SourceType:      models.DatabaseSource,
			BlueprintID:     &blueprintID,
			Created:         &routine.Created,
			Modified:        &routine.Modified,
			ChoreCount:      choreCount,
			CompletedChores: completedChores,
			FromRoutine:     &routine,
			FromBlueprint:   &blueprint,
		}
		if displayable.ImageUrl == "" {
			displayable.ImageUrl = blueprint.Image // Fallback to blueprint image
		}

		relevantRoutines = append(relevantRoutines, displayable)
		processedBlueprintIDs[blueprintID] = true // Mark this blueprint as processed
	}

	// 4. Process blueprints to generate virtual routines for today
	for _, blueprint := range blueprints {
		// Skip if a concrete instance for this blueprint was already added
		if processedBlueprintIDs[blueprint.ID] {
			// TODO: Revisit this based on AllowMultipleInstancesPerDay
			continue
		}

		// Check if blueprint is applicable today
		isApplicable := false
		switch blueprint.Recurrence {
		case models.Daily:
			isApplicable = true
		case models.Weekday:
			if today >= time.Monday && today <= time.Friday {
				isApplicable = true
			}
		case models.Weekly:
			// TODO: Implement weekly logic (e.g., check specific day)
			// isApplicable = (today == blueprint.TargetWeekday) // Example
			isApplicable = true // Assuming applicable for now
		}

		if isApplicable {
			// Fetch blueprint chores to get count and potentially image
			blueprintChores, err := database.GetBlueprintChores(s.db, blueprint.ID)
			if err != nil {
				// Log error but continue if possible
				log.Printf("Error fetching chores for blueprint %d: %v", blueprint.ID, err)
			}

			blueprintID := blueprint.ID
			displayable := models.DisplayableRoutine{
				ID:              -blueprintID, // Negative ID for virtual routines
				Name:            blueprint.Name,
				ToBeCompletedBy: blueprint.ToBeCompletedBy,
				ImageUrl:        blueprint.Image, // Use blueprint image by default
				OwnerID:         userID,
				SourceType:      models.BlueprintSource,
				BlueprintID:     &blueprintID,
				ChoreCount:      len(blueprintChores),
				CompletedChores: 0, // Virtual routines start incomplete
				FromBlueprint:   &blueprint,
			}

			// Try to get a better image from the first chore if blueprint image is missing
			if displayable.ImageUrl == "" && len(blueprintChores) > 0 {
				if blueprintChores[0].Image != "" {
					displayable.ImageUrl = blueprintChores[0].Image
				} else if blueprintChores[0].Chore != nil && blueprintChores[0].Chore.Image != "" {
					displayable.ImageUrl = blueprintChores[0].Chore.Image
				}
			}

			relevantRoutines = append(relevantRoutines, displayable)
		}
	}

	// Sort routines by ToBeCompletedBy
	sort.Slice(relevantRoutines, func(i, j int) bool {
		return getTimeOfDayPriority(relevantRoutines[i].ToBeCompletedBy) < getTimeOfDayPriority(relevantRoutines[j].ToBeCompletedBy)
	})

	return relevantRoutines, nil
}

// getTimeOfDayPriority returns a numerical priority for a given time of day string
// Lower numbers will be displayed first in the UI
func getTimeOfDayPriority(timeStr string) int {
	// Convert to lowercase for case-insensitive matching
	timeStr = strings.ToLower(timeStr)

	// Define priority map for common time periods
	priorities := map[string]int{
		"morning":   10,
		"breakfast": 20,
		"noon":      30,
		"lunch":     40,
		"afternoon": 50,
		"evening":   60,
		"dinner":    70,
		"night":     80,
		"bedtime":   90,
	}

	// Check for exact matches first
	if priority, exists := priorities[timeStr]; exists {
		return priority
	}

	// Check for partial matches
	for key, priority := range priorities {
		if strings.Contains(timeStr, key) {
			return priority
		}
	}

	// Handle specific times (like "8:00", "14:30", etc.)
	timeRegex := regexp.MustCompile(`(\d{1,2})[:\.]?(\d{2})?\s*(am|pm)?`)
	if matches := timeRegex.FindStringSubmatch(timeStr); matches != nil {
		hour, _ := strconv.Atoi(matches[1])

		// Handle AM/PM if present
		if len(matches) >= 4 && matches[3] != "" {
			if strings.ToLower(matches[3]) == "pm" && hour < 12 {
				hour += 12
			} else if strings.ToLower(matches[3]) == "am" && hour == 12 {
				hour = 0
			}
		}

		// Calculate priority: hours converted to minutes (0-23 hours -> 0-1380 minutes)
		minutes := 0
		if len(matches) >= 3 && matches[2] != "" {
			minutes, _ = strconv.Atoi(matches[2])
		}

		return hour*60 + minutes
	}

	// Default priority for unknown time strings - put them at the end
	return 1000
}

// getChoreCountsForRoutine fetches the total and completed chore counts for a specific routine instance.
func (s *RoutineService) getChoreCountsForRoutine(routineID int64) (total int, completed int) {
	// Call the database function to get the counts
	total, completed, err := database.GetChoreCountsForRoutine(s.db, routineID)
	if err != nil {
		log.Printf("Error counting chores for routine %d: %v", routineID, err)
		return 0, 0
	}

	return total, completed
}

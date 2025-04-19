package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/services"
	"github.com/bagvendt/chores/internal/templates"
)

// ChoreWithStatus adds completion status to a chore for display purposes
type ChoreWithStatus struct {
	Chore           models.Chore
	Completed       bool
	IsConcreteChore bool  // True if this is an actual chore_routine, false if synthetic
	ChoreRoutineID  int64 // ID of the chore_routine, or 0 if synthetic
}

// RoutineDetailHandler specifically handles the new routine detail view that displays chore cards
func RoutineDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the routine ID from the URL path
	path := strings.TrimPrefix(r.URL.Path, "/routine/")

	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	// Fetch the routine from the database
	routine, err := database.GetRoutine(database.DB, id)
	if err != nil {
		http.Error(w, "Failed to load routine", http.StatusInternalServerError)
		return
	}
	if routine == nil {
		http.Error(w, "Routine not found", http.StatusNotFound)
		return
	}

	// Use the new ChoreService to fetch chores for this routine
	choreService := services.NewChoreService(database.DB)
	choreRoutines, err := choreService.GetChoresForRoutine(id)
	if err != nil {
		http.Error(w, "Failed to load chores for routine", http.StatusInternalServerError)
		return
	}

	// Convert ChoreRoutines to Chores for the template
	chores := make([]models.Chore, 0, len(choreRoutines))
	choreStatuses := make(map[int64]bool) // Map to track completion status by chore ID

	for _, cr := range choreRoutines {
		if cr.Chore != nil {
			chores = append(chores, *cr.Chore)
			// Track completion status
			choreStatuses[cr.Chore.ID] = cr.CompletedAt != nil
		}
	}

	// Render the routine detail template with chore cards
	content := templates.RoutineDetailWithStatus(*routine, chores, choreStatuses)
	templates.Base(content).Render(r.Context(), w)
}

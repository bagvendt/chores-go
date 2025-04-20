package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/contextkeys"
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

	// Handle the case where we're creating a routine from a blueprint
	if strings.HasPrefix(path, "create-from-blueprint/") {
		blueprintIDStr := strings.TrimPrefix(path, "create-from-blueprint/")
		if r.Method == http.MethodGet || r.Method == http.MethodPost {
			createRoutineFromBlueprint(w, r, blueprintIDStr)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

func createRoutineFromBlueprint(w http.ResponseWriter, r *http.Request, blueprintIDStr string) {
	// Parse the blueprint ID
	blueprintID, err := strconv.ParseInt(blueprintIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	// Get the blueprint from the database
	blueprint, blueprintChores, err := database.GetBlueprint(database.DB, blueprintID)
	if err != nil {
		log.Printf("Failed to get blueprint: %v", err)
		http.Error(w, "Failed to load blueprint", http.StatusInternalServerError)
		return
	}
	if blueprint == nil {
		http.Error(w, "Blueprint not found", http.StatusNotFound)
		return
	}

	// Make sure the user is authenticated
	user, ok := r.Context().Value(contextkeys.UserContextKey).(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create a new routine from the blueprint
	routine := &models.Routine{
		OwnerID: user.ID,
	}

	// Setup RoutineBlueprintID
	routine.RoutineBlueprintID.Int64 = blueprint.ID
	routine.RoutineBlueprintID.Valid = true

	// Use the image from the blueprint if available
	routine.ImageUrl = blueprint.Image

	// Save the new routine to the database
	if err := database.CreateRoutine(database.DB, routine); err != nil {
		log.Printf("Failed to create routine: %v", err)
		http.Error(w, "Failed to create routine", http.StatusInternalServerError)
		return
	}

	// Create chore_routines for each blueprint chore
	for _, bc := range blueprintChores {
		_, err := database.UpsertChoreRoutine(database.DB, routine.ID, bc.ChoreID, false, user.ID)
		if err != nil {
			log.Printf("Warning: Failed to create chore_routine: %v", err)
			// Continue with the rest of the chores
		}
	}

	// Redirect to the new routine detail page
	http.Redirect(w, r, fmt.Sprintf("/routine/%d", routine.ID), http.StatusSeeOther)
}

package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/templates"
)

func BlueprintsHandler(w http.ResponseWriter, r *http.Request) {
	// Strip prefix to get the path
	path := strings.TrimPrefix(r.URL.Path, "/admin/blueprints")

	// Handle different routes
	switch {
	case path == "" || path == "/":
		// Handle list and create
		if r.Method == http.MethodPost {
			createBlueprint(w, r)
		} else {
			listBlueprints(w, r)
		}
	case strings.HasPrefix(path, "/new"):
		newBlueprint(w, r)
	default:
		// Handle detail/edit/delete routes
		idStr := strings.TrimPrefix(path, "/")
		if strings.HasSuffix(idStr, "/edit") {
			idStr = strings.TrimSuffix(idStr, "/edit")
			editBlueprint(w, r, idStr)
		} else {
			blueprintDetail(w, r, idStr)
		}
	}
}

func listBlueprints(w http.ResponseWriter, r *http.Request) {
	blueprints, err := database.GetBlueprints()
	if err != nil {
		http.Error(w, "Failed to load blueprints", http.StatusInternalServerError)
		return
	}

	content := templates.Blueprints(blueprints)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func blueprintDetail(w http.ResponseWriter, r *http.Request, idStr string) {
	switch r.Method {
	case http.MethodGet:
		getBlueprintDetail(w, r, idStr)
	case http.MethodPost:
		updateBlueprint(w, r, idStr)
	case http.MethodDelete:
		deleteBlueprint(w, r, idStr)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getBlueprintDetail(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	blueprint, chores, err := database.GetBlueprint(id)
	if err != nil {
		log.Printf("Failed to load blueprint (ID: %d): %v", id, err)
		http.Error(w, "Failed to load blueprint", http.StatusInternalServerError)
		return
	}
	if blueprint == nil {
		http.Error(w, "Blueprint not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		templates.BlueprintDetail(blueprint, chores).Render(r.Context(), w)
	} else {
		templates.AdminBase(templates.BlueprintDetail(blueprint, chores)).Render(r.Context(), w)
	}
}

func newBlueprint(w http.ResponseWriter, r *http.Request) {
	chores, err := database.GetChores()
	if err != nil {
		http.Error(w, "Failed to load chores", http.StatusInternalServerError)
		return
	}

	blueprint := &models.RoutineBlueprint{}
	content := templates.BlueprintForm(blueprint, chores)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func editBlueprint(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	blueprint, _, err := database.GetBlueprint(id)
	if err != nil {
		http.Error(w, "Failed to load blueprint", http.StatusInternalServerError)
		return
	}
	if blueprint == nil {
		http.Error(w, "Blueprint not found", http.StatusNotFound)
		return
	}

	chores, err := database.GetChores()
	if err != nil {
		http.Error(w, "Failed to load chores", http.StatusInternalServerError)
		return
	}

	content := templates.BlueprintForm(blueprint, chores)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func updateBlueprint(w http.ResponseWriter, r *http.Request, idStr string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	id, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	blueprint := &models.RoutineBlueprint{
		ID:                           id,
		Name:                         r.FormValue("name"),
		ToBeCompletedBy:              r.FormValue("to_be_completed_by"),
		AllowMultipleInstancesPerDay: r.FormValue("allow_multiple_instances_per_day") == "on",
		Recurrence:                   models.RecurrenceType(r.FormValue("recurrence")),
	}

	// Get selected chores
	choreIDs := []int64{}
	for _, idStr := range r.Form["chores"] {
		if choreID, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			choreIDs = append(choreIDs, choreID)
		}
	}

	var saveErr error
	if id == 0 {
		saveErr = database.CreateBlueprint(blueprint, choreIDs)
	} else {
		saveErr = database.UpdateBlueprint(blueprint, choreIDs)
	}

	if saveErr != nil {
		log.Printf("Error saving blueprint (ID: %d): %v", id, saveErr)
		http.Error(w, "Failed to save blueprint", http.StatusInternalServerError)
		return
	}

	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/blueprints")
	} else {
		http.Redirect(w, r, "/blueprints", http.StatusSeeOther)
	}
}

func deleteBlueprint(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	if err := database.DeleteBlueprint(id); err != nil {
		log.Printf("Error deleting blueprint (ID: %d): %v", id, err)
		http.Error(w, "Failed to delete blueprint", http.StatusInternalServerError)
		return
	}

	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/blueprints")
	} else {
		http.Redirect(w, r, "/blueprints", http.StatusSeeOther)
	}
}

// createBlueprint handles creation of new RoutineBlueprint
func createBlueprint(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	blueprint := &models.RoutineBlueprint{
		Name:                         r.FormValue("name"),
		ToBeCompletedBy:              r.FormValue("to_be_completed_by"),
		AllowMultipleInstancesPerDay: r.FormValue("allow_multiple_instances_per_day") == "on",
		Recurrence:                   models.RecurrenceType(r.FormValue("recurrence")),
	}
	// Get selected chores
	choreIDs := []int64{}
	for _, idStr := range r.Form["chores"] {
		if choreID, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			choreIDs = append(choreIDs, choreID)
		}
	}
	// Save new blueprint
	if err := database.CreateBlueprint(blueprint, choreIDs); err != nil {
		log.Printf("Error creating blueprint: %v", err)
		http.Error(w, "Failed to create blueprint", http.StatusInternalServerError)
		return
	}
	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/blueprints")
	} else {
		http.Redirect(w, r, "/blueprints", http.StatusSeeOther)
	}
}

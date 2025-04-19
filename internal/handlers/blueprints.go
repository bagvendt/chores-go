package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/templates"
)

// getImageFiles reads the static/img directory and returns a list of image filenames.
func getImageFiles() ([]string, error) {
	var files []string
	imgDir := "./static/img"
	items, err := os.ReadDir(imgDir)
	if err != nil {
		log.Printf("Error reading image directory %s: %v", imgDir, err)
		return nil, err
	}

	for _, item := range items {
		if !item.IsDir() {
			ext := filepath.Ext(item.Name())
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".svg" {
				files = append(files, item.Name())
			}
		}
	}
	return files, nil
}

func BlueprintsHandler(w http.ResponseWriter, r *http.Request) {
	// Strip prefix to get the path
	path := strings.TrimPrefix(r.URL.Path, "/blueprints")

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
	blueprints, err := database.GetBlueprints(database.DB)
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

	blueprint, chores, err := database.GetBlueprint(database.DB, id)
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
	chores, err := database.GetChores(database.DB)
	if err != nil {
		http.Error(w, "Failed to load chores", http.StatusInternalServerError)
		return
	}

	imageFiles, err := getImageFiles()
	if err != nil {
		// Log the error but continue, maybe show a message in the form?
		log.Printf("Warning: Failed to load image files: %v", err)
		// Or return an error response:
		// http.Error(w, "Failed to load image files", http.StatusInternalServerError)
		// return
	}

	blueprint := &models.RoutineBlueprint{}
	content := templates.BlueprintForm(blueprint, chores, imageFiles)
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

	blueprint, _, err := database.GetBlueprint(database.DB, id)
	if err != nil {
		http.Error(w, "Failed to load blueprint", http.StatusInternalServerError)
		return
	}
	if blueprint == nil {
		http.Error(w, "Blueprint not found", http.StatusNotFound)
		return
	}

	chores, err := database.GetChores(database.DB)
	if err != nil {
		http.Error(w, "Failed to load chores", http.StatusInternalServerError)
		return
	}

	imageFiles, err := getImageFiles()
	if err != nil {
		log.Printf("Warning: Failed to load image files: %v", err)
		// Continue without images rather than failing completely
	}

	content := templates.BlueprintForm(blueprint, chores, imageFiles)
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
		Image:                        r.FormValue("image"),
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
		saveErr = database.CreateBlueprint(database.DB, blueprint, choreIDs)
	} else {
		saveErr = database.UpdateBlueprint(database.DB, blueprint, choreIDs)
	}

	if saveErr != nil {
		log.Printf("Error saving blueprint (ID: %d): %v", id, saveErr)
		http.Error(w, "Failed to save blueprint", http.StatusInternalServerError)
		return
	}

	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/blueprints")
	} else {
		http.Redirect(w, r, "/admin/blueprints", http.StatusSeeOther)
	}
}

func deleteBlueprint(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blueprint ID", http.StatusBadRequest)
		return
	}

	if err := database.DeleteBlueprint(database.DB, id); err != nil {
		log.Printf("Error deleting blueprint (ID: %d): %v", id, err)
		http.Error(w, "Failed to delete blueprint", http.StatusInternalServerError)
		return
	}

	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/blueprints")
	} else {
		http.Redirect(w, r, "/admin/blueprints", http.StatusSeeOther)
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
		Image:                        r.FormValue("image"),
	}
	// Get selected chores
	choreIDs := []int64{}
	for _, idStr := range r.Form["chores"] {
		if choreID, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			choreIDs = append(choreIDs, choreID)
		}
	}
	// Save new blueprint
	if err := database.CreateBlueprint(database.DB, blueprint, choreIDs); err != nil {
		log.Printf("Error creating blueprint: %v", err)
		http.Error(w, "Failed to create blueprint", http.StatusInternalServerError)
		return
	}
	// Redirect to blueprint list
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/blueprints")
	} else {
		http.Redirect(w, r, "/admin/blueprints", http.StatusSeeOther)
	}
}

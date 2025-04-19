package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/templates"
)

func RoutinesHandler(w http.ResponseWriter, r *http.Request) {
	// Strip prefix to get the path
	path := strings.TrimPrefix(r.URL.Path, "/admin/routines")

	// Handle different routes
	switch {
	case path == "" || path == "/":
		listRoutines(w, r)
	case strings.HasPrefix(path, "/new"):
		newRoutine(w, r)
	default:
		// Assume it's a routine detail page
		routineDetail(w, r, strings.TrimPrefix(path, "/"))
	}
}

func listRoutines(w http.ResponseWriter, r *http.Request) {
	// TODO: Replace '1' with actual user ID
	routines, err := database.GetRoutines(database.DB, 1)
	if err != nil {
		log.Printf("Failed to load routines: %v", err)
		http.Error(w, "Failed to load routines", http.StatusInternalServerError)
		return
	}

	content := templates.Routines(routines)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func routineDetail(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	routine, err := database.GetRoutine(database.DB, id)
	if err != nil {
		http.Error(w, "Failed to load routine", http.StatusInternalServerError)
		return
	}
	if routine == nil {
		http.Error(w, "Routine not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		// If it's an HTMX request, only render the detail component
		templates.RoutineDetail(routine).Render(r.Context(), w)
	} else {
		// For regular requests, render the full page with the routine list
		// TODO: Replace '1' with actual user ID
		routines, err := database.GetRoutines(database.DB, 1)
		if err != nil {
			http.Error(w, "Failed to load routines", http.StatusInternalServerError)
			return
		}
		content := templates.Routines(routines)
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func newRoutine(w http.ResponseWriter, r *http.Request) {
	// TODO: Render new routine form
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

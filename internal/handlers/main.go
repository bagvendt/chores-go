package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/contextkeys"
	"github.com/bagvendt/chores/internal/database" // Add database import
	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	content := templates.MainContent()
	templates.AdminBase(content).Render(r.Context(), w)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch routines for a specific user (e.g., user ID 1 for now)
	// TODO: Replace '1' with actual user ID from session/context
	user, ok := r.Context().Value(contextkeys.UserContextKey).(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	routines, err := database.GetRoutines(database.DB, user.ID)
	if err != nil {
		// Handle error appropriately, maybe show an error page or log
		http.Error(w, "Failed to load routines", http.StatusInternalServerError)
		return
	}

	// Pass routines to the template
	content := templates.Home(routines)
	templates.Base(content).Render(r.Context(), w)
}

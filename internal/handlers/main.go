package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/database" // Add database import
	"github.com/bagvendt/chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	content := templates.MainContent()
	templates.AdminBase(content).Render(r.Context(), w)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch routines
	routines, err := database.GetRoutines()
	if err != nil {
		// Handle error appropriately, maybe show an error page or log
		http.Error(w, "Failed to load routines", http.StatusInternalServerError)
		return
	}

	// Pass routines to the template
	content := templates.Home(routines)
	templates.Base(content).Render(r.Context(), w)
}

package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	content := templates.MainContent()
	templates.AdminBase(content).Render(r.Context(), w)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//empty template
	content := templates.Home()
	templates.Base(content).Render(r.Context(), w)
}

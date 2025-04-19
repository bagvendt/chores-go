package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	content := templates.MainContent()
	templates.Main(content).Render(r.Context(), w)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//empty template
	templates.Home().Render(r.Context(), w)
}

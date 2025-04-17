package handlers

import (
	"net/http"

	"chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Main()
	component.Render(r.Context(), w)
} 
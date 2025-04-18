package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/templates"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	templates.Main().Render(r.Context(), w)
}

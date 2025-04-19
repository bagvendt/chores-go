package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"log"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/templates"
)

func ChoresHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/chores")

	switch {
	case path == "" || path == "/":
		if r.Method == http.MethodPost {
			createChore(w, r)
		} else {
			listChores(w, r)
		}
	case strings.HasPrefix(path, "/new"):
		newChore(w, r)
	default:
		idStr := strings.TrimPrefix(path, "/")
		if strings.HasSuffix(idStr, "/edit") {
			idStr = strings.TrimSuffix(idStr, "/edit")
			editChore(w, r, idStr)
		} else if r.Method == http.MethodDelete {
			deleteChore(w, r, idStr)
		} else {
			choreDetail(w, r, idStr)
		}
	}
}

func listChores(w http.ResponseWriter, r *http.Request) {
	chores, err := database.GetChores(database.DB)
	if err != nil {
		http.Error(w, "Failed to load chores", http.StatusInternalServerError)
		return
	}
	content := templates.Chores(chores)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func choreDetail(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chore ID", http.StatusBadRequest)
		return
	}
	chore, err := database.GetChore(database.DB, id)
	if err != nil {
		http.Error(w, "Failed to load chore", http.StatusInternalServerError)
		return
	}
	if chore == nil {
		http.Error(w, "Chore not found", http.StatusNotFound)
		return
	}
	content := templates.ChoreDetail(chore)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func newChore(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createChore(w, r)
		return
	}
	chore := &models.Chore{}

	imageFiles, err := getImageFiles()
	if err != nil {
		log.Printf("Error getting image files: %v", err)
		http.Error(w, "Failed to load image options", http.StatusInternalServerError)
		return
	}

	content := templates.ChoreForm(chore, imageFiles)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func createChore(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	chore := &models.Chore{
		Name:          r.FormValue("name"),
		DefaultPoints: atoiOrZero(r.FormValue("default_points")),
		Image:         r.FormValue("image"),
	}
	if err := database.CreateChore(database.DB, chore); err != nil {
		http.Error(w, "Failed to create chore", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/chores")
	} else {
		http.Redirect(w, r, "/admin/chores", http.StatusSeeOther)
	}
}

func editChore(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chore ID", http.StatusBadRequest)
		return
	}
	chore, err := database.GetChore(database.DB, id)
	if err != nil {
		http.Error(w, "Failed to load chore", http.StatusInternalServerError)
		return
	}
	if chore == nil {
		http.Error(w, "Chore not found", http.StatusNotFound)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		chore.Name = r.FormValue("name")
		chore.DefaultPoints = atoiOrZero(r.FormValue("default_points"))
		chore.Image = r.FormValue("image")
		if err := database.UpdateChore(database.DB, chore); err != nil {
			http.Error(w, "Failed to update chore", http.StatusInternalServerError)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/admin/chores")
		} else {
			http.Redirect(w, r, "/admin/chores", http.StatusSeeOther)
		}
		return
	}

	imageFiles, err := getImageFiles()
	if err != nil {
		log.Printf("Error getting image files: %v", err)
		http.Error(w, "Failed to load image options", http.StatusInternalServerError)
		return
	}

	content := templates.ChoreForm(chore, imageFiles)
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
	} else {
		templates.AdminBase(content).Render(r.Context(), w)
	}
}

func deleteChore(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid chore ID", http.StatusBadRequest)
		return
	}
	if err := database.DeleteChore(database.DB, id); err != nil {
		http.Error(w, "Failed to delete chore", http.StatusInternalServerError)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/chores")
	} else {
		http.Redirect(w, r, "/admin/chores", http.StatusSeeOther)
	}
}

func atoiOrZero(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

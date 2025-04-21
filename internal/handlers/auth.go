package handlers

import (
	"net/http"

	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/services"
	"github.com/bagvendt/chores/internal/templates"
)

// LoginHandler handles the login page
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle login form submission
		username := r.FormValue("username")
		password := r.FormValue("password")

		_, sessionToken, err := services.AuthenticateUser(database.DB, username, password)
		if err != nil {
			// Authentication failed, show error
			templates.LoginPage("Invalid username or password").Render(r.Context(), w)
			return
		}

		// Set the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Path:     "/",
			HttpOnly: true,
		})

		// Redirect to home page after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Show the login form for GET requests
	templates.LoginPage("").Render(r.Context(), w)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Clear the session from memory
		services.ClearSession(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

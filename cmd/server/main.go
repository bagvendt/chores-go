package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/bagvendt/chores/internal/contextkeys"
	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/handlers"
	"github.com/bagvendt/chores/internal/services"
)

// authMiddlewareHandler wraps a http.Handler with authentication logic,
// and attaches the authenticated user to the request context.
func authMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// No session cookie found, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate the session token
		user, valid := services.ValidateSession(cookie.Value)
		if !valid {
			// Invalid session, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Attach the authenticated user to context
		ctx := context.WithValue(r.Context(), contextkeys.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	// Initialize the database first
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Public routes (without auth)
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/login", handlers.LoginHandler)   // Login route
	publicMux.HandleFunc("/logout", handlers.LogoutHandler) // Logout route

	// Static files (should be accessible without auth)
	fs := http.FileServer(http.Dir(filepath.Join(".", "static")))
	publicMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Root mux for all routes (protected by auth)
	protectedMux := http.NewServeMux()

	// Public/Home route (now protected)
	protectedMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handlers.HomeHandler(w, r)
	})
	protectedMux.HandleFunc("/routine/", handlers.RoutineDetailHandler) // New route for routine detail view with chore cards

	// API routes
	protectedMux.HandleFunc("/api/", handlers.APIHandler)

	// Admin sub-mux for structured admin routes
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/", handlers.MainHandler)
	adminMux.HandleFunc("/routines/", handlers.RoutinesHandler)
	adminMux.HandleFunc("/blueprints", handlers.BlueprintsHandler)
	adminMux.HandleFunc("/blueprints/", handlers.BlueprintsHandler)
	adminMux.HandleFunc("/chores", handlers.ChoresHandler)
	adminMux.HandleFunc("/chores/", handlers.ChoresHandler)
	protectedMux.Handle("/admin/", http.StripPrefix("/admin", adminMux))

	// Wrap protected routes in auth middleware
	protectedHandler := authMiddlewareHandler(protectedMux)

	// Main mux that combines public and protected routes
	mainMux := http.NewServeMux()

	// Add public routes first (no auth)
	mainMux.HandleFunc("/login", publicMux.ServeHTTP)
	mainMux.HandleFunc("/logout", publicMux.ServeHTTP)
	mainMux.Handle("/static/", publicMux)

	// Add protected routes (with auth)
	mainMux.HandleFunc("/", protectedHandler.ServeHTTP)
	mainMux.HandleFunc("/routine/", protectedHandler.ServeHTTP)
	mainMux.HandleFunc("/api/", protectedHandler.ServeHTTP)
	mainMux.Handle("/admin/", protectedHandler)

	log.Println("Server is starting on port 8080...")
	if err := http.ListenAndServe(":8080", mainMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

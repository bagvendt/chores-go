package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bagvendt/chores/internal/contextkeys"
	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/handlers"
	"github.com/bagvendt/chores/internal/models"
)

// authMiddlewareHandler wraps a http.Handler with authentication logic,
// and attaches the authenticated user to the request context.
func authMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Replace with real auth logic (session, JWT, etc.)
		username, password, ok := r.BasicAuth()
		if !ok || !validateUser(username, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Attach the authenticated user to context
		user := &models.User{
			ID:       1, // TODO: Replace with actual user ID from DB
			Created:  time.Now(),
			Modified: time.Now(),
			Name:     username,
			Password: password,
		}
		ctx := context.WithValue(r.Context(), contextkeys.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateUser is a stub for user credential validation.
func validateUser(username, password string) bool {
	// TODO: Implement actual user validation, e.g., lookup in database
	return username == "admin" && password == "secret"
}

func main() {
	// Initialize the database first
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Root mux for all routes
	rootMux := http.NewServeMux()

	// Public/Home route (now protected)
	rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		handlers.HomeHandler(w, r)
	})
	rootMux.HandleFunc("/routine/", handlers.RoutineDetailHandler) // New route for routine detail view with chore cards

	// Static files (now protected)
	fs := http.FileServer(http.Dir(filepath.Join(".", "static")))
	rootMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Admin sub-mux for structured admin routes
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/", handlers.MainHandler)
	adminMux.HandleFunc("/routines/", handlers.RoutinesHandler)
	adminMux.HandleFunc("/blueprints", handlers.BlueprintsHandler)
	adminMux.HandleFunc("/blueprints/", handlers.BlueprintsHandler)
	adminMux.HandleFunc("/chores", handlers.ChoresHandler)
	adminMux.HandleFunc("/chores/", handlers.ChoresHandler)
	rootMux.Handle("/admin/", http.StripPrefix("/admin", adminMux))

	// Wrap all routes in auth middleware
	handler := authMiddlewareHandler(rootMux)

	log.Println("Server is starting on port 8080...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

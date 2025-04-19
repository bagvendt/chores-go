package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bagvendt/chores/internal/contextkeys"
	"github.com/bagvendt/chores/internal/database"
	"github.com/bagvendt/chores/internal/models"
)

// ChoreCompletionRequest is the request body for updating chore completion status
type ChoreCompletionRequest struct {
	Completed bool `json:"completed"`
}

// ChoreCompletionResponse is the response body after updating chore completion
type ChoreCompletionResponse struct {
	Success      bool                 `json:"success"`
	ChoreRoutine *models.ChoreRoutine `json:"chore_routine,omitempty"`
	Error        string               `json:"error,omitempty"`
}

// APIHandler handles API requests
func APIHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")

	// Route to appropriate handler based on path
	switch {
	case strings.HasPrefix(path, "/routine/"):
		handleRoutineAPI(w, r, strings.TrimPrefix(path, "/routine/"))
	default:
		http.Error(w, "API endpoint not found", http.StatusNotFound)
	}
}

// handleRoutineAPI handles routine-related API requests
func handleRoutineAPI(w http.ResponseWriter, r *http.Request, path string) {
	// Extract routineID from path
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid routine ID", http.StatusBadRequest)
		return
	}

	routineID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid routine ID format", http.StatusBadRequest)
		return
	}

	// Route to specific routine API endpoint
	restPath := strings.Join(parts[1:], "/")
	switch {
	case strings.HasPrefix(restPath, "chore/"):
		handleRoutineChoreAPI(w, r, routineID, strings.TrimPrefix(restPath, "chore/"))
	default:
		http.Error(w, "API endpoint not found", http.StatusNotFound)
	}
}

// handleRoutineChoreAPI handles chore-related API requests for a specific routine
func handleRoutineChoreAPI(w http.ResponseWriter, r *http.Request, routineID int64, path string) {
	// Extract choreID from path
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		http.Error(w, "Invalid chore ID", http.StatusBadRequest)
		return
	}

	choreID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid chore ID format", http.StatusBadRequest)
		return
	}

	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure user is authenticated
	user, ok := r.Context().Value(contextkeys.UserContextKey).(*models.User)
	if !ok || user == nil {
		sendJSONResponse(w, http.StatusUnauthorized, ChoreCompletionResponse{
			Success: false,
			Error:   "User not authenticated",
		})
		return
	}

	// Parse request body
	var req ChoreCompletionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, ChoreCompletionResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Update chore completion status
	choreRoutine, err := database.UpsertChoreRoutine(database.DB, routineID, choreID, req.Completed, user.ID)
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, ChoreCompletionResponse{
			Success: false,
			Error:   "Failed to update chore status: " + err.Error(),
		})
		return
	}

	// Return success response
	sendJSONResponse(w, http.StatusOK, ChoreCompletionResponse{
		Success:      true,
		ChoreRoutine: choreRoutine,
	})
}

// sendJSONResponse sends a JSON response with the given status code and data
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

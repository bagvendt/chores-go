package services

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/bagvendt/chores/internal/models"
	"github.com/bagvendt/chores/internal/utils/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
)

// Sessions is an in-memory session store
var Sessions = make(map[string]*models.User)

// GenerateSessionToken generates a simple session token
// In a production environment, use a more secure method of token generation
func GenerateSessionToken(userID int64) string {
	// Simple implementation for a toy app - using timestamp and userID
	return time.Now().Format("20060102150405") + "-" + strconv.FormatInt(userID, 10)
}

// AuthenticateUser authenticates a user with username and password
func AuthenticateUser(db *sql.DB, username, password string) (*models.User, string, error) {
	var user models.User
	var hashedPassword string
	var createdStr, modifiedStr string

	err := db.QueryRow(`
		SELECT id, created, modified, name, password
		FROM users
		WHERE name = ?
	`, username).Scan(&user.ID, &createdStr, &modifiedStr, &user.Name, &hashedPassword)

	if err == sql.ErrNoRows {
		return nil, "", ErrUserNotFound
	}
	if err != nil {
		return nil, "", err
	}

	// Parse timestamps
	user.Created, _ = time.Parse(time.RFC3339, createdStr)
	user.Modified, _ = time.Parse(time.RFC3339, modifiedStr)

	// Check password
	err = auth.ComparePasswords(hashedPassword, password)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate session token
	sessionToken := GenerateSessionToken(user.ID)

	// Store session
	Sessions[sessionToken] = &user

	return &user, sessionToken, nil
}

// GetUserByID returns a user by ID
func GetUserByID(db *sql.DB, id int64) (*models.User, error) {
	var user models.User
	var createdStr, modifiedStr string

	err := db.QueryRow(`
		SELECT id, created, modified, name, password
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &createdStr, &modifiedStr, &user.Name, &user.Password)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	// Parse timestamps
	user.Created, _ = time.Parse(time.RFC3339, createdStr)
	user.Modified, _ = time.Parse(time.RFC3339, modifiedStr)

	return &user, nil
}

// ValidateSession checks if a session token is valid
func ValidateSession(sessionToken string) (*models.User, bool) {
	user, exists := Sessions[sessionToken]
	return user, exists
}

// ClearSession removes a session from the session store
func ClearSession(sessionToken string) {
	delete(Sessions, sessionToken)
}

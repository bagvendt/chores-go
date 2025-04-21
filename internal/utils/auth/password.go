// filepath: /home/bagvendt/kode/chores-go/internal/utils/auth/password.go
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12 // Number of hashing iterations
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// ComparePasswords compares a hashed password with a plain-text password
func ComparePasswords(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

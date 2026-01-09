package auth

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword enforces password complexity policy
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// CheckPasswordHistory checks if the new password has been used recently (in last 5).
func CheckPasswordHistory(history []string, newPassword string) bool {
	for _, oldHash := range history {
		if CheckPasswordHash(newPassword, oldHash) {
			return true
		}
	}
	return false
}

// AppendPasswordHistory adds new hash and keeps max 5
func AppendPasswordHistory(history []string, newHash string) []string {
	// Add to end
	history = append(history, newHash)
	if len(history) > 5 {
		// Keep last 5
		history = history[len(history)-5:]
	}
	return history
}

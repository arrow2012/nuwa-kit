package auth

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	RecoveryCodeLength = 8  // 8-character recovery codes
	RecoveryCodeCount  = 10 // Generate 10 recovery codes
)

// GenerateRecoveryCodes generates N random recovery codes
// Returns plaintext codes for display to user
func GenerateRecoveryCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateRandomCode(RecoveryCodeLength)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// generateRandomCode generates a random alphanumeric code
func generateRandomCode(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude ambiguous characters
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	// Format as XXXX-XXXX for readability
	if length == 8 {
		return fmt.Sprintf("%s-%s", string(b[:4]), string(b[4:])), nil
	}
	return string(b), nil
}

// HashRecoveryCode hashes a recovery code using bcrypt
// This allows secure storage and one-time use verification
func HashRecoveryCode(code string) (string, error) {
	// Normalize: remove dashes and convert to uppercase
	normalized := strings.ToUpper(strings.ReplaceAll(code, "-", ""))

	// Use bcrypt for hashing (same as passwords)
	hashed, err := bcrypt.GenerateFromPassword([]byte(normalized), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// HashRecoveryCodes hashes multiple recovery codes
func HashRecoveryCodes(codes []string) ([]string, error) {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		h, err := HashRecoveryCode(code)
		if err != nil {
			return nil, err
		}
		hashed[i] = h
	}
	return hashed, nil
}

// VerifyRecoveryCode verifies a recovery code against hashed codes
// Returns (matched, index) where index is the position of the matched code
// Returns (-1, false) if no match found
func VerifyRecoveryCode(hashedCodes []string, inputCode string) (int, bool) {
	// Normalize input
	normalized := strings.ToUpper(strings.ReplaceAll(inputCode, "-", ""))

	for i, hashedCode := range hashedCodes {
		if hashedCode == "" {
			continue // Skip already-used codes
		}

		err := bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(normalized))
		if err == nil {
			return i, true
		}
	}
	return -1, false
}

// InvalidateRecoveryCode marks a recovery code as used by clearing it
func InvalidateRecoveryCode(hashedCodes []string, index int) []string {
	if index >= 0 && index < len(hashedCodes) {
		hashedCodes[index] = "" // Clear the used code
	}
	return hashedCodes
}

// CountRemainingCodes counts how many recovery codes are still valid
func CountRemainingCodes(hashedCodes []string) int {
	count := 0
	for _, code := range hashedCodes {
		if code != "" {
			count++
		}
	}
	return count
}

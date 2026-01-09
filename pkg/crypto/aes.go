package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Encrypt encrypts plain text string into base64 encoded string using AES-GCM
// key must be 32 bytes for AES-256
// Format: v1:nonce:ciphertext
func Encrypt(plaintext, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("v1:%s", base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// Decrypt decrypts base64 encoded string using AES-GCM
// It supports reading old plaintext data for migration (returns original if not matching encrypted format)
func Decrypt(ciphertext, key string) (string, error) {
	// 1. Compatibility Check: If not starting with v1:, assume it's plaintext (old data)
	if !strings.HasPrefix(ciphertext, "v1:") {
		return ciphertext, nil
	}

	if len(key) != 32 {
		return "", errors.New("encryption key must be 32 bytes")
	}

	// Remove prefix
	raw := strings.TrimPrefix(ciphertext, "v1:")

	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
)

// GenerateAccessKey ... (existing)

// GenerateRSAKeyPair generates a new RSA key pair of 2048 bits
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

// PrivateKeyToPEM encodes Private Key to PEM
func PrivateKeyToPEM(priv *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		},
	)
	return string(privPEM)
}

// PublicKeyToPEM encodes Public Key to PEM
func PublicKeyToPEM(pub *rsa.PublicKey) string {
	pubASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return ""
	}
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubASN1,
		},
	)
	return string(pubPEM)
}

// ParsePrivateKeyFromPEM parses a PEM encoded private key
func ParsePrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("key must be RSA PRIVATE KEY")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ParsePublicKeyFromPEM parses a PEM encoded public key
func ParsePublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	if block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("key must be RSA PUBLIC KEY")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key must be RSA PublicKey")
	}
}

// GenerateAccessKey generates a random Access Key (AK)
// Format: NW + 18 characters random alphanumeric (approx)
// Example: NWABC123...
func GenerateAccessKey() (string, error) {
	b := make([]byte, 14) // 14 bytes -> base64 -> ~19 chars.
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	// Use hex/base32/base64. AWS uses alphanumeric uppercase.
	// Let's use simple hex or base32 for simplicity and readability.
	// 20 chars total. "AK" prefix (2 chars) + 18 chars.
	// 18 chars hex = 9 bytes.

	randomBytes := make([]byte, 10) // 10 bytes = 20 hex chars
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}
	return "AK" + hex.EncodeToString(randomBytes), nil
}

// GenerateSecretKey generates a random Secret Key (SK)
// Typically 40 characters.
func GenerateSecretKey() (string, error) {
	b := make([]byte, 30) // 30 bytes * 8 / 6 = 40 base64 chars
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

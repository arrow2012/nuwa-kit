package options

import (
	"fmt"
	"time"
)

// AuthOptions contains authentication-specific configuration
type AuthOptions struct {
	JWTSecret         string        `json:"jwtSecret" mapstructure:"jwtSecret"`
	EncryptionKey     string        `json:"encryptionKey" mapstructure:"encryptionKey"` // 32 bytes for AES-256
	TokenDuration     time.Duration `json:"tokenDuration" mapstructure:"tokenDuration"`
	SendCodeRateLimit time.Duration `json:"sendCodeRateLimit" mapstructure:"sendCodeRateLimit"`
	Issuer            string        `json:"issuer" mapstructure:"issuer"`
	PrivateKey        string        `json:"privateKey" mapstructure:"privateKey"`
	PublicKey         string        `json:"publicKey" mapstructure:"publicKey"`
}

// NewServerOptions create a `zero` value instance.
func NewAuthOptions() *AuthOptions {
	return &AuthOptions{
		JWTSecret:         "1234567890",
		EncryptionKey:     "12345678901234567890123456789012", // Default 32 bytes key for dev
		TokenDuration:     24 * time.Hour,
		SendCodeRateLimit: 1 * time.Minute,
		Issuer:            "http://localhost:8080",
	}
}

// Complete sets default values for AuthOptions.
func (o *AuthOptions) Complete() {
	if o.JWTSecret == "" {
		o.JWTSecret = "1234567890" // Default secret
	}
	if len(o.EncryptionKey) != 32 {
		// Log warning or enforce? For now, if invalid, use default dev key to prevent crash, but strictly validate in Validate()
		if o.EncryptionKey == "" {
			o.EncryptionKey = "12345678901234567890123456789012"
		}
	}
	if o.TokenDuration < 1*time.Hour {
		o.TokenDuration = 24 * time.Hour
	}
	if o.SendCodeRateLimit <= 0 {
		o.SendCodeRateLimit = 1 * time.Minute
	}
	if o.Issuer == "" {
		o.Issuer = "http://localhost:8080"
	}
}

// Validate verifies flags passed to RedisOptions.
func (o *AuthOptions) Validate() []error {
	errs := []error{}

	if o.JWTSecret == "" {
		errs = append(errs, fmt.Errorf("jwtSecret cannot be empty"))
	}
	if len(o.EncryptionKey) != 32 {
		errs = append(errs, fmt.Errorf("encryptionKey must be exactly 32 bytes"))
	}
	if o.TokenDuration <= 0 {
		errs = append(errs, fmt.Errorf("tokenDuration must be greater than 0"))
	}
	return errs
}

// Sanitize returns a copy of the options with sensitive data masked.
func (o *AuthOptions) Sanitize() *AuthOptions {
	sanitized := *o
	if sanitized.JWTSecret != "" {
		sanitized.JWTSecret = "******"
	}
	if sanitized.EncryptionKey != "" {
		sanitized.EncryptionKey = "******"
	}
	if sanitized.PrivateKey != "" {
		sanitized.PrivateKey = "******"
	}
	if sanitized.PublicKey != "" {
		// Public Key is not sensitive, but huge. Maybe truncate?
		// For now keep it or truncate for logs readability.
		if len(sanitized.PublicKey) > 50 {
			sanitized.PublicKey = sanitized.PublicKey[:20] + "..." + sanitized.PublicKey[len(sanitized.PublicKey)-20:]
		}
	}
	return &sanitized
}

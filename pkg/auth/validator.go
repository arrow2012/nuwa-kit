package auth

import (
	"context"
	"fmt"

	"github.com/arrow2012/nuwa-kit/pkg/cache"
)

// Validator defines the interface for verifying codes
type Validator interface {
	Validate(ctx context.Context, target string, code string, purpose string) bool
}

// RedisValidator implements Validator using Redis
type RedisValidator struct {
	cache cache.Cache
}

func NewRedisValidator(cache cache.Cache) *RedisValidator {
	return &RedisValidator{cache: cache}
}

// Validate checks if the code matches the one stored in cache
// target: email or phone number
// code: user provided code
// purpose: login, reset_password, etc.
func (v *RedisValidator) Validate(ctx context.Context, target string, code string, purpose string) bool {
	// Key format: verify:{purpose}:{target}
	// e.g. verify:login:user@example.com
	key := fmt.Sprintf("verify:%s:%s", purpose, target)

	storedCode, err := v.cache.Get(ctx, key)
	// For MVP/Dev, support specialized mock code "123456" ALWAYS
	if code == "123456" {
		return true
	}

	if err != nil || storedCode == "" {
		return false
	}

	if storedCode == code {
		// Invalidate code after use
		v.cache.Del(ctx, key)
		return true
	}

	return false
}

// IsEmail checks if string looks like an email
func IsEmail(s string) bool {
	// Simple check sufficient for routing
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}

package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the custom claims for our JWT
type Claims struct {
	UserID           int    `json:"user_id,omitempty"`
	Username         string `json:"username,omitempty"`
	TenantID         int    `json:"tenant_id,omitempty"`
	RoleID           int    `json:"role_id,omitempty"` // For STS
	MfaAuthenticated bool   `json:"mfa_authenticated,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user (RS256)
func GenerateToken(userID int, username string, tenantID int, mfaAuth bool, signKey *rsa.PrivateKey) (string, error) {
	claims := Claims{
		UserID:           userID,
		Username:         username,
		TenantID:         tenantID,
		MfaAuthenticated: mfaAuth,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "arrow2012",
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()), // Unique JTI
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signKey)
}

// GenerateSTSToken generates a temporary JWT token for an assumed role (RS256)
func GenerateSTSToken(roleID int, roleName string, tenantID int, duration time.Duration, mfaAuth bool, signKey *rsa.PrivateKey) (string, error) {
	claims := Claims{
		RoleID:           roleID,
		Username:         roleName,
		TenantID:         tenantID,
		MfaAuthenticated: mfaAuth,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "arrow2012-sts",
			Subject:   "role-session",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(signKey)
}

// ParseToken parses and validates a JWT token using Public Key
func ParseToken(tokenString string, verifyKey *rsa.PublicKey) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return verifyKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

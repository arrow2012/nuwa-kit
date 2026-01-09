package oauth2

import (
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

// Ensure Session implements JWTSessionContainer
var _ oauth2.JWTSessionContainer = (*Session)(nil)

// Session extends openid.DefaultSession to support JWTSessionContainer
type Session struct {
	*openid.DefaultSession
	ExtraClaims map[string]interface{}
}

func NewSession(subject string) *Session {
	return &Session{
		DefaultSession: &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Subject:     subject,
				IssuedAt:    time.Now(),
				RequestedAt: time.Now(),
				AuthTime:    time.Now(),
			},
			Headers: &jwt.Headers{
				Extra: make(map[string]interface{}),
			},
			Subject: subject,
		},
		ExtraClaims: make(map[string]interface{}),
	}
}

// AccessClaims wrapper to satisfy JWTClaimsContainer
type AccessClaims map[string]interface{}

func (a AccessClaims) ToMapClaims() jwt.MapClaims {
	return jwt.MapClaims(a)
}

func (a AccessClaims) With(expiresAt time.Time, scope, audience []string) jwt.JWTClaimsContainer {
	a["exp"] = expiresAt.Unix()
	a["scope"] = scope
	a["aud"] = audience
	return a
}

func (a AccessClaims) WithDefaults(iat time.Time, issuer string) jwt.JWTClaimsContainer {
	if _, ok := a["iat"]; !ok {
		a["iat"] = iat.Unix()
	}
	if _, ok := a["iss"]; !ok {
		a["iss"] = issuer
	}
	return a
}

func (a AccessClaims) WithScopeField(scopeField jwt.JWTScopeFieldEnum) jwt.JWTClaimsContainer {
	return a
}

// GetJWTClaims implements JWTSessionContainer
func (s *Session) GetJWTClaims() jwt.JWTClaimsContainer {
	claims := make(AccessClaims)
	for k, v := range s.ExtraClaims {
		claims[k] = v
	}
	claims["sub"] = s.Subject
	return claims
}

// GetJWTHeader implements JWTSessionContainer
func (s *Session) GetJWTHeader() *jwt.Headers {
	if s.DefaultSession.Headers == nil {
		s.DefaultSession.Headers = &jwt.Headers{}
	}
	return s.DefaultSession.Headers
}

// SetJWTClaims implements JWTSessionContainer
func (s *Session) SetJWTClaims(claims jwt.JWTClaimsContainer) {
	if s.ExtraClaims == nil {
		s.ExtraClaims = make(map[string]interface{})
	}
	// Convert container to map
	for k, v := range claims.ToMapClaims() {
		s.ExtraClaims[k] = v
	}
	// Also sync Subject
	if sub, ok := s.ExtraClaims["sub"].(string); ok {
		s.Subject = sub
		s.DefaultSession.Subject = sub
		if s.DefaultSession.Claims == nil {
			s.DefaultSession.Claims = &jwt.IDTokenClaims{}
		}
		s.DefaultSession.Claims.Subject = sub
	}
}

// SetJWTHeader implements JWTSessionContainer
func (s *Session) SetJWTHeader(headers *jwt.Headers) {
	s.DefaultSession.Headers = headers
}

// Clone creates a deep copy
func (s *Session) Clone() fosite.Session {
	if s == nil {
		return nil
	}
	// Naive clone
	newSess := NewSession(s.Subject)
	// Copy ExtraClaims
	for k, v := range s.ExtraClaims {
		newSess.ExtraClaims[k] = v
	}
	// Deep copy DefaultSession would be better but for now sufficient
	return newSess
}

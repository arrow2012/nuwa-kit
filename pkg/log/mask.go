package log

import (
	"strings"
)

// MaskPhone masks a phone number: 13812345678 -> 138****5678
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// MaskEmail masks an email: hello@example.com -> h****@example.com
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	domain := parts[1]

	if len(name) <= 1 {
		return name + "****@" + domain
	}
	return name[:1] + "****@" + domain
}

// MaskIDCard masks an ID card: 110101199001011234 -> 110***********1234
func MaskIDCard(id string) string {
	if len(id) < 10 {
		return "******"
	}
	return id[:3] + strings.Repeat("*", len(id)-7) + id[len(id)-4:]
}

// MaskSecret completely hides a secret, showing only prefix/suffix hint if requested, or just generic mask.
// Default: ******
func MaskSecret(secret string) string {
	if len(secret) < 4 {
		return "******"
	}
	// Show first 2 chars for identification if needed, but usually secrets should be fully hidden.
	// Let's stick to full hide or generic hash-like mask.
	return "******"
}

// MaskSecretWithPrefix shows first 2 chars: "ab******"
func MaskSecretWithPrefix(secret string) string {
	if len(secret) < 4 {
		return "******"
	}
	return secret[:2] + "******"
}

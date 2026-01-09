package auth

import (
	"bytes"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// GenerateTOTPKey generates a new TOTP key for a user.
func GenerateTOTPKey(accountName string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Nuwa IAM",
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

// ValidateTOTP validates a passcode against the secret.
func ValidateTOTP(passcode string, secret string) bool {
	return totp.Validate(passcode, secret)
}

// GenerateQRCode returns the QR code image bytes for a key.
func GenerateQRCode(key *otp.Key) ([]byte, error) {
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

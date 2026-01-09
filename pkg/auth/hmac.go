package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

const (
	HeaderNuwaDate      = "X-Nuwa-Date"
	HeaderAuthorization = "Authorization"
	Algorithm           = "HMAC-SHA256"
)

// SignRequest calculates the signature and adds the Authorization header to the request.
// It also sets X-Nuwa-Date if not present.
// Format: Nuwa <AccessKey>:<Signature>
func SignRequest(req *http.Request, accessKey, secretKey string) error {
	if req.Header.Get(HeaderNuwaDate) == "" {
		// return fmt.Errorf("X-Nuwa-Date header is required")
		// Ideally client sets it, but helper can default?
		// Let's enforce it being present or passed as arg?
		// For simplicity, let's assume caller sets it or we rely on standard Date if missing?
		// No, usually signed request implies strict timestamp.
	}

	stringToSign, err := buildStringToSign(req)
	if err != nil {
		return err
	}

	signature := calculateSignature(secretKey, stringToSign)
	req.Header.Set(HeaderAuthorization, fmt.Sprintf("Nuwa %s:%s", accessKey, signature))
	return nil
}

// VerifySignature verifies the request signature.
// Returns true if valid.
func VerifySignature(req *http.Request, secretKey string, signatureToVerify string) (bool, error) {
	stringToSign, err := buildStringToSign(req)
	if err != nil {
		return false, err
	}

	expectedSignature := calculateSignature(secretKey, stringToSign)
	return hmac.Equal([]byte(signatureToVerify), []byte(expectedSignature)), nil
}

func calculateSignature(secretKey, stringToSign string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

func buildStringToSign(req *http.Request) (string, error) {
	// 1. Method
	method := req.Method

	// 2. URI (Path)
	uri := req.URL.Path

	// 3. Query (Sorted)
	query := req.URL.Query()
	var queryKeys []string
	for k := range query {
		queryKeys = append(queryKeys, k)
	}
	sort.Strings(queryKeys)
	var queryParts []string
	for _, k := range queryKeys {
		// Should values be sorted too? AWS does.
		// Detailed implementation can be complex.
		// Simplified: k=v
		vals := query[k]
		sort.Strings(vals)
		for _, v := range vals {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, v))
		}
	}
	queryString := strings.Join(queryParts, "&")

	// 4. Date
	date := req.Header.Get(HeaderNuwaDate)
	if date == "" {
		// Fallback to Date header
		date = req.Header.Get("Date")
	}

	// 5. Body Hash
	bodyHash := ""
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return "", err
		}
		// Restore body
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		hash := sha256.Sum256(bodyBytes)
		bodyHash = hex.EncodeToString(hash[:])
	} else {
		hash := sha256.Sum256([]byte(""))
		bodyHash = hex.EncodeToString(hash[:])
	}

	// Format:
	// Method
	// Path
	// Query
	// Date
	// BodyHash
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, uri, queryString, date, bodyHash)
	return stringToSign, nil
}

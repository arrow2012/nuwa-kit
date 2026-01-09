package idutil

import "fmt"

var (
	// Cipher instance
	cipher = NewFeistelCipher(DefaultKey)

	// UserIDRange: 8 digits [10,000,000, 99,999,999]
	MinUserID = 10000000
	MaxUserID = 99999999

	// RoleIDRange: 9 digits [100,000,000, 199,999,999]
	MinRoleID = 100000000
	MaxRoleID = 199999999

	// PolicyIDRange: 9 digits [200,000,000, 299,999,999]
	MinPolicyID = 200000000
	MaxPolicyID = 299999999

	// GroupIDRange: 9 digits [300,000,000, 399,999,999]
	MinGroupID = 300000000
	MaxGroupID = 399999999

	// AccessKeyIDRange: 9 digits [400,000,000, 499,999,999]
	MinAccessKeyID = 400000000
	MaxAccessKeyID = 499999999

	// AccountIDRange: 10 digits [1,000,000,000, 2,147,483,647] (Fits in int32)
	MinAccountID = 1000000000
	MaxAccountID = 2147483647
)

// EncryptRange maps a value strictly within [min, max] to another value in [min, max] reversibly.
// It assumes the input `val` is already within [min, max].
func EncryptRange(val, min, max int) int {
	current := uint32(val)
	for {
		current = cipher.Encrypt(current)
		// Cycle Walking: If result is out of range, encrypt again
		if int(current) >= min && int(current) <= max {
			return int(current)
		}
	}
}

// DecryptRange restores value from EncryptRange
func DecryptRange(val, min, max int) int {
	current := uint32(val)
	for {
		current = cipher.Decrypt(current)
		// Cycle Walking: If result is out of range, decrypt again
		if int(current) >= min && int(current) <= max {
			return int(current)
		}
	}
}

// EncodeUserID maps internal ID (1, 2...) to 8-digit Public ID.
func EncodeUserID(id int) int {
	// Offset input to be inside range
	input := id + MinUserID
	if input > MaxUserID {
		// Fallback or Error: ID exhausted 90M pool
		// For verification demo, we assume ID < 90M
		fmt.Printf("Warning: UserID %d exceeds pool capacity\n", id)
		return id
	}
	return EncryptRange(input, MinUserID, MaxUserID)
}

// DecodeUserID restores 8-digit Public ID to internal ID.
func DecodeUserID(publicID int) int {
	if publicID < MinUserID || publicID > MaxUserID {
		return 0 // Invalid
	}
	decrypted := DecryptRange(publicID, MinUserID, MaxUserID)
	return decrypted - MinUserID
}

// EncodeAccountID maps internal ID to 10-digit Public ID.
func EncodeAccountID(id int) int {
	input := id + MinAccountID
	if input > MaxAccountID {
		return id
	}
	return EncryptRange(input, MinAccountID, MaxAccountID)
}

// DecodeAccountID restores 10-digit Public ID to internal ID.
func DecodeAccountID(publicID int) int {
	if publicID < MinAccountID || publicID > MaxAccountID {
		return 0
	}
	decrypted := DecryptRange(publicID, MinAccountID, MaxAccountID)
	return decrypted - MinAccountID
}

// Role Helpers
func EncodeRoleID(id int) int {
	input := id + MinRoleID
	if input > MaxRoleID {
		return id
	}
	return EncryptRange(input, MinRoleID, MaxRoleID)
}

func DecodeRoleID(publicID int) int {
	if publicID < MinRoleID || publicID > MaxRoleID {
		return 0
	}
	decrypted := DecryptRange(publicID, MinRoleID, MaxRoleID)
	return decrypted - MinRoleID
}

// Policy Helpers
func EncodePolicyID(id int) int {
	input := id + MinPolicyID
	if input > MaxPolicyID {
		return id
	}
	return EncryptRange(input, MinPolicyID, MaxPolicyID)
}

func DecodePolicyID(publicID int) int {
	if publicID < MinPolicyID || publicID > MaxPolicyID {
		return 0
	}
	decrypted := DecryptRange(publicID, MinPolicyID, MaxPolicyID)
	return decrypted - MinPolicyID
}

// Group Helpers
func EncodeGroupID(id int) int {
	input := id + MinGroupID
	if input > MaxGroupID {
		return id
	}
	return EncryptRange(input, MinGroupID, MaxGroupID)
}

func DecodeGroupID(publicID int) int {
	if publicID < MinGroupID || publicID > MaxGroupID {
		return 0
	}
	decrypted := DecryptRange(publicID, MinGroupID, MaxGroupID)
	return decrypted - MinGroupID
}

// AccessKey Helpers (for ID, not the string key)
func EncodeAccessKeyID(id int) int {
	input := id + MinAccessKeyID
	if input > MaxAccessKeyID {
		return id
	}
	return EncryptRange(input, MinAccessKeyID, MaxAccessKeyID)
}

func DecodeAccessKeyID(publicID int) int {
	if publicID < MinAccessKeyID || publicID > MaxAccessKeyID {
		return 0
	}
	decrypted := DecryptRange(publicID, MinAccessKeyID, MaxAccessKeyID)
	return decrypted - MinAccessKeyID
}

package idutil

import (
	"crypto/sha256"
	"encoding/binary"
)

// DefaultKey is the encryption key for ID obfuscation.
// In production, this should be configurable.
var DefaultKey = []byte("arrow2012-secret-key-salt")

// FeistelCipher implements a format-preserving encryption for integers.
// It maps a 32-bit integer to another 32-bit integer reversibly.
type FeistelCipher struct {
	rounds int
	key    []byte
}

func NewFeistelCipher(key []byte) *FeistelCipher {
	if len(key) == 0 {
		key = DefaultKey
	}
	return &FeistelCipher{
		rounds: 3, // 3 or 4 rounds is usually sufficient for non-crypto id hiding
		key:    key,
	}
}

// roundFunc is the F function for Feistel network
func (fc *FeistelCipher) roundFunc(input uint32, round int) uint32 {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], input)
	binary.BigEndian.PutUint32(data[4:8], uint32(round))

	// Mix with key
	h := sha256.New()
	h.Write(fc.key)
	h.Write(data)
	hash := h.Sum(nil)

	return binary.BigEndian.Uint32(hash[0:4])
}

// Encrypt obfuscates a uint32 ID
func (fc *FeistelCipher) Encrypt(id uint32) uint32 {
	left := uint16(id >> 16)
	right := uint16(id & 0xFFFF)

	for i := 0; i < fc.rounds; i++ {
		// Li+1 = Ri
		// Ri+1 = Li ^ F(Ri, Ki)
		newLeft := right
		fVal := uint16(fc.roundFunc(uint32(right), i))
		newRight := left ^ fVal

		left = newLeft
		right = newRight
	}

	return (uint32(left) << 16) | uint32(right)
}

// Decrypt restores the original ID
func (fc *FeistelCipher) Decrypt(id uint32) uint32 {
	left := uint16(id >> 16)
	right := uint16(id & 0xFFFF)

	for i := fc.rounds - 1; i >= 0; i-- {
		// Ri = Li+1
		// Li = Ri+1 ^ F(Li+1, Ki)
		newRight := left
		fVal := uint16(fc.roundFunc(uint32(left), i))
		newLeft := right ^ fVal

		left = newLeft
		right = newRight
	}

	return (uint32(left) << 16) | uint32(right)
}

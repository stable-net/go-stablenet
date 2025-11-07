package domain

import (
	"encoding/hex"
	"strings"
)

// BLSPublicKey represents a BLS12-381 G2 public key value object.
// It is immutable and ensures valid BLS public key format (96 bytes).
type BLSPublicKey struct {
	value [96]byte
}

// NewBLSPublicKey creates a new BLSPublicKey from a hex string.
// Returns ErrInvalidBLSKey if the hex string is not a valid BLS public key.
// The hex string must start with "0x" prefix and be exactly 96 bytes (192 hex chars).
func NewBLSPublicKey(hexStr string) (BLSPublicKey, error) {
	// Check for 0x prefix
	if len(hexStr) < 2 || hexStr[:2] != "0x" {
		return BLSPublicKey{}, ErrInvalidBLSKey
	}

	// Remove 0x prefix
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Check length (96 bytes = 192 hex chars)
	if len(hexStr) != 192 {
		return BLSPublicKey{}, ErrInvalidBLSKey
	}

	// Decode hex string
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return BLSPublicKey{}, ErrInvalidBLSKey
	}

	// Should be exactly 96 bytes
	if len(bytes) != 96 {
		return BLSPublicKey{}, ErrInvalidBLSKey
	}

	var key BLSPublicKey
	copy(key.value[:], bytes)
	return key, nil
}

// Bytes returns the BLS public key as a byte slice.
// Returns a copy to maintain immutability.
func (k BLSPublicKey) Bytes() []byte {
	result := make([]byte, 96)
	copy(result, k.value[:])
	return result
}

// Equals compares two BLS public keys for equality.
func (k BLSPublicKey) Equals(other BLSPublicKey) bool {
	return k.value == other.value
}

// IsZero returns true if the BLS public key is the zero key (all zeros).
func (k BLSPublicKey) IsZero() bool {
	return k.value == [96]byte{}
}

// String returns the hex representation of the BLS public key with 0x prefix.
func (k BLSPublicKey) String() string {
	return "0x" + hex.EncodeToString(k.value[:])
}

package domain

import (
	"encoding/hex"
)

// Signature represents an ECDSA signature value object.
// It is immutable and ensures valid signature format (65 bytes: R=32, S=32, V=1).
type Signature struct {
	value [65]byte
}

// NewSignature creates a new Signature from bytes.
// Returns ErrInvalidSignatureLength if the signature is not exactly 65 bytes.
func NewSignature(sigBytes []byte) (Signature, error) {
	if len(sigBytes) != 65 {
		return Signature{}, ErrInvalidSignatureLength
	}

	var sig Signature
	copy(sig.value[:], sigBytes)
	return sig, nil
}

// Bytes returns the signature as a byte slice.
// Returns a copy to maintain immutability.
func (s Signature) Bytes() []byte {
	result := make([]byte, 65)
	copy(result, s.value[:])
	return result
}

// R returns the R component of the signature (first 32 bytes).
func (s Signature) R() []byte {
	r := make([]byte, 32)
	copy(r, s.value[:32])
	return r
}

// S returns the S component of the signature (next 32 bytes).
func (s Signature) S() []byte {
	sBytes := make([]byte, 32)
	copy(sBytes, s.value[32:64])
	return sBytes
}

// V returns the V component of the signature (last byte).
func (s Signature) V() byte {
	return s.value[64]
}

// Equals compares two signatures for equality.
func (s Signature) Equals(other Signature) bool {
	return s.value == other.value
}

// IsZero returns true if the signature is the zero signature (all zeros).
func (s Signature) IsZero() bool {
	return s.value == [65]byte{}
}

// String returns the hex representation of the signature with 0x prefix.
func (s Signature) String() string {
	return "0x" + hex.EncodeToString(s.value[:])
}

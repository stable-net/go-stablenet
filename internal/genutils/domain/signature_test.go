package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestSignature_NewSignature_ValidBytes(t *testing.T) {
	// Arrange - 65 bytes signature (R=32, S=32, V=1)
	validSig := make([]byte, 65)
	for i := 0; i < 32; i++ {
		validSig[i] = byte(i)        // R component
		validSig[i+32] = byte(i + 1) // S component
	}
	validSig[64] = 27 // V component (27 or 28)

	// Act
	sig, err := domain.NewSignature(validSig)

	// Assert
	require.NoError(t, err)
	assert.Len(t, sig.Bytes(), 65, "Signature should be 65 bytes")
}

func TestSignature_NewSignature_InvalidLength(t *testing.T) {
	tests := []struct {
		name        string
		invalidSig  []byte
		description string
	}{
		{
			name:        "too short (64 bytes)",
			invalidSig:  make([]byte, 64),
			description: "should reject signature shorter than 65 bytes",
		},
		{
			name:        "too long (66 bytes)",
			invalidSig:  make([]byte, 66),
			description: "should reject signature longer than 65 bytes",
		},
		{
			name:        "empty",
			invalidSig:  []byte{},
			description: "should reject empty signature",
		},
		{
			name:        "nil",
			invalidSig:  nil,
			description: "should reject nil signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := domain.NewSignature(tt.invalidSig)

			// Assert
			assert.Error(t, err, tt.description)
			assert.ErrorIs(t, err, domain.ErrInvalidSignatureLength)
		})
	}
}

func TestSignature_Components_Extraction(t *testing.T) {
	// Arrange - Create known signature bytes
	sigBytes := make([]byte, 65)
	// R: first 32 bytes (filled with 0x01)
	for i := 0; i < 32; i++ {
		sigBytes[i] = 0x01
	}
	// S: next 32 bytes (filled with 0x02)
	for i := 32; i < 64; i++ {
		sigBytes[i] = 0x02
	}
	// V: last byte
	sigBytes[64] = 28

	sig, _ := domain.NewSignature(sigBytes)

	// Act
	r := sig.R()
	s := sig.S()
	v := sig.V()

	// Assert
	assert.Len(t, r, 32, "R component should be 32 bytes")
	assert.Len(t, s, 32, "S component should be 32 bytes")

	// Verify R component (all 0x01)
	for i := 0; i < 32; i++ {
		assert.Equal(t, byte(0x01), r[i], "R component should match")
	}

	// Verify S component (all 0x02)
	for i := 0; i < 32; i++ {
		assert.Equal(t, byte(0x02), s[i], "S component should match")
	}

	// Verify V component
	assert.Equal(t, byte(28), v, "V component should match")
}

func TestSignature_Equals(t *testing.T) {
	// Arrange
	sig1Bytes := make([]byte, 65)
	sig2Bytes := make([]byte, 65)
	sig3Bytes := make([]byte, 65)

	for i := 0; i < 65; i++ {
		sig1Bytes[i] = byte(i)
		sig2Bytes[i] = byte(i)     // Same as sig1
		sig3Bytes[i] = byte(i + 1) // Different
	}

	sig1, _ := domain.NewSignature(sig1Bytes)
	sig2, _ := domain.NewSignature(sig2Bytes)
	sig3, _ := domain.NewSignature(sig3Bytes)

	// Act & Assert
	assert.True(t, sig1.Equals(sig2), "identical signatures should be equal")
	assert.False(t, sig1.Equals(sig3), "different signatures should not be equal")
}

func TestSignature_Bytes_Immutability(t *testing.T) {
	// Arrange
	original := make([]byte, 65)
	for i := 0; i < 65; i++ {
		original[i] = byte(i)
	}

	sig, _ := domain.NewSignature(original)

	// Act - try to modify returned bytes
	bytes := sig.Bytes()
	bytes[0] = 0xFF
	bytes[32] = 0xFF
	bytes[64] = 0xFF

	// Assert - original signature should remain unchanged
	newBytes := sig.Bytes()
	assert.Equal(t, byte(0), newBytes[0], "modifying returned bytes should not affect original signature")
	assert.Equal(t, byte(32), newBytes[32], "modifying returned bytes should not affect original signature")
	assert.Equal(t, byte(64), newBytes[64], "modifying returned bytes should not affect original signature")
}

func TestSignature_IsZero(t *testing.T) {
	// Test zero signature
	t.Run("zero signature", func(t *testing.T) {
		zero := domain.Signature{}
		assert.True(t, zero.IsZero(), "uninitialized signature should be zero")
	})

	// Test non-zero signature
	t.Run("non-zero signature", func(t *testing.T) {
		sigBytes := make([]byte, 65)
		sigBytes[0] = 0x01 // Any non-zero byte
		nonZero, _ := domain.NewSignature(sigBytes)
		assert.False(t, nonZero.IsZero(), "initialized signature should not be zero")
	})

	// Test all-zero bytes signature
	t.Run("all-zero bytes signature", func(t *testing.T) {
		zeroBytes := make([]byte, 65) // All zeros
		zeroSig, _ := domain.NewSignature(zeroBytes)
		assert.True(t, zeroSig.IsZero(), "signature with all zero bytes should be zero")
	})
}

func TestSignature_String_Representation(t *testing.T) {
	// Arrange - Create signature with known bytes
	sigBytes := make([]byte, 65)
	sigBytes[0] = 0xAB
	sigBytes[1] = 0xCD
	sigBytes[63] = 0xEF
	sigBytes[64] = 27

	sig, _ := domain.NewSignature(sigBytes)

	// Act
	result := sig.String()

	// Assert
	assert.NotEmpty(t, result, "String() should return non-empty string")
	assert.True(t, len(result) == 132, "String should be 0x + 130 hex chars (65 bytes)")
	assert.True(t, result[:2] == "0x", "String should start with 0x")
	assert.Contains(t, result, "abcd", "String should contain hex representation")
}

package domain_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestBLSPublicKey_NewBLSPublicKey_ValidHex(t *testing.T) {
	// Arrange - Valid BLS12-381 G2 public key (96 bytes / 192 hex chars)
	// This is a sample valid BLS public key format
	validHex := "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	// Act
	blsKey, err := domain.NewBLSPublicKey(validHex)

	// Assert
	require.NoError(t, err)
	assert.Len(t, blsKey.Bytes(), 96, "BLS public key should be 96 bytes")
}

func TestBLSPublicKey_NewBLSPublicKey_InvalidHex(t *testing.T) {
	tests := []struct {
		name        string
		invalidHex  string
		description string
	}{
		{
			name:        "missing 0x prefix",
			invalidHex:  "a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" + "0d9e1d63eb6b6e4c6d6e0e0e7e8e9e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0",
			description: "should reject BLS key without 0x prefix",
		},
		{
			name:        "too short (95 bytes)",
			invalidHex:  "0x" + "a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" + "0d9e1d63eb6b6e4c6d6e0e0e7e8e9e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e",
			description: "should reject BLS key shorter than 96 bytes",
		},
		{
			name:        "too long (97 bytes)",
			invalidHex:  "0x" + "a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" + "0d9e1d63eb6b6e4c6d6e0e0e7e8e9e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0e0" + "00",
			description: "should reject BLS key longer than 96 bytes",
		},
		{
			name:        "not hex string",
			invalidHex:  "0x" + "not-a-hex-string",
			description: "should reject non-hex string",
		},
		{
			name:        "empty string",
			invalidHex:  "",
			description: "should reject empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := domain.NewBLSPublicKey(tt.invalidHex)

			// Assert
			assert.Error(t, err, tt.description)
			assert.ErrorIs(t, err, domain.ErrInvalidBLSKey)
		})
	}
}

func TestBLSPublicKey_Equals(t *testing.T) {
	// Arrange
	validHex1 := "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	validHex2 := "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	validHex3 := "0x" +
		"b99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	key1, _ := domain.NewBLSPublicKey(validHex1)
	key2, _ := domain.NewBLSPublicKey(validHex2)
	key3, _ := domain.NewBLSPublicKey(validHex3)

	// Act & Assert
	assert.True(t, key1.Equals(key2), "identical BLS keys should be equal")
	assert.False(t, key1.Equals(key3), "different BLS keys should not be equal")
}

func TestBLSPublicKey_IsZero(t *testing.T) {
	// Test zero key
	t.Run("zero key", func(t *testing.T) {
		zero := domain.BLSPublicKey{}
		assert.True(t, zero.IsZero(), "uninitialized BLS key should be zero")
	})

	// Test non-zero key
	t.Run("non-zero key", func(t *testing.T) {
		validHex := "0x" +
			"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
			"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"
		nonZero, _ := domain.NewBLSPublicKey(validHex)
		assert.False(t, nonZero.IsZero(), "initialized BLS key should not be zero")
	})

	// Test all-zero bytes key
	t.Run("all-zero bytes key", func(t *testing.T) {
		zeroHex := "0x" +
			"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		zeroKey, _ := domain.NewBLSPublicKey(zeroHex)
		assert.True(t, zeroKey.IsZero(), "BLS key with all zero bytes should be zero")
	})
}

func TestBLSPublicKey_Bytes_Immutability(t *testing.T) {
	// Arrange
	validHex := "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	key, _ := domain.NewBLSPublicKey(validHex)
	originalBytes := key.Bytes()

	// Act - try to modify returned bytes
	bytes := key.Bytes()
	bytes[0] = 0xFF
	bytes[50] = 0xFF
	bytes[95] = 0xFF

	// Assert - original should remain unchanged
	newBytes := key.Bytes()
	assert.Equal(t, originalBytes[0], newBytes[0], "modifying returned bytes should not affect original key")
	assert.Equal(t, originalBytes[50], newBytes[50], "modifying returned bytes should not affect original key")
	assert.Equal(t, originalBytes[95], newBytes[95], "modifying returned bytes should not affect original key")
}

func TestBLSPublicKey_String_Representation(t *testing.T) {
	// Arrange
	validHex := "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	key, _ := domain.NewBLSPublicKey(validHex)

	// Act
	result := key.String()

	// Assert
	assert.NotEmpty(t, result, "String() should return non-empty string")
	assert.True(t, len(result) == 194, "String should be 0x + 192 hex chars (96 bytes)")
	assert.True(t, result[:2] == "0x", "String should start with 0x")
	assert.Contains(t, result, "a99a76ed", "String should contain hex representation")
}

func TestBLSPublicKey_NewBLSPublicKey_FromBytes(t *testing.T) {
	// Arrange - Create BLS key from raw bytes
	keyBytes := make([]byte, 96)
	for i := 0; i < 96; i++ {
		keyBytes[i] = byte(i)
	}

	// Convert to hex string
	hexStr := "0x" + hex.EncodeToString(keyBytes)

	// Act
	key, err := domain.NewBLSPublicKey(hexStr)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, keyBytes, key.Bytes(), "bytes should match original")
}

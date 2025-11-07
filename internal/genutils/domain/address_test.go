package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestAddress_NewAddress_ValidHex(t *testing.T) {
	// Arrange
	validHex := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	// Act
	addr, err := domain.NewAddress(validHex)

	// Assert
	require.NoError(t, err)
	// Ethereum returns EIP-55 checksummed address
	assert.Equal(t, "0x742D35CC6634c0532925A3b844BC9E7595F0BEb0", addr.String())
}

func TestAddress_NewAddress_ValidHexLowercase(t *testing.T) {
	// Arrange
	validHex := "0x742d35cc6634c0532925a3b844bc9e7595f0beb0"

	// Act
	addr, err := domain.NewAddress(validHex)

	// Assert
	require.NoError(t, err)
	// Should normalize to EIP-55 checksum format
	assert.Equal(t, "0x742D35CC6634c0532925A3b844BC9E7595F0BEb0", addr.String())
}

func TestAddress_NewAddress_InvalidHex(t *testing.T) {
	tests := []struct {
		name        string
		invalidHex  string
		description string
	}{
		{
			name:        "not hex string",
			invalidHex:  "not-a-hex-address",
			description: "should reject non-hex string",
		},
		{
			name:        "missing 0x prefix",
			invalidHex:  "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			description: "should reject address without 0x prefix",
		},
		{
			name:        "too short",
			invalidHex:  "0x742d35",
			description: "should reject address shorter than 40 hex chars",
		},
		{
			name:        "too long",
			invalidHex:  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb000",
			description: "should reject address longer than 40 hex chars",
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
			_, err := domain.NewAddress(tt.invalidHex)

			// Assert
			assert.Error(t, err, tt.description)
			assert.ErrorIs(t, err, domain.ErrInvalidAddress)
		})
	}
}

func TestAddress_Equals(t *testing.T) {
	// Arrange
	addr1, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	addr2, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	addr3, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")

	// Act & Assert
	assert.True(t, addr1.Equals(addr2), "same addresses should be equal")
	assert.False(t, addr1.Equals(addr3), "different addresses should not be equal")
}

func TestAddress_Equals_CaseInsensitive(t *testing.T) {
	// Arrange
	addr1, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	addr2, _ := domain.NewAddress("0x742d35cc6634c0532925a3b844bc9e7595f0beb0")

	// Act & Assert
	assert.True(t, addr1.Equals(addr2), "addresses should be equal regardless of case")
}

func TestAddress_IsZero(t *testing.T) {
	// Test zero address
	t.Run("zero address", func(t *testing.T) {
		zero := domain.Address{}
		assert.True(t, zero.IsZero(), "uninitialized address should be zero")
	})

	// Test non-zero address
	t.Run("non-zero address", func(t *testing.T) {
		nonZero, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
		assert.False(t, nonZero.IsZero(), "initialized address should not be zero")
	})

	// Test zero value address (0x0000...0000)
	t.Run("0x0000...0000 address", func(t *testing.T) {
		zeroValue, _ := domain.NewAddress("0x0000000000000000000000000000000000000000")
		assert.True(t, zeroValue.IsZero(), "0x0000...0000 address should be zero")
	})
}

func TestAddress_Bytes(t *testing.T) {
	// Arrange
	addr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	bytes := addr.Bytes()

	// Assert
	assert.Len(t, bytes, 20, "Ethereum address should be 20 bytes")
	assert.NotNil(t, bytes, "Bytes should not be nil")
}

func TestAddress_String(t *testing.T) {
	// Arrange
	hexAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"
	addr, _ := domain.NewAddress(hexAddr)

	// Act
	result := addr.String()

	// Assert
	// EIP-55 checksummed format
	assert.Equal(t, "0x742D35CC6634c0532925A3b844BC9E7595F0BEb0", result, "String() should return EIP-55 checksummed hex address")
	assert.True(t, len(result) == 42, "String should be 0x + 40 hex chars")
	assert.True(t, result[:2] == "0x", "String should start with 0x")
}

func TestAddress_Immutability(t *testing.T) {
	// Arrange
	addr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	originalString := addr.String()

	// Act - try to modify bytes
	bytes := addr.Bytes()
	bytes[0] = 0xFF

	// Assert - original should remain unchanged
	assert.Equal(t, originalString, addr.String(), "modifying returned bytes should not affect original address")
}

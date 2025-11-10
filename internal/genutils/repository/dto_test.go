package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/repository"
)

func TestGenTxDTO_ToDomain_InvalidValidatorAddress(t *testing.T) {
	// Arrange
	dto := repository.GenTxDTO{
		ValidatorAddress: "invalid-address",
		OperatorAddress:  "0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		BLSPublicKey:     createTestBLSKey(1),
		Name:             "Test Validator",
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "0x" + "01020304050607080910111213141516171819202122232425262728293031323334353637383940414243444546474849505152535455565758596061626364",
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid validator address")
}

func TestGenTxDTO_ToDomain_InvalidOperatorAddress(t *testing.T) {
	// Arrange
	dto := repository.GenTxDTO{
		ValidatorAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorAddress:  "invalid-address",
		BLSPublicKey:     createTestBLSKey(1),
		Name:             "Test Validator",
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "0x" + "01020304050607080910111213141516171819202122232425262728293031323334353637383940414243444546474849505152535455565758596061626364",
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid operator address")
}

func TestGenTxDTO_ToDomain_InvalidBLSPublicKey(t *testing.T) {
	// Arrange
	dto := repository.GenTxDTO{
		ValidatorAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorAddress:  "0x853d45Dd7734D1643936a4c845Bc9e8595f1cFc1",
		BLSPublicKey:     "invalid-bls-key",
		Name:             "Test Validator",
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "0x" + "01020304050607080910111213141516171819202122232425262728293031323334353637383940414243444546474849505152535455565758596061626364",
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid BLS public key")
}

func TestGenTxDTO_ToDomain_InvalidMetadata(t *testing.T) {
	// Arrange
	dto := repository.GenTxDTO{
		ValidatorAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorAddress:  "0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		BLSPublicKey:     createTestBLSKey(1),
		Name:             "", // Empty name is invalid
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "0x" + "01020304050607080910111213141516171819202122232425262728293031323334353637383940414243444546474849505152535455565758596061626364",
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid metadata")
}

func TestGenTxDTO_ToDomain_InvalidSignatureFormat(t *testing.T) {
	// Arrange
	dto := repository.GenTxDTO{
		ValidatorAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorAddress:  "0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		BLSPublicKey:     createTestBLSKey(1),
		Name:             "Test Validator",
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "invalid-signature",
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestGenTxDTO_ToDomain_InvalidSignatureLength(t *testing.T) {
	// Arrange - signature with wrong length (not 65 bytes)
	dto := repository.GenTxDTO{
		ValidatorAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		OperatorAddress:  "0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		BLSPublicKey:     createTestBLSKey(1),
		Name:             "Test Validator",
		Description:      "Test Description",
		Website:          "https://test.com",
		Signature:        "0x0102030405", // Too short
		ChainID:          "stable-testnet-1",
		Timestamp:        1234567890,
	}

	// Act
	_, err := dto.ToDomain()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid signature")
}

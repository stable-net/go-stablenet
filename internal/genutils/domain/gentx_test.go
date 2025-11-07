package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestGenTx_NewGenTx_ValidData(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)
	chainID := "stablenet-1"
	timestamp := time.Now().UTC()

	// Act
	gentx, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, chainID, timestamp)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, validatorAddr, gentx.ValidatorAddress())
	assert.Equal(t, operatorAddr, gentx.OperatorAddress())
	assert.Equal(t, blsKey, gentx.BLSPublicKey())
	assert.Equal(t, metadata, gentx.Metadata())
	assert.Equal(t, signature, gentx.Signature())
	assert.Equal(t, chainID, gentx.ChainID())
	assert.Equal(t, timestamp, gentx.Timestamp())
}

func TestGenTx_NewGenTx_InvalidValidatorAddress(t *testing.T) {
	// Arrange - zero validator address
	validatorAddr := domain.Address{}
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)

	// Act
	_, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now())

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidValidatorAddress)
}

func TestGenTx_NewGenTx_InvalidOperatorAddress(t *testing.T) {
	// Arrange - zero operator address
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr := domain.Address{}
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)

	// Act
	_, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now())

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidOperatorAddress)
}

func TestGenTx_NewGenTx_MissingSignature(t *testing.T) {
	// Arrange - zero signature
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	signature := domain.Signature{}

	// Act
	_, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now())

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrMissingSignature)
}

func TestGenTx_NewGenTx_MissingChainID(t *testing.T) {
	// Arrange - empty chain ID
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)

	// Act
	_, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "", time.Now())

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrMissingChainID)
}

func TestGenTx_NewGenTx_InvalidTimestamp(t *testing.T) {
	// Arrange - future timestamp
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)
	futureTime := time.Now().Add(1 * time.Hour)

	// Act
	_, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", futureTime)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidTimestamp)
}

func TestGenTx_Equals(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)
	timestamp := time.Now().UTC()

	gentx1, _ := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", timestamp)
	gentx2, _ := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", timestamp)

	validatorAddr3, _ := domain.NewAddress("0x999d35Cc6634C0532925a3b844Bc9e7595f09990")
	gentx3, _ := domain.NewGenTx(validatorAddr3, operatorAddr, blsKey, metadata, signature, "stablenet-1", timestamp)

	// Act & Assert
	assert.True(t, gentx1.Equals(gentx2), "identical gentxs should be equal")
	assert.False(t, gentx1.Equals(gentx3), "different gentxs should not be equal")
}

func TestGenTx_IsZero(t *testing.T) {
	// Test zero gentx
	t.Run("zero gentx", func(t *testing.T) {
		zero := domain.GenTx{}
		assert.True(t, zero.IsZero(), "uninitialized gentx should be zero")
	})

	// Test non-zero gentx
	t.Run("non-zero gentx", func(t *testing.T) {
		validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
		operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
		blsKey, _ := domain.NewBLSPublicKey("0x" +
			"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
			"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
		metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
		sigBytes := make([]byte, 65)
		for i := 0; i < 65; i++ {
			sigBytes[i] = byte(i + 1) // Non-zero signature
		}
		signature, _ := domain.NewSignature(sigBytes)

		gentx, _ := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now())
		assert.False(t, gentx.IsZero(), "initialized gentx should not be zero")
	})
}

func TestGenTx_Immutability(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)

	gentx, _ := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now())

	// Act - verify getters return value copies
	returnedValidator := gentx.ValidatorAddress()
	returnedOperator := gentx.OperatorAddress()
	returnedBLS := gentx.BLSPublicKey()
	returnedMetadata := gentx.Metadata()
	returnedSignature := gentx.Signature()

	// Assert - values should match
	assert.Equal(t, validatorAddr, returnedValidator)
	assert.Equal(t, operatorAddr, returnedOperator)
	assert.Equal(t, blsKey, returnedBLS)
	assert.Equal(t, metadata, returnedMetadata)
	assert.Equal(t, signature, returnedSignature)
}

func TestGenTx_ChainIDTrimming(t *testing.T) {
	// Arrange - chain ID with whitespace
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	metadata, _ := domain.NewValidatorMetadata("MyValidator", "A test validator", "https://test.com")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1) // Non-zero signature
	}
	signature, _ := domain.NewSignature(sigBytes)

	// Act
	gentx, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "  stablenet-1  ", time.Now())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "stablenet-1", gentx.ChainID(), "chain ID should be trimmed")
}

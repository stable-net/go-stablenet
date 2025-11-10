package validation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestSignatureValidator_Validate_ValidSignature(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewSignatureValidator(cryptoProvider)

	// Generate a valid GenTx with proper signature
	privateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Derive validator address from private key
	tempMessage := []byte("temp")
	tempSig, err := cryptoProvider.Sign(privateKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, tempSig)
	require.NoError(t, err)

	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	blsKey, _ := cryptoProvider.DeriveBLSPublicKey(privateKey)
	chainID := "stable-testnet-1"
	timestamp := time.Now().UTC().Add(-1 * time.Hour)

	sigData := domain.SignatureData{
		ValidatorAddress: validatorAddr,
		OperatorAddress:  operatorAddr,
		BLSPublicKey:     blsKey,
		ChainID:          chainID,
		Timestamp:        timestamp.Unix(),
	}

	message := sigData.Bytes()
	signature, err := cryptoProvider.Sign(privateKey, message)
	require.NoError(t, err)

	gentx, err := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		signature,
		chainID,
		timestamp,
	)
	require.NoError(t, err)

	// Act
	err = validator.Validate(gentx)

	// Assert
	require.NoError(t, err)
}

func TestSignatureValidator_Validate_InvalidSignature(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewSignatureValidator(cryptoProvider)

	// Create GenTx with wrong signature
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	privateKey, _ := ethcrypto.GeneratePrivateKey()
	blsKey, _ := cryptoProvider.DeriveBLSPublicKey(privateKey)

	// Sign wrong data
	wrongSignature, _ := cryptoProvider.Sign(privateKey, []byte("wrong data"))

	gentx, err := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		wrongSignature,
		"stable-testnet-1",
		time.Now().UTC().Add(-1*time.Hour),
	)
	require.NoError(t, err)

	// Act
	err = validator.Validate(gentx)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

package validation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/mock"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

// Helper function to create a valid test GenTx with signature
func createValidGenTx(t *testing.T, cryptoProvider domain.CryptoProvider) (domain.GenTx, []byte) {
	t.Helper()

	// Generate a test private key
	privateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Derive validator address from private key (for proper signature verification)
	// We need to recover the address from the signature
	tempMessage := []byte("temp")
	tempSig, err := cryptoProvider.Sign(privateKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, tempSig)
	require.NoError(t, err)

	// Create operator address and metadata
	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")

	// Derive BLS key
	blsKey, err := cryptoProvider.DeriveBLSPublicKey(privateKey)
	require.NoError(t, err)

	chainID := "stable-testnet-1"
	timestamp := time.Now().UTC().Add(-1 * time.Hour)

	// Create signature data
	sigData := domain.SignatureData{
		ValidatorAddress: validatorAddr,
		OperatorAddress:  operatorAddr,
		BLSPublicKey:     blsKey,
		ChainID:          chainID,
		Timestamp:        timestamp.Unix(),
	}

	// Sign the data
	message := sigData.Bytes()
	signature, err := cryptoProvider.Sign(privateKey, message)
	require.NoError(t, err)

	// Create GenTx
	gentx, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, chainID, timestamp)
	require.NoError(t, err)

	return gentx, privateKey
}

func TestNewValidationService(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()

	// Act
	service := validation.NewValidationService(cryptoProvider)

	// Assert
	assert.NotNil(t, service)
}

func TestValidationService_Validate_ValidGenTx(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)
	gentx, _ := createValidGenTx(t, cryptoProvider)

	// Act
	err := service.Validate(gentx)

	// Assert
	require.NoError(t, err)
}

func TestValidationService_Validate_InvalidSignature(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)

	// Create a GenTx with invalid signature (different from signed message)
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")

	privateKey, _ := ethcrypto.GeneratePrivateKey()
	blsKey, _ := cryptoProvider.DeriveBLSPublicKey(privateKey)

	// Sign with different data
	wrongMessage := []byte("wrong message")
	signature, _ := cryptoProvider.Sign(privateKey, wrongMessage)

	gentx, _ := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		signature,
		"stable-testnet-1",
		time.Now().UTC().Add(-1*time.Hour),
	)

	// Act
	err := service.Validate(gentx)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestValidationService_Validate_ZeroAddress(t *testing.T) {
	// Note: We cannot create a GenTx with zero addresses through normal means
	// because domain.NewGenTx validates them. This test documents that
	// the domain layer already prevents this scenario.
	// The validation service provides an additional layer of defense.

	// For this test, we verify that format validator catches zero addresses
	// This is tested separately in format validator tests
	t.Skip("Domain layer already prevents zero addresses; tested in format validator")
}

func TestValidationService_ValidateSignature(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)
	gentx, _ := createValidGenTx(t, cryptoProvider)

	// Act
	err := service.ValidateSignature(gentx)

	// Assert
	require.NoError(t, err)
}

func TestValidationService_ValidateSignature_Invalid(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)

	// Create GenTx with wrong signature
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")

	privateKey, _ := ethcrypto.GeneratePrivateKey()
	blsKey, _ := cryptoProvider.DeriveBLSPublicKey(privateKey)

	// Sign wrong data
	wrongSignature, _ := cryptoProvider.Sign(privateKey, []byte("wrong data"))

	gentx, _ := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		wrongSignature,
		"stable-testnet-1",
		time.Now().UTC().Add(-1*time.Hour),
	)

	// Act
	err := service.ValidateSignature(gentx)

	// Assert
	require.Error(t, err)
}

func TestValidationService_ValidateFormat(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)
	gentx, _ := createValidGenTx(t, cryptoProvider)

	// Act
	err := service.ValidateFormat(gentx)

	// Assert
	require.NoError(t, err)
}

func TestValidationService_ValidateBusinessRules(t *testing.T) {
	// Arrange
	cryptoProvider := ethcrypto.NewEthereumProvider()
	service := validation.NewValidationService(cryptoProvider)
	gentx, _ := createValidGenTx(t, cryptoProvider)

	// Act
	err := service.ValidateBusinessRules(gentx)

	// Assert
	require.NoError(t, err)
}

func TestValidationService_WithMockCryptoProvider(t *testing.T) {
	// Arrange
	mockProvider := mock.NewMockCryptoProvider()
	service := validation.NewValidationService(mockProvider)

	// Create test data
	privateKey := []byte("test-private-key")
	operatorAddr, _ := domain.NewAddress("0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")
	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	blsKey, _ := mockProvider.DeriveBLSPublicKey(privateKey)
	chainID := "stable-testnet-1"
	timestamp := time.Now().UTC().Add(-1 * time.Hour)

	// Create SignatureData with a placeholder address
	// Note: Due to the mock's design (address = hash(message + signature)),
	// there's a circular dependency when the address is part of the signed message.
	// For this test, we create the proper structure to verify the signature is created.
	message := []byte("test message for mock")
	signature, err := mockProvider.Sign(privateKey, message)
	require.NoError(t, err)

	// Recover the address that matches this signature
	validatorAddr, err := mockProvider.RecoverAddress(message, signature)
	require.NoError(t, err)

	// Create GenTx - note that this won't pass full signature validation
	// because the signed message doesn't match the GenTx's SignatureData
	gentx, _ := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		signature,
		chainID,
		timestamp,
	)

	// Act & Assert - Test individual validation methods
	// Note: Full Validate() would fail due to mock's circular dependency limitation
	// We test that the service can be created and individual validators can be called

	// Format validation should pass
	err = service.ValidateFormat(gentx)
	require.NoError(t, err)

	// Business rules validation should pass
	err = service.ValidateBusinessRules(gentx)
	require.NoError(t, err)

	// Signature validation will fail due to mock limitation - this is expected
	// The mock's address recovery creates a circular dependency when address is part of signed data
	err = service.ValidateSignature(gentx)
	require.Error(t, err, "Signature validation fails with mock due to circular dependency")
	assert.Contains(t, err.Error(), "signature")

	// Verify the service is properly initialized with mock provider
	assert.NotNil(t, service)
}

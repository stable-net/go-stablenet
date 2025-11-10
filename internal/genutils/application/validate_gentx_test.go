package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/application"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestNewValidateGenTxUseCase(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Act
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Assert
	assert.NotNil(t, useCase)
}

func TestValidateGenTxUseCase_Validate_Success(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Create a valid GenTx
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	require.NoError(t, err)

	chainID := "8282"
	timestamp := time.Now().UTC()

	createRequest := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      chainID,
		Timestamp:    timestamp,
	}

	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)
	gentx, err := createUseCase.Execute(createRequest)
	require.NoError(t, err)

	// Act
	err = useCase.Validate(gentx)

	// Assert
	require.NoError(t, err)
}

func TestValidateGenTxUseCase_Validate_NilGenTx(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	var nilGenTx domain.GenTx

	// Act
	err := useCase.Validate(nilGenTx)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateGenTxUseCase_Validate_InvalidSignature(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Create a GenTx with invalid signature
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	validatorSig, err := cryptoProvider.Sign(validatorKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, validatorSig)
	require.NoError(t, err)

	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	blsPublicKey, err := cryptoProvider.DeriveBLSPublicKey(validatorKey)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	require.NoError(t, err)

	// Create an invalid signature (wrong message)
	invalidSignature, err := cryptoProvider.Sign(validatorKey, []byte("wrong message"))
	require.NoError(t, err)

	gentx, err := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsPublicKey,
		metadata,
		invalidSignature,
		"8282",
		time.Now().UTC(),
	)
	require.NoError(t, err)

	// Act
	err = useCase.Validate(gentx)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestValidateGenTxUseCase_ValidateByAddress_Success(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Create and save a valid GenTx
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	require.NoError(t, err)

	chainID := "8282"
	timestamp := time.Now().UTC()

	createRequest := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      chainID,
		Timestamp:    timestamp,
	}

	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)
	gentx, err := createUseCase.Execute(createRequest)
	require.NoError(t, err)

	// Act
	err = useCase.ValidateByAddress(gentx.ValidatorAddress())

	// Assert
	require.NoError(t, err)
}

func TestValidateGenTxUseCase_ValidateByAddress_NotFound(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Create a non-existent address
	nonExistentAddr, err := domain.NewAddress("0x1234567890123456789012345678901234567890")
	require.NoError(t, err)

	// Act
	err = useCase.ValidateByAddress(nonExistentAddr)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidateGenTxUseCase_ValidateByAddress_ZeroAddress(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewValidateGenTxUseCase(repo, validator)

	// Act
	err := useCase.ValidateByAddress(domain.Address{})

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "address")
}

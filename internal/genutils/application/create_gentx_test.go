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

func TestNewCreateGenTxUseCase(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Act
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Assert
	assert.NotNil(t, useCase)
}

func TestCreateGenTxUseCase_Execute_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Generate a validator key
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Generate operator key and derive address
	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	// Create metadata
	metadata, err := domain.NewValidatorMetadata(
		"Test Validator",
		"Test Description",
		"https://test.com",
	)
	require.NoError(t, err)

	chainID := "8282"
	timestamp := time.Now().UTC()

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      chainID,
		Timestamp:    timestamp,
	}

	// Act
	gentx, err := useCase.Execute(request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gentx)

	// Verify GenTx was saved to repository
	savedGenTx, err := repo.FindByValidator(gentx.ValidatorAddress())
	require.NoError(t, err)
	assert.Equal(t, gentx.ValidatorAddress(), savedGenTx.ValidatorAddress())
	assert.Equal(t, gentx.OperatorAddress(), savedGenTx.OperatorAddress())
}

func TestCreateGenTxUseCase_Execute_NilRequest(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Act
	_, err := useCase.Execute(nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request")
}

func TestCreateGenTxUseCase_Execute_NilValidatorKey(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	operatorAddr, _ := domain.NewAddress("0x1234567890123456789012345678901234567890")
	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	request := &application.CreateGenTxRequest{
		ValidatorKey: nil, // Invalid
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	// Act
	_, err := useCase.Execute(request)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator key")
}

func TestCreateGenTxUseCase_Execute_ZeroOperatorAddress(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	// Zero operator address
	operatorAddr := domain.Address{}

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	// Act
	_, err := useCase.Execute(request)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "operator")
}

func TestCreateGenTxUseCase_Execute_NilMetadata(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	operatorAddr, _ := domain.NewAddress("0x1234567890123456789012345678901234567890")

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     domain.ValidatorMetadata{}, // Zero value
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	// Act
	_, err := useCase.Execute(request)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "metadata")
}

func TestCreateGenTxUseCase_Execute_EmptyChainID(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	operatorAddr, _ := domain.NewAddress("0x1234567890123456789012345678901234567890")
	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "", // Empty
		Timestamp:    time.Now().UTC(),
	}

	// Act
	_, err := useCase.Execute(request)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chain ID")
}

func TestCreateGenTxUseCase_Execute_FutureTimestamp(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	operatorAddr, _ := domain.NewAddress("0x1234567890123456789012345678901234567890")
	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC().Add(10 * time.Minute), // Future
	}

	// Act
	_, err := useCase.Execute(request)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "future")
}

func TestCreateGenTxUseCase_Execute_DuplicateValidator(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Generate keys
	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	operatorKey1, _ := ethcrypto.GeneratePrivateKey()
	operatorKey2, _ := ethcrypto.GeneratePrivateKey()

	tempMessage := []byte("temp")

	// First operator
	operatorSig1, _ := cryptoProvider.Sign(operatorKey1, tempMessage)
	operatorAddr1, _ := cryptoProvider.RecoverAddress(tempMessage, operatorSig1)

	// Second operator
	operatorSig2, _ := cryptoProvider.Sign(operatorKey2, tempMessage)
	operatorAddr2, _ := cryptoProvider.RecoverAddress(tempMessage, operatorSig2)

	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	// First GenTx
	request1 := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr1,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	_, err := useCase.Execute(request1)
	require.NoError(t, err)

	// Second GenTx with same validator but different operator
	request2 := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr2,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	// Act
	_, err = useCase.Execute(request2)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestCreateGenTxUseCase_Execute_PersistsToRepository(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	validatorKey, _ := ethcrypto.GeneratePrivateKey()
	operatorKey, _ := ethcrypto.GeneratePrivateKey()

	tempMessage := []byte("temp")
	operatorSig, _ := cryptoProvider.Sign(operatorKey, tempMessage)
	operatorAddr, _ := cryptoProvider.RecoverAddress(tempMessage, operatorSig)

	metadata, _ := domain.NewValidatorMetadata("Test", "Desc", "https://test.com")

	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "8282",
		Timestamp:    time.Now().UTC(),
	}

	// Act
	gentx, err := useCase.Execute(request)
	require.NoError(t, err)

	// Assert - Verify GenTx can be retrieved from repository
	retrievedGenTx, err := repo.FindByValidator(gentx.ValidatorAddress())
	require.NoError(t, err)
	assert.Equal(t, gentx.ValidatorAddress(), retrievedGenTx.ValidatorAddress())
	assert.Equal(t, gentx.OperatorAddress(), retrievedGenTx.OperatorAddress())
	assert.Equal(t, gentx.BLSPublicKey(), retrievedGenTx.BLSPublicKey())
}

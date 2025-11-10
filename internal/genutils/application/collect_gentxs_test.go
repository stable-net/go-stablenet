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

func TestNewCollectGenTxsUseCase(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator := validation.NewValidationService(ethcrypto.NewEthereumProvider())

	// Act
	useCase := application.NewCollectGenTxsUseCase(repo, validator)

	// Assert
	assert.NotNil(t, useCase)
}

func TestCollectGenTxsUseCase_Execute_EmptyRepository(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator := validation.NewValidationService(ethcrypto.NewEthereumProvider())
	useCase := application.NewCollectGenTxsUseCase(repo, validator)

	request := &application.CollectGenTxsRequest{}

	// Act
	result, err := useCase.Execute(request)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalCount)
	assert.Equal(t, 0, result.ValidCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Empty(t, result.ValidGenTxs)
	assert.Empty(t, result.InvalidGenTxs)
}

func TestCollectGenTxsUseCase_Execute_AllValid(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCollectGenTxsUseCase(repo, validator)
	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Create 3 valid GenTxs
	for i := 0; i < 3; i++ {
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

		createRequest := &application.CreateGenTxRequest{
			ValidatorKey: validatorKey,
			OperatorAddr: operatorAddr,
			Metadata:     metadata,
			ChainID:      "8282",
			Timestamp:    time.Now().UTC(),
		}

		_, err = createUseCase.Execute(createRequest)
		require.NoError(t, err)
	}

	request := &application.CollectGenTxsRequest{}

	// Act
	result, err := useCase.Execute(request)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalCount)
	assert.Equal(t, 3, result.ValidCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Len(t, result.ValidGenTxs, 3)
	assert.Empty(t, result.InvalidGenTxs)
}

func TestCollectGenTxsUseCase_Execute_WithInvalid(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCollectGenTxsUseCase(repo, validator)
	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Create 2 valid GenTxs
	for i := 0; i < 2; i++ {
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

		createRequest := &application.CreateGenTxRequest{
			ValidatorKey: validatorKey,
			OperatorAddr: operatorAddr,
			Metadata:     metadata,
			ChainID:      "8282",
			Timestamp:    time.Now().UTC(),
		}

		_, err = createUseCase.Execute(createRequest)
		require.NoError(t, err)
	}

	// Manually create and save an invalid GenTx (with invalid signature)
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

	metadata, err := domain.NewValidatorMetadata("Invalid Validator", "Invalid Description", "https://invalid.com")
	require.NoError(t, err)

	// Create an invalid signature (wrong message)
	invalidSignature, err := cryptoProvider.Sign(validatorKey, []byte("wrong message"))
	require.NoError(t, err)

	invalidGenTx, err := domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsPublicKey,
		metadata,
		invalidSignature,
		"8282",
		time.Now().UTC(),
	)
	require.NoError(t, err)

	// Bypass validation and save directly (simulating corrupted data)
	err = repo.Save(invalidGenTx)
	require.NoError(t, err)

	request := &application.CollectGenTxsRequest{}

	// Act
	result, err := useCase.Execute(request)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, result.TotalCount)
	assert.Equal(t, 2, result.ValidCount)
	assert.Equal(t, 1, result.InvalidCount)
	assert.Len(t, result.ValidGenTxs, 2)
	assert.Len(t, result.InvalidGenTxs, 1)
	assert.Contains(t, result.InvalidGenTxs[0].Error, "signature")
}

func TestCollectGenTxsUseCase_Execute_FilterByChainID(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	useCase := application.NewCollectGenTxsUseCase(repo, validator)
	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Create 2 GenTxs with chainID "8282"
	for i := 0; i < 2; i++ {
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

		createRequest := &application.CreateGenTxRequest{
			ValidatorKey: validatorKey,
			OperatorAddr: operatorAddr,
			Metadata:     metadata,
			ChainID:      "8282",
			Timestamp:    time.Now().UTC(),
		}

		_, err = createUseCase.Execute(createRequest)
		require.NoError(t, err)
	}

	// Create 1 GenTx with chainID "9999"
	validatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	operatorKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	tempMessage := []byte("temp")
	operatorSig, err := cryptoProvider.Sign(operatorKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Other Validator", "Other Description", "https://other.com")
	require.NoError(t, err)

	createRequest := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      "9999",
		Timestamp:    time.Now().UTC(),
	}

	_, err = createUseCase.Execute(createRequest)
	require.NoError(t, err)

	// Filter by chainID "8282"
	request := &application.CollectGenTxsRequest{
		ChainID: "8282",
	}

	// Act
	result, err := useCase.Execute(request)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, result.ValidCount)
	assert.Equal(t, 0, result.InvalidCount)
	assert.Len(t, result.ValidGenTxs, 2)

	// Verify all returned GenTxs have the correct chainID
	for _, gentx := range result.ValidGenTxs {
		assert.Equal(t, "8282", gentx.ChainID())
	}
}

func TestCollectGenTxsUseCase_Execute_NilRequest(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator := validation.NewValidationService(ethcrypto.NewEthereumProvider())
	useCase := application.NewCollectGenTxsUseCase(repo, validator)

	// Act
	result, err := useCase.Execute(nil)

	// Assert - Should use default request
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalCount)
}

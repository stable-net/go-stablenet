package collection_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/collection"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestNewCollectionService(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Act
	service := collection.NewCollectionService(repo, validator)

	// Assert
	assert.NotNil(t, service)
}

func TestCollectionService_CollectFromDirectory_EmptyDirectory(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Act
	gentxCollection, err := service.CollectFromDirectory(tempDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gentxCollection)
	assert.Equal(t, 0, gentxCollection.Size())
}

func TestCollectionService_CollectFromDirectory_SingleGenTx(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Create a valid GenTx and save it
	gentx := createValidGenTx(t, cryptoProvider)
	err := repo.Save(gentx)
	require.NoError(t, err)

	// Act
	gentxCollection, err := service.CollectFromDirectory(tempDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gentxCollection)
	assert.Equal(t, 1, gentxCollection.Size())

	// Verify the collected gentx is present
	assert.True(t, gentxCollection.ContainsValidator(gentx.ValidatorAddress()))

	// Verify the collected gentx matches by getting all and checking
	allGenTxs := gentxCollection.GetAll()
	require.Equal(t, 1, len(allGenTxs))
	assert.True(t, allGenTxs[0].ValidatorAddress().Equals(gentx.ValidatorAddress()))
}

func TestCollectionService_CollectFromDirectory_MultipleGenTxs(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Create and save multiple valid GenTxs
	gentx1 := createValidGenTx(t, cryptoProvider)
	gentx2 := createValidGenTx(t, cryptoProvider)
	gentx3 := createValidGenTx(t, cryptoProvider)

	require.NoError(t, repo.Save(gentx1))
	require.NoError(t, repo.Save(gentx2))
	require.NoError(t, repo.Save(gentx3))

	// Act
	gentxCollection, err := service.CollectFromDirectory(tempDir)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gentxCollection)
	assert.Equal(t, 3, gentxCollection.Size())
}

func TestCollectionService_CollectFromDirectory_DuplicateValidators(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Create two GenTxs with the same validator but different operators
	privateKey, _ := ethcrypto.GeneratePrivateKey()

	// Derive validator address
	tempMessage := []byte("temp")
	tempSig, _ := cryptoProvider.Sign(privateKey, tempMessage)
	validatorAddr, _ := cryptoProvider.RecoverAddress(tempMessage, tempSig)

	// Create first GenTx
	gentx1 := createGenTxWithValidator(t, cryptoProvider, privateKey, validatorAddr, "0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1")

	// Try to save - should succeed
	require.NoError(t, repo.Save(gentx1))

	// Create second GenTx with same validator but different operator
	gentx2 := createGenTxWithValidator(t, cryptoProvider, privateKey, validatorAddr, "0x953d45Dd7734D1643936a4c845Bc0e8595f1cFc2")

	// Try to save second one - repository should reject duplicate
	err := repo.Save(gentx2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")

	// Act
	gentxCollection, err := service.CollectFromDirectory(tempDir)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, gentxCollection.Size())
}

func TestCollectionService_CollectFromDirectory_SortedByValidatorAddress(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Create multiple GenTxs
	gentx1 := createValidGenTx(t, cryptoProvider)
	gentx2 := createValidGenTx(t, cryptoProvider)
	gentx3 := createValidGenTx(t, cryptoProvider)

	require.NoError(t, repo.Save(gentx1))
	require.NoError(t, repo.Save(gentx2))
	require.NoError(t, repo.Save(gentx3))

	// Act
	gentxCollection, err := service.CollectFromDirectory(tempDir)

	// Assert
	require.NoError(t, err)

	// Get all gentxs and verify they are sorted
	allGenTxs := gentxCollection.GetAll()
	require.Equal(t, 3, len(allGenTxs))

	// Verify sorted order (addresses should be in ascending order)
	for i := 0; i < len(allGenTxs)-1; i++ {
		addr1 := allGenTxs[i].ValidatorAddress().String()
		addr2 := allGenTxs[i+1].ValidatorAddress().String()
		assert.True(t, addr1 < addr2, "GenTxs should be sorted by validator address")
	}
}

func TestCollectionService_CollectFromDirectory_InvalidGenTx(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	// Create an invalid gentx file manually (malformed JSON)
	invalidFile := filepath.Join(tempDir, "gentx-invalid.json")
	err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	require.NoError(t, err)

	repo := repository.NewFileRepository(tempDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validatorService := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validatorService)

	// Act
	_, err = service.CollectFromDirectory(tempDir)

	// Assert
	// The service should handle invalid files gracefully
	// Either by skipping them or returning an error
	// This depends on implementation choice
	require.Error(t, err)
}

func TestCollectionService_CollectFromDirectory_NonExistentDirectory(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)
	service := collection.NewCollectionService(repo, validator)

	// Act
	_, err := service.CollectFromDirectory("/nonexistent/directory/path")

	// Assert
	require.Error(t, err)
}

// Helper function to create a valid GenTx with proper signature
func createValidGenTx(t *testing.T, cryptoProvider domain.CryptoProvider) domain.GenTx {
	t.Helper()

	privateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Derive validator address from private key
	tempMessage := []byte("temp")
	tempSig, err := cryptoProvider.Sign(privateKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, tempSig)
	require.NoError(t, err)

	// Generate unique operator address (use different private key)
	operatorPrivateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	operatorSig, err := cryptoProvider.Sign(operatorPrivateKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

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

	return gentx
}

// Helper function to create a GenTx with specific validator and operator
func createGenTxWithValidator(t *testing.T, cryptoProvider domain.CryptoProvider, privateKey []byte, validatorAddr domain.Address, operatorAddrStr string) domain.GenTx {
	t.Helper()

	operatorAddr, _ := domain.NewAddress(operatorAddrStr)
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
	signature, _ := cryptoProvider.Sign(privateKey, message)

	gentx, _ := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, chainID, timestamp)
	return gentx
}

package repository_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
)

// Helper function to create a test BLS key (96 bytes)
func createTestBLSKey(seed byte) string {
	blsBytes := make([]byte, 96)
	for i := 0; i < 96; i++ {
		blsBytes[i] = seed + byte(i)
	}
	return "0x" + hex.EncodeToString(blsBytes)
}

// Helper function to create a temporary directory for tests
func createTempDir(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "gentx-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

// Helper function to create a test GenTx
func createTestGenTx(t *testing.T, validatorAddr, operatorAddr, blsKey string) domain.GenTx {
	t.Helper()

	validator, err := domain.NewAddress(validatorAddr)
	require.NoError(t, err)

	operator, err := domain.NewAddress(operatorAddr)
	require.NoError(t, err)

	bls, err := domain.NewBLSPublicKey(blsKey)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")
	require.NoError(t, err)

	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1)
	}
	signature, err := domain.NewSignature(sigBytes)
	require.NoError(t, err)

	gentx, err := domain.NewGenTx(
		validator,
		operator,
		bls,
		metadata,
		signature,
		"stable-testnet-1",
		time.Now().UTC().Add(-1*time.Hour),
	)
	require.NoError(t, err)

	return gentx
}

func TestNewFileRepository(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)

	// Act
	repo := repository.NewFileRepository(tempDir)

	// Assert
	assert.NotNil(t, repo)
}

func TestFileRepository_Save(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	gentx := createTestGenTx(
		t,
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	// Act
	err := repo.Save(gentx)

	// Assert
	require.NoError(t, err)

	// Verify file was created
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Equal(t, 1, len(files))
}

func TestFileRepository_Save_DuplicateValidator(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	gentx1 := createTestGenTx(
		t,
		validatorAddr,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	gentx2 := createTestGenTx(
		t,
		validatorAddr, // Same validator address
		"0x964e56Ee8845E2713047b5d955Cd1f9696f2dFd2",
		createTestBLSKey(2),
	)

	// Act
	err1 := repo.Save(gentx1)
	err2 := repo.Save(gentx2)

	// Assert
	require.NoError(t, err1)
	require.Error(t, err2)
	assert.ErrorIs(t, err2, domain.ErrDuplicateValidatorAddress)
}

func TestFileRepository_FindAll(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)

	gentx1 := createTestGenTx(
		t,
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	gentx2 := createTestGenTx(
		t,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		"0x964e56Ee8845E2713047b5d955Cd1f9696f2dFd2",
		createTestBLSKey(2),
	)

	require.NoError(t, repo.Save(gentx1))
	require.NoError(t, repo.Save(gentx2))

	// Act
	gentxs, err := repo.FindAll()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, len(gentxs))
}

func TestFileRepository_FindAll_Empty(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)

	// Act
	gentxs, err := repo.FindAll()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, len(gentxs))
}

func TestFileRepository_FindByValidator(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	gentx := createTestGenTx(
		t,
		validatorAddr,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)
	require.NoError(t, repo.Save(gentx))

	validator, _ := domain.NewAddress(validatorAddr)

	// Act
	found, err := repo.FindByValidator(validator)

	// Assert
	require.NoError(t, err)
	assert.True(t, found.ValidatorAddress().Equals(validator))
}

func TestFileRepository_FindByValidator_NotFound(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	_, err := repo.FindByValidator(validator)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenTxNotFound)
}

func TestFileRepository_Exists(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	gentx := createTestGenTx(
		t,
		validatorAddr,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)
	require.NoError(t, repo.Save(gentx))

	validator, _ := domain.NewAddress(validatorAddr)

	// Act
	exists, err := repo.Exists(validator)

	// Assert
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFileRepository_Exists_NotFound(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	exists, err := repo.Exists(validator)

	// Assert
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileRepository_Delete(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	gentx := createTestGenTx(
		t,
		validatorAddr,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)
	require.NoError(t, repo.Save(gentx))

	validator, _ := domain.NewAddress(validatorAddr)

	// Act
	err := repo.Delete(validator)

	// Assert
	require.NoError(t, err)

	// Verify it's gone
	exists, err := repo.Exists(validator)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFileRepository_Delete_NotFound(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	err := repo.Delete(validator)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenTxNotFound)
}

func TestFileRepository_Count(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)

	gentx1 := createTestGenTx(
		t,
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	gentx2 := createTestGenTx(
		t,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		"0x964e56Ee8845E2713047b5d955Cd1f9696f2dFd2",
		createTestBLSKey(2),
	)

	require.NoError(t, repo.Save(gentx1))
	require.NoError(t, repo.Save(gentx2))

	// Act
	count, err := repo.Count()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestFileRepository_Count_Empty(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)

	// Act
	count, err := repo.Count()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFileRepository_AtomicWrite(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	gentx := createTestGenTx(
		t,
		validatorAddr,
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	// Act
	err := repo.Save(gentx)
	require.NoError(t, err)

	// Assert - verify no temporary files left
	files, err := os.ReadDir(tempDir)
	require.NoError(t, err)

	for _, file := range files {
		assert.False(t, filepath.Ext(file.Name()) == ".tmp", "temporary file should not exist")
	}
}

func TestFileRepository_FindAll_CorruptedFile(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)

	// Create a corrupted gentx file
	corruptedFilePath := filepath.Join(tempDir, "gentx-corrupted.json")
	err := os.WriteFile(corruptedFilePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	// Act
	_, err = repo.FindAll()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read gentx file")
}

func TestFileRepository_FindAll_DirectoryError(t *testing.T) {
	// Arrange - use a non-existent directory
	repo := repository.NewFileRepository("/path/that/does/not/exist/for/testing")

	// Act
	gentxs, err := repo.FindAll()

	// Assert - should return empty slice, not error (directory doesn't exist yet)
	require.NoError(t, err)
	assert.Equal(t, 0, len(gentxs))
}

func TestFileRepository_FindByValidator_CorruptedFile(t *testing.T) {
	// Arrange
	tempDir := createTempDir(t)
	repo := repository.NewFileRepository(tempDir)
	validatorAddr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"

	// Create a corrupted gentx file manually
	validator, _ := domain.NewAddress(validatorAddr)
	// Access private method through reflection or create file manually
	fileName := "gentx-742d35cc6634c0532925a3b844bc9e7595f0beb0.json"
	corruptedFilePath := filepath.Join(tempDir, fileName)
	err := os.WriteFile(corruptedFilePath, []byte("invalid json"), 0644)
	require.NoError(t, err)

	// Act
	_, err = repo.FindByValidator(validator)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read gentx file")
}

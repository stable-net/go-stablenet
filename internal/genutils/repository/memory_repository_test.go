package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
)

func TestNewMemoryRepository(t *testing.T) {
	// Act
	repo := repository.NewMemoryRepository()

	// Assert
	assert.NotNil(t, repo)
}

func TestMemoryRepository_Save(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
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

	// Verify it was saved
	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMemoryRepository_Save_DuplicateValidator(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
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

func TestMemoryRepository_FindAll(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()

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

func TestMemoryRepository_FindAll_Empty(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()

	// Act
	gentxs, err := repo.FindAll()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, len(gentxs))
}

func TestMemoryRepository_FindByValidator(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
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

func TestMemoryRepository_FindByValidator_NotFound(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	_, err := repo.FindByValidator(validator)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenTxNotFound)
}

func TestMemoryRepository_Exists(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
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

func TestMemoryRepository_Exists_NotFound(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	exists, err := repo.Exists(validator)

	// Assert
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMemoryRepository_Delete(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
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

func TestMemoryRepository_Delete_NotFound(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()
	validator, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	// Act
	err := repo.Delete(validator)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenTxNotFound)
}

func TestMemoryRepository_Count(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()

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

func TestMemoryRepository_Count_Empty(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()

	// Act
	count, err := repo.Count()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMemoryRepository_Concurrency(t *testing.T) {
	// Arrange
	repo := repository.NewMemoryRepository()

	// Act - concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			addr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb" + string(rune('0'+index))
			gentx := createTestGenTx(
				t,
				addr,
				"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
				createTestBLSKey(byte(index+1)),
			)
			err := repo.Save(gentx)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Assert
	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, 10, count)
}

func TestMemoryRepository_IsolationBetweenInstances(t *testing.T) {
	// Arrange
	repo1 := repository.NewMemoryRepository()
	repo2 := repository.NewMemoryRepository()

	gentx := createTestGenTx(
		t,
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		"0x853d45Dd7734D1643936a4c845Bc0e8595f1cFc1",
		createTestBLSKey(1),
	)

	// Act
	require.NoError(t, repo1.Save(gentx))

	// Assert - repo2 should not have the gentx
	count1, _ := repo1.Count()
	count2, _ := repo2.Count()

	assert.Equal(t, 1, count1)
	assert.Equal(t, 0, count2)
}

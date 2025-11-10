package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

const (
	testBLSKey1 = "0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	testBLSKey2 = "0x" +
		"b11a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"1d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"

	testBLSKey3 = "0x" +
		"c22a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"2d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c"
)

// Test helper to create a valid GenTx with unique addresses and BLS key
func createTestGenTx(t *testing.T, validatorAddrHex, operatorAddrHex, blsKeyHex string) domain.GenTx {
	t.Helper()

	validatorAddr, err := domain.NewAddress(validatorAddrHex)
	require.NoError(t, err)

	operatorAddr, err := domain.NewAddress(operatorAddrHex)
	require.NoError(t, err)

	blsKey, err := domain.NewBLSPublicKey(blsKeyHex)
	require.NoError(t, err)

	metadata, err := domain.NewValidatorMetadata("TestValidator", "Test description", "https://test.com")
	require.NoError(t, err)

	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i + 1)
	}
	signature, err := domain.NewSignature(sigBytes)
	require.NoError(t, err)

	gentx, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, "stablenet-1", time.Now().UTC())
	require.NoError(t, err)

	return gentx
}

func TestGenTxCollection_NewGenTxCollection(t *testing.T) {
	// Act
	collection := domain.NewGenTxCollection()

	// Assert
	assert.NotNil(t, collection)
	assert.Equal(t, 0, collection.Size(), "new collection should be empty")
	assert.True(t, collection.IsEmpty(), "new collection should report as empty")
}

func TestGenTxCollection_Add_ValidGenTx(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)

	// Act
	err := collection.Add(gentx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, collection.Size())
	assert.False(t, collection.IsEmpty())
}

func TestGenTxCollection_Add_DuplicateValidatorAddress(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx1 := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	gentx2 := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x999d35Cc6634C0532925a3b844Bc9e7595f09990", testBLSKey2)

	// Act
	err1 := collection.Add(gentx1)
	err2 := collection.Add(gentx2)

	// Assert
	require.NoError(t, err1)
	require.Error(t, err2)
	assert.ErrorIs(t, err2, domain.ErrDuplicateValidatorAddress)
	assert.Equal(t, 1, collection.Size(), "duplicate should not be added")
}

func TestGenTxCollection_Add_DuplicateOperatorAddress(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx1 := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	gentx2 := createTestGenTx(t, "0x999d35Cc6634C0532925a3b844Bc9e7595f09990", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey2)

	// Act
	err1 := collection.Add(gentx1)
	err2 := collection.Add(gentx2)

	// Assert
	require.NoError(t, err1)
	require.Error(t, err2)
	assert.ErrorIs(t, err2, domain.ErrDuplicateOperatorAddress)
	assert.Equal(t, 1, collection.Size(), "duplicate should not be added")
}

func TestGenTxCollection_Contains_ValidatorAddress(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	validatorAddr := gentx.ValidatorAddress()
	otherAddr, _ := domain.NewAddress("0x999d35Cc6634C0532925a3b844Bc9e7595f09990")

	// Act
	collection.Add(gentx)

	// Assert
	assert.True(t, collection.ContainsValidator(validatorAddr), "should contain added validator")
	assert.False(t, collection.ContainsValidator(otherAddr), "should not contain other validator")
}

func TestGenTxCollection_Contains_OperatorAddress(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	operatorAddr := gentx.OperatorAddress()
	otherAddr, _ := domain.NewAddress("0x999d35Cc6634C0532925a3b844Bc9e7595f09990")

	// Act
	collection.Add(gentx)

	// Assert
	assert.True(t, collection.ContainsOperator(operatorAddr), "should contain added operator")
	assert.False(t, collection.ContainsOperator(otherAddr), "should not contain other operator")
}

func TestGenTxCollection_GetAll(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx1 := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	gentx2 := createTestGenTx(t, "0x999d35Cc6634C0532925a3b844Bc9e7595f09990", "0x888d35Cc6634C0532925a3b844Bc9e7595f08880", testBLSKey2)

	// Act
	collection.Add(gentx1)
	collection.Add(gentx2)
	allGenTxs := collection.GetAll()

	// Assert
	assert.Equal(t, 2, len(allGenTxs))
	assert.Contains(t, allGenTxs, gentx1)
	assert.Contains(t, allGenTxs, gentx2)
}

func TestGenTxCollection_GetAll_ReturnsCopy(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	collection.Add(gentx)

	// Act - modify returned slice
	allGenTxs := collection.GetAll()
	_ = append(allGenTxs, domain.GenTx{})

	// Assert - original collection unchanged
	assert.Equal(t, 1, collection.Size(), "modifying returned slice should not affect collection")
}

func TestGenTxCollection_Remove_ExistingGenTx(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	collection.Add(gentx)

	// Act
	err := collection.Remove(gentx.ValidatorAddress())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, collection.Size())
	assert.True(t, collection.IsEmpty())
}

func TestGenTxCollection_Remove_NonExistingGenTx(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	nonExistingAddr, _ := domain.NewAddress("0x999d35Cc6634C0532925a3b844Bc9e7595f09990")

	// Act
	err := collection.Remove(nonExistingAddr)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrGenTxNotFound)
}

func TestGenTxCollection_Clear(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	gentx1 := createTestGenTx(t, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey1)
	gentx2 := createTestGenTx(t, "0x999d35Cc6634C0532925a3b844Bc9e7595f09990", "0x888d35Cc6634C0532925a3b844Bc9e7595f08880", testBLSKey2)
	collection.Add(gentx1)
	collection.Add(gentx2)

	// Act
	collection.Clear()

	// Assert
	assert.Equal(t, 0, collection.Size())
	assert.True(t, collection.IsEmpty())
}

func TestGenTxCollection_GetSorted_DeterministicOrder(t *testing.T) {
	// Arrange
	collection := domain.NewGenTxCollection()
	// Add in random order
	gentx3 := createTestGenTx(t, "0xCCCd35Cc6634C0532925a3b844Bc9e7595f0CCC0", "0x123d35Cc6634C0532925a3b844Bc9e7595f01230", testBLSKey3)
	gentx1 := createTestGenTx(t, "0xAAAd35Cc6634C0532925a3b844Bc9e7595f0AAA0", "0x888d35Cc6634C0532925a3b844Bc9e7595f08880", testBLSKey1)
	gentx2 := createTestGenTx(t, "0xBBBd35Cc6634C0532925a3b844Bc9e7595f0BBB0", "0x999d35Cc6634C0532925a3b844Bc9e7595f09990", testBLSKey2)

	collection.Add(gentx3)
	collection.Add(gentx1)
	collection.Add(gentx2)

	// Act
	sorted := collection.GetSorted()

	// Assert - should be sorted by validator address (lexicographically)
	assert.Equal(t, 3, len(sorted))
	assert.Equal(t, gentx1.ValidatorAddress(), sorted[0].ValidatorAddress())
	assert.Equal(t, gentx2.ValidatorAddress(), sorted[1].ValidatorAddress())
	assert.Equal(t, gentx3.ValidatorAddress(), sorted[2].ValidatorAddress())
}

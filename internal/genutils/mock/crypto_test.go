package mock_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/mock"
)

func TestMockCryptoProvider_Sign(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Act
	signature, err := provider.Sign(privateKey, message)

	// Assert
	require.NoError(t, err)
	assert.False(t, signature.IsZero(), "signature should not be zero")
}

func TestMockCryptoProvider_Sign_Deterministic(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Act - sign twice
	sig1, err1 := provider.Sign(privateKey, message)
	sig2, err2 := provider.Sign(privateKey, message)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.True(t, sig1.Equals(sig2), "signing should be deterministic")
}

func TestMockCryptoProvider_RecoverAddress(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Sign message to get signature
	signature, err := provider.Sign(privateKey, message)
	require.NoError(t, err)

	// Act
	address, err := provider.RecoverAddress(message, signature)

	// Assert
	require.NoError(t, err)
	assert.False(t, address.IsZero(), "address should not be zero")
}

func TestMockCryptoProvider_Verify(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Sign message
	signature, err := provider.Sign(privateKey, message)
	require.NoError(t, err)

	// Recover address from signature
	address, err := provider.RecoverAddress(message, signature)
	require.NoError(t, err)

	// Act
	valid, err := provider.Verify(address, message, signature)

	// Assert
	require.NoError(t, err)
	assert.True(t, valid, "signature should be valid")
}

func TestMockCryptoProvider_Verify_WrongAddress(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Sign message
	signature, err := provider.Sign(privateKey, message)
	require.NoError(t, err)

	// Create different address
	wrongAddress, err := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	require.NoError(t, err)

	// Act
	valid, err := provider.Verify(wrongAddress, message, signature)

	// Assert
	require.NoError(t, err)
	assert.False(t, valid, "signature should be invalid for wrong address")
}

func TestMockCryptoProvider_Verify_WrongMessage(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")
	message := []byte("test message")
	wrongMessage := []byte("wrong message")

	// Sign original message
	signature, err := provider.Sign(privateKey, message)
	require.NoError(t, err)

	// Recover address from signature
	address, err := provider.RecoverAddress(message, signature)
	require.NoError(t, err)

	// Act - verify with wrong message
	valid, err := provider.Verify(address, wrongMessage, signature)

	// Assert
	require.NoError(t, err)
	assert.False(t, valid, "signature should be invalid for wrong message")
}

func TestMockCryptoProvider_HashMessage(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	message := []byte("test message")

	// Act
	hash := provider.HashMessage(message)

	// Assert
	assert.NotNil(t, hash)
	assert.Greater(t, len(hash), 0, "hash should not be empty")

	// Verify deterministic
	hash2 := provider.HashMessage(message)
	assert.Equal(t, hash, hash2, "hashing should be deterministic")
}

func TestMockCryptoProvider_DeriveBLSPublicKey(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")

	// Act
	blsKey, err := provider.DeriveBLSPublicKey(privateKey)

	// Assert
	require.NoError(t, err)
	assert.False(t, blsKey.IsZero(), "BLS key should not be zero")
}

func TestMockCryptoProvider_DeriveBLSPublicKey_Deterministic(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	privateKey := []byte("test-private-key")

	// Act - derive twice
	blsKey1, err1 := provider.DeriveBLSPublicKey(privateKey)
	blsKey2, err2 := provider.DeriveBLSPublicKey(privateKey)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.True(t, blsKey1.Equals(blsKey2), "BLS key derivation should be deterministic")
}

func TestMockCryptoProvider_SetSignError(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	provider.SetSignError(true)
	privateKey := []byte("test-private-key")
	message := []byte("test message")

	// Act
	_, err := provider.Sign(privateKey, message)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock sign error")
}

func TestMockCryptoProvider_SetVerifyError(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	provider.SetVerifyError(true)
	address, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	signature, _ := domain.NewSignature(make([]byte, 65))
	message := []byte("test message")

	// Act
	_, err := provider.Verify(address, message, signature)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock verify error")
}

func TestMockCryptoProvider_SetRecoverError(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	provider.SetRecoverError(true)
	signature, _ := domain.NewSignature(make([]byte, 65))
	message := []byte("test message")

	// Act
	_, err := provider.RecoverAddress(message, signature)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock recover error")
}

func TestMockCryptoProvider_SetDeriveBLSError(t *testing.T) {
	// Arrange
	provider := mock.NewMockCryptoProvider()
	provider.SetDeriveBLSError(true)
	privateKey := []byte("test-private-key")

	// Act
	_, err := provider.DeriveBLSPublicKey(privateKey)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock derive BLS error")
}

package ethereum_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

// Test helper to generate a test private key
func generateTestPrivateKey(t *testing.T) (*ecdsa.PrivateKey, []byte) {
	t.Helper()
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	privateKeyBytes := crypto.FromECDSA(privateKey)
	return privateKey, privateKeyBytes
}

func TestEthereumProvider_Sign(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	_, privateKeyBytes := generateTestPrivateKey(t)
	message := []byte("test message")

	// Act
	signature, err := provider.Sign(privateKeyBytes, message)

	// Assert
	require.NoError(t, err)
	assert.False(t, signature.IsZero(), "signature should not be zero")
}

func TestEthereumProvider_Sign_InvalidPrivateKey(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	invalidKey := []byte("invalid key")
	message := []byte("test message")

	// Act
	_, err := provider.Sign(invalidKey, message)

	// Assert
	require.Error(t, err)
}

func TestEthereumProvider_RecoverAddress(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	privateKey, privateKeyBytes := generateTestPrivateKey(t)
	message := []byte("test message")

	// Create expected address from private key
	expectedAddress, err := domain.NewAddress(crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	require.NoError(t, err)

	// Sign message
	signature, err := provider.Sign(privateKeyBytes, message)
	require.NoError(t, err)

	// Act
	recoveredAddress, err := provider.RecoverAddress(message, signature)

	// Assert
	require.NoError(t, err)
	assert.True(t, expectedAddress.Equals(recoveredAddress), "recovered address should match expected")
}

func TestEthereumProvider_RecoverAddress_InvalidSignature(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	message := []byte("test message")

	// Create invalid signature (all zeros)
	invalidSigBytes := make([]byte, 65)
	invalidSig, _ := domain.NewSignature(invalidSigBytes)

	// Act
	_, err := provider.RecoverAddress(message, invalidSig)

	// Assert
	require.Error(t, err)
}

func TestEthereumProvider_Verify(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	privateKey, privateKeyBytes := generateTestPrivateKey(t)
	message := []byte("test message")

	// Create address from private key
	address, err := domain.NewAddress(crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	require.NoError(t, err)

	// Sign message
	signature, err := provider.Sign(privateKeyBytes, message)
	require.NoError(t, err)

	// Act
	valid, err := provider.Verify(address, message, signature)

	// Assert
	require.NoError(t, err)
	assert.True(t, valid, "signature should be valid")
}

func TestEthereumProvider_Verify_WrongAddress(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	_, privateKeyBytes := generateTestPrivateKey(t)
	message := []byte("test message")

	// Create different address
	wrongAddress, err := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	require.NoError(t, err)

	// Sign message
	signature, err := provider.Sign(privateKeyBytes, message)
	require.NoError(t, err)

	// Act
	valid, err := provider.Verify(wrongAddress, message, signature)

	// Assert
	require.NoError(t, err)
	assert.False(t, valid, "signature should be invalid for wrong address")
}

func TestEthereumProvider_Verify_WrongMessage(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	privateKey, privateKeyBytes := generateTestPrivateKey(t)
	message := []byte("test message")
	wrongMessage := []byte("wrong message")

	// Create address from private key
	address, err := domain.NewAddress(crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	require.NoError(t, err)

	// Sign original message
	signature, err := provider.Sign(privateKeyBytes, message)
	require.NoError(t, err)

	// Act - verify with wrong message
	valid, err := provider.Verify(address, wrongMessage, signature)

	// Assert
	require.NoError(t, err)
	assert.False(t, valid, "signature should be invalid for wrong message")
}

func TestEthereumProvider_HashMessage(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	message := []byte("test message")

	// Act
	hash := provider.HashMessage(message)

	// Assert
	assert.NotNil(t, hash)
	assert.Equal(t, 32, len(hash), "hash should be 32 bytes (Keccak256)")

	// Verify deterministic - same message should produce same hash
	hash2 := provider.HashMessage(message)
	assert.Equal(t, hash, hash2, "hashing should be deterministic")
}

func TestEthereumProvider_DeriveBLSPublicKey(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	_, privateKeyBytes := generateTestPrivateKey(t)

	// Act
	blsKey, err := provider.DeriveBLSPublicKey(privateKeyBytes)

	// Assert
	require.NoError(t, err)
	assert.False(t, blsKey.IsZero(), "BLS key should not be zero")
}

func TestEthereumProvider_DeriveBLSPublicKey_Deterministic(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	_, privateKeyBytes := generateTestPrivateKey(t)

	// Act - derive twice
	blsKey1, err1 := provider.DeriveBLSPublicKey(privateKeyBytes)
	blsKey2, err2 := provider.DeriveBLSPublicKey(privateKeyBytes)

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.True(t, blsKey1.Equals(blsKey2), "BLS key derivation should be deterministic")
}

func TestEthereumProvider_DeriveBLSPublicKey_InvalidPrivateKey(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	invalidKey := []byte("invalid key")

	// Act
	_, err := provider.DeriveBLSPublicKey(invalidKey)

	// Assert
	require.Error(t, err)
}

func TestEthereumProvider_SignAndVerify_Integration(t *testing.T) {
	// Arrange
	provider := ethcrypto.NewEthereumProvider()
	privateKey, privateKeyBytes := generateTestPrivateKey(t)
	address, _ := domain.NewAddress(crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
	message := []byte("integration test message")

	// Act
	signature, err := provider.Sign(privateKeyBytes, message)
	require.NoError(t, err)

	valid, err := provider.Verify(address, message, signature)
	require.NoError(t, err)

	recoveredAddr, err := provider.RecoverAddress(message, signature)
	require.NoError(t, err)

	// Assert
	assert.True(t, valid, "signature should be valid")
	assert.True(t, address.Equals(recoveredAddr), "recovered address should match original")
}

func TestValidatePrivateKey_Valid(t *testing.T) {
	// Arrange
	_, privateKeyBytes := generateTestPrivateKey(t)

	// Act
	err := ethcrypto.ValidatePrivateKey(privateKeyBytes)

	// Assert
	require.NoError(t, err)
}

func TestValidatePrivateKey_Invalid(t *testing.T) {
	// Arrange
	invalidKey := []byte("invalid key")

	// Act
	err := ethcrypto.ValidatePrivateKey(invalidKey)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key format")
}

func TestGeneratePrivateKey(t *testing.T) {
	// Act
	privateKeyBytes, err := ethcrypto.GeneratePrivateKey()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, privateKeyBytes)
	assert.Greater(t, len(privateKeyBytes), 0, "private key should not be empty")

	// Validate generated key
	err = ethcrypto.ValidatePrivateKey(privateKeyBytes)
	assert.NoError(t, err, "generated key should be valid")
}

func TestGeneratePrivateKey_Uniqueness(t *testing.T) {
	// Act - generate two keys
	key1, err1 := ethcrypto.GeneratePrivateKey()
	key2, err2 := ethcrypto.GeneratePrivateKey()

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, key1, key2, "generated keys should be unique")
}

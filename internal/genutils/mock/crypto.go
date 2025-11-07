package mock

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// MockCryptoProvider implements the CryptoProvider interface for testing purposes.
// It provides deterministic, predictable behavior without real cryptographic operations.
type MockCryptoProvider struct {
	signError      bool
	verifyError    bool
	recoverError   bool
	deriveBLSError bool
}

// NewMockCryptoProvider creates a new MockCryptoProvider instance.
func NewMockCryptoProvider() *MockCryptoProvider {
	return &MockCryptoProvider{}
}

// SetSignError configures the mock to return an error on Sign calls.
func (m *MockCryptoProvider) SetSignError(shouldError bool) {
	m.signError = shouldError
}

// SetVerifyError configures the mock to return an error on Verify calls.
func (m *MockCryptoProvider) SetVerifyError(shouldError bool) {
	m.verifyError = shouldError
}

// SetRecoverError configures the mock to return an error on RecoverAddress calls.
func (m *MockCryptoProvider) SetRecoverError(shouldError bool) {
	m.recoverError = shouldError
}

// SetDeriveBLSError configures the mock to return an error on DeriveBLSPublicKey calls.
func (m *MockCryptoProvider) SetDeriveBLSError(shouldError bool) {
	m.deriveBLSError = shouldError
}

// Sign creates a deterministic mock signature.
// The signature is derived from the private key and message using simple hashing.
func (m *MockCryptoProvider) Sign(privateKey []byte, message []byte) (domain.Signature, error) {
	if m.signError {
		return domain.Signature{}, errors.New("mock sign error")
	}

	// Create deterministic signature from privateKey + message
	data := append(privateKey, message...)
	hash := sha256.Sum256(data)

	// Create 65-byte signature (R: 32 bytes, S: 32 bytes, V: 1 byte)
	signatureBytes := make([]byte, 65)
	copy(signatureBytes[:32], hash[:]) // R

	// Hash again for S component
	hashS := sha256.Sum256(hash[:])
	copy(signatureBytes[32:64], hashS[:]) // S
	signatureBytes[64] = 27               // V (recovery ID)

	signature, err := domain.NewSignature(signatureBytes)
	if err != nil {
		return domain.Signature{}, fmt.Errorf("failed to create mock signature: %w", err)
	}

	return signature, nil
}

// Verify verifies a mock signature by comparing recovered address with expected address.
func (m *MockCryptoProvider) Verify(address domain.Address, message []byte, signature domain.Signature) (bool, error) {
	if m.verifyError {
		return false, errors.New("mock verify error")
	}

	// Recover address from signature
	recoveredAddress, err := m.RecoverAddress(message, signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover address: %w", err)
	}

	// Compare addresses
	return address.Equals(recoveredAddress), nil
}

// RecoverAddress recovers a mock address from the signature.
// The address is deterministically derived from the message and signature.
func (m *MockCryptoProvider) RecoverAddress(message []byte, signature domain.Signature) (domain.Address, error) {
	if m.recoverError {
		return domain.Address{}, errors.New("mock recover error")
	}

	// Create deterministic address from message + signature
	data := append(message, signature.Bytes()...)
	hash := sha256.Sum256(data)

	// Use first 20 bytes of hash as address
	addressHex := "0x" + hex.EncodeToString(hash[:20])

	address, err := domain.NewAddress(addressHex)
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to create mock address: %w", err)
	}

	return address, nil
}

// DeriveBLSPublicKey derives a mock BLS public key deterministically from the private key.
func (m *MockCryptoProvider) DeriveBLSPublicKey(privateKey []byte) (domain.BLSPublicKey, error) {
	if m.deriveBLSError {
		return domain.BLSPublicKey{}, errors.New("mock derive BLS error")
	}

	// Create deterministic 96-byte BLS key from private key
	result := make([]byte, 96)

	for i := 0; i < 3; i++ {
		data := append(privateKey, []byte("MOCK_BLS_KEY")...)
		data = append(data, byte(i))
		hash := sha256.Sum256(data)
		copy(result[i*32:], hash[:])
	}

	blsKeyHex := "0x" + hex.EncodeToString(result)
	blsKey, err := domain.NewBLSPublicKey(blsKeyHex)
	if err != nil {
		return domain.BLSPublicKey{}, fmt.Errorf("failed to create mock BLS key: %w", err)
	}

	return blsKey, nil
}

// HashMessage creates a simple hash of the message for testing.
func (m *MockCryptoProvider) HashMessage(message []byte) []byte {
	hash := sha256.Sum256(message)
	return hash[:]
}

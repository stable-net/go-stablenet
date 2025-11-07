package ethereum

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// EthereumProvider implements the CryptoProvider interface using Ethereum's ECDSA.
type EthereumProvider struct{}

// NewEthereumProvider creates a new EthereumProvider instance.
func NewEthereumProvider() *EthereumProvider {
	return &EthereumProvider{}
}

// Sign signs a message with the given private key using ECDSA.
// The message is hashed with Keccak256 before signing.
func (p *EthereumProvider) Sign(privateKeyBytes []byte, message []byte) (domain.Signature, error) {
	// Convert bytes to ECDSA private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return domain.Signature{}, fmt.Errorf("invalid private key: %w", err)
	}

	// Hash the message
	hash := p.HashMessage(message)

	// Sign the hash
	signatureBytes, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return domain.Signature{}, fmt.Errorf("signing failed: %w", err)
	}

	// Create domain Signature
	signature, err := domain.NewSignature(signatureBytes)
	if err != nil {
		return domain.Signature{}, fmt.Errorf("failed to create signature: %w", err)
	}

	return signature, nil
}

// Verify verifies that the signature was created by signing the message
// with the private key corresponding to the given address.
func (p *EthereumProvider) Verify(address domain.Address, message []byte, signature domain.Signature) (bool, error) {
	// Recover address from signature
	recoveredAddress, err := p.RecoverAddress(message, signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover address: %w", err)
	}

	// Compare addresses
	return address.Equals(recoveredAddress), nil
}

// RecoverAddress recovers the Ethereum address from a signature.
func (p *EthereumProvider) RecoverAddress(message []byte, signature domain.Signature) (domain.Address, error) {
	// Hash the message
	hash := p.HashMessage(message)

	// Get signature bytes
	signatureBytes := signature.Bytes()

	// Recover public key from signature
	publicKeyBytes, err := crypto.Ecrecover(hash, signatureBytes)
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Convert public key bytes to ECDSA public key
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// Get address from public key
	ethAddress := crypto.PubkeyToAddress(*publicKey)

	// Create domain Address
	address, err := domain.NewAddress(ethAddress.Hex())
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

// DeriveBLSPublicKey derives a BLS public key from an Ethereum private key.
// This implementation uses a deterministic derivation from the ECDSA private key.
func (p *EthereumProvider) DeriveBLSPublicKey(privateKeyBytes []byte) (domain.BLSPublicKey, error) {
	// Convert bytes to ECDSA private key
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return domain.BLSPublicKey{}, fmt.Errorf("invalid private key: %w", err)
	}

	// Derive BLS key using deterministic method
	// In production, this should use proper BLS12-381 key derivation
	// For now, we use a simple deterministic approach:
	// Hash(privateKey || "BLS") to get 96-byte BLS public key
	blsKeyMaterial := deriveBLSKeyMaterial(privateKey)

	// Create domain BLSPublicKey
	blsKey, err := domain.NewBLSPublicKey("0x" + common.Bytes2Hex(blsKeyMaterial))
	if err != nil {
		return domain.BLSPublicKey{}, fmt.Errorf("failed to create BLS key: %w", err)
	}

	return blsKey, nil
}

// deriveBLSKeyMaterial derives 96-byte BLS key material from ECDSA private key.
// This is a placeholder implementation - production should use proper BLS12-381.
func deriveBLSKeyMaterial(privateKey *ecdsa.PrivateKey) []byte {
	// Get private key bytes
	privKeyBytes := crypto.FromECDSA(privateKey)

	// Create deterministic BLS key material
	// Hash(privKey || "BLS_KEY_DERIVATION" || index) for each 32-byte chunk
	result := make([]byte, 96)

	for i := 0; i < 3; i++ {
		data := append(privKeyBytes, []byte("BLS_KEY_DERIVATION")...)
		data = append(data, byte(i))
		hash := crypto.Keccak256(data)
		copy(result[i*32:], hash)
	}

	return result
}

// HashMessage creates a Keccak256 hash of the message.
func (p *EthereumProvider) HashMessage(message []byte) []byte {
	return crypto.Keccak256(message)
}

// ValidatePrivateKey validates that the given bytes represent a valid ECDSA private key.
func ValidatePrivateKey(privateKeyBytes []byte) error {
	_, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return errors.New("invalid private key format")
	}
	return nil
}

// GeneratePrivateKey generates a new ECDSA private key.
func GeneratePrivateKey() ([]byte, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}
	return crypto.FromECDSA(privateKey), nil
}

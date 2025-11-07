package domain

// CryptoProvider defines the interface for cryptographic operations.
// This interface abstracts away the specific cryptographic implementation,
// allowing different implementations (e.g., Ethereum ECDSA, mock for testing).
// This follows the Dependency Inversion Principle from SOLID.
type CryptoProvider interface {
	// Sign signs a message with the given private key and returns the signature.
	// The message is typically a hash of the data to be signed.
	// Returns error if signing fails.
	Sign(privateKey []byte, message []byte) (Signature, error)

	// Verify verifies that the signature was created by signing the message
	// with the private key corresponding to the given address.
	// Returns true if verification succeeds, false otherwise.
	Verify(address Address, message []byte, signature Signature) (bool, error)

	// RecoverAddress recovers the Ethereum address from a signature.
	// This is useful for verifying that a signature was created by a specific key
	// without needing the public key explicitly.
	// Returns the recovered address or error if recovery fails.
	RecoverAddress(message []byte, signature Signature) (Address, error)

	// DeriveBLSPublicKey derives a BLS public key from an Ethereum private key.
	// This is used to generate the BLS public key for validator consensus.
	// Returns error if derivation fails.
	DeriveBLSPublicKey(privateKey []byte) (BLSPublicKey, error)

	// HashMessage creates a hash of the message suitable for signing.
	// This typically uses Keccak256 for Ethereum compatibility.
	HashMessage(message []byte) []byte
}

// SignatureData represents the structured data to be signed.
// This is used to create deterministic signatures for GenTx.
type SignatureData struct {
	ValidatorAddress Address
	OperatorAddress  Address
	BLSPublicKey     BLSPublicKey
	ChainID          string
	Timestamp        int64 // Unix timestamp
}

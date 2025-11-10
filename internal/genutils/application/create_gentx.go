package application

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// CreateGenTxRequest contains parameters for creating a new GenTx.
type CreateGenTxRequest struct {
	// ValidatorKey is the private key used for validator address derivation and signing
	ValidatorKey []byte

	// OperatorAddress is the address that will operate the validator (can be multisig)
	OperatorAddr domain.Address

	// Metadata contains validator information (name, description, website)
	Metadata domain.ValidatorMetadata

	// ChainID is the network identifier
	ChainID string

	// Timestamp is the GenTx creation time
	Timestamp time.Time
}

// CreateGenTxUseCase implements the use case for creating a new GenTx.
//
// The use case performs the following operations:
//  1. Validates the request parameters
//  2. Derives validator address from private key
//  3. Derives BLS public key from private key
//  4. Creates and signs the GenTx
//  5. Validates the GenTx using ValidationService
//  6. Saves the GenTx to repository
//  7. Returns the created GenTx
//
// This use case orchestrates domain objects and services to implement
// the complete GenTx creation workflow.
type CreateGenTxUseCase struct {
	repository     Repository
	cryptoProvider CryptoProvider
	validator      Validator
}

// Repository defines the interface for GenTx storage operations.
type Repository interface {
	// Save persists a GenTx to storage.
	Save(gentx domain.GenTx) error

	// FindByValidator retrieves a GenTx by validator address.
	FindByValidator(validatorAddr domain.Address) (domain.GenTx, error)

	// FindAll retrieves all GenTxs from storage.
	FindAll() ([]domain.GenTx, error)
}

// CryptoProvider defines the interface for cryptographic operations.
type CryptoProvider interface {
	// Sign creates a signature for the given message using the private key.
	Sign(privateKey []byte, message []byte) (domain.Signature, error)

	// RecoverAddress recovers the address from a signature and message.
	RecoverAddress(message []byte, signature domain.Signature) (domain.Address, error)

	// DeriveBLSPublicKey derives a BLS public key from an ECDSA private key.
	DeriveBLSPublicKey(privateKey []byte) (domain.BLSPublicKey, error)
}

// Validator defines the interface for GenTx validation.
type Validator interface {
	// Validate performs comprehensive validation on a GenTx.
	Validate(gentx domain.GenTx) error
}

// NewCreateGenTxUseCase creates a new CreateGenTxUseCase instance.
//
// Parameters:
//   - repository: Repository for persisting GenTx
//   - cryptoProvider: Crypto provider for signing and key derivation
//   - validator: Validator for validating GenTx
//
// Returns a new CreateGenTxUseCase ready for use.
func NewCreateGenTxUseCase(
	repository Repository,
	cryptoProvider CryptoProvider,
	validator Validator,
) *CreateGenTxUseCase {
	return &CreateGenTxUseCase{
		repository:     repository,
		cryptoProvider: cryptoProvider,
		validator:      validator,
	}
}

// Execute executes the CreateGenTx use case.
//
// The method performs comprehensive request validation, derives the validator
// address from the private key, creates and signs the GenTx, validates it,
// and persists it to the repository.
//
// Parameters:
//   - request: CreateGenTxRequest containing all parameters
//
// Returns:
//   - domain.GenTx: The created and persisted GenTx
//   - error: Error if validation fails or GenTx creation fails
//
// Errors:
//   - Request validation errors (nil request, missing fields)
//   - Cryptographic errors (key derivation, signing)
//   - Domain validation errors (invalid GenTx)
//   - Repository errors (duplicate validator, save failure)
//
// Example:
//
//	request := &CreateGenTxRequest{
//	    ValidatorKey: privateKey,
//	    OperatorAddr: operatorAddress,
//	    Metadata:     metadata,
//	    ChainID:      "stable-testnet-1",
//	    Timestamp:    time.Now().UTC(),
//	}
//	gentx, err := useCase.Execute(request)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Created GenTx for validator: %s\n", gentx.ValidatorAddress())
func (uc *CreateGenTxUseCase) Execute(request *CreateGenTxRequest) (domain.GenTx, error) {
	// Validate request
	if err := uc.validateRequest(request); err != nil {
		return domain.GenTx{}, fmt.Errorf("request validation failed: %w", err)
	}

	// Derive validator address from private key
	validatorAddr, err := uc.deriveValidatorAddress(request.ValidatorKey)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to derive validator address: %w", err)
	}

	// Derive BLS public key
	blsPublicKey, err := uc.cryptoProvider.DeriveBLSPublicKey(request.ValidatorKey)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to derive BLS public key: %w", err)
	}

	// Create signature data
	sigData := domain.SignatureData{
		ValidatorAddress: validatorAddr,
		OperatorAddress:  request.OperatorAddr,
		BLSPublicKey:     blsPublicKey,
		ChainID:          request.ChainID,
		Timestamp:        request.Timestamp.Unix(),
	}

	// Sign the data
	signature, err := uc.cryptoProvider.Sign(request.ValidatorKey, sigData.Bytes())
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to sign GenTx: %w", err)
	}

	// Create GenTx
	gentx, err := domain.NewGenTx(
		validatorAddr,
		request.OperatorAddr,
		blsPublicKey,
		request.Metadata,
		signature,
		request.ChainID,
		request.Timestamp,
	)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to create GenTx: %w", err)
	}

	// Validate GenTx
	if err := uc.validator.Validate(gentx); err != nil {
		return domain.GenTx{}, fmt.Errorf("GenTx validation failed: %w", err)
	}

	// Save to repository
	if err := uc.repository.Save(gentx); err != nil {
		return domain.GenTx{}, fmt.Errorf("failed to save GenTx: %w", err)
	}

	return gentx, nil
}

// validateRequest validates the CreateGenTxRequest.
func (uc *CreateGenTxUseCase) validateRequest(request *CreateGenTxRequest) error {
	if request == nil {
		return errors.New("request cannot be nil")
	}

	if request.ValidatorKey == nil || len(request.ValidatorKey) == 0 {
		return errors.New("validator key cannot be empty")
	}

	if request.OperatorAddr.IsZero() {
		return errors.New("operator address cannot be zero")
	}

	if request.Metadata.Name() == "" {
		return errors.New("metadata cannot be empty")
	}

	if request.ChainID == "" {
		return errors.New("chain ID cannot be empty")
	}

	if request.Timestamp.After(time.Now().UTC()) {
		return errors.New("timestamp cannot be in the future")
	}

	return nil
}

// deriveValidatorAddress derives the validator address from the private key.
func (uc *CreateGenTxUseCase) deriveValidatorAddress(privateKey []byte) (domain.Address, error) {
	// Create a temporary message to sign
	tempMessage := []byte("validator_address_derivation")

	// Sign the message
	signature, err := uc.cryptoProvider.Sign(privateKey, tempMessage)
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to sign temporary message: %w", err)
	}

	// Recover the address from the signature
	address, err := uc.cryptoProvider.RecoverAddress(tempMessage, signature)
	if err != nil {
		return domain.Address{}, fmt.Errorf("failed to recover address: %w", err)
	}

	return address, nil
}

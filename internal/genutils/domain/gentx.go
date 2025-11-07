package domain

import (
	"strings"
	"time"
)

// GenTx represents a genesis transaction aggregate root.
// It is immutable and ensures all business rules are enforced.
type GenTx struct {
	validatorAddress Address
	operatorAddress  Address
	blsPublicKey     BLSPublicKey
	metadata         ValidatorMetadata
	signature        Signature
	chainID          string
	timestamp        time.Time
}

// NewGenTx creates a new GenTx aggregate.
// All fields are required and validated.
// Returns appropriate errors if validation fails.
func NewGenTx(
	validatorAddr, operatorAddr Address,
	blsKey BLSPublicKey,
	metadata ValidatorMetadata,
	signature Signature,
	chainID string,
	timestamp time.Time,
) (GenTx, error) {
	// Trim whitespace from chainID
	chainID = strings.TrimSpace(chainID)

	// Validate validator address
	if validatorAddr.IsZero() {
		return GenTx{}, ErrInvalidValidatorAddress
	}

	// Validate operator address
	if operatorAddr.IsZero() {
		return GenTx{}, ErrInvalidOperatorAddress
	}

	// Validate BLS public key
	if blsKey.IsZero() {
		return GenTx{}, ErrInvalidBLSKey
	}

	// Validate metadata
	if metadata.IsZero() {
		return GenTx{}, ErrMissingValidatorName
	}

	// Validate signature
	if signature.IsZero() {
		return GenTx{}, ErrMissingSignature
	}

	// Validate chain ID
	if chainID == "" {
		return GenTx{}, ErrMissingChainID
	}

	// Validate timestamp (should not be in the future)
	if timestamp.After(time.Now().UTC()) {
		return GenTx{}, ErrInvalidTimestamp
	}

	return GenTx{
		validatorAddress: validatorAddr,
		operatorAddress:  operatorAddr,
		blsPublicKey:     blsKey,
		metadata:         metadata,
		signature:        signature,
		chainID:          chainID,
		timestamp:        timestamp,
	}, nil
}

// ValidatorAddress returns the validator address.
func (g GenTx) ValidatorAddress() Address {
	return g.validatorAddress
}

// OperatorAddress returns the operator address.
func (g GenTx) OperatorAddress() Address {
	return g.operatorAddress
}

// BLSPublicKey returns the BLS public key.
func (g GenTx) BLSPublicKey() BLSPublicKey {
	return g.blsPublicKey
}

// Metadata returns the validator metadata.
func (g GenTx) Metadata() ValidatorMetadata {
	return g.metadata
}

// Signature returns the signature.
func (g GenTx) Signature() Signature {
	return g.signature
}

// ChainID returns the chain ID.
func (g GenTx) ChainID() string {
	return g.chainID
}

// Timestamp returns the creation timestamp.
func (g GenTx) Timestamp() time.Time {
	return g.timestamp
}

// Equals compares two GenTx for equality.
// Two GenTx are equal if all their fields match.
func (g GenTx) Equals(other GenTx) bool {
	return g.validatorAddress.Equals(other.validatorAddress) &&
		g.operatorAddress.Equals(other.operatorAddress) &&
		g.blsPublicKey.Equals(other.blsPublicKey) &&
		g.metadata.Equals(other.metadata) &&
		g.signature.Equals(other.signature) &&
		g.chainID == other.chainID &&
		g.timestamp.Equal(other.timestamp)
}

// IsZero returns true if the GenTx is the zero value.
func (g GenTx) IsZero() bool {
	return g.validatorAddress.IsZero() &&
		g.operatorAddress.IsZero() &&
		g.blsPublicKey.IsZero() &&
		g.metadata.IsZero() &&
		g.signature.IsZero() &&
		g.chainID == "" &&
		g.timestamp.IsZero()
}

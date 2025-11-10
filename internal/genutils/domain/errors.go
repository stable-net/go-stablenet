package domain

import "errors"

// Domain errors for GenUtils
var (
	// Address errors
	ErrInvalidAddress = errors.New("invalid ethereum address")

	// Signature errors
	ErrInvalidSignature       = errors.New("invalid signature")
	ErrInvalidSignatureLength = errors.New("signature must be 65 bytes")
	ErrSignatureVerification  = errors.New("signature verification failed")

	// BLS key errors
	ErrInvalidBLSKey = errors.New("invalid BLS public key")

	// Metadata errors
	ErrMissingValidatorName = errors.New("validator name is required")
	ErrValidatorNameTooLong = errors.New("validator name too long (max 70 chars)")
	ErrDescriptionTooLong   = errors.New("description too long (max 280 chars)")

	// GenTx errors
	ErrInvalidValidatorAddress = errors.New("invalid validator address")
	ErrInvalidOperatorAddress  = errors.New("invalid operator address")
	ErrMissingSignature        = errors.New("signature is missing")
	ErrMissingChainID          = errors.New("chain ID is missing")
	ErrInvalidTimestamp        = errors.New("invalid timestamp")

	// Collection errors
	ErrDuplicateValidatorAddress = errors.New("duplicate validator address")
	ErrDuplicateOperatorAddress  = errors.New("duplicate operator address")
	ErrDuplicateBLSKey           = errors.New("duplicate BLS public key")
	ErrGenTxNotFound             = errors.New("gentx not found in collection")
	ErrChainIDMismatch           = errors.New("chain ID mismatch")
)

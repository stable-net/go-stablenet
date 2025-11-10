package validation

import (
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// FormatValidator validates GenTx format constraints.
type FormatValidator struct{}

// NewFormatValidator creates a new FormatValidator instance.
func NewFormatValidator() *FormatValidator {
	return &FormatValidator{}
}

// Validate checks GenTx format constraints.
// All format validation is already done by the domain objects,
// so this primarily validates that all required fields are present and non-zero.
func (v *FormatValidator) Validate(gentx domain.GenTx) error {
	// Check validator address
	if gentx.ValidatorAddress().IsZero() {
		return fmt.Errorf("validator address is zero")
	}

	// Check operator address
	if gentx.OperatorAddress().IsZero() {
		return fmt.Errorf("operator address is zero")
	}

	// Check BLS public key
	if gentx.BLSPublicKey().IsZero() {
		return fmt.Errorf("BLS public key is zero")
	}

	// Check metadata
	if gentx.Metadata().IsZero() {
		return fmt.Errorf("metadata is zero")
	}

	// Check signature
	if gentx.Signature().IsZero() {
		return fmt.Errorf("signature is zero")
	}

	// Check chain ID
	if gentx.ChainID() == "" {
		return fmt.Errorf("chain ID is empty")
	}

	// Check timestamp
	if gentx.Timestamp().IsZero() {
		return fmt.Errorf("timestamp is zero")
	}

	return nil
}

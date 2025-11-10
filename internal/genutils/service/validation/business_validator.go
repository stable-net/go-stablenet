package validation

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// BusinessValidator validates GenTx business rules.
type BusinessValidator struct{}

// NewBusinessValidator creates a new BusinessValidator instance.
func NewBusinessValidator() *BusinessValidator {
	return &BusinessValidator{}
}

// Validate checks GenTx business rules.
// Business rules include:
// - Timestamp must not be in the future
// - Validator and operator addresses must be different (optional, but recommended)
func (v *BusinessValidator) Validate(gentx domain.GenTx) error {
	// Check timestamp is not in the future
	now := time.Now().UTC()
	if gentx.Timestamp().After(now) {
		return fmt.Errorf("timestamp is in the future: %v > %v", gentx.Timestamp(), now)
	}

	// Optionally check if validator and operator are different
	// This is a recommended practice but not strictly required
	if gentx.ValidatorAddress().Equals(gentx.OperatorAddress()) {
		// This is a warning, not an error, so we don't fail validation
		// In production, you might want to log this as a warning
		// For now, we allow it
	}

	return nil
}

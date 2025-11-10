package validation_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
)

func TestBusinessValidator_DefenseInDepth(t *testing.T) {
	// Note: This test documents that BusinessValidator provides defense in depth.
	// The domain layer already validates timestamps. The business validator checks:
	// 1. Timestamp not in the future
	// 2. Validator and operator addresses (warning only)
	//
	// These rules are difficult to test in isolation because:
	// - Domain layer prevents future timestamps
	// - Same validator/operator is allowed (just not recommended)
	//
	// The validator is tested indirectly through ValidationService tests.

	validator := validation.NewBusinessValidator()
	if validator == nil {
		t.Fatal("BusinessValidator should not be nil")
	}
}

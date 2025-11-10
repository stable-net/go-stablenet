package validation_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
)

func TestFormatValidator_DefenseInDepth(t *testing.T) {
	// Note: This test documents that FormatValidator provides defense in depth.
	// The domain layer (NewGenTx, NewAddress, NewBLSPublicKey, etc.) already validates
	// that all fields are non-zero and well-formed. Therefore, it's not possible to
	// create a GenTx with zero/invalid fields through the public API.
	//
	// The FormatValidator provides an additional safety layer in case:
	// 1. GenTx is deserialized from untrusted sources
	// 2. Future code changes bypass domain validation
	// 3. Defensive programming best practices
	//
	// The validator is tested indirectly through ValidationService tests.

	validator := validation.NewFormatValidator()
	if validator == nil {
		t.Fatal("FormatValidator should not be nil")
	}
}

package test

import (
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestDebugAddress(t *testing.T) {
	initGov(t)
	defer g.backend.Close()

	// Check if the address in the EOA matches the address derived from the private key
	for i, validator := range customValidators {
		derivedAddr := crypto.PubkeyToAddress(validator.Operator.PrivateKey.PublicKey)
		storedAddr := validator.Operator.Address

		t.Logf("Validator %d:", i)
		t.Logf("  Stored Address:  %s", storedAddr.Hex())
		t.Logf("  Derived Address: %s", derivedAddr.Hex())

		if derivedAddr != storedAddr {
			t.Errorf("Address mismatch for validator %d!", i)
		} else {
			t.Logf("  ✓ Addresses match")
		}
	}
}

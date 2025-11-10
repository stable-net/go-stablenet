package validation

import (
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// SignatureValidator validates GenTx signatures.
type SignatureValidator struct {
	cryptoProvider domain.CryptoProvider
}

// NewSignatureValidator creates a new SignatureValidator instance.
func NewSignatureValidator(cryptoProvider domain.CryptoProvider) *SignatureValidator {
	return &SignatureValidator{
		cryptoProvider: cryptoProvider,
	}
}

// Validate verifies the GenTx signature.
// It reconstructs the signed message and verifies the signature against the validator address.
func (v *SignatureValidator) Validate(gentx domain.GenTx) error {
	// Reconstruct the signature data
	sigData := domain.SignatureData{
		ValidatorAddress: gentx.ValidatorAddress(),
		OperatorAddress:  gentx.OperatorAddress(),
		BLSPublicKey:     gentx.BLSPublicKey(),
		ChainID:          gentx.ChainID(),
		Timestamp:        gentx.Timestamp().Unix(),
	}

	// Get the message that should have been signed
	message := sigData.Bytes()

	// Verify the signature
	valid, err := v.cryptoProvider.Verify(gentx.ValidatorAddress(), message, gentx.Signature())
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}

	if !valid {
		return fmt.Errorf("signature verification failed: signature does not match validator address")
	}

	return nil
}

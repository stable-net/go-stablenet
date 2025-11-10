package validation

import (
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// ValidationService provides comprehensive GenTx validation.
// It orchestrates signature, format, and business rule validation.
type ValidationService struct {
	cryptoProvider     domain.CryptoProvider
	signatureValidator *SignatureValidator
	formatValidator    *FormatValidator
	businessValidator  *BusinessValidator
}

// NewValidationService creates a new ValidationService instance.
func NewValidationService(cryptoProvider domain.CryptoProvider) *ValidationService {
	return &ValidationService{
		cryptoProvider:     cryptoProvider,
		signatureValidator: NewSignatureValidator(cryptoProvider),
		formatValidator:    NewFormatValidator(),
		businessValidator:  NewBusinessValidator(),
	}
}

// Validate performs comprehensive validation on a GenTx.
// It checks signature, format, and business rules in that order.
// Returns error if any validation fails.
func (s *ValidationService) Validate(gentx domain.GenTx) error {
	// 1. Validate signature (most important)
	if err := s.signatureValidator.Validate(gentx); err != nil {
		return fmt.Errorf("signature validation failed: %w", err)
	}

	// 2. Validate format
	if err := s.formatValidator.Validate(gentx); err != nil {
		return fmt.Errorf("format validation failed: %w", err)
	}

	// 3. Validate business rules
	if err := s.businessValidator.Validate(gentx); err != nil {
		return fmt.Errorf("business rules validation failed: %w", err)
	}

	return nil
}

// ValidateSignature validates only the GenTx signature.
func (s *ValidationService) ValidateSignature(gentx domain.GenTx) error {
	return s.signatureValidator.Validate(gentx)
}

// ValidateFormat validates only the GenTx format.
func (s *ValidationService) ValidateFormat(gentx domain.GenTx) error {
	return s.formatValidator.Validate(gentx)
}

// ValidateBusinessRules validates only the GenTx business rules.
func (s *ValidationService) ValidateBusinessRules(gentx domain.GenTx) error {
	return s.businessValidator.Validate(gentx)
}

package application

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// ValidateGenTxUseCase implements the use case for validating a GenTx.
//
// The use case performs the following operations:
//  1. Validates the input parameters
//  2. Either validates a GenTx directly or loads it from repository
//  3. Delegates validation to the ValidationService
//  4. Returns validation result
//
// This use case provides application-level validation functionality,
// orchestrating repository and validation service operations.
type ValidateGenTxUseCase struct {
	repository Repository
	validator  Validator
}

// NewValidateGenTxUseCase creates a new ValidateGenTxUseCase instance.
//
// Parameters:
//   - repository: Repository for retrieving GenTx
//   - validator: Validator for validating GenTx
//
// Returns a new ValidateGenTxUseCase ready for use.
func NewValidateGenTxUseCase(
	repository Repository,
	validator Validator,
) *ValidateGenTxUseCase {
	return &ValidateGenTxUseCase{
		repository: repository,
		validator:  validator,
	}
}

// Validate validates a GenTx directly.
//
// The method performs comprehensive validation including:
//   - Input validation (non-empty GenTx)
//   - Signature validation
//   - Format validation
//   - Business rules validation
//
// Parameters:
//   - gentx: The GenTx to validate
//
// Returns:
//   - error: Error if validation fails, nil if valid
//
// Errors:
//   - Empty GenTx error
//   - Signature validation errors
//   - Format validation errors
//   - Business rules validation errors
//
// Example:
//
//	err := useCase.Validate(gentx)
//	if err != nil {
//	    log.Fatal("GenTx validation failed:", err)
//	}
//	fmt.Println("GenTx is valid")
func (uc *ValidateGenTxUseCase) Validate(gentx domain.GenTx) error {
	// Validate input
	if gentx.ValidatorAddress().IsZero() {
		return errors.New("cannot validate empty GenTx")
	}

	// Delegate to validation service
	if err := uc.validator.Validate(gentx); err != nil {
		return fmt.Errorf("GenTx validation failed: %w", err)
	}

	return nil
}

// ValidateByAddress loads a GenTx from repository and validates it.
//
// The method performs the following operations:
//  1. Validates the validator address
//  2. Loads the GenTx from repository
//  3. Validates the loaded GenTx
//
// Parameters:
//   - validatorAddr: The validator address to load and validate
//
// Returns:
//   - error: Error if validation fails or GenTx not found, nil if valid
//
// Errors:
//   - Zero address error
//   - Repository errors (not found, read failure)
//   - Validation errors
//
// Example:
//
//	err := useCase.ValidateByAddress(validatorAddr)
//	if err != nil {
//	    log.Fatal("GenTx validation failed:", err)
//	}
//	fmt.Println("GenTx is valid")
func (uc *ValidateGenTxUseCase) ValidateByAddress(validatorAddr domain.Address) error {
	// Validate address
	if validatorAddr.IsZero() {
		return errors.New("validator address cannot be zero")
	}

	// Load GenTx from repository
	gentx, err := uc.repository.FindByValidator(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to load GenTx: %w", err)
	}

	// Validate the loaded GenTx
	if err := uc.Validate(gentx); err != nil {
		return err
	}

	return nil
}

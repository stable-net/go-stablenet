package genesis

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/params"
)

// ContractInjector provides utilities for system contract configuration.
//
// This helper validates and prepares system contract configurations
// for genesis block injection.
type ContractInjector struct{}

// NewContractInjector creates a new ContractInjector instance.
func NewContractInjector() *ContractInjector {
	return &ContractInjector{}
}

// ValidateSystemContracts validates system contracts configuration.
//
// Validates:
//   - All required contracts are present
//   - Addresses are non-zero
//   - Versions are non-empty
//
// Parameters:
//   - contracts: System contracts configuration to validate
//
// Returns:
//   - error: Validation error if configuration is invalid
func (ci *ContractInjector) ValidateSystemContracts(contracts *params.SystemContracts) error {
	if contracts == nil {
		return errors.New("system contracts configuration is nil")
	}

	// Validate GovValidator
	if contracts.GovValidator == nil {
		return errors.New("GovValidator contract is missing")
	}
	if err := ci.validateContract("GovValidator", contracts.GovValidator); err != nil {
		return err
	}

	// Validate NativeCoinAdapter
	if contracts.NativeCoinAdapter == nil {
		return errors.New("NativeCoinAdapter contract is missing")
	}
	if err := ci.validateContract("NativeCoinAdapter", contracts.NativeCoinAdapter); err != nil {
		return err
	}

	// Validate GovMasterMinter
	if contracts.GovMasterMinter == nil {
		return errors.New("GovMasterMinter contract is missing")
	}
	if err := ci.validateContract("GovMasterMinter", contracts.GovMasterMinter); err != nil {
		return err
	}

	// Validate GovMinter
	if contracts.GovMinter == nil {
		return errors.New("GovMinter contract is missing")
	}
	if err := ci.validateContract("GovMinter", contracts.GovMinter); err != nil {
		return err
	}

	return nil
}

// validateContract validates a single system contract.
func (ci *ContractInjector) validateContract(name string, contract *params.SystemContract) error {
	if contract.Address.Hex() == "0x0000000000000000000000000000000000000000" {
		return fmt.Errorf("%s: address is zero", name)
	}

	if contract.Version == "" {
		return fmt.Errorf("%s: version is empty", name)
	}

	return nil
}

// MergeSystemContracts merges custom contracts with defaults.
//
// If custom contracts are provided, they override the defaults.
// Missing contracts in custom config use default values.
//
// Parameters:
//   - custom: Custom system contracts (may be partial)
//   - defaults: Default system contracts
//
// Returns merged system contracts configuration.
func (ci *ContractInjector) MergeSystemContracts(custom, defaults *params.SystemContracts) *params.SystemContracts {
	if custom == nil {
		return defaults
	}

	merged := &params.SystemContracts{
		GovValidator:      custom.GovValidator,
		NativeCoinAdapter: custom.NativeCoinAdapter,
		GovMasterMinter:   custom.GovMasterMinter,
		GovMinter:         custom.GovMinter,
	}

	// Use defaults for missing contracts
	if merged.GovValidator == nil && defaults != nil {
		merged.GovValidator = defaults.GovValidator
	}
	if merged.NativeCoinAdapter == nil && defaults != nil {
		merged.NativeCoinAdapter = defaults.NativeCoinAdapter
	}
	if merged.GovMasterMinter == nil && defaults != nil {
		merged.GovMasterMinter = defaults.GovMasterMinter
	}
	if merged.GovMinter == nil && defaults != nil {
		merged.GovMinter = defaults.GovMinter
	}

	return merged
}

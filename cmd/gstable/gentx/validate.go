// Copyright 2025 The go-stablenet Authors
// This file is part of go-stablenet.
//
// go-stablenet is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-stablenet. If not, see <http://www.gnu.org/licenses/>.

package gentx

import (
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/application"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
	"github.com/urfave/cli/v2"
)

// ValidateCommand implements the gentx validate command
var ValidateCommand = &cli.Command{
	Name:      "validate",
	Usage:     "Validate genesis transaction(s)",
	ArgsUsage: "",
	Description: `
The gentx validate command validates one or all genesis transactions.
It performs comprehensive validation including signature verification,
format validation, and business rules validation.

Example:
  gstable gentx validate --input-dir ./gentx
  gstable gentx validate --input-dir ./gentx --validator-addr 0x1234...
  gstable gentx validate --input-dir ./gentx --chain-id stable-testnet-1
`,
	Flags:  ValidateFlags,
	Action: runValidateCommand,
}

func runValidateCommand(ctx *cli.Context) error {
	// Get flags
	inputDir := ctx.String(InputDirFlag.Name)
	validatorAddrStr := ctx.String(ValidatorAddressFlag.Name)
	chainID := ctx.String(ChainIDFlag.Name)

	// Initialize dependencies
	repo := repository.NewFileRepository(inputDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Create use case
	validateUseCase := application.NewValidateGenTxUseCase(repo, validator)

	// Validate single validator if specified
	if validatorAddrStr != "" {
		return validateSingleGenTx(validateUseCase, validatorAddrStr)
	}

	// Otherwise validate all GenTxs
	return validateAllGenTxs(validateUseCase, repo, chainID)
}

func validateSingleGenTx(useCase *application.ValidateGenTxUseCase, validatorAddrStr string) error {
	// Parse validator address
	validatorAddr, err := domain.NewAddress(validatorAddrStr)
	if err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}

	// Validate
	err = useCase.ValidateByAddress(validatorAddr)
	if err != nil {
		fmt.Printf("❌ Validation failed for %s: %v\n", validatorAddrStr, err)
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("✅ GenTx for validator %s is valid\n", validatorAddrStr)
	return nil
}

func validateAllGenTxs(useCase *application.ValidateGenTxUseCase, repo *repository.FileRepository, chainID string) error {
	// Load all GenTxs
	allGenTxs, err := repo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load GenTxs: %w", err)
	}

	if len(allGenTxs) == 0 {
		fmt.Println("No GenTxs found to validate")
		return nil
	}

	// Filter by chain ID if specified
	var gentxsToValidate []domain.GenTx
	if chainID != "" {
		for _, gentx := range allGenTxs {
			if gentx.ChainID() == chainID {
				gentxsToValidate = append(gentxsToValidate, gentx)
			}
		}
	} else {
		gentxsToValidate = allGenTxs
	}

	if len(gentxsToValidate) == 0 {
		fmt.Printf("No GenTxs found for chain ID: %s\n", chainID)
		return nil
	}

	// Validate each GenTx
	validCount := 0
	invalidCount := 0

	fmt.Printf("Validating %d GenTx(s)...\n\n", len(gentxsToValidate))

	for _, gentx := range gentxsToValidate {
		err := useCase.Validate(gentx)
		if err != nil {
			fmt.Printf("❌ %s: %v\n", gentx.ValidatorAddress().String(), err)
			invalidCount++
		} else {
			fmt.Printf("✅ %s: Valid\n", gentx.ValidatorAddress().String())
			validCount++
		}
	}

	// Print summary
	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("  Total: %d\n", len(gentxsToValidate))
	fmt.Printf("  Valid: %d\n", validCount)
	fmt.Printf("  Invalid: %d\n", invalidCount)

	if invalidCount > 0 {
		return fmt.Errorf("validation failed: %d invalid GenTx(s)", invalidCount)
	}

	return nil
}

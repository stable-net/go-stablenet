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
	"os"
	"time"

	"github.com/ethereum/go-ethereum/internal/genutils/application"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
	"github.com/urfave/cli/v2"
)

// CreateCommand implements the gentx create command
var CreateCommand = &cli.Command{
	Name:      "create",
	Usage:     "Create a new genesis transaction (gentx)",
	ArgsUsage: "",
	Description: `
The gentx create command creates a new genesis transaction for a validator.
It generates a signed transaction that will be included in the genesis block
to initialize the validator set.

Example:
  gstable gentx create --validator-key ./validator.key --operator-addr 0x1234... --name "My Validator" --chain-id stable-testnet-1
`,
	Flags:  CreateFlags,
	Action: runCreateCommand,
}

func runCreateCommand(ctx *cli.Context) error {
	// Get flags
	validatorKeyPath := ctx.String(ValidatorKeyFlag.Name)
	operatorAddrStr := ctx.String(OperatorAddrFlag.Name)
	validatorName := ctx.String(ValidatorNameFlag.Name)
	validatorDescription := ctx.String(ValidatorDescriptionFlag.Name)
	validatorWebsite := ctx.String(ValidatorWebsiteFlag.Name)
	chainID := ctx.String(ChainIDFlag.Name)
	outputDir := ctx.String(OutputDirFlag.Name)

	// Validate required flags
	if validatorKeyPath == "" {
		return fmt.Errorf("--validator-key is required")
	}
	if operatorAddrStr == "" {
		return fmt.Errorf("--operator-addr is required")
	}
	if validatorName == "" {
		return fmt.Errorf("--name is required")
	}
	if chainID == "" {
		return fmt.Errorf("--chain-id is required")
	}

	// Load validator key
	validatorKey, err := os.ReadFile(validatorKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read validator key file: %w", err)
	}

	// Parse operator address
	operatorAddr, err := domain.NewAddress(operatorAddrStr)
	if err != nil {
		return fmt.Errorf("invalid operator address: %w", err)
	}

	// Create validator metadata
	metadata, err := domain.NewValidatorMetadata(validatorName, validatorDescription, validatorWebsite)
	if err != nil {
		return fmt.Errorf("invalid validator metadata: %w", err)
	}

	// Initialize dependencies
	repo := repository.NewFileRepository(outputDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Create use case
	createUseCase := application.NewCreateGenTxUseCase(repo, cryptoProvider, validator)

	// Create request
	request := &application.CreateGenTxRequest{
		ValidatorKey: validatorKey,
		OperatorAddr: operatorAddr,
		Metadata:     metadata,
		ChainID:      chainID,
		Timestamp:    time.Now().UTC(),
	}

	// Execute use case
	gentx, err := createUseCase.Execute(request)
	if err != nil {
		return fmt.Errorf("failed to create gentx: %w", err)
	}

	// Print success message
	fmt.Printf("Successfully created gentx for validator: %s\n", gentx.ValidatorAddress().String())
	fmt.Printf("Operator address: %s\n", gentx.OperatorAddress().String())
	fmt.Printf("Chain ID: %s\n", gentx.ChainID())
	fmt.Printf("Output directory: %s\n", outputDir)

	return nil
}

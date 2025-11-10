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
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/ethereum/go-ethereum/internal/genutils/service/validation"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
	"github.com/urfave/cli/v2"
)

// CollectCommand implements the gentx collect command
var CollectCommand = &cli.Command{
	Name:      "collect",
	Usage:     "Collect and validate all genesis transactions",
	ArgsUsage: "",
	Description: `
The gentx collect command collects all genesis transactions from a directory,
validates them, and provides a comprehensive report. This command is typically
used before genesis creation to ensure all GenTxs are valid.

Example:
  gstable gentx collect --input-dir ./gentx
  gstable gentx collect --input-dir ./gentx --chain-id stable-testnet-1
`,
	Flags:  CollectFlags,
	Action: runCollectCommand,
}

func runCollectCommand(ctx *cli.Context) error {
	// Get flags
	inputDir := ctx.String(InputDirFlag.Name)
	chainID := ctx.String(ChainIDFlag.Name)

	// Initialize dependencies
	repo := repository.NewFileRepository(inputDir)
	cryptoProvider := ethcrypto.NewEthereumProvider()
	validator := validation.NewValidationService(cryptoProvider)

	// Create use case
	collectUseCase := application.NewCollectGenTxsUseCase(repo, validator)

	// Create request
	request := &application.CollectGenTxsRequest{
		ChainID: chainID,
	}

	// Execute use case
	result, err := collectUseCase.Execute(request)
	if err != nil {
		return fmt.Errorf("failed to collect GenTxs: %w", err)
	}

	// Print results
	printCollectionResults(result, chainID)

	// Return error if any invalid GenTxs
	if result.InvalidCount > 0 {
		return fmt.Errorf("collection failed: %d invalid GenTx(s)", result.InvalidCount)
	}

	return nil
}

func printCollectionResults(result *application.CollectGenTxsResult, chainID string) {
	fmt.Println("==============================================")
	fmt.Println("       GenTx Collection Report")
	fmt.Println("==============================================")

	if chainID != "" {
		fmt.Printf("Chain ID Filter: %s\n", chainID)
	}
	fmt.Println()

	// Summary
	fmt.Printf("Total GenTxs: %d\n", result.TotalCount)
	fmt.Printf("Valid GenTxs: %d\n", result.ValidCount)
	fmt.Printf("Invalid GenTxs: %d\n", result.InvalidCount)
	fmt.Println()

	// Valid GenTxs
	if result.ValidCount > 0 {
		fmt.Println("✅ Valid GenTxs:")
		fmt.Println("----------------------------------------------")
		for i, gentx := range result.ValidGenTxs {
			fmt.Printf("%d. Validator: %s\n", i+1, gentx.ValidatorAddress().String())
			fmt.Printf("   Operator:  %s\n", gentx.OperatorAddress().String())
			fmt.Printf("   Name:      %s\n", gentx.Metadata().Name())
			fmt.Printf("   Chain ID:  %s\n", gentx.ChainID())
			fmt.Println()
		}
	}

	// Invalid GenTxs
	if result.InvalidCount > 0 {
		fmt.Println("❌ Invalid GenTxs:")
		fmt.Println("----------------------------------------------")
		for i, invalid := range result.InvalidGenTxs {
			fmt.Printf("%d. Validator: %s\n", i+1, invalid.GenTx.ValidatorAddress().String())
			fmt.Printf("   Error:     %s\n", invalid.Error)
			fmt.Println()
		}
	}

	fmt.Println("==============================================")

	if result.InvalidCount == 0 {
		fmt.Println("✅ All GenTxs are valid and ready for genesis!")
	} else {
		fmt.Println("❌ Please fix the invalid GenTxs before creating genesis.")
	}
	fmt.Println("==============================================")
}

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
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/repository"
	"github.com/urfave/cli/v2"
)

// InspectCommand implements the gentx inspect command
var InspectCommand = &cli.Command{
	Name:      "inspect",
	Usage:     "Inspect genesis transaction details",
	ArgsUsage: "",
	Description: `
The gentx inspect command displays detailed information about a specific
genesis transaction or all genesis transactions in a directory.

Example:
  gstable gentx inspect --input-dir ./gentx --validator-addr 0x1234...
  gstable gentx inspect --input-dir ./gentx
`,
	Flags:  InspectFlags,
	Action: runInspectCommand,
}

func runInspectCommand(ctx *cli.Context) error {
	// Get flags
	inputDir := ctx.String(InputDirFlag.Name)
	validatorAddrStr := ctx.String(ValidatorAddressFlag.Name)

	// Initialize repository
	repo := repository.NewFileRepository(inputDir)

	// Inspect single validator if specified
	if validatorAddrStr != "" {
		return inspectSingleGenTx(repo, validatorAddrStr)
	}

	// Otherwise inspect all GenTxs
	return inspectAllGenTxs(repo)
}

func inspectSingleGenTx(repo *repository.FileRepository, validatorAddrStr string) error {
	// Parse validator address
	validatorAddr, err := domain.NewAddress(validatorAddrStr)
	if err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}

	// Load GenTx
	gentx, err := repo.FindByValidator(validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to load GenTx: %w", err)
	}

	// Print details
	printGenTxDetails(gentx)

	return nil
}

func inspectAllGenTxs(repo *repository.FileRepository) error {
	// Load all GenTxs
	allGenTxs, err := repo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load GenTxs: %w", err)
	}

	if len(allGenTxs) == 0 {
		fmt.Println("No GenTxs found to inspect")
		return nil
	}

	fmt.Printf("Found %d GenTx(s)\n\n", len(allGenTxs))

	for i, gentx := range allGenTxs {
		fmt.Printf("═══════════════════════════════════════════════\n")
		fmt.Printf("GenTx #%d\n", i+1)
		fmt.Printf("═══════════════════════════════════════════════\n")
		printGenTxDetails(gentx)
		fmt.Println()
	}

	return nil
}

func printGenTxDetails(gentx domain.GenTx) {
	fmt.Println("Validator Information:")
	fmt.Printf("  Address:     %s\n", gentx.ValidatorAddress().String())
	fmt.Printf("  Name:        %s\n", gentx.Metadata().Name())
	fmt.Printf("  Description: %s\n", gentx.Metadata().Description())
	if gentx.Metadata().Website() != "" {
		fmt.Printf("  Website:     %s\n", gentx.Metadata().Website())
	}
	fmt.Println()

	fmt.Println("Operator Information:")
	fmt.Printf("  Address:     %s\n", gentx.OperatorAddress().String())
	fmt.Println()

	fmt.Println("BLS Public Key:")
	blsBytes := gentx.BLSPublicKey().Bytes()
	fmt.Printf("  %s\n", hex.EncodeToString(blsBytes))
	fmt.Println()

	fmt.Println("Network Information:")
	fmt.Printf("  Chain ID:    %s\n", gentx.ChainID())
	fmt.Printf("  Timestamp:   %s\n", gentx.Timestamp().UTC().Format("2006-01-02 15:04:05 UTC"))
	fmt.Println()

	fmt.Println("Signature:")
	sigBytes := gentx.Signature().Bytes()
	// Print signature in chunks for readability
	chunkSize := 64
	for i := 0; i < len(sigBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(sigBytes) {
			end = len(sigBytes)
		}
		fmt.Printf("  %s\n", hex.EncodeToString(sigBytes[i:end]))
	}
}

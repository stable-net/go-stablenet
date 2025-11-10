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
	"github.com/urfave/cli/v2"
)

var (
	// ValidatorKeyFlag specifies the validator private key file path
	ValidatorKeyFlag = &cli.StringFlag{
		Name:     "validator-key",
		Usage:    "Path to the validator private key file",
		Required: false,
	}

	// OperatorAddrFlag specifies the operator address
	OperatorAddrFlag = &cli.StringFlag{
		Name:     "operator-addr",
		Usage:    "Operator address (can be multisig)",
		Required: false,
	}

	// ValidatorNameFlag specifies the validator name
	ValidatorNameFlag = &cli.StringFlag{
		Name:     "name",
		Usage:    "Validator name (1-70 characters)",
		Required: false,
	}

	// ValidatorDescriptionFlag specifies the validator description
	ValidatorDescriptionFlag = &cli.StringFlag{
		Name:  "description",
		Usage: "Validator description (max 280 characters)",
	}

	// ValidatorWebsiteFlag specifies the validator website
	ValidatorWebsiteFlag = &cli.StringFlag{
		Name:  "website",
		Usage: "Validator website URL",
	}

	// ChainIDFlag specifies the chain ID
	ChainIDFlag = &cli.StringFlag{
		Name:     "chain-id",
		Usage:    "Network chain ID",
		Required: false,
	}

	// OutputDirFlag specifies the output directory for gentx files
	OutputDirFlag = &cli.StringFlag{
		Name:  "output-dir",
		Usage: "Output directory for gentx files",
		Value: "./gentx",
	}

	// InputDirFlag specifies the input directory for gentx files
	InputDirFlag = &cli.StringFlag{
		Name:  "input-dir",
		Usage: "Input directory containing gentx files",
		Value: "./gentx",
	}

	// ValidatorAddressFlag specifies the validator address to validate
	ValidatorAddressFlag = &cli.StringFlag{
		Name:  "validator-addr",
		Usage: "Validator address to validate",
	}
)

// Common flag groups for gentx commands
var (
	// CreateFlags are flags for the create command
	CreateFlags = []cli.Flag{
		ValidatorKeyFlag,
		OperatorAddrFlag,
		ValidatorNameFlag,
		ValidatorDescriptionFlag,
		ValidatorWebsiteFlag,
		ChainIDFlag,
		OutputDirFlag,
	}

	// ValidateFlags are flags for the validate command
	ValidateFlags = []cli.Flag{
		InputDirFlag,
		ValidatorAddressFlag,
		ChainIDFlag,
	}

	// CollectFlags are flags for the collect command
	CollectFlags = []cli.Flag{
		InputDirFlag,
		ChainIDFlag,
	}

	// InspectFlags are flags for the inspect command
	InspectFlags = []cli.Flag{
		InputDirFlag,
		ValidatorAddressFlag,
	}
)

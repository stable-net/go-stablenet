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

package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// hexPrefix is the expected prefix for hex-encoded strings
	hexPrefix = "0x"

	// addressLength is the expected total length of an Ethereum address string (0x + 40 hex chars)
	addressLength = 42

	// blsPublicKeyBytes is the size of a BLS12-381 G2 public key in bytes
	blsPublicKeyBytes = 192

	// blsPublicKeyHexChars is the number of hex characters needed to represent a BLS public key
	blsPublicKeyHexChars = blsPublicKeyBytes * 2 // 384

	// blsPublicKeyLength is the expected total length of a BLS public key string (0x + 384 hex chars)
	blsPublicKeyLength = len(hexPrefix) + blsPublicKeyHexChars // 386
)

// genTxData represents the minimal structure of a GenTx JSON file
// that we need to extract validator information.
type genTxData struct {
	ValidatorAddress string `json:"validator_address"`
	OperatorAddress  string `json:"operator_address"`
	BLSPublicKey     string `json:"bls_public_key"`
}

// ProcessGenTxs processes the genesis transactions and extracts validator information
// to populate the anzeon.init section of the genesis config.
//
// This function:
//  1. Parses each GenTx JSON from AnzeonGenTxs field
//  2. Extracts validator_address, operator_address, and bls_public_key
//  3. Validates the extracted data
//  4. Checks for duplicate validators
//  5. Updates the anzeon.init.validators, anzeon.init.operators, and anzeon.init.blsPublicKeys arrays
//
// Returns an error if:
//   - Anzeon config is not properly initialized
//   - Any GenTx has invalid JSON format
//   - Required fields are missing
//   - Validator addresses, operator addresses, or BLS keys are invalid
//   - Duplicate validators are detected
func (g *Genesis) ProcessGenTxs() error {
	// Validate anzeon config structure
	if g.Config == nil || g.Config.Anzeon == nil {
		return fmt.Errorf("anzeon config is not initialized")
	}
	if g.Config.Anzeon.Init == nil {
		return fmt.Errorf("anzeon init config is not initialized")
	}

	// Empty gentxs is valid - nothing to process
	if len(g.AnzeonGenTxs) == 0 {
		g.Config.Anzeon.Init.Validators = []common.Address{}
		g.Config.Anzeon.Init.Operators = []common.Address{}
		g.Config.Anzeon.Init.BLSPublicKeys = []string{}
		return nil
	}

	// Prepare result slices
	validators := make([]common.Address, 0, len(g.AnzeonGenTxs))
	operators := make([]common.Address, 0, len(g.AnzeonGenTxs))
	blsKeys := make([]string, 0, len(g.AnzeonGenTxs))
	seenValidators := make(map[common.Address]bool)

	// Process each GenTx
	for i, gentxRaw := range g.AnzeonGenTxs {
		var gentx genTxData
		if err := json.Unmarshal(gentxRaw, &gentx); err != nil {
			return fmt.Errorf("failed to unmarshal gentx at index %d: %w", i, err)
		}

		// Validate and extract validator address
		validatorAddr, err := validateAndParseAddress(gentx.ValidatorAddress, "validator_address")
		if err != nil {
			return fmt.Errorf("invalid validator address at index %d: %w", i, err)
		}

		// Validate and extract operator address
		operatorAddr, err := validateAndParseAddress(gentx.OperatorAddress, "operator_address")
		if err != nil {
			return fmt.Errorf("invalid operator address at index %d: %w", i, err)
		}

		// Check for duplicate validators
		if seenValidators[validatorAddr] {
			return fmt.Errorf("duplicate validator address detected at index %d: %s", i, validatorAddr.Hex())
		}
		seenValidators[validatorAddr] = true

		// Validate BLS public key
		if err := validateBLSPublicKey(gentx.BLSPublicKey); err != nil {
			return fmt.Errorf("invalid BLS public key at index %d: %w", i, err)
		}

		// Add to result slices
		validators = append(validators, validatorAddr)
		operators = append(operators, operatorAddr)
		blsKeys = append(blsKeys, gentx.BLSPublicKey)
	}

	// Update genesis config
	g.Config.Anzeon.Init.Validators = validators
	g.Config.Anzeon.Init.Operators = operators
	g.Config.Anzeon.Init.BLSPublicKeys = blsKeys

	return nil
}

// validateAndParseAddress validates and parses an Ethereum address string.
// fieldName is used in error messages for context (e.g., "validator_address", "operator_address").
// Returns error if:
//   - Address is empty
//   - Address doesn't have 0x prefix
//   - Address is not valid hex
//   - Address length is incorrect (must be 40 hex chars + 0x prefix)
//   - Address is zero address
func validateAndParseAddress(addrStr string, fieldName string) (common.Address, error) {
	// Check empty
	if addrStr == "" {
		return common.Address{}, fmt.Errorf("%s field is required", fieldName)
	}

	// Check 0x prefix
	if !strings.HasPrefix(addrStr, hexPrefix) {
		return common.Address{}, fmt.Errorf("%s must have 0x prefix, got: %s", fieldName, addrStr)
	}

	// Check length
	if len(addrStr) != addressLength {
		return common.Address{}, fmt.Errorf("%s must be %d characters (0x + 40 hex), got %d characters", fieldName, addressLength, len(addrStr))
	}

	// Parse using common.HexToAddress
	// This function is lenient and will not error on invalid hex,
	// so we need to validate the result
	addr := common.HexToAddress(addrStr)

	// Check if it's zero address
	if addr == (common.Address{}) {
		return common.Address{}, fmt.Errorf("%s cannot be zero address", fieldName)
	}

	// Validate hex characters by comparing the parsed address back to string
	// common.HexToAddress will silently accept invalid hex, so we verify
	expectedAddr := strings.ToLower(addrStr)
	actualAddr := strings.ToLower(addr.Hex())
	if expectedAddr != actualAddr {
		return common.Address{}, fmt.Errorf("%s contains invalid hex characters", fieldName)
	}

	return addr, nil
}

// validateBLSPublicKey validates a BLS public key string.
// BLS12-381 G2 public key format:
//   - Must have 0x prefix
//   - Must be exactly 192 bytes (384 hex characters + 0x prefix = 386 total)
//   - Must be valid hex
//
// Returns error if validation fails.
func validateBLSPublicKey(blsKey string) error {
	// Check empty
	if blsKey == "" {
		return fmt.Errorf("bls_public_key field is required")
	}

	// Check 0x prefix
	if !strings.HasPrefix(blsKey, hexPrefix) {
		return fmt.Errorf("BLS public key must have 0x prefix")
	}

	// Check length
	if len(blsKey) != blsPublicKeyLength {
		return fmt.Errorf("BLS public key must be %d characters (0x + %d hex), got %d characters",
			blsPublicKeyLength, blsPublicKeyHexChars, len(blsKey))
	}

	// Validate hex characters
	hexPart := blsKey[len(hexPrefix):] // Remove 0x prefix
	for i, c := range hexPart {
		if !isHexChar(c) {
			return fmt.Errorf("BLS public key contains invalid hex character at position %d: %c", i+len(hexPrefix), c)
		}
	}

	return nil
}

// isHexChar checks if a character is a valid hexadecimal character (0-9, a-f, A-F).
func isHexChar(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

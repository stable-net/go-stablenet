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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test BLS keys with correct length (384 hex chars + 0x prefix = 386 total)
const (
	validBLSKey1 = "0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6"
	validBLSKey2 = "0xb1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6"
)

func TestProcessGenTxs_Success(t *testing.T) {
	// Arrange
	genTx1 := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC",
		"operator_address":  "0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B",
		"bls_public_key":    validBLSKey1,
		"signature":         "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd1b",
		"metadata": map[string]interface{}{
			"name":        "Validator One",
			"description": "Test validator",
		},
		"timestamp": 1704067200,
	}

	genTx2 := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x1234567890123456789012345678901234567890",
		"operator_address":  "0x9876543210987654321098765432109876543210",
		"bls_public_key":    validBLSKey2,
		"signature":         "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab1b",
		"metadata": map[string]interface{}{
			"name":        "Validator Two",
			"description": "Another test validator",
		},
		"timestamp": 1704067300,
	}

	genTx1JSON, err := json.Marshal(genTx1)
	require.NoError(t, err)
	genTx2JSON, err := json.Marshal(genTx2)
	require.NoError(t, err)

	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{genTx1JSON, genTx2JSON},
	}

	// Act
	err = genesis.ProcessGenTxs()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, genesis.Config.Anzeon.Init)
	assert.Len(t, genesis.Config.Anzeon.Init.Validators, 2)
	assert.Len(t, genesis.Config.Anzeon.Init.BLSPublicKeys, 2)

	// Verify first validator
	assert.Equal(t, common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC"), genesis.Config.Anzeon.Init.Validators[0])
	assert.Equal(t, validBLSKey1, genesis.Config.Anzeon.Init.BLSPublicKeys[0])

	// Verify second validator
	assert.Equal(t, common.HexToAddress("0x1234567890123456789012345678901234567890"), genesis.Config.Anzeon.Init.Validators[1])
	assert.Equal(t, validBLSKey2, genesis.Config.Anzeon.Init.BLSPublicKeys[1])
}

func TestProcessGenTxs_EmptyInput(t *testing.T) {
	// Arrange
	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{},
	}

	// Act
	err := genesis.ProcessGenTxs()

	// Assert - Empty input should be valid (no gentxs to process)
	require.NoError(t, err)
	assert.Len(t, genesis.Config.Anzeon.Init.Validators, 0)
	assert.Len(t, genesis.Config.Anzeon.Init.BLSPublicKeys, 0)
}

func TestProcessGenTxs_NilAnzeonConfig(t *testing.T) {
	// Arrange
	genesis := &Genesis{
		Config:       &params.ChainConfig{},
		AnzeonGenTxs: []json.RawMessage{json.RawMessage(`{"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}`)},
	}

	// Act
	err := genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anzeon config")
}

func TestProcessGenTxs_NilInit(t *testing.T) {
	// Arrange
	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{},
		},
		AnzeonGenTxs: []json.RawMessage{json.RawMessage(`{"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}`)},
	}

	// Act
	err := genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "init")
}

func TestProcessGenTxs_InvalidJSON(t *testing.T) {
	// Arrange
	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{json.RawMessage(`{invalid json}`)},
	}

	// Act
	err := genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestProcessGenTxs_MissingValidatorAddress(t *testing.T) {
	// Arrange
	genTx := map[string]interface{}{
		"version":        "1.0",
		"chain_id":       "stablenet-1",
		"bls_public_key": "0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
		"timestamp":      1704067200,
	}
	genTxJSON, err := json.Marshal(genTx)
	require.NoError(t, err)

	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{genTxJSON},
	}

	// Act
	err = genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validator_address")
}

func TestProcessGenTxs_MissingBLSPublicKey(t *testing.T) {
	// Arrange
	genTx := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC",
		"timestamp":         1704067200,
	}
	genTxJSON, err := json.Marshal(genTx)
	require.NoError(t, err)

	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{genTxJSON},
	}

	// Act
	err = genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bls_public_key")
}

func TestProcessGenTxs_InvalidValidatorAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{
			name:    "not hex string",
			address: "not-a-hex-address",
		},
		{
			name:    "missing 0x prefix",
			address: "742d35Cc6634C0532925a3b844Bc9e7595f0bEbC",
		},
		{
			name:    "wrong length",
			address: "0x742d35Cc",
		},
		{
			name:    "empty string",
			address: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			genTx := map[string]interface{}{
				"version":           "1.0",
				"chain_id":          "stablenet-1",
				"validator_address": tt.address,
				"bls_public_key":    "0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
				"timestamp":         1704067200,
			}
			genTxJSON, err := json.Marshal(genTx)
			require.NoError(t, err)

			genesis := &Genesis{
				Config: &params.ChainConfig{
					Anzeon: &params.AnzeonConfig{
						Init: &params.WBFTInit{},
					},
				},
				AnzeonGenTxs: []json.RawMessage{genTxJSON},
			}

			// Act
			err = genesis.ProcessGenTxs()

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "validator address")
		})
	}
}

func TestProcessGenTxs_InvalidBLSPublicKey(t *testing.T) {
	tests := []struct {
		name   string
		blsKey string
	}{
		{
			name:   "not hex string",
			blsKey: "not-a-hex-key",
		},
		{
			name:   "missing 0x prefix",
			blsKey: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		},
		{
			name:   "wrong length (too short)",
			blsKey: "0xa1b2c3d4e5f6",
		},
		{
			name:   "wrong length (too long)",
			blsKey: "0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8",
		},
		{
			name:   "empty string",
			blsKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			genTx := map[string]interface{}{
				"version":           "1.0",
				"chain_id":          "stablenet-1",
				"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC",
				"bls_public_key":    tt.blsKey,
				"timestamp":         1704067200,
			}
			genTxJSON, err := json.Marshal(genTx)
			require.NoError(t, err)

			genesis := &Genesis{
				Config: &params.ChainConfig{
					Anzeon: &params.AnzeonConfig{
						Init: &params.WBFTInit{},
					},
				},
				AnzeonGenTxs: []json.RawMessage{genTxJSON},
			}

			// Act
			err = genesis.ProcessGenTxs()

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "BLS")
		})
	}
}

func TestProcessGenTxs_DuplicateValidators(t *testing.T) {
	// Arrange - Same validator address appears twice
	genTx1 := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC",
		"bls_public_key":    validBLSKey1,
		"timestamp":         1704067200,
	}
	genTx2 := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbC", // Same as genTx1
		"bls_public_key":    validBLSKey2,
		"timestamp":         1704067300,
	}

	genTx1JSON, err := json.Marshal(genTx1)
	require.NoError(t, err)
	genTx2JSON, err := json.Marshal(genTx2)
	require.NoError(t, err)

	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{genTx1JSON, genTx2JSON},
	}

	// Act
	err = genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestProcessGenTxs_ZeroAddress(t *testing.T) {
	// Arrange
	genTx := map[string]interface{}{
		"version":           "1.0",
		"chain_id":          "stablenet-1",
		"validator_address": "0x0000000000000000000000000000000000000000",
		"bls_public_key":    "0xa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6",
		"timestamp":         1704067200,
	}
	genTxJSON, err := json.Marshal(genTx)
	require.NoError(t, err)

	genesis := &Genesis{
		Config: &params.ChainConfig{
			Anzeon: &params.AnzeonConfig{
				Init: &params.WBFTInit{},
			},
		},
		AnzeonGenTxs: []json.RawMessage{genTxJSON},
	}

	// Act
	err = genesis.ProcessGenTxs()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "zero address")
}

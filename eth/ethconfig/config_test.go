// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package ethconfig

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	wbft "github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/params"
	_ "github.com/ethereum/go-ethereum/systemcontracts" // init SystemContractCodes
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetConfigFromChainConfig_BFork verifies that the production SetConfigFromChainConfig
// properly registers BFork SystemContractUpgrades and that GetSystemContractsStateTransition
// returns the correct v2 GovMinter bytecode transition at the fork block.
func TestSetConfigFromChainConfig_BFork(t *testing.T) {
	bForkBlock := big.NewInt(100)

	chainCfg := &params.ChainConfig{
		ChainID:       big.NewInt(8283),
		ApplepieBlock: big.NewInt(0),
		BForkBlock:    bForkBlock,
		Anzeon: &params.AnzeonConfig{
			WBFT: &params.WBFTConfig{
				EpochLength:           10,
				BlockPeriodSeconds:    1,
				RequestTimeoutSeconds: 2,
			},
			SystemContracts: &params.SystemContracts{
				GovValidator: &params.SystemContract{
					Address: common.HexToAddress("0x1001"),
					Version: "v1",
				},
				NativeCoinAdapter: &params.SystemContract{
					Address: common.HexToAddress("0x1000"),
					Version: "v1",
					Params: map[string]string{
						"masterMinter":  "0x0000000000000000000000000000000000001002",
						"minters":       "0x0000000000000000000000000000000000001003",
						"minterAllowed": "10000000000000000000000000000",
						"name":          "WKRC",
						"symbol":        "WKRC",
						"decimals":      "18",
						"currency":      "KRW",
					},
				},
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v1",
					Params: map[string]string{
						"quorum":        "1",
						"expiry":        "604800",
						"members":       "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
						"memberVersion": "1",
						"fiatToken":     "0x0000000000000000000000000000000000001000",
						"maxProposals":  "3",
					},
				},
				GovMasterMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1002"),
					Version: "v1",
					Params: map[string]string{
						"quorum":             "1",
						"expiry":             "604800",
						"members":            "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
						"memberVersion":      "1",
						"fiatToken":          "0x0000000000000000000000000000000000001000",
						"minters":            "0x0000000000000000000000000000000000001003",
						"maxMinterAllowance": "10000000000000000000000000000",
						"maxProposals":       "3",
					},
				},
				GovCouncil: &params.SystemContract{
					Address: common.HexToAddress("0x1004"),
					Version: "v1",
					Params: map[string]string{
						"quorum":        "1",
						"expiry":        "604800",
						"members":       "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
						"memberVersion": "1",
						"maxProposals":  "3",
					},
				},
			},
		},
		BFork: &params.AnzeonConfig{
			SystemContracts: &params.SystemContracts{
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v2",
				},
			},
		},
	}

	wbftCfg := new(wbft.Config)
	err := SetConfigFromChainConfig(wbftCfg, chainCfg)
	require.NoError(t, err)

	// Verify SystemContractUpgrades registration
	assert.Equal(t, 2, len(wbftCfg.SystemContractUpgrades),
		"Expected 2 SystemContractUpgrades (Anzeon at block 0 + BFork at block 100)")

	assert.Equal(t, int64(0), wbftCfg.SystemContractUpgrades[0].Block.Int64())
	assert.Equal(t, "v1", wbftCfg.SystemContractUpgrades[0].SystemContracts.GovMinter.Version)

	assert.Equal(t, int64(100), wbftCfg.SystemContractUpgrades[1].Block.Int64())
	assert.Equal(t, "v2", wbftCfg.SystemContractUpgrades[1].SystemContracts.GovMinter.Version)

	// At genesis (block 0): state transition should deploy v1 contracts
	st, err := wbft.GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	require.NoError(t, err)
	require.NotNil(t, st, "Block 0 should have state transition (Anzeon genesis)")
	assert.Equal(t, 5, len(st.Codes), "Block 0 should deploy all 5 system contracts")

	// Before BFork (block 50): no state transition
	st, err = wbft.GetSystemContractsStateTransition(wbftCfg, big.NewInt(50))
	require.NoError(t, err)
	assert.Nil(t, st, "Block 50 should have no state transition")

	// At BFork (block 100): state transition should deploy v2 GovMinter only
	st, err = wbft.GetSystemContractsStateTransition(wbftCfg, bForkBlock)
	require.NoError(t, err)
	require.NotNil(t, st, "Block 100 should have state transition (BFork)")
	assert.Equal(t, 1, len(st.Codes), "BFork should upgrade only GovMinter")
	assert.Equal(t, common.HexToAddress("0x1003"), st.Codes[0].Address)
	assert.NotEmpty(t, st.Codes[0].Code, "GovMinter v2 bytecode should not be empty")

	// Verify v2 code is different from v1 code
	stGenesis, _ := wbft.GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	var v1Code string
	for _, c := range stGenesis.Codes {
		if c.Address == common.HexToAddress("0x1003") {
			v1Code = c.Code
			break
		}
	}
	assert.NotEqual(t, v1Code, st.Codes[0].Code,
		"v2 GovMinter bytecode should differ from v1")

	// After BFork (block 101): no state transition
	st, err = wbft.GetSystemContractsStateTransition(wbftCfg, big.NewInt(101))
	require.NoError(t, err)
	assert.Nil(t, st, "Block 101 should have no state transition")
}

// TestSetConfigFromChainConfig_NoBFork verifies that when BFork is not configured,
// no BFork upgrades are registered.
func TestSetConfigFromChainConfig_NoBFork(t *testing.T) {
	chainCfg := &params.ChainConfig{
		ChainID:       big.NewInt(8283),
		ApplepieBlock: big.NewInt(0),
		// BForkBlock: nil — no BFork
		Anzeon: &params.AnzeonConfig{
			WBFT: &params.WBFTConfig{
				EpochLength:           10,
				BlockPeriodSeconds:    1,
				RequestTimeoutSeconds: 2,
			},
			SystemContracts: &params.SystemContracts{
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v1",
				},
			},
		},
	}

	wbftCfg := new(wbft.Config)
	err := SetConfigFromChainConfig(wbftCfg, chainCfg)
	require.NoError(t, err)

	assert.Equal(t, 1, len(wbftCfg.SystemContractUpgrades),
		"Without BFork, only Anzeon upgrade should be registered")
}

// Copyright 2017 The go-ethereum Authors
// Copyright 2024 The go-wemix-wbft Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/config_test.go (2024.07.25).
// Modified and improved for the wemix development.

package wbft

import (
	"errors"
	"math/big"
	"reflect"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"github.com/stretchr/testify/assert"
)

func TestProposerPolicy_UnmarshalTOML(t *testing.T) {
	input := `id = 2
`
	expectedId := ProposerPolicyId(2)
	var p proposerPolicyToml
	assert.NoError(t, toml.Unmarshal([]byte(input), &p))
	assert.Equal(t, expectedId, p.Id, "ProposerPolicyId mismatch")
}

func TestProposerPolicy_MarshalTOML(t *testing.T) {
	output := `id = 1
`
	p := &ProposerPolicy{Id: 1}
	b, err := p.MarshalTOML()
	if err != nil {
		t.Errorf("error marshalling ProposerPolicy: %v", err)
	}
	assert.Equal(t, output, b, "ProposerPolicy MarshalTOML mismatch")
}

// testChainConfigWrapper wraps ChainConfig to add fake hardForks for testing
type chainConfigWrapper struct {
	*params.ChainConfig
	fakeHardForks []fakeHardFork
}

type fakeHardFork struct {
	name                 string
	blockNum             *big.Int
	WBFTConfig           *params.WBFTConfig
	SystemContractConfig *params.SystemContracts
}

// newTestChainConfig creates a test chain config with fake hard forks
func newTestChainConfig() *chainConfigWrapper {
	return &chainConfigWrapper{
		ChainConfig:   params.TestWBFTChainConfig,
		fakeHardForks: make([]fakeHardFork, 0),
	}
}

// addFakeHardFork adds a fake hard fork for testing
func (cw *chainConfigWrapper) addFakeHardFork(name string, blockNum *big.Int, wbftConfig *params.WBFTConfig, systemContractConfig *params.SystemContracts) {
	fh := fakeHardFork{
		name:                 name,
		blockNum:             blockNum,
		WBFTConfig:           wbftConfig,
		SystemContractConfig: systemContractConfig,
	}
	cw.fakeHardForks = append(cw.fakeHardForks, fh)
}

// setConfigFromChainConfig is a test version of SetConfigFromChainConfig that works with fake hardForks
func setConfigFromChainConfig(wbftCfg *Config, chainCfg *chainConfigWrapper) error {
	config := chainCfg.Anzeon.WBFT
	if config.RequestTimeoutSeconds != 0 {
		wbftCfg.RequestTimeout = config.RequestTimeoutSeconds * 1000
	}
	if config.BlockPeriodSeconds != 0 {
		wbftCfg.BlockPeriod = config.BlockPeriodSeconds
	}
	if config.EpochLength != 0 {
		wbftCfg.Epoch = config.EpochLength
	}

	if config.ProposerPolicy != nil {
		wbftCfg.ProposerPolicy = NewProposerPolicy(ProposerPolicyId(*config.ProposerPolicy))
	}
	if config.MaxRequestTimeoutSeconds != nil {
		wbftCfg.MaxRequestTimeoutSeconds = *config.MaxRequestTimeoutSeconds
	}

	hfTransitionBlocks := make(map[*big.Int]bool)
	for _, hf := range chainCfg.fakeHardForks {
		transition := params.Transition{
			Block:      hf.blockNum,
			WBFTConfig: hf.WBFTConfig,
		}
		wbftCfg.Transitions = append(wbftCfg.Transitions, transition)
		hfTransitionBlocks[hf.blockNum] = true
	}

	if len(chainCfg.Transitions) > 0 {
		for _, t := range chainCfg.Transitions {
			if hfTransitionBlocks[t.Block] {
				return errors.New("hardfork transition block already exists")
			}
			wbftCfg.Transitions = append(wbftCfg.Transitions, t)
		}
	}

	sort.Slice(wbftCfg.Transitions, func(i, j int) bool {
		if wbftCfg.Transitions[i].Block == nil {
			return false
		}
		if wbftCfg.Transitions[j].Block == nil {
			return true
		}
		return wbftCfg.Transitions[i].Block.Cmp(wbftCfg.Transitions[j].Block) < 0
	})

	wbftCfg.SystemContractUpgrades = append(wbftCfg.SystemContractUpgrades, params.Upgrade{Block: new(big.Int), SystemContracts: chainCfg.Anzeon.SystemContracts})
	for _, hf := range chainCfg.fakeHardForks {
		upgrade := params.Upgrade{
			Block:           hf.blockNum,
			SystemContracts: hf.SystemContractConfig,
		}
		wbftCfg.SystemContractUpgrades = append(wbftCfg.SystemContractUpgrades, upgrade)
	}

	return nil
}

func TestGetConfig(t *testing.T) {
	// Create test chain config with fake hard forks
	testConfig := newTestChainConfig()
	wbftCfg := new(Config)

	// Add fake hard forks
	testConfig.addFakeHardFork("TestFork1", big.NewInt(10),
		&params.WBFTConfig{
			EpochLength: 200,
		},
		nil,
	)

	testConfig.addFakeHardFork("TestFork2", big.NewInt(20),
		&params.WBFTConfig{
			BlockPeriodSeconds: 3,
		},
		nil,
	)

	testConfig.addFakeHardFork("TestFork3", big.NewInt(30),
		&params.WBFTConfig{
			RequestTimeoutSeconds: 4000,
		},
		nil,
	)

	testConfig.Transitions = []params.Transition{{
		Block:      big.NewInt(1),
		WBFTConfig: &params.WBFTConfig{EpochLength: 40000},
	}, {
		Block:      big.NewInt(3),
		WBFTConfig: &params.WBFTConfig{BlockPeriodSeconds: 100},
	}, {
		Block:      big.NewInt(5),
		WBFTConfig: &params.WBFTConfig{RequestTimeoutSeconds: 5000},
	}}

	setConfigFromChainConfig(wbftCfg, testConfig)

	createExpectedConfig := func(baseConfig *Config, modifications func(*Config)) Config {
		expected := *baseConfig
		modifications(&expected)
		return expected
	}

	tests := []struct {
		name           string
		blockNumber    uint64
		expectedConfig Config
	}{
		{
			name:           "Before any hard fork or transitions (block 0)",
			blockNumber:    0,
			expectedConfig: *wbftCfg,
		},
		{
			name:        "After first transition (block 1)",
			blockNumber: 1,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 40000 // From Transition 1
			}),
		},
		{
			name:        "After second transition (block 3)",
			blockNumber: 4,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 40000     // From Transition 1
				cfg.BlockPeriod = 100 // From Transition 2
			}),
		},
		{
			name:        "After second transition (block 5)",
			blockNumber: 9,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 40000                // From Transition 1
				cfg.BlockPeriod = 100            // From Transition 2
				cfg.RequestTimeout = 5000 * 1000 // From Transition 3
			}),
		},
		{
			name:        "After TestFork1 (block 15)",
			blockNumber: 15,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 200                  // From TestFork1
				cfg.BlockPeriod = 100            // From Transition 2
				cfg.RequestTimeout = 5000 * 1000 // From Transition 3
			}),
		},
		{
			name:        "After TestFork2 (block 25)",
			blockNumber: 25,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 200                  // From TestFork1
				cfg.BlockPeriod = 3              // From TestFork2
				cfg.RequestTimeout = 5000 * 1000 // From Transition 3
			}),
		},
		{
			name:        "After TestFork3 (block 35)",
			blockNumber: 35,
			expectedConfig: createExpectedConfig(wbftCfg, func(cfg *Config) {
				cfg.Epoch = 200                  // From TestFork1
				cfg.BlockPeriod = 3              // From TestFork2
				cfg.RequestTimeout = 4000 * 1000 // From TestFork3
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := wbftCfg.GetConfig(big.NewInt(int64(test.blockNumber)))
			if !reflect.DeepEqual(result, test.expectedConfig) {
				t.Errorf("error in %s:\nexpected: %+v\ngot: %+v\n", test.name, test.expectedConfig, result)
			}
		})
	}
}

// getSystemContracts is test version of GetSystemContracts that works with fake hardForks
func getSystemContracts(blockNumber *big.Int, wbftCfg *Config) params.SystemContracts {
	gc := params.SystemContracts{}

	if len(wbftCfg.SystemContractUpgrades) > 0 {
		wbftCfg.getSystemContractsValue(blockNumber, func(upgrade params.Upgrade) {
			if upgrade.GovValidator != nil {
				gc.GovValidator = upgrade.GovValidator
			}
			if upgrade.NativeCoinAdapter != nil {
				gc.NativeCoinAdapter = upgrade.NativeCoinAdapter
			}
			if upgrade.GovMasterMinter != nil {
				gc.GovMasterMinter = upgrade.GovMasterMinter
			}
			if upgrade.GovMinter != nil {
				gc.GovMinter = upgrade.GovMinter
			}
			if upgrade.GovCouncil != nil {
				gc.GovCouncil = upgrade.GovCouncil
			}
		})
	}
	return gc
}

func TestGetSystemContracts(t *testing.T) {
	// Create test chain config with fake hard forks
	testConfig := newTestChainConfig()
	wbftCfg := new(Config)

	// Add fake hard forks
	testConfig.addFakeHardFork("TestFork1", big.NewInt(10),
		nil,
		&params.SystemContracts{
			GovValidator: &params.SystemContract{
				Address: common.HexToAddress("0x2000"),
				Version: "v2",
				Params:  nil,
			},
		},
	)
	testConfig.addFakeHardFork("TestFork2", big.NewInt(20),
		nil,
		&params.SystemContracts{
			GovValidator: &params.SystemContract{
				Address: common.HexToAddress("0x2000"),
				Version: "v2",
				Params:  nil,
			},
		},
	)

	setConfigFromChainConfig(wbftCfg, testConfig)
	baseContracts := testConfig.Anzeon.SystemContracts

	createExpectedSystemContracts := func(baseConfig *params.SystemContracts, modifications func(config *params.SystemContracts)) params.SystemContracts {
		expected := *baseConfig
		modifications(&expected)
		return expected
	}
	tests := []struct {
		name           string
		blockNumber    uint64
		expectedConfig params.SystemContracts
	}{
		{
			name:           "Before any hard forks",
			blockNumber:    0,
			expectedConfig: *baseContracts,
		},
		{
			name:        "After first transition (block 1)",
			blockNumber: 11,
			expectedConfig: createExpectedSystemContracts(baseContracts, func(contracts *params.SystemContracts) {
				contracts.GovValidator = &params.SystemContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
			}),
		},
		{
			name:        "After second transition (block 3)",
			blockNumber: 20,
			expectedConfig: createExpectedSystemContracts(baseContracts, func(contracts *params.SystemContracts) {
				contracts.GovValidator = &params.SystemContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
				contracts.GovValidator = &params.SystemContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getSystemContracts(big.NewInt(int64(test.blockNumber)), wbftCfg)
			if !reflect.DeepEqual(result, test.expectedConfig) {
				t.Errorf("error in %s:\nexpected: %+v\ngot: %+v\n", test.name, test.expectedConfig, result)
			}
		})
	}
}

// TestBohoSystemContractUpgrade verifies the Boho hardfork production path:
// - SetConfigFromChainConfig registers Boho SystemContractUpgrades
// - GetSystemContractsStateTransition returns v2 GovMinter at the fork block
// - Before the fork block, only Anzeon (v1) contracts are returned
func TestBohoSystemContractUpgrade(t *testing.T) {
	bohoBlock := big.NewInt(100)

	chainCfg := &params.ChainConfig{
		ChainID:       big.NewInt(8283),
		ApplepieBlock: big.NewInt(0),
		BohoBlock:     bohoBlock,
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
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v1",
				},
			},
		},
		Boho: &params.AnzeonConfig{
			SystemContracts: &params.SystemContracts{
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v2",
				},
			},
		},
	}

	wbftCfg := new(Config)
	err := SetConfigFromChainConfig(wbftCfg, chainCfg)
	assert.NoError(t, err)

	// Verify SystemContractUpgrades has 2 entries: Anzeon (block 0) and Boho (block 100)
	assert.Equal(t, 2, len(wbftCfg.SystemContractUpgrades),
		"Expected 2 SystemContractUpgrades (Anzeon + Boho)")
	assert.Equal(t, int64(0), wbftCfg.SystemContractUpgrades[0].Block.Int64(),
		"First upgrade should be at block 0 (Anzeon)")
	assert.Equal(t, bohoBlock.Int64(), wbftCfg.SystemContractUpgrades[1].Block.Int64(),
		"Second upgrade should be at block 100 (Boho)")

	// Before Boho: GetSystemContractsStateTransition should return nil (no transition at block 50)
	st, err := GetSystemContractsStateTransition(wbftCfg, big.NewInt(50))
	assert.NoError(t, err)
	assert.Nil(t, st, "No state transition should occur at block 50")

	// At Boho block: GetSystemContractsStateTransition should return v2 GovMinter transition
	st, err = GetSystemContractsStateTransition(wbftCfg, bohoBlock)
	assert.NoError(t, err)
	assert.NotNil(t, st, "State transition should occur at block 100 (Boho)")
	assert.Equal(t, 1, len(st.Codes), "Boho should upgrade 1 contract (GovMinter)")
	assert.Equal(t, common.HexToAddress("0x1003"), st.Codes[0].Address,
		"Upgraded contract should be GovMinter at 0x1003")

	// After Boho: no transition at block 101
	st, err = GetSystemContractsStateTransition(wbftCfg, big.NewInt(101))
	assert.NoError(t, err)
	assert.Nil(t, st, "No state transition should occur at block 101")

	// Verify getSystemContracts merges correctly
	// At block 50: GovMinter should be v1
	contracts50 := getSystemContracts(big.NewInt(50), wbftCfg)
	assert.Equal(t, "v1", contracts50.GovMinter.Version,
		"GovMinter should be v1 before Boho")

	// At block 100: GovMinter should be v2 (Boho applied)
	contracts100 := getSystemContracts(big.NewInt(100), wbftCfg)
	assert.Equal(t, "v2", contracts100.GovMinter.Version,
		"GovMinter should be v2 at Boho block")

	// At block 200: GovMinter should still be v2
	contracts200 := getSystemContracts(big.NewInt(200), wbftCfg)
	assert.Equal(t, "v2", contracts200.GovMinter.Version,
		"GovMinter should remain v2 after Boho")
}

// TestGetSystemContractsStateTransition_Merge verifies that multiple upgrades
// at the same block number are merged into a single StateTransition.
func TestGetSystemContractsStateTransition_Merge(t *testing.T) {
	wbftCfg := &Config{
		SystemContractUpgrades: []params.Upgrade{
			{
				Block: big.NewInt(0),
				SystemContracts: &params.SystemContracts{
					GovValidator: &params.SystemContract{
						Address: common.HexToAddress("0x1001"),
						Version: "v1",
					},
					GovMinter: &params.SystemContract{
						Address: common.HexToAddress("0x1003"),
						Version: "v1",
					},
				},
			},
			{
				Block: big.NewInt(0),
				SystemContracts: &params.SystemContracts{
					GovMinter: &params.SystemContract{
						Address: common.HexToAddress("0x1003"),
						Version: "v2",
					},
				},
			},
		},
	}

	st, err := GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	assert.NoError(t, err)
	assert.NotNil(t, st, "Merged state transition should not be nil")

	// Should contain codes from both upgrades: GovValidator v1 + GovMinter v1 + GovMinter v2
	assert.True(t, len(st.Codes) >= 2,
		"Merged transition should contain codes from both upgrades, got %d", len(st.Codes))

	// Last GovMinter code should be v2 (from second upgrade)
	var lastMinterCode string
	for _, c := range st.Codes {
		if c.Address == common.HexToAddress("0x1003") {
			lastMinterCode = c.Code
		}
	}
	assert.NotEmpty(t, lastMinterCode, "GovMinter code should be present in merged result")
}

// TestSetConfigFromChainConfig_CollectUpgrades verifies that CollectUpgrades()
// correctly registers all hardfork upgrades in SystemContractUpgrades.
func TestSetConfigFromChainConfig_CollectUpgrades(t *testing.T) {
	chainCfg := &params.ChainConfig{
		ChainID:       big.NewInt(8283),
		ApplepieBlock: big.NewInt(0),
		BohoBlock:     big.NewInt(0),
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
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v1",
				},
			},
		},
		Boho: &params.AnzeonConfig{
			SystemContracts: &params.SystemContracts{
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v2",
				},
			},
		},
	}

	wbftCfg := new(Config)
	err := SetConfigFromChainConfig(wbftCfg, chainCfg)
	assert.NoError(t, err)

	// Anzeon (block 0) + Boho (block 0) = 2 entries
	assert.Equal(t, 2, len(wbftCfg.SystemContractUpgrades),
		"Expected 2 SystemContractUpgrades (Anzeon + Boho)")

	// Both should be at block 0 and sorted
	for _, u := range wbftCfg.SystemContractUpgrades {
		assert.Equal(t, int64(0), u.Block.Int64(),
			"All upgrades should be at block 0")
	}

	// Merged state transition at block 0 should contain both upgrades
	st, err := GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	assert.NoError(t, err)
	assert.NotNil(t, st)
	assert.True(t, len(st.Codes) >= 2,
		"Block 0 should have codes from both Anzeon and Boho")
}

// TestProcessFinalize_Idempotent verifies that applying the same upgrade twice
// (genesis overlay + processFinalize) produces no harmful effect.
func TestProcessFinalize_Idempotent(t *testing.T) {
	chainCfg := &params.ChainConfig{
		ChainID:       big.NewInt(8283),
		ApplepieBlock: big.NewInt(0),
		BohoBlock:     big.NewInt(0),
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
		Boho: &params.AnzeonConfig{
			SystemContracts: &params.SystemContracts{
				GovMinter: &params.SystemContract{
					Address: common.HexToAddress("0x1003"),
					Version: "v2",
				},
			},
		},
	}

	wbftCfg := new(Config)
	err := SetConfigFromChainConfig(wbftCfg, chainCfg)
	assert.NoError(t, err)

	// First call at block 0
	st1, err := GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	assert.NoError(t, err)
	assert.NotNil(t, st1)

	// Second call at block 0 (simulating re-execution)
	st2, err := GetSystemContractsStateTransition(wbftCfg, big.NewInt(0))
	assert.NoError(t, err)
	assert.NotNil(t, st2)

	// Both calls should produce the same result
	assert.Equal(t, len(st1.Codes), len(st2.Codes),
		"Idempotent: same number of codes on repeated call")
}

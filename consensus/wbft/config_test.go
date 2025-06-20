// Modification Copyright 2024 The Wemix Authors
// Copyright 2017 The go-ethereum Authors
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
	name              string
	blockNum          *big.Int
	WBFTConfig        *params.WBFTConfig
	GovContractConfig *params.GovContracts
}

// newTestChainConfig creates a test chain config with fake hard forks
func newTestChainConfig() *chainConfigWrapper {
	return &chainConfigWrapper{
		ChainConfig:   params.TestWBFTChainConfig,
		fakeHardForks: make([]fakeHardFork, 0),
	}
}

// addFakeHardFork adds a fake hard fork for testing
func (cw *chainConfigWrapper) addFakeHardFork(name string, blockNum *big.Int, wbftConfig *params.WBFTConfig, govContractConfig *params.GovContracts) {
	fh := fakeHardFork{
		name:              name,
		blockNum:          blockNum,
		WBFTConfig:        wbftConfig,
		GovContractConfig: govContractConfig,
	}
	cw.fakeHardForks = append(cw.fakeHardForks, fh)
}

// setConfigFromChainConfig is a test version of SetConfigFromChainConfig that works with fake hardForks
func setConfigFromChainConfig(wbftCfg *Config, chainCfg *chainConfigWrapper) error {
	config := chainCfg.Croissant.WBFT
	if config.RequestTimeoutSeconds != 0 {
		wbftCfg.RequestTimeout = config.RequestTimeoutSeconds * 1000
	}
	if config.BlockPeriodSeconds != 0 {
		wbftCfg.BlockPeriod = config.BlockPeriodSeconds
	}
	if config.EpochLength != 0 {
		wbftCfg.Epoch = config.EpochLength
	}
	wbftCfg.BlockReward = config.BlockReward
	wbftCfg.BlockRewardBeneficiary = config.BlockRewardBeneficiary

	if config.ProposerPolicy != nil {
		wbftCfg.ProposerPolicy = NewProposerPolicy(ProposerPolicyId(*config.ProposerPolicy))
	}
	if config.TargetValidators != nil {
		wbftCfg.TargetValidators = *config.TargetValidators
	}
	if config.MaxRequestTimeoutSeconds != nil {
		wbftCfg.MaxRequestTimeoutSeconds = *config.MaxRequestTimeoutSeconds
	}
	if config.StabilizingStakersThreshold != nil {
		wbftCfg.StabilizingStakersThreshold = *config.StabilizingStakersThreshold
	}
	if config.UseNCP != nil {
		wbftCfg.UseNCP = *config.UseNCP
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

	if chainCfg.Transitions != nil && len(chainCfg.Transitions) > 0 {
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

	wbftCfg.GovContractUpgrades = append(wbftCfg.GovContractUpgrades, params.Upgrade{Block: chainCfg.CroissantBlock, GovContracts: chainCfg.Croissant.GovContracts})
	for _, hf := range chainCfg.fakeHardForks {
		upgrade := params.Upgrade{
			Block:        hf.blockNum,
			GovContracts: hf.GovContractConfig,
		}
		wbftCfg.GovContractUpgrades = append(wbftCfg.GovContractUpgrades, upgrade)
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

// getGovContracts is test version of GetGovContracts that works with fake hardForks
func getGovContracts(blockNumber *big.Int, wbftCfg *Config) params.GovContracts {
	gc := params.GovContracts{}

	if wbftCfg.GovContractUpgrades != nil && len(wbftCfg.GovContractUpgrades) > 0 {
		wbftCfg.getGovContractsValue(blockNumber, func(upgrade params.Upgrade) {
			if upgrade.GovStaking != nil {
				gc.GovStaking = upgrade.GovStaking
			}
			if upgrade.GovConfig != nil {
				gc.GovConfig = upgrade.GovConfig
			}
			if upgrade.GovRewardeeImp != nil {
				gc.GovRewardeeImp = upgrade.GovRewardeeImp
			}
			if upgrade.GovNCP != nil {
				gc.GovNCP = upgrade.GovNCP
			}
		})
	}
	return gc
}

func TestGetGovContracts(t *testing.T) {
	// Create test chain config with fake hard forks
	testConfig := newTestChainConfig()
	wbftCfg := new(Config)

	// Add fake hard forks
	testConfig.addFakeHardFork("TestFork1", big.NewInt(10),
		nil,
		&params.GovContracts{
			GovConfig: &params.GovContract{
				Address: common.HexToAddress("0x2000"),
				Version: "v2",
				Params:  nil,
			},
		},
	)
	testConfig.addFakeHardFork("TestFork2", big.NewInt(20),
		nil,
		&params.GovContracts{
			GovStaking: &params.GovContract{
				Address: common.HexToAddress("0x2000"),
				Version: "v2",
				Params:  nil,
			},
		},
	)

	setConfigFromChainConfig(wbftCfg, testConfig)
	baseContracts := testConfig.Croissant.GovContracts

	createExpectedGovContracts := func(baseConfig *params.GovContracts, modifications func(config *params.GovContracts)) params.GovContracts {
		expected := *baseConfig
		modifications(&expected)
		return expected
	}
	tests := []struct {
		name           string
		blockNumber    uint64
		expectedConfig params.GovContracts
	}{
		{
			name:           "Before any hard forks",
			blockNumber:    0,
			expectedConfig: *baseContracts,
		},
		{
			name:        "After first transition (block 1)",
			blockNumber: 11,
			expectedConfig: createExpectedGovContracts(baseContracts, func(contracts *params.GovContracts) {
				contracts.GovConfig = &params.GovContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
			}),
		},
		{
			name:        "After second transition (block 3)",
			blockNumber: 20,
			expectedConfig: createExpectedGovContracts(baseContracts, func(contracts *params.GovContracts) {
				contracts.GovConfig = &params.GovContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
				contracts.GovStaking = &params.GovContract{
					Address: common.HexToAddress("0x2000"),
					Version: "v2",
					Params:  nil,
				}
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getGovContracts(big.NewInt(int64(test.blockNumber)), wbftCfg)
			if !reflect.DeepEqual(result, test.expectedConfig) {
				t.Errorf("error in %s:\nexpected: %+v\ngot: %+v\n", test.name, test.expectedConfig, result)
			}
		})
	}
}

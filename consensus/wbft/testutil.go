// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.
//
// This file provides test utilities for WBFT consensus to avoid code duplication
// across test files while preventing cyclic imports.

package wbft

import (
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/params"
)

// SetConfigFromChainConfigForTest is a copy of ethconfig.SetConfigFromChainConfig
// This function is used in test files to avoid cyclic import issues
func SetConfigFromChainConfig(wbftCfg *Config, chainCfg *params.ChainConfig) error {
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

	//add hardforks that includes wbft config after anzeon here like :
	// transition := params.Transition{
	// 	Block:      chainCfg.DalgonaBlock,
	// 	WBFTConfig: chainCfg.Dalgona.WBFT,
	// }
	// wbftCfg.Transitions = append(wbftCfg.Transitions, transition)
	// hfTransitionBlocks[chainCfg.DalgonaBlock] = true

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
	// add hardforks that includes systemContracts after anzeon here like :
	// wbftCfg.SystemContractUpgrades = append(wbftCfg.SystemContractUpgrades, params.Upgrade{Block: chainCfg.DalgonaBlock, SystemContracts: chainCfg.Dalgona.SystemContracts})
	return nil
}

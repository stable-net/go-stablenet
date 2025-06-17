// Modification Copyright 2024 The Wemix Authors
//
// This file provides test utilities for QBFT consensus to avoid code duplication
// across test files while preventing cyclic imports.

package qbft

import (
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/params"
)

// SetConfigFromChainConfigForTest is a copy of ethconfig.SetConfigFromChainConfig
// This function is used in test files to avoid cyclic import issues
func SetConfigFromChainConfig(qbftCfg *Config, chainCfg *params.ChainConfig) error {
	config := chainCfg.MontBlanc.WBFT
	if config.RequestTimeoutSeconds != 0 {
		qbftCfg.RequestTimeout = config.RequestTimeoutSeconds * 1000
	}
	if config.BlockPeriodSeconds != 0 {
		qbftCfg.BlockPeriod = config.BlockPeriodSeconds
	}
	if config.EpochLength != 0 {
		qbftCfg.Epoch = config.EpochLength
	}
	qbftCfg.BlockReward = config.BlockReward
	qbftCfg.BlockRewardBeneficiary = config.BlockRewardBeneficiary

	if config.ProposerPolicy != nil {
		qbftCfg.ProposerPolicy = NewProposerPolicy(ProposerPolicyId(*config.ProposerPolicy))
	}
	if config.TargetValidators != nil {
		qbftCfg.TargetValidators = *config.TargetValidators
	}
	if config.MaxRequestTimeoutSeconds != nil {
		qbftCfg.MaxRequestTimeoutSeconds = *config.MaxRequestTimeoutSeconds
	}
	if config.StabilizingStakersThreshold != nil {
		qbftCfg.StabilizingStakersThreshold = *config.StabilizingStakersThreshold
	}
	if config.UseNCP != nil {
		qbftCfg.UseNCP = *config.UseNCP
	}

	hfTransitionBlocks := make(map[*big.Int]bool)

	//add hardforks that includes wbft config after montblanc here like :
	// transition := params.Transition{
	// 	Block:      chainCfg.DalgonaBlock,
	// 	WBFTConfig: chainCfg.Dalgona.WBFT,
	// }
	// qbftCfg.Transitions = append(qbftCfg.Transitions, transition)
	// hfTransitionBlocks[chainCfg.DalgonaBlock] = true

	if chainCfg.Transitions != nil && len(chainCfg.Transitions) > 0 {
		for _, t := range chainCfg.Transitions {
			if hfTransitionBlocks[t.Block] {
				return errors.New("hardfork transition block already exists")
			}
			qbftCfg.Transitions = append(qbftCfg.Transitions, t)
		}
	}

	sort.Slice(qbftCfg.Transitions, func(i, j int) bool {
		if qbftCfg.Transitions[i].Block == nil {
			return false
		}
		if qbftCfg.Transitions[j].Block == nil {
			return true
		}
		return qbftCfg.Transitions[i].Block.Cmp(qbftCfg.Transitions[j].Block) < 0
	})

	qbftCfg.GovContractUpgrades = append(qbftCfg.GovContractUpgrades, params.Upgrade{Block: chainCfg.MontBlancBlock, GovContracts: chainCfg.MontBlanc.GovContracts})
	// add hardforks that includes govContracts after montblanc here like :
	// qbftCfg.GovContractUpgrades = append(qbftCfg.GovContractUpgrades, params.Upgrade{Block: chainCfg.DalgonaBlock, GovContracts: chainCfg.Dalgona.GovContracts})
	return nil
}

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
// This file is derived from quorum/consensus/istanbul/config.go (2024.07.25).
// Modified and improved for the wemix development.

package qbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"

	"github.com/naoina/toml"
)

type ProposerPolicyId uint64

const (
	RoundRobin ProposerPolicyId = iota
	Sticky
)

// ProposerPolicy represents the Validator Proposer Policy
type ProposerPolicy struct {
	Id ProposerPolicyId    // Could be RoundRobin or Sticky
	By ValidatorSortByFunc // func that defines how the ValidatorSet should be sorted
}

// NewRoundRobinProposerPolicy returns a RoundRobin ProposerPolicy with ValidatorSortByString as default sort function
func NewRoundRobinProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(RoundRobin)
}

// NewStickyProposerPolicy return a Sticky ProposerPolicy with ValidatorSortByString as default sort function
func NewStickyProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(Sticky)
}

func NewProposerPolicy(id ProposerPolicyId) *ProposerPolicy {
	return NewProposerPolicyByIdAndSortFunc(id, ValidatorSortByString())
}

func NewProposerPolicyByIdAndSortFunc(id ProposerPolicyId, by ValidatorSortByFunc) *ProposerPolicy {
	return &ProposerPolicy{Id: id, By: by}
}

type proposerPolicyToml struct {
	Id ProposerPolicyId
}

func (p *ProposerPolicy) MarshalTOML() (interface{}, error) {
	if p == nil {
		return nil, nil
	}
	pp := &proposerPolicyToml{Id: p.Id}
	data, err := toml.Marshal(pp)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (p *ProposerPolicy) UnmarshalTOML(decode func(interface{}) error) error {
	var innerToml string
	err := decode(&innerToml)
	if err != nil {
		return err
	}
	var pp proposerPolicyToml
	err = toml.Unmarshal([]byte(innerToml), &pp)
	if err != nil {
		return err
	}
	p.Id = pp.Id
	p.By = ValidatorSortByString()
	return nil
}

// Use sets the ValidatorSortByFunc for the given ProposerPolicy and sorts the validatorSets according to it
func (p *ProposerPolicy) Use(v ValidatorSortByFunc) {
	p.By = v
}

type Config struct {
	RequestTimeout           uint64                  `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod              uint64                  `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	ProposerPolicy           *ProposerPolicy         `toml:",omitempty"` // The policy for proposer selection
	Epoch                    uint64                  `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
	AllowedFutureBlockTime   uint64                  `toml:",omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BlockReward              *math.HexOrDecimal256   `toml:",omitempty"` // Reward
	BlockRewardBeneficiary   *params.BeneficiaryInfo `toml:",omitempty"`
	Validators               []common.Address        `toml:",omitempty"`
	BLSPublicKeys            [][]byte                `toml:",omitempty"`
	TargetValidators         uint64                  `toml:",omitempty"`
	MaxRequestTimeoutSeconds uint64                  `toml:",omitempty"`
	GovParams                *params.GovParams       `toml:",omitempty"`
	Transitions              []params.Transition
}

var value, _ = new(big.Int).SetString("340282366920938463463374607431768211455", 10)

var DefaultConfig = &Config{
	RequestTimeout:         1000,
	BlockPeriod:            1,
	ProposerPolicy:         NewRoundRobinProposerPolicy(),
	Epoch:                  10,
	AllowedFutureBlockTime: 0,
	GovParams: &params.GovParams{
		MinimumStaking:     (*math.HexOrDecimal256)(new(big.Int).Mul(big.NewInt(params.Ether), big.NewInt(500_000))),
		MaximumStaking:     (*math.HexOrDecimal256)(value),
		UnbondingStaker:    604800, // 7 days
		UnbondingDelegator: 259200, // 3 days
		FeePrecision:       10000,  // 0.01%
		ChangeFeeDelay:     604800, // 7 days
		MinStakers:         1,
	},
}

func (c Config) GetConfig(blockNumber *big.Int) Config {
	newConfig := c

	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.RequestTimeoutSeconds != 0 {
			// RequestTimeout is on milliseconds
			newConfig.RequestTimeout = transition.RequestTimeoutSeconds * 1000
		}
		if transition.EpochLength != 0 {
			newConfig.Epoch = transition.EpochLength
		}
		if transition.BlockPeriodSeconds != 0 {
			newConfig.BlockPeriod = transition.BlockPeriodSeconds
		}
		if transition.BlockReward != nil {
			newConfig.BlockReward = transition.BlockReward
		}
		if transition.BlockRewardBeneficiary != nil {
			newConfig.BlockRewardBeneficiary = transition.BlockRewardBeneficiary
		}
		if transition.TargetValidators != nil {
			newConfig.TargetValidators = *transition.TargetValidators
		}
		if transition.MaxRequestTimeoutSeconds != nil {
			newConfig.MaxRequestTimeoutSeconds = *transition.MaxRequestTimeoutSeconds
		}
		if transition.GovParams != nil {
			newConfig.GovParams = transition.GovParams
		}
	})

	return newConfig
}

func (c *Config) getTransitionValue(num *big.Int, callback func(transition params.Transition)) {
	if c != nil && num != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(num) <= 0; i++ {
			callback(c.Transitions[i])
		}
	}
}

// String implements the stringer interface, returning the consensus engine details.
func (c *Config) String() string {
	return "qbft"
}

func GetFirstWbftBlockNumber(config *params.ChainConfig) *big.Int {
	if config.MontBlancBlock == nil {
		// wbft engine started from genesis
		return common.Big0
	} else {
		// wbft engine started with montblanc hardfork
		return config.MontBlancBlock
	}
}

func GetStateTransitions(chainConfig *params.ChainConfig, num *big.Int) []params.StateTransition {
	if chainConfig != nil && num != nil {
		transitions := make([]params.StateTransition, 0)

		if chainConfig.MontBlancBlock != nil && chainConfig.MontBlancBlock.Cmp(num) == 0 {
			transitions = append(transitions, getMontBlancTransition(chainConfig))
		}

		if st := chainConfig.GetStateTransitions(num); len(st) > 0 {
			transitions = append(transitions, st...)
		}
		return transitions
	}
	return nil
}

func getMontBlancTransition(config *params.ChainConfig) params.StateTransition {
	st := params.StateTransition{
		Codes: []params.CodeParam{
			{Address: govwbft.GovConfigAddress, Code: govwbft.GovConfigContract},
			{Address: govwbft.GovStakingAddress, Code: govwbft.GovStakingContract},
			{Address: govwbft.GovRewardeeImpAddress, Code: govwbft.GovRewardeeImpContract},
		},
		States: []params.StateParam{
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(0)), Value: common.BigToHash((*big.Int)(config.QBFT.GovParams.MinimumStaking))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(1)), Value: common.BigToHash((*big.Int)(config.QBFT.GovParams.MaximumStaking))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(2)), Value: common.BigToHash(new(big.Int).SetUint64(config.QBFT.GovParams.UnbondingStaker))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(3)), Value: common.BigToHash(new(big.Int).SetUint64(config.QBFT.GovParams.UnbondingDelegator))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(4)), Value: common.BigToHash(new(big.Int).SetUint64(config.QBFT.GovParams.FeePrecision))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(5)), Value: common.BigToHash(new(big.Int).SetUint64(config.QBFT.GovParams.ChangeFeeDelay))},
			{Address: govwbft.GovConfigAddress, Key: common.BigToHash(big.NewInt(6)), Value: common.BigToHash(new(big.Int).SetUint64(config.QBFT.GovParams.MinStakers))},
		},
	}

	if config.MontBlanc != nil && len(config.MontBlanc.NCPs) > 0 {
		st.Codes = append(st.Codes, params.CodeParam{Address: govwbft.GovNCPAddress, Code: govwbft.GovNCPContract})
		st.States = append(st.States, govwbft.InitializeNCP(config.MontBlanc.NCPs)...)
	}
	return st
}

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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
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
	RequestTimeout              uint64                  `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod                 uint64                  `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	ProposerPolicy              *ProposerPolicy         `toml:",omitempty"` // The policy for proposer selection
	Epoch                       uint64                  `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
	AllowedFutureBlockTime      uint64                  `toml:",omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BlockReward                 *math.HexOrDecimal256   `toml:",omitempty"` // Reward
	BlockRewardBeneficiary      *params.BeneficiaryInfo `toml:",omitempty"`
	TargetValidators            uint64                  `toml:",omitempty"`
	MaxRequestTimeoutSeconds    uint64                  `toml:",omitempty"`
	StabilizingStakersThreshold uint64                  `toml:",omitempty"`
	UseNCP                      bool                    `toml:",omitempty"` // Use NCP or not
	Transitions                 []params.Transition
}

var DefaultConfig = &Config{
	RequestTimeout:              1000,
	BlockPeriod:                 1,
	ProposerPolicy:              NewRoundRobinProposerPolicy(),
	Epoch:                       10,
	AllowedFutureBlockTime:      0,
	StabilizingStakersThreshold: 1,
	UseNCP:                      false,
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
		newConfig.UseNCP = transition.UseNCP
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
	return "wbft"
}

func GetMontBlancTransition(chainConfig *params.ChainConfig, num *big.Int) (*params.StateTransition, error) {
	if chainConfig == nil || chainConfig.MontBlancBlock == nil || num == nil {
		return nil, errors.New("nil montBlanc config or nil block number")
	}

	if num.Cmp(chainConfig.MontBlancBlock) == 0 {
		return govwbft.GetMontBlancTransition(chainConfig.MontBlanc.Init.GovContracts)
	}

	for _, upgrade := range chainConfig.MontBlanc.Upgrades {
		if num.Cmp(upgrade.Block) == 0 {
			return govwbft.GetMontBlancTransition(upgrade.GovContracts)
		} else if num.Cmp(upgrade.Block) < 0 {
			break
		}
	}
	return nil, nil
}

func CreateInitialExtraData(config *params.MontBlancConfig) ([]byte, error) {
	epochInfo, err := CreateInitialEpochInfo(config)
	if err != nil {
		return nil, err
	}

	extraData := &types.QBFTExtra{
		EpochInfo: epochInfo,
	}

	extraDataBytes, err := rlp.EncodeToBytes(extraData)
	if err != nil {
		return nil, err
	}

	return extraDataBytes, nil
}

func CreateInitialEpochInfo(config *params.MontBlancConfig) (*types.EpochInfo, error) {
	var (
		stakers       []common.Address
		blsPublicKeys []string
		epochInfo     = new(types.EpochInfo)
	)
	stakers = append(stakers, config.Init.Validators...)
	blsPublicKeys = append(blsPublicKeys, config.Init.BLSPublicKeys...)
	for i, addr := range stakers {
		epochInfo.Stakers = append(epochInfo.Stakers, &types.Staker{
			Addr:      addr,
			Diligence: types.DefaultDiligence,
		})
		epochInfo.Validators = append(epochInfo.Validators, uint32(i))
		epochInfo.BLSPublicKeys = append(epochInfo.BLSPublicKeys, hexutil.MustDecode(blsPublicKeys[i]))
		epochInfo.Stabilizing = true
	}

	log.Trace("initial epoch info", "validators", epochInfo.Validators)
	for i, staker := range epochInfo.Stakers {
		log.Trace(fmt.Sprintf("  - stakers[%d]", i), "addr", staker.Addr, "diligence", staker.Diligence)
	}

	return epochInfo, nil
}

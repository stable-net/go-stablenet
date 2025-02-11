// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/params/config.go (2024.07.25).
// Modified and improved for the wemix development.

package params

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// ## Quorum QBFT START
type QBFTConfig struct {
	EpochLength              uint64                `json:"epochLength"`                      // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds       uint64                `json:"blockPeriodSeconds"`               // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	RequestTimeoutSeconds    uint64                `json:"requestTimeoutSeconds"`            // Minimum request timeout for each QBFT round in milliseconds
	ProposerPolicy           uint64                `json:"proposerPolicy"`                   // The policy for proposer selection
	BlockReward              *math.HexOrDecimal256 `json:"blockReward,omitempty"`            // Reward from start, works only on QBFT consensus protocol
	BlockRewardBeneficiary   *BeneficiaryInfo      `json:"blockRewardBeneficiary,omitempty"` // Reward beneficiaries
	Validators               []common.Address      `json:"validators"`                       // Validators list
	MinStakers               uint64                `json:"minStakers"`                       // Minimum number of stakers before stabilization
	TargetValidators         uint64                `json:"targetValidators"`                 // Target number of validators
	MaxRequestTimeoutSeconds *uint64               `json:"maxRequestTimeoutSeconds"`         // The max round time
}

type BeneficiaryInfo struct {
	Denominator   uint64         `json:"denominator"`
	Beneficiaries []*Beneficiary `json:"beneficiaries"`
}

type Beneficiary struct {
	Name      string         `json:"name"`
	Addr      common.Address `json:"addr"`
	Numerator uint64         `json:"numerator"`
}

func (c *QBFTConfig) String() string {
	var blockReward, maxRequestTimeoutSeconds string

	if c.BlockReward != nil {
		blockReward = fmt.Sprintf("%v", c.BlockReward)
	} else {
		blockReward = "<nil>"
	}

	if c.MaxRequestTimeoutSeconds != nil {
		maxRequestTimeoutSeconds = fmt.Sprintf("%v", *c.MaxRequestTimeoutSeconds)
	} else {
		maxRequestTimeoutSeconds = "<nil>"
	}

	return fmt.Sprintf("{EpochLength: %v BlockPeriodSeconds: %v RequestTimeoutSeconds: %v, ProposerPolicy: %v, BlockReward: %v, BlockRewardBeneficiaries: %+v, Validators: %v, MinStakers: %v, TargetValidators: %v, MaxRequestTimeoutSeconds: %v}",
		c.EpochLength,
		c.BlockPeriodSeconds,
		c.RequestTimeoutSeconds,
		c.ProposerPolicy,
		blockReward,
		c.BlockRewardBeneficiary,
		c.Validators,
		c.MinStakers,
		c.TargetValidators,
		maxRequestTimeoutSeconds,
	)
}

type Transition struct {
	Block                        *big.Int              `json:"block"`
	EpochLength                  uint64                `json:"epochlength,omitempty"`                  // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds           uint64                `json:"blockperiodseconds,omitempty"`           // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	RequestTimeoutSeconds        uint64                `json:"requesttimeoutseconds,omitempty"`        // Minimum request timeout for each QBFT round in milliseconds
	ContractSizeLimit            uint64                `json:"contractsizelimit,omitempty"`            // Maximum smart contract code size
	Validators                   []common.Address      `json:"validators"`                             // List of validators
	EnhancedPermissioningEnabled *bool                 `json:"enhancedPermissioningEnabled,omitempty"` // aka QIP714Block
	PrivacyEnhancementsEnabled   *bool                 `json:"privacyEnhancementsEnabled,omitempty"`   // privacy enhancements (mandatory party, private state validation)
	PrivacyPrecompileEnabled     *bool                 `json:"privacyPrecompileEnabled,omitempty"`     // enable marker transactions support
	GasPriceEnabled              *bool                 `json:"gasPriceEnabled,omitempty"`              // enable gas price
	MinerGasLimit                uint64                `json:"miner.gaslimit,omitempty"`               // Gas Limit
	TransactionSizeLimit         uint64                `json:"transactionSizeLimit,omitempty"`         // Modify TransactionSizeLimit
	BlockReward                  *math.HexOrDecimal256 `json:"blockReward,omitempty"`                  // validation rewards
	BlockRewardBeneficiary       *BeneficiaryInfo      `json:"blockRewardBeneficiary,omitempty"`       // Reward beneficiaries
	MinStakers                   *uint64               `json:"minStakers,omitempty"`                   // Minimum number of stakers before stabilization
	TargetValidators             *uint64               `json:"targetValidators,omitempty"`             // Target number of validators
	MaxRequestTimeoutSeconds     *uint64               `json:"maxRequestTimeoutSeconds,omitempty"`     // The max a timeout should be for a round change
}

// gets value at or after a transition
func (c *ChainConfig) GetTransitionValue(num *big.Int, callback func(transition Transition)) {
	if c != nil && num != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(num) <= 0; i++ {
			callback(c.Transitions[i])
		}
	}
}

func (c *ChainConfig) GetBlockReward(num *big.Int) *big.Int {
	blockReward := big.NewInt(0)

	if c.QBFT != nil && c.QBFT.BlockReward != nil {
		blockReward = new(big.Int).Set((*big.Int)(c.QBFT.BlockReward))
	}

	c.GetTransitionValue(num, func(transition Transition) {
		if transition.BlockReward != nil {
			blockReward = new(big.Int).Set((*big.Int)(transition.BlockReward))
		}
	})

	return blockReward
}

// ## Quorum QBFT END

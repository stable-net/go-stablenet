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
	RequestTimeoutSeconds    uint64                `json:"requestTimeoutSeconds"`            // Minimum request timeout for each QBFT round in milliseconds
	BlockPeriodSeconds       uint64                `json:"blockPeriodSeconds"`               // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	ProposerPolicy           uint64                `json:"proposerPolicy"`                   // The policy for proposer selection
	EpochLength              uint64                `json:"epochLength"`                      // The duration during which a fixed validator set remains active
	BlockReward              *math.HexOrDecimal256 `json:"blockReward,omitempty"`            // Reward from start, works only on QBFT consensus protocol
	BlockRewardBeneficiary   *BeneficiaryInfo      `json:"blockRewardBeneficiary,omitempty"` // Reward beneficiaries
	Validators               []common.Address      `json:"validators"`                       // Validators list when the number of stakers is below the minimum stakers
	BLSPublicKeys            []string              `json:"blsPublicKeys"`                    // BLS PublicKey list of QBFTConfig.Validators
	TargetValidators         uint64                `json:"targetValidators"`                 // Target number of validators
	MaxRequestTimeoutSeconds *uint64               `json:"maxRequestTimeoutSeconds"`         // The max round time
	GovParams                *GovParams            `json:"govParams"`
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

var uint128Value, _ = new(big.Int).SetString("340282366920938463463374607431768211455", 10) //type(uint128).max;

var DefaultQBFTConfig = &QBFTConfig{
	RequestTimeoutSeconds: 2,
	BlockPeriodSeconds:    1,
	ProposerPolicy:        0,
	EpochLength:           10,
	BlockReward:           (*math.HexOrDecimal256)(new(big.Int).Mul(big.NewInt(Ether), big.NewInt(1))),
	GovParams: &GovParams{
		MinimumStaking:     (*math.HexOrDecimal256)(new(big.Int).Mul(big.NewInt(Ether), big.NewInt(500_000))),
		MaximumStaking:     (*math.HexOrDecimal256)(uint128Value),
		UnbondingStaker:    604800, // 7 days
		UnbondingDelegator: 259200, // 3 days
		FeePrecision:       10000,  // 0.01%
		ChangeFeeDelay:     604800, // 7 days
		MinStakers:         1,
	},
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

	return fmt.Sprintf("{EpochLength: %v BlockPeriodSeconds: %v RequestTimeoutSeconds: %v, ProposerPolicy: %v, BlockReward: %v, BlockRewardBeneficiaries: %+v, Validators: %v, BLSPublicKeys: %v, GovParams: %+v, TargetValidators: %v, MaxRequestTimeoutSeconds: %v}",
		c.EpochLength,
		c.BlockPeriodSeconds,
		c.RequestTimeoutSeconds,
		c.ProposerPolicy,
		blockReward,
		c.BlockRewardBeneficiary,
		c.Validators,
		c.BLSPublicKeys,
		c.GovParams,
		c.TargetValidators,
		maxRequestTimeoutSeconds,
	)
}

type Transition struct {
	Block                    *big.Int              `json:"block"`
	RequestTimeoutSeconds    uint64                `json:"requestTimeoutSeconds,omitempty"`  // Minimum request timeout for each QBFT round in milliseconds
	BlockPeriodSeconds       uint64                `json:"blockPeriodSeconds,omitempty"`     // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	EpochLength              uint64                `json:"epochLength,omitempty"`            // The duration during which a fixed validator set remains active
	BlockReward              *math.HexOrDecimal256 `json:"blockReward,omitempty"`            // Reward from start, works only on QBFT consensus protocol
	BlockRewardBeneficiary   *BeneficiaryInfo      `json:"blockRewardBeneficiary,omitempty"` // Reward beneficiaries
	GovParams                *GovParams            `json:"govParams,omitempty"`
	TargetValidators         *uint64               `json:"targetValidators,omitempty"`         // Target number of validators
	MaxRequestTimeoutSeconds *uint64               `json:"maxRequestTimeoutSeconds,omitempty"` // The max round time
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

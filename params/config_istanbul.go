// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/params/config.go (2024.07.25).
// Modified and improved for the wemix development.

package params

import (
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/log"
)

// ## Quorum QBFT START
const (
	ContractMode    = "contract"
	BlockHeaderMode = "blockheader"
)

type QBFTConfig struct {
	EpochLength              uint64                `json:"epochlength"`                       // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds       uint64                `json:"blockperiodseconds"`                // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	EmptyBlockPeriodSeconds  *uint64               `json:"emptyblockperiodseconds,omitempty"` // Minimum time between two consecutive QBFT a block and empty block’ timestamps in seconds
	RequestTimeoutSeconds    uint64                `json:"requesttimeoutseconds"`             // Minimum request timeout for each QBFT round in milliseconds
	ProposerPolicy           uint64                `json:"policy"`                            // The policy for proposer selection
	Ceil2Nby3Block           *big.Int              `json:"ceil2Nby3Block,omitempty"`          // Number of confirmations required to move from one state to next [2F + 1 to Ceil(2N/3)]
	BlockReward              *math.HexOrDecimal256 `json:"blockReward,omitempty"`             // Reward from start, works only on QBFT consensus protocol
	BeneficiaryMode          *string               `json:"beneficiaryMode,omitempty"`         // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	MiningBeneficiary        *common.Address       `json:"miningBeneficiary,omitempty"`       // Wallet address that benefits at every new block (besu mode)
	ValidatorSelectionMode   *string               `json:"validatorselectionmode,omitempty"`  // Select model for validators
	Validators               []common.Address      `json:"validators"`                        // Validators list
	MaxRequestTimeoutSeconds *uint64               `json:"maxRequestTimeoutSeconds"`          // The max round time
}

func (c QBFTConfig) String() string {
	return "qbft"
}

type Transition struct {
	Block                        *big.Int              `json:"block"`
	EpochLength                  uint64                `json:"epochlength,omitempty"`                  // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds           uint64                `json:"blockperiodseconds,omitempty"`           // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	EmptyBlockPeriodSeconds      *uint64               `json:"emptyblockperiodseconds,omitempty"`      // Minimum time between two consecutive QBFT a block and empty block’ timestamps in seconds
	RequestTimeoutSeconds        uint64                `json:"requesttimeoutseconds,omitempty"`        // Minimum request timeout for each QBFT round in milliseconds
	ContractSizeLimit            uint64                `json:"contractsizelimit,omitempty"`            // Maximum smart contract code size
	Validators                   []common.Address      `json:"validators"`                             // List of validators
	ValidatorSelectionMode       string                `json:"validatorselectionmode,omitempty"`       // Validator selection mode to switch to
	EnhancedPermissioningEnabled *bool                 `json:"enhancedPermissioningEnabled,omitempty"` // aka QIP714Block
	PrivacyEnhancementsEnabled   *bool                 `json:"privacyEnhancementsEnabled,omitempty"`   // privacy enhancements (mandatory party, private state validation)
	PrivacyPrecompileEnabled     *bool                 `json:"privacyPrecompileEnabled,omitempty"`     // enable marker transactions support
	GasPriceEnabled              *bool                 `json:"gasPriceEnabled,omitempty"`              // enable gas price
	MinerGasLimit                uint64                `json:"miner.gaslimit,omitempty"`               // Gas Limit
	TwoFPlusOneEnabled           *bool                 `json:"2FPlus1Enabled,omitempty"`               // Ceil(2N/3) is the default you need to explicitly use 2F + 1
	TransactionSizeLimit         uint64                `json:"transactionSizeLimit,omitempty"`         // Modify TransactionSizeLimit
	BlockReward                  *math.HexOrDecimal256 `json:"blockReward,omitempty"`                  // validation rewards
	BeneficiaryMode              *string               `json:"beneficiaryMode,omitempty"`              // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	MiningBeneficiary            *common.Address       `json:"miningBeneficiary,omitempty"`            // Wallet address that benefits at every new block (besu mode)
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

func (c *ChainConfig) GetRewardAccount(num *big.Int, coinbase common.Address) (common.Address, error) {
	beneficiaryMode := "validator"
	miningBeneficiary := common.Address{}

	if c.QBFT != nil && c.QBFT.MiningBeneficiary != nil {
		miningBeneficiary = *c.QBFT.MiningBeneficiary
		beneficiaryMode = "fixed"
	}

	if c.QBFT != nil && c.QBFT.BeneficiaryMode != nil {
		beneficiaryMode = *c.QBFT.BeneficiaryMode
	}

	c.GetTransitionValue(num, func(transition Transition) {
		if transition.BeneficiaryMode != nil && (*transition.BeneficiaryMode == "validators" || *transition.BeneficiaryMode == "validator") {
			beneficiaryMode = "validator"
		}
		if transition.MiningBeneficiary != nil && (transition.BeneficiaryMode == nil || *transition.BeneficiaryMode == "fixed") {
			miningBeneficiary = *transition.MiningBeneficiary
			beneficiaryMode = "fixed"
		}
	})

	switch strings.ToLower(beneficiaryMode) {
	case "fixed":
		log.Trace("fixed beneficiary mode", "miningBeneficiary", miningBeneficiary)
		return miningBeneficiary, nil
	case "validator":
		log.Trace("validator beneficiary mode", "coinbase", coinbase)
		return coinbase, nil
	}

	return common.Address{}, errors.New("BeneficiaryMode must be coinbase|fixed")
}

func (c *ChainConfig) GetBlockReward(num *big.Int) big.Int {
	blockReward := *math.NewHexOrDecimal256(0)

	if c.QBFT != nil && c.QBFT.BlockReward != nil {
		blockReward = *c.QBFT.BlockReward
	}

	c.GetTransitionValue(num, func(transition Transition) {
		if transition.BlockReward != nil {
			blockReward = *transition.BlockReward
		}
	})

	return big.Int(blockReward)
}

// ## Quorum QBFT END

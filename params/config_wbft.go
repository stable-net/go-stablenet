package params

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

const (
	GOV_CONFIG_ADDRESS       = "0x1000"
	GOV_STAKING_ADDRESS      = "0x1001"
	GOV_NCP_ADDRESS          = "0x1002"
	GOV_REWARDEE_IMP_ADDRESS = "0x1003"
)

type GovParams struct {
	MinimumStaking     *math.HexOrDecimal256 `json:"minimumStaking"`           // Minimum Staking Amount (WEI)
	MaximumStaking     *math.HexOrDecimal256 `json:"maximumStaking"`           // Maximum staking amount (WEI)
	UnbondingStaker    uint64                `json:"unbondingPeriodStaker"`    // Staker unbonding duration (seconds)
	UnbondingDelegator uint64                `json:"unbondingPeriodDelegator"` // Delegate unbundling period (seconds)
	FeePrecision       uint64                `json:"feePrecision"`             // Fee precision
	ChangeFeeDelay     uint64                `json:"changeFeeDelay"`           // Fee change latency (seconds)
	MinStakers         uint64                `json:"minStakers"`               // Minimum number of stakers
}

func (gp *GovParams) String() string {
	return fmt.Sprintf("{MinimumStaking: %v MaximumStaking: %v UnbondingStaker: %v UnbondingDelegator: %v FeePrecision: %v ChangeFeeDelay: %v MinStakers: %v}",
		((*big.Int)(gp.MinimumStaking)).String(),
		((*big.Int)(gp.MaximumStaking)).String(),
		gp.UnbondingStaker,
		gp.UnbondingDelegator,
		gp.FeePrecision,
		gp.ChangeFeeDelay,
		gp.MinStakers,
	)
}

type CodeParam struct {
	Address common.Address `json:"address"`
	Code    string         `json:"code"`
}

func (cp *CodeParam) String() string {
	return fmt.Sprintf("{Address: %v Code: %v}", cp.Address, cp.Code)
}

type StateParam struct {
	Address common.Address `json:"address"`
	Key     common.Hash    `json:"key"`
	Value   common.Hash    `json:"value"`
}

func (sp *StateParam) String() string {
	return fmt.Sprintf("{Address: %v Key: %v Value: %v}", sp.Address, sp.Key, sp.Value)
}

type StateTransition struct {
	Block  *big.Int     `json:"block"`
	Codes  []CodeParam  `json:"codes,omitempty"`
	States []StateParam `json:"states,omitempty"`
}

func (c *ChainConfig) GetStateTransitions(num *big.Int) []StateTransition {
	if c != nil && num != nil {
		transitions := make([]StateTransition, 0)

		for _, st := range c.StateTransitions {
			if st.Block.Cmp(num) == 0 {
				transitions = append(transitions, st)
			}
		}
		return transitions
	}
	return nil
}

func (st *StateTransition) String() string {
	return fmt.Sprintf("{Block: %v Codes: %v States: %v}", st.Block, st.Codes, st.States)
}

type MontBlancConfig struct {
	NCPs []common.Address `json:"ncps,omitempty"`
}

func (c *ChainConfig) GetBLSPublicKeys() [][]byte {
	blsPubKeys := make([][]byte, len(c.QBFT.BLSPublicKeys))
	for i, pk := range c.QBFT.BLSPublicKeys {
		blsPubKeys[i] = hexutil.MustDecode(pk)
	}
	return blsPubKeys
}

func (c *MontBlancConfig) String() string {
	return fmt.Sprintf("{NCPs: %v}",
		c.NCPs,
	)
}

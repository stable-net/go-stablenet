package params

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	GOV_CONST_ADDRESS        = "0x1000"
	GOV_STAKING_ADDRESS      = "0x1001"
	GOV_NCP_ADDRESS          = "0x1002"
	GOV_REWARDEE_IMP_ADDRESS = "0x1003"
)

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
	NCPs          []common.Address `json:"ncps,omitempty"`
	Validators    []common.Address `json:"validators"`    // Validators list when the number of stakers is below the minimum stakers
	BLSPublicKeys []string         `json:"blsPublicKeys"` // BLS PublicKey list of Validators
}

func (c *MontBlancConfig) GetBLSPublicKeys() [][]byte {
	blsPubKeys := make([][]byte, len(c.BLSPublicKeys))
	for i, pk := range c.BLSPublicKeys {
		blsPubKeys[i] = hexutil.MustDecode(pk)
	}
	return blsPubKeys
}

func (c *MontBlancConfig) String() string {
	return fmt.Sprintf("{NCPs: %v, Validators: %v, BLSPublicKeys: %v}",
		c.NCPs,
		c.Validators,
		c.BLSPublicKeys,
	)
}

// Copyright 2024 The go-wemix-wbft Authors
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
// This file is derived from quorum/params/config.go (2024.07.25).
// Modified and improved for the wemix development.

package params

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	DefaultGovValidatorAddress = common.HexToAddress("0x1001")
	DefaultGovVersion          = "v1"
)

var CheckSystemContractVersions func(systemContracts *SystemContracts) error

type WbftInit struct {
	Validators    []common.Address `json:"validators"`    // initial Wbft validators, order is matter
	BLSPublicKeys []string         `json:"blsPublicKeys"` // BLS public ket list of validators, order must be same as validators
}

type AnzeonConfig struct {
	Wbft            *WbftConfig      `json:"wbft"`
	Init            *WbftInit        `json:"init"`
	SystemContracts *SystemContracts `json:"systemContracts"`
}

func (c *AnzeonConfig) String() string {
	return fmt.Sprintf("{Wbft: %v Init: %v SystemContracts: %v}",
		c.Wbft,
		c.Init,
		c.SystemContracts)
}

func (c *AnzeonConfig) GetInitialBLSPublicKeys() [][]byte {
	blsPubKeys := make([][]byte, len(c.Init.BLSPublicKeys))
	for i, pk := range c.Init.BLSPublicKeys {
		blsPubKeys[i] = hexutil.MustDecode(pk)
	}
	return blsPubKeys
}

func (c *AnzeonConfig) CheckValidity() error {
	if c == nil {
		return errors.New("`anzeon`: missing `anzeon` section")
	}
	if c.Init == nil {
		return errors.New("`anzeon`: missing `init` section")
	}
	if c.Init.BLSPublicKeys == nil || len(c.Init.BLSPublicKeys) == 0 {
		return errors.New("`anzeon.init`: missing `blsPublicKeys` field")
	}
	if c.Init.Validators == nil || len(c.Init.Validators) == 0 {
		return errors.New("`anzeon.init`: missing `validators`")
	}
	if len(c.Init.Validators) != len(c.Init.BLSPublicKeys) {
		return fmt.Errorf(
			"`anzeon.init`: mismatched lengths: %d validators vs %d blsPublicKeys",
			len(c.Init.Validators), len(c.Init.BLSPublicKeys),
		)
	}
	if c.SystemContracts == nil {
		return errors.New("`anzeon: missing `systemContracts` section")
	}
	if c.SystemContracts.GovValidator == nil {
		return errors.New("`anzeon.systemContracts: missing `govValidator`")
	}
	if err := CheckSystemContractVersions(c.SystemContracts); err != nil {
		return fmt.Errorf("`anzeon.systemContracts`: %v", err)
	}

	if c.Wbft == nil {
		return errors.New("`anzeon`: missing `wBFT` section")
	}
	if c.Wbft.RequestTimeoutSeconds == 0 {
		return errors.New("`anzeon.wBFT`: `requestTimeoutSeconds` must be greater than 0")
	}
	if c.Wbft.BlockPeriodSeconds == 0 {
		return errors.New("`anzeon.wBFT`: `blockPeriodSeconds` must be greater than 0")
	}
	if c.Wbft.EpochLength <= 1 {
		return errors.New("`anzeon.wBFT`: `epochLength` must be greater than or equal to 2")
	}
	return nil
}

type Init struct {
	Validators      []common.Address `json:"validators"`      // initial Wbft validators, order is matter
	BLSPublicKeys   []string         `json:"blsPublicKeys"`   // BLS public ket list of validators, order must be same as validators
	SystemContracts *SystemContracts `json:"systemContracts"` // initial gov contracts, order must be same as validators
}

func (i *Init) String() string {
	return fmt.Sprintf("{Validators: %v BLSPublicKeys: %v SystemContracts: %v}",
		i.Validators,
		i.BLSPublicKeys,
		i.SystemContracts,
	)
}

type SystemContracts struct {
	GovValidator *SystemContract `json:"govValidator"`
}

func (c *SystemContracts) String() string {
	return fmt.Sprintf("{GovValidator: %v}",
		c.GovValidator,
	)
}

type SystemContract struct {
	Address common.Address    `json:"address"`
	Version string            `json:"version"`
	Params  map[string]string `json:"params"`
}

func (gc *SystemContract) String() string {
	return fmt.Sprintf("{Address: %v Version: %v Params: %v}",
		gc.Address,
		gc.Version,
		gc.Params,
	)
}

type Upgrade struct {
	Block            *big.Int `json:"block"`
	*SystemContracts `json:"systemContracts"`
}

func (u *Upgrade) String() string {
	return fmt.Sprintf("{Block: %v SystemContracts: %v}",
		u.Block.String(),
		u.SystemContracts.String(),
	)
}

type WbftConfig struct {
	RequestTimeoutSeconds    uint64  `json:"requestTimeoutSeconds"`            // Minimum request timeout for each Wbft round in milliseconds
	BlockPeriodSeconds       uint64  `json:"blockPeriodSeconds"`               // Minimum time between two consecutive Wbft blocks’ timestamps in seconds
	EpochLength              uint64  `json:"epochLength"`                      // The duration during which a fixed validator set remains active
	AllowedFutureBlockTime   uint64  `json:"allowedFutureBlockTime,omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	ProposerPolicy           *uint64 `json:"proposerPolicy"`                   // The policy for proposer selection
	MaxRequestTimeoutSeconds *uint64 `json:"maxRequestTimeoutSeconds"`         // The max round time
}

type Transition struct {
	Block *big.Int `json:"block"`
	*WbftConfig
}

func (t *Transition) String() string {
	return fmt.Sprintf("{Block: %v RequestTimeoutSeconds: %v BlockPeriodSeconds: %v EpochLength: %v MaxRequestTimeoutSeconds: %v}",
		t.Block.String(),
		t.RequestTimeoutSeconds,
		t.BlockPeriodSeconds,
		t.EpochLength,
		t.MaxRequestTimeoutSeconds,
	)
}

var DefaultAnzeonConfig = &AnzeonConfig{
	Wbft: &WbftConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		ProposerPolicy:        newUint64(0),
		EpochLength:           10,
	},
	SystemContracts: &SystemContracts{
		GovValidator: &SystemContract{
			Address: DefaultGovValidatorAddress,
			Version: DefaultGovVersion,
		},
	},
	Init: &WbftInit{},
}

func (c *WbftConfig) String() string {
	var maxRequestTimeoutSeconds string

	if c.MaxRequestTimeoutSeconds != nil {
		maxRequestTimeoutSeconds = fmt.Sprintf("%v", *c.MaxRequestTimeoutSeconds)
	} else {
		maxRequestTimeoutSeconds = "<nil>"
	}

	return fmt.Sprintf("{EpochLength: %v BlockPeriodSeconds: %v RequestTimeoutSeconds: %v, ProposerPolicy: %v, MaxRequestTimeoutSeconds: %v}",
		c.EpochLength,
		c.BlockPeriodSeconds,
		c.RequestTimeoutSeconds,
		c.ProposerPolicy,
		maxRequestTimeoutSeconds,
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

func (st *StateTransition) String() string {
	return fmt.Sprintf("{Block: %v Codes: %v States: %v}", st.Block, st.Codes, st.States)
}

// ## Anzeon CHAIN CONFIG END

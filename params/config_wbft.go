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
	DefaultGovValidatorAddress = common.HexToAddress("0x1000")
	DefaultGovVersion          = "v1"
)

var CheckGovContractVersions func(govContracts *GovContracts) error

type WbftInit struct {
	Validators    []common.Address `json:"validators"`    // initial WBFT validators, order is matter
	BLSPublicKeys []string         `json:"blsPublicKeys"` // BLS public ket list of validators, order must be same as validators
}

type CroissantConfig struct {
	WBFT         *WBFTConfig   `json:"wBFT"`
	Init         *WbftInit     `json:"init"`
	GovContracts *GovContracts `json:"govContracts"`
}

func (c *CroissantConfig) String() string {
	return fmt.Sprintf("{WBFT: %v Init: %v GovContracts: %v}",
		c.WBFT,
		c.Init,
		c.GovContracts)
}

func (c *CroissantConfig) GetInitialBLSPublicKeys() [][]byte {
	blsPubKeys := make([][]byte, len(c.Init.BLSPublicKeys))
	for i, pk := range c.Init.BLSPublicKeys {
		blsPubKeys[i] = hexutil.MustDecode(pk)
	}
	return blsPubKeys
}

func (c *CroissantConfig) CheckValidity() error {
	if c == nil {
		return errors.New("`croissant`: missing `croissant` section")
	}
	if c.Init == nil {
		return errors.New("`croissant`: missing `init` section")
	}
	if c.Init.BLSPublicKeys == nil || len(c.Init.BLSPublicKeys) == 0 {
		return errors.New("`croissant.init`: missing `blsPublicKeys` field")
	}
	if c.Init.Validators == nil || len(c.Init.Validators) == 0 {
		return errors.New("`croissant.init`: missing `validators`")
	}
	if len(c.Init.Validators) != len(c.Init.BLSPublicKeys) {
		return fmt.Errorf(
			"`croissant.init`: mismatched lengths: %d validators vs %d blsPublicKeys",
			len(c.Init.Validators), len(c.Init.BLSPublicKeys),
		)
	}
	if c.GovContracts == nil {
		return errors.New("`croissant: missing `govContracts` section")
	}
	if c.GovContracts.GovValidator == nil {
		return errors.New("`croissant.govContracts: missing `govValidator`")
	}
	if err := CheckGovContractVersions(c.GovContracts); err != nil {
		return fmt.Errorf("`croissant.govContracts`: %v", err)
	}

	if c.WBFT == nil {
		return errors.New("`croissant`: missing `wBFT` section")
	}
	if c.WBFT.RequestTimeoutSeconds == 0 {
		return errors.New("`croissant.wBFT`: `requestTimeoutSeconds` must be greater than 0")
	}
	if c.WBFT.BlockPeriodSeconds == 0 {
		return errors.New("`croissant.wBFT`: `blockPeriodSeconds` must be greater than 0")
	}
	if c.WBFT.EpochLength <= 1 {
		return errors.New("`croissant.wBFT`: `epochLength` must be greater than or equal to 2")
	}
	if c.WBFT.StabilizingStakersThreshold == nil {
		return errors.New("`croissant.wBFT`: missing `stabilizingStakersThreshold`")
	} else if *c.WBFT.StabilizingStakersThreshold == 0 {
		return errors.New("`croissant.wBFT`: `stabilizingStakersThreshold` must be greater than 0")
	}
	if c.WBFT.TargetValidators == nil {
		return errors.New("`croissant.wBFT`: missing `targetValidators`")
	} else if c.WBFT.EpochLength < *c.WBFT.TargetValidators {
		return fmt.Errorf("`croissant.wBFT`: `epochLength` (%d) must be greater than or equal to `targetValidators` (%d)",
			c.WBFT.EpochLength, *c.WBFT.TargetValidators)
	}
	return nil
}

type Init struct {
	Validators    []common.Address `json:"validators"`    // initial WBFT validators, order is matter
	BLSPublicKeys []string         `json:"blsPublicKeys"` // BLS public ket list of validators, order must be same as validators
	GovContracts  *GovContracts    `json:"govContracts"`  // initial gov contracts, order must be same as validators
}

func (i *Init) String() string {
	return fmt.Sprintf("{Validators: %v BLSPublicKeys: %v GovContracts: %v}",
		i.Validators,
		i.BLSPublicKeys,
		i.GovContracts,
	)
}

type GovContracts struct {
	GovValidator *GovContract `json:"govValidator"`
}

func (c *GovContracts) String() string {
	return fmt.Sprintf("{GovValidator: %v}",
		c.GovValidator,
	)
}

type GovContract struct {
	Address common.Address    `json:"address"`
	Version string            `json:"version"`
	Params  map[string]string `json:"params"`
}

func (gc *GovContract) String() string {
	return fmt.Sprintf("{Address: %v Version: %v Params: %v}",
		gc.Address,
		gc.Version,
		gc.Params,
	)
}

type Upgrade struct {
	Block         *big.Int `json:"block"`
	*GovContracts `json:"govContracts"`
}

func (u *Upgrade) String() string {
	return fmt.Sprintf("{Block: %v GovContracts: %v}",
		u.Block.String(),
		u.GovContracts.String(),
	)
}

type WBFTConfig struct {
	RequestTimeoutSeconds       uint64  `json:"requestTimeoutSeconds"`            // Minimum request timeout for each WBFT round in milliseconds
	BlockPeriodSeconds          uint64  `json:"blockPeriodSeconds"`               // Minimum time between two consecutive WBFT blocks’ timestamps in seconds
	EpochLength                 uint64  `json:"epochLength"`                      // The duration during which a fixed validator set remains active
	AllowedFutureBlockTime      uint64  `json:"allowedFutureBlockTime,omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	ProposerPolicy              *uint64 `json:"proposerPolicy"`                   // The policy for proposer selection
	TargetValidators            *uint64 `json:"targetValidators"`                 // Target number of validators
	MaxRequestTimeoutSeconds    *uint64 `json:"maxRequestTimeoutSeconds"`         // The max round time
	StabilizingStakersThreshold *uint64 `json:"stabilizingStakersThreshold"`      // initial stabilizing stakers threshold, default is 1
	UseNCP                      *bool   `json:"useNCP"`                           // Use NCP or not
}

type Transition struct {
	Block *big.Int `json:"block"`
	*WBFTConfig
}

func (t *Transition) String() string {
	return fmt.Sprintf("{Block: %v RequestTimeoutSeconds: %v BlockPeriodSeconds: %v EpochLength: %v TargetValidators: %v MaxRequestTimeoutSeconds: %v}",
		t.Block.String(),
		t.RequestTimeoutSeconds,
		t.BlockPeriodSeconds,
		t.EpochLength,
		t.TargetValidators,
		t.MaxRequestTimeoutSeconds,
	)
}

var DefaultCroissantConfig = &CroissantConfig{
	WBFT: &WBFTConfig{
		RequestTimeoutSeconds:       2,
		BlockPeriodSeconds:          1,
		ProposerPolicy:              newUint64(0),
		EpochLength:                 10,
		TargetValidators:            newUint64(1),
		StabilizingStakersThreshold: newUint64(1),
		UseNCP:                      newBool(false),
	},
	GovContracts: &GovContracts{
		GovValidator: &GovContract{
			Address: common.HexToAddress("0x1000"),
			Version: "v1",
		},
	},
	Init: &WbftInit{},
}

func (c *WBFTConfig) String() string {
	var maxRequestTimeoutSeconds string

	if c.MaxRequestTimeoutSeconds != nil {
		maxRequestTimeoutSeconds = fmt.Sprintf("%v", *c.MaxRequestTimeoutSeconds)
	} else {
		maxRequestTimeoutSeconds = "<nil>"
	}

	return fmt.Sprintf("{EpochLength: %v BlockPeriodSeconds: %v RequestTimeoutSeconds: %v, ProposerPolicy: %v, TargetValidators: %v, MaxRequestTimeoutSeconds: %v, StabilizingStakersThreshold: %v, UseNCP: %v}",
		c.EpochLength,
		c.BlockPeriodSeconds,
		c.RequestTimeoutSeconds,
		c.ProposerPolicy,
		c.TargetValidators,
		maxRequestTimeoutSeconds,
		c.StabilizingStakersThreshold,
		c.UseNCP,
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

// ## Croissant CHAIN CONFIG END

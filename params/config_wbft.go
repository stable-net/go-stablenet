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
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	DefaultGovConfigAddress      = common.HexToAddress("0x1000")
	DefaultGovStakingAddress     = common.HexToAddress("0x1001")
	DefaultGovRewardeeImpAddress = common.HexToAddress("0x1002")
	DefaultGovNCPAddress         = common.HexToAddress("0x1003")
	DefaultGovVersion            = "v1"
	DefaultGovConfigParams       = map[string]string{
		"minimumStaking":           "10000000000000000000000000",
		"maximumStaking":           "100000000000000000000000000",
		"unbondingPeriodStaker":    "604800", // 7 days
		"unbondingPeriodDelegator": "259200", // 3 days
		"feePrecision":             "10000",  // 0.01%
		"changeFeeDelay":           "604800", // 7 days
		"govCouncil":               common.HexToAddress("0x0").String(),
	}
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
	if c.GovContracts.GovStaking == nil {
		return errors.New("`croissant.govContracts: missing `govStaking`")
	}
	if c.GovContracts.GovConfig == nil {
		return errors.New("`croissant.govContracts: missing `govConfig`")
	}
	if c.GovContracts.GovRewardeeImp == nil {
		return errors.New("`croissant.govContracts: missing `govRewardeeImp`")
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
	if c.WBFT.EpochLength == 0 {
		return errors.New("`croissant.wBFT`: `epochLength` must be greater than 0")
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

	if err := checkSanityBeneficiaries(c.WBFT.BlockRewardBeneficiary); err != nil {
		return fmt.Errorf("`croissant.wBFT`: %v", err)
	}
	return nil
}

func checkSanityBeneficiaries(l *BeneficiaryInfo) error {
	var totNumerator uint64

	if l == nil {
		return nil
	}

	if l.Denominator == 0 {
		return fmt.Errorf("Denominator cannot be zero")
	}

	for _, beneficiary := range l.Beneficiaries {
		if beneficiary.Addr == (common.Address{}) {
			return fmt.Errorf("Beneficiary address cannot be zero address")
		}
		if beneficiary.Numerator > l.Denominator {
			return fmt.Errorf("Numerator (%v) > denominator (%v)", beneficiary.Numerator, l.Denominator)
		}
		totNumerator += beneficiary.Numerator
	}

	if totNumerator > l.Denominator {
		return fmt.Errorf("Total of numerator (%v) > denominator (%v)", totNumerator, l.Denominator)
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
	GovConfig      *GovContract `json:"govConfig"`
	GovStaking     *GovContract `json:"govStaking"`
	GovRewardeeImp *GovContract `json:"govRewardeeImp"`
	GovNCP         *GovContract `json:"govNCP"`
}

func (c *GovContracts) String() string {
	return fmt.Sprintf("{GovConfig: %v GovStaking: %v GovRewardeeImp: %v GovNCP: %v}",
		c.GovConfig,
		c.GovStaking,
		c.GovRewardeeImp,
		c.GovNCP,
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
	RequestTimeoutSeconds       uint64                `json:"requestTimeoutSeconds"`            // Minimum request timeout for each WBFT round in milliseconds
	BlockPeriodSeconds          uint64                `json:"blockPeriodSeconds"`               // Minimum time between two consecutive WBFT blocks’ timestamps in seconds
	EpochLength                 uint64                `json:"epochLength"`                      // The duration during which a fixed validator set remains active
	BlockReward                 *math.HexOrDecimal256 `json:"blockReward,omitempty"`            // Reward from start, works only on WBFT consensus protocol
	AllowedFutureBlockTime      uint64                `json:"allowedFutureBlockTime,omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BlockRewardBeneficiary      *BeneficiaryInfo      `json:"blockRewardBeneficiary,omitempty"` // Reward beneficiaries
	ProposerPolicy              *uint64               `json:"proposerPolicy"`                   // The policy for proposer selection
	TargetValidators            *uint64               `json:"targetValidators"`                 // Target number of validators
	MaxRequestTimeoutSeconds    *uint64               `json:"maxRequestTimeoutSeconds"`         // The max round time
	StabilizingStakersThreshold *uint64               `json:"stabilizingStakersThreshold"`      // initial stabilizing stakers threshold, default is 1
	UseNCP                      *bool                 `json:"useNCP"`                           // Use NCP or not
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

type Transition struct {
	Block *big.Int `json:"block"`
	*WBFTConfig
}

func (t *Transition) String() string {
	return fmt.Sprintf("{Block: %v RequestTimeoutSeconds: %v BlockPeriodSeconds: %v EpochLength: %v BlockReward: %v BlockRewardBeneficiary: %+v TargetValidators: %v MaxRequestTimeoutSeconds: %v}",
		t.Block.String(),
		t.RequestTimeoutSeconds,
		t.BlockPeriodSeconds,
		t.EpochLength,
		t.BlockReward,
		t.BlockRewardBeneficiary,
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
		BlockReward:                 (*math.HexOrDecimal256)(new(big.Int).Mul(big.NewInt(Ether), big.NewInt(1))),
		TargetValidators:            newUint64(1),
		StabilizingStakersThreshold: newUint64(1),
		UseNCP:                      newBool(false),
	},
	GovContracts: &GovContracts{
		GovConfig: &GovContract{
			Address: common.HexToAddress("0x1000"),
			Version: "v1",
			Params: map[string]string{
				"minimumStaking":           "10000000000000000000000000",
				"maximumStaking":           "100000000000000000000000000",
				"unbondingPeriodStaker":    "604800", // 7 days
				"unbondingPeriodDelegator": "259200", // 3 days
				"feePrecision":             "10000",  // 0.01%
				"changeFeeDelay":           "604800", // 7 days
			},
		},
		GovStaking: &GovContract{
			Address: common.HexToAddress("0x1001"),
			Version: "v1",
		},
		GovRewardeeImp: &GovContract{
			Address: common.HexToAddress("0x1002"),
			Version: "v1",
		},
	},
	Init: &WbftInit{},
}

func (c *WBFTConfig) String() string {
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

	return fmt.Sprintf("{EpochLength: %v BlockPeriodSeconds: %v RequestTimeoutSeconds: %v, ProposerPolicy: %v, BlockReward: %v, BlockRewardBeneficiaries: %+v, TargetValidators: %v, MaxRequestTimeoutSeconds: %v, StabilizingStakersThreshold: %v, UseNCP: %v}",
		c.EpochLength,
		c.BlockPeriodSeconds,
		c.RequestTimeoutSeconds,
		c.ProposerPolicy,
		blockReward,
		c.BlockRewardBeneficiary,
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

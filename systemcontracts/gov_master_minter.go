// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package systemcontracts

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_MASTER_MINTER_PARAM_FIAT_TOKEN           = "fiatToken"
	GOV_MASTER_MINTER_PARAM_MINTERS              = "minters"
	GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE = "maxMinterAllowance"

	// GovMasterMinter Storage Layout (extends GovBaseV2):
	// Slots 0x0-0xb: GovBaseV2 base storage
	// Slots 0xc-0x31: __gap (reserved)
	// Slot 0x32: fiatToken (address, 20 bytes)
	// Slot 0x33: emergencyPaused (bool, 1 byte)
	// Slot 0x34: maxMinterAllowance (uint256, 32 bytes)
	// Slot 0x35: isMinter (mapping(address => bool))
	// Slot 0x36: minterList (address[])
	// Slot 0x37: minterIndex (mapping(address => uint256))
	// Note: minterAllowances and totalMinterAllowance removed - FiatToken is source of truth
	SLOT_GOV_MASTER_MINTER_fiatToken          = "0x32"
	SLOT_GOV_MASTER_MINTER_emergencyPaused    = "0x33"
	SLOT_GOV_MASTER_MINTER_maxMinterAllowance = "0x34"
	SLOT_GOV_MASTER_MINTER_isMinter           = "0x35"
	SLOT_GOV_MASTER_MINTER_minterList         = "0x36"
	SLOT_GOV_MASTER_MINTER_minterIndex        = "0x37"
)

// Default maxMinterAllowance: 10B tokens (10000000000 * 10^18)
var DefaultMaxMinterAllowance = new(big.Int).Mul(
	big.NewInt(10000000000),
	new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
)

// initializeMasterMinter initializes the GovMasterMinter contract storage
func initializeMasterMinter(govMasterMinterAddress common.Address, param map[string]string) ([]params.StateParam, error) {
	// Initialize GovBase first
	sp, err := initializeBase(govMasterMinterAddress, param)
	if err != nil {
		return sp, err
	}

	// Initialize fiatToken address
	if fiatTokenStr, ok := param[GOV_MASTER_MINTER_PARAM_FIAT_TOKEN]; ok {
		fiatToken := common.HexToAddress(fiatTokenStr)
		if fiatToken == (common.Address{}) {
			return nil, fmt.Errorf("`systemContracts.govMasterMinter.params.fiatToken`: invalid address")
		}

		fiatTokenSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_fiatToken)
		sp = append(sp, params.StateParam{
			Address: govMasterMinterAddress,
			Key:     fiatTokenSlot,
			Value:   common.BytesToHash(fiatToken.Bytes()),
		})
	}

	// Initialize minters (comma-separated addresses)
	if mintersStr, ok := param[GOV_MASTER_MINTER_PARAM_MINTERS]; ok {
		minterAddressStrs := strings.Split(mintersStr, ",")
		if len(minterAddressStrs) == 0 {
			return nil, fmt.Errorf("`systemContracts.govMasterMinter.params.minters`: no addresses provided")
		}

		// Parse and validate all addresses
		var minterAddresses []common.Address
		for i, addrStr := range minterAddressStrs {
			minter := common.HexToAddress(strings.TrimSpace(addrStr))
			if minter == (common.Address{}) {
				return nil, fmt.Errorf("`systemContracts.govMasterMinter.params.minters[%d]`: invalid address %q", i, addrStr)
			}
			minterAddresses = append(minterAddresses, minter)
		}

		// Set minterList length (slot 0x36)
		minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
		sp = append(sp, params.StateParam{
			Address: govMasterMinterAddress,
			Key:     minterListSlot,
			Value:   common.BigToHash(big.NewInt(int64(len(minterAddresses)))),
		})

		// Set each minter in the array and mappings
		for i, minter := range minterAddresses {
			// Set minterList[i] = minter (keccak256(0x36) + i)
			arrayElementKey := CalculateDynamicSlot(minterListSlot, big.NewInt(int64(i)))
			sp = append(sp, params.StateParam{
				Address: govMasterMinterAddress,
				Key:     arrayElementKey,
				Value:   common.BytesToHash(minter.Bytes()),
			})

			// Set isMinter[minter] = true (slot 0x35)
			isMinterSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_isMinter)
			isMinterKey := CalculateMappingSlot(isMinterSlot, minter)
			sp = append(sp, params.StateParam{
				Address: govMasterMinterAddress,
				Key:     isMinterKey,
				Value:   common.BigToHash(big.NewInt(1)),
			})

			// Set minterIndex[minter] = i + 1 (1-based indexing, slot 0x37)
			minterIndexSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterIndex)
			minterIndexKey := CalculateMappingSlot(minterIndexSlot, minter)
			sp = append(sp, params.StateParam{
				Address: govMasterMinterAddress,
				Key:     minterIndexKey,
				Value:   common.BigToHash(big.NewInt(int64(i + 1))),
			})
		}
	}

	// Initialize maxMinterAllowance (optional, defaults to 10B tokens)
	maxMinterAllowance := DefaultMaxMinterAllowance
	if maxAllowanceStr, ok := param[GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE]; ok {
		var success bool
		maxMinterAllowance, success = new(big.Int).SetString(maxAllowanceStr, 10)
		if !success {
			return nil, fmt.Errorf("`systemContracts.govMasterMinter.params.maxMinterAllowance`: invalid number")
		}
		if maxMinterAllowance.Sign() <= 0 {
			return nil, fmt.Errorf("`systemContracts.govMasterMinter.params.maxMinterAllowance`: must be positive")
		}
	}

	maxMinterAllowanceSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance)
	sp = append(sp, params.StateParam{
		Address: govMasterMinterAddress,
		Key:     maxMinterAllowanceSlot,
		Value:   common.BigToHash(maxMinterAllowance),
	})

	// emergencyPaused defaults to false in slot 0x33 (no initialization needed)
	// All uninitialized storage slots default to 0/false, so explicit initialization is not required

	return sp, nil
}

// Note: GetMinterAllowance removed - query FiatToken directly for minter allowances
// Allowances are managed by FiatToken contract, not GovMasterMinter

// IsMinter checks if an address is registered as a minter
func IsMinter(govMasterMinterAddress common.Address, state StateReader, minter common.Address) bool {
	isMinterSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_isMinter)
	key := CalculateMappingSlot(isMinterSlot, minter)
	value := state.GetState(govMasterMinterAddress, key)
	return value.Big().Uint64() != 0
}

// GetMinterCount returns the total number of minters
func GetMinterCount(govMasterMinterAddress common.Address, state StateReader) uint64 {
	minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
	lengthValue := state.GetState(govMasterMinterAddress, minterListSlot)
	return lengthValue.Big().Uint64()
}

// GetMinterAt returns the minter address at the given index
func GetMinterAt(govMasterMinterAddress common.Address, state StateReader, index uint64) common.Address {
	minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
	key := CalculateDynamicSlot(minterListSlot, new(big.Int).SetUint64(index))
	value := state.GetState(govMasterMinterAddress, key)
	return common.BytesToAddress(value.Bytes())
}

// Note: GetTotalMinterAllowance removed - calculate sum by querying FiatToken for each minter
// Total allowance is computed on-demand in getMinterStats() view function

// GetMaxMinterAllowance returns the maximum allowance per minter
func GetMaxMinterAllowance(govMasterMinterAddress common.Address, state StateReader) *big.Int {
	maxMinterAllowanceSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance)
	value := state.GetState(govMasterMinterAddress, maxMinterAllowanceSlot)
	return value.Big()
}

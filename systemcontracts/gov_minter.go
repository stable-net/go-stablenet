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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_MINTER_PARAM_FIAT_TOKEN = "fiatToken"

	// GovMinter Storage Layout (extends GovBase):
	// Slots 0x0-0xb: GovBase base storage
	// Slots 0xc-0x31: __gap (reserved)
	// Slot 0x32: fiatToken (address, 20 bytes)
	// Slot 0x33: usedProofHashes (mapping(bytes32 => bool))
	// Slot 0x34: depositIdToProposalId (mapping(string => uint256))
	// Slot 0x35: executedDepositIds (mapping(string => bool))
	// Slot 0x36: withdrawalIdToProposalId (mapping(string => uint256))
	// Slot 0x37: executedWithdrawalIds (mapping(string => bool))
	// Slot 0x38: burnProposals (mapping(uint256 => BurnProposalData))
	// Slot 0x39: reservedMintAmount (uint256) - P0-1 security fix
	// Slot 0x3a: mintProposalAmounts (mapping(uint256 => uint256))
	// Slot 0x3b: burnBalance (mapping(address => uint256))
	// Slot 0x3c: emergencyPaused (bool, 1 byte)
	// Slot 0x3d: refundableBalance (mapping(address => uint256))
	SLOT_GOV_MINTER_fiatToken                = "0x32"
	SLOT_GOV_MINTER_usedProofHashes          = "0x33"
	SLOT_GOV_MINTER_depositIdToProposalId    = "0x34"
	SLOT_GOV_MINTER_executedDepositIds       = "0x35"
	SLOT_GOV_MINTER_withdrawalIdToProposalId = "0x36"
	SLOT_GOV_MINTER_executedWithdrawalIds    = "0x37"
	SLOT_GOV_MINTER_burnProposals            = "0x38"
	SLOT_GOV_MINTER_reservedMintAmount       = "0x39"
	SLOT_GOV_MINTER_mintProposalAmounts      = "0x3a"
	SLOT_GOV_MINTER_burnBalance              = "0x3b"
	SLOT_GOV_MINTER_emergencyPaused          = "0x3c"
	SLOT_GOV_MINTER_refundableBalance        = "0x3d"
)

// MintProof represents the proof data for minting operations
type MintProof struct {
	Beneficiary   common.Address
	Amount        *big.Int
	Timestamp     *big.Int
	DepositId     string
	BankReference string
	Memo          string
}

// BurnProof represents the proof data for burning operations
type BurnProof struct {
	From         common.Address
	Amount       *big.Int
	Timestamp    *big.Int
	WithdrawalId string
	ReferenceId  string
	Memo         string
}

// BurnProposalData represents the data stored for burn proposals
type BurnProposalData struct {
	Amount    *big.Int
	Requester common.Address
}

// initializeMinter initializes the GovMinter contract storage
func initializeMinter(govMinterAddress common.Address, param map[string]string) ([]params.StateParam, error) {
	// Initialize GovBase first
	sp, err := initializeBase(govMinterAddress, param)
	if err != nil {
		return sp, err
	}

	// Initialize fiatToken address
	if fiatTokenStr, ok := param[GOV_MINTER_PARAM_FIAT_TOKEN]; ok {
		fiatToken := common.HexToAddress(fiatTokenStr)
		if fiatToken == (common.Address{}) {
			return nil, fmt.Errorf("`systemContracts.govMinter.params.fiatToken`: invalid address")
		}

		// fiatToken is stored in slot 0x32 (address takes 20 bytes, leftmost in the slot)
		// emergencyPaused (bool) would be packed in the same slot but defaults to false (0)
		fiatTokenSlot := common.HexToHash(SLOT_GOV_MINTER_fiatToken)
		sp = append(sp, params.StateParam{
			Address: govMinterAddress,
			Key:     fiatTokenSlot,
			Value:   common.BytesToHash(fiatToken.Bytes()),
		})
	}

	return sp, nil
}


// GetReservedMintAmount returns the total reserved mint amount
func GetReservedMintAmount(govMinterAddress common.Address, state StateReader) *big.Int {
	slot := common.HexToHash(SLOT_GOV_MINTER_reservedMintAmount)
	value := state.GetState(govMinterAddress, slot)
	return value.Big()
}

// GetMintProposalAmount returns the mint amount reserved for a specific proposal
func GetMintProposalAmount(govMinterAddress common.Address, state StateReader, proposalId *big.Int) *big.Int {
	mintProposalAmountsSlot := common.HexToHash(SLOT_GOV_MINTER_mintProposalAmounts)
	key := CalculateMappingSlot(mintProposalAmountsSlot, common.BigToHash(proposalId))
	value := state.GetState(govMinterAddress, key)
	return value.Big()
}

// GetBurnBalance returns the burn balance for a given address
func GetBurnBalance(govMinterAddress common.Address, state StateReader, addr common.Address) *big.Int {
	burnBalanceSlot := common.HexToHash(SLOT_GOV_MINTER_burnBalance)
	key := CalculateMappingSlot(burnBalanceSlot, addr)
	value := state.GetState(govMinterAddress, key)
	return value.Big()
}

// GetRefundableBalance returns the refundable burn deposit balance for a given address
func GetRefundableBalance(govMinterAddress common.Address, state StateReader, addr common.Address) *big.Int {
	refundableBalanceSlot := common.HexToHash(SLOT_GOV_MINTER_refundableBalance)
	key := CalculateMappingSlot(refundableBalanceSlot, addr)
	value := state.GetState(govMinterAddress, key)
	return value.Big()
}

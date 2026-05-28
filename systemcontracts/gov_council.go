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
	"bytes"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_COUNCIL_PARAM_BLACKLIST           = "blacklist"
	GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS = "authorizedAccounts"

	// GovCouncil Storage Layout (extends GovBase):
	// Slots 0x0-0x31: GovBase storage (0-49)
	// Slots 0x32-0x35: GovCouncil storage (50-53)
	//
	// AddressSet storage structure (based on AddressSetLib.sol):
	// Each AddressSet occupies 2 consecutive slots:
	// - Base slot: contains array length (for _values dynamic array)
	// - Base slot + 1: mapping slot (for _positions mapping)
	//
	// Slot 0x32 (50): _currentBlacklist base slot
	//   - Array values stored at: keccak256(0x32) + index
	//   - Position mapping: keccak256(key || 0x33)
	// Slot 0x33 (51): _currentBlacklist positions mapping slot
	//
	// Slot 0x34 (52): _currentAuthorizedAccounts base slot
	//   - Array values stored at: keccak256(0x34) + index
	//   - Position mapping: keccak256(key || 0x35)
	// Slot 0x35 (53): _currentAuthorizedAccounts positions mapping slot
	//
	// Slot 0x36 (54): __accountManager address
	//   - Stores AccountManagerAddress (params.AccountManagerAddress)
	SLOT_GOV_COUNCIL_currentBlacklist_values             = "0x32" // Base slot for blacklist AddressSet
	SLOT_GOV_COUNCIL_currentBlacklist_positions          = "0x33" // Positions mapping slot for blacklist
	SLOT_GOV_COUNCIL_currentAuthorizedAccounts_values    = "0x34" // Base slot for authorized accounts AddressSet
	SLOT_GOV_COUNCIL_currentAuthorizedAccounts_positions = "0x35" // Positions mapping slot for authorized accounts
	SLOT_GOV_COUNCIL_accountManager                      = "0x36" // AccountManager address slot
)

// initializeGovCouncil initializes the GovCouncil contract storage.
// Reconciles blacklist and authorized account addresses from two sources:
// params (chain config) and alloc.Extra bits (genesis account state).
// The final sets are the union of both sources, applied to both alloc.Extra
// and the contract storage slots.
func initializeGovCouncil(govCouncilAddress common.Address, param map[string]string, alloc *types.GenesisAlloc) ([]params.StateParam, error) {
	// Initialize GovBase first
	sp, err := initializeBase(govCouncilAddress, param)
	if err != nil {
		return sp, err
	}

	type addrSetSpec struct {
		paramKey   string
		errPrefix  string
		isSet      func(uint64) bool
		setExtra   func(uint64) uint64
		valuesSlot common.Hash
		posSlot    common.Hash
	}
	specs := []addrSetSpec{
		{
			paramKey:   GOV_COUNCIL_PARAM_BLACKLIST,
			errPrefix:  "`systemContracts.govCouncil.params.blacklist`",
			isSet:      types.IsBlacklisted,
			setExtra:   types.SetBlacklisted,
			valuesSlot: common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values),
			posSlot:    common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_positions),
		},
		{
			paramKey:   GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS,
			errPrefix:  "`systemContracts.govCouncil.params.authorizedAccounts`",
			isSet:      types.IsAuthorized,
			setExtra:   types.SetAuthorized,
			valuesSlot: common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_values),
			posSlot:    common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_positions),
		},
	}

	// Build address sets from params first.
	addrSets := make([]map[common.Address]struct{}, len(specs))
	for i, spec := range specs {
		addrSets[i], err = parseParamAddresses(param[spec.paramKey])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", spec.errPrefix, err)
		}
	}

	if alloc != nil {
		// Merge alloc Extra bits into the param-seeded sets.
		// Zero addresses in alloc are skipped because they are unrelated to these sets.
		for addr, account := range *alloc {
			if addr == (common.Address{}) {
				continue
			}
			for i, spec := range specs {
				if spec.isSet(account.Extra) {
					addrSets[i][addr] = struct{}{}
				}
			}
		}

		// Sync alloc.Extra with the final merged sets.
		// Entries that only exist in params are added to alloc with a zero balance.
		for i, spec := range specs {
			for addr := range addrSets[i] {
				account := (*alloc)[addr]
				if account.Balance == nil {
					account.Balance = new(big.Int)
				}
				account.Extra = spec.setExtra(account.Extra)
				(*alloc)[addr] = account
			}
		}

		// Initialize contract storage only when Extra bits are also synced.
		// The hard fork path is not supported yet; if added later, storage and
		// Extra sync must be implemented together.
		for i, spec := range specs {
			if len(addrSets[i]) > 0 {
				sp = append(sp, initializeAddressSet(govCouncilAddress, spec.valuesSlot, spec.posSlot, addrSets[i])...)
			}
		}
	}

	// Initialize __accountManager
	sp = append(sp, params.StateParam{
		Address: govCouncilAddress,
		Key:     common.HexToHash(SLOT_GOV_COUNCIL_accountManager),
		Value:   common.BytesToHash(params.AccountManagerAddress.Bytes()),
	})

	return sp, nil
}

// initializeAddressSet initializes an AddressSet storage structure
// Follows AddressSetLib.sol implementation:
// - 1-based position indexing (0 means "not in set")
// - Array stored at keccak256(valuesSlot)
// - Positions mapping stored at keccak256(key || positionsSlot)
func initializeAddressSet(
	contractAddress common.Address,
	valuesSlot common.Hash,
	positionsSlot common.Hash,
	addresses map[common.Address]struct{},
) []params.StateParam {
	if len(addresses) == 0 {
		return nil
	}

	// Sort addresses to ensure deterministic slot index assignment.
	// Consistent ordering is required for a reproducible genesis block hash.
	sorted := make([]common.Address, 0, len(addresses))
	for addr := range addresses {
		sorted = append(sorted, addr)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i].Bytes(), sorted[j].Bytes()) < 0
	})

	sp := make([]params.StateParam, 0)

	// Set array length in base slot
	sp = append(sp, params.StateParam{
		Address: contractAddress,
		Key:     valuesSlot,
		Value:   common.BigToHash(big.NewInt(int64(len(sorted)))),
	})

	for i, addr := range sorted {
		// Array element slot: keccak256(valuesSlot) + index
		arraySlot := CalculateDynamicSlot(valuesSlot, big.NewInt(int64(i)))
		sp = append(sp, params.StateParam{
			Address: contractAddress,
			Key:     arraySlot,
			Value:   common.BytesToHash(addr.Bytes()),
		})

		// Position mapping: positions[addr] = index + 1 (1-based)
		positionKey := CalculateMappingSlot(positionsSlot, addr)
		sp = append(sp, params.StateParam{
			Address: contractAddress,
			Key:     positionKey,
			Value:   common.BigToHash(big.NewInt(int64(i + 1))),
		})
	}

	return sp
}

// parseParamAddresses parses a comma-separated address string into a map set.
// Returns an error if a zero address is encountered.
func parseParamAddresses(paramStr string) (map[common.Address]struct{}, error) {
	set := make(map[common.Address]struct{})
	for _, s := range splitAndTrim(paramStr, ",") {
		addr := common.HexToAddress(s)
		if addr == (common.Address{}) {
			return nil, fmt.Errorf("contains invalid zero address")
		}
		set[addr] = struct{}{}
	}
	return set, nil
}

// IsBlacklisted checks if an address is in the blacklist
func IsBlacklisted(govCouncilAddress common.Address, state StateReader, addr common.Address) bool {
	positionsSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_positions)
	positionKey := CalculateMappingSlot(positionsSlot, addr)
	position := state.GetState(govCouncilAddress, positionKey).Big()

	// Position 0 means not in set (1-based indexing)
	return position.Sign() > 0
}

// GetBlacklistCount returns the number of addresses in the blacklist
func GetBlacklistCount(govCouncilAddress common.Address, state StateReader) *big.Int {
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values)
	length := state.GetState(govCouncilAddress, valuesSlot)
	return length.Big()
}

// GetBlacklistedAddress returns the address at the given index in the blacklist
func GetBlacklistedAddress(govCouncilAddress common.Address, state StateReader, index *big.Int) common.Address {
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values)
	arraySlot := CalculateDynamicSlot(valuesSlot, index)
	value := state.GetState(govCouncilAddress, arraySlot)
	return common.BytesToAddress(value.Bytes())
}

// GetAllBlacklisted returns all blacklisted addresses
func GetAllBlacklisted(govCouncilAddress common.Address, state StateReader) []common.Address {
	count := GetBlacklistCount(govCouncilAddress, state)
	if count.Sign() == 0 {
		return nil
	}

	addresses := make([]common.Address, 0, count.Uint64())
	for i := int64(0); i < count.Int64(); i++ {
		addr := GetBlacklistedAddress(govCouncilAddress, state, big.NewInt(i))
		addresses = append(addresses, addr)
	}
	return addresses
}

// IsAuthorizedAccount checks if an address is in the authorized accounts list
func IsAuthorizedAccount(govCouncilAddress common.Address, state StateReader, addr common.Address) bool {
	positionsSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_positions)
	positionKey := CalculateMappingSlot(positionsSlot, addr)
	position := state.GetState(govCouncilAddress, positionKey).Big()

	// Position 0 means not in set (1-based indexing)
	return position.Sign() > 0
}

// GetAuthorizedAccountCount returns the number of authorized accounts
func GetAuthorizedAccountCount(govCouncilAddress common.Address, state StateReader) *big.Int {
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_values)
	length := state.GetState(govCouncilAddress, valuesSlot)
	return length.Big()
}

// GetAuthorizedAccountAddress returns the address at the given index in the authorized accounts list
func GetAuthorizedAccountAddress(govCouncilAddress common.Address, state StateReader, index *big.Int) common.Address {
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_values)
	arraySlot := CalculateDynamicSlot(valuesSlot, index)
	value := state.GetState(govCouncilAddress, arraySlot)
	return common.BytesToAddress(value.Bytes())
}

// GetAllAuthorizedAccounts returns all authorized account addresses
func GetAllAuthorizedAccounts(govCouncilAddress common.Address, state StateReader) []common.Address {
	count := GetAuthorizedAccountCount(govCouncilAddress, state)
	if count.Sign() == 0 {
		return nil
	}

	addresses := make([]common.Address, 0, count.Uint64())
	for i := int64(0); i < count.Int64(); i++ {
		addr := GetAuthorizedAccountAddress(govCouncilAddress, state, big.NewInt(i))
		addresses = append(addresses, addr)
	}
	return addresses
}

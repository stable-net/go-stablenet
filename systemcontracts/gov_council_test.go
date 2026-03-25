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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// MockStateReader implements StateReader interface for testing
type MockStateReader struct {
	state map[common.Address]map[common.Hash]common.Hash
}

func NewMockStateReader() *MockStateReader {
	return &MockStateReader{
		state: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func (m *MockStateReader) GetState(addr common.Address, key common.Hash) common.Hash {
	if addrState, ok := m.state[addr]; ok {
		if value, ok := addrState[key]; ok {
			return value
		}
	}
	return common.Hash{}
}

func (m *MockStateReader) SetState(addr common.Address, key common.Hash, value common.Hash) {
	if _, ok := m.state[addr]; !ok {
		m.state[addr] = make(map[common.Hash]common.Hash)
	}
	m.state[addr][key] = value
}

func TestInitializeCouncil_EmptyLists(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222,0x3333333333333333333333333333333333333333",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Should only have GovBase initialization params (no blacklist or authorized accounts)
	// Check that blacklist and authorized accounts slots are not set
	hasBlacklistSlot := false
	hasAuthorizedAccountSlot := false

	for _, param := range stateParams {
		if param.Key == common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values) {
			hasBlacklistSlot = true
		}
		if param.Key == common.HexToHash(SLOT_GOV_COUNCIL_currentAuthorizedAccounts_values) {
			hasAuthorizedAccountSlot = true
		}
	}

	require.False(t, hasBlacklistSlot, "Should not initialize empty blacklist")
	require.False(t, hasAuthorizedAccountSlot, "Should not initialize empty authorized accounts")
}

func TestInitializeCouncil_WithBlacklist(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_COUNCIL_PARAM_BLACKLIST:   "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify blacklist initialization
	blacklistCount := GetBlacklistCount(govCouncilAddress, mockState)
	require.Equal(t, int64(2), blacklistCount.Int64())

	// Verify addresses
	addr1 := GetBlacklistedAddress(govCouncilAddress, mockState, big.NewInt(0))
	addr2 := GetBlacklistedAddress(govCouncilAddress, mockState, big.NewInt(1))
	require.Equal(t, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), addr1)
	require.Equal(t, common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"), addr2)

	// Verify IsBlacklisted
	require.True(t, IsBlacklisted(govCouncilAddress, mockState, addr1))
	require.True(t, IsBlacklisted(govCouncilAddress, mockState, addr2))
	require.False(t, IsBlacklisted(govCouncilAddress, mockState, common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")))
}

func TestInitializeCouncil_WithAuthorizedAccounts(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:                "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:                 "2",
		GOV_BASE_PARAM_EXPIRY:                 "604800",
		GOV_BASE_PARAM_MEMBER_VERSION:         "1",
		GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS: "0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD,0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify authorized accounts initialization
	authorizedAccountCount := GetAuthorizedAccountCount(govCouncilAddress, mockState)
	require.Equal(t, int64(2), authorizedAccountCount.Int64())

	// Verify addresses
	addr1 := GetAuthorizedAccountAddress(govCouncilAddress, mockState, big.NewInt(0))
	addr2 := GetAuthorizedAccountAddress(govCouncilAddress, mockState, big.NewInt(1))
	require.Equal(t, common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"), addr1)
	require.Equal(t, common.HexToAddress("0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"), addr2)

	// Verify IsAuthorizedAccount
	require.True(t, IsAuthorizedAccount(govCouncilAddress, mockState, addr1))
	require.True(t, IsAuthorizedAccount(govCouncilAddress, mockState, addr2))
	require.False(t, IsAuthorizedAccount(govCouncilAddress, mockState, common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")))
}

func TestInitializeCouncil_WithBothLists(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:                "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:                 "2",
		GOV_BASE_PARAM_EXPIRY:                 "604800",
		GOV_BASE_PARAM_MEMBER_VERSION:         "1",
		GOV_COUNCIL_PARAM_BLACKLIST:           "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS: "0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify both lists
	require.Equal(t, int64(1), GetBlacklistCount(govCouncilAddress, mockState).Int64())
	require.Equal(t, int64(1), GetAuthorizedAccountCount(govCouncilAddress, mockState).Int64())

	blacklistedAddr := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	authorizedAddr := common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")

	require.True(t, IsBlacklisted(govCouncilAddress, mockState, blacklistedAddr))
	require.True(t, IsAuthorizedAccount(govCouncilAddress, mockState, authorizedAddr))
}

func TestInitializeCouncil_DuplicateAddresses(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		// Duplicate addresses
		GOV_COUNCIL_PARAM_BLACKLIST: "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Should deduplicate and only have 2 addresses
	blacklistCount := GetBlacklistCount(govCouncilAddress, mockState)
	require.Equal(t, int64(2), blacklistCount.Int64())
}

func TestInitializeCouncil_ZeroAddressError(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_COUNCIL_PARAM_BLACKLIST:   "0x0000000000000000000000000000000000000000",
	}

	_, err := initializeGovCouncil(govCouncilAddress, params)
	require.Error(t, err)
	require.Contains(t, err.Error(), "zero address")
}

func TestGetAllBlacklisted(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_COUNCIL_PARAM_BLACKLIST:   "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB,0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Get all blacklisted addresses
	addresses := GetAllBlacklisted(govCouncilAddress, mockState)
	require.Len(t, addresses, 3)

	// Check addresses (order should match initialization order)
	require.Equal(t, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), addresses[0])
	require.Equal(t, common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"), addresses[1])
	require.Equal(t, common.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"), addresses[2])
}

func TestGetAllAuthorizedAccounts(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	params := map[string]string{
		GOV_BASE_PARAM_MEMBERS:                "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:                 "2",
		GOV_BASE_PARAM_EXPIRY:                 "604800",
		GOV_BASE_PARAM_MEMBER_VERSION:         "1",
		GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS: "0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD,0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, params)
	require.NoError(t, err)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Get all authorized accounts
	addresses := GetAllAuthorizedAccounts(govCouncilAddress, mockState)
	require.Len(t, addresses, 2)

	// Check addresses
	require.Equal(t, common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD"), addresses[0])
	require.Equal(t, common.HexToAddress("0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE"), addresses[1])
}

func TestGetAllBlacklisted_Empty(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	mockState := NewMockStateReader()

	addresses := GetAllBlacklisted(govCouncilAddress, mockState)
	require.Nil(t, addresses)
}

func TestGetAllAuthorizedAccounts_Empty(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	mockState := NewMockStateReader()

	addresses := GetAllAuthorizedAccounts(govCouncilAddress, mockState)
	require.Nil(t, addresses)
}

func TestInitializeAddressSet_StorageLayout(t *testing.T) {
	contractAddress := common.HexToAddress("0x1000")
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values)
	positionsSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_positions)
	addresses := []string{
		"0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	}

	stateParams, err := initializeAddressSet(contractAddress, valuesSlot, positionsSlot, addresses, "test")
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify length stored in valuesSlot
	length := mockState.GetState(contractAddress, valuesSlot)
	require.Equal(t, int64(2), length.Big().Int64())

	// Verify first address in array
	arraySlot0 := CalculateDynamicSlot(valuesSlot, big.NewInt(0))
	addr0 := mockState.GetState(contractAddress, arraySlot0)
	require.Equal(t, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), common.BytesToAddress(addr0.Bytes()))

	// Verify second address in array
	arraySlot1 := CalculateDynamicSlot(valuesSlot, big.NewInt(1))
	addr1 := mockState.GetState(contractAddress, arraySlot1)
	require.Equal(t, common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"), common.BytesToAddress(addr1.Bytes()))

	// Verify position mappings (1-based indexing)
	posKey0 := CalculateMappingSlot(positionsSlot, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))
	pos0 := mockState.GetState(contractAddress, posKey0)
	require.Equal(t, int64(1), pos0.Big().Int64()) // 1-based

	posKey1 := CalculateMappingSlot(positionsSlot, common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"))
	pos1 := mockState.GetState(contractAddress, posKey1)
	require.Equal(t, int64(2), pos1.Big().Int64()) // 1-based
}

func TestInitializeAddressSet_EmptyList(t *testing.T) {
	contractAddress := common.HexToAddress("0x1000")
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values)
	positionsSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_positions)
	addresses := []string{}

	stateParams, err := initializeAddressSet(contractAddress, valuesSlot, positionsSlot, addresses, "test")
	require.NoError(t, err)
	require.Empty(t, stateParams, "Empty list should not generate state params")
}

func TestInitializeAddressSet_WithWhitespace(t *testing.T) {
	contractAddress := common.HexToAddress("0x1000")
	valuesSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_values)
	positionsSlot := common.HexToHash(SLOT_GOV_COUNCIL_currentBlacklist_positions)
	addresses := []string{
		"  0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA  ",
		"\t0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB\n",
	}

	stateParams, err := initializeAddressSet(contractAddress, valuesSlot, positionsSlot, addresses, "test")
	require.NoError(t, err)

	// Apply to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify addresses are trimmed correctly
	arraySlot0 := CalculateDynamicSlot(valuesSlot, big.NewInt(0))
	addr0 := mockState.GetState(contractAddress, arraySlot0)
	require.Equal(t, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"), common.BytesToAddress(addr0.Bytes()))
}

func TestInitializeCouncil_AccountManagerInitialization(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	testParams := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, testParams)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify __accountManager is initialized to AccountManagerAddress
	accountManagerSlot := common.HexToHash(SLOT_GOV_COUNCIL_accountManager)
	accountManagerValue := mockState.GetState(govCouncilAddress, accountManagerSlot)
	expectedAccountManager := common.BytesToHash(params.AccountManagerAddress.Bytes())

	require.Equal(t, expectedAccountManager, accountManagerValue,
		"__accountManager should be initialized to params.AccountManagerAddress")

	// Verify the address can be converted back
	storedAddress := common.BytesToAddress(accountManagerValue.Bytes())
	require.Equal(t, params.AccountManagerAddress, storedAddress,
		"Stored AccountManager address should match params.AccountManagerAddress")
}

func TestInitializeCouncil_AllSlots(t *testing.T) {
	govCouncilAddress := common.HexToAddress("0x1000")
	testParams := map[string]string{
		GOV_BASE_PARAM_MEMBERS:                "0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_QUORUM:                 "2",
		GOV_BASE_PARAM_EXPIRY:                 "604800",
		GOV_BASE_PARAM_MEMBER_VERSION:         "1",
		GOV_COUNCIL_PARAM_BLACKLIST:           "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS: "0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD",
	}

	stateParams, err := initializeGovCouncil(govCouncilAddress, testParams)
	require.NoError(t, err)
	require.NotNil(t, stateParams)

	// Apply state params to mock state
	mockState := NewMockStateReader()
	for _, param := range stateParams {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	// Verify all expected slots are initialized:
	// 1. GovBase slots (inherited)
	// 2. Blacklist slots (0x32, 0x33)
	// 3. Authorized accounts slots (0x34, 0x35)
	// 4. __accountManager slot (0x36)

	// Check blacklist
	blacklistCount := GetBlacklistCount(govCouncilAddress, mockState)
	require.Equal(t, int64(1), blacklistCount.Int64())

	// Check authorized accounts
	authorizedCount := GetAuthorizedAccountCount(govCouncilAddress, mockState)
	require.Equal(t, int64(1), authorizedCount.Int64())

	// Check __accountManager
	accountManagerSlot := common.HexToHash(SLOT_GOV_COUNCIL_accountManager)
	accountManagerValue := mockState.GetState(govCouncilAddress, accountManagerSlot)
	storedAddress := common.BytesToAddress(accountManagerValue.Bytes())
	require.Equal(t, params.AccountManagerAddress, storedAddress)
}

// =============================================================================
// upgradeGovCouncil tests
// =============================================================================

func TestUpgradeGovCouncil_EmptyParams(t *testing.T) {
	sp, err := upgradeGovCouncil(common.HexToAddress("0x1004"), map[string]string{})
	require.NoError(t, err)
	require.Equal(t, 0, len(sp), "Empty params should produce no StateParams")
}

func TestUpgradeGovCouncil_QuorumOnly(t *testing.T) {
	addr := common.HexToAddress("0x1004")
	sp, err := upgradeGovCouncil(addr, map[string]string{"quorum": "3"})
	require.NoError(t, err)

	// quorum: included
	quorumParam := findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_BASE_quorum))
	require.NotNil(t, quorumParam)
	require.Equal(t, common.BigToHash(big.NewInt(3)), quorumParam.Value)

	// maxProposals, expiry: not included (on-chain preserved)
	require.Nil(t, findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember)),
		"maxProposals should NOT be present")
	require.Nil(t, findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_BASE_proposalExpiry)),
		"expiry should NOT be present")

	// accountManager: not included in upgrade (already set at v1 init)
	require.Nil(t, findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_COUNCIL_accountManager)),
		"accountManager should NOT be present in upgrade")
}

func TestUpgradeGovCouncil_WithBlacklist(t *testing.T) {
	addr := common.HexToAddress("0x1004")
	sp, err := upgradeGovCouncil(addr, map[string]string{
		"blacklist": "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
	})
	require.NoError(t, err)

	// Apply to mock state and verify
	mockState := NewMockStateReader()
	for _, param := range sp {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	blacklistCount := GetBlacklistCount(addr, mockState)
	require.Equal(t, int64(2), blacklistCount.Int64())
	require.True(t, IsBlacklisted(addr, mockState, common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")))
	require.True(t, IsBlacklisted(addr, mockState, common.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")))
}

func TestUpgradeGovCouncil_WithAuthorizedAccounts(t *testing.T) {
	addr := common.HexToAddress("0x1004")
	sp, err := upgradeGovCouncil(addr, map[string]string{
		"authorizedAccounts": "0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD",
	})
	require.NoError(t, err)

	mockState := NewMockStateReader()
	for _, param := range sp {
		mockState.SetState(param.Address, param.Key, param.Value)
	}

	authorizedCount := GetAuthorizedAccountCount(addr, mockState)
	require.Equal(t, int64(1), authorizedCount.Int64())
	require.True(t, IsAuthorizedAccount(addr, mockState, common.HexToAddress("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")))
}

func TestUpgradeGovCouncil_QuorumAndBlacklist(t *testing.T) {
	addr := common.HexToAddress("0x1004")
	sp, err := upgradeGovCouncil(addr, map[string]string{
		"quorum":    "4",
		"blacklist": "0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	})
	require.NoError(t, err)

	// quorum included
	quorumParam := findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_BASE_quorum))
	require.NotNil(t, quorumParam)
	require.Equal(t, common.BigToHash(big.NewInt(4)), quorumParam.Value)

	// blacklist included
	mockState := NewMockStateReader()
	for _, param := range sp {
		mockState.SetState(param.Address, param.Key, param.Value)
	}
	require.Equal(t, int64(1), GetBlacklistCount(addr, mockState).Int64())

	// maxProposals, expiry, accountManager: not included
	require.Nil(t, findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember)))
	require.Nil(t, findUpgradeStateParam(sp, common.HexToHash(SLOT_GOV_COUNCIL_accountManager)))
}

func TestUpgradeGovCouncil_InvalidQuorum(t *testing.T) {
	_, err := upgradeGovCouncil(common.HexToAddress("0x1004"), map[string]string{"quorum": "0"})
	require.Error(t, err, "quorum=0 should be rejected")
}

func TestUpgradeGovCouncil_InvalidBlacklistZeroAddress(t *testing.T) {
	_, err := upgradeGovCouncil(common.HexToAddress("0x1004"), map[string]string{
		"blacklist": "0x0000000000000000000000000000000000000000",
	})
	require.Error(t, err, "zero address in blacklist should be rejected")
}

// findUpgradeStateParam is a helper to find a StateParam by Key (avoids conflict with systemcontracts_test.go)
func findUpgradeStateParam(sp []params.StateParam, key common.Hash) *params.StateParam {
	for i := range sp {
		if sp[i].Key == key {
			return &sp[i]
		}
	}
	return nil
}

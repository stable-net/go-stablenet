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

package test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

// newAllocSyncEnv creates a minimal GovCouncil environment for alloc sync tests.
// alloc entries can include Extra bits to simulate pre-existing account state.
func newAllocSyncEnv(t *testing.T, alloc types.GenesisAlloc, blacklistParam, authorizedParam string) (*GovWBFT, *EOA) {
	t.Helper()

	members := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
	nonMember := NewEOA()

	for _, m := range members {
		alloc[m.Operator.Address] = types.Account{Balance: towei(1_000_000)}
	}
	alloc[nonMember.Address] = types.Account{Balance: towei(1_000_000)}

	var memberAddrs, validators, blsPubKeys string
	for i, m := range members {
		if i > 0 {
			memberAddrs += ","
			validators += ","
			blsPubKeys += ","
		}
		memberAddrs += m.Operator.Address.String()
		validators += m.Validator.Address.String()
		blsPubKeys += hexutil.Encode(m.GetBLSPublicKey(t).Marshal())
	}

	govCouncil, err := NewGovWBFT(t, alloc,
		func(v *params.SystemContract) {
			v.Params = map[string]string{
				"members":       memberAddrs,
				"quorum":        "2",
				"expiry":        "604800",
				"memberVersion": "1",
				"validators":    validators,
				"blsPublicKeys": blsPubKeys,
			}
		},
		nil, nil, nil,
		func(c *params.SystemContract) {
			c.Params = map[string]string{
				sc.GOV_BASE_PARAM_MEMBERS:                memberAddrs,
				sc.GOV_BASE_PARAM_QUORUM:                 "2",
				sc.GOV_BASE_PARAM_EXPIRY:                 "604800",
				sc.GOV_BASE_PARAM_MEMBER_VERSION:         "1",
				sc.GOV_COUNCIL_PARAM_BLACKLIST:           blacklistParam,
				sc.GOV_COUNCIL_PARAM_AUTHORIZED_ACCOUNTS: authorizedParam,
			}
		},
		nil,
	)
	require.NoError(t, err)

	return govCouncil, nonMember
}

// TestAllocSync_ParamsOnly_Integration verifies that addresses in params but absent
// from alloc.Extra are registered in contract slots after genesis.
func TestAllocSync_ParamsOnly_Integration(t *testing.T) {
	addrA := NewEOA().Address
	addrB := NewEOA().Address

	govCouncil, nonMember := newAllocSyncEnv(t, types.GenesisAlloc{}, addrA.Hex(), addrB.Hex())
	defer govCouncil.backend.Close()

	isBlacklisted, err := govCouncil.IsBlacklisted(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isBlacklisted)

	isAuthorized, err := govCouncil.IsAuthorizedAccount(nonMember, addrB)
	require.NoError(t, err)
	require.True(t, isAuthorized)
}

// TestAllocSync_AllocOnly_Integration verifies that addresses with Extra bits set
// in alloc but absent from params are registered in contract slots after genesis.
func TestAllocSync_AllocOnly_Integration(t *testing.T) {
	addrA := NewEOA().Address
	addrB := NewEOA().Address

	alloc := types.GenesisAlloc{
		addrA: {Balance: big.NewInt(0), Extra: types.SetBlacklisted(0)},
		addrB: {Balance: big.NewInt(0), Extra: types.SetAuthorized(0)},
	}

	govCouncil, nonMember := newAllocSyncEnv(t, alloc, "", "")
	defer govCouncil.backend.Close()

	isBlacklisted, err := govCouncil.IsBlacklisted(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isBlacklisted)

	isAuthorized, err := govCouncil.IsAuthorizedAccount(nonMember, addrB)
	require.NoError(t, err)
	require.True(t, isAuthorized)
}

// TestAllocSync_Union_Integration verifies that addresses from both params and alloc.Extra
// are merged into the contract slots without duplication.
func TestAllocSync_Union_Integration(t *testing.T) {
	addrA := NewEOA().Address
	addrB := NewEOA().Address

	alloc := types.GenesisAlloc{
		addrB: {Balance: big.NewInt(0), Extra: types.SetBlacklisted(0)},
	}

	govCouncil, nonMember := newAllocSyncEnv(t, alloc, addrA.Hex(), "")
	defer govCouncil.backend.Close()

	count, err := govCouncil.GetBlacklistCount(nonMember)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(2), count)

	isBlacklistedA, err := govCouncil.IsBlacklisted(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isBlacklistedA)

	isBlacklistedB, err := govCouncil.IsBlacklisted(nonMember, addrB)
	require.NoError(t, err)
	require.True(t, isBlacklistedB)
}

// TestAllocSync_NoDuplication_Integration verifies that an address present in both
// params and alloc.Extra appears only once in the contract slots.
func TestAllocSync_NoDuplication_Integration(t *testing.T) {
	addrA := NewEOA().Address

	alloc := types.GenesisAlloc{
		addrA: {Balance: big.NewInt(0), Extra: types.SetBlacklisted(0)},
	}

	govCouncil, nonMember := newAllocSyncEnv(t, alloc, addrA.Hex(), "")
	defer govCouncil.backend.Close()

	count, err := govCouncil.GetBlacklistCount(nonMember)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(1), count)
}

// TestAllocSync_BothBitsSet_Integration verifies that an address with blacklist Extra bit
// but authorized in params is registered in both contract slots.
func TestAllocSync_BothBitsSet_Integration(t *testing.T) {
	addrA := NewEOA().Address

	alloc := types.GenesisAlloc{
		addrA: {Balance: big.NewInt(0), Extra: types.SetBlacklisted(0)},
	}

	govCouncil, nonMember := newAllocSyncEnv(t, alloc, "", addrA.Hex())
	defer govCouncil.backend.Close()

	isBlacklisted, err := govCouncil.IsBlacklisted(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isBlacklisted)

	isAuthorized, err := govCouncil.IsAuthorizedAccount(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isAuthorized)
}

// TestAllocSync_UnrelatedAllocPreserved_Integration verifies that alloc entries
// unrelated to blacklist/authorized are not affected by genesis sync.
func TestAllocSync_UnrelatedAllocPreserved_Integration(t *testing.T) {
	addrA := NewEOA().Address
	addrUnrelated := NewEOA().Address

	alloc := types.GenesisAlloc{
		addrUnrelated: {Balance: towei(999)},
	}

	govCouncil, nonMember := newAllocSyncEnv(t, alloc, addrA.Hex(), "")
	defer govCouncil.backend.Close()

	// addrUnrelated is not in the blacklist.
	isBlacklisted, err := govCouncil.IsBlacklisted(nonMember, addrUnrelated)
	require.NoError(t, err)
	require.False(t, isBlacklisted)

	// addrA is blacklisted as expected.
	isBlacklisted, err = govCouncil.IsBlacklisted(nonMember, addrA)
	require.NoError(t, err)
	require.True(t, isBlacklisted)

	// Verify balance preserved via contract call (balance check through chain).
	balance, err := govCouncil.backend.Client().BalanceAt(context.Background(), addrUnrelated, nil)
	require.NoError(t, err)
	require.Equal(t, towei(999), balance)
}

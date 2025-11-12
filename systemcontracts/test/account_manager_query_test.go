// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.

package test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAccountManagerQuery tests AccountManager query functions (isBlacklist, isAuthorized)
// These functions should return bool values via native precompiled contract
func TestAccountManagerQuery_IsBlacklist(t *testing.T) {
	t.Run("isBlacklist returns true for blacklisted account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify not blacklisted initially
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted initially")

		// Add to blacklist via proposal
		txPropose, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		// Approve proposal
		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Now check via AccountManager query
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "isBlacklist should return true for blacklisted account")
	})

	t.Run("isBlacklist returns false after removal from blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Add to blacklist
		txPropose, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted)

		// Remove from blacklist
		txProposeRemove, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txProposeRemove, err)
		require.NoError(t, err)

		txApproveRemove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txApproveRemove, err)
		require.NoError(t, err)

		// Check via AccountManager query - should return false
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "isBlacklist should return false after removal")
	})

	t.Run("isBlacklist works for proposal-added addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Create new addresses to add via proposal
		addr1 := NewEOA().Address
		addr2 := NewEOA().Address

		// Add first address to blacklist via proposal
		txPropose1, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, addr1)
		_, err = gCouncil.ExpectedOk(txPropose1, err)
		require.NoError(t, err)

		txApprove1, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove1, err)
		require.NoError(t, err)

		// Add second address to blacklist via proposal
		txPropose2, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, addr2)
		_, err = gCouncil.ExpectedOk(txPropose2, err)
		require.NoError(t, err)

		txApprove2, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txApprove2, err)
		require.NoError(t, err)

		// Check first address via AccountManager query
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, addr1)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "isBlacklist should return true for proposal-added address")

		// Check second address via AccountManager query
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, addr2)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "isBlacklist should return true for proposal-added address")
	})
}

func TestAccountManagerQuery_IsAuthorized(t *testing.T) {
	t.Run("isAuthorized returns true for authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify not authorized initially
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "Address should not be authorized initially")

		// Add to authorized accounts via proposal
		txPropose, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		// Approve proposal
		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Now check via AccountManager query
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "isAuthorized should return true for authorized account")
	})

	t.Run("isAuthorized returns false after removal", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Add to authorized accounts
		txPropose, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify authorized
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized)

		// Remove from authorized accounts
		txProposeRemove, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txProposeRemove, err)
		require.NoError(t, err)

		txApproveRemove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txApproveRemove, err)
		require.NoError(t, err)

		// Check via AccountManager query - should return false
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "isAuthorized should return false after removal")
	})

	t.Run("isAuthorized works for proposal-added addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Create new addresses to add via proposal
		addr1 := NewEOA().Address
		addr2 := NewEOA().Address

		// Add first address to authorized accounts via proposal
		txPropose1, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, addr1)
		_, err = gCouncil.ExpectedOk(txPropose1, err)
		require.NoError(t, err)

		txApprove1, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove1, err)
		require.NoError(t, err)

		// Add second address to authorized accounts via proposal
		txPropose2, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, addr2)
		_, err = gCouncil.ExpectedOk(txPropose2, err)
		require.NoError(t, err)

		txApprove2, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txApprove2, err)
		require.NoError(t, err)

		// Check first address via AccountManager query
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, addr1)
		require.NoError(t, err)
		require.True(t, isAuthorized, "isAuthorized should return true for proposal-added address")

		// Check second address via AccountManager query
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, addr2)
		require.NoError(t, err)
		require.True(t, isAuthorized, "isAuthorized should return true for proposal-added address")
	})
}

// Note: Direct call tests would require additional helper functions
// For now, we rely on GovCouncil's IsBlacklisted/IsAuthorizedAccount methods
// which internally call the AccountManager precompiled contract

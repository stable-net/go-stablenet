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

	t.Run("isBlacklist works for initially blacklisted addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// initialBlacklist is set in genesis (see helpers_edge_cases.go)
		require.NotEmpty(t, initialBlacklist, "Should have initial blacklist addresses")

		// Check first initially blacklisted address
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, initialBlacklist[0])
		require.NoError(t, err)
		require.True(t, isBlacklisted, "isBlacklist should return true for genesis blacklisted address")

		// Check second initially blacklisted address
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, initialBlacklist[1])
		require.NoError(t, err)
		require.True(t, isBlacklisted, "isBlacklist should return true for genesis blacklisted address")
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

	t.Run("isAuthorized works for initially authorized addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// initialAuthorizedAccounts is set in genesis
		require.NotEmpty(t, initialAuthorizedAccounts, "Should have initial authorized accounts")

		// Check first initially authorized address
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, initialAuthorizedAccounts[0])
		require.NoError(t, err)
		require.True(t, isAuthorized, "isAuthorized should return true for genesis authorized address")

		// Check second initially authorized address
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, initialAuthorizedAccounts[1])
		require.NoError(t, err)
		require.True(t, isAuthorized, "isAuthorized should return true for genesis authorized address")
	})
}

// Note: Direct call tests would require additional helper functions
// For now, we rely on GovCouncil's IsBlacklisted/IsAuthorizedAccount methods
// which internally call the AccountManager precompiled contract

// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.

package test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBlacklistTransferValidation tests that blacklisted accounts cannot send or receive transfers
// This implements Phase 3: Transfer Validation Hook - TDD approach
func TestBlacklistTransferValidation(t *testing.T) {
	t.Run("Transfer from blacklisted account should fail", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Use a fresh EOA with balance for testing
		sender := NewEOA()
		recipient := NewEOA()

		// Fund the sender account with native coins
		tx, err := TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, councilNonMember), towei(10_000), &sender.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err, "Funding sender should succeed")

		// Verify transfer works before blacklist
		tx, err = TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, sender), towei(100), &recipient.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err, "Transfer before blacklist should succeed")

		// Add sender to blacklist via proposal
		txPropose, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, sender.Address)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err, "Creating blacklist proposal should succeed")

		// Approve proposal (quorum is 2, proposer auto-approves)
		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err, "Approving proposal should succeed")

		// Verify sender is blacklisted in contract storage
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, sender.Address)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Sender should be blacklisted in contract storage")

		// Attempt transfer from blacklisted sender - should fail
		tx, err = TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, sender), towei(100), &recipient.Address)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Transfer from blacklisted account should fail")

		// Verify error mentions blacklist
		errorMsg := strings.ToLower(err.Error())
		require.True(t,
			strings.Contains(errorMsg, "blacklist") || strings.Contains(errorMsg, "execution reverted"),
			"Error should mention blacklist or execution reverted, got: %s", err.Error())
	})

	t.Run("Transfer to blacklisted account should fail", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Setup accounts
		sender := NewEOA()
		recipient := NewEOA()

		// Fund both accounts
		tx, err := TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, councilNonMember), towei(10_000), &sender.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, councilNonMember), towei(10_000), &recipient.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Add recipient to blacklist
		txPropose, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, recipient.Address)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify recipient is blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, recipient.Address)
		require.NoError(t, err)
		require.True(t, isBlacklisted)

		// Attempt transfer to blacklisted recipient - should fail
		tx, err = TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, sender), towei(100), &recipient.Address)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Transfer to blacklisted account should fail")

		errorMsg := strings.ToLower(err.Error())
		require.True(t,
			strings.Contains(errorMsg, "blacklist") || strings.Contains(errorMsg, "execution reverted"),
			"Error should mention blacklist or execution reverted, got: %s", err.Error())
	})

	t.Run("Transfer after removal from blacklist should succeed", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		sender := NewEOA()
		recipient := NewEOA()

		// Fund sender
		tx, err := TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, councilNonMember), towei(10_000), &sender.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Add to blacklist
		txPropose, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, sender.Address)
		_, err = gCouncil.ExpectedOk(txPropose, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, sender.Address)
		require.NoError(t, err)
		require.True(t, isBlacklisted)

		// Remove from blacklist
		txProposeRemove, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, sender.Address)
		_, err = gCouncil.ExpectedOk(txProposeRemove, err)
		require.NoError(t, err)

		txApproveRemove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txApproveRemove, err)
		require.NoError(t, err)

		// Verify not blacklisted
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, sender.Address)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Sender should not be blacklisted after removal")

		// Transfer should now succeed
		tx, err = TransferCoin(gCouncil.backend.Client(), NewTxOpts(t, sender), towei(100), &recipient.Address)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err, "Transfer after blacklist removal should succeed")
	})
}

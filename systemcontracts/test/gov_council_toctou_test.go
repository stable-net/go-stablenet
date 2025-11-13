// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"math/big"
	"testing"

	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

// TestGovCouncil_TOCTOU_DuplicateAddBlacklist tests TOCTOU fix for duplicate add blacklist proposals
func TestGovCouncil_TOCTOU_DuplicateAddBlacklist(t *testing.T) {
	t.Run("should handle duplicate add blacklist proposals gracefully", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify address is not blacklisted initially
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted initially")

		// Step 1: Member 0 proposes to add address to blacklist (Proposal #1)
		tx1, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx1, err)
		require.NoError(t, err)

		// Step 2: Member 1 proposes to add SAME address to blacklist (Proposal #2)
		tx2, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx2, err)
		require.NoError(t, err, "Should allow creating duplicate proposal at proposal time")

		// Verify both proposals were created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), proposalId, "Should have 2 proposals")

		// Step 3: Approve Proposal #1 (reaches quorum and auto-executes)
		tx, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt1, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #1 is executed
		proposal1, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusExecuted, proposal1.Status, "Proposal #1 should be executed")

		// Verify address is now blacklisted
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should be blacklisted after Proposal #1")

		// Verify AddressBlacklisted event was emitted for Proposal #1
		event1 := findEvent("AddressBlacklisted", receipt1.Logs)
		require.NotNil(t, event1, "AddressBlacklisted event should be emitted for Proposal #1")
		require.Equal(t, targetAddress, event1["account"], "Event should contain correct account")

		// Step 4: Approve Proposal #2 (reaches quorum and tries to execute)
		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(2))
		receipt2, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #2 is in Approved state (auto-execution returned false, stays Approved for retry)
		proposal2, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusApproved, proposal2.Status, "Proposal #2 should remain Approved after skipped execution")

		// Verify ProposalExecutionSkipped event was emitted during auto-execution
		skippedEvent := findEvent("ProposalExecutionSkipped", receipt2.Logs)
		require.NotNil(t, skippedEvent, "ProposalExecutionSkipped event should be emitted for Proposal #2")
		require.Equal(t, targetAddress, skippedEvent["account"], "Event should contain correct account")
		require.Equal(t, big.NewInt(2), skippedEvent["proposalId"], "Event should contain correct proposalId")
		require.Equal(t, "ALREADY_BLACKLISTED", skippedEvent["reason"], "Event should contain correct reason")

		// Step 5: Manually execute with failure flag to mark as Failed
		txFail, err := gCouncil.BaseTxExecuteWithFailure(t, gCouncil.govCouncil, councilMembers[0].Operator, big.NewInt(2))
		receiptFail, err := gCouncil.ExpectedOk(txFail, err)
		require.NoError(t, err)

		// Verify Proposal #2 is now Failed
		proposal2Final, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusFailed, proposal2Final.Status, "Proposal #2 should be Failed after executeWithFailure")

		// Verify ProposalExecutionSkipped event was emitted again
		skippedEvent2 := findEvent("ProposalExecutionSkipped", receiptFail.Logs)
		require.NotNil(t, skippedEvent2, "ProposalExecutionSkipped event should be emitted again")

		// Verify no duplicate AddressBlacklisted event for Proposal #2
		allBlacklistedEvents := findEvents("AddressBlacklisted", receipt2.Logs)
		require.Len(t, allBlacklistedEvents, 0, "Should not emit AddressBlacklisted event for duplicate")

		// Verify blacklist count is still correct (not duplicated)
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), count, "Blacklist count should be 3 (2 initial + 1 new), not 4")

		// Verify address is still blacklisted
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should remain blacklisted")
	})
}

// TestGovCouncil_TOCTOU_DuplicateRemoveBlacklist tests TOCTOU fix for duplicate remove blacklist proposals
func TestGovCouncil_TOCTOU_DuplicateRemoveBlacklist(t *testing.T) {
	t.Run("should handle duplicate remove blacklist proposals gracefully", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// First, add an address to blacklist and execute
		targetAddress := NewEOA().Address
		txAdd, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd, err)
		require.NoError(t, err)

		tx, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify address is blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should be blacklisted initially")

		// Step 1: Member 0 proposes to remove address from blacklist (Proposal #2, since #1 was Add)
		tx1, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx1, err)
		require.NoError(t, err)

		// Step 2: Member 1 proposes to remove SAME address from blacklist (Proposal #3)
		tx2, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx2, err)
		require.NoError(t, err, "Should allow creating duplicate proposal at proposal time")

		// Verify both remove proposals were created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), proposalId, "Should have 3 proposals total (1 add + 2 remove)")

		// Step 3: Approve Proposal #2 (first remove, reaches quorum and auto-executes)
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		receipt1, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #2 is executed
		proposal1, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusExecuted, proposal1.Status, "Proposal #2 should be executed")

		// Verify address is no longer blacklisted
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted after Proposal #2")

		// Verify AddressUnblacklisted event was emitted for Proposal #2
		event1 := findEvent("AddressUnblacklisted", receipt1.Logs)
		require.NotNil(t, event1, "AddressUnblacklisted event should be emitted for Proposal #2")

		// Step 4: Approve Proposal #3 (second remove, reaches quorum and tries to execute)
		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(3))
		receipt2, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #3 is in Approved state (auto-execution returned false, stays Approved for retry)
		proposal2, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusApproved, proposal2.Status, "Proposal #3 should remain Approved after skipped execution")

		// Verify ProposalExecutionSkipped event was emitted during auto-execution
		skippedEvent := findEvent("ProposalExecutionSkipped", receipt2.Logs)
		require.NotNil(t, skippedEvent, "ProposalExecutionSkipped event should be emitted for Proposal #3")
		require.Equal(t, targetAddress, skippedEvent["account"], "Event should contain correct account")
		require.Equal(t, big.NewInt(3), skippedEvent["proposalId"], "Event should contain correct proposalId")
		require.Equal(t, "NOT_IN_BLACKLIST", skippedEvent["reason"], "Event should contain correct reason")

		// Step 5: Manually execute with failure flag to mark as Failed
		txFail, err := gCouncil.BaseTxExecuteWithFailure(t, gCouncil.govCouncil, councilMembers[0].Operator, big.NewInt(3))
		_, err = gCouncil.ExpectedOk(txFail, err)
		require.NoError(t, err)

		// Verify Proposal #3 is now Failed
		proposal2Final, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusFailed, proposal2Final.Status, "Proposal #3 should be Failed after executeWithFailure")

		// Verify no duplicate AddressUnblacklisted event for Proposal #2
		allUnblacklistedEvents := findEvents("AddressUnblacklisted", receipt2.Logs)
		require.Len(t, allUnblacklistedEvents, 0, "Should not emit AddressUnblacklisted event for duplicate")

		// Verify blacklist count is correct (2 initial + 1 added - 1 removed = 2)
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), count, "Blacklist count should be 2 (2 initial + 1 added - 1 removed)")

		// Verify address is still not blacklisted
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should remain not blacklisted")
	})
}

// TestGovCouncil_TOCTOU_DuplicateAddAuthorized tests TOCTOU fix for duplicate add authorized account proposals
func TestGovCouncil_TOCTOU_DuplicateAddAuthorized(t *testing.T) {
	t.Run("should handle duplicate add authorized account proposals gracefully", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify address is not authorized initially
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "Address should not be authorized initially")

		// Step 1: Create two duplicate proposals
		tx1, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx1, err)
		require.NoError(t, err)

		tx2, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx2, err)
		require.NoError(t, err, "Should allow creating duplicate proposal")

		// Step 2: Approve and execute Proposal #1
		tx, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt1, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify address is now authorized
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "Address should be authorized after Proposal #1")

		// Verify AuthorizedAccountAdded event
		event1 := findEvent("AuthorizedAccountAdded", receipt1.Logs)
		require.NotNil(t, event1, "AuthorizedAccountAdded event should be emitted")

		// Step 3: Approve and execute Proposal #2 (should skip, remain Approved)
		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(2))
		receipt2, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #2 remains Approved (not Executed)
		proposal2, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusApproved, proposal2.Status, "Proposal #2 should remain Approved after skipped execution")

		// Verify ProposalExecutionSkipped event
		skippedEvent := findEvent("ProposalExecutionSkipped", receipt2.Logs)
		require.NotNil(t, skippedEvent, "ProposalExecutionSkipped event should be emitted")
		require.Equal(t, targetAddress, skippedEvent["account"], "Event should contain correct account")
		require.Equal(t, "ALREADY_AUTHORIZED", skippedEvent["reason"], "Event should contain correct reason")

		// Verify no duplicate event
		allAuthorizedEvents := findEvents("AuthorizedAccountAdded", receipt2.Logs)
		require.Len(t, allAuthorizedEvents, 0, "Should not emit duplicate event")

		// Step 4: Manually execute with failure flag to mark as Failed
		txFail, err := gCouncil.BaseTxExecuteWithFailure(t, gCouncil.govCouncil, councilMembers[0].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(txFail, err)
		require.NoError(t, err)

		// Verify Proposal #2 is now Failed
		proposal2Final, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusFailed, proposal2Final.Status, "Proposal #2 should be Failed after executeWithFailure")

		// Verify count is correct
		count, err := gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), count, "Count should be 3 (2 initial + 1 new)")
	})
}

// TestGovCouncil_TOCTOU_DuplicateRemoveAuthorized tests TOCTOU fix for duplicate remove authorized account proposals
func TestGovCouncil_TOCTOU_DuplicateRemoveAuthorized(t *testing.T) {
	t.Run("should handle duplicate remove authorized account proposals gracefully", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// First, add an address to authorized accounts and execute
		targetAddress := NewEOA().Address
		txAdd, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd, err)
		require.NoError(t, err)

		tx, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify address is authorized
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "Address should be authorized initially")

		// Step 1: Create two duplicate remove proposals (Proposal #2 and #3, since #1 was Add)
		tx1, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx1, err)
		require.NoError(t, err)

		tx2, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx2, err)
		require.NoError(t, err, "Should allow creating duplicate proposal")

		// Step 2: Approve and execute Proposal #2 (first remove)
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify address is no longer authorized
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "Address should not be authorized after Proposal #2")

		// Step 3: Approve and execute Proposal #3 (second remove, should skip)
		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(3))
		receipt2, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify Proposal #3 is in Approved state (auto-execution returned false, stays Approved for retry)
		proposal3, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusApproved, proposal3.Status, "Proposal #3 should remain Approved after skipped execution")

		// Verify ProposalExecutionSkipped event was emitted during auto-execution
		skippedEvent := findEvent("ProposalExecutionSkipped", receipt2.Logs)
		require.NotNil(t, skippedEvent, "ProposalExecutionSkipped event should be emitted")
		require.Equal(t, targetAddress, skippedEvent["account"], "Event should contain correct account")
		require.Equal(t, "NOT_AUTHORIZED", skippedEvent["reason"], "Event should contain correct reason")

		// Step 4: Manually execute with failure flag to mark as Failed
		txFail, err := gCouncil.BaseTxExecuteWithFailure(t, gCouncil.govCouncil, councilMembers[0].Operator, big.NewInt(3))
		_, err = gCouncil.ExpectedOk(txFail, err)
		require.NoError(t, err)

		// Verify Proposal #3 is now Failed
		proposal3Final, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusFailed, proposal3Final.Status, "Proposal #3 should be Failed after executeWithFailure")

		// Verify count is correct (2 initial + 1 added - 1 removed = 2)
		count, err := gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), count, "Count should be 2 (2 initial + 1 added - 1 removed)")
	})
}

// TestGovCouncil_TOCTOU_AlternatingProposals tests alternating add/remove proposals
func TestGovCouncil_TOCTOU_AlternatingProposals(t *testing.T) {
	t.Run("should handle alternating add and remove blacklist proposals", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Step 1: Add to blacklist first (and execute)
		txAdd1, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd1, err)
		require.NoError(t, err)

		tx, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		isBlacklisted, _ := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.True(t, isBlacklisted, "Should be blacklisted after first add")

		// Step 2: Now create Remove proposal (valid because address is in blacklist)
		txRemove1, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txRemove1, err)
		require.NoError(t, err)

		// Step 3: Execute Remove
		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(2))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		isBlacklisted, _ = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.False(t, isBlacklisted, "Should not be blacklisted after remove")

		// Step 4: Add again (and execute)
		txAdd2, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd2, err)
		require.NoError(t, err)

		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(3))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		isBlacklisted, _ = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.True(t, isBlacklisted, "Should be blacklisted after second add")

		// Step 5: Remove again (and execute)
		txRemove2, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[1].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txRemove2, err)
		require.NoError(t, err)

		tx, err = gCouncil.TxApprove(t, councilMembers[0].Operator, big.NewInt(4))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		isBlacklisted, _ = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.False(t, isBlacklisted, "Should not be blacklisted after second remove")
	})
}

// TestGovCouncil_TOCTOU_RapidFireDuplicates tests multiple rapid-fire duplicate proposals
func TestGovCouncil_TOCTOU_RapidFireDuplicates(t *testing.T) {
	t.Run("should handle multiple rapid-fire duplicate proposals", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Create 5 duplicate proposals rapidly
		proposalIds := make([]*big.Int, 5)
		for i := 0; i < 5; i++ {
			member := councilMembers[i%len(councilMembers)]
			tx, err := gCouncil.TxProposeAddBlacklist(t, member.Operator, targetAddress)
			_, err = gCouncil.ExpectedOk(tx, err)
			require.NoError(t, err)
			proposalIds[i] = big.NewInt(int64(i + 1))
		}

		// Execute all proposals in order
		successCount := 0
		skipCount := 0

		for i, proposalId := range proposalIds {
			// Each proposal was created by councilMembers[i%len(councilMembers)]
			// So we need a different member to approve (to reach quorum of 2)
			proposerIndex := i % len(councilMembers)
			approverIndex := (proposerIndex + 1) % len(councilMembers)
			approver := councilMembers[approverIndex].Operator

			// Approve to reach quorum
			tx, err := gCouncil.TxApprove(t, approver, proposalId)
			receipt, err := gCouncil.ExpectedOk(tx, err)
			require.NoError(t, err)

			// Check which event was emitted
			blacklistedEvent := findEvent("AddressBlacklisted", receipt.Logs)
			skippedEvent := findEvent("ProposalExecutionSkipped", receipt.Logs)

			if blacklistedEvent != nil {
				successCount++
				require.Equal(t, 0, i, "Only first proposal should succeed")
				// Verify Proposal #1 is Executed
				proposal1, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, proposalId)
				require.NoError(t, err)
				require.Equal(t, sc.ProposalStatusExecuted, proposal1.Status, "Proposal #1 should be Executed")
			}
			if skippedEvent != nil {
				skipCount++
				require.Greater(t, i, 0, "Only subsequent proposals should be skipped")
				require.Equal(t, "ALREADY_BLACKLISTED", skippedEvent["reason"], "Reason should be ALREADY_BLACKLISTED")
				// Verify skipped proposals remain Approved
				proposalSkipped, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, proposalId)
				require.NoError(t, err)
				require.Equal(t, sc.ProposalStatusApproved, proposalSkipped.Status, "Skipped proposals should remain Approved")
			}
		}

		// Verify exactly 1 succeeded and 4 skipped
		require.Equal(t, 1, successCount, "Exactly 1 proposal should succeed")
		require.Equal(t, 4, skipCount, "Exactly 4 proposals should be skipped")

		// Demonstrate cleanup: manually execute one of the skipped proposals with failure flag
		txFail, err := gCouncil.BaseTxExecuteWithFailure(t, gCouncil.govCouncil, councilMembers[0].Operator, big.NewInt(5))
		_, err = gCouncil.ExpectedOk(txFail, err)
		require.NoError(t, err)

		// Verify Proposal #5 is now Failed
		proposal5Final, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(5))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusFailed, proposal5Final.Status, "Proposal #5 should be Failed after executeWithFailure")

		// Verify final state
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should be blacklisted")

		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), count, "Count should be 3 (2 initial + 1 new)")
	})
}

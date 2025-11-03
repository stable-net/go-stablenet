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
	"time"

	"github.com/stretchr/testify/require"
)

// Test Suite: GovMinter Reservation Cleanup
// Purpose: Verify _cleanupMintReservation is called for ALL terminal proposal states
// Design: TDD approach - these tests will initially fail, then pass after hook implementation

// Test 1: Cancelled proposal should cleanup reservation
func TestGovMinter_CleanupOnCancelled(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member := minterMembers[0].Operator
	recipient := member.Address
	amount := big.NewInt(1_000_000)

	// Get initial reserved amount (should be 0)
	initialReserved, err := gMinter.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, initialReserved.Cmp(big.NewInt(0)), "Initial reserved amount should be 0")

	// Create mint proposal
	tx, err := gMinter.TxProposeMint(t, member, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Get proposal ID
	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member)
	require.NoError(t, err)

	// Verify reservation increased
	reservedAfterPropose, err := gMinter.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, amount, reservedAfterPropose, "Reservation should increase after proposal")

	// Verify proposal-specific reservation
	proposalReserved, err := gMinter.GetMintProposalAmount(member, proposalId)
	require.NoError(t, err)
	require.Equal(t, amount, proposalReserved, "Proposal-specific reservation should match")

	// Cancel proposal
	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Cancelled
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(4), uint8(proposal.Status), "Proposal should be Cancelled")

	// CRITICAL: Verify reservation is cleaned up
	reservedAfterCancel, err := gMinter.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterCancel.Cmp(big.NewInt(0)), "Reservation should be cleaned up after cancellation")

	// Verify proposal-specific reservation is cleared
	proposalReservedAfter, err := gMinter.GetMintProposalAmount(member, proposalId)
	require.NoError(t, err)
	require.Equal(t, 0, proposalReservedAfter.Cmp(big.NewInt(0)), "Proposal-specific reservation should be cleared")
}

// Test 2: Rejected proposal should cleanup reservation
func TestGovMinter_CleanupOnRejected(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	member2 := minterMembers[2].Operator
	recipient := member0.Address
	amount := big.NewInt(2_000_000)

	// Create mint proposal (member0 auto-approves)
	tx, err := gMinter.TxProposeMint(t, member0, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Verify reservation
	reservedAfterPropose, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, amount, reservedAfterPropose)

	// Member1 and Member2 reject (quorum=2, so 2 rejections will reject the proposal)
	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member2, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Rejected
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(7), uint8(proposal.Status), "Proposal should be Rejected")

	// CRITICAL: Verify reservation is cleaned up
	reservedAfterReject, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterReject.Cmp(big.NewInt(0)), "Reservation should be cleaned up after rejection")
}

// Test 3: Expired proposal (during vote) should cleanup reservation
func TestGovMinter_CleanupOnExpiredDuringVote(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	recipient := member0.Address
	amount := big.NewInt(3_000_000)

	// Create mint proposal
	tx, err := gMinter.TxProposeMint(t, member0, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Verify reservation
	reservedAfterPropose, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, amount, reservedAfterPropose)

	// Advance time past proposal expiry (7 days + 1 second)
	gMinter.backend.AdjustTime(7*24*time.Hour + time.Second)
	gMinter.backend.Commit()

	// Manually expire the proposal (expireProposal doesn't revert)
	tx, err = gMinter.BaseTxExpireProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err, "expireProposal should succeed")

	// CRITICAL: Verify reservation is cleaned up after manual expiry
	reservedAfterExpiry, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterExpiry.Cmp(big.NewInt(0)), "Reservation should be cleaned up after expiry")
}

// Test 4: Failed proposal should cleanup reservation
func TestGovMinter_CleanupOnFailed(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	recipient := member0.Address
	amount := big.NewInt(5_000_000)

	// Create and approve mint proposal
	tx, err := gMinter.TxProposeMint(t, member0, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Configure MockFiatToken to fail on mint BEFORE approval
	// This ensures auto-execution fails when quorum is reached
	tx, err = gMinter.SetMockFiatTokenMintShouldFail(t, member0, true)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Approve with member1 (auto-execution will fail but proposal stays Approved for retry)
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is still Approved (auto-execution failed but retry is possible)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be Approved (failed auto-exec, retry possible)")

	// Now do terminal execution to mark as Failed
	tx, err = gMinter.BaseTxExecuteWithFailure(t, gMinter.govMinter, member0, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is now Failed
	proposal, err = gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(6), uint8(proposal.Status), "Proposal should be Failed after terminal execution")

	// CRITICAL: Verify reservation is cleaned up
	reservedAfterFail, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterFail.Cmp(big.NewInt(0)), "Reservation should be cleaned up after failure")

	// Restore mock behavior
	tx, err = gMinter.SetMockFiatTokenMintShouldFail(t, member0, false)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
}

// Test 6: Executed proposal should cleanup reservation (regression test)
func TestGovMinter_CleanupOnExecuted(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	recipient := member0.Address
	amount := big.NewInt(6_000_000)

	// Create and execute mint proposal
	tx, err := gMinter.TxProposeMint(t, member0, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Verify reservation
	reservedAfterPropose, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, amount, reservedAfterPropose)

	// Approve (auto-executes on quorum)
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Executed (auto-executed when quorum reached)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")

	// CRITICAL: Verify reservation is cleaned up (should already work)
	reservedAfterExecute, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterExecute.Cmp(big.NewInt(0)), "Reservation should be cleaned up after execution")
}

// Test 7: Multiple proposals with independent cleanup
func TestGovMinter_CleanupMultipleProposals(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	member2 := minterMembers[2].Operator

	amount1 := big.NewInt(1_000_000)
	amount2 := big.NewInt(2_000_000)
	amount3 := big.NewInt(3_000_000)

	// Use member address as beneficiary (off-chain validation)
	beneficiary := member0.Address

	// Create 3 proposals with unique depositIds
	proofData1, err := CreateMintProof(beneficiary, amount1, "deposit-001", "bank-ref-001", "test mint 1")
	require.NoError(t, err)
	tx, err := gMinter.TxProposeMintWithProof(t, member0, proofData1)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId1, _ := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)

	proofData2, err := CreateMintProof(beneficiary, amount2, "deposit-002", "bank-ref-002", "test mint 2")
	require.NoError(t, err)
	tx, err = gMinter.TxProposeMintWithProof(t, member0, proofData2)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId2, _ := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)

	proofData3, err := CreateMintProof(beneficiary, amount3, "deposit-003", "bank-ref-003", "test mint 3")
	require.NoError(t, err)
	tx, err = gMinter.TxProposeMintWithProof(t, member0, proofData3)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId3, _ := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)

	// Verify total reservation = 6M
	totalReserved, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	expectedTotal := new(big.Int).Add(amount1, amount2)
	expectedTotal.Add(expectedTotal, amount3)
	require.Equal(t, expectedTotal, totalReserved, "Total reservation should be sum of all proposals")

	// Execute proposal 1 (auto-executes on quorum)
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId1)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Cancel proposal 2
	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member0, proposalId2)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Reject proposal 3
	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member1, proposalId3)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member2, proposalId3)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// CRITICAL: Verify all reservations cleaned up
	finalReserved, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, finalReserved.Cmp(big.NewInt(0)), "All reservations should be cleaned up")

	// Verify individual proposal reservations
	p1Reserved, _ := gMinter.GetMintProposalAmount(member0, proposalId1)
	p2Reserved, _ := gMinter.GetMintProposalAmount(member0, proposalId2)
	p3Reserved, _ := gMinter.GetMintProposalAmount(member0, proposalId3)
	require.Equal(t, 0, p1Reserved.Cmp(big.NewInt(0)), "Proposal 1 reservation should be cleared")
	require.Equal(t, 0, p2Reserved.Cmp(big.NewInt(0)), "Proposal 2 reservation should be cleared")
	require.Equal(t, 0, p3Reserved.Cmp(big.NewInt(0)), "Proposal 3 reservation should be cleared")
}

// Test 8: Cleanup idempotency
func TestGovMinter_CleanupIdempotency(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	recipient := member0.Address
	amount := big.NewInt(1_000_000)

	// Create and execute proposal
	tx, err := gMinter.TxProposeMint(t, member0, recipient, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Approve (auto-executes on quorum)
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify cleanup (auto-executed)
	reserved, err := gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(big.NewInt(0)), "Reservation should be cleaned up after auto-execution")

	// Verify proposal is Executed
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")

	// Try to execute again (should fail gracefully, no panic)
	// Note: This will fail because proposal is already executed
	tx, err = gMinter.BaseTxExecuteProposal(t, gMinter.govMinter, member0, proposalId)
	err = gMinter.ExpectedFail(tx, err)
	require.Error(t, err, "Second execution should fail gracefully")

	// Verify reservation still 0 (no negative values)
	reserved, err = gMinter.GetReservedMintAmount(member0)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(big.NewInt(0)), "Reservation should remain 0, not go negative")
}

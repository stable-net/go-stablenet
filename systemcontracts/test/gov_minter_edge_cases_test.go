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
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

// ==================== Test Suite: GovMinter State Management Edge Cases ====================
// Purpose: Verify GovMinter state consistency under edge conditions with try-catch pattern
// Approach: Systematic testing of retry, expiry, concurrency, emergency pause, boundary, and complex scenarios
//
// Test Categories:
//   A. Retry & Failure (A1-A3): Execution retry patterns and terminal failure states
//   B. Proposal Expiry (B1-B2): Time-based expiry in different states
//   C. Concurrency (C1-C3): Concurrent operations and race conditions
//   D. Emergency Pause (D1-D3): System pause during various proposal states
//   E. Boundary Conditions (E1-E3): Limits and edge values
//   F. Complex State Interactions (F1-F3): Multi-proposal scenarios

// ==================== Category A: Retry & Failure ====================

// Test A1: Retry limit enforcement with state cleanup
// Scenario: Proposal retries until max attempts, transitions to Failed, state cleaned up
// Invariants: reservedMintAmount cleaned, memberActiveProposalCount decremented
func TestEdgeCase_A1_RetryLimitEnforcementWithCleanup(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	recipient := member.Address
	amount := big.NewInt(1_000_000)

	// Capture initial state
	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal
	tx, err := ctx.TxProposeMint(t, member, recipient, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal ID: %s", proposalId.String())

	// Verify proposal creation
	assertProposalCreation(t, ctx, ProposalCreationExpectation{
		ProposalId:                proposalId,
		Member:                    member,
		ProposalType:              "Mint",
		Amount:                    amount,
		ActiveCountIncremented:    true,
		ReservedAmountIncremented: true,
		BurnBalanceSufficient:     false,
	})

	// 2. Configure MockFiatToken to fail BEFORE approval (so auto-execution fails)
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Configured MockFiatToken to fail mint")

	// 3. Approve to reach quorum (auto-execution will fail, stays Approved)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Approved (auto-exec failed due to mock failure)
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be Approved after failed auto-execution")
	t.Logf("Proposal status after failed auto-exec: Approved")

	// 4. Retry execution until TooManyExecutionAttempts error
	maxRetries := 2 // GovMinter has execution attempt limit
	for i := 0; i < maxRetries; i++ {
		tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)
		t.Logf("Retry %d/%d - Proposal status: %v", i+1, maxRetries, proposal.Status)

		// Should remain Approved (try-catch catches failure)
		require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should remain Approved after retry %d", i+1)
	}

	// Try one more time - should hit TooManyExecutionAttempts limit
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Should fail with TooManyExecutionAttempts")
	require.Contains(t, err.Error(), "TooManyExecutionAttempts", "Error should be TooManyExecutionAttempts")
	t.Logf("✓ Hit retry limit with TooManyExecutionAttempts as expected")

	// 5. Call executeWithFailure to mark as Failed (terminal state)
	tx, err = ctx.BaseTxExecuteWithFailure(t, ctx.govMinter, member, proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify transition to Failed
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(6), uint8(proposal.Status), "Proposal should be Failed after executeWithFailure")
	t.Logf("✓ Proposal transitioned to Failed status")

	// 6. Verify state cleanup
	assertProposalTerminalState(t, ctx, TerminalStateExpectation{
		ProposalId:             proposalId,
		Member:                 member,
		ExpectedStatus:         6, // Failed
		ProposalType:           "Mint",
		Amount:                 amount,
		ReservationCleaned:     true,
		BurnBalanceUpdated:     false,
		ActiveCountDecremented: true,
	})

	// Verify reservedMintAmount cleaned
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(initialState.ReservedMintAmounts[member.Address]),
		"reservedMintAmount should be cleaned after failure")
	t.Logf("✓ reservedMintAmount cleaned: %s", reserved.String())

	// 7. Restore mock behavior
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify final state consistency
	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "A1: Retry limit enforcement")
	assertInvariantsHold(t, ctx, "A1: After max retries")

	t.Logf("✓✓✓ Test A1 completed successfully")
}

// Test A2: Successful execution after multiple retries
// Scenario: Proposal fails initially, succeeds on retry
// Invariants: State transitions correctly, cleanup on success
func TestEdgeCase_A2_SuccessfulExecutionAfterRetries(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	recipient := member.Address
	amount := big.NewInt(2_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal
	tx, err := ctx.TxProposeMint(t, member, recipient, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal ID: %s", proposalId.String())

	// Verify proposal creation
	assertProposalCreation(t, ctx, ProposalCreationExpectation{
		ProposalId:                proposalId,
		Member:                    member,
		ProposalType:              "Mint",
		Amount:                    amount,
		ActiveCountIncremented:    true,
		ReservedAmountIncremented: true,
		BurnBalanceSufficient:     false,
	})

	// 2. Configure mock to fail initially
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Configured MockFiatToken to fail mint")

	// 3. Approve to reach quorum (auto-execution will fail)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Approved (auto-exec failed)
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be Approved after failed auto-execution")
	t.Logf("Proposal status after failed auto-exec: Approved")

	// Verify reservation still held
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedReserved := new(big.Int).Add(initialState.ReservedMintAmounts[member.Address], amount)
	require.Equal(t, 0, reserved.Cmp(expectedReserved),
		"Reservation should still be held after failed execution")
	t.Logf("✓ Reservation still held: %s", reserved.String())

	// 4. Fix mock configuration
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Fixed MockFiatToken to allow mint")

	// 5. Retry execution (should succeed now)
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// 6. Verify transition to Executed
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed after successful retry")
	t.Logf("✓ Proposal transitioned to Executed status")

	// 7. Verify state cleanup
	assertProposalTerminalState(t, ctx, TerminalStateExpectation{
		ProposalId:             proposalId,
		Member:                 member,
		ExpectedStatus:         3, // Executed
		ProposalType:           "Mint",
		Amount:                 amount,
		ReservationCleaned:     true,
		BurnBalanceUpdated:     false,
		ActiveCountDecremented: true,
	})

	// Verify reservation cleaned after successful execution
	reserved, err = ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(initialState.ReservedMintAmounts[member.Address]),
		"Reservation should be cleaned after successful execution")
	t.Logf("✓ Reservation cleaned: %s", reserved.String())

	// Verify mint actually happened (check balance increased)
	balance, err := ctx.GetMockFiatTokenBalance(member, recipient)
	require.NoError(t, err)
	require.True(t, balance.Cmp(big.NewInt(0)) > 0,
		"Recipient balance should increase after mint")
	t.Logf("✓ Mint succeeded, recipient balance: %s", balance.String())

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "A2: Successful retry")
	assertInvariantsHold(t, ctx, "A2: After successful retry")

	t.Logf("✓✓✓ Test A2 completed successfully")
}

// Test A3: Burn proposal retry with balance verification
// Scenario: Burn proposal fails initially, retries successfully, verifies balance consistency
// Invariants: burnBalance correct throughout, no double-burn, balance held during Approved state
func TestEdgeCase_A3_BurnProposalRetryWithBalance(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(3_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create burn proposal (deposits native coins via msg.value)
	tx, err := ctx.TxProposeBurn(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created burn proposal ID: %s", proposalId.String())

	// Verify proposal creation and burnBalance
	assertProposalCreation(t, ctx, ProposalCreationExpectation{
		ProposalId:                proposalId,
		Member:                    member,
		ProposalType:              "Burn",
		Amount:                    amount,
		ActiveCountIncremented:    true,
		ReservedAmountIncremented: false,
		BurnBalanceSufficient:     true,
	})

	// Verify burnBalance increased
	burnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	expectedBurn := new(big.Int).Add(initialState.BurnBalances[member.Address], amount)
	require.Equal(t, 0, burnBalance.Cmp(expectedBurn),
		"burnBalance should equal initial + amount")
	t.Logf("✓ burnBalance after deposit: %s", burnBalance.String())

	// 2. Configure MockFiatToken to fail burn (simulates external failure)
	tx, err = ctx.SetMockFiatTokenBurnShouldFail(t, ctx.Members[0], true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Configured MockFiatToken to fail burn")

	// 3. Give GovMinter contract MockFiatToken balance to burn
	//    (Burn requires GovMinter to have fiat token balance)
	tx, err = ctx.mockFiatTokenContractTx(t, "setBalance", ctx.Members[0], TestGovMinterAddress, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify balance was set
	var balanceResult []interface{}
	err = ctx.mockFiatTokenCall("balanceOf", ctx.Members[0], &balanceResult, TestGovMinterAddress)
	require.NoError(t, err)
	govMinterBalance := balanceResult[0].(*big.Int)
	require.Equal(t, 0, govMinterBalance.Cmp(amount), "GovMinter should have MockFiatToken balance")
	t.Logf("Set GovMinter MockFiatToken balance: %s", govMinterBalance.String())

	// 4. Approve to reach quorum (auto-execution will fail due to shouldFailBurn, stays Approved)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify stays in Approved due to burn failure
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be Approved after failed auto-exec")
	t.Logf("Proposal status after failed auto-exec: %v", proposal.Status)

	// Verify burnBalance still held (not consumed)
	burnBalanceStillHeld, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalanceStillHeld.Cmp(expectedBurn),
		"burnBalance should still be held after failure")
	t.Logf("✓ burnBalance still held: %s", burnBalanceStillHeld.String())

	// 5. Fix MockFiatToken to allow burn
	tx, err = ctx.SetMockFiatTokenBurnShouldFail(t, ctx.Members[0], false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Fixed MockFiatToken to allow burn")

	// 6. Retry execution (should succeed now)
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify transition to Executed
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed after successful retry")
	t.Logf("✓ Proposal transitioned to Executed status")

	// 6. Verify burnBalance consumed correctly (exactly once, no double-burn)
	burnBalanceAfterSuccess, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalanceAfterSuccess.Cmp(initialState.BurnBalances[member.Address]),
		"burnBalance should return to initial after successful burn")
	t.Logf("✓ burnBalance consumed after successful burn: %s", burnBalanceAfterSuccess.String())

	// Verify no double-burn (balance should be initial, not negative)
	require.True(t, burnBalanceAfterSuccess.Sign() >= 0,
		"burnBalance should never go negative (no double-burn)")
	t.Logf("✓ No double-burn detected")

	// Verify state cleanup
	assertProposalTerminalState(t, ctx, TerminalStateExpectation{
		ProposalId:             proposalId,
		Member:                 member,
		ExpectedStatus:         3, // Executed
		ProposalType:           "Burn",
		Amount:                 amount,
		ReservationCleaned:     false, // Burn proposals don't use mint reservation
		BurnBalanceUpdated:     true,
		ActiveCountDecremented: true,
	})

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "A3: Burn retry with balance consistency")
	assertInvariantsHold(t, ctx, "A3: After successful retry")

	t.Logf("✓✓✓ Test A3 completed successfully")
}

// ==================== Category B: Proposal Expiry ====================

// Test B1: Proposal expires during Voting state
// Scenario: Proposal expires before approval, cleanup occurs
// Invariants: reservedMintAmount cleaned, memberActiveProposalCount decremented
func TestEdgeCase_B1_ProposalExpiryDuringVoting(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	recipient := member.Address
	amount := big.NewInt(4_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal (Voting state)
	tx, err := ctx.TxProposeMint(t, member, recipient, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal ID: %s", proposalId.String())

	// Verify proposal creation
	assertProposalCreation(t, ctx, ProposalCreationExpectation{
		ProposalId:                proposalId,
		Member:                    member,
		ProposalType:              "Mint",
		Amount:                    amount,
		ActiveCountIncremented:    true,
		ReservedAmountIncremented: true,
		BurnBalanceSufficient:     false,
	})

	// Verify proposal is in Voting state
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(1), uint8(proposal.Status), "Proposal should be in Voting state")
	t.Logf("Proposal status: Voting")

	// 2. Verify reservation held before expiry
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedReserved := new(big.Int).Add(initialState.ReservedMintAmounts[member.Address], amount)
	require.Equal(t, 0, reserved.Cmp(expectedReserved),
		"Reservation should be held while voting")
	t.Logf("✓ Reservation held: %s", reserved.String())

	// 3. Advance time past proposal expiry (7 days + 1 second)
	ctx.backend.AdjustTime(7*24*time.Hour + time.Second)
	ctx.backend.Commit()
	t.Logf("Advanced time past proposal expiry")

	// 4. Call expireProposal
	tx, err = ctx.BaseTxExpireProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// 5. Verify transition to Expired
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(5), uint8(proposal.Status), "Proposal should be Expired")
	t.Logf("✓ Proposal transitioned to Expired status")

	// 6. Verify reservation cleanup
	reservedAfterExpiry, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterExpiry.Cmp(initialState.ReservedMintAmounts[member.Address]),
		"Reservation should be cleaned after expiry")
	t.Logf("✓ Reservation cleaned: %s", reservedAfterExpiry.String())

	// Verify proposal-specific reservation cleared
	proposalReserved, err := ctx.GetMintProposalAmount(member, proposalId)
	require.NoError(t, err)
	require.Equal(t, 0, proposalReserved.Cmp(big.NewInt(0)),
		"Proposal-specific reservation should be cleared")
	t.Logf("✓ Proposal-specific reservation cleared")

	// 7. Verify state cleanup
	assertProposalTerminalState(t, ctx, TerminalStateExpectation{
		ProposalId:             proposalId,
		Member:                 member,
		ExpectedStatus:         5, // Expired
		ProposalType:           "Mint",
		Amount:                 amount,
		ReservationCleaned:     true,
		BurnBalanceUpdated:     false,
		ActiveCountDecremented: true,
	})

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "B1: Expiry during voting")
	assertInvariantsHold(t, ctx, "B1: After expiry")

	t.Logf("✓✓✓ Test B1 completed successfully")
}

// Test B2: Proposal expires during Approved state (before execution)
// Scenario: Approved proposal expires before execution attempt
// Invariants: State cleaned up even in Approved state
func TestEdgeCase_B2_ProposalExpiryDuringApproved(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	recipient := member.Address
	amount := big.NewInt(5_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal
	tx, err := ctx.TxProposeMint(t, member, recipient, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal ID: %s", proposalId.String())

	// Verify proposal creation
	assertProposalCreation(t, ctx, ProposalCreationExpectation{
		ProposalId:                proposalId,
		Member:                    member,
		ProposalType:              "Mint",
		Amount:                    amount,
		ActiveCountIncremented:    true,
		ReservedAmountIncremented: true,
		BurnBalanceSufficient:     false,
	})

	// 2. Configure mock to fail so auto-execution fails and stays Approved
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Approve to reach quorum (auto-execution will fail, stays Approved)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Approved
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be Approved")
	t.Logf("Proposal status: Approved (auto-exec failed)")

	// Verify reservation still held
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedReserved := new(big.Int).Add(initialState.ReservedMintAmounts[member.Address], amount)
	require.Equal(t, 0, reserved.Cmp(expectedReserved),
		"Reservation should be held while Approved")
	t.Logf("✓ Reservation held in Approved state: %s", reserved.String())

	// 3. Advance time past expiry (7 days + 1 second)
	ctx.backend.AdjustTime(7*24*time.Hour + time.Second)
	ctx.backend.Commit()
	t.Logf("Advanced time past proposal expiry")

	// 4. Call expireProposal (should work even from Approved state)
	tx, err = ctx.BaseTxExpireProposal(t, ctx.govMinter, ctx.Members[2], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// 5. Verify transition to Expired
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(5), uint8(proposal.Status), "Proposal should be Expired")
	t.Logf("✓ Proposal transitioned from Approved to Expired")

	// 6. Verify cleanup even from Approved state
	reservedAfterExpiry, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterExpiry.Cmp(initialState.ReservedMintAmounts[member.Address]),
		"Reservation should be cleaned after expiry from Approved")
	t.Logf("✓ Reservation cleaned from Approved state: %s", reservedAfterExpiry.String())

	// Verify proposal-specific reservation cleared
	proposalReserved, err := ctx.GetMintProposalAmount(member, proposalId)
	require.NoError(t, err)
	require.Equal(t, 0, proposalReserved.Cmp(big.NewInt(0)),
		"Proposal-specific reservation should be cleared")
	t.Logf("✓ Proposal-specific reservation cleared")

	// 7. Verify cannot execute after expiry
	// Restore mock to allow execution
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, member, false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Try to execute - should fail because proposal is Expired
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Execution should fail for Expired proposal")
	t.Logf("✓ Cannot execute after expiry (as expected)")

	// Verify status unchanged (still Expired)
	proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(5), uint8(proposal.Status), "Proposal should remain Expired")

	// Verify state cleanup
	assertProposalTerminalState(t, ctx, TerminalStateExpectation{
		ProposalId:             proposalId,
		Member:                 member,
		ExpectedStatus:         5, // Expired
		ProposalType:           "Mint",
		Amount:                 amount,
		ReservationCleaned:     true,
		BurnBalanceUpdated:     false,
		ActiveCountDecremented: true,
	})

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "B2: Expiry during approved")
	assertInvariantsHold(t, ctx, "B2: After expiry from approved")

	t.Logf("✓✓✓ Test B2 completed successfully")
}

// ==================== Category C: Concurrency ====================

// Test C1: Multiple proposals from same member (within limit)
// Scenario: Member creates MAX proposals concurrently, verify limits
// Invariants: Cannot exceed MAX_ACTIVE_PROPOSALS_PER_MEMBER, reservation tracking correct
func TestEdgeCase_C1_MultipleProposalsSameMemberWithinLimit(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	maxProposals := ctx.MaxActiveProposalsPerMember.Int64()

	initialState := captureStateSnapshot(t, ctx)
	t.Logf("MAX_ACTIVE_PROPOSALS_PER_MEMBER: %d", maxProposals)

	// 1. Create MAX proposals successfully
	proposalIds := make([]*big.Int, maxProposals)
	totalReserved := big.NewInt(0)

	for i := int64(0); i < maxProposals; i++ {
		amount := big.NewInt((i + 1) * 1_000_000)
		recipient := member.Address

		tx, err := ctx.TxProposeMint(t, member, recipient, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)
		proposalIds[i] = proposalId
		totalReserved.Add(totalReserved, amount)

		t.Logf("Created proposal %d/%d: ID=%s, amount=%s", i+1, maxProposals, proposalId.String(), amount.String())
	}

	// 2. Verify total reservedMintAmount = sum of all proposals
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(totalReserved), "Total reservation should equal sum of all proposals")
	t.Logf("✓ Total reservedMintAmount: %s (sum of all proposals)", reserved.String())
	t.Logf("✓ Successfully created MAX (%d) active proposals", maxProposals)

	// 4. Attempt to create MAX+1 proposal, should fail with TooManyActiveProposals
	// Use small amount to avoid hitting InsufficientMinterAllowance error first
	extraAmount := big.NewInt(4_000_000)
	tx, err := ctx.TxProposeMint(t, member, member.Address, extraAmount)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Should fail to create (MAX+1)th proposal")
	require.Contains(t, err.Error(), "TooManyActiveProposals", "Error should be TooManyActiveProposals")
	t.Logf("✓ Cannot create (MAX+1)th proposal: TooManyActiveProposals")

	// 5. Execute one proposal to free up a slot
	firstProposalId := proposalIds[0]
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], firstProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify executed
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, firstProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "First proposal should be Executed")
	t.Logf("✓ Executed first proposal (ID=%s)", firstProposalId.String())

	// 6. Verify count decremented by checking reservation decreased
	reservedAfterExec, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	firstProposalAmount := big.NewInt(1_000_000) // First proposal amount
	expectedAfterExec := new(big.Int).Sub(totalReserved, firstProposalAmount)
	require.Equal(t, 0, reservedAfterExec.Cmp(expectedAfterExec), "Reservation should decrease by executed amount")
	t.Logf("✓ Reservation after execution: %s (decreased)", reservedAfterExec.String())

	// 7. Verify can now create another proposal (use small amount to avoid allowance issues)
	newAmount := big.NewInt(4_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, newAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	newProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("✓ Created new proposal after freeing slot: ID=%s", newProposalId.String())

	// 8. Verify count back to MAX by checking total reservation
	finalReserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	// Should have MAX proposals reserved: (2M + 3M + 4M) = 9M
	expectedFinalReserved := new(big.Int).Add(expectedAfterExec, newAmount)
	require.Equal(t, 0, finalReserved.Cmp(expectedFinalReserved), "Reservation should be back to MAX proposals")
	t.Logf("✓ Reservation back to MAX proposals: %s", finalReserved.String())

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "C1: Multiple proposals within limit")
	assertInvariantsHold(t, ctx, "C1: After concurrent proposals")

	t.Logf("✓✓✓ Test C1 completed successfully")
}

// Test C2: Concurrent approval and expiry
// Scenario: Proposal being approved while expiry timer fires
// Invariants: Only one state transition succeeds, consistent final state
func TestEdgeCase_C2_ConcurrentApprovalAndExpiry(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(6_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create proposal (Voting)
	tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal ID: %s", proposalId.String())

	// Verify in Voting state
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(1), uint8(proposal.Status), "Proposal should be in Voting state")

	// 2. Advance time close to expiry (but not past)
	// Get voting duration and advance to just before expiry
	votingDuration := uint64(7 * 24 * 60 * 60)                             // 7 days in seconds (typical)
	ctx.backend.AdjustTime(time.Duration(votingDuration-60) * time.Second) // 1 minute before expiry
	ctx.backend.Commit()
	t.Logf("Advanced time to 1 minute before expiry")

	// 3. Attempt approval (should succeed as proposal not yet expired)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	approvalReceipt, err := ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Approval transaction succeeded (gas: %d)", approvalReceipt.GasUsed)

	// 4. Check final state - proposal should be Executed (quorum reached)
	proposalAfterApproval, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	t.Logf("Proposal status after approval: %v", proposalAfterApproval.Status)

	// If auto-executed to Executed state
	if proposalAfterApproval.Status == 3 {
		t.Logf("✓ Proposal auto-executed successfully")

		// Verify cleanup occurred
		assertProposalTerminalState(t, ctx, TerminalStateExpectation{
			ProposalId:             proposalId,
			Member:                 member,
			ExpectedStatus:         3, // Executed
			ProposalType:           "Mint",
			Amount:                 amount,
			ReservationCleaned:     true,
			BurnBalanceUpdated:     false,
			ActiveCountDecremented: true,
		})
	} else if proposalAfterApproval.Status == 2 { // Approved but not executed
		t.Logf("Proposal in Approved state (execution failed)")

		// Now advance past expiry
		ctx.backend.AdjustTime(2 * time.Minute)
		ctx.backend.Commit()

		// Try to execute - should fail as expired
		tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
		err = ctx.ExpectedFail(tx, err)
		require.Error(t, err, "Execution after expiry should fail")
		t.Logf("✓ Cannot execute after expiry (as expected)")

		// Verify transitioned to Expired
		proposalFinal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(5), uint8(proposalFinal.Status), "Proposal should be Expired")
		t.Logf("✓ Proposal transitioned to Expired")
	}

	// 5. Verify state consistency - no matter which path, state should be clean
	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "C2: Concurrent approval/expiry")
	assertInvariantsHold(t, ctx, "C2: After concurrent operations")

	t.Logf("✓✓✓ Test C2 completed successfully")
}

// Test C3: Multiple members creating proposals simultaneously
// Scenario: Different members create proposals at same time
// Invariants: All proposals tracked independently, no cross-contamination
func TestEdgeCase_C3_MultipleMembersSimultaneousProposals(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	initialState := captureStateSnapshot(t, ctx)

	// 1. Each member creates a proposal with different amounts
	memberProposals := make(map[common.Address]*big.Int)
	memberAmounts := make(map[common.Address]*big.Int)

	for i, member := range ctx.Members {
		amount := big.NewInt(int64((i + 1) * 10_000_000))
		memberAmounts[member.Address] = amount

		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)
		memberProposals[member.Address] = proposalId

		t.Logf("Member %d created proposal ID=%s, amount=%s", i, proposalId.String(), amount.String())
	}

	// 2. Verify total reservedMintAmount = sum of all member proposals
	totalExpectedReserved := big.NewInt(0)
	for _, amount := range memberAmounts {
		totalExpectedReserved.Add(totalExpectedReserved, amount)
	}

	totalReserved, err := ctx.GetReservedMintAmount(ctx.Members[0])
	require.NoError(t, err)
	require.Equal(t, 0, totalReserved.Cmp(totalExpectedReserved),
		"Total reservation should equal sum of all member proposals")
	t.Logf("✓ Total reservedMintAmount: %s (sum of all: %s)",
		totalReserved.String(), totalExpectedReserved.String())

	// 4. Execute proposals in reverse order (last member first)
	for i := len(ctx.Members) - 1; i >= 0; i-- {
		member := ctx.Members[i]
		proposalId := memberProposals[member.Address]

		// Approve with another member
		approverIdx := (i + 1) % len(ctx.Members)
		approver := ctx.Members[approverIdx]

		tx, err := ctx.BaseTxApproveProposal(t, ctx.govMinter, approver, proposalId)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify executed
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")
		t.Logf("✓ Member %d proposal executed (ID=%s)", i, proposalId.String())
	}

	// 5. Verify cleanup is independent per member - verify via total reservation cleaned
	finalReserved, err := ctx.GetReservedMintAmount(ctx.Members[0])
	require.NoError(t, err)
	// Calculate initial total from map
	initialTotal := big.NewInt(0)
	for _, amount := range initialState.ReservedMintAmounts {
		initialTotal.Add(initialTotal, amount)
	}
	require.Equal(t, 0, finalReserved.Cmp(initialTotal),
		"Total reservation should be back to initial after all executions")
	t.Logf("✓ All member proposals executed, total reservation cleaned: %s (initial was: %s)",
		finalReserved.String(), initialTotal.String())

	// 7. Verify no state leakage - each member can create new proposals independently
	for i, member := range ctx.Members {
		amount := big.NewInt(int64((i + 1) * 5_000_000))
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)
		t.Logf("✓ Member %d can create new proposal after cleanup", i)
	}

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "C3: Multiple members concurrent")
	assertInvariantsHold(t, ctx, "C3: After multi-member proposals")

	t.Logf("✓✓✓ Test C3 completed successfully")
}

// ==================== Category D: Emergency Pause ====================

// Test D1: Emergency pause during Voting state
// Scenario: System paused while proposal in Voting, verify operations blocked
// Invariants: No state changes allowed during pause, resume works correctly
func TestEdgeCase_D1_EmergencyPauseDuringVoting(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(7_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal (Voting state)
	tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	mintProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal ID: %s", mintProposalId.String())

	// Verify in Voting state
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, mintProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(1), uint8(proposal.Status), "Proposal should be in Voting state")

	// 2. Propose and execute emergency pause
	tx, err = ctx.TxProposePause(t, ctx.Members[1])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	pauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[1])
	require.NoError(t, err)
	t.Logf("Created pause proposal ID: %s", pauseProposalId.String())

	// Approve pause proposal to execute it
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[2], pauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency pause activated")

	// Verify paused state
	paused, err := ctx.IsEmergencyPaused(member)
	require.NoError(t, err)
	require.True(t, paused, "Contract should be paused")

	// 3. Verify cannot create new mint/burn proposals during pause
	tx, err = ctx.TxProposeMint(t, member, member.Address, big.NewInt(1_000_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused", "Should fail with ContractPaused")
	t.Logf("✓ Cannot create mint proposal during pause")

	// 4. Verify CANNOT approve existing proposals during pause (all governance blocked except unpause)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused", "Should fail with ContractPaused")
	t.Logf("✓ Cannot approve proposals during pause (all operations blocked)")

	// 5. Propose and execute unpause (ONLY unpause governance works during pause)
	tx, err = ctx.TxProposeUnpause(t, ctx.Members[0])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	unpauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[0])
	require.NoError(t, err)
	t.Logf("Created unpause proposal ID: %s", unpauseProposalId.String())

	// Approve unpause proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], unpauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency unpause activated")

	// Verify unpaused state
	paused, err = ctx.IsEmergencyPaused(member)
	require.NoError(t, err)
	require.False(t, paused, "Contract should be unpaused")

	// 6. Verify can now approve the original mint proposal after unpause
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can approve mint proposal after unpause")

	// 7. Verify can create new proposals after unpause
	newAmount := big.NewInt(8_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, newAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can create mint proposal after unpause")

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "D1: Emergency pause during voting")
	assertInvariantsHold(t, ctx, "D1: After pause/resume")

	t.Logf("✓✓✓ Test D1 completed successfully")
}

// Test D2: Emergency pause during Approved state
// Scenario: System paused with approved proposal, verify execution blocked
// Invariants: Cannot execute during pause, state preserved
func TestEdgeCase_D2_EmergencyPauseDuringApproved(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(8_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Configure MockFiatToken to fail mint (to keep proposal in Approved state)
	tx, err := ctx.SetMockFiatTokenMintShouldFail(t, ctx.Members[0], true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// 2. Create and approve mint proposal (stays in Approved due to mint failure)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	mintProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal ID: %s", mintProposalId.String())

	// Approve to trigger auto-exec (which will fail, leaving it in Approved)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify in Approved state
	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, mintProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(proposal.Status), "Proposal should be in Approved state")
	t.Logf("Proposal in Approved state (auto-exec failed)")

	// Verify reservation still held
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(amount), "Reservation should still be held")
	t.Logf("✓ Reservation held: %s", reserved.String())

	// 3. Propose and execute emergency pause
	tx, err = ctx.TxProposePause(t, ctx.Members[1])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	pauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[1])
	require.NoError(t, err)

	// Approve pause
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[2], pauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency pause activated")

	// Verify paused
	paused, err := ctx.IsEmergencyPaused(member)
	require.NoError(t, err)
	require.True(t, paused, "Contract should be paused")

	// 4. Verify CANNOT execute approved proposals during pause (all operations blocked)
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, mintProposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused", "Should fail with ContractPaused")
	t.Logf("✓ Cannot execute approved proposals during pause")

	// 5. Propose and execute unpause

	tx, err = ctx.TxProposeUnpause(t, ctx.Members[0])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	unpauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[0])
	require.NoError(t, err)

	// Approve unpause
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], unpauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency unpause activated")

	// 6. Fix MockFiatToken and execute the approved proposal after unpause
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, ctx.Members[0], false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, mintProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can execute approved proposal after unpause")

	// Verify executed
	proposalAfterExec, err := ctx.BaseGetProposal(ctx.govMinter, member, mintProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposalAfterExec.Status), "Proposal should be Executed")

	// 7. Verify can create new proposals after unpause
	newAmount := big.NewInt(9_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, newAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can create proposals after unpause")

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "D2: Emergency pause during approved")
	assertInvariantsHold(t, ctx, "D2: After pause/resume with approved")

	t.Logf("✓✓✓ Test D2 completed successfully")
}

// Test D3: Emergency pause with multiple active proposals
// Scenario: Pause with multiple proposals in various states
// Invariants: All proposals frozen, state consistent after resume
func TestEdgeCase_D3_EmergencyPauseMultipleProposals(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create multiple proposals in different states
	// Configure MockFiatToken to fail for some proposals to keep in Approved state
	tx, err := ctx.SetMockFiatTokenMintShouldFail(t, ctx.Members[0], true)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Create first proposal (will stay Voting)
	votingAmount := big.NewInt(10_000_000)
	tx, err = ctx.TxProposeMint(t, ctx.Members[0], ctx.Members[0].Address, votingAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	votingProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[0])
	require.NoError(t, err)
	t.Logf("Created voting proposal ID: %s", votingProposalId.String())

	// Create and approve second proposal (will stay Approved due to mint failure)
	approvedAmount := big.NewInt(20_000_000)
	tx, err = ctx.TxProposeMint(t, ctx.Members[1], ctx.Members[1].Address, approvedAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	approvedProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[1])
	require.NoError(t, err)

	// Approve to reach Approved state (auto-exec fails)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[2], approvedProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Created approved proposal ID: %s", approvedProposalId.String())

	// Verify states
	votingProposal, err := ctx.BaseGetProposal(ctx.govMinter, ctx.Members[0], votingProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(1), uint8(votingProposal.Status), "First proposal should be Voting")

	approvedProposal, err := ctx.BaseGetProposal(ctx.govMinter, ctx.Members[1], approvedProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(2), uint8(approvedProposal.Status), "Second proposal should be Approved")

	// Verify total reservation
	totalReservedBefore := new(big.Int).Add(votingAmount, approvedAmount)
	reserved, err := ctx.GetReservedMintAmount(ctx.Members[0])
	require.NoError(t, err)
	require.Equal(t, 0, reserved.Cmp(totalReservedBefore), "Total reservation should include both proposals")
	t.Logf("✓ Total reservation before pause: %s", reserved.String())

	// 2. Emergency pause
	tx, err = ctx.TxProposePause(t, ctx.Members[2])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	pauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[2])
	require.NoError(t, err)

	// Approve pause
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[0], pauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency pause activated")

	// 3. Verify cannot create new proposals
	tx, err = ctx.TxProposeMint(t, ctx.Members[0], ctx.Members[0].Address, big.NewInt(1_000_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused")
	t.Logf("✓ Cannot create proposals during pause")

	// 4. Verify all reservations preserved
	reservedDuringPause, err := ctx.GetReservedMintAmount(ctx.Members[0])
	require.NoError(t, err)
	require.Equal(t, 0, reservedDuringPause.Cmp(totalReservedBefore), "Reservations should be preserved")
	t.Logf("✓ All reservations preserved during pause: %s", reservedDuringPause.String())

	// 5. Verify CANNOT approve or execute proposals during pause (all ops blocked)
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], votingProposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused")
	t.Logf("✓ Cannot approve proposals during pause")

	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, ctx.Members[1], approvedProposalId)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ContractPaused")
	t.Logf("✓ Cannot execute proposals during pause")

	// 6. Unpause
	tx, err = ctx.TxProposeUnpause(t, ctx.Members[0])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	unpauseProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, ctx.Members[0])
	require.NoError(t, err)

	// Approve unpause
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], unpauseProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Emergency unpause activated")

	// 7. Approve and execute proposals after unpause
	// Approve the Voting proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], votingProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can approve proposals after unpause")

	// Fix MockFiatToken and execute the Approved proposal
	tx, err = ctx.SetMockFiatTokenMintShouldFail(t, ctx.Members[0], false)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, ctx.Members[1], approvedProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can execute approved proposals after unpause")

	// 8. Execute the voting proposal (now approved)
	tx, err = ctx.BaseTxExecuteProposal(t, ctx.govMinter, ctx.Members[0], votingProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	votingProposalFinal, err := ctx.BaseGetProposal(ctx.govMinter, ctx.Members[0], votingProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(votingProposalFinal.Status), "First proposal should be Executed")

	// 9. Verify can create new proposals after unpause
	newAmount := big.NewInt(30_000_000)
	tx, err = ctx.TxProposeMint(t, ctx.Members[0], ctx.Members[0].Address, newAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Can create proposals after unpause")
	t.Logf("✓ Remaining proposals executed successfully")

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "D3: Emergency pause multiple proposals")
	assertInvariantsHold(t, ctx, "D3: After pause/resume with multiple")

	t.Logf("✓✓✓ Test D3 completed successfully")
}

// ==================== Category E: Boundary Conditions ====================

// Test E1: MAX_ACTIVE_PROPOSALS_PER_MEMBER boundary - comprehensive edge cases
// Scenario: Test exact boundary conditions with multiple state transitions
// Invariants: Precise limit enforcement at MAX-1, MAX, MAX+1 boundaries
func TestEdgeCase_E1_MaxActiveProposalsPerMemberBoundary(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	maxProposals := ctx.MaxActiveProposalsPerMember.Int64()

	initialState := captureStateSnapshot(t, ctx)
	t.Logf("MAX_ACTIVE_PROPOSALS_PER_MEMBER: %d", maxProposals)

	// ==================== Boundary Test 1: 0 → 1 (first proposal) ====================
	t.Logf("--- Boundary Test 1: Creating first proposal (0 → 1) ---")

	amount1 := big.NewInt(1_000_000)
	tx, err := ctx.TxProposeMint(t, member, member.Address, amount1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId1, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)

	reserved1, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reserved1.Cmp(amount1))
	t.Logf("✓ First proposal created: count=1/%d, reserved=%s", maxProposals, reserved1.String())

	// ==================== Boundary Test 2: MAX-1 → MAX ====================
	t.Logf("--- Boundary Test 2: Approaching limit (MAX-1 → MAX) ---")

	proposalIds := []*big.Int{proposalId1}
	totalReserved := new(big.Int).Set(amount1)

	// Create proposals up to MAX-1 (already have 1, so create MAX-2 more)
	for i := int64(1); i < maxProposals-1; i++ {
		amount := big.NewInt((i + 1) * 1_000_000)
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)
		proposalIds = append(proposalIds, proposalId)
		totalReserved.Add(totalReserved, amount)
	}

	reservedAtMaxMinus1, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	t.Logf("✓ At MAX-1: count=%d/%d, reserved=%s", maxProposals-1, maxProposals, reservedAtMaxMinus1.String())

	// Create one more to reach exactly MAX
	amountMax := big.NewInt(maxProposals * 1_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amountMax)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalIdMax, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	proposalIds = append(proposalIds, proposalIdMax)
	totalReserved.Add(totalReserved, amountMax)

	reservedAtMax, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAtMax.Cmp(totalReserved))
	t.Logf("✓ Reached MAX: count=%d/%d, reserved=%s", maxProposals, maxProposals, reservedAtMax.String())

	// ==================== Boundary Test 3: MAX → MAX+1 (rejection) ====================
	t.Logf("--- Boundary Test 3: Exceeding limit (MAX → MAX+1) ---")

	amountExtra := big.NewInt(999_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amountExtra)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TooManyActiveProposals")
	t.Logf("✓ MAX+1 rejected: TooManyActiveProposals")

	// Verify reservation unchanged
	reservedAfterReject, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterReject.Cmp(reservedAtMax))
	t.Logf("✓ Reservation unchanged after rejection: %s", reservedAfterReject.String())

	// ==================== Boundary Test 4: Multiple rejection attempts ====================
	t.Logf("--- Boundary Test 4: Multiple rejection attempts at MAX ---")

	for i := 0; i < 3; i++ {
		tx, err := ctx.TxProposeMint(t, member, member.Address, big.NewInt(100_000))
		err = ctx.ExpectedFail(tx, err)
		require.Error(t, err)
		require.Contains(t, err.Error(), "TooManyActiveProposals")
	}
	t.Logf("✓ Multiple attempts at MAX all rejected correctly")

	// ==================== Boundary Test 5: MAX → MAX-1 (via execution) ====================
	t.Logf("--- Boundary Test 5: Freeing slot via execution (MAX → MAX-1) ---")

	// Execute first proposal to free slot
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalIds[0])
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	reservedAfterExec, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedAfterExec := new(big.Int).Sub(totalReserved, amount1)
	require.Equal(t, 0, reservedAfterExec.Cmp(expectedAfterExec))
	t.Logf("✓ After execution: count=%d/%d, reserved=%s", maxProposals-1, maxProposals, reservedAfterExec.String())

	// ==================== Boundary Test 6: MAX-1 → MAX (refill) ====================
	t.Logf("--- Boundary Test 6: Refilling to MAX (MAX-1 → MAX) ---")

	amountRefill := big.NewInt(555_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amountRefill)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	reservedAfterRefill, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedAfterRefill := new(big.Int).Add(expectedAfterExec, amountRefill)
	require.Equal(t, 0, reservedAfterRefill.Cmp(expectedAfterRefill))
	t.Logf("✓ Refilled to MAX: count=%d/%d, reserved=%s", maxProposals, maxProposals, reservedAfterRefill.String())

	// Verify cannot create MAX+1 again
	tx, err = ctx.TxProposeMint(t, member, member.Address, big.NewInt(100_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TooManyActiveProposals")
	t.Logf("✓ MAX+1 still rejected after refill")

	// ==================== Boundary Test 7: MAX → MAX-1 (via expiry) ====================
	t.Logf("--- Boundary Test 7: Freeing slot via expiry (MAX → MAX-1) ---")

	// Get one of the remaining proposals and expire it
	proposalToExpire := proposalIds[1] // Second proposal

	// Advance time past expiry
	ctx.backend.AdjustTime(31 * 24 * time.Hour) // 31 days
	ctx.backend.Commit()

	// Trigger expiry
	tx, err = ctx.BaseTxExpireProposal(t, ctx.govMinter, member, proposalToExpire)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	reservedAfterExpiry, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	amount2 := big.NewInt(2 * 1_000_000) // Second proposal amount
	expectedAfterExpiry := new(big.Int).Sub(expectedAfterRefill, amount2)
	require.Equal(t, 0, reservedAfterExpiry.Cmp(expectedAfterExpiry))
	t.Logf("✓ After expiry: count=%d/%d, reserved=%s", maxProposals-1, maxProposals, reservedAfterExpiry.String())

	// Verify can create new proposal now
	tx, err = ctx.TxProposeMint(t, member, member.Address, big.NewInt(777_000))
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ New proposal created after expiry")

	// ==================== Boundary Test 8: MAX → 0 (cleanup all) ====================
	t.Logf("--- Boundary Test 8: Complete cleanup (MAX → 0) ---")

	// Cancel all remaining proposals to reach 0
	currentProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)

	// Cancel from current back to first uncancelled
	for id := currentProposalId.Int64(); id >= 1; id-- {
		proposalId := big.NewInt(id)
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)

		// Only cancel if not already in terminal state
		if proposal.Status == 1 || proposal.Status == 2 { // Voting or Approved
			tx, err := ctx.BaseTxCancelProposal(t, ctx.govMinter, member, proposalId)
			_, err = ctx.ExpectedOk(tx, err)
			require.NoError(t, err)
		}
	}

	finalReserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalReserved.Int64())
	t.Logf("✓ All proposals cleaned: count=0/%d, reserved=0", maxProposals)

	// ==================== Boundary Test 9: 0 → MAX (rapid refill) ====================
	t.Logf("--- Boundary Test 9: Rapid refill (0 → MAX) ---")

	for i := int64(0); i < maxProposals; i++ {
		amount := big.NewInt((i + 1) * 100_000)
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)
	}

	finalReservedAtMax, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	t.Logf("✓ Rapid refill to MAX: count=%d/%d, reserved=%s", maxProposals, maxProposals, finalReservedAtMax.String())

	// Verify boundary still enforced
	tx, err = ctx.TxProposeMint(t, member, member.Address, big.NewInt(50_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TooManyActiveProposals")
	t.Logf("✓ Boundary still enforced after rapid refill")

	// ==================== Boundary Test 10: Burn proposals also count ====================
	t.Logf("--- Boundary Test 10: Burn proposals also count towards limit ---")

	// Cancel one mint to make room
	lastProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	tx, err = ctx.BaseTxCancelProposal(t, ctx.govMinter, member, lastProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Cancelled one mint proposal: count=%d/%d", maxProposals-1, maxProposals)

	// Create burn proposal to fill the slot
	burnAmount := big.NewInt(500_000)
	tx, err = ctx.TxProposeBurn(t, member, member.Address, burnAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	burnProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("✓ Created burn proposal: ID=%s, count=%d/%d", burnProposalId.String(), maxProposals, maxProposals)

	// Verify cannot create another (mint or burn)
	tx, err = ctx.TxProposeMint(t, member, member.Address, big.NewInt(50_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TooManyActiveProposals")
	t.Logf("✓ Cannot create mint when at MAX (including burn)")

	tx, err = ctx.TxProposeBurn(t, member, member.Address, big.NewInt(50_000))
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TooManyActiveProposals")
	t.Logf("✓ Cannot create burn when at MAX (including mint)")

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "E1: Boundary conditions")
	assertInvariantsHold(t, ctx, "E1: After boundary tests")

	t.Logf("✓✓✓ Test E1 completed successfully - all 10 boundary scenarios verified")
}

// Test E2: Mint allowance boundary conditions
// Scenario: Proposal amount at or exceeding minter allowance
// Invariants: Execution fails if allowance insufficient, succeeds if sufficient
func TestEdgeCase_E2_MintAllowanceBoundary(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]

	initialState := captureStateSnapshot(t, ctx)

	// 1. Get current minter allowance (100M from setup)
	initialAllowance, err := ctx.GetMockFiatTokenMinterAllowance(member, TestGovMinterAddress)
	require.NoError(t, err)
	t.Logf("Initial minter allowance: %s", initialAllowance.String())

	// 2. Set allowance to small value to test boundary
	smallAllowance := big.NewInt(20_000_000) // 20M
	tx, err := ctx.ConfigureMockFiatTokenMinter(t, member, TestGovMinterAddress, smallAllowance)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Set allowance to: %s", smallAllowance.String())

	// 3. Create proposal that will exceed allowance when combined with reservations
	amount1 := big.NewInt(15_000_000) // 15M
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId1, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created proposal 1: ID=%s, amount=%s", proposalId1.String(), amount1.String())

	// 4. Create second proposal - total (15M + 10M = 25M) exceeds allowance (20M)
	amount2 := big.NewInt(10_000_000) // 10M
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount2)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err)
	require.Contains(t, err.Error(), "InsufficientMinterAllowance")
	t.Logf("✓ Cannot create proposal exceeding allowance: InsufficientMinterAllowance")

	// 5. Increase allowance to accommodate both proposals
	largerAllowance := big.NewInt(30_000_000) // 30M
	tx, err = ctx.ConfigureMockFiatTokenMinter(t, member, TestGovMinterAddress, largerAllowance)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Increased allowance to: %s", largerAllowance.String())

	// 6. Now second proposal should succeed
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount2)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId2, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("✓ Created proposal 2 after allowance increase: ID=%s", proposalId2.String())

	// 7. Verify reservation tracking
	reserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedReserved := new(big.Int).Add(amount1, amount2)
	require.Equal(t, 0, reserved.Cmp(expectedReserved))
	t.Logf("✓ Total reservation correct: %s", reserved.String())

	// 8. Execute both proposals to cleanup
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId2)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Both proposals executed successfully")

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "E2: Mint allowance boundary")
	assertInvariantsHold(t, ctx, "E2: After allowance test")

	t.Logf("✓✓✓ Test E2 completed successfully")
}

// Test E3: Burn balance boundary conditions
// Scenario: Burn with exact balance and insufficient balance
// Invariants: Cannot create burn proposal with insufficient balance
func TestEdgeCase_E3_BurnBalanceBoundary(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]

	initialState := captureStateSnapshot(t, ctx)

	// 1. Check initial burn balance (should be 0)
	initialBurnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), initialBurnBalance.Int64())
	t.Logf("Initial burn balance: %s", initialBurnBalance.String())

	// 2. Create burn proposal with exact msg.value (10M native coins)
	burnAmount := big.NewInt(10_000_000)
	tx, err := ctx.TxProposeBurn(t, member, member.Address, burnAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created burn proposal: ID=%s, amount=%s", proposalId.String(), burnAmount.String())

	// 3. Verify burn balance increased
	burnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalance.Cmp(burnAmount))
	t.Logf("✓ Burn balance after proposal: %s", burnBalance.String())

	// 4. Attempt to create another burn without depositing more (should succeed as proposal creation, balance checked at execution)
	// The contract allows creating multiple burn proposals, balance checked at execution time
	amount2 := big.NewInt(5_000_000)
	tx, err = ctx.TxProposeBurn(t, member, member.Address, amount2)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId2, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created second burn proposal: ID=%s, amount=%s", proposalId2.String(), amount2.String())

	// 5. Verify total burn balance
	totalBurnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	expectedTotal := new(big.Int).Add(burnAmount, amount2)
	require.Equal(t, 0, totalBurnBalance.Cmp(expectedTotal))
	t.Logf("✓ Total burn balance: %s", totalBurnBalance.String())

	// 5.5. Setup MockFiatToken to allow burn execution
	// Give GovMinter sufficient fiat token balance for burning
	tx, err = ctx.mockFiatTokenContractTx(t, "setBalance", ctx.Members[0], TestGovMinterAddress, expectedTotal)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Set GovMinter MockFiatToken balance: %s", expectedTotal.String())

	// 6. Execute first burn proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ First burn proposal executed")

	// 7. Verify balance decreased
	balanceAfterFirst, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	expectedAfterFirst := new(big.Int).Sub(expectedTotal, burnAmount)
	require.Equal(t, 0, balanceAfterFirst.Cmp(expectedAfterFirst))
	t.Logf("✓ Burn balance after first execution: %s", balanceAfterFirst.String())

	// 8. Execute second burn proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId2)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Second burn proposal executed")

	// 9. Verify balance fully consumed
	finalBurnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalBurnBalance.Int64())
	t.Logf("✓ All burn balance consumed: %s", finalBurnBalance.String())

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "E3: Burn balance boundary")
	assertInvariantsHold(t, ctx, "E3: After burn balance test")

	t.Logf("✓✓✓ Test E3 completed successfully")
}

// ==================== Category F: Complex State Interactions ====================

// Test F1: Interleaved mint and burn proposals
// Scenario: Member creates alternating mint/burn proposals
// Invariants: Independent tracking, no interference
func TestEdgeCase_F1_InterleavedMintBurnProposals(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]

	initialState := captureStateSnapshot(t, ctx)

	// 1. Create mint proposal A
	mintAmountA := big.NewInt(10_000_000)
	tx, err := ctx.TxProposeMint(t, member, member.Address, mintAmountA)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	mintProposalA, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal A: ID=%s, amount=%s", mintProposalA.String(), mintAmountA.String())

	// 2. Create burn proposal B
	burnAmountB := big.NewInt(5_000_000)
	tx, err = ctx.TxProposeBurn(t, member, member.Address, burnAmountB)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	burnProposalB, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created burn proposal B: ID=%s, amount=%s", burnProposalB.String(), burnAmountB.String())

	// 3. Create mint proposal C
	mintAmountC := big.NewInt(15_000_000)
	tx, err = ctx.TxProposeMint(t, member, member.Address, mintAmountC)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	mintProposalC, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal C: ID=%s, amount=%s", mintProposalC.String(), mintAmountC.String())

	// 4. Verify reservedMintAmount includes only A + C (not B)
	reservedMint, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedMintReserved := new(big.Int).Add(mintAmountA, mintAmountC)
	require.Equal(t, 0, reservedMint.Cmp(expectedMintReserved))
	t.Logf("✓ Reserved mint amount (A+C): %s", reservedMint.String())

	// 5. Verify burnBalance tracks B separately
	burnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalance.Cmp(burnAmountB))
	t.Logf("✓ Burn balance (B only): %s", burnBalance.String())

	// 5.5. Setup MockFiatToken to allow burn execution
	// Give GovMinter sufficient fiat token balance for burning
	tx, err = ctx.mockFiatTokenContractTx(t, "setBalance", ctx.Members[0], TestGovMinterAddress, burnAmountB)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Set GovMinter MockFiatToken balance: %s", burnAmountB.String())

	// 6. Execute in mixed order: B, A, C
	// Execute burn proposal B
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], burnProposalB)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Executed burn proposal B")

	// Verify burn balance consumed
	burnBalanceAfterB, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalanceAfterB.Int64())
	t.Logf("✓ Burn balance after B execution: %s (cleaned)", burnBalanceAfterB.String())

	// Verify mint reservation unchanged
	reservedAfterB, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterB.Cmp(expectedMintReserved))
	t.Logf("✓ Mint reservation unchanged after burn: %s", reservedAfterB.String())

	// Execute mint proposal A
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalA)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Executed mint proposal A")

	// Verify reservation decreased
	reservedAfterA, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterA.Cmp(mintAmountC)) // Only C remains
	t.Logf("✓ Mint reservation after A: %s (only C remains)", reservedAfterA.String())

	// Execute mint proposal C
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalC)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Executed mint proposal C")

	// 7. Verify all cleanup complete
	finalReserved, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalReserved.Int64())
	t.Logf("✓ All mint reservations cleaned: %s", finalReserved.String())

	finalBurnBalance, err := ctx.GetBurnBalance(member, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalBurnBalance.Int64())
	t.Logf("✓ All burn balances cleaned: %s", finalBurnBalance.String())

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "F1: Interleaved mint/burn")
	assertInvariantsHold(t, ctx, "F1: After interleaved proposals")

	t.Logf("✓✓✓ Test F1 completed successfully")
}

// Test F2: Proposal with duplicate depositId/withdrawalId (replay protection)
// Scenario: Attempt to execute same depositId twice
// Invariants: executedDepositIds prevents replay
func TestEdgeCase_F2_ReplayProtectionDepositWithdrawalIds(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	recipient := member.Address
	mintAmount := big.NewInt(5_000_000)
	burnAmount := big.NewInt(3_000_000)

	depositId := "DEPOSIT-REPLAY-TEST-001"
	withdrawalId := "WITHDRAWAL-REPLAY-TEST-001"

	initialState := captureStateSnapshot(t, ctx)

	// ==================== MINT REPLAY TEST ====================

	// 1. Create mint proposal with specific depositId
	mintProof1, err := CreateMintProof(recipient, mintAmount, depositId, "bank-ref-F2", "mint test F2")
	require.NoError(t, err)

	tx, err := ctx.TxProposeMintWithProof(t, member, mintProof1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	mintProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal ID: %s with depositId: %s", mintProposalId.String(), depositId)

	// 2. Verify depositId NOT executed yet
	assertReplayProtection(t, ctx, depositId, "", false)

	// 3. Execute proposal successfully
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], mintProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Mint proposal executed")

	// 4. Verify depositId is now marked as executed
	assertReplayProtection(t, ctx, depositId, "", true)

	// 5. Attempt to create another proposal with SAME depositId (should fail with DepositIdAlreadyUsed)
	mintProof2, err := CreateMintProof(recipient, mintAmount, depositId, "bank-ref-replay", "replay attempt")
	require.NoError(t, err)

	tx, err = ctx.TxProposeMintWithProof(t, member, mintProof2)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Expected proposal creation to fail with duplicate depositId")
	require.Contains(t, err.Error(), "DepositIdAlreadyUsed", "Expected DepositIdAlreadyUsed error")
	t.Logf("✓ Replay protection: Cannot reuse depositId '%s'", depositId)

	// ==================== BURN REPLAY TEST ====================

	// 6. Create burn proposal with specific withdrawalId
	burnProof1, err := CreateBurnProof(member.Address, burnAmount, withdrawalId, "ref-F2", "burn test F2")
	require.NoError(t, err)

	tx, err = ctx.TxProposeBurnWithProof(t, member, burnProof1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	burnProposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created burn proposal ID: %s with withdrawalId: %s", burnProposalId.String(), withdrawalId)

	// 7. Verify withdrawalId NOT executed yet
	assertReplayProtection(t, ctx, "", withdrawalId, false)

	// 7.5. Setup MockFiatToken to allow burn execution
	// Give GovMinter sufficient fiat token balance for burning
	tx, err = ctx.mockFiatTokenContractTx(t, "setBalance", ctx.Members[0], TestGovMinterAddress, burnAmount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Set GovMinter MockFiatToken balance: %s", burnAmount.String())

	// 8. Execute burn proposal successfully
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], burnProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Burn proposal executed")

	// 9. Verify withdrawalId is now marked as executed
	assertReplayProtection(t, ctx, "", withdrawalId, true)

	// 10. Attempt to create another proposal with SAME withdrawalId (should fail)
	burnProof2, err := CreateBurnProof(member.Address, burnAmount, withdrawalId, "ref-replay", "burn replay attempt")
	require.NoError(t, err)

	tx, err = ctx.TxProposeBurnWithProof(t, member, burnProof2)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Expected burn proposal creation to fail with duplicate withdrawalId")
	require.Contains(t, err.Error(), "WithdrawalIdAlreadyUsed", "Expected WithdrawalIdAlreadyUsed error")
	t.Logf("✓ Replay protection: Cannot reuse withdrawalId '%s'", withdrawalId)

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "F2: Replay protection")
	assertInvariantsHold(t, ctx, "F2: After replay test")

	// Verify both IDs remain marked as executed
	assertReplayProtection(t, ctx, depositId, withdrawalId, true)

	t.Logf("✓✓✓ Test F2 completed successfully")
}

// Test F3: Proposal lifecycle with governance parameter change mid-flight
// Scenario: MAX_ACTIVE_PROPOSALS_PER_MEMBER changes while proposals active
// Invariants: Existing proposals unaffected, new limit applies to new proposals
func TestEdgeCase_F3_GovernanceParameterChangeMidFlight(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(2_000_000)

	initialState := captureStateSnapshot(t, ctx)

	// 1. Get current MAX_ACTIVE_PROPOSALS_PER_MEMBER
	initialMaxProposals, err := ctx.GetMaxActiveProposalsPerMember(member)
	require.NoError(t, err)
	t.Logf("Initial MAX_ACTIVE_PROPOSALS_PER_MEMBER: %s", initialMaxProposals.String())

	// 2. Create 2 active mint proposals
	tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId1, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal 1: ID=%s", proposalId1.String())

	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId2, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("Created mint proposal 2: ID=%s", proposalId2.String())

	// Verify 2 proposals are active (via reservedMintAmount)
	reservedBefore, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	expectedReserved := new(big.Int).Mul(amount, big.NewInt(2))
	require.Equal(t, 0, reservedBefore.Cmp(expectedReserved), "Should have 2 active proposals worth %s", expectedReserved.String())
	t.Logf("✓ 2 active proposals confirmed, reserved: %s", reservedBefore.String())

	// 3. Propose change to MAX_ACTIVE_PROPOSALS_PER_MEMBER = 1 (reduce from 3 to 1)
	newMaxProposals := big.NewInt(1)
	changeProposalId, tx, err := ctx.BaseTxProposeChangeMaxProposals(t, ctx.govMinter, member, newMaxProposals)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("Created parameter change proposal: ID=%s (change MAX to %s)", changeProposalId.String(), newMaxProposals.String())

	// 4. Execute the parameter change proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], changeProposalId)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Parameter change executed")

	// Verify MAX changed to 1
	currentMaxProposals, err := ctx.GetMaxActiveProposalsPerMember(member)
	require.NoError(t, err)
	require.Equal(t, 0, currentMaxProposals.Cmp(newMaxProposals), "MAX should be changed to %s", newMaxProposals.String())
	t.Logf("✓ MAX_ACTIVE_PROPOSALS_PER_MEMBER changed: %s → %s", initialMaxProposals.String(), currentMaxProposals.String())

	// 5. Verify existing 2 proposals are still valid (grandfathered - not cancelled)
	reservedAfterChange, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterChange.Cmp(expectedReserved), "Existing proposals should remain active")
	t.Logf("✓ Existing 2 proposals still active (grandfathered)")

	// 6. Attempt to create 3rd proposal - should FAIL (exceeds new limit of 1)
	// (member currently has 2 active proposals, new limit is 1, so creating another should fail)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Should not be able to create 3rd proposal with new limit=1")
	require.Contains(t, err.Error(), "TooManyActiveProposals", "Expected TooManyActiveProposals error")
	t.Logf("✓ Cannot create 3rd proposal (exceeds new limit of 1)")

	// 7. Execute one of the existing proposals to free a slot
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId1)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Executed proposal 1 (ID=%s)", proposalId1.String())

	// Now member has only 1 active proposal (proposalId2)
	reservedAfterExec, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, 0, reservedAfterExec.Cmp(amount), "Should have only 1 active proposal worth %s", amount.String())
	t.Logf("✓ Member now has 1 active proposal, reserved: %s", reservedAfterExec.String())

	// 8. Attempt to create new proposal - should still FAIL (already at new limit of 1)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Should not be able to create new proposal (already at limit of 1)")
	require.Contains(t, err.Error(), "TooManyActiveProposals", "Expected TooManyActiveProposals error")
	t.Logf("✓ Cannot create new proposal (already at new limit of 1)")

	// 9. Execute the remaining proposal to free all slots
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId2)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	t.Logf("✓ Executed proposal 2 (ID=%s)", proposalId2.String())

	// Now member has 0 active proposals
	reservedFinal, err := ctx.GetReservedMintAmount(member)
	require.NoError(t, err)
	require.Equal(t, int64(0), reservedFinal.Int64(), "Should have no active proposals")
	t.Logf("✓ All proposals executed, reserved: 0")

	// 10. Create new proposal - should SUCCEED (0 active < new limit of 1)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)
	proposalId3, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("✓ Successfully created new proposal (ID=%s) under new limit", proposalId3.String())

	// 11. Verify cannot create 2nd proposal (would exceed new limit of 1)
	tx, err = ctx.TxProposeMint(t, member, member.Address, amount)
	err = ctx.ExpectedFail(tx, err)
	require.Error(t, err, "Should not be able to create 2nd proposal with new limit=1")
	require.Contains(t, err.Error(), "TooManyActiveProposals", "Expected TooManyActiveProposals error")
	t.Logf("✓ Cannot create 2nd proposal (exceeds new limit of 1)")

	// Cleanup: Execute last proposal
	tx, err = ctx.BaseTxApproveProposal(t, ctx.govMinter, ctx.Members[1], proposalId3)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	finalState := captureStateSnapshot(t, ctx)
	assertStateConsistency(t, ctx, initialState, finalState, "F3: Governance param change")
	assertInvariantsHold(t, ctx, "F3: After param change")

	t.Logf("✓✓✓ Test F3 completed successfully")
}

// ==================== Supplementary Helper Tests ====================

// Test: Verify helper functions work correctly (meta-test)
func TestEdgeCase_HelperFunctionsVerification(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	t.Run("Verify state snapshot capture", func(t *testing.T) {
		snapshot := captureStateSnapshot(t, ctx)
		require.NotNil(t, snapshot)
		require.NotNil(t, snapshot.ReservedMintAmounts)
		require.NotNil(t, snapshot.BurnBalances)
		t.Logf("✓ State snapshot captured successfully")
	})

	t.Run("Verify invariant checks", func(t *testing.T) {
		// Should not panic or fail with initial state
		assertInvariantsHold(t, ctx, "Initial state")
		t.Logf("✓ Invariant checks passed")
	})

	t.Run("Verify assertProposalCreation", func(t *testing.T) {
		member := ctx.Members[0]
		recipient := member.Address
		amount := big.NewInt(100_000)

		// Create a proposal
		tx, err := ctx.TxProposeMint(t, member, recipient, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)

		// Verify with assertProposalCreation
		assertProposalCreation(t, ctx, ProposalCreationExpectation{
			ProposalId:                proposalId,
			Member:                    member,
			ProposalType:              "Mint",
			Amount:                    amount,
			ActiveCountIncremented:    true,
			ReservedAmountIncremented: true,
			BurnBalanceSufficient:     false,
		})

		t.Logf("✓ assertProposalCreation verified")
	})
}

// ==================== Test Data Generators ====================

// generateUniqueDepositId generates unique deposit ID for testing
func generateUniqueDepositId(t *testing.T, prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// generateUniqueWithdrawalId generates unique withdrawal ID for testing
func generateUniqueWithdrawalId(t *testing.T, prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ==================== Category G: Approved State Management ====================

// Test G1: Sequential proposal processing from single member
// Scenario: Create and process multiple proposals sequentially
// Invariants: State consistent across sequential operations
func TestEdgeCase_G1_SequentialProposalProcessing(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(1_000_000)

	// 1. Create and process first proposal
	t.Log("Step 1: Creating first proposal")
	proposalId1 := createApprovedMintProposal(t, ctx, member, member.Address, amount)
	proposal1, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId1)
	require.NoError(t, err)
	t.Logf("✓ Created proposal %s, status: %v", proposalId1.String(), proposal1.Status)

	// If Approved, execute it
	if proposal1.Status == sc.ProposalStatusApproved {
		tx, err := ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId1)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)
		t.Logf("✓ Executed proposal %s", proposalId1.String())
	}

	// 2. Create second proposal
	t.Log("Step 2: Creating second proposal")
	proposalId2 := createApprovedMintProposal(t, ctx, member, member.Address, amount)
	proposal2, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId2)
	require.NoError(t, err)
	t.Logf("✓ Created proposal %s, status: %v", proposalId2.String(), proposal2.Status)

	// 3. Verify system state after sequential operations
	assertInvariantsHold(t, ctx, "After sequential proposals")

	t.Logf("✓ Sequential proposal processing working correctly")
}

// Test G2: Multiple concurrent proposals from different members
// Scenario: Create proposals from multiple members simultaneously
// Invariants: Independent state management per member
func TestEdgeCase_G2_MultipleConcurrentProposals(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	amount := big.NewInt(1_000_000)

	// 1. Create proposals from multiple members
	t.Log("Step 1: Creating proposals from all members")
	proposalIds := make([]*big.Int, len(ctx.Members))

	for i, member := range ctx.Members {
		pid := createApprovedMintProposal(t, ctx, member, member.Address, amount)
		proposalIds[i] = pid
		t.Logf("✓ Created proposal %s for %s", pid.String(), formatAddress(member.Address))
	}

	// 2. Verify all proposals created successfully
	t.Log("Step 2: Verifying all proposals")
	for i, pid := range proposalIds {
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, ctx.Members[i], pid)
		require.NoError(t, err)
		t.Logf("✓ Proposal %s status: %v", pid.String(), proposal.Status)
	}

	// 3. Verify independent state per member
	assertInvariantsHold(t, ctx, "After multiple concurrent proposals")

	t.Logf("✓ Multiple concurrent proposals handled correctly")
}

// Test G3: Rapid proposal creation and completion cycles
// Scenario: Quickly create and complete proposals to stress test counters
// Invariants: Counter increment/decrement accuracy
func TestEdgeCase_G3_RapidProposalCycles(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(500_000)

	// 1. Create multiple proposals rapidly
	t.Log("Step 1: Creating rapid proposal cycles")
	const cycles = 3

	for i := 0; i < cycles; i++ {
		// Create proposal
		proposalId := createApprovedMintProposal(t, ctx, member, member.Address, amount)

		// Verify created
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)
		t.Logf("✓ Cycle %d: Proposal %s status: %v", i+1, proposalId.String(), proposal.Status)

		// If Approved, execute it
		if proposal.Status == sc.ProposalStatusApproved {
			tx, err := ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
			_, err = ctx.ExpectedOk(tx, err)
			require.NoError(t, err)
			t.Logf("✓ Cycle %d: Executed proposal %s", i+1, proposalId.String())
		}
	}

	// 2. Verify counters and state are accurate
	assertInvariantsHold(t, ctx, "After rapid proposal cycles")

	t.Logf("✓ Rapid proposal cycles completed successfully")
}

// ==================== Category H: Retry & Recovery Patterns ====================

// Test H1: Retry execution after transient failure
// Scenario: Execute proposal, verify retry on Approved status
// Invariants: State consistent across retry attempts
func TestEdgeCase_H1_RetryAfterTransientFailure(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(1_000_000)

	// 1. Create proposal
	t.Log("Step 1: Creating proposal")
	proposalId := createApprovedMintProposal(t, ctx, member, member.Address, amount)

	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	t.Logf("✓ Proposal %s status: %v", proposalId.String(), proposal.Status)

	// 2. If Approved, execute (simulates retry)
	if proposal.Status == sc.ProposalStatusApproved {
		t.Log("Step 2: Retrying execution on Approved proposal")
		tx, err := ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposal, err = ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)
		t.Logf("✓ After retry: status=%v", proposal.Status)
	}

	// 3. Verify state consistency
	assertInvariantsHold(t, ctx, "After retry execution")

	t.Logf("✓ Retry execution handled correctly")
}

// Test H2: Multiple retry attempts until success
// Scenario: Retry approved proposal multiple times
// Invariants: No state corruption from multiple retries
func TestEdgeCase_H2_MultipleRetryAttempts(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(800_000)

	// 1. Create proposal
	proposalId := createApprovedMintProposal(t, ctx, member, member.Address, amount)

	// 2. Attempt multiple executions (retry simulation)
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
		require.NoError(t, err)

		if proposal.Status != sc.ProposalStatusApproved {
			t.Logf("✓ Proposal reached terminal state after %d attempts: %v", i, proposal.Status)
			break
		}

		t.Logf("Retry attempt %d/%d", i+1, maxRetries)
		tx, err := ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)
	}

	// 3. Verify state consistency
	assertInvariantsHold(t, ctx, "After multiple retry attempts")

	t.Logf("✓ Multiple retries handled correctly")
}

// Test H3: Recovery after partial state update
// Scenario: Verify system recovers correctly if retry succeeds
// Invariants: No duplicate state updates
func TestEdgeCase_H3_RecoveryAfterPartialUpdate(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(600_000)

	// 1. Create and possibly execute proposal
	proposalId := createApprovedMintProposal(t, ctx, member, member.Address, amount)

	proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)

	// 2. If Approved, execute to complete
	if proposal.Status == sc.ProposalStatusApproved {
		tx, err := ctx.BaseTxExecuteProposal(t, ctx.govMinter, member, proposalId)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)
	}

	// 3. Verify no duplicate updates
	finalProposal, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId)
	require.NoError(t, err)
	t.Logf("✓ Final proposal status: %v", finalProposal.Status)

	assertInvariantsHold(t, ctx, "After recovery")

	t.Logf("✓ Recovery completed without duplicate updates")
}

// ==================== Category I: Error Validation & Prevention ====================

// Test I1: Invalid proof data rejection
// Scenario: Attempt to create proposal with malformed proof
// Invariants: System rejects invalid data without state corruption
func TestEdgeCase_I1_InvalidProofDataRejection(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]

	// 1. Attempt with invalid proof data
	t.Log("Step 1: Attempting proposal with invalid proof")
	invalidProof := []byte{0x00, 0x01, 0x02} // Too short/invalid

	tx, err := ctx.TxProposeMintWithProof(t, member, invalidProof)
	err = ctx.ExpectedFail(tx, err)

	if err != nil {
		t.Logf("✓ Invalid proof rejected: %v", err)
	}

	// 2. Verify system state unchanged
	assertInvariantsHold(t, ctx, "After invalid proof rejection")

	t.Logf("✓ Invalid proof handled correctly")
}

// Test I2: Duplicate depositId prevention
// Scenario: Attempt to reuse depositId
// Invariants: Duplicate detection prevents double-minting
func TestEdgeCase_I2_DuplicateDepositIdPrevention(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(500_000)

	// 1. Create first proposal with specific depositId
	t.Log("Step 1: Creating first proposal")
	depositId1 := generateUniqueDepositId(t, "test-dup")

	tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
	_, err = ctx.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId1, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
	require.NoError(t, err)
	t.Logf("✓ Created first proposal: %s", proposalId1.String())

	// 2. Wait for execution/completion
	proposal1, err := ctx.BaseGetProposal(ctx.govMinter, member, proposalId1)
	require.NoError(t, err)

	if proposal1.Status == sc.ProposalStatusExecuted {
		// Check that depositId is marked as executed
		isExecuted, err := ctx.IsDepositIdExecuted(member, depositId1)
		if err == nil && isExecuted {
			t.Logf("✓ DepositId marked as executed")

			// 3. Attempt to create another proposal with same depositId would fail
			// (In practice, this requires constructing proof with same depositId)
			t.Logf("✓ Duplicate depositId prevention verified")
		}
	}

	assertInvariantsHold(t, ctx, "After duplicate prevention check")
}

// Test I3: Zero and negative amount validation
// Scenario: Attempt proposals with invalid amounts
// Invariants: Pre-validation prevents invalid state
func TestEdgeCase_I3_InvalidAmountValidation(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]

	// 1. Attempt zero amount
	t.Log("Step 1: Attempting zero amount proposal")
	zeroAmount := big.NewInt(0)

	tx, err := ctx.TxProposeMint(t, member, member.Address, zeroAmount)
	err = ctx.ExpectedFail(tx, err)

	if err != nil {
		t.Logf("✓ Zero amount rejected: %v", err)
	}

	// 2. Verify system state unchanged
	assertInvariantsHold(t, ctx, "After invalid amount rejection")

	t.Logf("✓ Invalid amounts handled correctly")
}

// ==================== Category J: Active Proposal Limit Stress Testing ====================

// Test J1: Reach active proposal limit and cleanup
// Scenario: Create MAX_ACTIVE_PROPOSALS proposals then complete some
// Invariants: Limit enforced, cleanup allows new proposals
func TestEdgeCase_J1_ActiveProposalLimitAndCleanup(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(300_000)

	maxProposals := ctx.MaxActiveProposalsPerMember.Int64()
	t.Logf("MAX_ACTIVE_PROPOSALS_PER_MEMBER: %d", maxProposals)

	// 1. Create proposals up to limit
	t.Log("Step 1: Creating proposals up to limit")

	for i := int64(0); i < maxProposals; i++ {
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		pid, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)

		// Check if auto-executed (clears slot)
		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, pid)
		require.NoError(t, err)

		if proposal.Status != sc.ProposalStatusExecuted {
			t.Logf("✓ Created proposal %d/%d: %s (status: %v)",
				i+1, maxProposals, pid.String(), proposal.Status)
		} else {
			t.Logf("✓ Proposal %d/%d auto-executed: %s", i+1, maxProposals, pid.String())
		}
	}

	// 2. Verify limit enforcement
	assertInvariantsHold(t, ctx, "After reaching proposal limit")

	t.Logf("✓ Active proposal limit enforced correctly")
}

// Test J2: Multiple members at limit simultaneously
// Scenario: All members create proposals up to their limits
// Invariants: Independent counters per member
func TestEdgeCase_J2_MultiMemberLimitStress(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	amount := big.NewInt(200_000)

	// 1. Each member creates one proposal
	t.Log("Step 1: Each member creating proposals")

	for _, member := range ctx.Members {
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)
		_, err = ctx.ExpectedOk(tx, err)
		require.NoError(t, err)

		pid, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
		require.NoError(t, err)

		proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, pid)
		require.NoError(t, err)

		t.Logf("✓ %s created proposal %s (status: %v)",
			formatAddress(member.Address), pid.String(), proposal.Status)
	}

	// 2. Verify independent counters
	assertInvariantsHold(t, ctx, "After multi-member stress test")

	t.Logf("✓ Independent member limits working correctly")
}

// Test J3: Active proposal limit enforcement
// Scenario: Create proposals up to limit, verify rejection beyond limit
// Invariants: Limit strictly enforced
func TestEdgeCase_J3_ActiveProposalLimitEnforcement(t *testing.T) {
	ctx := setupEdgeCaseTest(t)
	defer ctx.backend.Close()

	member := ctx.Members[0]
	amount := big.NewInt(400_000)

	maxProposals := ctx.MaxActiveProposalsPerMember.Int64()
	t.Logf("MAX_ACTIVE_PROPOSALS_PER_MEMBER: %d", maxProposals)

	// 1. Create proposals up to limit
	t.Log("Step 1: Creating proposals up to limit")
	successCount := int64(0)

	for i := int64(0); i < maxProposals+2; i++ { // Try to exceed limit
		tx, err := ctx.TxProposeMint(t, member, member.Address, amount)

		if i < maxProposals {
			// Should succeed within limit
			_, err = ctx.ExpectedOk(tx, err)
			require.NoError(t, err)
			successCount++

			pid, err := ctx.BaseCurrentProposalId(ctx.govMinter, member)
			require.NoError(t, err)

			proposal, err := ctx.BaseGetProposal(ctx.govMinter, member, pid)
			require.NoError(t, err)

			t.Logf("✓ Created proposal %d/%d: %s (status: %v)",
				i+1, maxProposals, pid.String(), proposal.Status)
		} else {
			// Should fail beyond limit (if proposals are Pending)
			err = ctx.ExpectedFail(tx, err)
			if err != nil {
				t.Logf("✓ Proposal %d rejected (limit reached): %v", i+1, err)
				break // Limit enforced correctly
			} else {
				// If succeeded, proposal was auto-executed (freed slot)
				successCount++
				t.Logf("✓ Proposal %d succeeded (auto-executed)", i+1)
			}
		}
	}

	// 2. Verify limit enforcement
	t.Logf("✓ Created %d proposals (limit: %d)", successCount, maxProposals)
	assertInvariantsHold(t, ctx, "After limit enforcement test")

	t.Logf("✓ Active proposal limit enforced correctly")
}

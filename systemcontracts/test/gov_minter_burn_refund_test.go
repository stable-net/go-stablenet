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
	"time"

	"github.com/stretchr/testify/require"
)

// Test Suite: [Burn-2] burnBalance Refund on Non-Executed Proposal Termination
//
// Fix: Hybrid approach — _cleanupBurnDeposit moves burnBalance → refundableBalance
// on non-executed finalization (no external calls in hook), then claimBurnRefund()
// allows Pull withdrawal of native coins.

// Test 1: Cancelled burn proposal should refund burnBalance
func TestBurnRefund_CancelledProposalShouldRefundBurnBalance(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member := minterMembers[0].Operator
	amount := big.NewInt(1_000_000)

	// Record initial burnBalance (should be 0)
	initialBurnBalance, err := gMinter.GetBurnBalance(minterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), initialBurnBalance.Int64(), "Initial burnBalance should be 0")

	// Step 1: proposeBurn deposits native coins and credits burnBalance
	tx, err := gMinter.TxProposeBurn(t, member, member.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify burnBalance was credited
	burnBalanceAfterPropose, err := gMinter.GetBurnBalance(minterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalanceAfterPropose.Cmp(amount),
		"burnBalance should be %s after proposeBurn, got %s", amount.String(), burnBalanceAfterPropose.String())

	// Get proposal ID
	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member)
	require.NoError(t, err)

	// Step 2: Cancel the proposal
	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Cancelled (status 4)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(4), uint8(proposal.Status), "Proposal should be Cancelled")

	// burnBalance should be 0 (moved to refundableBalance)
	burnBalanceAfterCancel, err := gMinter.GetBurnBalance(minterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalanceAfterCancel.Int64(),
		"[Burn-2] burnBalance should be 0 after cancellation, but got %s",
		burnBalanceAfterCancel.String())

	// refundableBalance should have the amount
	refundable, err := gMinter.GetRefundableBalance(minterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(amount),
		"refundableBalance should be %s after cancellation, got %s", amount.String(), refundable.String())

	// Step 3: Claim refund via Pull withdrawal
	govBalBefore := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)

	tx, err = gMinter.TxClaimBurnRefund(t, member)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// refundableBalance should be 0 after claim
	refundableAfterClaim, err := gMinter.GetRefundableBalance(minterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundableAfterClaim.Int64(), "refundableBalance should be 0 after claim")

	// GovMinter native balance should decrease by the refund amount
	govBalAfter := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	govBalDiff := new(big.Int).Sub(govBalBefore, govBalAfter)
	require.Equal(t, 0, govBalDiff.Cmp(amount),
		"GovMinter balance should decrease by refund amount, diff: %s", govBalDiff.String())
}

// Test 2: Rejected burn proposal should refund burnBalance
func TestBurnRefund_RejectedProposalShouldRefundBurnBalance(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	member2 := minterMembers[2].Operator
	amount := big.NewInt(2_000_000)

	// Step 1: proposeBurn
	tx, err := gMinter.TxProposeBurn(t, member0, member0.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Verify burnBalance credited
	burnBalance, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalance.Cmp(amount), "burnBalance should match deposited amount")

	// Step 2: Reject proposal (quorum=2, two rejections)
	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = gMinter.BaseTxDisapproveProposal(t, gMinter.govMinter, member2, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Rejected (status 7)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(7), uint8(proposal.Status), "Proposal should be Rejected")

	// burnBalance should be 0
	burnBalanceAfterReject, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalanceAfterReject.Int64(),
		"[Burn-2] burnBalance should be 0 after rejection, but got %s",
		burnBalanceAfterReject.String())

	// refundableBalance should have the amount
	refundable, err := gMinter.GetRefundableBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(amount),
		"refundableBalance should be %s after rejection, got %s", amount.String(), refundable.String())
}

// Test 3: Expired burn proposal should refund burnBalance
func TestBurnRefund_ExpiredProposalShouldRefundBurnBalance(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	amount := big.NewInt(3_000_000)

	// Step 1: proposeBurn
	tx, err := gMinter.TxProposeBurn(t, member0, member0.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Verify burnBalance credited
	burnBalance, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalance.Cmp(amount), "burnBalance should match deposited amount")

	// Step 2: Advance time past proposal expiry (7 days + 1 second)
	gMinter.backend.AdjustTime(7*24*time.Hour + time.Second)
	gMinter.backend.Commit()

	// Expire the proposal
	tx, err = gMinter.BaseTxExpireProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// burnBalance should be 0
	burnBalanceAfterExpiry, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalanceAfterExpiry.Int64(),
		"[Burn-2] burnBalance should be 0 after expiry, but got %s",
		burnBalanceAfterExpiry.String())

	// refundableBalance should have the amount
	refundable, err := gMinter.GetRefundableBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(amount),
		"refundableBalance should be %s after expiry, got %s", amount.String(), refundable.String())
}

// Test 4: Full reproduction scenario from audit report
// Demonstrates cumulative asset locking across cancel + re-propose
func TestBurnRefund_AuditScenario_CancelAndRepropose(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	firstAmount := big.NewInt(50)
	secondAmount := big.NewInt(100)

	// Step 1: proposeBurn{value: 50}
	tx, err := gMinter.TxProposeBurn(t, member0, member0.Address, firstAmount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId1, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	burnBalance, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(50), burnBalance.Int64(), "burnBalance should be 50 after first proposal")

	// Record GovMinter native balance after first deposit
	govMinterBalanceAfterFirst := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	t.Logf("GovMinter native balance after first deposit: %s", govMinterBalanceAfterFirst.String())

	// Step 2: cancelProposal — burnBalance moves to refundableBalance
	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member0, proposalId1)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	burnBalanceAfterCancel, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)

	// Step 3: proposeBurn{value: 100} — new proposal
	tx, err = gMinter.TxProposeBurn(t, member0, member0.Address, secondAmount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId2, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	burnBalanceAfterRepropose, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)

	// After cancel, burnBalance should be 0 (moved to refundable)
	require.Equal(t, int64(0), burnBalanceAfterCancel.Int64(),
		"[Burn-2] After cancel, burnBalance should be 0 but got %s — first deposit is locked",
		burnBalanceAfterCancel.String())

	// After re-propose, burnBalance should be only 100 (second deposit only)
	require.Equal(t, secondAmount.Int64(), burnBalanceAfterRepropose.Int64(),
		"[Burn-2] After re-propose, burnBalance should be %d (only second deposit), but got %s — cancelled deposit accumulated",
		secondAmount.Int64(), burnBalanceAfterRepropose.String())

	// Step 4: Execute second proposal (give GovMinter fiat token balance first)
	tx, err = gMinter.mockFiatTokenContractTx(t, "setBalance", member0, TestGovMinterAddress, secondAmount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId2)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// After successful execution, burnBalance should be 0
	finalBurnBalance, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalBurnBalance.Int64(),
		"[Burn-2] After execution, burnBalance should be 0 but got %s — residual from cancelled proposal",
		finalBurnBalance.String())
}

// Test 5: GovMinter native balance should decrease after burn refund claim
// Verifies that actual native coins are returned, not just the accounting
func TestBurnRefund_NativeCoinsReturnedOnCancel(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member := minterMembers[0].Operator
	amount := big.NewInt(1_000_000)

	// Record initial GovMinter native balance
	initialGovBalance := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)

	// proposeBurn deposits native coins
	tx, err := gMinter.TxProposeBurn(t, member, member.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member)
	require.NoError(t, err)

	// GovMinter should have received the native coins
	govBalanceAfterDeposit := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	depositIncrease := new(big.Int).Sub(govBalanceAfterDeposit, initialGovBalance)
	require.Equal(t, 0, depositIncrease.Cmp(amount),
		"GovMinter native balance should increase by deposit amount")

	// Cancel proposal
	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// After cancel, coins are still in GovMinter (refundableBalance, not yet claimed)
	govBalanceAfterCancel := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	require.Equal(t, 0, govBalanceAfterCancel.Cmp(govBalanceAfterDeposit),
		"GovMinter balance should remain same after cancel (Pull pattern, not yet claimed)")

	// Claim refund
	tx, err = gMinter.TxClaimBurnRefund(t, member)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// GovMinter native balance should return to initial level after claim
	govBalanceAfterClaim := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	remainingIncrease := new(big.Int).Sub(govBalanceAfterClaim, initialGovBalance)

	require.Equal(t, int64(0), remainingIncrease.Int64(),
		"[Burn-2] GovMinter native balance should return to initial after claim, but %s remains locked",
		remainingIncrease.String())
}

// Test 6: Executed proposal should NOT credit refundableBalance
func TestBurnRefund_ExecutedProposalNoRefund(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member0 := minterMembers[0].Operator
	member1 := minterMembers[1].Operator
	amount := big.NewInt(500_000)

	// proposeBurn
	tx, err := gMinter.TxProposeBurn(t, member0, member0.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member0)
	require.NoError(t, err)

	// Give GovMinter fiat token balance for burn execution
	tx, err = gMinter.mockFiatTokenContractTx(t, "setBalance", member0, TestGovMinterAddress, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Approve → Execute
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, member1, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Executed (status 3)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")

	// burnBalance should be 0 (decremented by _safeBurn on success)
	burnBal, err := gMinter.GetBurnBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBal.Int64(), "burnBalance should be 0 after successful execution")

	// refundableBalance should also be 0 (no refund for executed proposals)
	refundable, err := gMinter.GetRefundableBalance(minterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundable.Int64(),
		"refundableBalance should be 0 for executed proposal, got %s", refundable.String())
}

// Test 7: Double claim should revert with NoRefundAvailable
func TestBurnRefund_DoubleClaimReverts(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	member := minterMembers[0].Operator
	amount := big.NewInt(1_000_000)

	// proposeBurn → cancel
	tx, err := gMinter.TxProposeBurn(t, member, member.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, member)
	require.NoError(t, err)

	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, member, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// First claim should succeed
	tx, err = gMinter.TxClaimBurnRefund(t, member)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Second claim should revert (NoRefundAvailable)
	tx, err = gMinter.TxClaimBurnRefund(t, member)
	ExpectedRevert(t, gMinter.ExpectedFail(tx, err), "NoRefundAvailable")
}

// Test 8: Other accounts cannot claim someone else's refund
func TestBurnRefund_OtherAccountCannotClaim(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	memberA := minterMembers[0].Operator
	memberB := minterMembers[1].Operator
	amount := big.NewInt(1_000_000)

	// Member A proposes burn → cancel → refundableBalance credited to A
	tx, err := gMinter.TxProposeBurn(t, memberA, memberA.Address, amount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, memberA)
	require.NoError(t, err)

	tx, err = gMinter.BaseTxCancelProposal(t, gMinter.govMinter, memberA, proposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify Member A has refundable balance
	refundable, err := gMinter.GetRefundableBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(amount), "Member A should have refundableBalance")

	// Member B tries to claim — should revert (B has no refundableBalance)
	tx, err = gMinter.TxClaimBurnRefund(t, memberB)
	ExpectedRevert(t, gMinter.ExpectedFail(tx, err), "NoRefundAvailable")

	// Non-member tries to claim — should also revert
	tx, err = gMinter.TxClaimBurnRefund(t, minterNonMember)
	ExpectedRevert(t, gMinter.ExpectedFail(tx, err), "NoRefundAvailable")

	// Member A's refundable balance is still intact
	refundableAfter, err := gMinter.GetRefundableBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundableAfter.Cmp(amount),
		"Member A's refundableBalance should be unchanged after others' failed claims")
}

// Test 9: Removed member should still be able to claim refund for expired burn proposal
// Scenario: Member A proposes burn → Member A is removed from governance → proposal expires → Member A claims refund
func TestBurnRefund_RemovedMemberCanClaimAfterExpiry(t *testing.T) {
	initGovMinterV2(t)
	defer gMinter.backend.Close()

	memberA := minterMembers[0].Operator
	memberB := minterMembers[1].Operator
	memberC := minterMembers[2].Operator
	burnAmount := big.NewInt(5_000_000)

	// Step 1: Member A proposes burn
	tx, err := gMinter.TxProposeBurn(t, memberA, memberA.Address, burnAmount)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	burnProposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, memberA)
	require.NoError(t, err)
	t.Logf("Burn proposal ID: %s", burnProposalId.String())

	// Verify burnBalance credited
	burnBal, err := gMinter.GetBurnBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBal.Cmp(burnAmount), "burnBalance should be %s", burnAmount.String())

	// Step 2: Member B proposes to remove Member A (quorum stays 2)
	// proposeRemoveMember(address member, uint32 newQuorum)
	tx, err = gMinter.govMinter.Transact(NewTxOptsWithValue(t, memberB, nil), "proposeRemoveMember", memberA.Address, uint32(2))
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	removeProposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, memberB)
	require.NoError(t, err)
	t.Logf("Remove member proposal ID: %s", removeProposalId.String())

	// Step 3: Member C approves removal → Member A is removed (quorum reached)
	tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, memberC, removeProposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify Member A is removed
	memberAInfo, err := gMinter.BaseMembers(gMinter.govMinter, memberB, memberA.Address)
	require.NoError(t, err)
	require.False(t, memberAInfo.IsActive, "Member A should be removed from governance")
	t.Log("Member A removed from governance")

	// Step 4: Advance time past proposal expiry (7 days + 1 second)
	gMinter.backend.AdjustTime(7*24*time.Hour + time.Second)
	gMinter.backend.Commit()

	// Step 5: Member B expires the burn proposal
	// (Member A can't call expireProposal because it requires onlyProposalMember,
	//  but the burn proposal was created at version 1, so Member A could still call it.
	//  However, let's use Member B to expire it — the important part is claim, not expire.)
	tx, err = gMinter.BaseTxExpireProposal(t, gMinter.govMinter, memberB, burnProposalId)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal is Expired
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, memberB, burnProposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(5), uint8(proposal.Status), "Burn proposal should be Expired")
	t.Log("Burn proposal expired")

	// Step 6: Verify burnBalance moved to refundableBalance
	burnBalAfter, err := gMinter.GetBurnBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalAfter.Int64(), "burnBalance should be 0 after expiry")

	refundable, err := gMinter.GetRefundableBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(burnAmount),
		"refundableBalance should be %s for removed member", burnAmount.String())

	// Step 7: Member A (now removed) claims refund — this is the critical assertion
	// claimBurnRefund() has no access control, so removed members can call it
	govBalBefore := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)

	tx, err = gMinter.TxClaimBurnRefund(t, memberA)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err, "Removed member should be able to claim burn refund")
	t.Log("Removed member successfully claimed burn refund")

	// Verify refundableBalance is now 0
	refundableAfter, err := gMinter.GetRefundableBalance(minterNonMember, memberA.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundableAfter.Int64(), "refundableBalance should be 0 after claim")

	// Verify GovMinter native balance decreased by refund amount
	govBalAfter := gMinter.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	govBalDiff := new(big.Int).Sub(govBalBefore, govBalAfter)
	require.Equal(t, 0, govBalDiff.Cmp(burnAmount),
		"GovMinter balance should decrease by %s after claim, diff: %s", burnAmount.String(), govBalDiff.String())
}

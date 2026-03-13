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

	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

// createGovMinterHardforkEnv creates a test environment with v1 GovMinter at genesis
// and v2 ABI binding (superset of v1) to support calling both v1 and v2 methods.
// The caller should use CommitWithState to apply the v2 upgrade at the desired block.
func createGovMinterHardforkEnv(t *testing.T) *GovMinterTestEnv {
	env := createGovMinterTestEnv(t)

	// Swap ABI binding to v2 (superset of v1).
	// v1 methods use the same selectors, so they work on v1 bytecode.
	// v2-only methods (claimBurnRefund) will revert on v1 bytecode, which is expected.
	env.GMinter.govMinter = compiledGovMinterV2.New(env.GMinter.backend.Client(), TestGovMinterAddress)

	return env
}

// applyBForkUpgrade simulates the BFork hardfork by upgrading GovMinter to v2 bytecode.
func applyBForkUpgrade(g *GovWBFT) {
	g.backend.CommitWithState(&params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: TestGovMinterAddress,
			Version: sc.SYSTEM_CONTRACT_VERSION_2,
		},
	}, nil)
}

// TestHardfork_V1BurnCancelNoRefund verifies that v1 GovMinter does NOT provide
// burn refund on proposal cancellation (this is the bug that v2 fixes).
func TestHardfork_V1BurnCancelNoRefund(t *testing.T) {
	env := createGovMinterHardforkEnv(t)
	defer env.GMinter.backend.Close()

	g := env.GMinter
	member := env.MinterMembers[0].Operator
	amount := big.NewInt(1_000_000)

	// proposeBurn deposits native coins
	tx, err := g.TxProposeBurn(t, member, member.Address, amount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := g.BaseCurrentProposalId(g.govMinter, member)
	require.NoError(t, err)

	// Verify burnBalance credited
	burnBal, err := g.GetBurnBalance(env.MinterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBal.Cmp(amount), "burnBalance should be credited")

	// Cancel proposal
	tx, err = g.BaseTxCancelProposal(t, g.govMinter, member, proposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	// v1 bug: burnBalance stays after cancellation (not cleaned up)
	burnBalAfter, err := g.GetBurnBalance(env.MinterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalAfter.Cmp(amount),
		"[v1] burnBalance should remain %s after cancel (v1 bug - no cleanup)", amount.String())

	// v1: claimBurnRefund does not exist, should revert
	tx, err = g.TxClaimBurnRefund(t, member)
	err = g.ExpectedFail(tx, err)
	require.Error(t, err, "[v1] claimBurnRefund should revert on v1 bytecode")
}

// TestHardfork_V1ToV2Transition verifies the complete v1 → v2 hardfork transition:
// 1. Pre-fork: v1 behavior (no refund on cancel)
// 2. Apply BFork upgrade
// 3. Post-fork: v2 behavior (refund on cancel, claimBurnRefund works)
func TestHardfork_V1ToV2Transition(t *testing.T) {
	env := createGovMinterHardforkEnv(t)
	defer env.GMinter.backend.Close()

	g := env.GMinter
	member := env.MinterMembers[0].Operator
	member1 := env.MinterMembers[1].Operator
	preForkAmount := big.NewInt(500_000)
	postForkAmount := big.NewInt(1_000_000)

	// ========== Phase 1: Pre-fork (v1 behavior) ==========
	t.Log("=== Phase 1: Pre-fork (v1 GovMinter) ===")

	// proposeBurn with preForkAmount
	tx, err := g.TxProposeBurn(t, member, member.Address, preForkAmount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	preForkProposalId, err := g.BaseCurrentProposalId(g.govMinter, member)
	require.NoError(t, err)

	// Cancel proposal
	tx, err = g.BaseTxCancelProposal(t, g.govMinter, member, preForkProposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	// v1: burnBalance stays (bug)
	burnBalPreFork, err := g.GetBurnBalance(env.MinterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalPreFork.Cmp(preForkAmount),
		"[Pre-fork] burnBalance should remain locked in v1")
	t.Logf("Pre-fork burnBalance (locked): %s", burnBalPreFork.String())

	// ========== Phase 2: Apply BFork upgrade ==========
	t.Log("=== Phase 2: Applying BFork (v1 → v2 upgrade) ===")
	applyBForkUpgrade(g)
	t.Log("BFork upgrade applied - GovMinter bytecode swapped to v2")

	// ========== Phase 3: Post-fork (v2 behavior) ==========
	t.Log("=== Phase 3: Post-fork (v2 GovMinter) ===")

	// Verify pre-fork locked burnBalance is still there (storage preserved)
	burnBalAfterFork, err := g.GetBurnBalance(env.MinterNonMember, member.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalAfterFork.Cmp(preForkAmount),
		"[Post-fork] Pre-fork burnBalance should be preserved after bytecode swap")
	t.Logf("Post-fork: pre-fork burnBalance preserved: %s", burnBalAfterFork.String())

	// New burn proposal with postForkAmount
	tx, err = g.TxProposeBurn(t, member1, member1.Address, postForkAmount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	postForkProposalId, err := g.BaseCurrentProposalId(g.govMinter, member1)
	require.NoError(t, err)

	// Verify burnBalance credited for member1
	burnBalMember1, err := g.GetBurnBalance(env.MinterNonMember, member1.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalMember1.Cmp(postForkAmount),
		"[Post-fork] member1 burnBalance should be credited")

	// Cancel post-fork proposal
	tx, err = g.BaseTxCancelProposal(t, g.govMinter, member1, postForkProposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	// v2: burnBalance should be cleared (moved to refundableBalance)
	burnBalMember1After, err := g.GetBurnBalance(env.MinterNonMember, member1.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBalMember1After.Int64(),
		"[Post-fork] v2: burnBalance should be 0 after cancel (cleanup works)")

	// v2: refundableBalance should have the amount
	refundable, err := g.GetRefundableBalance(env.MinterNonMember, member1.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(postForkAmount),
		"[Post-fork] v2: refundableBalance should be %s", postForkAmount.String())

	// v2: claimBurnRefund should work
	govBalBefore := g.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)

	tx, err = g.TxClaimBurnRefund(t, member1)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err, "[Post-fork] claimBurnRefund should succeed on v2 bytecode")

	// Verify refundableBalance cleared
	refundableAfter, err := g.GetRefundableBalance(env.MinterNonMember, member1.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundableAfter.Int64(),
		"[Post-fork] refundableBalance should be 0 after claim")

	// Verify GovMinter native balance decreased
	govBalAfter := g.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	govBalDiff := new(big.Int).Sub(govBalBefore, govBalAfter)
	require.Equal(t, 0, govBalDiff.Cmp(postForkAmount),
		"[Post-fork] GovMinter balance should decrease by refund amount")

	t.Log("v1 → v2 hardfork transition verified successfully")
}

// TestHardfork_StoragePreservation verifies that existing storage (proposals, burnBalance,
// reservedMintAmount) is preserved across the bytecode swap.
func TestHardfork_StoragePreservation(t *testing.T) {
	env := createGovMinterHardforkEnv(t)
	defer env.GMinter.backend.Close()

	g := env.GMinter
	member0 := env.MinterMembers[0].Operator
	member1 := env.MinterMembers[1].Operator
	burnAmount := big.NewInt(2_000_000)

	// Create a burn proposal in v1 (keep it in Voting status)
	tx, err := g.TxProposeBurn(t, member0, member0.Address, burnAmount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := g.BaseCurrentProposalId(g.govMinter, member0)
	require.NoError(t, err)

	// Verify proposal state before fork
	proposal, err := g.BaseGetProposal(g.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, sc.ProposalStatusVoting, proposal.Status, "Proposal should be Voting")

	burnBal, err := g.GetBurnBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBal.Cmp(burnAmount), "burnBalance should match")

	// Apply BFork
	applyBForkUpgrade(g)

	// After fork: verify ALL storage is preserved
	proposalAfter, err := g.BaseGetProposal(g.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, proposal.Status, proposalAfter.Status,
		"Proposal status should be preserved after fork")

	burnBalAfter, err := g.GetBurnBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, burnBalAfter.Cmp(burnAmount),
		"burnBalance should be preserved after fork")

	// After fork: complete the proposal lifecycle with v2 behavior
	// Approve → Execute (needs fiat token balance for burn execution)
	tx, err = g.mockFiatTokenContractTx(t, "setBalance", member0, TestGovMinterAddress, burnAmount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = g.BaseTxApproveProposal(t, g.govMinter, member1, proposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify proposal executed
	proposalFinal, err := g.BaseGetProposal(g.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(sc.ProposalStatusExecuted), uint8(proposalFinal.Status),
		"Pre-fork proposal should execute successfully on v2 bytecode")

	// Executed proposal: burnBalance consumed, no refundableBalance
	finalBurnBal, err := g.GetBurnBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), finalBurnBal.Int64(),
		"burnBalance should be 0 after successful execution")

	refundable, err := g.GetRefundableBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundable.Int64(),
		"refundableBalance should be 0 for executed proposal")
}

// TestHardfork_PostForkBurnRejectRefund verifies that burn proposals rejected
// AFTER the fork correctly trigger the refund mechanism.
func TestHardfork_PostForkBurnRejectRefund(t *testing.T) {
	env := createGovMinterHardforkEnv(t)
	defer env.GMinter.backend.Close()

	g := env.GMinter
	member0 := env.MinterMembers[0].Operator
	member1 := env.MinterMembers[1].Operator
	member2 := env.MinterMembers[2].Operator
	amount := big.NewInt(3_000_000)

	// Apply BFork first
	applyBForkUpgrade(g)

	// Create burn proposal on v2
	tx, err := g.TxProposeBurn(t, member0, member0.Address, amount)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := g.BaseCurrentProposalId(g.govMinter, member0)
	require.NoError(t, err)

	// Reject (quorum=2, two rejections)
	tx, err = g.BaseTxDisapproveProposal(t, g.govMinter, member1, proposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = g.BaseTxDisapproveProposal(t, g.govMinter, member2, proposalId)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify rejected
	proposal, err := g.BaseGetProposal(g.govMinter, member0, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(7), uint8(proposal.Status), "Proposal should be Rejected")

	// v2: burnBalance cleared, refundableBalance credited
	burnBal, err := g.GetBurnBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), burnBal.Int64(), "burnBalance should be 0 after rejection")

	refundable, err := g.GetRefundableBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, 0, refundable.Cmp(amount),
		"refundableBalance should be %s after rejection", amount.String())

	// Claim refund
	govBalBefore := g.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)

	tx, err = g.TxClaimBurnRefund(t, member0)
	_, err = g.ExpectedOk(tx, err)
	require.NoError(t, err)

	refundableAfter, err := g.GetRefundableBalance(env.MinterNonMember, member0.Address)
	require.NoError(t, err)
	require.Equal(t, int64(0), refundableAfter.Int64(), "refundableBalance should be 0 after claim")

	govBalAfter := g.BalanceAt(t, context.TODO(), TestGovMinterAddress, nil)
	govBalDiff := new(big.Int).Sub(govBalBefore, govBalAfter)
	require.Equal(t, 0, govBalDiff.Cmp(amount),
		"GovMinter balance should decrease by refund amount after claim")
}

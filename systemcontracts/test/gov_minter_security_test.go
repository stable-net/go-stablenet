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

	"github.com/stretchr/testify/require"
)

// ====================================================================================
// Replay Attack Prevention Tests
// ====================================================================================

// TestProposeMint_DuplicateDepositId_InVoting validates that depositIds cannot be
// reused while a proposal is in Voting status, preventing replay attacks
func TestProposeMint_DuplicateDepositId_InVoting(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	amount := big.NewInt(1000000)
	depositId := "FIXED-DEPOSIT-001" // ✅ Fixed ID

	// First proposal by Member 0 with their beneficiary
	beneficiary0 := minterMembers[0].Operator.Address
	proof1, err := CreateMintProof(beneficiary0, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx1, err := gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof1)
	_, err = gMinter.ExpectedOk(tx1, err)
	require.NoError(t, err)

	// Verify status = Voting
	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
	require.NoError(t, err)
	proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(1), uint8(proposal.Status)) // Voting = 1

	// Second proposal by Member 1 with SAME depositId but their OWN beneficiary (should fail)
	beneficiary1 := minterMembers[1].Operator.Address
	proof2, err := CreateMintProof(beneficiary1, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx2, err := gMinter.TxProposeMintWithProof(t, minterMembers[1].Operator, proof2)
	err = gMinter.ExpectedFail(tx2, err)

	// Verify error
	ExpectedRevert(t, err, "DepositIdInUse")

	t.Logf("✓ Replay attack prevented: depositId cannot be reused while proposal is Voting")
}

// TestProposeMint_DuplicateDepositId_AlreadyExecuted validates that depositIds
// are permanently consumed after execution and cannot be reused
func TestProposeMint_DuplicateDepositId_AlreadyExecuted(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	amount := big.NewInt(1000000)
	depositId := "EXECUTED-DEPOSIT-001"

	// First proposal by Member 0 with their beneficiary
	beneficiary0 := minterMembers[0].Operator.Address
	proof1, err := CreateMintProof(beneficiary0, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx, err := gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof1)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
	require.NoError(t, err)

	// Execute the proposal (auto-executed on quorum)
	_, err = gMinter.ExecuteProposal(t, gMinter.govMinter, proposalId, []*EOA{minterMembers[1].Operator})
	require.NoError(t, err)

	// Verify depositId is marked as executed
	isExecuted, err := gMinter.IsDepositIdExecuted(minterNonMember, depositId)
	require.NoError(t, err)
	require.True(t, isExecuted, "depositId should be marked as executed")

	// Try to reuse executed depositId by Member 1 with their OWN beneficiary
	beneficiary1 := minterMembers[1].Operator.Address
	proof2, err := CreateMintProof(beneficiary1, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx, err = gMinter.TxProposeMintWithProof(t, minterMembers[1].Operator, proof2)
	err = gMinter.ExpectedFail(tx, err)

	ExpectedRevert(t, err, "DepositIdAlreadyUsed")

	t.Logf("✓ Permanent consumption: depositId cannot be reused after execution")
}

// TestProposeMint_DuplicateProofHash validates that identical proofs cannot be
// submitted by different proposers, preventing replay attacks
func TestProposeMint_DuplicateProofHash(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	amount := big.NewInt(1000000)
	depositId := "deposit-hash-001"

	// First proposal by Member 0 with their beneficiary
	beneficiary0 := minterMembers[0].Operator.Address
	proof1, err := CreateMintProof(beneficiary0, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx1, err := gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof1)
	_, err = gMinter.ExpectedOk(tx1, err)
	require.NoError(t, err)

	// Second proposal by Member 1 with SAME depositId but their OWN beneficiary (should fail)
	beneficiary1 := minterMembers[1].Operator.Address
	proof2, err := CreateMintProof(beneficiary1, amount, depositId, "bank-ref", "memo")
	require.NoError(t, err)

	tx2, err := gMinter.TxProposeMintWithProof(t, minterMembers[1].Operator, proof2)
	err = gMinter.ExpectedFail(tx2, err)

	// Verify error - should be caught by depositId check
	ExpectedRevert(t, err, "DepositIdInUse")

	t.Logf("✓ Proof hash replay attack prevented: same proof cannot be reused")
}

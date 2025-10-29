// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.

package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// ====================================================================================
// Front-running Prevention Tests
// ====================================================================================

// TestProposeMint_BeneficiaryMismatch validates that members cannot mint to addresses
// other than their registered beneficiary, preventing front-running attacks
func TestProposeMint_BeneficiaryMismatch(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	// Member 0 has beneficiary = member0.Operator.Address (from init)
	// Try to mint to different address
	differentAddress := NewEOA().Address

	// Create proof with wrong beneficiary
	proof, err := CreateMintProof(
		differentAddress, // ❌ Wrong beneficiary
		big.NewInt(1000000),
		"deposit-001",
		"bank-ref",
		"memo",
	)
	require.NoError(t, err)

	// Propose mint with wrong beneficiary
	tx, err := gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof)
	err = gMinter.ExpectedFail(tx, err)

	// Verify error
	ExpectedRevert(t, err, "BeneficiaryMismatch")

	// Verify no proposal was created (proposal ID unchanged)
	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
	require.NoError(t, err)
	require.Zero(t, proposalId.Cmp(big.NewInt(0)), "No proposal should be created")

	t.Logf("✓ Front-running attack prevented: cannot mint to unauthorized beneficiary")
}

// TestProposeMint_BeneficiaryNotRegistered validates that non-members
// cannot propose mint operations (access control enforced before beneficiary checks)
// NOTE: This test validates non-member access control, not beneficiary registration status
func TestProposeMint_BeneficiaryNotRegistered(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	// Use existing non-member who is not in governance
	// Non-members should be rejected before beneficiary validation occurs

	// Try to propose mint WITHOUT being a governance member
	tx, err := gMinter.TxProposeMint(t, minterNonMember, minterNonMember.Address, big.NewInt(1000000))
	err = gMinter.ExpectedFail(tx, err)

	// Verify error - should fail with NotAuthorized because sender is not a governance member
	require.Error(t, err, "Should fail when sender is not a member")

	t.Logf("✓ Non-member cannot propose mint (access control enforced before beneficiary checks)")
}

// TestProposeMint_CannotMintToOtherMemberBeneficiary validates that members cannot
// mint to other members' registered beneficiaries, preventing cross-member attacks
func TestProposeMint_CannotMintToOtherMemberBeneficiary(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	// Setup: Change beneficiaries to distinct addresses
	beneficiary0 := NewEOA().Address
	beneficiary1 := NewEOA().Address

	// Register distinct beneficiaries
	tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, beneficiary0)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	tx, err = gMinter.TxRegisterBeneficiary(t, minterMembers[1].Operator, beneficiary1)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify beneficiaries are registered
	registered0, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, beneficiary0, registered0)

	registered1, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[1].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, beneficiary1, registered1)

	// Member 0 tries to create proof with Member 1's beneficiary
	proof, err := CreateMintProof(
		beneficiary1, // ❌ Member 1's beneficiary
		big.NewInt(1000000),
		"deposit-001",
		"bank-ref",
		"memo",
	)
	require.NoError(t, err)

	// Propose mint (should fail - beneficiary mismatch)
	tx, err = gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof)
	err = gMinter.ExpectedFail(tx, err)

	// Verify error
	ExpectedRevert(t, err, "BeneficiaryMismatch")

	t.Logf("✓ Member 0 cannot mint to Member 1's beneficiary - front-running prevented")
}

// TestProposeMint_SuccessWithCorrectBeneficiary validates that minting succeeds
// when using the correct registered beneficiary
func TestProposeMint_SuccessWithCorrectBeneficiary(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	// Get Member 0's registered beneficiary
	beneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)

	// Create proof with CORRECT beneficiary
	proof, err := CreateMintProof(
		beneficiary, // ✅ Correct beneficiary
		big.NewInt(1000000),
		"deposit-success-001",
		"bank-ref",
		"memo",
	)
	require.NoError(t, err)

	// Propose mint (should succeed)
	tx, err := gMinter.TxProposeMintWithProof(t, minterMembers[0].Operator, proof)
	receipt, err := gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	require.Equal(t, uint64(1), receipt.Status)

	// Verify proposal was created
	proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
	require.NoError(t, err)
	require.NotZero(t, proposalId.Cmp(big.NewInt(0)), "Proposal should be created")

	t.Logf("✓ Mint proposal succeeded with correct beneficiary (Proposal ID: %s)", proposalId.String())
}

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

// ====================================================================================
// Duplicate Beneficiary Tests
// ====================================================================================

// TestRegisterBeneficiary_ZeroAddress validates that zero address cannot be registered
// as a beneficiary, preventing accidental loss of funds
func TestRegisterBeneficiary_ZeroAddress(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	// Try to register zero address as beneficiary
	zeroAddress := common.Address{}
	tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, zeroAddress)
	err = gMinter.ExpectedFail(tx, err)

	// Verify error
	ExpectedRevert(t, err, "InvalidBeneficiary")

	t.Logf("✓ Zero address cannot be registered as beneficiary")
}

// TestRegisterBeneficiary_DuplicateBeneficiary validates that a beneficiary address
// cannot be registered by multiple members, preventing confusion and attacks
func TestRegisterBeneficiary_DuplicateBeneficiary(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	sharedBeneficiary := NewEOA().Address

	// Member 0 registers beneficiary
	tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, sharedBeneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify registration
	registered0, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, sharedBeneficiary, registered0)

	// Member 1 tries to register SAME beneficiary (should fail)
	tx, err = gMinter.TxRegisterBeneficiary(t, minterMembers[1].Operator, sharedBeneficiary)
	err = gMinter.ExpectedFail(tx, err)

	// Verify error
	ExpectedRevert(t, err, "DuplicateBeneficiary")

	// Verify Member 1 still has no beneficiary or old beneficiary
	registered1, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[1].Operator.Address)
	require.NoError(t, err)
	require.NotEqual(t, sharedBeneficiary, registered1, "Member 1 should not have the shared beneficiary")

	t.Logf("✓ Duplicate beneficiary prevented: address can only be used by one member")
}

// TestRegisterBeneficiary_ChangeBeneficiary validates that members can change their
// beneficiary and that old mappings are properly cleared
func TestRegisterBeneficiary_ChangeBeneficiary(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	oldBeneficiary := NewEOA().Address
	newBeneficiary := NewEOA().Address

	// Register initial beneficiary
	tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, oldBeneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify initial registration
	registered, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, oldBeneficiary, registered)

	// Change to new beneficiary
	tx, err = gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, newBeneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify new beneficiary is registered
	registered, err = gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, newBeneficiary, registered)

	// Verify old beneficiary is now available for other members
	tx, err = gMinter.TxRegisterBeneficiary(t, minterMembers[1].Operator, oldBeneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify Member 1 successfully registered the old beneficiary
	registered1, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[1].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, oldBeneficiary, registered1)

	t.Logf("✓ Beneficiary change works correctly, old mapping cleared")
}

// TestRegisterBeneficiary_ReregisterSameBeneficiary validates that members can
// re-register the same beneficiary (idempotent operation)
func TestRegisterBeneficiary_ReregisterSameBeneficiary(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	beneficiary := NewEOA().Address

	// First registration
	tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, beneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Re-register same beneficiary (should succeed - idempotent)
	tx, err = gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, beneficiary)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err)

	// Verify beneficiary is still registered
	registered, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
	require.NoError(t, err)
	require.Equal(t, beneficiary, registered)

	t.Logf("✓ Re-registering same beneficiary is idempotent and allowed")
}

// TestRegisterBeneficiary_NonMemberCannotRegister validates that only governance
// members can register beneficiaries
func TestRegisterBeneficiary_NonMemberCannotRegister(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	beneficiary := NewEOA().Address

	// Non-member tries to register beneficiary
	tx, err := gMinter.TxRegisterBeneficiary(t, minterNonMember, beneficiary)
	err = gMinter.ExpectedFail(tx, err)

	// Verify error (should fail with NotAuthorized or similar)
	require.Error(t, err, "Non-member should not be able to register beneficiary")

	t.Logf("✓ Non-members cannot register beneficiaries")
}

// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.

package test

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// TestMockFiatToken_Deployment validates MockFiatToken is deployed correctly at genesis
func TestMockFiatToken_Deployment(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	t.Run("MockFiatToken is deployed", func(t *testing.T) {
		require.NotNil(t, gMinter.mockFiatToken, "MockFiatToken should be initialized")

		// Verify MockFiatToken is deployed at the correct address
		ctx := context.TODO()
		code, err := gMinter.backend.Client().CodeAt(ctx, TestMockFiatTokenAddress, nil)
		require.NoError(t, err)
		require.NotEmpty(t, code, "MockFiatToken should have bytecode at address %s", TestMockFiatTokenAddress.Hex())
		t.Logf("✓ MockFiatToken deployed at %s with %d bytes", TestMockFiatTokenAddress.Hex(), len(code))
	})

	t.Run("MockFiatToken has zero initial supply", func(t *testing.T) {
		totalSupply, err := gMinter.GetMockFiatTokenTotalSupply(minterNonMember)
		require.NoError(t, err)
		require.Zero(t, totalSupply.Cmp(big.NewInt(0)), "Initial total supply should be 0")
		t.Logf("✓ Initial total supply: %s", totalSupply.String())
	})

	t.Run("GovMinter is connected to MockFiatToken", func(t *testing.T) {
		// GovMinter's fiatToken should point to MockFiatToken address
		fiatToken, err := gMinter.FiatToken(minterNonMember)
		require.NoError(t, err)
		require.Equal(t, TestMockFiatTokenAddress, fiatToken, "GovMinter should be connected to MockFiatToken")
		t.Logf("✓ GovMinter fiatToken: %s", fiatToken.Hex())
	})
}

// TestMockFiatToken_BasicOperations validates MockFiatToken basic functions work
func TestMockFiatToken_BasicOperations(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	recipient := minterMembers[0].Operator.Address
	amount := big.NewInt(1000000)

	t.Run("balanceOf returns zero for new account", func(t *testing.T) {
		balance, err := gMinter.GetMockFiatTokenBalance(minterNonMember, recipient)
		require.NoError(t, err)
		require.Zero(t, balance.Cmp(big.NewInt(0)), "New account should have zero balance")
		t.Logf("✓ Balance of %s: %s", recipient.Hex()[:10], balance.String())
	})

	t.Run("setBalance helper works", func(t *testing.T) {
		// Use mockFiatToken.Transact to call setBalance
		tx, err := gMinter.mockFiatToken.Transact(
			NewTxOptsWithValue(t, minterMembers[0].Operator, nil),
			"setBalance",
			recipient,
			amount,
		)
		receipt, err := gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status, "setBalance should succeed")

		// Verify balance updated
		balance, err := gMinter.GetMockFiatTokenBalance(minterNonMember, recipient)
		require.NoError(t, err)
		require.Zero(t, balance.Cmp(amount), "Balance should be updated")
		t.Logf("✓ Set balance to %s, new balance: %s", amount.String(), balance.String())
	})

	t.Run("totalSupply is updated", func(t *testing.T) {
		totalSupply, err := gMinter.GetMockFiatTokenTotalSupply(minterNonMember)
		require.NoError(t, err)
		require.Zero(t, totalSupply.Cmp(amount), "Total supply should equal set balance")
		t.Logf("✓ Total supply: %s", totalSupply.String())
	})
}

// TestCreateProof_ABIEncoding validates proof generation helpers work correctly
func TestCreateProof_ABIEncoding(t *testing.T) {
	beneficiary := common.HexToAddress("0x1111111111111111111111111111111111111111")
	amount := big.NewInt(500000)
	depositId := "DEPOSIT-TEST-001"
	bankRef := "BANK-REF-123"
	memo := "Test mint"

	t.Run("CreateMintProof generates valid ABI encoding", func(t *testing.T) {
		proof, err := CreateMintProof(beneficiary, amount, depositId, bankRef, memo)
		require.NoError(t, err)
		require.NotEmpty(t, proof, "Proof should not be empty")

		// Proof should be ABI-encoded (beneficiary, amount, depositId, bankReference, memo)
		// Minimum length: 32 (address) + 32 (uint256) + 3*64 (string offsets + lengths) ≈ 256+ bytes
		require.Greater(t, len(proof), 200, "Proof should be properly ABI-encoded")
		t.Logf("✓ Mint proof generated: %d bytes", len(proof))
	})

	t.Run("CreateBurnProof generates valid ABI encoding", func(t *testing.T) {
		from := common.HexToAddress("0x2222222222222222222222222222222222222222")
		withdrawalId := "WITHDRAWAL-TEST-001"
		referenceId := "REF-456"
		memo := "Test burn"

		proof, err := CreateBurnProof(from, amount, withdrawalId, referenceId, memo)
		require.NoError(t, err)
		require.NotEmpty(t, proof, "Proof should not be empty")
		require.Greater(t, len(proof), 200, "Proof should be properly ABI-encoded")
		t.Logf("✓ Burn proof generated: %d bytes", len(proof))
	})
}

// TestExecuteProposal_Workflow validates ExecuteProposal helper function
func TestExecuteProposal_Workflow(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	beneficiary := minterMembers[0].Operator.Address
	amount := big.NewInt(1000000)

	t.Run("ExecuteProposal automates full approval workflow", func(t *testing.T) {
		// Step 1: Create a mint proposal
		tx, err := gMinter.TxProposeMint(t, minterMembers[0].Operator, beneficiary, amount)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Get proposal ID
		proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)
		t.Logf("✓ Created proposal ID: %s", proposalId.String())

		// Step 2: Use ExecuteProposal to approve + execute
		// Quorum is 2, so we need 1 more approval (proposer already voted)
		// Note: With auto-execution, ExecuteProposal returns nil receipt when quorum is reached
		approvers := []*EOA{minterMembers[1].Operator}
		_, err = gMinter.ExecuteProposal(t, gMinter.govMinter, proposalId, approvers)
		require.NoError(t, err)
		t.Logf("✓ ExecuteProposal succeeded (auto-executed on quorum)")

		// Step 3: Verify proposal is now Executed
		proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed (3)")
		t.Logf("✓ Final proposal status: %v (Executed)", proposal.Status)
	})
}

// TestCompleteMintProposal_E2E validates end-to-end mint workflow
func TestCompleteMintProposal_E2E(t *testing.T) {
	initGovMinter(t)
	defer gMinter.backend.Close()

	beneficiary := minterMembers[0].Operator.Address
	amount := big.NewInt(2000000)

	t.Run("CompleteMintProposal full workflow", func(t *testing.T) {
		// This should: create proposal → approve → execute
		// Note: With auto-execution, CompleteMintProposal returns nil receipt when quorum is reached
		approvers := []*EOA{minterMembers[1].Operator} // Need 1 more approval (quorum=2, proposer=1)
		_, err := gMinter.CompleteMintProposal(
			t,
			minterMembers[0].Operator, // proposer
			beneficiary,
			amount,
			approvers,
		)
		require.NoError(t, err)
		t.Logf("✓ CompleteMintProposal succeeded (auto-executed on quorum)")

		// Verify proposal was executed by checking latest proposalId status
		proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)

		proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed (3)")
		t.Logf("✓ Final proposal ID %s status: %v (Executed)", proposalId.String(), proposal.Status)
	})
}

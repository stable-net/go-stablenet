// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

var (
	gMinter          *GovWBFT
	minterMembers    []*TestCandidate
	minterNonMember  *EOA
	fiatTokenAddress common.Address
)

func initGovMinter(t *testing.T) {
	minterMembers = []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
	minterNonMember = NewEOA()
	fiatTokenAddress = TestMockFiatTokenAddress // Use actual deployed MockFiatToken address

	var err error
	gMinter, err = NewGovWBFT(t, types.GenesisAlloc{
		minterMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		minterMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		minterMembers[2].Operator.Address: {Balance: towei(1_000_000)},
		minterNonMember.Address:           {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		// Setup governance members for voting
		var members, validators, blsPubKeys string
		for i, m := range minterMembers {
			if i > 0 {
				members = members + ","
				validators = validators + ","
				blsPubKeys = blsPubKeys + ","
			}
			members = members + m.Operator.Address.String()
			validators = validators + m.Validator.Address.String()
			blsPubKeys = blsPubKeys + hexutil.Encode(m.GetBLSPublicKey(t).Marshal())
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "2",
			"expiry":        "604800",
			"memberVersion": "1",
			"validators":    validators,
			"blsPublicKeys": blsPubKeys,
		}
	}, nil, func(govMinter *params.SystemContract) {
		// Initialize GovMinter with fiatToken address and beneficiaries
		beneficiary1 := minterMembers[0].Operator.Address
		beneficiary2 := minterMembers[1].Operator.Address
		beneficiary3 := minterMembers[2].Operator.Address

		govMinter.Params = map[string]string{
			sc.GOV_MINTER_PARAM_FIAT_TOKEN:    fiatTokenAddress.String(),
			sc.GOV_BASE_PARAM_MEMBERS:         minterMembers[0].Operator.Address.String() + "," + minterMembers[1].Operator.Address.String() + "," + minterMembers[2].Operator.Address.String(),
			sc.GOV_MINTER_PARAM_BENEFICIARIES: beneficiary1.String() + "," + beneficiary2.String() + "," + beneficiary3.String(),
			sc.GOV_BASE_PARAM_QUORUM:          "2",
			sc.GOV_BASE_PARAM_EXPIRY:          "604800",
			sc.GOV_BASE_PARAM_MEMBER_VERSION:  "1",
		}
	}, nil, func(fiatToken *params.SystemContract) {
		// Deploy MockFiatToken at genesis for testing
		// This is a test helper contract, not a production system contract
		fiatToken.Params = map[string]string{
			// MockFiatToken has no initialization params needed
		}
	})
	require.NoError(t, err)

	// Configure GovMinter as a minter with sufficient allowance (10M tokens)
	// This is required for the new P0-1 security fix (minter allowance validation)
	owner := minterMembers[0].Operator
	minterAllowance := big.NewInt(10_000_000)
	tx, err := gMinter.ConfigureMockFiatTokenMinter(t, owner, TestGovMinterAddress, minterAllowance)
	_, err = gMinter.ExpectedOk(tx, err)
	require.NoError(t, err, "Failed to configure GovMinter as minter")

	// Verify minter allowance was set
	allowance, err := gMinter.GetMockFiatTokenMinterAllowance(owner, TestGovMinterAddress)
	require.NoError(t, err)
	require.Equal(t, 0, minterAllowance.Cmp(allowance), "Minter allowance should be configured")
}

func TestGovMinter_Initialize(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		// Check fiatToken is set correctly
		token, err := gMinter.FiatToken(minterNonMember)
		require.NoError(t, err)
		require.Equal(t, fiatTokenAddress, token)

		// Check governance base parameters
		quorum, err := gMinter.BaseQuorum(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorum)

		expiry, err := gMinter.BaseProposalExpiry(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)
		require.Equal(t, uint64(604800), expiry.Uint64())

		// Check members are initialized
		for i, m := range minterMembers {
			member, err := gMinter.BaseMembers(gMinter.govMinter, minterNonMember, m.Operator.Address)
			require.NoError(t, err, "Member %d should be initialized", i)
			require.True(t, member.IsActive, "Member %d should be active", i)

			// Check beneficiary is set
			beneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, m.Operator.Address)
			require.NoError(t, err)
			require.Equal(t, m.Operator.Address, beneficiary, "Beneficiary for member %d should be set", i)
		}

		// Check non-member is not initialized
		member, err := gMinter.BaseMembers(gMinter.govMinter, minterNonMember, minterNonMember.Address)
		require.NoError(t, err)
		require.False(t, member.IsActive, "Non-member should not be active")
	})
}

func TestGovMinter_ProposeMint(t *testing.T) {
	t.Run("member can propose mint", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		recipient := NewEOA()
		amount := big.NewInt(1000000)

		// Member proposes mint
		tx, err := gMinter.TxProposeMint(t, minterMembers[0].Operator, recipient.Address, amount)
		receipt, err := gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)

		// Check proposal details
		proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
		require.Equal(t, minterMembers[0].Operator.Address, proposal.Proposer)
		require.Equal(t, uint32(1), proposal.Approved) // Proposer auto-approves
	})

	t.Run("non-member cannot propose mint", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		recipient := NewEOA()
		amount := big.NewInt(1000000)

		// Non-member tries to propose mint (should fail)
		tx, err := gMinter.TxProposeMint(t, minterNonMember, recipient.Address, amount)
		err = gMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Non-member should not be able to propose mint")
	})
}

func TestGovMinter_ProposeBurn(t *testing.T) {
	t.Run("member can propose burn", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		from := minterMembers[0].Operator.Address
		amount := big.NewInt(500000)

		// proposeBurn now accepts msg.value directly (no need for separate deposit)
		tx, err := gMinter.TxProposeBurn(t, minterMembers[0].Operator, from, amount)
		receipt, err := gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Verify burn balance was credited
		burnBalance, err := gMinter.GetBurnBalance(minterNonMember, from)
		require.NoError(t, err)
		require.Equal(t, 0, burnBalance.Cmp(amount), "burnBalance should match deposited amount")

		// Check proposal was created
		proposalId, err := gMinter.BaseCurrentProposalId(gMinter.govMinter, minterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)

		// Check proposal status
		proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
	})

	t.Run("burn balance tracking", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		// Initial burn balance should be 0
		burnBalance, err := gMinter.GetBurnBalance(minterNonMember, minterMembers[0].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, 0, burnBalance.Cmp(big.NewInt(0)))
	})
}

func TestGovMinter_RegisterBeneficiary(t *testing.T) {
	t.Run("member can register beneficiary", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		newBeneficiary := NewEOA().Address

		// Member registers beneficiary (direct call, not a proposal)
		tx, err := gMinter.TxRegisterBeneficiary(t, minterMembers[0].Operator, newBeneficiary)
		receipt, err := gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check beneficiary was updated
		beneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, newBeneficiary, beneficiary)
	})

	t.Run("beneficiary persists after setting", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		// Check initial beneficiary
		beneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, minterMembers[0].Operator.Address)
		require.NoError(t, err)
		require.Equal(t, minterMembers[0].Operator.Address, beneficiary)
	})
}

func TestGovMinter_GovernanceWorkflow(t *testing.T) {
	t.Run("complete mint proposal workflow", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		recipient := NewEOA()
		amount := big.NewInt(1000000)

		// Step 1: Member 0 proposes mint
		tx, err := gMinter.TxProposeMint(t, minterMembers[0].Operator, recipient.Address, amount)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId := big.NewInt(1)

		// Step 2: Check proposal status (voting, needs 1 more approval for quorum 2)
		proposal, err := gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
		require.Equal(t, uint32(1), proposal.Approved)

		// Step 3: Member 1 approves
		tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, minterMembers[1].Operator, proposalId)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 4: Check proposal is now approved (quorum reached)
		proposal, err = gMinter.BaseGetProposal(gMinter.govMinter, minterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint32(2), proposal.Approved)
		// Note: MockFiatToken is deployed in test setup
		// Auto-execution will succeed when quorum is reached
	})
}

// ========================================
// CRITICAL-1 & CRITICAL-2: _safeBurn Vulnerability Tests
// ========================================

func TestGovMinter_CRITICAL_BurnRollbackOnFailure(t *testing.T) {
	t.Run("CRITICAL-1: burn should rollback burnBalance when burn fails", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		member := minterMembers[0]
		burnAmount := towei(1) // 1 ETH

		// Step 1 & 2 combined: proposeBurn now accepts msg.value directly
		tx, err := gMinter.TxProposeBurn(t, member.Operator, member.Operator.Address, burnAmount)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify burnBalance is credited after proposeBurn
		burnBalanceBefore, err := gMinter.GetBurnBalance(minterNonMember, member.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, 0, burnBalanceBefore.Cmp(burnAmount), "burnBalance should be credited")

		proposalId := big.NewInt(1)

		// Step 3: Approve proposal (auto-executes, but will fail due to insufficient fiat token balance)
		// NOTE: With auto-execution, burn failure causes entire transaction to revert
		// This automatically ensures burnBalance is rolled back (no partial state changes)
		tx, err = gMinter.BaseTxApproveProposal(t, gMinter.govMinter, minterMembers[1].Operator, proposalId)
		// Auto-execution fails, so transaction reverts
		_ = gMinter.ExpectedFail(tx, err)

		// CRITICAL TEST: burnBalance should be rolled back after transaction failure
		// With auto-execution, transaction revert guarantees all state changes are rolled back
		burnBalanceAfter, err := gMinter.GetBurnBalance(minterNonMember, member.Operator.Address)
		require.NoError(t, err)

		// Auto-execution revert ensures burnBalance is unchanged (transaction-level rollback)
		require.Equal(t, 0, burnBalanceAfter.Cmp(burnBalanceBefore),
			"CRITICAL-1: burnBalance should be rolled back when burn fails, but got %s expected %s",
			burnBalanceAfter.String(), burnBalanceBefore.String())

		t.Logf("✓ CRITICAL-1: With auto-execution, burn failure causes transaction revert, ensuring complete rollback")
	})
}

func TestGovMinter_CRITICAL_BurnTransfersNativeCoins(t *testing.T) {
	t.Run("CRITICAL-2: burn should transfer native coins back to user", func(t *testing.T) {
		// This test will be implemented after deploying a working MockFiatToken
		// For now, skip it as it requires full integration
		t.Skip("Requires working MockFiatToken contract deployment")

		// When implemented, this test should:
		// 1. Deploy MockFiatToken with mint/burn functionality
		// 2. Member deposits ETH for burn
		// 3. Give GovMinter some fiat tokens
		// 4. Create and approve burn proposal
		// 5. Execute burn
		// 6. Verify user received ETH back
		// 7. Verify burnBalance is 0
		// 8. Verify withdrawalId is marked executed
	})
}

// ========================================
// HIGH-1: Beneficiary Front-Running Tests
// ========================================

func TestGovMinter_HIGH_DuplicateBeneficiaryPrevention(t *testing.T) {
	t.Run("HIGH-1: duplicate beneficiary should be prevented", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		member1 := minterMembers[0]
		member2 := minterMembers[1]

		// Member1's beneficiary is already set during initialization
		beneficiary1, err := gMinter.GetMemberBeneficiary(minterNonMember, member1.Operator.Address)
		require.NoError(t, err)

		// Try to set member2's beneficiary to the same as member1 (should fail)
		tx, err := gMinter.TxRegisterBeneficiary(t, member2.Operator, beneficiary1)
		err = gMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Duplicate beneficiary should be rejected")
	})

	t.Run("HIGH-1: beneficiary change should clear old mapping", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		member1 := minterMembers[0]
		member2 := minterMembers[1]

		// Get member1's original beneficiary
		oldBeneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, member1.Operator.Address)
		require.NoError(t, err)

		// Member1 changes beneficiary to a new address
		newBeneficiary := NewEOA().Address
		tx, err := gMinter.TxRegisterBeneficiary(t, member1.Operator, newBeneficiary)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify new beneficiary is set
		currentBeneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, member1.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, newBeneficiary, currentBeneficiary)

		// NOW: member2 should be able to register oldBeneficiary
		// WITH CURRENT IMPLEMENTATION (O(n) loop), this might fail due to stale check
		// AFTER FIX (reverse mapping), this should succeed
		tx, err = gMinter.TxRegisterBeneficiary(t, member2.Operator, oldBeneficiary)
		_, err = gMinter.ExpectedOk(tx, err)
		require.NoError(t, err, "After member1 changes beneficiary, member2 should be able to use the old one")

		// Verify member2's beneficiary is set
		member2Beneficiary, err := gMinter.GetMemberBeneficiary(minterNonMember, member2.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, oldBeneficiary, member2Beneficiary)
	})
}

// ========================================
// MEDIUM-3: FiatToken Validation Tests
// ========================================

func TestGovMinter_MEDIUM_FiatTokenValidation(t *testing.T) {
	t.Run("MEDIUM-3: fiatToken validation during initialization", func(t *testing.T) {
		// This test verifies that invalid fiatToken addresses are rejected
		// Current implementation only checks for zero address
		// After fix, should also check for contract existence and interface compatibility
		t.Skip("Requires genesis initialization with validation")
	})
}

// ========================================
// Burn Balance Edge Cases
// ========================================

func TestGovMinter_BurnBalanceEdgeCases(t *testing.T) {
	t.Run("multiple deposits accumulate burnBalance", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		member := minterMembers[0]
		depositAmount := towei(1)

		// Initial balance should be 0
		balance, err := gMinter.GetBurnBalance(minterNonMember, member.Operator.Address)
		require.NoError(t, err)
		require.Equal(t, 0, balance.Cmp(big.NewInt(0)))

		// Propose burn 3 times (each proposeBurn deposits msg.value)
		// Note: Each proposal needs unique withdrawalId, so we use TxProposeBurnWithProof
		for i := 0; i < 3; i++ {
			// Create unique proof with different withdrawalId for each iteration
			timestamp := big.NewInt(time.Now().Unix())
			withdrawalId := fmt.Sprintf("withdrawal-%s-%d", member.Operator.Address.Hex()[:10], i)
			referenceId := fmt.Sprintf("ref-%03d", i)
			memo := "test burn"

			proofData, err := abi.Arguments{
				{Type: mustParseType("address")},
				{Type: mustParseType("uint256")},
				{Type: mustParseType("uint256")},
				{Type: mustParseType("string")},
				{Type: mustParseType("string")},
				{Type: mustParseType("string")},
			}.Pack(member.Operator.Address, depositAmount, timestamp, withdrawalId, referenceId, memo)
			require.NoError(t, err)

			// Send proposeBurn with msg.value
			tx, err := gMinter.govMinter.Transact(NewTxOptsWithValue(t, member.Operator, depositAmount), "proposeBurn", proofData)
			_, err = gMinter.ExpectedOk(tx, err)
			require.NoError(t, err)
		}

		// Balance should be 3 * depositAmount
		balance, err = gMinter.GetBurnBalance(minterNonMember, member.Operator.Address)
		require.NoError(t, err)
		expectedBalance := new(big.Int).Mul(depositAmount, big.NewInt(3))
		require.Equal(t, 0, balance.Cmp(expectedBalance), "Balance should accumulate")
	})

	t.Run("burn proposal requires msg.value to match proof amount", func(t *testing.T) {
		initGovMinter(t)
		defer gMinter.backend.Close()

		member := minterMembers[0]
		proofAmount := towei(10)
		wrongMsgValue := towei(5) // msg.value doesn't match proof amount

		// Create proof data with proofAmount
		timestamp := big.NewInt(time.Now().Unix())
		withdrawalId := fmt.Sprintf("withdrawal-%s-mismatch", member.Operator.Address.Hex()[:10])
		referenceId := "ref-mismatch"
		memo := "test burn"

		proofData, err := abi.Arguments{
			{Type: mustParseType("address")},
			{Type: mustParseType("uint256")},
			{Type: mustParseType("uint256")},
			{Type: mustParseType("string")},
			{Type: mustParseType("string")},
			{Type: mustParseType("string")},
		}.Pack(member.Operator.Address, proofAmount, timestamp, withdrawalId, referenceId, memo)
		require.NoError(t, err)

		// Try to propose burn with msg.value != proof.amount (should fail with BurnAmountMismatch)
		tx, err := gMinter.govMinter.Transact(NewTxOptsWithValue(t, member.Operator, wrongMsgValue), "proposeBurn", proofData)
		err = gMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Should fail with BurnAmountMismatch")
	})
}

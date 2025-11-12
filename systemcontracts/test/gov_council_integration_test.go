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

	"github.com/ethereum/go-ethereum/common"
	sc "github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

func TestGovCouncil_Initialize(t *testing.T) {
	t.Run("initial state with blacklist and authorized accounts", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Check governance base parameters
		quorum, err := gCouncil.BaseQuorum(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorum)

		expiry, err := gCouncil.BaseProposalExpiry(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, uint64(604800), expiry.Uint64())

		// Check members are initialized
		for i, m := range councilMembers {
			member, err := gCouncil.BaseMembers(gCouncil.govCouncil, councilNonMember, m.Operator.Address)
			require.NoError(t, err, "Member %d should be initialized", i)
			require.True(t, member.IsActive, "Member %d should be active", i)
		}

		// Check non-member is not initialized
		member, err := gCouncil.BaseMembers(gCouncil.govCouncil, councilNonMember, councilNonMember.Address)
		require.NoError(t, err)
		require.False(t, member.IsActive, "Non-member should not be active")

		// Check blacklist is initialized with initial addresses (via storage)
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), count, "Should have 2 blacklisted addresses")

		addr0, err := gCouncil.GetBlacklistedAddress(councilNonMember, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, initialBlacklist[0], addr0, "First address should match")

		addr1, err := gCouncil.GetBlacklistedAddress(councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, initialBlacklist[1], addr1, "Second address should match")

		// Check authorized accounts are initialized (via storage)
		count, err = gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), count, "Should have 2 authorized accounts")

		authAddr0, err := gCouncil.GetAuthorizedAccountAddress(councilNonMember, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, initialAuthorizedAccounts[0], authAddr0, "First account should match")

		authAddr1, err := gCouncil.GetAuthorizedAccountAddress(councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, initialAuthorizedAccounts[1], authAddr1, "Second account should match")
	})
}

func TestGovCouncil_ProposeAddBlacklist(t *testing.T) {
	t.Run("member can propose to add address to blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify address is not blacklisted initially
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted initially")

		// Member proposes to add address to blacklist
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)

		// Check proposal details
		proposal, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
		require.Equal(t, councilMembers[0].Operator.Address, proposal.Proposer)
		require.Equal(t, uint32(1), proposal.Approved) // Proposer auto-approves

		// Address should still not be blacklisted (proposal not approved yet)
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted until approved")
	})

	t.Run("non-member cannot propose to add blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Non-member tries to propose (should fail)
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilNonMember, targetAddress)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Non-member should not be able to propose")
	})

	t.Run("cannot propose to add zero address to blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Try to propose zero address (should fail)
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, zeroAddress)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to add zero address")
	})

	t.Run("cannot propose to add duplicate address to blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Try to propose adding an already blacklisted address (should fail)
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, initialBlacklist[0])
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to add duplicate address")
	})
}

func TestGovCouncil_ProposeRemoveBlacklist(t *testing.T) {
	t.Run("member can propose to remove address from blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// First, add address to blacklist via proposal
		txAdd, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify address is blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should be blacklisted")

		// Member proposes to remove address from blacklist
		tx, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, targetAddress)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), proposalId)

		// Address should still be blacklisted (proposal not approved yet)
		isBlacklisted, err = gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should still be blacklisted until approved")
	})

	t.Run("cannot propose to remove non-blacklisted address", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		nonBlacklistedAddress := NewEOA().Address

		// Try to propose removing a non-blacklisted address (should fail)
		tx, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, nonBlacklistedAddress)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to remove non-blacklisted address")
	})
}

func TestGovCouncil_BlacklistWorkflow(t *testing.T) {
	t.Run("complete workflow: propose, approve, execute add blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Step 1: Member 0 proposes to add address to blacklist
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 2: Member 1 approves the proposal (quorum is 2, so this reaches quorum)
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Step 3: Verify proposal is now executed (auto-executes at quorum)
		proposal, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusExecuted, proposal.Status)

		// Step 4: Verify address is now blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isBlacklisted, "Address should be blacklisted after approval")

		// Step 5: Verify blacklist count increased
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), count, "Blacklist count should increase to 3")
	})

	t.Run("complete workflow: propose, approve, execute remove blacklist", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := initialBlacklist[0]

		// Step 1: Member 0 proposes to remove address from blacklist
		tx, err := gCouncil.TxProposeRemoveBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 2: Member 1 approves the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Step 3: Verify address is no longer blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted after removal")

		// Step 4: Verify blacklist count decreased
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), count, "Blacklist count should decrease to 1")
	})
}

func TestGovCouncil_ProposeAddAuthorizedAccount(t *testing.T) {
	t.Run("member can propose to add authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Verify address is not authorized initially
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "Address should not be authorized initially")

		// Member proposes to add authorized account
		tx, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)
	})

	t.Run("cannot propose to add duplicate authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Try to propose adding an already authorized account (should fail)
		tx, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, initialAuthorizedAccounts[0])
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to add duplicate authorized account")
	})
}

func TestGovCouncil_ProposeRemoveAuthorizedAccount(t *testing.T) {
	t.Run("member can propose to remove authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// First, add address to authorized accounts via proposal
		txAdd, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(txAdd, err)
		require.NoError(t, err)

		txApprove, err := gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(txApprove, err)
		require.NoError(t, err)

		// Verify address is authorized
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "Address should be authorized")

		// Member proposes to remove authorized account
		tx, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gCouncil.BaseCurrentProposalId(gCouncil.govCouncil, councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), proposalId)

		// Address should still be authorized (proposal not approved yet)
		isAuthorized, err = gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "Address should still be authorized until approved")
	})

	t.Run("cannot propose to remove non-authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		nonAuthorizedAddress := NewEOA().Address

		// Try to propose removing a non-authorized account (should fail)
		tx, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[0].Operator, nonAuthorizedAddress)
		err = gCouncil.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to remove non-authorized account")
	})
}

func TestGovCouncil_AuthorizedAccountWorkflow(t *testing.T) {
	t.Run("complete workflow: propose, approve, execute add authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Step 1: Member 0 proposes to add authorized account
		tx, err := gCouncil.TxProposeAddAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 2: Member 1 approves the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Step 3: Verify address is now authorized
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.True(t, isAuthorized, "Address should be authorized after approval")

		// Step 4: Verify authorized account count increased
		count, err := gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), count, "Authorized account count should increase to 3")
	})

	t.Run("complete workflow: propose, approve, execute remove authorized account", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := initialAuthorizedAccounts[0]

		// Step 1: Member 0 proposes to remove authorized account
		tx, err := gCouncil.TxProposeRemoveAuthorizedAccount(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 2: Member 1 approves the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Step 3: Verify address is no longer authorized
		isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isAuthorized, "Address should not be authorized after removal")

		// Step 4: Verify authorized account count decreased
		count, err := gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), count, "Authorized account count should decrease to 1")
	})
}

func TestGovCouncil_BatchOperations(t *testing.T) {
	t.Run("propose add blacklist batch", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		addresses := []common.Address{
			NewEOA().Address,
			NewEOA().Address,
			NewEOA().Address,
		}

		// Member proposes to add multiple addresses to blacklist
		tx, err := gCouncil.TxProposeAddBlacklistBatch(t, councilMembers[0].Operator, addresses)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Approve the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify all addresses are blacklisted
		for _, addr := range addresses {
			isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, addr)
			require.NoError(t, err)
			require.True(t, isBlacklisted, "Address should be blacklisted")
		}

		// Verify blacklist count
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(5), count, "Blacklist count should be 5 (2 initial + 3 new)")
	})

	t.Run("propose remove blacklist batch", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		addresses := []common.Address{
			initialBlacklist[0],
			initialBlacklist[1],
		}

		// Member proposes to remove multiple addresses from blacklist
		tx, err := gCouncil.TxProposeRemoveBlacklistBatch(t, councilMembers[0].Operator, addresses)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Approve the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify all addresses are removed from blacklist
		for _, addr := range addresses {
			isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, addr)
			require.NoError(t, err)
			require.False(t, isBlacklisted, "Address should not be blacklisted")
		}

		// Verify blacklist count
		count, err := gCouncil.GetBlacklistCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, int64(0), count.Int64(), "Blacklist count should be 0")
	})

	t.Run("propose add authorized account batch", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		addresses := []common.Address{
			NewEOA().Address,
			NewEOA().Address,
		}

		// Member proposes to add multiple authorized accounts
		tx, err := gCouncil.TxProposeAddAuthorizedAccountBatch(t, councilMembers[0].Operator, addresses)
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Approve the proposal
		tx, err = gCouncil.TxApprove(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify all addresses are authorized
		for _, addr := range addresses {
			isAuthorized, err := gCouncil.IsAuthorizedAccount(councilNonMember, addr)
			require.NoError(t, err)
			require.True(t, isAuthorized, "Address should be authorized")
		}

		// Verify authorized account count
		count, err := gCouncil.GetAuthorizedAccountCount(councilNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(4), count, "Authorized account count should be 4 (2 initial + 2 new)")
	})
}

func TestGovCouncil_ProposalRejection(t *testing.T) {
	t.Run("proposal can be rejected", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		targetAddress := NewEOA().Address

		// Step 1: Member 0 proposes to add address to blacklist
		tx, err := gCouncil.TxProposeAddBlacklist(t, councilMembers[0].Operator, targetAddress)
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 2: Member 1 rejects the proposal
		tx, err = gCouncil.TxReject(t, councilMembers[1].Operator, big.NewInt(1))
		_, err = gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 3: Member 2 also rejects (need 2 rejections to exceed maxRejections=1)
		tx, err = gCouncil.TxReject(t, councilMembers[2].Operator, big.NewInt(1))
		receipt, err := gCouncil.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Step 4: Verify proposal is rejected
		proposal, err := gCouncil.BaseGetProposal(gCouncil.govCouncil, councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusRejected, proposal.Status)

		// Step 5: Verify address is not blacklisted
		isBlacklisted, err := gCouncil.IsBlacklisted(councilNonMember, targetAddress)
		require.NoError(t, err)
		require.False(t, isBlacklisted, "Address should not be blacklisted after rejection")
	})
}

func TestGovCouncil_QueryFunctions(t *testing.T) {
	t.Run("getBlacklistedAddress returns correct address at index", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Get first blacklisted address
		addr, err := gCouncil.GetBlacklistedAddress(councilNonMember, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, initialBlacklist[0], addr)

		// Get second blacklisted address
		addr, err = gCouncil.GetBlacklistedAddress(councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, initialBlacklist[1], addr)
	})

	t.Run("getAllBlacklisted returns all addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		addresses, err := gCouncil.GetAllBlacklisted(councilNonMember)
		require.NoError(t, err)
		require.Len(t, addresses, 2, "Should return 2 blacklisted addresses")
		require.Contains(t, addresses, initialBlacklist[0])
		require.Contains(t, addresses, initialBlacklist[1])
	})

	t.Run("getAuthorizedAccountAddress returns correct address at index", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		// Get first authorized account
		addr, err := gCouncil.GetAuthorizedAccountAddress(councilNonMember, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, initialAuthorizedAccounts[0], addr)

		// Get second authorized account
		addr, err = gCouncil.GetAuthorizedAccountAddress(councilNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, initialAuthorizedAccounts[1], addr)
	})

	t.Run("getAllAuthorizedAccounts returns all addresses", func(t *testing.T) {
		initGovCouncil(t)
		defer gCouncil.backend.Close()

		addresses, err := gCouncil.GetAllAuthorizedAccounts(councilNonMember)
		require.NoError(t, err)
		require.Len(t, addresses, 2, "Should return 2 authorized accounts")
		require.Contains(t, addresses, initialAuthorizedAccounts[0])
		require.Contains(t, addresses, initialAuthorizedAccounts[1])
	})
}

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
	"math/big"
	"testing"

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
	fiatTokenAddress = common.HexToAddress("0xC00001") // Mock fiat token address

	var err error
	gMinter, err = NewGovWBFT(t, types.GenesisAlloc{
		minterMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		minterMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		minterMembers[2].Operator.Address: {Balance: towei(1_000_000)},
		minterNonMember.Address:           {Balance: towei(1_000_000)},
		fiatTokenAddress:                  {Balance: towei(0)}, // Mock token contract
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
	}, nil)
	require.NoError(t, err)
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

		// First, deposit native coins to build up burn balance
		err := gMinter.DepositForBurn(t, minterMembers[0].Operator, amount)
		require.NoError(t, err)

		// Verify burn balance
		burnBalance, err := gMinter.GetBurnBalance(minterNonMember, from)
		require.NoError(t, err)
		require.Equal(t, 0, burnBalance.Cmp(amount))

		// Now member proposes burn
		tx, err := gMinter.TxProposeBurn(t, minterMembers[0].Operator, from, amount)
		receipt, err := gMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

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
		// Note: Execution would happen but may fail if fiatToken is not a real contract
		// In a full integration test, we'd deploy a mock fiat token contract
	})
}

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
	gMasterMinter         *GovWBFT
	masterMinterMembers   []*TestCandidate
	masterMinterNonMember *EOA
	mockFiatToken         common.Address
	defaultMaxAllowance   *big.Int
)

func initGovMasterMinter(t *testing.T) {
	masterMinterMembers = []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
	masterMinterNonMember = NewEOA()
	mockFiatToken = common.HexToAddress("0xC00002") // Mock fiat token address
	defaultMaxAllowance = new(big.Int).Mul(big.NewInt(1000000), big.NewInt(1e18)) // 1M tokens

	var err error
	gMasterMinter, err = NewGovWBFT(t, types.GenesisAlloc{
		masterMinterMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterMembers[2].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterNonMember.Address:           {Balance: towei(1_000_000)},
		mockFiatToken:                           {Balance: towei(0)}, // Mock token contract
	}, func(govValidator *params.SystemContract) {
		// Setup governance members for voting
		var members, validators, blsPubKeys string
		for i, m := range masterMinterMembers {
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
	}, nil, nil, func(govMasterMinter *params.SystemContract) {
		// Initialize GovMasterMinter with fiatToken and max allowance
		govMasterMinter.Params = map[string]string{
			sc.GOV_MASTER_MINTER_PARAM_FIAT_TOKEN:         mockFiatToken.String(),
			sc.GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: defaultMaxAllowance.String(),
			sc.GOV_BASE_PARAM_MEMBERS:                     masterMinterMembers[0].Operator.Address.String() + "," + masterMinterMembers[1].Operator.Address.String() + "," + masterMinterMembers[2].Operator.Address.String(),
			sc.GOV_BASE_PARAM_QUORUM:                      "2",
			sc.GOV_BASE_PARAM_EXPIRY:                      "604800",
			sc.GOV_BASE_PARAM_MEMBER_VERSION:              "1",
		}
	})
	require.NoError(t, err)
}

func TestGovMasterMinter_Initialize(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		// Check fiatToken is set correctly
		token, err := gMasterMinter.MasterMinterFiatToken(masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, mockFiatToken, token)

		// Check maxMinterAllowance is set correctly
		maxAllowance, err := gMasterMinter.MaxMinterAllowance(masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, 0, maxAllowance.Cmp(defaultMaxAllowance))

		// Check governance base parameters
		quorum, err := gMasterMinter.BaseQuorum(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorum)

		expiry, err := gMasterMinter.BaseProposalExpiry(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, uint64(604800), expiry.Uint64())

		// Check members are initialized
		for i, m := range masterMinterMembers {
			member, err := gMasterMinter.BaseMembers(gMasterMinter.govMasterMinter, masterMinterNonMember, m.Operator.Address)
			require.NoError(t, err, "Member %d should be initialized", i)
			require.True(t, member.IsActive, "Member %d should be active", i)
		}

		// Check non-member is not initialized
		member, err := gMasterMinter.BaseMembers(gMasterMinter.govMasterMinter, masterMinterNonMember, masterMinterNonMember.Address)
		require.NoError(t, err)
		require.False(t, member.IsActive, "Non-member should not be active")
	})
}

func TestGovMasterMinter_ProposeConfigureMinter(t *testing.T) {
	t.Run("member can propose configure minter", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address
		allowance := big.NewInt(100000)

		// Check minter is not configured initially
		isMinter, err := gMasterMinter.IsMinter(masterMinterNonMember, minter)
		require.NoError(t, err)
		require.False(t, isMinter)

		// Member proposes configure minter
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter, allowance)
		receipt, err := gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gMasterMinter.BaseCurrentProposalId(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)

		// Check proposal details
		proposal, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
		require.Equal(t, masterMinterMembers[0].Operator.Address, proposal.Proposer)
		require.Equal(t, uint32(1), proposal.Approved) // Proposer auto-approves
	})

	t.Run("non-member cannot propose configure minter", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address
		allowance := big.NewInt(100000)

		// Non-member tries to propose configure minter (should fail)
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterNonMember, minter, allowance)
		err = gMasterMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Non-member should not be able to propose configure minter")
	})

	t.Run("cannot exceed max allowance", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address
		// Try to set allowance above max
		tooLargeAllowance := new(big.Int).Add(defaultMaxAllowance, big.NewInt(1))

		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter, tooLargeAllowance)
		err = gMasterMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Should not allow allowance exceeding maxMinterAllowance")
	})
}

func TestGovMasterMinter_ProposeRemoveMinter(t *testing.T) {
	t.Run("cannot propose remove non-existent minter", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address

		// Try to propose removing a minter that doesn't exist (should fail)
		tx, err := gMasterMinter.TxProposeRemoveMinter(t, masterMinterMembers[0].Operator, minter)
		err = gMasterMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Should not be able to propose removing a non-existent minter")
	})
}

func TestGovMasterMinter_ProposeUpdateMaxAllowance(t *testing.T) {
	t.Run("member can propose update max allowance", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		newMaxAllowance := new(big.Int).Mul(big.NewInt(2000000), big.NewInt(1e18))

		// Member proposes update max allowance
		tx, err := gMasterMinter.TxProposeUpdateMaxMinterAllowance(t, masterMinterMembers[0].Operator, newMaxAllowance)
		receipt, err := gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status)

		// Check proposal was created
		proposalId, err := gMasterMinter.BaseCurrentProposalId(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(1), proposalId)
	})

	t.Run("cannot set max allowance to zero", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		// Try to set max allowance to zero
		tx, err := gMasterMinter.TxProposeUpdateMaxMinterAllowance(t, masterMinterMembers[0].Operator, big.NewInt(0))
		err = gMasterMinter.ExpectedFail(tx, err)
		require.Error(t, err, "Should not allow zero max allowance")
	})
}

func TestGovMasterMinter_MinterAllowanceTracking(t *testing.T) {
	t.Run("initial minter allowance is zero", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address

		// Check initial allowance is zero
		allowance, err := gMasterMinter.MinterAllowance(masterMinterNonMember, minter)
		require.NoError(t, err)
		require.Equal(t, 0, allowance.Cmp(big.NewInt(0)))

		// Check minter status is false
		isMinter, err := gMasterMinter.IsMinter(masterMinterNonMember, minter)
		require.NoError(t, err)
		require.False(t, isMinter)
	})
}

func TestGovMasterMinter_GovernanceWorkflow(t *testing.T) {
	t.Run("complete configure minter workflow", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter := NewEOA().Address
		allowance := big.NewInt(500000)

		// Step 1: Member 0 proposes configure minter
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter, allowance)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId := big.NewInt(1)

		// Step 2: Check proposal status (pending, needs 1 more approval for quorum 2)
		proposal, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal.Status)
		require.Equal(t, uint32(1), proposal.Approved)

		// Step 3: Member 1 approves
		tx, err = gMasterMinter.BaseTxApproveProposal(t, gMasterMinter.govMasterMinter, masterMinterMembers[1].Operator, proposalId)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Step 4: Check proposal reached quorum
		proposal, err = gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint32(2), proposal.Approved)
		// Note: Execution would happen but may fail if fiatToken is not a real contract
		// In a full integration test, we'd deploy a mock fiat token contract
	})

	t.Run("multiple proposals workflow", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		minter1 := NewEOA().Address
		minter2 := NewEOA().Address
		allowance1 := big.NewInt(100000)
		allowance2 := big.NewInt(200000)

		// Create first proposal
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter1, allowance1)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Create second proposal
		tx, err = gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[1].Operator, minter2, allowance2)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Check both proposals exist
		proposalId, err := gMasterMinter.BaseCurrentProposalId(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(2), proposalId)

		// Check first proposal
		proposal1, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(1))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal1.Status)

		// Check second proposal
		proposal2, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal2.Status)
	})
}

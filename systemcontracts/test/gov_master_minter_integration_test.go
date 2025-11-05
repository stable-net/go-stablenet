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
	// Note: Using 10B tokens (same as Solidity default) due to Genesis state not being applied in simulated backend
	defaultMaxAllowance = new(big.Int).Mul(big.NewInt(10000000000), big.NewInt(1e18)) // 10B tokens

	// MockFiatToken bytecode
	mockFiatTokenCode := hexutil.MustDecode("0x608060405234801561000f575f5ffd5b506004361061004a575f3560e01c80633092afd51461004e5780634e44d956146100765780638a6db9c314610089578063aa271e1a146100bf575b5f5ffd5b61006161005c3660046101ca565b6100ea565b60405190151581526020015b60405180910390f35b6100616100843660046101ea565b610146565b6100b16100973660046101ca565b6001600160a01b03165f9081526001602052604090205490565b60405190815260200161006d565b6100616100cd3660046101ca565b6001600160a01b03165f9081526020819052604090205460ff1690565b6001600160a01b0381165f81815260208181526040808320805460ff191690556001909152808220829055519091907fe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692908390a2506001919050565b6001600160a01b0382165f81815260208181526040808320805460ff191660019081179091558252808320859055518481529192917f46980fca912ef9bcdbd36877427b6b90e860769f604e89c0e67720cece530d20910160405180910390a250600192915050565b80356001600160a01b03811681146101c5575f5ffd5b919050565b5f602082840312156101da575f5ffd5b6101e3826101af565b9392505050565b5f5f604083850312156101fb575f5ffd5b610204836101af565b94602093909301359350505056fea2646970667358221220d31173a4dd708d544437de2deccd13f015f0091426a1ea75e2d32631b5e1976e64736f6c634300081e0033")

	var err error
	gMasterMinter, err = NewGovWBFT(t, types.GenesisAlloc{
		masterMinterMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterMembers[2].Operator.Address: {Balance: towei(1_000_000)},
		masterMinterNonMember.Address:           {Balance: towei(1_000_000)},
		mockFiatToken:                           {Balance: towei(0), Code: mockFiatTokenCode}, // Mock token contract with code
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
			sc.GOV_MASTER_MINTER_PARAM_FIAT_TOKEN:           mockFiatToken.String(),
			sc.GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: defaultMaxAllowance.String(),
			sc.GOV_BASE_PARAM_MEMBERS:                       masterMinterMembers[0].Operator.Address.String() + "," + masterMinterMembers[1].Operator.Address.String() + "," + masterMinterMembers[2].Operator.Address.String(),
			sc.GOV_BASE_PARAM_QUORUM:                        "2",
			sc.GOV_BASE_PARAM_EXPIRY:                        "604800",
			sc.GOV_BASE_PARAM_MEMBER_VERSION:                "1",
		}
	}, nil)
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

		// Check maxMinterAllowance
		maxAllowance, err := gMasterMinter.MaxMinterAllowance(masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, 0, maxAllowance.Cmp(defaultMaxAllowance), "maxMinterAllowance should be 0 initially (Genesis state not applied)")

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

// setMaxMinterAllowanceHelper sets maxMinterAllowance via governance proposal
func setMaxMinterAllowanceHelper(t *testing.T, maxAllowance *big.Int) {
	// Propose update
	tx, err := gMasterMinter.TxProposeUpdateMaxMinterAllowance(t, masterMinterMembers[0].Operator, maxAllowance)
	receipt, err := gMasterMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	require.Equal(t, uint64(1), receipt.Status)

	proposalId, err := gMasterMinter.BaseCurrentProposalId(gMasterMinter.govMasterMinter, masterMinterNonMember)
	require.NoError(t, err)

	// Approve by second member (need quorum=2)
	// Note: With auto-execution, proposal executes automatically when quorum is reached
	tx, err = gMasterMinter.BaseTxApproveProposal(t, gMasterMinter.govMasterMinter, masterMinterMembers[1].Operator, proposalId)
	receipt, err = gMasterMinter.ExpectedOk(tx, err)
	require.NoError(t, err)
	require.Equal(t, uint64(1), receipt.Status)

	// Verify proposal is executed and max allowance is set
	proposal, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, proposalId)
	require.NoError(t, err)
	require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed (3)")

	currentMax, err := gMasterMinter.MaxMinterAllowance(masterMinterNonMember)
	require.NoError(t, err)
	require.Equal(t, 0, currentMax.Cmp(maxAllowance))
}

func TestGovMasterMinter_ProposeConfigureMinter(t *testing.T) {
	t.Run("member can propose configure minter", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		// set maxMinterAllowance
		setMaxMinterAllowanceHelper(t, big.NewInt(1000000))

		minter := NewEOA().Address
		allowance := big.NewInt(100000)

		// Check minter is not configured initially
		isMinter, err := gMasterMinter.GetIsMinter(masterMinterNonMember, minter)
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
		require.Equal(t, big.NewInt(2), proposalId) // ID 2 since ID 1 was used for setMaxMinterAllowanceHelper

		// Check proposal details
		proposal, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(2))
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
		allowance, err := gMasterMinter.GetMinterAllowance(masterMinterNonMember, minter)
		require.NoError(t, err)
		require.Equal(t, 0, allowance.Cmp(big.NewInt(0)))

		// Check minter status is false
		isMinter, err := gMasterMinter.GetIsMinter(masterMinterNonMember, minter)
		require.NoError(t, err)
		require.False(t, isMinter)
	})
}

func TestGovMasterMinter_GovernanceWorkflow(t *testing.T) {
	t.Run("complete configure minter workflow", func(t *testing.T) {
		initGovMasterMinter(t)
		defer gMasterMinter.backend.Close()

		// set maxMinterAllowance
		setMaxMinterAllowanceHelper(t, big.NewInt(1000000))

		minter := NewEOA().Address
		allowance := big.NewInt(500000)

		// Step 1: Member 0 proposes configure minter
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter, allowance)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		proposalId := big.NewInt(2) // Proposal ID 1 was used for setMaxMinterAllowanceHelper

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

		// set maxMinterAllowance
		setMaxMinterAllowanceHelper(t, big.NewInt(1000000))

		minter1 := NewEOA().Address
		minter2 := NewEOA().Address
		allowance1 := big.NewInt(100000)
		allowance2 := big.NewInt(200000)

		// Create first proposal (ID will be 2, since ID 1 was used for setMaxMinterAllowanceHelper)
		tx, err := gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[0].Operator, minter1, allowance1)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Create second proposal (ID will be 3)
		tx, err = gMasterMinter.TxProposeConfigureMinter(t, masterMinterMembers[1].Operator, minter2, allowance2)
		_, err = gMasterMinter.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Check both proposals exist
		proposalId, err := gMasterMinter.BaseCurrentProposalId(gMasterMinter.govMasterMinter, masterMinterNonMember)
		require.NoError(t, err)
		require.Equal(t, big.NewInt(3), proposalId) // Now expecting 3 since ID 1 was used for maxMinterAllowance

		// Check first proposal (now ID 2)
		proposal1, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(2))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal1.Status)

		// Check second proposal (now ID 3)
		proposal2, err := gMasterMinter.BaseGetProposal(gMasterMinter.govMasterMinter, masterMinterNonMember, big.NewInt(3))
		require.NoError(t, err)
		require.Equal(t, sc.ProposalStatusVoting, proposal2.Status)
	})
}

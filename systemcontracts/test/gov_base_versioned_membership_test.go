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

// TestGovBase_VersionedMembershipScenario tests the critical scenario where:
// 1. Member B creates Proposal 1 at memberVersion 1
// 2. Member B proposes to remove Member A (Proposal 2)
// 3. Member C votes to approve removal → Member A is removed (memberVersion becomes 2)
// 4. Member A should still be able to vote on Proposal 1 (created at version 1)
// 5. Member C creates Proposal 3 at memberVersion 2
// 6. Member A should NOT be able to create new proposals (not active)
// 7. Member A should NOT be able to vote on Proposal 3 (created at version 2)
func TestGovBase_VersionedMembershipScenario(t *testing.T) {
	// Setup: 4 members initially (A, B, C, D), quorum = 2
	memberA := NewTestCandidate()
	memberB := NewTestCandidate()
	memberC := NewTestCandidate()
	memberD := NewTestCandidate()
	members := []*TestCandidate{memberA, memberB, memberC, memberD}

	mockFiatToken := common.HexToAddress("0xC00002")
	mockFiatTokenCode := hexutil.MustDecode("0x608060405234801561000f575f5ffd5b506004361061004a575f3560e01c80633092afd51461004e5780634e44d956146100765780638a6db9c314610089578063aa271e1a146100bf575b5f5ffd5b61006161005c3660046101ca565b6100ea565b60405190151581526020015b60405180910390f35b6100616100843660046101ea565b610146565b6100b16100973660046101ca565b6001600160a01b03165f9081526001602052604090205490565b60405190815260200161006d565b6100616100cd3660046101ca565b6001600160a01b03165f9081526020819052604090205460ff1690565b6001600160a01b0381165f81815260208181526040808320805460ff191690556001909152808220829055519091907fe94479a9f7e1952cc78f2d6baab678adc1b772d936c6583def489e524cb66692908390a2506001919050565b6001600160a01b0382165f81815260208181526040808320805460ff191660019081179091558252808320859055518481529192917f46980fca912ef9bcdbd36877427b6b90e860769f604e89c0e67720cece530d20910160405180910390a250600192915050565b80356001600160a01b03811681146101c5575f5ffd5b919050565b5f602082840312156101da575f5ffd5b6101e3826101af565b9392505050565b5f5f604083850312156101fb575f5ffd5b610204836101af565b94602093909301359350505056fea2646970667358221220d31173a4dd708d544437de2deccd13f015f0091426a1ea75e2d32631b5e1976e64736f6c634300081e0033")
	defaultMaxAllowance := new(big.Int).Mul(big.NewInt(10000000000), big.NewInt(1e18))

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		memberA.Operator.Address: {Balance: towei(1_000_000)},
		memberB.Operator.Address: {Balance: towei(1_000_000)},
		memberC.Operator.Address: {Balance: towei(1_000_000)},
		memberD.Operator.Address: {Balance: towei(1_000_000)},
		mockFiatToken:            {Balance: towei(0), Code: mockFiatTokenCode},
	}, func(govValidator *params.SystemContract) {
		var memberAddrs, validators, blsPubKeys string
		for i, m := range members {
			if i > 0 {
				memberAddrs += ","
				validators += ","
				blsPubKeys += ","
			}
			memberAddrs += m.Operator.Address.String()
			validators += m.Validator.Address.String()
			blsPubKeys += hexutil.Encode(m.GetBLSPublicKey(t).Marshal())
		}
		govValidator.Params = map[string]string{
			"members":       memberAddrs,
			"quorum":        "2",
			"expiry":        "604800",
			"memberVersion": "1",
			"validators":    validators,
			"blsPublicKeys": blsPubKeys,
		}
	}, nil, nil, func(govMasterMinter *params.SystemContract) {
		govMasterMinter.Params = map[string]string{
			sc.GOV_MASTER_MINTER_PARAM_FIAT_TOKEN:           mockFiatToken.String(),
			sc.GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: defaultMaxAllowance.String(),
			sc.GOV_BASE_PARAM_MEMBERS:                       memberA.Operator.Address.String() + "," + memberB.Operator.Address.String() + "," + memberC.Operator.Address.String() + "," + memberD.Operator.Address.String(),
			sc.GOV_BASE_PARAM_QUORUM:                        "2",
			sc.GOV_BASE_PARAM_EXPIRY:                        "604800",
			sc.GOV_BASE_PARAM_MEMBER_VERSION:                "1",
		}
	}, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	// Verify initial state: memberVersion = 1, all 4 members active
	version, err := g.BaseMemberVersion(g.govMasterMinter, memberA.Operator)
	require.NoError(t, err)
	require.Equal(t, uint64(1), version.Uint64(), "Initial memberVersion should be 1")

	// First, we need to update maxMinterAllowance since simulated backend doesn't initialize it from genesis
	// This is a known limitation mentioned in GovMasterMinter.sol line 61
	t.Log("Setup: Update maxMinterAllowance via governance proposal")
	proposalId0 := big.NewInt(1)
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberA.Operator, nil), "proposeUpdateMaxMinterAllowance", defaultMaxAllowance))
	require.NoError(t, err)
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberB.Operator, nil), "approveProposal", proposalId0))
	require.NoError(t, err)

	// Step 1: Member B creates Proposal 1 at memberVersion 1
	t.Log("Step 1: Member B creates Proposal 1 at memberVersion 1")
	minter1 := common.HexToAddress("0x0000000000000000000000000000000000001001")
	minterAllowance := big.NewInt(10_000_000)

	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberB.Operator, nil), "proposeConfigureMinter", minter1, minterAllowance))
	require.NoError(t, err, "Member B should be able to create proposal at version 1")

	proposalIdByMemberB := big.NewInt(2)
	proposalByB, err := g.BaseGetProposal(g.govMasterMinter, memberA.Operator, proposalIdByMemberB)
	require.NoError(t, err)
	require.Equal(t, sc.ProposalStatusVoting, proposalByB.Status, "Proposal by Member B should be in Voting status")
	require.Equal(t, uint64(1), proposalByB.MemberVersion.Uint64(), "Proposal by Member B should capture memberVersion 1")

	// Step 2: Member B proposes to remove Member A
	t.Log("Step 2: Member B proposes to remove Member A")
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberB.Operator, nil), "proposeRemoveMember", memberA.Operator.Address, uint32(3)))
	require.NoError(t, err)

	proposalRemoveA := big.NewInt(3)
	removalProposal, err := g.BaseGetProposal(g.govMasterMinter, memberB.Operator, proposalRemoveA)
	require.NoError(t, err)
	require.Equal(t, sc.ProposalStatusVoting, removalProposal.Status, "Removal proposal should be in Voting status")

	// Step 3: Member C votes to approve removal → Member A is removed (memberVersion becomes 2)
	t.Log("Step 3: Member C votes to approve removal → Member A is removed")
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberC.Operator, nil), "approveProposal", proposalRemoveA))
	require.NoError(t, err, "Member C should be able to approve removal proposal")

	// Verify member A is removed and memberVersion incremented
	version, err = g.BaseMemberVersion(g.govMasterMinter, memberB.Operator)
	require.NoError(t, err)
	require.Equal(t, uint64(2), version.Uint64(), "MemberVersion should be 2 after Member A removal")

	memberAInfo, err := g.BaseMembers(g.govMasterMinter, memberB.Operator, memberA.Operator.Address)
	require.NoError(t, err)
	require.False(t, memberAInfo.IsActive, "Member A should not be active after removal")

	// Step 4: Member A should still be able to vote on Proposal 1 (created by Member B at version 1)
	t.Log("Step 4: Member A votes on Member B's Proposal 1 (should succeed - proposal at version 1)")
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberA.Operator, nil), "approveProposal", proposalIdByMemberB))
	require.NoError(t, err, "Member A should be able to vote on Proposal 1 created at version 1")

	proposalByB, err = g.BaseGetProposal(g.govMasterMinter, memberB.Operator, proposalIdByMemberB)
	require.NoError(t, err)
	require.Equal(t, sc.ProposalStatusExecuted, proposalByB.Status, "Proposal 1 should be Executed after reaching quorum")

	// Step 5: Member C creates Proposal 3 at memberVersion 2
	t.Log("Step 5: Member C creates Proposal 3 at memberVersion 2")
	minter3 := common.HexToAddress("0x0000000000000000000000000000000000001003")
	_, err = g.ExpectedOk(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberC.Operator, nil), "proposeConfigureMinter", minter3, minterAllowance))
	require.NoError(t, err, "Member C should be able to create proposal at version 2")

	proposalIdByMemberC := big.NewInt(4)
	proposalByC, err := g.BaseGetProposal(g.govMasterMinter, memberB.Operator, proposalIdByMemberC)
	require.NoError(t, err)
	require.Equal(t, sc.ProposalStatusVoting, proposalByC.Status, "Proposal by Member C should be in Voting status")
	require.Equal(t, uint64(2), proposalByC.MemberVersion.Uint64(), "Proposal by Member C should capture memberVersion 2")

	// Step 6: Member A should NOT be able to create new proposals (not active)
	t.Log("Step 6: Member A tries to create new proposal (should fail - not active member)")
	minter2 := common.HexToAddress("0x0000000000000000000000000000000000001002")
	ExpectedRevert(t,
		g.ExpectedFail(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberA.Operator, nil), "proposeConfigureMinter", minter2, minterAllowance)),
		"NotAMember",
	)

	// Step 7: Member A should NOT be able to vote on Proposal 3 (created by Member C at version 2)
	t.Log("Step 7: Member A tries to vote on Member C's Proposal 3 (should fail - not in version 2 snapshot)")
	ExpectedRevert(t,
		g.ExpectedFail(g.govMasterMinter.Transact(NewTxOptsWithValue(t, memberA.Operator, nil), "approveProposal", proposalIdByMemberC)),
		"NotAMember",
	)

	t.Log("✓ Versioned membership works correctly:")
	t.Log("  - Removed member CAN vote on proposals created before removal")
	t.Log("  - Removed member CANNOT create new proposals")
	t.Log("  - Removed member CANNOT vote on proposals created after removal")
}

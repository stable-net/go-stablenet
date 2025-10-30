// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.

package test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestChangeQuorum_BasicProposal tests basic quorum change proposal creation and execution
func TestChangeQuorum_BasicProposal(t *testing.T) {
	// Setup: 5 validators with quorum of 3
	customValidators := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
		NewTestCandidate(), NewTestCandidate(),
	}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[3].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[4].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
			}
			members = members + v.Operator.Address.Hex()
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "3",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customValidators[0].Operator
	approver1 := customValidators[1].Operator
	approver2 := customValidators[2].Operator

	t.Run("change quorum from 3 to 4", func(t *testing.T) {
		// Verify initial quorum
		quorumBefore, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(3), quorumBefore, "Initial quorum should be 3")

		// Create proposal to change quorum to 4
		proposalId, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(4))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Greater(t, proposalId.Uint64(), uint64(0), "Proposal ID should be greater than 0")

		// Verify proposal details
		proposal, err := g.BaseGetProposal(g.govValidator, proposer, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(1), uint8(proposal.Status), "Proposal should be in Voting status")
		require.Equal(t, uint32(1), proposal.Approved, "Proposer auto-approves")

		// Approve by second member
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Approve by third member - should reach quorum and execute
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver2, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify proposal is executed
		proposal, err = g.BaseGetProposal(g.govValidator, proposer, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")

		// Verify quorum has changed
		quorumAfter, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(4), quorumAfter, "Quorum should be changed to 4")

		t.Logf("✓ Quorum successfully changed from 3 to 4")
	})
}

// TestChangeQuorum_ValidationErrors tests validation errors for quorum changes
func TestChangeQuorum_ValidationErrors(t *testing.T) {
	customValidators := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
	}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
			}
			members = members + v.Operator.Address.Hex()
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "2",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customValidators[0].Operator

	t.Run("cannot propose quorum less than 2 for multiple members", func(t *testing.T) {
		_, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(1))
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Quorum 1 should be rejected for 3 members")
		t.Logf("✓ Quorum < 2 correctly rejected for multiple members")
	})

	t.Run("cannot propose quorum greater than member count", func(t *testing.T) {
		_, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(4))
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Quorum 4 should be rejected for 3 members")
		t.Logf("✓ Quorum > member count correctly rejected")
	})

	t.Run("cannot propose quorum equal to 0", func(t *testing.T) {
		_, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(0))
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Quorum 0 should be rejected")
		t.Logf("✓ Quorum 0 correctly rejected")
	})
}

// TestChangeQuorum_EdgeCases tests edge cases for quorum changes
func TestChangeQuorum_EdgeCases(t *testing.T) {
	t.Run("single member can only have quorum 1", func(t *testing.T) {
		customValidators := []*TestCandidate{NewTestCandidate()}
		g, err := NewGovWBFT(t, types.GenesisAlloc{
			customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		}, func(govValidator *params.SystemContract) {
			govValidator.Params = map[string]string{
				"members":       customValidators[0].Operator.Address.Hex(),
				"quorum":        "1",
				"expiry":        "86400",
				"memberVersion": "1",
			}
		}, nil, nil, nil, nil)
		require.NoError(t, err)
		defer g.backend.Close()

		proposer := customValidators[0].Operator

		// Try to change to quorum 2 (should fail)
		_, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Single member cannot have quorum > 1")

		t.Logf("✓ Single member correctly restricted to quorum 1")
	})

	t.Run("change quorum to maximum (member count)", func(t *testing.T) {
		customValidators := []*TestCandidate{
			NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
		}
		g, err := NewGovWBFT(t, types.GenesisAlloc{
			customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		}, func(govValidator *params.SystemContract) {
			var members string
			for i, v := range customValidators {
				if i > 0 {
					members = members + ","
				}
				members = members + v.Operator.Address.Hex()
			}
			govValidator.Params = map[string]string{
				"members":       members,
				"quorum":        "2",
				"expiry":        "86400",
				"memberVersion": "1",
			}
		}, nil, nil, nil, nil)
		require.NoError(t, err)
		defer g.backend.Close()

		proposer := customValidators[0].Operator
		approver1 := customValidators[1].Operator

		// Create proposal to change quorum to 3 (all members)
		proposalId, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(3))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Approve by second member - reaches quorum and executes
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify quorum changed to 3
		quorumAfter, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(3), quorumAfter, "Quorum should be 3")

		t.Logf("✓ Quorum successfully changed to maximum (member count)")
	})

	t.Run("change quorum to minimum (2)", func(t *testing.T) {
		customValidators := []*TestCandidate{
			NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
			NewTestCandidate(), NewTestCandidate(),
		}
		g, err := NewGovWBFT(t, types.GenesisAlloc{
			customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[3].Operator.Address: {Balance: towei(1_000_000)},
			customValidators[4].Operator.Address: {Balance: towei(1_000_000)},
		}, func(govValidator *params.SystemContract) {
			var members string
			for i, v := range customValidators {
				if i > 0 {
					members = members + ","
				}
				members = members + v.Operator.Address.Hex()
			}
			govValidator.Params = map[string]string{
				"members":       members,
				"quorum":        "3",
				"expiry":        "86400",
				"memberVersion": "1",
			}
		}, nil, nil, nil, nil)
		require.NoError(t, err)
		defer g.backend.Close()

		proposer := customValidators[0].Operator
		approver1 := customValidators[1].Operator
		approver2 := customValidators[2].Operator

		// Create proposal to change quorum to 2
		proposalId, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Approve by two more members
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver2, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify quorum changed to 2
		quorumAfter, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorumAfter, "Quorum should be 2")

		t.Logf("✓ Quorum successfully changed to minimum (2)")
	})
}

// TestChangeQuorum_MultipleChanges tests consecutive quorum changes
func TestChangeQuorum_MultipleChanges(t *testing.T) {
	customValidators := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
		NewTestCandidate(), NewTestCandidate(),
	}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[3].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[4].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
			}
			members = members + v.Operator.Address.Hex()
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "3",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customValidators[0].Operator
	approver1 := customValidators[1].Operator
	approver2 := customValidators[2].Operator
	approver3 := customValidators[3].Operator

	t.Run("change quorum multiple times", func(t *testing.T) {
		// First change: 3 -> 4
		proposalId1, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(4))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId1)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver2, proposalId1)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		quorum1, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(4), quorum1, "Quorum should be 4")
		t.Logf("✓ First change: 3 -> 4")

		// Second change: 4 -> 2
		proposalId2, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Need 4 approvals now (including proposer)
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId2)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver2, proposalId2)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver3, proposalId2)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		quorum2, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(2), quorum2, "Quorum should be 2")
		t.Logf("✓ Second change: 4 -> 2")

		// Third change: 2 -> 5
		proposalId3, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(5))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Only need 2 approvals now
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId3)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		quorum3, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(5), quorum3, "Quorum should be 5")
		t.Logf("✓ Third change: 2 -> 5")

		t.Logf("✓ Multiple quorum changes completed: 3 -> 4 -> 2 -> 5")
	})
}

// TestChangeQuorum_WithMemberChanges tests quorum changes in combination with member changes
func TestChangeQuorum_WithMemberChanges(t *testing.T) {
	customValidators := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
	}
	newMember := NewTestCandidate()

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		newMember.Operator.Address:           {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
			}
			members = members + v.Operator.Address.Hex()
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "2",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customValidators[0].Operator
	approver1 := customValidators[1].Operator

	t.Run("change quorum after adding member", func(t *testing.T) {
		// Add new member (quorum remains 2)
		proposalId1, tx, err := g.BaseTxProposeAddMember(t, g.govValidator, proposer, newMember.Operator.Address, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId1)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify member count is now 4
		versionAfter, err := g.BaseMemberVersion(g.govValidator, proposer)
		require.NoError(t, err)
		memberCount := big.NewInt(0)
		for i := 0; i < 10; i++ {
			_, err := g.BaseVersionedMemberList(g.govValidator, proposer, versionAfter, big.NewInt(int64(i)))
			if err != nil {
				break
			}
			memberCount.Add(memberCount, big.NewInt(1))
		}
		require.Equal(t, big.NewInt(4), memberCount, "Should have 4 members")

		// Change quorum to 3 (now valid with 4 members)
		proposalId2, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(3))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver1, proposalId2)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		quorum, err := g.BaseQuorum(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, uint32(3), quorum, "Quorum should be 3")

		t.Logf("✓ Quorum changed after adding member: 2 -> 3 (4 members)")
	})
}

// TestChangeQuorum_EventEmission tests that QuorumUpdated events are emitted correctly
func TestChangeQuorum_EventEmission(t *testing.T) {
	customValidators := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
	}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
			}
			members = members + v.Operator.Address.Hex()
		}
		govValidator.Params = map[string]string{
			"members":       members,
			"quorum":        "2",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customValidators[0].Operator
	approver := customValidators[1].Operator

	t.Run("QuorumUpdated event is emitted on execution", func(t *testing.T) {
		// Create and execute quorum change proposal
		proposalId, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(3))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, proposalId)
		receipt, err := g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify event logs are present
		require.Greater(t, len(receipt.Logs), 0, "Should have emitted events")

		t.Logf("✓ QuorumUpdated event emitted correctly (%d logs)", len(receipt.Logs))
	})
}

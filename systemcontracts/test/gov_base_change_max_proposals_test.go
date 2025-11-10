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

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestChangeMaxProposals_BasicProposal tests basic maxProposals change proposal creation and execution
func TestChangeMaxProposals_BasicProposal(t *testing.T) {
	// Setup: 3 members with quorum of 2, default maxProposals = 3
	customMembers := []*TestCandidate{
		NewTestCandidate(), NewTestCandidate(), NewTestCandidate(),
	}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[2].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customMembers {
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
			"gasTip":        "1",
		}
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customMembers[0].Operator
	approver := customMembers[1].Operator

	t.Run("change maxProposals from 3 to 10", func(t *testing.T) {
		// Verify initial value
		maxBefore, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, int64(3), maxBefore.Int64(), "Initial maxProposals should be 3")

		// Create proposal to change maxProposals to 10
		proposalId, tx, err := g.BaseTxProposeChangeMaxProposals(t, g.govValidator, proposer, big.NewInt(10))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Greater(t, proposalId.Uint64(), uint64(0), "Proposal ID should be greater than 0")

		// Verify proposal details
		proposal, err := g.BaseGetProposal(g.govValidator, proposer, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(1), uint8(proposal.Status), "Proposal should be in Voting status")
		require.Equal(t, uint32(1), proposal.Approved, "Proposer auto-approves")

		// Approve by second member - should reach quorum and execute
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, proposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify proposal is executed
		proposal, err = g.BaseGetProposal(g.govValidator, proposer, proposalId)
		require.NoError(t, err)
		require.Equal(t, uint8(3), uint8(proposal.Status), "Proposal should be Executed")

		// Verify maxProposals was updated
		maxAfter, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, int64(10), maxAfter.Int64(), "maxProposals should be updated to 10")
	})
}

// TestChangeMaxProposals_InvalidValues tests validation of maxProposals range
func TestChangeMaxProposals_InvalidValues(t *testing.T) {
	customMembers := []*TestCandidate{NewTestCandidate(), NewTestCandidate()}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		govValidator.Params = map[string]string{
			"members":       customMembers[0].Operator.Address.Hex() + "," + customMembers[1].Operator.Address.Hex(),
			"quorum":        "2",
			"expiry":        "86400",
			"memberVersion": "1",
		}
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customMembers[0].Operator

	testCases := []struct {
		name     string
		value    int64
		expected string
	}{
		{"zero value", 0, "InvalidMaxProposals"},
		{"below minimum", -1, "InvalidMaxProposals"},
		{"above maximum", 51, "InvalidMaxProposals"},
		{"far above maximum", 100, "InvalidMaxProposals"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, tx, err := g.BaseTxProposeChangeMaxProposals(t, g.govValidator, proposer, big.NewInt(tc.value))
			err = g.ExpectedFail(tx, err)
			ExpectedRevert(t, err, tc.expected)
		})
	}
}

// TestChangeMaxProposals_BoundaryValues tests valid boundary values
func TestChangeMaxProposals_BoundaryValues(t *testing.T) {
	customMembers := []*TestCandidate{NewTestCandidate(), NewTestCandidate()}
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		govValidator.Params = map[string]string{
			"members":       customMembers[0].Operator.Address.Hex() + "," + customMembers[1].Operator.Address.Hex(),
			"quorum":        "2",
			"expiry":        "86400",
			"memberVersion": "1",
			"gasTip":        "5000000000000",
		}
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customMembers[0].Operator
	approver := customMembers[1].Operator

	testCases := []struct {
		name  string
		value int64
	}{
		{"minimum value 1", 1},
		{"maximum value 50", 50},
		{"mid-range value 25", 25},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proposalId, tx, err := g.BaseTxProposeChangeMaxProposals(t, g.govValidator, proposer, big.NewInt(tc.value))
			_, err = g.ExpectedOk(tx, err)
			require.NoError(t, err)

			// Approve to execute
			tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, proposalId)
			_, err = g.ExpectedOk(tx, err)
			require.NoError(t, err)

			// Verify the value was set
			maxAfter, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, proposer)
			require.NoError(t, err)
			require.Equal(t, tc.value, maxAfter.Int64(), "maxProposals should be updated to %d", tc.value)
		})
	}
}

// TestChangeMaxProposals_LimitEnforcement tests that the limit is actually enforced
func TestChangeMaxProposals_LimitEnforcement(t *testing.T) {
	customMembers := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}

	// Initialize with maxProposals=2 for easier testing
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[2].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customMembers {
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
			"maxProposals":  "2", // Set limit to 2
		}
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customMembers[0].Operator

	t.Run("can create up to limit", func(t *testing.T) {
		// Verify initial value
		maxProposals, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, int64(2), maxProposals.Int64(), "maxProposals should be 2")

		// Create first proposal
		_, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Create second proposal
		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Try to create third proposal - should fail
		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		err = g.ExpectedFail(tx, err)
		ExpectedRevert(t, err, "TooManyActiveProposals")
	})
}

// TestChangeMaxProposals_DynamicLimitChange tests changing the limit and its effect
func TestChangeMaxProposals_DynamicLimitChange(t *testing.T) {
	customMembers := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}

	// Initialize with maxProposals=2
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
		customMembers[2].Operator.Address: {Balance: towei(1_000_000)},
	}, func(govValidator *params.SystemContract) {
		var members string
		for i, v := range customMembers {
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
			"maxProposals":  "2",
		}
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	proposer := customMembers[0].Operator
	approver := customMembers[1].Operator

	t.Run("increase limit and create more proposals", func(t *testing.T) {
		// Create first proposal
		firstProposalId, tx, err := g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Execute the first proposal to free up a slot
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, firstProposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Now create 2 more proposals (at limit again)
		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Try third - should fail
		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		err = g.ExpectedFail(tx, err)
		ExpectedRevert(t, err, "TooManyActiveProposals")

		// Execute one of the pending proposals to free up a slot
		secondProposalId := big.NewInt(2)
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, secondProposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Now we can create a proposal to increase the limit to 5
		changeProposalId, tx, err := g.BaseTxProposeChangeMaxProposals(t, g.govValidator, proposer, big.NewInt(5))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Approve and execute the change maxProposals proposal
		tx, err = g.BaseTxApproveProposal(t, g.govValidator, approver, changeProposalId)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		// Verify new limit
		maxAfter, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, proposer)
		require.NoError(t, err)
		require.Equal(t, int64(5), maxAfter.Int64())

		// Now should be able to create more proposals up to the new limit
		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		_, tx, err = g.BaseTxProposeChangeQuorum(t, g.govValidator, proposer, uint32(2))
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)
	})
}

// TestChangeMaxProposals_GenesisInitialization tests custom genesis initialization
func TestChangeMaxProposals_GenesisInitialization(t *testing.T) {
	customMembers := []*TestCandidate{NewTestCandidate(), NewTestCandidate()}

	testCases := []struct {
		name          string
		genesisValue  string
		expectedValue int64
	}{
		{"default value when not specified", "", 3},
		{"custom value 10", "10", 10},
		{"custom value 1", "1", 1},
		{"custom value 50", "50", 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGovWBFT(t, types.GenesisAlloc{
				customMembers[0].Operator.Address: {Balance: towei(1_000_000)},
				customMembers[1].Operator.Address: {Balance: towei(1_000_000)},
			}, func(govValidator *params.SystemContract) {
				govValidator.Params = map[string]string{
					"members":       customMembers[0].Operator.Address.Hex() + "," + customMembers[1].Operator.Address.Hex(),
					"quorum":        "2",
					"expiry":        "86400",
					"memberVersion": "1",
				}
				if tc.genesisValue != "" {
					govValidator.Params["maxProposals"] = tc.genesisValue
				}
			}, nil, nil, nil, nil, nil)
			require.NoError(t, err)
			defer g.backend.Close()

			// Verify the genesis value
			maxProposals, err := g.BaseMaxActiveProposalsPerMember(g.govValidator, customMembers[0].Operator)
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, maxProposals.Int64(), "maxProposals should be %d", tc.expectedValue)
		})
	}
}

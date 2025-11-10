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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestChangeMember_SelfChange tests that an active member can change their own address
func TestChangeMember_SelfChange(t *testing.T) {
	// Setup: 3 validators
	customValidators := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
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
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	oldMember := customValidators[0].Operator
	newMemberAddress := NewEOA().Address

	t.Run("active member can change their own address", func(t *testing.T) {
		// Verify old member is active
		memberBefore, err := g.BaseMembers(g.govValidator, oldMember, oldMember.Address)
		require.NoError(t, err)
		require.True(t, memberBefore.IsActive, "Old member should be active")

		// Get version before change (should NOT change)
		versionBefore, err := g.BaseMemberVersion(g.govValidator, oldMember)
		require.NoError(t, err)

		// Find old member's index in current version
		var oldIndex int
		for i := 0; i < 3; i++ {
			addr, err := g.BaseVersionedMemberList(g.govValidator, oldMember, versionBefore, big.NewInt(int64(i)))
			require.NoError(t, err)
			if addr == oldMember.Address {
				oldIndex = i
				break
			}
		}

		// Change member address
		tx, err := g.BaseTxChangeMember(t, g.govValidator, oldMember, newMemberAddress)
		receipt, err := g.ExpectedOk(tx, err)
		require.NoError(t, err)
		require.Equal(t, uint64(1), receipt.Status, "changeMember should succeed")

		// Verify version NOT incremented (same version)
		versionAfter, err := g.BaseMemberVersion(g.govValidator, oldMember)
		require.NoError(t, err)
		require.Equal(t, versionBefore, versionAfter, "Version should NOT change")

		// Verify old member is now inactive
		memberAfter, err := g.BaseMembers(g.govValidator, oldMember, oldMember.Address)
		require.NoError(t, err)
		require.False(t, memberAfter.IsActive, "Old member should be inactive")

		// Verify new member is active
		newMember, err := g.BaseMembers(g.govValidator, oldMember, newMemberAddress)
		require.NoError(t, err)
		require.True(t, newMember.IsActive, "New member should be active")

		// Verify new member is at same index in SAME version (replaced in-place)
		newMemberAtIndex, err := g.BaseVersionedMemberList(g.govValidator, oldMember, versionAfter, big.NewInt(int64(oldIndex)))
		require.NoError(t, err)
		require.Equal(t, newMemberAddress, newMemberAtIndex, "New member should replace old member at same index")

		t.Logf("✓ Member address changed from %s to %s", oldMember.Address.Hex()[:10], newMemberAddress.Hex()[:10])
		t.Logf("✓ Version unchanged: %s", versionAfter.String())
	})
}

// TestChangeMember_ValidationErrors tests validation errors
func TestChangeMember_ValidationErrors(t *testing.T) {
	customValidators := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
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
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	activeMember := customValidators[0].Operator
	nonMember := NewEOA()
	otherActiveMember := customValidators[1].Operator

	t.Run("non-member cannot change address", func(t *testing.T) {
		newAddress := NewEOA().Address
		tx, err := g.BaseTxChangeMember(t, g.govValidator, nonMember, newAddress)
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Non-member should not be able to change address")
		t.Logf("✓ Non-member correctly prevented from changing address")
	})

	t.Run("cannot change to zero address", func(t *testing.T) {
		tx, err := g.BaseTxChangeMember(t, g.govValidator, activeMember, common.Address{})
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Cannot change to zero address")
		t.Logf("✓ Zero address correctly rejected")
	})

	t.Run("cannot change to already active member address", func(t *testing.T) {
		tx, err := g.BaseTxChangeMember(t, g.govValidator, activeMember, otherActiveMember.Address)
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Cannot change to address of existing active member")
		t.Logf("✓ Existing active member address correctly rejected")
	})

	t.Run("cannot change to own address", func(t *testing.T) {
		tx, err := g.BaseTxChangeMember(t, g.govValidator, activeMember, activeMember.Address)
		err = g.ExpectedFail(tx, err)
		require.Error(t, err, "Cannot change to own address")
		t.Logf("✓ Self address correctly rejected")
	})
}

// TestChangeMember_MultipleChanges tests consecutive address changes
func TestChangeMember_MultipleChanges(t *testing.T) {
	customValidators := []*TestCandidate{NewTestCandidate(), NewTestCandidate(), NewTestCandidate()}
	newAddress1 := NewEOA()
	newAddress2 := NewEOA()

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		customValidators[0].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[1].Operator.Address: {Balance: towei(1_000_000)},
		customValidators[2].Operator.Address: {Balance: towei(1_000_000)},
		newAddress1.Address:                  {Balance: towei(1_000_000)},
		newAddress2.Address:                  {Balance: towei(1_000_000)},
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
	}, nil, nil, nil, nil, nil)
	require.NoError(t, err)
	defer g.backend.Close()

	member0 := customValidators[0].Operator
	member1 := customValidators[1].Operator

	t.Run("member can change address multiple times", func(t *testing.T) {
		versionBefore, err := g.BaseMemberVersion(g.govValidator, member0)
		require.NoError(t, err)

		// First change: member0 → newAddress1
		tx, err := g.BaseTxChangeMember(t, g.govValidator, member0, newAddress1.Address)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		versionAfterFirst, err := g.BaseMemberVersion(g.govValidator, member0)
		require.NoError(t, err)
		require.Equal(t, versionBefore, versionAfterFirst, "Version should NOT change")

		// Verify newAddress1 is active
		member, err := g.BaseMembers(g.govValidator, member0, newAddress1.Address)
		require.NoError(t, err)
		require.True(t, member.IsActive)

		// Second change: newAddress1 → newAddress2
		tx, err = g.BaseTxChangeMember(t, g.govValidator, newAddress1, newAddress2.Address)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		versionAfterSecond, err := g.BaseMemberVersion(g.govValidator, member0)
		require.NoError(t, err)
		require.Equal(t, versionAfterFirst, versionAfterSecond, "Version should still NOT change")

		// Verify newAddress2 is active and newAddress1 is inactive
		member2, err := g.BaseMembers(g.govValidator, member0, newAddress2.Address)
		require.NoError(t, err)
		require.True(t, member2.IsActive, "newAddress2 should be active")

		member1, err := g.BaseMembers(g.govValidator, member0, newAddress1.Address)
		require.NoError(t, err)
		require.False(t, member1.IsActive, "newAddress1 should be inactive")

		t.Logf("✓ Multiple address changes completed successfully")
		t.Logf("✓ member0 → newAddress1 → newAddress2 (version unchanged)")
	})

	t.Run("different members can change independently", func(t *testing.T) {
		versionBefore, err := g.BaseMemberVersion(g.govValidator, member1)
		require.NoError(t, err)

		newAddress3 := NewEOA().Address
		tx, err := g.BaseTxChangeMember(t, g.govValidator, member1, newAddress3)
		_, err = g.ExpectedOk(tx, err)
		require.NoError(t, err)

		versionAfter, err := g.BaseMemberVersion(g.govValidator, member1)
		require.NoError(t, err)
		require.Equal(t, versionBefore, versionAfter, "Version should NOT change")

		// Verify member1 is inactive and newAddress3 is active
		member1After, err := g.BaseMembers(g.govValidator, member1, member1.Address)
		require.NoError(t, err)
		require.False(t, member1After.IsActive)

		member3, err := g.BaseMembers(g.govValidator, member1, newAddress3)
		require.NoError(t, err)
		require.True(t, member3.IsActive)

		t.Logf("✓ Independent member address changes work correctly")
	})
}

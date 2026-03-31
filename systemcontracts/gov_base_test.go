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

package systemcontracts

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// CompareParam compares two StateParam objects for testing
func CompareParam(result params.StateParam, expect params.StateParam, t *testing.T, testName string) {
	if result.Address != expect.Address {
		t.Errorf("[test=%v] Address mismatch: result %s, expect %s", testName, result.Address.Hex(), expect.Address.Hex())
	}
	if result.Key != expect.Key {
		t.Errorf("[test=%v] Key mismatch: result %s, expect %s", testName, result.Key.Hex(), expect.Key.Hex())
	}
	if result.Value != expect.Value {
		t.Errorf("[test=%v] Value mismatch: result %s, expect %s", testName, result.Value.Hex(), expect.Value.Hex())
	}
}

// TestMemberToHash tests the Member struct encoding for Solidity storage
func TestMemberToHash(t *testing.T) {
	testCases := []struct {
		name     string
		member   Member
		validate func(common.Hash) bool
	}{
		{
			name:   "active member with joinedAt=0",
			member: Member{IsActive: true, JoinedAt: 0},
			validate: func(h common.Hash) bool {
				// Byte 31 should be 0x01 (isActive=true)
				return h[31] == 0x01
			},
		},
		{
			name:   "inactive member with joinedAt=0",
			member: Member{IsActive: false, JoinedAt: 0},
			validate: func(h common.Hash) bool {
				// All bytes should be 0x00
				return h == common.Hash{}
			},
		},
		{
			name:   "active member with joinedAt=1000000",
			member: Member{IsActive: true, JoinedAt: 1000000},
			validate: func(h common.Hash) bool {
				// Byte 31 should be 0x01, and joinedAt should be encoded in bytes 27-30
				return h[31] == 0x01 && h != common.Hash{}
			},
		},
		{
			name:   "active member with max uint32",
			member: Member{IsActive: true, JoinedAt: 4294967295},
			validate: func(h common.Hash) bool {
				// Byte 31 should be 0x01
				return h[31] == 0x01
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.member.ToHash()
			if !tc.validate(result) {
				t.Errorf("validation failed for %s, got %s", tc.name, result.Hex())
			}
		})
	}
}

// TestInitializeBase tests initialization logic with dynamic hash calculation
func TestInitializeBase(t *testing.T) {
	testCases := []struct {
		name          string
		param         map[string]string
		expectErr     string
		validateCount int
		validateFunc  func(*testing.T, []params.StateParam)
	}{
		{
			name:          "empty param",
			param:         map[string]string{},
			expectErr:     "",
			validateCount: 1, // maxProposals is always initialized with default value 3
		},
		{
			name: "member without version - should fail",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			},
			expectErr: "`systemContracts.govBase.params.memberVersion` is required when `systemContracts.govBase.params.members` is set",
		},
		{
			name: "invalid version format",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "v2",
			},
			expectErr: "`systemContracts.govBase.params.memberVersion`: strconv.ParseUint: parsing \"v2\": invalid syntax",
		},
		{
			name: "quorum > member count",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "2",
			},
			expectErr: "`systemContracts.govBase.params.quorum` (2) must not be greater than unique member count (1)",
		},
		{
			name: "invalid quorum format",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "2.5",
			},
			expectErr: "`systemContracts.govBase.params.quorum`: strconv.ParseUint: parsing \"2.5\": invalid syntax",
		},
		{
			name: "invalid expiry format",
			param: map[string]string{
				GOV_BASE_PARAM_EXPIRY: "invalid",
			},
			expectErr: "`systemContracts.govBase.params.expiry`: strconv.ParseUint: parsing \"invalid\": invalid syntax",
		},
		{
			name: "single member with version",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
			},
			expectErr:     "",
			validateCount: 6, // members, versionedMemberList[v][0], memberIndexByVersion, versionedMemberList.length, memberVersion, maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify memberVersion is set correctly
				found := false
				for _, p := range sp {
					if p.Key == common.HexToHash(SLOT_GOV_BASE_version) {
						if p.Value != common.BigToHash(big.NewInt(1)) {
							t.Errorf("memberVersion: expected 1, got %s", p.Value.Big().String())
						}
						found = true
					}
				}
				if !found {
					t.Error("memberVersion not found in state params")
				}
			},
		},
		{
			name: "single member with quorum",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "1",
			},
			expectErr:     "",
			validateCount: 8, // +quorum, +quorumByVersion, +maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify quorum is set
				foundQuorum := false
				foundQuorumByVersion := false
				for _, p := range sp {
					if p.Key == common.HexToHash(SLOT_GOV_BASE_quorum) {
						if p.Value != common.BigToHash(big.NewInt(1)) {
							t.Errorf("quorum: expected 1, got %s", p.Value.Big().String())
						}
						foundQuorum = true
					}
					// quorumByVersion should also be set
					quorumByVersionSlot := CalculateMappingSlot(common.HexToHash(SLOT_GOV_BASE_quorumByVersion), big.NewInt(1))
					if p.Key == quorumByVersionSlot {
						foundQuorumByVersion = true
					}
				}
				if !foundQuorum {
					t.Error("quorum not found")
				}
				if !foundQuorumByVersion {
					t.Error("quorumByVersion not found")
				}
			},
		},
		{
			name: "member with expiry",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_EXPIRY:         "604800",
			},
			expectErr:     "",
			validateCount: 7, // +proposalExpiry, +maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify expiry is set
				found := false
				for _, p := range sp {
					if p.Key == common.HexToHash(SLOT_GOV_BASE_proposalExpiry) {
						if p.Value != common.BigToHash(big.NewInt(604800)) {
							t.Errorf("proposalExpiry: expected 604800, got %s", p.Value.Big().String())
						}
						found = true
					}
				}
				if !found {
					t.Error("proposalExpiry not found")
				}
			},
		},
		{
			name: "multiple members",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "2",
			},
			expectErr:     "",
			validateCount: 14, // 3*(members+versionedMemberList+memberIndexByVersion) + versionedMemberList.length + memberVersion + quorum + quorumByVersion + maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify length is 3
				versionedMemberListSlot := CalculateMappingSlot(common.HexToHash(SLOT_GOV_BASE_versionedMemberList), big.NewInt(1))
				found := false
				for _, p := range sp {
					if p.Key == versionedMemberListSlot {
						if p.Value != common.BigToHash(big.NewInt(3)) {
							t.Errorf("versionedMemberList length: expected 3, got %s", p.Value.Big().String())
						}
						found = true
					}
				}
				if !found {
					t.Error("versionedMemberList length not found")
				}
			},
		},
		{
			name: "duplicate members - should deduplicate",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "2",
			},
			expectErr:     "",
			validateCount: 11, // 2 unique members + maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify length is 2 (not 3)
				versionedMemberListSlot := CalculateMappingSlot(common.HexToHash(SLOT_GOV_BASE_versionedMemberList), big.NewInt(1))
				found := false
				for _, p := range sp {
					if p.Key == versionedMemberListSlot {
						if p.Value != common.BigToHash(big.NewInt(2)) {
							t.Errorf("versionedMemberList length: expected 2 (deduplicated), got %s", p.Value.Big().String())
						}
						found = true
					}
				}
				if !found {
					t.Error("versionedMemberList length not found")
				}
			},
		},
		{
			name: "quorum without members",
			param: map[string]string{
				GOV_BASE_PARAM_QUORUM: "3",
			},
			expectErr:     "",
			validateCount: 2, // quorum + maxProposals
		},
		{
			name: "expiry without members",
			param: map[string]string{
				GOV_BASE_PARAM_EXPIRY: "86400",
			},
			expectErr:     "",
			validateCount: 2, // expiry + maxProposals
		},
		{
			name: "quorum=0 should be rejected (SECURITY FIX)",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "0",
			},
			expectErr: "`systemContracts.govBase.params.quorum` must be greater than 0",
		},
		{
			name: "multi-member with quorum=1 should fail (SECURITY - matches Solidity)",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "1",
			},
			expectErr: "`systemContracts.govBase.params.quorum` (1) must be at least 2 for multi-member governance (member count: 3)",
		},
		{
			name: "single member with quorum=2 should fail",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "2",
			},
			expectErr: "`systemContracts.govBase.params.quorum` (2) must not be greater than unique member count (1)",
		},
		{
			name: "single member with quorum=1 should succeed",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_QUORUM:         "1",
			},
			expectErr:     "",
			validateCount: 8, // members, versionedMemberList[v][0], memberIndexByVersion, versionedMemberList.length, memberVersion, quorum, quorumByVersion, maxProposals
			validateFunc: func(t *testing.T, sp []params.StateParam) {
				// Verify single member governance is properly configured
				foundMember := false
				foundQuorum := false
				for _, p := range sp {
					if p.Key == common.HexToHash(SLOT_GOV_BASE_quorum) {
						if p.Value != common.BigToHash(big.NewInt(1)) {
							t.Errorf("quorum: expected 1, got %s", p.Value.Big().String())
						}
						foundQuorum = true
					}
					// Verify member is registered
					memberSlot := CalculateMappingSlot(common.HexToHash(SLOT_GOV_BASE_members), common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"))
					if p.Key == memberSlot {
						foundMember = true
					}
				}
				if !foundQuorum {
					t.Error("quorum not found")
				}
				if !foundMember {
					t.Error("member not found")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := initializeBase(common.Address{}, tc.param)

			// Check error expectation
			if tc.expectErr != "" {
				if err == nil {
					t.Errorf("expected error: %v, got nil", tc.expectErr)
				} else if err.Error() != tc.expectErr {
					t.Errorf("expected error: %v, got: %v", tc.expectErr, err.Error())
				}
				return
			}

			// No error expected
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check count
			if len(result) != tc.validateCount {
				t.Errorf("expected %d params, got %d", tc.validateCount, len(result))
			}

			// Run custom validation if provided
			if tc.validateFunc != nil {
				tc.validateFunc(t, result)
			}
		})
	}
}

// TestCalculateMappingSlot tests mapping slot calculation
func TestCalculateMappingSlot(t *testing.T) {
	// Test that function produces consistent results
	addr := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	slot := common.HexToHash(SLOT_GOV_BASE_members)

	result1 := CalculateMappingSlot(slot, addr)
	result2 := CalculateMappingSlot(slot, addr)

	if result1 != result2 {
		t.Error("CalculateMappingSlot should be deterministic")
	}

	// Test with big.Int
	version := big.NewInt(1)
	versionSlot := common.HexToHash(SLOT_GOV_BASE_versionedMemberList)

	versionResult1 := CalculateMappingSlot(versionSlot, version)
	versionResult2 := CalculateMappingSlot(versionSlot, version)

	if versionResult1 != versionResult2 {
		t.Error("CalculateMappingSlot with big.Int should be deterministic")
	}
}

// TestCalculateDynamicSlot tests dynamic array slot calculation
func TestCalculateDynamicSlot(t *testing.T) {
	baseSlot := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	result1 := CalculateDynamicSlot(baseSlot, big.NewInt(0))
	result2 := CalculateDynamicSlot(baseSlot, big.NewInt(0))

	if result1 != result2 {
		t.Error("CalculateDynamicSlot should be deterministic")
	}

	// Sequential indices should produce different hashes
	result3 := CalculateDynamicSlot(baseSlot, big.NewInt(1))
	if result1 == result3 {
		t.Error("Different indices should produce different slots")
	}
}

// TestGovernanceConfig tests the GovernanceConfig struct
func TestGovernanceConfig(t *testing.T) {
	config := GovernanceConfig{
		Members: []common.Address{
			common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			common.HexToAddress("0x1111111111111111111111111111111111111111"),
		},
		Quorum:         2,
		ProposalExpiry: big.NewInt(604800),
	}

	if len(config.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(config.Members))
	}
	if config.Quorum != 2 {
		t.Errorf("expected quorum 2, got %d", config.Quorum)
	}
	if config.ProposalExpiry.Cmp(big.NewInt(604800)) != 0 {
		t.Errorf("expected proposalExpiry 604800, got %s", config.ProposalExpiry.String())
	}
}

// TestStorageSlotConstants verifies that storage slot constants match GovBase.sol
func TestStorageSlotConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"proposalExpiry", SLOT_GOV_BASE_proposalExpiry, "0x0"},
		{"memberVersion", SLOT_GOV_BASE_version, "0x1"},
		{"currentProposalId", SLOT_GOV_BASE_currentProposalId, "0x2"},
		{"reentrancyGuard", SLOT_GOV_BASE_reentrancyGuard, "0x3"},
		{"quorum", SLOT_GOV_BASE_quorum, "0x4"},
		{"members", SLOT_GOV_BASE_members, "0x5"},
		{"versionedMemberList", SLOT_GOV_BASE_versionedMemberList, "0x6"},
		{"proposals", SLOT_GOV_BASE_proposals, "0x7"},
		{"memberIndexByVersion", SLOT_GOV_BASE_memberIndexByVersion, "0x8"},
		{"quorumByVersion", SLOT_GOV_BASE_quorumByVersion, "0x9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("SLOT_%s = %s, expected %s", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestQuorumZeroValidation tests that quorum=0 is rejected
func TestQuorumZeroValidation(t *testing.T) {
	param := map[string]string{
		GOV_BASE_PARAM_QUORUM: "0",
	}
	_, err := initializeBase(common.Address{}, param)
	if err == nil {
		t.Error("expected error for quorum=0, got nil")
	}
	expectedMsg := "`systemContracts.govBase.params.quorum` must be greater than 0"
	if err.Error() != expectedMsg {
		t.Errorf("expected error: %v, got: %v", expectedMsg, err.Error())
	}
}

// TestZeroAddressMember tests that zero address in members is rejected
func TestZeroAddressMember(t *testing.T) {
	param := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0x0000000000000000000000000000000000000000",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_BASE_PARAM_QUORUM:         "1",
	}
	_, err := initializeBase(common.Address{}, param)
	if err == nil {
		t.Error("expected error for zero address member, got nil")
	}
	if !strings.Contains(err.Error(), "invalid zero address") {
		t.Errorf("expected error about zero address, got: %v", err.Error())
	}
}

// TestMemberIndexOverflow tests that member count exceeding MAX_MEMBERS is rejected
func TestMemberIndexOverflow(t *testing.T) {
	// Create 256 members (exceeds MAX_MEMBERS = 255)
	members := make([]string, 256)
	for i := 0; i < 256; i++ {
		// Generate unique addresses
		members[i] = fmt.Sprintf("0x%040x", i+1)
	}

	param := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        strings.Join(members, ","),
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_BASE_PARAM_QUORUM:         "1",
	}
	_, err := initializeBase(common.Address{}, param)
	if err == nil {
		t.Error("expected error for member overflow, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds maximum allowed") {
		t.Errorf("expected error about max members, got: %v", err.Error())
	}
}

// TestQuorumAfterDeduplication tests that quorum is validated after deduplication
func TestQuorumAfterDeduplication(t *testing.T) {
	param := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_BASE_PARAM_QUORUM:         "3", // 3 addresses but only 2 unique
	}
	_, err := initializeBase(common.Address{}, param)
	if err == nil {
		t.Error("expected error for quorum > unique members, got nil")
	}
	if !strings.Contains(err.Error(), "must not be greater than unique member count") {
		t.Errorf("expected error about quorum vs unique members, got: %v", err.Error())
	}
}

// TestValidConfiguration tests that valid configuration passes all checks
func TestValidConfiguration(t *testing.T) {
	param := map[string]string{
		GOV_BASE_PARAM_MEMBERS:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd,0x1111111111111111111111111111111111111111,0x2222222222222222222222222222222222222222",
		GOV_BASE_PARAM_MEMBER_VERSION: "1",
		GOV_BASE_PARAM_QUORUM:         "2",
		GOV_BASE_PARAM_EXPIRY:         "604800",
	}
	result, err := initializeBase(common.Address{}, param)
	if err != nil {
		t.Errorf("unexpected error for valid config: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty state params")
	}
}


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
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_BASE_PARAM_MEMBERS        = "members"
	GOV_BASE_PARAM_QUORUM         = "quorum"
	GOV_BASE_PARAM_EXPIRY         = "expiry"
	GOV_BASE_PARAM_MEMBER_VERSION = "memberVersion"
	GOV_BASE_PARAM_MAX_PROPOSALS  = "maxProposals"

	// GovBase Storage Layout (matches GovBaseV2.sol):
	// Slot 0: proposalExpiry (uint256)
	// Slot 1: memberVersion (uint256)
	// Slot 2: currentProposalId (uint256)
	// Slot 3: _reentrancyGuard (uint256)
	// Slot 4: quorum (uint32)
	// Slot 5: members (mapping)
	// Slot 6: versionedMemberList (mapping)
	// Slot 7: proposals (mapping)
	// Slot 8: memberIndexByVersion (mapping(uint256 => mapping(address => uint32)))
	// Slot 9: quorumByVersion (mapping(uint256 => uint32))
	// Slot 10: proposalExecutionCount (mapping(uint256 => uint256))
	// Slot 11: memberActiveProposalCount (mapping(address => uint256))
	// Slot 12: maxActiveProposalsPerMember (uint256)
	// Slot 13-49: __gap (reserved storage)
	SLOT_GOV_BASE_proposalExpiry               = "0x0"
	SLOT_GOV_BASE_version                      = "0x1"
	SLOT_GOV_BASE_currentProposalId            = "0x2"
	SLOT_GOV_BASE_reentrancyGuard              = "0x3"
	SLOT_GOV_BASE_quorum                       = "0x4"
	SLOT_GOV_BASE_members                      = "0x5"
	SLOT_GOV_BASE_versionedMemberList          = "0x6"
	SLOT_GOV_BASE_proposals                    = "0x7"
	SLOT_GOV_BASE_memberIndexByVersion         = "0x8"
	SLOT_GOV_BASE_quorumByVersion              = "0x9"
	SLOT_GOV_BASE_proposalExecutionCount       = "0xa"
	SLOT_GOV_BASE_memberActiveProposalCount    = "0xb"
	SLOT_GOV_BASE_maxActiveProposalsPerMember  = "0xc"
)

type Member struct {
	IsActive bool
	JoinedAt uint32
}

type GovernanceConfig struct {
	Members        []common.Address
	Quorum         uint32
	ProposalExpiry *big.Int
}

// ProposalStatus represents the state of a governance proposal
// Matches the ProposalStatus enum in GovBaseV2.sol
type ProposalStatus uint8

const (
	ProposalStatusNone      ProposalStatus = 0 // No proposal exists
	ProposalStatusVoting    ProposalStatus = 1 // Voting in progress - voting and execution allowed
	ProposalStatusApproved  ProposalStatus = 2 // Quorum reached - voting and execution allowed
	ProposalStatusExecuted  ProposalStatus = 3 // Successfully executed - all operations disallowed
	ProposalStatusCancelled ProposalStatus = 4 // Cancelled by proposer - all operations disallowed
	ProposalStatusExpired   ProposalStatus = 5 // Expired due to timeout - all operations disallowed
	ProposalStatusFailed    ProposalStatus = 6 // Execution failed - all operations disallowed
	ProposalStatusRejected  ProposalStatus = 7 // Rejected by votes - all operations disallowed
)

type Proposal struct {
	ActionType        [32]byte
	MemberVersion     *big.Int
	VotedBitmap       *big.Int
	CreatedAt         *big.Int
	ExecutedAt        *big.Int
	Proposer          common.Address
	RequiredApprovals uint32
	Approved          uint32
	Rejected          uint32
	Status            ProposalStatus
	CallData          []byte
}

func (m Member) ToHash() common.Hash {
	var result common.Hash
	// Solidity tight packing for structs in storage:
	// Fields are packed from right to left (lower-order first)
	// Byte 31: isActive (bool) - rightmost byte
	// Bytes 27-30: joinedAt (uint32)
	// Bytes 0-26: padding (zeros)
	if m.IsActive {
		result[31] = 0x01
	} else {
		result[31] = 0x00
	}
	binary.BigEndian.PutUint32(result[27:31], m.JoinedAt)
	return result
}

func initializeBase(govBaseAddress common.Address, param map[string]string) ([]params.StateParam, error) {
	sp := make([]params.StateParam, 0)

	quorum := uint64(0)
	if quorumStr, ok := param[GOV_BASE_PARAM_QUORUM]; ok {
		var err error
		quorum, err = strconv.ParseUint(quorumStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`systemContracts.govBase.params.quorum`: %w", err)
		}

		// Validate quorum is greater than 0
		if quorum == 0 {
			return nil, fmt.Errorf("`systemContracts.govBase.params.quorum` must be greater than 0")
		}

		sp = append(sp,
			params.StateParam{
				Address: govBaseAddress,
				Key:     common.HexToHash(SLOT_GOV_BASE_quorum),
				Value:   common.BigToHash(big.NewInt(int64(quorum))),
			},
		)
	}

	if expiryStr, ok := param[GOV_BASE_PARAM_EXPIRY]; ok {
		expiry, err := strconv.ParseUint(expiryStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`systemContracts.govBase.params.expiry`: %w", err)
		}
		sp = append(sp,
			params.StateParam{
				Address: govBaseAddress,
				Key:     common.HexToHash(SLOT_GOV_BASE_proposalExpiry),
				Value:   common.BigToHash(big.NewInt(int64(expiry))),
			},
		)
	}

	if membersStr, ok := param[GOV_BASE_PARAM_MEMBERS]; ok {
		memberAddresses := strings.Split(membersStr, ",")

		// Pre-validate and deduplicate members
		uniqueMembers := make(map[common.Address]struct{})
		for _, addrStr := range memberAddresses {
			member := common.HexToAddress(addrStr)

			// Validate zero address
			if member == (common.Address{}) {
				return nil, fmt.Errorf("`systemContracts.govBase.params.members` contains invalid zero address")
			}

			uniqueMembers[member] = struct{}{}
		}

		// Check quorum after deduplication
		if quorum > 0 && uint64(len(uniqueMembers)) < quorum {
			return nil, fmt.Errorf("`systemContracts.govBase.params.quorum` (%d) must not be greater than unique member count (%d)", quorum, len(uniqueMembers))
		}

		// Enforce maximum member limit (matches Solidity MAX_MEMBER_INDEX)
		const MAX_MEMBERS = 255
		if len(uniqueMembers) > MAX_MEMBERS {
			return nil, fmt.Errorf("`systemContracts.govBase.params.members` count (%d) exceeds maximum allowed (%d)", len(uniqueMembers), MAX_MEMBERS)
		}

		// Quorum validation matching GovBaseV2.sol logic
		// This ensures consistency between genesis initialization and smart contract behavior
		// Only validate if quorum is explicitly provided (quorum > 0)
		if quorum > 0 {
			memberCount := uint64(len(uniqueMembers))
			if memberCount == 1 {
				// Single member governance: quorum must be exactly 1
				// WARNING: Single-member governance is centralized and not recommended for production
				if quorum != 1 {
					return nil, fmt.Errorf("`systemContracts.govBase.params.quorum` must be exactly 1 for single-member governance, got %d", quorum)
				}
			} else {
				// Multi-member governance: require at least 2 approvals (proposer + 1 reviewer)
				// This prevents any single member from executing proposals without peer review
				// Security: quorum=1 would allow single-member unilateral decisions
				if quorum < 2 {
					return nil, fmt.Errorf("`systemContracts.govBase.params.quorum` (%d) must be at least 2 for multi-member governance (member count: %d)", quorum, memberCount)
				}
			}
		}

		membersSlot := common.HexToHash(SLOT_GOV_BASE_members)
		versionedMemberListSlot := common.HexToHash(SLOT_GOV_BASE_versionedMemberList)
		versionSlot := common.HexToHash(SLOT_GOV_BASE_version)
		memberIndexByVersionSlot := common.HexToHash(SLOT_GOV_BASE_memberIndexByVersion)
		quorumByVersionSlot := common.HexToHash(SLOT_GOV_BASE_quorumByVersion)

		versionStr, ok2 := param[GOV_BASE_PARAM_MEMBER_VERSION]
		if !ok2 {
			return nil, fmt.Errorf("`systemContracts.govBase.params.memberVersion` is required when `systemContracts.govBase.params.members` is set")
		}
		versionInt, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`systemContracts.govBase.params.memberVersion`: %w", err)
		}
		version := new(big.Int).SetUint64(versionInt)

		// joinedAt is set to 0 for all initial members.
		// We do not use the current time in the genesis (configuration).
		// This is because it makes the genesis generation non-deterministic.
		memberData := Member{
			IsActive: true,
			JoinedAt: 0,
		}.ToHash()

		currentIdx := uint64(0)
		// Use uniqueMembers map for iteration
		for member := range uniqueMembers {
			// Additional overflow check (defensive programming)
			if currentIdx >= MAX_MEMBERS {
				return nil, fmt.Errorf("member index overflow: currentIdx (%d) >= MAX_MEMBERS (%d)", currentIdx, MAX_MEMBERS)
			}
			sp = append(sp,
				// Set member active status and joinedAt timestamp
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateMappingSlot(membersSlot, member),
					Value:   memberData,
				},
				// Add member to versionedMemberList[version][currentIdx]
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateDynamicSlot(CalculateMappingSlot(versionedMemberListSlot, version), new(big.Int).SetUint64(currentIdx)),
					Value:   common.BytesToHash(member.Bytes()),
				},
				// Set memberIndexByVersion[version][member] = currentIdx + 1
				// Note: Index is 1-based (0 means not a member)
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateMappingSlot(CalculateMappingSlot(memberIndexByVersionSlot, version), member),
					Value:   common.BigToHash(new(big.Int).SetUint64(currentIdx + 1)),
				},
			)
			currentIdx++
		}

		if currentIdx > 0 {
			sp = append(sp,
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateMappingSlot(versionedMemberListSlot, version),
					Value:   common.BigToHash(new(big.Int).SetUint64(currentIdx)),
				},
			)
		}
		sp = append(sp,
			params.StateParam{
				Address: govBaseAddress,
				Key:     versionSlot,
				Value:   common.BigToHash(version),
			},
		)

		// Set quorumByVersion[version] = quorum
		// This creates a snapshot of the quorum value at this member version
		if quorum > 0 {
			sp = append(sp,
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateMappingSlot(quorumByVersionSlot, version),
					Value:   common.BigToHash(new(big.Int).SetUint64(quorum)),
				},
			)
		}
	}

	// Initialize maxActiveProposalsPerMember (default: 3, range: 1-50)
	maxProposals := uint64(3) // Default value
	if maxProposalsStr, ok := param[GOV_BASE_PARAM_MAX_PROPOSALS]; ok {
		var err error
		maxProposals, err = strconv.ParseUint(maxProposalsStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`systemContracts.govBase.params.maxProposals`: %w", err)
		}

		// Validate range: 1 to 50 (inclusive)
		if maxProposals < 1 || maxProposals > 50 {
			return nil, fmt.Errorf("`systemContracts.govBase.params.maxProposals` must be between 1 and 50, got %d", maxProposals)
		}
	}

	sp = append(sp,
		params.StateParam{
			Address: govBaseAddress,
			Key:     common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember),
			Value:   common.BigToHash(big.NewInt(int64(maxProposals))),
		},
	)

	return sp, nil
}

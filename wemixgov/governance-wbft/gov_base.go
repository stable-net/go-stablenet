// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

package govwbft

import (
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"strconv"
	"strings"
)

const (
	GOV_BASE_PARAM_MEMBERS        = "members"
	GOV_BASE_PARAM_QUORUM         = "quorum"
	GOV_BASE_PARAM_EXPIRY         = "expiry"
	GOV_BASE_PARAM_MEMBER_VERSION = "memberVersion"

	SLOT_GOV_BASE_quorum              = "0x0"
	SLOT_GOV_BASE_proposalExpiry      = "0x1"
	SLOT_GOV_BASE_members             = "0x2"
	SLOT_GOV_BASE_versionedMemberList = "0x3"
	SLOT_GOV_BASE_version             = "0x4"
)

type Member struct {
	IsActive bool
	JoinedAt uint32
}

func (m Member) ToHash() common.Hash {
	var result common.Hash
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
		if quorum > 0 && uint64(len(memberAddresses)) < quorum {
			return nil, fmt.Errorf("`systemContracts.govBase.params.quorum` must not be greater than the number of members")
		}

		membersSlot := common.HexToHash(SLOT_GOV_BASE_members)
		versionedMemberListSlot := common.HexToHash(SLOT_GOV_BASE_versionedMemberList)
		versionSlot := common.HexToHash(SLOT_GOV_BASE_version)

		versionStr, ok2 := param[GOV_BASE_PARAM_MEMBER_VERSION]
		if !ok2 {
			return nil, fmt.Errorf("`systemContracts.govBase.params.memberVersion` is required when `systemContracts.govBase.params.members` is set")
		}
		versionInt, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`systemContracts.govBase.params.memberVersion`: %w", err)
		}
		version := new(big.Int).SetUint64(versionInt)

		duplicated := make(map[common.Address]struct{})

		// joinedAt is set to 0 for all initial members.
		// We do not use the current time in the genesis (configuration).
		// This is because it makes the genesis generation non-deterministic.
		memberData := Member{
			IsActive: true,
			JoinedAt: 0,
		}.ToHash()

		currentIdx := uint64(0)
		for _, memberAddr := range memberAddresses {
			member := common.HexToAddress(memberAddr)
			if _, ok := duplicated[member]; ok {
				continue
			}
			sp = append(sp,
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateMappingSlot(membersSlot, member),
					Value:   memberData,
				},
				params.StateParam{
					Address: govBaseAddress,
					Key:     CalculateDynamicSlot(CalculateMappingSlot(versionedMemberListSlot, version), new(big.Int).SetUint64(currentIdx)),
					Value:   common.BytesToHash(member.Bytes()),
				},
			)
			duplicated[member] = struct{}{}
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
	}

	return sp, nil
}

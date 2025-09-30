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
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_BASE_PARAM_MEMBERS = "members"
	GOV_BASE_PARAM_QUORUM  = "quorum"
	GOV_BASE_PARAM_EXPIRY  = "expiry"

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

func initializeBase(govBaseAddress common.Address, members []common.Address, quorum uint64, expiry uint64) []params.StateParam {
	param := make([]params.StateParam, 0)

	param = append(param,
		params.StateParam{
			Address: govBaseAddress,
			Key:     common.HexToHash(SLOT_GOV_BASE_quorum),
			Value:   common.BigToHash(big.NewInt(int64(quorum))),
		},
		params.StateParam{
			Address: govBaseAddress,
			Key:     common.HexToHash(SLOT_GOV_BASE_proposalExpiry),
			Value:   common.BigToHash(big.NewInt(int64(expiry))),
		},
	)

	membersSlot := common.HexToHash(SLOT_GOV_BASE_members)
	versionedMemberListSlot := common.HexToHash(SLOT_GOV_BASE_versionedMemberList)
	versionSlot := common.HexToHash(SLOT_GOV_BASE_version)
	version := new(big.Int).SetUint64(1)
	duplicated := make(map[common.Address]struct{})

	memberData := Member{
		IsActive: true,
		JoinedAt: uint32(time.Now().Unix()),
	}.ToHash()

	currentIdx := uint64(0)
	for _, member := range members {
		if _, ok := duplicated[member]; ok {
			continue
		}
		param = append(param,
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
		param = append(param,
			params.StateParam{
				Address: govBaseAddress,
				Key:     CalculateMappingSlot(versionedMemberListSlot, version),
				Value:   common.BigToHash(new(big.Int).SetUint64(currentIdx)),
			},
		)
	}
	param = append(param,
		params.StateParam{
			Address: govBaseAddress,
			Key:     versionSlot,
			Value:   common.BigToHash(version),
		},
	)
	return param
}

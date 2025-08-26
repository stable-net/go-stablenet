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

package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// staker state
const (
	SLOT_BLS_POP_PRECOMPILED_ADDRESS = "0x0"
	SLOT_GOV_CONFIG_ADDRESS          = "0x1"
	SLOT_GOV_REWARDEE_IMP_ADDRESS    = "0x2"
	SLOT_TOTAL_STAKING               = "0x3"
	SLOT_STAKER_SET                  = "0x4"
	SLOT_STAKER_INFO                 = "0x6"
	SLOT_STAKER_BY_OPERATOR          = "0x7"
	SLOT_USER_REWARD_INFO            = "0xd"
	SLOT_DANGLING_DELEGATED          = "0xe"
)

type Staker struct {
	Operator            common.Address
	Rewardee            common.Address
	FeeRecipient        common.Address
	BLSPublicKey        []byte
	TotalStaked         *big.Int
	Delegated           *big.Int
	FeeRate             *big.Int
	AccRewardPerStaking *big.Int
	AccFeePerStaking    *big.Int
	LastRewardBalance   *big.Int
}

type UserRewardInfo struct {
	StakingAmount    *big.Int
	PendingReward    *big.Int
	PendingFee       *big.Int
	RewardPerStaking *big.Int
	FeePerStaking    *big.Int
}

// stakerInfo slot
const (
	StakerInfo_Rewardee int64 = iota + 1
	StakerInfo_FeeRecipient
	StakerInfo_FeeRate
	StakerInfo_BLSPublicKey
	StakerInfo_TotalStaked
	StakerInfo_AccRewardPerStaking
	StakerInfo_AccFeePerStaking
	StakerInfo_LastRewardBalance
)

func TotalStaking(govStakingAddress common.Address, state StateReader) *big.Int {
	return state.GetState(govStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func DanglingDelegated(govStakingAddress common.Address, state StateReader) *big.Int {
	return state.GetState(govStakingAddress, common.HexToHash(SLOT_DANGLING_DELEGATED)).Big()
}

func StakerLength(govStakingAddress common.Address, state StateReader) uint64 {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Length(state, govStakingAddress)
}

func IsStaker(govStakingAddress common.Address, state StateReader, staker common.Address) bool {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Contains(state, govStakingAddress, staker)
}

func Stakers(govStakingAddress common.Address, state StateReader) []common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Values(state, govStakingAddress)
}

func StakerAt(govStakingAddress common.Address, state StateReader, index *big.Int) common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.At(state, govStakingAddress, index)
}

func StakerByOperator(govStakingAddress common.Address, state StateReader, operator common.Address) common.Address {
	staker := state.GetState(govStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_STAKER_BY_OPERATOR), operator))
	return HashToAddress(staker)
}

func StakerInfo(govStakingAddress common.Address, state StateReader, staker common.Address) Staker {
	baseSlot := stakerInfoSlot(staker)

	stakerInfo := Staker{
		Operator:            getOperator(govStakingAddress, state, baseSlot),
		Rewardee:            HashToAddress(state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_Rewardee)))),
		FeeRecipient:        HashToAddress(state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_FeeRecipient)))),
		BLSPublicKey:        getBLSPublicKey(govStakingAddress, state, baseSlot),
		FeeRate:             state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_FeeRate))).Big(),
		TotalStaked:         state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_TotalStaked))).Big(),
		AccRewardPerStaking: state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_AccRewardPerStaking))).Big(),
		AccFeePerStaking:    state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_AccFeePerStaking))).Big(),
		LastRewardBalance:   state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_LastRewardBalance))).Big(),
	}
	userInfo := UserInfo(govStakingAddress, state, staker, staker)
	x := new(big.Int).Set(stakerInfo.TotalStaked)
	stakerInfo.Delegated = x.Sub(x, userInfo.StakingAmount)
	return stakerInfo
}

func UserInfo(govStakingAddress common.Address, state StateReader, staker common.Address, user common.Address) UserRewardInfo {
	baseSlot := userInfoSlot(staker, user)

	return UserRewardInfo{
		StakingAmount:    state.GetState(govStakingAddress, baseSlot).Big(),
		PendingReward:    state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(1))).Big(),
		PendingFee:       state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big(),
		RewardPerStaking: state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
		FeePerStaking:    state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(4))).Big(),
	}
}

func StakerInfoMap(govStakingAddress common.Address, state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := Stakers(govStakingAddress, state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(govStakingAddress, state, v)
	}
	return stakerInfos
}

func GetTotalStaked(govStakingAddress common.Address, state StateReader, staker common.Address) *big.Int {
	return getTotalStaked(govStakingAddress, state, stakerInfoSlot(staker))
}

func GetBLSPublicKey(govStakingAddress common.Address, state StateReader, staker common.Address) []byte {
	return getBLSPublicKey(govStakingAddress, state, stakerInfoSlot(staker))
}

func getOperator(govStakingAddress common.Address, state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(govStakingAddress, baseSlot))
}

func getTotalStaked(govStakingAddress common.Address, state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_TotalStaked))).Big()
}

func getBLSPublicKey(govStakingAddress common.Address, state StateReader, baseSlot common.Hash) []byte {
	return GetBytes(state, govStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_BLSPublicKey)))
}

func stakerInfoSlot(staker common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_STAKER_INFO), staker)
}

func userInfoSlot(staker common.Address, user common.Address) common.Hash {
	stakerMap := CalculateMappingSlot(common.HexToHash(SLOT_USER_REWARD_INFO), staker)
	return CalculateMappingSlot(stakerMap, user)
}

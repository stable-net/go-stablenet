package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// staker state
const (
	SLOT_TOTAL_STAKING       = "0x0"
	SLOT_STAKER_SET          = "0x1" // ,0x2
	SLOT_STAKER_INFO         = "0x3"
	SLOT_STAKER_BY_OPERATOR  = "0x4"
	SLOT_USER_REWARD_INFO    = "0x9"
	SLOT_DANGLING_DELEGATED  = "0xa"
	SLOT_AFTER_STABILIZATION = "0xb"
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

func IsAfterStabilization(state StateReader) bool {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_AFTER_STABILIZATION)).Big().Sign() > 0
}

func TotalStaking(state StateReader) *big.Int {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
}

func DanglingDelegated(state StateReader) *big.Int {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_DANGLING_DELEGATED)).Big()
}

func StakerLength(state StateReader) uint64 {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Length(state, GovStakingAddress)
}

func IsStaker(state StateReader, staker common.Address) bool {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Contains(state, GovStakingAddress, staker)
}

func Stakers(state StateReader) []common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.Values(state, GovStakingAddress)
}

func StakerAt(state StateReader, index *big.Int) common.Address {
	stakerSet := NewAddressSet(common.HexToHash(SLOT_STAKER_SET))
	return stakerSet.At(state, GovStakingAddress, index)
}

func StakerByOperator(state StateReader, operator common.Address) common.Address {
	staker := state.GetState(GovStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_STAKER_BY_OPERATOR), operator))
	return HashToAddress(staker)
}

func StakerInfo(state StateReader, staker common.Address) Staker {
	baseSlot := stakerInfoSlot(staker)

	stakerInfo := Staker{
		Operator:            getOperator(state, baseSlot),
		Rewardee:            HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_Rewardee)))),
		FeeRecipient:        HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_FeeRecipient)))),
		BLSPublicKey:        getBLSPublicKey(state, baseSlot),
		FeeRate:             state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_FeeRate))).Big(),
		TotalStaked:         state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_TotalStaked))).Big(),
		AccRewardPerStaking: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_AccRewardPerStaking))).Big(),
		AccFeePerStaking:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_AccFeePerStaking))).Big(),
		LastRewardBalance:   state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_LastRewardBalance))).Big(),
	}
	userInfo := UserInfo(state, staker, stakerInfo.Operator)
	x := new(big.Int).Set(stakerInfo.TotalStaked)
	stakerInfo.Delegated = x.Sub(x, userInfo.StakingAmount)
	return stakerInfo
}

func UserInfo(state StateReader, staker common.Address, user common.Address) UserRewardInfo {
	baseSlot := userInfoSlot(staker, user)

	return UserRewardInfo{
		StakingAmount:    state.GetState(GovStakingAddress, baseSlot).Big(),
		PendingReward:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1))).Big(),
		PendingFee:       state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big(),
		RewardPerStaking: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
		FeePerStaking:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(4))).Big(),
	}
}

func StakerInfoMap(state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := Stakers(state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(state, v)
	}
	return stakerInfos
}

func GetTotalStaked(state StateReader, staker common.Address) *big.Int {
	return getTotalStaked(state, stakerInfoSlot(staker))
}

func GetBLSPublicKey(state StateReader, staker common.Address) []byte {
	return getBLSPublicKey(state, stakerInfoSlot(staker))
}

func getOperator(state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(GovStakingAddress, baseSlot))
}

func getTotalStaked(state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_TotalStaked))).Big()
}

func getBLSPublicKey(state StateReader, baseSlot common.Hash) []byte {
	return GetBytes(state, GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_BLSPublicKey)))
}

func stakerInfoSlot(staker common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_STAKER_INFO), staker)
}

func userInfoSlot(staker common.Address, user common.Address) common.Hash {
	stakerMap := CalculateMappingSlot(common.HexToHash(SLOT_USER_REWARD_INFO), staker)
	return CalculateMappingSlot(stakerMap, user)
}

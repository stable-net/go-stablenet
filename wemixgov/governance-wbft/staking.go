package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// staker state
const (
	StakerState_None int64 = iota
	StakerState_Inactive
	StakerState_Active
)

// stakerInfo slot
const (
	StakerInfo_Rewardee int64 = iota + 1
	StakerInfo_Staking
	StakerInfo_Delegated
	StakerInfo_BLSPublicKey
	StakerInfo_State
)

const (
	SLOT_TOTAL_STAKING       = "0x0"
	SLOT_STAKER_SET          = "0x1" // ,0x2
	SLOT_STAKER_INFO         = "0x3"
	SLOT_STAKER_BY_OPERATOR  = "0x4"
	SLOT_AFTER_STABILIZATION = "0x9"
)

type Staker struct {
	Operator     common.Address
	Rewardee     common.Address
	Staking      *big.Int
	Delegated    *big.Int
	BLSPublicKey []byte
	State        int64
}

func IsAfterStabilization(state StateReader) bool {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_AFTER_STABILIZATION)).Big().Sign() > 0
}

func TotalStaking(state StateReader) *big.Int {
	return state.GetState(GovStakingAddress, common.HexToHash(SLOT_TOTAL_STAKING)).Big()
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

	return Staker{
		Operator:     getOperator(state, baseSlot),
		Rewardee:     HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_Rewardee)))),
		Staking:      getStaking(state, baseSlot),
		Delegated:    state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_Delegated))).Big(),
		BLSPublicKey: getBLSPublicKey(state, baseSlot),
		State:        getStakerState(state, baseSlot),
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

func GetStaking(state StateReader, staker common.Address) *big.Int {
	return getStaking(state, stakerInfoSlot(staker))
}

func GetBLSPublicKey(state StateReader, staker common.Address) []byte {
	return getBLSPublicKey(state, stakerInfoSlot(staker))
}

func IsActive(state StateReader, staker common.Address) bool {
	return GetStakerState(state, staker) == StakerState_Active
}

func GetStakerState(state StateReader, staker common.Address) int64 {
	return getStakerState(state, stakerInfoSlot(staker))
}

func getOperator(state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(GovStakingAddress, baseSlot))
}

func getStaking(state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_Staking))).Big()
}

func getBLSPublicKey(state StateReader, baseSlot common.Hash) []byte {
	return GetBytes(state, GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_BLSPublicKey)))
}

func getStakerState(state StateReader, baseSlot common.Hash) int64 {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(StakerInfo_State))).Big().Int64()
}

func stakerInfoSlot(staker common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_STAKER_INFO), staker)
}

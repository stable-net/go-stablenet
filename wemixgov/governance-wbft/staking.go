package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SLOT_TOTAL_STAKING    = "0x0"
	SLOT_STAKER_SET       = "0x1" // ,0x2
	SLOT_STAKER_INFO      = "0x3"
	SLOT_STAKER_BY_STAKER = "0x4"
)

type Staker struct {
	Operator  common.Address
	Rewardee  common.Address
	Staking   *big.Int
	Delegated *big.Int
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
	staker := state.GetState(GovStakingAddress, CalculateMappingSlot(common.HexToHash(SLOT_STAKER_BY_STAKER), operator))
	return HashToAddress(staker)
}

func StakerInfo(state StateReader, staker common.Address) Staker {
	baseSlot := stakerInfoSlot(staker)

	return Staker{
		Operator:  getOperator(state, baseSlot),
		Rewardee:  HashToAddress(state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(1)))),
		Staking:   getStaking(state, baseSlot),
		Delegated: state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(3))).Big(),
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

func getOperator(state StateReader, baseSlot common.Hash) common.Address {
	return HashToAddress(state.GetState(GovStakingAddress, baseSlot))
}

func getStaking(state StateReader, baseSlot common.Hash) *big.Int {
	return state.GetState(GovStakingAddress, IncrementHash(baseSlot, big.NewInt(2))).Big()
}

func stakerInfoSlot(staker common.Address) common.Hash {
	return CalculateMappingSlot(common.HexToHash(SLOT_STAKER_INFO), staker)
}

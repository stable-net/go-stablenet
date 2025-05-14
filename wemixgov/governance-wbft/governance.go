package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	GovConfigAddress      = common.HexToAddress(params.GOV_CONFIG_ADDRESS)
	GovStakingAddress     = common.HexToAddress(params.GOV_STAKING_ADDRESS)
	GovNCPAddress         = common.HexToAddress(params.GOV_NCP_ADDRESS)
	GovRewardeeImpAddress = common.HexToAddress(params.GOV_REWARDEE_IMP_ADDRESS)
)

func InitializeNCP(ncps []common.Address) []params.StateParam {
	param := make([]params.StateParam, 0)

	valueSlot := common.HexToHash(SLOT_NCP_LIST)
	indexSlot := IncrementHash(valueSlot, big.NewInt(1))
	duplicated := make(map[common.Address]struct{})

	currentIdx := uint64(0)
	newLength := new(big.Int)
	for _, ncp := range ncps {
		if _, ok := duplicated[ncp]; ok {
			continue
		}
		newLength = new(big.Int).SetUint64(currentIdx + 1)

		param = append(param,
			// set index slot
			params.StateParam{
				Address: GovNCPAddress,
				Key:     CalculateMappingSlot(indexSlot, ncp),
				Value:   common.BigToHash(newLength),
			},
			// set value slot
			params.StateParam{
				Address: GovNCPAddress,
				Key:     CalculateDynamicSlot(valueSlot, new(big.Int).SetUint64(currentIdx)),
				Value:   common.BytesToHash(ncp.Bytes()),
			},
		)
		duplicated[ncp] = struct{}{}
		currentIdx++
	}
	if newLength.Sign() > 0 {
		param = append(param, params.StateParam{
			Address: GovNCPAddress,
			Key:     valueSlot,
			Value:   common.BigToHash(newLength),
		})
	}
	return param
}

func IsNCPStaker(state StateReader, staker common.Address) bool {
	if !IsStaker(state, staker) {
		return false
	}
	return IsNCP(state, getOperator(state, stakerInfoSlot(staker)))
}

func NCPStakers(state StateReader) []common.Address {
	stakers := make([]common.Address, 0)
	ncps := NCPList(state)
	for _, ncp := range ncps {
		v := StakerByOperator(state, ncp)
		if v != (common.Address{}) {
			stakers = append(stakers, v)
		}
	}
	return stakers
}

func NCPTotalStaking(state StateReader) *big.Int {
	totalStaking := new(big.Int)
	stakers := NCPStakers(state)
	for _, v := range stakers {
		totalStaking.Add(totalStaking, GetTotalStaked(state, v))
	}
	return totalStaking
}

func NCPStakerInfoMap(state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := NCPStakers(state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(state, v)
	}
	return stakerInfos
}

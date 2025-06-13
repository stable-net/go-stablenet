package govwbft

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

func init() {
	// to avoid import cycle
	params.CheckInitGovContractVersions = checkInitGovContractVersions
	params.CheckUpgradeGovContractVersions = checkUpgradeGovContractVersions
}

func checkInitGovContractVersions(govContracts *params.GovContracts) error {
	if GovContractCodes[CONTRACT_GOV_CONFIG][govContracts.GovConfig.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govConfig`: unsupported version %s", govContracts.GovConfig.Version)
	}
	if GovContractCodes[CONTRACT_GOV_STAKING][govContracts.GovStaking.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govStaking`: unsupported version %s", govContracts.GovStaking.Version)
	}
	if GovContractCodes[CONTRACT_GOV_REWARDEE_IMP][govContracts.GovRewardeeImp.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govRewardeeImp`: unsupported version %s", govContracts.GovRewardeeImp.Version)
	}
	if govContracts.GovNCP != nil && GovContractCodes[CONTRACT_GOV_NCP][govContracts.GovNCP.Version] == "" {
		return fmt.Errorf("`montblanc.init.govContracts.govNCP`: unsupported version %s", govContracts.GovNCP.Version)
	}
	return nil
}

func checkUpgradeGovContractVersions(govContracts *params.GovContracts) error {
	if govContracts.GovConfig != nil && GovContractCodes[CONTRACT_GOV_CONFIG][govContracts.GovConfig.Version] == "" {
		return fmt.Errorf("`montblanc.upgrades.govContracts.govConfig`: unsupported version %s", govContracts.GovConfig.Version)
	}
	if govContracts.GovStaking != nil && GovContractCodes[CONTRACT_GOV_STAKING][govContracts.GovStaking.Version] == "" {
		return fmt.Errorf("`montblanc.upgrades.govContracts.govStaking`: unsupported version %s", govContracts.GovStaking.Version)
	}
	if govContracts.GovRewardeeImp != nil && GovContractCodes[CONTRACT_GOV_REWARDEE_IMP][govContracts.GovRewardeeImp.Version] == "" {
		return fmt.Errorf("`montblanc.upgrades.govContracts.govRewardeeImp`: unsupported version %s", govContracts.GovRewardeeImp.Version)
	}
	if govContracts.GovNCP != nil && GovContractCodes[CONTRACT_GOV_NCP][govContracts.GovNCP.Version] == "" {
		return fmt.Errorf("`montblanc.upgrades.govContracts.govNCP`: unsupported version %s", govContracts.GovNCP.Version)
	}
	return nil
}

func GetMontBlancTransition(govContracts *params.GovContracts) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if govContracts.GovConfig != nil {
		minStaking, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_MINIMUM_STAKING], 10)
		maxStaking, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_MAXIMUM_STAKING], 10)
		unbondingStaker, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_UNBONDING_STAKER], 10)
		unbondingDelegator, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_UNBONDING_DELEGATOR], 10)
		feePrecision, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_FEE_PRECISION], 10)
		changeFeeDelay, _ := new(big.Int).SetString(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_CHANGE_FEE_DELAY], 10)
		govCouncil := common.HexToAddress(govContracts.GovConfig.Params[GOV_CONFIG_PARAM_GOV_COUNCIL])
		if minStaking == nil || maxStaking == nil || unbondingStaker == nil || unbondingDelegator == nil ||
			feePrecision == nil || changeFeeDelay == nil {
			return nil, errors.New("invalid gov config params")
		}

		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovConfig.Address, Code: GovContractCodes[CONTRACT_GOV_CONFIG][govContracts.GovConfig.Version]})
		st.States = append(st.States, []params.StateParam{
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_MINIMUM_STAKING), Value: common.BigToHash(minStaking)},
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_MAXIMUM_STAKING), Value: common.BigToHash(maxStaking)},
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_UNBONDING_STAKER), Value: common.BigToHash(unbondingStaker)},
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_UNBONDING_DELEGATOR), Value: common.BigToHash(unbondingDelegator)},
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_FEE_PRECISION), Value: common.BigToHash(feePrecision)},
			{Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_CHANGE_FEE_DELAY), Value: common.BigToHash(changeFeeDelay)},
		}...)
		if govCouncil != (common.Address{}) {
			st.States = append(st.States, params.StateParam{
				Address: govContracts.GovConfig.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_GOV_COUNCIL), Value: common.BytesToHash(govCouncil.Bytes())})
		}
	}

	if govContracts.GovStaking != nil {
		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovStaking.Address, Code: GovContractCodes[CONTRACT_GOV_STAKING][govContracts.GovStaking.Version]})

		// initialize govConfig, govRewardeeImp addresses of GovStaking contract
		if govContracts.GovConfig != nil {
			st.States = append(st.States, params.StateParam{
				Address: govContracts.GovStaking.Address, Key: common.HexToHash(SLOT_GOV_CONFIG_ADDRESS), Value: common.BytesToHash(govContracts.GovConfig.Address.Bytes())})
		}
		if govContracts.GovRewardeeImp != nil {
			st.States = append(st.States, params.StateParam{
				Address: govContracts.GovStaking.Address, Key: common.HexToHash(SLOT_GOV_REWARDEE_IMP_ADDRESS), Value: common.BytesToHash(govContracts.GovRewardeeImp.Address.Bytes())})
		}
	}

	if govContracts.GovRewardeeImp != nil {
		st.Codes = append(st.Codes, params.CodeParam{
			Address: govContracts.GovRewardeeImp.Address, Code: GovContractCodes[CONTRACT_GOV_REWARDEE_IMP][govContracts.GovRewardeeImp.Version]})
	}

	if govContracts.GovNCP != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: govContracts.GovNCP.Address, Code: GovContractCodes[CONTRACT_GOV_NCP][govContracts.GovNCP.Version]})
		ncpAddresses := strings.Split(govContracts.GovNCP.Params[GOV_NCP_PARAM_NCPS], ",")
		ncps := make([]common.Address, 0)
		for _, ncp := range ncpAddresses {
			ncps = append(ncps, common.HexToAddress(ncp))
		}
		st.States = append(st.States, initializeNCP(govContracts.GovNCP.Address, ncps)...)
	}
	return st, nil
}

func initializeNCP(govNCPAddress common.Address, ncps []common.Address) []params.StateParam {
	param := make([]params.StateParam, 0)

	valueSlot := common.HexToHash(SLOT_NCP_LIST)
	indexSlot := IncrementHash(valueSlot, big.NewInt(1))
	duplicated := make(map[common.Address]struct{})

	currentIdx := uint64(0)
	newLength := new(big.Int)
	ncpID := new(big.Int)
	for _, ncp := range ncps {
		if _, ok := duplicated[ncp]; ok {
			continue
		}
		newLength = new(big.Int).SetUint64(currentIdx + 1)

		ncpID = new(big.Int).Add(ncpID, big.NewInt(1))
		param = append(param,
			// set index slot
			params.StateParam{
				Address: govNCPAddress,
				Key:     CalculateMappingSlot(indexSlot, ncp),
				Value:   common.BigToHash(newLength),
			},
			// set value slot
			params.StateParam{
				Address: govNCPAddress,
				Key:     CalculateDynamicSlot(valueSlot, new(big.Int).SetUint64(currentIdx)),
				Value:   common.BytesToHash(ncp.Bytes()),
			},

			// set id to address mapping
			params.StateParam{
				Address: govNCPAddress,
				Key:     CalculateMappingSlot(common.HexToHash(SLOT_NCP_ID_TO_ADDRESS), ncpID),
				Value:   common.BytesToHash(ncp.Bytes()),
			},
			// set address to id mapping
			params.StateParam{
				Address: govNCPAddress,
				Key:     CalculateMappingSlot(common.HexToHash(SLOT_NCP_ADDRESS_TO_ID), ncp),
				Value:   common.BigToHash(ncpID),
			},
		)
		duplicated[ncp] = struct{}{}
		currentIdx++
	}
	if newLength.Sign() > 0 {
		param = append(param,
			params.StateParam{
				Address: govNCPAddress,
				Key:     valueSlot,
				Value:   common.BigToHash(newLength),
			},
			params.StateParam{
				Address: govNCPAddress,
				Key:     common.HexToHash(SLOT_NCP_LAST_ID),
				Value:   common.BigToHash(ncpID),
			})
	}
	return param
}

func NCPStakers(govStakingAddress, govNCPAddress common.Address, state StateReader) []common.Address {
	stakers := make([]common.Address, 0)
	ncps := NCPList(govNCPAddress, state)
	for _, ncp := range ncps {
		v := StakerByOperator(govStakingAddress, state, ncp)
		if v != (common.Address{}) {
			stakers = append(stakers, v)
		}
	}
	return stakers
}

func NCPTotalStaking(govStakingAddress, govNCPAddress common.Address, state StateReader) *big.Int {
	totalStaking := new(big.Int)
	stakers := NCPStakers(govStakingAddress, govNCPAddress, state)
	for _, v := range stakers {
		totalStaking.Add(totalStaking, GetTotalStaked(govStakingAddress, state, v))
	}
	return totalStaking
}

func NCPStakerInfoMap(govStakingAddress, govNCPAddress common.Address, state StateReader) map[common.Address]Staker {
	stakerInfos := make(map[common.Address]Staker)
	stakers := NCPStakers(govStakingAddress, govNCPAddress, state)
	for _, v := range stakers {
		stakerInfos[v] = StakerInfo(govStakingAddress, state, v)
	}
	return stakerInfos
}

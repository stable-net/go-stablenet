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
	_ "embed"
)

const (
	GOV_CONTRACT_VERSION_1 = "v1"
	GOV_CONTRACT_VERSION_2 = "v2"

	CONTRACT_GOV_STAKING      = "GovStaking"
	CONTRACT_GOV_NCP          = "GovNCP"
	CONTRACT_GOV_CONFIG       = "GovConfig"
	CONTRACT_GOV_REWARDEE_IMP = "GovRewardeeImp"
	CONTRACT_GOV_REWARDEE     = "GovRewardee"
	CONTRACT_MULTISIG_WALLET  = "MultiSigWallet"
	CONTRACT_OPERATOR_SAMPLE  = "OperatorSample"

	GOV_CONFIG_PARAM_MINIMUM_STAKING     = "minimumStaking"
	GOV_CONFIG_PARAM_MAXIMUM_STAKING     = "maximumStaking"
	GOV_CONFIG_PARAM_UNBONDING_STAKER    = "unbondingPeriodStaker"
	GOV_CONFIG_PARAM_UNBONDING_DELEGATOR = "unbondingPeriodDelegator"
	GOV_CONFIG_PARAM_FEE_PRECISION       = "feePrecision"
	GOV_CONFIG_PARAM_CHANGE_FEE_DELAY    = "changeFeeDelay"
	GOV_CONFIG_PARAM_GOV_COUNCIL         = "govCouncil"

	GOV_NCP_PARAM_NCPS = "ncps"

	SLOT_GOV_CONFIG_MINIMUM_STAKING     = "0x0" //
	SLOT_GOV_CONFIG_MAXIMUM_STAKING     = "0x1" //
	SLOT_GOV_CONFIG_UNBONDING_STAKER    = "0x2" //
	SLOT_GOV_CONFIG_UNBONDING_DELEGATOR = "0x3" //
	SLOT_GOV_CONFIG_FEE_PRECISION       = "0x4" //
	SLOT_GOV_CONFIG_CHANGE_FEE_DELAY    = "0x5" //
	SLOT_GOV_CONFIG_GOV_COUNCIL         = "0x6" //
)

var (
	//go:embed govcontracts/v1/GovStaking
	GovStakingContractV1 string
	//go:embed govcontracts/v1/GovNCP
	GovNCPContractV1 string
	//go:embed govcontracts/v1/GovConfig
	GovConfigContractV1 string
	//go:embed govcontracts/v1/GovRewardeeImp
	GovRewardeeImpContractV1 string

	//go:embed govcontracts/v2/GovStaking
	GovStakingContractV2 string

	GovContractCodes map[string]map[string]string
)

func init() {
	GovContractCodes = make(map[string]map[string]string)

	GovContractCodes[CONTRACT_GOV_CONFIG] = make(map[string]string)
	GovContractCodes[CONTRACT_GOV_NCP] = make(map[string]string)
	GovContractCodes[CONTRACT_GOV_STAKING] = make(map[string]string)
	GovContractCodes[CONTRACT_GOV_REWARDEE_IMP] = make(map[string]string)

	GovContractCodes[CONTRACT_GOV_CONFIG][GOV_CONTRACT_VERSION_1] = GovConfigContractV1
	GovContractCodes[CONTRACT_GOV_NCP][GOV_CONTRACT_VERSION_1] = GovNCPContractV1
	GovContractCodes[CONTRACT_GOV_STAKING][GOV_CONTRACT_VERSION_1] = GovStakingContractV1
	GovContractCodes[CONTRACT_GOV_REWARDEE_IMP][GOV_CONTRACT_VERSION_1] = GovRewardeeImpContractV1
	GovContractCodes[CONTRACT_GOV_STAKING][GOV_CONTRACT_VERSION_2] = GovStakingContractV2
}

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
	_ "embed"
)

const (
	SYSTEM_CONTRACT_VERSION_1 = "v1"
	SYSTEM_CONTRACT_VERSION_2 = "v2"

	CONTRACT_GOV_VALIDATOR     = "GovValidator"
	CONTRACT_COIN_ADAPTER      = "NativeCoinAdapter"
	CONTRACT_GOV_MINTER        = "GovMinter"
	CONTRACT_GOV_MASTER_MINTER = "GovMasterMinter"
	CONTRACT_GOV_COUNCIL       = "GovCouncil"
)

var (
	//go:embed artifacts/v1/GovValidator
	GovValidatorContractV1 string

	//go:embed artifacts/v1/NativeCoinAdapter
	CoinAdapterContractV1 string

	//go:embed artifacts/v1/GovMinter
	GovMinterContractV1 string

	//go:embed artifacts/v1/GovMasterMinter
	GovMasterMinterContractV1 string

	//go:embed artifacts/v1/GovCouncil
	GovCouncilContractV1 string

	SystemContractCodes map[string]map[string]string
)

func init() {
	SystemContractCodes = make(map[string]map[string]string)

	SystemContractCodes[CONTRACT_GOV_VALIDATOR] = make(map[string]string)
	SystemContractCodes[CONTRACT_GOV_VALIDATOR][SYSTEM_CONTRACT_VERSION_1] = GovValidatorContractV1

	SystemContractCodes[CONTRACT_COIN_ADAPTER] = make(map[string]string)
	SystemContractCodes[CONTRACT_COIN_ADAPTER][SYSTEM_CONTRACT_VERSION_1] = CoinAdapterContractV1

	SystemContractCodes[CONTRACT_GOV_MINTER] = make(map[string]string)
	SystemContractCodes[CONTRACT_GOV_MINTER][SYSTEM_CONTRACT_VERSION_1] = GovMinterContractV1

	SystemContractCodes[CONTRACT_GOV_MASTER_MINTER] = make(map[string]string)
	SystemContractCodes[CONTRACT_GOV_MASTER_MINTER][SYSTEM_CONTRACT_VERSION_1] = GovMasterMinterContractV1

	SystemContractCodes[CONTRACT_GOV_COUNCIL] = make(map[string]string)
	SystemContractCodes[CONTRACT_GOV_COUNCIL][SYSTEM_CONTRACT_VERSION_1] = GovCouncilContractV1
}

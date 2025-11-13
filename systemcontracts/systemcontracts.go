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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"

	"fmt"
)

func init() {
	// to avoid import cycle
	params.CheckSystemContractVersions = checkSystemContractVersions
}

func checkSystemContractVersions(systemContracts *params.SystemContracts) error {
	if SystemContractCodes[CONTRACT_GOV_VALIDATOR][systemContracts.GovValidator.Version] == "" {
		return fmt.Errorf("`systemContracts.govValidator`: unsupported version %s", systemContracts.GovValidator.Version)
	}

	if SystemContractCodes[CONTRACT_COIN_ADAPTER][systemContracts.NativeCoinAdapter.Version] == "" {
		return fmt.Errorf("`systemContracts.nativeCoinAdapter`: unsupported version %s", systemContracts.NativeCoinAdapter.Version)
	}

	if systemContracts.GovMinter != nil {
		if SystemContractCodes[CONTRACT_GOV_MINTER][systemContracts.GovMinter.Version] == "" {
			return fmt.Errorf("`systemContracts.govMinter`: unsupported version %s", systemContracts.GovMinter.Version)
		}
	}

	if systemContracts.GovMasterMinter != nil {
		if SystemContractCodes[CONTRACT_GOV_MASTER_MINTER][systemContracts.GovMasterMinter.Version] == "" {
			return fmt.Errorf("`systemContracts.govMasterMinter`: unsupported version %s", systemContracts.GovMasterMinter.Version)
		}
	}

	if systemContracts.GovCouncil != nil {
		if SystemContractCodes[CONTRACT_GOV_COUNCIL][systemContracts.GovCouncil.Version] == "" {
			return fmt.Errorf("`systemContracts.govCouncil`: unsupported version %s", systemContracts.GovCouncil.Version)
		}
	}

	return nil
}

func GetSystemContractsTransition(systemContracts *params.SystemContracts, alloc *types.GenesisAlloc) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if systemContracts.GovValidator != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovValidator.Address, Code: SystemContractCodes[CONTRACT_GOV_VALIDATOR][systemContracts.GovValidator.Version]})
		sp, err := initializeValidator(systemContracts.GovValidator.Address, systemContracts.GovValidator.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}

	if systemContracts.NativeCoinAdapter != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.NativeCoinAdapter.Address, Code: SystemContractCodes[CONTRACT_COIN_ADAPTER][systemContracts.NativeCoinAdapter.Version]})
		sp, err := initializeCoinAdapter(systemContracts.NativeCoinAdapter.Address, systemContracts.NativeCoinAdapter.Params, alloc)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}

	if systemContracts.GovMinter != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovMinter.Address, Code: SystemContractCodes[CONTRACT_GOV_MINTER][systemContracts.GovMinter.Version]})
		sp, err := initializeMinter(systemContracts.GovMinter.Address, systemContracts.GovMinter.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}

	if systemContracts.GovMasterMinter != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovMasterMinter.Address, Code: SystemContractCodes[CONTRACT_GOV_MASTER_MINTER][systemContracts.GovMasterMinter.Version]})
		sp, err := initializeMasterMinter(systemContracts.GovMasterMinter.Address, systemContracts.GovMasterMinter.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}

	if systemContracts.GovCouncil != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovCouncil.Address, Code: SystemContractCodes[CONTRACT_GOV_COUNCIL][systemContracts.GovCouncil.Version]})
		sp, err := initializeGovCouncil(systemContracts.GovCouncil.Address, systemContracts.GovCouncil.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}

	return st, nil
}

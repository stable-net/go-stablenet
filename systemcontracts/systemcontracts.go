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
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
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

// GetSystemContractsTransition builds a StateTransition for the given SystemContracts.
//
// Upgrade principle — Code only:
// Hardfork upgrades ONLY deploy new contract code. On-chain state (storage slots)
// is NEVER modified during an upgrade. State values set at genesis (v1) or changed
// through governance proposals are always preserved as-is.
//
// Version-based behavior:
//   - "v1": Code deployment + State initialization (genesis only)
//   - other: Code deployment only, no state changes
//
// IMPORTANT: Each contract's initial version MUST be "v1". The initialize*() functions
// (e.g., initializeValidator, initializeCoinAdapter) are only invoked when Version == "v1",
// and they set up the required storage layout (owner, quorum, members, etc.).
// Starting with any other version will skip initialization, leaving the contract in an
// uninitialized state with empty storage — which will cause runtime failures.
func GetSystemContractsTransition(systemContracts *params.SystemContracts, alloc *types.GenesisAlloc) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if systemContracts.GovValidator != nil {
		code, err := getContractCode(CONTRACT_GOV_VALIDATOR, systemContracts.GovValidator.Version)
		if err != nil {
			return nil, err
		}
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovValidator.Address, Code: code})
		if systemContracts.GovValidator.Params != nil && systemContracts.GovValidator.Version == SYSTEM_CONTRACT_VERSION_1 {
			sp, err := initializeValidator(systemContracts.GovValidator.Address, systemContracts.GovValidator.Params)
			if err != nil {
				return nil, err
			}
			st.States = append(st.States, sp...)
		}
	}

	if systemContracts.NativeCoinAdapter != nil {
		code, err := getContractCode(CONTRACT_COIN_ADAPTER, systemContracts.NativeCoinAdapter.Version)
		if err != nil {
			return nil, err
		}
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.NativeCoinAdapter.Address, Code: code})
		if systemContracts.NativeCoinAdapter.Params != nil && systemContracts.NativeCoinAdapter.Version == SYSTEM_CONTRACT_VERSION_1 {
			sp, err := initializeCoinAdapter(systemContracts.NativeCoinAdapter.Address, systemContracts.NativeCoinAdapter.Params, alloc)
			if err != nil {
				return nil, err
			}
			st.States = append(st.States, sp...)
		}
	}

	if systemContracts.GovMinter != nil {
		code, err := getContractCode(CONTRACT_GOV_MINTER, systemContracts.GovMinter.Version)
		if err != nil {
			return nil, err
		}
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovMinter.Address, Code: code})
		if systemContracts.GovMinter.Params != nil && systemContracts.GovMinter.Version == SYSTEM_CONTRACT_VERSION_1 {
			sp, err := initializeMinter(systemContracts.GovMinter.Address, systemContracts.GovMinter.Params)
			if err != nil {
				return nil, err
			}
			st.States = append(st.States, sp...)
		}
	}

	if systemContracts.GovMasterMinter != nil {
		code, err := getContractCode(CONTRACT_GOV_MASTER_MINTER, systemContracts.GovMasterMinter.Version)
		if err != nil {
			return nil, err
		}
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovMasterMinter.Address, Code: code})
		if systemContracts.GovMasterMinter.Params != nil && systemContracts.GovMasterMinter.Version == SYSTEM_CONTRACT_VERSION_1 {
			sp, err := initializeMasterMinter(systemContracts.GovMasterMinter.Address, systemContracts.GovMasterMinter.Params)
			if err != nil {
				return nil, err
			}
			st.States = append(st.States, sp...)
		}
	}

	if systemContracts.GovCouncil != nil {
		code, err := getContractCode(CONTRACT_GOV_COUNCIL, systemContracts.GovCouncil.Version)
		if err != nil {
			return nil, err
		}
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovCouncil.Address, Code: code})
		if systemContracts.GovCouncil.Params != nil && systemContracts.GovCouncil.Version == SYSTEM_CONTRACT_VERSION_1 {
			sp, err := initializeGovCouncil(systemContracts.GovCouncil.Address, systemContracts.GovCouncil.Params, alloc)
			if err != nil {
				return nil, err
			}
			st.States = append(st.States, sp...)
		}
	}

	return st, nil
}

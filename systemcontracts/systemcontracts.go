// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// Copyright 2025 The stable-one Authors
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

package systemcontracts

import (
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
	return nil
}

func GetSystemContractsTransition(systemContracts *params.SystemContracts) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if systemContracts.GovValidator != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: systemContracts.GovValidator.Address, Code: SystemContractCodes[CONTRACT_GOV_VALIDATOR][systemContracts.GovValidator.Version]})
		sp, err := initializeValidator(systemContracts.GovValidator.Address, systemContracts.GovValidator.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}
	return st, nil
}

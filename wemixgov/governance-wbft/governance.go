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

package govwbft

import (
	"github.com/ethereum/go-ethereum/params"

	"fmt"
)

func init() {
	// to avoid import cycle
	params.CheckGovContractVersions = checkGovContractVersions
}

func checkGovContractVersions(govContracts *params.GovContracts) error {
	if GovContractCodes[CONTRACT_GOV_VALIDATOR][govContracts.GovValidator.Version] == "" {
		return fmt.Errorf("`govContracts.govConfig`: unsupported version %s", govContracts.GovValidator.Version)
	}
	return nil
}

func GetGovContractsTransition(govContracts *params.GovContracts) (*params.StateTransition, error) {
	st := &params.StateTransition{}

	if govContracts.GovValidator != nil {
		st.Codes = append(st.Codes, params.CodeParam{Address: govContracts.GovValidator.Address, Code: GovContractCodes[CONTRACT_GOV_VALIDATOR][govContracts.GovValidator.Version]})
		sp, err := initializeValidator(govContracts.GovValidator.Address, govContracts.GovValidator.Params)
		if err != nil {
			return nil, err
		}
		st.States = append(st.States, sp...)
	}
	return st, nil
}

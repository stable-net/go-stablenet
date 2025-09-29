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
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/params"
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
		memberAddresses := strings.Split(govContracts.GovValidator.Params[GOV_BASE_PARAM_MEMBERS], ",")
		members := make([]common.Address, 0)
		for _, memberAddr := range memberAddresses {
			members = append(members, common.HexToAddress(memberAddr))
		}
		valAddresses := strings.Split(govContracts.GovValidator.Params[GOV_VALIDATOR_PARAM_VALIDATORS], ",")
		validators := make([]common.Address, 0)
		for _, valAddr := range valAddresses {
			validators = append(validators, common.HexToAddress(valAddr))
		}
		blsKeyStrings := strings.Split(govContracts.GovValidator.Params[GOV_VALIDATOR_PARAM_BLS_KEYS], ",")

		if len(members) != len(validators) {
			return nil, fmt.Errorf("the number of members and validators must be the same")
		}
		if len(validators) != len(blsKeyStrings) {
			return nil, fmt.Errorf("the number of validators and BLS public keys must be the same")
		}

		quorum, err := strconv.ParseUint(govContracts.GovValidator.Params[GOV_BASE_PARAM_QUORUM], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("`govContracts.govValidator.params.quorum`: %w", err)
		}
		expiry, err2 := strconv.ParseUint(govContracts.GovValidator.Params[GOV_BASE_PARAM_EXPIRY], 10, 64)
		if err2 != nil {
			return nil, fmt.Errorf("`govContracts.govValidator.params.expiry`: %w", err2)
		}
		blsKeys := make([]bls.PublicKey, 0)
		for _, key := range blsKeyStrings {
			pk, err3 := bls.PublicKeyFromBytes(common.Hex2Bytes(key))
			if err3 != nil {
				return nil, fmt.Errorf("invalid BLS public key: %w", err3)
			}
			blsKeys = append(blsKeys, pk)
		}
		st.States = append(st.States,
			initializeValidator(
				govContracts.GovValidator.Address,
				members,
				validators,
				blsKeys,
				quorum,
				expiry)...)
	}
	return st, nil
}

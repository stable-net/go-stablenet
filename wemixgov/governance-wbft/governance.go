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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
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
		if v != (common.Address{}) && IsStaker(govStakingAddress, state, v) {
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

// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

package govwbft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	GOV_VALIDATOR_PARAM_VALIDATORS = "validators"
	GOV_VALIDATOR_PARAM_BLS_KEYS   = "blsPublicKeys"

	SLOT_VALIDATOR_blsPop              = "0x32"
	SLOT_VALIDATOR_validators          = "0x33"
	SLOT_VALIDATOR_validatorToOperator = "0x35"
	SLOT_VALIDATOR_operatorToValidator = "0x36"
	SLOT_VALIDATOR_validatorToBlsKey   = "0x37"
	SLOT_VALIDATOR_blsKeyToValidator   = "0x38"
)

type blsWrapper struct {
	blsKey []byte
}

func (w *blsWrapper) Bytes() []byte {
	return w.blsKey
}

func initializeValidator(govValidatorAddress common.Address, members []common.Address, validators []common.Address, blsKeys [][]byte, quorum uint64, expiry uint64) []params.StateParam {
	param := initializeBase(govValidatorAddress, members, quorum, expiry)

	param = append(param,
		params.StateParam{
			Address: govValidatorAddress,
			Key:     common.HexToHash(SLOT_VALIDATOR_blsPop),
			Value:   common.BytesToHash(params.BLSPoPPrecompileAddress.Bytes())},
	)

	valueSlot := common.HexToHash(SLOT_VALIDATOR_validators)
	indexSlot := IncrementHash(valueSlot, big.NewInt(1))
	duplicated := make(map[common.Address]struct{})

	currentIdx := uint64(0)
	newLength := new(big.Int)

	for i, val := range validators {
		if _, ok := duplicated[val]; ok {
			continue
		}
		newLength = new(big.Int).SetUint64(currentIdx + 1)

		param = append(param,
			// set index slot
			params.StateParam{
				Address: govValidatorAddress,
				Key:     CalculateMappingSlot(indexSlot, val),
				Value:   common.BigToHash(newLength),
			},
			// set value slot
			params.StateParam{
				Address: govValidatorAddress,
				Key:     CalculateDynamicSlot(valueSlot, new(big.Int).SetUint64(currentIdx)),
				Value:   common.BytesToHash(val.Bytes()),
			},

			// validator to operator(member) mapping
			params.StateParam{
				Address: govValidatorAddress,
				Key:     CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_validatorToOperator), val),
				Value:   common.BytesToHash(members[i].Bytes()),
			},
			// operator(member) to validator mapping
			params.StateParam{
				Address: govValidatorAddress,
				Key:     CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_operatorToValidator), members[i]),
				Value:   common.BytesToHash(val.Bytes()),
			},
		)

		param = append(param, MakeMultipleParam(govValidatorAddress, CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_validatorToBlsKey), val), VarLenBytesToMultipleHash(blsKeys[i]))...)

		param = append(param,
			params.StateParam{
				Address: govValidatorAddress,
				Key: CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_blsKeyToValidator), &blsWrapper{
					blsKey: blsKeys[i],
				}),
				Value: common.BytesToHash(val.Bytes()),
			},
		)

		duplicated[val] = struct{}{}
		currentIdx++
	}
	if newLength.Sign() > 0 {
		param = append(param,
			params.StateParam{
				Address: govValidatorAddress,
				Key:     valueSlot,
				Value:   common.BigToHash(newLength),
			},
		)
	}

	return param
}

func MakeMultipleParam(govValidatorAddress common.Address, baseSlot common.Hash, value []common.Hash) []params.StateParam {
	result := make([]params.StateParam, 0)
	result = append(result, params.StateParam{
		Address: govValidatorAddress,
		Key:     baseSlot,
		Value:   value[0],
	})
	for i := uint64(1); i < uint64(len(value)); i++ {
		result = append(result, params.StateParam{
			Address: govValidatorAddress,
			Key:     CalculateDynamicSlot(baseSlot, big.NewInt(int64(i-1))),
			Value:   value[i],
		})
	}
	return result
}

func ValidatorList(govValidatorAddress common.Address, state StateReader) []common.Address {
	ncpSet := NewAddressSet(common.HexToHash(SLOT_VALIDATOR_validators))
	return ncpSet.Values(state, govValidatorAddress)
}

func GetBLSPublicKey(govValidatorAddress common.Address, state StateReader, val common.Address) []byte {
	return GetBytes(state, govValidatorAddress, CalculateMappingSlot(common.HexToHash(SLOT_VALIDATOR_validatorToBlsKey), val))
}

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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

func TestInitializeValidator(t *testing.T) {
	sampleMemberAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	sampleBlsKey := "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b"
	derivedKeyHashForMembers := common.HexToHash("0xbf7ef88dc1e3e88da4ae9b3f61a5fa07847ffcf591e366db12c183df7df576c3")

	testCases := []struct {
		name        string
		param       map[string]string
		expectErr   string
		expectParam []params.StateParam
	}{
		{
			name:      "empty param",
			param:     map[string]string{},
			expectErr: "",
			expectParam: []params.StateParam{
				{
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000032"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000b00001"),
				},
			},
		},
		{
			name: "1 member, no validator",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
			},
			expectErr: "",
			expectParam: []params.StateParam{
				{ // members
					Address: common.Address{},
					Key:     derivedKeyHashForMembers,
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // versionedMemberList.members
					Address: common.Address{},
					Key:     common.HexToHash("0x2c644dcf44e265ba93879b2da89e1b16ab48fc5eb8e31bc16b0612d6da8463f1"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // versionedMemberList.length
					Address: common.Address{},
					Key:     common.HexToHash("0xa15bc60c955c405d20d9149c709e2460f1c2d9a497496a7f46004d1772c3054c"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // memberVersion
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // blsPop
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000032"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000b00001"),
				},
			},
		},
		{
			name: "1 member, 1 validator, no bls key",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:         sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION:  "1",
				GOV_VALIDATOR_PARAM_VALIDATORS: sampleMemberAddress,
			},
			expectErr:   "`systemContracts.govValidator.params`: missing parameter: blsPublicKeys",
			expectParam: nil,
		},
		{
			name: "1 member, no validator, 1 bls key",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_VALIDATOR_PARAM_BLS_KEYS:  sampleBlsKey,
			},
			expectErr:   "`systemContracts.govValidator.params`: missing parameter: validators",
			expectParam: nil,
		},
		{
			name: "1 member, 2 validator, 1 bls key",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:         sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION:  "1",
				GOV_VALIDATOR_PARAM_VALIDATORS: sampleMemberAddress + "," + sampleMemberAddress,
				GOV_VALIDATOR_PARAM_BLS_KEYS:   sampleBlsKey,
			},
			expectErr:   "`systemContracts.govValidator.params`: the number of members and validators must be the same",
			expectParam: nil,
		},
		{
			name: "1 member, 1 validator, 2 bls key",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:         sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION:  "1",
				GOV_VALIDATOR_PARAM_VALIDATORS: sampleMemberAddress,
				GOV_VALIDATOR_PARAM_BLS_KEYS:   sampleBlsKey + "," + sampleBlsKey,
			},
			expectErr:   "`systemContracts.govValidator.params`: the number of validators and BLS public keys must be the same",
			expectParam: nil,
		},
		{
			name: "1 member, 1 validator, 1 bls key",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:         sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION:  "1",
				GOV_VALIDATOR_PARAM_VALIDATORS: sampleMemberAddress,
				GOV_VALIDATOR_PARAM_BLS_KEYS:   sampleBlsKey,
			},
			expectErr: "",
			expectParam: []params.StateParam{
				{ // members
					Address: common.Address{},
					Key:     derivedKeyHashForMembers,
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // versionedMemberList.members
					Address: common.Address{},
					Key:     common.HexToHash("0x2c644dcf44e265ba93879b2da89e1b16ab48fc5eb8e31bc16b0612d6da8463f1"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // versionedMemberList.length
					Address: common.Address{},
					Key:     common.HexToHash("0xa15bc60c955c405d20d9149c709e2460f1c2d9a497496a7f46004d1772c3054c"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // memberVersion
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // blsPop
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000032"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000b00001"),
				},
				{ // __validators.index
					Address: common.Address{},
					Key:     common.HexToHash("0x5d156553fedc0e3ad6b77dfb4190223d769a4e8575263d506d55e35ca385ec4f"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // __validators.value(validator)
					Address: common.Address{},
					Key:     common.HexToHash("0x82a75bdeeae8604d839476ae9efd8b0e15aa447e21bfd7f41283bb54e22c9a82"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // validatorToOperator
					Address: common.Address{},
					Key:     common.HexToHash("0x72d3e02218551170037da0841c2a16050467f113cb761dcd5ea0d4edd206e3c7"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // operatorToValidator
					Address: common.Address{},
					Key:     common.HexToHash("0xb47b937a548fdbc8eeb6153348801b91dd067e9110633c5d95d8fde2c500b131"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // validatorToBlsKey.length | 0x01
					Address: common.Address{},
					Key:     common.HexToHash("0x3de6e5bb5ca8d1f2605fc1b641bad3e8725ac54e8e7ef4222ef9361a93df8491"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000061"),
				},
				{ // validatorToBlsKey
					Address: common.Address{},
					Key:     common.HexToHash("0x704ccd9af691691ff31c7b42662363b6c5d56eb9d93ac624e6cceda3b8b9af77"),
					Value:   common.HexToHash("0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6f"),
				},
				{ // validatorToBlsKey
					Address: common.Address{},
					Key:     common.HexToHash("0x704ccd9af691691ff31c7b42662363b6c5d56eb9d93ac624e6cceda3b8b9af78"),
					Value:   common.HexToHash("0xbe4f4adc21e1e454bc37522eb3478b9b00000000000000000000000000000000"),
				},
				{ // blsKeyToValidator
					Address: common.Address{},
					Key:     common.HexToHash("0xbe042d13e4dc3c69d08493aab6f511fa8f0029eacc43ede3af636620ce697bc8"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // __validators.length
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000033"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := initializeValidator(common.Address{}, tc.param)
			if tc.expectErr != "" {
				if err == nil || err.Error() != tc.expectErr {
					t.Errorf("[test=%v] expected error: %v, got: %v", tc.name, tc.expectErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("[test=%v] unexpected error: %v", tc.name, err)
				}
				if len(result) != len(tc.expectParam) {
					t.Errorf("[test=%v] expected params length: %d, got: %d", tc.name, len(tc.expectParam), len(result))
				}
				for i, v := range tc.expectParam {
					CompareParam(result[i], v, t, tc.name)
				}
			}
		})
	}
}

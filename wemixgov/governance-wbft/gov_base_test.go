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

var anyHash common.Hash

func init() {
	anyHash = common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234")
}

func CompareParam(result params.StateParam, expect params.StateParam, t *testing.T, testName string) {
	if result.Address != expect.Address {
		t.Errorf("[test=%v] Address mismatch: result %s, expect %s", testName, result.Address.Hex(), expect.Address.Hex())
	}
	if result.Key != expect.Key {
		t.Errorf("[test=%v] Key mismatch: result %s, expect %s", testName, result.Key.Hex(), expect.Key.Hex())
	}
	if expect.Value != anyHash && result.Value != expect.Value {
		t.Errorf("[test=%v] Value mismatch: result %s, expect %s", testName, result.Value.Hex(), expect.Value.Hex())
	}
}

func TestInitializeBase(t *testing.T) {
	sampleMemberAddress := "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"
	derivedKeyHashForMembers := common.HexToHash("0xbf7ef88dc1e3e88da4ae9b3f61a5fa07847ffcf591e366db12c183df7df576c3")

	testCases := []struct {
		name        string
		param       map[string]string
		expectErr   string
		expectParam []params.StateParam
	}{
		{
			name:        "empty param",
			param:       map[string]string{},
			expectErr:   "",
			expectParam: []params.StateParam{},
		},
		{
			name: "1 member, no version",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS: sampleMemberAddress,
			},
			expectErr:   "`govContracts.govBase.params.memberVersion` is required when `govContracts.govBase.params.members` is set",
			expectParam: nil,
		},
		{
			name: "version not number",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "v2",
			},
			expectErr:   "`govContracts.govBase.params.memberVersion`: strconv.ParseUint: parsing \"v2\": invalid syntax",
			expectParam: nil,
		},
		{
			name: "1 member, version",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "2",
			},
			expectErr: "",
			expectParam: []params.StateParam{
				{ // members
					Address: common.Address{},
					Key:     derivedKeyHashForMembers,
					Value:   anyHash,
				},
				{ // versionedMemberList.members
					Address: common.Address{},
					Key:     common.HexToHash("0x19b5847ec9d8983e32da86b2c2bedc7b0bcabd1d214557fda78706fe7ba568ce"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // versionedMemberList.length
					Address: common.Address{},
					Key:     common.HexToHash("0xc3a24b0501bd2c13a7e57f2db4369ec4c223447539fc0724a9d55ac4a06ebd4d"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // memberVersion
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				},
			},
		},
		{
			name: "1 member, version, quorum",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "2",
				GOV_BASE_PARAM_QUORUM:         "1",
			},
			expectErr: "",
			expectParam: []params.StateParam{
				{ // quorum
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // members
					Address: common.Address{},
					Key:     derivedKeyHashForMembers,
					Value:   anyHash,
				},
				{ // versionedMemberList.members
					Address: common.Address{},
					Key:     common.HexToHash("0x19b5847ec9d8983e32da86b2c2bedc7b0bcabd1d214557fda78706fe7ba568ce"),
					Value:   common.HexToHash("0x000000000000000000000000abcdefabcdefabcdefabcdefabcdefabcdefabcd"),
				},
				{ // versionedMemberList.length
					Address: common.Address{},
					Key:     common.HexToHash("0xc3a24b0501bd2c13a7e57f2db4369ec4c223447539fc0724a9d55ac4a06ebd4d"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // memberVersion
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000004"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002"),
				},
			},
		},
		{
			name: "too large quorum",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "2",
				GOV_BASE_PARAM_QUORUM:         "2",
			},
			expectErr:   "`govContracts.govBase.params.quorum` must not be greater than the number of members",
			expectParam: nil,
		},
		{
			name: "quorum not number",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "2",
				GOV_BASE_PARAM_QUORUM:         "2.2",
			},
			expectErr:   "`govContracts.govBase.params.quorum`: strconv.ParseUint: parsing \"2.2\": invalid syntax",
			expectParam: nil,
		},
		{
			name: "expiry not number",
			param: map[string]string{
				GOV_BASE_PARAM_EXPIRY: "T123",
			},
			expectErr:   "`govContracts.govBase.params.expiry`: strconv.ParseUint: parsing \"T123\": invalid syntax",
			expectParam: nil,
		},
		{
			name: "1 member, version, quorum, expiry",
			param: map[string]string{
				GOV_BASE_PARAM_MEMBERS:        sampleMemberAddress,
				GOV_BASE_PARAM_MEMBER_VERSION: "1",
				GOV_BASE_PARAM_EXPIRY:         "604800",
				GOV_BASE_PARAM_QUORUM:         "1",
			},
			expectErr: "",
			expectParam: []params.StateParam{
				{ // quorum
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
				},
				{ // expiry
					Address: common.Address{},
					Key:     common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
					Value:   common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000093a80"),
				},
				{ // members
					Address: common.Address{},
					Key:     derivedKeyHashForMembers,
					Value:   anyHash,
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
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := initializeBase(common.Address{}, tc.param)
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

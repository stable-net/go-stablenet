// Modification Copyright 2024 The Wemix Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/core/types/istanbul_test.go (2024.07.25).
// Modified and improved for the wemix development.

package types

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rlp"
)

// ## Quorum QBFT START
func TestHeaderHash(t *testing.T) {
	// 0x06f58ebb96f233516c4a0eeedad007e8b8e0965b24c88d7407eb08a6ee7bc4ff
	expectedExtra := common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000f89af8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b440b8410000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0")
	expectedHash := common.HexToHash("0x06f58ebb96f233516c4a0eeedad007e8b8e0965b24c88d7407eb08a6ee7bc4ff")

	// for istanbul consensus
	header := &Header{Difficulty: QBFTDefaultDifficulty, Extra: expectedExtra}
	if !reflect.DeepEqual(header.Hash(), expectedHash) {
		t.Errorf("expected: %v, but got: %v", expectedHash.Hex(), header.Hash().Hex())
	}

	// append useless information to extra-data
	unexpectedExtra := append(expectedExtra, []byte{1, 2, 3}...)
	header.Extra = unexpectedExtra
	if !reflect.DeepEqual(header.Hash(), rlpHash(header)) {
		t.Errorf("expected: %v, but got: %v", rlpHash(header).Hex(), header.Hash().Hex())
	}
}

func TestExtractToQBFTExtra(t *testing.T) {
	testCases := []struct {
		istRawData     []byte
		expectedResult *QBFTExtra
		expectedErr    error
	}{
		{
			// normal case
			hexutil.MustDecode("0xf85a80f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b440c080c0"),
			&QBFTExtra{
				VanityData: []byte{},
				Validators: []common.Address{
					common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
					common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
					common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
					common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440")),
				},
				CommittedSeal: [][]byte{},
				Round:         0,
				Vote:          nil,
			},
			nil,
		},
	}
	for _, test := range testCases {
		h := &Header{Extra: test.istRawData}
		istanbulExtra, err := ExtractQBFTExtra(h)
		if err != test.expectedErr {
			t.Errorf("expected: %v, but got: %v", test.expectedErr, err)
		}
		if !reflect.DeepEqual(istanbulExtra, test.expectedResult) {
			t.Errorf("expected: %v, but got: %v", test.expectedResult, istanbulExtra)
		}
	}
}

func TestGenerateExtra(t *testing.T) {
	sampleExtra := &QBFTExtra{
		VanityData: []byte("WEMIX MontBlanc chain block"),
		Validators: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697")),
		},
		Round: 0,
	}
	b, _ := rlp.EncodeToBytes(sampleExtra)
	t.Logf("extra bytes: %v\n", hexutil.Encode(b))
}

// ## Quorum QBFT END

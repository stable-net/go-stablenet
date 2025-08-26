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

package testutils

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestGeneratingGenesisExtra(t *testing.T) {
	WBFTExtra := &types.WBFTExtra{
		VanityData:        []byte{},
		RandaoReveal:      []byte{},
		PrevRound:         0,
		PreparedSeal:      &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		PrevPreparedSeal:  &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		Round:             0,
		CommittedSeal:     &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		PrevCommittedSeal: &types.WBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		EpochInfo: &types.EpochInfo{
			Stakers: []*types.Staker{
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xa62987A40E094CbE020313f19F71aeeB3E48B86f")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xE1bD8108149FCa703B9FCD8ea967B6b7e660f13b")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0xd19f9374f4549B2fB182ED766d6b7501494a3634")), Diligence: types.DefaultDiligence},
				{Addr: common.BytesToAddress(hexutil.MustDecode("0x02cF1E577C79EF0E93947cCd82a4D41E0485Be73")), Diligence: types.DefaultDiligence},
			},
			Stabilizing:   true,
			Validators:    []uint32{0, 1, 2, 3},
			BLSPublicKeys: [][]byte{{}, {}, {}, {}},
		},
	}
	genesis := GenesisWithSeals(WBFTExtra.EpochInfo.GetStakers(), WBFTExtra.EpochInfo.BLSPublicKeys)
	t.Log("Genesis Extra Data: ", hex.EncodeToString(genesis.ExtraData))
	wbftExtra := new(types.WBFTExtra)
	err := rlp.DecodeBytes(genesis.ExtraData[:], wbftExtra)
	if err != nil {
		t.Errorf("Failed to decode genesis wbft extra data : %v", err)
	}
	wbftExtra.VanityData = []byte{} // clear vanity
	if !reflect.DeepEqual(WBFTExtra, wbftExtra) {
		t.Errorf("decoded extra object is different from origin(decoded=%v, origin=%v)", WBFTExtra, wbftExtra)
	}
}

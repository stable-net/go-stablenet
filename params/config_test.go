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

package params

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

func TestCheckCompatible(t *testing.T) {
	type test struct {
		stored, new   *ChainConfig
		headBlock     uint64
		headTimestamp uint64
		wantErr       *ConfigCompatError
	}
	tests := []test{
		{stored: AllEthashProtocolChanges, new: AllEthashProtocolChanges, headBlock: 0, headTimestamp: 0, wantErr: nil},
		{stored: AllEthashProtocolChanges, new: AllEthashProtocolChanges, headBlock: 0, headTimestamp: uint64(time.Now().Unix()), wantErr: nil},
		{stored: AllEthashProtocolChanges, new: AllEthashProtocolChanges, headBlock: 100, wantErr: nil},
		{
			stored:    &ChainConfig{EIP150Block: big.NewInt(10)},
			new:       &ChainConfig{EIP150Block: big.NewInt(20)},
			headBlock: 9,
			wantErr:   nil,
		},
		{
			stored:    AllEthashProtocolChanges,
			new:       &ChainConfig{HomesteadBlock: nil},
			headBlock: 3,
			wantErr: &ConfigCompatError{
				What:          "Homestead fork block",
				StoredBlock:   big.NewInt(0),
				NewBlock:      nil,
				RewindToBlock: 0,
			},
		},
		{
			stored:    AllEthashProtocolChanges,
			new:       &ChainConfig{HomesteadBlock: big.NewInt(1)},
			headBlock: 3,
			wantErr: &ConfigCompatError{
				What:          "Homestead fork block",
				StoredBlock:   big.NewInt(0),
				NewBlock:      big.NewInt(1),
				RewindToBlock: 0,
			},
		},
		{
			stored:    &ChainConfig{HomesteadBlock: big.NewInt(30), EIP150Block: big.NewInt(10)},
			new:       &ChainConfig{HomesteadBlock: big.NewInt(25), EIP150Block: big.NewInt(20)},
			headBlock: 25,
			wantErr: &ConfigCompatError{
				What:          "EIP150 fork block",
				StoredBlock:   big.NewInt(10),
				NewBlock:      big.NewInt(20),
				RewindToBlock: 9,
			},
		},
		{
			stored:    &ChainConfig{ConstantinopleBlock: big.NewInt(30)},
			new:       &ChainConfig{ConstantinopleBlock: big.NewInt(30), PetersburgBlock: big.NewInt(30)},
			headBlock: 40,
			wantErr:   nil,
		},
		{
			stored:    &ChainConfig{ConstantinopleBlock: big.NewInt(30)},
			new:       &ChainConfig{ConstantinopleBlock: big.NewInt(30), PetersburgBlock: big.NewInt(31)},
			headBlock: 40,
			wantErr: &ConfigCompatError{
				What:          "Petersburg fork block",
				StoredBlock:   nil,
				NewBlock:      big.NewInt(31),
				RewindToBlock: 30,
			},
		},
		{
			stored:        &ChainConfig{ShanghaiTime: newUint64(10)},
			new:           &ChainConfig{ShanghaiTime: newUint64(20)},
			headTimestamp: 9,
			wantErr:       nil,
		},
		{
			stored:        &ChainConfig{ShanghaiTime: newUint64(10)},
			new:           &ChainConfig{ShanghaiTime: newUint64(20)},
			headTimestamp: 25,
			wantErr: &ConfigCompatError{
				What:         "Shanghai fork timestamp",
				StoredTime:   newUint64(10),
				NewTime:      newUint64(20),
				RewindToTime: 9,
			},
		},
	}

	for _, test := range tests {
		err := test.stored.CheckCompatible(test.new, test.headBlock, test.headTimestamp)
		if !reflect.DeepEqual(err, test.wantErr) {
			t.Errorf("error mismatch:\nstored: %v\nnew: %v\nheadBlock: %v\nheadTimestamp: %v\nerr: %v\nwant: %v", test.stored, test.new, test.headBlock, test.headTimestamp, err, test.wantErr)
		}
	}
}

func TestConfigRules(t *testing.T) {
	c := &ChainConfig{
		LondonBlock:  new(big.Int),
		ShanghaiTime: newUint64(500),
	}
	var stamp uint64
	if r := c.Rules(big.NewInt(0), true, stamp); r.IsShanghai {
		t.Errorf("expected %v to not be shanghai", stamp)
	}
	stamp = 500
	if r := c.Rules(big.NewInt(0), true, stamp); !r.IsShanghai {
		t.Errorf("expected %v to be shanghai", stamp)
	}
	stamp = math.MaxInt64
	if r := c.Rules(big.NewInt(0), true, stamp); !r.IsShanghai {
		t.Errorf("expected %v to be shanghai", stamp)
	}
}

func TestCollectUpgrades_Order(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *ChainConfig
		wantCount int
		wantBlock int64
		wantVer   string
	}{
		{
			name: "Boho at block 0 (genesis overlay)",
			cfg: &ChainConfig{
				BohoBlock: big.NewInt(0),
				Boho: &AnzeonConfig{
					SystemContracts: &SystemContracts{
						GovMinter: &SystemContract{Address: common.HexToAddress("0x1003"), Version: "v2"},
					},
				},
			},
			wantCount: 1,
			wantBlock: 0,
			wantVer:   "v2",
		},
		{
			name: "Boho at block 100 (runtime upgrade)",
			cfg: &ChainConfig{
				BohoBlock: big.NewInt(100),
				Boho: &AnzeonConfig{
					SystemContracts: &SystemContracts{
						GovMinter: &SystemContract{Address: common.HexToAddress("0x1003"), Version: "v2"},
					},
				},
			},
			wantCount: 1,
			wantBlock: 100,
			wantVer:   "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upgrades := tt.cfg.CollectUpgrades()
			if len(upgrades) != tt.wantCount {
				t.Fatalf("expected %d upgrade(s), got %d", tt.wantCount, len(upgrades))
			}
			if upgrades[0].Block.Int64() != tt.wantBlock {
				t.Errorf("expected block %d, got %d", tt.wantBlock, upgrades[0].Block.Int64())
			}
			if upgrades[0].GovMinter.Version != tt.wantVer {
				t.Errorf("expected GovMinter %s, got %s", tt.wantVer, upgrades[0].GovMinter.Version)
			}
		})
	}
}

func TestCollectUpgrades_NilHandling(t *testing.T) {
	tests := []struct {
		name string
		cfg  *ChainConfig
		want int
	}{
		{
			name: "BohoBlock nil",
			cfg:  &ChainConfig{BohoBlock: nil, Boho: &AnzeonConfig{SystemContracts: &SystemContracts{}}},
			want: 0,
		},
		{
			name: "Boho nil",
			cfg:  &ChainConfig{BohoBlock: big.NewInt(0), Boho: nil},
			want: 0,
		},
		{
			name: "Boho.SystemContracts nil",
			cfg:  &ChainConfig{BohoBlock: big.NewInt(0), Boho: &AnzeonConfig{SystemContracts: nil}},
			want: 0,
		},
		{
			name: "both nil",
			cfg:  &ChainConfig{},
			want: 0,
		},
		{
			name: "all set",
			cfg: &ChainConfig{
				BohoBlock: big.NewInt(0),
				Boho: &AnzeonConfig{
					SystemContracts: &SystemContracts{
						GovMinter: &SystemContract{Address: common.HexToAddress("0x1003"), Version: "v2"},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.CollectUpgrades()
			if len(got) != tt.want {
				t.Errorf("CollectUpgrades() returned %d upgrades, want %d", len(got), tt.want)
			}
		})
	}
}

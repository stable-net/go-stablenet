// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/qbft/engine/engine_test.go (2024.07.25).
// Modified and improved for the wemix development.

package qbftengine

import (
	"bytes"
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func TestPrepareExtra(t *testing.T) {
	validators := make([]common.Address, 4)
	validators[0] = common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a"))
	validators[1] = common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212"))
	validators[2] = common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6"))
	validators[3] = common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440"))

	expectedResult := hexutil.MustDecode("0xf87da00000000000000000000000000000000000000000000000000000000000000000f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b44080c0c080c0c0")

	h := &types.Header{}
	err := ApplyHeaderQBFTExtra(
		h,
		WriteValidators(validators),
	)
	if err != nil {
		t.Errorf("error mismatch: have %v, want: nil", err)
	}
	if !reflect.DeepEqual(h.Extra, expectedResult) {
		t.Errorf("payload mismatch: have %v, want %v", h.Extra, expectedResult)
	}
}

func TestWriteCommittedSeals(t *testing.T) {
	istRawData := hexutil.MustDecode("0xf8a180f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b44080c0c080c0f843b8410102030000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	expectedCommittedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedIstExtra := &types.QBFTExtra{
		VanityData: []byte{},
		Validators: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
			common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
			common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
			common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440")),
		},
		PrevRound:         0,
		PrevCommittedSeal: [][]byte{},
		PrevPreparedSeal:  [][]byte{},
		Round:             0,
		CommittedSeal:     [][]byte{expectedCommittedSeal},
		PreparedSeal:      [][]byte{},
	}

	h := &types.Header{
		Extra: istRawData,
	}

	// normal case
	err := ApplyHeaderQBFTExtra(
		h,
		writeCommittedSeals([][]byte{expectedCommittedSeal}),
	)
	if err != nil {
		t.Errorf("error mismatch: have %v, want: nil", err)
	}

	// verify qbft extra-data
	istExtra, err := getExtra(h)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
	if !reflect.DeepEqual(istExtra, expectedIstExtra) {
		t.Errorf("extra data mismatch: have %v, want %v", istExtra, expectedIstExtra)
	}

	// invalid seal
	unexpectedCommittedSeal := append(expectedCommittedSeal, make([]byte, 1)...)
	err = ApplyHeaderQBFTExtra(
		h,
		writeCommittedSeals([][]byte{unexpectedCommittedSeal}),
	)
	if err != qbftcommon.ErrInvalidCommittedSeals {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidCommittedSeals)
	}
}

func TestWritePreparedSeals(t *testing.T) {
	istRawData := hexutil.MustDecode("0xf8a180f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b44080c0c080f843b8410102030000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0")
	expectedPreparedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedIstExtra := &types.QBFTExtra{
		VanityData: []byte{},
		Validators: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
			common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
			common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
			common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440")),
		},
		PrevRound:         0,
		PrevCommittedSeal: [][]byte{},
		PrevPreparedSeal:  [][]byte{},
		Round:             0,
		CommittedSeal:     [][]byte{},
		PreparedSeal:      [][]byte{expectedPreparedSeal},
	}

	h := &types.Header{
		Extra: istRawData,
	}

	// normal case
	err := ApplyHeaderQBFTExtra(
		h,
		writePreparedSeals([][]byte{expectedPreparedSeal}),
	)
	if err != nil {
		t.Errorf("error mismatch: have %v, want: nil", err)
	}

	// verify qbft extra-data
	istExtra, err := getExtra(h)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
	if !reflect.DeepEqual(istExtra, expectedIstExtra) {
		t.Errorf("extra data mismatch: have %v, want %v", istExtra, expectedIstExtra)
	}

	// invalid seal
	unexpectedPreparedSeal := append(expectedPreparedSeal, make([]byte, 1)...)
	err = ApplyHeaderQBFTExtra(
		h,
		writePreparedSeals([][]byte{unexpectedPreparedSeal}),
	)
	if err != qbftcommon.ErrInvalidPreparedSeals {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidPreparedSeals)
	}
}

func TestWriteRoundNumber(t *testing.T) {
	istRawData := hexutil.MustDecode("0xf85d80f8549444add0ec310f115a0e603b2d7db9f067778eaf8a94294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212946beaaed781d2d2ab6350f5c4566a2c6eaac407a6948be76812f765c24641ec63dc2852b378aba2b44080c0c080c0c0")
	expectedIstExtra := &types.QBFTExtra{
		VanityData: []byte{},
		Validators: []common.Address{
			common.BytesToAddress(hexutil.MustDecode("0x44add0ec310f115a0e603b2d7db9f067778eaf8a")),
			common.BytesToAddress(hexutil.MustDecode("0x294fc7e8f22b3bcdcf955dd7ff3ba2ed833f8212")),
			common.BytesToAddress(hexutil.MustDecode("0x6beaaed781d2d2ab6350f5c4566a2c6eaac407a6")),
			common.BytesToAddress(hexutil.MustDecode("0x8be76812f765c24641ec63dc2852b378aba2b440")),
		},
		PrevRound:         0,
		PrevCommittedSeal: [][]byte{},
		PrevPreparedSeal:  [][]byte{},
		Round:             0,
		CommittedSeal:     [][]byte{},
		PreparedSeal:      [][]byte{},
	}

	var expectedErr error

	h := &types.Header{
		Extra: istRawData,
	}

	// normal case
	err := ApplyHeaderQBFTExtra(
		h,
		writeRoundNumber(big.NewInt(5)),
	)
	if err != expectedErr {
		t.Errorf("error mismatch: have %v, want %v", err, expectedErr)
	}

	// verify qbft extra-data
	istExtra, err := getExtra(h)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
	if istExtra.Round != 5 {
		t.Errorf("writing round does not effected")
	}
	istExtra.Round = expectedIstExtra.Round
	if !reflect.DeepEqual(istExtra, expectedIstExtra) {
		t.Errorf("extra data mismatch: have %v, want %v", istExtra, expectedIstExtra)
	}
}

func TestIsEpochBlock(t *testing.T) {
	engine := NewEngine(nil, common.Address{}, nil)

	testCases := []struct {
		chainConfig         params.ChainConfig
		config              qbft.Config
		blockNumber         *big.Int
		expectedResult      bool
		expectedLatestEpoch *big.Int
		expectedError       error
	}{
		// case 1: no montblanc fork, zero block is an epoch block
		{
			params.ChainConfig{},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int),
			true,
			new(big.Int),
			nil,
		},
		// case 2: no montblanc fork, 1 block is not an epoch block
		{
			params.ChainConfig{},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(1),
			false,
			new(big.Int),
			nil,
		},
		// case 3: no montblanc fork, epoch - 1 block is not an epoch block
		{
			params.ChainConfig{},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(99),
			false,
			new(big.Int),
			nil,
		},
		// case 4: no montblanc fork, epoch block is an epoch block
		{
			params.ChainConfig{},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(100),
			true,
			new(big.Int).SetUint64(100),
			nil,
		},
		// case 5: no montblanc fork, epoch block * n is an epoch block
		{
			params.ChainConfig{},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(300),
			true,
			new(big.Int).SetUint64(300),
			nil,
		},
		// case 6: montblanc fork, error before fork
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1),
			},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int),
			false,
			nil,
			qbftcommon.ErrIsNotWBFTBlock,
		},
		// case 7: montblanc fork, fork block is an epoch block
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(13),
			},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(13),
			true,
			new(big.Int).SetUint64(13),
			nil,
		},
		// case 8: montblanc fork, next of fork block is not an epoch block
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(13),
			},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(14),
			false,
			new(big.Int).SetUint64(13),
			nil,
		},
		// case 9: montblanc fork, fork block + epoch is an epoch block
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(13),
			},
			qbft.Config{
				Epoch: 100,
			},
			new(big.Int).SetUint64(113),
			true,
			new(big.Int).SetUint64(113),
			nil,
		},
		// case 10: montblanc fork, transition exist, before transition
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1101), EpochLength: 200},
				},
			},
			new(big.Int).SetUint64(1100), // before transition
			true,
			new(big.Int).SetUint64(1100),
			nil,
		},
		// case 11: montblanc fork, transition exist, just on transition
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1100), EpochLength: 200},
				},
			},
			new(big.Int).SetUint64(1100), // on transition
			true,
			new(big.Int).SetUint64(1100),
			nil,
		},
		// case 12: montblanc fork, transition exist, after transition, applied new epoch length
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1100), EpochLength: 200},
				},
			},
			new(big.Int).SetUint64(1200),
			false,
			new(big.Int).SetUint64(1100),
			nil,
		},
		// case 13: edge case; transition before montblanc fork?
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(950), EpochLength: 50},
				},
			},
			new(big.Int).SetUint64(1050),
			true,
			new(big.Int).SetUint64(1050),
			nil,
		},
		// case 14: montblanc fork, transitions exist, after transition, applied new epoch length
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1100), EpochLength: 200},
					{Block: new(big.Int).SetUint64(1300), EpochLength: 100},
				},
			},
			new(big.Int).SetUint64(1200),
			false,
			new(big.Int).SetUint64(1100),
			nil,
		},
		// case 15: montblanc fork, transitions exist, after transitions, applied new epoch length
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1100), EpochLength: 200},
					{Block: new(big.Int).SetUint64(1300), EpochLength: 50},
				},
			},
			new(big.Int).SetUint64(1350),
			true,
			new(big.Int).SetUint64(1350),
			nil,
		},
		// case 16: montblanc fork, transitions exist, after transitions, applied new epoch length
		{
			params.ChainConfig{
				MontBlancBlock: new(big.Int).SetUint64(1000),
			},
			qbft.Config{
				Epoch: 100,
				Transitions: []params.Transition{
					{Block: new(big.Int).SetUint64(1100), EpochLength: 10},
					{Block: new(big.Int).SetUint64(1300), EpochLength: 50},
				},
			},
			new(big.Int).SetUint64(1310),
			false,
			new(big.Int).SetUint64(1300),
			nil,
		},
	}

	for i, tc := range testCases {
		testConfig := tc.config
		engine.cfg = &testConfig
		if r, epoch, err := engine.IsEpochBlockNumber(&tc.chainConfig, tc.blockNumber); err != nil {
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("[case %d] unexpected error: have %v, want %v", i+1, err, tc.expectedError)
			}
			if epoch != nil {
				t.Errorf("[case %d] unexpected epoch: have %v, want nil", i+1, epoch)
			}
		} else {
			if r != tc.expectedResult {
				t.Errorf("[case %d] unexpected result: have %v, want %v", i+1, r, tc.expectedResult)
			}
			if epoch.Cmp(tc.expectedLatestEpoch) != 0 {
				t.Errorf("[case %d] unexpected epoch: have %v, want %v", i+1, epoch, tc.expectedLatestEpoch)
			}
		}
	}
}

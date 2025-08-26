// Copyright 2017 The go-ethereum Authors
// Copyright 2024 The go-wemix-wbft Authors
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
// This file is derived from quorum/consensus/istanbul/validator/default_test.go (2024.07.25).
// Modified and improved for the wemix development.

package validator

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
)

var (
	testAddress  = "70524d664ffe731100208a0154e556f9bb679ae6"
	testAddress2 = "b37866a925bccd69cfa98d43b510f1d23d78a851"
)

func TestValidatorSet(t *testing.T) {
	testNewValidatorSet(t)
	testNormalValSet(t)
	testEmptyValSet(t)
	testStickyProposer(t)
	testAddAndRemoveValidator(t)
}

func extractValidators(extraData []byte) []wbft.Validator {
	// get the validator addresses
	validators := make([]wbft.Validator, (len(extraData) / common.AddressLength))
	for i := 0; i < len(validators); i++ {
		var addr common.Address
		copy(addr[:], extraData[i*common.AddressLength:])
		validators = append(validators, New(addr, nil))
	}

	return validators
}

func newTestValidator() wbft.Validator {
	key, _ := crypto.GenerateKey()
	blsKey, _ := bls.DeriveFromECDSA(key)
	return New(crypto.PubkeyToAddress(key.PublicKey), blsKey.PublicKey().Marshal())
}

func testNewValidatorSet(t *testing.T) {
	var validators []wbft.Validator
	const ValCnt = 100

	// Create 100 validators with random addresses
	for i := 0; i < ValCnt; i++ {
		_ = append(validators, newTestValidator())
	}

	// Create ValidatorSet
	valSet := NewSetByValidators(validators, wbft.NewRoundRobinProposerPolicy())
	if valSet == nil {
		t.Errorf("the validator byte array cannot be parsed")
		t.FailNow()
	}
}

func testNormalValSet(t *testing.T) {
	b1 := common.Hex2Bytes(testAddress)
	b2 := common.Hex2Bytes(testAddress2)
	addr1 := common.BytesToAddress(b1)
	addr2 := common.BytesToAddress(b2)
	val1 := New(addr1, nil)
	val2 := New(addr2, nil)

	valSet := newDefaultSet([]wbft.Validator{val1, val2}, wbft.NewRoundRobinProposerPolicy())
	if valSet == nil {
		t.Errorf("the format of validator set is invalid")
		t.FailNow()
	}

	// check size
	if size := valSet.Size(); size != 2 {
		t.Errorf("the size of validator set is wrong: have %v, want 2", size)
	}
	// test get by index
	if val := valSet.GetByIndex(uint64(0)); !reflect.DeepEqual(val, val1) {
		t.Errorf("validator mismatch: have %v, want %v", val, val1)
	}
	// test get by invalid index
	if val := valSet.GetByIndex(uint64(2)); val != nil {
		t.Errorf("validator mismatch: have %v, want nil", val)
	}
	// test get by address
	if _, val := valSet.GetByAddress(addr2); !reflect.DeepEqual(val, val2) {
		t.Errorf("validator mismatch: have %v, want %v", val, val2)
	}
	// test get by invalid address
	invalidAddr := common.HexToAddress("0x9535b2e7faaba5288511d89341d94a38063a349b")
	if _, val := valSet.GetByAddress(invalidAddr); val != nil {
		t.Errorf("validator mismatch: have %v, want nil", val)
	}
	// test get proposer
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val1) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val1)
	}
	// test calculate proposer
	lastProposer := addr1
	valSet.CalcProposer(lastProposer, uint64(0))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val2) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val2)
	}
	valSet.CalcProposer(lastProposer, uint64(3))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val1) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val1)
	}
	// test empty last proposer
	lastProposer = common.Address{}
	valSet.CalcProposer(lastProposer, uint64(3))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val2) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val2)
	}
}

func testEmptyValSet(t *testing.T) {
	valSet := NewSetByValidators(extractValidators([]byte{}), wbft.NewRoundRobinProposerPolicy())
	if valSet == nil {
		t.Errorf("validator set should not be nil")
	}
}

func testAddAndRemoveValidator(t *testing.T) {
	valSet := NewSetByValidators(extractValidators([]byte{}), wbft.NewRoundRobinProposerPolicy())
	if !valSet.AddValidator(common.BytesToAddress([]byte("2")), nil) {
		t.Error("the validator should be added")
	}
	if valSet.AddValidator(common.BytesToAddress([]byte("2")), nil) {
		t.Error("the existing validator should not be added")
	}
	valSet.AddValidator(common.BytesToAddress([]byte("1")), nil)
	valSet.AddValidator(common.BytesToAddress([]byte("0")), nil)
	if len(valSet.List()) != 3 {
		t.Error("the size of validator set should be 3")
	}

	if !valSet.RemoveValidator(common.BytesToAddress([]byte("2"))) {
		t.Error("the validator should be removed")
	}
	if valSet.RemoveValidator(common.BytesToAddress([]byte("2"))) {
		t.Error("the non-existing validator should not be removed")
	}
	if len(valSet.List()) != 2 {
		t.Error("the size of validator set should be 2")
	}
	valSet.RemoveValidator(common.BytesToAddress([]byte("1")))
	if len(valSet.List()) != 1 {
		t.Error("the size of validator set should be 1")
	}
	valSet.RemoveValidator(common.BytesToAddress([]byte("0")))
	if len(valSet.List()) != 0 {
		t.Error("the size of validator set should be 0")
	}
}

func testStickyProposer(t *testing.T) {
	b1 := common.Hex2Bytes(testAddress)
	b2 := common.Hex2Bytes(testAddress2)
	addr1 := common.BytesToAddress(b1)
	addr2 := common.BytesToAddress(b2)
	val1 := New(addr1, nil)
	val2 := New(addr2, nil)

	valSet := newDefaultSet([]wbft.Validator{val1, val2}, wbft.NewStickyProposerPolicy())

	// test get proposer
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val1) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val1)
	}
	// test calculate proposer
	lastProposer := addr1
	valSet.CalcProposer(lastProposer, uint64(0))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val1) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val1)
	}

	valSet.CalcProposer(lastProposer, uint64(1))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val2) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val2)
	}
	// test empty last proposer
	lastProposer = common.Address{}
	valSet.CalcProposer(lastProposer, uint64(3))
	if val := valSet.GetProposer(); !reflect.DeepEqual(val, val2) {
		t.Errorf("proposer mismatch: have %v, want %v", val, val2)
	}
}

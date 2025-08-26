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
// This file is derived from quorum/consensus/istanbul/validator/default.go (2024.07.25).
// Modified and improved for the wemix development.

package validator

import (
	"math"
	"reflect"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
)

type defaultValidator struct {
	address      common.Address
	blsPublicKey []byte
}

func (val *defaultValidator) Address() common.Address {
	return val.address
}

func (val *defaultValidator) BLSPublicKey() []byte {
	return val.blsPublicKey
}

func (val *defaultValidator) String() string {
	return val.Address().String()
}

func (val *defaultValidator) Copy() wbft.Validator {
	newVal := &defaultValidator{
		address:      val.address,
		blsPublicKey: make([]byte, len(val.blsPublicKey)),
	}
	copy(newVal.blsPublicKey, val.blsPublicKey)

	return newVal
}

// ----------------------------------------------------------------------------

type defaultSet struct {
	validators wbft.Validators
	policy     *wbft.ProposerPolicy

	proposer    wbft.Validator
	validatorMu sync.RWMutex
	selector    wbft.ProposalSelector
}

// create an ordered validator set with given addrs order
func newDefaultSet(validators wbft.Validators, policy *wbft.ProposerPolicy) *defaultSet {
	valSet := &defaultSet{}

	valSet.policy = policy

	// init validators
	valSet.validators = validators

	// init proposer
	if valSet.Size() > 0 {
		valSet.proposer = valSet.GetByIndex(0)
	}
	valSet.selector = roundRobinProposer
	if policy.Id == wbft.Sticky {
		valSet.selector = stickyProposer
	}

	return valSet
}

func (valSet *defaultSet) Size() int {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return len(valSet.validators)
}

func (valSet *defaultSet) List() []wbft.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return valSet.validators
}

func (valSet *defaultSet) AddressList() []common.Address {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	result := make([]common.Address, len(valSet.validators))
	for i, v := range valSet.validators {
		result[i] = v.Address()
	}
	return result
}

func (valSet *defaultSet) GetByIndex(i uint64) wbft.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	if i < uint64(valSet.Size()) {
		return valSet.validators[i]
	}
	return nil
}

func (valSet *defaultSet) GetByAddress(addr common.Address) (int, wbft.Validator) {
	for i, val := range valSet.List() {
		if addr == val.Address() {
			return i, val
		}
	}
	return -1, nil
}

func (valSet *defaultSet) GetProposer() wbft.Validator {
	return valSet.proposer
}

func (valSet *defaultSet) IsProposer(address common.Address) bool {
	_, val := valSet.GetByAddress(address)
	return reflect.DeepEqual(valSet.GetProposer(), val)
}

func (valSet *defaultSet) CalcProposer(lastProposer common.Address, round uint64) {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	valSet.proposer = valSet.selector(valSet, lastProposer, round)
}

func calcSeed(valSet wbft.ValidatorSet, proposer common.Address, round uint64) uint64 {
	offset := 0
	if idx, val := valSet.GetByAddress(proposer); val != nil {
		offset = idx
	}
	return uint64(offset) + round
}

func emptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

func roundRobinProposer(valSet wbft.ValidatorSet, proposer common.Address, round uint64) wbft.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round) + 1
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

func stickyProposer(valSet wbft.ValidatorSet, proposer common.Address, round uint64) wbft.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round)
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

func (valSet *defaultSet) AddValidator(address common.Address, blsPublicKey []byte) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()
	for _, v := range valSet.validators {
		if v.Address() == address {
			return false
		}
	}
	valSet.validators = append(valSet.validators, New(address, blsPublicKey))
	return true
}

func (valSet *defaultSet) RemoveValidator(address common.Address) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()

	for i, v := range valSet.validators {
		if v.Address() == address {
			valSet.validators = append(valSet.validators[:i], valSet.validators[i+1:]...)
			return true
		}
	}
	return false
}

func (valSet *defaultSet) Copy() wbft.ValidatorSet {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	validators := make([]wbft.Validator, 0, len(valSet.validators))
	for _, v := range valSet.validators {
		validators = append(validators, v.Copy())
	}
	return NewSetByValidators(validators, valSet.policy)
}

func (valSet *defaultSet) F() float64 { return float64(valSet.Size()) / 3 }

func (valSet *defaultSet) Policy() wbft.ProposerPolicy { return *valSet.policy }

func (valSet *defaultSet) QuorumSize() int {
	//c.currentLogger(true, nil).Trace("WBFT: confirmation Formula used ceil(2N/3)")
	return int(math.Ceil(float64(valSet.Size()) - valSet.F()))
}

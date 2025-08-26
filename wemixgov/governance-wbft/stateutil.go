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
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

type StateReader interface {
	GetState(addr common.Address, hash common.Hash) common.Hash
}

func CalculateMappingSlot(baseSlot common.Hash, key interface{ Bytes() []byte }) common.Hash {
	// keccak256(encode(key) . encode(slot))
	hash := sha3.NewLegacyKeccak256()

	keyBytes := append(common.LeftPadBytes(key.Bytes(), 32), baseSlot.Bytes()...)
	hash.Write(keyBytes)
	return common.BytesToHash(hash.Sum(nil))
}

func CalculateDynamicSlot(baseSlot interface{ Bytes() []byte }, index *big.Int) common.Hash {
	// keccak256(baseSlot)으로 배열의 시작 위치를 계산
	hash := sha3.NewLegacyKeccak256()
	hash.Write(common.LeftPadBytes(baseSlot.Bytes(), 32))
	arrayStartSlot := new(big.Int).SetBytes(hash.Sum(nil))

	// arrayStartSlot + index
	elementSlot := new(big.Int).Add(arrayStartSlot, index)

	return common.BigToHash(elementSlot)
}

func IncrementHash(baseSlot common.Hash, increment *big.Int) common.Hash {
	return common.BigToHash(new(big.Int).Add(baseSlot.Big(), increment))
}

type EnumerableSet[T interface{ Bytes() []byte }] struct {
	indexSlot common.Hash
	valueSlot common.Hash
	convertFn func(common.Hash) T
}

func NewEnumerableSet[T interface{ Bytes() []byte }](baseSlot common.Hash) *EnumerableSet[T] {
	return &EnumerableSet[T]{
		valueSlot: baseSlot,
		indexSlot: IncrementHash(baseSlot, big.NewInt(1)),
	}
}

func (es *EnumerableSet[T]) Length(state StateReader, address common.Address) uint64 {
	return state.GetState(address, es.valueSlot).Big().Uint64()
}

func (es *EnumerableSet[T]) Contains(state StateReader, address common.Address, value T) bool {
	index := state.GetState(address, CalculateMappingSlot(es.indexSlot, value)).Big()

	return index.Sign() > 0
}

func (es *EnumerableSet[T]) Values(state StateReader, address common.Address) []T {
	len := es.Length(state, address)
	values := make([]T, len)
	for i := uint64(0); i < len; i++ {
		values[i] = es.convertFn(state.GetState(address, CalculateDynamicSlot(es.valueSlot, new(big.Int).SetUint64(i))))
	}
	return values
}

func (es *EnumerableSet[T]) At(state StateReader, address common.Address, index *big.Int) T {
	return es.convertFn(state.GetState(address, CalculateDynamicSlot(es.valueSlot, index)))
}

func NewAddressSet(baseSlot common.Hash) *EnumerableSet[common.Address] {
	es := NewEnumerableSet[common.Address](baseSlot)
	es.convertFn = HashToAddress
	return es
}

func HashToAddress(hash common.Hash) common.Address {
	return common.BytesToAddress(hash.Bytes())
}

// If retrieving a string, use string(GetBytes(...))
func GetBytes(stateDB StateReader, contractAddress common.Address, baseSlot common.Hash) []byte {
	// Retrieve the data from the baseSlot
	storageValue := stateDB.GetState(contractAddress, baseSlot)

	// If the slot is empty, return an empty byte array
	if storageValue.Cmp(common.Hash{}) == 0 {
		return []byte{}
	}

	var bytesLength int
	// Check the last bit to determine if the data length exceeds 31 bytes
	if (storageValue[31] & 0x01) == 0 {
		// If 0 - the length is 31 bytes or less
		// Extract the data length from the last byte
		bytesLength = int(storageValue[31] >> 1)

		// return the data stored directly in the base slot
		return storageValue[:bytesLength]
	}
	// If 1 - the length exceeds 31 bytes, the slot stores only the length
	hashInt := storageValue.Big()

	// Ignore the least significant bit and extract the actual length
	bytesLength = int(hashInt.Rsh(hashInt, 1).Int64())

	// Prepare a byte slice to store the data
	bytesData := make([]byte, 0)

	// For data longer than 31 bytes, traverse the slots using Keccak-256
	for i := int64(0); ; i++ {
		// Calculate the current slot
		currentSlot := CalculateDynamicSlot(baseSlot, big.NewInt(i))
		// Retrieve the data from the current slot
		slotData := stateDB.GetState(contractAddress, currentSlot)

		// Append the data from the slot to the byte slice
		bytesData = append(bytesData, slotData[:]...)

		// Stop the loop when the entire data has been collected (data is split into 32-byte chunks)
		if len(bytesData) >= bytesLength {
			break
		}
	}

	// Return the exact length of the data
	return bytesData[:bytesLength]
}

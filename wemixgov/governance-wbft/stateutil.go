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

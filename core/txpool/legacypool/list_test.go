// Copyright 2016 The go-ethereum Authors
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

package legacypool

import (
	"container/heap"
	"math/big"
	"math/rand"
	"testing"

	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Tests that transactions can be added to strict lists and list contents and
// nonce boundaries are correctly maintained.
func TestStrictListAdd(t *testing.T) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make(types.Transactions, 1024)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), 0, key)
	}
	// Insert the transactions in a random order
	list := newList(true)
	for _, v := range rand.Perm(len(txs)) {
		list.Add(txs[v], DefaultConfig.PriceBump)
	}
	// Verify internal state
	if len(list.txs.items) != len(txs) {
		t.Errorf("transaction count mismatch: have %d, want %d", len(list.txs.items), len(txs))
	}
	for i, tx := range txs {
		if list.txs.items[tx.Nonce()] != tx {
			t.Errorf("item %d: transaction mismatch: have %v, want %v", i, list.txs.items[tx.Nonce()], tx)
		}
	}
}

// TestListAddVeryExpensive tests adding txs which exceed 256 bits in cost. It is
// expected that the list does not panic.
func TestListAddVeryExpensive(t *testing.T) {
	key, _ := crypto.GenerateKey()
	list := newList(true)
	for i := 0; i < 3; i++ {
		value := big.NewInt(100)
		gasprice, _ := new(big.Int).SetString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 0)
		gaslimit := uint64(i)
		tx, _ := types.SignTx(types.NewTransaction(uint64(i), common.Address{}, value, gaslimit, gasprice, nil), types.HomesteadSigner{}, key)
		t.Logf("cost: %x bitlen: %d\n", tx.Cost(), tx.Cost().BitLen())
		list.Add(tx, DefaultConfig.PriceBump)
	}
}

func BenchmarkListAdd(b *testing.B) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make(types.Transactions, 100000)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), 0, key)
	}
	// Insert the transactions in a random order
	priceLimit := uint256.NewInt(DefaultConfig.PriceLimit)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := newList(true)
		for _, v := range rand.Perm(len(txs)) {
			list.Add(txs[v], DefaultConfig.PriceBump)
			list.Filter(false, nil, priceLimit, DefaultConfig.PriceBump)
		}
	}
}

func BenchmarkListCapOneTx(b *testing.B) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make(types.Transactions, 32)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), 0, key)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := newList(true)
		// Insert the transactions in a random order
		for _, v := range rand.Perm(len(txs)) {
			list.Add(txs[v], DefaultConfig.PriceBump)
		}
		b.StartTimer()
		list.Cap(list.Len() - 1)
		b.StopTimer()
	}
}

// TestPricedListIntegration tests the complete pricedList functionality including
// Put, PutAnzeon, Pop, Underpriced, SetBaseFee, SetHeaderGasTip, and all cmp/Less conditions
func TestPricedListIntegration(t *testing.T) {
	lookup := newLookup()
	priced := newPricedList(lookup)

	// Generate keys for different accounts
	authorizedKey1, _ := crypto.GenerateKey()
	authorizedKey2, _ := crypto.GenerateKey()
	normalKey1, _ := crypto.GenerateKey()
	normalKey2, _ := crypto.GenerateKey()
	normalKey3, _ := crypto.GenerateKey()

	// ============================================
	// Test Case 1: Anzeon Disabled, baseFee = nil, headerGasTip = nil
	// ============================================
	t.Logf("=== Test Case 1: Anzeon Disabled, baseFee = nil, headerGasTip = nil ===")

	// Create transactions with different feeCaps (no baseFee, no headerGasTip, so compare by feeCap)
	tx1 := dynamicFeeTx(0, 21000, big.NewInt(50000), big.NewInt(20000), authorizedKey1) // feeCap=50000
	tx2 := dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(15000), authorizedKey2) // feeCap=40000
	tx3 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(10000), normalKey1)     // feeCap=30000

	lookup.Add(tx1, false)
	lookup.Add(tx2, false)
	lookup.Add(tx3, false)

	priced.Put(tx1, false)
	priced.Put(tx2, false)
	priced.Put(tx3, false)

	// Reheap to initialize heap
	priced.Reheap()

	// Verify order: tx1 (50000) > tx2 (40000) > tx3 (30000)
	popped1_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped1_1.Hash() != tx3.Hash() {
		t.Errorf("Expected tx3 (feeCap=30000) to be popped first, got tx with feeCap=%s", popped1_1.GasFeeCap().String())
	}

	popped1_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped1_2.Hash() != tx2.Hash() {
		t.Errorf("Expected tx2 (feeCap=40000) to be popped second, got tx with feeCap=%s", popped1_2.GasFeeCap().String())
	}

	popped1_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped1_3.Hash() != tx1.Hash() {
		t.Errorf("Expected tx1 (feeCap=50000) to be popped last, got tx with feeCap=%s", popped1_3.GasFeeCap().String())
	}

	// ============================================
	// Test Case 2: Anzeon Disabled, baseFee exists, headerGasTip = nil
	// ============================================
	t.Logf("=== Test Case 2: Anzeon Disabled, baseFee = 10000, headerGasTip = nil ===")

	// Clear and reset
	lookup = newLookup()
	priced = newPricedList(lookup)

	// Set baseFee
	baseFee := big.NewInt(10000)
	priced.SetBaseFee(baseFee)

	// effectiveTip = min(tipCap, feeCap - baseFee)

	// Create transactions with same feeCap but different effectiveTips
	// tx4: feeCap=30000, tipCap=20000 -> effectiveTip=min(20000, 20000)=20000
	tx4 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)
	// tx5: feeCap=30000, tipCap=15000 -> effectiveTip=min(15000, 20000)=15000
	tx5 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(15000), authorizedKey2)
	// tx6: feeCap=30000, tipCap=10000 -> effectiveTip=min(10000, 20000)=10000
	tx6 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(10000), normalKey1)

	lookup.Add(tx4, false)
	lookup.Add(tx5, false)
	lookup.Add(tx6, false)

	priced.Put(tx4, false)
	priced.Put(tx5, false)
	priced.Put(tx6, false)

	priced.Reheap()

	// Verify order: tx4 (effectiveTip=20000) > tx5 (15000) > tx6 (10000)
	popped2_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped2_1.Hash() != tx6.Hash() {
		t.Errorf("Expected tx6 (effectiveTip=10000) to be popped first")
	}

	popped2_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped2_2.Hash() != tx5.Hash() {
		t.Errorf("Expected tx5 (effectiveTip=15000) to be popped second")
	}

	popped2_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped2_3.Hash() != tx4.Hash() {
		t.Errorf("Expected tx4 (effectiveTip=20000) to be popped last")
	}

	// ============================================
	// Test Case 3: Anzeon Enabled, baseFee exists, headerGasTip exists
	// ============================================
	t.Logf("=== Test Case 3: Anzeon Enabled, baseFee = 10000, headerGasTip = 10000 ===")

	// Clear and reset
	lookup = newLookup()
	priced = newPricedList(lookup)

	baseFee = big.NewInt(10000)
	headerGasTip := big.NewInt(10000)
	priced.SetBaseFee(baseFee)
	priced.SetHeaderGasTip(headerGasTip)

	//effectiveTip = min(tipCap, feeCap - baseFee)

	// Authorized transactions use their own GasTipCap
	// authTx1: feeCap=30000, tipCap=20000 -> effectiveTip=min(20000, 20000)=20000
	authTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)
	// authTx2: feeCap=30000, tipCap=15000 -> effectiveTip=min(15000, 20000)=15000
	authTx2 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(15000), authorizedKey2)

	// Normal transactions use headerGasTip (10000)
	// normTx1: feeCap=30000, tipCap=25000 (ignored) -> effectiveTip=min(10000, 20000)=10000
	normTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(25000), normalKey1)
	// normTx2: feeCap=25000, tipCap=20000 (ignored) -> effectiveTip=min(10000, 15000)=10000
	normTx2 := dynamicFeeTx(0, 21000, big.NewInt(25000), big.NewInt(20000), normalKey2)

	lookup.Add(authTx1, false)
	lookup.Add(authTx2, false)
	lookup.Add(normTx1, false)
	lookup.Add(normTx2, false)

	priced.PutAnzeon(authTx1, false, true)
	priced.PutAnzeon(authTx2, false, true)
	priced.PutAnzeon(normTx1, false, false)
	priced.PutAnzeon(normTx2, false, false)

	priced.Reheap()

	// Verify order: authTx1 (20000) > authTx2 (15000) > normTx1 (10000, feeCap=30000) > normTx2 (10000, feeCap=25000)
	popped3_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3_1.Hash() != normTx2.Hash() {
		t.Errorf("Expected normTx2 (effectiveTip=10000, feeCap=25000) to be popped first")
	}

	popped3_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3_2.Hash() != normTx1.Hash() {
		t.Errorf("Expected normTx1 (effectiveTip=10000, feeCap=30000) to be popped second")
	}

	popped3_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3_3.Hash() != authTx2.Hash() {
		t.Errorf("Expected authTx2 (effectiveTip=15000) to be popped third")
	}

	popped3_4 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3_4.Hash() != authTx1.Hash() {
		t.Errorf("Expected authTx1 (effectiveTip=20000) to be popped last")
	}

	// ============================================
	// Test Case 4: Same effectiveTip, compare by feeCap
	// ============================================
	t.Logf("=== Test Case 4: Same effectiveTip, compare by feeCap ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	priced.SetHeaderGasTip(big.NewInt(10000))

	// effectiveTip = min(tipCap, feeCap - baseFee)

	// Normal transactions use headerGasTip (10000)
	// All have effectiveTip=10000, but different feeCaps
	// normTx3: feeCap=40000, tipCap=30000(ignored) -> effectiveTip=min(10000, 40000-10000)=10000
	normTx3 := dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(30000), normalKey3)
	// normTx4: feeCap=35000, tipCap=30000(ignored) -> effectiveTip=min(10000, 35000-10000)=10000
	normTx4 := dynamicFeeTx(0, 21000, big.NewInt(35000), big.NewInt(30000), normalKey1)
	// normTx5: feeCap=30000, tipCap=30000(ignored) -> effectiveTip=min(10000, 30000-10000)=10000
	normTx5 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(30000), normalKey2)

	lookup.Add(normTx3, false)
	lookup.Add(normTx4, false)
	lookup.Add(normTx5, false)

	priced.PutAnzeon(normTx3, false, false)
	priced.PutAnzeon(normTx4, false, false)
	priced.PutAnzeon(normTx5, false, false)

	priced.Reheap()

	// Verify order: normTx5 (30000) < normTx4 (35000) < normTx3 (40000)
	popped4_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped4_1.Hash() != normTx5.Hash() {
		t.Errorf("Expected normTx5 (feeCap=30000) to be popped first")
	}

	popped4_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped4_2.Hash() != normTx4.Hash() {
		t.Errorf("Expected normTx4 (feeCap=35000) to be popped second")
	}

	popped4_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped4_3.Hash() != normTx3.Hash() {
		t.Errorf("Expected normTx3 (feeCap=40000) to be popped last")
	}

	// ============================================
	// Test Case 5: Same effectiveTip and feeCap, compare by tipCap
	// ============================================
	t.Logf("=== Test Case 5: Same effectiveTip and feeCap, compare by tipCap ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	priced.SetHeaderGasTip(big.NewInt(10000))

	// effectiveTip = min(tipCap, feeCap - baseFee)
	// All have same feeCap=30000, but different tipCaps (for authorized accounts)
	// authTx3: tipCap=20000 -> effectiveTip=min(20000, 30000-10000)=min(20000, 20000)=20000
	authTx3 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)
	// authTx4: tipCap=15000 -> effectiveTip=min(15000, 30000-10000)=min(15000, 20000)=15000
	authTx4 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(15000), authorizedKey2)

	lookup.Add(authTx3, false)
	lookup.Add(authTx4, false)

	priced.PutAnzeon(authTx3, false, true)
	priced.PutAnzeon(authTx4, false, true)

	priced.Reheap()

	// Verify order: authTx4 (effectiveTip=15000) < authTx3 (effectiveTip=20000)
	popped5_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped5_1.Hash() != authTx4.Hash() {
		t.Errorf("Expected authTx4 (effectiveTip=15000) to be popped first")
	}

	popped5_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped5_2.Hash() != authTx3.Hash() {
		t.Errorf("Expected authTx3 (effectiveTip=20000) to be popped last")
	}

	// ============================================
	// Test Case 6a: Same effectiveTip, feeCap, tipCap - compare by nonce (reverse)
	// ============================================
	t.Logf("=== Test Case 6: Same effectiveTip, feeCap, tipCap - compare by nonce (reverse) ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	priced.SetHeaderGasTip(big.NewInt(10000))

	// All values identical, different nonces
	// sameTx1: nonce=0
	sameTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)
	// sameTx2: nonce=1
	sameTx2 := dynamicFeeTx(1, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)
	// sameTx3: nonce=2
	sameTx3 := dynamicFeeTx(2, 21000, big.NewInt(30000), big.NewInt(20000), authorizedKey1)

	lookup.Add(sameTx1, false)
	lookup.Add(sameTx2, false)
	lookup.Add(sameTx3, false)

	priced.PutAnzeon(sameTx1, false, true)
	priced.PutAnzeon(sameTx2, false, true)
	priced.PutAnzeon(sameTx3, false, true)

	priced.Reheap()

	// Verify order: nonce reverse (higher nonce first)
	// sameTx3 (nonce=2) < sameTx2 (nonce=1) < sameTx1 (nonce=0)
	popped6a_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6a_1.Hash() != sameTx3.Hash() {
		t.Errorf("Expected sameTx3 (nonce=2) to be popped first")
	}

	popped6a_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6a_2.Hash() != sameTx2.Hash() {
		t.Errorf("Expected sameTx2 (nonce=1) to be popped second")
	}

	popped6a_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6a_3.Hash() != sameTx1.Hash() {
		t.Errorf("Expected sameTx1 (nonce=0) to be popped last")
	}

	// ============================================
	// Test Case 6b: nonce comparison (Anzeon disabled, no baseFee)
	// ============================================
	t.Logf("=== Test Case 6b: nonce comparison (Anzeon disabled, no baseFee) ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	// No baseFee, no headerGasTip (Anzeon disabled)
	// All transactions have same feeCap and tipCap
	// nonceTx1: nonce=0, feeCap=30000, tipCap=20000
	nonceTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)
	// nonceTx2: nonce=1, feeCap=30000, tipCap=20000
	nonceTx2 := dynamicFeeTx(1, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)
	// nonceTx3: nonce=2, feeCap=30000, tipCap=20000
	nonceTx3 := dynamicFeeTx(2, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)

	lookup.Add(nonceTx1, false)
	lookup.Add(nonceTx2, false)
	lookup.Add(nonceTx3, false)

	priced.Put(nonceTx1, false)
	priced.Put(nonceTx2, false)
	priced.Put(nonceTx3, false)

	priced.Reheap()

	// Verify order: nonce reverse (higher nonce first)
	// nonceTx3 (nonce=2) < nonceTx2 (nonce=1) < nonceTx1 (nonce=0)
	popped6b_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6b_1.Hash() != nonceTx3.Hash() {
		t.Errorf("Expected nonceTx3 (nonce=2) to be popped first")
	}

	popped6b_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6b_2.Hash() != nonceTx2.Hash() {
		t.Errorf("Expected nonceTx2 (nonce=1) to be popped second")
	}

	popped6b_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6b_3.Hash() != nonceTx1.Hash() {
		t.Errorf("Expected nonceTx1 (nonce=0) to be popped last")
	}

	// ============================================
	// Test Case 6c: nonce comparison (Anzeon disabled, with baseFee)
	// ============================================
	t.Logf("=== Test Case 6c: nonce comparison (Anzeon disabled, with baseFee) ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	// No headerGasTip (Anzeon disabled)

	// All transactions have same feeCap, tipCap, and effectiveTip
	// baseNonceTx1: nonce=0, feeCap=30000, tipCap=20000 -> effectiveTip=min(20000, 20000)=20000
	baseNonceTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)
	// baseNonceTx2: nonce=1, feeCap=30000, tipCap=20000 -> effectiveTip=min(20000, 20000)=20000
	baseNonceTx2 := dynamicFeeTx(1, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)
	// baseNonceTx3: nonce=2, feeCap=30000, tipCap=20000 -> effectiveTip=min(20000, 20000)=20000
	baseNonceTx3 := dynamicFeeTx(2, 21000, big.NewInt(30000), big.NewInt(20000), normalKey1)

	lookup.Add(baseNonceTx1, false)
	lookup.Add(baseNonceTx2, false)
	lookup.Add(baseNonceTx3, false)

	priced.Put(baseNonceTx1, false)
	priced.Put(baseNonceTx2, false)
	priced.Put(baseNonceTx3, false)

	priced.Reheap()

	// Verify order: nonce reverse (higher nonce first)
	// baseNonceTx3 (nonce=2) < baseNonceTx2 (nonce=1) < baseNonceTx1 (nonce=0)
	popped6c_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6c_1.Hash() != baseNonceTx3.Hash() {
		t.Errorf("Expected baseNonceTx3 (nonce=2) to be popped first")
	}

	popped6c_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6c_2.Hash() != baseNonceTx2.Hash() {
		t.Errorf("Expected baseNonceTx2 (nonce=1) to be popped second")
	}

	popped6c_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6c_3.Hash() != baseNonceTx1.Hash() {
		t.Errorf("Expected baseNonceTx1 (nonce=0) to be popped last")
	}

	// ============================================
	// Test Case 6d: nonce comparison (Anzeon enabled, mixed accounts)
	// ============================================
	t.Logf("=== Test Case 6d: All nonce comparison (Anzeon enabled, mixed accounts) ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	priced.SetHeaderGasTip(big.NewInt(20000)) // Same as tipCap for authorized

	// All transactions have same effectiveTip, feeCap
	// mixedNonceTx1: nonce=0, authorized, feeCap=40000, tipCap=20000 -> effectiveTip=min(20000, 40000-10000)=20000
	mixedNonceTx1 := dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(20000), authorizedKey1)
	// mixedNonceTx2: nonce=1, normal, feeCap=40000, tipCap=30000(ignored) -> effectiveTip=min(20000, 40000-10000)=20000
	mixedNonceTx2 := dynamicFeeTx(1, 21000, big.NewInt(40000), big.NewInt(30000), normalKey1)
	// mixedNonceTx3: nonce=2, authorized, feeCap=40000, tipCap=20000 -> effectiveTip=min(20000, 40000-10000)=20000
	mixedNonceTx3 := dynamicFeeTx(2, 21000, big.NewInt(40000), big.NewInt(20000), authorizedKey2)

	lookup.Add(mixedNonceTx1, false)
	lookup.Add(mixedNonceTx2, false)
	lookup.Add(mixedNonceTx3, false)

	priced.PutAnzeon(mixedNonceTx1, false, true)
	priced.PutAnzeon(mixedNonceTx2, false, false)
	priced.PutAnzeon(mixedNonceTx3, false, true)

	priced.Reheap()

	// All have same effectiveTip (20000), feeCap (40000), tipCap (20000), normal account uses headerGasTip=20000)
	// Should compare by nonce reverse: nonce=2 < nonce=1 < nonce=0
	popped6d_1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6d_1.Hash() != mixedNonceTx3.Hash() {
		t.Errorf("Expected mixedNonceTx3 (nonce=2) to be popped first")
	}

	popped6d_2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6d_2.Hash() != mixedNonceTx2.Hash() {
		t.Errorf("Expected mixedNonceTx2 (nonce=1) to be popped second")
	}

	popped6d_3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped6d_3.Hash() != mixedNonceTx1.Hash() {
		t.Errorf("Expected mixedNonceTx1 (nonce=0) to be popped last")
	}

	// ============================================
	// Test Case 7: Underpriced check
	// ============================================
	t.Logf("=== Test Case 7: Underpriced check ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	priced.SetBaseFee(big.NewInt(10000))
	priced.SetHeaderGasTip(big.NewInt(10000))

	// Add multiple transactions to ensure both urgent and floating heaps have items
	// highTx: effectiveTip=min(30000, 40000)=30000
	highTx := dynamicFeeTx(0, 21000, big.NewInt(50000), big.NewInt(30000), authorizedKey1)
	// midTx: effectiveTip=min(20000, 40000)=20000
	midTx := dynamicFeeTx(0, 21000, big.NewInt(45000), big.NewInt(20000), authorizedKey2)
	// lowTx: effectiveTip=min(10000, 30000)=10000
	veryLowTx := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(10000), normalKey1)

	lookup.Add(highTx, false)
	lookup.Add(midTx, false)
	lookup.Add(veryLowTx, false)

	priced.PutAnzeon(highTx, false, true)
	priced.PutAnzeon(midTx, false, true)
	priced.PutAnzeon(veryLowTx, false, false)

	priced.Reheap()

	// After Reheap, some transactions may be in floating heap
	// The minimum priority transaction (worst) should be in one of the heaps

	// Create a lower-priority transaction than the worst one
	// lowTx: effectiveTip=min(5000, 10000)=5000
	lowTx := dynamicFeeTx(0, 21000, big.NewInt(20000), big.NewInt(5000), normalKey2)

	// lowTx should be underpriced (worse than the worst transaction in heaps)
	if !priced.Underpriced(lowTx) {
		t.Errorf("Expected lowTx to be underpriced")
	}

	// Create a much higher-priority transaction
	// higherTx: effectiveTip=min(40000, 50000)=40000, feeCap=60000
	higherTx := dynamicFeeTx(0, 21000, big.NewInt(60000), big.NewInt(40000), authorizedKey1)

	// higherTx should NOT be underpriced (better than all transactions in heaps)
	if priced.Underpriced(higherTx) {
		t.Errorf("Expected higherTx NOT to be underpriced (effectiveTip=40000 > all others)")
	}

	// ============================================
	// Test Case 8: Change headerGasTip
	// ============================================
	t.Logf("=== Test Case 8: Change headerGasTip ===")

	lookup = newLookup()
	priced = newPricedList(lookup)

	// Initial settings: baseFee=10000, headerGasTip=10000
	baseFee = big.NewInt(10000)
	headerGasTip = big.NewInt(10000)
	priced.SetBaseFee(baseFee)
	priced.SetHeaderGasTip(headerGasTip)

	// Add multiple transactions to see sorting effects
	// Conditions:
	// - GasTipCap >= headerGasTip (all accounts)
	// - GasFeeCap >= headerGasTip + baseFee (normal accounts)
	// - GasFeeCap >= GasTipCap + baseFee (authorized accounts)

	// Authorized transactions
	// dynAuthTx1: effectiveTip=min(20000, 40000-10000)=min(20000, 30000)=20000
	dynAuthTx1 := dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(20000), authorizedKey1)
	// dynAuthTx2: effectiveTip=min(15000, 30000-10000)=min(15000, 20000)=15000
	dynAuthTx2 := dynamicFeeTx(1, 21000, big.NewInt(30000), big.NewInt(15000), authorizedKey2)

	// Normal transactions (use headerGasTip=10000)
	// dynNormTx1: effectiveTip=min(10000, 30000-10000)=min(10000, 20000)=10000
	dynNormTx1 := dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(25000), normalKey1)
	// dynNormTx2: effectiveTip=min(10000, 35000-10000)=min(10000, 25000)=10000
	dynNormTx2 := dynamicFeeTx(0, 21000, big.NewInt(35000), big.NewInt(30000), normalKey2)
	// dynNormTx3: effectiveTip=min(10000, 25000-10000)=min(10000, 15000)=10000
	dynNormTx3 := dynamicFeeTx(0, 21000, big.NewInt(25000), big.NewInt(20000), normalKey3)

	lookup.Add(dynAuthTx1, false)
	lookup.Add(dynAuthTx2, false)
	lookup.Add(dynNormTx1, false)
	lookup.Add(dynNormTx2, false)
	lookup.Add(dynNormTx3, false)

	priced.PutAnzeon(dynAuthTx1, false, true)
	priced.PutAnzeon(dynAuthTx2, false, true)
	priced.PutAnzeon(dynNormTx1, false, false)
	priced.PutAnzeon(dynNormTx2, false, false)
	priced.PutAnzeon(dynNormTx3, false, false)

	priced.Reheap()

	// Note: Reheap() moves some transactions to floating heap
	// With 5 transactions: floatingCount = 5 * 1 / (4 + 1) = 1
	// So urgent has 4, floating has 1

	// Initial order (headerGasTip=10000, min-heap: smallest first):
	// All normal transactions have effectiveTip=10000, compare by feeCap
	// 1. dynNormTx3 (effectiveTip=10000, feeCap=35000) - smallest (lowest priority, moved to floating)
	// 2. dynNormTx1 (effectiveTip=10000, feeCap=40000)
	// 3. dynNormTx2 (effectiveTip=10000, feeCap=45000)
	// 4. dynAuthTx2 (effectiveTip=15000)
	// 5. dynAuthTx1 (effectiveTip=20000) - largest (highest priority)

	// Verify initial order in urgent heap (4 transactions remain)
	// Note: Reheap() balances heaps, so some transactions move to floating
	urgentCount := priced.urgent.Len()
	floatingCount := priced.floating.Len()
	if urgentCount != 4 {
		t.Errorf("Initial: expected 4 transactions in urgent, got %d (floating has %d)", urgentCount, floatingCount)
	}

	// Verify floating heap has the smallest transaction (moved by Reheap)
	if priced.floating.Len() != 1 {
		t.Errorf("Initial: expected 1 transaction in floating, got %d", priced.floating.Len())
	}

	// Pop from urgent and verify sorting order
	// After Reheap(), urgent has 4 transactions, floating has 1
	// The smallest transaction (dynNormTx3) is moved to floating
	// Expected order in urgent heap (after smallest moved to floating):
	// 1. dynNormTx1 (effectiveTip=10000, feeCap=40000) - smallest remaining
	// 2. dynNormTx2 (effectiveTip=10000, feeCap=45000)
	// 3. dynAuthTx2 (effectiveTip=15000)
	// 4. dynAuthTx1 (effectiveTip=20000)
	popped1 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped1.Hash() != dynNormTx1.Hash() {
		t.Errorf("Initial: expected dynNormTx1 (effectiveTip=10000, feeCap=40000) to be popped first, got feeCap=%s", popped1.GasFeeCap().String())
	}

	popped2 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped2.Hash() != dynNormTx2.Hash() {
		t.Errorf("Initial: expected dynNormTx2 (effectiveTip=10000, feeCap=45000) to be popped second, got feeCap=%s", popped2.GasFeeCap().String())
	}

	popped3 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3.Hash() != dynAuthTx2.Hash() {
		t.Errorf("Initial: expected dynAuthTx2 (effectiveTip=15000) to be popped third")
	}

	popped4 := heap.Pop(&priced.urgent).(*types.Transaction)
	if popped4.Hash() != dynAuthTx1.Hash() {
		t.Errorf("Initial: expected dynAuthTx1 (effectiveTip=20000) to be popped fourth")
	}

	floatingTx := heap.Pop(&priced.floating).(*types.Transaction)
	if floatingTx.Hash() != dynNormTx3.Hash() {
		t.Errorf("Initial: expected dynNormTx3 (effectiveTip=10000, feeCap=35000) to be in floating heap")
	}

	t.Logf("Initial sorting verified: urgent has 4, floating has 1, all in correct order")

	// Re-add transactions for headerGasTip change test
	// Need to create new transactions that satisfy conditions for headerGasTip=25000
	lookup = newLookup()
	priced = newPricedList(lookup)

	baseFee = big.NewInt(10000)
	headerGasTip = big.NewInt(10000)
	priced.SetBaseFee(baseFee)
	priced.SetHeaderGasTip(headerGasTip)

	// Create transactions that satisfy conditions for both headerGasTip=10000 and 25000
	// Authorized transactions (unchanged)
	// dynAuthTx1: tipCap=20000 >= headerGasTip(10000, 25000) ✓, feeCap=40000 >= tipCap(20000) + baseFee(10000) = 30000 ✓
	dynAuthTx1 = dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(20000), authorizedKey1)
	// dynAuthTx2: tipCap=15000 >= headerGasTip(10000) ✓, feeCap=30000 >= tipCap(15000) + baseFee(10000) = 25000 ✓
	dynAuthTx2 = dynamicFeeTx(0, 21000, big.NewInt(30000), big.NewInt(15000), authorizedKey2)

	// dynNormTx1: headerGasTip=25000: effectiveTip=min(25000, 40000-10000)=min(25000, 30000)=25000
	dynNormTx1 = dynamicFeeTx(0, 21000, big.NewInt(40000), big.NewInt(25000), normalKey1)
	// dynNormTx2: headerGasTip=25000: effectiveTip=min(25000, 45000-10000)=min(25000, 35000)=25000
	dynNormTx2 = dynamicFeeTx(0, 21000, big.NewInt(45000), big.NewInt(30000), normalKey2)
	// dynNormTx3: headerGasTip=25000: effectiveTip=min(25000, 35000-10000)=min(25000, 25000)=25000
	dynNormTx3 = dynamicFeeTx(0, 21000, big.NewInt(35000), big.NewInt(20000), normalKey3)

	lookup.Add(dynAuthTx1, false)
	lookup.Add(dynAuthTx2, false)
	lookup.Add(dynNormTx1, false)
	lookup.Add(dynNormTx2, false)
	lookup.Add(dynNormTx3, false)

	priced.PutAnzeon(dynAuthTx1, false, true)
	priced.PutAnzeon(dynAuthTx2, false, true)
	priced.PutAnzeon(dynNormTx1, false, false)
	priced.PutAnzeon(dynNormTx2, false, false)
	priced.PutAnzeon(dynNormTx3, false, false)

	priced.Reheap()

	// Initial order (headerGasTip=10000, min-heap: smallest first):
	// All normal transactions have effectiveTip=10000, compare by feeCap
	// 1. dynNormTx3 (effectiveTip=10000, feeCap=35000) - smallest (lowest priority)
	// 2. dynNormTx1 (effectiveTip=10000, feeCap=40000)
	// 3. dynNormTx2 (effectiveTip=10000, feeCap=45000)
	// 4. dynAuthTx2 (effectiveTip=15000)
	// 5. dynAuthTx1 (effectiveTip=20000) - largest (highest priority)

	// Change headerGasTip from 10000 to 15000
	priced.SetHeaderGasTip(big.NewInt(15000))
	priced.Reheap()

	// After headerGasTip change to 15000:
	// Normal transactions now use headerGasTip=15000 instead of 10000
	// dynAuthTx1: effectiveTip=min(20000, 40000-10000)=min(20000, 30000)=20000 (unchanged)
	// dynAuthTx2: effectiveTip=min(15000, 30000-10000)=min(15000, 20000)=15000 (unchanged)
	// dynNormTx1: effectiveTip=min(15000, 40000-10000)=min(15000, 30000)=15000 (changed from 10000!)
	// dynNormTx2: effectiveTip=min(15000, 45000-10000)=min(15000, 35000)=15000 (changed from 10000!)
	// dynNormTx3: effectiveTip=min(15000, 35000-10000)=min(15000, 25000)=15000 (changed from 10000!)

	// New order should be (urgent has 4, floating has 1):
	// 1. dynAuthTx2 (effectiveTip=15000, feeCap=30000, nonce=1) - smallest (moved to floating)
	// 2. dynNormTx3 (effectiveTip=15000, feeCap=35000, nonce=0)
	// 3. dynNormTx1 (effectiveTip=15000, feeCap=40000, nonce=0)
	// 4. dynNormTx2 (effectiveTip=15000, feeCap=45000, nonce=0)
	// 5. dynAuthTx1 (effectiveTip=20000, feeCap=40000, nonce=0) - largest (highest priority)

	// Verify urgent heap has 4 transactions after headerGasTip change
	urgentCountAfter := priced.urgent.Len()
	floatingCountAfter := priced.floating.Len()
	if urgentCountAfter != 4 {
		t.Errorf("After headerGasTip change: expected 4 transactions in urgent, got %d (floating has %d)", urgentCountAfter, floatingCountAfter)
	}

	// Verify floating heap has 1 transaction
	if priced.floating.Len() != 1 {
		t.Errorf("After headerGasTip change: expected 1 transaction in floating, got %d", priced.floating.Len())
	}

	// Pop from urgent and verify sorting order (min-heap: smallest first)
	// Note: Reheap() moves the smallest transaction to floating heap using heap.Pop()
	// After headerGasTip=25000, the smallest transaction (dynAuthTx2) is moved to floating
	// Expected order in urgent heap (after smallest moved to floating):
	// 1. dynNormTx3 (effectiveTip=15000, feeCap=35000, nonce=0)
	// 2. dynNormTx1 (effectiveTip=15000, feeCap=40000, nonce=0)
	// 3. dynNormTx2 (effectiveTip=15000, feeCap=45000, nonce=0)
	// 4. dynAuthTx1 (effectiveTip=20000, feeCap=40000, nonce=0) - largest (highest priority)
	popped1 = heap.Pop(&priced.urgent).(*types.Transaction)
	if popped1.Hash() != dynNormTx3.Hash() {
		t.Errorf("After headerGasTip change: expected dynNormTx3 (effectiveTip=15000, feeCap=35000) to be popped first, got feeCap=%s", popped1.GasFeeCap().String())
	}

	popped2 = heap.Pop(&priced.urgent).(*types.Transaction)
	if popped2.Hash() != dynNormTx1.Hash() {
		t.Errorf("After headerGasTip change: expected dynNormTx1 (effectiveTip=15000, feeCap=40000) to be popped second, got feeCap=%s", popped2.GasFeeCap().String())
	}

	popped3 = heap.Pop(&priced.urgent).(*types.Transaction)
	if popped3.Hash() != dynNormTx2.Hash() {
		t.Errorf("After headerGasTip change: expected dynNormTx2 (effectiveTip=15000, feeCap=45000) to be popped third, got feeCap=%s", popped3.GasFeeCap().String())
	}

	popped4 = heap.Pop(&priced.urgent).(*types.Transaction)
	if popped4.Hash() != dynAuthTx1.Hash() {
		t.Errorf("After headerGasTip change: expected dynAuthTx1 (effectiveTip=20000, feeCap=40000) to be popped fourth, got feeCap=%s", popped4.GasFeeCap().String())
	}

	floatingTx = heap.Pop(&priced.floating).(*types.Transaction)
	if floatingTx.Hash() != dynAuthTx2.Hash() {
		t.Errorf("After headerGasTip change: expected dynAuthTx2 (effectiveTip=15000, feeCap=30000) to be in floating heap")
	}

	t.Logf("After headerGasTip change: urgent has 4, floating has 1, all in correct order")
	t.Logf("HeaderGasTip change from 10000 to 25000 successfully affected normal transactions' effectiveTip")

	t.Logf("=== All test cases passed! ===")
}

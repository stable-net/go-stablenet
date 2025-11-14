// Copyright 2014 The go-ethereum Authors
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

package miner

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
)

var (
	// Test accounts for authorized account tests
	testAuthorizedKey1, _ = crypto.GenerateKey()
	testAuthorizedAddr1   = crypto.PubkeyToAddress(testAuthorizedKey1.PublicKey)
	testAuthorizedKey2, _ = crypto.GenerateKey()
	testAuthorizedAddr2   = crypto.PubkeyToAddress(testAuthorizedKey2.PublicKey)
	testAuthorizedKey3, _ = crypto.GenerateKey()
	testAuthorizedAddr3   = crypto.PubkeyToAddress(testAuthorizedKey3.PublicKey)

	// Test accounts for normal account tests
	testNormalKey1, _ = crypto.GenerateKey()
	testNormalAddr1   = crypto.PubkeyToAddress(testNormalKey1.PublicKey)
	testNormalKey2, _ = crypto.GenerateKey()
	testNormalAddr2   = crypto.PubkeyToAddress(testNormalKey2.PublicKey)
	testNormalKey3, _ = crypto.GenerateKey()
	testNormalAddr3   = crypto.PubkeyToAddress(testNormalKey3.PublicKey)

	// Pre-configured stateDB for tests
	testStateDB = setupTestStateDB()
)

// setupTestStateDB creates a stateDB for testing and sets up accounts.
func setupTestStateDB() *state.StateDB {
	// Create stateDB for testing (prepared for future StateAccount.Extra field usage)
	db := rawdb.NewMemoryDatabase()
	sdb := state.NewDatabase(db)
	stateDB, _ := state.New(types.EmptyRootHash, sdb, nil)

	// Include all test accounts in stateDB
	allAddrs := []common.Address{
		testAuthorizedAddr1, testAuthorizedAddr2, testAuthorizedAddr3,
		testNormalAddr1, testNormalAddr2, testNormalAddr3,
	}
	for _, addr := range allAddrs {
		stateDB.CreateAccount(addr)
	}

	if stateDB != nil {
		stateDB.SetAuthorized(testAuthorizedAddr1)
		stateDB.SetAuthorized(testAuthorizedAddr2)
		stateDB.SetAuthorized(testAuthorizedAddr3)
	}
	return stateDB
}

func TestTransactionPriceNonceSortLegacy(t *testing.T) {
	t.Parallel()
	testTransactionPriceNonceSort(t, nil)
}

func TestTransactionPriceNonceSort1559(t *testing.T) {
	t.Parallel()
	testTransactionPriceNonceSort(t, big.NewInt(0))
	testTransactionPriceNonceSort(t, big.NewInt(5))
	testTransactionPriceNonceSort(t, big.NewInt(50))
}

// Tests that transactions can be correctly sorted according to their price in
// decreasing order, but at the same time with increasing nonces when issued by
// the same account.
func testTransactionPriceNonceSort(t *testing.T, baseFee *big.Int) {
	// Generate a batch of accounts to start with
	keys := make([]*ecdsa.PrivateKey, 25)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
	}
	signer := types.LatestSignerForChainID(common.Big1)

	// Generate a batch of transactions with overlapping values, but shifted nonces
	groups := map[common.Address][]*txpool.LazyTransaction{}
	expectedCount := 0
	for start, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		count := 25
		for i := 0; i < 25; i++ {
			var tx *types.Transaction
			gasFeeCap := rand.Intn(50)
			if baseFee == nil {
				tx = types.NewTx(&types.LegacyTx{
					Nonce:    uint64(start + i),
					To:       &common.Address{},
					Value:    big.NewInt(100),
					Gas:      100,
					GasPrice: big.NewInt(int64(gasFeeCap)),
					Data:     nil,
				})
			} else {
				tx = types.NewTx(&types.DynamicFeeTx{
					Nonce:     uint64(start + i),
					To:        &common.Address{},
					Value:     big.NewInt(100),
					Gas:       100,
					GasFeeCap: big.NewInt(int64(gasFeeCap)),
					GasTipCap: big.NewInt(int64(rand.Intn(gasFeeCap + 1))),
					Data:      nil,
				})
				if count == 25 && int64(gasFeeCap) < baseFee.Int64() {
					count = i
				}
			}
			tx, err := types.SignTx(tx, signer, key)
			if err != nil {
				t.Fatalf("failed to sign tx: %s", err)
			}
			groups[addr] = append(groups[addr], &txpool.LazyTransaction{
				Hash:      tx.Hash(),
				Tx:        tx,
				Time:      tx.Time(),
				GasFeeCap: uint256.MustFromBig(tx.GasFeeCap()),
				GasTipCap: uint256.MustFromBig(tx.GasTipCap()),
				Gas:       tx.Gas(),
				BlobGas:   tx.BlobGas(),
			})
		}
		expectedCount += count
	}
	// Sort the transactions and cross check the nonce ordering
	txset := newTransactionsByPriceAndNonce(signer, groups, baseFee, false, nil)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}
	if len(txs) != expectedCount {
		t.Errorf("expected %d transactions, found %d", expectedCount, len(txs))
	}
	for i, txi := range txs {
		fromi, _ := types.Sender(signer, txi)

		// Make sure the nonce order is valid
		for j, txj := range txs[i+1:] {
			fromj, _ := types.Sender(signer, txj)
			if fromi == fromj && txi.Nonce() > txj.Nonce() {
				t.Errorf("invalid nonce ordering: tx #%d (A=%x N=%v) < tx #%d (A=%x N=%v)", i, fromi[:4], txi.Nonce(), i+j, fromj[:4], txj.Nonce())
			}
		}
		// If the next tx has different from account, the price must be lower than the current one
		if i+1 < len(txs) {
			next := txs[i+1]
			fromNext, _ := types.Sender(signer, next)
			tip, err := txi.EffectiveGasTip(baseFee)
			nextTip, nextErr := next.EffectiveGasTip(baseFee)
			if err != nil || nextErr != nil {
				t.Errorf("error calculating effective tip: %v, %v", err, nextErr)
			}
			if fromi != fromNext && tip.Cmp(nextTip) < 0 {
				t.Errorf("invalid gasprice ordering: tx #%d (A=%x P=%v) < tx #%d (A=%x P=%v)", i, fromi[:4], txi.GasPrice(), i+1, fromNext[:4], next.GasPrice())
			}
		}
	}
}

// Tests that if multiple transactions have the same price, the ones seen earlier
// are prioritized to avoid network spam attacks aiming for a specific ordering.
func TestTransactionTimeSort(t *testing.T) {
	t.Parallel()
	// Generate a batch of accounts to start with
	keys := make([]*ecdsa.PrivateKey, 5)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
	}
	signer := types.HomesteadSigner{}

	// Generate a batch of transactions with overlapping prices, but different creation times
	groups := map[common.Address][]*txpool.LazyTransaction{}
	for start, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)

		tx, _ := types.SignTx(types.NewTransaction(0, common.Address{}, big.NewInt(100), 100, big.NewInt(1), nil), signer, key)
		tx.SetTime(time.Unix(0, int64(len(keys)-start)))

		groups[addr] = append(groups[addr], &txpool.LazyTransaction{
			Hash:      tx.Hash(),
			Tx:        tx,
			Time:      tx.Time(),
			GasFeeCap: uint256.MustFromBig(tx.GasFeeCap()),
			GasTipCap: uint256.MustFromBig(tx.GasTipCap()),
			Gas:       tx.Gas(),
			BlobGas:   tx.BlobGas(),
		})
	}
	// Sort the transactions and cross check the nonce ordering
	txset := newTransactionsByPriceAndNonce(signer, groups, nil, false, nil)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}
	if len(txs) != len(keys) {
		t.Errorf("expected %d transactions, found %d", len(keys), len(txs))
	}
	for i, txi := range txs {
		fromi, _ := types.Sender(signer, txi)
		if i+1 < len(txs) {
			next := txs[i+1]
			fromNext, _ := types.Sender(signer, next)

			if txi.GasPrice().Cmp(next.GasPrice()) < 0 {
				t.Errorf("invalid gasprice ordering: tx #%d (A=%x P=%v) < tx #%d (A=%x P=%v)", i, fromi[:4], txi.GasPrice(), i+1, fromNext[:4], next.GasPrice())
			}
			// Make sure time order is ascending if the txs have the same gas price
			if txi.GasPrice().Cmp(next.GasPrice()) == 0 && txi.Time().After(next.Time()) {
				t.Errorf("invalid received time ordering: tx #%d (A=%x T=%v) > tx #%d (A=%x T=%v)", i, fromi[:4], txi.Time(), i+1, fromNext[:4], next.Time())
			}
		}
	}
}

// TestAuthorizedAccountPriority tests the priority ordering for authorized accounts
func TestAuthorizedAccountPriority(t *testing.T) {
	t.Parallel()

	signer := types.LatestSignerForChainID(common.Big1)

	// Use pre-configured stateDB with authorized accounts
	stateDB := testStateDB

	groups := map[common.Address][]*txpool.LazyTransaction{}

	// Authorized account 1
	tx1, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(20),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testAuthorizedKey1)
	tx1.SetTime(time.Unix(4, 0))
	groups[testAuthorizedAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx1.Hash(),
		Tx:        tx1,
		Time:      tx1.Time(),
		GasFeeCap: uint256.MustFromBig(tx1.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx1.GasTipCap()),
		Gas:       tx1.Gas(),
		BlobGas:   tx1.BlobGas(),
	}}

	// Authorized account 2
	tx2, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(15),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testAuthorizedKey2)
	tx2.SetTime(time.Unix(3, 0))
	groups[testAuthorizedAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx2.Hash(),
		Tx:        tx2,
		Time:      tx2.Time(),
		GasFeeCap: uint256.MustFromBig(tx2.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx2.GasTipCap()),
		Gas:       tx2.Gas(),
		BlobGas:   tx2.BlobGas(),
	}}

	// Normal account 1
	tx3, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(200),
		GasTipCap: big.NewInt(100),
		Data:      nil,
	}), signer, testNormalKey1)
	tx3.SetTime(time.Unix(2, 0))
	groups[testNormalAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx3.Hash(),
		Tx:        tx3,
		Time:      tx3.Time(),
		GasFeeCap: uint256.MustFromBig(tx3.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx3.GasTipCap()),
		Gas:       tx3.Gas(),
		BlobGas:   tx3.BlobGas(),
	}}

	// Normal account 2
	tx4, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(15),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey2)
	tx4.SetTime(time.Unix(1, 0))
	groups[testNormalAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx4.Hash(),
		Tx:        tx4,
		Time:      tx4.Time(),
		GasFeeCap: uint256.MustFromBig(tx4.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx4.GasTipCap()),
		Gas:       tx4.Gas(),
		BlobGas:   tx4.BlobGas(),
	}}

	txset := newTransactionsByPriceAndNonce(signer, groups, big.NewInt(0), true, stateDB)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}

	// fee order: tx3 > tx1 > tx4 = tx2
	// time order: tx4 > tx3 > tx2 > tx1
	// tx1(testAuthorizedAddr1), tx2(testAuthorizedAddr2) : Authorized account
	// tx3(testNormalAddr1),tx4(testNormalAddr2) : Normal account

	// Expected order:
	// tx1 > tx2 > tx4 > tx3
	if len(txs) != 4 {
		t.Fatalf("expected 4 transactions, found %d", len(txs))
	}

	from1, _ := types.Sender(signer, txs[0])
	from2, _ := types.Sender(signer, txs[1])
	from3, _ := types.Sender(signer, txs[2])
	from4, _ := types.Sender(signer, txs[3])

	// First two should be authorized accounts (ordered by fee)
	if from1 != testAuthorizedAddr1 {
		t.Errorf("expected first tx from authorized account 1, got %x", from1)
	}
	if from2 != testAuthorizedAddr2 {
		t.Errorf("expected second tx from authorized account 2, got %x", from2)
	}

	// Last two should be normal accounts (ordered by time/FIFO)
	if from3 != testNormalAddr2 {
		t.Errorf("expected third tx from normal account 1, got %x", from3)
	}
	if from4 != testNormalAddr1 {
		t.Errorf("expected fourth tx from normal account 2, got %x", from4)
	}
}

// TestAuthorizedAccountPrioritySameFee tests authorized accounts with same fee (should use FIFO)
func TestAuthorizedAccountPrioritySameFee(t *testing.T) {
	t.Parallel()

	signer := types.LatestSignerForChainID(common.Big1)

	// Use pre-configured stateDB with authorized accounts
	stateDB := testStateDB

	groups := map[common.Address][]*txpool.LazyTransaction{}

	// Authorized account 1 - same fee
	tx1, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(20),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testAuthorizedKey1)
	tx1.SetTime(time.Unix(1, 0))
	groups[testAuthorizedAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx1.Hash(),
		Tx:        tx1,
		Time:      tx1.Time(),
		GasFeeCap: uint256.MustFromBig(tx1.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx1.GasTipCap()),
		Gas:       tx1.Gas(),
		BlobGas:   tx1.BlobGas(),
	}}

	// Authorized account 2 - same fee
	tx2, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(20),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testAuthorizedKey2)
	tx2.SetTime(time.Unix(3, 0))
	groups[testAuthorizedAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx2.Hash(),
		Tx:        tx2,
		Time:      tx2.Time(),
		GasFeeCap: uint256.MustFromBig(tx2.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx2.GasTipCap()),
		Gas:       tx2.Gas(),
		BlobGas:   tx2.BlobGas(),
	}}

	// Authorized account 3 - same fee
	tx3, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(20),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testAuthorizedKey3)
	tx3.SetTime(time.Unix(2, 0))
	groups[testAuthorizedAddr3] = []*txpool.LazyTransaction{{
		Hash:      tx3.Hash(),
		Tx:        tx3,
		Time:      tx3.Time(),
		GasFeeCap: uint256.MustFromBig(tx3.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx3.GasTipCap()),
		Gas:       tx3.Gas(),
		BlobGas:   tx3.BlobGas(),
	}}

	txset := newTransactionsByPriceAndNonce(signer, groups, big.NewInt(0), true, stateDB)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}

	// fee : same for all authorized accounts
	// time order: tx1 > tx3 > tx2

	// Expected order:
	// tx1 > tx3 > tx2
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, found %d", len(txs))
	}

	from1, _ := types.Sender(signer, txs[0])
	from2, _ := types.Sender(signer, txs[1])
	from3, _ := types.Sender(signer, txs[2])

	// Should be ordered by time (FIFO) when fees are equal
	if from1 != testAuthorizedAddr1 {
		t.Errorf("expected first tx from authorized account 1, got %x", from1)
	}
	if from2 != testAuthorizedAddr3 {
		t.Errorf("expected second tx from authorized account 3, got %x", from2)
	}
	if from3 != testAuthorizedAddr2 {
		t.Errorf("expected third tx from authorized account 2, got %x", from3)
	}
}

// TestNotAuthorizedFIFO tests that normal accounts are ordered by FIFO regardless of fee
func TestNotAuthorizedFIFO(t *testing.T) {
	t.Parallel()

	signer := types.LatestSignerForChainID(common.Big1)

	// Use pre-configured stateDB
	stateDB := testStateDB

	groups := map[common.Address][]*txpool.LazyTransaction{}

	// Normal account 1
	tx1, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(30),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testNormalKey1)
	tx1.SetTime(time.Unix(2, 0))
	groups[testNormalAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx1.Hash(),
		Tx:        tx1,
		Time:      tx1.Time(),
		GasFeeCap: uint256.MustFromBig(tx1.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx1.GasTipCap()),
		Gas:       tx1.Gas(),
		BlobGas:   tx1.BlobGas(),
	}}

	// Normal account 2
	tx2, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(40),
		GasTipCap: big.NewInt(20),
		Data:      nil,
	}), signer, testNormalKey2)
	tx2.SetTime(time.Unix(3, 0))
	groups[testNormalAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx2.Hash(),
		Tx:        tx2,
		Time:      tx2.Time(),
		GasFeeCap: uint256.MustFromBig(tx2.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx2.GasTipCap()),
		Gas:       tx2.Gas(),
		BlobGas:   tx2.BlobGas(),
	}}

	// Normal account 3
	tx3, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(10),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey3)
	tx3.SetTime(time.Unix(1, 0))
	groups[testNormalAddr3] = []*txpool.LazyTransaction{{
		Hash:      tx3.Hash(),
		Tx:        tx3,
		Time:      tx3.Time(),
		GasFeeCap: uint256.MustFromBig(tx3.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx3.GasTipCap()),
		Gas:       tx3.Gas(),
		BlobGas:   tx3.BlobGas(),
	}}

	txset := newTransactionsByPriceAndNonce(signer, groups, big.NewInt(0), true, stateDB)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}

	// fee order : Normal account is not considered, only authorized accounts are considered for fee ordering in anzeon enabled mode
	// time order: tx3 > tx1 > tx2

	// Expected order:
	// tx3 > tx1 > tx2
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, found %d", len(txs))
	}

	from1, _ := types.Sender(signer, txs[0])
	from2, _ := types.Sender(signer, txs[1])
	from3, _ := types.Sender(signer, txs[2])

	// Should be ordered by time (FIFO) regardless of fee in anzeon enabled mode
	if from1 != testNormalAddr3 {
		t.Errorf("expected first tx from normal account 3, got %x", from1)
	}
	if from2 != testNormalAddr1 {
		t.Errorf("expected second tx from normal account 1, got %x", from2)
	}
	if from3 != testNormalAddr2 {
		t.Errorf("expected third tx from normal account 2, got %x", from3)
	}
}

// TestAnzeonDisabledAndFIFO tests that normal accounts are ordered by FIFO regardless of fee
func TestAnzeonDisabledAndFIFO(t *testing.T) {
	t.Parallel()

	signer := types.LatestSignerForChainID(common.Big1)

	groups := map[common.Address][]*txpool.LazyTransaction{}

	// Normal account 1
	tx1, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(10),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey1)
	tx1.SetTime(time.Unix(3, 0))
	groups[testNormalAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx1.Hash(),
		Tx:        tx1,
		Time:      tx1.Time(),
		GasFeeCap: uint256.MustFromBig(tx1.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx1.GasTipCap()),
		Gas:       tx1.Gas(),
		BlobGas:   tx1.BlobGas(),
	}}

	// Normal account 2
	tx2, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(10),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey2)
	tx2.SetTime(time.Unix(1, 0))
	groups[testNormalAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx2.Hash(),
		Tx:        tx2,
		Time:      tx2.Time(),
		GasFeeCap: uint256.MustFromBig(tx2.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx2.GasTipCap()),
		Gas:       tx2.Gas(),
		BlobGas:   tx2.BlobGas(),
	}}

	// Normal account 3
	tx3, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(10),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey3)
	tx3.SetTime(time.Unix(2, 0))
	groups[testNormalAddr3] = []*txpool.LazyTransaction{{
		Hash:      tx3.Hash(),
		Tx:        tx3,
		Time:      tx3.Time(),
		GasFeeCap: uint256.MustFromBig(tx3.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx3.GasTipCap()),
		Gas:       tx3.Gas(),
		BlobGas:   tx3.BlobGas(),
	}}

	txset := newTransactionsByPriceAndNonce(signer, groups, big.NewInt(0), false, nil)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}

	// fee order : same for all normal accounts
	// time order: tx2 > tx3 > tx1

	// Expected order:
	// tx2 > tx3 > tx1
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, found %d", len(txs))
	}

	from1, _ := types.Sender(signer, txs[0])
	from2, _ := types.Sender(signer, txs[1])
	from3, _ := types.Sender(signer, txs[2])

	// Should be ordered by time (FIFO) when fees are equal
	if from1 != testNormalAddr2 {
		t.Errorf("expected first tx from normal account 2, got %x", from1)
	}
	if from2 != testNormalAddr3 {
		t.Errorf("expected second tx from normal account 3, got %x", from2)
	}
	if from3 != testNormalAddr1 {
		t.Errorf("expected third tx from normal account 1, got %x", from3)
	}
}

// TestAnzeonDisabledAndHigherFeeFirst tests that when Anzeon is disabled, original fee-based ordering is used and higher fee first
func TestAnzeonDisabledAndHigherFeeFirst(t *testing.T) {
	t.Parallel()

	signer := types.LatestSignerForChainID(common.Big1)

	groups := map[common.Address][]*txpool.LazyTransaction{}

	// Account 1
	tx1, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(10),
		GasTipCap: big.NewInt(5),
		Data:      nil,
	}), signer, testNormalKey1)
	tx1.SetTime(time.Unix(1, 0))
	groups[testNormalAddr1] = []*txpool.LazyTransaction{{
		Hash:      tx1.Hash(),
		Tx:        tx1,
		Time:      tx1.Time(),
		GasFeeCap: uint256.MustFromBig(tx1.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx1.GasTipCap()),
		Gas:       tx1.Gas(),
		BlobGas:   tx1.BlobGas(),
	}}

	// Account 2
	tx2, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(30),
		GasTipCap: big.NewInt(20),
		Data:      nil,
	}), signer, testNormalKey2)
	tx1.SetTime(time.Unix(2, 0))
	groups[testNormalAddr2] = []*txpool.LazyTransaction{{
		Hash:      tx2.Hash(),
		Tx:        tx2,
		Time:      tx2.Time(),
		GasFeeCap: uint256.MustFromBig(tx2.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx2.GasTipCap()),
		Gas:       tx2.Gas(),
		BlobGas:   tx2.BlobGas(),
	}}

	// Account 3
	tx3, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(100),
		Gas:       100,
		GasFeeCap: big.NewInt(25),
		GasTipCap: big.NewInt(10),
		Data:      nil,
	}), signer, testNormalKey3)
	tx3.SetTime(time.Unix(3, 0))
	groups[testNormalAddr3] = []*txpool.LazyTransaction{{
		Hash:      tx3.Hash(),
		Tx:        tx3,
		Time:      tx3.Time(),
		GasFeeCap: uint256.MustFromBig(tx3.GasFeeCap()),
		GasTipCap: uint256.MustFromBig(tx3.GasTipCap()),
		Gas:       tx3.Gas(),
		BlobGas:   tx3.BlobGas(),
	}}

	txset := newTransactionsByPriceAndNonce(signer, groups, big.NewInt(0), false, nil)

	txs := types.Transactions{}
	for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
		txs = append(txs, tx.Tx)
		txset.Shift()
	}

	// fee order : tx2 > tx3 > tx1
	// time order: tx1 > tx2 > tx3

	// Expected order:
	// tx2 > tx3 > tx1
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, found %d", len(txs))
	}

	from1, _ := types.Sender(signer, txs[0])
	from2, _ := types.Sender(signer, txs[1])
	from3, _ := types.Sender(signer, txs[2])

	// Should be ordered by fee (higher first) when Anzeon is disabled
	if from1 != testNormalAddr2 {
		t.Errorf("expected first tx from account 2, got %x", from1)
	}
	if from2 != testNormalAddr3 {
		t.Errorf("expected second tx from account 3, got %x", from2)
	}
	if from3 != testNormalAddr1 {
		t.Errorf("expected third tx from account 1, got %x", from3)
	}
}

// Benchmark helpers
func generateTransactions(numAccounts, txsPerAccount int, authorizedRatio float64, rng *rand.Rand) (map[common.Address][]*txpool.LazyTransaction, *state.StateDB) {
	signer := types.LatestSignerForChainID(common.Big1)
	db := rawdb.NewMemoryDatabase()
	sdb := state.NewDatabase(db)
	stateDB, _ := state.New(types.EmptyRootHash, sdb, nil)

	groups := make(map[common.Address][]*txpool.LazyTransaction)
	authorizedCount := int(float64(numAccounts) * authorizedRatio)

	for i := 0; i < numAccounts; i++ {
		key, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(key.PublicKey)
		stateDB.CreateAccount(addr)

		isAuthorized := i < authorizedCount
		if isAuthorized {
			stateDB.SetAuthorized(addr)
		}

		for j := 0; j < txsPerAccount; j++ {
			gasFeeCap := big.NewInt(int64(10 + rng.Intn(100)))
			gasTipCap := big.NewInt(int64(1 + rng.Intn(int(gasFeeCap.Int64()))))
			txTime := time.Unix(int64(rng.Intn(1000000)), 0)

			tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
				Nonce:     uint64(j),
				To:        &common.Address{},
				Value:     big.NewInt(100),
				Gas:       100,
				GasFeeCap: gasFeeCap,
				GasTipCap: gasTipCap,
				Data:      nil,
			}), signer, key)
			tx.SetTime(txTime)

			groups[addr] = append(groups[addr], &txpool.LazyTransaction{
				Hash:      tx.Hash(),
				Tx:        tx,
				Time:      tx.Time(),
				GasFeeCap: uint256.MustFromBig(tx.GasFeeCap()),
				GasTipCap: uint256.MustFromBig(tx.GasTipCap()),
				Gas:       tx.Gas(),
				BlobGas:   tx.BlobGas(),
			})
		}
	}

	return groups, stateDB
}

func benchmarkOrdering(b *testing.B, numAccounts, txsPerAccount int, authorizedRatio float64, anzeonEnabled bool) {
	// Generate transactions once before starting the timer
	rng := rand.New(rand.NewSource(42))
	groups, stateDB := generateTransactions(numAccounts, txsPerAccount, authorizedRatio, rng)
	signer := types.LatestSignerForChainID(common.Big1)
	baseFee := big.NewInt(0)

	// Prepare groups copy template (will be copied in each iteration)
	groupsTemplate := make(map[common.Address][]*txpool.LazyTransaction)
	for addr, txs := range groups {
		txsCopy := make([]*txpool.LazyTransaction, len(txs))
		copy(txsCopy, txs)
		groupsTemplate[addr] = txsCopy
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Recreate groups for each iteration
		groupsCopy := make(map[common.Address][]*txpool.LazyTransaction)
		for addr, txs := range groupsTemplate {
			txsCopy := make([]*txpool.LazyTransaction, len(txs))
			copy(txsCopy, txs)
			groupsCopy[addr] = txsCopy
		}

		txset := newTransactionsByPriceAndNonce(signer, groupsCopy, baseFee, anzeonEnabled, stateDB)

		// Consume all transactions to measure full sorting performance
		for tx, _ := txset.Peek(); tx != nil; tx, _ = txset.Peek() {
			txset.Shift()
		}
	}
}

// Benchmark_10Accounts_10Txs benchmarks ordering with 10 accounts, 10 transactions per account
// Compares different account types and configurations
func Benchmark_10Accounts_10Txs(b *testing.B) {
	b.Run("AuthorizedOnly", func(b *testing.B) {
		benchmarkOrdering(b, 10, 10, 1.0, true)
	})
	b.Run("NormalOnly", func(b *testing.B) {
		benchmarkOrdering(b, 10, 10, 0.0, true)
	})
	b.Run("Mixed_50Percent", func(b *testing.B) {
		benchmarkOrdering(b, 10, 10, 0.5, true)
	})
	b.Run("AnzeonDisabled", func(b *testing.B) {
		benchmarkOrdering(b, 10, 10, 0.5, false)
	})
}

// Benchmark_50Accounts_10Txs benchmarks ordering with 50 accounts, 10 transactions per account
// Compares different account types and configurations
func Benchmark_50Accounts_10Txs(b *testing.B) {
	b.Run("AuthorizedOnly", func(b *testing.B) {
		benchmarkOrdering(b, 50, 10, 1.0, true)
	})
	b.Run("NormalOnly", func(b *testing.B) {
		benchmarkOrdering(b, 50, 10, 0.0, true)
	})
	b.Run("Mixed_50Percent", func(b *testing.B) {
		benchmarkOrdering(b, 50, 10, 0.5, true)
	})
	b.Run("AnzeonDisabled", func(b *testing.B) {
		benchmarkOrdering(b, 50, 10, 0.5, false)
	})
}

// Benchmark_100Accounts_10Txs benchmarks ordering with 100 accounts, 10 transactions per account
// Compares different account types and configurations
func Benchmark_100Accounts_10Txs(b *testing.B) {
	b.Run("AuthorizedOnly", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 1.0, true)
	})
	b.Run("NormalOnly", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 0.0, true)
	})
	b.Run("Mixed_50Percent", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 0.5, true)
	})
	b.Run("Mixed_20Percent", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 0.2, true)
	})
	b.Run("Mixed_80Percent", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 0.8, true)
	})
	b.Run("AnzeonDisabled", func(b *testing.B) {
		benchmarkOrdering(b, 100, 10, 0.5, false)
	})
}

// Benchmark_100Accounts_50Txs benchmarks ordering with 100 accounts, 50 transactions per account
// Compares different account types and configurations
func Benchmark_100Accounts_50Txs(b *testing.B) {
	b.Run("AuthorizedOnly", func(b *testing.B) {
		benchmarkOrdering(b, 100, 50, 1.0, true)
	})
	b.Run("NormalOnly", func(b *testing.B) {
		benchmarkOrdering(b, 100, 50, 0.0, true)
	})
	b.Run("Mixed_50Percent", func(b *testing.B) {
		benchmarkOrdering(b, 100, 50, 0.5, true)
	})
	b.Run("AnzeonDisabled", func(b *testing.B) {
		benchmarkOrdering(b, 100, 50, 0.5, false)
	})
}

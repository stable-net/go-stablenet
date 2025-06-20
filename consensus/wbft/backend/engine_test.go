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
// This file is derived from quorum/consensus/istanbul/backend/engine_test.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	wbftcore "github.com/ethereum/go-ethereum/consensus/wbft/core"
	wbftengine "github.com/ethereum/go-ethereum/consensus/wbft/engine"
	"github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/consensus/wbft/testutils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/eth/fetcher"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)

var blockEnqueueChannel chan *types.Block

type fakeBroadcaster struct {
	blockFetcher *fetcher.BlockFetcher
}

type otherNode struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
	balance    *big.Int
}

func makeFakeBroadcaster(chain *core.BlockChain) *fakeBroadcaster {
	blockEnqueueChannel = make(chan *types.Block)
	validator := func(header *types.Header) error {
		return chain.Engine().VerifyHeader(chain, header)
	}

	broadcastBlock := func(block *types.Block, propagate bool) {}

	heighter := func() uint64 {
		return chain.CurrentBlock().Number.Uint64()
	}

	//inserter := func(blocks types.Blocks) (int, error) {
	//	idx, err := chain.InsertChain(blocks)
	//	if err == nil {
	//		header := blocks[len(blocks)-1].Header()
	//		chain.SetFinalized(header)
	//		chain.SetSafe(header)
	//	}
	//	// ## Wemix END
	//	return idx, err
	//}

	fb := fakeBroadcaster{
		blockFetcher: fetcher.NewBlockFetcher(false, nil, chain.GetBlockByHash, validator, broadcastBlock, heighter, nil, nil, nil),
	}
	fb.blockFetcher.Start()
	return &fb
}

func (fb *fakeBroadcaster) Enqueue(id string, block *types.Block) {
	go func() { blockEnqueueChannel <- block }()
}

func (fb *fakeBroadcaster) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	return m
}

func newBlockchainFromConfig(genesis *core.Genesis, nodeKeys []*ecdsa.PrivateKey, cfg *wbft.Config) (*core.BlockChain, *Backend, []otherNode) {
	memDB := rawdb.NewMemoryDatabase()

	// Use the first key as private key
	backend := New(cfg, nodeKeys[0], memDB)

	genesis.MustCommit(memDB, triedb.NewDatabase(memDB, triedb.HashDefaults))

	blockchain, err := core.NewBlockChain(memDB, nil, genesis, nil, backend, vm.Config{}, nil, nil)
	if err != nil {
		panic(err)
	}

	state, err := blockchain.StateAt(blockchain.Genesis().Root())
	if state == nil || err != nil {
		panic(err)
	}

	// Make virtual node struct for simulation
	nodes := make([]otherNode, 0)
	for i := 0; i < len(nodeKeys); i++ {
		address := crypto.PubkeyToAddress(nodeKeys[i].PublicKey)
		b := state.GetBalance(address).ToBig()
		nodes = append(nodes, otherNode{address, nodeKeys[i], b})
	}

	fb := makeFakeBroadcaster(blockchain)
	backend.broadcaster = fb

	backend.Start(blockchain, blockchain.CurrentFullBlock, rawdb.HasBadBlock, nil)
	genesisBlock := blockchain.GetHeaderByHash(blockchain.Genesis().Hash())
	valSet, err := backend.Engine().GetValidators(blockchain, common.Big1, genesisBlock.Hash(), nil)

	if err != nil {
		panic(err)
	}
	if valSet == nil {
		panic("failed to get validator set")
	}
	proposerAddr := valSet.GetProposer().Address()

	// find proposer key
	for i, key := range nodeKeys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		if addr.String() == proposerAddr.String() {
			backend.privateKey = key
			backend.address = addr
			backend.wbftEngine = wbftengine.NewEngine(backend.config, addr, backend.Sign, backend.CheckSignature)
			backend.blsSecretKey, _ = bls.DeriveFromECDSA(key)
			// set backend's node to first index of nodes for convenient
			nodes[0], nodes[i] = nodes[i], nodes[0]
			break
		}
	}

	return blockchain, backend, nodes
}

// in this test, we can set n to 1, and it means we can process Istanbul and commit a
// block by one node. Otherwise, if n is larger than 1, we have to generate
// other fake events to process Istanbul.
func newBlockChain(n int) (*core.BlockChain, *Backend, []otherNode) {
	genesis, nodeKeys, _ := testutils.GenesisAndKeys(n)

	config := new(wbft.Config)
	wbft.SetConfigFromChainConfig(config, genesis.Config)

	return newBlockchainFromConfig(genesis, nodeKeys, config)
}

func newBlockChainWithCustom(n int, customizeConfig func(config *wbft.Config)) (*core.BlockChain, *Backend, []otherNode) {
	genesis, nodeKeys, _ := testutils.GenesisAndKeys(n)

	config := new(wbft.Config)
	wbft.SetConfigFromChainConfig(config, genesis.Config)
	customizeConfig(config)

	return newBlockchainFromConfig(genesis, nodeKeys, config)
}

// makeHeader create header executing no txs
func makeHeader(chainConfig *params.ChainConfig, engineConfig *wbft.Config, parent *types.Block) *types.Header {
	blockNumber := parent.Number().Add(parent.Number(), common.Big1)
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     blockNumber,
		GasLimit:   parent.GasLimit(),
		GasUsed:    0, // empty tx
		Time:       parent.Time() + engineConfig.GetConfig(blockNumber).BlockPeriod,
		Difficulty: types.WBFTDefaultDifficulty,
		BaseFee:    eip1559.CalcBaseFee(chainConfig, parent.Header()),
	}
	return header
}

// makeBlock create block executing no txs
func makeBlock(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	block := makeBlockWithoutSeal(chain, engine, parent)
	state, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), state, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	engine.Seal(chain, block, resultCh, nil)
	blk := <-resultCh
	return blk
}

// makeBlock create block executing no txs without seal
func makeBlockWithoutSeal(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	engine.NewChainHead() // progress to next sequence
	if engine.core.GetState() != wbftcore.StateAcceptRequest {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			<-ticker.C
			if engine.core.GetState() == wbftcore.StateAcceptRequest {
				ticker.Stop()
				break
			}
		}
	}
	header := makeHeader(chain.Config(), engine.config, parent)
	engine.Prepare(chain, header)
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}

func TestWBFTPrepare(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	header := makeHeader(chain.Config(), engine.config, chain.Genesis())
	err := engine.Prepare(chain, header)
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}

	header.ParentHash = common.BytesToHash([]byte("1234567890"))
	err = engine.Prepare(chain, header)
	if err != consensus.ErrUnknownAncestor {
		t.Errorf("error mismatch: have %v, want %v", err, consensus.ErrUnknownAncestor)
	}
}

func TestSealStopChannel(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	stop := make(chan struct{}, 1)
	eventSub := engine.EventMux().Subscribe(wbft.RequestEvent{})
	eventLoop := func() {
		ev := <-eventSub.Chan()
		_, ok := ev.Data.(wbft.RequestEvent)
		if !ok {
			t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
		}
		stop <- struct{}{}
		eventSub.Unsubscribe()
	}
	resultCh := make(chan *types.Block, 10)
	go func() {
		err := engine.Seal(chain, block, resultCh, stop)
		if err != nil {
			t.Errorf("error mismatch: have %v, want nil", err)
		}
	}()
	go eventLoop()

	finalBlock := <-resultCh
	if finalBlock != nil {
		t.Errorf("block mismatch: have %v, want nil", finalBlock)
	}
}

func TestSealCommittedOtherHash(t *testing.T) {
	chain, engine, _ := newBlockChain(2)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	otherBlock := makeBlockWithoutSeal(chain, engine, block)
	expectedPreparedSeal := engine.blsSecretKey.Sign(append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)).Marshal()
	expectedCommittedSeal := engine.blsSecretKey.Sign(append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)).Marshal()

	eventSub := engine.EventMux().Subscribe(wbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	blockOutputChannel := make(chan *types.Block)
	stopChannel := make(chan struct{})

	if err := engine.Seal(chain, block, blockOutputChannel, stopChannel); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(wbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(otherBlock,
		[]wbft.SealData{{Sealer: 0, Seal: expectedPreparedSeal}},
		[]wbft.SealData{{Sealer: 0, Seal: expectedCommittedSeal}},
		big.NewInt(0)); err != nil {
		t.Error(err.Error())
	}

	select {
	case <-blockOutputChannel:
		t.Error("Wrong block found!")
	case <-time.After(time.Second):
		//no block found, stop the sealing
		close(stopChannel)
	}

	output := <-blockOutputChannel
	if output != nil {
		t.Error("Block not nil!")
	}
}

func updateWBFTBlock(block *types.Block, addr common.Address) *types.Block {
	header := block.Header()
	header.Coinbase = addr
	return block.WithSeal(header)
}

func TestSealCommitted(t *testing.T) {
	chain, engine, _ := newBlockChain(2)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	expectedBlock := updateWBFTBlock(block, engine.Address())
	expectedPreparedSeal := engine.blsSecretKey.Sign(append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)).Marshal()
	expectedCommittedSeal := engine.blsSecretKey.Sign(append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)).Marshal()

	eventSub := engine.EventMux().Subscribe(wbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	resultCh := make(chan *types.Block, 10)

	if err := engine.Seal(chain, block, resultCh, make(chan struct{})); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(wbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(expectedBlock,
		[]wbft.SealData{{Sealer: 0, Seal: expectedPreparedSeal}},
		[]wbft.SealData{{Sealer: 0, Seal: expectedCommittedSeal}},
		big.NewInt(0)); err != nil {
		t.Error(err.Error())
	}

	finalBlock := <-resultCh
	if finalBlock.Hash() != expectedBlock.Hash() {
		t.Errorf("hash mismatch: have %v, want %v", finalBlock.Hash(), expectedBlock.Hash())
	}
}

func TestVerifyHeaderForChainedBlock(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()

	firstWbftBlock := makeBlock(chain, engine, chain.Genesis())
	_, err := chain.InsertChain(types.Blocks{firstWbftBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	secondWbftBlock := makeBlock(chain, engine, firstWbftBlock)
	_, err = chain.InsertChain(types.Blocks{secondWbftBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	//create chain consists of genesisblock - croissant hardfork block - regular wbft block
	testCases := []struct {
		block                  *types.Block
		headerManipulationFunc func(*types.Block) *types.Header
		expectedError          error
	}{
		{
			firstWbftBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractWBFTExtra(header)
				// write invalid round
				if _, err := wbftengine.ApplyHeaderWBFTExtra(header, wbftengine.WritePrevSeals(extra.PrevRound+1, extra.PrevPreparedSeal, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			wbftcommon.ErrInvalidPreparedSeals, // PrevPreparedSeal changed -> block hash changed -> prepare seal invalid
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractWBFTExtra(header)
				if _, err := wbftengine.ApplyHeaderWBFTExtra(header, wbftengine.WritePrevSeals(extra.PrevRound, nil, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			wbftcommon.ErrInvalidPreparedSeals, // PrevPreparedSeal changed -> block hash changed -> prepare seal invalid
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractWBFTExtra(header)
				if _, err := wbftengine.ApplyHeaderWBFTExtra(header, wbftengine.WritePrevSeals(extra.PrevRound, extra.PrevPreparedSeal, nil)); err != nil {
					return nil
				}
				return header
			},
			// for wbftBlock, invalid preparedSeal occurs when validating preparedSeal because header is changed and gets wrong signer address
			wbftcommon.ErrInvalidPreparedSeals,
		},
		{
			firstWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractWBFTExtra(header)
				if _, err := wbftengine.ApplyHeaderWBFTExtra(header, wbftengine.WritePrevSeals(extra.PrevRound, nil, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			nil,
		},
		{
			firstWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractWBFTExtra(header)
				if _, err := wbftengine.ApplyHeaderWBFTExtra(header, wbftengine.WritePrevSeals(extra.PrevRound, extra.PrevPreparedSeal, nil)); err != nil {
					return nil
				}
				return header
			},
			// for croissant block, invalid preparedSeal "does not" occurs when validating preparedSeal because header is  "not" changed
			nil,
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				wbftExtra, _ := types.ExtractWBFTExtra(header)
				wbftExtra.PreparedSeal = nil

				payload, err := rlp.EncodeToBytes(wbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			wbftcommon.ErrEmptyPreparedSeals,
		},
		{
			secondWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				wbftExtra, _ := types.ExtractWBFTExtra(header)
				wbftExtra.CommittedSeal = nil

				payload, err := rlp.EncodeToBytes(wbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			wbftcommon.ErrEmptyCommittedSeals,
		},
		{
			firstWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				wbftExtra, _ := types.ExtractWBFTExtra(header)
				wbftExtra.PreparedSeal = nil

				payload, err := rlp.EncodeToBytes(wbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			wbftcommon.ErrEmptyPreparedSeals,
		},
		{
			firstWbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				wbftExtra, _ := types.ExtractWBFTExtra(header)
				wbftExtra.CommittedSeal = nil

				payload, err := rlp.EncodeToBytes(wbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			wbftcommon.ErrEmptyCommittedSeals,
		},
	}

	for i, tc := range testCases {
		headerToTest := tc.headerManipulationFunc(tc.block)
		err := engine.VerifyHeader(chain, headerToTest)
		if !errors.Is(err, tc.expectedError) {
			t.Errorf("error mismatch for case %d: have %v, want %v", i, err, tc.expectedError)
		}
	}
}

func TestVerifyHeaderForSingleBlock(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()

	// wbftcommon.ErrEmptyPrevPreparedSeals case
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	block = updateWBFTBlock(block, engine.Address())
	err := engine.VerifyHeader(chain, block.Header())

	if err != wbftcommon.ErrEmptyPreparedSeals {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrEmptyPreparedSeals)
	}

	// short extra data
	header := block.Header()
	header.Extra = []byte{}
	err = engine.VerifyHeader(chain, header)
	if err != wbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrInvalidExtraDataFormat)
	}
	// incorrect extra format
	header.Extra = []byte("0000000000000000000000000000000012300000000000000000000000000000000000000000000000000000000000000000")
	err = engine.VerifyHeader(chain, header)
	if err != wbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrInvalidExtraDataFormat)
	}

	// invalid uncles hash
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.UncleHash = common.BytesToHash([]byte("123456789"))
	err = engine.VerifyHeader(chain, header)
	if err != wbftcommon.ErrInvalidUncleHash {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrInvalidUncleHash)
	}

	// invalid difficulty
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Difficulty = big.NewInt(2)
	err = engine.VerifyHeader(chain, header)
	if err != wbftcommon.ErrInvalidDifficulty {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrInvalidDifficulty)
	}

	// invalid timestamp
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = chain.Genesis().Time() + (engine.config.GetConfig(block.Number()).BlockPeriod - 1)
	err = engine.VerifyHeader(chain, header)
	if err != wbftcommon.ErrInvalidTimestamp {
		t.Errorf("error mismatch: have %v, want %v", err, wbftcommon.ErrInvalidTimestamp)
	}

	// future block
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = uint64(time.Now().Unix() + 10)
	err = engine.VerifyHeader(chain, header)
	if err != consensus.ErrFutureBlock {
		t.Errorf("error mismatch: have %v, want %v", err, consensus.ErrFutureBlock)
	}

	// future block which is within AllowedFutureBlockTime
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = new(big.Int).Add(big.NewInt(time.Now().Unix()), new(big.Int).SetUint64(10)).Uint64()
	priorValue := engine.config.AllowedFutureBlockTime
	engine.config.AllowedFutureBlockTime = 10
	err = engine.VerifyHeader(chain, header)
	engine.config.AllowedFutureBlockTime = priorValue //restore changed value
	if err == consensus.ErrFutureBlock {
		t.Errorf("error mismatch: have %v, want nil", err)
	}
}

func TestVerifyHeaders(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	genesis := chain.Genesis()

	// success case
	headers := []*types.Header{}
	blocks := []*types.Block{}
	size := 100

	for i := 0; i < size; i++ {
		var b *types.Block
		if i == 0 {
			b = makeBlockWithoutSeal(chain, engine, genesis)
			b = updateWBFTBlock(b, engine.Address())
		} else {
			b = makeBlockWithoutSeal(chain, engine, blocks[i-1])
			b = updateWBFTBlock(b, engine.Address())
		}
		blocks = append(blocks, b)
		headers = append(headers, blocks[i].Header())
	}
	// now = func() time.Time {
	// 	return time.Unix(int64(headers[size-1].Time), 0)
	// }
	_, results := engine.VerifyHeaders(chain, headers)
	const timeoutDura = 2 * time.Second
	timeout := time.NewTimer(timeoutDura)
	index := 0
OUT1:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != wbftcommon.ErrEmptyPrevPreparedSeals && err != wbftcommon.ErrEmptyPreparedSeals && err != wbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want wbftcommon.ErrEmptyCommittedSeals|wbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
					break OUT1
				}
			}
			index++
			if index == size {
				break OUT1
			}
		case <-timeout.C:
			break OUT1
		}
	}
	_, results = engine.VerifyHeaders(chain, headers)
	timeout = time.NewTimer(timeoutDura)
OUT2:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != wbftcommon.ErrEmptyPrevPreparedSeals && err != wbftcommon.ErrEmptyPreparedSeals && err != wbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want wbftcommon.ErrEmptyCommittedSeals|wbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
					break OUT2
				}
			}
		case <-timeout.C:
			break OUT2
		}
	}
	// error header cases
	headers[2].Number = big.NewInt(100)
	_, results = engine.VerifyHeaders(chain, headers)
	timeout = time.NewTimer(timeoutDura)
	index = 0
	errors := 0
	expectedErrors := 0
OUT3:
	for {
		select {
		case err := <-results:
			if err != nil {
				if err != wbftcommon.ErrEmptyPrevPreparedSeals && err != wbftcommon.ErrEmptyPreparedSeals && err != wbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					errors++
				}
			}
			index++
			if index == size {
				if errors != expectedErrors {
					t.Errorf("error mismatch: have %v, want %v", errors, expectedErrors)
				}
				break OUT3
			}
		case <-timeout.C:
			break OUT3
		}
	}
}

// Test that the next block has the correct previous prepared seals and previous committed seals.
// First, create a block with a seal and commit it.
// Then, when creating the next block, it should have the correct previous prepared seals and previous committed seals.
func TestPrevSeals(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	// Create an insert a new block into the chain.
	block := makeBlock(chain, engine, chain.Genesis())

	blockExtra, err := types.ExtractWBFTExtra(block.Header())
	if err != nil {
		t.Error(err.Error())
	}

	_, err = chain.InsertChain(types.Blocks{block})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	if err = engine.NewChainHead(); err != nil {
		t.Errorf("Error posting NewChainHead Event: %v", err)
	}

	nextBlock := makeBlockWithoutSeal(chain, engine, block)

	nextBlockExtra, err := types.ExtractWBFTExtra(nextBlock.Header())
	if err != nil {
		t.Error(err.Error())
	}

	if len(nextBlockExtra.PrevPreparedSeal.Sealers.GetSealers()) != 1 {
		t.Errorf("prev prepared seals mismatch: have %v, want 1", len(nextBlockExtra.PrevPreparedSeal.Sealers.GetSealers()))
	}

	if !bytes.Equal(blockExtra.PreparedSeal.Signature, nextBlockExtra.PrevPreparedSeal.Signature) {
		t.Errorf("prev prepared seals mismatch: have %v, want %v", nextBlockExtra.PrevPreparedSeal.Signature, blockExtra.PreparedSeal.Signature)
	}

	if len(nextBlockExtra.PrevCommittedSeal.Sealers.GetSealers()) != 1 {
		t.Errorf("committed seals mismatch: have %v, want 1", len(nextBlockExtra.PrevCommittedSeal.Sealers.GetSealers()))
	}

	if !bytes.Equal(blockExtra.CommittedSeal.Signature, nextBlockExtra.PrevCommittedSeal.Signature) {
		t.Errorf("committed seals mismatch: have %v, want %v", nextBlockExtra.PrevCommittedSeal.Signature, blockExtra.CommittedSeal.Signature)
	}
}

func postMsgEventToBackend(wbftEngine *Backend, message messages.WBFTMessage, payload []byte) error {
	if err := wbftEngine.istanbulEventMux.Post(wbft.MessageEvent{
		Code:    message.Code(),
		Payload: payload,
	}); err != nil {
		return err
	}
	return nil
}

func makeWBFTMessagePayload(message messages.WBFTMessage, node otherNode) ([]byte, error) {
	// set source of the message
	message.SetSource(node.address)
	encodedPayload, err := message.EncodePayloadForSigning()
	if err != nil {
		return nil, err
	}
	hashData := crypto.Keccak256(encodedPayload)
	signature, err := crypto.Sign(hashData, node.privateKey)
	if err != nil {
		return nil, err
	}
	message.SetSignature(signature)
	payload, err := rlp.EncodeToBytes(&message)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func nodeSendPrepareMsg(wbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	blsKey, err := bls.DeriveFromECDSA(node.privateKey)
	if err != nil {
		return err
	}
	prepareSeal := blsKey.Sign(wbftcore.PrepareSeal(targetBlock.Header(), uint32(round.Uint64()), wbftcore.SealTypePrepare)).Marshal()
	prepare := messages.NewPrepare(sequence, round, targetBlock.Hash(), prepareSeal)
	payload, err := makeWBFTMessagePayload(prepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(wbftEngine, prepare, payload)

	return nil
}

func nodeSendCommitMsg(wbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	blsKey, err := bls.DeriveFromECDSA(node.privateKey)
	if err != nil {
		return err
	}
	commitSeal := blsKey.Sign(wbftcore.PrepareSeal(targetBlock.Header(), uint32(round.Uint64()), wbftcore.SealTypeCommit)).Marshal()
	if err != nil {
		return err
	}
	commit := messages.NewCommit(sequence, round, targetBlock.Hash(), commitSeal)
	payload, err := makeWBFTMessagePayload(commit, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(wbftEngine, commit, payload)

	return nil
}

func TestAddingExtraSeals(t *testing.T) {
	// assume 3 nodes,
	// one is myself, two is normal node, three is slow node that sends extraSeals
	expectedAdditionalPreparedSealCnt := 3
	expectedAdditionalCommittedSealCnt := 3

	chain, engine, nodes := newBlockChain(3)
	normalNode := nodes[1]
	slowNode := nodes[2]

	parentBlock := chain.Genesis()
	eventSub := engine.EventMux().Subscribe(wbft.RequestEvent{})

	block := makeBlockWithoutSeal(chain, engine, parentBlock)
	currState, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	stopCh := make(chan struct{})

	go func() {
		engine.Seal(chain, block, resultCh, stopCh)
	}()
	ev := <-eventSub.Chan()
	request, ok := ev.Data.(wbft.RequestEvent)
	if !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}
	eventSub.Unsubscribe()
	proposedBlock, _ := request.Proposal.(*types.Block)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	executed := make(map[wbftcore.State]bool)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			<-ticker.C
			consensusState := engine.core.GetState()
			if executed[consensusState] {
				continue
			}

			switch consensusState {
			case wbftcore.StatePreprepared:
				if err := nodeSendPrepareMsg(engine, normalNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("normal node failed to send prepare msg. err :  %v", err)
				}
				executed[consensusState] = true
			case wbftcore.StatePrepared:
				// valid extra seal msg
				if err := nodeSendPrepareMsg(engine, slowNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("slow node failed to send prepare msg. err :  %v", err)
				}
				if err := nodeSendCommitMsg(engine, normalNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("normal node failed to send commit msg. err :  %v", err)
				}
				executed[consensusState] = true
			case wbftcore.StateCommitted:
				// valid extra seal msg
				if err := nodeSendCommitMsg(engine, slowNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("slow node failed to send commit msg. err :  %v", err)
				}
				executed[consensusState] = true
			case wbftcore.StateAcceptRequest:
				// valid extra seal msg
				if err := nodeSendCommitMsg(engine, normalNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("slow node failed to send prepare msg. err :  %v", err)
				}
				// invalid extra seal msg - wrong sequence
				if err := nodeSendPrepareMsg(engine, slowNode, new(big.Int).Add(proposedBlock.Number(), common.Big1), big.NewInt(1), proposedBlock); err != nil {
					t.Errorf("slow node failed to send commit msg. err :  %v", err)
				}
				// invalid extra seal msg - wrong round
				if err := nodeSendPrepareMsg(engine, slowNode, proposedBlock.Number(), big.NewInt(1), proposedBlock); err != nil {
					t.Errorf("slow node failed to send commit msg. err :  %v", err)
				}
				executed[consensusState] = true
				return
			}
		}
	}()

	select {
	case blockFromResultCh := <-resultCh:
		if _, err := chain.InsertChain(types.Blocks{blockFromResultCh}); err != nil {
			t.Errorf("failed to insert block1. err %v", err)
		}
		// give time for state Committed to be caught
		time.Sleep(time.Second)
		if err := engine.NewChainHead(); err != nil {
			t.Errorf("Error posting NewChainHead Event: %v", err)
		}
		// assume it's block time delay
		time.Sleep(time.Second)
	case <-stopCh:
		t.Errorf("engine stopped")
	}
	wg.Wait()

	extraPrepared, extraCommitted := engine.core.ProcessExtraSeal(proposedBlock, big.NewInt(0), engine.core.PriorValidators())
	if len(extraPrepared) != expectedAdditionalPreparedSealCnt {
		t.Errorf("unexpected prepared extra seal count. want %v, have %v", expectedAdditionalPreparedSealCnt, len(extraPrepared))
	}
	if len(extraCommitted) != expectedAdditionalCommittedSealCnt {
		t.Errorf("unexpected prepared extra seal count. want %v, have %v", expectedAdditionalCommittedSealCnt, len(extraCommitted))
	}
}

func setExtra(h *types.Header, wbftExtra *types.WBFTExtra) error {
	payload, err := rlp.EncodeToBytes(wbftExtra)
	if err != nil {
		return err
	}
	h.Extra = payload
	return nil
}

func TestReusePreparedSeal(t *testing.T) {
	chain, engine, _ := newBlockChain(1)
	defer engine.Stop()
	normalBlock := makeBlock(chain, engine, chain.Genesis())

	err := engine.VerifyHeader(chain, normalBlock.Header())
	if err != nil {
		t.Errorf("error mismatch: have %v, want nil", err)
	}

	prepareReused := normalBlock.Header()
	extra, _ := types.ExtractWBFTExtra(prepareReused)
	preparedSeal := extra.PreparedSeal
	extra.CommittedSeal = preparedSeal // reuse prepared seal for committed seal
	setExtra(prepareReused, extra)
	err = engine.VerifyHeader(chain, prepareReused)
	if err == nil {
		t.Errorf("prepared seal can be used for committed seal as well")
	} else if !errors.Is(err, wbftcommon.ErrInvalidCommittedSeals) {
		t.Errorf("error is not ErrInvalidCommittedSeals")
	}
}

/*
*
When verifying a proposal, since the block does not contain PrepareSeal and CommitSeal,
an error is returned from verifyCascadingFields. In VerifyBlockProposal, this error was implemented to be ignored.
However, the issue was that the subsequent checks in the code after the error was returned were not executed.
*/
func TestVerifyProposalBug(t *testing.T) {
	chain, engine, _ := newBlockChainWithCustom(1, func(config *wbft.Config) {
		config.BlockPeriod = 1
	})
	defer engine.Stop()

	firstBlock := makeBlock(chain, engine, chain.Genesis())
	_, err := chain.InsertChain(types.Blocks{firstBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	secondProposalBlock := makeBlockWithoutSeal(chain, engine, firstBlock)
	state, _ := chain.State()
	secondProposalBlock, _ = engine.FinalizeAndAssemble(chain, secondProposalBlock.Header(), state, nil, nil, nil, nil)

	invalidPrevCommittedSealBlockHeader := secondProposalBlock.Header()
	extra, _ := types.ExtractWBFTExtra(invalidPrevCommittedSealBlockHeader)
	extra.PrevCommittedSeal = extra.PrevPreparedSeal // invalid prevCommittedSeal
	setExtra(invalidPrevCommittedSealBlockHeader, extra)

	valSet, _ := engine.Engine().GetValidators(chain, firstBlock.Number(), firstBlock.ParentHash(), nil)
	invalidBlock := types.NewBlock(invalidPrevCommittedSealBlockHeader, nil, nil, nil, trie.NewStackTrie(nil))

	time.Sleep(time.Second) // wait for the block time
	_, err = engine.Engine().VerifyBlockProposal(chain, invalidBlock, valSet, valSet)
	if err == nil {
		t.Errorf("engine fails to verify a proposal which has invalid PrevCommittedSeal")
	} else if !errors.Is(err, wbftcommon.ErrInvalidPrevCommittedSeals) {
		t.Errorf("error is not ErrInvalidPrevCommittedSeals: %v", err)
	}
}

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
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	qbftcore "github.com/ethereum/go-ethereum/consensus/qbft/core"
	qbftengine "github.com/ethereum/go-ethereum/consensus/qbft/engine"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
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

func newBlockchainFromConfig(genesis *core.Genesis, nodeKeys []*ecdsa.PrivateKey, cfg *qbft.Config) (*core.BlockChain, *Backend, []otherNode) {
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
			backend.qbftEngine = qbftengine.NewEngine(backend.config, addr, backend.Sign)
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
	genesis, nodeKeys := testutils.GenesisAndKeys(n)

	config := new(qbft.Config)
	setConfigFromChainConfig(config, genesis.Config)

	return newBlockchainFromConfig(genesis, nodeKeys, config)
}

func newBlockChainWithCustom(n int, customizeConfig func(config *qbft.Config) *qbft.Config) (*core.BlockChain, *Backend, []otherNode) {
	genesis, nodeKeys := testutils.GenesisAndKeys(n)

	config := new(qbft.Config)
	setConfigFromChainConfig(config, genesis.Config)
	config = customizeConfig(config)

	return newBlockchainFromConfig(genesis, nodeKeys, config)
}

// this is a copy of ethconfig.SetConfigFromChainConfig; avoiding cyclic import
func setConfigFromChainConfig(qbftCfg *qbft.Config, config *params.ChainConfig) {
	if len(config.Transitions) > 0 {
		qbftCfg.Transitions = config.Transitions
	}
	if config.QBFT.BlockPeriodSeconds != 0 {
		qbftCfg.BlockPeriod = config.QBFT.BlockPeriodSeconds
	}
	if config.QBFT.RequestTimeoutSeconds != 0 {
		qbftCfg.RequestTimeout = config.QBFT.RequestTimeoutSeconds * 1000
	}
	if config.QBFT.EpochLength != 0 {
		qbftCfg.Epoch = config.QBFT.EpochLength
	}

	qbftCfg.ProposerPolicy = qbft.NewProposerPolicy(qbft.ProposerPolicyId(config.QBFT.ProposerPolicy))
	qbftCfg.BlockReward = config.QBFT.BlockReward
	qbftCfg.Validators = config.QBFT.Validators

	if config.QBFT.MaxRequestTimeoutSeconds != nil && *config.QBFT.MaxRequestTimeoutSeconds > 0 {
		qbftCfg.MaxRequestTimeoutSeconds = *config.QBFT.MaxRequestTimeoutSeconds
	}
}

// makeHeader create header executing no txs
func makeHeader(chainConfig *params.ChainConfig, engineConfig *qbft.Config, parent *types.Block) *types.Header {
	blockNumber := parent.Number().Add(parent.Number(), common.Big1)
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     blockNumber,
		GasLimit:   parent.GasLimit(),
		GasUsed:    0, // empty tx
		Time:       parent.Time() + engineConfig.GetConfig(blockNumber).BlockPeriod,
		Difficulty: types.QBFTDefaultDifficulty,
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
	header := makeHeader(chain.Config(), engine.config, parent)
	engine.Prepare(chain, header)
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}

func TestQBFTPrepare(t *testing.T) {
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
	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	eventLoop := func() {
		ev := <-eventSub.Chan()
		_, ok := ev.Data.(qbft.RequestEvent)
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
	expectedPreparedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedCommittedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)

	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	blockOutputChannel := make(chan *types.Block)
	stopChannel := make(chan struct{})

	if err := engine.Seal(chain, block, blockOutputChannel, stopChannel); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(qbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(otherBlock, [][]byte{expectedPreparedSeal}, [][]byte{expectedCommittedSeal}, big.NewInt(0)); err != nil {
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

func updateQBFTBlock(block *types.Block, addr common.Address) *types.Block {
	header := block.Header()
	header.Coinbase = addr
	return block.WithSeal(header)
}

func TestSealCommitted(t *testing.T) {
	chain, engine, _ := newBlockChain(2)
	defer engine.Stop()
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	expectedBlock := updateQBFTBlock(block, engine.Address())
	expectedPreparedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)
	expectedCommittedSeal := append([]byte{1, 2, 3}, bytes.Repeat([]byte{0x00}, types.IstanbulExtraSeal-3)...)

	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})
	defer eventSub.Unsubscribe()
	resultCh := make(chan *types.Block, 10)

	if err := engine.Seal(chain, block, resultCh, make(chan struct{})); err != nil {
		t.Error(err.Error())
	}

	ev := <-eventSub.Chan()
	if _, ok := ev.Data.(qbft.RequestEvent); !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}

	if err := engine.Commit(expectedBlock, [][]byte{expectedPreparedSeal}, [][]byte{expectedCommittedSeal}, big.NewInt(0)); err != nil {
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

	firstQbftBlock := makeBlock(chain, engine, chain.Genesis())
	_, err := chain.InsertChain(types.Blocks{firstQbftBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	secondQbftBlock := makeBlock(chain, engine, firstQbftBlock)
	_, err = chain.InsertChain(types.Blocks{secondQbftBlock})
	if err != nil {
		t.Errorf("Error inserting block: %v", err)
	}

	//create chain consists of genesisblock - montblanc hardfork block - regular qbft block
	testCases := []struct {
		block                  *types.Block
		headerManipulationFunc func(*types.Block) *types.Header
		expectedError          error
	}{
		{
			firstQbftBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header { return block.Header() },
			nil,
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractQBFTExtra(header)
				// write invalid round
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevSeals(extra.PrevRound+1, extra.PrevPreparedSeal, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			qbftcommon.ErrInvalidPreparedSeals, // PrevPreparedSeal changed -> block hash changed -> prepare seal invalid
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractQBFTExtra(header)
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevSeals(extra.PrevRound, [][]byte{}, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			qbftcommon.ErrInvalidPreparedSeals, // PrevPreparedSeal changed -> block hash changed -> prepare seal invalid
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractQBFTExtra(header)
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevSeals(extra.PrevRound, extra.PrevPreparedSeal, [][]byte{})); err != nil {
					return nil
				}
				return header
			},
			// for qbftBlock, invalid preparedSeal occurs when validating preparedSeal because header is changed and gets wrong signer address
			qbftcommon.ErrInvalidPreparedSeals,
		},
		{
			firstQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractQBFTExtra(header)
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevSeals(extra.PrevRound, [][]byte{}, extra.PrevCommittedSeal)); err != nil {
					return nil
				}
				return header
			},
			nil,
		},
		{
			firstQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				extra, _ := types.ExtractQBFTExtra(header)
				if err := qbftengine.ApplyHeaderQBFTExtra(header, qbftengine.WritePrevSeals(extra.PrevRound, extra.PrevPreparedSeal, [][]byte{})); err != nil {
					return nil
				}
				return header
			},
			// for montblanc block, invalid preparedSeal "does not" occurs when validating preparedSeal because header is  "not" changed
			nil,
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.PreparedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyPreparedSeals,
		},
		{
			secondQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.CommittedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyCommittedSeals,
		},
		{
			firstQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.PreparedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyPreparedSeals,
		},
		{
			firstQbftBlock,
			func(block *types.Block) *types.Header {
				header := block.Header()
				qbftExtra, _ := types.ExtractQBFTExtra(header)
				qbftExtra.CommittedSeal = [][]byte{}

				payload, err := rlp.EncodeToBytes(qbftExtra)
				if err != nil {
					return nil
				}
				header.Extra = payload
				return header
			},
			qbftcommon.ErrEmptyCommittedSeals,
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

	// qbftcommon.ErrEmptyPrevPreparedSeals case
	block := makeBlockWithoutSeal(chain, engine, chain.Genesis())
	block = updateQBFTBlock(block, engine.Address())
	err := engine.VerifyHeader(chain, block.Header())

	if err != qbftcommon.ErrEmptyPreparedSeals {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrEmptyPreparedSeals)
	}

	// short extra data
	header := block.Header()
	header.Extra = []byte{}
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidExtraDataFormat)
	}
	// incorrect extra format
	header.Extra = []byte("0000000000000000000000000000000012300000000000000000000000000000000000000000000000000000000000000000")
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidExtraDataFormat {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidExtraDataFormat)
	}

	// invalid uncles hash
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.UncleHash = common.BytesToHash([]byte("123456789"))
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidUncleHash {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidUncleHash)
	}

	// invalid difficulty
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Difficulty = big.NewInt(2)
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidDifficulty {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidDifficulty)
	}

	// invalid timestamp
	block = makeBlockWithoutSeal(chain, engine, chain.Genesis())
	header = block.Header()
	header.Time = chain.Genesis().Time() + (engine.config.GetConfig(block.Number()).BlockPeriod - 1)
	err = engine.VerifyHeader(chain, header)
	if err != qbftcommon.ErrInvalidTimestamp {
		t.Errorf("error mismatch: have %v, want %v", err, qbftcommon.ErrInvalidTimestamp)
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
			b = updateQBFTBlock(b, engine.Address())
		} else {
			b = makeBlockWithoutSeal(chain, engine, blocks[i-1])
			b = updateQBFTBlock(b, engine.Address())
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
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want qbftcommon.ErrEmptyCommittedSeals|qbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
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
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
					t.Errorf("error mismatch: have %v, want qbftcommon.ErrEmptyCommittedSeals|qbftcommon.ErrInvalidCommittedSeals|ErrUnknownAncestor", err)
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
				if err != qbftcommon.ErrEmptyPrevPreparedSeals && err != qbftcommon.ErrEmptyPreparedSeals && err != qbftcommon.ErrInvalidCommittedSeals && err != consensus.ErrUnknownAncestor {
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

	blockExtra, err := types.ExtractQBFTExtra(block.Header())
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

	nextBlockExtra, err := types.ExtractQBFTExtra(nextBlock.Header())
	if err != nil {
		t.Error(err.Error())
	}

	if len(nextBlockExtra.PrevPreparedSeal) != 1 {
		t.Errorf("prev prepared seals mismatch: have %v, want 1", len(nextBlockExtra.PrevPreparedSeal))
	}

	if !bytes.Equal(blockExtra.PreparedSeal[0], nextBlockExtra.PrevPreparedSeal[0]) {
		t.Errorf("prev prepared seals mismatch: have %v, want %v", nextBlockExtra.PrevPreparedSeal[0], blockExtra.PreparedSeal[0])
	}

	if len(nextBlockExtra.PrevCommittedSeal) != 1 {
		t.Errorf("committed seals mismatch: have %v, want 1", len(nextBlockExtra.PrevCommittedSeal))
	}

	if !bytes.Equal(blockExtra.CommittedSeal[0], nextBlockExtra.PrevCommittedSeal[0]) {
		t.Errorf("committed seals mismatch: have %v, want %v", nextBlockExtra.PrevCommittedSeal[0], blockExtra.CommittedSeal[0])
	}
}

func postMsgEventToBackend(qbftEngine *Backend, message messages.QBFTMessage, payload []byte) error {
	if err := qbftEngine.istanbulEventMux.Post(qbft.MessageEvent{
		Code:    message.Code(),
		Payload: payload,
	}); err != nil {
		return err
	}
	return nil
}

func makeQBFTMessagePayload(message messages.QBFTMessage, node otherNode) ([]byte, error) {
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

func nodeSendPreprepareMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	preprepare := messages.NewPreprepare(sequence, round, targetBlock)
	payload, err := makeQBFTMessagePayload(preprepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, preprepare, payload)

	return nil
}

func nodeSendPrepareMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	prepareSeal, err := crypto.Sign(qbftcore.PrepareSeal(targetBlock.Header(), uint32(round.Uint64()), qbftcore.SealTypePrepare), node.privateKey)
	if err != nil {
		return err
	}
	prepare := messages.NewPrepare(sequence, round, targetBlock.Hash(), prepareSeal)
	payload, err := makeQBFTMessagePayload(prepare, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, prepare, payload)

	return nil
}

func nodeSendCommitMsg(qbftEngine *Backend, node otherNode, sequence, round *big.Int, targetBlock *types.Block) error {
	commitSeal, err := crypto.Sign(qbftcore.PrepareSeal(targetBlock.Header(), uint32(round.Uint64()), qbftcore.SealTypeCommit), node.privateKey)
	if err != nil {
		return err
	}
	commit := messages.NewCommit(sequence, round, targetBlock.Hash(), commitSeal)
	payload, err := makeQBFTMessagePayload(commit, node)
	if err != nil {
		return err
	}
	go postMsgEventToBackend(qbftEngine, commit, payload)

	return nil
}

func makeBlockThroughConsensus(chain *core.BlockChain, engine *Backend, nodes []otherNode, parentBlock *types.Block) (*types.Block, error) {
	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})

	block := makeBlockWithoutSeal(chain, engine, parentBlock)
	currState, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	stopCh := make(chan struct{})
	go func() {
		engine.Seal(chain, block, resultCh, stopCh)
	}()

	ev := <-eventSub.Chan()
	request, ok := ev.Data.(qbft.RequestEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}
	eventSub.Unsubscribe()

	proposedBlock, _ := request.Proposal.(*types.Block)
	go func() error {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		executed := make(map[qbftcore.State]bool)
		for {
			<-ticker.C
			consensusState := engine.core.GetState()
			if executed[consensusState] {
				continue
			}

			switch consensusState {
			case qbftcore.StateAcceptRequest:
				for _, node := range nodes {
					if engine.getProposerForTest() == node.address {
						header := proposedBlock.Header()
						header.Coinbase = node.address
						statedb, err := chain.State()
						if err != nil {
							return fmt.Errorf("failed to get statedb err %v", err)
						}
						engine.Finalize(chain, header, statedb, nil, nil, nil)
						proposedBlock = proposedBlock.WithSeal(header)
						if err := nodeSendPreprepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
							return fmt.Errorf("failed to send preprepare msg. err :  %v", err)
						}
					}
				}
				executed[consensusState] = true
			case qbftcore.StatePreprepared:
				for _, node := range nodes {
					if err := nodeSendPrepareMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
						return fmt.Errorf("failed to send prepare msg. err :  %v", err)
					}
				}
				executed[consensusState] = true
			case qbftcore.StatePrepared:
				for _, node := range nodes {
					if err := nodeSendCommitMsg(engine, node, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
						return fmt.Errorf("failed to send commit msg. err :  %v", err)
					}
				}
				executed[consensusState] = true
			}
		}
	}()

	select {
	case blockFromResultCh := <-resultCh:
		return blockFromResultCh, nil
	case blockFromEnquequeCh := <-blockEnqueueChannel:
		stopCh <- struct{}{}
		return blockFromEnquequeCh, nil
	}
}

func TestMakingBlock(t *testing.T) {
	chain, engine, nodes := newBlockChain(4)
	others := nodes[1:]
	parentBlock := chain.Genesis()

	for i := 0; i < 3; i++ {
		finalBlock, err := makeBlockThroughConsensus(chain, engine, others, parentBlock)
		if err != nil {
			t.Errorf("failed to make block1 through consensus. err %v", err)
		}
		if parentBlock.Hash() != finalBlock.ParentHash() {
			t.Errorf("parent hash mismatch: have %v, want %v", finalBlock.ParentHash(), parentBlock.Hash())
		}

		if _, err := chain.InsertChain(types.Blocks{finalBlock}); err != nil {
			t.Errorf("failed to insert block1. err %v", err)
		}

		// check if generated final block is included in chain properly
		block := chain.GetBlockByHash(finalBlock.Hash())
		if block == nil {
			t.Errorf("block number %v is not generated correctly", finalBlock.Hash())
		}
		parentBlock = finalBlock
	}
}

func TestAddingExtraSeals(t *testing.T) {
	// assume 3 nodes,
	// one is myself, two is normal node, three is slow node that sends extraSeals
	expectedAdditionalPreparedSealCnt := 1
	expectedAdditionalCommittedSealCnt := 2

	chain, engine, nodes := newBlockChain(3)
	normalNode := nodes[1]
	slowNode := nodes[2]

	parentBlock := chain.Genesis()
	eventSub := engine.EventMux().Subscribe(qbft.RequestEvent{})

	block := makeBlockWithoutSeal(chain, engine, parentBlock)
	currState, _ := chain.State()
	block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
	resultCh := make(chan *types.Block, 10)
	stopCh := make(chan struct{})

	go func() {
		engine.Seal(chain, block, resultCh, stopCh)
	}()
	ev := <-eventSub.Chan()
	request, ok := ev.Data.(qbft.RequestEvent)
	if !ok {
		t.Errorf("unexpected event comes: %v", reflect.TypeOf(ev.Data))
	}
	eventSub.Unsubscribe()
	proposedBlock, _ := request.Proposal.(*types.Block)

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	executed := make(map[qbftcore.State]bool)

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
			case qbftcore.StatePreprepared:
				if err := nodeSendPrepareMsg(engine, normalNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("normal node failed to send prepare msg. err :  %v", err)
				}
				executed[consensusState] = true
			case qbftcore.StatePrepared:
				// valid extra seal msg
				if err := nodeSendPrepareMsg(engine, slowNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("slow node failed to send prepare msg. err :  %v", err)
				}
				if err := nodeSendCommitMsg(engine, normalNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("normal node failed to send commit msg. err :  %v", err)
				}
				executed[consensusState] = true
			case qbftcore.StateCommitted:
				// valid extra seal msg
				if err := nodeSendCommitMsg(engine, slowNode, proposedBlock.Number(), big.NewInt(0), proposedBlock); err != nil {
					t.Errorf("slow node failed to send commit msg. err :  %v", err)
				}
				executed[consensusState] = true
			case qbftcore.StateAcceptRequest:
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

	extraPrepared, extraCommitted := engine.core.ProcessExtraSeal(proposedBlock, big.NewInt(0))
	if len(extraPrepared) != expectedAdditionalPreparedSealCnt {
		t.Errorf("unexpected prepared extra seal count. want %v, have %v", expectedAdditionalPreparedSealCnt, len(extraPrepared))
	}
	if len(extraCommitted) != expectedAdditionalCommittedSealCnt {
		t.Errorf("unexpected prepared extra seal count. want %v, have %v", expectedAdditionalCommittedSealCnt, len(extraCommitted))
	}
}

func TestLackingSealsFromPropagatedBlock(t *testing.T) {
	chain, engine, nodes := newBlockChain(4)
	others := nodes[1:]

	// 1. generate block through consensus
	validBlock, err := makeBlockThroughConsensus(chain, engine, others, chain.Genesis())
	if err != nil {
		t.Errorf("failed to make valid block through consensus. err %v", err)
	}

	// 2. remove some preparedSeals from valid block ( len(preparedSeal) < 2F+1 ) and make malformed block
	header := validBlock.Header()
	qbftExtra, _ := types.ExtractQBFTExtra(header)

	qbftExtra.PreparedSeal = qbftExtra.PreparedSeal[2:]
	payload, err := rlp.EncodeToBytes(qbftExtra)
	if err != nil {
		t.Errorf("failed to encode qbftExtra. err %v", err)
	}
	header.Extra = payload
	malformedBlock := validBlock.WithSeal(header)

	// 3. test the case when malformed block with lack of preparedSeal is propagated
	_, err = chain.InsertChain(types.Blocks{malformedBlock})
	if !errors.Is(err, qbftcommon.ErrInvalidPreparedSeals) {
		t.Errorf("unexpected error. expect %v, got %v", qbftcommon.ErrInvalidPreparedSeals, err)
	}
}

func setExtra(h *types.Header, qbftExtra *types.QBFTExtra) error {
	payload, err := rlp.EncodeToBytes(qbftExtra)
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
	extra, _ := types.ExtractQBFTExtra(prepareReused)
	preparedSeal := extra.PreparedSeal
	extra.CommittedSeal = preparedSeal // reuse prepared seal for committed seal
	setExtra(prepareReused, extra)
	err = engine.VerifyHeader(chain, prepareReused)
	if err == nil {
		t.Errorf("prepared seal can be used for committed seal as well")
	} else if !errors.Is(err, qbftcommon.ErrInvalidCommittedSeals) {
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
	chain, engine, _ := newBlockChainWithCustom(1, func(config *qbft.Config) *qbft.Config {
		config.BlockPeriod = 1
		return config
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
	extra, _ := types.ExtractQBFTExtra(invalidPrevCommittedSealBlockHeader)
	extra.PrevCommittedSeal = extra.PrevPreparedSeal // invalid prevCommittedSeal
	setExtra(invalidPrevCommittedSealBlockHeader, extra)

	valSet, _ := engine.Engine().GetValidators(chain, firstBlock.Number(), firstBlock.ParentHash(), nil)
	invalidBlock := types.NewBlock(invalidPrevCommittedSealBlockHeader, nil, nil, nil, trie.NewStackTrie(nil))

	time.Sleep(time.Second) // wait for the block time
	_, err = engine.Engine().VerifyBlockProposal(chain, invalidBlock, valSet, valSet)
	if err == nil {
		t.Errorf("engine fails to verify a proposal which has invalid PrevCommittedSeal")
	} else if !errors.Is(err, qbftcommon.ErrInvalidPrevCommittedSeals) {
		t.Errorf("error is not ErrInvalidPrevCommittedSeals: %v", err)
	}
}

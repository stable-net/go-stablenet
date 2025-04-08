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
// This file is derived from quorum/consensus/istanbul/backend/backend.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"crypto/ecdsa"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	qbftcore "github.com/ethereum/go-ethereum/consensus/qbft/core"
	qbftengine "github.com/ethereum/go-ethereum/consensus/qbft/engine"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// ## Wemix QBFT START
// 1. related code to istanbul engine is erased
// 2. package "github.com/hashicorp/golang-lru" is replaced to  "github.com/ethereum/go-ethereum/common/lru"
// ## Wemix QBFT END

const (
	// fetcherID is the ID indicates the block is from Istanbul engine
	fetcherID = "istanbul"
)

type SimApplier interface {
	Apply(chainConfig *params.ChainConfig, config *qbft.Config, blockNum *big.Int)
}

// New creates an Ethereum backend for Istanbul core engine.
func New(config *qbft.Config, privateKey *ecdsa.PrivateKey, db ethdb.Database) *Backend {
	recentMessages := lru.NewCache[common.Address, *lru.Cache[common.Hash, bool]](inmemoryPeers)
	knownMessages := lru.NewCache[common.Hash, bool](inmemoryMessages)

	blsSecretKey, err := bls.DeriveFromECDSA(privateKey)
	if err != nil {
		return nil
	}

	sb := &Backend{
		config:           config,
		istanbulEventMux: new(event.TypeMux),
		privateKey:       privateKey,
		blsSecretKey:     blsSecretKey,
		address:          crypto.PubkeyToAddress(privateKey.PublicKey),
		logger:           log.New(),
		db:               db,
		commitCh:         make(chan *types.Block, 1),
		coreStarted:      false,
		recentMessages:   recentMessages,
		knownMessages:    knownMessages,
	}

	sb.qbftEngine = qbftengine.NewEngine(sb.config, sb.address, sb.Sign)
	return sb
}

// ----------------------------------------------------------------------------

type Backend struct {
	config *qbft.Config

	blsSecretKey bls.SecretKey
	privateKey   *ecdsa.PrivateKey
	address      common.Address

	core *qbftcore.Core

	qbftEngine *qbftengine.Engine

	istanbulEventMux *event.TypeMux

	logger log.Logger

	db ethdb.Database

	chain        consensus.ChainHeaderReader
	currentBlock func() *types.Block
	hasBadBlock  func(db ethdb.Reader, hash common.Hash) bool

	// the channels for qbft engine notifications
	commitCh          chan *types.Block
	proposedBlockHash common.Hash
	sealMu            sync.Mutex
	coreStarted       bool
	coreMu            sync.RWMutex

	// event subscription for ChainHeadEvent event
	broadcaster consensus.Broadcaster

	recentMessages *lru.Cache[common.Address, *lru.Cache[common.Hash, bool]] // the cache of peer's messages
	knownMessages  *lru.Cache[common.Hash, bool]                             // the cache of self messages

	simApplier SimApplier

	notifyNewRound func(isProposer bool, waitTime time.Duration, round *big.Int)
}

func (sb *Backend) InjectSimApplier(applier SimApplier) {
	sb.simApplier = applier
}

func (sb *Backend) IsRunning() bool {
	return sb.coreStarted
}

func (sb *Backend) Engine() *qbftengine.Engine {
	return sb.qbftEngine // ## Wemix QBFT : currently return only qbft engine
}

// zekun: HACK
func (sb *Backend) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return sb.Engine().CalcDifficulty(chain, time, parent)
}

// Address implements qbft.Backend.Address
func (sb *Backend) Address() common.Address {
	return sb.Engine().Address()
}

// Broadcast implements qbft.Backend.Broadcast
func (sb *Backend) Broadcast(valSet qbft.ValidatorSet, code uint64, payload []byte) error {
	_, validator := valSet.GetByAddress(sb.address)
	if validator == nil {
		return qbft.ErrUnauthorizedAddress
	}

	// send to others
	sb.Gossip(valSet, code, payload)
	// send to self
	msg := qbft.MessageEvent{
		Code:    code,
		Payload: payload,
	}
	go sb.istanbulEventMux.Post(msg)
	return nil
}

// Gossip implements qbft.Backend.Gossip
func (sb *Backend) Gossip(valSet qbft.ValidatorSet, code uint64, payload []byte) error {
	hash := qbft.RLPHash(payload)
	sb.knownMessages.Add(hash, true)

	targets := make(map[common.Address]bool)
	for _, val := range valSet.List() {
		if val.Address() != sb.Address() {
			targets[val.Address()] = true
		}
	}
	if sb.broadcaster != nil && len(targets) > 0 {
		ps := sb.broadcaster.FindPeers(targets)
		for addr, p := range ps {
			m, ok := sb.recentMessages.Get(addr)
			if ok {
				if _, ok := m.Get(hash); ok {
					// This peer had this event, skip it
					continue
				}
			} else {
				m = lru.NewCache[common.Hash, bool](inmemoryMessages)
			}

			m.Add(hash, true)
			sb.recentMessages.Add(addr, m)

			var outboundCode uint64 = istanbulMsg
			if _, ok := qbftmessage.MessageCodes()[code]; ok {
				outboundCode = code
			}
			go p.SendQBFTConsensus(outboundCode, payload)
		}
	}
	return nil
}

// Commit implements qbft.Backend.Commit
func (sb *Backend) Commit(proposal qbft.Proposal, preparedSeals, committedSeals []qbft.SealData, round *big.Int) (err error) {
	// Check if the proposal is a valid block
	block, ok := proposal.(*types.Block)
	if !ok {
		sb.logger.Error("BFT: invalid block proposal", "proposal", proposal)
		return qbftcommon.ErrInvalidProposal
	}

	// Commit header
	h := block.Header()
	err = sb.Engine().CommitHeader(h, preparedSeals, committedSeals, round)
	if err != nil {
		return
	}

	// update block's header
	block = block.WithSeal(h)

	sb.logger.Info("BFT: block proposal committed", "author", sb.Address(), "hash", proposal.Hash(), "number", proposal.Number().Uint64())

	// - if the proposed and committed blocks are the same, send the proposed hash
	//   to commit channel, which is being watched inside the engine.Seal() function.
	// - otherwise, we try to insert the block.
	// -- if success, the ChainHeadEvent event will be broadcasted, try to build
	//    the next block and the previous Seal() will be stopped.
	// -- otherwise, a error will be returned and a round change event will be fired.
	if sb.proposedBlockHash == block.Hash() {
		// feed block hash to Seal() and wait the Seal() result
		sb.commitCh <- block
		return nil
	}

	if sb.broadcaster != nil {
		sb.broadcaster.Enqueue(fetcherID, block)
	}

	return nil
}

// EventMux implements qbft.Backend.EventMux
func (sb *Backend) EventMux() *event.TypeMux {
	return sb.istanbulEventMux
}

// Verify implements qbft.Backend.Verify
func (sb *Backend) Verify(proposal qbft.Proposal) (time.Duration, error) {
	// Check if the proposal is a valid block
	block, ok := proposal.(*types.Block)
	if !ok {
		sb.logger.Error("BFT: invalid block proposal", "proposal", proposal)
		return 0, qbftcommon.ErrInvalidProposal
	}

	// check bad block
	if sb.HasBadProposal(block.Hash()) {
		sb.logger.Warn("BFT: bad block proposal", "proposal", proposal)
		return 0, qbftcommon.ErrBlacklistedHash
	}

	header := block.Header()
	valSet, prevValSet, err := sb.GetValidatorsForVerifying(sb.chain, header, nil)
	if err != nil {
		return 0, err
	}
	return sb.Engine().VerifyBlockProposal(sb.chain, block, valSet, prevValSet)
}

// Sign implements qbft.Backend.Sign
func (sb *Backend) Sign(data []byte) ([]byte, error) {
	hashData := crypto.Keccak256(data)
	return crypto.Sign(hashData, sb.privateKey)
}

// SignWithoutHashing implements qbft.Backend.SignWithoutHashing and signs input data with the backend's bls secret key without hashing the input data
func (sb *Backend) SignWithoutHashing(data []byte) []byte {
	return sb.blsSecretKey.Sign(data).Marshal()
}

// CheckSignature implements qbft.Backend.CheckSignature
func (sb *Backend) CheckSignature(data []byte, address common.Address, sig []byte) error {
	signer, err := qbft.GetSignatureAddress(data, sig)
	if err != nil {
		return err
	}
	// Compare derived addresses
	if signer != address {
		return qbftcommon.ErrInvalidSignature
	}

	return nil
}

// HasPropsal implements qbft.Backend.HashBlock
func (sb *Backend) HasPropsal(hash common.Hash, number *big.Int) bool {
	return sb.chain.GetHeader(hash, number.Uint64()) != nil
}

// GetProposer implements qbft.Backend.GetProposer
func (sb *Backend) GetProposer(number uint64) common.Address {
	if h := sb.chain.GetHeaderByNumber(number); h != nil {
		a, _ := sb.Author(h)
		return a
	}
	return common.Address{}
}

func (sb *Backend) Validators(proposal qbft.Proposal) qbft.ValidatorSet {
	valSet, err := sb.Engine().GetValidators(sb.chain, new(big.Int).Add(proposal.Number(), common.Big1), proposal.Hash(), nil)
	if err != nil {
		return validator.NewSet(nil, nil, sb.config.ProposerPolicy)
	}
	return valSet
}

func (sb *Backend) getProposerForTest() common.Address {
	// used only for testing
	// if you want actual using, you should lock coreMu for it
	return sb.core.GetProposer()
}

func (sb *Backend) LastProposal() (qbft.Proposal, common.Address) {
	block := sb.currentBlock()

	var proposer common.Address
	if block.Number().Cmp(common.Big0) > 0 {
		var err error
		proposer, err = sb.Author(block.Header())
		if err != nil {
			sb.logger.Error("BFT: last block proposal invalid", "err", err)
			return nil, common.Address{}
		}
	}

	// Return header only block here since we don't need block body
	return block, proposer
}

func (sb *Backend) HasBadProposal(hash common.Hash) bool {
	if sb.hasBadBlock == nil {
		return false
	}
	return sb.hasBadBlock(sb.db, hash)
}

func (sb *Backend) Close() error {
	return nil
}

func (sb *Backend) startQBFT() error {
	sb.logger.Info("BFT: activate QBFT")
	sb.logger.Trace("BFT: set ProposerPolicy sorter to ValidatorSortByByteFunc")
	sb.config.ProposerPolicy.Use(qbft.ValidatorSortByByte())

	sb.core = qbftcore.New(sb, sb.config)
	if err := sb.core.Start(); err != nil {
		sb.logger.Error("BFT: failed to activate QBFT", "err", err)
		return err
	}

	return nil
}

func (sb *Backend) stop() error {
	core := sb.core
	sb.core = nil

	if core != nil {
		sb.logger.Info("BFT: deactivate")
		if err := core.Stop(); err != nil {
			sb.logger.Error("BFT: failed to deactivate", "err", err)
			return err
		}
	}

	return nil
}

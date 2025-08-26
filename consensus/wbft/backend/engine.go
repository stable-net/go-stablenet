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
// This file is derived from quorum/consensus/istanbul/backend/engine.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	wbftengine "github.com/ethereum/go-ethereum/consensus/wbft/engine"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	inmemoryPeers    = 40
	inmemoryMessages = 1024
)

// 1. code related to ibft engine is erased
// 2. fix interface function to get fit with consensus.Engine interface

// Author retrieves the Ethereum address of the account that minted the given
// block, which may be different from the header's coinbase if a consensus
// engine is based on signatures.
func (sb *Backend) Author(header *types.Header) (common.Address, error) {
	return sb.Engine().Author(header)
}

// PrepareSigners extracts all the addresses who have signed the given header
// during the prepare phase. It will extract for each seal who signed it,
// regardless of if the seal is repeated
func (sb *Backend) PrepareSigners(chain consensus.ChainHeaderReader, header *types.Header) ([]common.Address, error) {
	return sb.Engine().PrepareSigners(chain, header)
}

// CommitSigners extracts all the addresses who have signed the given header
// during the commit phase. It will extract for each seal who signed it,
// regardless of if the seal is repeated
func (sb *Backend) CommitSigners(chain consensus.ChainHeaderReader, header *types.Header) ([]common.Address, error) {
	return sb.Engine().CommitSigners(chain, header)
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (sb *Backend) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return sb.verifyHeader(chain, header, nil)
}

func (sb *Backend) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	valSet, prevValSet, err := sb.GetValidatorsForVerifying(chain, header, parents)
	if err != nil {
		return err
	}
	return sb.Engine().VerifyHeader(chain, header, parents, valSet, prevValSet, true)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (sb *Backend) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))
	go func() {
		errored := false
		for i, header := range headers {
			var err error
			if errored {
				err = consensus.ErrUnknownAncestor
			} else {
				err = sb.verifyHeader(chain, header, headers[:i])
			}

			if err != nil {
				errored = true
			}

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (sb *Backend) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	return sb.Engine().VerifyUncles(chain, block)
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (sb *Backend) VerifySeal(chain consensus.ChainHeaderReader, header *types.Header) error {
	// get parent header and ensure the signer is in parent's validator set
	number := header.Number.Uint64()
	if number == 0 {
		return wbftcommon.ErrUnknownBlock
	}

	// Assemble the voting snapshot
	valSet, err := sb.Engine().GetValidators(chain, header.Number, header.ParentHash, nil)
	if err != nil {
		return err
	}

	return sb.Engine().VerifySeal(chain, header, valSet)
}

// timeForNextWork returns the time to wait for next work namely next block time
func (sb *Backend) timeForNextWork() uint64 {
	if sb.currentBlock == nil {
		return 0 // if it has no current block function then it returns zero time
	}
	latestBlock := sb.currentBlock()
	if latestBlock == nil {
		return 0
	}
	next := new(big.Int).Set(latestBlock.Number())
	next = next.Add(next, big.NewInt(1))
	return latestBlock.Time() + sb.Engine().PeriodToNextBlock(next)
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (sb *Backend) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	valSet, err := sb.Engine().GetValidators(chain, header.Number, header.ParentHash, nil)
	if err != nil {
		return err
	}

	if sb.simApplier != nil {
		sb.simApplier.Apply(chain.Config(), sb.config, header.Number)
	}

	extraPreparedSeal, extraCommittedSeal := sb.processExtraSeals()

	err = sb.Engine().Prepare(chain, header, valSet, extraPreparedSeal, extraCommittedSeal)
	if err != nil {
		return err
	}

	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
//
// Note, the block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (sb *Backend) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, _ []*types.Withdrawal) error {
	return sb.Engine().Finalize(chain, header, state, txs, uncles)
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (sb *Backend) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt, _ []*types.Withdrawal) (*types.Block, error) {
	return sb.Engine().FinalizeAndAssemble(chain, header, state, txs, uncles, receipts)
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (sb *Backend) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	go func() {
		// get the proposed block hash and clear it if the seal() is completed.
		sb.sealMu.Lock()
		sb.proposedBlockHash = block.Hash()

		defer func() {
			sb.proposedBlockHash = common.Hash{}
			sb.sealMu.Unlock()
		}()

		// post block into Istanbul engine
		go sb.EventMux().Post(wbft.RequestEvent{
			Proposal: block,
		})
		for {
			select {
			case sealed := <-sb.commitCh:
				// if the block hash and the hash from channel are the same,
				// return the result. Otherwise, keep waiting the next hash.
				if sealed != nil {
					if block.Hash() == sealed.Hash() {
						results <- sealed
						return
					}
				}
			case <-stop:
				results <- nil
				return
			}
		}
	}()
	return nil
}

func (sb *Backend) processExtraSeals() ([]wbft.SealData, []wbft.SealData) {
	sb.coreMu.RLock()
	defer sb.coreMu.RUnlock()
	if sb.core == nil {
		sb.logger.Warn("WBFT: fail to process extra seals due to nil core")
		return nil, nil
	} else {
		return sb.core.ProcessExtraSeal(sb.currentBlock(), sb.core.PriorRound(), sb.core.PriorValidators())
	}
}

// APIs returns the RPC APIs this consensus engine provides.
func (sb *Backend) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{{
		Namespace: "istanbul",
		Version:   "1.0",
		Service:   &API{chain: chain, backend: sb},
		Public:    true,
	}}
}

// Start implements consensus.Istanbul.Start
func (sb *Backend) Start(
	chain consensus.ChainHeaderReader,
	currentBlock func() *types.Block,
	hasBadBlock func(db ethdb.Reader, hash common.Hash) bool,
	notifyNewRound func(waitTime time.Duration, round *big.Int)) error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if sb.coreStarted {
		return wbft.ErrStartedEngine
	}

	// clear previous data
	sb.proposedBlockHash = common.Hash{}
	if sb.commitCh != nil {
		close(sb.commitCh)
	}
	sb.commitCh = make(chan *types.Block, 1)

	sb.chain = chain
	sb.currentBlock = currentBlock
	sb.hasBadBlock = hasBadBlock
	sb.notifyNewRound = notifyNewRound

	log.Info("start WBFT")
	err := sb.startWBFT()

	if err != nil {
		return err
	}

	sb.coreStarted = true

	return nil
}

func (sb *Backend) NotifyNewRound(round *big.Int) {
	if sb.notifyNewRound != nil {
		waitDuration := time.Duration(0)
		if round.Uint64() == 0 {
			waitDuration = time.Until(time.Unix(int64(sb.timeForNextWork()), 0))
		}
		sb.notifyNewRound(waitDuration, round)
	}
}

// Stop implements consensus.Istanbul.Stop
func (sb *Backend) Stop() error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if !sb.coreStarted {
		return wbft.ErrStoppedEngine
	}
	if err := sb.stop(); err != nil {
		return err
	}
	sb.coreStarted = false

	return nil
}

// CallEngineSpecific implements consensus.Engine
func (sb *Backend) CallEngineSpecific(method string, args ...interface{}) interface{} {
	switch method {
	case "Start":
		if len(args) != 4 {
			return wbftcommon.ErrInvalidSpecificCall
		}
		chain, ok := args[0].(consensus.ChainHeaderReader)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		currentBlock, ok := args[1].(func() *types.Block)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		hasBadBlock, ok := args[2].(func(db ethdb.Reader, hash common.Hash) bool)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		notifyNewRound, ok := args[3].(func(waitTime time.Duration, round *big.Int))
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		if sb.coreStarted {
			_ = sb.Stop()
		}
		return sb.Start(chain, currentBlock, hasBadBlock, notifyNewRound)
	case "InheritExtra":
		if len(args) != 2 {
			return wbftcommon.ErrInvalidSpecificCall
		}
		parent, ok := args[0].(*types.Header)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		header, ok := args[1].(*types.Header)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		extra, err := types.ExtractWBFTExtra(parent)
		if err != nil {
			return err
		}

		if parent.Number.Sign() > 0 {
			if extra.PreparedSeal == nil {
				return wbftcommon.ErrEmptyPreparedSeals
			} else if extra.CommittedSeal == nil {
				// TODO : what if there is not committedSeal that node collected?
				return wbftcommon.ErrEmptyCommittedSeals
			}
		}

		prevPreparedSeal := extra.PreparedSeal
		prevCommittedSeal := extra.CommittedSeal
		// add lastBlock committers to extraData's prevCommittedSeal section
		// validators are stored in genesis block
		wbftengine.ApplyHeaderWBFTExtra(
			header,
			sb.Engine().WriteRandao(sb.chain.Config(), header),
			wbftengine.WritePrevSeals(extra.Round, prevPreparedSeal, prevCommittedSeal))
		return nil

	case "SetMixDigest":
		if len(args) != 2 {
			return wbftcommon.ErrInvalidSpecificCall
		}
		parent, ok := args[0].(*types.Header)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		header, ok := args[1].(*types.Header)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		extra, err := types.ExtractWBFTExtra(header)
		if err != nil {
			return err
		}

		header.MixDigest = wbftengine.CalculateRandaoMix(parent.MixDigest, extra.RandaoReveal)
		return nil

	case "NewChainHead":
		return sb.NewChainHead()

	case "SetCoinbase":
		header, ok := args[0].(*types.Header)
		if !ok {
			return wbftcommon.ErrInvalidSpecificCall
		}
		header.Coinbase = sb.Engine().Address()
		return nil

	default:
		return wbftcommon.ErrInvalidSpecificCall
	}
}

// SealHash returns the hash of a block prior to it being sealed.
func (sb *Backend) SealHash(header *types.Header) common.Hash {
	return sb.Engine().SealHash(header)
}

func (sb *Backend) GetValidatorsForVerifying(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) (wbft.ValidatorSet, wbft.ValidatorSet, error) {
	var valSet, prevValSet wbft.ValidatorSet
	var err error

	// Retrieve the ValidatorSet for the block height
	if valSet, err = sb.Engine().GetValidators(chain, header.Number, header.ParentHash, parents); err != nil {
		return nil, nil, consensus.ErrUnknownAncestor
	}

	if header.Number.Uint64() >= chain.Config().CroissantBlock.Uint64()+2 {
		var parent *types.Header
		if len(parents) == 0 {
			parent = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
			parents = nil
		} else {
			parent = parents[len(parents)-1]
			parents = parents[:len(parents)-1]
		}
		if parent == nil {
			return nil, nil, consensus.ErrUnknownAncestor
		}
		// Retrieve the ValidatorSet of the previous block
		if prevValSet, err = sb.Engine().GetValidators(chain, parent.Number, parent.ParentHash, parents); err != nil {
			return nil, nil, err
		}
	} else {
		prevValSet = valSet
	}

	return valSet, prevValSet, nil
}

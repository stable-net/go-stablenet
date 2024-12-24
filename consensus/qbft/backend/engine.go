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
// This file is derived from quorum/consensus/istanbul/backend/engine.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	qbftengine "github.com/ethereum/go-ethereum/consensus/qbft/engine"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	checkpointInterval = 1024 // Number of blocks after which to save the vote snapshot to the database
	inmemorySnapshots  = 128  // Number of recent vote snapshots to keep in memory
	inmemoryPeers      = 40
	inmemoryMessages   = 1024
)

// ## Wemix QBFT START
// 1. code related to ibft engine is erased
// 2. fix interface function to get fit with consensus.Engine interface
// ## Wemix QBFT END

// Author retrieves the Ethereum address of the account that minted the given
// block, which may be different from the header's coinbase if a consensus
// engine is based on signatures.
func (sb *Backend) Author(header *types.Header) (common.Address, error) {
	return sb.Engine().Author(header)
}

// PrepareSigners extracts all the addresses who have signed the given header
// during the prepare phase. It will extract for each seal who signed it,
// regardless of if the seal is repeated
func (sb *Backend) PrepareSigners(header *types.Header) ([]common.Address, error) {
	return sb.Engine().PrepareSigners(header)
}

// CommitSigners extracts all the addresses who have signed the given header
// during the commit phase. It will extract for each seal who signed it,
// regardless of if the seal is repeated
func (sb *Backend) CommitSigners(header *types.Header) ([]common.Address, error) {
	return sb.Engine().CommitSigners(header)
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (sb *Backend) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	return sb.verifyHeader(chain, header, nil)
}

func (sb *Backend) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) error {
	// Assemble the voting snapshot
	var snap, prevSnap *Snapshot
	var err error

	if snap, err = sb.snapshot(chain, header.Number.Uint64()-1, header.ParentHash, parents); err != nil {
		return err
	} else if header.Number.Uint64() < 2 {
		return sb.Engine().VerifyHeader(chain, header, parents, snap.ValSet, snap.ValSet, true)
	} else if len(parents) < 2 {
		var parent *types.Header
		if len(parents) == 1 {
			parent = parents[0]
		} else {
			parent = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
			if parent == nil {
				return consensus.ErrUnknownAncestor
			}
		}
		if prevSnap, err = sb.snapshot(chain, parent.Number.Uint64()-1, parent.ParentHash, nil); err != nil {
			return err
		}
	} else {
		h := parents[len(parents)-2]
		if h.Number.Uint64() != header.Number.Uint64()-2 {
			return errors.New("unexpected parents block")
		}
		if prevSnap, err = sb.snapshot(chain, h.Number.Uint64(), h.Hash(), parents[:len(parents)-1]); err != nil {
			return err
		}
	}

	return sb.Engine().VerifyHeader(chain, header, parents, snap.ValSet, prevSnap.ValSet, true)
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
		return qbftcommon.ErrUnknownBlock
	}

	// Assemble the voting snapshot
	snap, err := sb.snapshot(chain, number-1, header.ParentHash, nil)
	if err != nil {
		return err
	}

	return sb.Engine().VerifySeal(chain, header, snap.ValSet)
}

// TimeForNextWork returns the time to wait for next work namely next block time
func (sb *Backend) TimeForNextWork() uint64 {
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
	// Assemble the voting snapshot
	snap, err := sb.snapshot(chain, header.Number.Uint64()-1, header.ParentHash, nil)
	if err != nil {
		return err
	}

	if sb.simApplier != nil {
		sb.simApplier.Apply(chain.Config(), sb.config, header.Number)
	}

	extraPreparedSeal, extraCommittedSeal := sb.processExtraSeals()

	err = sb.Engine().Prepare(chain, header, snap.ValSet, extraPreparedSeal, extraCommittedSeal)
	if err != nil {
		return err
	}

	// get valid candidate list
	sb.candidatesLock.RLock()
	var addresses []common.Address
	var authorizes []bool
	for address, authorize := range sb.candidates {
		if snap.checkVote(address, authorize) {
			addresses = append(addresses, address)
			authorizes = append(authorizes, authorize)
		}
	}
	sb.candidatesLock.RUnlock()

	if len(addresses) > 0 {
		index := rand.Intn(len(addresses))

		err = sb.Engine().WriteVote(header, addresses[index], authorizes[index])
		if err != nil {
			log.Error("BFT: error writing validator vote", "err", err)
			return err
		}
	}
	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
//
// Note, the block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (sb *Backend) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, _ []*types.Withdrawal) {
	sb.Engine().Finalize(chain, header, state, txs, uncles)
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
		go sb.EventMux().Post(qbft.RequestEvent{
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

func (sb *Backend) processExtraSeals() (map[common.Hash][]byte, map[common.Hash][]byte) {
	if sb.core == nil {
		return nil, nil
	} else {
		lastProposal := sb.currentBlock()
		extraPreparedSeal, extraCommittedSeal := sb.core.ProcessExtraSeal(lastProposal, sb.core.PriorRound())
		return extraPreparedSeal, extraCommittedSeal
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
func (sb *Backend) Start(chain consensus.ChainHeaderReader, currentBlock func() *types.Block, hasBadBlock func(db ethdb.Reader, hash common.Hash) bool) error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if sb.coreStarted {
		return qbft.ErrStartedEngine
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

	log.Info("start QBFT")
	err := sb.startQBFT()

	if err != nil {
		return err
	}

	sb.coreStarted = true

	return nil
}

// Stop implements consensus.Istanbul.Stop
func (sb *Backend) Stop() error {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()
	if !sb.coreStarted {
		return qbft.ErrStoppedEngine
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
		if len(args) != 3 {
			return qbftcommon.ErrInvalidSpecificCall
		}
		chain, ok := args[0].(consensus.ChainHeaderReader)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		currentBlock, ok := args[1].(func() *types.Block)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		hasBadBlock, ok := args[2].(func(db ethdb.Reader, hash common.Hash) bool)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		if sb.coreStarted {
			_ = sb.Stop()
		}
		return sb.Start(chain, currentBlock, hasBadBlock)
	case "SetExtra":
		if len(args) != 2 {
			return qbftcommon.ErrInvalidSpecificCall
		}
		val, ok := args[0].(common.Address)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		header, ok := args[1].(*types.Header)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		vals := make([]common.Address, 1)
		vals[0] = val
		qbftengine.ApplyHeaderQBFTExtra(
			header,
			func(qbftExtra *types.QBFTExtra) error {
				qbftExtra.Validators = vals
				return nil
			})
		return nil
	case "InheritExtra":
		if len(args) != 2 {
			return qbftcommon.ErrInvalidSpecificCall
		}
		parent, ok := args[0].(*types.Header)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		header, ok := args[1].(*types.Header)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		extra, err := types.ExtractQBFTExtra(parent)
		if err != nil {
			return err
		} else if extra.PreparedSeal == nil {
			return qbftcommon.ErrEmptyPreparedSeals
		} else if extra.CommittedSeal == nil {
			// TODO : what if there is not committedSeal that node collected?
			return qbftcommon.ErrEmptyCommittedSeals
		}

		prevPreparedSeal := extra.PreparedSeal
		prevCommittedSeal := extra.CommittedSeal
		// add validators in snapshot to extraData's validators section and lastBlock committers to extraData's prevCommittedSeal section
		qbftengine.ApplyHeaderQBFTExtra(
			header,
			qbftengine.WriteValidators(extra.Validators),
			qbftengine.WritePrevPreparedSeal(prevPreparedSeal),
			qbftengine.WritePrevCommittedSeal(prevCommittedSeal))
		return nil
	case "NewChainHead":
		return sb.NewChainHead()

	case "SetCoinbase":
		header, ok := args[0].(*types.Header)
		if !ok {
			return qbftcommon.ErrInvalidSpecificCall
		}
		header.Coinbase = sb.Engine().Address()
		return nil
	default:
		return qbftcommon.ErrInvalidSpecificCall
	}
}

func addrsToString(addrs []common.Address) []string {
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = addr.String()
	}
	return strs
}

func (sb *Backend) snapLogger(snap *Snapshot) log.Logger {
	return sb.logger.New(
		"snap.number", snap.Number,
		"snap.hash", snap.Hash.String(),
		"snap.epoch", snap.Epoch,
		"snap.validators", addrsToString(snap.validators()),
		"snap.votes", snap.Votes,
	)
}

func (sb *Backend) storeSnap(snap *Snapshot) error {
	logger := sb.snapLogger(snap)
	logger.Debug("BFT: store snapshot to database")
	if err := snap.store(sb.db); err != nil {
		logger.Error("BFT: failed to store snapshot to database", "err", err)
		return err
	}

	return nil
}

// snapshot retrieves the authorization snapshot at a given point in time.
func (sb *Backend) snapshot(chain consensus.ChainHeaderReader, number uint64, hash common.Hash, parents []*types.Header) (*Snapshot, error) {
	// Search for a snapshot in memory or on disk for checkpoints
	var (
		headers []*types.Header
		snap    *Snapshot
	)
	for snap == nil {
		// If an in-memory snapshot was found, use that
		if s, ok := sb.recents.Get(hash); ok {
			snap = s
			sb.snapLogger(snap).Trace("BFT: loaded voting snapshot from cache")
			break
		}
		// If an on-disk checkpoint snapshot can be found, use that
		if number%checkpointInterval == 0 {
			if s, err := loadSnapshot(sb.config.GetConfig(new(big.Int).SetUint64(number)).Epoch, sb.db, hash); err == nil {
				snap = s
				sb.snapLogger(snap).Trace("BFT: loaded voting snapshot from database")
				break
			}
		}

		// If we're at block zero, make a snapshot
		if number == 0 {
			genesis := chain.GetHeaderByNumber(0)
			if err := sb.Engine().VerifyHeader(chain, genesis, nil, nil, nil, false); err != nil {
				sb.logger.Error("BFT: invalid genesis block", "err", err)
				return nil, err
			}

			var validators []common.Address
			validatorsFromConfig := sb.config.GetValidatorsAt(big.NewInt(0))
			if len(validatorsFromConfig) > 0 {
				validators = validatorsFromConfig
				log.Info("BFT: Initialising snap with config validators", "validators", validators)
			} else {
				var err error
				validators, err = sb.Engine().ExtractGenesisValidators(genesis)
				log.Info("BFT: Initialising snap with extradata", "validators", validators)
				if err != nil {
					log.Error("BFT: invalid genesis block", "err", err)
					return nil, err
				}
			}

			snap = newSnapshot(sb.config.GetConfig(new(big.Int).SetUint64(number)).Epoch, 0, genesis.Hash(), validator.NewSet(validators, sb.config.ProposerPolicy))
			if err := sb.storeSnap(snap); err != nil {
				return nil, err
			}
			break
		}

		// No snapshot for this header, gather the header and move backward
		var header *types.Header
		if len(parents) > 0 {
			// If we have explicit parents, pick from there (enforced)
			header = parents[len(parents)-1]
			if header.Hash() != hash || header.Number.Uint64() != number {
				return nil, consensus.ErrUnknownAncestor
			}
			parents = parents[:len(parents)-1]
		} else {
			// No explicit parents (or no more left), reach out to the database
			header = chain.GetHeader(hash, number)
			if header == nil {
				return nil, consensus.ErrUnknownAncestor
			}
		}

		headers = append(headers, header)
		number, hash = number-1, header.ParentHash
	}

	// Previous snapshot found, apply any pending headers on top of it
	for i := 0; i < len(headers)/2; i++ {
		headers[i], headers[len(headers)-1-i] = headers[len(headers)-1-i], headers[i]
	}

	snapApplied, err := sb.snapApply(snap, headers)
	if err != nil {
		return nil, err
	}
	sb.recents.Add(snapApplied.Hash, snapApplied)

	targetBlockHeight := new(big.Int).SetUint64(number)
	// we only need to update the validator set if it's a new block
	if validatorsFromTransitions := sb.config.GetValidatorsAt(targetBlockHeight); len(validatorsFromTransitions) > 0 && sb.config.GetValidatorSelectionMode(targetBlockHeight) == params.BlockHeaderMode {
		//Note! we only want to set this once at this block height. Subsequent blocks will be propagated with the same
		// 		validator as they are copied into the block header on the next block. Then normal voting can take place
		// 		again.
		valSet := validator.NewSet(validatorsFromTransitions, sb.config.ProposerPolicy)
		snapApplied.ValSet = valSet
	}

	// If we've generated a new checkpoint snapshot, save to disk
	if snapApplied.Number%checkpointInterval == 0 && len(headers) > 0 {
		if err = sb.storeSnap(snapApplied); err != nil {
			return nil, err
		}
	}

	return snapApplied, err
}

// SealHash returns the hash of a block prior to it being sealed.
func (sb *Backend) SealHash(header *types.Header) common.Hash {
	return sb.Engine().SealHash(header)
}

func (sb *Backend) snapApply(snap *Snapshot, headers []*types.Header) (*Snapshot, error) {
	// Allow passing in no headers for cleaner code
	if len(headers) == 0 {
		return snap, nil
	}
	// Sanity check that the headers can be applied
	for i := 0; i < len(headers)-1; i++ {
		if headers[i+1].Number.Uint64() != headers[i].Number.Uint64()+1 {
			return nil, qbftcommon.ErrInvalidVotingChain
		}
	}
	if headers[0].Number.Uint64() != snap.Number+1 {
		return nil, qbftcommon.ErrInvalidVotingChain
	}
	// Iterate through the headers and create a new snapshot
	snapCpy := snap.copy()

	for _, header := range headers {
		err := sb.snapApplyHeader(snapCpy, header)
		if err != nil {
			return nil, err
		}
	}
	snapCpy.Number += uint64(len(headers))
	snapCpy.Hash = headers[len(headers)-1].Hash()

	return snapCpy, nil
}

func (sb *Backend) snapApplyHeader(snap *Snapshot, header *types.Header) error {
	logger := sb.snapLogger(snap).New("header.number", header.Number.Uint64(), "header.hash", header.Hash().String())

	logger.Trace("BFT: apply header to voting snapshot")

	// Remove any votes on checkpoint blocks
	number := header.Number.Uint64()
	if number%snap.Epoch == 0 {
		snap.Votes = nil
		snap.Tally = make(map[common.Address]Tally)
	}

	// Resolve the authorization key and check against validators
	validator, err := sb.Engine().Author(header)
	if err != nil {
		logger.Error("BFT: invalid header author", "err", err)
		return err
	}

	logger = logger.New("header.author", validator)

	if _, v := snap.ValSet.GetByAddress(validator); v == nil {
		logger.Error("BFT: header author is not a validator", "Validators", snap.ValSet, "Author", validator)
		return qbftcommon.ErrUnauthorized
	}

	// Read vote from header
	candidate, authorize, err := sb.Engine().ReadVote(header)
	if err != nil {
		logger.Error("BFT: invalid header vote", "err", err)
		return err
	}

	logger = logger.New("candidate", candidate.String(), "authorize", authorize)
	// Header authorized, discard any previous votes from the validator
	for i, vote := range snap.Votes {
		if vote.Validator == validator && vote.Address == candidate {
			logger.Trace("BFT: discard previous vote from tally", "old.authorize", vote.Authorize)
			// Uncast the vote from the cached tally
			snap.uncast(vote.Address, vote.Authorize)

			// Uncast the vote from the chronological list
			snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)
			break // only one vote allowed
		}
	}

	logger.Debug("BFT: add vote to tally")
	if snap.cast(candidate, authorize) {
		snap.Votes = append(snap.Votes, &Vote{
			Validator: validator,
			Block:     number,
			Address:   candidate,
			Authorize: authorize,
		})
	}

	// If the vote passed, update the list of validators
	if tally := snap.Tally[candidate]; tally.Votes > snap.ValSet.Size()/2 {
		if tally.Authorize {
			logger.Info("BFT: reached majority to add validator")
			snap.ValSet.AddValidator(candidate)
		} else {
			logger.Info("BFT: reached majority to remove validator")
			snap.ValSet.RemoveValidator(candidate)

			// Discard any previous votes the deauthorized validator cast
			for i := 0; i < len(snap.Votes); i++ {
				if snap.Votes[i].Validator == candidate {
					// Uncast the vote from the cached tally
					snap.uncast(snap.Votes[i].Address, snap.Votes[i].Authorize)

					// Uncast the vote from the chronological list
					snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)

					i--
				}
			}
		}
		// Discard any previous votes around the just changed account
		for i := 0; i < len(snap.Votes); i++ {
			if snap.Votes[i].Address == candidate {
				snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)
				i--
			}
		}
		delete(snap.Tally, candidate)
	}
	return nil
}

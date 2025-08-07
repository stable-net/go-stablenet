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
// This file is derived from quorum/consensus/istanbul/wbft/core/core.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	metrics "github.com/ethereum/go-ethereum/metrics"
)

type SealType byte

const (
	SealTypePrepare SealType = iota
	SealTypeCommit
)

var (
	roundMeter        = metrics.NewRegisteredMeter("consensus/wbft/core/round", nil)
	sequenceMeter     = metrics.NewRegisteredMeter("consensus/wbft/core/sequence", nil)
	consensusTimer    = metrics.NewRegisteredTimer("consensus/wbft/core/consensus", nil)
	timeoutRoundMeter = metrics.NewRegisteredMeter("consensus/wbft/core/timeout_round", nil)
)

// New creates a WBFT consensus core
func New(backend Backend, config *wbft.Config) *Core {
	c := &Core{
		config:             config,
		address:            backend.Address(),
		state:              StateAcceptRequest,
		handlerWg:          new(sync.WaitGroup),
		logger:             log.New("address", backend.Address()),
		backend:            backend,
		backlogs:           make(map[common.Address]*prque.Prque[int64, wbftmessage.WBFTMessage]),
		backlogsMu:         new(sync.Mutex),
		prepareExtraSeals:  make(map[common.Address]*wbftmessage.Prepare),
		commitExtraSeals:   make(map[common.Address]*wbftmessage.Commit),
		extraSealsMu:       new(sync.Mutex),
		pendingRequests:    prque.New[int64, *Request](nil),
		pendingRequestsMu:  new(sync.Mutex),
		consensusTimestamp: time.Time{},
		priorState:         priorState{new(sync.RWMutex), common.Big0, nil, nil},
	}

	c.validateFn = c.checkValidatorSignature
	return c
}

// ----------------------------------------------------------------------------

type Core struct {
	config  *wbft.Config
	address common.Address
	state   State
	logger  log.Logger

	backend               Backend
	events                *event.TypeMuxSubscription
	finalCommittedSub     *event.TypeMuxSubscription
	timeoutSub            *event.TypeMuxSubscription
	retryTimeoutSub       *event.TypeMuxSubscription
	futurePreprepareTimer *time.Timer

	valSet     wbft.ValidatorSet
	validateFn func([]byte, []byte, wbft.View) (common.Address, error)

	backlogs   map[common.Address]*prque.Prque[int64, wbftmessage.WBFTMessage]
	backlogsMu *sync.Mutex

	prepareExtraSeals map[common.Address]*wbftmessage.Prepare
	commitExtraSeals  map[common.Address]*wbftmessage.Commit
	extraSealsMu      *sync.Mutex
	priorState        priorState

	current      *roundState
	currentMutex sync.Mutex
	handlerWg    *sync.WaitGroup

	roundChangeSet          *roundChangeSet
	roundChangeTimer        *time.Timer
	lastSentTimeoutCanceled *bool

	retrySendingRoundChangeTimer *time.Timer

	WBFTPreparedPrepares []*wbftmessage.Prepare

	pendingRequests   *prque.Prque[int64, *Request]
	pendingRequestsMu *sync.Mutex

	consensusTimestamp time.Time
}

func (c *Core) currentView() *wbft.View {
	return &wbft.View{
		Sequence: new(big.Int).Set(c.current.Sequence()),
		Round:    new(big.Int).Set(c.current.Round()),
	}
}

func (c *Core) PriorRound() *big.Int {
	return c.priorState.Round()
}

func (c *Core) PriorValidators() wbft.ValidatorSet {
	return c.priorState.Validators()
}

func (c *Core) IsProposer() bool {
	v := c.valSet
	if v == nil {
		return false
	}
	return v.IsProposer(c.backend.Address())
}

func (c *Core) GetProposer() common.Address {
	if c.valSet != nil {
		return c.valSet.GetProposer().Address()
	}
	return common.Address{}
}

func (c *Core) IsCurrentProposal(blockHash common.Hash) bool {
	return c.current != nil && c.current.pendingRequest != nil && c.current.pendingRequest.Proposal.Hash() == blockHash
}

// startNewRound starts a new round. if round equals to 0, it means to starts a new sequence
func (c *Core) startNewRound(round *big.Int) {
	c.currentMutex.Lock()
	defer c.currentMutex.Unlock()

	var logger log.Logger
	if c.current == nil {
		logger = c.logger.New("old.round", -1, "old.seq", 0)
	} else {
		logger = c.currentLogger(false, nil)
	}
	logger = logger.New("target.round", round)

	roundChange := false

	// Try to get last proposal
	lastProposal, lastProposer := c.backend.LastProposal()
	if lastProposal != nil {
		logger = logger.New("lastProposal.number", lastProposal.Number().Uint64(), "lastProposal.hash", lastProposal.Hash())
	}

	logger.Info("WBFT: initialize new round")

	if c.current == nil {
		logger.Debug("WBFT: start at the initial round")
	} else if lastProposal.Number().Cmp(c.current.Sequence()) >= 0 {
		diff := new(big.Int).Sub(lastProposal.Number(), c.current.Sequence())
		sequenceMeter.Mark(new(big.Int).Add(diff, common.Big1).Int64())

		if !c.consensusTimestamp.IsZero() {
			consensusTimer.UpdateSince(c.consensusTimestamp)
			c.consensusTimestamp = time.Time{}
		}
		logger.Debug("WBFT: catch up last block proposal")
	} else if lastProposal.Number().Cmp(big.NewInt(c.current.Sequence().Int64()-1)) == 0 {
		if round.Cmp(common.Big0) == 0 {
			// same seq and round, don't need to start new round
			logger.Debug("WBFT: same round, no need to start new round")
			return
		} else if round.Cmp(c.current.Round()) < 0 {
			logger.Warn("WBFT: next round is inferior to current round")
			return
		}
		roundChange = true
	} else {
		logger.Warn("WBFT: next sequence is before last block proposal")
		return
	}

	var oldLogger log.Logger
	if c.current == nil {
		oldLogger = c.logger.New("old.round", -1, "old.seq", 0)
	} else {
		oldLogger = c.logger.New("old.round", c.current.Round().Uint64(), "old.sequence", c.current.Sequence().Uint64(), "old.state", c.state.String(), "old.proposer", c.valSet.GetProposer())
	}

	if c.current != nil && round.Cmp(c.current.Round()) > 0 {
		roundMeter.Mark(new(big.Int).Sub(round, c.current.Round()).Int64())
	}

	// Create next view
	var newView *wbft.View
	var nextValSet wbft.ValidatorSet
	if roundChange {
		newView = &wbft.View{
			Sequence: new(big.Int).Set(c.current.Sequence()),
			Round:    new(big.Int).Set(round),
		}
		nextValSet = c.valSet
	} else {
		newView = &wbft.View{
			Sequence: new(big.Int).Add(lastProposal.Number(), common.Big1),
			Round:    new(big.Int),
		}
		nextValSet = c.backend.Validators(lastProposal)
	}

	// Add extra seal that contributed to consensus
	c.addEffectiveSealToExtraSeal()

	// New snapshot for new round
	c.updateRoundState(nextValSet, newView, roundChange)

	// Calculate new proposer
	c.valSet.CalcProposer(lastProposer, newView.Round.Uint64())
	c.setState(StateAcceptRequest)

	// Update RoundChangeSet by deleting older round messages
	if round.Uint64() == 0 {
		c.WBFTPreparedPrepares = nil
		c.roundChangeSet = newRoundChangeSet(c.valSet)
		c.ClearExtraSeals(lastProposal.Number())
	} else {
		// Clear earlier round messages
		c.roundChangeSet.ClearLowerThan(round)
	}
	c.roundChangeSet.NewRound(round)

	c.backend.NotifyNewRound(round)

	// the order of NotifyNewRound() and newRoundChangeTimer() does not matter on actual consensus, but
	// it matters on multi-engine test, so we keep the order as it is
	c.newRoundChangeTimer()

	oldLogger.Info("WBFT: start new round", "next.round", newView.Round, "next.seq", newView.Sequence, "next.proposer", c.valSet.GetProposer(), "next.valSet", c.valSet.List(), "next.size", c.valSet.Size(), "next.IsProposer", c.IsProposer())
}

// updateRoundState updates round state by checking if locking block is necessary
func (c *Core) updateRoundState(nextValSet wbft.ValidatorSet, view *wbft.View, roundChange bool) {
	if roundChange && c.current != nil {
		if c.current.preparedBlock != nil && c.backend.HasBadProposal(c.current.preparedBlock.Hash()) {
			c.currentLogger(false, nil).Warn("[QBFT] Discarding prepared block due to bad proposal", "hash", c.current.preparedBlock.Hash())
			// clear prepared round and block if we have a bad proposal
			c.current.preparedRound = nil
			c.current.preparedBlock = nil
			// clear prepare and commit seals for the current sequence
			c.ClearExtraSeals(new(big.Int).Add(c.current.sequence, common.Big1))
		}
		c.current = newRoundState(view, nextValSet, c.current.Preprepare, c.current.preparedRound, c.current.preparedBlock, c.current.pendingRequest, c.backend.HasBadProposal)
	} else {
		if c.current != nil {
			// priorState is only set for finalCommitted block
			c.updatePriorState()
		}
		c.current = newRoundState(view, nextValSet, nil, nil, nil, nil, c.backend.HasBadProposal)
	}
	c.valSet = nextValSet
}

func (c *Core) setState(state State) {
	if c.state != state {
		oldState := c.state
		c.state = state
		c.currentLogger(false, nil).Info("WBFT: changed state", "old.state", oldState.String(), "new.state", state.String())
	}
	if state == StateAcceptRequest {
		c.processPendingRequests()
	}

	// each time we change state, we process backlog for possible message that are
	// now ready
	c.processBacklog()
}

func (c *Core) GetState() State {
	return c.state
}

func (c *Core) Address() common.Address {
	return c.address
}

func (c *Core) stopFuturePreprepareTimer() {
	if c.futurePreprepareTimer != nil {
		c.futurePreprepareTimer.Stop()
	}
}

func (c *Core) stopTimer() {
	c.stopFuturePreprepareTimer()

	// Stop retry sending ROUND-CHANGE retry timer
	c.stopRetrySendingRoundChangeTimer()

	if c.roundChangeTimer != nil {
		c.roundChangeTimer.Stop()
	}
	if c.lastSentTimeoutCanceled != nil {
		*c.lastSentTimeoutCanceled = true
	}
}

func (c *Core) newRoundChangeTimer() {
	c.stopTimer()

	for c.current == nil { // wait because it is asynchronous in handleRequest
		time.Sleep(10 * time.Millisecond)
	}

	// set timeout based on the round number
	cfg := c.config.GetConfig(c.current.Sequence())
	baseTimeout := time.Duration(cfg.RequestTimeout) * time.Millisecond
	round := c.current.Round().Uint64()
	maxRequestTimeout := time.Duration(cfg.MaxRequestTimeoutSeconds) * time.Second

	// If the upper limit of the request timeout is capped by small maxRequestTimeout, round can be a quite large number,
	// which leads to float64 overflow, making its value negative or zero forever after some point.
	// In this case we cannot simply use math.Pow and have to implement a safeguard on our own, at the cost of performance (which is not important in this case).
	var timeout time.Duration
	if maxRequestTimeout > time.Duration(0) {
		timeout = baseTimeout
		for i := uint64(0); i < round; i++ {
			timeout = timeout * 2
			if timeout > maxRequestTimeout {
				timeout = maxRequestTimeout
				break
			}
		}
		// prevent log storm when unexpected overflow happens
		if timeout < baseTimeout {
			c.currentLogger(true, nil).Warn("WBFT: Possible request timeout overflow detected, setting timeout value to maxRequestTimeout",
				"timeout", timeout.Seconds(),
				"max_request_timeout", maxRequestTimeout.Seconds(),
			)
			timeout = maxRequestTimeout
		}
	} else {
		timeoutFloat64 := math.Pow(2, float64(round)) * float64(baseTimeout)

		if math.IsNaN(timeoutFloat64) || math.IsInf(timeoutFloat64, 0) || timeoutFloat64 > float64(math.MaxInt64) {
			c.currentLogger(true, nil).Warn("WBFT: Timeout overflow detected, setting timeout value to MaxInt64",
				"round", round,
				"adjusted_timeout", time.Duration(math.MaxInt64).Seconds(),
			)
			timeout = time.Duration(math.MaxInt64)
		} else {
			timeout = time.Duration(int64(timeoutFloat64))
		}
	}

	c.currentLogger(true, nil).Trace("WBFT: start new ROUND-CHANGE timer", "timeout", timeout.Seconds())
	c.lastSentTimeoutCanceled = new(bool)
	*c.lastSentTimeoutCanceled = false
	c.roundChangeTimer = time.AfterFunc(timeout, func() {
		c.sendEvent(timeoutEvent{c.lastSentTimeoutCanceled})
	})
}

// stopRetrySendingRoundChangeTimer stops the round-change retry timer if running.
func (c *Core) stopRetrySendingRoundChangeTimer() {
	if c.retrySendingRoundChangeTimer != nil {
		c.retrySendingRoundChangeTimer.Stop()
	}
}

// newRetrySendingRoundChangeTimer sets a retry timer to reattempt round-change after a timeout
func (c *Core) newRetrySendingRoundChangeTimer() {
	c.stopRetrySendingRoundChangeTimer()

	// set timeout based on the round number
	cfg := c.config.GetConfig(c.current.Sequence())
	timeout := time.Duration(cfg.RequestTimeout) * time.Millisecond

	c.currentLogger(true, nil).Trace("WBFT: set ROUND-CHANGE retry timer", "round", c.current.Round(), "timeout", timeout.Seconds())
	c.retrySendingRoundChangeTimer = time.AfterFunc(timeout, func() {
		c.sendEvent(retryTimeoutEvent{c.current.Round()})
	})
}

func (c *Core) checkValidatorSignature(data []byte, sig []byte, view wbft.View) (common.Address, error) {
	if view.Cmp(c.currentView()) < 0 && c.state == StateAcceptRequest && view.Cmp(&wbft.View{
		Sequence: new(big.Int).Sub(c.current.Sequence(), common.Big1),
		Round:    c.PriorRound(),
	}) == 0 {
		if valSet := c.PriorValidators(); valSet != nil {
			return wbft.CheckValidatorSignature(valSet, data, sig)
		}
	}
	return wbft.CheckValidatorSignature(c.valSet, data, sig)
}

// PrepareSeal returns a committed seal for the given header and takes current round under consideration
func PrepareSeal(header *types.Header, round uint32, sealType SealType) []byte {
	h := types.CopyHeader(header)
	roundHeader := h.WBFTHashWithRoundNumber(round).Bytes()
	return crypto.Keccak256Hash(append(roundHeader, byte(sealType))).Bytes()
}

func verifySeal(valSet wbft.ValidatorSet, header *types.Header, round uint32, sealType SealType, seal []byte, sealer common.Address) error {
	_, validator := valSet.GetByAddress(sealer)

	pubkey, err := bls.PublicKeyFromBytes(validator.BLSPublicKey())
	if err != nil {
		return err
	}

	sig, err := bls.SignatureFromBytes(seal)
	if err != nil {
		return errInvalidSeal
	}

	if !sig.Verify(pubkey, PrepareSeal(header, round, sealType)) {
		return errInvalidSigner
	}
	return nil
}

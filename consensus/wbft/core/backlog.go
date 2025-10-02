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
// This file is derived from quorum/consensus/istanbul/qbft/core/backlog.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
)

// 1. package "gopkg.in/karalabe/cookiejar.v2/collections/prque" is replaced to  "github.com/ethereum/go-ethereum/common/prque"

var (
	// msgPriority is defined for calculating processing priority to speedup consensus
	// msgPreprepare > msgCommit > msgPrepare
	msgPriority = map[uint64]int{
		wbfmessage.PreprepareCode: 1,
		wbfmessage.CommitCode:     2,
		wbfmessage.PrepareCode:    3,
	}
)

const (
	sequenceThreshold = 1  // Allow up to 1 future sequence
	roundThreshold    = 10 // Allow up to 10 future rounds
)

// isSequenceTooFarAhead returns true if the sequence difference exceeds the threshold
func (c *Core) isSequenceTooFarAhead(viewSeq, currSeq *big.Int, threshold int64) (*big.Int, bool) {
	seqDiff := new(big.Int).Sub(viewSeq, currSeq)
	return seqDiff, seqDiff.Cmp(big.NewInt(threshold)) >= 0
}

// isRoundTooFarAhead returns true if the round difference exceeds the threshold
func (c *Core) isRoundTooFarAhead(viewRound, currRound *big.Int, threshold int64) (*big.Int, bool) {
	roundDiff := new(big.Int).Sub(viewRound, currRound)
	return roundDiff, roundDiff.Cmp(big.NewInt(threshold)) >= 0
}

// isTooFarFutureMessage filters out messages too far ahead in sequence or round
// This function prevents the node from processing messages that are excessively ahead,
// which could disrupt the normal flow of the consensus process
func (c *Core) isTooFarFutureMessage(view *wbft.View) bool {
	curr := c.currentView()

	if view.Sequence.Cmp(curr.Sequence) > 0 {
		// In the initial phase of block consensus, a message with a sequence number one higher than the current sequence may be received.
		if seqDiff, tooFar := c.isSequenceTooFarAhead(view.Sequence, curr.Sequence, sequenceThreshold); tooFar {
			c.logger.Trace("WBFT: future message too far ahead in sequence, dropped",
				"msg_seq", view.Sequence.String(),
				"curr_seq", curr.Sequence.String(),
				"diff", seqDiff.String(),
			)
			return true
		}
	}

	if view.Sequence.Cmp(curr.Sequence) == 0 && view.Round.Cmp(curr.Round) > 0 {
		if roundDiff, tooFar := c.isRoundTooFarAhead(view.Round, curr.Round, roundThreshold); tooFar {
			c.logger.Trace("WBFT: future message too far ahead in round, dropped",
				"msg_round", view.Round.String(),
				"curr_round", curr.Round.String(),
				"diff", roundDiff.String(),
			)
			return true
		}
	}

	return false
}

// checkMessage checks that a message matches our current WBFT state
//
// In particular it ensures that
// - message has the expected round
// - message has the expected sequence
// - message type is expected given our current state

// return errInvalidMessage if the message is invalid
// return errFutureMessage if the message view is larger than current view
// return errOldMessage if the message view is smaller than current view
func (c *Core) checkMessage(msgCode uint64, view *wbft.View) error {
	if view == nil || view.Sequence == nil || view.Round == nil {
		return errInvalidMessage
	}

	// Drop the message if the view's sequence or round number is too far ahead
	// This prevents processing of messages that may disrupt consensus due to excessive lead
	if c.isTooFarFutureMessage(view) {
		return errFutureViewTooFar
	}

	if msgCode == wbfmessage.RoundChangeCode {
		// if ROUND-CHANGE message
		// check that
		// - sequence matches our current sequence
		// - round is in the future
		if view.Sequence.Cmp(c.currentView().Sequence) > 0 {
			return errFutureMessage
		} else if view.Cmp(c.currentView()) < 0 {
			return errOldMessage
		}
		return nil
	}

	// If not ROUND-CHANGE
	// check that round and sequence equals our current round and sequence
	if view.Cmp(c.currentView()) > 0 {
		return errFutureMessage
	}

	if view.Cmp(c.currentView()) < 0 {
		// save prepare and commit message to extraSeal under below condition
		// 1. view's sequence is right before current.Sequence &&
		// 2. view's round is same as prior round &&
		// 3. c.state is AcceptRequest
		if new(big.Int).Sub(c.currentView().Sequence, view.Sequence).Cmp(common.Big1) == 0 && view.Round.Cmp(c.PriorRound()) == 0 && c.state == StateAcceptRequest {
			return errExtraSealMessage
		}
		return errOldMessage
	}

	switch c.state {
	case StateAcceptRequest:
		// StateAcceptRequest only accepts msgPreprepare and msgRoundChange
		// other messages are future messages
		if msgCode > wbfmessage.PreprepareCode {
			return errFutureMessage
		}
		return nil
	case StatePreprepared:
		// StatePreprepared only accepts msgPrepare and msgRoundChange
		// message less than msgPrepare are invalid and greater are future messages
		if msgCode < wbfmessage.PrepareCode {
			return errInvalidMessage
		} else if msgCode > wbfmessage.PrepareCode {
			return errFutureMessage
		}
		return nil
	case StatePrepared:
		// StatePrepared only accepts msgCommit and msgRoundChange
		// other messages are invalid messages
		if msgCode == wbfmessage.PrepareCode {
			return errExtraSealMessage
		} else if msgCode < wbfmessage.CommitCode {
			return errInvalidMessage
		}
		return nil
	case StateCommitted:
		// for prepare, commit message with same view as current
		// return err extraSealMessage
		if msgCode >= wbfmessage.PrepareCode {
			return errExtraSealMessage
		}
		// StateCommit rejects all messages other than msgRoundChange
		return errInvalidMessage
	}
	return nil
}

// addToBacklog allows to postpone the processing of future messages

// it adds the message to backlog which is read on every state change
func (c *Core) addToBacklog(msg wbfmessage.WBFTMessage) {
	logger := c.currentLogger(true, msg)

	src := msg.Source()
	if src == c.Address() {
		logger.Warn("WBFT: backlog from self")
		return
	}

	logger.Trace("WBFT: new backlog message", "backlogs_size", len(c.backlogs))

	c.backlogsMu.Lock()
	defer c.backlogsMu.Unlock()

	backlog := c.backlogs[src]
	if backlog == nil {
		backlog = prque.New[int64, wbfmessage.WBFTMessage](nil)
		c.backlogs[src] = backlog
	}
	view := msg.View()
	backlog.Push(msg, toNegatePriority(msg.Code(), &view))
}

// processBacklog lookup for future messages that have been backlogged and post it on
// the event channel so main handler loop can handle it

// It is called on every state change
func (c *Core) processBacklog() {
	c.backlogsMu.Lock()
	defer c.backlogsMu.Unlock()

	for srcAddress, backlog := range c.backlogs {
		if backlog == nil {
			continue
		}
		_, src := c.valSet.GetByAddress(srcAddress)
		if src == nil {
			// validator is not available
			delete(c.backlogs, srcAddress)
			continue
		}
		logger := c.logger.New("from", src, "state", c.state)
		isFuture := false

		logger.Trace("WBFT: process backlog")

		// We stop processing if
		//   1. backlog is empty
		//   2. The first message in queue is a future message
		for !(backlog.Empty() || isFuture) {
			msg, prio := backlog.Pop()

			var code uint64
			var view wbft.View
			var event backlogEvent

			code = msg.Code()
			view = msg.View()
			event.msg = msg

			// Push back if it's a future message
			err := c.checkMessage(code, &view)
			if err != nil && err != errExtraSealMessage {
				if err == errFutureMessage {
					// this is still a future message
					logger.Trace("WBFT: stop processing backlog", "msg", msg)
					backlog.Push(msg, prio)
					isFuture = true
					break
				}
				logger.Trace("WBFT: skip backlog message", "msg", msg, "err", err)
				continue
			}
			logger.Trace("WBFT: post backlog event", "msg", msg)

			event.src = src
			go c.sendEvent(event)
		}
	}
}

func toNegatePriority(msgCode uint64, view *wbft.View) int64 {
	if msgCode == wbfmessage.RoundChangeCode {
		// For msgRoundChange, set the message priority based on its sequence
		return -int64(view.Sequence.Uint64() * 1000)
	}
	// FIXME: round will be reset as 0 while new sequence
	// 10 * Round limits the range of message code is from 0 to 9
	// 1000 * Sequence limits the range of round is from 0 to 99
	return -int64(view.Sequence.Uint64()*1000 + view.Round.Uint64()*10 + uint64(msgPriority[msgCode]))
}

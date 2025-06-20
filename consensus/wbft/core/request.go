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
// This file is derived from quorum/consensus/istanbul/wbft/core/request.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"github.com/ethereum/go-ethereum/consensus/wbft"
)

// handleRequest is called by proposer in reaction to `miner.Seal()`
// (this is the starting of the WBFT validation process)

// It
// - validates block proposal is not empty and number correspond to the current sequence
// - creates and send PRE-PREPARE message to other validators
func (c *Core) handleRequest(request *Request) error {
	logger := c.currentLogger(true, nil)

	logger.Info("WBFT: handle block proposal request")

	if err := c.checkRequestMsg(request); err != nil {
		if err == errInvalidMessage {
			logger.Error("WBFT: invalid request")
			return err
		}
		logger.Error("WBFT: unexpected request", "err", err, "number", request.Proposal.Number(), "hash", request.Proposal.Hash())
		return err
	}

	c.current.pendingRequest = request
	if c.state == StateAcceptRequest {
		if c.current.Round().Uint64() == 0 {
			// if round > 0, then we don't send preprepare because it would be failed due to having no justification
			// after that we will send preprepare when 2/3+ round change messages are received
			c.sendPreprepareMsg(request)
		}
	}

	return nil
}

// check request state
// return errInvalidMessage if the message is invalid
// return errFutureMessage if the sequence of proposal is larger than current sequence
// return errOldMessage if the sequence of proposal is smaller than current sequence
func (c *Core) checkRequestMsg(request *Request) error {
	if request == nil || request.Proposal == nil {
		return errInvalidMessage
	}
	if c.current == nil {
		return errCurrentIsNil
	}

	if cmp := c.current.sequence.Cmp(request.Proposal.Number()); cmp > 0 {
		return errOldMessage
	} else if cmp < 0 {
		return errFutureMessage
	} else {
		return nil
	}
}

func (c *Core) storeRequestMsg(request *Request) {
	logger := c.currentLogger(true, nil).New("proposal.number", request.Proposal.Number(), "proposal.hash", request.Proposal.Hash())

	logger.Trace("WBFT: store block proposal request for future treatment")

	c.pendingRequestsMu.Lock()
	defer c.pendingRequestsMu.Unlock()

	c.pendingRequests.Push(request, -request.Proposal.Number().Int64()) // ## WBFT
}

// processPendingRequests is called each time WBFT state is re-initialized
// it lookup over pending requests and re-input its so they can be treated
func (c *Core) processPendingRequests() {
	c.pendingRequestsMu.Lock()
	defer c.pendingRequestsMu.Unlock()

	logger := c.currentLogger(true, nil)
	logger.Debug("WBFT: lookup for pending block proposal requests")

	for !(c.pendingRequests.Empty()) {
		r, prio := c.pendingRequests.Pop()
		// Push back if it's a future message
		err := c.checkRequestMsg(r)
		if err != nil {
			if err == errFutureMessage {
				logger.Trace("WBFT: stop looking up for pending block proposal request")
				c.pendingRequests.Push(r, prio)
				break
			}
			logger.Trace("WBFT: skip pending invalid block proposal request", "number", r.Proposal.Number(), "hash", r.Proposal.Hash(), "err", err)
			continue
		}
		logger.Debug("WBFT: found pending block proposal request", "proposal.number", r.Proposal.Number(), "proposal.hash", r.Proposal.Hash())

		go c.sendEvent(wbft.RequestEvent{
			Proposal: r.Proposal,
		})
	}
}

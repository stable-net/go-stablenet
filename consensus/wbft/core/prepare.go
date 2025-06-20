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
// This file is derived from quorum/consensus/istanbul/wbft/core/prepare.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// broadcastPrepare is called after receiving PRE-PREPARE from proposer node

// It
// - creates a PREPARE message
// - broadcast PREPARE message to other validators
func (c *Core) broadcastPrepare() {
	logger := c.currentLogger(true, nil)

	// Create PREPARE message from the current proposal
	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Prepare Seal
	prepareSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypePrepare))
	prepare := wbfmessage.NewPrepare(sub.View.Sequence, sub.View.Round, sub.Digest, prepareSeal)
	prepare.SetSource(c.Address())

	// Sign Message
	encodedPayload, err := prepare.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to encode payload of PREPARE message", "err", err)
		return
	}
	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to sign PREPARE message", "err", err)
		return
	}
	prepare.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&prepare)
	if err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to encode PREPARE message", "err", err)
		return
	}

	withMsg(logger, prepare).Info("WBFT: broadcast PREPARE message", "payload", hexutil.Encode(payload))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, prepare.Code(), payload); err != nil {
		withMsg(logger, prepare).Error("WBFT: failed to broadcast PREPARE message", "err", err)
		return
	}
}

// handlePrepareMsg is called when receiving a PREPARE message

// It
// - validates PREPARE message digest matches the current block proposal
// - accumulates valid PREPARE message until reaching quorum
// - when quorum is reached update states to "Prepared" and broadcast COMMIT
func (c *Core) handlePrepareMsg(prepare *wbfmessage.Prepare) error {
	logger := c.currentLogger(true, prepare).New()

	logger.Info("WBFT: handle PREPARE message", "prepares.count", c.current.WBFTPrepares.Size(), "quorum", c.valSet.QuorumSize())

	// Check digest
	if prepare.Digest != c.current.Proposal().Hash() {
		logger.Error("WBFT: invalid PREPARE message digest")
		return errInvalidMessage
	}

	// Check prepareSeal
	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		return errInvalidMessage
	}

	if verifySeal(c.valSet, block.Header(), uint32(prepare.CommonPayload.Round.Uint64()), SealTypePrepare,
		prepare.PrepareSeal, prepare.Source()) != nil {
		return errInvalidMessage
	}

	// Save PREPARE messages
	if err := c.current.WBFTPrepares.Add(prepare); err != nil {
		logger.Error("WBFT: failed to save PREPARE message", "err", err)
		return err
	}

	logger = logger.New("prepares.count", c.current.WBFTPrepares.Size(), "quorum", c.valSet.QuorumSize())

	// Change to "Prepared" state if we've received quorum of PREPARE messages
	// and we are in earlier state than "Prepared"
	if (c.current.WBFTPrepares.Size() >= c.valSet.QuorumSize()) && c.state.Cmp(StatePrepared) < 0 {
		logger.Info("WBFT: received quorum of PREPARE messages")

		// Accumulates PREPARE messages
		c.current.preparedRound = c.currentView().Round
		c.WBFTPreparedPrepares = make([]*wbfmessage.Prepare, 0)
		for _, m := range c.current.WBFTPrepares.Values() {
			c.WBFTPreparedPrepares = append(
				c.WBFTPreparedPrepares,
				wbfmessage.NewPrepareWithSigAndSource(
					m.View().Sequence, m.View().Round, m.(*wbfmessage.Prepare).Digest, m.Signature(), m.Source(), m.(*wbfmessage.Prepare).PrepareSeal))
		}

		if c.current.Proposal() != nil && c.current.Proposal().Hash() == prepare.Digest {
			logger.Debug("WBFT: PREPARE message matches proposal", "proposal", c.current.Proposal().Hash(), "prepare", prepare.Digest)
			c.current.preparedBlock = c.current.Proposal()
		}

		c.setState(StatePrepared)
		c.broadcastCommit()
	} else {
		logger.Debug("WBFT: accepted PREPARE messages")
	}

	return nil
}

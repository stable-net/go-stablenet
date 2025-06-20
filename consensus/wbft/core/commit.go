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
// This file is derived from quorum/consensus/istanbul/wbft/core/commit.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// broadcastCommit is called when receiving quorum of PREPARE message

// It
// - creates a COMMIT message from current proposal
// - broadcast COMMIT message to other validators
func (c *Core) broadcastCommit() {
	var err error

	logger := c.currentLogger(true, nil)

	sub := c.current.Subject()

	var header *types.Header
	if block, ok := c.current.Proposal().(*types.Block); ok {
		header = block.Header()
	}

	// Create Commit Seal
	commitSeal := c.backend.SignWithoutHashing(PrepareSeal(header, uint32(c.currentView().Round.Uint64()), SealTypeCommit))
	commit := wbfmessage.NewCommit(sub.View.Sequence, sub.View.Round, sub.Digest, commitSeal)
	commit.SetSource(c.Address())

	// Sign Message
	encodedPayload, err := commit.EncodePayloadForSigning()
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to encode payload of COMMIT message", "err", err)
		return
	}

	signature, err := c.backend.Sign(encodedPayload)
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to sign COMMIT message", "err", err)
		return
	}
	commit.SetSignature(signature)

	// RLP-encode message
	payload, err := rlp.EncodeToBytes(&commit)
	if err != nil {
		withMsg(logger, commit).Error("WBFT: failed to encode COMMIT message", "err", err)
		return
	}

	withMsg(logger, commit).Info("WBFT: broadcast COMMIT message", "payload", hexutil.Encode(payload))

	// Broadcast RLP-encoded message
	if err = c.backend.Broadcast(c.valSet, commit.Code(), payload); err != nil {
		withMsg(logger, commit).Error("WBFT: failed to broadcast COMMIT message", "err", err)
		return
	}
}

// handleCommitMsg is called when receiving a COMMIT message from another validator

// It
// - validates COMMIT message digest matches the current block proposal
// - accumulates valid COMMIT messages until reaching quorum
// - when quorum of COMMIT messages is reached then update state and commits
func (c *Core) handleCommitMsg(commit *wbfmessage.Commit) error {
	logger := c.currentLogger(true, commit)

	logger.Info("WBFT: handle COMMIT message", "commits.count", c.current.WBFTCommits.Size(), "quorum", c.valSet.QuorumSize())

	// Check digest
	if commit.Digest != c.current.Proposal().Hash() {
		logger.Error("WBFT: invalid COMMIT message digest", "digest", commit.Digest, "proposal", c.current.Proposal().Hash().String())
		return errInvalidMessage
	}

	// Check commitSeal
	block, ok := c.current.Proposal().(*types.Block)
	if !ok {
		return errInvalidMessage
	}

	if verifySeal(c.valSet, block.Header(), uint32(commit.CommonPayload.Round.Uint64()), SealTypeCommit,
		commit.CommitSeal, commit.Source()) != nil {
		return errInvalidMessage
	}

	// Add to received msgs
	if err := c.current.WBFTCommits.Add(commit); err != nil {
		c.logger.Error("WBFT: failed to save COMMIT message", "err", err)
		return err
	}

	logger = logger.New("commits.count", c.current.WBFTCommits.Size(), "quorum", c.valSet.QuorumSize())

	// If we reached threshold
	if c.current.WBFTCommits.Size() >= c.valSet.QuorumSize() {
		logger.Info("WBFT: received quorum of COMMIT messages")
		c.commitWBFT()
	} else {
		logger.Debug("WBFT: accepted new COMMIT messages")
	}

	return nil
}

// commitWBFT is called once quorum of commits is reached
// - computes committedSeals from each received commit messages
// - then commits block proposal to database with committed seals
// - broadcast round change
func (c *Core) commitWBFT() {
	c.setState(StateCommitted)

	proposal := c.current.Proposal()
	if proposal != nil {
		// Compute prepared seals
		preparedSeals := make([]wbft.SealData, c.current.WBFTPrepares.Size())
		for i, msg := range c.current.WBFTPrepares.Values() {
			idx, _ := c.valSet.GetByAddress(msg.Source())
			if idx < 0 {
				continue
			}
			preparedSeals[i] = wbft.SealData{
				Sealer: uint32(idx),
				Seal:   make([]byte, types.IstanbulExtraSeal),
			}
			prepareMsg := msg.(*wbfmessage.Prepare)
			copy(preparedSeals[i].Seal[:], prepareMsg.PrepareSeal[:])
		}

		// Compute committed seals
		committedSeals := make([]wbft.SealData, c.current.WBFTCommits.Size())
		for i, msg := range c.current.WBFTCommits.Values() {
			idx, _ := c.valSet.GetByAddress(msg.Source())
			if idx < 0 {
				continue
			}
			committedSeals[i] = wbft.SealData{
				Sealer: uint32(idx),
				Seal:   make([]byte, types.IstanbulExtraSeal),
			}
			commitMsg := msg.(*wbfmessage.Commit)
			copy(committedSeals[i].Seal[:], commitMsg.CommitSeal[:])
		}

		// Commit proposal to database
		if err := c.backend.Commit(proposal, preparedSeals, committedSeals, c.currentView().Round); err != nil {
			c.currentLogger(true, nil).Error("WBFT: error committing proposal", "err", err)
			c.broadcastNextRoundChange()
			return
		}
	}
}

// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
)

// addToExtraSeal adds a seal received after consensus to extraSeals.
func (c *Core) addToExtraSeal(msg wbfmessage.WBFTMessage) error {
	logger := c.currentLogger(true, msg)
	var (
		block    *types.Block
		ok       bool
		sealType SealType
		valSet   wbft.ValidatorSet
	)

	if c.state == StateAcceptRequest {
		block, ok = c.priorState.Proposal().(*types.Block)
		valSet = c.priorState.Validators()
	} else {
		block, ok = c.current.Proposal().(*types.Block)
		valSet = c.valSet
	}
	if !ok {
		// ignore if block is not found
		return nil
	}

	// validate seal
	if prepareMsg, ok := msg.(*wbfmessage.Prepare); ok {
		sealType = SealTypePrepare
		// Check digest
		if prepareMsg.Digest != block.Hash() {
			logger.Error("WBFT: invalid extra PREPARE message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(valSet, block.Header(), uint32(prepareMsg.CommonPayload.Round.Uint64()), sealType,
			prepareMsg.PrepareSeal, prepareMsg.Source()); err != nil {
			return err
		}

		c.addToPrepareExtraSeal(msg.(*wbfmessage.Prepare))
	} else if commitMsg, ok := msg.(*wbfmessage.Commit); ok {
		sealType = SealTypeCommit
		// Check digest
		if commitMsg.Digest != block.Hash() {
			logger.Error("WBFT: invalid extra COMMIT message digest")
			return errInvalidMessage
		}
		// verify msg seal is matched with msg digest and seal type
		if err := verifySeal(valSet, block.Header(), uint32(commitMsg.CommonPayload.Round.Uint64()), sealType,
			commitMsg.CommitSeal, commitMsg.Source()); err != nil {
			return err
		}

		c.addToCommitExtraSeal(msg.(*wbfmessage.Commit))
	} else {
		return errInvalidExtraSealMessage
	}
	return nil
}

func (c *Core) addToPrepareExtraSeal(prepareMsg *wbfmessage.Prepare) {
	logger := c.currentLogger(true, prepareMsg)

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	if c.prepareExtraSeals[prepareMsg.Source()] != nil {
		if existingView, incomingView := c.prepareExtraSeals[prepareMsg.Source()].View(), prepareMsg.View(); existingView.Cmp(&incomingView) >= 0 {
			return
		}
	}
	c.prepareExtraSeals[prepareMsg.Source()] = prepareMsg
	logger.Debug("WBFT: new extra prepare seal message")
}

func (c *Core) addToCommitExtraSeal(commitMsg *wbfmessage.Commit) {
	logger := c.currentLogger(true, commitMsg)

	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	if c.commitExtraSeals[commitMsg.Source()] != nil {
		if existingView, incomingView := c.commitExtraSeals[commitMsg.Source()].View(), commitMsg.View(); existingView.Cmp(&incomingView) >= 0 {
			return
		}
	}
	c.commitExtraSeals[commitMsg.Source()] = commitMsg
	logger.Debug("WBFT: new extra commit seal message")
}

// addEffectiveSealToExtraSeal adds a consensus-effective seal to extraSeals used during block creation.
func (c *Core) addEffectiveSealToExtraSeal() error {
	// c.current may be nil on the first function call after system boot.
	if c.current != nil {
		for _, m := range c.current.WBFTPrepares.Values() {
			c.addToPrepareExtraSeal(m.(*wbfmessage.Prepare))
		}
		for _, m := range c.current.WBFTCommits.Values() {
			c.addToCommitExtraSeal(m.(*wbfmessage.Commit))
		}
	}
	return nil
}

// ProcessExtraSeal collects prepare and commit messages that have been stored in extraSeal
// and pass it to backend preparing new block
func (c *Core) ProcessExtraSeal(lastProposal wbft.Proposal, priorRound *big.Int, valSet wbft.ValidatorSet) ([]wbft.SealData, []wbft.SealData) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()

	preparedSeal := make([]wbft.SealData, 0)
	committedSeal := make([]wbft.SealData, 0)
	// latestView is view to process extra seal
	latestView := wbft.View{
		Round:    priorRound,
		Sequence: lastProposal.Number(),
	}

	// process prepare seal
	for _, msg := range c.prepareExtraSeals {
		if msg != nil {
			view := msg.View()
			if latestView.Cmp(&view) == 0 && msg.Digest == lastProposal.Hash() {
				// this seal(c.prepareExtraSeals[addr]) is valid and re-usable for this sequence
				idx, _ := valSet.GetByAddress(msg.Source())
				if idx < 0 {
					continue
				}
				preparedSeal = append(preparedSeal, wbft.SealData{
					Sealer: uint32(idx),
					Seal:   append([]byte{}, msg.PrepareSeal...),
				})
			}
		}
	}

	// process commit seal
	for _, msg := range c.commitExtraSeals {
		if msg != nil {
			view := msg.View()
			if latestView.Cmp(&view) == 0 && msg.Digest == lastProposal.Hash() {
				// this seal(c.commitExtraSeals[addr]) is valid and re-usable for this sequence
				idx, _ := valSet.GetByAddress(msg.Source())
				if idx < 0 {
					continue
				}
				committedSeal = append(committedSeal, wbft.SealData{
					Sealer: uint32(idx),
					Seal:   append([]byte{}, msg.CommitSeal...),
				})
			}
		}
	}

	return preparedSeal, committedSeal
}

// Delete all extraSeals prior to the previous proposal
func (c *Core) ClearExtraSeals(lastNum *big.Int) {
	c.extraSealsMu.Lock()
	defer c.extraSealsMu.Unlock()
	// process prepare seal
	for addr, msg := range c.prepareExtraSeals {
		if msg != nil && msg.Sequence.Cmp(lastNum) < 0 {
			delete(c.prepareExtraSeals, addr) // erase invalid seal
		}
	}

	// process commit seal
	for addr, msg := range c.commitExtraSeals {
		if msg != nil && msg.Sequence.Cmp(lastNum) < 0 {
			delete(c.commitExtraSeals, addr) // erase invalid seal
		}
	}
}

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
// This file is derived from quorum/consensus/istanbul/wbft/core/roundstate.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbfmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
)

// newRoundState creates a new roundState instance with the given view and validatorSet
func newRoundState(view *wbft.View, validatorSet wbft.ValidatorSet, preprepare *wbfmessage.Preprepare, preparedRound *big.Int, preparedBlock wbft.Proposal, pendingRequest *Request, hasBadProposal func(hash common.Hash) bool) *roundState {
	return &roundState{
		round:      view.Round,
		sequence:   view.Sequence,
		Preprepare: preprepare,
		//Prepares:       newMessageSet(validatorSet),
		//Commits:        newMessageSet(validatorSet),
		WBFTPrepares:   newWBFTMsgSet(validatorSet),
		WBFTCommits:    newWBFTMsgSet(validatorSet),
		preparedRound:  preparedRound,
		preparedBlock:  preparedBlock,
		mu:             new(sync.RWMutex),
		pendingRequest: pendingRequest,
		hasBadProposal: hasBadProposal,
		preprepareSent: big.NewInt(0),
	}
}

// roundState stores the consensus state
type roundState struct {
	round      *big.Int
	sequence   *big.Int
	Preprepare *wbfmessage.Preprepare

	WBFTPrepares *wbftMsgSet
	WBFTCommits  *wbftMsgSet

	pendingRequest *Request
	preparedRound  *big.Int
	preparedBlock  wbft.Proposal

	mu             *sync.RWMutex
	hasBadProposal func(hash common.Hash) bool

	// Keep track of preprepare sent messages
	preprepareSent *big.Int
}

func (s *roundState) Subject() *Subject {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	return &Subject{
		View: &wbft.View{
			Round:    new(big.Int).Set(s.round),
			Sequence: new(big.Int).Set(s.sequence),
		},
		Digest: s.Preprepare.Proposal.Hash(),
	}
}

func (s *roundState) SetPreprepare(preprepare *wbfmessage.Preprepare) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Preprepare = preprepare
}

func (s *roundState) Proposal() wbft.Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Preprepare != nil {
		return s.Preprepare.Proposal
	}

	return nil
}

func (s *roundState) SetRound(r *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.round = new(big.Int).Set(r)
}

func (s *roundState) Round() *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.round
}

func (s *roundState) SetSequence(seq *big.Int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequence = seq
}

func (s *roundState) Sequence() *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sequence
}

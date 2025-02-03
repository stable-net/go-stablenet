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
// This file is derived from quorum/consensus/istanbul/qbft/core/types.go
// and quorum/consensus/istanbul/backend.go (2024.07.25).
// Modified and improved for the wemix development.

package core

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftmessage "github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/event"
)

type State uint64

const (
	StateAcceptRequest State = iota
	StatePreprepared
	StatePrepared
	StateCommitted
)

func (s State) String() string {
	if s == StateAcceptRequest {
		return "Accept request"
	} else if s == StatePreprepared {
		return "Preprepared"
	} else if s == StatePrepared {
		return "Prepared"
	} else if s == StateCommitted {
		return "Committed"
	} else {
		return "Unknown"
	}
}

// Cmp compares s and y and returns:
//
//	-1 if s is the previous state of y
//	 0 if s and y are the same state
//	+1 if s is the next state of y
func (s State) Cmp(y State) int {
	if uint64(s) < uint64(y) {
		return -1
	}
	if uint64(s) > uint64(y) {
		return 1
	}
	return 0
}

// Request is used to construct a Preprepare message
type Request struct {
	Proposal        qbft.Proposal
	RCMessages      *qbftMsgSet
	PrepareMessages []*qbftmessage.Prepare
}

// Subject represents the message sent when msgPrepare and msgCommit is broadcasted
type Subject struct {
	View   *qbft.View
	Digest common.Hash
}

func (b *Subject) String() string {
	return fmt.Sprintf("{View: %v, Proposal: %v}", b.View, b.Digest.String())
}

// ----------------------------------------------------------------------------

// Backend provides application specific functions for QBFT core
type Backend interface {
	// Address returns the owner's address
	Address() common.Address

	// Validators returns the validator set
	Validators(proposal qbft.Proposal) qbft.ValidatorSet

	// EventMux returns the event mux in backend
	EventMux() *event.TypeMux

	// Broadcast sends a message to all validators (include self)
	Broadcast(valSet qbft.ValidatorSet, code uint64, payload []byte) error

	// Gossip sends a message to all validators (exclude self)
	Gossip(valSet qbft.ValidatorSet, code uint64, payload []byte) error

	// Commit delivers an approved proposal to backend.
	// The delivered proposal will be put into blockchain.
	Commit(proposal qbft.Proposal, preparedSeals, committedSeals [][]byte, round *big.Int) error

	// Verify verifies the proposal. If a consensus.ErrFutureBlock error is returned,
	// the time difference of the proposal and current time is also returned.
	Verify(qbft.Proposal) (time.Duration, error)

	// Sign signs input data with the backend's private key
	Sign([]byte) ([]byte, error)

	// SignWithoutHashing sign input data with the backend's private key without hashing the input data
	SignWithoutHashing([]byte) ([]byte, error)

	// CheckSignature verifies the signature by checking if it's signed by
	// the given validator
	CheckSignature(data []byte, addr common.Address, sig []byte) error

	// LastProposal retrieves latest committed proposal and the address of proposer
	LastProposal() (qbft.Proposal, common.Address)

	// HasPropsal checks if the combination of the given hash and height matches any existing blocks
	HasPropsal(hash common.Hash, number *big.Int) bool

	// GetProposer returns the proposer of the given block height
	GetProposer(number uint64) common.Address

	// HasBadProposal returns whether the block with the hash is a bad block
	HasBadProposal(hash common.Hash) bool

	Close() error

	NotifyNewRound(round *big.Int)
}

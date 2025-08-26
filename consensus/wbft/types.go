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
// This file is derived from quorum/consensus/istanbul/types.go
// and quorum/consensus/istanbul/events.go
// and quorum/consensus/istanbul/validator.go (2024.07.25).
// Modified and improved for the wemix development.

package wbft

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Proposal supports retrieving height and serialized block to be used during Istanbul consensus.
type Proposal interface {
	// Number retrieves the sequence number of this proposal.
	Number() *big.Int

	// Hash retrieves the hash of this proposal.
	Hash() common.Hash

	EncodeRLP(w io.Writer) error

	DecodeRLP(s *rlp.Stream) error
}

var _ Proposal = &types.Block{}

type Request struct {
	Proposal Proposal
}

// View includes a round number and a sequence number.
// Sequence is the block number we'd like to commit.
// Each round has a number and is composed by 3 steps: preprepare, prepare and commit.
//
// If the given block is not accepted by validators, a round change will occur
// and the validators start a new round with round+1.
type View struct {
	Round    *big.Int
	Sequence *big.Int
}

// EncodeRLP serializes b into the Ethereum RLP format.
func (v *View) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{v.Round, v.Sequence})
}

// DecodeRLP implements rlp.Decoder, and load the consensus fields from a RLP stream.
func (v *View) DecodeRLP(s *rlp.Stream) error {
	var view struct {
		Round    *big.Int
		Sequence *big.Int
	}

	if err := s.Decode(&view); err != nil {
		return err
	}
	v.Round, v.Sequence = view.Round, view.Sequence
	return nil
}

func (v *View) String() string {
	return fmt.Sprintf("{Round: %d, Sequence: %d}", v.Round.Uint64(), v.Sequence.Uint64())
}

// Cmp compares v and y and returns:
//
//	-1 if v <  y
//	 0 if v == y
//	+1 if v >  y
func (v *View) Cmp(y *View) int {
	if v.Sequence.Cmp(y.Sequence) != 0 {
		return v.Sequence.Cmp(y.Sequence)
	}
	if v.Round.Cmp(y.Round) != 0 {
		return v.Round.Cmp(y.Round)
	}
	return 0
}

type Preprepare struct {
	View     *View
	Proposal Proposal
}

// EncodeRLP serializes b into the Ethereum RLP format.
func (b *Preprepare) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{b.View, b.Proposal})
}

// DecodeRLP implements rlp.Decoder, and load the consensus fields from a RLP stream.
func (b *Preprepare) DecodeRLP(s *rlp.Stream) error {
	var preprepare struct {
		View     *View
		Proposal *types.Block
	}

	if err := s.Decode(&preprepare); err != nil {
		return err
	}
	b.View, b.Proposal = preprepare.View, preprepare.Proposal

	return nil
}

type Subject struct {
	View   *View
	Digest common.Hash
}

// EncodeRLP serializes b into the Ethereum RLP format.
func (b *Subject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{b.View, b.Digest})
}

// DecodeRLP implements rlp.Decoder, and load the consensus fields from a RLP stream.
func (b *Subject) DecodeRLP(s *rlp.Stream) error {
	var subject struct {
		View   *View
		Digest common.Hash
	}

	if err := s.Decode(&subject); err != nil {
		return err
	}
	b.View, b.Digest = subject.View, subject.Digest
	return nil
}

func (b *Subject) String() string {
	return fmt.Sprintf("{View: %v, Digest: %v}", b.View, b.Digest.String())
}

// ----------------------------------------------------------------------------

type Validator interface {
	// Address returns address
	Address() common.Address

	BLSPublicKey() []byte

	// String representation of Validator
	String() string

	Copy() Validator
}

type Validators []Validator

func (vs validatorSorter) Len() int {
	return len(vs.validators)
}

func (vs validatorSorter) Swap(i, j int) {
	vs.validators[i], vs.validators[j] = vs.validators[j], vs.validators[i]
}

func (vs validatorSorter) Less(i, j int) bool {
	return vs.by(vs.validators[i], vs.validators[j])
}

type validatorSorter struct {
	validators Validators
	by         ValidatorSortByFunc
}

type ValidatorSortByFunc func(v1 Validator, v2 Validator) bool

func ValidatorSortByString() ValidatorSortByFunc {
	return func(v1 Validator, v2 Validator) bool {
		return strings.Compare(v1.String(), v2.String()) < 0
	}
}

func ValidatorSortByByte() ValidatorSortByFunc {
	return func(v1 Validator, v2 Validator) bool {
		return bytes.Compare(v1.Address().Bytes(), v2.Address().Bytes()) < 0
	}
}

func (by ValidatorSortByFunc) Sort(validators []Validator) {
	v := &validatorSorter{
		validators: validators,
		by:         by,
	}
	sort.Sort(v)
}

// ----------------------------------------------------------------------------

type ValidatorSet interface {
	// Calculate the proposer
	CalcProposer(lastProposer common.Address, round uint64)
	// Return the validator size
	Size() int
	// Return the validator array
	List() []Validator
	// Return the validator address array
	AddressList() []common.Address
	// Get validator by index
	GetByIndex(i uint64) Validator
	// Get validator by given address
	GetByAddress(addr common.Address) (int, Validator)
	// Get current proposer
	GetProposer() Validator
	// Check whether the validator with given address is a proposer
	IsProposer(address common.Address) bool
	// Add validator
	AddValidator(address common.Address, blsPublicKey []byte) bool
	// Remove validator
	RemoveValidator(address common.Address) bool
	// Copy validator set
	Copy() ValidatorSet
	// Get the maximum number of faulty nodes
	F() float64
	// Get quorum size ( which is n-f)
	QuorumSize() int
	// Get proposer policy
	Policy() ProposerPolicy
}

// ----------------------------------------------------------------------------

type ProposalSelector func(ValidatorSet, common.Address, uint64) Validator

// ----------------------------------------------------------------------------

// RequestEvent is posted to propose a proposal
type RequestEvent struct {
	Proposal Proposal
}

// MessageEvent is posted for Istanbul engine communication
type MessageEvent struct {
	Code    uint64
	Payload []byte
}

// FinalCommittedEvent is posted when a proposal is committed
type FinalCommittedEvent struct {
}

type SealData struct {
	Sealer uint32
	Seal   []byte
}

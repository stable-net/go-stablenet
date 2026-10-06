// Copyright 2026 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/consensus/wbft/messages"
)

// two64 is 2^64, the smallest value whose low 64 bits alias zero.
var two64 = new(big.Int).Lsh(common.Big1, 64)

// TestHandlePreprepareRejectsMismatchedProposalNumber verifies that a
// PRE-PREPARE whose proposal number differs from its sequence is rejected,
// including numbers that only match the sequence in their low 64 bits.
func TestHandlePreprepareRejectsMismatchedProposalNumber(t *testing.T) {
	const sequence = 100
	seq := big.NewInt(sequence)

	tests := []struct {
		name   string
		number *big.Int
	}{
		{"2^64 + sequence", new(big.Int).Add(two64, seq)},
		{"2^65 + sequence", new(big.Int).Add(new(big.Int).Lsh(two64, 1), seq)},
		{"sequence + 1", big.NewInt(sequence + 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The backend is nil: a rejected PRE-PREPARE must never reach Verify.
			c := makeCoreForTest(common.Big0, common.Big0, seq, nil)
			preprepare := messages.NewPreprepare(seq, common.Big0, makeProposal(tt.number))
			preprepare.SetSource(c.valSet.GetProposer().Address())

			err := c.handlePreprepareMsg(preprepare)
			if !errors.Is(err, errInvalidPreparedBlock) {
				t.Fatalf("handlePreprepareMsg(number=%s) = %v; want %v", tt.number, err, errInvalidPreparedBlock)
			}
			if c.state != StateAcceptRequest {
				t.Fatalf("state = %v; want %v", c.state, StateAcceptRequest)
			}
		})
	}
}

// TestCheckMessageRejectsOversizedView verifies that a view whose sequence or
// round exceeds 64 bits is never accepted, backlogged or kept as an extra
// seal. The big.Int comparisons in checkMessage already catch these values,
// so no separate IsUint64 guard is needed on the view.
func TestCheckMessageRejectsOversizedView(t *testing.T) {
	const sequence = 100
	seq := big.NewInt(sequence)

	tests := []struct {
		name string
		code uint64
		view wbft.View
		want error
	}{
		{
			name: "PRE-PREPARE with 2^64 + sequence",
			code: messages.PreprepareCode,
			view: wbft.View{Sequence: new(big.Int).Add(two64, seq), Round: common.Big0},
			want: errFutureViewTooFar,
		},
		{
			name: "PRE-PREPARE with round 2^64",
			code: messages.PreprepareCode,
			view: wbft.View{Sequence: seq, Round: two64},
			want: errFutureViewTooFar,
		},
		{
			name: "ROUND-CHANGE with round 2^64",
			code: messages.RoundChangeCode,
			view: wbft.View{Sequence: seq, Round: two64},
			want: errFutureViewTooFar,
		},
		{
			name: "COMMIT for the prior sequence with round 2^64",
			code: messages.CommitCode,
			view: wbft.View{Sequence: big.NewInt(sequence - 1), Round: two64},
			want: errOldMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := makeCoreForTest(common.Big0, common.Big0, seq, nil)
			if err := c.checkMessage(tt.code, &tt.view); !errors.Is(err, tt.want) {
				t.Fatalf("checkMessage(%v) = %v; want %v", tt.view, err, tt.want)
			}
		})
	}
}

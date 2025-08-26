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
//
// This file is derived from quorum/consensus/istanbul/qbft/types/commit.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// A WBFT COMMIT message.
type Commit struct {
	CommonPayload
	Digest     common.Hash
	CommitSeal []byte
}

func NewCommit(sequence *big.Int, round *big.Int, digest common.Hash, seal []byte) *Commit {
	return &Commit{
		CommonPayload: CommonPayload{
			code:     CommitCode,
			Sequence: sequence,
			Round:    round,
		},
		Digest:     digest,
		CommitSeal: seal,
	}
}

func (m *Commit) EncodePayloadForSigning() ([]byte, error) {
	return rlp.EncodeToBytes(
		[]interface{}{
			m.Code(),
			[]interface{}{m.Sequence, m.Round, m.Digest, m.CommitSeal},
		})
}

func (m *Commit) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		[]interface{}{m.Sequence, m.Round, m.Digest, m.CommitSeal},
		m.signature})
}

func (m *Commit) DecodeRLP(stream *rlp.Stream) error {
	var message struct {
		Payload struct {
			Sequence   *big.Int
			Round      *big.Int
			Digest     common.Hash
			CommitSeal []byte
		}
		Signature []byte
	}
	if err := stream.Decode(&message); err != nil {
		return err
	}
	m.code = CommitCode
	m.Sequence = message.Payload.Sequence
	m.Round = message.Payload.Round
	m.Digest = message.Payload.Digest
	m.CommitSeal = message.Payload.CommitSeal
	m.signature = message.Signature
	return nil
}

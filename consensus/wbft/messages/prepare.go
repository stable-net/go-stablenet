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
// This file is derived from quorum/consensus/istanbul/qbft/types/prepare.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// A WBFT PREPARE message.
type Prepare struct {
	CommonPayload
	Digest      common.Hash
	PrepareSeal []byte
}

func NewPrepare(sequence *big.Int, round *big.Int, digest common.Hash, seal []byte) *Prepare {
	return &Prepare{
		CommonPayload: CommonPayload{
			code:     PrepareCode,
			Sequence: sequence,
			Round:    round,
		},
		Digest:      digest,
		PrepareSeal: seal,
	}
}

func NewPrepareWithSigAndSource(sequence *big.Int, round *big.Int, digest common.Hash, signature []byte, source common.Address, seal []byte) *Prepare {
	prepare := NewPrepare(sequence, round, digest, seal)
	prepare.signature = signature
	prepare.source = source
	return prepare
}

func (p *Prepare) String() string {
	return fmt.Sprintf("Prepare {seq=%v, round=%v, digest=%v}", p.Sequence, p.Round, p.Digest.Hex())
}

func (p *Prepare) EncodePayloadForSigning() ([]byte, error) {
	return rlp.EncodeToBytes(
		[]interface{}{
			p.Code(),
			[]interface{}{p.Sequence, p.Round, p.Digest, p.PrepareSeal},
		})
}

func (p *Prepare) EncodeRLP(w io.Writer) error {
	return rlp.Encode(
		w,
		[]interface{}{
			[]interface{}{
				p.Sequence,
				p.Round,
				p.Digest,
				p.PrepareSeal},
			p.signature,
		})
}

func (p *Prepare) DecodeRLP(stream *rlp.Stream) error {
	var message struct {
		Payload struct {
			Sequence    *big.Int
			Round       *big.Int
			Digest      common.Hash
			PrepareSeal []byte
		}
		Signature []byte
	}
	if err := stream.Decode(&message); err != nil {
		return err
	}
	p.code = PrepareCode
	p.Sequence = message.Payload.Sequence
	p.Round = message.Payload.Round
	p.Digest = message.Payload.Digest
	p.PrepareSeal = message.Payload.PrepareSeal
	p.signature = message.Signature
	return nil
}

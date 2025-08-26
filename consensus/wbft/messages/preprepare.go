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
// This file is derived from quorum/consensus/istanbul/qbft/types/preprepare.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Preprepare struct {
	CommonPayload
	Proposal                  wbft.Proposal
	JustificationRoundChanges []*SignedRoundChangePayload
	JustificationPrepares     []*Prepare
}

func NewPreprepare(sequence *big.Int, round *big.Int, proposal wbft.Proposal) *Preprepare {
	return &Preprepare{
		CommonPayload: CommonPayload{
			code:     PreprepareCode,
			Sequence: sequence,
			Round:    round,
		},
		Proposal: proposal,
	}
}

func (m *Preprepare) EncodePayloadForSigning() ([]byte, error) {
	return rlp.EncodeToBytes(
		[]interface{}{
			m.Code(),
			[]interface{}{m.Sequence, m.Round, m.Proposal},
		})
}

func (m *Preprepare) EncodeRLP(w io.Writer) error {
	return rlp.Encode(
		w,
		[]interface{}{
			[]interface{}{
				[]interface{}{m.Sequence, m.Round, m.Proposal},
				m.signature,
			},
			[]interface{}{
				m.JustificationRoundChanges,
				m.JustificationPrepares,
			},
		})
}

func (m *Preprepare) DecodeRLP(stream *rlp.Stream) error {
	var message struct {
		SignedPayload struct {
			Payload struct {
				Sequence *big.Int
				Round    *big.Int
				Proposal *types.Block
			}
			Signature []byte
		}
		Justification struct {
			RoundChanges []rlp.RawValue
			Prepares     []*Prepare
		}
	}

	if err := stream.Decode(&message); err != nil {
		return fmt.Errorf("failed to decode preprepare: %w", err)
	}

	var decodedSignedRoundChange []*SignedRoundChangePayload
	for i, rawItem := range message.Justification.RoundChanges {
		var rc SignedRoundChangePayload
		if err := rlp.DecodeBytes(rawItem, &rc); err != nil {
			return fmt.Errorf("failed to decode SignedRoundChange[%d]: %w", i, err)
		}
		decodedSignedRoundChange = append(decodedSignedRoundChange, &rc)
	}

	m.code = PreprepareCode
	m.Sequence = message.SignedPayload.Payload.Sequence
	m.Round = message.SignedPayload.Payload.Round
	m.Proposal = message.SignedPayload.Payload.Proposal
	m.signature = message.SignedPayload.Signature
	m.JustificationPrepares = message.Justification.Prepares
	m.JustificationRoundChanges = decodedSignedRoundChange
	return nil
}

func (m *Preprepare) String() string {
	return fmt.Sprintf("code: %d, sequence: %d, round: %d, proposal: %v", m.code, m.Sequence, m.Round, m.Proposal.Hash().Hex())
}

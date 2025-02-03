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
// This file is derived from quorum/core/types/istanbul.go (2024.07.25).
// Modified and improved for the wemix development.

package types

import (
	"errors"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// ## Quorum QBFT START
var (
	IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
	IstanbulExtraSeal   = 65 // Fixed number of extra-data bytes reserved for validator seal

	// QBFTDefaultDifficulty is used to identify whether the block is from QBFT consensus engine.
	// we use this value on behalf of the role IstanbulDigest
	QBFTDefaultDifficulty = big.NewInt(1) // ## Wemix

	// ErrInvalidIstanbulHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidIstanbulHeaderExtra = errors.New("invalid qbft header extra-data")
)

// QBFTExtra represents header extradata for qbft protocol
type QBFTExtra struct {
	VanityData        []byte
	Validators        []common.Address
	PrevRound         uint32
	PrevPreparedSeal  [][]byte
	PrevCommittedSeal [][]byte // committedSeal of previous local block
	Round             uint32
	PreparedSeal      [][]byte
	CommittedSeal     [][]byte
}

// EncodeRLP serializes qist into the Ethereum RLP format.
func (qst *QBFTExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		qst.VanityData,
		qst.Validators,
		qst.PrevRound,
		qst.PrevPreparedSeal,
		qst.PrevCommittedSeal,
		qst.Round,
		qst.PreparedSeal,
		qst.CommittedSeal,
	})
}

// DecodeRLP implements rlp.Decoder, and load the QBFTExtra fields from a RLP stream.
func (qst *QBFTExtra) DecodeRLP(s *rlp.Stream) error {
	var qbftExtra struct {
		VanityData        []byte
		Validators        []common.Address
		PrevRound         uint32
		PrevPreparedSeal  [][]byte
		PrevCommittedSeal [][]byte
		Round             uint32
		PreparedSeal      [][]byte
		CommittedSeal     [][]byte
	}
	if err := s.Decode(&qbftExtra); err != nil {
		return err
	}

	qst.VanityData, qst.Validators, qst.PrevRound, qst.PrevPreparedSeal, qst.PrevCommittedSeal, qst.Round, qst.PreparedSeal, qst.CommittedSeal =
		qbftExtra.VanityData, qbftExtra.Validators, qbftExtra.PrevRound, qbftExtra.PrevPreparedSeal, qbftExtra.PrevCommittedSeal, qbftExtra.Round, qbftExtra.PreparedSeal, qbftExtra.CommittedSeal

	return nil
}

// ExtractQBFTExtra extracts all values of the QBFTExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func ExtractQBFTExtra(h *Header) (*QBFTExtra, error) {
	qbftExtra := new(QBFTExtra)
	err := rlp.DecodeBytes(h.Extra[:], qbftExtra)
	if err != nil {
		return nil, err
	}
	return qbftExtra, nil
}

// QBFTFilteredHeader returns a filtered header which some information (like committed seals, round)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func QBFTFilteredHeader(h *Header) *Header {
	return QBFTFilteredHeaderWithRound(h, 0)
}

// QBFTFilteredHeaderWithRound returns the copy of the header with round number set to the given round number
// and commit seal set to its null value
func QBFTFilteredHeaderWithRound(h *Header, round uint32) *Header {
	newHeader := CopyHeader(h)
	qbftExtra, err := ExtractQBFTExtra(newHeader)
	if err != nil {
		return nil
	}

	qbftExtra.PreparedSeal = [][]byte{}
	qbftExtra.CommittedSeal = [][]byte{}
	qbftExtra.Round = round

	payload, err := rlp.EncodeToBytes(&qbftExtra)
	if err != nil {
		return nil
	}

	newHeader.Extra = payload

	return newHeader
}

// ## Quorum QBFT END

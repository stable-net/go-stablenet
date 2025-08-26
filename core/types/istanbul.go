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
// This file is derived from quorum/core/types/istanbul.go (2024.07.25).
// Modified and improved for the wemix development.

package types

import (
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
	IstanbulExtraSeal   = 96 // Fixed number of extra-data bytes reserved for validator seal (BLS_SIGNATURE_LENGTH)

	// WBFTDefaultDifficulty is used to identify whether the block is from WBFT consensus engine.
	// we use this value on behalf of the role IstanbulDigest
	WBFTDefaultDifficulty = big.NewInt(1) // ## Wemix

	// Diligence is used to choose validators for next epoch
	// Diligence has maximum value of 2 * DiligenceDenominator.
	DiligenceDenominator = uint64(1_000_000)

	// DefaultDiligence is 95% of maximum diligence.
	DefaultDiligence = 2 * DiligenceDenominator * 95 / 100

	// ErrInvalidIstanbulHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidIstanbulHeaderExtra = errors.New("invalid wbft header extra-data")
)

type WBFTAggregatedSeal struct {
	Sealers   SealerSet
	Signature []byte
}

func (as *WBFTAggregatedSeal) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		as.Sealers,
		as.Signature,
	})
}

func (as *WBFTAggregatedSeal) DecodeRLP(s *rlp.Stream) error {
	var aggregatedSeal struct {
		Sealers   SealerSet
		Signature []byte
	}
	if err := s.Decode(&aggregatedSeal); err != nil {
		return err
	}
	as.Sealers, as.Signature = aggregatedSeal.Sealers, aggregatedSeal.Signature
	return nil
}

func (as *WBFTAggregatedSeal) String() string {
	return fmt.Sprintf("{sealers: %v, signature: %x}", as.Sealers, as.Signature)
}

// WBFTExtra represents header extradata for wbft protocol
type WBFTExtra struct {
	VanityData        []byte
	RandaoReveal      []byte // bls signature of the block number
	PrevRound         uint32
	PrevPreparedSeal  *WBFTAggregatedSeal
	PrevCommittedSeal *WBFTAggregatedSeal // committedSeal of previous local block
	Round             uint32
	PreparedSeal      *WBFTAggregatedSeal
	CommittedSeal     *WBFTAggregatedSeal
	EpochInfo         *EpochInfo // epoch info is filled only for last block of epoch
}

type Staker struct {
	Addr      common.Address
	Diligence uint64 // unit: 10^-6
}

type EpochInfo struct {
	Stakers       []*Staker // staker list for next epoch (staker index may be changed for each epoch)
	Validators    []uint32  // validator list for next epoch (using indices of staker list)
	BLSPublicKeys [][]byte  // bls public key list for next epoch
	Stabilizing   bool      // initial epochs are stabilizing epochs, which means that the stakers are less than `stabilizingStakersThreshold`
}

// EncodeRLP serializes qist into the Ethereum RLP format.
func (qst *WBFTExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		qst.VanityData,
		qst.RandaoReveal,
		qst.PrevRound,
		qst.PrevPreparedSeal,
		qst.PrevCommittedSeal,
		qst.Round,
		qst.PreparedSeal,
		qst.CommittedSeal,
		qst.EpochInfo,
	})
}

// DecodeRLP implements rlp.Decoder, and load the WBFTExtra fields from a RLP stream.
func (qst *WBFTExtra) DecodeRLP(s *rlp.Stream) error {
	var wbftExtra struct {
		VanityData        []byte
		RandaoReveal      []byte
		PrevRound         uint32
		PrevPreparedSeal  *WBFTAggregatedSeal `rlp:"nil"`
		PrevCommittedSeal *WBFTAggregatedSeal `rlp:"nil"`
		Round             uint32
		PreparedSeal      *WBFTAggregatedSeal `rlp:"nil"`
		CommittedSeal     *WBFTAggregatedSeal `rlp:"nil"`
		EpochInfo         *EpochInfo          `rlp:"nil"`
	}
	if err := s.Decode(&wbftExtra); err != nil {
		return err
	}

	qst.VanityData, qst.RandaoReveal, qst.PrevRound, qst.PrevPreparedSeal, qst.PrevCommittedSeal, qst.Round, qst.PreparedSeal, qst.CommittedSeal, qst.EpochInfo =
		wbftExtra.VanityData, wbftExtra.RandaoReveal, wbftExtra.PrevRound, wbftExtra.PrevPreparedSeal, wbftExtra.PrevCommittedSeal, wbftExtra.Round, wbftExtra.PreparedSeal, wbftExtra.CommittedSeal, wbftExtra.EpochInfo

	return nil
}

func (ei *EpochInfo) GetStakers() []common.Address {
	if ei == nil {
		return nil
	}

	l := make([]common.Address, len(ei.Stakers))
	for i, staker := range ei.Stakers {
		l[i] = staker.Addr
	}
	return l
}

func (ei *EpochInfo) FindStakerByAddress(addr common.Address) (uint32, *Staker) {
	if ei == nil {
		return 0, nil
	}

	for i, staker := range ei.Stakers {
		if staker.Addr == addr {
			return uint32(i), staker
		}
	}

	return uint32(len(ei.Stakers)), nil
}

func (ei *EpochInfo) GetValidators() []common.Address {
	if ei == nil {
		return nil
	}

	l := make([]common.Address, len(ei.Validators))
	for i, validator := range ei.Validators {
		l[i] = ei.GetValidator(validator)
	}
	return l
}

func (ei *EpochInfo) GetValidator(index uint32) common.Address {
	if ei == nil || int(index) >= len(ei.Stakers) {
		return common.Address{}
	}
	return ei.Stakers[index].Addr
}

func (ei *EpochInfo) GetValidatorIndexMap() map[common.Address]uint32 {
	validatorIndexMap := make(map[common.Address]uint32)
	for _, idx := range ei.Validators {
		validatorIndexMap[ei.GetValidator(idx)] = idx
	}
	return validatorIndexMap
}

// EncodeRLP serializes epochInfo into the Ethereum RLP format.
func (ei *EpochInfo) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		ei.Stakers,
		ei.Validators,
		ei.BLSPublicKeys,
		ei.Stabilizing,
	})
}

// DecodeRLP implements rlp.Decoder, and load the EpochInfo fields from a RLP stream.
func (ei *EpochInfo) DecodeRLP(s *rlp.Stream) error {
	var epochInfo struct {
		Stakers       []*Staker
		Validators    []uint32
		BLSPublicKeys [][]byte
		Stabilizing   bool
	}
	if err := s.Decode(&epochInfo); err != nil {
		return err
	}
	ei.Stakers, ei.Validators, ei.BLSPublicKeys, ei.Stabilizing = epochInfo.Stakers, epochInfo.Validators, epochInfo.BLSPublicKeys, epochInfo.Stabilizing
	return nil
}

// EncodeRLP serializes Staker into the Ethereum RLP format.
func (stkr *Staker) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		stkr.Addr,
		stkr.Diligence,
	})
}

// DecodeRLP implements rlp.Decoder, and load the Staker fields from a RLP stream.
func (stkr *Staker) DecodeRLP(s *rlp.Stream) error {
	var staker struct {
		Addr      common.Address
		Diligence uint64
	}
	if err := s.Decode(&staker); err != nil {
		return err
	}
	stkr.Addr, stkr.Diligence = staker.Addr, staker.Diligence
	return nil
}

// ExtractWBFTExtra extracts all values of the WBFTExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func ExtractWBFTExtra(h *Header) (*WBFTExtra, error) {
	wbftExtra := new(WBFTExtra)
	err := rlp.DecodeBytes(h.Extra[:], wbftExtra)
	if err != nil {
		return nil, err
	}
	return wbftExtra, nil
}

// WBFTFilteredHeader returns a filtered header which some information (like committed seals, round)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func WBFTFilteredHeader(h *Header) *Header {
	return WBFTFilteredHeaderWithRound(h, 0)
}

// WBFTFilteredHeaderWithRound returns the copy of the header with round number set to the given round number
// and commit seal set to its null value
func WBFTFilteredHeaderWithRound(h *Header, round uint32) *Header {
	newHeader := CopyHeader(h)
	wbftExtra, err := ExtractWBFTExtra(newHeader)
	if err != nil {
		return nil
	}

	wbftExtra.PreparedSeal = nil
	wbftExtra.CommittedSeal = nil
	wbftExtra.Round = round

	payload, err := rlp.EncodeToBytes(&wbftExtra)
	if err != nil {
		return nil
	}

	newHeader.Extra = payload

	return newHeader
}

type SealerSet []byte

func (s *SealerSet) SetSealer(index uint32) {
	byteIndex := int(index / 8)
	if len(*s) <= byteIndex {
		*s = append(*s, make([]byte, byteIndex+1-len(*s))...)
	}
	(*s)[byteIndex] |= 1 << (index % 8)
}

func (s SealerSet) ClearSealer(index uint32) {
	byteIndex := int(index / 8)
	if byteIndex < len(s) {
		s[byteIndex] &^= 1 << (index % 8)
	}
}

func (s SealerSet) IsSealer(index uint32) bool {
	byteIndex := int(index / 8)
	if byteIndex >= len(s) {
		return false
	}
	return ((s)[byteIndex] & (1 << (index % 8))) != 0
}

func (s SealerSet) GetSealers() []uint32 {
	sealers := make([]uint32, 0)
	for byteIndex := 0; byteIndex < len(s); byteIndex++ {
		for bitOffset := 0; bitOffset < 8; bitOffset++ {
			if s[byteIndex]&(1<<bitOffset) != 0 {
				sealers = append(sealers, uint32(byteIndex*8+bitOffset))
			}
		}
	}
	return sealers
}

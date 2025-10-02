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
// This file is derived from quorum/consensus/istanbul/qbft/types/message.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
)

// WBFT message codes
const (
	PreprepareCode  = 0x12
	PrepareCode     = 0x13
	CommitCode      = 0x14
	RoundChangeCode = 0x15
)

// A set containing the messages codes for all WBFT messages.
func MessageCodes() map[uint64]struct{} {
	return map[uint64]struct{}{
		PreprepareCode:  {},
		PrepareCode:     {},
		CommitCode:      {},
		RoundChangeCode: {},
	}
}

// Common interface for all WBFT messages
type WBFTMessage interface {
	Code() uint64
	View() wbft.View
	Source() common.Address
	SetSource(address common.Address)
	EncodePayloadForSigning() ([]byte, error)
	Signature() []byte
	SetSignature(signature []byte)
}

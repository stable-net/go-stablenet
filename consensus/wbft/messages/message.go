// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/wbft/types/message.go (2024.07.25).
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

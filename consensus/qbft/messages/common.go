// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/qbft/types/common.go (2024.07.25).
// Modified and improved for the wemix development.

package messages

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
)

// Data that is common to all QBFT messages. Used for composition.
type CommonPayload struct {
	code      uint64
	source    common.Address
	Sequence  *big.Int
	Round     *big.Int
	signature []byte
}

func (m *CommonPayload) Code() uint64 {
	return m.code
}

func (m *CommonPayload) Source() common.Address {
	return m.source
}

func (m *CommonPayload) SetSource(address common.Address) {
	m.source = address
}

func (m *CommonPayload) View() qbft.View {
	return qbft.View{Sequence: m.Sequence, Round: m.Round}
}

func (m *CommonPayload) Signature() []byte {
	return m.signature
}

func (m *CommonPayload) SetSignature(signature []byte) {
	m.signature = signature
}

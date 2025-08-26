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

package eth

import (
	"errors"

	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// Quorum: quorum_protocol enables the eth service to return two different protocols, one for the eth mainnet "eth" service,
//         and one for the quorum specific consensus algo, obtained from engine.consensus
//         2021 Jan in the future consensus (istanbul) may run from its own service and use a single subprotocol there,
//         instead of overloading the eth service.

var (
	// errEthPeerNil is returned when no eth peer is found to be associated with a p2p peer.
	errEthPeerNil           = errors.New("eth peer was nil")
	errEthPeerNotRegistered = errors.New("eth peer was not registered")
)

const (
	Istanbul100 = 100
	// quorum consensus Protocol variables are optionally set in addition to the "eth" protocol variables (eth/protocol.go).
	quorumConsensusProtocolName = "istanbul"
)

func (s *Ethereum) quorumConsensusProtocols(backend eth.Backend, network uint64, dnsdisc enode.Iterator) []p2p.Protocol {
	// Set protocol Name/Version
	// keep `var protocolName = "eth"` as is, and only update the quorum consensus specific protocol
	// This is used to enable the eth service to return multiple devp2p subprotocols.
	// With this change, support is added so that the "eth" subprotocol remains and optionally a consensus subprotocol
	// can be added allowing the node to communicate over "eth" and an optional consensus subprotocol, e.g. "eth" and "istanbul/100"

	// ProtocolVersions are the supported versions of the quorum consensus protocol (first is primary), e.g. []uint{Istanbul100}.
	quorumConsensusProtocolVersions := []uint{Istanbul100}
	// protocol Length describe the number of messages support by the protocol/version map[uint]uint64{Istanbul100: 18}
	quorumConsensusProtocolLengths := map[uint]uint64{Istanbul100: 22}

	protos := make([]p2p.Protocol, len(quorumConsensusProtocolVersions))
	for i, vsn := range quorumConsensusProtocolVersions {
		// if we have a legacy protocol, e.g. istanbul/99, istanbul/64 then the protocol handler is will be the "eth"
		// protocol handler, and the subprotocol "eth" will not be used, but rather the legacy subprotocol will handle
		// both eth messages and consensus messages.
		length, ok := quorumConsensusProtocolLengths[vsn]
		if !ok {
			panic("makeQuorumConsensusProtocol for unknown version")
		}
		protos[i] = s.handler.makeQuorumConsensusProtocol(quorumConsensusProtocolName, vsn, length, backend, network, dnsdisc)
	}
	return protos
}

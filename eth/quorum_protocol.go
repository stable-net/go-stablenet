package eth

import (
	"errors"

	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// ## Quorum QBFT START
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
	// Previously, for istanbul/64 istnbul/99 and clique (v2.6) `protocolName` would be overridden and
	// set to the consensus subprotocol name instead of "eth", meaning the node would no longer
	// communicate over the "eth" subprotocol, e.g. "eth" or "istanbul/99" but not eth" and "istanbul/99".
	// With this change, support is added so that the "eth" subprotocol remains and optionally a consensus subprotocol
	// can be added allowing the node to communicate over "eth" and an optional consensus subprotocol, e.g. "eth" and "istanbul/100"

	// ProtocolVersions are the supported versions of the quorum consensus protocol (first is primary), e.g. []uint{Istanbul64, Istanbul99, Istanbul100}.
	quorumConsensusProtocolVersions := []uint{Istanbul100}
	// protocol Length describe the number of messages support by the protocol/version map[uint]uint64{Istanbul64: 18, Istanbul99: 18, Istanbul100: 18}
	quorumConsensusProtocolLengths := map[uint]uint64{Istanbul100: 22}

	protos := make([]p2p.Protocol, len(quorumConsensusProtocolVersions))
	for i, vsn := range quorumConsensusProtocolVersions {
		// if we have a legacy protocol, e.g. istanbul/99, istanbul/64 then the protocol handler is will be the "eth"
		// protocol handler, and the subprotocol "eth" will not be used, but rather the legacy subprotocol will handle
		// both eth messages and consensus messages.
		if isLegacyProtocol(quorumConsensusProtocolName, vsn) {
			length, ok := quorumConsensusProtocolLengths[vsn]
			if !ok {
				panic("makeProtocol for unknown version")
			}
			lp := s.handler.makeLegacyProtocol(quorumConsensusProtocolName, vsn, length, backend, network, dnsdisc)
			protos[i] = lp
		} else {
			length, ok := quorumConsensusProtocolLengths[vsn]
			if !ok {
				panic("makeQuorumConsensusProtocol for unknown version")
			}
			protos[i] = s.handler.makeQuorumConsensusProtocol(quorumConsensusProtocolName, vsn, length, backend, network, dnsdisc)
		}
	}
	return protos
}

// istanbul/64, istanbul/99, clique/63, clique/64 all override the "eth" subprotocol.
func isLegacyProtocol(name string, version uint) bool {
	// protocols that override "eth" subprotocol and run only the quorum subprotocol.
	quorumLegacyProtocols := map[string][]uint{"istanbul": {64, 99}, "clique": {63, 64}}
	for lpName, lpVersions := range quorumLegacyProtocols {
		if lpName == name {
			for _, v := range lpVersions {
				if v == version {
					return true
				}
			}
		}
	}
	return false
}

// ## Quorum QBFT END

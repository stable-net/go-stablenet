// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/eth/protocols/eth/qlight_deps.go (2024.07.25).
// Modified and improved for the wemix development.

package eth

import (
	"github.com/ethereum/go-ethereum/core"
)

// ## Quorum QBFT START
func CurrentENREntry(chain *core.BlockChain) *enrEntry {
	return currentENREntry(chain)
}

func NodeInfoFunc(chain *core.BlockChain, network uint64) *NodeInfo {
	return nodeInfo(chain, network)
}

// ## Quorum QBFT END

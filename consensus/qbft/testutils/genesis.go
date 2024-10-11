// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/testutils/genesis.go (2024.07.25).
// Modified and improved for the wemix development.

package testutils

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// ## Wemix QBFT START
// 1. remove ibft engine related test code
// ## Wemix QBFT END

func Genesis(validators []common.Address) *core.Genesis {
	// generate genesis block
	genesis := core.TestGenesisBlock()
	genesis.Config = params.TestChainConfig
	genesis.Config.Ethash = nil
	genesis.Difficulty = types.QBFTDefaultDifficulty
	genesis.Nonce = qbftcommon.EmptyBlockNonce.Uint64()

	appendValidators(genesis, validators)

	return genesis
}

func GenesisAndKeys(n int) (*core.Genesis, []*ecdsa.PrivateKey) {
	// Setup validators
	var nodeKeys = make([]*ecdsa.PrivateKey, n)
	var addrs = make([]common.Address, n)
	for i := 0; i < n; i++ {
		nodeKeys[i], _ = crypto.GenerateKey()
		addrs[i] = crypto.PubkeyToAddress(nodeKeys[i].PublicKey)
	}

	// generate genesis block
	genesis := Genesis(addrs)

	return genesis, nodeKeys
}

func appendValidators(genesis *core.Genesis, addrs []common.Address) {
	vanity := append(genesis.ExtraData, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(genesis.ExtraData))...)
	ist := &types.QBFTExtra{
		VanityData:    vanity,
		Validators:    addrs,
		Vote:          nil,
		CommittedSeal: [][]byte{},
		Round:         0,
	}

	istPayload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		panic("failed to encode qbft extra")
	}
	genesis.ExtraData = istPayload
}

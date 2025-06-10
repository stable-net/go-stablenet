// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/testutils/genesis.go (2024.07.25).
// Modified and improved for the wemix development.

package testutils

import (
	"bytes"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// ## Wemix QBFT START
// 1. remove ibft engine related test code
// ## Wemix QBFT END

func GenesisWithSeals(validators []common.Address, blsPublicKeys [][]byte) *core.Genesis {
	// generate genesis block
	genesis := core.DefaultGenesisBlock()
	genesis.Config = params.TestQBFTChainConfig
	genesis.Config.Ethash = nil
	genesis.Difficulty = types.QBFTDefaultDifficulty
	genesis.Nonce = qbftcommon.EmptyBlockNonce.Uint64()

	appendValidatorsAndPrevSeals(genesis, validators, blsPublicKeys)

	return genesis
}

func Genesis(validators []common.Address, blsPublicKeys [][]byte) *core.Genesis {
	// generate genesis block
	genesis := core.TestGenesisBlock()
	genesis.Config = params.TestQBFTChainConfig
	genesis.Config.Ethash = nil
	genesis.Difficulty = types.QBFTDefaultDifficulty
	genesis.Nonce = qbftcommon.EmptyBlockNonce.Uint64()

	_ = core.InjectContracts(genesis, genesis.Config)

	appendValidators(genesis, validators, blsPublicKeys)

	return genesis
}

func GenesisAndKeys(n int) (*core.Genesis, []*ecdsa.PrivateKey, []common.Address) {
	// Setup validators
	var nodeKeys = make([]*ecdsa.PrivateKey, n)
	var addrs = make([]common.Address, n)
	var blsPubKeys = make([][]byte, n)
	for i := 0; i < n; i++ {
		nodeKeys[i], _ = crypto.GenerateKey()
		blsKey, _ := bls.DeriveFromECDSA(nodeKeys[i])
		addrs[i] = crypto.PubkeyToAddress(nodeKeys[i].PublicKey)
		blsPubKeys[i] = blsKey.PublicKey().Marshal()
	}

	// generate genesis block
	genesis := Genesis(addrs, blsPubKeys)

	return genesis, nodeKeys, addrs
}

func GenesisAndFixedKeys(n int) (*core.Genesis, []*ecdsa.PrivateKey, []common.Address) {
	// Setup validators
	var nodeKeys = make([]*ecdsa.PrivateKey, n)
	var addrs = make([]common.Address, n)
	var blsPubKeys = make([][]byte, n)
	// use fixed keys for testing
	// generated addresses are:
	// 0: 0xB9Dd267FDb07316f89De27Ed37Ae010525B728Fc
	// 1: 0x209b41AA27e00828C33564DCB3339B8FF7F49304
	// 2: 0x2ddA32341F88F502Dbfb4854dcf66e88aCc2B4b3
	// 3: 0xAf6D46d1E55AA87772Fb1538FE4d36AAA70f4e06
	for i := 0; i < n; i++ {
		if i == 0 {
			nodeKeys[0], _ = crypto.HexToECDSA("e478a31539810867949701d8a78835451c38b5fe84045f2d7b1b0e2c1f1e0d0a")
		} else if i == 1 {
			nodeKeys[1], _ = crypto.HexToECDSA("2b7e151628aed2a6abf7158809cf4f3c7a57928f9e758f9c7e44106c9b2938b9")
		} else if i == 2 {
			nodeKeys[2], _ = crypto.HexToECDSA("8f7d38a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9")
		} else if i == 3 {
			nodeKeys[3], _ = crypto.HexToECDSA("1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b")
		} else if i > 3 {
			nodeKeys[i], _ = crypto.GenerateKey()
		}
		blsKey, _ := bls.DeriveFromECDSA(nodeKeys[i])
		addrs[i] = crypto.PubkeyToAddress(nodeKeys[i].PublicKey)
		blsPubKeys[i] = blsKey.PublicKey().Marshal()
	}

	// generate genesis block
	genesis := Genesis(addrs, blsPubKeys)

	return genesis, nodeKeys, addrs
}

func appendValidators(genesis *core.Genesis, addrs []common.Address, blsPublicKeys [][]byte) {
	setQBFTExtra(genesis, addrs, blsPublicKeys, false)
}

func appendValidatorsAndPrevSeals(genesis *core.Genesis, validators []common.Address, blsPublicKeys [][]byte) {
	setQBFTExtra(genesis, validators, blsPublicKeys, true)
}

func setQBFTExtra(genesis *core.Genesis, validators []common.Address, blsPublicKeys [][]byte, withPrev bool) {
	vanity := append(genesis.ExtraData, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(genesis.ExtraData))...)

	epochInfo := new(types.EpochInfo)
	blsPubKeys := make([]string, len(validators))
	for i, addr := range validators {
		epochInfo.Stakers = append(epochInfo.Stakers, &types.Staker{
			Addr:      addr,
			Diligence: types.DefaultDiligence,
		})
		epochInfo.Validators = append(epochInfo.Validators, uint32(i))
		epochInfo.BLSPublicKeys = append(epochInfo.BLSPublicKeys, blsPublicKeys[i])
		blsPubKeys[i] = hexutil.Encode(blsPublicKeys[i])
	}
	epochInfo.Stabilizing = true
	ist := &types.QBFTExtra{
		VanityData:    vanity,
		Round:         0,
		PreparedSeal:  &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		CommittedSeal: &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}},
		EpochInfo:     epochInfo,
	}

	if withPrev {
		ist.PrevRound = 0
		ist.PrevPreparedSeal = &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}}
		ist.PrevCommittedSeal = &types.QBFTAggregatedSeal{Signature: []byte{}, Sealers: types.SealerSet{}}
	}

	istPayload, err := rlp.EncodeToBytes(&ist)
	if err != nil {
		panic("failed to encode qbft extra")
	}
	genesis.ExtraData = istPayload
}

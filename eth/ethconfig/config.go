// Modification Copyright 2024 The Wemix Authors
// Copyright 2021 The go-ethereum Authors
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

// Package ethconfig contains the configuration of the ETH and LES protocols.
package ethconfig

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftBackend "github.com/ethereum/go-ethereum/consensus/wbft/backend"
	"github.com/ethereum/go-ethereum/consensus/wemix"
	"github.com/ethereum/go-ethereum/consensus/wpoa"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool/blobpool"
	"github.com/ethereum/go-ethereum/core/txpool/legacypool"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/miner"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/wemixgov"
)

// FullNodeGPO contains default gasprice oracle settings for full node.
var FullNodeGPO = gasprice.Config{
	Blocks:           20,
	Percentile:       60,
	MaxHeaderHistory: 1024,
	MaxBlockHistory:  1024,
	MaxPrice:         gasprice.DefaultMaxPrice,
	IgnorePrice:      gasprice.DefaultIgnorePrice,
}

// Defaults contains default settings for use on the Ethereum main net.
var Defaults = Config{
	//SyncMode: downloader.SnapSync,
	// make full sync the default sync mode in wbft (as opposed to upstream geth)
	SyncMode:       downloader.FullSync,
	ForceSyncCycle: common.Duration(10 * time.Second), // Time interval to force syncs, even if few peers are available
	TdSyncInterval: common.Duration(10 * time.Second), // Time interval to verify TD changes and detect sync stalling

	NetworkId:          0, // enable auto configuration of networkID == chainID
	TxLookupLimit:      31536000,
	TransactionHistory: 31536000,
	StateHistory:       params.FullImmutabilityThreshold,
	LightPeers:         100,
	DatabaseCache:      512,
	TrieCleanCache:     154,
	TrieDirtyCache:     256,
	TrieTimeout:        common.Duration(60 * time.Minute),
	SnapshotCache:      102,
	FilterLogCacheSize: 32,
	Miner:              miner.DefaultConfig,
	TxPool:             legacypool.DefaultConfig,
	BlobPool:           blobpool.DefaultConfig,
	RPCGasCap:          50000000,
	RPCEVMTimeout:      common.Duration(5 * time.Second),
	GPO:                FullNodeGPO,
	RPCTxFeeCap:        1, // 1 ether
}

//go:generate go run github.com/fjl/gencodec -type Config -formats toml -out gen_config.go

// Config contains configuration options for ETH and LES protocols.
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ethereum main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Network ID separates blockchains on the peer-to-peer networking level. When left
	// zero, the chain ID is used as network ID.
	NetworkId uint64
	SyncMode  downloader.SyncMode

	ForceSyncCycle common.Duration `toml:"ForceSyncCycle"`
	TdSyncInterval common.Duration `toml:"TdSyncInterval"`

	// This can be set to list of enrtree:// URLs which will be queried for
	// for nodes to connect to.
	EthDiscoveryURLs  []string
	SnapDiscoveryURLs []string

	NoPruning  bool // Whether to disable pruning and flush everything to disk
	NoPrefetch bool // Whether to disable prefetching and only load state on demand

	// Deprecated, use 'TransactionHistory' instead.
	TxLookupLimit      uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	TransactionHistory uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.
	StateHistory       uint64 `toml:",omitempty"` // The maximum number of blocks from head whose state histories are reserved.

	// State scheme represents the scheme used to store ethereum states and trie
	// nodes on top. It can be 'hash', 'path', or none which means use the scheme
	// consistent with persistent state.
	StateScheme string `toml:",omitempty"`

	// RequiredBlocks is a set of block number -> hash mappings which must be in the
	// canonical chain of all remote peers. Setting the option makes geth verify the
	// presence of these blocks for every new peer connection.
	RequiredBlocks map[uint64]common.Hash `toml:"-"`

	// Light client options
	LightServ        int  `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightIngress     int  `toml:",omitempty"` // Incoming bandwidth limit for light servers
	LightEgress      int  `toml:",omitempty"` // Outgoing bandwidth limit for light servers
	LightPeers       int  `toml:",omitempty"` // Maximum number of LES client peers
	LightNoPrune     bool `toml:",omitempty"` // Whether to disable light chain pruning
	LightNoSyncServe bool `toml:",omitempty"` // Whether to serve light clients before syncing

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	DatabaseFreezer    string

	TrieCleanCache int
	TrieDirtyCache int
	TrieTimeout    common.Duration `toml:"TrieTimeout"`
	SnapshotCache  int
	Preimages      bool

	// This is the number of blocks for which logs will be cached in the filter system.
	FilterLogCacheSize int

	// Mining options
	Miner miner.Config

	// Transaction pool options
	TxPool   legacypool.Config
	BlobPool blobpool.Config

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// RPCGasCap is the global gas cap for eth-call variants.
	RPCGasCap uint64

	// RPCEVMTimeout is the global timeout for eth-call.
	RPCEVMTimeout common.Duration `toml:"RPCEVMTimeout"`

	// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	// send-transaction variants. The unit is ether.
	RPCTxFeeCap float64

	// OverrideCancun (TODO: remove after the fork)
	OverrideCancun *uint64 `toml:",omitempty"`

	// OverrideVerkle (TODO: remove after the fork)
	OverrideVerkle *uint64 `toml:",omitempty"`
}

// CreateConsensusEngine creates a consensus engine for the given chain config.
// Clique is allowed for now to live standalone, but ethash is forbidden and can
// only exist on already merged networks.
func CreateConsensusEngine(govCli wemixgov.GovBackend, config *params.ChainConfig, privKey *ecdsa.PrivateKey, db ethdb.Database) (consensus.Engine, error) {
	// If proof-of-authority is requested, set it up
	if config.Clique != nil {
		if config.TerminalTotalDifficulty == nil {
			// clique engine without supporting beacon logic
			return clique.New(config.Clique, db), nil
		}
		return beacon.New(clique.New(config.Clique, db)), nil
	}

	if config.CroissantEnabled() {
		wbftCfg := new(wbft.Config)
		err := SetConfigFromChainConfig(wbftCfg, config)

		if err != nil {
			return nil, err
		}

		if config.CroissantBlock.Cmp(common.Big1) >= 0 {
			return wemix.NewCroissantEngine(wpoa.NewWemixPoAEngine(govCli), wbftCfg, privKey, db), nil
		}
		return wbftBackend.New(wbftCfg, privKey, db), nil
	}

	// If defaulting to proof-of-work, enforce an already merged network since
	// we cannot run PoW algorithms anymore, so we cannot even follow a chain
	// not coordinated by a beacon node.
	if !config.TerminalTotalDifficultyPassed {
		// no beacon and pure ethash faker
		return ethash.NewFaker(), nil
	}
	return beacon.New(ethash.NewFaker()), nil
}

func SetConfigFromChainConfig(wbftCfg *wbft.Config, chainCfg *params.ChainConfig) error {
	config := chainCfg.Croissant.WBFT
	if config.RequestTimeoutSeconds != 0 {
		wbftCfg.RequestTimeout = config.RequestTimeoutSeconds * 1000
	}
	if config.BlockPeriodSeconds != 0 {
		wbftCfg.BlockPeriod = config.BlockPeriodSeconds
	}
	if config.EpochLength != 0 {
		wbftCfg.Epoch = config.EpochLength
	}
	if config.AllowedFutureBlockTime != 0 {
		wbftCfg.AllowedFutureBlockTime = config.AllowedFutureBlockTime
	}
	wbftCfg.BlockReward = config.BlockReward
	wbftCfg.BlockRewardBeneficiary = config.BlockRewardBeneficiary

	if config.ProposerPolicy != nil {
		wbftCfg.ProposerPolicy = wbft.NewProposerPolicy(wbft.ProposerPolicyId(*config.ProposerPolicy))
	}
	if config.TargetValidators != nil {
		wbftCfg.TargetValidators = *config.TargetValidators
	}
	if config.MaxRequestTimeoutSeconds != nil {
		wbftCfg.MaxRequestTimeoutSeconds = *config.MaxRequestTimeoutSeconds
	}
	if config.StabilizingStakersThreshold != nil {
		wbftCfg.StabilizingStakersThreshold = *config.StabilizingStakersThreshold
	}
	if config.UseNCP != nil {
		wbftCfg.UseNCP = *config.UseNCP
	}

	hfTransitionBlocks := make(map[*big.Int]bool)

	//add hardforks that includes wbft config after croissant here like :
	// transition := params.Transition{
	// 	Block:      chainCfg.DalgonaBlock,
	// 	WBFTConfig: chainCfg.Dalgona.WBFT,
	// }
	// wbftCfg.Transitions = append(wbftCfg.Transitions, transition)
	// hfTransitionBlocks[chainCfg.DalgonaBlock] = true

	if chainCfg.Transitions != nil && len(chainCfg.Transitions) > 0 {
		for _, t := range chainCfg.Transitions {
			if hfTransitionBlocks[t.Block] {
				return errors.New("hardfork transition block already exists")
			}
			wbftCfg.Transitions = append(wbftCfg.Transitions, t)
		}
	}

	sort.Slice(wbftCfg.Transitions, func(i, j int) bool {
		if wbftCfg.Transitions[i].Block == nil {
			return false
		}
		if wbftCfg.Transitions[j].Block == nil {
			return true
		}
		return wbftCfg.Transitions[i].Block.Cmp(wbftCfg.Transitions[j].Block) < 0
	})

	wbftCfg.GovContractUpgrades = append(wbftCfg.GovContractUpgrades, params.Upgrade{Block: chainCfg.CroissantBlock, GovContracts: chainCfg.Croissant.GovContracts})
	// add hardforks that includes govContracts after croissant here like :
	// wbftCfg.GovContractUpgrades = append(wbftCfg.GovContractUpgrades, params.Upgrade{Block: chainCfg.DalgonaBlock, GovContracts: chainCfg.Dalgona.GovContracts})
	return nil
}

func CreateEthashFakeEngine(config *params.ChainConfig) (consensus.Engine, error) {
	if !config.TerminalTotalDifficultyPassed {
		return nil, errors.New("ethash is only supported as a historical component of already merged networks")
	}
	return beacon.New(ethash.NewFaker()), nil
}

func CreateFakeConsensusEngine(prvKey *ecdsa.PrivateKey, govCli wemixgov.GovBackend, config *params.ChainConfig, db ethdb.Database) (consensus.Engine, error) {
	// If proof-of-authority is requested, set it up
	engine := wpoa.NewWemixFakeEngine(prvKey, govCli)
	return beacon.New(engine), nil
}

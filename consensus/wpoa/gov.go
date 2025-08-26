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

package wpoa

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/wemixgov"
	"github.com/pkg/errors"
)

type WemixGov struct {
	coinbaseEnodeCache   *sync.Map
	height2enode         *LruCache
	blockBuildParamsLock sync.Mutex
	blockBuildParams     *BlockBuildParameters
	backend              wemixgov.GovBackend
}

func NewWemixGov(backend wemixgov.GovBackend) *WemixGov {
	wg := &WemixGov{}
	wg.coinbaseEnodeCache = &sync.Map{}
	wg.height2enode = NewLruCache(10000, true)
	wg.blockBuildParams = &BlockBuildParameters{}
	wg.backend = backend
	return wg
}

// cached governance data to derive miner's enode
type coinbaseEnodeEntry struct {
	modifiedBlock  *big.Int
	nodes          []*WemixNode
	coinbase2enode map[string][]byte // string(common.Address[:]) => []byte
	enode2index    map[string]int    // string([]byte) => int
}

func (wg *WemixGov) GetGovInfo(blockNumber *big.Int) (WemixGovInfo, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	govApi, err := wg.backend.GetGovApiWithHeight(ctx, blockNumber)
	if err != nil {
		return WemixGovInfo{}, err
	}

	result := WemixGovInfo{}

	result.Registry = govApi.GetRegistryAddress()
	result.Gov = govApi.GetGovAddress()
	result.Staking = govApi.GetStakingAddress()

	result.ModifiedBlock, err = govApi.GetModifiedBlock()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.BlockInterval, err = govApi.GetBlockCreationTime()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.BlocksPer, err = govApi.GetBlocksPer()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.BlockReward, err = govApi.GetBlockRewardAmount()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.MaxPriorityFeePerGas, err = govApi.GetMaxPriorityFeePerGas()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.MaxBaseFee, err = govApi.GetMaxBaseFee()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.GasLimit, result.BaseFeeMaxChangeRate, result.BaseFeeMaxChangeRate, err = govApi.GetGasLimitAndBaseFee()
	if err != nil {
		return WemixGovInfo{}, err
	}

	result.DefaultBriocheBlockReward = new(big.Int).Set(params.DefaultBriocheBlockReward)

	nodes := make([]*WemixNode, 0)
	nodeLength, err := govApi.GetNodeLength()
	if err != nil {
		return WemixGovInfo{}, err
	}
	count := nodeLength.Int64()
	for i := int64(1); i <= count; i++ {
		node, err := govApi.GetNode(big.NewInt(i))
		if err != nil {
			return WemixGovInfo{}, err
		}
		member, err := govApi.GetMember(big.NewInt(i))
		if err != nil {
			return WemixGovInfo{}, err
		}

		sid := hex.EncodeToString(node.Enode)
		if len(sid) != 128 {
			return WemixGovInfo{}, ErrInvalidEnode
		}
		idv4, _ := toIdv4(sid)
		nodes = append(nodes, &WemixNode{
			Name:  string(node.Name),
			Enode: sid,
			Ip:    string(node.Ip),
			Id:    idv4,
			Port:  int(node.Port.Int64()),
			Addr:  member,
		})
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	result.Nodes = nodes
	return result, nil
}

func (wg *WemixGov) GetLegacyBlockRewardAmount(height *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	govApi, err := wg.backend.GetGovApiWithHeight(ctx, height)
	if err != nil {
		return nil, err
	}

	rewardAmount, err := govApi.GetBlockRewardAmount()
	if err != nil {
		return nil, err
	}
	return rewardAmount, nil
}

func (wg *WemixGov) GetMaxPriorityFeePerGas(height *big.Int) (*big.Int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	govApi, err := wg.backend.GetGovApiWithHeight(ctx, height)
	if err != nil {
		return nil, err
	}

	fee, err := govApi.GetMaxPriorityFeePerGas()
	if err != nil {
		return nil, err
	}
	return fee, nil
}

func (wg *WemixGov) GetRewardParams(height *big.Int) (*RewardParameters, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rp := &RewardParameters{}
	govApi, err := wg.backend.GetGovApiWithHeight(ctx, height)
	if err != nil {
		return nil, err
	}

	rp.RewardAmount, err = govApi.GetBlockRewardAmount()
	if err != nil {
		return nil, err
	}

	distributionMethod1, distributionMethod2, distributionMethod3, distributionMethod4, err := govApi.GetBlockRewardDistributionMethod()
	if err != nil {
		return nil, err
	}
	rp.DistributionMethod = []*big.Int{distributionMethod1, distributionMethod2, distributionMethod3, distributionMethod4}

	staker, err := govApi.GetStakingRewardAddress()
	if err != nil {
		return nil, err
	}
	rp.Staker = &staker

	ecoSystem, err := govApi.GetEcosystemAddress()
	if err != nil {
		return nil, err
	}
	rp.EcoSystem = &ecoSystem

	maintenance, err := govApi.GetMaintenanceAddress()
	if err != nil {
		return nil, err
	}
	rp.Maintenance = &maintenance

	feeCollector, err := govApi.GetFeeCollectorAddress()
	if err != nil {
		rp.FeeCollector = nil
	} else {
		rp.FeeCollector = &feeCollector
	}

	blocksPer, err := govApi.GetBlocksPer()
	if err != nil {
		return nil, err
	}
	rp.BlocksPer = blocksPer.Int64()

	if countBig, err := govApi.GetMemberLength(); err != nil {
		return nil, err
	} else {
		count := countBig.Int64()
		for i := int64(1); i <= count; i++ {
			index := big.NewInt(i)
			if member, err := govApi.GetMember(index); err != nil {
				return nil, err
			} else if reward, err := govApi.GetReward(index); err != nil {
				return nil, err
			} else if stake, err := govApi.GetLockedBalanceOf(member); err != nil {
				return nil, err
			} else {
				rp.Members = append(rp.Members, &WemixMember{
					Staker: member,
					Reward: reward,
					Stake:  stake,
				})
			}
		}
	}
	return rp, nil
}

func (wg *WemixGov) VerifyBlockSig(height *big.Int, chain consensus.ChainHeaderReader, coinbase common.Address, nodeId []byte, hash common.Hash, sig []byte, checkMinerLimit bool) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get nodeid from the coinbase
	num := new(big.Int).Sub(height, common.Big1)
	govApi, err := wg.backend.GetGovApiWithHeight(ctx, num)
	if err != nil {
		return err == ErrNotInitialized || errors.Is(err, ethereum.NotFound)
	} else if count, err := govApi.GetMemberLength(); err != nil || count.Sign() == 0 {
		return err == ErrNotInitialized || count.Sign() == 0
	}
	// if minerNodeId is given, i.e. present in block header, use it,
	// otherwise, derive it from the codebase
	var data []byte
	if len(nodeId) == 0 {
		nodeId, err = wg.coinbaseExists(govApi, &coinbase)
		if err != nil || len(nodeId) == 0 {
			return false
		}
		data = append(height.Bytes(), hash.Bytes()...)
		data = crypto.Keccak256(data)
	} else {
		if _, err := wg.enodeExists(govApi, nodeId); err != nil {
			return false
		}
		data = hash.Bytes()
	}
	pubKey, err := crypto.Ecrecover(data, sig)
	if err != nil || len(pubKey) < 1 || !bytes.Equal(nodeId, pubKey[1:]) {
		return false
	}
	// check miner limit
	if !checkMinerLimit {
		return true
	}
	ok, err := wg.verifyMinerLimit(chain, height, govApi, &coinbase, nodeId)
	return err == nil && ok
}

func (wg *WemixGov) GetBlockBuildParameters(height *big.Int) (blockInterval int64, maxBaseFee, gasLimit *big.Int, baseFeeMaxChangeRate, gasTargetPercentage int64, err error) {
	err = ErrNotInitialized

	wg.blockBuildParamsLock.Lock()
	if wg.blockBuildParams != nil && wg.blockBuildParams.Height == height.Uint64() {
		// use cached
		blockInterval = wg.blockBuildParams.BlockInterval
		maxBaseFee = wg.blockBuildParams.MaxBaseFee
		gasLimit = wg.blockBuildParams.GasLimit
		baseFeeMaxChangeRate = wg.blockBuildParams.BaseFeeMaxChangeRate
		gasTargetPercentage = wg.blockBuildParams.GasTargetPercentage
		wg.blockBuildParamsLock.Unlock()
		err = nil
		return
	}
	wg.blockBuildParamsLock.Unlock()

	// default values
	blockInterval = 15
	maxBaseFee = big.NewInt(0)
	gasLimit = big.NewInt(0)
	baseFeeMaxChangeRate = 0
	gasTargetPercentage = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	govApi, err2 := wg.backend.GetGovApiWithHeight(ctx, height)
	if err2 != nil {
		err = ErrNotInitialized
		return
	}

	if count, err2 := govApi.GetMemberLength(); err2 != nil || count.Sign() == 0 {
		err = ErrNotInitialized
		return
	}
	if v, err2 := govApi.GetBlockCreationTime(); err2 != nil {
		err = ErrNotInitialized
		return
	} else {
		blockInterval = v.Int64()
	}

	if GasLimit, BaseFeeMaxChangeRate, GasTargetPercentage, err2 := govApi.GetGasLimitAndBaseFee(); err2 != nil {
		err = ErrNotInitialized
		return
	} else {
		gasLimit = GasLimit
		baseFeeMaxChangeRate = BaseFeeMaxChangeRate.Int64()
		gasTargetPercentage = GasTargetPercentage.Int64()
	}

	if maxBaseFee, err = govApi.GetMaxBaseFee(); err != nil {
		err = ErrNotInitialized
		return
	}

	// cache it
	wg.blockBuildParamsLock.Lock()
	wg.blockBuildParams = &BlockBuildParameters{
		Height:               height.Uint64(),
		BlockInterval:        blockInterval,
		MaxBaseFee:           maxBaseFee,
		GasLimit:             gasLimit,
		BaseFeeMaxChangeRate: baseFeeMaxChangeRate,
		GasTargetPercentage:  gasTargetPercentage,
	}
	wg.blockBuildParamsLock.Unlock()
	err = nil
	return
}

func (wg *WemixGov) verifyMinerLimit(chain consensus.ChainHeaderReader, height *big.Int, govApi wemixgov.GovContractApi, coinbase *common.Address, enode []byte) (bool, error) {
	// parent block number
	prev := new(big.Int).Sub(height, common.Big1)
	e, err := wg.getCoinbaseEnodeCache(govApi)
	if err != nil {
		return false, err
	}
	// if count <= 2, not enforced
	if len(e.nodes) <= 2 {
		return true, nil
	}
	// if enode is not given, derive it from the coinbase
	if len(enode) == 0 {
		enode2, ok := e.coinbase2enode[string(coinbase[:])]
		if !ok {
			return false, nil
		}
		enode = enode2
	}
	// the enode should not appear within the last (member count / 2) blocks
	limit := len(e.nodes) / 2
	if limit > int(height.Int64()-e.modifiedBlock.Int64()-1) {
		limit = int(height.Int64() - e.modifiedBlock.Int64() - 1)
	}
	for h := new(big.Int).Set(prev); limit > 0; h, limit = h.Sub(h, common.Big1), limit-1 {
		blockMinerEnode, err := wg.getBlockMiner(chain, e, h)
		if err != nil {
			return false, err
		}
		if bytes.Equal(enode[:], blockMinerEnode[:]) {
			return false, nil
		}
	}
	return true, nil
}

// return block's miner node id
func (wg *WemixGov) getBlockMiner(chain consensus.ChainHeaderReader, entry *coinbaseEnodeEntry, height *big.Int) ([]byte, error) {
	// if already cached, use it
	if enode := wg.height2enode.Get(height.Uint64()); enode != nil {
		return enode.([]byte), nil
	}
	block := chain.GetHeaderByNumber(height.Uint64())
	if len(block.MinerNodeId) == 0 {
		enode, ok := entry.coinbase2enode[string(block.Coinbase[:])]
		if !ok {
			return nil, nil
		}
		wg.height2enode.Put(height.Uint64(), enode)
		return enode, nil
	} else {
		if _, ok := entry.enode2index[string(block.MinerNodeId)]; !ok {
			return nil, nil
		}
		wg.height2enode.Put(height.Uint64(), block.MinerNodeId)
		return block.MinerNodeId, nil
	}
}

// returns coinbase's enode if exists in governance at given height - 1
func (wg *WemixGov) coinbaseExists(govApi wemixgov.GovContractApi, coinbase *common.Address) ([]byte, error) {
	e, err := wg.getCoinbaseEnodeCache(govApi)
	if err != nil {
		return nil, err
	}
	enode, ok := e.coinbase2enode[string(coinbase[:])]
	if !ok {
		return nil, nil
	}
	return enode, nil
}

func (wg *WemixGov) getCoinbaseEnodeCache(govApi wemixgov.GovContractApi) (*coinbaseEnodeEntry, error) {
	modifiedBlock, err := govApi.GetModifiedBlock()
	if err != nil {
		return nil, err
	}
	if modifiedBlock.Sign() == 0 {
		return nil, ErrNotInitialized
	}

	// if found in cache, use it
	if e, ok := wg.coinbaseEnodeCache.Load(modifiedBlock.Int64()); ok {
		return e.(*coinbaseEnodeEntry), nil
	}
	// otherwise, load it from the governance
	var (
		count       *big.Int
		addr        common.Address
		name, enode []byte
		e           = &coinbaseEnodeEntry{
			modifiedBlock:  modifiedBlock,
			coinbase2enode: map[string][]byte{},
			enode2index:    map[string]int{},
		}
	)
	if count, err = govApi.GetNodeLength(); err != nil {
		return nil, err
	}
	for i := int64(1); i <= count.Int64(); i++ {
		ix := big.NewInt(i)
		if addr, err = govApi.GetReward(ix); err != nil {
			return nil, err
		}

		if output, err := govApi.GetNode(ix); err != nil {
			return nil, err
		} else {
			name, enode = output.Name, output.Enode
		}

		idv4, _ := toIdv4(hex.EncodeToString(enode))
		e.nodes = append(e.nodes, &WemixNode{
			Name:  string(name),
			Enode: string(enode), // note that this is not in hex unlike wemixAdmin
			Id:    idv4,
			Addr:  addr,
		})
		e.coinbase2enode[string(addr[:])] = enode
		e.enode2index[string(enode)] = int(i) // 1-based, not 0-based
	}
	wg.coinbaseEnodeCache.Store(modifiedBlock.Int64(), e)
	return e, nil
}

// returns true if enode exists in governance at given height-1
func (wg *WemixGov) enodeExists(govApi wemixgov.GovContractApi, enode []byte) (common.Address, error) {
	e, err := wg.getCoinbaseEnodeCache(govApi)
	if err != nil {
		return common.Address{}, err
	}
	ix, ok := e.enode2index[string(enode)]
	if !ok {
		return common.Address{}, ErrNotFound
	}
	return e.nodes[ix-1].Addr, nil
}

func toIdv4(id string) (string, error) {
	if len(id) == 64 {
		return id, nil
	} else if len(id) == 128 {
		idv4, err := enode.ParseV4(fmt.Sprintf("enode://%v@127.0.0.1:8589", id))
		if err != nil {
			return "", err
		} else {
			return idv4.ID().String(), nil
		}
	} else {
		return "", fmt.Errorf("invalid V5 Identifier")
	}
}

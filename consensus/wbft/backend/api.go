// Copyright 2017 The go-ethereum Authors
// Copyright 2024 The go-wemix-wbft Authors
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
// This file is derived from quorum/consensus/istanbul/backend/api.go (2024.07.25).
// Modified and improved for the wemix development.

package backend

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// API is a user facing RPC API to dump Istanbul state
type API struct {
	chain   consensus.ChainHeaderReader
	backend *Backend
}

// BlockSigners is contains who created and who signed a particular block, denoted by its number and hash
type BlockSigners struct {
	Number     uint64
	Hash       common.Hash
	Author     common.Address
	Committers []common.Address
}

// SealerActivity contains seal signature counts by type
type SealerActivity struct {
	Total         map[common.Address]int `json:"total"`         // Total seal signatures
	Prepared      map[common.Address]int `json:"prepared"`      // PreparedSeal signatures
	Committed     map[common.Address]int `json:"committed"`     // CommittedSeal signatures
	PrevPrepared  map[common.Address]int `json:"prevPrepared"`  // PrevPreparedSeal signatures
	PrevCommitted map[common.Address]int `json:"prevCommitted"` // PrevCommittedSeal signatures
}

// BlockRange contains block range information for status collection
type BlockRange struct {
	StartBlock  uint64 `json:"startBlock"`  // Starting block number
	EndBlock    uint64 `json:"endBlock"`    // Ending block number
	TotalBlocks uint64 `json:"totalBlocks"` // Total number of blocks processed
}

// RoundStats contains round distribution statistics
type RoundStats struct {
	RoundDistribution map[uint64]uint64 `json:"roundDistribution"` // Map of round number to occurrence count
}

// Status contains validator activity statistics
type Status struct {
	SealerActivity SealerActivity         `json:"sealerActivity"` // Seal signature counts by type
	AuthorCounts   map[common.Address]int `json:"author"`         // Block proposal counts
	BlockRange     BlockRange             `json:"blockRange"`     // Block range information
	RoundStats     RoundStats             `json:"roundStats"`     // Round distribution statistics
}

// NodeAddress returns the public address that is used to sign block headers in IBFT
func (api *API) NodeAddress() common.Address {
	return api.backend.Address()
}

// GetCommitSignersFromBlock returns the signers and minter for a given block number, or the
// latest block available if none is specified
func (api *API) GetCommitSignersFromBlock(number *rpc.BlockNumber) (*BlockSigners, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}

	if header == nil {
		return nil, wbftcommon.ErrUnknownBlock
	}

	return api.commitSigners(header)
}

// GetSignersFromBlockByHash returns the signers and minter for a given block hash
func (api *API) GetCommitSignersFromBlockByHash(hash common.Hash) (*BlockSigners, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, wbftcommon.ErrUnknownBlock
	}

	return api.commitSigners(header)
}

func (api *API) commitSigners(header *types.Header) (*BlockSigners, error) {
	author, err := api.backend.Author(header)
	if err != nil {
		return nil, err
	}

	committers, err := api.backend.CommitSigners(api.chain, header)
	if err != nil {
		return nil, err
	}

	return &BlockSigners{
		Number:     header.Number.Uint64(),
		Hash:       header.Hash(),
		Author:     author,
		Committers: committers,
	}, nil
}

// GetValidators retrieves the list of authorized validators at the specified block.
func (api *API) GetValidators(number *rpc.BlockNumber) ([]common.Address, error) {
	// Retrieve the requested block number (or current if none requested)
	var header *types.Header
	if number == nil || *number == rpc.LatestBlockNumber {
		header = api.chain.CurrentHeader()
	} else {
		header = api.chain.GetHeaderByNumber(uint64(number.Int64()))
	}
	// Ensure we have an actually valid block and return the validators from its snapshot
	if header == nil {
		return nil, wbftcommon.ErrUnknownBlock
	}
	valSet, err := api.backend.Engine().GetValidators(api.chain, header.Number, header.ParentHash, nil)
	if err != nil {
		return nil, err
	}
	return valSet.AddressList(), nil
}

// GetValidatorsAtHash retrieves the state snapshot at a given block.
func (api *API) GetValidatorsAtHash(hash common.Hash) ([]common.Address, error) {
	header := api.chain.GetHeaderByHash(hash)
	if header == nil {
		return nil, wbftcommon.ErrUnknownBlock
	}
	valSet, err := api.backend.Engine().GetValidators(api.chain, header.Number, header.ParentHash, nil)
	if err != nil {
		return nil, err
	}
	return valSet.AddressList(), nil
}

// Status returns validator activity statistics and round statistics for the specified block range
func (api *API) Status(startBlockNum *rpc.BlockNumber, endBlockNum *rpc.BlockNumber) (*Status, error) {
	start, end, numBlocks, err := api.calculateBlockRange(startBlockNum, endBlockNum)
	if err != nil {
		return nil, err
	}

	activity := SealerActivity{
		Total:         make(map[common.Address]int),
		Prepared:      make(map[common.Address]int),
		Committed:     make(map[common.Address]int),
		PrevPrepared:  make(map[common.Address]int),
		PrevCommitted: make(map[common.Address]int),
	}
	authorCounts := make(map[common.Address]int)

	var cachedCurVals, cachedPrevVals []common.Address

	roundDistribution := make(map[uint64]uint64)
	for n := start; n <= end; n++ {
		round, err := api.analyzeBlock(n, &activity, authorCounts, &cachedCurVals, &cachedPrevVals)
		if err != nil {
			return nil, err
		}
		roundDistribution[round]++
	}

	return &Status{
		SealerActivity: activity,
		AuthorCounts:   authorCounts,
		BlockRange: BlockRange{
			StartBlock:  start,
			EndBlock:    end,
			TotalBlocks: numBlocks,
		},
		RoundStats: RoundStats{
			RoundDistribution: roundDistribution,
		},
	}, nil
}

// maxStatusBlockRange is the maximum number of blocks allowed per Status request to prevent DoS.
const maxStatusBlockRange = 1024

// calculateBlockRange calculates the block range for status collection
func (api *API) calculateBlockRange(startBlockNum *rpc.BlockNumber, endBlockNum *rpc.BlockNumber) (uint64, uint64, uint64, error) {
	if startBlockNum != nil && endBlockNum == nil {
		return 0, 0, 0, errors.New("pass the end block number")
	}

	if startBlockNum == nil && endBlockNum != nil {
		return 0, 0, 0, errors.New("pass the start block number")
	}

	currentNum := api.chain.CurrentHeader().Number.Uint64()

	var start, end uint64
	if startBlockNum == nil && endBlockNum == nil {
		// Default: last 64 blocks. When end < 63, start remains 0 (genesis included).
		end = currentNum
		if end >= 63 {
			start = end - 63
		}
	} else {
		resolve := func(n rpc.BlockNumber) (uint64, error) {
			if n >= 0 {
				return uint64(n), nil
			}
			if n == rpc.LatestBlockNumber {
				return currentNum, nil
			}
			return 0, fmt.Errorf("unsupported block number: %d", n)
		}
		var err error
		if end, err = resolve(*endBlockNum); err != nil {
			return 0, 0, 0, err
		}
		if start, err = resolve(*startBlockNum); err != nil {
			return 0, 0, 0, err
		}
		if start > end {
			return 0, 0, 0, errors.New("start block number should be less than end block number")
		}
		if end > currentNum {
			return 0, 0, 0, errors.New("end block number should be less than or equal to current block height")
		}
	}

	numBlocks := end - start + 1
	if numBlocks > maxStatusBlockRange {
		return 0, 0, 0, fmt.Errorf("requested range too large: %d blocks (max %d)", numBlocks, maxStatusBlockRange)
	}
	return start, end, numBlocks, nil
}

// analyzeBlock analyzes a single block and updates counters.
// Validator sets are cached and refreshed only on epoch transition to avoid redundant DB calls.
func (api *API) analyzeBlock(blockNum uint64, activity *SealerActivity, authorCounts map[common.Address]int, cachedCurVals, cachedPrevVals *[]common.Address) (uint64, error) {
	header := api.chain.GetHeaderByNumber(blockNum)
	if header == nil {
		return 0, fmt.Errorf("block %d not found", blockNum)
	}

	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return 0, fmt.Errorf("block %d: failed to extract WBFT extra: %w", blockNum, err)
	}

	// Refresh validator cache on first block or after epoch transition
	if *cachedCurVals == nil {
		curValidators, prevValidators, err := api.backend.GetValidatorsForVerifying(api.chain, header, nil)
		if err != nil {
			return 0, fmt.Errorf("block %d: failed to get validators: %w", blockNum, err)
		}
		*cachedCurVals = curValidators.AddressList()
		*cachedPrevVals = prevValidators.AddressList()

		// Initialize zero baseline for validators entering the range or new epoch
		initZero := func(addr common.Address, maps ...map[common.Address]int) {
			for _, m := range maps {
				if _, ok := m[addr]; !ok {
					m[addr] = 0
				}
			}
		}
		for _, addr := range *cachedCurVals {
			initZero(addr, activity.Prepared, activity.Committed, activity.Total, authorCounts)
		}
		for _, addr := range *cachedPrevVals {
			initZero(addr, activity.PrevPrepared, activity.PrevCommitted, activity.Total, authorCounts)
		}
	}
	curVals := *cachedCurVals
	prevVals := *cachedPrevVals

	// All data collected — update counters to avoid partial updates on error.
	// Count block author (proposal creator)
	author, err := api.backend.Author(header)
	if err == nil {
		authorCounts[author]++
	}

	addCounts := func(indices []uint32, vals []common.Address, target map[common.Address]int, addToTotal bool) {
		for _, idx := range indices {
			if int(idx) < len(vals) {
				addr := vals[idx]
				target[addr]++
				// addToTotal is false for null seals to exclude them from the total count
				if addToTotal {
					activity.Total[addr]++
				}
			}
		}
	}

	if extra.PreparedSeal != nil {
		addCounts(extra.PreparedSeal.Sealers.GetSealers(), curVals, activity.Prepared, true)
	}
	if extra.CommittedSeal != nil {
		addCounts(extra.CommittedSeal.Sealers.GetSealers(), curVals, activity.Committed, true)
	}
	if extra.PrevPreparedSeal != nil {
		addCounts(extra.PrevPreparedSeal.Sealers.GetSealers(), prevVals, activity.PrevPrepared, true)
	}
	if extra.PrevCommittedSeal != nil {
		addCounts(extra.PrevCommittedSeal.Sealers.GetSealers(), prevVals, activity.PrevCommitted, true)
	}

	if extra.EpochInfo != nil {
		// Next block starts a new epoch: reset cache to trigger refresh.
		// prevVals sync is skipped — GetValidatorsForVerifying will return correct values on next block.
		*cachedCurVals = nil
	} else {
		// Sync prevVals to curVals for subsequent blocks in the same epoch.
		*cachedPrevVals = *cachedCurVals
	}

	return uint64(extra.Round), nil
}

func (api *API) IsValidator(blockNum *rpc.BlockNumber) (bool, error) {
	var blockNumber rpc.BlockNumber
	if blockNum != nil {
		blockNumber = *blockNum
	} else {
		header := api.chain.CurrentHeader()
		blockNumber = rpc.BlockNumber(header.Number.Int64())
	}
	s, _ := api.GetValidators(&blockNumber)

	for _, v := range s {
		if v == api.backend.address {
			return true, nil
		}
	}
	return false, nil
}

func sealForJSON(seal *types.WBFTAggregatedSeal, valSet []common.Address) map[string]interface{} {
	if seal == nil {
		return nil
	}

	sealerIndxs := seal.Sealers.GetSealers()

	sealers := make([]string, 0, len(sealerIndxs))

	for _, idx := range sealerIndxs {
		if int(idx) < len(valSet) {
			sealers = append(sealers, valSet[idx].Hex())
		}
	}

	return map[string]interface{}{
		"sealers":   sealers,
		"signature": "0x" + hex.EncodeToString(seal.Signature),
	}
}

func epochForJSON(epoch *types.EpochInfo) map[string]interface{} {
	if epoch == nil {
		return nil
	}
	// Candidates
	candidates := make([]map[string]interface{}, 0, len(epoch.Candidates))

	for _, s := range epoch.Candidates {
		candidates = append(candidates, map[string]interface{}{
			"addr":      s.Addr.Hex(),
			"diligence": fmt.Sprintf("0x%x", s.Diligence),
		})
	}

	// Validators
	validators := make([]map[string]interface{}, 0, len(epoch.Validators))
	for i, idx := range epoch.Validators {
		validators = append(validators, map[string]interface{}{
			"index": fmt.Sprintf("0x%x", idx),
			"addr":  epoch.GetCandidate(idx).Hex(),
			"bls":   "0x" + hex.EncodeToString(epoch.BLSPublicKeys[i]),
		})
	}

	return map[string]interface{}{
		"candidates": candidates,
		"validators": validators,
	}
}

// DecodeVanityData decodes a 32-byte vanityData field.
// It detects if the input is UTF-8 or RLP encoded, and decodes accordingly.
func DecodeVanityData(vanity []byte) string {
	clean := bytes.TrimRight(vanity, "\x00")

	if utf8.Valid(clean) {
		return string(clean)
	}

	var val []interface{}

	err := rlp.DecodeBytes(clean, &val)
	versionBytes := val[0].([]uint8)
	if len(versionBytes) < 3 {
		tempBytes := make([]uint8, 3)
		copy(tempBytes[3-len(versionBytes):], versionBytes)
		versionBytes = tempBytes
	}
	version := uint32(versionBytes[0])<<16 | uint32(versionBytes[1])<<8 | uint32(versionBytes[2])
	if err == nil && version > 0 {
		major, minor, patch := versionBytes[0], versionBytes[1], versionBytes[2]
		clientBytes := val[1].([]byte)
		goVerBytes := val[2].([]byte)
		goOSBytes := val[3].([]byte)
		return fmt.Sprintf("[version: v%d.%d.%d, client: %s, go: %s, os: %s]", major, minor, patch, string(clientBytes), string(goVerBytes), string(goOSBytes))
	}

	return fmt.Sprintf("Unknown vanityData format, hex: 0x%s", hex.EncodeToString(clean))
}

func (api *API) GetWbftExtraInfo(number rpc.BlockNumber) (map[string]interface{}, error) {
	bNumber := big.NewInt(int64(number))

	if !api.chain.Config().AnzeonEnabled() {
		return nil, wbftcommon.ErrIsNotWBFTBlock
	}

	header := api.chain.GetHeaderByNumber(uint64(number))
	if header == nil {
		return nil, fmt.Errorf("block %d not found", bNumber)
	}

	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return nil, err
	}

	curValidators, prevValidators, err := api.backend.GetValidatorsForVerifying(api.chain, header, nil)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"vanityData":        DecodeVanityData(extra.VanityData),
		"randaoReveal":      "0x" + hex.EncodeToString(extra.RandaoReveal),
		"prevRound":         fmt.Sprintf("0x%x", extra.PrevRound),
		"prevPreparedSeal":  sealForJSON(extra.PrevPreparedSeal, prevValidators.AddressList()),
		"prevCommittedSeal": sealForJSON(extra.PrevCommittedSeal, prevValidators.AddressList()),
		"round":             fmt.Sprintf("0x%x", extra.Round),
		"preparedSeal":      sealForJSON(extra.PreparedSeal, curValidators.AddressList()),
		"committedSeal":     sealForJSON(extra.CommittedSeal, curValidators.AddressList()),
		"gasTip":            extra.GasTip.String(),
		"epochInfo":         epochForJSON(extra.EpochInfo),
	}

	return result, nil
}

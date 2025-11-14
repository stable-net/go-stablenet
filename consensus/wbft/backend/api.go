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
	TotalRounds       uint64            `json:"totalRounds"`       // Total number of rounds processed
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
	// Calculate block range
	start, end, numBlocks, blockNumber, err := api.calculateBlockRange(startBlockNum, endBlockNum)
	if err != nil {
		return nil, err
	}

	// Get validators
	signers, err := api.GetValidators(&blockNumber)
	if err != nil {
		return nil, err
	}

	// Initialize counters
	authorCounts, preparedCounts, committedCounts, prevPreparedCounts, prevCommittedCounts, totalSealCounts := api.initializeCounters(signers)

	// Analyze blocks and collect statistics
	roundDistribution := make(map[uint64]uint64)
	for n := start; n <= end; n++ {
		round := api.analyzeBlock(n, authorCounts, preparedCounts, committedCounts, prevPreparedCounts, prevCommittedCounts, totalSealCounts)
		roundDistribution[round]++
	}

	// Calculate total rounds (weighted sum) and remove rounds with zero count
	var totalRounds uint64
	for round, count := range roundDistribution {
		if count == 0 {
			delete(roundDistribution, round)
		} else {
			totalRounds += round * count
		}
	}
	return &Status{
		SealerActivity: SealerActivity{
			Total:         totalSealCounts,
			Prepared:      preparedCounts,
			Committed:     committedCounts,
			PrevPrepared:  prevPreparedCounts,
			PrevCommitted: prevCommittedCounts,
		},
		AuthorCounts: authorCounts,
		BlockRange: BlockRange{
			StartBlock:  start,
			EndBlock:    end,
			TotalBlocks: numBlocks,
		},
		RoundStats: RoundStats{
			RoundDistribution: roundDistribution,
			TotalRounds:       totalRounds,
		},
	}, nil
}

// calculateBlockRange calculates the block range for status collection
func (api *API) calculateBlockRange(startBlockNum *rpc.BlockNumber, endBlockNum *rpc.BlockNumber) (uint64, uint64, uint64, rpc.BlockNumber, error) {
	if startBlockNum != nil && endBlockNum == nil {
		return 0, 0, 0, 0, errors.New("pass the end block number")
	}

	if startBlockNum == nil && endBlockNum != nil {
		return 0, 0, 0, 0, errors.New("pass the start block number")
	}

	var start, end uint64
	var blockNumber rpc.BlockNumber

	if startBlockNum == nil && endBlockNum == nil {
		// Default: last 64 blocks
		header := api.chain.CurrentHeader()
		end = header.Number.Uint64()
		if end >= 63 {
			start = end - 63 // 64 blocks total
		} else {
			start = 1 // Start from block 1 if not enough blocks
		}
		blockNumber = rpc.BlockNumber(header.Number.Int64())
	} else {
		end = uint64(*endBlockNum)
		start = uint64(*startBlockNum)
		if start > end {
			return 0, 0, 0, 0, errors.New("start block number should be less than end block number")
		}

		if end > api.chain.CurrentHeader().Number.Uint64() {
			return 0, 0, 0, 0, errors.New("end block number should be less than or equal to current block height")
		}

		blockNumber = rpc.BlockNumber(end)
	}

	numBlocks := end - start + 1
	return start, end, numBlocks, blockNumber, nil
}

// initializeCounters initializes all counter maps for validators
func (api *API) initializeCounters(signers []common.Address) (map[common.Address]int, map[common.Address]int, map[common.Address]int, map[common.Address]int, map[common.Address]int, map[common.Address]int) {
	authorCounts := make(map[common.Address]int)
	preparedCounts := make(map[common.Address]int)
	committedCounts := make(map[common.Address]int)
	prevPreparedCounts := make(map[common.Address]int)
	prevCommittedCounts := make(map[common.Address]int)
	totalSealCounts := make(map[common.Address]int)

	for _, s := range signers {
		authorCounts[s] = 0
		preparedCounts[s] = 0
		committedCounts[s] = 0
		prevPreparedCounts[s] = 0
		prevCommittedCounts[s] = 0
		totalSealCounts[s] = 0
	}

	return authorCounts, preparedCounts, committedCounts, prevPreparedCounts, prevCommittedCounts, totalSealCounts
}

// analyzeBlock analyzes a single block and updates counters
func (api *API) analyzeBlock(blockNum uint64, authorCounts, preparedCounts, committedCounts, prevPreparedCounts, prevCommittedCounts, totalSealCounts map[common.Address]int) uint64 {
	// Fetch header
	header := api.chain.GetHeaderByNumber(blockNum)
	if header == nil {
		return 0
	}

	// Count block author (proposal creator)
	author, err := api.backend.Author(header)
	if err == nil {
		if _, ok := authorCounts[author]; !ok {
			authorCounts[author] = 0
		}
		authorCounts[author]++
	}

	// Count signers from prepared/committed and previous-round seals
	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return 0
	}

	curValidators, prevValidators, err := api.backend.GetValidatorsForVerifying(api.chain, header, nil)
	if err != nil {
		return uint64(extra.Round)
	}
	curVals := curValidators.AddressList()
	prevVals := prevValidators.AddressList()

	// helper to add counts safely
	addCounts := func(indices []uint32, vals []common.Address, target map[common.Address]int, addToTotal bool) {
		for _, idx := range indices {
			if int(idx) < len(vals) {
				addr := vals[idx]
				// ensure key exists on target
				if _, ok := target[addr]; !ok {
					target[addr] = 0
				}
				target[addr]++
				// only add to total if requested (for actual seals, not null ones)
				if addToTotal {
					if _, ok := totalSealCounts[addr]; !ok {
						totalSealCounts[addr] = 0
					}
					totalSealCounts[addr]++
				}
			}
		}
	}

	if extra.PreparedSeal != nil {
		addCounts(extra.PreparedSeal.Sealers.GetSealers(), curVals, preparedCounts, true)
	}
	if extra.CommittedSeal != nil {
		addCounts(extra.CommittedSeal.Sealers.GetSealers(), curVals, committedCounts, true)
	}
	if extra.PrevPreparedSeal != nil {
		addCounts(extra.PrevPreparedSeal.Sealers.GetSealers(), prevVals, prevPreparedCounts, true)
	}
	if extra.PrevCommittedSeal != nil {
		addCounts(extra.PrevCommittedSeal.Sealers.GetSealers(), prevVals, prevCommittedCounts, true)
	}

	return uint64(extra.Round)
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

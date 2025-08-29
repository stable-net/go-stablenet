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
//
// This file is derived from quorum/consensus/istanbul/qbft/engine/engine.go (2024.07.25).
// Modified and improved for the wemix development.

package wbftengine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	"github.com/ethereum/go-ethereum/consensus/wbft/core"
	"github.com/ethereum/go-ethereum/consensus/wbft/validator"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

const inmemoryCache = 128 // Number of recent validator set to keep in memory

var (
	nilUncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
)

type SignerFn func(data []byte) ([]byte, error)
type CheckSignatureFn func(data []byte, address common.Address, sig []byte) error

type Candidate struct {
	Addr      common.Address
	Power     *big.Int
	Diligence uint64
}

type Engine struct {
	cfg        *wbft.Config
	signer     common.Address   // Ethereum address of the signing key
	sign       SignerFn         // Signer function to authorize hashes with
	checkSig   CheckSignatureFn // Check signature function to verify signatures
	epochCache *lru.Cache[uint64, *types.EpochInfo]
}

func NewEngine(cfg *wbft.Config, signer common.Address, sign SignerFn, checkSig CheckSignatureFn) *Engine {
	return &Engine{
		cfg:        cfg,
		signer:     signer,
		sign:       sign,
		checkSig:   checkSig,
		epochCache: lru.NewCache[uint64, *types.EpochInfo](inmemoryCache),
	}
}

func mustHavePrevSeals(header *types.Header, config *params.ChainConfig) bool {
	return header.Number.Cmp(config.CroissantBlock) > 0 && header.Number.Cmp(common.Big1) > 0
}

func (e *Engine) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (e *Engine) CommitHeader(header *types.Header, preparedSeals, committedSeals []wbft.SealData, round *big.Int) error {
	_, err := ApplyHeaderWBFTExtra(
		header,
		writePreparedSeals(preparedSeals),
		writeCommittedSeals(committedSeals),
		writeRoundNumber(round),
	)
	return err
}

// writePreparedSeals writes the extra-data field of a block header with given prepared seals.
func writePreparedSeals(preparedSeals []wbft.SealData) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		if len(preparedSeals) == 0 {
			return wbftcommon.ErrInvalidPreparedSeals
		}
		aggregatedSeal, err := aggregateSeal(preparedSeals)
		if err != nil {
			return err
		}
		wbftExtra.PreparedSeal = aggregatedSeal
		return nil
	}
}

// writeCommittedSeals writes the extra-data field of a block header with given committed seals.
func writeCommittedSeals(committedSeals []wbft.SealData) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		if len(committedSeals) == 0 {
			return wbftcommon.ErrInvalidCommittedSeals
		}
		aggregatedSeal, err := aggregateSeal(committedSeals)
		if err != nil {
			return err
		}
		wbftExtra.CommittedSeal = aggregatedSeal
		return nil
	}
}

func aggregateSeal(sealDatas []wbft.SealData) (*types.WBFTAggregatedSeal, error) {
	seals := make([][]byte, 0)
	sealers := types.SealerSet{}
	for _, seal := range sealDatas {
		if len(seal.Seal) != types.IstanbulExtraSeal {
			return nil, wbftcommon.ErrInvalidSeal
		}
		sealers.SetSealer(seal.Sealer)
		seals = append(seals, seal.Seal)
	}

	aggregatedSeal, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		return nil, err
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   sealers,
		Signature: aggregatedSeal.Marshal(),
	}, nil
}

// writeRoundNumber writes the extra-data field of a block header with given round.
func writeRoundNumber(round *big.Int) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		wbftExtra.Round = uint32(round.Uint64())
		return nil
	}
}

func (e *Engine) VerifyBlockProposal(chain consensus.ChainHeaderReader, block *types.Block, validators wbft.ValidatorSet, prevValidators wbft.ValidatorSet) (time.Duration, error) {
	// check block body
	txnHash := types.DeriveSha(block.Transactions(), trie.NewStackTrie(nil))
	if txnHash != block.Header().TxHash {
		return 0, wbftcommon.ErrMismatchTxhashes
	}

	uncleHash := types.CalcUncleHash(block.Uncles())
	if uncleHash != nilUncleHash {
		return 0, wbftcommon.ErrInvalidUncleHash
	}

	// verify the header of proposed block
	err := e.VerifyHeader(chain, block.Header(), nil, validators, prevValidators, false)
	if errors.Is(err, consensus.ErrFutureBlock) {
		return time.Until(time.Unix(int64(block.Header().Time), 0)), consensus.ErrFutureBlock
	} else if err != nil {
		return 0, err
	}

	parentHeader := chain.GetHeaderByHash(block.ParentHash())
	if parentHeader == nil {
		return 0, fmt.Errorf("unknown parent hash")
	}

	return 0, nil
}

func (e *Engine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators wbft.ValidatorSet, prevValidators wbft.ValidatorSet, checkSeals bool) error {
	return e.verifyHeader(chain, header, parents, validators, prevValidators, checkSeals)
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (e *Engine) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators wbft.ValidatorSet, prevValidators wbft.ValidatorSet, checkSeals bool) error {
	if header.Number == nil {
		return wbftcommon.ErrUnknownBlock
	}

	// Don't waste time checking blocks from the future (adjusting for allowed threshold)
	adjustedTimeNow := time.Now().Add(time.Duration(e.cfg.AllowedFutureBlockTime) * time.Second).Unix()
	if header.Time > uint64(adjustedTimeNow) {
		return consensus.ErrFutureBlock
	}

	// Ensure that the block doesn't contain any uncles which are meaningless in Istanbul
	if header.UncleHash != nilUncleHash {
		return wbftcommon.ErrInvalidUncleHash
	}

	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if header.Difficulty == nil || header.Difficulty.Cmp(types.WBFTDefaultDifficulty) != 0 {
		return wbftcommon.ErrInvalidDifficulty
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}
	if chain.Config().IsShanghai(header.Number, header.Time) {
		return errors.New("wbft does not support shanghai fork")
	}
	// Verify the non-existence of withdrawalsHash.
	if header.WithdrawalsHash != nil {
		return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
	}
	if chain.Config().IsCancun(header.Number, header.Time) {
		return errors.New("wbft does not support cancun fork")
	}
	// Verify the non-existence of cancun-specific header fields
	switch {
	case header.ExcessBlobGas != nil:
		return fmt.Errorf("invalid excessBlobGas: have %d, expected nil", header.ExcessBlobGas)
	case header.BlobGasUsed != nil:
		return fmt.Errorf("invalid blobGasUsed: have %d, expected nil", header.BlobGasUsed)
	case header.ParentBeaconRoot != nil:
		return fmt.Errorf("invalid parentBeaconRoot, have %#x, expected nil", header.ParentBeaconRoot)
	}

	// All basic checks passed, verify cascading fields
	return e.verifyCascadingFields(chain, header, validators, prevValidators, parents, checkSeals)
}

// verifyCascadingFields verifies all the header fields that are not standalone,
// rather depend on a batch of previous headers. The caller may optionally pass
// in a batch of parents (ascending order) to avoid looking those up from the
// database. This is useful for concurrently verifying a batch of new headers.
func (e *Engine) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, validators wbft.ValidatorSet, prevValidators wbft.ValidatorSet, parents []*types.Header, checkSeal bool) error {
	// The genesis block is the always valid dead-end
	number := header.Number.Uint64()
	if number == 0 {
		return nil
	}

	// Check parent
	var parent *types.Header
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	} else {
		parent = chain.GetHeader(header.ParentHash, number-1)
	}

	// Ensure that the block's parent has right number and hash
	if parent == nil || parent.Number.Uint64() != number-1 || parent.Hash() != header.ParentHash {
		return consensus.ErrUnknownAncestor
	}

	// Ensure that the block's timestamp isn't too close to it's parent
	// When the BlockPeriod is reduced it is reduced for the proposal.
	// e.g when blockperiod is 1 from block 10 the block period between 9 and 10 is 1
	if parent.Time+e.cfg.GetConfig(header.Number).BlockPeriod > header.Time {
		return wbftcommon.ErrInvalidTimestamp
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil {
			return fmt.Errorf("invalid baseFee before fork: have %d, want <nil>", header.BaseFee)
		}
		if err := misc.VerifyGaslimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else if err := eip1559.VerifyEIP1559Header(chain.Config(), parent, header); err != nil {
		// Verify the header's EIP-1559 attributes.
		return err
	}

	// Verify signer
	if err := e.verifySigner(chain, header, parents, validators); err != nil {
		return err
	}

	// extract the extra data from the header
	currentExtra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return wbftcommon.ErrInvalidExtraDataFormat
	}

	// Verify seals
	if checkSeal {
		if err := e.verifySeals(header, validators, currentExtra); err != nil {
			return err
		}
	}

	if err := e.checkSig(makeRandaoData(chain.Config(), header.Number), header.Coinbase, currentExtra.RandaoReveal); err != nil {
		return fmt.Errorf("failed to verify randao reveal signature: %w", err)
	}
	calculatedRandaoMix := CalculateRandaoMix(parent.MixDigest, currentExtra.RandaoReveal)
	if calculatedRandaoMix != header.MixDigest {
		return fmt.Errorf("invalid randao mix: have %x, want %x", header.MixDigest, calculatedRandaoMix)
	}

	// prev seals validation for monblanc block or first block after genesis is skipped because it's empty
	if mustHavePrevSeals(header, chain.Config()) {
		// if croissant == 0: croissant+1(== 1) has no prev seals;
		// if croissant > 0: croissant+1 has prev seals;
		// Verify prevPreparedSeals and prevCommittedSeals
		if err := e.verifyPrevSeals(header, parent, prevValidators, currentExtra); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) verifySigner(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators wbft.ValidatorSet) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return wbftcommon.ErrUnknownBlock
	}

	// Resolve the authorization key and check against signers
	signer, err := e.Author(header)
	if err != nil {
		return err
	}

	// Signer should be in the validator set of previous block's extraData.
	if _, v := validators.GetByAddress(signer); v == nil {
		return wbftcommon.ErrUnauthorized
	}

	return nil
}

// verifyPrevSeals checks whether every prevPreparedSeals and prevCommittedSeals are signed by one of the parent's validators
func (e *Engine) verifyPrevSeals(header *types.Header, parent *types.Header, prevValidators wbft.ValidatorSet, extra *types.WBFTExtra) error {
	if parent.Number.Sign() == 0 {
		// We don't need to verify prepared seals in the genesis block
		return nil
	}

	prevPreparedSeal := extra.PrevPreparedSeal
	if prevPreparedSeal == nil || len(prevPreparedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyPrevPreparedSeals
	} else {
		//check whether prevPrepared seals are generated by prevValidators
		if err := verifyAggregatedSeal(prevValidators, parent, extra.PrevRound, prevPreparedSeal, core.SealTypePrepare); err != nil {
			log.Error("Failed to verify seal", "err", err)
			return wbftcommon.ErrInvalidPrevPreparedSeals
		}
	}

	prevCommittedSeal := extra.PrevCommittedSeal
	if prevCommittedSeal == nil || len(prevCommittedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyPrevCommittedSeals
	} else {
		if err := verifyAggregatedSeal(prevValidators, parent, extra.PrevRound, prevCommittedSeal, core.SealTypeCommit); err != nil {
			log.Error("Failed to verify seal", "err", err)
			return wbftcommon.ErrInvalidPrevCommittedSeals
		}
	}
	return nil
}

// verifySeals checks whether every prepared seals and committed seals are signed by one of validators
func (e *Engine) verifySeals(header *types.Header, validators wbft.ValidatorSet, extra *types.WBFTExtra) error {
	number := header.Number.Uint64()

	if number == 0 {
		// We don't need to verify committed seals in the genesis block
		return nil
	}

	preparedSeal := extra.PreparedSeal
	// The length of Prepared seals should be larger than 0
	if preparedSeal == nil || len(preparedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyPreparedSeals
	}

	if err := verifyAggregatedSeal(validators, header, extra.Round, preparedSeal, core.SealTypePrepare); err != nil {
		log.Error("Failed to verify seal", "err", err)
		return wbftcommon.ErrInvalidPreparedSeals
	}

	committedSeal := extra.CommittedSeal
	// The length of Committed seals should be larger than 0
	if committedSeal == nil || len(committedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyCommittedSeals
	}

	if err := verifyAggregatedSeal(validators, header, extra.Round, committedSeal, core.SealTypeCommit); err != nil {
		log.Error("Failed to verify seal", "err", err)
		return wbftcommon.ErrInvalidCommittedSeals
	}

	return nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (e *Engine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return wbftcommon.ErrInvalidUncleHash
	}
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (e *Engine) VerifySeal(chain consensus.ChainHeaderReader, header *types.Header, validators wbft.ValidatorSet) error {
	// get parent header and ensure the signer is in parent's validator set
	number := header.Number.Uint64()
	if number == 0 {
		return wbftcommon.ErrUnknownBlock
	}

	// ensure that the difficulty equals to wbft.DefaultDifficulty
	if header.Difficulty.Cmp(types.WBFTDefaultDifficulty) != 0 {
		return wbftcommon.ErrInvalidDifficulty
	}

	return e.verifySigner(chain, header, nil, validators)
}

func (e *Engine) PeriodToNextBlock(blockNumber *big.Int) uint64 {
	return e.cfg.GetConfig(blockNumber).BlockPeriod
}

func (e *Engine) Prepare(chain consensus.ChainHeaderReader, header *types.Header, extraPreparedSeal, extraCommittedSeal []wbft.SealData) error {
	header.Coinbase = e.Address()
	header.Nonce = wbftcommon.EmptyBlockNonce

	// copy the parent extra data as the header extra data
	number := header.Number.Uint64()

	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}

	// use the same difficulty for all blocks
	header.Difficulty = types.WBFTDefaultDifficulty

	// set header's timestamp
	header.Time = parent.Time + e.cfg.GetConfig(header.Number).BlockPeriod
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}

	var madeExtra *types.WBFTExtra
	var err error
	if mustHavePrevSeals(header, chain.Config()) {
		lastCanonicalHeader := chain.GetHeaderByNumber(header.Number.Uint64() - 1)
		if lastCanonicalHeader.Number.Sign() == 0 {
			madeExtra, err = ApplyHeaderWBFTExtra(header, e.WriteRandao(chain.Config(), header))
		} else {
			extra, err2 := types.ExtractWBFTExtra(lastCanonicalHeader)
			if err2 != nil {
				return err2
			}
			if extra.PreparedSeal == nil {
				return wbftcommon.ErrEmptyPreparedSeals
			}

			if extra.CommittedSeal == nil {
				return wbftcommon.ErrEmptyCommittedSeals
			}

			// make final prevSeals by merging existing seals and extra seals
			prevPreparedSeal := mergeSeals(extra.PreparedSeal, extraPreparedSeal)
			prevCommittedSeal := mergeSeals(extra.CommittedSeal, extraCommittedSeal)

			// add validators in snapshot to extraData's validators section and lastBlock committers to extraData's prevCommittedSeal section
			madeExtra, err = ApplyHeaderWBFTExtra(
				header,
				e.WriteRandao(chain.Config(), header),
				WritePrevSeals(extra.Round, prevPreparedSeal, prevCommittedSeal),
			)
		}
	} else {
		// monblac hardFork block has empty prev seal
		// next block of genesis croissant block has empty prev seal
		madeExtra, err = ApplyHeaderWBFTExtra(header, e.WriteRandao(chain.Config(), header))
	}
	if err != nil {
		return fmt.Errorf("failed to write wbft extra: %w", err)
	}
	header.MixDigest = CalculateRandaoMix(parent.MixDigest, madeExtra.RandaoReveal)
	return nil
}

func (e *Engine) WriteRandao(config *params.ChainConfig, header *types.Header) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		randaoReveal, err2 := e.sign(makeRandaoData(config, header.Number))
		if err2 != nil {
			return fmt.Errorf("failed to sign randao reveal: %w", err2)
		}

		wbftExtra.RandaoReveal = randaoReveal
		return nil
	}
}

func makeRandaoData(config *params.ChainConfig, number *big.Int) []byte {
	var data []byte
	chainId := config.ChainID
	randaoVersion := new(big.Int)
	if config.IsCroissant(number) {
		randaoVersion.SetUint64(1) // croissant randao version is 1
	}
	data = append(data, chainId.Bytes()...)
	data = append(data, randaoVersion.Bytes()...)
	data = append(data, number.Bytes()...)

	return crypto.Keccak256(data)
}

func CalculateRandaoMix(prevRandaoMix common.Hash, randaoReveal []byte) common.Hash {
	// Calculate the new RandaoMix by XORing the previous RandaoMix with the new RandaoReveal
	bigA := new(big.Int).SetBytes(prevRandaoMix.Bytes())
	bigB := new(big.Int).SetBytes(crypto.Keccak256Hash(randaoReveal).Bytes())
	resultBigInt := new(big.Int).Xor(bigA, bigB)
	return common.BigToHash(resultBigInt)
}

func WritePrevSeals(prevRound uint32, prevPreparedSeal, prevCommittedSeal *types.WBFTAggregatedSeal) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		wbftExtra.PrevRound = prevRound
		wbftExtra.PrevPreparedSeal = prevPreparedSeal
		wbftExtra.PrevCommittedSeal = prevCommittedSeal
		return nil
	}
}

func WriteEpochInfo(epochInfo *types.EpochInfo) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		wbftExtra.EpochInfo = epochInfo
		return nil
	}
}

// GetStakers
// If number of stakers >= minStakers (after stabilization stage), use staker list from gov.
// If number of stakers < minStakers , use validator list (regarded as staker list) from previous epoch.
func (e *Engine) GetStakers(config *params.ChainConfig, latestEpochInfo *types.EpochInfo, state govwbft.StateReader, num *big.Int) ([]common.Address, bool) {
	var (
		stakers     []common.Address
		stabilizing bool = latestEpochInfo.Stabilizing
	)
	govContracts := e.cfg.GetGovContracts(num, config)
	govStakingAddress := govContracts.GovStaking.Address
	if e.cfg.GetConfig(num).UseNCP {
		govNCPAddress := govContracts.GovNCP.Address
		stakers = govwbft.NCPStakers(govStakingAddress, govNCPAddress, state)
	} else {
		stakers = govwbft.Stakers(govStakingAddress, state)
	}
	if stabilizing {
		if uint64(len(stakers)) >= e.cfg.StabilizingStakersThreshold {
			// finally we have enough stakers to be stabilized!!
			stabilizing = false
		} else {
			// epoch in a stabilization stage has the validator set which is same to the previous epoch
			stakers = latestEpochInfo.GetValidators()
		}
	}

	return stakers, stabilizing
}

type stakerInfo struct {
	isValidator  bool // in current epoch
	wasValidator bool // in previous epoch
	staker       *types.Staker
}

// verifyHeader() must catch inconsistent seals before calling this.
func (e *Engine) buildEpochInfo(chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) (*types.EpochInfo, error) {
	var newEpoch types.EpochInfo

	config := chain.Config()
	if isEpoch, _, err := e.IsEpochBlockNumber(config, header.Number); err != nil {
		log.Error("IsEpochBlockNumber failed", "number", header.Number, "err", err)
		return nil, err
	} else if !isEpoch {
		return nil, nil
	}

	// Generate initial epoch block if a transition occurs.
	if config.CroissantBlock != nil && header.Number.Cmp(config.CroissantBlock) == 0 {
		return wbft.CreateInitialEpochInfo(config.Croissant)
	}

	proposedSealsInEpoch := make(map[common.Address]int)
	submittedSealsInEpoch := make(map[common.Address]int)
	proposedCountsInEpoch := make(map[common.Address]int)
	proposers := []common.Address{}
	epochLength := uint64(0)

	var lastProposer common.Address
	latestEpoch, latestEpochInfo, err := e.GetEpochInfo(chain, header, nil)
	if err != nil {
		log.Error("failed to get latest epoch info", "err", err)
		return nil, err
	}

	// Traverse blocks until reaching the epoch block.
	// Stop counting if the block reaches to the epoch block.
	it := header
	for latestEpoch.Cmp(it.Number) != 0 {
		extra, err := types.ExtractWBFTExtra(it)
		if err != nil {
			log.Error("failed to extract wbft extra data", "err", err)
			return nil, err
		}

		proposer, err := e.Author(it)
		if err != nil {
			log.Error("failed to get proposer", "err", err)
			return nil, err
		}
		proposers = append(proposers, proposer)

		parent := chain.GetHeader(it.ParentHash, it.Number.Uint64()-1)
		epochInfo := latestEpochInfo
		if latestEpoch.Cmp(parent.Number) == 0 {
			_, info, err := e.GetEpochInfo(chain, parent, nil)
			if err != nil {
				log.Error("failed to get prev epoch info", "number(parent)", parent.Number, "err", err)
				return nil, err
			}
			lastProposer, _ = e.Author(parent)
			epochInfo = info
		}

		// Accumulate PrevPreparedSeal counts.
		if parent.Number.Sign() == 0 {
			// If the parent block is the genesis block, PrevPreparedSeal is empty.
			// So we don't count proposed seal for the first block.
			proposedCountsInEpoch[proposer]-- // it will be -1
		} else {
			prepareSigners, err := getSignerAddress(epochInfo, extra.PrevPreparedSeal)
			if err != nil {
				log.Error("failed to get prev prepare signers", "err", err)
				return nil, err
			}

			proposedSealsInEpoch[proposer] += len(prepareSigners)
			for _, addr := range prepareSigners {
				submittedSealsInEpoch[addr]++
			}

			// Accumulate PrevCommittedSeal counts.
			commitSigners, err := getSignerAddress(epochInfo, extra.PrevCommittedSeal)
			if err != nil {
				log.Error("failed to get prev commit signers", "err", err)
				return nil, err
			}
			proposedSealsInEpoch[proposer] += len(commitSigners)
			for _, addr := range commitSigners {
				submittedSealsInEpoch[addr]++
			}

			log.Trace("Seals count", "current block number", it.Number, "prepareSigners", prepareSigners, "commitSigners", commitSigners)
		}

		// Update current header.
		it = parent
		epochLength++
	}

	stakerMap := make(map[common.Address]*stakerInfo)
	validators := []common.Address{}
	for _, staker := range latestEpochInfo.Stakers {
		stakerMap[staker.Addr] = &stakerInfo{staker: staker}
	}
	for _, val := range latestEpochInfo.Validators {
		addr := latestEpochInfo.GetValidator(val)
		stakerMap[addr].isValidator = true
		validators = append(validators, addr)
	}

	// set prior validator
	if it.Number.Cmp(config.CroissantBlock) > 0 {
		parent := chain.GetHeader(it.ParentHash, it.Number.Uint64()-1)
		_, priorEpochInfo, err2 := e.GetEpochInfo(chain, parent, nil)
		if err2 != nil {
			log.Error("failed to get prior epoch info", "err", err2)
			return nil, err2
		}
		for _, val := range priorEpochInfo.Validators {
			addr := priorEpochInfo.GetValidator(val)
			if stakerMap[addr] != nil { // if stakerMap[addr] == nil -> this is not current staker
				stakerMap[addr].wasValidator = true
			}
		}
	} else if it.Number.Sign() != 0 {
		// exeption case: if "it" is the croissant block, all current stakers were validators in the croissant block
		// in other words, all current stakers can have the previous seals in the point of the first block
		for _, st := range stakerMap {
			st.wasValidator = true
		}
	}

	// Accumulate proposer counts being selected within epoch.
	valSet := validator.NewSet(validators, latestEpochInfo.BLSPublicKeys, e.cfg.GetConfig(header.Number).ProposerPolicy)
	for i := len(proposers) - 1; i >= 0; i-- {
		proposer := proposers[i]
		for round := 0; ; round++ {
			// NOTE: WEMIX uses a round-robin policy to select proposers.
			// If round change occurs for every validators more than once,
			// latest round change cycle window will be used for counting.
			if round >= len(validators) {
				err := errors.New("failed to find valid proposer")
				log.Error("Invalid round", "err", err)
				return nil, err
			}

			valSet.CalcProposer(lastProposer, uint64(round))
			currP := valSet.GetProposer().Address()
			proposedCountsInEpoch[currP]++

			if currP == proposer {
				break
			}
		}
		lastProposer = proposer
	}

	log.Trace("Seals counts in epoch", "header.number", header.Number,
		"current block number", latestEpoch,
		"proposedSealsInEpoch", proposedSealsInEpoch,
		"submittedSealsInEpoch", submittedSealsInEpoch,
	)

	// Update epoch info.
	newStakers, stabilizing := e.GetStakers(config, latestEpochInfo, state, header.Number)
	newEpoch.Stakers = make([]*types.Staker, len(newStakers))
	for i, staker := range newStakers {
		var d uint64

		stakerInfo := stakerMap[staker]
		if stakerInfo == nil {
			// Assign default diligence for new staker.
			d = types.DefaultDiligence
		} else {
			// Calculate validator's diligence for current epoch.
			//
			// If validator proposed any blocks, d(h) = p / (2*v*w) + s / (2*e),
			// Otherwise, d(h) = 1 + s / (2*e)

			// applyingRate:
			// - prior_not && current_validator: e-1
			// - prior_validator && current_not: 1
			// - prior_validator && current_validator: e
			applyingRate := epochLength
			d = uint64(submittedSealsInEpoch[staker]) * types.DiligenceDenominator / (2 * epochLength)
			if !stakerInfo.isValidator {
				applyingRate = 1
				d = uint64(submittedSealsInEpoch[staker]) * types.DiligenceDenominator / 2
			} else if !stakerInfo.wasValidator {
				applyingRate = epochLength - 1
				if uint64(submittedSealsInEpoch[staker]) > 2*(epochLength-1) {
					return nil, errors.New("seal count exceed the range for non validator in prior epoch")
				}
				d = uint64(submittedSealsInEpoch[staker]) * types.DiligenceDenominator / (2 * (epochLength - 1))
			}

			if proposedCountsInEpoch[staker] > 0 {
				d += uint64(proposedSealsInEpoch[staker]) * types.DiligenceDenominator /
					uint64(2*len(latestEpochInfo.Validators)*proposedCountsInEpoch[staker])
			} else {
				d += types.DiligenceDenominator
			}

			// Calculate validator's cumulative diligence for next epoch.
			//
			// (n-1)validator-(n)validator:     D(h) = D(h-1) * 9/10 +          d(h) * 1/10
			// (n-1)non-validator-(n)validator: D(h) = D(h-1) * (9e + 1)/10e +  d(h) * (e-1)/10e
			// (n-1)validator-(n)non-validator: D(h) = D(h-1) * (10e - 1)/10e + d(h) * 1/10e
			d = (stakerInfo.staker.Diligence*(10*epochLength-applyingRate) + d*applyingRate) / 10 / epochLength
		}

		// Ensure Diligence is within valid range
		if d > 2*types.DiligenceDenominator {
			return nil, fmt.Errorf("WBFT: Invalid Diligence %d exceeds maximum", d)
		}

		newEpoch.Stakers[i] = &types.Staker{
			Addr:      staker,
			Diligence: d,
		}
	}

	// epoch in a stabilization stage has the validator set which is same to the previous epoch
	if stabilizing {
		newEpoch.Validators = latestEpochInfo.Validators
		newEpoch.BLSPublicKeys = latestEpochInfo.BLSPublicKeys
	} else {
		govStakingAddress := e.cfg.GetGovContracts(header.Number, config).GovStaking.Address
		candidates := make([]Candidate, 0, len(newStakers))
		for _, s := range newEpoch.Stakers {
			candidate := Candidate{
				Addr:      s.Addr,
				Diligence: s.Diligence,
				Power:     govwbft.GetTotalStaked(govStakingAddress, state, s.Addr),
			}
			candidates = append(candidates, candidate)
		}
		newEpoch.Validators, err = e.decideValidators(header, candidates, e.cfg.GetConfig(header.Number).TargetValidators)
		if err != nil {
			log.Error("Failed to decide validators", "err", err)
			return nil, err
		}
		newEpoch.BLSPublicKeys = make([][]byte, len(newEpoch.Validators))
		for i, addr := range newEpoch.GetValidators() {
			pk := govwbft.GetBLSPublicKey(govStakingAddress, state, addr)
			if len(pk) == 0 {
				err := errors.New("bls public key is zero")
				log.Error("Invalid BLS Public Key", "err", err)
				return nil, err
			}
			newEpoch.BLSPublicKeys[i] = pk
		}
	}
	newEpoch.Stabilizing = stabilizing

	log.Trace("update epoch info", "header.Number", header.Number, "validators", newEpoch.Validators)
	for i, staker := range newEpoch.Stakers {
		log.Trace(fmt.Sprintf("  - stakers[%d]", i), "addr", staker.Addr, "diligence", staker.Diligence)
	}

	e.epochCache.Add(header.Number.Uint64(), &newEpoch)
	return &newEpoch, nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
//
// Note, the block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (e *Engine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) error {
	return e.processFinalize(chain, header, state, txs, uncles, verifyEpoch)
}

// processFinalize is the internal implementation of Finalize.
//
// Parameters:
//   - epochHandler: A function that is executed when the block is an EpochBlock.
//     It processes actions specific to the EpochBlock, which records the ValidatorList for the next Epoch,
//     and is the last block of the (N-1)th Epoch for the (N)th Epoch.
func (e *Engine) processFinalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, epochHandler func(*Engine, consensus.ChainHeaderReader, *types.Header, govwbft.StateReader) error) error {
	// Accumulate any block and uncle rewards and commit the final state root
	if err := e.accumulateRewards(chain, state, header); err != nil {
		return err
	}

	if st, err := wbft.GetGovContractsStateTransition(e.cfg, header.Number); err != nil {
		return err
	} else if st != nil {
		for _, c := range st.Codes {
			state.SetCode(c.Address, hexutil.MustDecode(c.Code))
		}
		for _, s := range st.States {
			state.SetState(s.Address, s.Key, s.Value)
		}
	}

	if isEpoch, _, err := e.IsEpochBlockNumber(chain.Config(), header.Number); err != nil {
		return err
	} else if isEpoch && epochHandler != nil {
		if err = epochHandler(e, chain, header, state); err != nil {
			return err
		}
	}

	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = nilUncleHash
	return nil
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (e *Engine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Add the validatorList to the extra field of the header.
	if err := e.processFinalize(chain, header, state, txs, uncles, writeEpoch); err != nil {
		return nil, err
	}

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

func (e *Engine) SealHash(header *types.Header) common.Hash {
	return sigHash(header)
}

func (e *Engine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int).Set(types.WBFTDefaultDifficulty)
}

// IsEpochBlockNumber returns whether the given block number is an epoch block.
// it returns whether the given block number is an epoch block and the last epoch block number.
func (e *Engine) IsEpochBlockNumber(config *params.ChainConfig, number *big.Int) (bool, *big.Int, error) {
	if config.CroissantBlock != nil && !config.IsCroissant(number) {
		return false, nil, wbftcommon.ErrIsNotWBFTBlock
	}

	epochLength := new(big.Int).SetUint64(e.cfg.Epoch)
	firstNewEpoch := new(big.Int).SetUint64(0)
	if config.CroissantBlock != nil {
		firstNewEpoch.Set(config.CroissantBlock)
	}
	for _, transition := range e.cfg.Transitions {
		if transition.Block.Cmp(number) > 0 {
			break
		}
		if transition.EpochLength == 0 {
			continue
		}
		// EPOCH RULE: all epoch transition blocks are an epoch block
		firstNewEpoch.Set(transition.Block)
		epochLength.SetUint64(transition.EpochLength)
	}
	rem := new(big.Int).Sub(number, firstNewEpoch)
	rem.Rem(rem, epochLength)
	return rem.Sign() == 0, new(big.Int).Sub(number, rem), nil
}

// GetValidators retrieve the validator list of the epoch to which block of given number belongs.
// If the given block is an epoch block, it returns the validators of prior epoch.
// `parents` is a hint for backward traverse.
// exceptional case: blockNumber is genesis block number or croissant hard fork block number, then
// it returns the validators from chain config.
func (e *Engine) GetValidators(chain consensus.ChainHeaderReader, blockNumber *big.Int, parentHash common.Hash, parents []*types.Header) (wbft.ValidatorSet, error) {
	chainConfig := chain.Config()
	// 1. Check if the block is not a WBFT block
	if chainConfig.CroissantBlock != nil && !chainConfig.IsCroissant(blockNumber) {
		return nil, wbftcommon.ErrIsNotWBFTBlock
	}

	var (
		epochInfo *types.EpochInfo
		err       error
	)
	if blockNumber.Cmp(chainConfig.CroissantBlock) == 0 {
		//croissant hard fork validators from croissant config
		vs := validator.NewSet(chainConfig.Croissant.Init.Validators, chainConfig.Croissant.GetInitialBLSPublicKeys(), e.cfg.ProposerPolicy)
		return vs, nil
	}

	_, epochInfo, err = e.getEpochInfo(chain, blockNumber, parentHash, parents)
	if err != nil {
		log.Error("failed to get epochInfo", "err", err)
		return nil, err
	}

	vs := validator.NewSet(epochInfo.GetValidators(), epochInfo.BLSPublicKeys, e.cfg.GetConfig(blockNumber).ProposerPolicy)
	return vs, nil
}

func (e *Engine) PrepareSigners(chain consensus.ChainHeaderReader, header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	_, epochInfo, err := e.GetEpochInfo(chain, header, nil)
	if err != nil {
		return []common.Address{}, err
	}
	return getSignerAddress(epochInfo, extra.PreparedSeal)
}

func (e *Engine) CommitSigners(chain consensus.ChainHeaderReader, header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	_, epochInfo, err := e.GetEpochInfo(chain, header, nil)
	if err != nil {
		return []common.Address{}, err
	}
	return getSignerAddress(epochInfo, extra.CommittedSeal)
}

func (e *Engine) Address() common.Address {
	return e.signer
}

// FIXME: Need to update this for Istanbul
// sigHash returns the hash which is used as input for the Istanbul
// signing. It is the hash of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// Note, the method requires the extra data to be at least 65 bytes, otherwise it
// panics. This is done to avoid accidentally using both forms (signature present
// or not), which could be abused to produce different hashes for the same header.
func sigHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()
	rlp.Encode(hasher, types.WBFTFilteredHeader(header))
	hasher.Sum(hash[:0])
	return hash
}

func getExtra(header *types.Header) (*types.WBFTExtra, error) {
	if len(header.Extra) < types.IstanbulExtraVanity {
		// In this scenario, the header extradata only contains client specific information, hence create a new wbftExtra and set vanity
		vanity := append(header.Extra, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(header.Extra))...)
		return &types.WBFTExtra{
			VanityData:        vanity,
			RandaoReveal:      []byte{},
			PrevRound:         0,
			PrevPreparedSeal:  nil,
			PrevCommittedSeal: nil,
			Round:             0,
			PreparedSeal:      nil,
			CommittedSeal:     nil,
			EpochInfo:         nil,
		}, nil
	}

	// This is the case when Extra has already been set
	return types.ExtractWBFTExtra(header)
}

func setExtra(h *types.Header, wbftExtra *types.WBFTExtra) error {
	payload, err := rlp.EncodeToBytes(wbftExtra)
	if err != nil {
		return err
	}
	h.Extra = payload
	return nil
}

func makeRewardFunc(state *state.StateDB, blockReward *big.Int) func(*govwbft.Staker, *big.Int) {
	validatorReward := new(big.Int).Set(blockReward)
	return func(staker *govwbft.Staker, tot *big.Int) {
		r := new(big.Int).Set(validatorReward)
		r.Mul(r, staker.TotalStaked)
		r.Div(r, tot)
		if r.Sign() > 0 {
			state.AddBalance(staker.Rewardee, uint256.MustFromBig(r))
			blockReward.Sub(blockReward, r)
			log.Trace("WBFT: accumulate rewards to", "rewardee", staker.Rewardee, "block reward", r)
		} else {
			log.Trace("WBFT: skip accumulating rewards to", "rewardee", staker.Rewardee)
		}
	}
}

// AccumulateRewards credits the beneficiary of the given block with a reward.
func (e *Engine) accumulateRewards(chain consensus.ChainHeaderReader, state *state.StateDB, header *types.Header) error {
	var blockReward *big.Int

	// if brioche config exists, use it
	// else, get block reward from wbft config
	if chain.Config().IsBrioche(header.Number) {
		blockReward = chain.Config().Brioche.GetBriocheBlockReward(params.DefaultBriocheBlockReward, header.Number)
	} else {
		cfgBlockReward := e.cfg.GetConfig(header.Number).BlockReward
		if cfgBlockReward != nil {
			blockReward = new(big.Int).Set((*big.Int)(cfgBlockReward))
		}
	}

	// Deduct rewards of beneficiaries.
	if blockReward.Sign() > 0 {
		bReward := new(big.Int)
		beneficiaryInfo := e.cfg.GetConfig(header.Number).BlockRewardBeneficiary
		if beneficiaryInfo != nil {
			for _, beneficiary := range beneficiaryInfo.Beneficiaries {
				r := new(big.Int).Set(blockReward)
				r.Mul(r, new(big.Int).SetUint64(beneficiary.Numerator))
				r.Div(r, new(big.Int).SetUint64(beneficiaryInfo.Denominator))

				log.Debug("WBFT: accumulate rewards to", "beneficiary", beneficiary.Addr, "block reward", r)
				state.AddBalance(beneficiary.Addr, uint256.MustFromBig(r))
				bReward.Add(bReward, r)
			}
		}

		if blockReward.Cmp(bReward) < 0 {
			// Unreachable if genesis block is set correctly.
			log.Error("block reward underflow", "blockReward", blockReward, "bReward", bReward)
			return errors.New("block reward underflow")
		}
		blockReward.Sub(blockReward, bReward)
	}

	govStakingAddress := e.cfg.GetGovContracts(header.Number, chain.Config()).GovStaking.Address
	getStakerInfo := func(addr common.Address) *govwbft.Staker {
		// If the staker is removed from gov, use fallback staker info with zero staking.
		if !govwbft.IsStaker(govStakingAddress, state, addr) {
			log.Trace("WBFT: fallback staker info", "staker", addr)
			return &govwbft.Staker{
				Operator:    addr,
				Rewardee:    addr,
				TotalStaked: big.NewInt(0),
				Delegated:   big.NewInt(0),
			}
		}

		s := govwbft.StakerInfo(govStakingAddress, state, addr)
		return &s
	}

	// Distribute remaining block reward to validators (including proposer) who signed the block.
	if blockReward.Sign() > 0 {
		if err := e.calculateRewards(
			chain,
			header,
			makeRewardFunc(state, blockReward),
			getStakerInfo,
		); err != nil {
			log.Error("Error while calculating rewards", "err", err)
			return err
		}
	}

	// The reward left rewards to the proposer.
	if blockReward.Sign() > 0 {
		proposer, _ := e.Author(header)
		staker := getStakerInfo(proposer)
		state.AddBalance(staker.Rewardee, uint256.MustFromBig(blockReward))
		log.Trace("Block reward left rewards to", "rewardee", staker.Rewardee, "amount", blockReward)
	}
	return nil
}

// calculateRewards calculates the reward for the given block.
// Currently, seals are not considered for rewards because which we cannot determine malicious validators.
// Instead, we use diligence score to give faithful validator opportunity to propose more blocks.
func (e *Engine) calculateRewards(chain consensus.ChainHeaderReader, header *types.Header, rewardFn func(*govwbft.Staker, *big.Int), getStakerInfo func(common.Address) *govwbft.Staker) error {
	valSet, err := e.GetValidators(chain, header.Number, header.ParentHash, nil)
	if err != nil {
		return err
	}
	validators := valSet.AddressList()

	// Get staking amounts for rewardees.
	stakers := make([]*govwbft.Staker, len(validators))
	totStakingAmount := big.NewInt(0)
	for i, val := range validators {
		staker := getStakerInfo(val)
		stakers[i] = staker
		totStakingAmount.Add(totStakingAmount, staker.TotalStaked)
	}

	log.Debug("Calculating block reward", "currentBlock", header.Number, "totStakingAmount", totStakingAmount, "validator", validators)

	if rewardFn != nil && totStakingAmount.Sign() > 0 {
		for _, staker := range stakers {
			rewardFn(staker, totStakingAmount)
		}
	}

	return nil
}

func writeEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) error {
	newEpoch, err := e.buildEpochInfo(chain, header, state)
	if err != nil {
		return err
	}

	_, err = ApplyHeaderWBFTExtra(header, WriteEpochInfo(newEpoch))
	return err
}

// verifyEpoch is a handler that performs default actions when the block is an EpochBlock,
// and is called during the Finalize process.
// It validates the validity of the ValidatorList associated with the EpochBlock.
func verifyEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) error {
	bHeader := types.CopyHeader(header)
	epoch, err := e.buildEpochInfo(chain, bHeader, state)
	if err != nil {
		return err
	}

	extra, err := types.ExtractWBFTExtra(header)
	if err != nil {
		return err
	} else if extra.EpochInfo == nil {
		return errors.New("WBFT: epochInfo is nil")
	}

	// Check Stakers.
	if len(epoch.Stakers) != len(extra.EpochInfo.Stakers) {
		return errors.New("WBFT: mismatch in staker sizes")
	}
	for i := range epoch.Stakers {
		if epoch.Stakers[i].Addr != extra.EpochInfo.Stakers[i].Addr {
			return errors.New("WBFT: The two stakers do not match")
		}
		// Validate Diligence matches
		if epoch.Stakers[i].Diligence != extra.EpochInfo.Stakers[i].Diligence {
			return fmt.Errorf("WBFT: Diligence mismatch at index %d: expected %d, got %d",
				i, epoch.Stakers[i].Diligence, extra.EpochInfo.Stakers[i].Diligence)
		}
	}

	// Check validators.
	if len(epoch.Validators) != len(extra.EpochInfo.Validators) {
		return errors.New("WBFT: mismatch in validator sizes")
	}
	for i := range extra.EpochInfo.Validators {
		if epoch.Validators[i] != extra.EpochInfo.Validators[i] {
			return errors.New("WBFT: The two validators do not match")
		}
	}

	// Check BLS public keys.
	if len(epoch.BLSPublicKeys) != len(extra.EpochInfo.BLSPublicKeys) {
		return errors.New("WBFT: mismatch in BLS public key sizes")
	}
	for i, pk := range epoch.BLSPublicKeys {
		if !bytes.Equal(pk, extra.EpochInfo.BLSPublicKeys[i]) {
			return errors.New("WBFT: The two BLS public keys do not match")
		}
	}

	// Check Stabilizing flag.
	if epoch.Stabilizing != extra.EpochInfo.Stabilizing {
		return errors.New("WBFT: mismatch in stabilizing flag")
	}

	return nil
}

// 1. WEMIX 3.5
// - sort by (staking amount, diligence) descending
// - cut off the list at targetValidators
// - shuffle the list randomly
//
// 2. WEMIX 4.0 (not implemented yet)
// - randomSelect(staking amount, diligence) until targetValidators
// - shuffle the list randomly
func (e *Engine) decideValidators(header *types.Header, newStakers []Candidate, targetValidators uint64) ([]uint32, error) {
	indices := sortCandidates(newStakers)
	if uint64(len(indices)) > targetValidators {
		// If the number of candidates is greater than targetValidators,
		// we cut off the list at targetValidators.
		indices = indices[:targetValidators]
	}
	validators := make([]uint32, len(indices))
	for i := range indices {
		// Convert the index to uint32 and store it in the validators slice.
		shuffledIdx, err := computeShuffledIndex(uint64(i), uint64(len(validators)), [32]byte(header.MixDigest), true)
		if err != nil {
			log.Error("Failed to compute shuffled index", "index", i, "err", err)
			return nil, err
		}
		validators[i] = uint32(indices[shuffledIdx])
	}
	return validators, nil
}

func sortCandidates(candidates []Candidate) []int {
	indices := make([]int, len(candidates))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		originalIndexI := indices[i]
		originalIndexJ := indices[j]

		candidateI := candidates[originalIndexI]
		candidateJ := candidates[originalIndexJ]

		powerComparison := candidateI.Power.Cmp(candidateJ.Power)

		if powerComparison != 0 {
			return powerComparison > 0
		}

		return candidateI.Diligence > candidateJ.Diligence
	})

	return indices
}

func getSignerAddress(epochInfo *types.EpochInfo, signedSeal *types.WBFTAggregatedSeal) ([]common.Address, error) {
	if signedSeal == nil {
		return nil, wbftcommon.ErrEmptySeals
	}
	sealers := signedSeal.Sealers.GetSealers()
	signers := make([]common.Address, len(sealers))
	for i, idx := range sealers {
		v := epochInfo.GetValidator(epochInfo.Validators[idx])
		if v == (common.Address{}) {
			return nil, errors.New("validator address is zero")
		}
		signers[i] = v
	}
	return signers, nil
}

func verifyAggregatedSeal(valSet wbft.ValidatorSet, header *types.Header, round uint32, signedSeal *types.WBFTAggregatedSeal, sealType core.SealType) error {
	sealers := signedSeal.Sealers.GetSealers()
	// verify sealers
	if len(sealers) < valSet.QuorumSize() {
		return errors.New("lack of seal count")
	}

	blsPubKeys := make([][]byte, 0)
	for _, sealer := range sealers {
		val := valSet.GetByIndex(uint64(sealer))
		if val == nil {
			return errors.New("sealer is not validator")
		}
		blsPubKeys = append(blsPubKeys, val.BLSPublicKey())
	}

	sealData := core.PrepareSeal(header, round, sealType)
	aggregatedPubKey, err := bls.AggregatePublicKeys(blsPubKeys)
	if err != nil {
		return err
	}
	signature, err := bls.SignatureFromBytes(signedSeal.Signature)
	if err != nil {
		return err
	}

	if !signature.Verify(aggregatedPubKey, sealData) {
		return wbftcommon.ErrInvalidSeal
	}

	return nil
}

func mergeSeals(seal *types.WBFTAggregatedSeal, extraSeals []wbft.SealData) *types.WBFTAggregatedSeal {
	if len(extraSeals) == 0 {
		return seal
	}

	seals := [][]byte{seal.Signature}
	sealers := make(types.SealerSet, len(seal.Sealers))
	copy(sealers[:], seal.Sealers[:])

	for _, extraSeal := range extraSeals {
		if sealers.IsSealer(extraSeal.Sealer) {
			continue
		}
		sealers.SetSealer(extraSeal.Sealer)
		seals = append(seals, extraSeal.Seal)
	}

	aggregatedSeal, err := bls.AggregateCompressedSignatures(seals)
	if err != nil {
		return seal
	}

	return &types.WBFTAggregatedSeal{
		Sealers:   sealers,
		Signature: aggregatedSeal.Marshal(),
	}
}

func (e *Engine) GetEpochInfo(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header) (*big.Int, *types.EpochInfo, error) {
	if chain.Config().CroissantBlock.Cmp(header.Number) == 0 {
		if epochInfo, ok := e.epochCache.Get(header.Number.Uint64()); ok {
			return header.Number, epochInfo, nil
		}
		return e.extractEpochInfo(header)
	}

	return e.getEpochInfo(chain, header.Number, header.ParentHash, parents)
}

func (e *Engine) getEpochInfo(chain consensus.ChainHeaderReader, blockNumber *big.Int, parentHash common.Hash, parents []*types.Header) (*big.Int, *types.EpochInfo, error) {
	parentNumber := blockNumber.Uint64() - 1
	_, epochNum, err := e.IsEpochBlockNumber(chain.Config(), new(big.Int).SetUint64(parentNumber))
	if err != nil {
		return nil, nil, err
	}

	if epochInfo, ok := e.epochCache.Get(epochNum.Uint64()); ok {
		return epochNum, epochInfo, nil
	}

	if epochHeader := chain.GetHeaderByNumber(epochNum.Uint64()); epochHeader != nil {
		return e.extractEpochInfo(epochHeader)
	}

	// epoch block is not a canonical block
	var block *types.Header
	for {
		if len(parents) > 0 {
			block = parents[len(parents)-1]
			parents = parents[:len(parents)-1]
		} else {
			block = chain.GetHeader(parentHash, parentNumber)
		}
		if block == nil {
			return nil, nil, consensus.ErrUnknownAncestor
		}
		if epochNum.Cmp(block.Number) == 0 {
			break
		}
		parentHash, parentNumber = block.ParentHash, block.Number.Uint64()-1
	}
	return e.extractEpochInfo(block)
}

func (e *Engine) extractEpochInfo(epochHeader *types.Header) (*big.Int, *types.EpochInfo, error) {
	epochExtra, err := types.ExtractWBFTExtra(epochHeader)
	if err != nil {
		return nil, nil, err
	} else if epochExtra.EpochInfo == nil {
		return nil, nil, errors.New("WBFT: epochInfo is nil")
	}
	e.epochCache.Add(epochHeader.Number.Uint64(), epochExtra.EpochInfo)

	return epochHeader.Number, epochExtra.EpochInfo, nil
}

const seedSize = int8(32)
const roundSize = int8(1)
const positionWindowSize = int8(4)
const pivotViewSize = seedSize + roundSize
const totalSize = seedSize + roundSize + positionWindowSize
const shuffleRoundCount = uint8(33)

func fromBytes8(x []byte) uint64 {
	if len(x) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(x)
}

// computeShuffledIndex is from prysm's computeShuffledIndex function.
// this function follows ethereum beacon chain's shuffling algorithm.
func computeShuffledIndex(index uint64, indexCount uint64, seed [32]byte, shuffle bool) (uint64, error) {
	if index >= indexCount {
		return 0, fmt.Errorf("input index %d out of bounds: %d", index, indexCount)
	}
	rounds := shuffleRoundCount
	round := uint8(0)
	if !shuffle {
		// Starting last round and iterating through the rounds in reverse, un-swaps everything,
		// effectively un-shuffling the list.
		round = rounds - 1
	}
	buf := make([]byte, totalSize)
	posBuffer := make([]byte, 8)
	hashfunc := crypto.Keccak256Hash

	// Seed is always the first 32 bytes of the hash input, we never have to change this part of the buffer.
	copy(buf[:32], seed[:])
	for {
		buf[seedSize] = round
		h := hashfunc(buf[:pivotViewSize])
		hash8 := h[:8]
		hash8Int := fromBytes8(hash8)
		pivot := hash8Int % indexCount
		flip := (pivot + indexCount - index) % indexCount
		// Consider every pair only once by picking the highest pair index to retrieve randomness.
		position := index
		if flip > position {
			position = flip
		}
		// Add position except its last byte to []buf for randomness,
		// it will be used later to select a bit from the resulting hash.
		binary.LittleEndian.PutUint64(posBuffer[:8], position>>8)
		copy(buf[pivotViewSize:], posBuffer[:4])
		source := hashfunc(buf)
		// Effectively keep the first 5 bits of the byte value of the position,
		// and use it to retrieve one of the 32 (= 2^5) bytes of the hash.
		byteV := source[(position&0xff)>>3]
		// Using the last 3 bits of the position-byte, determine which bit to get from the hash-byte (note: 8 bits = 2^3)
		bitV := (byteV >> (position & 0x7)) & 0x1
		// index = flip if bit else index
		if bitV == 1 {
			index = flip
		}
		if shuffle {
			round++
			if round == rounds {
				break
			}
		} else {
			if round == 0 {
				break
			}
			round--
		}
	}
	return index, nil
}

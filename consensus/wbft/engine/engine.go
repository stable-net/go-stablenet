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
	"github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/crypto/sha3"
)

var (
	nilUncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
)

type SignerFn func(data []byte) ([]byte, error)
type CheckSignatureFn func(data []byte, address common.Address, sig []byte) error

type PoweredCandidate struct {
	Addr      common.Address
	Power     *big.Int
	Diligence uint64
}

type Engine struct {
	cfg      *wbft.Config
	signer   common.Address   // Ethereum address of the signing key
	sign     SignerFn         // Signer function to authorize hashes with
	checkSig CheckSignatureFn // Check signature function to verify signatures
	govTip   *big.Int         // tip value agreed through governance voting (in Wei)
}

func NewEngine(cfg *wbft.Config, signer common.Address, sign SignerFn, checkSig CheckSignatureFn) *Engine {
	return &Engine{
		cfg:      cfg,
		signer:   signer,
		sign:     sign,
		checkSig: checkSig,
		govTip:   big.NewInt(100 * params.GWei),
	}
}

// SetGovTip updates the governance-agreed tip value used by miners (in Wei).
func (e *Engine) SetGovTip(tip *big.Int) error {
	if tip == nil {
		return fmt.Errorf("invalid govTip: value is nil")
	}
	// keep a copy to avoid external mutation
	e.govTip = new(big.Int).Set(tip)
	return nil
}

func mustHavePrevSeals(header *types.Header) bool {
	return header.Number.Cmp(common.Big1) > 0
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

	// prev seals validation for anzeon block or first block after genesis is skipped because it's empty
	if mustHavePrevSeals(header) {
		// Verify prevPreparedSeals and prevCommittedSeals
		if err := e.verifyPrevSeals(header, parent, prevValidators, currentExtra); err != nil {
			return err
		}
	}

	if header.Number.Uint64() > 0 && currentExtra.GovTip != nil && currentExtra.GovTip.Cmp(e.govTip) != 0 {
		return fmt.Errorf("invalid gov tip: have %d, want %d", currentExtra.GovTip, e.govTip)
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
			log.Error("WBFT: failed to verify previous prepared seal", "number", parent.Number, "round", extra.PrevRound, "err", err)
			return wbftcommon.ErrInvalidPrevPreparedSeals
		}
	}

	prevCommittedSeal := extra.PrevCommittedSeal
	if prevCommittedSeal == nil || len(prevCommittedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyPrevCommittedSeals
	} else {
		if err := verifyAggregatedSeal(prevValidators, parent, extra.PrevRound, prevCommittedSeal, core.SealTypeCommit); err != nil {
			log.Error("WBFT: failed to verify previous committed seal", "number", parent.Number, "round", extra.PrevRound, "err", err)
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
		// NOTE: The caller already logs a WARN for invalid previous seals.
		// Here, the log is kept at TRACE to retain debug details while avoiding duplicate logs.
		log.Error("WBFT: failed to verify prepared seal", "err", err)
		return wbftcommon.ErrInvalidPreparedSeals
	}

	committedSeal := extra.CommittedSeal
	// The length of Committed seals should be larger than 0
	if committedSeal == nil || len(committedSeal.Signature) == 0 {
		return wbftcommon.ErrEmptyCommittedSeals
	}

	if err := verifyAggregatedSeal(validators, header, extra.Round, committedSeal, core.SealTypeCommit); err != nil {
		// NOTE: The caller already logs a WARN for invalid previous seals.
		// Here, the log is kept at TRACE to retain debug details while avoiding duplicate logs.
		log.Error("WBFT: failed to verify committed seal", "err", err)
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
	if mustHavePrevSeals(header) {
		lastCanonicalHeader := chain.GetHeaderByNumber(header.Number.Uint64() - 1)
		if lastCanonicalHeader.Number.Sign() == 0 {
			madeExtra, err = ApplyHeaderWBFTExtra(header, e.WriteRandao(chain.Config(), header), WriteGovTip(e.govTip))
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
				WriteGovTip(e.govTip),
			)
		}
	} else {
		// croissant hardFork block has empty prev seal
		// next block of genesis croissant block has empty prev seal
		madeExtra, err = ApplyHeaderWBFTExtra(header, e.WriteRandao(chain.Config(), header), WriteGovTip(e.govTip))
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
	if config.AnzeonEnabled() {
		randaoVersion.SetUint64(1) // anzeon randao version is 1
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

func WriteGovTip(govTip *big.Int) ApplyWBFTExtra {
	return func(wbftExtra *types.WBFTExtra) error {
		wbftExtra.GovTip = govTip
		return nil
	}
}

// GetGovCandidates retrieves the list of current validators from the GovValidator contract.
func (e *Engine) GetGovCandidates(config *params.ChainConfig, state systemcontracts.StateReader, num *big.Int) []common.Address {
	systemContracts := e.cfg.GetSystemContracts(num, config)
	govValidatorAddress := systemContracts.GovValidator.Address

	return systemcontracts.ValidatorList(govValidatorAddress, state)
}

type candidateInfo struct {
	isValidator  bool // in current epoch
	wasValidator bool // in previous epoch
	candidate    *types.Candidate
}

// verifyHeader() must catch inconsistent seals before calling this.
func (e *Engine) buildEpochInfo(chain consensus.ChainHeaderReader, header *types.Header, state systemcontracts.StateReader) (*types.EpochInfo, error) {
	var newEpoch types.EpochInfo

	config := chain.Config()
	if isEpoch, _, err := e.IsEpochBlockNumber(config, header.Number); err != nil {
		log.Error("WBFT: failed to determine epoch block", "number", header.Number, "err", err)
		return nil, err
	} else if !isEpoch {
		return nil, nil
	}

	// Generate initial epoch block if a transition occurs.
	if header.Number.Sign() == 0 {
		return wbft.CreateInitialEpochInfo(config.Anzeon)
	}

	proposedSealsInEpoch := make(map[common.Address]int)
	submittedSealsInEpoch := make(map[common.Address]int)
	beingProposerCountInEpoch := make(map[common.Address]int)
	proposers := []common.Address{}
	epochLength := uint64(0)

	var firstProposer, lastProposer common.Address
	latestEpoch, latestEpochInfo, err := e.GetEpochInfo(chain, header, nil)
	if err != nil {
		log.Error("WBFT: failed to get latest epoch info", "number", header.Number, "err", err)
		return nil, err
	}

	validatorsDiff := 0 // latestEpochInfo.Validators - previousEpochInfo.Validators
	// Traverse blocks until reaching the epoch block.
	// Stop counting if the block reaches to the epoch block.
	it := header
	for latestEpoch.Cmp(it.Number) != 0 {
		extra, err := types.ExtractWBFTExtra(it)
		if err != nil {
			log.Error("WBFT: failed to extract wbft extra data", "number", it.Number, "err", err)
			return nil, err
		}

		proposer, err := e.Author(it)
		if err != nil {
			log.Error("WBFT: failed to get proposer", "err", "number", it.Number, err)
			return nil, err
		}
		proposers = append(proposers, proposer)

		parent := chain.GetHeader(it.ParentHash, it.Number.Uint64()-1)
		epochInfo := latestEpochInfo
		if latestEpoch.Cmp(parent.Number) == 0 {
			_, info, err := e.GetEpochInfo(chain, parent, nil)
			if err != nil {
				log.Error("WBFT: failed to get previous epoch info", "number", parent.Number, "err", err)
				return nil, err
			}
			firstProposer, _ = e.Author(it)
			lastProposer, _ = e.Author(parent)
			epochInfo = info

			// Calculate the difference of validator set between two epochs. can be negative.
			validatorsDiff = len(latestEpochInfo.Validators) - len(info.Validators)
		}

		// Accumulate PrevPreparedSeal counts.
		if parent.Number.Sign() == 0 {
			// If the parent block is the genesis block, PrevPreparedSeal is empty.
			// So we don't count proposed seal for the first block.
			beingProposerCountInEpoch[proposer]-- // it will be -1
		} else {
			prepareSigners, err := getSignerAddress(epochInfo, extra.PrevPreparedSeal)
			if err != nil {
				log.Error("WBFT: failed to get previous prepare signers", "number", it.Number, "err", err)
				return nil, err
			}

			proposedSealsInEpoch[proposer] += len(prepareSigners)
			for _, addr := range prepareSigners {
				submittedSealsInEpoch[addr]++
			}

			// Accumulate PrevCommittedSeal counts.
			commitSigners, err := getSignerAddress(epochInfo, extra.PrevCommittedSeal)
			if err != nil {
				log.Error("WBFT: failed to get previous commit signers", "number", it.Number, "err", err)
				return nil, err
			}
			proposedSealsInEpoch[proposer] += len(commitSigners)
			for _, addr := range commitSigners {
				submittedSealsInEpoch[addr]++
			}

			log.Trace("WBFT: Seals count", "current block number", it.Number, "prepareSigners", prepareSigners, "commitSigners", commitSigners)
		}

		// Update current header.
		it = parent
		epochLength++
	}

	candidateMap := make(map[common.Address]*candidateInfo)
	validators := []common.Address{}
	for _, candi := range latestEpochInfo.Candidates {
		candidateMap[candi.Addr] = &candidateInfo{candidate: candi}
	}
	for _, val := range latestEpochInfo.Validators {
		addr := latestEpochInfo.GetCandidate(val)
		candidateMap[addr].isValidator = true
		validators = append(validators, addr)
	}

	// set prior validator
	if it.Number.Cmp(new(big.Int)) > 0 {
		parent := chain.GetHeader(it.ParentHash, it.Number.Uint64()-1)
		_, priorEpochInfo, err2 := e.GetEpochInfo(chain, parent, nil)
		if err2 != nil {
			log.Error("WBFT: failed to get prior epoch info", "number", parent.Number, "err", err2)
			return nil, err2
		}
		for _, val := range priorEpochInfo.Validators {
			addr := priorEpochInfo.GetCandidate(val)
			if candidateMap[addr] != nil { // if candidateMap[addr] == nil -> this is not current candidate
				candidateMap[addr].wasValidator = true
			}
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
				log.Error("WBFT: Invalid round", "num", header.Number.Uint64(), "err", err)
				return nil, err
			}

			valSet.CalcProposer(lastProposer, uint64(round))
			currP := valSet.GetProposer().Address()
			beingProposerCountInEpoch[currP]++

			if currP == proposer {
				break
			}
		}
		lastProposer = proposer
	}

	log.Trace("WBFT: Seals counts in epoch", "header.number", header.Number,
		"current block number", latestEpoch,
		"proposedSealsInEpoch", proposedSealsInEpoch,
		"submittedSealsInEpoch", submittedSealsInEpoch,
	)

	// Update epoch info.
	newCandidates := e.GetGovCandidates(config, state, header.Number)
	newEpoch.Candidates = make([]*types.Candidate, len(newCandidates))
	for i, candidate := range newCandidates {
		var d uint64

		candiInfo := candidateMap[candidate]
		if candiInfo == nil {
			// Assign default diligence for new candidate.
			d = types.DefaultDiligence
		} else {
			// Calculate validator's diligence for current epoch.
			//
			// p: number of seals of proposed blocks in current epoch
			// v: number of validators in current epoch
			// e: epoch length (in blocks)
			// w: number of times being proposer in current epoch
			// s: number of submitted seals in current epoch
			//
			// If validator proposed any blocks, d(h) = p / (2*v*w) + s / (2*e),
			// Otherwise, d(h) = 1 + s / (2*e)

			// applyingRate:
			// - prior_not && current_validator: e-1
			// - prior_validator && current_not: 1
			// - prior_validator && current_validator: e
			applyingRate := epochLength
			d = uint64(submittedSealsInEpoch[candidate]) * types.DiligenceDenominator / (2 * epochLength)
			if !candiInfo.isValidator {
				applyingRate = 1
				d = uint64(submittedSealsInEpoch[candidate]) * types.DiligenceDenominator / 2
			} else if !candiInfo.wasValidator {
				applyingRate = epochLength - 1
				if uint64(submittedSealsInEpoch[candidate]) > 2*(epochLength-1) {
					return nil, errors.New("seal count exceed the range for non validator in prior epoch")
				}
				d = uint64(submittedSealsInEpoch[candidate]) * types.DiligenceDenominator / (2 * (epochLength - 1))
			}

			if beingProposerCountInEpoch[candidate] > 0 {
				maxProposedSeals := 2 * len(latestEpochInfo.Validators) * beingProposerCountInEpoch[candidate]
				if candidate.Cmp(firstProposer) == 0 && candiInfo.wasValidator {
					// first proposer can put more prev seals as much as -validatorsDiff
					maxProposedSeals -= 2 * validatorsDiff
				}
				d += uint64(proposedSealsInEpoch[candidate]) * types.DiligenceDenominator / uint64(maxProposedSeals)
			} else {
				d += types.DiligenceDenominator
			}

			// Calculate validator's cumulative diligence for next epoch.
			//
			// (n-1)validator-(n)validator:     D(h) = D(h-1) * 9/10 +          d(h) * 1/10
			// (n-1)non-validator-(n)validator: D(h) = D(h-1) * (9e + 1)/10e +  d(h) * (e-1)/10e
			// (n-1)validator-(n)non-validator: D(h) = D(h-1) * (10e - 1)/10e + d(h) * 1/10e
			d = (candiInfo.candidate.Diligence*(10*epochLength-applyingRate) + d*applyingRate) / 10 / epochLength
		}

		// Ensure Diligence is within valid range
		if d > 2*types.DiligenceDenominator {
			return nil, fmt.Errorf("WBFT: Invalid Diligence %d exceeds maximum", d)
		}

		newEpoch.Candidates[i] = &types.Candidate{
			Addr:      candidate,
			Diligence: d,
		}
	}

	// epoch in a stabilization stage has the validator set which is same to the previous epoch
	govValidatorAddress := e.cfg.GetSystemContracts(header.Number, config).GovValidator.Address
	poweredCandidates := make([]PoweredCandidate, 0, len(newCandidates))
	for _, c := range newEpoch.Candidates {
		candi := PoweredCandidate{
			Addr:      c.Addr,
			Diligence: c.Diligence,
			Power:     big.NewInt(1), // Each validator has the same power in StableNet
		}
		poweredCandidates = append(poweredCandidates, candi)
	}
	newValidators, err := e.decideValidators(header, poweredCandidates)
	if err != nil {
		log.Error("WBFT: Failed to decide validators", "err", err)
		return nil, err
	}
	newEpoch.Validators = make([]uint32, 0)
	newEpoch.BLSPublicKeys = make([][]byte, 0)
	for _, newVal := range newValidators {
		addr := newEpoch.Candidates[newVal].Addr
		pk := systemcontracts.GetBLSPublicKey(govValidatorAddress, state, addr)
		if len(pk) == 0 {
			log.Warn("WBFT: no BLS public key for the validator", "validator", addr)
			// Although a specific validator's BLS key should not be missing from the GovValidator (due to potential bugs),
			// the consensus engine must not halt if this occurs.
			// Therefore, the engine is designed to exclude the affected validator and continue consensus.
			continue
		}
		newEpoch.Validators = append(newEpoch.Validators, newVal)
		newEpoch.BLSPublicKeys = append(newEpoch.BLSPublicKeys, pk)
	}

	log.Trace("WBFT: update epoch info", "header.Number", header.Number, "validators", newEpoch.Validators)
	for i, candidate := range newEpoch.Candidates {
		log.Trace(fmt.Sprintf("WBFT:   - candidates[%d]", i), "addr", candidate.Addr, "diligence", candidate.Diligence)
	}

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
func (e *Engine) processFinalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, epochHandler func(*Engine, consensus.ChainHeaderReader, *types.Header, systemcontracts.StateReader) error) error {
	if st, err := wbft.GetSystemContractsStateTransition(e.cfg, header.Number); err != nil {
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
	} else {
		extra, err := types.ExtractWBFTExtra(header)
		if err != nil {
			return err
		} else if extra.EpochInfo != nil {
			return wbftcommon.ErrEpochInfoIsNotNil
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
	if !config.AnzeonEnabled() {
		return false, nil, wbftcommon.ErrIsNotWBFTBlock
	}

	epochLength := new(big.Int).SetUint64(e.cfg.Epoch)
	firstNewEpoch := new(big.Int).SetUint64(0)
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
// exceptional case: blockNumber is genesis block number, then
// it returns the validators from chain config.
func (e *Engine) GetValidators(chain consensus.ChainHeaderReader, blockNumber *big.Int, parentHash common.Hash, parents []*types.Header) (wbft.ValidatorSet, error) {
	chainConfig := chain.Config()
	// 1. Check if the block is not a WBFT block
	if !chainConfig.AnzeonEnabled() {
		return nil, wbftcommon.ErrIsNotWBFTBlock
	}

	var (
		epochInfo *types.EpochInfo
		err       error
	)
	if blockNumber.Sign() == 0 {
		// if it is genesis block, get validators from anzeon config
		vs := validator.NewSet(chainConfig.Anzeon.Init.Validators, chainConfig.Anzeon.GetInitialBLSPublicKeys(), e.cfg.ProposerPolicy)
		return vs, nil
	}

	_, epochInfo, err = e.getEpochInfo(chain, blockNumber, parentHash, parents)
	if err != nil {
		log.Error("WBFT: failed to get epochInfo", "number", blockNumber, "err", err)
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

func writeEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state systemcontracts.StateReader) error {
	if valSet, err := e.GetValidators(chain, header.Number, header.ParentHash, nil); err != nil {
		return err
	} else if idx, _ := valSet.GetByAddress(e.Address()); idx < 0 {
		return nil
	}

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
func verifyEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state systemcontracts.StateReader) error {
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

	// Check Candidates.
	if len(epoch.Candidates) != len(extra.EpochInfo.Candidates) {
		return errors.New("WBFT: mismatch in candidate sizes")
	}
	for i := range epoch.Candidates {
		if epoch.Candidates[i].Addr != extra.EpochInfo.Candidates[i].Addr {
			return fmt.Errorf("WBFT: The two candidates do not match at index %d: expected %v, got %v",
				i, extra.EpochInfo.Candidates[i].Addr.Hex(), epoch.Candidates[i].Addr.Hex())
		}
		// Validate Diligence matches
		if epoch.Candidates[i].Diligence != extra.EpochInfo.Candidates[i].Diligence {
			return fmt.Errorf("WBFT: Diligence mismatch at index %d: expected %d, got %d",
				i, extra.EpochInfo.Candidates[i].Diligence, epoch.Candidates[i].Diligence)
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

	return nil
}

func (e *Engine) decideValidators(header *types.Header, newStakers []PoweredCandidate) ([]uint32, error) {
	indices := sortCandidates(newStakers)
	validators := make([]uint32, len(indices))
	for i := range indices {
		// Convert the index to uint32 and store it in the validators slice.
		shuffledIdx, err := computeShuffledIndex(uint64(i), uint64(len(validators)), header.MixDigest, true)
		if err != nil {
			log.Error("WBFT: Failed to compute shuffled index", "index", i, "err", err)
			return nil, err
		}
		validators[i] = uint32(indices[shuffledIdx])
	}
	return validators, nil
}

func sortCandidates(candidates []PoweredCandidate) []int {
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
		v := epochInfo.GetCandidate(epochInfo.Validators[idx])
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
	if header.Number.Sign() == 0 {
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

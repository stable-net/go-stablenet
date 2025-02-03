// Modification Copyright 2024 The Wemix Authors
//
// This file is derived from quorum/consensus/istanbul/qbft/engine/engine.go (2024.07.25).
// Modified and improved for the wemix development.

package qbftengine

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/lru"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/consensus/qbft/core"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
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

type Engine struct {
	cfg         *qbft.Config
	valSetCache *lru.Cache[uint64, qbft.ValidatorSet]

	signer common.Address // Ethereum address of the signing key
	sign   SignerFn       // Signer function to authorize hashes with
}

func NewEngine(cfg *qbft.Config, signer common.Address, sign SignerFn) *Engine {
	return &Engine{
		cfg:         cfg,
		valSetCache: lru.NewCache[uint64, qbft.ValidatorSet](inmemoryCache),
		signer:      signer,
		sign:        sign,
	}
}

func (e *Engine) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (e *Engine) CommitHeader(header *types.Header, preparedSeals, committedSeals [][]byte, round *big.Int) error {
	return ApplyHeaderQBFTExtra(
		header,
		writePreparedSeals(preparedSeals),
		writeCommittedSeals(committedSeals),
		writeRoundNumber(round),
	)
}

// writePreparedSeals writes the extra-data field of a block header with given prepared seals.
func writePreparedSeals(preparedSeals [][]byte) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		if len(preparedSeals) == 0 {
			return qbftcommon.ErrInvalidPreparedSeals
		}

		for _, seal := range preparedSeals {
			if len(seal) != types.IstanbulExtraSeal {
				return qbftcommon.ErrInvalidPreparedSeals
			}
		}

		qbftExtra.PreparedSeal = make([][]byte, len(preparedSeals))
		copy(qbftExtra.PreparedSeal, preparedSeals)

		return nil
	}
}

// writeCommittedSeals writes the extra-data field of a block header with given committed seals.
func writeCommittedSeals(committedSeals [][]byte) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		if len(committedSeals) == 0 {
			return qbftcommon.ErrInvalidCommittedSeals
		}

		for _, seal := range committedSeals {
			if len(seal) != types.IstanbulExtraSeal {
				return qbftcommon.ErrInvalidCommittedSeals
			}
		}

		qbftExtra.CommittedSeal = make([][]byte, len(committedSeals))
		copy(qbftExtra.CommittedSeal, committedSeals)

		return nil
	}
}

// writeRoundNumber writes the extra-data field of a block header with given round.
func writeRoundNumber(round *big.Int) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		qbftExtra.Round = uint32(round.Uint64())
		return nil
	}
}

func (e *Engine) VerifyBlockProposal(chain consensus.ChainHeaderReader, block *types.Block, validators qbft.ValidatorSet, prevValidators qbft.ValidatorSet) (time.Duration, error) {
	// check block body
	txnHash := types.DeriveSha(block.Transactions(), trie.NewStackTrie(nil))
	if txnHash != block.Header().TxHash {
		return 0, qbftcommon.ErrMismatchTxhashes
	}

	uncleHash := types.CalcUncleHash(block.Uncles())
	if uncleHash != nilUncleHash {
		return 0, qbftcommon.ErrInvalidUncleHash
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

func (e *Engine) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators qbft.ValidatorSet, prevValidators qbft.ValidatorSet, checkSeals bool) error {
	return e.verifyHeader(chain, header, parents, validators, prevValidators, checkSeals)
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (e *Engine) verifyHeader(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators qbft.ValidatorSet, prevValidators qbft.ValidatorSet, checkSeals bool) error {
	if header.Number == nil {
		return qbftcommon.ErrUnknownBlock
	}

	// Don't waste time checking blocks from the future (adjusting for allowed threshold)
	adjustedTimeNow := time.Now().Add(time.Duration(e.cfg.AllowedFutureBlockTime) * time.Second).Unix()
	if header.Time > uint64(adjustedTimeNow) {
		return consensus.ErrFutureBlock
	}

	if _, err := types.ExtractQBFTExtra(header); err != nil {
		return qbftcommon.ErrInvalidExtraDataFormat
	}

	// Ensure that the mix digest is zero as we don't have fork protection currently
	//if header.MixDigest != types.IstanbulDigest {
	//	return qbftcommon.ErrInvalidMixDigest
	//}

	// Ensure that the block doesn't contain any uncles which are meaningless in Istanbul
	if header.UncleHash != nilUncleHash {
		return qbftcommon.ErrInvalidUncleHash
	}

	// Ensure that the block's difficulty is meaningful (may not be correct at this point)
	if header.Difficulty == nil || header.Difficulty.Cmp(types.QBFTDefaultDifficulty) != 0 {
		return qbftcommon.ErrInvalidDifficulty
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
func (e *Engine) verifyCascadingFields(chain consensus.ChainHeaderReader, header *types.Header, validators qbft.ValidatorSet, prevValidators qbft.ValidatorSet, parents []*types.Header, checkSeal bool) error {
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
		return qbftcommon.ErrInvalidTimestamp
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

	// Verify seals
	if checkSeal {
		if err := e.verifySeals(header, validators); err != nil {
			return err
		}
	}

	// Verify prevPreparedSeals and prevCommittedSeals
	if err := e.verifyPrevSeals(chain, header, parent, prevValidators); err != nil {
		return err
	}

	return nil
}

func (e *Engine) verifySigner(chain consensus.ChainHeaderReader, header *types.Header, parents []*types.Header, validators qbft.ValidatorSet) error {
	// Verifying the genesis block is not supported
	number := header.Number.Uint64()
	if number == 0 {
		return qbftcommon.ErrUnknownBlock
	}

	// Resolve the authorization key and check against signers
	signer, err := e.Author(header)
	if err != nil {
		return err
	}

	// Signer should be in the validator set of previous block's extraData.
	if _, v := validators.GetByAddress(signer); v == nil {
		return qbftcommon.ErrUnauthorized
	}

	return nil
}

func verifySealers(sealers []common.Address, validators qbft.ValidatorSet) error {
	validatorsCpy := validators.Copy()
	validSealCnt := 0

	for _, sealer := range sealers {
		if validatorsCpy.RemoveValidator(sealer) {
			validSealCnt++
			continue
		}
		return fmt.Errorf("sealer is not validator")
	}

	if validSealCnt < validators.QuorumSize() {
		return fmt.Errorf("lack of seal count")
	}
	return nil
}

// verifyPrevSeals checks whether every prevPreparedSeals and prevCommittedSeals are signed by one of the parent's validators
func (e *Engine) verifyPrevSeals(chain consensus.ChainHeaderReader, header *types.Header, parent *types.Header, prevValidators qbft.ValidatorSet) error {
	number := header.Number.Uint64()

	if number == 0 {
		// We don't need to verify prepared seals in the genesis block
		return nil
	}

	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return err
	}

	var firstWbftBlockNum *big.Int

	if chain.Config().MontBlancBlock == nil {
		// wbft engine started from genesis
		firstWbftBlockNum = common.Big0
	} else {
		// wbft engine started with montblanc hardfork
		firstWbftBlockNum = chain.Config().MontBlancBlock
	}

	prevPreparedSeal := extra.PrevPreparedSeal
	if len(prevPreparedSeal) == 0 {
		// prevPreparedSeal validation for monblanc block or first block after genesis is skipped because it's empty
		if firstWbftBlockNum.Cmp(header.Number) != 0 && number != 1 {
			return qbftcommon.ErrEmptyPrevPreparedSeals
		}
	} else {
		//check whether prevPrepared seals are generated by prevValidators
		var prevPreparers []common.Address
		prevPreparers, err = e.GetSignerAddress(parent, extra.PrevRound, prevPreparedSeal, core.SealTypePrepare)
		if err != nil {
			return err
		}

		if err := verifySealers(prevPreparers, prevValidators); err != nil {
			return qbftcommon.ErrInvalidPrevPreparedSeals
		}
	}

	prevCommittedSeal := extra.PrevCommittedSeal
	if len(prevCommittedSeal) == 0 {
		// prevCommittedSeal validation for monblanc block is skipped because it's empty
		if firstWbftBlockNum.Cmp(header.Number) != 0 && number != 1 {
			return qbftcommon.ErrEmptyPrevCommittedSeals
		}
	} else {
		var prevCommitters []common.Address
		prevCommitters, err = e.GetSignerAddress(parent, extra.PrevRound, prevCommittedSeal, core.SealTypeCommit)
		if err != nil {
			return err
		}

		if err := verifySealers(prevCommitters, prevValidators); err != nil {
			return qbftcommon.ErrInvalidPrevCommittedSeals
		}
	}
	return nil
}

// verifySeals checks whether every prepared seals and committed seals are signed by one of validators
func (e *Engine) verifySeals(header *types.Header, validators qbft.ValidatorSet) error {
	number := header.Number.Uint64()

	if number == 0 {
		// We don't need to verify committed seals in the genesis block
		return nil
	}

	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return err
	}

	preparedSeal := extra.PreparedSeal
	// The length of Prepared seals should be larger than 0
	if len(preparedSeal) == 0 {
		return qbftcommon.ErrEmptyPreparedSeals
	}

	// Check whether the prepared seals are generated by validators
	preparers, err := e.GetSignerAddress(header, extra.Round, preparedSeal, core.SealTypePrepare)
	if err != nil {
		return err
	}

	if err := verifySealers(preparers, validators); err != nil {
		return qbftcommon.ErrInvalidPreparedSeals
	}

	committedSeal := extra.CommittedSeal
	// The length of Committed seals should be larger than 0
	if len(committedSeal) == 0 {
		return qbftcommon.ErrEmptyCommittedSeals
	}

	// Check whether the committed seals are generated by validator
	committers, err := e.GetSignerAddress(header, extra.Round, committedSeal, core.SealTypeCommit)
	if err != nil {
		return err
	}

	if err := verifySealers(committers, validators); err != nil {
		return qbftcommon.ErrInvalidCommittedSeals
	}

	return nil
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (e *Engine) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return qbftcommon.ErrInvalidUncleHash
	}
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (e *Engine) VerifySeal(chain consensus.ChainHeaderReader, header *types.Header, validators qbft.ValidatorSet) error {
	// get parent header and ensure the signer is in parent's validator set
	number := header.Number.Uint64()
	if number == 0 {
		return qbftcommon.ErrUnknownBlock
	}

	// ensure that the difficulty equals to qbft.DefaultDifficulty
	if header.Difficulty.Cmp(types.QBFTDefaultDifficulty) != 0 {
		return qbftcommon.ErrInvalidDifficulty
	}

	return e.verifySigner(chain, header, nil, validators)
}

func (e *Engine) PeriodToNextBlock(blockNumber *big.Int) uint64 {
	return e.cfg.GetConfig(blockNumber).BlockPeriod
}

func (e *Engine) Prepare(chain consensus.ChainHeaderReader, header *types.Header, validators qbft.ValidatorSet, extraPreparedSeal, extraCommittedSeal map[common.Hash][]byte) error {
	if _, v := validators.GetByAddress(e.signer); v == nil {
		return qbftcommon.ErrUnauthorized
	}

	header.Coinbase = e.signer
	header.Nonce = qbftcommon.EmptyBlockNonce

	// copy the parent extra data as the header extra data
	number := header.Number.Uint64()

	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}

	// use the same difficulty for all blocks
	header.Difficulty = types.QBFTDefaultDifficulty

	// set header's timestamp
	header.Time = parent.Time + e.cfg.GetConfig(header.Number).BlockPeriod
	if header.Time < uint64(time.Now().Unix()) {
		header.Time = uint64(time.Now().Unix())
	}

	var firstWbftBlockNum *big.Int
	if chain.Config().MontBlancBlock == nil {
		// wbft engine started from genesis
		firstWbftBlockNum = common.Big0
	} else {
		// wbft engine started with montblanc hardfork
		firstWbftBlockNum = chain.Config().MontBlancBlock
	}

	if firstWbftBlockNum.Cmp(header.Number) == 0 {
		// monblac hardFork block has empty prev seal
		// validators will be written at FinalizeAndAssemble
		return ApplyHeaderQBFTExtra(header)
	} else {
		lastCanonicalHeader := chain.GetHeaderByNumber(header.Number.Uint64() - 1)
		extra, err := types.ExtractQBFTExtra(lastCanonicalHeader)
		if err != nil {
			return err
		} else if extra.PreparedSeal == nil {
			return qbftcommon.ErrEmptyPreparedSeals
		} else if extra.CommittedSeal == nil {
			return qbftcommon.ErrEmptyCommittedSeals
		}

		// make final prevSeals by merging existing seals and extra seals
		prevPreparedSeal := mergeSeals(extra.PreparedSeal, extraPreparedSeal)
		prevCommittedSeal := mergeSeals(extra.CommittedSeal, extraCommittedSeal)

		// add validators in snapshot to extraData's validators section and lastBlock committers to extraData's prevCommittedSeal section
		return ApplyHeaderQBFTExtra(
			header,
			WritePrevSeals(extra.Round, prevPreparedSeal, prevCommittedSeal),
		)
	}
}

func WriteValidators(validators []common.Address) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		qbftExtra.Validators = validators
		return nil
	}
}

func WritePrevSeals(prevRound uint32, prevPreparedSeal [][]byte, prevCommittedSeal [][]byte) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		qbftExtra.PrevRound = prevRound
		qbftExtra.PrevPreparedSeal = prevPreparedSeal
		qbftExtra.PrevCommittedSeal = prevCommittedSeal
		return nil
	}
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
//
// Note, the block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (e *Engine) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	e.processFinalize(chain, header, state, txs, uncles, verifyEpoch)
}

// processFinalize is the internal implementation of Finalize.
//
// Parameters:
//   - epochHandler: A function that is executed when the block is an EpochBlock.
//     It processes actions specific to the EpochBlock, which records the ValidatorList for the next Epoch,
//     and is the last block of the (N-1)th Epoch for the (N)th Epoch.
func (e *Engine) processFinalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, epochHandler func(*Engine, consensus.ChainHeaderReader, *types.Header, govwbft.StateReader) error) error {
	// Accumulate any block and uncle rewards and commit the final state root
	e.accumulateRewards(chain, state, header)

	if transitions := qbft.GetStateTransitions(chain.Config(), header.Number); len(transitions) > 0 {
		for _, st := range transitions {
			for _, c := range st.Codes {
				state.SetCode(c.Address, hexutil.MustDecode(c.Code))
			}
			for _, s := range st.States {
				state.SetState(s.Address, s.Key, s.Value)
			}
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
	if err := e.processFinalize(chain, header, state, txs, uncles, writeValidatorsToEpoch); err != nil {
		return nil, err
	}

	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

func (e *Engine) SealHash(header *types.Header) common.Hash {
	return sigHash(header)
}

func (e *Engine) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	return new(big.Int).Set(types.QBFTDefaultDifficulty)
}

func (e *Engine) ExtractGenesisValidators(header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return nil, err
	}

	return extra.Validators, nil
}

// IsEpochBlockNumber returns whether the given block number is an epoch block.
// it returns whether the given block number is an epoch block and the last epoch block number.
func (e *Engine) IsEpochBlockNumber(config *params.ChainConfig, number *big.Int) (bool, *big.Int, error) {
	if config.MontBlancBlock != nil && !config.IsMontBlanc(number) {
		return false, nil, qbftcommon.ErrIsNotWBFTBlock
	}

	epochLength := new(big.Int).SetUint64(e.cfg.Epoch)
	firstNewEpoch := new(big.Int).SetUint64(0)
	if config.MontBlancBlock != nil {
		firstNewEpoch.Set(config.MontBlancBlock)
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
// exceptional case: blockNumber is genesis block number or montblanc hard fork block number, then
// it returns the validators from chain config.
func (e *Engine) GetValidators(chain consensus.ChainHeaderReader, blockNumber *big.Int, parentHash common.Hash, parents []*types.Header) (qbft.ValidatorSet, error) {
	if vs, ok := e.valSetCache.Get(blockNumber.Uint64()); ok {
		return vs.Copy(), nil
	}

	// 1. Check if the block is not a WBFT block
	if chain.Config().MontBlancBlock != nil && !chain.Config().IsMontBlanc(blockNumber) {
		return nil, qbftcommon.ErrIsNotWBFTBlock
	}

	if (chain.Config().MontBlancBlock == nil && blockNumber.Cmp(common.Big0) == 0) ||
		(chain.Config().MontBlancBlock != nil && chain.Config().MontBlancBlock.Cmp(blockNumber) == 0) {
		// genesis validators or montblanc hard fork validators from wbft config
		return validator.NewSet(e.cfg.Validators, e.cfg.ProposerPolicy), nil
	}

	// traverse back to the last epoch block
	parentNumber := new(big.Int).Sub(blockNumber, common.Big1)
	isEpoch, latestEpoch, err := e.IsEpochBlockNumber(chain.Config(), parentNumber)
	if err != nil {
		return nil, err
	}
	var block *types.Header
	if len(parents) > 0 {
		block = parents[len(parents)-1]
		parents = parents[:len(parents)-1]
	} else {
		block = chain.GetHeader(parentHash, parentNumber.Uint64())
	}
	if block == nil {
		return nil, consensus.ErrUnknownAncestor
	}
	for !isEpoch && block.Number.Cmp(latestEpoch) > 0 {
		if vs, ok := e.valSetCache.Get(block.Number.Uint64()); ok {
			// this block is not epoch and this block's epoch is same with requested block
			// so if cache hit, it must be same validator set with requested one
			return vs.Copy(), nil
		}

		if len(parents) > 0 {
			block = parents[len(parents)-1]
			parents = parents[:len(parents)-1]
		} else {
			block = chain.GetHeader(block.ParentHash, block.Number.Uint64()-1)
		}
		if block == nil {
			return nil, consensus.ErrUnknownAncestor
		}
	}

	// block must be an epoch block
	qbftExtra, err := types.ExtractQBFTExtra(block)
	if err != nil {
		log.Error("BFT: invalid epoch header", "err", err)
		return nil, err
	}
	vs := validator.NewSet(qbftExtra.Validators, e.cfg.ProposerPolicy)
	e.valSetCache.Add(blockNumber.Uint64(), vs)
	return vs, nil
}

func (e *Engine) GetSignerAddress(header *types.Header, round uint32, signedSeal [][]byte, sealType core.SealType) ([]common.Address, error) {
	sealData := core.PrepareSeal(header, round, sealType)
	var addrs []common.Address

	for _, seal := range signedSeal {
		// Get the original address by seal and block hash
		addr, err := qbft.GetSignatureAddressNoHashing(sealData, seal)
		if err != nil {
			return nil, qbftcommon.ErrInvalidSignature
		}
		addrs = append(addrs, addr)
	}

	return addrs, nil
}

func (e *Engine) PrepareSigners(header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	preparedSeal := extra.PreparedSeal
	return e.GetSignerAddress(header, extra.Round, preparedSeal, core.SealTypePrepare)
}

func (e *Engine) CommitSigners(header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	committedSeal := extra.CommittedSeal
	return e.GetSignerAddress(header, extra.Round, committedSeal, core.SealTypeCommit)
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
	rlp.Encode(hasher, types.QBFTFilteredHeader(header))
	hasher.Sum(hash[:0])
	return hash
}

func getExtra(header *types.Header) (*types.QBFTExtra, error) {
	if len(header.Extra) < types.IstanbulExtraVanity {
		// In this scenario, the header extradata only contains client specific information, hence create a new qbftExtra and set vanity
		vanity := append(header.Extra, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(header.Extra))...)
		return &types.QBFTExtra{
			VanityData:        vanity,
			Validators:        []common.Address{},
			PrevRound:         0,
			PrevPreparedSeal:  [][]byte{},
			PrevCommittedSeal: [][]byte{},
			Round:             0,
			PreparedSeal:      [][]byte{},
			CommittedSeal:     [][]byte{},
		}, nil
	}

	// This is the case when Extra has already been set
	return types.ExtractQBFTExtra(header)
}

func setExtra(h *types.Header, qbftExtra *types.QBFTExtra) error {
	payload, err := rlp.EncodeToBytes(qbftExtra)
	if err != nil {
		return err
	}
	h.Extra = payload
	return nil
}

// AccumulateRewards credits the beneficiary of the given block with a reward.
func (e *Engine) accumulateRewards(chain consensus.ChainHeaderReader, state *state.StateDB, header *types.Header) {
	blockReward := chain.Config().GetBlockReward(header.Number)
	if blockReward.Cmp(big.NewInt(0)) > 0 {
		coinbase := header.Coinbase
		if (coinbase == common.Address{}) {
			coinbase = e.signer
		}
		rewardAccount, _ := chain.Config().GetRewardAccount(header.Number, coinbase)
		log.Trace("QBFT: accumulate rewards to", "rewardAccount", rewardAccount, "blockReward", blockReward)

		state.AddBalance(rewardAccount, uint256.MustFromBig(&blockReward))

		if err := e.calculateRewards(
			chain,
			header,
			func(addr common.Address, amt *big.Int) { state.AddBalance(addr, uint256.MustFromBig(amt)) },
			func(addr common.Address, amt *big.Int) { state.AddBalance(addr, uint256.MustFromBig(amt)) },
		); err != nil {
			// TODO: how to handle err here?
			log.Warn("Error while calculating rewards", "err", err)
		}
	}
}

func (e *Engine) calculateRewards(chain consensus.ChainHeaderReader, header *types.Header, prepareRewardFn, commitRewardFn func(common.Address, *big.Int)) error {
	// TODO : need proper calculation when distribution rule is decided. Current work is to get rewardee's address from preCommittedSeal

	parentHeader := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	lastQbftExtra, err := types.ExtractQBFTExtra(parentHeader)
	if err != nil {
		return qbftcommon.ErrInvalidExtraDataFormat
	}

	// proposal seals of prior block
	proposalPrepareSeal := core.PrepareSeal(parentHeader, lastQbftExtra.Round, core.SealTypePrepare)
	proposalCommitSeal := core.PrepareSeal(parentHeader, lastQbftExtra.Round, core.SealTypeCommit)

	var prepareRewardees []common.Address
	var commitRewardees []common.Address

	qbftExtra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return qbftcommon.ErrInvalidExtraDataFormat
	}

	// get prev prepared address
	for _, seal := range qbftExtra.PrevPreparedSeal {
		addr, err := qbft.GetSignatureAddressNoHashing(proposalPrepareSeal, seal)
		if err != nil {
			return qbftcommon.ErrInvalidSignature
		}
		prepareRewardees = append(prepareRewardees, addr)
	}

	// get prev committed address
	for _, seal := range qbftExtra.PrevCommittedSeal {
		addr, err := qbft.GetSignatureAddressNoHashing(proposalCommitSeal, seal)
		if err != nil {
			return qbftcommon.ErrInvalidSignature
		}
		commitRewardees = append(commitRewardees, addr)
	}
	log.Trace("Calculating block reward", "currentBlock", header.Number, "calculatingBlock", parentHeader.Number, "prepareReward", prepareRewardees, "commitReward", commitRewardees)
	prepareReward := chain.Config().GetPrepareReward(header.Number)
	commitReward := chain.Config().GetCommitReward(header.Number)

	if prepareRewardFn != nil {
		for _, addr := range prepareRewardees {
			prepareRewardFn(addr, &prepareReward)
		}
	}

	if commitRewardFn != nil {
		for _, addr := range commitRewardees {
			commitRewardFn(addr, &commitReward)
		}
	}

	return nil
}

func mergeSeals(seals [][]byte, extraSeals map[common.Hash][]byte) [][]byte {
	if extraSeals == nil {
		return seals
	}
	mergedSeals := [][]byte{}

	for _, s := range extraSeals {
		mergedSeals = append(mergedSeals, s)
	}
	for _, s := range seals {
		if extraSeals[common.BytesToHash(s)] != nil {
			continue
		}
		mergedSeals = append(mergedSeals, s)
	}
	return mergedSeals
}

func writeValidatorsToEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) error {
	vals, err := decideValidators(e, chain, header, state)
	if err != nil {
		return err
	}
	return ApplyHeaderQBFTExtra(header, WriteValidators(vals))
}

// verifyEpoch is a handler that performs default actions when the block is an EpochBlock,
// and is called during the Finalize process.
// It validates the validity of the ValidatorList associated with the EpochBlock.
func verifyEpoch(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) error {
	vals, err := decideValidators(e, chain, header, state)
	if err != nil {
		return err
	}

	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return err
	}

	if len(vals) != len(extra.Validators) {
		return errors.New("WBFT: mismatch in ValidatorList sizes")
	}

	// Compare each element
	for i := range vals {
		if vals[i] != extra.Validators[i] {
			return errors.New("WBFT: The two validators do not match")
		}
	}
	return nil
}

func decideValidators(e *Engine, chain consensus.ChainHeaderReader, header *types.Header, state govwbft.StateReader) ([]common.Address, error) {
	// WEMIX 3.5:
	// - Stabilization stage:
	//   - validator list(ordered) defined at wbft config
	// - After stabilization stage:
	//   - NCP & Staker
	//   - order: address alphabetically (temporary, vrf random order possible)
	// WEMIX 4.0:
	// - vrf random selection from Stakers
	// - order: vrf random order (selected order)
	//
	// this is a temporary solution; stabilization stage is not implemented yet
	vals := govwbft.NCPStakers(state)
	if len(vals) == 0 { // TODO: check this with config.minStakers rather than zero
		valSet, err := e.GetValidators(chain, header.Number, header.ParentHash, nil)
		if err != nil {
			return nil, err
		}
		vals = valSet.AddressList()
	}
	return vals, nil
}

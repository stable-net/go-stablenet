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
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftcommon "github.com/ethereum/go-ethereum/consensus/qbft/common"
	"github.com/ethereum/go-ethereum/consensus/qbft/core"
	"github.com/ethereum/go-ethereum/consensus/qbft/validator"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

var (
	nilUncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
)

type SignerFn func(data []byte) ([]byte, error)

type Engine struct {
	cfg *qbft.Config

	signer common.Address // Ethereum address of the signing key
	sign   SignerFn       // Signer function to authorize hashes with
}

func NewEngine(cfg *qbft.Config, signer common.Address, sign SignerFn) *Engine {
	return &Engine{
		cfg:    cfg,
		signer: signer,
		sign:   sign,
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
	config := e.cfg.GetConfig(parentHeader.Number)

	if config.EmptyBlockPeriod > config.BlockPeriod && len(block.Transactions()) == 0 {
		if block.Header().Time < parentHeader.Time+config.EmptyBlockPeriod {
			return 0, fmt.Errorf("empty block verification fail")
		}
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

	prevPreparedSeal := extra.PrevPreparedSeal
	if len(prevPreparedSeal) == 0 {
		// prevPreparedSeal validation for monblanc block or first block after genesis is skipped because it's empty
		if chain.Config().MontBlancBlock.Cmp(header.Number) != 0 && number != 1 {
			return qbftcommon.ErrEmptyPrevPreparedSeals
		}
	} else {
		//check whether prevPrepared seals are generated by prevValidators
		var prevPreparers []common.Address
		prevPreparers, err = e.GetSignerAddress(parent, prevPreparedSeal, core.SealTypePrepare)
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
		if chain.Config().MontBlancBlock.Cmp(header.Number) != 0 && number != 1 {
			return qbftcommon.ErrEmptyPrevCommittedSeals
		}
	} else {
		var prevCommitters []common.Address
		prevCommitters, err = e.GetSignerAddress(parent, prevCommittedSeal, core.SealTypeCommit)
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
	preparers, err := e.GetSignerAddress(header, preparedSeal, core.SealTypePrepare)
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
	committers, err := e.GetSignerAddress(header, committedSeal, core.SealTypeCommit)
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

func (e *Engine) Prepare(chain consensus.ChainHeaderReader, header *types.Header, validators qbft.ValidatorSet) error {
	header.Coinbase = common.Address{}
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

	currentBlockNumber := big.NewInt(0).SetUint64(number - 1)
	// ## Wemix QBFT START : removed
	// validatorContract := e.cfg.GetValidatorContractAddress(currentBlockNumber)
	// if validatorContract != (common.Address{}) && e.cfg.GetValidatorSelectionMode(currentBlockNumber) == params.ContractMode {
	// 	return ApplyHeaderQBFTExtra(
	// 		header,
	// 		WriteValidators([]common.Address{}),
	// 	)
	// } else {
	// ## Wemix QBFT EMD
	for _, transition := range e.cfg.Transitions {
		if transition.Block.Cmp(currentBlockNumber) == 0 && len(transition.Validators) > 0 {
			toRemove := make([]qbft.Validator, 0, validators.Size())
			l := validators.List()
			toRemove = append(toRemove, l...)
			for i := range toRemove {
				validators.RemoveValidator(toRemove[i].Address())
			}
			for i := range transition.Validators {
				validators.AddValidator(transition.Validators[i])
			}
			break
		}
	}
	validatorsList := validator.SortedAddresses(validators.List())
	if chain.Config().MontBlancBlock.Cmp(header.Number) == 0 {
		// monblac hardFork block has empty prevCommittedSeal
		return ApplyHeaderQBFTExtra(
			header,
			WriteValidators(validatorsList),
		)
	} else {
		lastCanonicalHeader := chain.GetHeaderByNumber(header.Number.Uint64() - 1)
		extra, err := types.ExtractQBFTExtra(lastCanonicalHeader)
		if err != nil {
			return err
		} else if extra.PreparedSeal == nil {
			// TODO : what if there is not preparedSeal that node collected?
			return qbftcommon.ErrEmptyPreparedSeals
		} else if extra.CommittedSeal == nil {
			// TODO : what if there is not committedSeal that node collected?
			return qbftcommon.ErrEmptyCommittedSeals
		}

		prevPreparedSeal := extra.PreparedSeal
		prevCommittedSeal := extra.CommittedSeal
		// add validators in snapshot to extraData's validators section and lastBlock committers to extraData's prevCommittedSeal section
		return ApplyHeaderQBFTExtra(
			header,
			WriteValidators(validatorsList),
			WritePrevPreparedSeal(prevPreparedSeal),
			WritePrevCommittedSeal(prevCommittedSeal),
		)
	}
}

func WriteValidators(validators []common.Address) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		qbftExtra.Validators = validators
		return nil
	}
}

func WritePrevPreparedSeal(prevPreparedSeal [][]byte) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		qbftExtra.PrevPreparedSeal = prevPreparedSeal
		return nil
	}
}

func WritePrevCommittedSeal(prevCommittedSeal [][]byte) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
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
	// Accumulate any block and uncle rewards and commit the final state root
	e.accumulateRewards(chain, state, header)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = nilUncleHash
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (e *Engine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	e.Finalize(chain, header, state, txs, uncles)
	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil)), nil
}

// Seal generates a new block for the given input block with the local miner's
// seal place on top.
func (e *Engine) Seal(chain consensus.ChainHeaderReader, block *types.Block, validators qbft.ValidatorSet) (*types.Block, error) {
	if _, v := validators.GetByAddress(e.signer); v == nil {
		return block, qbftcommon.ErrUnauthorized
	}

	header := block.Header()
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return block, consensus.ErrUnknownAncestor
	}

	// Set Coinbase
	header.Coinbase = e.signer

	return block.WithSeal(header), nil
}

func (e *Engine) SealHash(header *types.Header) common.Hash {
	header.Coinbase = e.signer
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

func (e *Engine) GetSignerAddress(header *types.Header, signedSeal [][]byte, sealType core.SealType) ([]common.Address, error) {
	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	sealData := core.PrepareSeal(header, extra.Round, sealType)
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
	return e.GetSignerAddress(header, preparedSeal, core.SealTypePrepare)
}

func (e *Engine) CommitSigners(header *types.Header) ([]common.Address, error) {
	extra, err := types.ExtractQBFTExtra(header)
	if err != nil {
		return []common.Address{}, err
	}
	committedSeal := extra.CommittedSeal
	return e.GetSignerAddress(header, committedSeal, core.SealTypeCommit)
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

func (e *Engine) WriteVote(header *types.Header, candidate common.Address, authorize bool) error {
	return ApplyHeaderQBFTExtra(
		header,
		WriteVote(candidate, authorize),
	)
}

func WriteVote(candidate common.Address, authorize bool) ApplyQBFTExtra {
	return func(qbftExtra *types.QBFTExtra) error {
		voteType := types.QBFTDropVote
		if authorize {
			voteType = types.QBFTAuthVote
		}

		vote := &types.ValidatorVote{RecipientAddress: candidate, VoteType: voteType}
		qbftExtra.Vote = vote
		return nil
	}
}

func (e *Engine) ReadVote(header *types.Header) (candidate common.Address, authorize bool, err error) {
	qbftExtra, err := getExtra(header)
	if err != nil {
		return common.Address{}, false, err
	}

	var vote *types.ValidatorVote
	if qbftExtra.Vote == nil {
		vote = &types.ValidatorVote{RecipientAddress: common.Address{}, VoteType: types.QBFTDropVote}
	} else {
		vote = qbftExtra.Vote
	}

	// Tally up the new vote from the validator
	switch {
	case vote.VoteType == types.QBFTAuthVote:
		authorize = true
	case vote.VoteType == types.QBFTDropVote:
		authorize = false
	default:
		return common.Address{}, false, qbftcommon.ErrInvalidVote
	}

	return vote.RecipientAddress, authorize, nil
}

func getExtra(header *types.Header) (*types.QBFTExtra, error) {
	if len(header.Extra) < types.IstanbulExtraVanity {
		// In this scenario, the header extradata only contains client specific information, hence create a new qbftExtra and set vanity
		vanity := append(header.Extra, bytes.Repeat([]byte{0x00}, types.IstanbulExtraVanity-len(header.Extra))...)
		return &types.QBFTExtra{
			VanityData:        vanity,
			Validators:        []common.Address{},
			PreparedSeal:      [][]byte{},
			CommittedSeal:     [][]byte{},
			PrevPreparedSeal:  [][]byte{},
			PrevCommittedSeal: [][]byte{},
			Round:             0,
			Vote:              nil,
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

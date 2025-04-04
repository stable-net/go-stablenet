// Copyright 2017 The go-ethereum Authors
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

package wpoa

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/holiman/uint256"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/wemixgov"
)

var (
	two256                        = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
	maxUncles                     = 2         // Maximum number of uncles allowed in a single block
	allowedFutureBlockTimeSeconds = int64(15) // Max seconds from current time allowed for blocks, before they're considered future blocks
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	ErrNotInitialized = errors.New("not initialized")
	ErrInvalidEnode   = errors.New("invalid enode")
	ErrNotFound       = errors.New("not found")

	errTooManyUncles    = errors.New("too many uncles")
	errDuplicateUncle   = errors.New("duplicate uncle")
	errUncleIsAncestor  = errors.New("uncle is ancestor")
	errDanglingUncle    = errors.New("uncle's parent is not ancestor")
	errUnauthorized     = errors.New("unauthorized block")
	errInvalidMixDigest = errors.New("invalid mix digest")
	errInvalidPoW       = errors.New("invalid proof-of-work")
)

type WemixPoA struct {
	govCli *WemixGov
}

type WemixNode struct {
	Name  string         `json:"name"`
	Enode string         `json:"enode"`
	Id    string         `json:"id"`
	Ip    string         `json:"ip"`
	Port  int            `json:"port"`
	Addr  common.Address `json:"addr"`

	Status string `json:"status"`
	Miner  bool   `json:"miner"`
}

// block build parameters for caching
type BlockBuildParameters struct {
	Height               uint64
	BlockInterval        int64
	MaxBaseFee           *big.Int
	GasLimit             *big.Int
	BaseFeeMaxChangeRate int64
	GasTargetPercentage  int64
}

type WemixMember struct {
	Staker common.Address `json:"address"`
	Reward common.Address `json:"reward"`
	Stake  *big.Int       `json:"stake"`
}

type RewardParameters struct {
	RewardAmount       *big.Int
	Staker             *common.Address
	EcoSystem          *common.Address
	Maintenance        *common.Address
	FeeCollector       *common.Address
	Members            []*WemixMember
	DistributionMethod []*big.Int
	BlocksPer          int64
}

type reward struct {
	Addr   common.Address `json:"addr"`
	Reward *big.Int       `json:"reward"`
}

type WemixGovInfo struct {
	Registry                  common.Address
	Gov                       common.Address
	Staking                   common.Address
	ModifiedBlock             *big.Int
	BlockInterval             *big.Int
	BlocksPer                 *big.Int
	BlockReward               *big.Int
	MaxPriorityFeePerGas      *big.Int
	MaxBaseFee                *big.Int
	GasLimit                  *big.Int
	BaseFeeMaxChangeRate      *big.Int
	GasTargetPercentage       *big.Int
	DefaultBriocheBlockReward *big.Int
	Nodes                     []*WemixNode
}

func NewWemixPoAEngine(backend wemixgov.GovBackend) consensus.Engine {
	wpoa := &WemixPoA{
		govCli: NewWemixGov(backend),
	}

	SetWemixPoA(wpoa)
	return wpoa
}

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (wpoa *WemixPoA) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
func (wpoa *WemixPoA) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	// Short circuit if the header is known, or its parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return wpoa.verifyHeader(chain, header, parent, false, time.Now().Unix())
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (wpoa *WemixPoA) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	if len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}
	abort := make(chan struct{})
	results := make(chan error, len(headers))
	unixNow := time.Now().Unix()

	go func() {
		for i, header := range headers {
			var parent *types.Header
			if i == 0 {
				parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
			} else if headers[i-1].Hash() == headers[i].ParentHash {
				parent = headers[i-1]
			}
			var err error
			if parent == nil {
				err = consensus.ErrUnknownAncestor
			} else {
				err = wpoa.verifyHeader(chain, header, parent, false, unixNow)
			}
			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of the stock Ethereum ethash engine.
func (wpoa *WemixPoA) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// Verify that there are at most 2 uncles included in this block
	if len(block.Uncles()) > maxUncles {
		return errTooManyUncles
	}
	if len(block.Uncles()) == 0 {
		return nil
	}
	// Gather the set of past uncles and ancestors
	uncles, ancestors := mapset.NewSet[common.Hash](), make(map[common.Hash]*types.Header)

	number, parent := block.NumberU64()-1, block.ParentHash()
	for i := 0; i < 7; i++ {
		ancestorHeader := chain.GetHeader(parent, number)
		if ancestorHeader == nil {
			break
		}
		ancestors[parent] = ancestorHeader
		// If the ancestor doesn't have any uncles, we don't have to iterate them
		if ancestorHeader.UncleHash != types.EmptyUncleHash {
			// Need to add those uncles to the banned list too
			ancestor := chain.GetBlock(parent, number)
			if ancestor == nil {
				break
			}
			for _, uncle := range ancestor.Uncles() {
				uncles.Add(uncle.Hash())
			}
		}
		parent, number = ancestorHeader.ParentHash, number-1
	}
	ancestors[block.Hash()] = block.Header()
	uncles.Add(block.Hash())

	// Verify each of the uncles that it's recent, but not an ancestor
	for _, uncle := range block.Uncles() {
		// Make sure every uncle is rewarded only once
		hash := uncle.Hash()
		if uncles.Contains(hash) {
			return errDuplicateUncle
		}
		uncles.Add(hash)

		// Make sure the uncle has a valid ancestry
		if ancestors[hash] != nil {
			return errUncleIsAncestor
		}
		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			return errDanglingUncle
		}
		if err := wpoa.verifyHeader(chain, uncle, ancestors[uncle.ParentHash], true, time.Now().Unix()); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
// See YP section 4.3.4. "Block Header Validity"
func (wpoa *WemixPoA) verifyHeader(chain consensus.ChainHeaderReader, header, parent *types.Header, uncle bool, unixNow int64) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if !uncle {
		if header.Time > uint64(unixNow+allowedFutureBlockTimeSeconds) {
			return consensus.ErrFutureBlock
		}
	}

	// WEMIX poa uses 1 for the difficulty
	expected := big.NewInt(1)
	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	if header.GasLimit > params.MaxGasLimit {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, params.MaxGasLimit)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}
	// Verify the block's gas usage and (if applicable) verify the base fee.
	if !chain.Config().IsLondon(header.Number) {
		// Verify BaseFee not present before EIP-1559 fork.
		if header.BaseFee != nil && header.BaseFee.Cmp(new(big.Int)) > 0 {
			// A block before london hard fork may have `BaseFee` field in WEMIX because
			// rlp.Decode generates the zero big.Int field for `BaseFee` instead of nil.
			return fmt.Errorf("invalid baseFee before fork: have %d, expected 'nil'", header.BaseFee)
		}
		if err := wpoa.VerifyGasLimit(parent.GasLimit, header.GasLimit); err != nil {
			return err
		}
	} else {
		_, _, _, _, gasTargetPercentage, err := wpoa.govCli.GetBlockBuildParameters(parent.Number)
		if errors.Is(err, ErrNotInitialized) {
			return nil
		}
		if err := wpoa.VerifyDynamicGasHeader(chain.Config(), parent, header, uint64(gasTargetPercentage)); err != nil {
			// Verify the header's EIP-1559 attributes.
			return err
		}
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	if chain.Config().IsShanghai(header.Number, header.Time) {
		return errors.New("ethash does not support shanghai fork")
	}
	// Verify the non-existence of withdrawalsHash.
	if header.WithdrawalsHash != nil {
		return fmt.Errorf("invalid withdrawalsHash: have %x, expected nil", header.WithdrawalsHash)
	}
	if chain.Config().IsCancun(header.Number, header.Time) {
		return errors.New("ethash does not support cancun fork")
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

	// verify Wemix block seal
	if err := wpoa.verifySeal(header); err != nil {
		return err
	}

	// If all checks passed, validate any special fields for hard forks
	// WEMIX doesn't verify dao extra data
	// if err := misc.VerifyDAOHeaderExtraData(chain.Config(), header); err != nil {return err}
	// WEMIX: Check if it's generated and signed by a registered node
	if !wpoa.govCli.VerifyBlockSig(header.Number, chain, header.Coinbase, header.MinerNodeId, header.Root, header.MinerNodeSig, chain.Config().IsPangyo(header.Number)) {
		return errUnauthorized
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (wpoa *WemixPoA) CalcDifficulty(consensus.ChainHeaderReader, uint64, *types.Header) *big.Int {
	return big.NewInt(params.FixedDifficulty)
}

// VerifyGaslimit verifies the header gas limit according increase/decrease
// in relation to the parent gas limit.
func (wpoa *WemixPoA) VerifyGasLimit(parentGasLimit, headerGasLimit uint64) error {
	// now WEMIX does not check block gas limit
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
func (wpoa *WemixPoA) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	return errors.New("wpoa `Prepare` is disabled")
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards.
func (wpoa *WemixPoA) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) error {
	// Accumulate any block and uncle rewards
	wpoa.accumulateRewards(chain.Config(), state, header, uncles)
	return nil
}

// FinalizeAndAssemble implements consensus.Engine, accumulating the block and
// uncle rewards, setting the final state and assembling the block.
func (wpoa *WemixPoA) FinalizeAndAssemble(consensus.ChainHeaderReader, *types.Header, *state.StateDB, []*types.Transaction,
	[]*types.Header, []*types.Receipt, []*types.Withdrawal) (*types.Block, error) {
	// Wpoa mining is disabled. If this function returns an error, miner worker will not do anything
	return nil, errors.New("wpoa `FinalizeAndAssemble` is disabled")
}

func (wpoa *WemixPoA) Seal(consensus.ChainHeaderReader, *types.Block, chan<- *types.Block, <-chan struct{}) error {
	return errors.New("wpoa `Seal` is disabled")
}

func (wpoa *WemixPoA) APIs(consensus.ChainHeaderReader) []rpc.API {
	return []rpc.API{}
}

func (wpoa *WemixPoA) Close() error {
	return nil
}

// CallEngineSpecific implements consensus.Engine
func (wpoa *WemixPoA) CallEngineSpecific(method string, args ...interface{}) interface{} {
	return nil
}

// SealHash returns the hash of a block prior to it being sealed.
func (wpoa *WemixPoA) SealHash(*types.Header) (hash common.Hash) {
	// wpoa `SealHash` is disabled
	return common.Hash{}
}

// accumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func (wpoa *WemixPoA) accumulateRewards(config *params.ChainConfig, stateDB *state.StateDB, header *types.Header, uncles []*types.Header) {
	rewards, err := wpoa.calculateRewards(
		config, header.Number, header.Fees,
		func(addr common.Address, amt *big.Int) {
			stateDB.AddBalance(addr, uint256.MustFromBig(amt))
		})
	if err == nil {
		header.Rewards = rewards
	} else {
		if errors.Is(err, ErrNotInitialized) {
			reward := new(big.Int)
			if header.Fees != nil {
				reward.Add(reward, header.Fees)
			}
			stateDB.AddBalance(header.Coinbase, uint256.MustFromBig(reward))
		}
	}
}

func (wpoa *WemixPoA) calculateRewards(config *params.ChainConfig, num, fees *big.Int, addBalance func(common.Address, *big.Int)) ([]byte, error) {
	rp, err := wpoa.govCli.GetRewardParams(big.NewInt(num.Int64() - 1))
	if err != nil {
		// all goes to the coinbase
		return nil, ErrNotInitialized
	}

	return wpoa.calculateRewardsWithParams(config, rp, num, fees, addBalance)
}

func (wpoa *WemixPoA) calculateRewardsWithParams(config *params.ChainConfig, rp *RewardParameters, num, fees *big.Int, addBalance func(common.Address, *big.Int)) (rewards []byte, err error) {
	if (rp.Staker == nil && rp.EcoSystem == nil && rp.Maintenance == nil) || len(rp.Members) == 0 {
		// handle testnet block 94 rewards
		if rewards94 := wpoa.handleBlock94Rewards(num, rp); rewards94 != nil {
			if addBalance != nil {
				for _, i := range rewards94 {
					addBalance(i.Addr, i.Reward)
				}
			}
			rewards, err = json.Marshal(rewards94)
			return
		}
		err = ErrNotInitialized
		return
	}

	var blockReward *big.Int
	if config.IsBrioche(num) {
		blockReward = config.Brioche.GetBriocheBlockReward(params.DefaultBriocheBlockReward, num)
	} else {
		// if the wemix chain is not on brioche hard fork, use the `RewardAmount` from gov contract
		blockReward = new(big.Int).Set(rp.RewardAmount)
	}

	// block reward
	// - not brioche chain: use `EnvStorageImp.getBlockRewardAmount()`
	// - brioche chain
	//   - config.Brioche.BlockReward != nil: config.Brioche.BlockReward
	//   - config.Brioche.BlockReward == nil: 1e18
	//   - apply halving for BlockReward
	rr, errr := wpoa.distributeRewards(num, rp, blockReward, fees)
	if errr != nil {
		err = errr
		return
	}

	if addBalance != nil {
		for _, i := range rr {
			addBalance(i.Addr, i.Reward)
		}
	}

	rewards, err = json.Marshal(rr)
	return
}

var testnetBlock94Rewards = []reward{
	{
		Addr:   common.HexToAddress("0x6f488615e6b462ce8909e9cd34c3f103994ab2fb"),
		Reward: new(big.Int).SetInt64(100000000000000000),
	},
	{
		Addr:   common.HexToAddress("0x6bd26c4a45e7d7cac2a389142f99f12e5713d719"),
		Reward: new(big.Int).SetInt64(250000000000000000),
	},
	{
		Addr:   common.HexToAddress("0x816e30b6c314ba5d1a67b1b54be944ce4554ed87"),
		Reward: new(big.Int).SetInt64(306213253695614752),
	},
}

func (wpoa *WemixPoA) handleBlock94Rewards(height *big.Int, rp *RewardParameters) []reward {
	if height.Int64() == 94 && len(rp.Members) == 0 &&
		bytes.Equal(rp.Staker[:], testnetBlock94Rewards[0].Addr[:]) &&
		bytes.Equal(rp.EcoSystem[:], testnetBlock94Rewards[1].Addr[:]) &&
		bytes.Equal(rp.Maintenance[:], testnetBlock94Rewards[2].Addr[:]) {
		return testnetBlock94Rewards
	}
	return nil
}

// distributeRewards divides the RewardAmount among Members according to their
// stakes, and allocates rewards to Staker, EcoSystem, and Maintenance accounts.
func (wpoa *WemixPoA) distributeRewards(height *big.Int, rp *RewardParameters, blockReward *big.Int, fees *big.Int) ([]reward, error) {
	dm := new(big.Int)
	for i := 0; i < len(rp.DistributionMethod); i++ {
		dm.Add(dm, rp.DistributionMethod[i])
	}
	if dm.Int64() != 10000 {
		return nil, ErrNotInitialized
	}

	v10000 := big.NewInt(10000)
	minerAmount := new(big.Int).Set(blockReward)
	minerAmount.Div(minerAmount.Mul(minerAmount, rp.DistributionMethod[0]), v10000)
	stakerAmount := new(big.Int).Set(blockReward)
	stakerAmount.Div(stakerAmount.Mul(stakerAmount, rp.DistributionMethod[1]), v10000)
	ecoSystemAmount := new(big.Int).Set(blockReward)
	ecoSystemAmount.Div(ecoSystemAmount.Mul(ecoSystemAmount, rp.DistributionMethod[2]), v10000)
	// the rest goes to Maintenance
	maintenanceAmount := new(big.Int).Set(blockReward)
	maintenanceAmount.Sub(maintenanceAmount, minerAmount)
	maintenanceAmount.Sub(maintenanceAmount, stakerAmount)
	maintenanceAmount.Sub(maintenanceAmount, ecoSystemAmount)

	// if FeeCollector is not specified, i.e. nil, fees go to Maintenance
	if rp.FeeCollector == nil {
		maintenanceAmount.Add(maintenanceAmount, fees)
	}

	var rewards []reward
	if n := len(rp.Members); n > 0 {
		stakeTotal, equalStakes := big.NewInt(0), true
		for i := 0; i < n; i++ {
			if equalStakes && i < n-1 && rp.Members[i].Stake.Cmp(rp.Members[i+1].Stake) != 0 {
				equalStakes = false
			}
			stakeTotal.Add(stakeTotal, rp.Members[i].Stake)
		}

		if equalStakes {
			v0, v1 := big.NewInt(0), big.NewInt(1)
			vn := big.NewInt(int64(n))
			b := new(big.Int).Set(minerAmount)
			d := new(big.Int)
			d.Div(b, vn)
			for i := 0; i < n; i++ {
				rewards = append(rewards, reward{
					Addr:   rp.Members[i].Reward,
					Reward: new(big.Int).Set(d),
				})
			}
			d.Mul(d, vn)
			b.Sub(b, d)
			for ix := height.Int64() % int64(n); b.Cmp(v0) > 0; ix = (ix + 1) % int64(n) {
				rewards[ix].Reward.Add(rewards[ix].Reward, v1)
				b.Sub(b, v1)
			}
		} else {
			// rewards distributed according to stakes
			v0, v1 := big.NewInt(0), big.NewInt(1)
			remainder := new(big.Int).Set(minerAmount)
			for i := 0; i < n; i++ {
				memberReward := new(big.Int).Mul(minerAmount, rp.Members[i].Stake)
				memberReward.Div(memberReward, stakeTotal)
				remainder.Sub(remainder, memberReward)
				rewards = append(rewards, reward{
					Addr:   rp.Members[i].Reward,
					Reward: memberReward,
				})
			}
			for ix := height.Int64() % int64(n); remainder.Cmp(v0) > 0; ix = (ix + 1) % int64(n) {
				rewards[ix].Reward.Add(rewards[ix].Reward, v1)
				remainder.Sub(remainder, v1)
			}
		}
	}
	rewards = append(rewards, reward{
		Addr:   *rp.Staker,
		Reward: stakerAmount,
	})
	rewards = append(rewards, reward{
		Addr:   *rp.EcoSystem,
		Reward: ecoSystemAmount,
	})
	rewards = append(rewards, reward{
		Addr:   *rp.Maintenance,
		Reward: maintenanceAmount,
	})
	if rp.FeeCollector != nil {
		rewards = append(rewards, reward{
			Addr:   *rp.FeeCollector,
			Reward: fees,
		})
	}
	return rewards, nil
}

func (wpoa *WemixPoA) verifySeal(header *types.Header) error {
	digest, result := hashimeta(sealHash(header).Bytes(), header.Nonce.Uint64())
	// Verify the calculated values against the ones provided in the header
	if !bytes.Equal(header.MixDigest[:], digest) {
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(two256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

func sealHash(header *types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	enc := []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	}
	if header.BaseFee != nil {
		enc = append(enc, header.BaseFee)
	}
	rlp.Encode(hasher, enc)
	hasher.Sum(hash[:0])
	return hash
}

func hashimeta(hash []byte, nonce uint64) ([]byte, []byte) {
	// Combine header+nonce into a 64 byte seed
	seed := make([]byte, 40)
	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)

	result := crypto.Keccak256(seed)
	return result, result
}

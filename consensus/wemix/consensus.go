package wemix

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftBackend "github.com/ethereum/go-ethereum/consensus/wbft/backend"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

type CroissantConsensus struct {
	legacy consensus.Engine
	wbft   *wbftBackend.Backend
	stopCh chan struct{}
}

func NewCroissantEngine(legacyEngine consensus.Engine, config *wbft.Config, privateKey *ecdsa.PrivateKey, db ethdb.Database) consensus.Engine {
	wbft := wbftBackend.New(config, privateKey, db)

	result := &CroissantConsensus{
		legacy: legacyEngine,
		wbft:   wbft,
	}
	return result
}

func (we *CroissantConsensus) Start(
	config *params.ChainConfig,
	chain consensus.ChainHeaderReader,
	currentBlock func() *types.Block,
	subscribeChainHead func(ch chan<- core.ChainHeadEvent) event.Subscription,
	notifyNewRound func(waitTime time.Duration, round *big.Int)) {
	we.stopCh = make(chan struct{})
	we.wbft.Start(chain, currentBlock, rawdb.HasBadBlock, notifyNewRound)
}

func (we *CroissantConsensus) Stop() {
	we.wbft.Stop()
	close(we.stopCh)
}

func (we *CroissantConsensus) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (we *CroissantConsensus) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	// The fact that this engine is used implies that the Croissant hard fork is configured. (Croissant is not nil)
	if chain.Config().IsCroissant(header.Number) {
		return we.wbft.VerifyHeader(chain, header)
	}
	return we.legacy.VerifyHeader(chain, header)
}

func (we *CroissantConsensus) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	if chain.Config().IsCroissant(headers[0].Number) {
		return we.wbft.VerifyHeaders(chain, headers)
	} else if !chain.Config().IsCroissant(headers[len(headers)-1].Number) {
		return we.legacy.VerifyHeaders(chain, headers)
	}

	abort := make(chan struct{})
	results := make(chan error, len(headers))
	go func() {
		errored := false
		for _, header := range headers {
			var err error
			if errored {
				err = consensus.ErrUnknownAncestor
			} else {
				err = we.VerifyHeader(chain, header)
			}

			if err != nil {
				errored = true
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

func (we *CroissantConsensus) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if chain.Config().IsCroissant(block.Number()) {
		return we.wbft.VerifyUncles(chain, block)
	}
	return we.legacy.VerifyUncles(chain, block)
}

func (we *CroissantConsensus) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	if chain.Config().IsCroissant(header.Number) {
		return we.wbft.Prepare(chain, header)
	}
	return we.legacy.Prepare(chain, header)
}

func (we *CroissantConsensus) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) error {
	if chain.Config().IsCroissant(header.Number) {
		return we.wbft.Finalize(chain, header, state, txs, uncles, withdrawals)
	} else {
		return we.legacy.Finalize(chain, header, state, txs, uncles, withdrawals)
	}
}

func (we *CroissantConsensus) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if chain.Config().IsCroissant(header.Number) {
		return we.wbft.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
	}
	return we.legacy.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
}

func (we *CroissantConsensus) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	if chain.Config().IsCroissant(block.Number()) {
		return we.wbft.Seal(chain, block, results, stop)
	}
	return we.legacy.Seal(chain, block, results, stop)
}

func (we *CroissantConsensus) SealHash(header *types.Header) common.Hash {
	// Wpoa does not support SealHash, so we suppose this block to be a wbft block
	return we.wbft.SealHash(header)
}

func (we *CroissantConsensus) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	num := new(big.Int).Set(parent.Number)
	num = num.Add(num, big.NewInt(1))
	if chain.Config().IsCroissant(num) {
		return we.wbft.CalcDifficulty(chain, time, parent)
	}
	return we.legacy.CalcDifficulty(chain, time, parent)
}

func (we *CroissantConsensus) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	// legacy has no consensus API
	return we.wbft.APIs(chain)
}

func (we *CroissantConsensus) Close() error {
	err := we.legacy.Close()
	if err != nil {
		return err
	}
	return we.wbft.Close()
}

// CallEngineSpecific implements consensus.Engine
func (we *CroissantConsensus) CallEngineSpecific(method string, args ...interface{}) interface{} {
	return we.wbft.CallEngineSpecific(method, args)
}

func (we *CroissantConsensus) NewChainHead() error {
	return we.wbft.NewChainHead()
}

func (we *CroissantConsensus) HandleMsg(address common.Address, data p2p.Msg) (bool, error) {
	return we.wbft.HandleMsg(address, data)
}

func (we *CroissantConsensus) SetBroadcaster(broadcaster consensus.Broadcaster) {
	we.wbft.SetBroadcaster(broadcaster)
}

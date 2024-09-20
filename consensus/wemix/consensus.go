package wemix

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	qbftBackend "github.com/ethereum/go-ethereum/consensus/qbft/backend"
	"github.com/ethereum/go-ethereum/consensus/wpoa"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/wemixgov"
)

type WemixConsensus struct {
	wpoa   consensus.Engine
	wbft   consensus.Engine
	stopCh chan struct{}
}

func NewWemixEngine(backend wemixgov.GovBackend, config *qbft.Config, privateKey *ecdsa.PrivateKey, db ethdb.Database) consensus.Engine {
	wpoa := wpoa.NewWemixPoAEngine(backend)
	wbft := qbftBackend.New(config, privateKey, db)

	return &WemixConsensus{
		wpoa:   wpoa,
		wbft:   wbft,
		stopCh: make(chan struct{}),
	}
}

func (we *WemixConsensus) Start(config *params.ChainConfig, chain consensus.ChainHeaderReader, currentBlock func() *types.Block, subscribeChainHead func(ch chan<- core.ChainHeadEvent) event.Subscription) {
	chainHeadCh := make(chan core.ChainHeadEvent, 10)
	chainHeadSub := subscribeChainHead(chainHeadCh)

	// WEMIX engine is waiting for MontBlanc hard fork then triggers qbft engine and quits its loop
	go func() {
	loop:
		for {
			select {
			case head := <-chainHeadCh:
				if config.IsMontBlanc(head.Block.Number()) {
					log.Info("MontBlanc hard fork is activated. Starting WEMIX BFT engine")
					err := we.wbft.(*qbftBackend.Backend).Start(chain, currentBlock, rawdb.HasBadBlock)
					if err != nil {
						log.Error("cannot start WEMIX consensus engine", "err", err)
					}
					break loop
				}
			case err := <-chainHeadSub.Err():
				log.Warn("wemix consensus engine loop exits abnormally", "err", err)
				break loop
			case <-we.stopCh:
				break loop
			}
		}
		chainHeadSub.Unsubscribe()
	}()
}

func (we *WemixConsensus) Stop() {
	close(we.stopCh)
}

func (we *WemixConsensus) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

func (we *WemixConsensus) VerifyHeader(chain consensus.ChainHeaderReader, header *types.Header) error {
	if chain.Config().IsMontBlanc(header.Number) {
		return we.wbft.VerifyHeader(chain, header)
	}
	return we.wpoa.VerifyHeader(chain, header)
}

func (we *WemixConsensus) VerifyHeaders(chain consensus.ChainHeaderReader, headers []*types.Header) (chan<- struct{}, <-chan error) {
	if chain.Config().IsMontBlanc(headers[0].Number) {
		return we.wbft.VerifyHeaders(chain, headers)
	} else if !chain.Config().IsMontBlanc(headers[len(headers)-1].Number) {
		return we.wpoa.VerifyHeaders(chain, headers)
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

func (we *WemixConsensus) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if chain.Config().IsMontBlanc(block.Number()) {
		return we.wbft.VerifyUncles(chain, block)
	}
	return we.wpoa.VerifyUncles(chain, block)
}

func (we *WemixConsensus) Prepare(chain consensus.ChainHeaderReader, header *types.Header) error {
	if chain.Config().IsMontBlanc(header.Number) {
		return we.wbft.Prepare(chain, header)
	}
	return we.wpoa.Prepare(chain, header)
}

func (we *WemixConsensus) Finalize(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, withdrawals []*types.Withdrawal) {
	if chain.Config().IsMontBlanc(header.Number) {
		we.wbft.Finalize(chain, header, state, txs, uncles, withdrawals)
	} else {
		we.wpoa.Finalize(chain, header, state, txs, uncles, withdrawals)
	}
}

func (we *WemixConsensus) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt, withdrawals []*types.Withdrawal) (*types.Block, error) {
	if chain.Config().IsMontBlanc(header.Number) {
		return we.wbft.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
	}
	return we.wpoa.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts, withdrawals)
}

func (we *WemixConsensus) Seal(chain consensus.ChainHeaderReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	if chain.Config().IsMontBlanc(block.Number()) {
		return we.wbft.Seal(chain, block, results, stop)
	}
	return we.wpoa.Seal(chain, block, results, stop)
}

func (we *WemixConsensus) SealHash(header *types.Header) common.Hash {
	// Wpoa does not support SealHash, so we suppose this block to be a wbft block
	return we.wbft.SealHash(header)
}

func (we *WemixConsensus) CalcDifficulty(chain consensus.ChainHeaderReader, time uint64, parent *types.Header) *big.Int {
	num := new(big.Int).Set(parent.Number)
	num = num.Add(num, big.NewInt(1))
	if chain.Config().IsMontBlanc(num) {
		return we.wbft.CalcDifficulty(chain, time, parent)
	}
	return we.wpoa.CalcDifficulty(chain, time, parent)
}

func (we *WemixConsensus) APIs(chain consensus.ChainHeaderReader) []rpc.API {
	// wpoa has no consensus API
	return we.wbft.APIs(chain)
}

func (we *WemixConsensus) Close() error {
	err := we.wpoa.Close()
	if err != nil {
		return err
	}
	return we.wbft.Close()
}

// CallEngineSpecific implements consensus.Engine
func (we *WemixConsensus) CallEngineSpecific(method string, args ...interface{}) interface{} {
	return nil
}

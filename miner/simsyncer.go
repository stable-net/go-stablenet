package miner

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/params"
)

type simSyncer struct {
	worker              *worker
	workCh              chan *newWorkReq
	resultCh            chan common.Hash
	latestBlockNum      *big.Int
	adjustedBlockPeriod map[uint64]uint64
}

func (ss *simSyncer) Apply(config *qbft.Config, num *big.Int) {
	if ss.adjustedBlockPeriod[num.Uint64()] > 0 {
		config.Transitions = append(config.Transitions, params.Transition{
			Block:              num,
			BlockPeriodSeconds: ss.adjustedBlockPeriod[num.Uint64()],
		})
	}
}

func newSimSyncer(worker *worker) *simSyncer {
	return &simSyncer{
		worker:              worker,
		workCh:              make(chan *newWorkReq),
		resultCh:            make(chan common.Hash),
		latestBlockNum:      new(big.Int),
		adjustedBlockPeriod: make(map[uint64]uint64),
	}
}

func (ss *simSyncer) queueCommitReq(req *newWorkReq) {
	if work, err := ss.worker.prepareWork(&generateParams{timestamp: uint64(req.timestamp), coinbase: ss.worker.etherbase()}); err == nil {
		ss.worker.updateSnapshot(work.copy())
	}
	currentBlock := ss.worker.chain.CurrentBlock()
	if currentBlock.Number.Cmp(ss.latestBlockNum) > 0 {
		ss.latestBlockNum = new(big.Int).Set(currentBlock.Number)
		ss.notifyCommitResult(currentBlock.Hash())
	}
	ss.workCh <- req
}

func (ss *simSyncer) commit() common.Hash {
	for {
		select {
		case req := <-ss.workCh:
			ss.worker.commitWork(req.interrupt, req.timestamp)
		case result := <-ss.resultCh:
			return result
		}
	}
}

func (ss *simSyncer) commitWithPeriod(duration time.Duration) common.Hash {
	for {
		select {
		case req := <-ss.workCh:
			ss.adjustedBlockPeriod[ss.worker.chain.CurrentBlock().Number.Uint64()+1] = uint64(duration.Seconds())
			ss.worker.commitWork(req.interrupt, req.timestamp)
		case result := <-ss.resultCh:
			return result
		}
	}
}

func (ss *simSyncer) notifyCommitResult(head common.Hash) {
	ss.resultCh <- head
}

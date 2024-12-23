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
	adjustedBlockPeriod map[uint64]uint64
	stateTransitions    map[uint64]params.StateTransition
}

func (ss *simSyncer) Apply(chainConfig *params.ChainConfig, config *qbft.Config, num *big.Int) {
	number := num.Uint64()
	if ss.adjustedBlockPeriod[number] > 0 {
		config.Transitions = append(config.Transitions, params.Transition{
			Block:              num,
			BlockPeriodSeconds: ss.adjustedBlockPeriod[number],
		})
	}
	if transition, ok := ss.stateTransitions[number]; ok {
		chainConfig.StateTransitions = append(chainConfig.StateTransitions, transition)
	}
}

func (ss *simSyncer) close() {
	close(ss.resultCh)
	select {
	case <-ss.workCh:
	default:
	}
	close(ss.workCh)
}

func newSimSyncer(worker *worker) *simSyncer {
	return &simSyncer{
		worker:              worker,
		workCh:              make(chan *newWorkReq),
		resultCh:            make(chan common.Hash),
		adjustedBlockPeriod: make(map[uint64]uint64),
		stateTransitions:    make(map[uint64]params.StateTransition),
	}
}

func (ss *simSyncer) queueCommitReq(req *newWorkReq) {
	if work, err := ss.worker.prepareWork(&generateParams{timestamp: uint64(req.timestamp), coinbase: ss.worker.etherbase()}); err == nil {
		ss.worker.updateSnapshot(work.copy())
	}
	currentBlock := ss.worker.chain.CurrentBlock()
	if currentBlock.Number.Sign() > 0 {
		ss.resultCh <- currentBlock.Hash()
	}
	ss.workCh <- req
}

func (ss *simSyncer) commit() common.Hash {
	req := <-ss.workCh
	ss.commitWork(req)
	return <-ss.resultCh
}

func (ss *simSyncer) commitWithPeriod(duration time.Duration) common.Hash {
	req := <-ss.workCh
	ss.adjustedBlockPeriod[ss.worker.chain.CurrentBlock().Number.Uint64()+1] = uint64(duration.Seconds())
	ss.commitWork(req)
	return <-ss.resultCh
}

func (ss *simSyncer) commitWithState(transition params.StateTransition) common.Hash {
	req := <-ss.workCh
	if transition.Block == nil {
		transition.Block = new(big.Int).Add(ss.worker.chain.CurrentBlock().Number, common.Big1)
	}
	ss.stateTransitions[transition.Block.Uint64()] = transition
	ss.commitWork(req)
	return <-ss.resultCh
}

func (ss *simSyncer) commitWork(req *newWorkReq) {
	if err := ss.worker.eth.TxPool().Sync(); err != nil {
		panic(err)
	}
	ss.worker.commitWork(req.interrupt, req.timestamp)
}

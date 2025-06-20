package miner

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	"github.com/ethereum/go-ethereum/params"
)

type simSyncer struct {
	worker              *worker
	workCh              chan *newWorkReq
	resultCh            chan common.Hash
	adjustedBlockPeriod map[uint64]uint64
	upgradeContracts    map[uint64]*params.GovContracts
}

func combineGovContracts(source *params.GovContracts, target *params.GovContracts) {
	if target.GovConfig != nil {
		source.GovConfig = target.GovConfig
	}
	if target.GovStaking != nil {
		source.GovStaking = target.GovStaking
	}
	if target.GovNCP != nil {
		source.GovNCP = target.GovNCP
	}
	if target.GovRewardeeImp != nil {
		source.GovRewardeeImp = target.GovRewardeeImp
	}
}

func (ss *simSyncer) Apply(chainConfig *params.ChainConfig, config *wbft.Config, num *big.Int) {
	number := num.Uint64()
	if ss.adjustedBlockPeriod[number] > 0 {
		config.Transitions = append(config.Transitions, params.Transition{
			Block: num,
			WBFTConfig: &params.WBFTConfig{
				BlockPeriodSeconds: ss.adjustedBlockPeriod[number],
			},
		})
	}
	if upgradeContracts, ok := ss.upgradeContracts[number]; ok {
		if chainConfig.CroissantBlock.Cmp(num) == 0 {
			if chainConfig.Croissant.GovContracts == nil {
				chainConfig.Croissant.GovContracts = new(params.GovContracts)
			}
			combineGovContracts(chainConfig.Croissant.GovContracts, upgradeContracts)
		} else {
			newUpgrade := params.Upgrade{
				Block:        num,
				GovContracts: upgradeContracts,
			}

			config.GovContractUpgrades = append(config.GovContractUpgrades, newUpgrade)
		}
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
		upgradeContracts:    make(map[uint64]*params.GovContracts),
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

func (ss *simSyncer) commitWithState(upgradeContracts *params.GovContracts, num *big.Int) common.Hash {
	req := <-ss.workCh
	if num == nil {
		num = new(big.Int).Add(ss.worker.chain.CurrentBlock().Number, common.Big1)
	}
	ss.upgradeContracts[num.Uint64()] = upgradeContracts
	ss.commitWork(req)
	return <-ss.resultCh
}

func (ss *simSyncer) commitWork(req *newWorkReq) {
	if err := ss.worker.eth.TxPool().Sync(); err != nil {
		panic(err)
	}
	ss.worker.commitWork(req.interrupt, req.timestamp)
}

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

func (ss *simSyncer) Apply(chainConfig *params.ChainConfig, config *qbft.Config, num *big.Int) {
	number := num.Uint64()
	if ss.adjustedBlockPeriod[number] > 0 {
		config.Transitions = append(config.Transitions, params.Transition{
			Block:              num,
			BlockPeriodSeconds: ss.adjustedBlockPeriod[number],
		})
	}
	if upgradeContracts, ok := ss.upgradeContracts[number]; ok {
		if chainConfig.MontBlancBlock.Cmp(num) == 0 {
			if chainConfig.MontBlanc.Init.GovContracts == nil {
				chainConfig.MontBlanc.Init.GovContracts = new(params.GovContracts)
			}
			combineGovContracts(chainConfig.MontBlanc.Init.GovContracts, upgradeContracts)
		} else {
			newUpgrade := params.Upgrade{
				Block:        num,
				GovContracts: upgradeContracts,
			}
			if chainConfig.MontBlanc.Upgrades == nil {
				chainConfig.MontBlanc.Upgrades = make([]params.Upgrade, 0)
				chainConfig.MontBlanc.Upgrades = append(chainConfig.MontBlanc.Upgrades, newUpgrade)
			} else {
				for i, upgrade := range chainConfig.MontBlanc.Upgrades {
					if upgrade.Block.Cmp(num) == 0 {
						combineGovContracts(upgrade.GovContracts, upgradeContracts)
						return
					} else if upgrade.Block.Cmp(num) > 0 {
						chainConfig.MontBlanc.Upgrades = append(chainConfig.MontBlanc.Upgrades, newUpgrade)
						for j := len(chainConfig.MontBlanc.Upgrades) - 1; j > i; j-- {
							chainConfig.MontBlanc.Upgrades[j] = chainConfig.MontBlanc.Upgrades[j-1]
						}
						chainConfig.MontBlanc.Upgrades[i] = newUpgrade
						return
					}
				}
				chainConfig.MontBlanc.Upgrades = append(chainConfig.MontBlanc.Upgrades, newUpgrade)
			}
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

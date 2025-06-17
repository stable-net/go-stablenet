package backend

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/qbft"
	"github.com/ethereum/go-ethereum/consensus/qbft/messages"
	"github.com/ethereum/go-ethereum/consensus/qbft/testutils"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
)

const ANY_PROPOSER int = 12345678
const ANY_ROUND uint32 = 12345678

type testEnv struct {
	addrs          []common.Address
	index          map[common.Address]int
	chains         map[common.Address]*core.BlockChain
	engines        map[common.Address]*Backend
	newRoundReady  map[common.Address]chan uint64
	roundStartChan map[common.Address]chan struct{}
	results        map[common.Address]chan *types.Block
	subResult      chan *types.Block // a block from Enqueue rather than Seal; only for a proposer
	parent         *types.Block
	stopSealingCh  chan struct{}

	// properties for test scenario
	down        map[common.Address]bool
	msgDisabled map[common.Address]map[uint64]bool
}

// make n engines with given genesis and cfg
func MakeMultiEngineTestEnv(n int) (env *testEnv) {
	env = &testEnv{}

	// validators are ordered so the proposer will be selected in index order
	genesis, nodeKeys, _ := testutils.GenesisAndFixedKeys(n)
	env.parent = genesis.ToBlock()
	env.down = make(map[common.Address]bool)
	env.msgDisabled = make(map[common.Address]map[uint64]bool)
	env.results = make(map[common.Address]chan *types.Block)

	config := new(qbft.Config)
	qbft.SetConfigFromChainConfig(config, genesis.Config)
	config.BlockPeriod = 1
	config.Epoch = 4
	config.RequestTimeout = 2000
	config.MaxRequestTimeoutSeconds = 2
	config.AllowedFutureBlockTime = 100000000 // to skip future block check; this makes block creation time to be very short

	env.addrs = make([]common.Address, n)
	env.index = make(map[common.Address]int)
	env.chains = make(map[common.Address]*core.BlockChain)
	env.engines = make(map[common.Address]*Backend)
	env.subResult = make(chan *types.Block)
	for i, nodeKey := range nodeKeys {
		addr := crypto.PubkeyToAddress(nodeKey.PublicKey)
		memDB := rawdb.NewMemoryDatabase()
		env.addrs[i] = addr
		env.index[addr] = i
		env.engines[addr] = New(config, nodeKey, memDB)
		env.results[addr] = make(chan *types.Block)

		genesis.MustCommit(memDB, triedb.NewDatabase(memDB, triedb.HashDefaults))

		var err error
		env.chains[addr], err = core.NewBlockChain(memDB, nil, genesis, nil, env.engines[addr], vm.Config{}, nil, nil)
		if err != nil {
			panic(err)
		}
	}
	env.newRoundReady = make(map[common.Address]chan uint64)
	env.roundStartChan = make(map[common.Address]chan struct{})
	for addr, engine := range env.engines {
		engine.broadcaster = makeSimBroadcaster(env, addr)
		env.newRoundReady[addr] = make(chan uint64, 1) // it should have a buffer
		env.roundStartChan[addr] = make(chan struct{})

		// engine tries to `NotifyNewRound` at `Start`, so it will be blocked until we call `waitToSync`
		go engine.Start(env.chains[addr], env.chains[addr].CurrentFullBlock, rawdb.HasBadBlock, makeNotifyNewRound(env.newRoundReady[addr], env.roundStartChan[addr]))
	}
	return
}

func makeBlockNoNewChainHead(chain *core.BlockChain, engine *Backend, parent *types.Block) *types.Block {
	header := makeHeader(chain.Config(), engine.config, parent)
	engine.Prepare(chain, header)
	block := types.NewBlock(header, nil, nil, nil, trie.NewStackTrie(nil))
	return block
}

type scenarioFunc func(index int) string

type scenario struct {
	target []int
	set    scenarioFunc
}

// all engines are waiting for this to be called
func (env *testEnv) GoNewRound(t *testing.T, sc *scenario, rounds ...uint64) {
	// wait for new round

	actualRound := make([]uint64, len(env.addrs))
	i := 0
	for addr, newRoundReady := range env.newRoundReady {
		round := <-newRoundReady
		if len(rounds) > 0 && rounds[env.index[addr]] != round {
			t.Errorf("rounds are mismatch: have %d, want %d", round, rounds[env.index[addr]])
		}
		actualRound[i] = round
		i++
	}

	// define scenario if any exists
	if sc != nil {
		message := ""
		for _, index := range sc.target {
			message += sc.set(index) + " "
		}
		t.Logf("[NEW ROUND]: %d with scenario: %s", actualRound, message)
	} else {
		t.Logf("[NEW ROUND]: %d without scenario", actualRound)
	}

	// round go ahead
	for _, roundStartChan := range env.roundStartChan {
		roundStartChan <- struct{}{}
	}

	// commit new work
	env.commitNewWork()
}

func (env *testEnv) commitNewWork() {
	env.stopSealingCh = make(chan struct{})
	for _, engine := range env.engines {
		chain := env.chains[engine.address]
		block := makeBlockNoNewChainHead(chain, engine, chain.CurrentFullBlock())
		currState, _ := chain.State()
		block, _ = engine.FinalizeAndAssemble(chain, block.Header(), currState, nil, nil, nil, nil)
		// all engines try to seal
		engine.Seal(chain, block, env.results[engine.address], env.stopSealingCh)
	}
}

func (env *testEnv) MustSucceed(t *testing.T, allowRoundChange bool, expectedProposer int, expectedRound uint32) *types.Block {
	var result *types.Block
	var proposer common.Address
	timeoutCh := make(chan struct{})
	timer := time.AfterFunc(500*time.Millisecond, func() {
		close(timeoutCh)
	})

	for _, addr := range env.addrs {
		go func(addr common.Address) {
			block := <-env.results[addr]
			if block != nil {
				result = block
				proposer = addr
				close(env.stopSealingCh) // stop other `Seal`
			}
		}(addr)
	}

	select {
	case <-env.stopSealingCh:
		env.stopSealingCh = nil
	case result = <-env.subResult:
		// special case: read simBroadcaster.Enqueue comment
		close(env.stopSealingCh)
		env.stopSealingCh = nil
		proposer = result.Coinbase()
	case <-timeoutCh:
		close(env.stopSealingCh)
		env.stopSealingCh = nil
		if allowRoundChange {
			t.Logf("Round changing. Go next round")
			return nil
		}
		t.Errorf("Block producing is timeout")
		return nil
	}
	timer.Stop()

	if result.ParentHash() != env.parent.Hash() {
		t.Errorf("parent hash mismatch: have %v, want %v", result.ParentHash(), env.parent.Hash())
		return nil
	}
	if _, err := env.chains[proposer].InsertChain(types.Blocks{result}); err != nil {
		t.Errorf("failed to insert a block. err %v", err)
	}
	env.engines[proposer].NewChainHead() // progress to next sequence

	// propagate block to down engine
	for _, addr := range env.addrs {
		if env.down[addr] {
			env.chains[addr].InsertChain(types.Blocks{result})
			env.engines[addr].NewChainHead() // progress to next sequence
		}
	}

	env.parent = result

	if expectedProposer != ANY_PROPOSER && expectedProposer != env.index[proposer] {
		t.Errorf("unexpected proposer (expected=%d, got=%d)", expectedProposer, env.index[proposer])
	}

	extra, err := types.ExtractQBFTExtra(result.Header())
	if err != nil {
		t.Errorf("error on extracting extra: %v", err)
	}
	if expectedRound != ANY_ROUND {
		if expectedRound != extra.Round {
			t.Errorf("unexpected round in header (expected=%d, got=%d)", expectedRound, extra.Round)
		}
	}

	t.Logf("[RESULT] A block is created successfully. (proposer=%d, round=%d)", env.index[proposer], extra.Round)
	return result
}

func (env *testEnv) makeScenarioEngineDown(index ...int) *scenario {
	return &scenario{
		target: append([]int{}, index...),
		set: func(i int) string {
			env.down[env.addrs[i]] = true
			return fmt.Sprintf("[%d:down]", i)
		},
	}
}

func (env *testEnv) makeScenarioEngineUp(index ...int) *scenario {
	return &scenario{
		target: append([]int{}, index...),
		set: func(i int) string {
			env.down[env.addrs[i]] = false
			return fmt.Sprintf("[%d:up]", i)
		},
	}
}

func (env *testEnv) makeScenarioDisableCommitMsg(index ...int) *scenario {
	return &scenario{
		target: append([]int{}, index...),
		set: func(i int) string {
			if env.msgDisabled[env.addrs[i]] == nil {
				env.msgDisabled[env.addrs[i]] = make(map[uint64]bool)
			}
			env.msgDisabled[env.addrs[i]][messages.CommitCode] = true
			return fmt.Sprintf("[%d:commit_msg_disabled]", i)
		},
	}
}

func (env *testEnv) makeScenarioEnableCommitMsg(index ...int) *scenario {
	return &scenario{
		target: append([]int{}, index...),
		set: func(i int) string {
			if env.msgDisabled[env.addrs[i]] == nil {
				env.msgDisabled[env.addrs[i]] = make(map[uint64]bool)
			}
			env.msgDisabled[env.addrs[i]][messages.CommitCode] = false
			return fmt.Sprintf("[%d:commit_msg_enabled]", i)
		},
	}
}

func (env *testEnv) makeScenarioOnlyOneDown(index int) *scenario {
	return &scenario{
		target: append([]int{}, index),
		set: func(i int) string {
			for j := 0; j < len(env.addrs); j++ {
				if j == i {
					env.down[env.addrs[j]] = true
				} else {
					env.down[env.addrs[j]] = false
				}
			}
			return fmt.Sprintf("[only %d:down]", i)
		},
	}
}

func (env *testEnv) makeScenarioRandomDown(index ...int) *scenario {
	x := rand.Int() % len(index)
	return &scenario{
		target: append([]int{}, index...),
		set: func(i int) string {
			if i == x {
				env.down[env.addrs[i]] = true
				return fmt.Sprintf("[%d:down]", i)
			} else {
				env.down[env.addrs[i]] = false
				return fmt.Sprintf("[%d:up]", i)
			}
		},
	}
}

func makeNotifyNewRound(notifyChan chan uint64, startChan chan struct{}) func(waitTime time.Duration, round *big.Int) {
	return func(waitTime time.Duration, round *big.Int) {
		notifyChan <- round.Uint64()
		<-startChan
	}
}

type simBroadcaster struct {
	env      *testEnv
	myAddr   common.Address
	myChain  *core.BlockChain
	myEngine *Backend
	peers    map[common.Address]consensus.Peer
}

func makeSimBroadcaster(env *testEnv, myAddr common.Address) *simBroadcaster {
	simBr := &simBroadcaster{
		env:      env,
		myAddr:   myAddr,
		myChain:  env.chains[myAddr],
		myEngine: env.engines[myAddr],
	}
	peers := make(map[common.Address]consensus.Peer)
	for addr, engine := range env.engines {
		if addr == myAddr {
			continue
		}
		peers[addr] = &simPeer{
			br:         simBr,
			myAddr:     myAddr,
			peerAddr:   addr,
			peerEngine: engine,
		}
	}
	simBr.peers = peers
	return simBr
}

func (sb *simBroadcaster) Enqueue(id string, block *types.Block) {
	if sb.myAddr == block.Coinbase() {
		// SPECIAL CASE: a proposer seals a block but committed block is different with that.
		// committed block should be proposed prior round and it is possible.
		// in this case, Seal() does not receive the result so we should use the enqueued block for the result
		sb.env.subResult <- block
	}
	sb.myChain.InsertChain(types.Blocks{block})
	sb.myEngine.NewChainHead() // progress to next sequence
}

func (sb *simBroadcaster) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for addr, p := range sb.peers {
		if targets[addr] {
			m[addr] = p
		}
	}
	return m
}

type simPeer struct {
	br         *simBroadcaster
	myAddr     common.Address
	peerAddr   common.Address
	peerEngine *Backend
}

func (sp *simPeer) SendQBFTConsensus(msgcode uint64, payload []byte) error {
	if (sp.br.env.msgDisabled[sp.br.myAddr] != nil && sp.br.env.msgDisabled[sp.br.myAddr][msgcode]) ||
		(sp.br.env.msgDisabled[sp.peerAddr] != nil && sp.br.env.msgDisabled[sp.peerAddr][msgcode]) ||
		sp.br.env.down[sp.br.myAddr] || sp.br.env.down[sp.peerAddr] {
		return nil
	}
	if err := sp.peerEngine.istanbulEventMux.Post(qbft.MessageEvent{
		Code:    msgcode,
		Payload: payload,
	}); err != nil {
		return err
	}
	return nil
}

func TestWBFTSimpleCase(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	for i := 0; i < 5; i++ {
		// wait for all engine to ready for new round
		env.GoNewRound(t, nil, 0, 0, 0, 0)
		env.MustSucceed(t, false, i%4, 0)
	}
}

func TestWBFTOneEngineDown(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.GoNewRound(t, nil, 0, 0, 0, 0)
	env.MustSucceed(t, false, 0, 0)

	// make second block with 3 normal engine and 1 disconnected engine
	env.GoNewRound(t, env.makeScenarioEngineDown(3), 0, 0, 0, 0)
	env.MustSucceed(t, false, 1, 0)
}

func TestWBFTTwoEngineDown(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.GoNewRound(t, nil, 0, 0, 0, 0)
	env.MustSucceed(t, false, 0, 0) // proposer is engine 0

	// make second block with 2 normal engine and 2 disconnected engine
	env.GoNewRound(t, env.makeScenarioEngineDown(2, 3), 0, 0, 0, 0)

	env.GoNewRound(t, nil, 1, 1, 1, 1)
	t.Log("round changed to 1")

	env.GoNewRound(t, nil, 2, 2, 2, 2)
	t.Log("round changed to 2")

	env.GoNewRound(t, env.makeScenarioEngineUp(2), 3, 3, 3, 3)
	t.Log("round changed to 3")
	env.MustSucceed(t, false, 0, 3) // proposer is engine 0 again after circulation
}

func TestWBFTPreparedAndRoundChange(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.GoNewRound(t, nil, 0, 0, 0, 0)
	env.MustSucceed(t, false, 0, 0) // proposer is engine 0

	// make second block and 2 engines do not send commit => prepared and round change
	env.GoNewRound(t, env.makeScenarioDisableCommitMsg(2, 3), 0, 0, 0, 0)

	env.GoNewRound(t, nil, 1, 1, 1, 1)
	t.Log("round changed to 1")

	env.GoNewRound(t, env.makeScenarioEnableCommitMsg(2), 2, 2, 2, 2)
	t.Log("round changed to 2")
	env.MustSucceed(t, false, 1, 2) // proposer index should be 1 (second proposer); round changed twice but first proposal is prepared!
}

func TestWBFTRandomEngineDown(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	for i := 0; i < 10; i++ {
		env.GoNewRound(t, env.makeScenarioRandomDown(0, 1, 2, 3))
		result := env.MustSucceed(t, true, ANY_PROPOSER, ANY_ROUND)
		for result == nil {
			env.GoNewRound(t, nil)
			result = env.MustSucceed(t, true, ANY_PROPOSER, ANY_ROUND)
		}
	}
}

func TestWBFTExtraSeal(t *testing.T) {
	env := MakeMultiEngineTestEnv(4)

	// make first block with 4 normal engine
	env.GoNewRound(t, nil, 0, 0, 0, 0)
	env.MustSucceed(t, false, 0, 0) // proposer is engine 0

	// engine 1 down and round change
	env.GoNewRound(t, env.makeScenarioOnlyOneDown(1), 0, 0, 0, 0)

	// engine 2 down and round change
	env.GoNewRound(t, env.makeScenarioOnlyOneDown(2), 1, 1, 1, 1)

	// engine 3 down and round change
	env.GoNewRound(t, env.makeScenarioOnlyOneDown(3), 2, 2, 2, 2)

	// engine 0 down and round change
	env.GoNewRound(t, env.makeScenarioOnlyOneDown(0), 3, 3, 3, 3)

	// all engines up
	env.GoNewRound(t, nil, 4, 4, 4, 4)
	block := env.MustSucceed(t, false, 1, 4)
	extra, err := types.ExtractQBFTExtra(block.Header())
	if err != nil {
		t.Errorf("failed to extract extra: %v", err)
	}
	if len(extra.PrevPreparedSeal.Sealers.GetSealers()) != 4 {
		t.Errorf("lack of PrevPreparedSeal: have %d, want %d", len(extra.PrevPreparedSeal.Sealers.GetSealers()), 4)
	}
}

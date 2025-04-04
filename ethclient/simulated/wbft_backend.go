package simulated

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	qbftbackend "github.com/ethereum/go-ethereum/consensus/qbft/backend"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// WbftBackend is a simulated blockchain for WBFT. You can use it to test your contracts or
// other code that interacts with the Ethereum chain.
type WbftBackend struct {
	eth    *eth.Ethereum
	client WbftClient
}

type WbftClient simClient

// NewWbftBackend creates a new simulated blockchain for WBFT that can be used as a backend for
// contract bindings in unit tests.
//
// A simulated backend always uses chainID 1337.
func NewWbftBackend(alloc types.GenesisAlloc, options ...func(nodeConf *node.Config, ethConf *ethconfig.Config)) *WbftBackend {
	// Create the default configurations for the outer node shell and the Ethereum
	// service to mutate with the options afterwards
	nodeConf := node.DefaultConfig
	nodeConf.DataDir = ""
	nodeConf.P2P = p2p.Config{
		PrivateKey:  nodeConf.NodeKey(),
		NoDiscovery: true,
	}

	ethConf := ethconfig.Defaults
	ethConf.Genesis = &core.Genesis{
		Config:     params.TestChainConfig,
		GasLimit:   ethconfig.Defaults.Miner.GasCeil,
		Alloc:      alloc,
		Difficulty: new(big.Int).SetUint64(1),
		BaseFee:    big.NewInt(1000000000),
	}
	ethConf.Istanbul.AllowedFutureBlockTime = 3153600000 // disable time verification of a block ( == 100 years )
	ethConf.Genesis.Config.QBFT.BlockPeriodSeconds = 1
	ethConf.Genesis.Config.QBFT.EpochLength = 1000
	ethConf.Genesis.Config.QBFT.MinStakers = 999
	validator := crypto.PubkeyToAddress(nodeConf.P2P.PrivateKey.PublicKey)
	blsKey, _ := bls.DeriveFromECDSA(nodeConf.P2P.PrivateKey)
	blsPubKey := blsKey.PublicKey().Marshal()
	ethConf.Genesis.ExtraData = genExtraData(validator, blsPubKey) // simulated chain block
	ethConf.SyncMode = downloader.FullSync
	ethConf.Miner.GasPrice = big.NewInt(1)
	ethConf.Miner.SimulatedEnabled = true
	ethConf.Miner.Recommit = 10 * time.Second // prevent block interruption
	ethConf.TxPool.NoLocals = true

	for _, option := range options {
		option(&nodeConf, &ethConf)
	}

	// Assemble the Ethereum stack to run the chain with
	stack, err := node.New(&nodeConf)
	if err != nil {
		panic(err) // this should never happen
	}

	sim, err := newWbftWithNode(stack, &ethConf)
	if err != nil {
		panic(err) // this should never happen
	}

	return sim
}

func genExtraData(validator common.Address, blsPubKey []byte) []byte {
	sampleExtra := &types.QBFTExtra{
		VanityData: []byte("WEMIX MontBlanc chain block"),
		EpochInfo: &types.EpochInfo{
			Stakers: []*types.Staker{
				{Addr: validator, Diligence: types.DefaultDiligence},
			},
			Validators:    []uint32{0},
			BLSPublicKeys: [][]byte{blsPubKey},
		},
		Round: 0,
	}
	b, _ := rlp.EncodeToBytes(sampleExtra)
	return b
}

// newWithNode sets up a simulated backend on an existing node. The provided node
// must not be started and will be started by this method.
func newWbftWithNode(stack *node.Node, conf *eth.Config) (*WbftBackend, error) {
	backend, err := eth.New(stack, conf)
	if err != nil {
		return nil, err
	}
	// Register the filter system
	filterSystem := filters.NewFilterSystem(backend.APIBackend, filters.Config{})
	stack.RegisterAPIs([]rpc.API{{
		Namespace: "eth",
		Service:   filters.NewFilterAPI(filterSystem, false),
	}})
	// Start the node
	if err := stack.Start(); err != nil {
		return nil, err
	}
	// Start Miner & QBFT Engine
	backend.StartMining()
	qbftEngine := backend.Engine().(*qbftbackend.Backend)
	backend.Miner().InjectSimApplierTo(qbftEngine)
	if !qbftEngine.IsRunning() {
		ticker := time.NewTicker(0.1e9) // 0.1s
		for range ticker.C {
			if qbftEngine.IsRunning() {
				ticker.Stop()
				break
			}
		}
	}

	return &WbftBackend{
		eth:    backend,
		client: WbftClient{ethclient.NewClient(stack.Attach())},
	}, nil
}

// Close shuts down the simWbftBackend.
// The simulated backend can't be used afterwards.
func (n *WbftBackend) Close() error {
	if n.client.Client != nil {
		n.client.Close()
		n.client = WbftClient{}
	}
	if qbftEngine, ok := n.Engine().(*qbftbackend.Backend); ok {
		if err := qbftEngine.Stop(); err != nil {
			return err
		}
	}
	if n.eth.Miner().Mining() {
		n.eth.Miner().Close()
	}
	return nil
}

func (n *WbftBackend) Engine() consensus.Engine {
	return n.eth.Engine()
}

// Commit seals a block and moves the chain forward to a new empty block.
func (n *WbftBackend) Commit() common.Hash {
	return n.eth.Miner().CommitSimulated()
}

func (n *WbftBackend) CommitWithState(transition params.StateTransition) common.Hash {
	return n.eth.Miner().CommitSimulatedWithState(transition)
}

func (n *WbftBackend) AdjustTime(duration time.Duration) common.Hash {
	return n.eth.Miner().CommitSimulatedWithPeriod(duration)
}

// Client returns a client that accesses the simulated chain.
func (n *WbftBackend) Client() Client {
	return n.client
}

// EstimateGas tries to estimate the gas needed to execute a specific transaction based on
// the current pending state of the backend blockchain. There is no guarantee that this is
// the true gas limit requirement as other transactions may be added or removed by miners,
// but it should provide a basis for setting a reasonable default.
func (c WbftClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	var hex hexutil.Uint64
	err := c.Client.Client().CallContext(ctx, &hex, "eth_estimateGas", toCallArg(msg), rpc.PendingBlockNumber)
	if err != nil {
		return 0, err
	}

	return uint64(hex), nil
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["input"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	if msg.GasFeeCap != nil {
		arg["maxFeePerGas"] = (*hexutil.Big)(msg.GasFeeCap)
	}
	if msg.GasTipCap != nil {
		arg["maxPriorityFeePerGas"] = (*hexutil.Big)(msg.GasTipCap)
	}
	if msg.AccessList != nil {
		arg["accessList"] = msg.AccessList
	}
	if msg.FeePayer != nil {
		arg["feePayer"] = msg.FeePayer
	}
	if msg.BlobGasFeeCap != nil {
		arg["maxFeePerBlobGas"] = (*hexutil.Big)(msg.BlobGasFeeCap)
	}
	if msg.BlobHashes != nil {
		arg["blobVersionedHashes"] = msg.BlobHashes
	}
	return arg
}

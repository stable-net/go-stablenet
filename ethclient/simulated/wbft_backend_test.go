package simulated

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func simTestWbftBackend(testAddr common.Address) *WbftBackend {
	return NewWbftBackend(
		types.GenesisAlloc{
			testAddr: {Balance: big.NewInt(10000000000000000)},
		},
	)
}

func TestNewWbftBackend(t *testing.T) {
	sim := NewWbftBackend(types.GenesisAlloc{})
	defer sim.Close()

	client := sim.Client()
	num, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if num != 0 {
		t.Fatalf("expected 0 got %v", num)
	}
	// Create a block
	hash := sim.Commit()
	num, err = client.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if num != 1 {
		t.Fatalf("expected 1 got %v", num)
	}
	block, err := client.BlockByHash(context.Background(), hash)
	if err != nil {
		t.Fatal(err)
	}
	if block.Number().Uint64() != num {
		t.Fatal("committed block number is not 1")
	}
	if block.Hash() != hash {
		t.Fatal("committed block hash is different")
	}

	var result map[string]interface{}
	err = sim.client.Client.Client().CallContext(context.Background(), &result, "istanbul_getWbftExtraInfo", "0x1")

	if err != nil {
		t.Fatalf("istanbul_getWbftExtraInfo failed: %v", err)
	}
}

func TestWbftAdjustTime(t *testing.T) {
	sim := NewWbftBackend(types.GenesisAlloc{})
	defer sim.Close()

	client := sim.Client()

	// Create a block
	sim.Commit()
	block1, _ := client.BlockByNumber(context.Background(), nil)
	prevTime := block1.Time()
	sim.AdjustTime(2 * time.Second)
	block2, _ := client.BlockByNumber(context.Background(), nil)
	newTime := block2.Time()
	if newTime-prevTime < uint64(2*time.Second.Seconds()) {
		t.Errorf("adjusted time not equal to 2 seconds. prev: %v, new: %v", prevTime, newTime)
	}
}

func TestWbftSendTransaction(t *testing.T) {
	sim := simTestWbftBackend(testAddr)
	defer sim.Close()

	client := sim.Client()
	ctx := context.Background()

	signedTx, err := newTx(client, testKey)
	if err != nil {
		t.Errorf("could not create transaction: %v", err)
	}
	// send tx to simulated backend
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		t.Errorf("could not add tx to pending block: %v", err)
	}
	sim.Commit()
	block, err := client.BlockByNumber(ctx, big.NewInt(1))
	if err != nil {
		t.Errorf("could not get block at height 1: %v", err)
	}

	if signedTx.Hash() != block.Transactions()[0].Hash() {
		t.Errorf("did not commit sent transaction. expected hash %v got hash %v", block.Transactions()[0].Hash(), signedTx.Hash())
	}
}

// Tests that the simulator starts with the initial gas limit in the genesis block,
// and that it keeps the same target value.
func TestWbftWithBlockGasLimitOption(t *testing.T) {
	// Construct a simulator, targeting a different gas limit
	sim := NewWbftBackend(types.GenesisAlloc{}, WithBlockGasLimit(12_345_678))
	defer sim.Close()

	client := sim.Client()
	genesis, err := client.BlockByNumber(context.Background(), big.NewInt(0))
	if err != nil {
		t.Fatalf("failed to retrieve genesis block: %v", err)
	}
	if genesis.GasLimit() != 12_345_678 {
		t.Errorf("genesis gas limit mismatch: have %v, want %v", genesis.GasLimit(), 12_345_678)
	}
	// Produce a number of blocks and verify the locked in gas target
	sim.Commit()
	head, err := client.BlockByNumber(context.Background(), big.NewInt(1))
	if err != nil {
		t.Fatalf("failed to retrieve head block: %v", err)
	}
	if head.GasLimit() != 12_345_678 {
		t.Errorf("head gas limit mismatch: have %v, want %v", head.GasLimit(), 12_345_678)
	}
}

// Tests that the simulator honors the RPC call caps set by the options.
func TestWbftWithCallGasLimitOption(t *testing.T) {
	// Construct a simulator, targeting a different gas limit
	sim := NewWbftBackend(types.GenesisAlloc{
		testAddr: {Balance: big.NewInt(10000000000000000)},
	}, WithCallGasLimit(params.TxGas-1))
	defer sim.Close()

	client := sim.Client()
	_, err := client.CallContract(context.Background(), ethereum.CallMsg{
		From: testAddr,
		To:   &testAddr,
		Gas:  21000000,
	}, nil)
	if !strings.Contains(err.Error(), core.ErrIntrinsicGas.Error()) {
		t.Fatalf("error mismatch: have %v, want %v", err, core.ErrIntrinsicGas)
	}
}

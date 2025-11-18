package test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/systemcontracts"
	compile "github.com/ethereum/go-ethereum/systemcontracts/compile/compiler"
	"github.com/stretchr/testify/require"
)

func TestTransferLog(t *testing.T) {
	var (
		ctx = context.TODO()

		masterMinter = NewEOA()
		minter       = NewEOA()
		sender       = NewEOA()
		recipient    = NewEOA()

		amount = towei(1_000_000)
	)

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		minter.Address: {Balance: new(big.Int).Mul(amount, common.Big3)},
	}, nil, func(coinAdapter *params.SystemContract) {
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MASTER_MINTER] = masterMinter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTERS] = minter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTER_ALLOWED] = MAX_UINT_128.String()
	}, nil, nil, nil, nil)
	require.NoError(t, err)

	checkGasFeeTransferEvent := func(event map[string]interface{}, sender common.Address, receipt *types.Receipt) {
		blockNumber := receipt.BlockNumber
		block, err := g.backend.Client().BlockByNumber(ctx, blockNumber)
		require.NoError(t, err)

		coinBase := block.Coinbase()
		diff := new(big.Int).Sub(
			g.BalanceAt(t, ctx, coinBase, blockNumber),
			g.BalanceAt(t, ctx, coinBase, new(big.Int).Sub(blockNumber, common.Big1)),
		)
		actualGas := event["value"].(*big.Int)
		require.Equal(t, sender, event["from"].(common.Address))
		require.Equal(t, coinBase, event["to"].(common.Address))
		require.True(t, diff.Cmp(actualGas) == 0)
		require.True(t, new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice).Cmp(actualGas) == 0)
	}

	t.Run("mint", func(t *testing.T) {
		mintAmount := new(big.Int).Mul(amount, common.Big3)
		receipt, err := g.ExpectedOk(g.Mint(t, minter, sender.Address, mintAmount))
		require.NoError(t, err)

		transferEvents := findEvents("Transfer", receipt.Logs)
		require.Equal(t, 2, len(transferEvents)) // mint, gas

		eventMint := transferEvents[0]
		require.Equal(t, common.Address{}, eventMint["from"].(common.Address))
		require.Equal(t, sender.Address, eventMint["to"].(common.Address))
		require.True(t, mintAmount.Cmp(eventMint["value"].(*big.Int)) == 0)

		checkGasFeeTransferEvent(transferEvents[1], minter.Address, receipt)
	})

	t.Run("tx.value", func(t *testing.T) {
		receipt, err := g.ExpectedOk(TransferCoin(g.backend.Client(), NewTxOpts(t, sender), amount, &recipient.Address))
		require.NoError(t, err)

		transferEvents := findEvents("Transfer", receipt.Logs)
		require.Equal(t, 2, len(transferEvents)) // mint, gas

		eventMint := transferEvents[0]
		require.Equal(t, sender.Address, eventMint["from"].(common.Address))
		require.Equal(t, recipient.Address, eventMint["to"].(common.Address))
		require.True(t, amount.Cmp(eventMint["value"].(*big.Int)) == 0)

		checkGasFeeTransferEvent(transferEvents[1], sender.Address, receipt)
	})

	t.Run("transfer", func(t *testing.T) {
		receipt, err := g.ExpectedOk(g.Transfer(t, sender, recipient.Address, amount))
		require.NoError(t, err)

		transferEvents := findEvents("Transfer", receipt.Logs)
		require.Equal(t, 2, len(transferEvents)) // mint, gas

		eventTransfer := transferEvents[0]
		require.Equal(t, sender.Address, eventTransfer["from"].(common.Address))
		require.Equal(t, recipient.Address, eventTransfer["to"].(common.Address))
		require.True(t, amount.Cmp(eventTransfer["value"].(*big.Int)) == 0)

		checkGasFeeTransferEvent(transferEvents[1], sender.Address, receipt)
	})

	t.Run("transfer 0-value", func(t *testing.T) {
		receipt, err := g.ExpectedOk(g.Transfer(t, sender, recipient.Address, common.Big0))
		require.NoError(t, err)

		transferEvents := findEvents("Transfer", receipt.Logs)
		require.Equal(t, 2, len(transferEvents)) // mint, gas

		eventTransfer := transferEvents[0]
		require.Equal(t, sender.Address, eventTransfer["from"].(common.Address))
		require.Equal(t, recipient.Address, eventTransfer["to"].(common.Address))
		require.True(t, eventTransfer["value"].(*big.Int).Sign() == 0)

		checkGasFeeTransferEvent(transferEvents[1], sender.Address, receipt)
	})

	t.Run("tx.value + transfer", func(t *testing.T) {
		dir := t.TempDir()
		filename := "Test.sol"
		testSource := `pragma solidity ^0.8.0;
		interface CoinAdapter { function transfer(address to, uint256 amount) external; } 
		contract TestContract{
			function transfer(address _to) payable external {
				CoinAdapter(` + TestCoinAdapterAddress.String() + `).transfer(_to, msg.value);
			}
		}`

		require.NoError(t, os.WriteFile(filepath.Join(dir, filename), []byte(testSource), 0700))
		compiled, err := compile.Compile(dir, filepath.Join(dir, filename))
		require.NoError(t, err)

		testContract, err := newBindContract(compiled["TestContract"])
		require.NoError(t, err)

		tcAddr, deployTx, tc, err := testContract.Deploy(g.backend.Client(), g.owner)
		_, err = g.ExpectedOk(deployTx, err)
		require.NoError(t, err)
		// transfer with tx.value
		{
			receipt, err := g.ExpectedOk(tc.Transact(NewTxOptsWithValue(t, recipient, amount), "transfer", sender.Address))
			require.NoError(t, err)

			transferEvents := findEvents("Transfer", receipt.Logs)
			require.Equal(t, 3, len(transferEvents)) // mint, gas

			eventTxValue := transferEvents[0]
			require.Equal(t, recipient.Address, eventTxValue["from"].(common.Address))
			require.Equal(t, tcAddr, eventTxValue["to"].(common.Address))
			require.True(t, amount.Cmp(eventTxValue["value"].(*big.Int)) == 0)

			eventTransfer := transferEvents[1]
			require.Equal(t, tcAddr, eventTransfer["from"].(common.Address))
			require.Equal(t, sender.Address, eventTransfer["to"].(common.Address))
			require.True(t, amount.Cmp(eventTransfer["value"].(*big.Int)) == 0)

			checkGasFeeTransferEvent(transferEvents[2], recipient.Address, receipt)
		}
	})

	t.Run("burn", func(t *testing.T) {
		burnAmount := new(big.Int).Mul(amount, common.Big2)
		receipt, err := g.ExpectedOk(g.Burn(t, minter, burnAmount))
		require.NoError(t, err)

		transferEvents := findEvents("Transfer", receipt.Logs)
		require.Equal(t, 2, len(transferEvents)) // mint, gas

		eventMint := transferEvents[0]
		require.Equal(t, minter.Address, eventMint["from"].(common.Address))
		require.Equal(t, common.Address{}, eventMint["to"].(common.Address))
		require.True(t, burnAmount.Cmp(eventMint["value"].(*big.Int)) == 0)

		checkGasFeeTransferEvent(transferEvents[1], minter.Address, receipt)
	})
}

func TestNativeCoinAdapter(t *testing.T) {
	var (
		ctx                = context.Background()
		masterMinter       = NewEOA()
		minter1            = NewEOA()
		decimals     uint8 = 18

		amount         = toWeiN(10_000, decimals)
		initialBalance = toWeiN(1_000_000, decimals) // for gas
		allowedAmount  = toWeiN(100_000_000, decimals)

		testAccount1 = NewEOA()
		testAccount2 = NewEOA()
	)

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		masterMinter.Address: {Balance: initialBalance},
		minter1.Address:      {Balance: initialBalance},
	}, nil, func(coinAdapter *params.SystemContract) {
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MASTER_MINTER] = masterMinter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTERS] = minter1.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTER_ALLOWED] = allowedAmount.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_DECIMALS] = strconv.Itoa(int(decimals))
	}, nil, nil, nil, nil)
	require.NoError(t, err)

	calcGasCost := func(receipt *types.Receipt) *big.Int {
		return new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
	}

	t.Run("initialize", func(t *testing.T) {
		// masterMinter
		require.Equal(t, masterMinter.Address, contractCall(t, g.coinAdapter, "masterMinter")[0].(common.Address))
		require.False(t, contractCall(t, g.coinAdapter, "isMinter", masterMinter.Address)[0].(bool))

		// minter
		require.True(t, contractCall(t, g.coinAdapter, "isMinter", minter1.Address)[0].(bool))
		require.True(t, allowedAmount.Cmp(g.MinterAllowance(t, minter1.Address)) == 0)

		// decimals
		require.Equal(t, decimals, contractCall(t, g.coinAdapter, "decimals")[0].(uint8))

		// allocation & totalSupply
		{
			ownerAddr := g.owner.From
			ownerBalance := g.BalanceAt(t, ctx, ownerAddr, nil)
			require.True(t, ownerBalance.Cmp(g.BalanceOf(t, ownerAddr)) == 0)

			masterMinterBalance := g.BalanceAt(t, ctx, masterMinter.Address, nil)
			require.True(t, masterMinterBalance.Cmp(initialBalance) == 0)
			require.True(t, masterMinterBalance.Cmp(g.BalanceOf(t, masterMinter.Address)) == 0)

			minter1Balance := g.BalanceAt(t, ctx, minter1.Address, nil)
			require.True(t, minter1Balance.Cmp(initialBalance) == 0)
			require.True(t, minter1Balance.Cmp(g.BalanceOf(t, minter1.Address)) == 0)

			expectedTotalSupply := new(big.Int).Add(ownerBalance, new(big.Int).Add(masterMinterBalance, minter1Balance))
			actualTotalSupply := g.TotalSupply(t)
			require.True(t, actualTotalSupply.Cmp(expectedTotalSupply) == 0)
		}
	})

	t.Run("mint", func(t *testing.T) {
		beforeTotalSupply := g.TotalSupply(t)
		beforeAllowance := g.MinterAllowance(t, minter1.Address)
		require.True(t, g.BalanceOf(t, testAccount1.Address).Sign() == 0)

		// failure case
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.Mint(t, masterMinter, testAccount1.Address, initialBalance)),
				"NativeCoinAdapter: caller is not a minter",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Mint(t, minter1, common.Address{}, initialBalance)),
				"NativeCoinAdapter: mint to the zero address",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Mint(t, minter1, testAccount1.Address, common.Big0)),
				"NativeCoinAdapter: mint amount not greater than 0",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Mint(t, minter1, testAccount1.Address, new(big.Int).Add(allowedAmount, common.Big1))),
				"NativeCoinAdapter: mint amount exceeds minterAllowance",
			)
		}

		receipt, err := g.ExpectedOk(g.Mint(t, minter1, testAccount1.Address, initialBalance))
		require.NoError(t, err)

		require.True(t, initialBalance.Cmp(g.BalanceOf(t, testAccount1.Address)) == 0)
		require.True(t, new(big.Int).Add(beforeTotalSupply, initialBalance).Cmp(g.TotalSupply(t)) == 0)
		require.True(t, new(big.Int).Sub(beforeAllowance, initialBalance).Cmp(g.MinterAllowance(t, minter1.Address)) == 0)

		// mint event
		mintEvent := findEvent("Mint", receipt.Logs)
		require.Equal(t, minter1.Address, mintEvent["minter"])
		require.Equal(t, testAccount1.Address, mintEvent["to"])
		require.True(t, initialBalance.Cmp(mintEvent["amount"].(*big.Int)) == 0)
	})

	t.Run("transfer", func(t *testing.T) {
		from, to := testAccount1, testAccount2
		// failure case
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.Transfer(t, from, common.Address{}, amount)),
				"NativeCoinAdapter: transfer to the zero address",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Transfer(t, from, to.Address, new(big.Int).Add(initialBalance, common.Big1))),
				"NativeCoinAdapter: transfer amount exceeds balance",
			)
		}
		require.True(t, g.BalanceOf(t, to.Address).Sign() == 0)
		beforeBalance := g.BalanceOf(t, from.Address)

		receipt, err := g.ExpectedOk(g.Transfer(t, from, to.Address, amount))
		require.NoError(t, err)

		// before balance - amount - gas cost
		expectedBalance := new(big.Int).Sub(new(big.Int).Sub(beforeBalance, amount), calcGasCost(receipt))
		require.True(t, amount.Cmp(g.BalanceOf(t, to.Address)) == 0)
		require.True(t, expectedBalance.Cmp(g.BalanceOf(t, from.Address)) == 0)

		// transfer event
		transferEvent := findEvent("Transfer", receipt.Logs)
		require.Equal(t, from.Address, transferEvent["from"].(common.Address))
		require.Equal(t, to.Address, transferEvent["to"].(common.Address))
		require.True(t, amount.Cmp(transferEvent["value"].(*big.Int)) == 0)
	})

	t.Run("burn", func(t *testing.T) {
		beforeTotalSupply := g.TotalSupply(t)
		beforeBalance := g.BalanceOf(t, minter1.Address)

		// failure case
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.Burn(t, masterMinter, amount)),
				"NativeCoinAdapter: caller is not a minter",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Burn(t, minter1, common.Big0)),
				"NativeCoinAdapter: burn amount not greater than 0",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Burn(t, minter1, new(big.Int).Add(beforeBalance, common.Big1))),
				"NativeCoinAdapter: burn amount exceeds balance",
			)
		}

		receipt, err := g.ExpectedOk(g.Burn(t, minter1, amount))
		require.NoError(t, err)

		require.True(t, new(big.Int).Sub(beforeTotalSupply, amount).Cmp(g.TotalSupply(t)) == 0)
		// before balance - amount - gas cost
		expectedBalance := new(big.Int).Sub(
			new(big.Int).Sub(beforeBalance, amount),
			new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice), // gas cost
		)
		require.True(t, expectedBalance.Cmp(g.BalanceOf(t, minter1.Address)) == 0)

		// burn event
		burnEvent := findEvent("Burn", receipt.Logs)
		require.Equal(t, minter1.Address, burnEvent["burner"])
		require.True(t, amount.Cmp(burnEvent["amount"].(*big.Int)) == 0)
	})

	t.Run("approve and transferFrom", func(t *testing.T) {
		var (
			owner, spender = testAccount2, testAccount1
			approveAmount  = new(big.Int).Div(amount, big.NewInt(10))
		)

		// failure case - approve
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.Approve(t, owner, common.Address{}, approveAmount)),
				"NativeCoinAdapter: approve to the zero address",
			)
		}
		_, err := g.ExpectedOk(g.Approve(t, owner, spender.Address, approveAmount))
		require.NoError(t, err)

		require.True(t, approveAmount.Cmp(g.Allowance(t, owner.Address, spender.Address)) == 0)

		// failure case - transferFrom
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferFrom(t, minter1, owner.Address, spender.Address, approveAmount)),
				"NativeCoinAdapter: transfer amount exceeds allowance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.TransferFrom(t, spender, owner.Address, spender.Address, new(big.Int).Add(approveAmount, common.Big1))),
				"NativeCoinAdapter: transfer amount exceeds allowance",
			)
		}
		beforeBalanceAccount1 := g.BalanceOf(t, spender.Address)
		beforeBalanceAccount2 := g.BalanceOf(t, owner.Address)

		receipt, err := g.ExpectedOk(g.TransferFrom(t, spender, owner.Address, spender.Address, approveAmount))
		require.NoError(t, err)

		expectedBalanceAccount1 := new(big.Int).Sub(new(big.Int).Add(beforeBalanceAccount1, approveAmount), calcGasCost(receipt))
		expectedBalanceAccount2 := new(big.Int).Sub(beforeBalanceAccount2, approveAmount)
		require.True(t, expectedBalanceAccount1.Cmp(g.BalanceOf(t, spender.Address)) == 0)
		require.True(t, expectedBalanceAccount2.Cmp(g.BalanceOf(t, owner.Address)) == 0)

		// transfer event
		transferEvent := findEvent("Transfer", receipt.Logs)
		require.Equal(t, owner.Address, transferEvent["from"].(common.Address))
		require.Equal(t, spender.Address, transferEvent["to"].(common.Address))
		require.True(t, approveAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
	})

	t.Run("permit", func(t *testing.T) {
		var (
			owner, spender = testAccount2, testAccount1
			approveAmount  = new(big.Int).Div(amount, big.NewInt(10))
		)

		permitSig, r, s, v := g.BuildPermitSig(t, owner, spender.Address, approveAmount, nil) // deadline == MAX_UINT_256
		// permit by r,s,v
		{
			receipt, err := g.ExpectedOk(g.Permit(t, minter1, owner.Address, spender.Address, approveAmount, nil, v, r, s))
			require.NoError(t, err)

			require.True(t, approveAmount.Cmp(g.Allowance(t, owner.Address, spender.Address)) == 0)

			// approval event
			approvalEvent := findEvent("Approval", receipt.Logs)
			require.Equal(t, owner.Address, approvalEvent["owner"].(common.Address))
			require.Equal(t, spender.Address, approvalEvent["spender"].(common.Address))
			require.True(t, approveAmount.Cmp(approvalEvent["value"].(*big.Int)) == 0)
		}

		block, err := g.backend.Client().BlockByNumber(ctx, nil)
		require.NoError(t, err)
		blockTime := new(big.Int).SetUint64(block.Time())

		// failure case
		{
			// replay
			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, minter1, owner.Address, spender.Address, approveAmount, nil, permitSig)),
				"NativeCoinAdapter: invalid signature (EIP2612)",
			)

			expectedFailSig, _, _, _ := g.BuildPermitSig(t, owner, spender.Address, approveAmount, blockTime)

			// expired
			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, spender, owner.Address, spender.Address, approveAmount, blockTime, expectedFailSig)),
				"NativeCoinAdapter: permit is expired",
			)

			// invalid deadline
			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, spender, owner.Address, spender.Address, approveAmount, nil, expectedFailSig)),
				"NativeCoinAdapter: invalid signature (EIP2612)",
			)
		}

		// permit by signature
		{
			newApproveAmount := new(big.Int).Mul(approveAmount, common.Big2)
			deadline := new(big.Int).Add(blockTime, big.NewInt(86_400))
			permitWithDeadlineSig, _, _, _ := g.BuildPermitSig(t, owner, spender.Address, newApproveAmount, deadline)

			receipt, err := g.ExpectedOk(g.Permit(t, spender, owner.Address, spender.Address, newApproveAmount, deadline, permitWithDeadlineSig))
			require.NoError(t, err)

			require.True(t, newApproveAmount.Cmp(g.Allowance(t, owner.Address, spender.Address)) == 0)

			// approval event
			approvalEvent := findEvent("Approval", receipt.Logs)
			require.Equal(t, owner.Address, approvalEvent["owner"].(common.Address))
			require.Equal(t, spender.Address, approvalEvent["spender"].(common.Address))
			require.True(t, newApproveAmount.Cmp(approvalEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("transfer with authorization", func(t *testing.T) {
		var (
			from, to       = testAccount1, testAccount2
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)

		transferNonce := ToBytes32("transfer_1")
		transferSig, r, s, v := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, nil, nil, transferNonce) // 0 - MAX_UINT_256
		// transfer by r,s,v
		{
			balnaceFrom := g.BalanceOf(t, from.Address)
			balnaceTo := g.BalanceOf(t, to.Address)

			receipt, err := g.ExpectedOk(
				g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, nil, nil, transferNonce, v, r, s),
			)
			require.NoError(t, err)

			require.True(t, new(big.Int).Sub(balnaceFrom, transferAmount).Cmp(g.BalanceOf(t, from.Address)) == 0)
			require.True(t, new(big.Int).Add(balnaceTo, transferAmount).Cmp(g.BalanceOf(t, to.Address)) == 0)

			// approval event
			approvalEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, from.Address, approvalEvent["from"].(common.Address))
			require.Equal(t, to.Address, approvalEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(approvalEvent["value"].(*big.Int)) == 0)
		}

		// failure case
		{
			// replay
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, nil, nil, transferNonce, transferSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)

			block, err := g.backend.Client().BlockByNumber(ctx, nil)
			require.NoError(t, err)

			expectedFailNonce := ToBytes32("failed")
			validAfter := new(big.Int).SetUint64(block.Time() + 10)
			validBefore := new(big.Int).SetUint64(block.Time() + 100)
			expectedFailSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, validAfter, validBefore, expectedFailNonce)

			// not yet valid
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: authorization is not yet valid",
			)

			g.AdjustTime(10 * time.Second)

			// invalid signature - invalid to.address
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter1, from.Address, minter1.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: invalid signature (EIP3009)",
			)

			g.AdjustTime(150 * time.Second)
			// expired
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: authorization is expired",
			)
		}

		// transfer by signature
		{
			block, err := g.backend.Client().BlockByNumber(ctx, nil)
			require.NoError(t, err)

			transferNonce = ToBytes32("transfer_2")
			validAfter := new(big.Int).SetUint64(block.Time())
			validBefore := new(big.Int).SetUint64(block.Time() + 100)
			transferSig, _, _, _ = g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, validAfter, validBefore, transferNonce)

			balnaceFrom := g.BalanceOf(t, from.Address)
			balnaceTo := g.BalanceOf(t, to.Address)

			receipt, err := g.ExpectedOk(g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, validAfter, validBefore, transferNonce, transferSig))
			require.NoError(t, err)

			require.True(t, new(big.Int).Sub(balnaceFrom, transferAmount).Cmp(g.BalanceOf(t, from.Address)) == 0)
			require.True(t, new(big.Int).Add(balnaceTo, transferAmount).Cmp(g.BalanceOf(t, to.Address)) == 0)

			// transfer event
			transferEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, from.Address, transferEvent["from"].(common.Address))
			require.Equal(t, to.Address, transferEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("receive with authorization", func(t *testing.T) {
		var (
			from, to       = testAccount1, testAccount2
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)

		receiveNonce := ToBytes32("receive_1")
		receiveSig, r, s, v := g.BuildReceiveWithAuthSig(t, from, to.Address, transferAmount, nil, nil, receiveNonce) // 0 - MAX_UINT_256
		// receive by r,s,v
		{
			balnaceFrom := g.BalanceOf(t, from.Address)
			balnaceTo := g.BalanceOf(t, to.Address)

			receipt, err := g.ExpectedOk(
				g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, nil, nil, receiveNonce, v, r, s),
			)
			require.NoError(t, err)

			require.True(t, new(big.Int).Sub(balnaceFrom, transferAmount).Cmp(g.BalanceOf(t, from.Address)) == 0)
			expectedBalance := new(big.Int).Sub(new(big.Int).Add(balnaceTo, transferAmount), calcGasCost(receipt))
			require.True(t, expectedBalance.Cmp(g.BalanceOf(t, to.Address)) == 0)

			// approval event
			approvalEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, from.Address, approvalEvent["from"].(common.Address))
			require.Equal(t, to.Address, approvalEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(approvalEvent["value"].(*big.Int)) == 0)
		}

		// failure case
		{
			// replay
			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, nil, nil, receiveNonce, receiveSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)

			block, err := g.backend.Client().BlockByNumber(ctx, nil)
			require.NoError(t, err)

			expectedFailNonce := ToBytes32("failed")
			validAfter := new(big.Int).SetUint64(block.Time() + 10)
			validBefore := new(big.Int).SetUint64(block.Time() + 100)
			expectedFailSig, _, _, _ := g.BuildReceiveWithAuthSig(t, from, to.Address, transferAmount, validAfter, validBefore, expectedFailNonce)

			// not yet valid
			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: authorization is not yet valid",
			)

			g.AdjustTime(10 * time.Second)

			// msg.sender != to.address
			ExpectedRevert(t,
				g.ExpectedFail(g.coinAdapter.Transact(
					NewTxOpts(t, minter1),
					"receiveWithAuthorization",
					from.Address,
					to.Address,
					transferAmount,
					validAfter,
					validBefore,
					expectedFailNonce,
					expectedFailSig,
				)),
				"NativeCoinAdapter: caller must be the payee",
			)

			// invalid signature
			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, minter1, from.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: invalid signature (EIP3009)",
			)

			g.AdjustTime(150 * time.Second)
			// expired
			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, validAfter, validBefore, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: authorization is expired",
			)
		}

		// receive by signature
		{
			block, err := g.backend.Client().BlockByNumber(ctx, nil)
			require.NoError(t, err)

			receiveNonce = ToBytes32("receive_2")
			validAfter := new(big.Int).SetUint64(block.Time())
			validBefore := new(big.Int).SetUint64(block.Time() + 100)
			receiveSig, _, _, _ = g.BuildReceiveWithAuthSig(t, from, to.Address, transferAmount, validAfter, validBefore, receiveNonce)

			balnaceFrom := g.BalanceOf(t, from.Address)
			balnaceTo := g.BalanceOf(t, to.Address)

			receipt, err := g.ExpectedOk(
				g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, validAfter, validBefore, receiveNonce, receiveSig),
			)
			require.NoError(t, err)

			require.True(t, new(big.Int).Sub(balnaceFrom, transferAmount).Cmp(g.BalanceOf(t, from.Address)) == 0)
			expectedBalance := new(big.Int).Sub(new(big.Int).Add(balnaceTo, transferAmount), calcGasCost(receipt))
			require.True(t, expectedBalance.Cmp(g.BalanceOf(t, to.Address)) == 0)

			// transfer event
			transferEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, from.Address, transferEvent["from"].(common.Address))
			require.Equal(t, to.Address, transferEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("cancel authorization", func(t *testing.T) {
		var (
			from, to       = testAccount1, testAccount2
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)

		cancelNonce := ToBytes32("cancel_1")
		cancelSig, r, s, v := g.BuildCancelAuthSig(t, from, cancelNonce)
		// cancel by v, r, s
		{
			_, err := g.ExpectedOk(g.CancelAuthorization(t, from, from.Address, cancelNonce, v, r, s))
			require.NoError(t, err)

			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, nil, nil, cancelNonce) // 0 - MAX_UINT_256
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, to, from.Address, to.Address, transferAmount, nil, nil, cancelNonce, transferSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)
		}

		// failure case
		{
			// replay
			ExpectedRevert(t,
				g.ExpectedFail(g.CancelAuthorization(t, from, from.Address, cancelNonce, cancelSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)

			expectedFailNonce := ToBytes32("expectedFail")
			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, nil, nil, expectedFailNonce) // 0 - MAX_UINT_256
			expectedFailSig, _, _, _ := g.BuildCancelAuthSig(t, from, cancelNonce)

			// invalid signature - mismatched nonce
			ExpectedRevert(t,
				g.ExpectedFail(g.CancelAuthorization(t, from, from.Address, ToBytes32("mismatched nonce"), expectedFailSig)),
				"NativeCoinAdapter: invalid signature (EIP3009)",
			)

			// invalid signature - mismatched authorizer
			ExpectedRevert(t,
				g.ExpectedFail(g.CancelAuthorization(t, from, to.Address, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: invalid signature (EIP3009)",
			)

			// transfer
			{
				balnaceFrom := g.BalanceOf(t, from.Address)
				balnaceTo := g.BalanceOf(t, to.Address)

				receipt, err := g.ExpectedOk(g.TransferWithAuthorization(t, minter1, from.Address, to.Address, transferAmount, nil, nil, expectedFailNonce, transferSig))
				require.NoError(t, err)

				require.True(t, new(big.Int).Sub(balnaceFrom, transferAmount).Cmp(g.BalanceOf(t, from.Address)) == 0)
				require.True(t, new(big.Int).Add(balnaceTo, transferAmount).Cmp(g.BalanceOf(t, to.Address)) == 0)

				// transfer event
				transferEvent := findEvent("Transfer", receipt.Logs)
				require.Equal(t, from.Address, transferEvent["from"].(common.Address))
				require.Equal(t, to.Address, transferEvent["to"].(common.Address))
				require.True(t, transferAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
			}

			// already used
			ExpectedRevert(t,
				g.ExpectedFail(g.CancelAuthorization(t, from, from.Address, expectedFailNonce, expectedFailSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)
		}

		// cancel by signature
		{
			cancelNonce := ToBytes32("cancel_2")
			cancelSig, _, _, _ := g.BuildCancelAuthSig(t, from, cancelNonce)

			_, err := g.ExpectedOk(g.CancelAuthorization(t, from, from.Address, cancelNonce, cancelSig))
			require.NoError(t, err)

			transferSig, _, _, _ := g.BuildReceiveWithAuthSig(t, from, to.Address, transferAmount, nil, nil, cancelNonce) // 0 - MAX_UINT_256
			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, nil, nil, cancelNonce, transferSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)
		}
	})

	t.Run("manage minter", func(t *testing.T) {
		// beforeTotalSupply := g.TotalSupply(t)
		beforeAllowance := g.MinterAllowance(t, minter1.Address)
		newAllowance := new(big.Int).Add(beforeAllowance, allowedAmount)

		// configure minter - registered minter
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.ConfigureMinter(t, minter1, minter1.Address, newAllowance)),
				"NativeCoinAdapter: caller is not the masterMinter",
			)

			_, err := g.ExpectedOk(g.ConfigureMinter(t, masterMinter, minter1.Address, newAllowance))
			require.NoError(t, err)
			require.True(t, newAllowance.Cmp(g.MinterAllowance(t, minter1.Address)) == 0)
		}

		// configure minter - new minter
		{
			newMinter := NewEOA()
			require.False(t, g.IsMinter(t, newMinter.Address))
			require.True(t, g.MinterAllowance(t, newMinter.Address).Sign() == 0)

			_, err := g.ExpectedOk(g.ConfigureMinter(t, masterMinter, newMinter.Address, newAllowance))
			require.NoError(t, err)

			require.True(t, g.IsMinter(t, newMinter.Address))
			require.True(t, newAllowance.Cmp(g.MinterAllowance(t, newMinter.Address)) == 0)
		}

		// remove minter
		{
			ExpectedRevert(t,
				g.ExpectedFail(g.RemoveMinter(t, minter1, minter1.Address)),
				"NativeCoinAdapter: caller is not the masterMinter",
			)

			_, err := g.ExpectedOk(g.RemoveMinter(t, masterMinter, minter1.Address))
			require.NoError(t, err)

			require.False(t, g.IsMinter(t, minter1.Address))
			require.True(t, g.MinterAllowance(t, minter1.Address).Sign() == 0)
		}
	})
}

func TestNativeCoinAdapter_Blacklist(t *testing.T) {
	var (
		// ctx                = context.Background()
		masterMinter       = NewEOA()
		minter             = NewEOA()
		decimals     uint8 = 18

		amount         = toWeiN(10_000, decimals)
		initialBalance = toWeiN(1_000_000, decimals) // for gas
		allowedAmount  = toWeiN(100_000_000, decimals)

		normalAccount      = NewEOA()
		blacklistedAccount = NewEOA()

		expectedRevertMsg = "account is blacklisted" // Revert triggered in contract(NativeCoinAdapter)
		expectedErrMsg    = "blacklisted sender"     // Error occurred in EVM
	)

	councilMember := NewEOA()
	g, err := NewGovWBFT(t, types.GenesisAlloc{
		masterMinter.Address:       {Balance: initialBalance},
		minter.Address:             {Balance: initialBalance},
		blacklistedAccount.Address: {Balance: initialBalance, Extra: types.SetBlacklisted(uint64(0))},
		councilMember.Address:      {Balance: initialBalance},
	}, nil, func(coinAdapter *params.SystemContract) {
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MASTER_MINTER] = masterMinter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTERS] = minter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTER_ALLOWED] = allowedAmount.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_DECIMALS] = strconv.Itoa(int(decimals))
	}, nil, nil, func(govCouncil *params.SystemContract) {
		govCouncil.Params = map[string]string{
			systemcontracts.GOV_BASE_PARAM_MEMBERS:        councilMember.Address.String(),
			systemcontracts.GOV_BASE_PARAM_QUORUM:         "1",
			systemcontracts.GOV_BASE_PARAM_EXPIRY:         "86400",
			systemcontracts.GOV_BASE_PARAM_MEMBER_VERSION: "1",
		}
	}, nil)
	require.NoError(t, err)

	t.Run("mint", func(t *testing.T) {
		// not blacklisted
		{
			beforeTotalSupply := g.TotalSupply(t)
			beforeAllowance := g.MinterAllowance(t, minter.Address)
			require.True(t, g.BalanceOf(t, normalAccount.Address).Sign() == 0)

			_, err := g.ExpectedOk(g.Mint(t, minter, normalAccount.Address, initialBalance))
			require.NoError(t, err)

			require.True(t, initialBalance.Cmp(g.BalanceOf(t, normalAccount.Address)) == 0)
			require.True(t, new(big.Int).Add(beforeTotalSupply, initialBalance).Cmp(g.TotalSupply(t)) == 0)
			require.True(t, new(big.Int).Sub(beforeAllowance, initialBalance).Cmp(g.MinterAllowance(t, minter.Address)) == 0)
		}

		beforeBalance := g.BalanceOf(t, blacklistedAccount.Address)

		// mint to blacklisted address
		ExpectedRevert(t, g.ExpectedFail(g.Mint(t, minter, blacklistedAccount.Address, initialBalance)), expectedRevertMsg)

		require.True(t, beforeBalance.Cmp(g.BalanceOf(t, blacklistedAccount.Address)) == 0)
	})

	t.Run("transfer", func(t *testing.T) {
		from, to := normalAccount, blacklistedAccount
		// not blacklisted
		{
			beforeBalance := g.BalanceOf(t, minter.Address)
			_, err := g.ExpectedOk(g.Transfer(t, normalAccount, minter.Address, amount))
			require.NoError(t, err)

			expectedBalance := new(big.Int).Add(beforeBalance, amount)
			require.True(t, expectedBalance.Cmp(g.BalanceOf(t, minter.Address)) == 0)
		}

		// transfer to blacklisted address
		ExpectedRevert(t, g.ExpectedFail(g.Transfer(t, from, to.Address, new(big.Int))), expectedRevertMsg)
		ExpectedRevert(t, g.ExpectedFail(g.Transfer(t, from, to.Address, amount)), expectedRevertMsg)

		// blacklisted address transfer
		ExpectedRevert(t, g.ExpectedFail(g.Transfer(t, to, from.Address, amount)), expectedErrMsg)
	})

	t.Run("approve and transferFrom", func(t *testing.T) {
		var (
			owner, spender = normalAccount, blacklistedAccount
			approveAmount  = new(big.Int).Div(amount, big.NewInt(10))
			transferAmount = new(big.Int).Div(approveAmount, common.Big2)
		)
		// not blacklisted
		{
			_, err := g.ExpectedOk(g.Approve(t, owner, minter.Address, approveAmount))
			require.NoError(t, err)

			require.True(t, approveAmount.Cmp(g.Allowance(t, owner.Address, minter.Address)) == 0)

			_, err = g.ExpectedOk(g.TransferFrom(t, minter, owner.Address, minter.Address, transferAmount))
			require.NoError(t, err)

			_, err = g.ExpectedOk(g.TransferFrom(t, owner, minter.Address, owner.Address, new(big.Int)))
			require.NoError(t, err)
		}
		// owner -> blacklist
		ExpectedRevert(t, g.ExpectedFail(g.Approve(t, owner, spender.Address, approveAmount)), expectedRevertMsg)
		// blacklist -> owner
		ExpectedRevert(t, g.ExpectedFail(g.Approve(t, spender, owner.Address, approveAmount)), expectedErrMsg)

		// transfer to blacklist
		ExpectedRevert(t, g.ExpectedFail(g.TransferFrom(t, minter, owner.Address, spender.Address, transferAmount)), expectedRevertMsg)
		// transfer from blacklist
		ExpectedRevert(t, g.ExpectedFail(g.TransferFrom(t, owner, spender.Address, minter.Address, new(big.Int))), expectedRevertMsg)
		// msg.sender is blacklist
		ExpectedRevert(t, g.ExpectedFail(g.TransferFrom(t, spender, owner.Address, minter.Address, new(big.Int))), expectedErrMsg)
	})

	t.Run("permit", func(t *testing.T) {
		var (
			owner, spender = normalAccount, blacklistedAccount
			approveAmount  = new(big.Int).Div(amount, big.NewInt(10))
		)
		// spender is blacklisted
		{
			permitSig, _, _, _ := g.BuildPermitSig(t, owner, spender.Address, approveAmount, nil) // deadline == MAX_UINT_256

			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, minter, owner.Address, spender.Address, approveAmount, nil, permitSig)),
				expectedRevertMsg,
			)
		}
		// onwer is blacklisted
		{
			permitSig, _, _, _ := g.BuildPermitSig(t, spender, owner.Address, approveAmount, nil) // deadline == MAX_UINT_256

			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, minter, spender.Address, owner.Address, approveAmount, nil, permitSig)),
				expectedRevertMsg,
			)
		}
		// msg.sender is blacklisted - should fail due to Go-level sender validation
		{
			permitSig, _, _, _ := g.BuildPermitSig(t, minter, owner.Address, approveAmount, nil)

			ExpectedRevert(t,
				g.ExpectedFail(g.Permit(t, spender, minter.Address, owner.Address, approveAmount, nil, permitSig)),
				expectedErrMsg,
			)
		}
		// not blacklisted - success case
		{
			permitSig, _, _, _ := g.BuildPermitSig(t, minter, owner.Address, approveAmount, nil)

			receipt, err := g.ExpectedOk(g.Permit(t, minter, minter.Address, owner.Address, approveAmount, nil, permitSig))
			require.NoError(t, err)

			require.True(t, approveAmount.Cmp(g.Allowance(t, minter.Address, owner.Address)) == 0)

			// approval event
			approvalEvent := findEvent("Approval", receipt.Logs)
			require.Equal(t, minter.Address, approvalEvent["owner"].(common.Address))
			require.Equal(t, owner.Address, approvalEvent["spender"].(common.Address))
			require.True(t, approveAmount.Cmp(approvalEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("transfer with authorization", func(t *testing.T) {
		var (
			from, to       = normalAccount, blacklistedAccount
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)
		// transfer to blacklist
		{
			transferNonce := ToBytes32("transfer_1")
			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, nil, nil, transferNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter, from.Address, to.Address, transferAmount, nil, nil, transferNonce, transferSig)),
				expectedRevertMsg,
			)
		}
		// transfer from blacklist
		{
			transferNonce := ToBytes32("transfer_2")
			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, to, from.Address, transferAmount, nil, nil, transferNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, minter, to.Address, from.Address, transferAmount, nil, nil, transferNonce, transferSig)),
				expectedRevertMsg,
			)
		}
		// msg.sender is blacklisted - should fail due to Go-level sender validation
		{
			transferNonce := ToBytes32("transfer_3")
			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, minter.Address, transferAmount, nil, nil, transferNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, to, from.Address, minter.Address, transferAmount, nil, nil, transferNonce, transferSig)),
				expectedErrMsg,
			)
		}
		// not blacklisted - success case
		{
			transferNonce := ToBytes32("transfer_4")
			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, minter.Address, transferAmount, nil, nil, transferNonce)

			receipt, err := g.ExpectedOk(g.TransferWithAuthorization(t, minter, from.Address, minter.Address, transferAmount, nil, nil, transferNonce, transferSig))
			require.NoError(t, err)

			// transfer event
			transferEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, from.Address, transferEvent["from"].(common.Address))
			require.Equal(t, minter.Address, transferEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("receive with authorization", func(t *testing.T) {
		var (
			from, to       = normalAccount, blacklistedAccount
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)
		// msg.sender is blacklisted
		{
			receiveNonce := ToBytes32("receive_1")
			receiveSig, _, _, _ := g.BuildReceiveWithAuthSig(t, from, to.Address, transferAmount, nil, nil, receiveNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, to, from.Address, transferAmount, nil, nil, receiveNonce, receiveSig)),
				expectedErrMsg,
			)
		}
		// receive from blacklist
		{
			receiveNonce := ToBytes32("receive_2")
			receiveSig, _, _, _ := g.BuildReceiveWithAuthSig(t, to, from.Address, transferAmount, nil, nil, receiveNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.ReceiveWithAuthorization(t, from, to.Address, transferAmount, nil, nil, receiveNonce, receiveSig)),
				expectedRevertMsg,
			)
		}
		// not blacklisted
		{
			receiveNonce := ToBytes32("receive_3")
			receiveSig, _, _, _ := g.BuildReceiveWithAuthSig(t, minter, from.Address, transferAmount, nil, nil, receiveNonce)

			receipt, err := g.ExpectedOk(
				g.ReceiveWithAuthorization(t, from, minter.Address, transferAmount, nil, nil, receiveNonce, receiveSig),
			)
			require.NoError(t, err)

			// transfer event
			transferEvent := findEvent("Transfer", receipt.Logs)
			require.Equal(t, minter.Address, transferEvent["from"].(common.Address))
			require.Equal(t, from.Address, transferEvent["to"].(common.Address))
			require.True(t, transferAmount.Cmp(transferEvent["value"].(*big.Int)) == 0)
		}
	})

	t.Run("cancel authorization", func(t *testing.T) {
		var (
			from, to       = normalAccount, minter
			transferAmount = new(big.Int).Div(amount, big.NewInt(10))
		)

		// msg.sender is blacklisted - should fail due to Go-level sender validation
		{
			cancelNonce := ToBytes32("cancel_1")
			cancelSig, _, _, _ := g.BuildCancelAuthSig(t, from, cancelNonce)

			ExpectedRevert(t,
				g.ExpectedFail(g.CancelAuthorization(t, blacklistedAccount, from.Address, cancelNonce, cancelSig)),
				expectedErrMsg,
			)
		}

		// not blacklisted - success case
		{
			cancelNonce := ToBytes32("cancel_2")
			cancelSig, _, _, _ := g.BuildCancelAuthSig(t, from, cancelNonce)

			_, err := g.ExpectedOk(g.CancelAuthorization(t, to, from.Address, cancelNonce, cancelSig))
			require.NoError(t, err)

			transferSig, _, _, _ := g.BuildTransferWithAuthSig(t, from, to.Address, transferAmount, nil, nil, cancelNonce)
			ExpectedRevert(t,
				g.ExpectedFail(g.TransferWithAuthorization(t, to, from.Address, to.Address, transferAmount, nil, nil, cancelNonce, transferSig)),
				"NativeCoinAdapter: authorization is used or canceled",
			)
		}
	})
}

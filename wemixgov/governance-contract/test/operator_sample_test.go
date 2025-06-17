package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"

	"github.com/stretchr/testify/require"
)

func TestOperatorContractMultiSig(t *testing.T) {
	var (
		owner1       = getTxOpt(t, "operatorContractOwner1")
		owner2       = getTxOpt(t, "operatorContractOwner2")
		owner3       = getTxOpt(t, "operatorContractOwner3")
		owner4       = getTxOpt(t, "operatorContractOwner4")
		fundManager  = getTxOpt(t, "fundManager")
		minStaking   = towei(500000)
		feeRate      = new(big.Int).SetUint64(1500)
		ctx          = context.Background()
		rewardAmount = towei(10)
		delegator1   = NewEOA()
		err          error
		receipt      *types.Receipt
	)
	// initiate gov
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		owner1.From:        {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner2.From:        {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner3.From:        {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner4.From:        {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		fundManager.From:   {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		delegator1.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
	})
	require.NoError(t, err)
	defer g.backend.Close()
	setWbftGovConfig(g)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	// deploy operatorSample
	owners := []common.Address{owner1.From, owner2.From, owner3.From, owner4.From}
	fundManagers := []common.Address{fundManager.From}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, fundManagers, common.Big2)

	var s1 = NewTestStakerWithOperatorCA(&CA{Address: operatorSampleAddr})

	t.Run("Register Staker", func(t *testing.T) {
		callOpts := new(bind.CallOpts)
		// 1. owners sends value to operator contract
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), owner1, new(big.Int).Div(minStaking, big.NewInt(4)), &operatorSampleAddr))
		require.NoError(t, err)
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), owner2, new(big.Int).Div(minStaking, big.NewInt(4)), &operatorSampleAddr))
		require.NoError(t, err)
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), owner3, new(big.Int).Div(minStaking, big.NewInt(4)), &operatorSampleAddr))
		require.NoError(t, err)
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), owner4, new(big.Int).Div(minStaking, big.NewInt(4)), &operatorSampleAddr))
		require.NoError(t, err)

		var tracedUnstaked *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedUnstaked}, "unstakedAmount"))
		require.True(t, tracedUnstaked.Cmp(minStaking) == 0)

		//2. registerStaker - failure cases
		t.Run("Failure cases", func(t *testing.T) {
			// 2-1. cannot call function for single owner
			ExpectedRevert(t, g.ExpectedFail(g.SingleOwnerRegisterStaker(owner1, s1, minStaking, feeRate)), "Operator: only wallet can access")

			// 2-2. cannot submit multisig transaction that calls govContract's registerStaker function
			blsPubkey, _ := s1.GetBLSPublicKey()
			blsSig, _ := s1.GetBLSPoPSignature()
			callData, _ := g.stakingContract.Pack("registerStaker", minStaking, s1.Staker.Address, operatorSampleAddr, feeRate, blsPubkey.Marshal(), blsSig.Marshal())
			ExpectedRevert(
				t,
				g.ExpectedFail(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData)),
				"Operator: use proper stake functions",
			)
		})

		//3. registerStaker - success case
		t.Run("Success case", func(t *testing.T) {
			// submit multiSig transaction that calls operatorContract's registerStaker function
			blsPubkey, _ := s1.GetBLSPublicKey()
			blsSig, _ := s1.GetBLSPoPSignature()
			callData, _ := g.operatorContract.Pack("registerStaker", minStaking, s1.Staker.Address, operatorSampleAddr, feeRate, blsPubkey.Marshal(), blsSig.Marshal())
			receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
			txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)
			// confirm and execute
			_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner4, 2, txId)
			require.NoError(t, err)
			require.Equal(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked, minStaking)
		})
	})

	t.Run("Stake", func(t *testing.T) {
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), owner4, minStaking, &operatorSampleAddr))
		require.NoError(t, err)

		t.Run("Failure cases", func(t *testing.T) {
			// 1. cannot call function for single owner
			ExpectedRevert(t, g.ExpectedFail(g.SingleOwnerStake(owner1, minStaking)), "Operator: only wallet can access")

			// 2. cannot submit multisig transaction that calls govContract's stake function
			callData, _ := g.stakingContract.Pack("stake", minStaking)
			ExpectedRevert(
				t,
				g.ExpectedFail(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData)),
				"Operator: use proper stake functions",
			)
		})

		t.Run("Success cases", func(t *testing.T) {
			callData, _ := g.operatorContract.Pack("stake", minStaking)
			receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
			txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)
			// confirm and execute
			_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner4, 2, txId)
			require.NoError(t, err)
			require.Equal(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked, new(big.Int).Mul(minStaking, common.Big2))
		})
	})

	t.Run("Claim for Reward, restake", func(t *testing.T) {
		callOpts := new(bind.CallOpts)
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		t.Run("Failure cases", func(t *testing.T) {
			// cannot directly call claimWithRestake. need to submit tx
			ExpectedRevert(
				t,
				g.ExpectedFail(g.ClaimWithRestake(owner1, s1)),
				"Operator: only wallet can access",
			)

			//cannot submit claim tx whose destination is govStaking contract
			callData, _ := g.stakingContract.Pack("claim", s1.Staker.Address, true)
			ExpectedRevert(
				t,
				g.ExpectedFail(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData)),
				"Operator: use proper claim/withdraw functions",
			)
		})

		t.Run("Success cases", func(t *testing.T) {
			var (
				pendingReward *big.Int
				tracedReward  *big.Int
			)
			// 1. owners must multiSig claimWithRestake function of operator contract
			beforeStakedAmt := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked

			distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
			callData, _ := g.operatorContract.Pack("claimWithRestake", s1.Staker.Address)
			receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
			require.NoError(t, err)
			txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

			receipt, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner4, 2, txId)
			require.NoError(t, err)
			pendingReward = findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingReward"].(*big.Int)
			require.Equal(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked, beforeStakedAmt.Add(beforeStakedAmt, pendingReward))
			require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedReward}, "rewardAmount"))
			require.True(t, tracedReward.Cmp(common.Big0) == 0)
		})
		var claimedReward *big.Int

		t.Run("Claim for Reward, not restake", func(t *testing.T) {
			t.Run("Failure cases", func(t *testing.T) {
				// submitting claimWithoutRestake is banned
				callData, _ := g.operatorContract.Pack("claimWithoutRestake", s1.Staker.Address)
				ExpectedRevert(
					t,
					g.ExpectedFail(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData)),
					"Operator: use proper claim/withdraw functions",
				)

				// cannot submit claim tx whose destination is govStaking contract
				callData, _ = g.stakingContract.Pack("claim", s1.Staker.Address, false)
				ExpectedRevert(
					t,
					g.ExpectedFail(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData)),
					"Operator: use proper claim/withdraw functions",
				)
			})

			t.Run("Success cases", func(t *testing.T) {
				callOpts := new(bind.CallOpts)
				var (
					beforeTracedReward *big.Int
					tracedReward       *big.Int
				)
				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&beforeTracedReward}, "rewardAmount"))
				// owner can call claim function of operator contract right away
				distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
				receipt, err = g.ExpectedOk(g.ClaimWithoutRestake(owner1, s1))
				require.NoError(t, err)
				pendingReward := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingReward"].(*big.Int)

				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedReward}, "rewardAmount"))
				require.Equal(t, tracedReward, beforeTracedReward.Add(beforeTracedReward, pendingReward))
				claimedReward = tracedReward
			})
		})

		t.Run("Withdraw reward", func(t *testing.T) {
			t.Run("Failure case", func(t *testing.T) {
				// only fundManager can access this function
				ExpectedRevert(t, g.ExpectedFail(g.WithdrawRewardAmount(owner1, owner1.From, claimedReward)), "Operator: only fundManager can access")
				// multiSig obviously not works
				callData, _ := g.operatorContract.Pack("withdrawReward", owner4.From, claimedReward)
				receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
				require.NoError(t, err)
				txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

				_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner1, 2, txId)
				ExpectedRevert(t, err, "Operator: transaction failed: Operator: only fundManager can access")
			})
			t.Run("Success case", func(t *testing.T) {
				var tracedReward *big.Int

				beforeBalance := g.balanceAt(t, ctx, owner4.From, nil)
				// fundManager withdraw claimedReward to owner4
				receipt, err = g.ExpectedOk(g.WithdrawRewardAmount(fundManager, owner4.From, claimedReward))
				require.NoError(t, err)
				// check operator traces reward well
				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedReward}, "rewardAmount"))
				require.True(t, tracedReward.Cmp(common.Big0) == 0)
				require.Equal(t, g.balanceAt(t, ctx, owner4.From, nil), beforeBalance.Add(beforeBalance, claimedReward))
			})
		})
		var claimedFee *big.Int

		t.Run("Get fee from undelegation", func(t *testing.T) {
			callOpts := new(bind.CallOpts)
			// 1. delegator1 delegates, distribute reward
			_, err := g.ExpectedOk(g.Delegate(t, delegator1, s1.Staker.Address, minStaking))
			require.NoError(t, err)
			distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)

			// 2. Delegator claims reward, fee will be sent to operatorContract
			beforeBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
			receipt, err = g.ExpectedOk(g.Claim(t, delegator1, s1.Staker.Address, false))
			require.NoError(t, err)
			pendingFee := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingFee"].(*big.Int)
			afterBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
			require.Equal(t, afterBalance, beforeBalance.Add(beforeBalance, pendingFee))

			var tracedFee *big.Int
			require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedFee}, "feeAmount"))
			require.True(t, tracedFee.Cmp(pendingFee) == 0)
			claimedFee = tracedFee
		})

		t.Run("Withdraw fee", func(t *testing.T) {
			//callOpts := new(bind.CallOpts)
			// 1. withdraw fee - failure case
			t.Run("Failure case", func(t *testing.T) {
				// cannot call function for single owner
				ExpectedRevert(t, g.ExpectedFail(g.WithdrawFeeAmount(owner1, owner1.From, claimedFee)), "Operator: only fundManager can access")

				// multiSig obviously not works
				callData, _ := g.operatorContract.Pack("withdrawFee", owner4.From, claimedFee)
				receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
				require.NoError(t, err)
				txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

				_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner1, 2, txId)
				ExpectedRevert(t, err, "Operator: transaction failed: Operator: only fundManager can access")
			})
			//2. withdraw fee - success case
			t.Run("Success case", func(t *testing.T) {
				var tracedFee *big.Int
				beforeBalance := g.balanceAt(t, ctx, owner4.From, nil)
				// withdraw claimedFee to owner4
				_, err = g.ExpectedOk(g.WithdrawFeeAmount(fundManager, owner4.From, claimedFee))
				require.NoError(t, err)

				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedFee}, "feeAmount"))
				require.True(t, tracedFee.Cmp(common.Big0) == 0)
				require.Equal(t, g.balanceAt(t, ctx, owner4.From, nil), beforeBalance.Add(beforeBalance, claimedFee))
			})
		})

		var unbonding *big.Int
		var stakedAmt *big.Int
		t.Run("Unstake and withdraw unstaked amount", func(t *testing.T) {
			stakedAmt = govwbft.UserInfo(TestGovStakingAddress, stateDB, s1.Staker.Address, s1.Staker.Address).StakingAmount
			t.Run("Unstake failure case", func(t *testing.T) {
				// need multiSig
				ExpectedRevert(t, g.ExpectedFail(g.SingleOwnerUnstake(owner1, stakedAmt)), "Operator: only wallet can access")
			})
			// unstake the staked amount
			t.Run("Unstake success case", func(t *testing.T) {
				// 1. owner can multiSig unstake function of operator contract
				callData, _ := g.operatorContract.Pack("unstake", minStaking)
				receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
				require.NoError(t, err)
				txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

				_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner4, 2, txId)
				require.NoError(t, err)

				// 2. owner can submit claim function of governance contract
				callData, _ = g.stakingContract.Pack("unstake", new(big.Int).Sub(stakedAmt, minStaking))
				receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData))
				require.NoError(t, err)
				txId = findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

				receipt, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner4, 2, txId)
				require.NoError(t, err)
				unbonding = findEvents("NewCredential", receipt.Logs)[0]["unbonding"].(*big.Int)
			})
		})

		t.Run("Withdraw unstaked amount from govContract", func(t *testing.T) {
			//callOpts := new(bind.CallOpts)
			g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)

			t.Run("Failure case", func(t *testing.T) {
				// submitting withdraw is banned
				callData, _ := g.operatorContract.Pack("withdraw", common.Big1)
				ExpectedRevert(
					t,
					g.ExpectedFail(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData)),
					"Operator: use proper claim/withdraw functions",
				)

				// cannot submit withdraw tx whose destination is govStaking contract
				callData, _ = g.stakingContract.Pack("withdraw", common.Big1)
				ExpectedRevert(
					t,
					g.ExpectedFail(g.SubmitTransaction(owner1, TestGovStakingAddress, common.Big0, callData)),
					"Operator: use proper claim/withdraw functions",
				)
			})

			t.Run("Success case", func(t *testing.T) {
				//1. withdraw first unstake amount which is minimumStaking
				_, err := g.ExpectedOk(g.WithdrawViaOperatorContract(owner1, common.Big1))
				require.NoError(t, err)

				var tracedUnstaked *big.Int
				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedUnstaked}, "unstakedAmount"))
				require.Equal(t, tracedUnstaked, minStaking)

				// 2. withdraw second unstake amount which is totalStaked-minimumStaking
				_, err = g.ExpectedOk(g.WithdrawViaOperatorContract(owner1, common.Big1))
				require.NoError(t, err)

				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedUnstaked}, "unstakedAmount"))
				require.Equal(t, tracedUnstaked, stakedAmt)
			})
		})

		t.Run("Withdraw unstaked amount from operator contarct", func(t *testing.T) {
			callOpts := new(bind.CallOpts)
			var tracedUnstaked *big.Int
			// 1. withdraw fee - failure case
			t.Run("Failure case", func(t *testing.T) {
				// 1-1. cannot call function for single owner
				ExpectedRevert(t, g.ExpectedFail(g.WithdrawUnstakedAmount(owner1, owner1.From, stakedAmt)), "Operator: only wallet can access")
				// custom withdraw should be failed
				functionSignature := []byte("maliciousFunction(address)")
				methodID := crypto.Keccak256(functionSignature)[:4]
				args := abi.Arguments{
					{Type: mustParseType("address")},
				}
				packedArgs, err := args.Pack(TestGovConfigAddress) // just used as random contract address. nothing to do with govConst contract
				if err != nil {
					panic(fmt.Sprintf("Failed to pack arguments: %v", err))
				}
				callData := append(methodID, packedArgs...)
				ExpectedRevert(t, g.ExpectedFail(g.SubmitTransaction(owner1, operatorSampleAddr, stakedAmt, callData)), "Operator: use proper withdraw functions to transfer value from the contract")
				require.NoError(t, err)
			})
			//2. withdraw fee - success case
			t.Run("Success case", func(t *testing.T) {
				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedUnstaked}, "unstakedAmount"))
				beforeBalance := g.balanceAt(t, ctx, owner4.From, nil)
				// withdraw claimedFee to owner4
				callData, _ := g.operatorContract.Pack("withdrawUnstaked", owner4.From, stakedAmt)
				receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
				require.NoError(t, err)
				txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

				receipt, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner1, 2, txId)
				require.NoError(t, err)

				require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedUnstaked}, "unstakedAmount"))
				require.Equal(t, g.balanceAt(t, ctx, owner4.From, nil), beforeBalance.Add(beforeBalance, stakedAmt))
				require.True(t, tracedUnstaked.Cmp(common.Big0) == 0)
			})
		})
	})
}

func multiSigConfirmTxAndExecute(g *GovWBFT, owners []*bind.TransactOpts, executer *bind.TransactOpts, quorum int, txId *big.Int) (*types.Receipt, error) {
	for i := 0; i < quorum; i++ {
		_, err := g.ExpectedOk(g.ConfirmTransaction(owners[i], txId))
		if err != nil {
			return nil, err
		}
	}
	receipt, err := g.ExpectedOk(g.ExecuteTransaction(executer, txId))
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func TestOperatorContractSingleOwner(t *testing.T) {
	var (
		operatorContractSingleOwner = getTxOpt(t, "operatorContractOwner")
		minStaking                  = towei(500000)
		feeRate                     = new(big.Int).SetUint64(1500)
		ctx                         = context.Background()
		rewardAmount                = towei(10)
		fundManager                 = getTxOpt(t, "fundManager")
		delegator1                  = NewEOA()
	)
	// initiate gov
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		operatorContractSingleOwner.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		delegator1.Address:               {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		fundManager.From:                 {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	// deploy operatorSample
	owners := []common.Address{operatorContractSingleOwner.From}
	fundManagers := []common.Address{fundManager.From}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, fundManagers, new(big.Int))

	var s1 = NewTestStakerWithOperatorCA(&CA{Address: operatorSampleAddr})

	t.Run("Register Staker", func(t *testing.T) {
		// 1. owner sends value to operator contract
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), operatorContractSingleOwner, minStaking, &operatorSampleAddr))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, operatorSampleAddr, nil), minStaking)

		//2. owner executes registerStaker function of operator contract
		_, err := g.ExpectedOk(g.SingleOwnerRegisterStaker(operatorContractSingleOwner, s1, minStaking, feeRate))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, TestGovStakingAddress, nil), minStaking)
		require.True(t, big.NewInt(0).Cmp(g.balanceAt(t, ctx, operatorSampleAddr, nil)) == 0)
	})

	t.Run("Claim for Reward, restake", func(t *testing.T) {
		// disribute reward manually - 10 ether
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		// restake the reward
		_, err := g.ExpectedOk(g.ClaimWithRestake(operatorContractSingleOwner, s1))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, TestGovStakingAddress, nil), new(big.Int).Add(minStaking, rewardAmount))
	})

	var claimedReward *big.Int

	t.Run("Claim for Reward, not restake", func(t *testing.T) {
		// disribute reward manually
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		// transfer reward from rewardee to operator contract
		receipt, err := g.ExpectedOk(g.ClaimWithoutRestake(operatorContractSingleOwner, s1))
		require.NoError(t, err)
		userRewardEvent := findEvents("UserRewardUpdated", receipt.Logs)
		claimedReward = userRewardEvent[0]["pendingReward"].(*big.Int)
		require.Equal(t, userRewardEvent[0]["pendingReward"], g.balanceAt(t, ctx, operatorSampleAddr, nil))
	})

	t.Run("Withdraw reward", func(t *testing.T) {
		ExpectedRevert(t, g.ExpectedFail(g.WithdrawRewardAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, claimedReward)), "Operator: only fundManager can access")
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		_, err := g.ExpectedOk(g.WithdrawRewardAmount(fundManager, operatorContractSingleOwner.From, claimedReward))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, claimedReward))
	})

	var claimedFee *big.Int

	t.Run("Get fee from undelegation", func(t *testing.T) {
		// 1. delegator1 delegates, distribute reward
		_, err := g.ExpectedOk(g.Delegate(t, delegator1, s1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		functionSignature := []byte("changeFeeRecipient(address)")
		methodID := crypto.Keccak256(functionSignature)[:4]
		args := abi.Arguments{
			{Type: mustParseType("address")},
		}
		packedArgs, err := args.Pack(operatorSampleAddr)
		if err != nil {
			panic(fmt.Sprintf("Failed to pack arguments: %v", err))
		}
		callData := append(methodID, packedArgs...)
		// 2. submit tx to set operatorContract as feeRecipient

		receipt, err := g.ExpectedOk(g.SubmitTransaction(operatorContractSingleOwner, TestGovStakingAddress, new(big.Int), callData))
		require.NoError(t, err)
		txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

		// 3.confirm the submitted tx
		_, err = g.ExpectedOk(g.ConfirmTransaction(operatorContractSingleOwner, txId))
		require.NoError(t, err)

		// 4. execute the tx
		_, err = g.ExpectedOk(g.ExecuteTransaction(operatorContractSingleOwner, txId))
		require.NoError(t, err)

		// 5. Check if FeeRecipient has changed
		require.Equal(t, operatorSampleAddr, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).FeeRecipient)

		beforeBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
		// 6. Delegator claims reward, fee will be sent to operatorContract
		receipt, err = g.ExpectedOk(g.Claim(t, delegator1, s1.Staker.Address, false))
		require.NoError(t, err)
		pendingFee := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingFee"].(*big.Int)
		afterBalance := g.balanceAt(t, ctx, operatorSampleAddr, nil)
		require.Equal(t, afterBalance, beforeBalance.Add(beforeBalance, pendingFee))

		// 7. Check if feeAmount tracking works
		callOpts := new(bind.CallOpts)
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&claimedFee}, "feeAmount"))
		require.Equal(t, pendingFee, claimedFee)
	})

	t.Run("Withdraw fee", func(t *testing.T) {
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		ExpectedRevert(t, g.ExpectedFail(g.WithdrawFeeAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, claimedFee)), "Operator: only fundManager can access")
		// owner withdraw fee
		_, err := g.ExpectedOk(g.WithdrawFeeAmount(fundManager, operatorContractSingleOwner.From, claimedFee))
		require.NoError(t, err)
		afterBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		// check if balance change is valid
		require.Equal(t, afterBalance, beforeBalance.Add(beforeBalance, claimedFee))

		// remaining fee amount in operator contract should be 0
		callOpts := new(bind.CallOpts)
		var feeAmount *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&feeAmount}, "feeAmount"))
		require.True(t, feeAmount.Cmp(new(big.Int)) == 0)
	})

	t.Run("Unstake and withdraw undstaked amount", func(t *testing.T) {
		// unstake the staked amount
		stakedAmt := govwbft.UserInfo(TestGovStakingAddress, stateDB, s1.Staker.Address, s1.Staker.Address).StakingAmount
		receipt, err := g.ExpectedOk(g.SingleOwnerUnstake(operatorContractSingleOwner, stakedAmt))
		require.NoError(t, err)
		unbondingPeriod := findEvents("NewCredential", receipt.Logs)[0]["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbondingPeriod.Int64()) * time.Second)

		_, err = g.ExpectedOk(g.WithdrawViaOperatorContract(operatorContractSingleOwner, new(big.Int)))
		require.NoError(t, err)

		callOpts := new(bind.CallOpts)
		var unstakedAmount *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&unstakedAmount}, "unstakedAmount"))
		require.Equal(t, unstakedAmount, stakedAmt)

		// withdraw it
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		receipt, err = g.ExpectedOk(g.WithdrawUnstakedAmount(operatorContractSingleOwner, operatorContractSingleOwner.From, unstakedAmount))
		require.NoError(t, err)
		gasUsed := calcTxGasCost(receipt)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, new(big.Int).Sub(unstakedAmount, gasUsed)))
	})

	t.Run("Withdraw remaining reward", func(t *testing.T) {
		distributeReward(t, g, stateDB, rewardAmount, s1.Staker.Address)
		receipt, err := g.ExpectedOk(g.ClaimWithoutRestake(operatorContractSingleOwner, s1))
		require.NoError(t, err)
		claimedReward := findEvents("UserRewardUpdated", receipt.Logs)[0]["pendingReward"].(*big.Int)

		callOpts := new(bind.CallOpts)
		var tracedReward *big.Int
		require.NoError(t, g.operatorContract.Call(callOpts, &[]interface{}{&tracedReward}, "rewardAmount"))
		require.Equal(t, tracedReward, claimedReward)

		// withdraw it
		beforeBalance := g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil)
		_, err = g.ExpectedOk(g.WithdrawRewardAmount(fundManager, operatorContractSingleOwner.From, claimedReward))
		require.NoError(t, err)
		require.Equal(t, g.balanceAt(t, ctx, operatorContractSingleOwner.From, nil), beforeBalance.Add(beforeBalance, claimedReward))
	})

	t.Run("Activate staker by staking again", func(t *testing.T) {
		_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), operatorContractSingleOwner, minStaking, &operatorSampleAddr))
		require.NoError(t, err)
		_, err := g.ExpectedOk(g.SingleOwnerStake(operatorContractSingleOwner, minStaking))
		require.NoError(t, err)

		var isStaker bool
		require.NoError(t, g.stakingContract.Call(new(bind.CallOpts), &[]interface{}{&isStaker}, "isStaker", s1.Staker.Address))
		require.True(t, isStaker)
	})
}

func TestMultiSig(t *testing.T) {
	var (
		owner1      = getTxOpt(t, "operatorContractOwner1")
		owner2      = getTxOpt(t, "operatorContractOwner2")
		owner3      = getTxOpt(t, "operatorContractOwner3")
		notOwner    = getTxOpt(t, "notOwner")
		fundManager = getTxOpt(t, "fundManger")
		err         error
		receipt     *types.Receipt
	)
	// initiate gov
	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		owner1.From:   {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner2.From:   {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner3.From:   {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		notOwner.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	// deploy operatorSample with single owner
	owners := []common.Address{owner1.From}
	fundManagers := []common.Address{fundManager.From}
	// quorum will be set to 1 anyway since there's only one owner
	operatorSampleAddr := g.DeployOperatorSample(t, owners, fundManagers, common.Big2)

	t.Run("Changing owners", func(t *testing.T) {
		// cannot remove single owner
		ExpectedRevert(t, g.ExpectedFail(g.RemoveOwner(owner1, owner1.From, true)), "Operator: cannot remove single owner")
		// add owner -> not single owner anymore, quorum is 2
		_, err := g.ExpectedOk(g.AddOwner(owner1, owner2.From, true))
		require.NoError(t, err)

		t.Run("Failure cases", func(t *testing.T) {
			ExpectedRevert(t, g.ExpectedFail(g.AddOwner(owner1, owner3.From, false)), "Operator: only wallet can access")
			ExpectedRevert(t, g.ExpectedFail(g.RemoveOwner(owner1, owner1.From, true)), "Operator: only wallet can access")
			ExpectedRevert(t, g.ExpectedFail(g.ReplaceOwner(owner1, owner1.From, owner2.From)), "Operator: only wallet can access")
			ExpectedRevert(t, g.ExpectedFail(g.ChangeQuorum(owner1, common.Big2)), "Operator: only wallet can access")
		})

		t.Run("Quorum size auto changed", func(t *testing.T) {
			callData, _ := g.operatorContract.Pack("removeOwner", owner2.From, false)
			receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
			require.NoError(t, err)
			txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)
			// confirm it
			g.ExpectedOk(g.ConfirmTransaction(owner1, txId))
			g.ExpectedOk(g.ConfirmTransaction(owner2, txId))

			receipt, err = g.ExpectedOk(g.ExecuteTransaction(owner1, txId))
			require.NoError(t, err)

			newQuorum := findEvents("ChangeQuorum", receipt.Logs)[0]["quorum"].(*big.Int)
			require.Equal(t, newQuorum, common.Big1)
		})

		// cannot remove non-existing owner
		callData, _ := g.operatorContract.Pack("removeOwner", owner2.From, true)
		receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
		require.NoError(t, err)
		txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)
		// single owner, quorum is 1
		receipt, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner1}, owner1, 1, txId)
		ExpectedRevert(t, err, "Operator: transaction failed: Operator: only owner can access")

		// test onlyOwner modifier
		ExpectedRevert(t, g.ExpectedFail(g.SubmitTransaction(notOwner, operatorSampleAddr, common.Big0, []byte{})), "Operator: only owner can access")
	})

	t.Run("Changing fundManagers", func(t *testing.T) {
		t.Run("Failure cases", func(t *testing.T) {
			ExpectedRevert(t, g.ExpectedFail(g.AddFundManager(owner2, owner3.From)), "Operator: only owner can access")
			ExpectedRevert(t, g.ExpectedFail(g.RemoveFundManager(owner2, owner1.From)), "Operator: only owner can access")
		})
		t.Run("Add fundManager", func(t *testing.T) {
			_, err := g.ExpectedOk(g.AddFundManager(owner1, owner3.From))
			require.NoError(t, err)

			_, err = g.ExpectedOk(g.AddOwner(owner1, owner2.From, false))
			require.NoError(t, err)

			callData, _ := g.operatorContract.Pack("addFundManager", owner2.From, true)
			receipt, err = g.ExpectedOk(g.SubmitTransaction(owner1, operatorSampleAddr, common.Big0, callData))
			require.NoError(t, err)
			txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)
			receipt, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner1}, owner1, 1, txId)
			require.NoError(t, err)

			var fundManagers []common.Address
			require.NoError(t, g.operatorContract.Call(new(bind.CallOpts), &[]interface{}{&fundManagers}, "getFundManagers"))
			require.True(t, len(fundManagers) == 2)
		})
	})
}

func TestCallingNCPContract(t *testing.T) {
	var (
		// owner1,2,3 is the owner of operatorContarct, and operatorContract will be ncp1
		owner1 = getTxOpt(t, "operatorContractOwner1")
		owner2 = getTxOpt(t, "operatorContractOwner2")
		owner3 = getTxOpt(t, "operatorContractOwner3")
		ncp2   = new(EOA)
		ncp3   = new(EOA)
		ncp4   = new(EOA)
		err    error
	)
	ncpList := []common.Address{owner1.From, ncp2.Address, ncp3.Address, ncp4.Address}
	// initiate gov
	g, err := NewGovWBFT(t, ncpList, types.GenesisAlloc{
		owner1.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner2.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
		owner3.From: {Balance: new(big.Int).Add(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	setWbftGovConfig(g)
	defer g.backend.Close()

	require.NoError(t, err)

	// deploy operatorSample with single owner
	owners := []common.Address{owner1.From, owner2.From, owner3.From}
	operatorSampleAddr := g.DeployOperatorSample(t, owners, []common.Address{}, common.Big2)

	t.Run("Set operator contract as ncp", func(t *testing.T) {
		_, err = g.ExpectedOk(g.ncpContract.Transact(owner1, "changeNCP", operatorSampleAddr))
		require.NoError(t, err)
		var isNCP bool
		require.NoError(t, g.ncpContract.Call(new(bind.CallOpts), &[]interface{}{&isNCP}, "isNCP", operatorSampleAddr))
		require.True(t, isNCP)
	})

	t.Run("operator contract call ncpContract methods", func(t *testing.T) {
		// call change ncp method - change ncp to owner1 address again
		callData, _ := g.ncpContract.Pack("changeNCP", owner1.From)
		receipt, err := g.ExpectedOk(g.SubmitTransaction(owner1, TestGovNCPAddress, common.Big0, callData))
		require.NoError(t, err)
		txId := findEvents("SubmitTransaction", receipt.Logs)[0]["txIndex"].(*big.Int)

		_, err = multiSigConfirmTxAndExecute(g, []*bind.TransactOpts{owner2, owner3}, owner3, 2, txId)
		require.NoError(t, err)

		var isNCP bool
		require.NoError(t, g.ncpContract.Call(new(bind.CallOpts), &[]interface{}{&isNCP}, "isNCP", owner1.From))
		require.True(t, isNCP)
	})
}

package test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/stretchr/testify/require"
)

type TestStateDB struct {
	getState func(addr common.Address, hash common.Hash) common.Hash
}

func (db *TestStateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	return db.getState(addr, hash)
}

func TestGovWithoutNCP(t *testing.T) {
	var (
		ctx          = context.TODO()
		minStaking   = towei(500000)
		totalStaking = new(big.Int)
		stakers      = make([]common.Address, 0)

		v1        = NewTestStaker()
		v2        = NewTestStaker()
		delegator = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		v2.Operator.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator.Address:   {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
	})
	require.NoError(t, err)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	checkGovBalanceFn := func() {
		require.True(t, totalStaking.Cmp(g.balanceAt(t, ctx, govwbft.GovStakingAddress, nil)) == 0)
	}

	t.Run("New Validtor", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("add staker", func(t *testing.T) {
			require.True(t, govwbft.TotalStaking(stateDB).Sign() == 0)
			require.True(t, len(govwbft.Stakers(stateDB)) == 0)
			beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking))
			require.NoError(t, err)
			stakers = append(stakers, v1.Staker.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
			require.Equal(t, stakers, govwbft.Stakers(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Operator.Address, nil))
		})

		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t,
					"registerStaker",
					v2.Operator, minStaking,
					new(big.Int).Sub(minStaking, big.NewInt(1)),
					v2.Staker.Address,
					v2.Rewardee.Address,
				)),
				"amount and msg.value mismatch",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, v2, new(big.Int).Sub(minStaking, big.NewInt(1)))),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, v2, new(big.Int).Add(MAX_UINT_128, big.NewInt(1)))),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v2.Operator, v2.Operator, v2.Operator}, minStaking)),
				"operator cannot be staker or rewardee",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{&EOA{Address: common.Address{}}, v2.Operator, v2.Staker}, minStaking)),
				"zero address",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v2.Staker, v2.Operator, v2.Staker}, minStaking)),
				"staker cannot be rewardee",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v2.Staker, v1.Operator, v2.Rewardee}, minStaking)),
				"operator is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v1.Operator, v2.Operator, v2.Rewardee}, minStaking)),
				"staker is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v2.Staker, v2.Operator, v1.Rewardee}, minStaking)),
				"rewardee is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker{v1.Staker, v2.Operator, v2.Rewardee}, minStaking)),
				"staker exists",
			)
		})

		t.Run("add another staker", func(t *testing.T) {
			require.Equal(t, minStaking, govwbft.TotalStaking(stateDB))
			require.Equal(t, stakers, govwbft.Stakers(stateDB))
			beforeBalance := g.balanceAt(t, ctx, v2.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterStaker(t, v2, minStaking))
			require.NoError(t, err)

			stakers = append(stakers, v2.Staker.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
			require.Equal(t, stakers, govwbft.Stakers(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v2.Operator.Address, nil))
		})
	})

	t.Run("Stake", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t, "stake", v1.Operator, common.Big2, minStaking)),
				"amount and msg.value mismatch",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, delegator, minStaking)),
				"unregistered staker",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, v1.Operator, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("stake more", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.Stake(t, v1.Operator, minStaking))
			require.NoError(t, err)

			totalStaking = totalStaking.Add(totalStaking, minStaking)
			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
			require.Equal(t, new(big.Int).Mul(minStaking, common.Big2), govwbft.StakerInfo(stateDB, v1.Staker.Address).Staking)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Operator.Address, nil))
		})
	})

	t.Run("Unstake & Withdraw", func(t *testing.T) {
		defer checkGovBalanceFn()
		var unstakeEvent map[string]interface{}

		t.Run("unstake failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, delegator, minStaking)),
				"unregistered staker",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Operator, common.Big0)),
				"amount is zero",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Operator, new(big.Int).Add(minStaking, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v2.Operator, new(big.Int).Sub(minStaking, common.Big1))),
				"amount must equal balance to remove staker",
			)
		})

		t.Run("unstake", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
			require.Equal(t, minStaking, govwbft.StakerInfo(stateDB, v1.Staker.Address).Staking)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Operator.Address, nil))

			unstakeEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, unstakeEvent)
		})

		t.Run("withdraw failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t, "withdraw", v1.Operator, nil, common.Big0)),
				"invalid credential",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, v2.Operator, unstakeEvent["credentialID"].(*big.Int))),
				"msg.sender is not requester",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, v1.Operator, unstakeEvent["credentialID"].(*big.Int))),
				"not yet time to withdraw",
			)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)

			unbonding := unstakeEvent["unbonding"].(*big.Int)
			g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, v1.Operator, unstakeEvent["credentialID"].(*big.Int)))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Operator.Address, nil))
		})

		t.Run("remove staker", func(t *testing.T) {
			{ // unstake
				receipt, err := g.ExpectedOk(g.Unstake(t, v2.Operator, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Sub(totalStaking, minStaking)
				stakers = removeElement(stakers, v2.Staker.Address)

				require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
				require.Equal(t, stakers, govwbft.Stakers(stateDB))
				require.True(t, govwbft.StakerInfo(stateDB, v2.Staker.Address).Staking.Sign() == 0)

				unstakeEvent = findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)
			}

			{ // withdraw
				beforeBalance := g.balanceAt(t, ctx, v2.Operator.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
				receipt, err := g.ExpectedOk(g.Withdraw(t, v2.Operator, unstakeEvent["credentialID"].(*big.Int)))
				require.NoError(t, err)

				gasCost := calcTxGasCost(receipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v2.Operator.Address, nil))
			}
		})
	})

	t.Run("Delegate", func(t *testing.T) {
		defer checkGovBalanceFn()
		delegateAmount := towei(100_000)
		t.Run("failure case", func(t *testing.T) {
			{
				_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &v1.Staker.Address))
				require.NoError(t, err)
				_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &v1.Rewardee.Address))
				require.NoError(t, err)
			}
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Staker, v1.Staker.Address, delegateAmount)),
				"staker cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Operator, v1.Staker.Address, delegateAmount)),
				"operator(rewardee) cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, v1.Rewardee, v1.Staker.Address, delegateAmount)),
				"operator(rewardee) cannot delegate",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, v2.Staker.Address, delegateAmount)),
				"unregistered staker",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, v1.Staker.Address, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("delegate", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)
			beforeInfo_v1 := govwbft.StakerInfo(stateDB, v1.Staker.Address)

			receipt, err := g.ExpectedOk(g.Delegate(t, delegator, v1.Staker.Address, delegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Add(totalStaking, delegateAmount)
			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))

			afterInfo_v1 := govwbft.StakerInfo(stateDB, v1.Staker.Address)
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_v1.Staking, beforeInfo_v1.Staking))
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_v1.Delegated, beforeInfo_v1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(delegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})

	t.Run("Undelegate & Withdraw", func(t *testing.T) {
		defer checkGovBalanceFn()
		var (
			undelegateEvent  map[string]interface{}
			undelegateAmount = towei(50_000)
		)

		t.Run("undelegate", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)
			beforeInfo_v1 := govwbft.StakerInfo(stateDB, v1.Staker.Address)

			receipt, err := g.ExpectedOk(g.Unelegate(t, delegator, v1.Staker.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, undelegateAmount)
			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))

			afterInfo_v1 := govwbft.StakerInfo(stateDB, v1.Staker.Address)
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_v1.Staking, afterInfo_v1.Staking))
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_v1.Delegated, afterInfo_v1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))

			undelegateEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, undelegateEvent)
		})

		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, delegator, undelegateEvent["credentialID"].(*big.Int))),
				"not yet time to withdraw",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unelegate(t, delegator, v1.Staker.Address, new(big.Int).Add(undelegateAmount, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unelegate(t, delegator, v2.Staker.Address, undelegateAmount)),
				"insufficient balance",
			)

			// try unstake, including the delegated amount
			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, v1.Operator, govwbft.StakerInfo(stateDB, v1.Staker.Address).Staking)),
				"insufficient balance",
			)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			unbonding := undelegateEvent["unbonding"].(*big.Int)
			g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, delegator, undelegateEvent["credentialID"].(*big.Int)))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})

		t.Run("undelegate to removed staker", func(t *testing.T) {
			// unstake and remove staker
			{
				receipt, err := g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Sub(totalStaking, minStaking)
				stakers = removeElement(stakers, v1.Staker.Address)

				require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
				require.Equal(t, stakers, govwbft.Stakers(stateDB))

				unstakeEvent := findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)

				beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
				withdrawReceipt, err := g.ExpectedOk(g.Withdraw(t, v1.Operator, unstakeEvent["credentialID"].(*big.Int)))
				require.NoError(t, err)

				gasCost := calcTxGasCost(withdrawReceipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, v1.Operator.Address, nil))
			}
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			receipt, err := g.ExpectedOk(g.Unelegate(t, delegator, v1.Staker.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking = totalStaking.Sub(totalStaking, undelegateAmount)
			require.True(t, totalStaking.Sign() == 0)
			require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})
}

var (
	ProposalType_None       = common.Big0
	ProposalType_NCPAdd     = common.Big1
	ProposalType_NCPRemoval = common.Big2

	Voting_Period = time.Duration(604800) * time.Second
)

func TestGovWithNCP(t *testing.T) {
	var (
		ctx             = context.TODO()
		minStaking      = towei(500000)
		totalStaking    = new(big.Int)
		ncpTotalStaking = new(big.Int)
		stakers         = make([]common.Address, 0)
		ncps            = make([]common.Address, 0)
		ncpStakers      = make([]common.Address, 0)

		ncp1 = NewTestStaker()
		ncp2 = NewTestStaker()
		ncp3 = NewTestStaker()
		ncp4 = NewTestStaker()
	)

	ncps = append(ncps, ncp1.Operator.Address, ncp2.Operator.Address)
	g, err := NewGovWBFT(t, ncps, types.GenesisAlloc{
		ncp1.Operator.Address: {Balance: MAX_UINT_128},
		ncp2.Operator.Address: {Balance: MAX_UINT_128},
		ncp3.Operator.Address: {Balance: MAX_UINT_128},
		ncp4.Operator.Address: {Balance: MAX_UINT_128},
	})
	require.NoError(t, err)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	checkNCPStaker := func() {
		require.True(t, totalStaking.Cmp(govwbft.TotalStaking(stateDB)) == 0)
		require.Equal(t, stakers, govwbft.Stakers(stateDB))
		require.Equal(t, ncps, govwbft.NCPList(stateDB))
		require.Equal(t, ncpTotalStaking, govwbft.NCPTotalStaking(stateDB))
		require.Equal(t, ncpStakers, govwbft.NCPStakers(stateDB))
	}

	t.Run("NCP Staking", func(t *testing.T) {
		require.True(t, govwbft.TotalStaking(stateDB).Sign() == 0)
		require.True(t, govwbft.NCPTotalStaking(stateDB).Sign() == 0)
		require.Equal(t, stakers, govwbft.Stakers(stateDB))
		require.Equal(t, ncps, govwbft.NCPList(stateDB))
		require.Equal(t, ncpStakers, govwbft.NCPStakers(stateDB))

		t.Run("NCP staking", func(t *testing.T) {
			defer checkNCPStaker()
			_, err := g.ExpectedOk(g.RegisterStaker(t, ncp1, minStaking))
			require.NoError(t, err)

			stakers = append(stakers, ncp1.Staker.Address)
			ncpStakers = append(ncpStakers, ncp1.Staker.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)
			ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, minStaking)
		})

		t.Run("non-NCP staking", func(t *testing.T) {
			defer checkNCPStaker()
			_, err := g.ExpectedOk(g.RegisterStaker(t, ncp3, minStaking))
			require.NoError(t, err)

			stakers = append(stakers, ncp3.Staker.Address)
			totalStaking = totalStaking.Add(totalStaking, minStaking)
		})

		t.Run("stake more", func(t *testing.T) {
			defer checkNCPStaker()
			// ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp1.Operator, minStaking))
				require.NoError(t, err)

				totalStaking = totalStaking.Add(totalStaking, minStaking)
				ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, minStaking)
			}

			// non-ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp3.Operator, minStaking))
				require.NoError(t, err)
				totalStaking = totalStaking.Add(totalStaking, minStaking)
			}
		})
	})

	t.Run("Add NCP", func(t *testing.T) {
		var proposalEvent map[string]interface{}

		t.Run("new proposal to add ncp", func(t *testing.T) {
			defer checkNCPStaker()

			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)

			proposalEvent = findEvent("NewProposal", receipt.Logs)
			require.Equal(t, ProposalType_NCPAdd, proposalEvent["proposalType"].(*big.Int))
		})

		t.Run("failure case", func(t *testing.T) {
			defer checkNCPStaker()

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp4.Operator, ncp4.Operator.Address)),
				"msg.sender is not ncp",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Operator, ncp2.Operator.Address)),
				"ncp exists",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Operator, ncp4.Operator.Address)),
				"previous vote is in progress",
			)
		})
		t.Run("vote & add ncp", func(t *testing.T) {
			defer checkNCPStaker()

			_, err := g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			// Vote not finalized
			checkNCPStaker()

			_, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ncps = append(ncps, ncp3.Operator.Address)
			ncpStakers = append(ncpStakers, ncp3.Staker.Address)
			ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, govwbft.StakerInfo(stateDB, ncp3.Staker.Address).Staking)
		})
	})

	t.Run("Remove NCP", func(t *testing.T) {
		var proposalEvent map[string]interface{}

		t.Run("new proposal to remove ncp", func(t *testing.T) {
			defer checkNCPStaker()

			receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)

			proposalEvent = findEvent("NewProposal", receipt.Logs)
			require.Equal(t, ProposalType_NCPRemoval, proposalEvent["proposalType"].(*big.Int))
		})

		t.Run("failure case", func(t *testing.T) {
			defer checkNCPStaker()

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToRemoveNCP(t, ncp1.Operator, ncp4.Operator.Address)),
				"invalid ncp",
			)
		})
		t.Run("vote & remove ncp", func(t *testing.T) {
			defer checkNCPStaker()

			_, err := g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			// not finalized
			checkNCPStaker()

			_, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ncps = removeElement(ncps, ncp3.Operator.Address)
			ncpStakers = removeElement(ncpStakers, ncp3.Staker.Address)
			ncpTotalStaking = ncpTotalStaking.Sub(ncpTotalStaking, govwbft.StakerInfo(stateDB, ncp3.Staker.Address).Staking)
		})
	})

	t.Run("Cancel Proposal", func(t *testing.T) {
		defer checkNCPStaker()

		t.Run("cancel by proposer", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.CancelProposal(t, ncp2.Operator, proposalEvent["id"].(*big.Int))),
				"non-proposer cannot cancel before timeout",
			)

			receipt, err = g.ExpectedOk(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int)))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])

			t.Run("cannot cancel after vote", func(t *testing.T) {
				receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
				require.NoError(t, err)
				proposalEvent := findEvent("NewProposal", receipt.Logs)

				_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
				require.NoError(t, err)

				ExpectedRevert(t,
					g.ExpectedFail(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int))),
					"cannot cancel after vote",
				)

				g.backend.AdjustTime(Voting_Period)

				// cancel for next test
				{
					receipt, err = g.ExpectedOk(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int)))
					require.NoError(t, err)
					require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
				}
			})
		})

		t.Run("canceled due to timeout", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.CancelProposal(t, ncp2.Operator, proposalEvent["id"].(*big.Int))),
				"non-proposer cannot cancel before timeout",
			)

			g.adjustTime(Voting_Period)

			receipt, err = g.ExpectedOk(g.CancelProposal(t, ncp2.Operator, proposalEvent["id"].(*big.Int)))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
		})

		t.Run("timeout & new proposal", func(t *testing.T) {
			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			ExpectedRevert(t,
				g.ExpectedFail(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address)),
				"previous vote is in progress",
			)

			g.adjustTime(Voting_Period)

			receipt, err = g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)
			require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])

			// cancel for next test
			{
				proposalEvent := findEvent("NewProposal", receipt.Logs)
				receipt, err := g.ExpectedOk(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int)))
				require.NoError(t, err)
				require.Equal(t, proposalEvent["id"], findEvent("ProposalCanceled", receipt.Logs)["proposalID"])
			}
		})
	})

	t.Run("Vote", func(t *testing.T) {
		t.Run("failure case", func(t *testing.T) {
			defer checkNCPStaker()

			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
			require.NoError(t, err)
			proposalEvent := findEvent("NewProposal", receipt.Logs)

			_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ExpectedRevert(t,
				g.ExpectedFail(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true)),
				"already voted",
			)

			g.adjustTime(Voting_Period)

			ExpectedRevert(t,
				g.ExpectedFail(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true)),
				"already closed vote",
			)
		})

		t.Run("majority", func(t *testing.T) {
			t.Run("2 ncp", func(t *testing.T) {
				t.Run("reject", func(t *testing.T) {
					defer checkNCPStaker()

					// 1 NCP is required for reject
					receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, false, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
				t.Run("accept", func(t *testing.T) {
					defer checkNCPStaker()

					// 2 NCP is required for accept
					receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp3.Operator.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					ncps = append(ncps, ncp3.Operator.Address)
					ncpStakers = append(ncpStakers, ncp3.Staker.Address)
					ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, govwbft.StakerInfo(stateDB, ncp3.Staker.Address).Staking)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, true, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
			})
			t.Run("3 ncp", func(t *testing.T) {
				t.Run("reject", func(t *testing.T) {
					defer checkNCPStaker()

					// 2 NCP is required for reject
					receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp2.Operator, ncp3.Operator.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), false))
					require.NoError(t, err)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, false, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
				t.Run("accept", func(t *testing.T) {
					defer checkNCPStaker()

					// 2 NCP is required for accept
					receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp3.Operator, ncp3.Operator.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp3.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					ncps = removeElement(ncps, ncp3.Operator.Address)
					ncpStakers = removeElement(ncpStakers, ncp3.Staker.Address)
					ncpTotalStaking = ncpTotalStaking.Sub(ncpTotalStaking, govwbft.StakerInfo(stateDB, ncp3.Staker.Address).Staking)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, true, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
			})
		})
	})
}

func TestSetCode(t *testing.T) {
	/*
		contract TestGovConst{
		    uint256 public constant MINIMUM_STAKING = 100000e18;
		    uint256 public constant MAXIMUM_STAKING = type(uint128).max;
		    uint256 public constant UNBONDING_PERIOD_STAKER = 3 hours;
		    uint256 public constant UNBONDING_PERIOD_DELEGATOR = 72 hours;
		}
	*/
	var testGovConst = "0x6080604052348015600f57600080fd5b506004361060465760003560e01c8063129060ab14604b578063840c1771146073578063ba631d3f14607c578063fde7f37114608c575b600080fd5b60616fffffffffffffffffffffffffffffffff81565b60405190815260200160405180910390f35b60616203f48081565b606169152d02c7e14af680000081565b6061612a308156fea264697066735822122002e0d8fab314d33e504da31ee2aa938cea7fe0526a24782b3a3ea3b107464a2b64736f6c634300080e0033"
	var (
		ctx          = context.TODO()
		minStaking1  = towei(500000)
		minStaking2  = towei(100000)
		totalStaking = new(big.Int)
		stakers      = make([]common.Address, 0)

		ncp1 = NewTestStaker()
		ncp2 = NewTestStaker()
		ncp3 = NewTestStaker()
	)

	// for duplicate test
	ncpInput := []common.Address{ncp3.Operator.Address, ncp3.Operator.Address}

	g, err := NewGovWBFT(t, ncpInput, types.GenesisAlloc{
		ncp1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		ncp2.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
	})
	require.NoError(t, err)
	defer g.backend.Close()

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	t.Run("duplicated ncp", func(t *testing.T) {
		ncps := []common.Address{ncp3.Operator.Address}
		require.Equal(t, ncps, govwbft.NCPList(stateDB))
	})

	t.Run("add staker", func(t *testing.T) {
		require.True(t, govwbft.TotalStaking(stateDB).Sign() == 0)
		require.True(t, len(govwbft.Stakers(stateDB)) == 0)
		beforeBalance := g.balanceAt(t, ctx, ncp1.Operator.Address, nil)

		receipt, err := g.ExpectedOk(g.RegisterStaker(t, ncp1, minStaking1))
		require.NoError(t, err)
		stakers = append(stakers, ncp1.Staker.Address)
		totalStaking = totalStaking.Add(totalStaking, minStaking1)

		require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
		require.Equal(t, stakers, govwbft.Stakers(stateDB))

		gasCost := calcTxGasCost(receipt)
		expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking1, gasCost))
		require.Equal(t, expectedBalance, g.balanceAt(t, ctx, ncp1.Operator.Address, nil))

		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, minStaking2)),
			"out of bounds",
		)
	})

	t.Run("upgrade contract", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, new(big.Int).Sub(minStaking1, common.Big1))),
			"out of bounds",
		)

		// upgrade contract
		g.backend.CommitWithState(params.StateTransition{
			Codes: []params.CodeParam{{Address: govwbft.GovConstAddress, Code: testGovConst}},
		})
	})

	t.Run("retry add staker", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, new(big.Int).Sub(minStaking2, common.Big1))),
			"out of bounds",
		)

		beforeBalance := g.balanceAt(t, ctx, ncp2.Operator.Address, nil)

		receipt, err := g.ExpectedOk(g.RegisterStaker(t, ncp2, minStaking2))
		require.NoError(t, err)
		stakers = append(stakers, ncp2.Staker.Address)
		totalStaking = totalStaking.Add(totalStaking, minStaking2)

		require.Equal(t, totalStaking, govwbft.TotalStaking(stateDB))
		require.Equal(t, stakers, govwbft.Stakers(stateDB))

		gasCost := calcTxGasCost(receipt)
		expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking2, gasCost))
		require.Equal(t, expectedBalance, g.balanceAt(t, ctx, ncp2.Operator.Address, nil))
	})
}

func removeElement(slice []common.Address, value common.Address) []common.Address {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

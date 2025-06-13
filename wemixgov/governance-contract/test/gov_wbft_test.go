package test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/bls/blst"
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
		feeRate      = new(big.Int).SetUint64(1500)
		minStaking   = towei(500000)
		totalStaking = new(big.Int)
		stakers      = make([]common.Address, 0)

		s1        = NewTestStaker()
		s2        = NewTestStaker()
		delegator = NewEOA()
		newStaker = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		s1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		s2.Operator.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator.Address:   {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		newStaker.Address:   {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	checkGovBalanceFn := func() {
		require.True(t, totalStaking.Cmp(g.balanceAt(t, ctx, TestGovStakingAddress, nil)) == 0)
	}

	t.Run("New Staker", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("add staker", func(t *testing.T) {
			require.True(t, govwbft.TotalStaking(TestGovStakingAddress, stateDB).Sign() == 0)
			require.True(t, len(govwbft.Stakers(TestGovStakingAddress, stateDB)) == 0)
			beforeBalance := g.balanceAt(t, ctx, s1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterStaker(t, s1, minStaking, feeRate))
			require.NoError(t, err)
			stakers = append(stakers, s1.Staker.Address)
			totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s1.Operator.Address, nil))
		})

		t.Run("failure case", func(t *testing.T) {
			s2_bls_pk, err := s2.GetBLSPublicKey()
			require.NoError(t, err)
			s2_bls_pk_byte := s2_bls_pk.Marshal()
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t,
					"registerStaker",
					s2.Operator, minStaking,
					new(big.Int).Sub(minStaking, big.NewInt(1)),
					s2.Staker.Address,
					s2.FeeRecipient.Address,
					feeRate,
					s2_bls_pk_byte,
				)),
				"amount and msg.value mismatch",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t,
					"registerStaker",
					s2.Operator, minStaking,
					minStaking,
					s2.Staker.Address,
					s2.FeeRecipient.Address,
					feeRate,
					s2_bls_pk_byte[1:],
				)),
				"invalid bls public key",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, s2, new(big.Int).Sub(minStaking, big.NewInt(1)), feeRate)),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, s2, new(big.Int).Add(MAX_UINT_128, big.NewInt(1)), feeRate)),
				"out of bounds",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{&EOA{Address: common.Address{}}, s2.Operator, s2.FeeRecipient}, minStaking, feeRate)),
				"received ikm is invalid",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{s2.Staker, s2.Staker, s2.FeeRecipient}, minStaking, feeRate)),
				"insufficient funds for transfer",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{s2.Staker, s1.Operator, s2.FeeRecipient}, minStaking, feeRate)),
				"operator is already registered",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{s1.Staker, s2.Operator, s2.FeeRecipient}, minStaking, feeRate)),
				"already registered staker",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{s2.Staker, s2.Operator, &EOA{Address: common.Address{}}}, minStaking, feeRate)),
				"fee recipient is zero address",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.RegisterStaker(t, &TestStaker[*EOA]{s2.Staker, s2.Operator, s2.FeeRecipient}, minStaking, new(big.Int).SetUint64(10001))),
				"fee rate exceeds precision",
			)
		})

		t.Run("add another staker", func(t *testing.T) {
			require.Equal(t, minStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
			beforeBalance := g.balanceAt(t, ctx, s2.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.RegisterStaker(t, s2, minStaking, feeRate))
			require.NoError(t, err)

			stakers = append(stakers, s2.Staker.Address)
			totalStaking.Add(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s2.Operator.Address, nil))
		})
	})

	t.Run("Stake", func(t *testing.T) {
		defer checkGovBalanceFn()
		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.stakingContractTx(t, "stake", s1.Operator, common.Big2, minStaking)),
				"amount and msg.value mismatch",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, delegator, minStaking)),
				"unregistered staker",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Stake(t, s1.Operator, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("stake more", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, s1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.Stake(t, s1.Operator, minStaking))
			require.NoError(t, err)

			totalStaking.Add(totalStaking, minStaking)
			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, new(big.Int).Mul(minStaking, common.Big2), govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s1.Operator.Address, nil))
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
				g.ExpectedFail(g.Unstake(t, s2.Operator, common.Big0)),
				"amount is zero",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, s2.Operator, new(big.Int).Add(minStaking, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, s2.Operator, new(big.Int).Sub(minStaking, common.Big1))),
				"amount must equal balance to deactivate staker",
			)
		})

		t.Run("unstake", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, s1.Operator.Address, nil)

			receipt, err := g.ExpectedOk(g.Unstake(t, s1.Operator, minStaking))
			require.NoError(t, err)

			totalStaking.Sub(totalStaking, minStaking)

			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, minStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s1.Operator.Address, nil))

			unstakeEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, unstakeEvent)
		})

		t.Run("withdraw failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, s2.Operator, common.Big0)),
				"no credential to withdraw",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, s1.Operator, common.Big1)),
				"withdrawal time not reached",
			)

			// no error, no withdrawal
			beforeBalance := g.balanceAt(t, ctx, TestGovStakingAddress, nil)
			_, err := g.ExpectedOk(g.Withdraw(t, s1.Operator, common.Big0))
			require.NoError(t, err)
			afterBalance := g.balanceAt(t, ctx, TestGovStakingAddress, nil)
			require.Equal(t, beforeBalance, afterBalance)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, s1.Operator.Address, nil)

			unbonding := unstakeEvent["unbonding"].(*big.Int)
			g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, s1.Operator, common.Big0))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s1.Operator.Address, nil))
		})

		t.Run("disable staker", func(t *testing.T) {
			{ // unstake
				receipt, err := g.ExpectedOk(g.Unstake(t, s2.Operator, minStaking))
				require.NoError(t, err)

				totalStaking.Sub(totalStaking, minStaking)
				stakers = removeElement(stakers, s2.Staker.Address)

				require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
				require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
				require.True(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s2.Staker.Address).TotalStaked.Sign() == 0)

				unstakeEvent = findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)
			}

			{ // withdraw
				beforeBalance := g.balanceAt(t, ctx, s2.Operator.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
				receipt, err := g.ExpectedOk(g.Withdraw(t, s2.Operator, common.Big0))
				require.NoError(t, err)

				gasCost := calcTxGasCost(receipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s2.Operator.Address, nil))
			}
		})
	})

	t.Run("Delegate", func(t *testing.T) {
		defer checkGovBalanceFn()
		delegateAmount := towei(100_000)
		t.Run("failure case", func(t *testing.T) {
			{
				_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &s1.Staker.Address))
				require.NoError(t, err)
				rewardee := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).Rewardee
				_, err = g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, minStaking, &rewardee))
				require.NoError(t, err)
			}
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, s1.Staker, s1.Staker.Address, delegateAmount)),
				"staker cannot delegate to self",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, s1.Operator, s1.Staker.Address, delegateAmount)),
				"operator cannot delegate to self",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, s2.Staker.Address, delegateAmount)),
				"staker is inactive",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, s2.Operator.Address, delegateAmount)),
				"unregistered staker",
			)
			ExpectedRevert(t,
				g.ExpectedFail(g.Delegate(t, delegator, s1.Staker.Address, MAX_UINT_128)),
				"exceeded the maximum",
			)
		})

		t.Run("delegate", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)
			beforeInfo_s1 := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address)

			receipt, err := g.ExpectedOk(g.Delegate(t, delegator, s1.Staker.Address, delegateAmount))
			require.NoError(t, err)

			totalStaking.Add(totalStaking, delegateAmount)
			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))

			afterInfo_s1 := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address)
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_s1.TotalStaked, beforeInfo_s1.TotalStaked))
			require.Equal(t, delegateAmount, new(big.Int).Sub(afterInfo_s1.Delegated, beforeInfo_s1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(delegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})

	t.Run("Deactivate & Reactivate Staker", func(t *testing.T) {
		delegatedAmount := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).Delegated
		receipt, err := g.ExpectedOk(g.Unstake(t, s1.Operator, minStaking))
		require.NoError(t, err)
		stakers = removeElement(stakers, s1.Staker.Address)

		require.Equal(t, delegatedAmount, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, []common.Address{}, stakers)
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

		unstakeEvent := findEvent("NewCredential", receipt.Logs)
		require.NotNil(t, unstakeEvent)
		unbonding := unstakeEvent["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)

		_, err = g.ExpectedOk(g.Withdraw(t, s1.Operator, common.Big0))
		require.NoError(t, err)

		t.Run("Failed to reactivate", func(t *testing.T) {
			ExpectedRevert(t, g.ExpectedFail(g.RegisterStaker(t, s1, minStaking, feeRate)), "already registered staker")
		})

		t.Run("Reactivate Staker", func(t *testing.T) {
			_, err := g.ExpectedOk(g.Stake(t, s1.Operator, minStaking))
			require.NoError(t, err)

			stakers = append(stakers, s1.Staker.Address)
			require.Equal(t, totalStaking, new(big.Int).Add(minStaking, delegatedAmount))
			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
			require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
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
			beforeInfo_s1 := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address)

			receipt, err := g.ExpectedOk(g.Undelegate(t, delegator, s1.Staker.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking.Sub(totalStaking, undelegateAmount)
			require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))

			afterInfo_s1 := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address)
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_s1.TotalStaked, afterInfo_s1.TotalStaked))
			require.Equal(t, undelegateAmount, new(big.Int).Sub(beforeInfo_s1.Delegated, afterInfo_s1.Delegated))

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Sub(beforeBalance, gasCost)
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))

			undelegateEvent = findEvent("NewCredential", receipt.Logs)
			require.NotNil(t, undelegateEvent)
		})

		t.Run("failure case", func(t *testing.T) {
			ExpectedRevert(t,
				g.ExpectedFail(g.Withdraw(t, delegator, common.Big1)),
				"withdrawal time not reached",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Undelegate(t, delegator, s1.Staker.Address, new(big.Int).Add(undelegateAmount, common.Big1))),
				"insufficient balance",
			)

			ExpectedRevert(t,
				g.ExpectedFail(g.Undelegate(t, delegator, s2.Staker.Address, undelegateAmount)),
				"insufficient balance",
			)

			// try unstake, including the delegated amount
			ExpectedRevert(t,
				g.ExpectedFail(g.Unstake(t, s1.Operator, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).TotalStaked)),
				"insufficient balance",
			)
		})

		t.Run("withdraw", func(t *testing.T) {
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			unbonding := undelegateEvent["unbonding"].(*big.Int)
			g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
			receipt, err := g.ExpectedOk(g.Withdraw(t, delegator, common.Big0))
			require.NoError(t, err)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateEvent["amount"].(*big.Int), gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})

		t.Run("undelegate to removed staker", func(t *testing.T) {
			// unstake and remove staker
			{
				receipt, err := g.ExpectedOk(g.Unstake(t, s1.Operator, minStaking))
				require.NoError(t, err)

				totalStaking.Sub(totalStaking, minStaking)

				delegated := govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).Delegated
				require.Equal(t, delegated, govwbft.DanglingDelegated(TestGovStakingAddress, stateDB))

				stakers = removeElement(stakers, s1.Staker.Address)

				require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

				unstakeEvent := findEvent("NewCredential", receipt.Logs)
				require.NotNil(t, unstakeEvent)

				beforeBalance := g.balanceAt(t, ctx, s1.Operator.Address, nil)

				unbonding := unstakeEvent["unbonding"].(*big.Int)
				g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
				withdrawReceipt, err := g.ExpectedOk(g.Withdraw(t, s1.Operator, common.Big1))
				require.NoError(t, err)

				gasCost := calcTxGasCost(withdrawReceipt)
				expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(unstakeEvent["amount"].(*big.Int), gasCost))
				require.Equal(t, expectedBalance, g.balanceAt(t, ctx, s1.Operator.Address, nil))
			}
			beforeBalance := g.balanceAt(t, ctx, delegator.Address, nil)

			receipt, err := g.ExpectedOk(g.Undelegate(t, delegator, s1.Staker.Address, undelegateAmount))
			require.NoError(t, err)

			totalStaking.Sub(totalStaking, undelegateAmount)

			gasCost := calcTxGasCost(receipt)
			expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateAmount, gasCost))
			require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator.Address, nil))
		})
	})

	t.Run("Transfer operator ship", func(t *testing.T) {
		_, err := g.ExpectedOk(g.TransferOperatorShip(t, s1.Operator, newStaker.Address))
		require.NoError(t, err)

		require.Equal(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).Operator, newStaker.Address)

		ExpectedRevert(t,
			g.ExpectedFail(g.TransferOperatorShip(t, s1.Operator, newStaker.Address)),
			"unregistered staker",
		)

		_, err = g.ExpectedOk(g.TransferOperatorShip(t, newStaker, s1.Operator.Address))
		require.NoError(t, err)

		require.Equal(t, govwbft.StakerInfo(TestGovStakingAddress, stateDB, s1.Staker.Address).Operator, s1.Operator.Address)
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
		feeRate         = new(big.Int).SetUint64(1500)
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
	setWbftGovConfigWithGovCouncil(g)

	stateDB := &TestStateDB{
		getState: func(addr common.Address, hash common.Hash) (result common.Hash) {
			value, _ := g.backend.Client().StorageAt(ctx, addr, hash, nil)
			return common.BytesToHash(value)
		},
	}

	checkNCPStaker := func() {
		require.True(t, totalStaking.Cmp(govwbft.TotalStaking(TestGovStakingAddress, stateDB)) == 0)
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, ncps, govwbft.NCPList(TestGovNCPAddress, stateDB))
		require.Equal(t, ncpTotalStaking, govwbft.NCPTotalStaking(TestGovStakingAddress, TestGovNCPAddress, stateDB))
		require.Equal(t, ncpStakers, govwbft.NCPStakers(TestGovStakingAddress, TestGovNCPAddress, stateDB))
	}

	t.Run("NCP Staking", func(t *testing.T) {
		require.True(t, govwbft.TotalStaking(TestGovStakingAddress, stateDB).Sign() == 0)
		require.True(t, govwbft.NCPTotalStaking(TestGovStakingAddress, TestGovNCPAddress, stateDB).Sign() == 0)
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, ncps, govwbft.NCPList(TestGovNCPAddress, stateDB))
		require.Equal(t, ncpStakers, govwbft.NCPStakers(TestGovStakingAddress, TestGovNCPAddress, stateDB))

		t.Run("NCP staking", func(t *testing.T) {
			defer checkNCPStaker()
			_, err := g.ExpectedOk(g.RegisterStaker(t, ncp1, minStaking, feeRate))
			require.NoError(t, err)

			stakers = append(stakers, ncp1.Staker.Address)
			ncpStakers = append(ncpStakers, ncp1.Staker.Address)
			totalStaking.Add(totalStaking, minStaking)
			ncpTotalStaking.Add(ncpTotalStaking, minStaking)
		})

		t.Run("non-NCP staking", func(t *testing.T) {
			defer checkNCPStaker()
			_, err := g.ExpectedOk(g.RegisterStaker(t, ncp3, minStaking, feeRate))
			require.NoError(t, err)

			stakers = append(stakers, ncp3.Staker.Address)
			totalStaking.Add(totalStaking, minStaking)
		})

		t.Run("stake more", func(t *testing.T) {
			defer checkNCPStaker()
			// ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp1.Operator, minStaking))
				require.NoError(t, err)

				totalStaking.Add(totalStaking, minStaking)
				ncpTotalStaking.Add(ncpTotalStaking, minStaking)
			}

			// non-ncp stake more
			{
				_, err := g.ExpectedOk(g.Stake(t, ncp3.Operator, minStaking))
				require.NoError(t, err)
				totalStaking.Add(totalStaking, minStaking)
			}
		})
	})

	t.Run("Remove self and add again", func(t *testing.T) {
		t.Run("remove ncp2", func(t *testing.T) {
			defer checkNCPStaker()

			_, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp2.Operator, ncp2.Operator.Address))
			require.NoError(t, err)

			ncps = removeElement(ncps, ncp2.Operator.Address)
		})

		t.Run("add ncp2 again", func(t *testing.T) {
			defer checkNCPStaker()

			receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp2.Operator.Address))
			require.NoError(t, err)

			proposalEvent := findEvent("NewProposal", receipt.Logs)
			require.Equal(t, ProposalType_NCPAdd, proposalEvent["proposalType"].(*big.Int))

			_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
			require.NoError(t, err)

			ncps = append(ncps, ncp2.Operator.Address)
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
			ncpTotalStaking.Add(ncpTotalStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, ncp3.Staker.Address).TotalStaked)
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
			ncpTotalStaking.Sub(ncpTotalStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, ncp3.Staker.Address).TotalStaked)
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
					ncpTotalStaking.Add(ncpTotalStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, ncp3.Staker.Address).TotalStaked)

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
					receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp2.Operator, ncp3.Operator.Address))
					require.NoError(t, err)
					proposalEvent := findEvent("NewProposal", receipt.Logs)

					_, err = g.ExpectedOk(g.Vote(t, ncp2.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					receipt, err = g.ExpectedOk(g.Vote(t, ncp3.Operator, proposalEvent["id"].(*big.Int), true))
					require.NoError(t, err)

					ncps = removeElement(ncps, ncp3.Operator.Address)
					ncpStakers = removeElement(ncpStakers, ncp3.Staker.Address)
					ncpTotalStaking.Sub(ncpTotalStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, ncp3.Staker.Address).TotalStaked)

					finalizedEvent := findEvent("ProposalFinalized", receipt.Logs)
					require.NotNil(t, finalizedEvent)
					require.Equal(t, true, findEvent("ProposalFinalized", receipt.Logs)["accepted"].(bool))
				})
			})
		})
	})

	t.Run("Change NCP", func(t *testing.T) {
		defer checkNCPStaker()

		_, err := g.ExpectedOk(g.ChangeNCP(t, ncp2.Operator, ncp4.Operator.Address))
		require.NoError(t, err)

		ncps = removeElement(ncps, ncp2.Operator.Address)
		ncps = append(ncps, ncp4.Operator.Address)
	})

	t.Run("Can change NCP even if there is an ongoing proposal", func(t *testing.T) {
		defer checkNCPStaker()

		receipt, err := g.ExpectedOk(g.NewProposalToRemoveNCP(t, ncp1.Operator, ncp4.Operator.Address))
		require.NoError(t, err)
		proposalEvent := findEvent("NewProposal", receipt.Logs)

		// must success even if there is an ongoing proposal
		g.ExpectedOk(g.ChangeNCP(t, ncp1.Operator, ncp2.Operator.Address))
		// back to ncp1
		g.ExpectedOk(g.ChangeNCP(t, ncp2.Operator, ncp1.Operator.Address))

		// must success even if there is an ongoing proposal
		g.ExpectedOk(g.ChangeNCP(t, ncp4.Operator, ncp2.Operator.Address))
		// back to ncp4
		g.ExpectedOk(g.ChangeNCP(t, ncp2.Operator, ncp4.Operator.Address))

		_, err = g.ExpectedOk(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int)))
		require.NoError(t, err)
	})

	t.Run("Cannot change the ncp to an address that is proposed as the new ncp", func(t *testing.T) {
		defer checkNCPStaker()

		receipt, err := g.ExpectedOk(g.NewProposalToAddNCP(t, ncp1.Operator, ncp2.Operator.Address))
		require.NoError(t, err)
		proposalEvent := findEvent("NewProposal", receipt.Logs)

		ExpectedRevert(t,
			g.ExpectedFail(g.ChangeNCP(t, ncp1.Operator, ncp2.Operator.Address)),
			"cannot change the ncp to an address that is proposed as the new ncp",
		)

		_, err = g.ExpectedOk(g.CancelProposal(t, ncp1.Operator, proposalEvent["id"].(*big.Int)))
		require.NoError(t, err)
	})

	t.Run("Set emergency", func(t *testing.T) {
		defer checkNCPStaker()

		// set emergency mode
		receipt, err := g.ExpectedOk(g.NewProposalEmergencyMode(t, ncp1.Operator, true))
		require.NoError(t, err)
		proposalEvent := findEvent("NewProposal", receipt.Logs)

		_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Vote(t, ncp4.Operator, proposalEvent["id"].(*big.Int), true))
		require.NoError(t, err)

		ExpectedRevert(t,
			g.ExpectedFail(g.Stake(t, ncp1.Operator, minStaking)),
			"operation not permitted by council",
		)

		// set emergency mode off
		receipt, err = g.ExpectedOk(g.NewProposalEmergencyMode(t, ncp1.Operator, false))
		require.NoError(t, err)
		proposalEvent = findEvent("NewProposal", receipt.Logs)

		_, err = g.ExpectedOk(g.Vote(t, ncp1.Operator, proposalEvent["id"].(*big.Int), true))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Vote(t, ncp4.Operator, proposalEvent["id"].(*big.Int), true))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Stake(t, ncp1.Operator, minStaking))
		require.NoError(t, err)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		ncpTotalStaking = ncpTotalStaking.Add(ncpTotalStaking, minStaking)
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

func distributeReward(t *testing.T, g *GovWBFT, stateDB *TestStateDB, rewardAmount *big.Int, stakers ...common.Address) {
	for _, _staker := range stakers {
		_rewardee := govwbft.StakerInfo(TestGovStakingAddress, stateDB, _staker).Rewardee
		_, err := g.ExpectedOk(TransferCoin(g.backend.Client(), g.owner, rewardAmount, &_rewardee))
		require.NoError(t, err)
	}
}

func TestGovReward(t *testing.T) {
	var (
		ctx                  = context.TODO()
		feeRate1             = new(big.Int).SetUint64(1000)
		feeRate2             = new(big.Int).SetUint64(500)
		minStaking           = towei(500000)
		totalStaking         = new(big.Int)
		calcRewardPerStaking = new(big.Int)
		stakers              = make([]common.Address, 0)
		rewardAmount         = towei(10)

		v1         = NewTestStaker()
		v2         = NewTestStaker()
		delegator1 = NewEOA()
		delegator2 = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		v2.Operator.Address: {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator1.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator2.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	checkGovBalanceFn := func() {
		if totalStaking.Cmp(g.balanceAt(t, ctx, TestGovStakingAddress, nil)) != 0 {
			t.Logf("total = %v, balance = %v", totalStaking, g.balanceAt(t, ctx, TestGovStakingAddress, nil))
		}
		require.True(t, totalStaking.Cmp(g.balanceAt(t, ctx, TestGovStakingAddress, nil)) == 0)
	}

	defer checkGovBalanceFn()

	t.Run("first staking", func(t *testing.T) {
		// v1.stake = 500000
		// v1.operator = 500000
		_, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking, feeRate1))
		require.NoError(t, err)
		stakers = append(stakers, v1.Staker.Address)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, minStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, 0, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking.Sign())
		require.Equal(t, 0, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance.Sign())
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).PendingReward.Sign())
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).RewardPerStaking.Sign())
	})

	t.Run("second staking", func(t *testing.T) {
		// v1.staking = 500000
		// v2.staking = 500000
		// v1.operator = 500000
		// v2.operator = 500000
		_, err = g.ExpectedOk(g.RegisterStaker(t, v2, minStaking, feeRate2))
		require.NoError(t, err)
		stakers = append(stakers, v2.Staker.Address)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, minStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).TotalStaked)
		require.Equal(t, 0, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).AccRewardPerStaking.Sign())
		require.Equal(t, 0, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).LastRewardBalance.Sign())
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, v2.Staker.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, v2.Staker.Address).PendingReward.Sign())
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, v2.Staker.Address).RewardPerStaking.Sign())
	})

	t.Run("delegator1 delegates to v1", func(t *testing.T) {
		// delegator1 delegates to v1
		// v1.staking = 1000000
		// v2.staking = 500000
		// v1.operator = 500000
		// v2.operator = 500000
		// delegator1 = 500000
		// v1.rewardee = 10
		// v2.rewardee = 10
		distributeReward(t, g, stateDB, rewardAmount, v1.Staker.Address, v2.Staker.Address)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		calcRewardPerStaking = toRewardPerStaking(rewardAmount, minStaking)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, new(big.Int).Add(minStaking, minStaking), govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, calcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking)
		require.Equal(t, rewardAmount, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance)
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).PendingReward.Sign())
		require.Equal(t, calcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).RewardPerStaking)
	})

	t.Run("v1 claim reward", func(t *testing.T) {
		// distribute reward to v1, v2
		// v1.staking = 1000000
		// v2.staking = 500000
		// v1.rewardee = 20 => 5
		// v2.rewardee = 20
		distributeReward(t, g, stateDB, rewardAmount, v1.Staker.Address, v2.Staker.Address)

		beforeBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)
		// v1 claim reward
		// v1.rewardee = 20 => v1.pendingReward = 15, delegator1.pendingReward = 5
		receipt, err := g.ExpectedOk(g.Claim(t, v1.Operator, v1.Staker.Address, false))
		require.NoError(t, err)
		gasCost := calcTxGasCost(receipt)
		calcRewardPerStaking = calcRewardPerStaking.Add(calcRewardPerStaking, toRewardPerStaking(rewardAmount, new(big.Int).Add(minStaking, minStaking)))
		expectedClaimed := towei(15)
		expectedBalance := towei(5)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, new(big.Int).Add(minStaking, minStaking), govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, calcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking)
		require.Equal(t, expectedBalance, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance)
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).PendingReward.Sign())
		require.Equal(t, calcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, v1.Staker.Address).RewardPerStaking)
		afterBalance := g.balanceAt(t, ctx, v1.Operator.Address, nil)
		require.Equal(t, expectedClaimed, afterBalance.Sub(afterBalance, beforeBalance).Add(afterBalance, gasCost))
	})

	t.Run("delegator1 delegates to v2", func(t *testing.T) {
		// delegator1 delegates to v2
		// v1.staking = 1000000
		// v2.staking = 1000000
		// v1.operator = 500000
		// v2.operator = 500000
		// delegator1 = 500000(to v1), 500000(to v2)
		// v1.rewardee = 15
		// v2.rewardee = 30
		distributeReward(t, g, stateDB, rewardAmount, v1.Staker.Address, v2.Staker.Address)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v2.Staker.Address, minStaking))
		require.NoError(t, err)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		v2Reward := new(big.Int).Mul(rewardAmount, new(big.Int).SetUint64(3))
		v2CalcRewardPerStaking := toRewardPerStaking(v2Reward, minStaking)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, new(big.Int).Add(minStaking, minStaking), govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).TotalStaked)
		require.Equal(t, v2CalcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).AccRewardPerStaking)
		require.Equal(t, v2Reward, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v2.Staker.Address).LastRewardBalance)
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, delegator1.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, delegator1.Address).PendingReward.Sign())
		require.Equal(t, v2CalcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v2.Staker.Address, delegator1.Address).RewardPerStaking)
	})

	t.Run("delegator2 delegates to v1", func(t *testing.T) {
		// delegator2 delegates to v1
		// v1.staking = 1500000
		// v2.staking = 1000000
		// v1.operator = 500000
		// v2.operator = 500000
		// delegator1 = 500000(to v1), 500000(to v2)
		// delegator2 = 500000(to v1)
		// v1.rewardee = 25
		// v2.rewardee = 40
		distributeReward(t, g, stateDB, rewardAmount, v1.Staker.Address, v2.Staker.Address)

		_, err = g.ExpectedOk(g.Delegate(t, delegator2, v1.Staker.Address, minStaking))
		require.NoError(t, err)
		totalStaking = totalStaking.Add(totalStaking, minStaking)
		v1Reward := towei(25)
		v1Staking := new(big.Int).Mul(minStaking, new(big.Int).SetUint64(2))
		calcRewardPerStaking = calcRewardPerStaking.Add(calcRewardPerStaking, toRewardPerStaking(new(big.Int).Mul(rewardAmount, new(big.Int).SetUint64(2)), v1Staking))
		v1Staking = v1Staking.Add(v1Staking, minStaking)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, v1Staking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, calcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking)
		require.Equal(t, v1Reward, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance)
		require.Equal(t, minStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator2.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator2.Address).PendingReward.Sign())
		require.Equal(t, calcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator2.Address).RewardPerStaking)
	})

	t.Run("delegator1 undelegates from v1", func(t *testing.T) {
		// delegator2 delegates to v1
		// v1.staking = 1000000
		// v2.staking = 1000000
		// v1.operator = 500000
		// v2.operator = 500000
		// delegator1 = 500000(to v2)
		// delegator2 = 500000(to v1)
		// v1.rewardee = 55
		// v2.rewardee = 80
		newRewardAmount := new(big.Int).Mul(rewardAmount, new(big.Int).SetUint64(3))
		distributeReward(t, g, stateDB, newRewardAmount, v1.Staker.Address, v2.Staker.Address)

		receipt, err := g.ExpectedOk(g.Undelegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)
		undelegateEvent := findEvent("NewCredential", receipt.Logs)
		require.NotNil(t, undelegateEvent)

		totalStaking = totalStaking.Sub(totalStaking, minStaking)
		v1Reward := towei(55)
		v1Staking := new(big.Int).Mul(minStaking, new(big.Int).SetUint64(3))
		calcRewardPerStaking = calcRewardPerStaking.Add(calcRewardPerStaking, toRewardPerStaking(newRewardAmount, v1Staking))
		v1Staking = v1Staking.Sub(v1Staking, minStaking)
		expectedReward := towei(25)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, v1Staking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, calcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking)
		require.Equal(t, v1Reward, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).StakingAmount.Sign())
		require.Equal(t, expectedReward, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).PendingReward)
		require.Equal(t, calcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).RewardPerStaking)

		// withdraw
		beforeBalance := g.balanceAt(t, ctx, delegator1.Address, nil)

		unbonding := undelegateEvent["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
		receipt, err = g.ExpectedOk(g.Withdraw(t, delegator1, common.Big0))
		require.NoError(t, err)

		gasCost := calcTxGasCost(receipt)
		expectedBalance := new(big.Int).Add(beforeBalance, new(big.Int).Sub(undelegateEvent["amount"].(*big.Int), gasCost))
		require.Equal(t, expectedBalance, g.balanceAt(t, ctx, delegator1.Address, nil))
	})

	t.Run("delegator1 claim reward to v1", func(t *testing.T) {
		// distribute reward to v1, v2
		// v1.staking = 1000000 + 25
		// v2.staking = 1000000
		// v1.operator = 500000
		// v2.operator = 500000
		// delegator1 = 25(to v1), 500000(to v2)
		// delegator2 = 500000(to v1)
		// v1.rewardee = 65 - 25 = 40
		// v2.rewardee = 90
		distributeReward(t, g, stateDB, rewardAmount, v1.Staker.Address, v2.Staker.Address)

		beforeBalance := g.balanceAt(t, ctx, delegator1.Address, nil)
		// v1 claim reward
		receipt, err := g.ExpectedOk(g.Claim(t, delegator1, v1.Staker.Address, true))
		require.NoError(t, err)

		gasCost := calcTxGasCost(receipt)
		v1Staking := new(big.Int).Add(minStaking, minStaking)
		calcRewardPerStaking = calcRewardPerStaking.Add(calcRewardPerStaking, toRewardPerStaking(rewardAmount, v1Staking))
		expectedClaimed := towei(25)
		fee := new(big.Int).Mul(expectedClaimed, feeRate1)
		fee = fee.Div(fee, new(big.Int).SetUint64(10000))
		expectedClaimed = expectedClaimed.Sub(expectedClaimed, fee)
		expectedBalance := towei(40)
		v1Staking = v1Staking.Add(v1Staking, expectedClaimed)
		totalStaking = totalStaking.Add(totalStaking, expectedClaimed)

		// check list
		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))
		require.Equal(t, v1Staking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).TotalStaked)
		require.Equal(t, calcRewardPerStaking, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).AccRewardPerStaking)
		require.Equal(t, expectedBalance, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).LastRewardBalance)
		require.Equal(t, expectedClaimed, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).StakingAmount)
		require.Equal(t, 0, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).PendingReward.Sign())
		require.Equal(t, calcRewardPerStaking, govwbft.UserInfo(TestGovStakingAddress, stateDB, v1.Staker.Address, delegator1.Address).RewardPerStaking)
		afterBalance := g.balanceAt(t, ctx, delegator1.Address, nil)
		require.Equal(t, afterBalance, beforeBalance.Sub(beforeBalance, gasCost))
		require.Equal(t, fee, g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil))
	})
}

func TestGovChangeFeeRate(t *testing.T) {
	var (
		ctx        = context.TODO()
		feeRate1   = new(big.Int).SetUint64(1000)
		feeRate2   = new(big.Int).SetUint64(500)
		minStaking = towei(500000)

		v1         = NewTestStaker()
		delegator1 = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		delegator1.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	checkFee := func(t *testing.T, staker common.Address, expectedFee *big.Int) {
		require.Equal(t, expectedFee, govwbft.StakerInfo(TestGovStakingAddress, stateDB, staker).FeeRate)
	}

	getConst := func(t *testing.T, method string) *big.Int {
		var out []interface{}
		err := g.govConst.Call(nil, &out, method)
		require.NoError(t, err)
		return *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	}

	t.Run("first staking", func(t *testing.T) {
		// v1.stake = 500000
		// v1.operator = 500000
		_, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking, feeRate1))
		require.NoError(t, err)

		// check list
		checkFee(t, v1.Staker.Address, feeRate1)
	})

	t.Run("change fee rate when there is no delegator", func(t *testing.T) {
		_, err := g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate2))
		require.NoError(t, err)

		// check list
		checkFee(t, v1.Staker.Address, feeRate2)
	})

	t.Run("cannot execute if there is no request", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.ExecuteChangingFee(t, v1.Operator, delegator1.Address)), "no request exists")

		ExpectedRevert(t,
			g.ExpectedFail(g.ExecuteChangingFee(t, v1.Operator, v1.Staker.Address)), "no request exists")
	})

	t.Run("change fee rate when there is a delegator", func(t *testing.T) {
		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		_, err := g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate1))
		require.NoError(t, err)

		// check list
		checkFee(t, v1.Staker.Address, feeRate2) // changing not applied yet
	})

	t.Run("after CHANGE_FEE_DELAY", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.ExecuteChangingFee(t, v1.Operator, v1.Staker.Address)), "the request cannot be executed before delay time")

		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)

		_, err := g.ExpectedOk(g.ExecuteChangingFee(t, delegator1, v1.Staker.Address)) // anyone can ExecuteChangingFee
		require.NoError(t, err)

		// check list
		checkFee(t, v1.Staker.Address, feeRate1)
	})
}

func TestGovFeeRateConsistency(t *testing.T) {
	var (
		ctx        = context.TODO()
		feeRate1   = new(big.Int).SetUint64(1000)
		feeRate2   = new(big.Int).SetUint64(2000)
		feeRate3   = new(big.Int).SetUint64(3000)
		feeRate4   = new(big.Int).SetUint64(4000)
		feeRate5   = new(big.Int).SetUint64(5000)
		minStaking = towei(500000)

		v1         = NewTestStaker()
		delegator1 = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		delegator1.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	getConst := func(t *testing.T, method string) *big.Int {
		var out []interface{}
		err := g.govConst.Call(nil, &out, method)
		require.NoError(t, err)
		return *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	}

	claimAndCheck := func(t *testing.T, who *EOA, expectedClaimed *big.Int, expectedFee *big.Int) {
		beforeBalance := g.balanceAt(t, ctx, who.Address, nil)
		beforeFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		receipt, err := g.ExpectedOk(g.Claim(t, who, v1.Staker.Address, false))
		require.NoError(t, err)

		gasCost := calcTxGasCost(receipt)
		afterBalance := g.balanceAt(t, ctx, who.Address, nil)
		afterFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		if expectedClaimed.Sign() == 0 {
			require.Equal(t, 0, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)).Sign())
		} else {
			require.Equal(t, expectedClaimed, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)))
		}

		if expectedFee.Sign() == 0 {
			require.Equal(t, 0, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance).Sign())
		} else {
			require.Equal(t, expectedFee, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance))
		}
	}

	t.Run("first staking and delegation", func(t *testing.T) {
		// v1.stake = 1000000
		// v1.operator = 500000
		// delegator = 500000
		_, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking, feeRate1))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)
	})

	t.Run("check fee at first claiming", func(t *testing.T) {
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		// v1.Operator claims
		claimAndCheck(t, v1.Operator, towei(50), common.Big0)

		// delegator1 claims
		claimAndCheck(t, delegator1, towei(45), towei(5))
	})

	t.Run("fee of claiming after changing fee", func(t *testing.T) {
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		_, err := g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate2))
		require.NoError(t, err)
		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)
		claimAndCheck(t, delegator1, towei(45), towei(5)) // feeRate1 should be applied

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		claimAndCheck(t, delegator1, towei(40), towei(10)) // feeRate2 should be applied
	})

	t.Run("fee changes several", func(t *testing.T) {
		// fee2 should be applied
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		_, err := g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate3))
		require.NoError(t, err)
		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)
		claimAndCheck(t, v1.Operator, towei(150), common.Big0)

		// fee3 should be applied
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		_, err = g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate4))
		require.NoError(t, err)
		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)
		claimAndCheck(t, v1.Operator, towei(50), common.Big0)

		// fee4 should be applied
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		_, err = g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate5))
		require.NoError(t, err)
		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)
		claimAndCheck(t, v1.Operator, towei(50), common.Big0)

		// fee5 should be applied
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)
		expectedClaimed := new(big.Int).Add(towei(40), towei(35))
		expectedClaimed = expectedClaimed.Add(expectedClaimed, towei(30))
		expectedClaimed = expectedClaimed.Add(expectedClaimed, towei(25))
		expectedFee := new(big.Int).Add(towei(10), towei(15))
		expectedFee = expectedFee.Add(expectedFee, towei(20))
		expectedFee = expectedFee.Add(expectedFee, towei(25))
		claimAndCheck(t, delegator1, expectedClaimed, expectedFee)
	})
}

func TestClaimForUnstakedStaker(t *testing.T) {
	var (
		ctx        = context.TODO()
		feeRate1   = new(big.Int).SetUint64(1000)
		feeRate2   = new(big.Int).SetUint64(2000)
		minStaking = towei(500000)

		v1         = NewTestStaker()
		delegator1 = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		delegator1.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	getConst := func(t *testing.T, method string) *big.Int {
		var out []interface{}
		err := g.govConst.Call(nil, &out, method)
		require.NoError(t, err)
		return *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	}

	claimAndCheck := func(t *testing.T, who *EOA, expectedClaimed *big.Int, expectedFee *big.Int) {
		beforeBalance := g.balanceAt(t, ctx, who.Address, nil)
		beforeFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		receipt, err := g.ExpectedOk(g.Claim(t, who, v1.Staker.Address, false))
		require.NoError(t, err)

		gasCost := calcTxGasCost(receipt)
		afterBalance := g.balanceAt(t, ctx, who.Address, nil)
		afterFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		if expectedClaimed.Sign() == 0 {
			require.Equal(t, 0, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)).Sign())
		} else {
			require.Equal(t, expectedClaimed, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)))
		}

		if expectedFee.Sign() == 0 {
			require.Equal(t, 0, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance).Sign())
		} else {
			require.Equal(t, expectedFee, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance))
		}
	}

	t.Run("preparation", func(t *testing.T) {
		// preparation:
		//  - register staker
		//  - delegation
		//  - request changing fee
		//  - distribute reward
		//  - unstake
		_, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking, feeRate1))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.RequestChangingFee(t, v1.Operator, feeRate2))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		_, err = g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
		require.NoError(t, err)
	})

	t.Run("v1 can claim", func(t *testing.T) {
		claimAndCheck(t, v1.Operator, towei(50), common.Big0)
	})

	t.Run("delegator1 can claim", func(t *testing.T) {
		g.adjustTime(time.Duration(int64(getConst(t, "changeFeeDelay").Uint64())) * time.Second)

		// cannot re-stake to unregistered staker
		ExpectedRevert(t,
			g.ExpectedFail(g.Claim(t, delegator1, v1.Staker.Address, true)), "staker is inactive")

		// claim and execute changing fee
		claimAndCheck(t, delegator1, towei(45), towei(5))
	})

	t.Run("cannot operate any more", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.ExecuteChangingFee(t, v1.Operator, v1.Staker.Address)), "no request exists")

		ExpectedRevert(t,
			g.ExpectedFail(g.Delegate(t, delegator1, v1.Staker.Address, minStaking)), "staker is inactive")
	})

	t.Run("staker can restake", func(t *testing.T) {
		_, err := g.ExpectedOk(g.Stake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		// check list
		newStakers := govwbft.Stakers(TestGovStakingAddress, stateDB)
		require.Equal(t, 1, len(newStakers))
		require.Equal(t, v1.Staker.Address, newStakers[0])
	})

	t.Run("delegator1 can undelegate and claim with changed fee", func(t *testing.T) {
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		_, err = g.ExpectedOk(g.Undelegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address) // will not be applied
		claimAndCheck(t, delegator1, towei(40), towei(10))             // fee2 should be applied
	})

	t.Run("delegator1 can delegate again", func(t *testing.T) {
		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		claimAndCheck(t, delegator1, towei(40), towei(10))
	})
}

func TestZeroTotalStaking(t *testing.T) {
	var (
		ctx        = context.TODO()
		feeRate    = new(big.Int).SetUint64(1000)
		minStaking = towei(500000)

		v1         = NewTestStaker()
		delegator1 = NewEOA()
		delegator2 = NewEOA()
		delegator3 = NewEOA()
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		v1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		delegator1.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator2.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
		delegator3.Address:  {Balance: new(big.Int).Add(MAX_UINT_128, minStaking)},
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

	claimAndCheck := func(t *testing.T, who *EOA, expectedClaimed *big.Int, expectedFee *big.Int) {
		beforeBalance := g.balanceAt(t, ctx, who.Address, nil)
		beforeFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		receipt, err := g.ExpectedOk(g.Claim(t, who, v1.Staker.Address, false))
		require.NoError(t, err)

		gasCost := calcTxGasCost(receipt)
		afterBalance := g.balanceAt(t, ctx, who.Address, nil)
		afterFeeBalance := g.balanceAt(t, ctx, govwbft.StakerInfo(TestGovStakingAddress, stateDB, v1.Staker.Address).FeeRecipient, nil)

		if expectedClaimed.Sign() == 0 {
			require.Equal(t, 0, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)).Sign())
		} else {
			require.Equal(t, expectedClaimed, afterBalance.Sub(afterBalance, new(big.Int).Sub(beforeBalance, gasCost)))
		}

		if expectedFee.Sign() == 0 {
			require.Equal(t, 0, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance).Sign())
		} else {
			require.Equal(t, expectedFee, afterFeeBalance.Sub(afterFeeBalance, beforeFeeBalance))
		}
	}

	t.Run("preparation", func(t *testing.T) {
		// preparation:
		//  - register staker
		//  - delegation
		//  - request changing fee
		//  - distribute reward
		//  - unstake
		_, err := g.ExpectedOk(g.RegisterStaker(t, v1, minStaking, feeRate))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		_, err = g.ExpectedOk(g.Undelegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		// make total staking zero
		receipt, err := g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		unstakeEvent := findEvent("NewCredential", receipt.Logs)
		require.NotNil(t, unstakeEvent)

		unbonding := unstakeEvent["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)

		// withdraw all
		_, err = g.ExpectedOk(g.Withdraw(t, v1.Operator, common.Big0))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Withdraw(t, delegator1, common.Big0))
		require.NoError(t, err)

		stakingBalance := g.balanceAt(t, ctx, TestGovStakingAddress, nil)
		require.Equal(t, stakingBalance.Sign(), 0)
	})

	t.Run("delegator1 can claim", func(t *testing.T) {
		// miss distribution for no staking; no one can take this reward
		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		// this should not include miss distribution
		claimAndCheck(t, delegator1, towei(45), towei(5))
	})

	t.Run("v1 restake", func(t *testing.T) {
		_, err := g.ExpectedOk(g.Stake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address)

		_, err = g.ExpectedOk(g.Delegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Delegate(t, delegator2, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		_, err = g.ExpectedOk(g.Delegate(t, delegator3, v1.Staker.Address, minStaking))
		require.NoError(t, err)
	})

	t.Run("v1, delegator1 claim", func(t *testing.T) {
		distributeReward(t, g, stateDB, towei(200), v1.Staker.Address)

		claimAndCheck(t, v1.Operator, towei(300), common.Big0)
		claimAndCheck(t, delegator1, towei(45), towei(5))
	})

	t.Run("all unstake", func(t *testing.T) {
		_, err := g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(300), v1.Staker.Address) // reward for delegator1, 2, 3

		_, err = g.ExpectedOk(g.Undelegate(t, delegator1, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(200), v1.Staker.Address) // reward for delegator2, 3

		_, err = g.ExpectedOk(g.Undelegate(t, delegator2, v1.Staker.Address, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address) // reward for delegator3

		_, err = g.ExpectedOk(g.Undelegate(t, delegator3, v1.Staker.Address, minStaking))
		require.NoError(t, err)
	})

	t.Run("v1 restake", func(t *testing.T) {
		_, err := g.ExpectedOk(g.Stake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		distributeReward(t, g, stateDB, towei(100), v1.Staker.Address) // reward for v1
	})

	t.Run("all claim", func(t *testing.T) {
		claimAndCheck(t, delegator1, towei(90), towei(10))
		claimAndCheck(t, delegator2, towei(225), towei(25))

		_, err := g.ExpectedOk(g.Unstake(t, v1.Operator, minStaking))
		require.NoError(t, err)

		claimAndCheck(t, delegator3, towei(315), towei(35))
	})
}

func TestSetCode(t *testing.T) {
	var testParams = map[string]string{
		"minimumStaking":           "100000000000000000000000",
		"maximumStaking":           "100000000000000000000000000",
		"unbondingPeriodStaker":    "10800",  // 3 hours
		"unbondingPeriodDelegator": "259200", // 3 days
		"feePrecision":             "100",
		"changeFeeDelay":           "3600", // 1 hour
	}

	var (
		testVersion  = "test_version"
		ctx          = context.TODO()
		minStaking1  = towei(500000)
		minStaking2  = towei(100000)
		feeRate1     = new(big.Int).SetUint64(15)
		feeRate2     = new(big.Int).SetUint64(1500)
		totalStaking = new(big.Int)
		stakers      = make([]common.Address, 0)

		ncp1 = NewTestStaker()
		ncp2 = NewTestStaker()
		ncp3 = NewTestStaker()
	)

	// register upgrading contract
	govwbft.GovContractCodes[govwbft.CONTRACT_GOV_CONFIG][testVersion] = govwbft.GovContractCodes[govwbft.CONTRACT_GOV_CONFIG][params.DefaultGovVersion]

	// for duplicate test
	ncpInput := []common.Address{ncp3.Operator.Address, ncp3.Operator.Address}

	g, err := NewGovWBFT(t, ncpInput, types.GenesisAlloc{
		ncp1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
		ncp2.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
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

	t.Run("duplicated ncp", func(t *testing.T) {
		ncps := []common.Address{ncp3.Operator.Address}
		require.Equal(t, ncps, govwbft.NCPList(TestGovNCPAddress, stateDB))
	})

	t.Run("add staker", func(t *testing.T) {
		require.True(t, govwbft.TotalStaking(TestGovStakingAddress, stateDB).Sign() == 0)
		require.True(t, len(govwbft.Stakers(TestGovStakingAddress, stateDB)) == 0)
		beforeBalance := g.balanceAt(t, ctx, ncp1.Operator.Address, nil)

		receipt, err := g.ExpectedOk(g.RegisterStaker(t, ncp1, minStaking1, feeRate1))
		require.NoError(t, err)
		stakers = append(stakers, ncp1.Staker.Address)
		totalStaking.Add(totalStaking, minStaking1)

		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

		gasCost := calcTxGasCost(receipt)
		expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking1, gasCost))
		require.Equal(t, expectedBalance, g.balanceAt(t, ctx, ncp1.Operator.Address, nil))

		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, minStaking2, feeRate1)),
			"out of bounds",
		)
	})

	t.Run("upgrade contract", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, new(big.Int).Sub(minStaking1, common.Big1), feeRate1)),
			"out of bounds",
		)

		// upgrade contract
		g.backend.CommitWithState(&params.GovContracts{
			GovConfig: &params.GovContract{
				Address: TestGovConfigAddress,
				Version: testVersion,
				Params:  testParams,
			},
		}, nil)
	})

	t.Run("retry add staker", func(t *testing.T) {
		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, new(big.Int).Sub(minStaking2, common.Big1), feeRate1)),
			"out of bounds",
		)
		ExpectedRevert(t,
			g.ExpectedFail(g.RegisterStaker(t, ncp2, minStaking2, feeRate2)),
			"fee rate exceeds precision",
		)

		beforeBalance := g.balanceAt(t, ctx, ncp2.Operator.Address, nil)

		receipt, err := g.ExpectedOk(g.RegisterStaker(t, ncp2, minStaking2, feeRate1))
		require.NoError(t, err)
		stakers = append(stakers, ncp2.Staker.Address)
		totalStaking.Add(totalStaking, minStaking2)

		require.Equal(t, totalStaking, govwbft.TotalStaking(TestGovStakingAddress, stateDB))
		require.Equal(t, stakers, govwbft.Stakers(TestGovStakingAddress, stateDB))

		gasCost := calcTxGasCost(receipt)
		expectedBalance := new(big.Int).Sub(beforeBalance, new(big.Int).Add(minStaking2, gasCost))
		require.Equal(t, expectedBalance, g.balanceAt(t, ctx, ncp2.Operator.Address, nil))

		// restore GovConfig
		g.backend.CommitWithState(&params.GovContracts{
			GovConfig: &params.GovContract{
				Address: TestGovConfigAddress,
				Version: govwbft.GOV_CONTRACT_VERSION_1,
				Params: map[string]string{
					govwbft.GOV_CONFIG_PARAM_MINIMUM_STAKING:     towei(500000).String(),
					govwbft.GOV_CONFIG_PARAM_MAXIMUM_STAKING:     (new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))).String(),
					govwbft.GOV_CONFIG_PARAM_UNBONDING_STAKER:    "604800",
					govwbft.GOV_CONFIG_PARAM_UNBONDING_DELEGATOR: "259200",
					govwbft.GOV_CONFIG_PARAM_FEE_PRECISION:       "10000",
					govwbft.GOV_CONFIG_PARAM_CHANGE_FEE_DELAY:    "604800",
				},
			},
		}, nil)
	})
}

func TestGovGetBls(t *testing.T) {
	var (
		ctx        = context.TODO()
		minStaking = towei(500000)
		s1         = NewTestStaker()
		feeRate    = new(big.Int).SetUint64(100)
	)

	g, err := NewGovWBFT(t, nil, types.GenesisAlloc{
		s1.Operator.Address: {Balance: new(big.Int).Mul(MAX_UINT_128, common.Big2)},
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

	g.ExpectedOk(g.RegisterStaker(t, s1, minStaking, feeRate))

	t.Run("Get BLS PublicKey", func(t *testing.T) {
		expected, err := s1.GetBLSPublicKey()
		require.NoError(t, err)
		pubKeyFromState := govwbft.GetBLSPublicKey(TestGovStakingAddress, stateDB, s1.Staker.Address)

		actual, err := blst.PublicKeyFromBytes(pubKeyFromState)
		require.NoError(t, err)

		require.True(t, expected.Equals(actual))
	})

	t.Run("From Disabled Staker", func(t *testing.T) {
		receipt, err := g.ExpectedOk(g.Unstake(t, s1.Operator, minStaking))
		require.NoError(t, err)

		require.True(t, govwbft.TotalStaking(TestGovStakingAddress, stateDB).Sign() == 0)

		unstakeEvent := findEvent("NewCredential", receipt.Logs)
		require.NotNil(t, unstakeEvent)

		unbonding := unstakeEvent["unbonding"].(*big.Int)
		g.adjustTime(time.Duration(unbonding.Int64()) * time.Second)
		_, err = g.ExpectedOk(g.Withdraw(t, s1.Operator, common.Big0))
		require.NoError(t, err)

		expected, err := s1.GetBLSPublicKey()
		require.NoError(t, err)
		pubKeyFromState := govwbft.GetBLSPublicKey(TestGovStakingAddress, stateDB, s1.Staker.Address)

		actual, err := blst.PublicKeyFromBytes(pubKeyFromState)
		require.NoError(t, err)

		require.True(t, expected.Equals(actual))
	})
}

func setWbftGovConfig(g *GovWBFT) {
	g.backend.CommitWithState(&params.GovContracts{
		GovConfig: &params.GovContract{
			Address: TestGovConfigAddress,
			Version: govwbft.GOV_CONTRACT_VERSION_1,
			Params: map[string]string{
				govwbft.GOV_CONFIG_PARAM_MINIMUM_STAKING:     towei(500000).String(),
				govwbft.GOV_CONFIG_PARAM_MAXIMUM_STAKING:     (new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))).String(),
				govwbft.GOV_CONFIG_PARAM_UNBONDING_STAKER:    "604800",
				govwbft.GOV_CONFIG_PARAM_UNBONDING_DELEGATOR: "259200",
				govwbft.GOV_CONFIG_PARAM_FEE_PRECISION:       "10000",
				govwbft.GOV_CONFIG_PARAM_CHANGE_FEE_DELAY:    "604800",
			},
		},
	}, nil)
}

func setWbftGovConfigWithGovCouncil(g *GovWBFT) {
	g.backend.CommitWithState(&params.GovContracts{
		GovConfig: &params.GovContract{
			Address: TestGovConfigAddress,
			Version: govwbft.GOV_CONTRACT_VERSION_1,
			Params: map[string]string{
				govwbft.GOV_CONFIG_PARAM_MINIMUM_STAKING:     towei(500000).String(),
				govwbft.GOV_CONFIG_PARAM_MAXIMUM_STAKING:     (new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))).String(),
				govwbft.GOV_CONFIG_PARAM_UNBONDING_STAKER:    "604800",
				govwbft.GOV_CONFIG_PARAM_UNBONDING_DELEGATOR: "259200",
				govwbft.GOV_CONFIG_PARAM_FEE_PRECISION:       "10000",
				govwbft.GOV_CONFIG_PARAM_CHANGE_FEE_DELAY:    "604800",
				govwbft.GOV_CONFIG_PARAM_GOV_COUNCIL:         TestGovNCPAddress.String(),
			},
		},
	}, nil)
}

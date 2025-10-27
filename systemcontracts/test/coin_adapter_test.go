package test

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/stretchr/testify/require"
)

func TestNativeCoinAdapter(t *testing.T) {
	var (
		ctx                = context.Background()
		masterMinter       = NewEOA()
		minter1            = NewEOA()
		decimals     uint8 = 18

		initialBalance = toWeiN(1_000_000, decimals) // for gas
		allowedAmount  = toWeiN(10_000_000, decimals)
	)

	g, err := NewGovWBFT(t, types.GenesisAlloc{
		masterMinter.Address: {Balance: initialBalance},
		minter1.Address:      {Balance: initialBalance},
	}, nil, func(coinAdapter *params.SystemContract) {
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MASTER_MINTER] = masterMinter.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTERS] = minter1.Address.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_MINTER_ALLOWED] = allowedAmount.String()
		coinAdapter.Params[systemcontracts.COIN_ADAPTER_PARAM_DECIMALS] = strconv.Itoa(int(decimals))
	}, nil, nil)
	require.NoError(t, err)

	t.Run("initialize", func(t *testing.T) {
		// masterMinter
		require.Equal(t, masterMinter.Address, contractCall(t, g.coinAdapter, "masterMinter")[0].(common.Address))
		require.False(t, contractCall(t, g.coinAdapter, "isMinter", masterMinter.Address)[0].(bool))

		// minter
		require.True(t, contractCall(t, g.coinAdapter, "isMinter", minter1.Address)[0].(bool))
		require.True(t, allowedAmount.Cmp(contractCall(t, g.coinAdapter, "minterAllowance", minter1.Address)[0].(*big.Int)) == 0)

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
			actualTotalSupply := contractCall(t, g.coinAdapter, "totalSupply")[0].(*big.Int)
			require.True(t, actualTotalSupply.Cmp(expectedTotalSupply) == 0)
		}
	})

	t.Run("mint", func(t *testing.T) {

	})
}

func (g *GovWBFT) BalanceOf(t *testing.T, address common.Address) *big.Int {
	return contractCall(t, g.coinAdapter, "balanceOf", address)[0].(*big.Int)
}

// TODO
// transfer
// // contract function
// // call
// // to precompiled

// mint/burn
// mint allowance
// add minter, remove minter
// permit
// transferWithAuthorization

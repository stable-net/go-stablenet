// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package txpool

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

type account struct {
	privKey *ecdsa.PrivateKey
	address common.Address
}

var (
	chainConfig       *params.ChainConfig
	signer            types.Signer
	feeDelegateSigner types.Signer
	initialBalance    *uint256.Int
)

func init() {
	chainConfig = params.TestWBFTChainConfig
	signer = types.LatestSigner(chainConfig)
	feeDelegateSigner = types.NewFeeDelegateSigner(chainConfig.ChainID)
	initialBalance = uint256.MustFromDecimal("1000000000000000000000")
}

func newStateDB() *state.StateDB {
	statedb, _ := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	return statedb
}

func newOptions(statedb *state.StateDB) *ValidationOptionsWithState {
	opts := &ValidationOptionsWithState{
		Config:              params.TestWBFTChainConfig,
		State:               statedb,
		FirstNonceGap:       nil,
		UsedAndLeftSlots:    nil,
		ExistingExpenditure: nil,
		ExistingCost:        nil,
	}
	return opts
}

func newTestState() (*state.StateDB, *ValidationOptionsWithState) {
	statedb := newStateDB()
	return statedb, newOptions(statedb)
}

func newAccount(statedb *state.StateDB) *account {
	key, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(key.PublicKey)

	account := &account{key, address}
	statedb.CreateAccount(account.address)
	statedb.AddBalance(account.address, initialBalance)

	return account
}

func newDynamicFeeTx(to *common.Address) types.DynamicFeeTx {
	return types.DynamicFeeTx{
		ChainID:    chainConfig.ChainID,
		Nonce:      0,
		GasTipCap:  new(big.Int).SetUint64(params.InitialGasTip),
		GasFeeCap:  new(big.Int).Add(new(big.Int).SetUint64(params.MinBaseFee), new(big.Int).SetUint64(params.InitialGasTip)),
		Gas:        params.TxGas,
		To:         to,
		Value:      big.NewInt(1000000000000000000),
		Data:       nil,
		AccessList: nil,
	}
}

func signDynamicFeeTx(tx types.DynamicFeeTx, sender *account) (*types.Transaction, error) {
	return types.SignNewTx(sender.privKey, signer, &tx)
}

func signFeeDelegateTx(rawTx *types.Transaction, feePayer *account) (*types.Transaction, error) {
	V, R, S := rawTx.RawSignatureValues()
	senderTx := types.DynamicFeeTx{
		To:         rawTx.To(),
		ChainID:    rawTx.ChainId(),
		Nonce:      rawTx.Nonce(),
		Gas:        rawTx.Gas(),
		GasFeeCap:  rawTx.GasFeeCap(),
		GasTipCap:  rawTx.GasTipCap(),
		Value:      rawTx.Value(),
		Data:       rawTx.Data(),
		AccessList: rawTx.AccessList(),
		V:          V,
		R:          R,
		S:          S,
	}

	feeDelegateTx := &types.FeeDelegateDynamicFeeTx{
		FeePayer: &feePayer.address,
	}
	feeDelegateTx.SetSenderTx(senderTx)
	tx := types.NewTx(feeDelegateTx)

	return types.SignTx(tx, feeDelegateSigner, feePayer.privKey)
}

func TestBlacklistedAccountTx(t *testing.T) {
	t.Run("DynamicFeeTx", func(t *testing.T) {
		tests := []struct {
			name            string
			blacklistedRole core.BlacklistRole
			expectErr       bool
		}{
			{
				name:            "unrelated to any blacklisted account",
				blacklistedRole: core.NoneRole,
				expectErr:       false,
			},
			{
				name:            "sender is blacklisted",
				blacklistedRole: core.SenderRole,
				expectErr:       true,
			},
			{
				name:            "recipient is blacklisted",
				blacklistedRole: core.RecipientRole,
				expectErr:       true,
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb, opts := newTestState()
				testAccts := map[core.BlacklistRole]*account{
					core.SenderRole:    newAccount(statedb),
					core.RecipientRole: newAccount(statedb),
				}

				if tc.blacklistedRole != core.NoneRole {
					blacklistedAcct := testAccts[tc.blacklistedRole]
					statedb.SetBlacklisted(blacklistedAcct.address)
				}

				sender := testAccts[core.SenderRole]
				recipient := testAccts[core.RecipientRole]

				tx := newDynamicFeeTx(&recipient.address)
				signedTx, _ := signDynamicFeeTx(tx, sender)

				err := ValidateTransactionWithState(signedTx, signer, opts)

				if tc.expectErr {
					require.Error(t, err)

					var haveErr *core.ErrBlacklistedAccount
					require.ErrorAs(t, err, &haveErr)
					require.Equal(t, tc.blacklistedRole, haveErr.Role)
					require.Equal(t, testAccts[tc.blacklistedRole].address, haveErr.Address)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("FeeDelegateDynamicFeeTx", func(t *testing.T) {
		tests := []struct {
			name            string
			blacklistedRole core.BlacklistRole
			expectErr       bool
		}{
			{
				name:            "unrelated to any blacklisted account",
				blacklistedRole: core.NoneRole,
				expectErr:       false,
			},
			{
				name:            "sender is blacklisted",
				blacklistedRole: core.SenderRole,
				expectErr:       true,
			},
			{
				name:            "recipient is blacklisted",
				blacklistedRole: core.RecipientRole,
				expectErr:       true,
			},
			{
				name:            "feePayer is blacklisted",
				blacklistedRole: core.FeePayerRole,
				expectErr:       true,
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb, opts := newTestState()
				testAccts := map[core.BlacklistRole]*account{
					core.SenderRole:    newAccount(statedb),
					core.RecipientRole: newAccount(statedb),
					core.FeePayerRole:  newAccount(statedb),
				}

				if tc.blacklistedRole != core.NoneRole {
					blacklistedAcct := testAccts[tc.blacklistedRole]
					statedb.SetBlacklisted(blacklistedAcct.address)
				}

				sender := testAccts[core.SenderRole]
				recipient := testAccts[core.RecipientRole]
				feePayer := testAccts[core.FeePayerRole]

				tx := newDynamicFeeTx(&recipient.address)
				signedTx, _ := signDynamicFeeTx(tx, sender)
				feePayerSignedTx, _ := signFeeDelegateTx(signedTx, feePayer)

				err := ValidateTransactionWithState(feePayerSignedTx, signer, opts)

				if tc.expectErr {
					require.Error(t, err)

					var haveErr *core.ErrBlacklistedAccount
					require.ErrorAs(t, err, &haveErr)

					require.Equal(t, tc.blacklistedRole, haveErr.Role)
					require.Equal(t, testAccts[tc.blacklistedRole].address, haveErr.Address)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

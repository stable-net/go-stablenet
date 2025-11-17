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
	"fmt"
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

type testAccounts struct {
	sender    *account
	recipient *account
	feePayer  *account
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

func buildAccounts(statedb *state.StateDB, senderBlacklisted, recipientBlacklisted, feePayerBlacklisted bool) testAccounts {
	var accounts testAccounts

	accounts.sender = newAccount(statedb)
	if senderBlacklisted {
		statedb.SetBlacklisted(accounts.sender.address)
	}
	accounts.recipient = newAccount(statedb)
	if recipientBlacklisted {
		statedb.SetBlacklisted(accounts.recipient.address)
	}
	accounts.feePayer = newAccount(statedb)
	if feePayerBlacklisted {
		statedb.SetBlacklisted(accounts.feePayer.address)
	}

	return accounts
}

func TestBlacklistedAccountTx(t *testing.T) {
	t.Run("DynamicFeeTx", func(t *testing.T) {
		tests := []struct {
			name                 string
			senderBlacklisted    bool
			recipientBlacklisted bool
			expectErr            bool
			getErrPart           func(accts testAccounts) string
		}{
			{
				name:                 "unrelated to any blacklisted account",
				senderBlacklisted:    false,
				recipientBlacklisted: false,
				expectErr:            false,
			},
			{
				name:                 "sender is blacklisted",
				senderBlacklisted:    true,
				recipientBlacklisted: false,
				expectErr:            true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("from %s", accts.sender.address.Hex())
				},
			},
			{
				name:                 "recipient is blacklisted",
				senderBlacklisted:    false,
				recipientBlacklisted: true,
				expectErr:            true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("to %s", accts.recipient.address.Hex())
				},
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb, opts := newTestState()

				accts := buildAccounts(statedb, tc.senderBlacklisted, tc.recipientBlacklisted, false)

				sender := accts.sender
				recipient := accts.recipient

				tx := newDynamicFeeTx(&recipient.address)
				signedTx, _ := signDynamicFeeTx(tx, sender)

				err := ValidateTransactionWithState(signedTx, signer, opts)
				if tc.expectErr {
					require.ErrorIs(t, err, core.ErrBlacklistedAccount)
					require.Contains(t, err.Error(), tc.getErrPart(accts))
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("FeeDelegateDynamicFeeTx", func(t *testing.T) {
		tests := []struct {
			name                 string
			senderBlacklisted    bool
			recipientBlacklisted bool
			feePayerBlacklisted  bool
			expectErr            bool
			getErrPart           func(accts testAccounts) string
		}{
			{
				name:                 "unrelated to any blacklisted account",
				senderBlacklisted:    false,
				recipientBlacklisted: false,
				feePayerBlacklisted:  false,
				expectErr:            false,
			},
			{
				name:                 "sender is blacklisted",
				senderBlacklisted:    true,
				recipientBlacklisted: false,
				feePayerBlacklisted:  false,
				expectErr:            true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("from %s", accts.sender.address.Hex())
				},
			},
			{
				name:                 "recipient is blacklisted",
				senderBlacklisted:    false,
				recipientBlacklisted: true,
				feePayerBlacklisted:  false,
				expectErr:            true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("to %s", accts.recipient.address.Hex())
				},
			},
			{
				name:                 "fee payer is blacklisted",
				senderBlacklisted:    false,
				recipientBlacklisted: false,
				feePayerBlacklisted:  true,
				expectErr:            true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("fee payer %s", accts.feePayer.address.Hex())
				},
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb, opts := newTestState()

				accts := buildAccounts(statedb, tc.senderBlacklisted, tc.recipientBlacklisted, tc.feePayerBlacklisted)

				sender := accts.sender
				recipient := accts.recipient
				feePayer := accts.feePayer

				tx := newDynamicFeeTx(&recipient.address)
				signedTx, _ := signDynamicFeeTx(tx, sender)
				feePayerSignedTx, _ := signFeeDelegateTx(signedTx, feePayer)

				err := ValidateTransactionWithState(feePayerSignedTx, signer, opts)
				if tc.expectErr {
					require.ErrorIs(t, err, core.ErrBlacklistedAccount)
					require.Contains(t, err.Error(), tc.getErrPart(accts))
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

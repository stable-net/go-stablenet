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

package vm

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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
	caller *account
	target *account
}

func buildAccounts(statedb StateDB, callerBlacklisted, targetBlacklisted bool) testAccounts {
	var accounts testAccounts

	accounts.caller = newAccount(statedb)
	if callerBlacklisted {
		statedb.SetBlacklisted(accounts.caller.address)
	}
	accounts.target = newAccount(statedb)
	if targetBlacklisted {
		statedb.SetBlacklisted(accounts.target.address)
	}

	return accounts
}

func newAccount(statedb StateDB) *account {
	key, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(key.PublicKey)

	account := &account{key, address}
	statedb.CreateAccount(account.address)
	statedb.AddBalance(account.address, uint256.NewInt(0))

	return account
}

func newStateDB() *state.StateDB {
	statedb, _ := state.New(types.EmptyRootHash, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	return statedb
}

func newTestEvm(statedb StateDB) *EVM {
	vmctx := BlockContext{
		Transfer:    func(StateDB, common.Address, common.Address, *uint256.Int) {},
		CanTransfer: func(sd StateDB, a common.Address, i *uint256.Int) bool { return true },
	}
	return NewEVM(vmctx, TxContext{}, statedb, params.TestWBFTChainConfig, Config{})
}

func TestBlacklistedAccountTx(t *testing.T) {
	t.Run("Call", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name              string
			callerBlacklisted bool
			targetBlacklisted bool
			expectErr         bool
			getErrPart        func(accts testAccounts) string
		}{
			{
				name:              "unrelated to any blacklisted account",
				callerBlacklisted: false,
				targetBlacklisted: false,
				expectErr:         false,
			},
			{
				name:              "caller is blacklisted",
				callerBlacklisted: true,
				targetBlacklisted: false,
				expectErr:         true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("caller %s", accts.caller.address.Hex())
				},
			},
			{
				name:              "target is blacklisted",
				callerBlacklisted: false,
				targetBlacklisted: true,
				expectErr:         true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("target %s", accts.target.address.Hex())
				},
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb := newStateDB()
				evm := newTestEvm(statedb)
				accts := buildAccounts(statedb, tc.callerBlacklisted, tc.targetBlacklisted)

				caller := accts.caller
				target := accts.target

				callerRef := AccountRef(caller.address)

				_, _, err := evm.Call(callerRef, target.address, []byte{}, 0, uint256.NewInt(0))
				if tc.expectErr {
					require.ErrorIs(t, err, ErrBlacklistedAccount)
					require.Contains(t, err.Error(), tc.getErrPart(accts))
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Create", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name              string
			callerBlacklisted bool
			expectErr         bool
			getErrPart        func(accts testAccounts) string
		}{
			{
				name:              "unrelated to any blacklisted account",
				callerBlacklisted: false,
				expectErr:         false,
			},
			{
				name:              "caller is blacklisted",
				callerBlacklisted: true,
				expectErr:         true,
				getErrPart: func(accts testAccounts) string {
					return fmt.Sprintf("caller %s", accts.caller.address.Hex())
				},
			},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb := newStateDB()
				evm := newTestEvm(statedb)
				accts := buildAccounts(statedb, tc.callerBlacklisted, false)

				caller := accts.caller
				callerRef := AccountRef(caller.address)

				constructorCode := []byte{0x00}
				codeAndHash := codeAndHash{
					code: constructorCode,
				}

				_, _, _, err := evm.create(callerRef, &codeAndHash, 0, uint256.NewInt(0), common.Address{}, CREATE)
				if tc.expectErr {
					require.ErrorIs(t, err, ErrBlacklistedAccount)
					require.Contains(t, err.Error(), tc.getErrPart(accts))
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

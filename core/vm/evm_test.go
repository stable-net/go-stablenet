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

const testGas = 100000

type BlacklistRole uint8

const (
	noneRole BlacklistRole = iota
	callerRole
	targetRole
	contractRole
	beneficiaryRole
)

type account struct {
	privKey *ecdsa.PrivateKey
	address common.Address
}

func newAccount(statedb StateDB) *account {
	key, _ := crypto.GenerateKey()
	address := crypto.PubkeyToAddress(key.PublicKey)

	account := &account{key, address}
	statedb.CreateAccount(account.address)
	statedb.AddBalance(account.address, uint256.NewInt(params.Ether))

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

type callFn func(evm *EVM, caller *Contract, target common.Address) error

func invokeCall(evm *EVM, caller *Contract, target common.Address) error {
	_, _, err := evm.Call(caller, target, []byte{}, testGas, uint256.NewInt(0))
	return err
}

func invokeDelegateCall(evm *EVM, caller *Contract, target common.Address) error {
	_, _, err := evm.DelegateCall(caller, target, []byte{}, testGas)
	return err
}

func invokeCallCode(evm *EVM, caller *Contract, target common.Address) error {
	_, _, err := evm.CallCode(caller, target, []byte{}, testGas, uint256.NewInt(0))
	return err
}

func invokeStaticCall(evm *EVM, caller *Contract, target common.Address) error {
	_, _, err := evm.StaticCall(caller, target, []byte{}, testGas)
	return err
}

func TestBlacklistedAccountExecution(t *testing.T) {
	t.Run("Call", func(t *testing.T) {
		tests := []struct {
			name            string
			invoke          callFn
			blacklistedRole BlacklistRole
			expectErr       bool
		}{
			{
				name:            "Call: unrelated to any blacklisted account",
				invoke:          invokeCall,
				blacklistedRole: noneRole,
				expectErr:       false,
			},
			{
				name:            "Call: caller is blacklisted",
				invoke:          invokeCall,
				blacklistedRole: callerRole,
				expectErr:       true,
			},
			{
				name:            "Call: target is blacklisted",
				invoke:          invokeCall,
				blacklistedRole: targetRole,
				expectErr:       true,
			},
			{
				name:            "DelegateCall: unrelated to any blacklisted account",
				invoke:          invokeDelegateCall,
				blacklistedRole: noneRole,
				expectErr:       false,
			},
			{
				name:            "DelegateCall: caller is blacklisted",
				invoke:          invokeDelegateCall,
				blacklistedRole: callerRole,
				expectErr:       true,
			},
			{
				name:            "DelegateCall: target is blacklisted",
				invoke:          invokeDelegateCall,
				blacklistedRole: targetRole,
				expectErr:       true,
			},
			{
				name:            "CallCode: unrelated to any blacklisted account",
				invoke:          invokeCallCode,
				blacklistedRole: noneRole,
				expectErr:       false,
			},
			{
				name:            "CallCode: caller is blacklisted",
				invoke:          invokeCallCode,
				blacklistedRole: callerRole,
				expectErr:       true,
			},
			{
				name:            "CallCode: target is blacklisted",
				invoke:          invokeCallCode,
				blacklistedRole: targetRole,
				expectErr:       true,
			},
			{
				name:            "StaticCall: unrelated to any blacklisted account",
				invoke:          invokeStaticCall,
				blacklistedRole: noneRole,
				expectErr:       false,
			},
			{
				name:            "StaticCall: caller is blacklisted",
				invoke:          invokeStaticCall,
				blacklistedRole: callerRole,
				expectErr:       true,
			},
			{
				name:            "StaticCall: target is blacklisted",
				invoke:          invokeStaticCall,
				blacklistedRole: targetRole,
				expectErr:       true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb := newStateDB()
				evm := newTestEvm(statedb)

				testAccts := map[BlacklistRole]*account{
					callerRole: newAccount(statedb),
					targetRole: newAccount(statedb),
				}

				if tc.blacklistedRole != noneRole {
					blacklistedAcct := testAccts[tc.blacklistedRole]
					statedb.SetBlacklisted(blacklistedAcct.address)
				}

				caller := testAccts[callerRole]
				target := testAccts[targetRole]

				callerRef := NewContract(AccountRef(caller.address), AccountRef(caller.address), uint256.NewInt(0), 0)

				err := tc.invoke(evm, callerRef, target.address)
				if tc.expectErr {
					require.Error(t, err)

					var haveErr *ErrBlacklistedAccount
					require.ErrorAs(t, err, &haveErr)
					require.Equal(t, testAccts[tc.blacklistedRole].address, haveErr.Address)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name            string
			blacklistedRole BlacklistRole
			expectErr       bool
		}{
			{
				name:            "Create: unrelated to any blacklisted account",
				blacklistedRole: noneRole,
				expectErr:       false,
			},
			{
				name:            "Create: caller is blacklisted",
				blacklistedRole: callerRole,
				expectErr:       true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				statedb := newStateDB()
				evm := newTestEvm(statedb)

				testAccts := map[BlacklistRole]*account{
					callerRole: newAccount(statedb),
				}

				if tc.blacklistedRole != noneRole {
					blacklistedAcct := testAccts[tc.blacklistedRole]
					statedb.SetBlacklisted(blacklistedAcct.address)
				}

				caller := testAccts[callerRole]
				callerRef := AccountRef(caller.address)

				constructorCode := []byte{0x00}
				codeAndHash := codeAndHash{
					code: constructorCode,
				}
				_, _, _, err := evm.create(callerRef, &codeAndHash, testGas, uint256.NewInt(0), common.Address{}, CREATE)
				if tc.expectErr {
					require.Error(t, err)

					var haveErr *ErrBlacklistedAccount
					require.ErrorAs(t, err, &haveErr)
					require.Equal(t, testAccts[tc.blacklistedRole].address, haveErr.Address)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func contractCall(t *testing.T, contract *bind.BoundContract, method string, args ...interface{}) (result []interface{}) {
	require.NoError(t, contract.Call(&bind.CallOpts{}, &result, method, args...))
	return
}

func commitTx(backend *simulated.WBFTBackend, tx *types.Transaction, txErr error) (*types.Receipt, error) {
	backend.Commit()
	if txErr != nil {
		return nil, NewRevertError(txErr)
	}

	return bind.WaitMined(context.TODO(), backend.Client(), tx)
}

func ExpectedRevert(t *testing.T, err error, args ...interface{}) {
	require.Error(t, err)
	length := len(args)
	if revert, ok := err.(*RevertError); ok {
		if length > 0 {
			name, ok := args[0].(string)
			require.True(t, ok)
			require.Equal(t, name, revert.ABI.Name)
		}
		if length > 1 {
			output, ok := revert.Output.([]interface{})
			require.True(t, ok)
			for i := 1; i < length; i++ {
				arg := args[i]
				if arg != nil {
					require.Equal(t, args[i], output[i-1])
				}
			}
		}
	} else {
		if length > 0 {
			errStr, _ := strings.CutPrefix(err.Error(), vm.ErrExecutionReverted.Error()+":")
			message, ok := args[0].(string)
			require.True(t, ok)
			require.Equal(t, strings.TrimSpace(message), strings.TrimSpace(errStr))
		}
	}
}

var eoas = make(map[string]*bind.TransactOpts)

func getTxOpt(t *testing.T, alias string) *bind.TransactOpts {
	if eoa, ok := eoas[alias]; ok {
		return eoa
	} else {
		pk, err := crypto.GenerateKey()
		require.NoError(t, err)
		opts, err := bind.NewKeyedTransactorWithChainID(pk, params.TestWBFTChainConfig.ChainID)
		require.NoError(t, err)
		eoas[alias] = opts
		return opts
	}
}

type IBackend interface {
	bind.ContractBackend
	bind.DeployBackend
	SuggestGasTipCap(context.Context) (*big.Int, error)
}

func CreateDynamicTx(backend IBackend, opts *bind.TransactOpts, to *common.Address, input []byte) (*types.Transaction, error) {
	// Normalize value
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	// Estimate TipCap
	gasTipCap := opts.GasTipCap
	if gasTipCap == nil {
		tip, err := backend.SuggestGasTipCap(ensureContext(opts.Context))
		if err != nil {
			return nil, err
		}
		gasTipCap = tip
	}
	// Estimate FeeCap
	gasFeeCap := opts.GasFeeCap
	if gasFeeCap == nil {
		gasFeeCap = new(big.Int).Add(gasTipCap, big.NewInt(1e9)) // 101gwei is recommended for maxFeeCap
	}
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		return nil, fmt.Errorf("maxFeePerGas (%v) < maxPriorityFeePerGas (%v)", gasFeeCap, gasTipCap)
	}
	// Estimate GasLimit
	gasLimit := opts.GasLimit
	if opts.GasLimit == 0 {
		var err error
		gasLimit, err = estimateGasLimit(backend, opts, to, input, nil, gasTipCap, gasFeeCap, value)
		if err != nil {
			return nil, err
		}
	}
	// create the transaction
	nonce, err := getNonce(backend, opts)
	if err != nil {
		return nil, err
	}

	baseTx := &types.DynamicFeeTx{
		To:        to,
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		Value:     value,
		Data:      input,
	}

	return opts.Signer(opts.From, types.NewTx(baseTx))
}

func TransferCoin(backend IBackend, opts *bind.TransactOpts, value *big.Int, to *common.Address) (*types.Transaction, error) {
	opts.Value = value
	defer func() { opts.Value = nil }()
	if tx, err := CreateDynamicTx(backend, opts, to, []byte{}); err != nil {
		return nil, err
	} else {
		return tx, backend.SendTransaction(ensureContext(opts.Context), tx)
	}
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func estimateGasLimit(backend interface {
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error)
}, opts *bind.TransactOpts, to *common.Address, input []byte, gasPrice, gasTipCap, gasFeeCap, value *big.Int) (uint64, error) {
	msg := ethereum.CallMsg{
		From:      opts.From,
		To:        to,
		GasPrice:  gasPrice,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Value:     value,
		Data:      input,
	}
	return backend.EstimateGas(ensureContext(opts.Context), msg)
}

func getNonce(backend interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}, opts *bind.TransactOpts) (uint64, error) {
	if opts.Nonce == nil {
		return backend.PendingNonceAt(ensureContext(opts.Context), opts.From)
	} else {
		return opts.Nonce.Uint64(), nil
	}
}

func NewTxOpts(t *testing.T, eoa *EOA) *bind.TransactOpts {
	return NewTxOptsWithValue(t, eoa, nil)
}

func NewTxOptsWithValue(t *testing.T, eoa *EOA, value *big.Int) *bind.TransactOpts {
	opts, err := bind.NewKeyedTransactorWithChainID(eoa.PrivateKey, params.TestWBFTChainConfig.ChainID)
	require.NoError(t, err)
	if value != nil && value.Cmp(new(big.Int)) > 0 {
		opts.Value = new(big.Int).Set(value)
	}
	return opts
}

func towei(x int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(x), big.NewInt(params.Ether))
}

func toWeiN(x int64, decimals uint8) *big.Int {
	mul := new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(uint64(decimals)), nil)
	return new(big.Int).Mul(big.NewInt(x), mul)
}

func ToBytes32(str string) [32]byte {
	bytes := []byte(str)
	if len(bytes) > 32 {
		bytes = bytes[:32]
	}
	var copied = [32]byte{}
	copy(copied[:], bytes)
	return copied
}

func concatBytes(parts ...[]byte) []byte {
	return bytes.Join(parts, nil)
}

func SignEIP712Hash(t *testing.T, domainSeparator, data common.Hash, signer *EOA) (sig []byte, r, s common.Hash, v uint8) {
	finalEncoded := concatBytes([]byte("\x19\x01"), domainSeparator.Bytes(), data.Bytes())
	sig, err := crypto.Sign(crypto.Keccak256Hash(finalEncoded).Bytes(), signer.PrivateKey)
	require.NoError(t, err)

	if sig[64] < 27 {
		sig[64] += 27
	}

	r = common.BytesToHash(sig[:32])
	s = common.BytesToHash(sig[32:64])
	v = sig[64]

	return
}

func CheckSignatureArgs(args ...interface{}) error {
	switch len(args) {
	case 1:
		// case: bytes signature
		if _, ok := args[0].([]byte); !ok {
			return fmt.Errorf("invalid signature: expected bytes, got %T", args[0])
		}
		return nil

	case 3:
		// case: v, r, s
		if _, ok := args[0].(uint8); !ok {
			return fmt.Errorf("invalid v: expected uint8, got %T", args[0])
		}

		isValidHash := func(v interface{}) bool {
			switch v.(type) {
			case [32]byte, common.Hash:
				return true
			default:
				return false
			}
		}

		if !isValidHash(args[1]) {
			return fmt.Errorf("invalid r: expected [32]byte or common.Hash, got %T", args[1])
		}
		if !isValidHash(args[2]) {
			return fmt.Errorf("invalid s: expected [32]byte or common.Hash, got %T", args[2])
		}
		return nil

	default:
		return fmt.Errorf("invalid signature arguments: expected signature(bytes) or v, r, s(uint8,bytes32,byte32)")
	}
}

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
	"encoding/json"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type bindBackend interface {
	bind.ContractBackend
	bind.DeployBackend
}

type bindContract struct {
	Bin []byte
	Abi abi.ABI
}

func newBindContract(contract *compiler.Contract) (*bindContract, error) {
	if contract == nil {
		return nil, errors.New("nil contracts")
	}

	if parsedAbi, err := parseABI(contract.Info.AbiDefinition); err != nil {
		return nil, err
	} else {
		code := contract.Code
		if !strings.HasPrefix(code, "0x") {
			code = "0x" + code
		}
		collectErrors(parsedAbi)
		collectEvent(parsedAbi)
		return &bindContract{Bin: hexutil.MustDecode(code), Abi: *parsedAbi}, err
	}
}

func parseABI(abiDefinition interface{}) (*abi.ABI, error) {
	s, ok := abiDefinition.(string)
	if !ok {
		if bytes, err := json.Marshal(abiDefinition); err != nil {
			return nil, err
		} else {
			s = string(bytes)
		}
	}
	if abi, err := abi.JSON(strings.NewReader(s)); err != nil {
		return nil, err
	} else {
		return &abi, nil
	}
}

func (bc *bindContract) New(backend bindBackend, address common.Address) *bind.BoundContract {
	return bind.NewBoundContract(address, bc.Abi, backend, backend, backend)
}

func (bc *bindContract) Deploy(backend bindBackend, opts *bind.TransactOpts, args ...interface{}) (common.Address, *types.Transaction, *bind.BoundContract, error) {
	return bind.DeployContract(opts, bc.Abi, bc.Bin, backend, args...)
}

type nodeInfo struct {
	name  []byte
	enode []byte
	ip    []byte
	port  *big.Int
}

type MemberInfo struct {
	Staker     common.Address `json:"staker"`
	Voter      common.Address `json:"voter"`
	Reward     common.Address `json:"reward"`
	Name       []byte         `json:"name"`
	Enode      []byte         `json:"enode"`
	Ip         []byte         `json:"ip"`
	Port       *big.Int       `json:"port"`
	LockAmount *big.Int       `json:"lockAmount"`
	Memo       []byte         `json:"memo"`
	Duration   *big.Int       `json:"duration"`
}

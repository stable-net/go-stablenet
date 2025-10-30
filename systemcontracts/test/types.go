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
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

var (
	LOCK_AMOUNT  *big.Int = towei(1500000)
	MAX_UINT_256          = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	MAX_INT_256           = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
	MAX_UINT_128          = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 128), common.Big1)
)

type bindBackend interface {
	bind.ContractBackend
	bind.DeployBackend
}

type bindContract struct {
	Bin        []byte
	RuntimeBin []byte
	Abi        abi.ABI
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

		runtimeCode := contract.RuntimeCode
		if !strings.HasPrefix(runtimeCode, "0x") {
			runtimeCode = "0x" + runtimeCode
		}

		collectErrors(parsedAbi)
		collectEvent(parsedAbi)
		return &bindContract{
			Bin:        hexutil.MustDecode(code),
			RuntimeBin: hexutil.MustDecode(runtimeCode),
			Abi:        *parsedAbi,
		}, err
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

// evnet_unpack
type allEventsType map[string]abi.Event

var (
	allEventsLock = sync.RWMutex{}
	allEvents     = allEventsType{}
)

func collectEvent(abi *abi.ABI) error {
	for n, e := range abi.Events {
		func() {
			allEventsLock.Lock()
			defer allEventsLock.Unlock()
			if exist, ok := allEvents[n]; ok {
				if exist.String() == e.String() {
					goto skip
				}
			}
			allEvents[n] = e
		skip:
		}()
	}

	return nil
}

// error_unpack
type allErrorsType map[[4]byte]abi.Error

var (
	allErrorsLock = sync.RWMutex{}
	allErrors     = allErrorsType{}
)

func collectErrors(abi *abi.ABI) error {
	for _, e := range abi.Errors {
		func() {
			allErrorsLock.Lock()
			defer allErrorsLock.Unlock()
			sig := [4]byte{}
			copy(sig[:], e.ID[:4])
			if _, ok := allErrors[sig]; !ok {
				allErrors[sig] = e
			}
		}()
	}

	return nil
}

type RevertError struct {
	ABI    abi.Error
	Output interface{}
}

func (r *RevertError) Error() string {
	return fmt.Sprintf("%s: %s %v", vm.ErrExecutionReverted, r.ABI.Sig, r.Output)
}

// ErrorCode returns the JSON error code for a revert.
// See: https://github.com/ethereum/wiki/wiki/JSON-RPC-Error-Codes-Improvement-Proposal
func NewRevertError(err error) error {
	if revert, ok := err.(interface {
		ErrorCode() int
		ErrorData() interface{}
	}); !ok || revert.ErrorCode() != 3 {
		return err
	} else {
		if data, ok := revert.ErrorData().(string); !ok {
			return err
		} else {
			datas := hexutil.MustDecode(data)
			if revertErr, ok := UnpackError(datas); ok {
				return revertErr
			} else {
				reason, errUnpack := abi.UnpackRevert(datas)
				if errUnpack == nil {
					return fmt.Errorf("execution reverted: %v", reason)
				} else {
					return errors.New("execution reverted")
				}
			}
		}
	}
}

func UnpackError(result []byte) (error, bool) {
	sig := [4]byte{}
	copy(sig[:], result[:4])
	if errABI, ok := allErrors[sig]; !ok {
		return nil, false
	} else if output, err := errABI.Unpack(result); err != nil {
		return nil, false
	} else {
		return &RevertError{errABI, output}, true
	}
}

type EOA struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

type CA struct {
	Address common.Address
}

func NewEOA() (eoa *EOA) {
	pk, _ := crypto.GenerateKey()
	return &EOA{
		PrivateKey: pk,
		Address:    crypto.PubkeyToAddress(pk.PublicKey),
	}
}

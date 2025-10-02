// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the stable-one library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

const (
	SELECTOR_LENGTH  int = 4
	LENGTH_PER_PARAM int = 32
)

var (
	zeroAddress = common.Address{}
)

var (
	ErrNotAvailable          = errors.New("CoinManager: not available before Anzeon")
	ErrUnauthorized          = errors.New("CoinManager: caller is not authorized")
	ErrInvalidCallContext    = errors.New("CoinManager: invalid call context (CALL required)")
	ErrInvalidSelectorLength = errors.New("CoinManager: invalid method selector length")
	ErrInvalidMethod         = errors.New("CoinManager: invalid method")
	ErrInvalidInputLength    = errors.New("CoinManager: invalid input length")
)

type CoinManagerMethodSelector [SELECTOR_LENGTH]byte

// bytesToCoinManagerMethodSelector converts a byte slice to CoinManagerMethodSelector
func bytesToCoinManagerMethodSelector(b []byte) CoinManagerMethodSelector {
	var selctor CoinManagerMethodSelector
	copy(selctor[:], b)
	return selctor
}

type CoinManagerMethod interface {
	ParamCount() int
	Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) // Run runs the coin manager method
}

// Update method selectors below if method names, parameters, etc. are added or changed:
var (
	CoinManagerMintSelector     = bytesToCoinManagerMethodSelector(crypto.Keccak256([]byte("mint(address,uint256)"))[:4])
	CoinManagerBurnSelector     = bytesToCoinManagerMethodSelector(crypto.Keccak256([]byte("burn(address,uint256)"))[:4])
	CoinManagerTransferSelector = bytesToCoinManagerMethodSelector(crypto.Keccak256([]byte("transfer(address,address,uint256)"))[:4])
)

// Add, Remove, Modify, etc. CoinManager methods per hardfork as follows:
var CoinManagerMethodsAnzeon = map[CoinManagerMethodSelector]CoinManagerMethod{
	CoinManagerMintSelector:     &coinManagerMint{},
	CoinManagerBurnSelector:     &coinManagerBurn{},
	CoinManagerTransferSelector: &coinManagerTransfer{},
}

func ActiveCoinManager(rules params.Rules) *common.Address {
	if !rules.IsAnzeon {
		return nil
	}
	return &params.NativeCoinManagerAddress
}

func selectCoinManagerMethod(evm *EVM, selector CoinManagerMethodSelector) (CoinManagerMethod, bool) {
	var methods map[CoinManagerMethodSelector]CoinManagerMethod
	switch {
	// <HardforkName> is a placeholder, e.g., London, Berlin
	// case evm.chainRules.Is<HardforkName>:
	// 	methods = CoinManagerMethods<HardforkName>
	default: // Same as: case evm.chainRules.IsAnzeon:
		methods = CoinManagerMethodsAnzeon
	}
	m, ok := methods[selector]
	return m, ok
}

func (evm *EVM) checkCoinManagerCall(typ OpCode, caller ContractRef, addr common.Address) (bool, error) {
	if addr != params.NativeCoinManagerAddress {
		return false, nil
	}
	if typ != CALL {
		return false, ErrInvalidCallContext
	}
	if !evm.chainRules.IsAnzeon {
		return false, ErrNotAvailable
	}
	if caller.Address() != evm.chainConfig.Anzeon.SystemContracts.FiatToken.Address {
		return false, ErrUnauthorized
	}
	return true, nil
}

// It returns
// - the returned bytes,
// - the _remaining_ gas,
// - any error that occurred
func (evm *EVM) runNativeCoinManager(input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	if len(input) < SELECTOR_LENGTH {
		return nil, suppliedGas, ErrInvalidInputLength
	}

	method, ok := selectCoinManagerMethod(evm, bytesToCoinManagerMethodSelector(input[:SELECTOR_LENGTH]))
	if !ok {
		return nil, suppliedGas, ErrInvalidMethod
	}

	data := input[SELECTOR_LENGTH:]
	if len(data) != method.ParamCount()*LENGTH_PER_PARAM {
		return nil, suppliedGas, ErrInvalidInputLength
	}

	return method.Run(evm, data, suppliedGas)
}

// coinManagerMint implemented as a native contract method.
type coinManagerMint struct{}

func (c *coinManagerMint) ParamCount() int { return 2 }
func (c *coinManagerMint) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	to := common.BytesToAddress(data[0:32])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[32:64]))

	gasCost := params.UpdateBalanceGas
	if !evm.StateDB.Exist(to) {
		gasCost += params.CallNewAccountGas
	}

	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= gasCost

	// Temporarily increase 0x0 balance to enable mint transfer to receiver
	evm.StateDB.AddBalance(zeroAddress, amount)

	// Note: evm.depth is not incremented for coin_manager calls because
	// it is already increased when the coin_manager itself is invoked.
	// No further increment is needed here.
	return evm.Call(AccountRef(zeroAddress), to, nil, suppliedGas, amount)
}

// coinManagerBurn implemented as a native contract method.
type coinManagerBurn struct{}

func (c *coinManagerBurn) ParamCount() int { return 2 }
func (c *coinManagerBurn) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	from := common.BytesToAddress(data[0:32])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[32:64]))

	if !evm.Context.CanTransfer(evm.StateDB, from, amount) {
		return nil, suppliedGas, ErrInsufficientBalance
	}

	if suppliedGas < params.UpdateBalanceGas {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= params.UpdateBalanceGas

	evm.StateDB.SubBalance(from, amount)
	evm.AddTransferLog(from, zeroAddress, amount)

	return nil, suppliedGas, nil
}

// coinManagerTransfer implemented as a native contract method.
type coinManagerTransfer struct{}

func (c *coinManagerTransfer) ParamCount() int { return 3 }
func (c *coinManagerTransfer) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	from := common.BytesToAddress(data[0:32])
	to := common.BytesToAddress(data[32:64])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[64:96]))

	gasCost := 2 * params.UpdateBalanceGas // addBalance, subBalance
	if !evm.StateDB.Exist(to) {
		gasCost += params.CallNewAccountGas
	}

	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= gasCost

	// Note: evm.depth is not incremented for coin_manager calls because
	// it is already increased when the coin_manager itself is invoked.
	// No further increment is needed here.
	return evm.Call(AccountRef(from), to, nil, suppliedGas, amount)
}

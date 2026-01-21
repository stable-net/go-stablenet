// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-stablenet library.
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
	"fmt"
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

// ============================================================================
// Error definitions for native manager
// ============================================================================
var (
	ErrUnauthorized          = errors.New("native manager: caller is not authorized")
	ErrInvalidSelectorLength = errors.New("native manager: invalid method selector length")
	ErrInvalidMethod         = errors.New("native manager: invalid method")
	ErrInvalidInputLength    = errors.New("native manager: invalid input length")
)

// ErrInvalidCallContext is returned when the call context (opcode) does not match the required
type ErrInvalidCallContext struct {
	callContext OpCode
	required    OpCode
}

func (e *ErrInvalidCallContext) Error() string {
	return fmt.Sprintf("native manager: invalid call context, required %s but got %s", e.required, e.callContext)
}

// ============================================================================
// native manager
// ============================================================================

type NativeManagerContract map[NativeManagerMethodSelector]NativeManagerMethod

type NativeManagerMethodSelector [SELECTOR_LENGTH]byte

// NativeManagerMethod defines the interface for native manager contract methods.
type NativeManagerMethod interface {
	ParamCount() int                                                        // returns the number of expected parameters for the method.
	CanRun(evm *EVM, op OpCode, caller ContractRef) error                   // checks whether the method can be executed in the current context.
	Run(evm *EVM, input []byte, suppliedGas uint64) ([]byte, uint64, error) // executes the native manager method
}

// bytesToManagerMethodSelector converts a byte slice to ManagerMethodSelector
func bytesToManagerMethodSelector(b []byte) NativeManagerMethodSelector {
	var selector NativeManagerMethodSelector
	copy(selector[:], b[:SELECTOR_LENGTH])
	return selector
}

// Update method selectors below if method names, parameters, etc. are added or changed:
var (
	CoinManagerMintSelector     = bytesToManagerMethodSelector(crypto.Keccak256([]byte("mint(address,uint256)")))
	CoinManagerBurnSelector     = bytesToManagerMethodSelector(crypto.Keccak256([]byte("burn(address,uint256)")))
	CoinManagerTransferSelector = bytesToManagerMethodSelector(crypto.Keccak256([]byte("transfer(address,address,uint256)")))

	AccountManagerBlacklistSelector     = bytesToManagerMethodSelector(crypto.Keccak256([]byte("blacklist(address)")))
	AccountManagerUnBlacklistSelector   = bytesToManagerMethodSelector(crypto.Keccak256([]byte("unBlacklist(address)")))
	AccountManagerIsBlacklistedSelector = bytesToManagerMethodSelector(crypto.Keccak256([]byte("isBlacklisted(address)")))

	AccountManagerAuthorizeSelector    = bytesToManagerMethodSelector(crypto.Keccak256([]byte("authorize(address)")))
	AccountManagerUnAuthorizeSelector  = bytesToManagerMethodSelector(crypto.Keccak256([]byte("unAuthorize(address)")))
	AccountManagerIsAuthorizedSelector = bytesToManagerMethodSelector(crypto.Keccak256([]byte("isAuthorized(address)")))
)

// Add, Remove, Modify, etc. Managers per hardfork as follows:
var NativeManagerContractsAnzeon = map[common.Address]NativeManagerContract{
	params.NativeCoinManagerAddress: {
		CoinManagerMintSelector:     &coinManagerMint{},
		CoinManagerBurnSelector:     &coinManagerBurn{},
		CoinManagerTransferSelector: &coinManagerTransfer{},
	},
	params.AccountManagerAddress: {
		AccountManagerBlacklistSelector:     &accountManagerBlacklist{},
		AccountManagerUnBlacklistSelector:   &accountManagerUnBlacklist{},
		AccountManagerIsBlacklistedSelector: &accountManagerIsBlacklisted{},
		AccountManagerAuthorizeSelector:     &accountManagerAuthorize{},
		AccountManagerUnAuthorizeSelector:   &accountManagerUnAuthorize{},
		AccountManagerIsAuthorizedSelector:  &accountManagerIsAuthorized{},
	},
}

var (
	NativeManagerAddressesAnzeon []common.Address
)

func init() {
	for k := range NativeManagerContractsAnzeon {
		NativeManagerAddressesAnzeon = append(NativeManagerAddressesAnzeon, k)
	}
}

// ActiveNativeManagers returns the native manager enabled with the current configuration.
func ActiveNativeManagers(rules params.Rules) []common.Address {
	switch {
	case rules.IsAnzeon:
		return NativeManagerAddressesAnzeon
	default:
		return nil
	}
}

// It returns
// - the returned bytes,
// - the _remaining_ gas,
// - any error that occurred
func (evm *EVM) runNativeManager(m NativeManagerContract, op OpCode, caller ContractRef, input []byte, suppliedGas uint64) (ret []byte, remainingGas uint64, err error) {
	if len(input) < SELECTOR_LENGTH {
		return nil, suppliedGas, ErrInvalidSelectorLength
	}

	method, ok := m[bytesToManagerMethodSelector(input)]
	if !ok {
		return nil, suppliedGas, ErrInvalidMethod
	}

	if err := method.CanRun(evm, op, caller); err != nil {
		return nil, suppliedGas, err
	}

	if len(input)-SELECTOR_LENGTH != method.ParamCount()*LENGTH_PER_PARAM {
		return nil, suppliedGas, ErrInvalidInputLength
	}

	return method.Run(evm, input, suppliedGas)
}

// ============================================================================
// Coin Manager helper functions
// ============================================================================

// nativeTransfer performs a direct balance transfer without invoking receive()/fallback().
// It handles account creation if needed and emits a Transfer log.
// Returns remaining gas and any error.
func nativeTransfer(evm *EVM, from, to common.Address, amount *uint256.Int, suppliedGas uint64) (uint64, error) {
	if !amount.IsZero() {
		gasCost := params.UpdateBalanceGas
		transfer := from != zeroAddress
		if transfer {
			gasCost += params.UpdateBalanceGas // subBalance for transfer
		}
		if !evm.StateDB.Exist(to) {
			gasCost += params.CallNewAccountGas
		}

		if suppliedGas < gasCost {
			return 0, ErrOutOfGas
		}
		suppliedGas -= gasCost

		if transfer {
			evm.StateDB.SubBalance(from, amount)
		}
		evm.StateDB.AddBalance(to, amount)
	}
	evm.AddTransferLog(from, to, amount)

	return suppliedGas, nil
}

// ============================================================================
// Coin Manager methods
// ============================================================================

// coinManagerMint implemented as a native contract method.
type coinManagerMint struct{}

func (c *coinManagerMint) ParamCount() int { return 2 }
func (c *coinManagerMint) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunCoinManger(evm, op, caller)
}
func (c *coinManagerMint) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	to := common.BytesToAddress(data[0:32])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[32:64]))

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, zeroAddress, to, input, suppliedGas, amount.ToBig())
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(nil, startGas-retGas, err)
		}(suppliedGas)
	}

	retGas, err = nativeTransfer(evm, zeroAddress, to, amount, suppliedGas)
	return nil, retGas, err
}

// coinManagerBurn implemented as a native contract method.
type coinManagerBurn struct{}

func (c *coinManagerBurn) ParamCount() int { return 2 }
func (c *coinManagerBurn) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunCoinManger(evm, op, caller)
}
func (c *coinManagerBurn) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	from := common.BytesToAddress(data[0:32])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[32:64]))

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, from, zeroAddress, input, suppliedGas, amount.ToBig())
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(nil, startGas-retGas, err)
		}(suppliedGas)
	}

	if !evm.Context.CanTransfer(evm.StateDB, from, amount) {
		return nil, suppliedGas, ErrInsufficientBalance
	}

	gasCost := params.UpdateBalanceGas
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	evm.StateDB.SubBalance(from, amount)
	evm.AddTransferLog(from, zeroAddress, amount)

	return nil, suppliedGas - gasCost, nil
}

// coinManagerTransfer implemented as a native contract method.
type coinManagerTransfer struct{}

func (c *coinManagerTransfer) ParamCount() int { return 3 }
func (c *coinManagerTransfer) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunCoinManger(evm, op, caller)
}
func (c *coinManagerTransfer) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	from := common.BytesToAddress(data[0:32])
	to := common.BytesToAddress(data[32:64])
	amount := uint256.MustFromBig(new(big.Int).SetBytes(data[64:96]))

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, from, to, input, suppliedGas, amount.ToBig())
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(nil, startGas-retGas, err)
		}(suppliedGas)
	}

	// Validate balance before transfer
	if !evm.Context.CanTransfer(evm.StateDB, from, amount) {
		return nil, suppliedGas, ErrInsufficientBalance
	}

	retGas, err = nativeTransfer(evm, from, to, amount, suppliedGas)
	return nil, retGas, err
}

// ============================================================================
// Account Manager methods - Blacklist
// ============================================================================

// accountManagerBlacklist implemented as a native contract method.
type accountManagerBlacklist struct{}

func (c *accountManagerBlacklist) ParamCount() int { return 1 }
func (c *accountManagerBlacklist) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerBlacklist) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	gasCost := params.UpdateAccountExtraGas
	if !evm.StateDB.Exist(address) {
		gasCost += params.CallNewAccountGas
	}

	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	evm.StateDB.SetBlacklisted(address)

	return nil, suppliedGas - gasCost, nil
}

// accountManagerUnBlacklist implemented as a native contract method.
type accountManagerUnBlacklist struct{}

func (c *accountManagerUnBlacklist) ParamCount() int { return 1 }
func (c *accountManagerUnBlacklist) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerUnBlacklist) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	gasCost := params.UpdateAccountExtraGas
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	evm.StateDB.ClearBlacklisted(address)

	return nil, suppliedGas - gasCost, nil
}

// accountManagerIsBlacklisted implemented as a native contract method.
type accountManagerIsBlacklisted struct{}

func (c *accountManagerIsBlacklisted) ParamCount() int { return 1 }
func (c *accountManagerIsBlacklisted) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return nil
}
func (c *accountManagerIsBlacklisted) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(STATICCALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	ret = make([]byte, 32)
	if evm.StateDB.IsBlacklisted(address) {
		ret[31] = 1
	}
	return ret, suppliedGas, nil
}

// ============================================================================
// Account Manager methods - Authorized Account
// ============================================================================

// accountManagerAuthorize implemented as a native contract method.
type accountManagerAuthorize struct{}

func (c *accountManagerAuthorize) ParamCount() int { return 1 }
func (c *accountManagerAuthorize) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerAuthorize) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	gasCost := params.UpdateAccountExtraGas
	if !evm.StateDB.Exist(address) {
		gasCost += params.CallNewAccountGas
	}

	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	evm.StateDB.SetAuthorized(address)

	return nil, suppliedGas - gasCost, nil
}

// accountManagerUnAuthorize implemented as a native contract method.
type accountManagerUnAuthorize struct{}

func (c *accountManagerUnAuthorize) ParamCount() int { return 1 }
func (c *accountManagerUnAuthorize) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerUnAuthorize) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(CALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	gasCost := params.UpdateAccountExtraGas
	if suppliedGas < gasCost {
		return nil, 0, ErrOutOfGas
	}

	evm.StateDB.ClearAuthorized(address)

	return nil, suppliedGas - gasCost, nil
}

// accountManagerIsAuthorized implemented as a native contract method.
type accountManagerIsAuthorized struct{}

func (c *accountManagerIsAuthorized) ParamCount() int { return 1 }
func (c *accountManagerIsAuthorized) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return nil
}
func (c *accountManagerIsAuthorized) Run(evm *EVM, input []byte, suppliedGas uint64) (ret []byte, retGas uint64, err error) {
	data := input[SELECTOR_LENGTH:]
	address := common.BytesToAddress(data[0:32])

	if evm.Config.Tracer != nil {
		evm.Config.Tracer.CaptureEnter(STATICCALL, params.AccountManagerAddress, address, input, suppliedGas, nil)
		defer func(startGas uint64) {
			evm.Config.Tracer.CaptureExit(ret, startGas-retGas, err)
		}(suppliedGas)
	}

	ret = make([]byte, 32)
	if evm.StateDB.IsAuthorized(address) {
		ret[31] = 1
	}
	return ret, suppliedGas, nil
}

// ============================================================================
// validation functions
// ============================================================================

func canRunCoinManger(evm *EVM, op OpCode, caller ContractRef) error {
	if err := validateCallContext(CALL, op); err != nil {
		return err
	}
	return validateCaller(evm.chainConfig.Anzeon.SystemContracts.NativeCoinAdapter.Address, caller.Address())
}

func canRunAccountManager(evm *EVM, op OpCode, caller ContractRef) error {
	if err := validateCallContext(CALL, op); err != nil {
		return err
	}
	return validateCaller(evm.chainConfig.Anzeon.SystemContracts.GovCouncil.Address, caller.Address())
}

func validateCaller(required, caller common.Address) error {
	if caller != required {
		return ErrUnauthorized
	}
	return nil
}

func validateCallContext(required, callContext OpCode) error {
	if callContext != required {
		return &ErrInvalidCallContext{required: required, callContext: callContext}
	}
	return nil
}

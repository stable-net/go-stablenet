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
	ParamCount() int                                                       // returns the number of expected parameters for the method.
	CanRun(evm *EVM, op OpCode, caller ContractRef) error                  // checks whether the method can be executed in the current context.
	Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) // executes the native manager method
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
	AccountManagerUnblacklistSelector   = bytesToManagerMethodSelector(crypto.Keccak256([]byte("unblacklist(address)")))
	AccountManagerIsBlacklistedSelector = bytesToManagerMethodSelector(crypto.Keccak256([]byte("isBlacklist(address)")))

	AccountManagerAuthorizeSelector    = bytesToManagerMethodSelector(crypto.Keccak256([]byte("authorize(address)")))
	AccountManagerUnauthorizeSelector  = bytesToManagerMethodSelector(crypto.Keccak256([]byte("unauthorize(address)")))
	AccountManagerIsAuthorizedSelector = bytesToManagerMethodSelector(crypto.Keccak256([]byte("isAuthorized(address)")))
)

// Add, Remove, Modify, etc. Managers per hardfork as follows:
var NativeManagerContractsAnzeon = map[common.Address]NativeManagerContract{
	params.NativeCoinManagerAddress: NativeManagerContract{
		CoinManagerMintSelector:     &coinManagerMint{},
		CoinManagerBurnSelector:     &coinManagerBurn{},
		CoinManagerTransferSelector: &coinManagerTransfer{},
	},
	params.AccountManagerAddress: NativeManagerContract{
		AccountManagerBlacklistSelector:     &accountManagerBlacklist{},
		AccountManagerUnblacklistSelector:   &accountManagerUnblacklist{},
		AccountManagerIsBlacklistedSelector: &accountManagerIsBlacklisted{},
		AccountManagerAuthorizeSelector:     &accountManagerAuthorize{},
		AccountManagerUnauthorizeSelector:   &accountManagerUnauthorize{},
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

	data := input[SELECTOR_LENGTH:]
	if len(data) != method.ParamCount()*LENGTH_PER_PARAM {
		return nil, suppliedGas, ErrInvalidInputLength
	}

	return method.Run(evm, data, suppliedGas)
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
	//
	// evm.Call is used so that when transferring to a contract address,
	// its receive() or fallback() function is properly invoked instead of
	// just updating balances directly.
	return evm.Call(AccountRef(zeroAddress), to, nil, suppliedGas, amount)
}

// coinManagerBurn implemented as a native contract method.
type coinManagerBurn struct{}

func (c *coinManagerBurn) ParamCount() int { return 2 }
func (c *coinManagerBurn) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunCoinManger(evm, op, caller)
}
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
func (c *coinManagerTransfer) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunCoinManger(evm, op, caller)
}
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
	//
	// evm.Call is used so that when transferring to a contract address,
	// its receive() or fallback() function is properly invoked instead of
	// just updating balances directly.
	ret, leftOverGas, err := evm.Call(AccountRef(from), to, nil, suppliedGas, amount)

	// Transfer event emission for 0-value transfer (ERC20)
	if err == nil && amount.IsZero() {
		evm.AddTransferLog(from, to, amount)
	}
	return ret, leftOverGas, err
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
func (c *accountManagerBlacklist) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	if suppliedGas < params.UpdateBalanceGas {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= params.UpdateBalanceGas

	evm.StateDB.SetBlacklisted(address)

	return nil, suppliedGas, nil
}

// accountManagerUnBlacklist implemented as a native contract method.
type accountManagerUnblacklist struct{}

func (c *accountManagerUnblacklist) ParamCount() int { return 1 }
func (c *accountManagerUnblacklist) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerUnblacklist) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	if suppliedGas < params.UpdateBalanceGas {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= params.UpdateBalanceGas

	evm.StateDB.ClearBlacklisted(address)

	return nil, suppliedGas, nil
}

// accountManagerIsBlacklisted implemented as a native contract method.
type accountManagerIsBlacklisted struct{}

func (c *accountManagerIsBlacklisted) ParamCount() int { return 1 }
func (c *accountManagerIsBlacklisted) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return nil
}
func (c *accountManagerIsBlacklisted) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	evm.StateDB.IsBlacklisted(address)

	return nil, suppliedGas, nil
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
func (c *accountManagerAuthorize) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	if suppliedGas < params.UpdateBalanceGas {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= params.UpdateBalanceGas

	evm.StateDB.SetAuthorized(address)

	return nil, suppliedGas, nil
}

// accountManagerUnauthorize implemented as a native contract method.
type accountManagerUnauthorize struct{}

func (c *accountManagerUnauthorize) ParamCount() int { return 1 }
func (c *accountManagerUnauthorize) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return canRunAccountManager(evm, op, caller)
}
func (c *accountManagerUnauthorize) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	if suppliedGas < params.UpdateBalanceGas {
		return nil, 0, ErrOutOfGas
	}
	suppliedGas -= params.UpdateBalanceGas

	evm.StateDB.ClearAuthorized(address)

	return nil, suppliedGas, nil
}

// accountManagerIsAuthorized implemented as a native contract method.
type accountManagerIsAuthorized struct{}

func (c *accountManagerIsAuthorized) ParamCount() int { return 1 }
func (c *accountManagerIsAuthorized) CanRun(evm *EVM, op OpCode, caller ContractRef) error {
	return nil
}
func (c *accountManagerIsAuthorized) Run(evm *EVM, data []byte, suppliedGas uint64) ([]byte, uint64, error) {
	address := common.BytesToAddress(data[0:32])

	evm.StateDB.IsAuthorized(address)

	return nil, suppliedGas, nil
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
	return validateCaller(evm.chainConfig.Anzeon.SystemContracts.NativeCoinAdapter.Address, caller.Address())
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

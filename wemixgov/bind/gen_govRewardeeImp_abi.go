// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package gov

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// GovRewardeeImpMetaData contains all meta data concerning the GovRewardeeImp contract.
var GovRewardeeImpMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FeePaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RewardPaid\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"GOV_STAKING\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"sendFeeTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"sendRewardTo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Sigs: map[string]string{
		"e3966bc7": "GOV_STAKING()",
		"8530869a": "sendFeeTo(address,uint256)",
		"a2cc9ab6": "sendRewardTo(address,uint256)",
	},
	Bin: "0x608060405234801561001057600080fd5b506107ea806100206000396000f3fe6080604052600436106100385760003560e01c80638530869a14610044578063a2cc9ab614610066578063e3966bc71461008657600080fd5b3661003f57005b600080fd5b34801561005057600080fd5b5061006461005f366004610641565b6100b8565b005b34801561007257600080fd5b50610064610081366004610641565b6104a0565b34801561009257600080fd5b5061009c61100181565b6040516001600160a01b03909116815260200160405180910390f35b33611001146100e25760405162461bcd60e51b81526004016100d990610679565b60405180910390fd5b6001600160a01b0382166101085760405162461bcd60e51b81526004016100d9906106cb565b600081116101585760405162461bcd60e51b815260206004820152601b60248201527f476f7652657761726465653a20616d6f756e74206973207a65726f000000000060448201526064016100d9565b478111156101785760405162461bcd60e51b81526004016100d990610715565b813b80156103e3576040516301ffc9a760e01b815263f4f2499160e01b60048201526001600160a01b038416906301ffc9a790602401602060405180830381865afa9250505080156101e7575060408051601f3d908101601f191682019092526101e491810190610756565b60015b610264576000836001600160a01b03168360405160006040518083038185875af1925050503d8060008114610238576040519150601f19603f3d011682016040523d82523d6000602084013e61023d565b606091505b505090508061025e5760405162461bcd60e51b81526004016100d99061077f565b50610458565b801561036957604051630969acb360e01b8152600481018490526001600160a01b03851690630969acb39085906024016000604051808303818588803b1580156102ad57600080fd5b505af1935050505080156102bf575060015b6103205760405162461bcd60e51b815260206004820152602c60248201527f476f7652657761726465653a2066656520726563697069656e7420636f6e747260448201526b1858dd081c995d995c9d195960a21b60648201526084016100d9565b836001600160a01b03167f075a2720282fdf622141dae0b048ef90a21a7e57c134c76912d19d006b3b3f6f8460405161035b91815260200190565b60405180910390a250505050565b6000846001600160a01b03168460405160006040518083038185875af1925050503d80600081146103b6576040519150601f19603f3d011682016040523d82523d6000602084013e6103bb565b606091505b50509050806103dc5760405162461bcd60e51b81526004016100d99061077f565b5050610458565b6000836001600160a01b03168360405160006040518083038185875af1925050503d8060008114610430576040519150601f19603f3d011682016040523d82523d6000602084013e610435565b606091505b50509050806104565760405162461bcd60e51b81526004016100d99061077f565b505b826001600160a01b03167f075a2720282fdf622141dae0b048ef90a21a7e57c134c76912d19d006b3b3f6f8360405161049391815260200190565b60405180910390a2505050565b33611001146104c15760405162461bcd60e51b81526004016100d990610679565b6001600160a01b0382166104e75760405162461bcd60e51b81526004016100d9906106cb565b600081116105375760405162461bcd60e51b815260206004820152601b60248201527f476f7652657761726465653a20616d6f756e74206973207a65726f000000000060448201526064016100d9565b478111156105575760405162461bcd60e51b81526004016100d990610715565b6000826001600160a01b03168260405160006040518083038185875af1925050503d80600081146105a4576040519150601f19603f3d011682016040523d82523d6000602084013e6105a9565b606091505b50509050806106065760405162461bcd60e51b815260206004820152602360248201527f476f7652657761726465653a20726577617264207472616e73666572206661696044820152621b195960ea1b60648201526084016100d9565b826001600160a01b03167fe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e04868360405161049391815260200190565b6000806040838503121561065457600080fd5b82356001600160a01b038116811461066b57600080fd5b946020939093013593505050565b60208082526032908201527f476f7652657761726465653a2063616c6c6572206973206e6f742074686520476040820152711bdd94dd185ada5b99c818dbdb9d1c9858dd60721b606082015260800190565b6020808252602a908201527f476f7652657761726465653a20726563697069656e7420697320746865207a65604082015269726f206164647265737360b01b606082015260800190565b60208082526021908201527f476f7652657761726465653a20696e73756666696369656e742062616c616e636040820152606560f81b606082015260800190565b60006020828403121561076857600080fd5b8151801515811461077857600080fd5b9392505050565b6020808252818101527f476f7652657761726465653a20666565207472616e73666572206661696c656460408201526060019056fea2646970667358221220061b8473af1433193e94c5378c9e1b8dfdb1119d95c23d31121c1028b8815e5864736f6c634300080e0033",
}

// GovRewardeeImpABI is the input ABI used to generate the binding from.
// Deprecated: Use GovRewardeeImpMetaData.ABI instead.
var GovRewardeeImpABI = GovRewardeeImpMetaData.ABI

// Deprecated: Use GovRewardeeImpMetaData.Sigs instead.
// GovRewardeeImpFuncSigs maps the 4-byte function signature to its string representation.
var GovRewardeeImpFuncSigs = GovRewardeeImpMetaData.Sigs

// GovRewardeeImpBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovRewardeeImpMetaData.Bin instead.
var GovRewardeeImpBin = GovRewardeeImpMetaData.Bin

// DeployGovRewardeeImp deploys a new Ethereum contract, binding an instance of GovRewardeeImp to it.
func DeployGovRewardeeImp(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovRewardeeImp, error) {
	parsed, err := GovRewardeeImpMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovRewardeeImpBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovRewardeeImp{GovRewardeeImpCaller: GovRewardeeImpCaller{contract: contract}, GovRewardeeImpTransactor: GovRewardeeImpTransactor{contract: contract}, GovRewardeeImpFilterer: GovRewardeeImpFilterer{contract: contract}}, nil
}

// GovRewardeeImp is an auto generated Go binding around an Ethereum contract.
type GovRewardeeImp struct {
	GovRewardeeImpCaller     // Read-only binding to the contract
	GovRewardeeImpTransactor // Write-only binding to the contract
	GovRewardeeImpFilterer   // Log filterer for contract events
}

// GovRewardeeImpCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovRewardeeImpCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovRewardeeImpTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovRewardeeImpTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovRewardeeImpFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovRewardeeImpFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovRewardeeImpSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovRewardeeImpSession struct {
	Contract     *GovRewardeeImp   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovRewardeeImpCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovRewardeeImpCallerSession struct {
	Contract *GovRewardeeImpCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// GovRewardeeImpTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovRewardeeImpTransactorSession struct {
	Contract     *GovRewardeeImpTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// GovRewardeeImpRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovRewardeeImpRaw struct {
	Contract *GovRewardeeImp // Generic contract binding to access the raw methods on
}

// GovRewardeeImpCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovRewardeeImpCallerRaw struct {
	Contract *GovRewardeeImpCaller // Generic read-only contract binding to access the raw methods on
}

// GovRewardeeImpTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovRewardeeImpTransactorRaw struct {
	Contract *GovRewardeeImpTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovRewardeeImp creates a new instance of GovRewardeeImp, bound to a specific deployed contract.
func NewGovRewardeeImp(address common.Address, backend bind.ContractBackend) (*GovRewardeeImp, error) {
	contract, err := bindGovRewardeeImp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImp{GovRewardeeImpCaller: GovRewardeeImpCaller{contract: contract}, GovRewardeeImpTransactor: GovRewardeeImpTransactor{contract: contract}, GovRewardeeImpFilterer: GovRewardeeImpFilterer{contract: contract}}, nil
}

// NewGovRewardeeImpCaller creates a new read-only instance of GovRewardeeImp, bound to a specific deployed contract.
func NewGovRewardeeImpCaller(address common.Address, caller bind.ContractCaller) (*GovRewardeeImpCaller, error) {
	contract, err := bindGovRewardeeImp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImpCaller{contract: contract}, nil
}

// NewGovRewardeeImpTransactor creates a new write-only instance of GovRewardeeImp, bound to a specific deployed contract.
func NewGovRewardeeImpTransactor(address common.Address, transactor bind.ContractTransactor) (*GovRewardeeImpTransactor, error) {
	contract, err := bindGovRewardeeImp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImpTransactor{contract: contract}, nil
}

// NewGovRewardeeImpFilterer creates a new log filterer instance of GovRewardeeImp, bound to a specific deployed contract.
func NewGovRewardeeImpFilterer(address common.Address, filterer bind.ContractFilterer) (*GovRewardeeImpFilterer, error) {
	contract, err := bindGovRewardeeImp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImpFilterer{contract: contract}, nil
}

// bindGovRewardeeImp binds a generic wrapper to an already deployed contract.
func bindGovRewardeeImp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovRewardeeImpMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovRewardeeImp *GovRewardeeImpRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovRewardeeImp.Contract.GovRewardeeImpCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovRewardeeImp *GovRewardeeImpRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.GovRewardeeImpTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovRewardeeImp *GovRewardeeImpRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.GovRewardeeImpTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovRewardeeImp *GovRewardeeImpCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovRewardeeImp.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovRewardeeImp *GovRewardeeImpTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovRewardeeImp *GovRewardeeImpTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.contract.Transact(opts, method, params...)
}

// GOVSTAKING is a free data retrieval call binding the contract method 0xe3966bc7.
//
// Solidity: function GOV_STAKING() view returns(address)
func (_GovRewardeeImp *GovRewardeeImpCaller) GOVSTAKING(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GovRewardeeImp.contract.Call(opts, &out, "GOV_STAKING")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GOVSTAKING is a free data retrieval call binding the contract method 0xe3966bc7.
//
// Solidity: function GOV_STAKING() view returns(address)
func (_GovRewardeeImp *GovRewardeeImpSession) GOVSTAKING() (common.Address, error) {
	return _GovRewardeeImp.Contract.GOVSTAKING(&_GovRewardeeImp.CallOpts)
}

// GOVSTAKING is a free data retrieval call binding the contract method 0xe3966bc7.
//
// Solidity: function GOV_STAKING() view returns(address)
func (_GovRewardeeImp *GovRewardeeImpCallerSession) GOVSTAKING() (common.Address, error) {
	return _GovRewardeeImp.Contract.GOVSTAKING(&_GovRewardeeImp.CallOpts)
}

// SendFeeTo is a paid mutator transaction binding the contract method 0x8530869a.
//
// Solidity: function sendFeeTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpTransactor) SendFeeTo(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.contract.Transact(opts, "sendFeeTo", recipient, amount)
}

// SendFeeTo is a paid mutator transaction binding the contract method 0x8530869a.
//
// Solidity: function sendFeeTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpSession) SendFeeTo(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.SendFeeTo(&_GovRewardeeImp.TransactOpts, recipient, amount)
}

// SendFeeTo is a paid mutator transaction binding the contract method 0x8530869a.
//
// Solidity: function sendFeeTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpTransactorSession) SendFeeTo(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.SendFeeTo(&_GovRewardeeImp.TransactOpts, recipient, amount)
}

// SendRewardTo is a paid mutator transaction binding the contract method 0xa2cc9ab6.
//
// Solidity: function sendRewardTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpTransactor) SendRewardTo(opts *bind.TransactOpts, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.contract.Transact(opts, "sendRewardTo", recipient, amount)
}

// SendRewardTo is a paid mutator transaction binding the contract method 0xa2cc9ab6.
//
// Solidity: function sendRewardTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpSession) SendRewardTo(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.SendRewardTo(&_GovRewardeeImp.TransactOpts, recipient, amount)
}

// SendRewardTo is a paid mutator transaction binding the contract method 0xa2cc9ab6.
//
// Solidity: function sendRewardTo(address recipient, uint256 amount) returns()
func (_GovRewardeeImp *GovRewardeeImpTransactorSession) SendRewardTo(recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.SendRewardTo(&_GovRewardeeImp.TransactOpts, recipient, amount)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GovRewardeeImp *GovRewardeeImpTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovRewardeeImp.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GovRewardeeImp *GovRewardeeImpSession) Receive() (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.Receive(&_GovRewardeeImp.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_GovRewardeeImp *GovRewardeeImpTransactorSession) Receive() (*types.Transaction, error) {
	return _GovRewardeeImp.Contract.Receive(&_GovRewardeeImp.TransactOpts)
}

// GovRewardeeImpFeePaidIterator is returned from FilterFeePaid and is used to iterate over the raw logs and unpacked data for FeePaid events raised by the GovRewardeeImp contract.
type GovRewardeeImpFeePaidIterator struct {
	Event *GovRewardeeImpFeePaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *GovRewardeeImpFeePaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovRewardeeImpFeePaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(GovRewardeeImpFeePaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *GovRewardeeImpFeePaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovRewardeeImpFeePaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovRewardeeImpFeePaid represents a FeePaid event raised by the GovRewardeeImp contract.
type GovRewardeeImpFeePaid struct {
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterFeePaid is a free log retrieval operation binding the contract event 0x075a2720282fdf622141dae0b048ef90a21a7e57c134c76912d19d006b3b3f6f.
//
// Solidity: event FeePaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) FilterFeePaid(opts *bind.FilterOpts, recipient []common.Address) (*GovRewardeeImpFeePaidIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _GovRewardeeImp.contract.FilterLogs(opts, "FeePaid", recipientRule)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImpFeePaidIterator{contract: _GovRewardeeImp.contract, event: "FeePaid", logs: logs, sub: sub}, nil
}

// WatchFeePaid is a free log subscription operation binding the contract event 0x075a2720282fdf622141dae0b048ef90a21a7e57c134c76912d19d006b3b3f6f.
//
// Solidity: event FeePaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) WatchFeePaid(opts *bind.WatchOpts, sink chan<- *GovRewardeeImpFeePaid, recipient []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _GovRewardeeImp.contract.WatchLogs(opts, "FeePaid", recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovRewardeeImpFeePaid)
				if err := _GovRewardeeImp.contract.UnpackLog(event, "FeePaid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFeePaid is a log parse operation binding the contract event 0x075a2720282fdf622141dae0b048ef90a21a7e57c134c76912d19d006b3b3f6f.
//
// Solidity: event FeePaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) ParseFeePaid(log types.Log) (*GovRewardeeImpFeePaid, error) {
	event := new(GovRewardeeImpFeePaid)
	if err := _GovRewardeeImp.contract.UnpackLog(event, "FeePaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovRewardeeImpRewardPaidIterator is returned from FilterRewardPaid and is used to iterate over the raw logs and unpacked data for RewardPaid events raised by the GovRewardeeImp contract.
type GovRewardeeImpRewardPaidIterator struct {
	Event *GovRewardeeImpRewardPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *GovRewardeeImpRewardPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovRewardeeImpRewardPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(GovRewardeeImpRewardPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *GovRewardeeImpRewardPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovRewardeeImpRewardPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovRewardeeImpRewardPaid represents a RewardPaid event raised by the GovRewardeeImp contract.
type GovRewardeeImpRewardPaid struct {
	Recipient common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRewardPaid is a free log retrieval operation binding the contract event 0xe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e0486.
//
// Solidity: event RewardPaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) FilterRewardPaid(opts *bind.FilterOpts, recipient []common.Address) (*GovRewardeeImpRewardPaidIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _GovRewardeeImp.contract.FilterLogs(opts, "RewardPaid", recipientRule)
	if err != nil {
		return nil, err
	}
	return &GovRewardeeImpRewardPaidIterator{contract: _GovRewardeeImp.contract, event: "RewardPaid", logs: logs, sub: sub}, nil
}

// WatchRewardPaid is a free log subscription operation binding the contract event 0xe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e0486.
//
// Solidity: event RewardPaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) WatchRewardPaid(opts *bind.WatchOpts, sink chan<- *GovRewardeeImpRewardPaid, recipient []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _GovRewardeeImp.contract.WatchLogs(opts, "RewardPaid", recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovRewardeeImpRewardPaid)
				if err := _GovRewardeeImp.contract.UnpackLog(event, "RewardPaid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRewardPaid is a log parse operation binding the contract event 0xe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e0486.
//
// Solidity: event RewardPaid(address indexed recipient, uint256 amount)
func (_GovRewardeeImp *GovRewardeeImpFilterer) ParseRewardPaid(log types.Log) (*GovRewardeeImpRewardPaid, error) {
	event := new(GovRewardeeImpRewardPaid)
	if err := _GovRewardeeImp.contract.UnpackLog(event, "RewardPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

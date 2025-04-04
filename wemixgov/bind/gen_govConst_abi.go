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

// GovConstMetaData contains all meta data concerning the GovConst contract.
var GovConstMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BLS_PUBLIC_KEY_LENGTH\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CHANGE_FEE_DELAY\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"FEE_PRECISION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAXIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MINIMUM_STAKING\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_STAKERS\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"REWARD_PRECISION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_DELEGATOR\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UNBONDING_PERIOD_STAKER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"8280a25a": "BLS_PUBLIC_KEY_LENGTH()",
		"af2a57bb": "CHANGE_FEE_DELAY()",
		"e63a391f": "FEE_PRECISION()",
		"129060ab": "MAXIMUM_STAKING()",
		"ba631d3f": "MINIMUM_STAKING()",
		"decf0206": "MIN_STAKERS()",
		"3d6aa5e1": "REWARD_PRECISION()",
		"840c1771": "UNBONDING_PERIOD_DELEGATOR()",
		"fde7f371": "UNBONDING_PERIOD_STAKER()",
	},
	Bin: "0x608060405234801561001057600080fd5b50610130806100206000396000f3fe6080604052348015600f57600080fd5b506004361060875760003560e01c8063af2a57bb11605e578063af2a57bb1460d2578063ba631d3f1460db578063decf02061460eb578063e63a391f1460f2578063fde7f3711460d257600080fd5b8063129060ab14608c5780633d6aa5e11460b45780638280a25a1460c2578063840c17711460c9575b600080fd5b60a26fffffffffffffffffffffffffffffffff81565b60405190815260200160405180910390f35b60a2670de0b6b3a764000081565b60a2603081565b60a26203f48081565b60a262093a8081565b60a26969e10de76676d080000081565b60a2600581565b60a26127108156fea2646970667358221220807711e629d1b15676d0e65f365dd144b684790d6d106aaa67a0f91e3ef7117064736f6c634300080e0033",
}

// GovConstABI is the input ABI used to generate the binding from.
// Deprecated: Use GovConstMetaData.ABI instead.
var GovConstABI = GovConstMetaData.ABI

// Deprecated: Use GovConstMetaData.Sigs instead.
// GovConstFuncSigs maps the 4-byte function signature to its string representation.
var GovConstFuncSigs = GovConstMetaData.Sigs

// GovConstBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovConstMetaData.Bin instead.
var GovConstBin = GovConstMetaData.Bin

// DeployGovConst deploys a new Ethereum contract, binding an instance of GovConst to it.
func DeployGovConst(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovConst, error) {
	parsed, err := GovConstMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovConstBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovConst{GovConstCaller: GovConstCaller{contract: contract}, GovConstTransactor: GovConstTransactor{contract: contract}, GovConstFilterer: GovConstFilterer{contract: contract}}, nil
}

// GovConst is an auto generated Go binding around an Ethereum contract.
type GovConst struct {
	GovConstCaller     // Read-only binding to the contract
	GovConstTransactor // Write-only binding to the contract
	GovConstFilterer   // Log filterer for contract events
}

// GovConstCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovConstCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConstTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovConstTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConstFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovConstFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConstSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovConstSession struct {
	Contract     *GovConst         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovConstCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovConstCallerSession struct {
	Contract *GovConstCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// GovConstTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovConstTransactorSession struct {
	Contract     *GovConstTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// GovConstRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovConstRaw struct {
	Contract *GovConst // Generic contract binding to access the raw methods on
}

// GovConstCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovConstCallerRaw struct {
	Contract *GovConstCaller // Generic read-only contract binding to access the raw methods on
}

// GovConstTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovConstTransactorRaw struct {
	Contract *GovConstTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovConst creates a new instance of GovConst, bound to a specific deployed contract.
func NewGovConst(address common.Address, backend bind.ContractBackend) (*GovConst, error) {
	contract, err := bindGovConst(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovConst{GovConstCaller: GovConstCaller{contract: contract}, GovConstTransactor: GovConstTransactor{contract: contract}, GovConstFilterer: GovConstFilterer{contract: contract}}, nil
}

// NewGovConstCaller creates a new read-only instance of GovConst, bound to a specific deployed contract.
func NewGovConstCaller(address common.Address, caller bind.ContractCaller) (*GovConstCaller, error) {
	contract, err := bindGovConst(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovConstCaller{contract: contract}, nil
}

// NewGovConstTransactor creates a new write-only instance of GovConst, bound to a specific deployed contract.
func NewGovConstTransactor(address common.Address, transactor bind.ContractTransactor) (*GovConstTransactor, error) {
	contract, err := bindGovConst(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovConstTransactor{contract: contract}, nil
}

// NewGovConstFilterer creates a new log filterer instance of GovConst, bound to a specific deployed contract.
func NewGovConstFilterer(address common.Address, filterer bind.ContractFilterer) (*GovConstFilterer, error) {
	contract, err := bindGovConst(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovConstFilterer{contract: contract}, nil
}

// bindGovConst binds a generic wrapper to an already deployed contract.
func bindGovConst(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovConstMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovConst *GovConstRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovConst.Contract.GovConstCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovConst *GovConstRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovConst.Contract.GovConstTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovConst *GovConstRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovConst.Contract.GovConstTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovConst *GovConstCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovConst.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovConst *GovConstTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovConst.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovConst *GovConstTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovConst.Contract.contract.Transact(opts, method, params...)
}

// BLSPUBLICKEYLENGTH is a free data retrieval call binding the contract method 0x8280a25a.
//
// Solidity: function BLS_PUBLIC_KEY_LENGTH() view returns(uint256)
func (_GovConst *GovConstCaller) BLSPUBLICKEYLENGTH(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "BLS_PUBLIC_KEY_LENGTH")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BLSPUBLICKEYLENGTH is a free data retrieval call binding the contract method 0x8280a25a.
//
// Solidity: function BLS_PUBLIC_KEY_LENGTH() view returns(uint256)
func (_GovConst *GovConstSession) BLSPUBLICKEYLENGTH() (*big.Int, error) {
	return _GovConst.Contract.BLSPUBLICKEYLENGTH(&_GovConst.CallOpts)
}

// BLSPUBLICKEYLENGTH is a free data retrieval call binding the contract method 0x8280a25a.
//
// Solidity: function BLS_PUBLIC_KEY_LENGTH() view returns(uint256)
func (_GovConst *GovConstCallerSession) BLSPUBLICKEYLENGTH() (*big.Int, error) {
	return _GovConst.Contract.BLSPUBLICKEYLENGTH(&_GovConst.CallOpts)
}

// CHANGEFEEDELAY is a free data retrieval call binding the contract method 0xaf2a57bb.
//
// Solidity: function CHANGE_FEE_DELAY() view returns(uint256)
func (_GovConst *GovConstCaller) CHANGEFEEDELAY(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "CHANGE_FEE_DELAY")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHANGEFEEDELAY is a free data retrieval call binding the contract method 0xaf2a57bb.
//
// Solidity: function CHANGE_FEE_DELAY() view returns(uint256)
func (_GovConst *GovConstSession) CHANGEFEEDELAY() (*big.Int, error) {
	return _GovConst.Contract.CHANGEFEEDELAY(&_GovConst.CallOpts)
}

// CHANGEFEEDELAY is a free data retrieval call binding the contract method 0xaf2a57bb.
//
// Solidity: function CHANGE_FEE_DELAY() view returns(uint256)
func (_GovConst *GovConstCallerSession) CHANGEFEEDELAY() (*big.Int, error) {
	return _GovConst.Contract.CHANGEFEEDELAY(&_GovConst.CallOpts)
}

// FEEPRECISION is a free data retrieval call binding the contract method 0xe63a391f.
//
// Solidity: function FEE_PRECISION() view returns(uint256)
func (_GovConst *GovConstCaller) FEEPRECISION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "FEE_PRECISION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FEEPRECISION is a free data retrieval call binding the contract method 0xe63a391f.
//
// Solidity: function FEE_PRECISION() view returns(uint256)
func (_GovConst *GovConstSession) FEEPRECISION() (*big.Int, error) {
	return _GovConst.Contract.FEEPRECISION(&_GovConst.CallOpts)
}

// FEEPRECISION is a free data retrieval call binding the contract method 0xe63a391f.
//
// Solidity: function FEE_PRECISION() view returns(uint256)
func (_GovConst *GovConstCallerSession) FEEPRECISION() (*big.Int, error) {
	return _GovConst.Contract.FEEPRECISION(&_GovConst.CallOpts)
}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstCaller) MAXIMUMSTAKING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "MAXIMUM_STAKING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstSession) MAXIMUMSTAKING() (*big.Int, error) {
	return _GovConst.Contract.MAXIMUMSTAKING(&_GovConst.CallOpts)
}

// MAXIMUMSTAKING is a free data retrieval call binding the contract method 0x129060ab.
//
// Solidity: function MAXIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstCallerSession) MAXIMUMSTAKING() (*big.Int, error) {
	return _GovConst.Contract.MAXIMUMSTAKING(&_GovConst.CallOpts)
}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstCaller) MINIMUMSTAKING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "MINIMUM_STAKING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstSession) MINIMUMSTAKING() (*big.Int, error) {
	return _GovConst.Contract.MINIMUMSTAKING(&_GovConst.CallOpts)
}

// MINIMUMSTAKING is a free data retrieval call binding the contract method 0xba631d3f.
//
// Solidity: function MINIMUM_STAKING() view returns(uint256)
func (_GovConst *GovConstCallerSession) MINIMUMSTAKING() (*big.Int, error) {
	return _GovConst.Contract.MINIMUMSTAKING(&_GovConst.CallOpts)
}

// MINSTAKERS is a free data retrieval call binding the contract method 0xdecf0206.
//
// Solidity: function MIN_STAKERS() view returns(uint256)
func (_GovConst *GovConstCaller) MINSTAKERS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "MIN_STAKERS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINSTAKERS is a free data retrieval call binding the contract method 0xdecf0206.
//
// Solidity: function MIN_STAKERS() view returns(uint256)
func (_GovConst *GovConstSession) MINSTAKERS() (*big.Int, error) {
	return _GovConst.Contract.MINSTAKERS(&_GovConst.CallOpts)
}

// MINSTAKERS is a free data retrieval call binding the contract method 0xdecf0206.
//
// Solidity: function MIN_STAKERS() view returns(uint256)
func (_GovConst *GovConstCallerSession) MINSTAKERS() (*big.Int, error) {
	return _GovConst.Contract.MINSTAKERS(&_GovConst.CallOpts)
}

// REWARDPRECISION is a free data retrieval call binding the contract method 0x3d6aa5e1.
//
// Solidity: function REWARD_PRECISION() view returns(uint256)
func (_GovConst *GovConstCaller) REWARDPRECISION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "REWARD_PRECISION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// REWARDPRECISION is a free data retrieval call binding the contract method 0x3d6aa5e1.
//
// Solidity: function REWARD_PRECISION() view returns(uint256)
func (_GovConst *GovConstSession) REWARDPRECISION() (*big.Int, error) {
	return _GovConst.Contract.REWARDPRECISION(&_GovConst.CallOpts)
}

// REWARDPRECISION is a free data retrieval call binding the contract method 0x3d6aa5e1.
//
// Solidity: function REWARD_PRECISION() view returns(uint256)
func (_GovConst *GovConstCallerSession) REWARDPRECISION() (*big.Int, error) {
	return _GovConst.Contract.REWARDPRECISION(&_GovConst.CallOpts)
}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovConst *GovConstCaller) UNBONDINGPERIODDELEGATOR(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "UNBONDING_PERIOD_DELEGATOR")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovConst *GovConstSession) UNBONDINGPERIODDELEGATOR() (*big.Int, error) {
	return _GovConst.Contract.UNBONDINGPERIODDELEGATOR(&_GovConst.CallOpts)
}

// UNBONDINGPERIODDELEGATOR is a free data retrieval call binding the contract method 0x840c1771.
//
// Solidity: function UNBONDING_PERIOD_DELEGATOR() view returns(uint256)
func (_GovConst *GovConstCallerSession) UNBONDINGPERIODDELEGATOR() (*big.Int, error) {
	return _GovConst.Contract.UNBONDINGPERIODDELEGATOR(&_GovConst.CallOpts)
}

// UNBONDINGPERIODSTAKER is a free data retrieval call binding the contract method 0xfde7f371.
//
// Solidity: function UNBONDING_PERIOD_STAKER() view returns(uint256)
func (_GovConst *GovConstCaller) UNBONDINGPERIODSTAKER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConst.contract.Call(opts, &out, "UNBONDING_PERIOD_STAKER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UNBONDINGPERIODSTAKER is a free data retrieval call binding the contract method 0xfde7f371.
//
// Solidity: function UNBONDING_PERIOD_STAKER() view returns(uint256)
func (_GovConst *GovConstSession) UNBONDINGPERIODSTAKER() (*big.Int, error) {
	return _GovConst.Contract.UNBONDINGPERIODSTAKER(&_GovConst.CallOpts)
}

// UNBONDINGPERIODSTAKER is a free data retrieval call binding the contract method 0xfde7f371.
//
// Solidity: function UNBONDING_PERIOD_STAKER() view returns(uint256)
func (_GovConst *GovConstCallerSession) UNBONDINGPERIODSTAKER() (*big.Int, error) {
	return _GovConst.Contract.UNBONDINGPERIODSTAKER(&_GovConst.CallOpts)
}

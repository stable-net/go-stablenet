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

// GovConfigMetaData contains all meta data concerning the GovConfig contract.
var GovConfigMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"changeFeeDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feePrecision\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maximumStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minStakers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minimumStaking\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unbondingPeriodDelegator\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unbondingPeriodStaker\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"3e9e0ade": "changeFeeDelay()",
		"35ff1e28": "feePrecision()",
		"478b6653": "maximumStaking()",
		"5f0b96e5": "minStakers()",
		"d479ed71": "minimumStaking()",
		"4dbd834d": "unbondingPeriodDelegator()",
		"ea3e16f6": "unbondingPeriodStaker()",
	},
	Bin: "0x608060405234801561001057600080fd5b5060f88061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060735760003560e01c80634dbd834d1160545780634dbd834d1460a25780635f0b96e51460aa578063d479ed711460b2578063ea3e16f61460ba57600080fd5b806335ff1e281460785780633e9e0ade146092578063478b665314609a575b600080fd5b608060045481565b60405190815260200160405180910390f35b608060055481565b608060015481565b608060035481565b608060065481565b608060005481565b60806002548156fea26469706673582212204624dc23658fe20a00f7f252c41a22deea300b5ef72cf6150f78cc27758cc6e664736f6c634300080e0033",
}

// GovConfigABI is the input ABI used to generate the binding from.
// Deprecated: Use GovConfigMetaData.ABI instead.
var GovConfigABI = GovConfigMetaData.ABI

// Deprecated: Use GovConfigMetaData.Sigs instead.
// GovConfigFuncSigs maps the 4-byte function signature to its string representation.
var GovConfigFuncSigs = GovConfigMetaData.Sigs

// GovConfigBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovConfigMetaData.Bin instead.
var GovConfigBin = GovConfigMetaData.Bin

// DeployGovConfig deploys a new Ethereum contract, binding an instance of GovConfig to it.
func DeployGovConfig(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovConfig, error) {
	parsed, err := GovConfigMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovConfigBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovConfig{GovConfigCaller: GovConfigCaller{contract: contract}, GovConfigTransactor: GovConfigTransactor{contract: contract}, GovConfigFilterer: GovConfigFilterer{contract: contract}}, nil
}

// GovConfig is an auto generated Go binding around an Ethereum contract.
type GovConfig struct {
	GovConfigCaller     // Read-only binding to the contract
	GovConfigTransactor // Write-only binding to the contract
	GovConfigFilterer   // Log filterer for contract events
}

// GovConfigCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovConfigCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConfigTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovConfigTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConfigFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovConfigFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovConfigSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovConfigSession struct {
	Contract     *GovConfig        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovConfigCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovConfigCallerSession struct {
	Contract *GovConfigCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// GovConfigTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovConfigTransactorSession struct {
	Contract     *GovConfigTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// GovConfigRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovConfigRaw struct {
	Contract *GovConfig // Generic contract binding to access the raw methods on
}

// GovConfigCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovConfigCallerRaw struct {
	Contract *GovConfigCaller // Generic read-only contract binding to access the raw methods on
}

// GovConfigTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovConfigTransactorRaw struct {
	Contract *GovConfigTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovConfig creates a new instance of GovConfig, bound to a specific deployed contract.
func NewGovConfig(address common.Address, backend bind.ContractBackend) (*GovConfig, error) {
	contract, err := bindGovConfig(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovConfig{GovConfigCaller: GovConfigCaller{contract: contract}, GovConfigTransactor: GovConfigTransactor{contract: contract}, GovConfigFilterer: GovConfigFilterer{contract: contract}}, nil
}

// NewGovConfigCaller creates a new read-only instance of GovConfig, bound to a specific deployed contract.
func NewGovConfigCaller(address common.Address, caller bind.ContractCaller) (*GovConfigCaller, error) {
	contract, err := bindGovConfig(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovConfigCaller{contract: contract}, nil
}

// NewGovConfigTransactor creates a new write-only instance of GovConfig, bound to a specific deployed contract.
func NewGovConfigTransactor(address common.Address, transactor bind.ContractTransactor) (*GovConfigTransactor, error) {
	contract, err := bindGovConfig(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovConfigTransactor{contract: contract}, nil
}

// NewGovConfigFilterer creates a new log filterer instance of GovConfig, bound to a specific deployed contract.
func NewGovConfigFilterer(address common.Address, filterer bind.ContractFilterer) (*GovConfigFilterer, error) {
	contract, err := bindGovConfig(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovConfigFilterer{contract: contract}, nil
}

// bindGovConfig binds a generic wrapper to an already deployed contract.
func bindGovConfig(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovConfigMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovConfig *GovConfigRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovConfig.Contract.GovConfigCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovConfig *GovConfigRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovConfig.Contract.GovConfigTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovConfig *GovConfigRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovConfig.Contract.GovConfigTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovConfig *GovConfigCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovConfig.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovConfig *GovConfigTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovConfig.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovConfig *GovConfigTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovConfig.Contract.contract.Transact(opts, method, params...)
}

// ChangeFeeDelay is a free data retrieval call binding the contract method 0x3e9e0ade.
//
// Solidity: function changeFeeDelay() view returns(uint256)
func (_GovConfig *GovConfigCaller) ChangeFeeDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "changeFeeDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ChangeFeeDelay is a free data retrieval call binding the contract method 0x3e9e0ade.
//
// Solidity: function changeFeeDelay() view returns(uint256)
func (_GovConfig *GovConfigSession) ChangeFeeDelay() (*big.Int, error) {
	return _GovConfig.Contract.ChangeFeeDelay(&_GovConfig.CallOpts)
}

// ChangeFeeDelay is a free data retrieval call binding the contract method 0x3e9e0ade.
//
// Solidity: function changeFeeDelay() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) ChangeFeeDelay() (*big.Int, error) {
	return _GovConfig.Contract.ChangeFeeDelay(&_GovConfig.CallOpts)
}

// FeePrecision is a free data retrieval call binding the contract method 0x35ff1e28.
//
// Solidity: function feePrecision() view returns(uint256)
func (_GovConfig *GovConfigCaller) FeePrecision(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "feePrecision")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeePrecision is a free data retrieval call binding the contract method 0x35ff1e28.
//
// Solidity: function feePrecision() view returns(uint256)
func (_GovConfig *GovConfigSession) FeePrecision() (*big.Int, error) {
	return _GovConfig.Contract.FeePrecision(&_GovConfig.CallOpts)
}

// FeePrecision is a free data retrieval call binding the contract method 0x35ff1e28.
//
// Solidity: function feePrecision() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) FeePrecision() (*big.Int, error) {
	return _GovConfig.Contract.FeePrecision(&_GovConfig.CallOpts)
}

// MaximumStaking is a free data retrieval call binding the contract method 0x478b6653.
//
// Solidity: function maximumStaking() view returns(uint256)
func (_GovConfig *GovConfigCaller) MaximumStaking(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "maximumStaking")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaximumStaking is a free data retrieval call binding the contract method 0x478b6653.
//
// Solidity: function maximumStaking() view returns(uint256)
func (_GovConfig *GovConfigSession) MaximumStaking() (*big.Int, error) {
	return _GovConfig.Contract.MaximumStaking(&_GovConfig.CallOpts)
}

// MaximumStaking is a free data retrieval call binding the contract method 0x478b6653.
//
// Solidity: function maximumStaking() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) MaximumStaking() (*big.Int, error) {
	return _GovConfig.Contract.MaximumStaking(&_GovConfig.CallOpts)
}

// MinStakers is a free data retrieval call binding the contract method 0x5f0b96e5.
//
// Solidity: function minStakers() view returns(uint256)
func (_GovConfig *GovConfigCaller) MinStakers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "minStakers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStakers is a free data retrieval call binding the contract method 0x5f0b96e5.
//
// Solidity: function minStakers() view returns(uint256)
func (_GovConfig *GovConfigSession) MinStakers() (*big.Int, error) {
	return _GovConfig.Contract.MinStakers(&_GovConfig.CallOpts)
}

// MinStakers is a free data retrieval call binding the contract method 0x5f0b96e5.
//
// Solidity: function minStakers() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) MinStakers() (*big.Int, error) {
	return _GovConfig.Contract.MinStakers(&_GovConfig.CallOpts)
}

// MinimumStaking is a free data retrieval call binding the contract method 0xd479ed71.
//
// Solidity: function minimumStaking() view returns(uint256)
func (_GovConfig *GovConfigCaller) MinimumStaking(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "minimumStaking")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimumStaking is a free data retrieval call binding the contract method 0xd479ed71.
//
// Solidity: function minimumStaking() view returns(uint256)
func (_GovConfig *GovConfigSession) MinimumStaking() (*big.Int, error) {
	return _GovConfig.Contract.MinimumStaking(&_GovConfig.CallOpts)
}

// MinimumStaking is a free data retrieval call binding the contract method 0xd479ed71.
//
// Solidity: function minimumStaking() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) MinimumStaking() (*big.Int, error) {
	return _GovConfig.Contract.MinimumStaking(&_GovConfig.CallOpts)
}

// UnbondingPeriodDelegator is a free data retrieval call binding the contract method 0x4dbd834d.
//
// Solidity: function unbondingPeriodDelegator() view returns(uint256)
func (_GovConfig *GovConfigCaller) UnbondingPeriodDelegator(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "unbondingPeriodDelegator")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnbondingPeriodDelegator is a free data retrieval call binding the contract method 0x4dbd834d.
//
// Solidity: function unbondingPeriodDelegator() view returns(uint256)
func (_GovConfig *GovConfigSession) UnbondingPeriodDelegator() (*big.Int, error) {
	return _GovConfig.Contract.UnbondingPeriodDelegator(&_GovConfig.CallOpts)
}

// UnbondingPeriodDelegator is a free data retrieval call binding the contract method 0x4dbd834d.
//
// Solidity: function unbondingPeriodDelegator() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) UnbondingPeriodDelegator() (*big.Int, error) {
	return _GovConfig.Contract.UnbondingPeriodDelegator(&_GovConfig.CallOpts)
}

// UnbondingPeriodStaker is a free data retrieval call binding the contract method 0xea3e16f6.
//
// Solidity: function unbondingPeriodStaker() view returns(uint256)
func (_GovConfig *GovConfigCaller) UnbondingPeriodStaker(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovConfig.contract.Call(opts, &out, "unbondingPeriodStaker")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UnbondingPeriodStaker is a free data retrieval call binding the contract method 0xea3e16f6.
//
// Solidity: function unbondingPeriodStaker() view returns(uint256)
func (_GovConfig *GovConfigSession) UnbondingPeriodStaker() (*big.Int, error) {
	return _GovConfig.Contract.UnbondingPeriodStaker(&_GovConfig.CallOpts)
}

// UnbondingPeriodStaker is a free data retrieval call binding the contract method 0xea3e16f6.
//
// Solidity: function unbondingPeriodStaker() view returns(uint256)
func (_GovConfig *GovConfigCallerSession) UnbondingPeriodStaker() (*big.Int, error) {
	return _GovConfig.Contract.UnbondingPeriodStaker(&_GovConfig.CallOpts)
}

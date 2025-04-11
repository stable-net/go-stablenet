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

// GovNCPMetaData contains all meta data concerning the GovNCP contract.
var GovNCPMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"}],\"name\":\"NCPAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldNCP\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newNCP\",\"type\":\"address\"}],\"name\":\"NCPChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"}],\"name\":\"NCPRemoved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proposalType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"ncp\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"proposer\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"time\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"endtime\",\"type\":\"uint256\"}],\"name\":\"NewProposal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalID\",\"type\":\"uint256\"}],\"name\":\"ProposalCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"accepted\",\"type\":\"bool\"}],\"name\":\"ProposalFinalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"proposalID\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"accept\",\"type\":\"bool\"}],\"name\":\"Vote\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"VOTING_PERIOD\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_proposalID\",\"type\":\"uint256\"}],\"name\":\"cancelProposal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ncp\",\"type\":\"address\"}],\"name\":\"changeNCP\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentProposalID\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ncp\",\"type\":\"address\"}],\"name\":\"isNCP\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ncpList\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_newNCP\",\"type\":\"address\"}],\"name\":\"newProposalToAddNCP\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_ncp\",\"type\":\"address\"}],\"name\":\"newProposalToRemoveNCP\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_proposalID\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"_accept\",\"type\":\"bool\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Sigs: map[string]string{
		"b1610d7e": "VOTING_PERIOD()",
		"e0a8f6f5": "cancelProposal(uint256)",
		"45fcff21": "changeNCP(address)",
		"63966190": "currentProposalID()",
		"701693e3": "isNCP(address)",
		"34afb7bd": "ncpList()",
		"a998cb3c": "newProposalToAddNCP(address)",
		"fba44fe0": "newProposalToRemoveNCP(address)",
		"c9d27afe": "vote(uint256,bool)",
	},
	Bin: "0x608060405234801561001057600080fd5b50610f2b806100206000396000f3fe608060405234801561001057600080fd5b50600436106100935760003560e01c8063a998cb3c11610066578063a998cb3c14610105578063b1610d7e14610118578063c9d27afe14610122578063e0a8f6f514610135578063fba44fe01461014857600080fd5b806334afb7bd1461009857806345fcff21146100b657806363966190146100cb578063701693e3146100e2575b600080fd5b6100a061015b565b6040516100ad9190610d43565b60405180910390f35b6100c96100c4366004610d90565b61016c565b005b6100d460025481565b6040519081526020016100ad565b6100f56100f0366004610d90565b61029a565b60405190151581526020016100ad565b6100c9610113366004610d90565b6102ac565b6100d462093a8081565b6100c9610130366004610db9565b610326565b6100c9610143366004610dee565b61053e565b6100c9610156366004610d90565b61064a565b606061016760006106eb565b905090565b6101776000336106ff565b61019c5760405162461bcd60e51b815260040161019390610e07565b60405180910390fd5b6101a76000826106ff565b156101e95760405162461bcd60e51b81526020600482015260126024820152716e637020616c72656164792065786973747360701b6044820152606401610193565b3360009081526004602052604090205460ff16156102495760405162461bcd60e51b815260206004820152601e60248201527f62656c6f6e6720696e20616e206f6e2d676f696e672070726f706f73616c00006044820152606401610193565b610254600033610720565b50610260600082610735565b506040516001600160a01b0382169033907f52aef81f7bac2d78a9ab8d4601b95ca4a5e699c7d04290ceae4bb90e4a59901990600090a350565b60006102a681836106ff565b92915050565b6102b76000336106ff565b6102d35760405162461bcd60e51b815260040161019390610e07565b6102de6000826106ff565b156103185760405162461bcd60e51b815260206004820152600a6024820152696e63702065786973747360b01b6044820152606401610193565b61032381600161074a565b50565b6103316000336106ff565b61034d5760405162461bcd60e51b815260040161019390610e07565b600061035883610913565b905080600201544211156103a45760405162461bcd60e51b8152602060048201526013602482015272616c726561647920636c6f73656420766f746560681b6044820152606401610193565b33600090815260088201602052604081205460ff1660028111156103ca576103ca610e36565b146104075760405162461bcd60e51b815260206004820152600d60248201526c185b1c9958591e481d9bdd1959609a1b6044820152606401610193565b6000821561042e5750600681018054600191600061042483610e62565b9190505550610449565b50600781018054600291600061044383610e62565b91905055505b600582018054600181810183556000928352602080842090920180546001600160a01b0319163390811790915583526008850190915260409091208054839260ff19909116908360028111156104a1576104a1610e36565b021790555060408051338152841515602082015285917fcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26910160405180910390a260006104ee60006109c5565b905080836006015460026105029190610e7b565b118061051e5750808360070154600261051b9190610e7b565b10155b15610537576105378384600701548560060154116109cf565b5050505050565b6105496000336106ff565b6105655760405162461bcd60e51b815260040161019390610e07565b600061057082610913565b90508060020154421161063d5760038101546001600160a01b031633146105eb5760405162461bcd60e51b815260206004820152602960248201527f6e6f6e2d70726f706f7365722063616e6e6f742063616e63656c206265666f7260448201526819481d1a5b595bdd5d60ba1b6064820152608401610193565b60058101541561063d5760405162461bcd60e51b815260206004820152601860248201527f63616e6e6f742063616e63656c20616674657220766f746500000000000000006044820152606401610193565b61064681610b56565b5050565b6106556000336106ff565b6106715760405162461bcd60e51b815260040161019390610e07565b61067c6000826106ff565b6106b65760405162461bcd60e51b815260206004820152600b60248201526a0696e76616c6964206e63760ac1b6044820152606401610193565b6106c181600261074a565b6001600160a01b03811633036103235760006106de600254610913565b90506106468160016109cf565b606060006106f883610ba5565b9392505050565b6001600160a01b031660009081526001919091016020526040902054151590565b60006106f8836001600160a01b038416610c01565b60006106f8836001600160a01b038416610cf4565b600254600090815260036020526040812090815460ff16600481111561077257610772610e36565b146107e657428160020154106107ca5760405162461bcd60e51b815260206004820152601c60248201527f70726576696f757320766f746520697320696e2070726f6772657373000000006044820152606401610193565b6107d381610b56565b5060025460009081526003602052604090205b6003810180546001600160a01b0319163317905542600182018190556108109062093a8090610e9a565b6002808301919091556004820180546001600160a01b0386166001600160a01b031982168117835585936001600160a81b03199092161790600160a01b90849081111561085f5761085f610e36565b02179055508054600160ff19918216811783553360009081526004602052604080822080548516841790556001600160a01b03871682529020805490921617905560028054907fd6b5f9a920e142ce49d64e77e2d2b84fadbc05ff9ea207438a58d4b50375556c9084908111156108d8576108d8610e36565b6002840154604080519283526001600160a01b03881660208401523390830152426060830152608082015260a00160405180910390a2505050565b6000600254821461095c5760405162461bcd60e51b81526020600482015260136024820152721a5b9d985b1a59081c1c9bdc1bdcd85b081a59606a1b6044820152606401610193565b5060008181526003602052604090206001815460ff16600481111561098357610983610e36565b146109c05760405162461bcd60e51b815260206004820152600d60248201526c6e6f7420696e20766f74696e6760981b6044820152606401610193565b919050565b60006102a6825490565b8015610abc5760016004830154600160a01b900460ff1660028111156109f7576109f7610e36565b03610a56576004820154610a16906000906001600160a01b0316610735565b5060048201546040516001600160a01b03909116907ffdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a932990600090a2610aac565b6004820154610a70906000906001600160a01b0316610720565b5060048201546040516001600160a01b03909116907f9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c90600090a25b815460ff19166002178255610ac8565b815460ff191660031782555b60038201546001600160a01b039081166000908152600460208181526040808420805460ff199081169091559287015490941683529183902080549091169055600254915183151581527fb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7910160405180910390a260028054906000610b4d83610e62565b91905055505050565b805460ff191660041781556002546040517f789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c90600090a260028054906000610b9d83610e62565b919050555050565b606081600001805480602002602001604051908101604052809291908181526020018280548015610bf557602002820191906000526020600020905b815481526020019060010190808311610be1575b50505050509050919050565b60008181526001830160205260408120548015610cea576000610c25600183610eb2565b8554909150600090610c3990600190610eb2565b9050818114610c9e576000866000018281548110610c5957610c59610ec9565b9060005260206000200154905080876000018481548110610c7c57610c7c610ec9565b6000918252602080832090910192909255918252600188019052604090208390555b8554869080610caf57610caf610edf565b6001900381819060005260206000200160009055905585600101600086815260200190815260200160002060009055600193505050506102a6565b60009150506102a6565b6000818152600183016020526040812054610d3b575081546001818101845560008481526020808220909301849055845484825282860190935260409020919091556102a6565b5060006102a6565b6020808252825182820181905260009190848201906040850190845b81811015610d845783516001600160a01b031683529284019291840191600101610d5f565b50909695505050505050565b600060208284031215610da257600080fd5b81356001600160a01b03811681146106f857600080fd5b60008060408385031215610dcc57600080fd5b8235915060208301358015158114610de357600080fd5b809150509250929050565b600060208284031215610e0057600080fd5b5035919050565b60208082526015908201527406d73672e73656e646572206973206e6f74206e637605c1b604082015260600190565b634e487b7160e01b600052602160045260246000fd5b634e487b7160e01b600052601160045260246000fd5b600060018201610e7457610e74610e4c565b5060010190565b6000816000190483118215151615610e9557610e95610e4c565b500290565b60008219821115610ead57610ead610e4c565b500190565b600082821015610ec457610ec4610e4c565b500390565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052603160045260246000fdfea26469706673582212202eab1b413d5f90946377825f18475e847d2fdfad97c57b88b02bab14dae7f81664736f6c634300080e0033",
}

// GovNCPABI is the input ABI used to generate the binding from.
// Deprecated: Use GovNCPMetaData.ABI instead.
var GovNCPABI = GovNCPMetaData.ABI

// Deprecated: Use GovNCPMetaData.Sigs instead.
// GovNCPFuncSigs maps the 4-byte function signature to its string representation.
var GovNCPFuncSigs = GovNCPMetaData.Sigs

// GovNCPBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GovNCPMetaData.Bin instead.
var GovNCPBin = GovNCPMetaData.Bin

// DeployGovNCP deploys a new Ethereum contract, binding an instance of GovNCP to it.
func DeployGovNCP(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GovNCP, error) {
	parsed, err := GovNCPMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GovNCPBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GovNCP{GovNCPCaller: GovNCPCaller{contract: contract}, GovNCPTransactor: GovNCPTransactor{contract: contract}, GovNCPFilterer: GovNCPFilterer{contract: contract}}, nil
}

// GovNCP is an auto generated Go binding around an Ethereum contract.
type GovNCP struct {
	GovNCPCaller     // Read-only binding to the contract
	GovNCPTransactor // Write-only binding to the contract
	GovNCPFilterer   // Log filterer for contract events
}

// GovNCPCaller is an auto generated read-only Go binding around an Ethereum contract.
type GovNCPCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovNCPTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GovNCPTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovNCPFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GovNCPFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GovNCPSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GovNCPSession struct {
	Contract     *GovNCP           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovNCPCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GovNCPCallerSession struct {
	Contract *GovNCPCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// GovNCPTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GovNCPTransactorSession struct {
	Contract     *GovNCPTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GovNCPRaw is an auto generated low-level Go binding around an Ethereum contract.
type GovNCPRaw struct {
	Contract *GovNCP // Generic contract binding to access the raw methods on
}

// GovNCPCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GovNCPCallerRaw struct {
	Contract *GovNCPCaller // Generic read-only contract binding to access the raw methods on
}

// GovNCPTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GovNCPTransactorRaw struct {
	Contract *GovNCPTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGovNCP creates a new instance of GovNCP, bound to a specific deployed contract.
func NewGovNCP(address common.Address, backend bind.ContractBackend) (*GovNCP, error) {
	contract, err := bindGovNCP(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GovNCP{GovNCPCaller: GovNCPCaller{contract: contract}, GovNCPTransactor: GovNCPTransactor{contract: contract}, GovNCPFilterer: GovNCPFilterer{contract: contract}}, nil
}

// NewGovNCPCaller creates a new read-only instance of GovNCP, bound to a specific deployed contract.
func NewGovNCPCaller(address common.Address, caller bind.ContractCaller) (*GovNCPCaller, error) {
	contract, err := bindGovNCP(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GovNCPCaller{contract: contract}, nil
}

// NewGovNCPTransactor creates a new write-only instance of GovNCP, bound to a specific deployed contract.
func NewGovNCPTransactor(address common.Address, transactor bind.ContractTransactor) (*GovNCPTransactor, error) {
	contract, err := bindGovNCP(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GovNCPTransactor{contract: contract}, nil
}

// NewGovNCPFilterer creates a new log filterer instance of GovNCP, bound to a specific deployed contract.
func NewGovNCPFilterer(address common.Address, filterer bind.ContractFilterer) (*GovNCPFilterer, error) {
	contract, err := bindGovNCP(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GovNCPFilterer{contract: contract}, nil
}

// bindGovNCP binds a generic wrapper to an already deployed contract.
func bindGovNCP(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GovNCPMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovNCP *GovNCPRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovNCP.Contract.GovNCPCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovNCP *GovNCPRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovNCP.Contract.GovNCPTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovNCP *GovNCPRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovNCP.Contract.GovNCPTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GovNCP *GovNCPCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GovNCP.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GovNCP *GovNCPTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GovNCP.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GovNCP *GovNCPTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GovNCP.Contract.contract.Transact(opts, method, params...)
}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_GovNCP *GovNCPCaller) VOTINGPERIOD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovNCP.contract.Call(opts, &out, "VOTING_PERIOD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_GovNCP *GovNCPSession) VOTINGPERIOD() (*big.Int, error) {
	return _GovNCP.Contract.VOTINGPERIOD(&_GovNCP.CallOpts)
}

// VOTINGPERIOD is a free data retrieval call binding the contract method 0xb1610d7e.
//
// Solidity: function VOTING_PERIOD() view returns(uint256)
func (_GovNCP *GovNCPCallerSession) VOTINGPERIOD() (*big.Int, error) {
	return _GovNCP.Contract.VOTINGPERIOD(&_GovNCP.CallOpts)
}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_GovNCP *GovNCPCaller) CurrentProposalID(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GovNCP.contract.Call(opts, &out, "currentProposalID")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_GovNCP *GovNCPSession) CurrentProposalID() (*big.Int, error) {
	return _GovNCP.Contract.CurrentProposalID(&_GovNCP.CallOpts)
}

// CurrentProposalID is a free data retrieval call binding the contract method 0x63966190.
//
// Solidity: function currentProposalID() view returns(uint256)
func (_GovNCP *GovNCPCallerSession) CurrentProposalID() (*big.Int, error) {
	return _GovNCP.Contract.CurrentProposalID(&_GovNCP.CallOpts)
}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_GovNCP *GovNCPCaller) IsNCP(opts *bind.CallOpts, _ncp common.Address) (bool, error) {
	var out []interface{}
	err := _GovNCP.contract.Call(opts, &out, "isNCP", _ncp)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_GovNCP *GovNCPSession) IsNCP(_ncp common.Address) (bool, error) {
	return _GovNCP.Contract.IsNCP(&_GovNCP.CallOpts, _ncp)
}

// IsNCP is a free data retrieval call binding the contract method 0x701693e3.
//
// Solidity: function isNCP(address _ncp) view returns(bool)
func (_GovNCP *GovNCPCallerSession) IsNCP(_ncp common.Address) (bool, error) {
	return _GovNCP.Contract.IsNCP(&_GovNCP.CallOpts, _ncp)
}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_GovNCP *GovNCPCaller) NcpList(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _GovNCP.contract.Call(opts, &out, "ncpList")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_GovNCP *GovNCPSession) NcpList() ([]common.Address, error) {
	return _GovNCP.Contract.NcpList(&_GovNCP.CallOpts)
}

// NcpList is a free data retrieval call binding the contract method 0x34afb7bd.
//
// Solidity: function ncpList() view returns(address[])
func (_GovNCP *GovNCPCallerSession) NcpList() ([]common.Address, error) {
	return _GovNCP.Contract.NcpList(&_GovNCP.CallOpts)
}

// CancelProposal is a paid mutator transaction binding the contract method 0xe0a8f6f5.
//
// Solidity: function cancelProposal(uint256 _proposalID) returns()
func (_GovNCP *GovNCPTransactor) CancelProposal(opts *bind.TransactOpts, _proposalID *big.Int) (*types.Transaction, error) {
	return _GovNCP.contract.Transact(opts, "cancelProposal", _proposalID)
}

// CancelProposal is a paid mutator transaction binding the contract method 0xe0a8f6f5.
//
// Solidity: function cancelProposal(uint256 _proposalID) returns()
func (_GovNCP *GovNCPSession) CancelProposal(_proposalID *big.Int) (*types.Transaction, error) {
	return _GovNCP.Contract.CancelProposal(&_GovNCP.TransactOpts, _proposalID)
}

// CancelProposal is a paid mutator transaction binding the contract method 0xe0a8f6f5.
//
// Solidity: function cancelProposal(uint256 _proposalID) returns()
func (_GovNCP *GovNCPTransactorSession) CancelProposal(_proposalID *big.Int) (*types.Transaction, error) {
	return _GovNCP.Contract.CancelProposal(&_GovNCP.TransactOpts, _proposalID)
}

// ChangeNCP is a paid mutator transaction binding the contract method 0x45fcff21.
//
// Solidity: function changeNCP(address _ncp) returns()
func (_GovNCP *GovNCPTransactor) ChangeNCP(opts *bind.TransactOpts, _ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.contract.Transact(opts, "changeNCP", _ncp)
}

// ChangeNCP is a paid mutator transaction binding the contract method 0x45fcff21.
//
// Solidity: function changeNCP(address _ncp) returns()
func (_GovNCP *GovNCPSession) ChangeNCP(_ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.ChangeNCP(&_GovNCP.TransactOpts, _ncp)
}

// ChangeNCP is a paid mutator transaction binding the contract method 0x45fcff21.
//
// Solidity: function changeNCP(address _ncp) returns()
func (_GovNCP *GovNCPTransactorSession) ChangeNCP(_ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.ChangeNCP(&_GovNCP.TransactOpts, _ncp)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_GovNCP *GovNCPTransactor) NewProposalToAddNCP(opts *bind.TransactOpts, _newNCP common.Address) (*types.Transaction, error) {
	return _GovNCP.contract.Transact(opts, "newProposalToAddNCP", _newNCP)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_GovNCP *GovNCPSession) NewProposalToAddNCP(_newNCP common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.NewProposalToAddNCP(&_GovNCP.TransactOpts, _newNCP)
}

// NewProposalToAddNCP is a paid mutator transaction binding the contract method 0xa998cb3c.
//
// Solidity: function newProposalToAddNCP(address _newNCP) returns()
func (_GovNCP *GovNCPTransactorSession) NewProposalToAddNCP(_newNCP common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.NewProposalToAddNCP(&_GovNCP.TransactOpts, _newNCP)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_GovNCP *GovNCPTransactor) NewProposalToRemoveNCP(opts *bind.TransactOpts, _ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.contract.Transact(opts, "newProposalToRemoveNCP", _ncp)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_GovNCP *GovNCPSession) NewProposalToRemoveNCP(_ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.NewProposalToRemoveNCP(&_GovNCP.TransactOpts, _ncp)
}

// NewProposalToRemoveNCP is a paid mutator transaction binding the contract method 0xfba44fe0.
//
// Solidity: function newProposalToRemoveNCP(address _ncp) returns()
func (_GovNCP *GovNCPTransactorSession) NewProposalToRemoveNCP(_ncp common.Address) (*types.Transaction, error) {
	return _GovNCP.Contract.NewProposalToRemoveNCP(&_GovNCP.TransactOpts, _ncp)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_GovNCP *GovNCPTransactor) Vote(opts *bind.TransactOpts, _proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _GovNCP.contract.Transact(opts, "vote", _proposalID, _accept)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_GovNCP *GovNCPSession) Vote(_proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _GovNCP.Contract.Vote(&_GovNCP.TransactOpts, _proposalID, _accept)
}

// Vote is a paid mutator transaction binding the contract method 0xc9d27afe.
//
// Solidity: function vote(uint256 _proposalID, bool _accept) returns()
func (_GovNCP *GovNCPTransactorSession) Vote(_proposalID *big.Int, _accept bool) (*types.Transaction, error) {
	return _GovNCP.Contract.Vote(&_GovNCP.TransactOpts, _proposalID, _accept)
}

// GovNCPNCPAddedIterator is returned from FilterNCPAdded and is used to iterate over the raw logs and unpacked data for NCPAdded events raised by the GovNCP contract.
type GovNCPNCPAddedIterator struct {
	Event *GovNCPNCPAdded // Event containing the contract specifics and raw log

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
func (it *GovNCPNCPAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPNCPAdded)
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
		it.Event = new(GovNCPNCPAdded)
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
func (it *GovNCPNCPAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPNCPAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPNCPAdded represents a NCPAdded event raised by the GovNCP contract.
type GovNCPNCPAdded struct {
	Ncp common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterNCPAdded is a free log retrieval operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_GovNCP *GovNCPFilterer) FilterNCPAdded(opts *bind.FilterOpts, ncp []common.Address) (*GovNCPNCPAddedIterator, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "NCPAdded", ncpRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPNCPAddedIterator{contract: _GovNCP.contract, event: "NCPAdded", logs: logs, sub: sub}, nil
}

// WatchNCPAdded is a free log subscription operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_GovNCP *GovNCPFilterer) WatchNCPAdded(opts *bind.WatchOpts, sink chan<- *GovNCPNCPAdded, ncp []common.Address) (event.Subscription, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "NCPAdded", ncpRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPNCPAdded)
				if err := _GovNCP.contract.UnpackLog(event, "NCPAdded", log); err != nil {
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

// ParseNCPAdded is a log parse operation binding the contract event 0xfdebb9fe9f62427c1a85b79db458c2f8e39a46226369241ddbc06d92395a9329.
//
// Solidity: event NCPAdded(address indexed ncp)
func (_GovNCP *GovNCPFilterer) ParseNCPAdded(log types.Log) (*GovNCPNCPAdded, error) {
	event := new(GovNCPNCPAdded)
	if err := _GovNCP.contract.UnpackLog(event, "NCPAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPNCPChangedIterator is returned from FilterNCPChanged and is used to iterate over the raw logs and unpacked data for NCPChanged events raised by the GovNCP contract.
type GovNCPNCPChangedIterator struct {
	Event *GovNCPNCPChanged // Event containing the contract specifics and raw log

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
func (it *GovNCPNCPChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPNCPChanged)
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
		it.Event = new(GovNCPNCPChanged)
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
func (it *GovNCPNCPChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPNCPChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPNCPChanged represents a NCPChanged event raised by the GovNCP contract.
type GovNCPNCPChanged struct {
	OldNCP common.Address
	NewNCP common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNCPChanged is a free log retrieval operation binding the contract event 0x52aef81f7bac2d78a9ab8d4601b95ca4a5e699c7d04290ceae4bb90e4a599019.
//
// Solidity: event NCPChanged(address indexed oldNCP, address indexed newNCP)
func (_GovNCP *GovNCPFilterer) FilterNCPChanged(opts *bind.FilterOpts, oldNCP []common.Address, newNCP []common.Address) (*GovNCPNCPChangedIterator, error) {

	var oldNCPRule []interface{}
	for _, oldNCPItem := range oldNCP {
		oldNCPRule = append(oldNCPRule, oldNCPItem)
	}
	var newNCPRule []interface{}
	for _, newNCPItem := range newNCP {
		newNCPRule = append(newNCPRule, newNCPItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "NCPChanged", oldNCPRule, newNCPRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPNCPChangedIterator{contract: _GovNCP.contract, event: "NCPChanged", logs: logs, sub: sub}, nil
}

// WatchNCPChanged is a free log subscription operation binding the contract event 0x52aef81f7bac2d78a9ab8d4601b95ca4a5e699c7d04290ceae4bb90e4a599019.
//
// Solidity: event NCPChanged(address indexed oldNCP, address indexed newNCP)
func (_GovNCP *GovNCPFilterer) WatchNCPChanged(opts *bind.WatchOpts, sink chan<- *GovNCPNCPChanged, oldNCP []common.Address, newNCP []common.Address) (event.Subscription, error) {

	var oldNCPRule []interface{}
	for _, oldNCPItem := range oldNCP {
		oldNCPRule = append(oldNCPRule, oldNCPItem)
	}
	var newNCPRule []interface{}
	for _, newNCPItem := range newNCP {
		newNCPRule = append(newNCPRule, newNCPItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "NCPChanged", oldNCPRule, newNCPRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPNCPChanged)
				if err := _GovNCP.contract.UnpackLog(event, "NCPChanged", log); err != nil {
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

// ParseNCPChanged is a log parse operation binding the contract event 0x52aef81f7bac2d78a9ab8d4601b95ca4a5e699c7d04290ceae4bb90e4a599019.
//
// Solidity: event NCPChanged(address indexed oldNCP, address indexed newNCP)
func (_GovNCP *GovNCPFilterer) ParseNCPChanged(log types.Log) (*GovNCPNCPChanged, error) {
	event := new(GovNCPNCPChanged)
	if err := _GovNCP.contract.UnpackLog(event, "NCPChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPNCPRemovedIterator is returned from FilterNCPRemoved and is used to iterate over the raw logs and unpacked data for NCPRemoved events raised by the GovNCP contract.
type GovNCPNCPRemovedIterator struct {
	Event *GovNCPNCPRemoved // Event containing the contract specifics and raw log

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
func (it *GovNCPNCPRemovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPNCPRemoved)
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
		it.Event = new(GovNCPNCPRemoved)
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
func (it *GovNCPNCPRemovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPNCPRemovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPNCPRemoved represents a NCPRemoved event raised by the GovNCP contract.
type GovNCPNCPRemoved struct {
	Ncp common.Address
	Raw types.Log // Blockchain specific contextual infos
}

// FilterNCPRemoved is a free log retrieval operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_GovNCP *GovNCPFilterer) FilterNCPRemoved(opts *bind.FilterOpts, ncp []common.Address) (*GovNCPNCPRemovedIterator, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "NCPRemoved", ncpRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPNCPRemovedIterator{contract: _GovNCP.contract, event: "NCPRemoved", logs: logs, sub: sub}, nil
}

// WatchNCPRemoved is a free log subscription operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_GovNCP *GovNCPFilterer) WatchNCPRemoved(opts *bind.WatchOpts, sink chan<- *GovNCPNCPRemoved, ncp []common.Address) (event.Subscription, error) {

	var ncpRule []interface{}
	for _, ncpItem := range ncp {
		ncpRule = append(ncpRule, ncpItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "NCPRemoved", ncpRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPNCPRemoved)
				if err := _GovNCP.contract.UnpackLog(event, "NCPRemoved", log); err != nil {
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

// ParseNCPRemoved is a log parse operation binding the contract event 0x9f091a55550f1f18f409a75ec4f0ac28d9571c610c14670b40e18f6be1494e2c.
//
// Solidity: event NCPRemoved(address indexed ncp)
func (_GovNCP *GovNCPFilterer) ParseNCPRemoved(log types.Log) (*GovNCPNCPRemoved, error) {
	event := new(GovNCPNCPRemoved)
	if err := _GovNCP.contract.UnpackLog(event, "NCPRemoved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPNewProposalIterator is returned from FilterNewProposal and is used to iterate over the raw logs and unpacked data for NewProposal events raised by the GovNCP contract.
type GovNCPNewProposalIterator struct {
	Event *GovNCPNewProposal // Event containing the contract specifics and raw log

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
func (it *GovNCPNewProposalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPNewProposal)
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
		it.Event = new(GovNCPNewProposal)
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
func (it *GovNCPNewProposalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPNewProposalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPNewProposal represents a NewProposal event raised by the GovNCP contract.
type GovNCPNewProposal struct {
	Id           *big.Int
	ProposalType *big.Int
	Ncp          common.Address
	Proposer     common.Address
	Time         *big.Int
	Endtime      *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNewProposal is a free log retrieval operation binding the contract event 0xd6b5f9a920e142ce49d64e77e2d2b84fadbc05ff9ea207438a58d4b50375556c.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer, uint256 time, uint256 endtime)
func (_GovNCP *GovNCPFilterer) FilterNewProposal(opts *bind.FilterOpts, id []*big.Int) (*GovNCPNewProposalIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "NewProposal", idRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPNewProposalIterator{contract: _GovNCP.contract, event: "NewProposal", logs: logs, sub: sub}, nil
}

// WatchNewProposal is a free log subscription operation binding the contract event 0xd6b5f9a920e142ce49d64e77e2d2b84fadbc05ff9ea207438a58d4b50375556c.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer, uint256 time, uint256 endtime)
func (_GovNCP *GovNCPFilterer) WatchNewProposal(opts *bind.WatchOpts, sink chan<- *GovNCPNewProposal, id []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "NewProposal", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPNewProposal)
				if err := _GovNCP.contract.UnpackLog(event, "NewProposal", log); err != nil {
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

// ParseNewProposal is a log parse operation binding the contract event 0xd6b5f9a920e142ce49d64e77e2d2b84fadbc05ff9ea207438a58d4b50375556c.
//
// Solidity: event NewProposal(uint256 indexed id, uint256 proposalType, address ncp, address proposer, uint256 time, uint256 endtime)
func (_GovNCP *GovNCPFilterer) ParseNewProposal(log types.Log) (*GovNCPNewProposal, error) {
	event := new(GovNCPNewProposal)
	if err := _GovNCP.contract.UnpackLog(event, "NewProposal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPProposalCanceledIterator is returned from FilterProposalCanceled and is used to iterate over the raw logs and unpacked data for ProposalCanceled events raised by the GovNCP contract.
type GovNCPProposalCanceledIterator struct {
	Event *GovNCPProposalCanceled // Event containing the contract specifics and raw log

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
func (it *GovNCPProposalCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPProposalCanceled)
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
		it.Event = new(GovNCPProposalCanceled)
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
func (it *GovNCPProposalCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPProposalCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPProposalCanceled represents a ProposalCanceled event raised by the GovNCP contract.
type GovNCPProposalCanceled struct {
	ProposalID *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalCanceled is a free log retrieval operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed proposalID)
func (_GovNCP *GovNCPFilterer) FilterProposalCanceled(opts *bind.FilterOpts, proposalID []*big.Int) (*GovNCPProposalCanceledIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "ProposalCanceled", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPProposalCanceledIterator{contract: _GovNCP.contract, event: "ProposalCanceled", logs: logs, sub: sub}, nil
}

// WatchProposalCanceled is a free log subscription operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed proposalID)
func (_GovNCP *GovNCPFilterer) WatchProposalCanceled(opts *bind.WatchOpts, sink chan<- *GovNCPProposalCanceled, proposalID []*big.Int) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "ProposalCanceled", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPProposalCanceled)
				if err := _GovNCP.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
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

// ParseProposalCanceled is a log parse operation binding the contract event 0x789cf55be980739dad1d0699b93b58e806b51c9d96619bfa8fe0a28abaa7b30c.
//
// Solidity: event ProposalCanceled(uint256 indexed proposalID)
func (_GovNCP *GovNCPFilterer) ParseProposalCanceled(log types.Log) (*GovNCPProposalCanceled, error) {
	event := new(GovNCPProposalCanceled)
	if err := _GovNCP.contract.UnpackLog(event, "ProposalCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPProposalFinalizedIterator is returned from FilterProposalFinalized and is used to iterate over the raw logs and unpacked data for ProposalFinalized events raised by the GovNCP contract.
type GovNCPProposalFinalizedIterator struct {
	Event *GovNCPProposalFinalized // Event containing the contract specifics and raw log

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
func (it *GovNCPProposalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPProposalFinalized)
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
		it.Event = new(GovNCPProposalFinalized)
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
func (it *GovNCPProposalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPProposalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPProposalFinalized represents a ProposalFinalized event raised by the GovNCP contract.
type GovNCPProposalFinalized struct {
	ProposalID *big.Int
	Accepted   bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterProposalFinalized is a free log retrieval operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_GovNCP *GovNCPFilterer) FilterProposalFinalized(opts *bind.FilterOpts, proposalID []*big.Int) (*GovNCPProposalFinalizedIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "ProposalFinalized", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPProposalFinalizedIterator{contract: _GovNCP.contract, event: "ProposalFinalized", logs: logs, sub: sub}, nil
}

// WatchProposalFinalized is a free log subscription operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_GovNCP *GovNCPFilterer) WatchProposalFinalized(opts *bind.WatchOpts, sink chan<- *GovNCPProposalFinalized, proposalID []*big.Int) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "ProposalFinalized", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPProposalFinalized)
				if err := _GovNCP.contract.UnpackLog(event, "ProposalFinalized", log); err != nil {
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

// ParseProposalFinalized is a log parse operation binding the contract event 0xb5ac567fcf1b069e0235e4f16734625c6cf54c1b40517fd9eb85517f6e1265a7.
//
// Solidity: event ProposalFinalized(uint256 indexed proposalID, bool accepted)
func (_GovNCP *GovNCPFilterer) ParseProposalFinalized(log types.Log) (*GovNCPProposalFinalized, error) {
	event := new(GovNCPProposalFinalized)
	if err := _GovNCP.contract.UnpackLog(event, "ProposalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GovNCPVoteIterator is returned from FilterVote and is used to iterate over the raw logs and unpacked data for Vote events raised by the GovNCP contract.
type GovNCPVoteIterator struct {
	Event *GovNCPVote // Event containing the contract specifics and raw log

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
func (it *GovNCPVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GovNCPVote)
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
		it.Event = new(GovNCPVote)
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
func (it *GovNCPVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GovNCPVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GovNCPVote represents a Vote event raised by the GovNCP contract.
type GovNCPVote struct {
	ProposalID *big.Int
	Voter      common.Address
	Accept     bool
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVote is a free log retrieval operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_GovNCP *GovNCPFilterer) FilterVote(opts *bind.FilterOpts, proposalID []*big.Int) (*GovNCPVoteIterator, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.FilterLogs(opts, "Vote", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return &GovNCPVoteIterator{contract: _GovNCP.contract, event: "Vote", logs: logs, sub: sub}, nil
}

// WatchVote is a free log subscription operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_GovNCP *GovNCPFilterer) WatchVote(opts *bind.WatchOpts, sink chan<- *GovNCPVote, proposalID []*big.Int) (event.Subscription, error) {

	var proposalIDRule []interface{}
	for _, proposalIDItem := range proposalID {
		proposalIDRule = append(proposalIDRule, proposalIDItem)
	}

	logs, sub, err := _GovNCP.contract.WatchLogs(opts, "Vote", proposalIDRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GovNCPVote)
				if err := _GovNCP.contract.UnpackLog(event, "Vote", log); err != nil {
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

// ParseVote is a log parse operation binding the contract event 0xcfa82ef0390c8f3e57ebe6c0665352a383667e792af012d350d9786ee5173d26.
//
// Solidity: event Vote(uint256 indexed proposalID, address voter, bool accept)
func (_GovNCP *GovNCPFilterer) ParseVote(log types.Log) (*GovNCPVote, error) {
	event := new(GovNCPVote)
	if err := _GovNCP.contract.UnpackLog(event, "Vote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

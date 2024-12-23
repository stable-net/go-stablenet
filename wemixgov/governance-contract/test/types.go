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

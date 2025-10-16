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
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/status-im/keycard-go/hexutils"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
	govwbft "github.com/ethereum/go-ethereum/wemixgov/governance-wbft"
	"github.com/stretchr/testify/require"
)

var (
	TestGovValidatorAddress = params.DefaultGovValidatorAddress
)

var (
	compiledWBFT compiledContractWBFT
)

func init() {
	compiledWBFT.Compile("../contracts-wbft/v1", "../contracts")
}

type compiledContractWBFT struct {
	GovValidator *bindContract
}

func (c *compiledContractWBFT) Compile(root, openzeppelinPath string) {
	if contracts, err := compile.Compile(openzeppelinPath,
		filepath.Join(root, "GovValidator.sol"),
	); err != nil {
		panic(err)
	} else {
		if c.GovValidator, err = newBindContract(contracts["GovValidator"]); err != nil {
			panic(err)
		}
	}
}

type GovWBFT struct {
	backend      *simulated.WBFTBackend
	owner        *bind.TransactOpts
	govValidator *bind.BoundContract
}

type TestCandidate struct {
	Validator *EOA
	Operator  *EOA
}

func NewTestCandidate() *TestCandidate {
	return &TestCandidate{
		Validator: NewEOA(),
		Operator:  NewEOA(),
	}
}

func (s *TestCandidate) GetBLSSecretKey() (bls.SecretKey, error) {
	blsSecretKey, err := bls.DeriveFromECDSA(s.Validator.PrivateKey)
	if err != nil {
		return nil, err
	}
	return blsSecretKey, nil
}

func (s *TestCandidate) BLSSign(t *testing.T, msg []byte) bls.Signature {
	sk, err := s.GetBLSSecretKey()
	if err != nil {
		t.Errorf("failed to get bls secret key: %v", err)
		return nil
	}
	return sk.Sign(msg)
}

func (s *TestCandidate) GetBLSPublicKey(t *testing.T) bls.PublicKey {
	blsSecretKey, err := s.GetBLSSecretKey()
	if err != nil {
		t.Errorf("failed to get bls secret key: %v", err)
		return nil
	}
	return blsSecretKey.PublicKey()
}

func (s *TestCandidate) GetBLSPoPSignature(t *testing.T) bls.Signature {
	pk := s.GetBLSPublicKey(t)
	return s.BLSSign(t, pk.Marshal())
}

var defaultBlockPeriod time.Duration

func NewGovWBFT(t *testing.T, customValidators []*TestCandidate, alloc types.GenesisAlloc) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")

	if alloc == nil {
		alloc = make(types.GenesisAlloc)
	}
	alloc[owner.From] = types.Account{Balance: MAX_UINT_128}
	var customEthConf *ethconfig.Config
	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWBFTBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			customEthConf = ethConf
			defaultBlockPeriod = time.Duration(ethConf.Genesis.Config.Anzeon.WBFT.BlockPeriodSeconds) * time.Second
		}),
	}
	members := ""
	validators := ""
	blsPubKeys := ""
	if len(customValidators) > 0 {
		for i, v := range customValidators {
			if i > 0 {
				members = members + ","
				validators = validators + ","
				blsPubKeys = blsPubKeys + ","
			}
			members = members + v.Operator.Address.String()
			validators = validators + v.Validator.Address.String()
			blsKey := v.GetBLSPublicKey(t)
			blsPubKeys = blsPubKeys + hexutils.BytesToHex(blsKey.Marshal())
		}
	} else {
		members = customEthConf.Genesis.Config.Anzeon.Init.Validators[0].String()
		validators = customEthConf.Genesis.Config.Anzeon.Init.Validators[0].String()
		blsPubKeys = customEthConf.Genesis.Config.Anzeon.Init.BLSPublicKeys[0]
	}
	g.backend.CommitWithState(&params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: TestGovValidatorAddress,
			Version: govwbft.SYSTEM_CONTRACT_VERSION_1,
			Params: map[string]string{
				"members":       members,
				"quorum":        "2",
				"expiry":        "604800", // 7 days
				"memberVersion": "1",
				"validators":    validators,
				"blsPublicKeys": blsPubKeys,
			},
		},
	}, nil)
	g.govValidator = compiledWBFT.GovValidator.New(g.backend.Client(), TestGovValidatorAddress)
	return g, nil
}

func (g *GovWBFT) ExpectedOk(tx *types.Transaction, txErr error) (*types.Receipt, error) {
	receipt, err := commitTx(g.backend, tx, txErr)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (g *GovWBFT) ExpectedFail(tx *types.Transaction, txErr error) error {
	_, err := commitTx(g.backend, tx, txErr)
	if err != nil {
		return err
	}
	return nil
}

// GovBase Contract
func (g *GovWBFT) BaseQuorum(contract *bind.BoundContract, sender *EOA) (uint32, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "quorum")
	if err != nil {
		return 0, err
	}
	return result[0].(uint32), nil
}

func (g *GovWBFT) BaseProposalExpiry(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "proposalExpiry")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseMembers(contract *bind.BoundContract, sender *EOA, member common.Address) (govwbft.Member, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "members", member)
	if err != nil {
		return govwbft.Member{}, err
	}
	resultMember := govwbft.Member{
		IsActive: result[0].(bool),
		JoinedAt: result[1].(uint32),
	}

	return resultMember, nil
}

func (g *GovWBFT) BaseVersionedMemberList(contract *bind.BoundContract, sender *EOA, version *big.Int, index *big.Int) (common.Address, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "versionedMemberList", version, index)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) BaseMemberVersion(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "memberVersion")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseGetProposal(contract *bind.BoundContract, sender *EOA) (govwbft.Proposal, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "getProposal")
	if err != nil {
		return govwbft.Proposal{}, err
	}
	proposal := *abi.ConvertType(result[0], new(govwbft.Proposal)).(*govwbft.Proposal)
	return proposal, nil
}

func (g *GovWBFT) BaseTxProposeAddMember(t *testing.T, contract *bind.BoundContract, sender *EOA, newMember common.Address, newQuorum uint32) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "proposeAddMember", newMember, newQuorum)
}

func (g *GovWBFT) BaseTxApproveProposal(t *testing.T, contract *bind.BoundContract, sender *EOA) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "approveProposal")
}

// GovValidator Contract
func (g *GovWBFT) ConfigureValidator(t *testing.T, v *TestCandidate) (*types.Transaction, error) {
	blsPubKey := v.GetBLSPublicKey(t)
	blsPoPSig := v.GetBLSPoPSignature(t)
	return g.validatorContractTx(t, "configureValidator", v.Operator, v.Validator.Address, blsPubKey.Marshal(), blsPoPSig.Marshal())
}

func (g *GovWBFT) ValidatorList(sender *EOA) ([]common.Address, error) {
	var result []interface{}
	err := g.validatorCall("validatorList", sender, &result)
	if err != nil {
		return nil, err
	}
	validators := make([]common.Address, len(result))
	for i, v := range result {
		validators[i] = v.(common.Address)
	}
	return validators, nil
}

func (g *GovWBFT) IsValidator(sender *EOA, addr common.Address) (bool, error) {
	var result []interface{}
	err := g.validatorCall("isValidator", sender, &result, addr)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

func (g *GovWBFT) ValidatorCount(sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := g.validatorCall("validatorCount", sender, &result)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) ValidatorToOperator(sender *EOA, addr common.Address) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("validatorToOperator", sender, &result, addr)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) OperatorToValidator(sender *EOA, addr common.Address) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("operatorToValidator", sender, &result, addr)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) ValidatorToBlsKey(sender *EOA, addr common.Address) ([]byte, error) {
	var result []interface{}
	err := g.validatorCall("validatorToBlsKey", sender, &result, addr)
	if err != nil {
		return nil, err
	}
	return result[0].([]byte), nil
}

func (g *GovWBFT) BlsKeyToValidator(sender *EOA, blsKey []byte) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("blsKeyToValidator", sender, &result, blsKey)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) validatorContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.govValidator.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
}

func (g *GovWBFT) validatorCall(method string, sender *EOA, result *[]interface{}, params ...interface{}) error {
	return g.govValidator.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, result, method, params...)
}

// General Functions
func (g *GovWBFT) BalanceAt(t *testing.T, ctx context.Context, addr common.Address, num *big.Int) *big.Int {
	balance, err := g.backend.Client().BalanceAt(ctx, addr, num)
	require.NoError(t, err)

	return balance
}

func (g *GovWBFT) AdjustTime(adjustment time.Duration) {
	g.backend.AdjustTime(adjustment)
	g.backend.AdjustTime(defaultBlockPeriod)
}

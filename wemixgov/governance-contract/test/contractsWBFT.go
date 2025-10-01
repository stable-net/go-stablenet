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
	backend      *simulated.WbftBackend
	owner        *bind.TransactOpts
	govValidator *bind.BoundContract
}

var defaultBlockPeriod time.Duration

func NewGovWBFT(t *testing.T, ncpList []common.Address, alloc types.GenesisAlloc) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")

	if alloc == nil {
		alloc = make(types.GenesisAlloc)
	}
	alloc[owner.From] = types.Account{Balance: MAX_UINT_128}
	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWbftBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			defaultBlockPeriod = time.Duration(ethConf.Genesis.Config.Anzeon.Wbft.BlockPeriodSeconds) * time.Second
		}),
	}
	if len(ncpList) > 0 {
		g.backend.CommitWithState(&params.SystemContracts{
			GovValidator: &params.SystemContract{
				Address: TestGovValidatorAddress,
				Version: govwbft.SYSTEM_CONTRACT_VERSION_1,
				Params: map[string]string{
					"members":       "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697",
					"memberVersion": "1",
					"validators":    "0xaA5FAA65e9cC0F74a85b6fDfb5f6991f5C094697",
					"blsPublicKeys": "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b",
				},
			},
		}, nil)
		g.govValidator = compiledWBFT.GovValidator.New(g.backend.Client(), TestGovValidatorAddress)
	}
	return g, nil
}

func (g *GovWBFT) Deploy(address common.Address, tx *types.Transaction, contract *bind.BoundContract, txErr error) (common.Address, *bind.BoundContract, error) {
	if txErr != nil {
		return common.Address{}, nil, txErr
	}
	_, err := g.ExpectedOk(tx, txErr)
	return address, contract, err
}

func (g *GovWBFT) ExpectedOk(tx *types.Transaction, txErr error) (*types.Receipt, error) {
	return expectedOk(g.backend, tx, txErr)
}

func (g *GovWBFT) ExpectedFail(tx *types.Transaction, txErr error) error {
	_, err := expectedFail(g.backend, tx, txErr)
	return err
}

// GovBase Contract

// GovValidator Contract
func (g *GovWBFT) ConfigureValidator(t *testing.T, v *TestStaker[*EOA]) (*types.Transaction, error) {
	blsPubKey, err := v.GetBLSPublicKey()
	if err != nil {
		return nil, err
	}
	blsPoPSig, err := v.GetBLSPoPSignature()
	if err != nil {
		return nil, err
	}
	return g.validatorContractTx(t, "configureValidator", v.Operator, v.Staker.Address, blsPubKey.Marshal(), blsPoPSig.Marshal())
}

func (g *GovWBFT) validatorContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.govValidator.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
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

type OperatorType interface {
	*EOA | *CA
}

type TestStaker[T OperatorType] struct {
	Staker   *EOA
	Operator T
}

func NewTestStaker() *TestStaker[*EOA] {
	return &TestStaker[*EOA]{
		Staker:   NewEOA(),
		Operator: NewEOA(),
	}
}

func NewTestStakerWithOperatorCA(opperator *CA) *TestStaker[*CA] {
	return &TestStaker[*CA]{
		Staker:   NewEOA(),
		Operator: opperator,
	}
}

func (s *TestStaker[T]) GetBLSSecretKey() (bls.SecretKey, error) {
	blsSecretKey, err := bls.DeriveFromECDSA(s.Staker.PrivateKey)
	if err != nil {
		return nil, err
	}
	return blsSecretKey, nil
}

func (s *TestStaker[T]) BLSSign(msg []byte) (bls.Signature, error) {
	sk, err := s.GetBLSSecretKey()
	if err != nil {
		return nil, err
	}
	return sk.Sign(msg), nil
}

func (s *TestStaker[T]) GetBLSPublicKey() (bls.PublicKey, error) {
	blsSecretKey, err := s.GetBLSSecretKey()
	if err != nil {
		return nil, err
	}
	return blsSecretKey.PublicKey(), nil
}

func (s *TestStaker[T]) GetBLSPoPSignature() (bls.Signature, error) {
	pk, err := s.GetBLSPublicKey()
	if err != nil {
		return nil, err
	}
	return s.BLSSign(pk.Marshal())
}

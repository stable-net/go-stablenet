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
	"math/big"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	compile "github.com/ethereum/go-ethereum/wemixgov/governance-contract"
	"github.com/stretchr/testify/require"
)

type Governance struct {
	backend   *simulated.WbftBackend
	owner     *bind.TransactOpts
	nodeInfos []nodeInfo

	registry common.Address
	Registry,
	Gov, GovImp,
	NCPExit, NCPExitImp,
	Staking, StakingImp,
	BallotStorage, BallotStorageImp,
	EnvStorage, EnvStorageImp *bind.BoundContract
}

func NewGovernance(t *testing.T) *Governance {
	owner := getTxOpt(t, "owner")
	backend := simulated.NewWbftBackend(types.GenesisAlloc{
		owner.From: {Balance: MAX_UINT_128},
	})

	return &Governance{
		backend: backend,
		owner:   owner,
		nodeInfos: []nodeInfo{{
			[]byte("name"),
			hexutil.MustDecode("0x6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec012937307647bd3b9a82abe2974e1407241d54947bbb39763a4cac9f77166ad92a0"),
			[]byte("127.0.0.1"),
			big.NewInt(8542),
		}},
	}
}

func (g *Governance) DeployContracts(t *testing.T) *Governance {
	// deploy registry
	registry, Registry, err := g.Deploy(compiled.Registry.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	// deploy impls
	govImp, _, err := g.Deploy(compiled.GovImp.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	ncpExitImp, _, err := g.Deploy(compiled.NCPExitImp.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	stakingImp, _, err := g.Deploy(compiled.StakingImp.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	ballotStorageImp, _, err := g.Deploy(compiled.BallotStorageImp.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)
	envStorageImp, _, err := g.Deploy(compiled.EnvStorageImp.Deploy(g.backend.Client(), g.owner))
	require.NoError(t, err)

	// deploy proxies
	gov, Gov, err := g.Deploy(compiled.Gov.Deploy(g.backend.Client(), g.owner, govImp))
	require.NoError(t, err)
	ncpExit, NCPExit, err := g.Deploy(compiled.NCPExit.Deploy(g.backend.Client(), g.owner, ncpExitImp))
	require.NoError(t, err)
	staking, Staking, err := g.Deploy(compiled.Staking.Deploy(g.backend.Client(), g.owner, stakingImp))
	require.NoError(t, err)
	ballotStorage, BallotStorage, err := g.Deploy(compiled.BallotStorage.Deploy(g.backend.Client(), g.owner, ballotStorageImp))
	require.NoError(t, err)
	envStorage, EnvStorage, err := g.Deploy(compiled.EnvStorage.Deploy(g.backend.Client(), g.owner, envStorageImp))
	require.NoError(t, err)

	// set up g
	g.registry = registry
	g.Registry = Registry
	g.Gov = Gov
	g.NCPExit = NCPExit
	g.Staking = Staking
	g.BallotStorage = BallotStorage
	g.EnvStorage = EnvStorage

	g.GovImp = compiled.GovImp.New(g.backend.Client(), gov)
	g.NCPExitImp = compiled.NCPExitImp.New(g.backend.Client(), ncpExit)
	g.StakingImp = compiled.StakingImp.New(g.backend.Client(), staking)
	g.BallotStorageImp = compiled.BallotStorageImp.New(g.backend.Client(), ballotStorage)
	g.EnvStorageImp = compiled.EnvStorageImp.New(g.backend.Client(), envStorage)

	// set Domains
	require.NoError(t, g.ExpectedOk(g.Registry.Transact(g.owner, "setContractDomain", ToBytes32("GovernanceContract"), gov)))
	require.NoError(t, g.ExpectedOk(g.Registry.Transact(g.owner, "setContractDomain", ToBytes32("Staking"), staking)))
	require.NoError(t, g.ExpectedOk(g.Registry.Transact(g.owner, "setContractDomain", ToBytes32("EnvStorage"), envStorage)))
	require.NoError(t, g.ExpectedOk(g.Registry.Transact(g.owner, "setContractDomain", ToBytes32("BallotStorage"), ballotStorage)))

	// initialize
	require.NoError(t, g.ExpectedOk(g.NCPExitImp.Transact(g.owner, "initialize", registry)))
	require.NoError(t, g.ExpectedOk(g.StakingImp.Transact(g.owner, "init", registry, []byte{})))
	require.NoError(t, g.ExpectedOk(g.BallotStorageImp.Transact(g.owner, "initialize", registry)))
	envNames, envValues := makeEnvParams(
		EnvConstants.BLOCKS_PER,
		EnvConstants.BALLOT_DURATION_MIN,
		EnvConstants.BALLOT_DURATION_MAX,
		EnvConstants.STAKING_MIN,
		EnvConstants.STAKING_MAX,
		EnvConstants.MAX_IDLE_BLOCK_INTERVAL,
		EnvConstants.BLOCK_CREATION_TIME,
		EnvConstants.BLOCK_REWARD_AMOUNT,
		EnvConstants.MAX_PRIORITY_FEE_PER_GAS,
		EnvConstants.BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER,
		EnvConstants.BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD,
		EnvConstants.BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM,
		EnvConstants.BLOCK_REWARD_DISTRIBUTION_MAINTENANCE,
		EnvConstants.MAX_BASE_FEE,
		EnvConstants.BLOCK_GASLIMIT,
		EnvConstants.BASE_FEE_MAX_CHANGE_RATE,
		EnvConstants.GAS_TARGET_PERCENTAGE,
	)
	require.NoError(t, g.ExpectedOk(g.EnvStorageImp.Transact(g.owner, "initialize", registry, envNames, envValues)))

	// staking
	{
		g.owner.Value = LOCK_AMOUNT
		require.NoError(t, g.ExpectedOk(g.StakingImp.Transact(g.owner, "deposit")))
		g.owner.Value = nil
	}
	node := g.nodeInfos[0]
	require.NoError(t, g.ExpectedOk(g.GovImp.Transact(g.owner, "init", registry, LOCK_AMOUNT, node.name, node.enode, node.ip, node.port)))

	return g
}

func (r *Governance) Deploy(address common.Address, tx *types.Transaction, contract *bind.BoundContract, err error) (common.Address, *bind.BoundContract, error) {
	if err != nil {
		return common.Address{}, nil, err
	}
	return address, contract, r.ExpectedOk(tx, err)
}

func (r *Governance) ExpectedOk(tx *types.Transaction, txErr error) error {
	_, err := expectedOk(r.backend, tx, txErr)
	return err
}

func (r *Governance) ExpectedFail(tx *types.Transaction, txErr error) error {
	_, err := expectedFail(r.backend, tx, txErr)
	return err
}

var (
	compiled compiledContract
)

func init() {
	compiled.Compile("../contracts")
}

type compiledContract struct {
	Registry,
	Gov, GovImp,
	NCPExit, NCPExitImp,
	Staking, StakingImp,
	BallotStorage, BallotStorageImp,
	EnvStorage, EnvStorageImp *bindContract
}

func (c *compiledContract) Compile(root string) {
	if contracts, err := compile.Compile("",
		filepath.Join(root, "Registry.sol"),
		filepath.Join(root, "Gov.sol"),
		filepath.Join(root, "GovImp.sol"),
		filepath.Join(root, "NCPExit.sol"),
		filepath.Join(root, "NCPExitImp.sol"),
		filepath.Join(root, "Staking.sol"),
		filepath.Join(root, "StakingImp.sol"),
		filepath.Join(root, "storage", "BallotStorage.sol"),
		filepath.Join(root, "storage", "BallotStorageImp.sol"),
		filepath.Join(root, "storage", "EnvStorage.sol"),
		filepath.Join(root, "storage", "EnvStorageImp.sol"),
	); err != nil {
		panic(err)
	} else {
		if c.Registry, err = newBindContract(contracts["Registry"]); err != nil {
			panic(err)
		} else if c.Gov, err = newBindContract(contracts["Gov"]); err != nil {
			panic(err)
		} else if c.GovImp, err = newBindContract(contracts["GovImp"]); err != nil {
			panic(err)
		} else if c.NCPExit, err = newBindContract(contracts["NCPExit"]); err != nil {
			panic(err)
		} else if c.NCPExitImp, err = newBindContract(contracts["NCPExitImp"]); err != nil {
			panic(err)
		} else if c.Staking, err = newBindContract(contracts["Staking"]); err != nil {
			panic(err)
		} else if c.StakingImp, err = newBindContract(contracts["StakingImp"]); err != nil {
			panic(err)
		} else if c.BallotStorage, err = newBindContract(contracts["BallotStorage"]); err != nil {
			panic(err)
		} else if c.BallotStorageImp, err = newBindContract(contracts["BallotStorageImp"]); err != nil {
			panic(err)
		} else if c.EnvStorage, err = newBindContract(contracts["EnvStorage"]); err != nil {
			panic(err)
		} else if c.EnvStorageImp, err = newBindContract(contracts["EnvStorageImp"]); err != nil {
			panic(err)
		}
	}
}

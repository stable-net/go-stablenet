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
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	client, opts := func() (*simulated.WbftBackend, *bind.TransactOpts) {
		g := NewGovernance(t)
		return g.backend, g.owner
	}()
	go func() {
		for {
			time.Sleep(1e9)
			client.Commit()
		}
	}()
	contracts, err := gov.DeployGovContracts(opts, client.Client(), nil)
	require.NoError(t, err)
	lockAmount := gov.DefaultInitEnvStorage.STAKING_MIN
	gov.ExecuteInitialize(contracts, opts, client.Client(), lockAmount, gov.DefaultInitEnvStorage, gov.InitMembers{
		{
			Staker:  opts.From,
			Voter:   opts.From,
			Reward:  opts.From,
			Name:    "name",
			Enode:   "0x6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec012937307647bd3b9a82abe2974e1407241d54947bbb39763a4cac9f77166ad92a0",
			Ip:      "127.0.0.1",
			Port:    8542,
			Deposit: lockAmount,
		},
	})

	checkMainnetEnvStorageValues(t, new(bind.CallOpts), contracts.EnvStorageImp)
}

func TestCheckMainnetEnvStorageValues(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := ethclient.DialContext(ctx, "https://api.wemix.com/")
	require.NoError(t, err)

	block, err := client.BlockByNumber(ctx, common.Big0)
	require.NoError(t, err)

	callOpts := &bind.CallOpts{Context: ctx}
	contracts, err := gov.GetGovContractsByOwner(callOpts, client, block.Coinbase())
	require.NoError(t, err)

	checkMainnetEnvStorageValues(t, callOpts, contracts.EnvStorageImp)
}

func checkMainnetEnvStorageValues(t *testing.T, callOpts *bind.CallOpts, envStorage *gov.EnvStorageImp) {
	BLOCKS_PER, _ := envStorage.GetBlocksPer(callOpts)
	t.Log("BLOCKS_PER:", BLOCKS_PER)

	BALLOT_DURATION_MIN, _ := envStorage.GetBallotDurationMin(callOpts)
	t.Log("BALLOT_DURATION_MIN:", BALLOT_DURATION_MIN)

	BALLOT_DURATION_MAX, _ := envStorage.GetBallotDurationMax(callOpts)
	t.Log("BALLOT_DURATION_MAX:", BALLOT_DURATION_MAX)

	STAKING_MIN, _ := envStorage.GetStakingMin(callOpts)
	t.Log("STAKING_MIN:", STAKING_MIN)

	STAKING_MAX, _ := envStorage.GetStakingMax(callOpts)
	t.Log("STAKING_MAX:", STAKING_MAX)

	MAX_IDLE_BLOCK_INTERVAL, _ := envStorage.GetMaxIdleBlockInterval(callOpts)
	t.Log("MAX_IDLE_BLOCK_INTERVAL:", MAX_IDLE_BLOCK_INTERVAL)

	BLOCK_CREATION_TIME, _ := envStorage.GetBlockCreationTime(callOpts)
	t.Log("BLOCK_CREATION_TIME:", BLOCK_CREATION_TIME)

	BLOCK_REWARD_AMOUNT, _ := envStorage.GetBlockRewardAmount(callOpts)
	t.Log("BLOCK_REWARD_AMOUNT:", BLOCK_REWARD_AMOUNT)

	MAX_PRIORITY_FEE_PER_GAS, _ := envStorage.GetMaxPriorityFeePerGas(callOpts)
	t.Log("MAX_PRIORITY_FEE_PER_GAS:", MAX_PRIORITY_FEE_PER_GAS)

	BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("blockRewardDistributionBlockProducer")))
	t.Log("BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER:", BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER)

	BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("blockRewardDistributionStakingReward")))
	t.Log("BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD:", BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD)

	BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("blockRewardDistributionEcosystem")))
	t.Log("BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM:", BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM)

	BLOCK_REWARD_DISTRIBUTION_MAINTENANCE, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("blockRewardDistributionMaintenance")))
	t.Log("BLOCK_REWARD_DISTRIBUTION_MAINTENANCE:", BLOCK_REWARD_DISTRIBUTION_MAINTENANCE)

	MAX_BASE_FEE, _ := envStorage.GetMaxBaseFee(callOpts)
	t.Log("MAX_BASE_FEE:", MAX_BASE_FEE)

	BLOCK_GASLIMIT, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("blockGasLimit")))
	t.Log("BLOCK_GASLIMIT:", BLOCK_GASLIMIT)

	BASE_FEE_MAX_CHANGE_RATE, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("baseFeeMaxChangeRate")))
	t.Log("BASE_FEE_MAX_CHANGE_RATE:", BASE_FEE_MAX_CHANGE_RATE)

	GAS_TARGET_PERCENTAGE, _ := envStorage.GetUint(callOpts, crypto.Keccak256Hash([]byte("gasTargetPercentage")))
	t.Log("GAS_TARGET_PERCENTAGE:", GAS_TARGET_PERCENTAGE)
}

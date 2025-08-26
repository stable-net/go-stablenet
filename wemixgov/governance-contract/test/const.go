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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func towei(x int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(x), big.NewInt(params.Ether))
}

func toRewardPerStaking(reward *big.Int, totalStaking *big.Int) *big.Int {
	x := new(big.Int).Mul(reward, new(big.Int).Mul(big.NewInt(params.Ether), big.NewInt(1e9)))
	x.Div(x, totalStaking)
	return x
}

func toGwei(x int64) *big.Int {
	return new(big.Int).Mul(big.NewInt(x), big.NewInt(params.GWei))
}

var (
	LOCK_AMOUNT  *big.Int = towei(1500000)
	MAX_UINT_256          = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
	MAX_INT_256           = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 255), common.Big1)
	MAX_UINT_128          = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 128), common.Big1)
)

var EnvConstants = struct {
	BLOCKS_PER                               env
	BALLOT_DURATION_MIN                      env
	BALLOT_DURATION_MAX                      env
	STAKING_MIN                              env
	STAKING_MAX                              env
	MAX_IDLE_BLOCK_INTERVAL                  env
	BLOCK_CREATION_TIME                      env
	BLOCK_REWARD_AMOUNT                      env
	MAX_PRIORITY_FEE_PER_GAS                 env
	BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER env
	BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD env
	BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM      env
	BLOCK_REWARD_DISTRIBUTION_MAINTENANCE    env
	MAX_BASE_FEE                             env
	BLOCK_GASLIMIT                           env
	BASE_FEE_MAX_CHANGE_RATE                 env
	GAS_TARGET_PERCENTAGE                    env
}{
	BLOCKS_PER:                               newEnvInt64("blocksPer", 1),
	BALLOT_DURATION_MIN:                      newEnvInt64("ballotDurationMin", 86400),
	BALLOT_DURATION_MAX:                      newEnvInt64("ballotDurationMax", 604800),
	STAKING_MIN:                              newEnvBig("stakingMin", towei(1500000)),
	STAKING_MAX:                              newEnvBig("stakingMax", towei(1500000)),
	MAX_IDLE_BLOCK_INTERVAL:                  newEnvInt64("MaxIdleBlockInterval", 5),
	BLOCK_CREATION_TIME:                      newEnvInt64("blockCreationTime", 1000),
	BLOCK_REWARD_AMOUNT:                      newEnvBig("blockRewardAmount", towei(1)),
	MAX_PRIORITY_FEE_PER_GAS:                 newEnvBig("maxPriorityFeePerGas", toGwei(100)),
	BLOCK_REWARD_DISTRIBUTION_BLOCK_PRODUCER: newEnvInt64("blockRewardDistributionBlockProducer", 4000),
	BLOCK_REWARD_DISTRIBUTION_STAKING_REWARD: newEnvInt64("blockRewardDistributionStakingReward", 1000),
	BLOCK_REWARD_DISTRIBUTION_ECOSYSTEM:      newEnvInt64("blockRewardDistributionEcosystem", 2500),
	BLOCK_REWARD_DISTRIBUTION_MAINTENANCE:    newEnvInt64("blockRewardDistributionMaintenance", 2500),
	MAX_BASE_FEE:                             newEnvBig("maxBaseFee", toGwei(5000)),
	BLOCK_GASLIMIT:                           newEnvInt64("blockGasLimit", 1050000000),
	BASE_FEE_MAX_CHANGE_RATE:                 newEnvInt64("baseFeeMaxChangeRate", 46),
	GAS_TARGET_PERCENTAGE:                    newEnvInt64("gasTargetPercentage", 30),
}

type env struct {
	Name  [32]byte
	Value *big.Int
}

func newEnvInt64(name string, value int64) env {
	return env{
		Name:  crypto.Keccak256Hash([]byte(name)),
		Value: big.NewInt(value),
	}
}

func newEnvBig(name string, value *big.Int) env {
	if value == nil {
		value = common.Big0
	}
	return env{
		Name:  crypto.Keccak256Hash([]byte(name)),
		Value: value,
	}
}

func makeEnvParams(envs ...env) (names [][32]byte, values []*big.Int) {
	length := len(envs)
	names = make([][32]byte, length)
	values = make([]*big.Int, length)
	for i, e := range envs {
		names[i] = e.Name
		values[i] = e.Value
	}
	return
}

var EnvTypes = struct {
	Invalid *big.Int
	Int     *big.Int
	Uint    *big.Int
	Address *big.Int
	Bytes32 *big.Int
	Bytes   *big.Int
	String  *big.Int
}{big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5), big.NewInt(6)}

var BallotStates = struct {
	Invalid    *big.Int
	Ready      *big.Int
	InProgress *big.Int
	Accepted   *big.Int
	Rejected   *big.Int
}{big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}

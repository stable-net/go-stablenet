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

package wemixgov

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type NodeInfo struct {
	Name  []byte
	Enode []byte
	Ip    []byte
	Port  *big.Int
}

type GovContractApi interface {
	GetRegistryAddress() common.Address
	GetGovAddress() common.Address
	GetStakingAddress() common.Address
	GetModifiedBlock() (*big.Int, error)
	GetBlockCreationTime() (*big.Int, error)
	GetBlocksPer() (*big.Int, error)
	GetBlockRewardAmount() (*big.Int, error)
	GetMaxPriorityFeePerGas() (*big.Int, error)
	GetMaxBaseFee() (*big.Int, error)
	GetGasLimitAndBaseFee() (*big.Int, *big.Int, *big.Int, error)
	GetNodeLength() (*big.Int, error)
	GetNode(index *big.Int) (NodeInfo, error)
	GetMemberLength() (*big.Int, error)
	GetMember(index *big.Int) (common.Address, error)
	GetBlockRewardDistributionMethod() (*big.Int, *big.Int, *big.Int, *big.Int, error)
	GetStakingRewardAddress() (common.Address, error)
	GetEcosystemAddress() (common.Address, error)
	GetMaintenanceAddress() (common.Address, error)
	GetFeeCollectorAddress() (common.Address, error)
	GetReward(index *big.Int) (common.Address, error)
	GetLockedBalanceOf(common.Address) (*big.Int, error)
}

type GovBackend interface {
	GetGovApiWithHeight(ctx context.Context, height *big.Int) (GovContractApi, error)
}

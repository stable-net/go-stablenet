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

package wpoa

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	wemixPoA *WemixPoA
)

func SetWemixPoA(wpoa *WemixPoA) {
	wemixPoA = wpoa
}

func GetLegacyBlockRewardAmount(height *big.Int) (*big.Int, error) {
	return wemixPoA.govCli.GetLegacyBlockRewardAmount(height)
}

func SuggestGasPrice() *big.Int {
	defaultFee := big.NewInt(100 * params.GWei)
	if wemixPoA == nil || wemixPoA.govCli == nil {
		return defaultFee
	}
	amount, err := wemixPoA.govCli.GetMaxPriorityFeePerGas(nil)
	if err != nil {
		return defaultFee
	}
	return amount
}

func CalcBaseFee(config *params.ChainConfig, parent *types.Header) *big.Int {
	return wemixPoA.CalcBaseFee(config, parent)
}

func Info(block *types.Header) interface{} {
	govInfo, err := wemixPoA.govCli.GetGovInfo(block.Number)
	if err != nil {
		return ""
	}

	info := &map[string]interface{}{
		"registry":                  govInfo.Registry,
		"governance":                govInfo.Gov,
		"staking":                   govInfo.Staking,
		"modifiedblock":             govInfo.ModifiedBlock,
		"BlocksPer":                 govInfo.BlocksPer,
		"blockInterval":             govInfo.BlockInterval,
		"blockReward":               govInfo.BlockReward,
		"maxPriorityFeePerGas":      govInfo.MaxPriorityFeePerGas,
		"blockGasLimit":             govInfo.GasLimit,
		"maxBaseFee":                govInfo.MaxBaseFee,
		"baseFeeMaxChangeRate":      govInfo.BaseFeeMaxChangeRate,
		"gasTargetPercentage":       govInfo.GasTargetPercentage,
		"nodes":                     govInfo.Nodes,
		"DefaultBriocheBlockReward": params.DefaultBriocheBlockReward,
	}
	return info
}

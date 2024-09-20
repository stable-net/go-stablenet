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

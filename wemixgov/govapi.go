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

package cli

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/wemixgov"
	gov "github.com/ethereum/go-ethereum/wemixgov/bind"
)

type govCli struct {
	cli         bind.ContractBackend
	bootAccount common.Address
}

func NewGovCli(cli bind.ContractBackend) (wemixgov.GovBackend, error) {
	return &govCli{
		cli: cli,
	}, nil
}

type govHeightEnv struct {
	contracts *gov.GovContracts
	opts      *bind.CallOpts
}

func (govCli *govCli) GetGovApiWithHeight(ctx context.Context, height *big.Int) (wemixgov.GovContractApi, error) {
	if (govCli.bootAccount == common.Address{}) {
		genesisHeader, err := govCli.cli.HeaderByNumber(ctx, new(big.Int))
		if err != nil {
			return nil, err
		}
		govCli.bootAccount = genesisHeader.Coinbase
	}
	opts := &bind.CallOpts{Context: ctx, BlockNumber: height}
	contracts, err := gov.GetGovContractsByOwner(opts, govCli.cli, govCli.bootAccount)
	if err != nil {
		return nil, err
	}
	return &govHeightEnv{
		contracts: contracts,
		opts:      opts,
	}, nil
}

func (govEnv *govHeightEnv) GetRegistryAddress() common.Address {
	return govEnv.contracts.Address().Registry
}

func (govEnv *govHeightEnv) GetGovAddress() common.Address {
	return govEnv.contracts.Address().Gov
}

func (govEnv *govHeightEnv) GetStakingAddress() common.Address {
	return govEnv.contracts.Address().Staking
}

func (govEnv *govHeightEnv) GetModifiedBlock() (*big.Int, error) {
	return govEnv.contracts.GovImp.ModifiedBlock(govEnv.opts)
}

func (govEnv *govHeightEnv) GetBlockCreationTime() (*big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetBlockCreationTime(govEnv.opts)
}

func (govEnv *govHeightEnv) GetBlocksPer() (*big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetBlocksPer(govEnv.opts)
}

func (govEnv *govHeightEnv) GetBlockRewardAmount() (*big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetBlockRewardAmount(govEnv.opts)
}

func (govEnv *govHeightEnv) GetMaxPriorityFeePerGas() (*big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetMaxPriorityFeePerGas(govEnv.opts)
}

func (govEnv *govHeightEnv) GetMaxBaseFee() (*big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetMaxBaseFee(govEnv.opts)
}

func (govEnv *govHeightEnv) GetGasLimitAndBaseFee() (*big.Int, *big.Int, *big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetGasLimitAndBaseFee(govEnv.opts)
}

func (govEnv *govHeightEnv) GetNodeLength() (*big.Int, error) {
	return govEnv.contracts.GovImp.GetNodeLength(govEnv.opts)
}

func (govEnv *govHeightEnv) GetNode(index *big.Int) (wemixgov.NodeInfo, error) {
	return govEnv.contracts.GovImp.GetNode(govEnv.opts, index)
}

func (govEnv *govHeightEnv) GetMemberLength() (*big.Int, error) {
	return govEnv.contracts.GovImp.GetMemberLength(govEnv.opts)
}

func (govEnv *govHeightEnv) GetMember(index *big.Int) (common.Address, error) {
	return govEnv.contracts.GovImp.GetMember(govEnv.opts, index)
}

func (govEnv *govHeightEnv) GetBlockRewardDistributionMethod() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return govEnv.contracts.EnvStorageImp.GetBlockRewardDistributionMethod(govEnv.opts)
}

func (govEnv *govHeightEnv) GetStakingRewardAddress() (common.Address, error) {
	return govEnv.contracts.Registry.GetContractAddress(govEnv.opts, toBytes32(gov.DOMAIN_StakingReward))
}

func (govEnv *govHeightEnv) GetEcosystemAddress() (common.Address, error) {
	return govEnv.contracts.Registry.GetContractAddress(govEnv.opts, toBytes32(gov.DOMAIN_Ecosystem))
}

func (govEnv *govHeightEnv) GetMaintenanceAddress() (common.Address, error) {
	return govEnv.contracts.Registry.GetContractAddress(govEnv.opts, toBytes32(gov.DOMAIN_Maintenance))
}

func (govEnv *govHeightEnv) GetFeeCollectorAddress() (common.Address, error) {
	return govEnv.contracts.Registry.GetContractAddress(govEnv.opts, toBytes32(gov.DOMAIN_FeeCollector))
}

func (govEnv *govHeightEnv) GetReward(index *big.Int) (common.Address, error) {
	return govEnv.contracts.GovImp.GetReward(govEnv.opts, index)
}

func (govEnv *govHeightEnv) GetLockedBalanceOf(member common.Address) (*big.Int, error) {
	return govEnv.contracts.StakingImp.LockedBalanceOf(govEnv.opts, member)
}

func toBytes32(b string) [32]byte {
	var b32 [32]byte
	if len(b) > len(b32) {
		b = b[len(b)-len(b32):]
	}
	copy(b32[:], []byte(b))
	return b32
}

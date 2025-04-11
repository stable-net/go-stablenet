package test

import (
	"context"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	compiledWBFT compiledContractWBFT
)

func init() {
	compiledWBFT.Compile("../contracts-wbft", "../contracts")
}

type compiledContractWBFT struct {
	GovConst, GovStaking, GovNCP, GovRewardeeImp *bindContract
}

func (c *compiledContractWBFT) Compile(root, openzeppelinPath string) {
	if contracts, err := compile.Compile(openzeppelinPath,
		filepath.Join(root, "GovConst.sol"),
		filepath.Join(root, "GovStaking.sol"),
		filepath.Join(root, "GovNCP.sol"),
		filepath.Join(root, "GovRewardeeImp.sol"),
	); err != nil {
		panic(err)
	} else {
		if c.GovConst, err = newBindContract(contracts["GovConst"]); err != nil {
			panic(err)
		} else if c.GovStaking, err = newBindContract(contracts["GovStaking"]); err != nil {
			panic(err)
		} else if c.GovNCP, err = newBindContract(contracts["GovNCP"]); err != nil {
			panic(err)
		} else if c.GovRewardeeImp, err = newBindContract(contracts["GovRewardeeImp"]); err != nil {
			panic(err)
		}
	}
}

type GovWBFT struct {
	backend         *simulated.WbftBackend
	owner           *bind.TransactOpts
	govConst        *bind.BoundContract
	stakingContract *bind.BoundContract
	ncpContract     *bind.BoundContract
}

var defaultBlockPeriod time.Duration

func NewGovWBFT(t *testing.T, ncpList []common.Address, alloc types.GenesisAlloc) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")

	if alloc == nil {
		alloc = make(types.GenesisAlloc)
	}
	alloc[owner.From] = types.Account{Balance: MAX_UINT_128}
	alloc[govwbft.GovConstAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovConstContract)}
	alloc[govwbft.GovStakingAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovStakingContract)}
	alloc[govwbft.GovRewardeeImpAddress] = types.Account{Code: hexutil.MustDecode(govwbft.GovRewardeeImpContract)}

	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWbftBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			defaultBlockPeriod = time.Duration(ethConf.Genesis.Config.QBFT.BlockPeriodSeconds) * time.Second
		}),
	}
	if len(ncpList) > 0 {
		g.backend.CommitWithState(params.StateTransition{
			Codes:  []params.CodeParam{{Address: govwbft.GovNCPAddress, Code: govwbft.GovNCPContract}},
			States: govwbft.InitializeNCP(ncpList),
		})
	}

	g.govConst = compiledWBFT.GovConst.New(g.backend.Client(), govwbft.GovConstAddress)
	g.stakingContract = compiledWBFT.GovStaking.New(g.backend.Client(), govwbft.GovStakingAddress)
	g.ncpContract = compiledWBFT.GovNCP.New(g.backend.Client(), govwbft.GovNCPAddress)
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

// Staking Contract
func (g *GovWBFT) RegisterStaker(t *testing.T, v *TestStaker, amount *big.Int, fee *big.Int) (*types.Transaction, error) {
	blsPubKey, err := v.GetBLSPublicKey()
	if err != nil {
		return nil, err
	}
	return g.stakingContractTx(t, "registerStaker", v.Operator, amount, amount, v.Staker.Address, v.FeeRecipient.Address, fee, blsPubKey.Marshal())
}

func (g *GovWBFT) Stake(t *testing.T, operator *EOA, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "stake", operator, amount, amount)
}

func (g *GovWBFT) Unstake(t *testing.T, operator *EOA, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "unstake", operator, nil, amount)
}

func (g *GovWBFT) Delegate(t *testing.T, delegator *EOA, staker common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "delegate", delegator, amount, staker, amount)
}

func (g *GovWBFT) Undelegate(t *testing.T, delegator *EOA, staker common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "undelegate", delegator, nil, staker, amount)
}

func (g *GovWBFT) Claim(t *testing.T, user *EOA, staker common.Address, restake bool) (*types.Transaction, error) {
	return g.stakingContractTx(t, "claim", user, nil, staker, restake)
}

func (g *GovWBFT) Withdraw(t *testing.T, sender *EOA, count *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "withdraw", sender, nil, count)
}

func (g *GovWBFT) RequestChangingFee(t *testing.T, sender *EOA, newFeeRate *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "requestChangingFee", sender, nil, newFeeRate)
}

func (g *GovWBFT) ExecuteChangingFee(t *testing.T, sender *EOA, staker common.Address) (*types.Transaction, error) {
	return g.stakingContractTx(t, "executeChangingFee", sender, nil, staker)
}

func (g *GovWBFT) RequestChangeFee(t *testing.T, sender *EOA, newFeeRate *big.Int) (*types.Transaction, error) {
	return g.stakingContractTx(t, "requestChangeFee", sender, nil, newFeeRate)
}

func (g *GovWBFT) ExecuteChangeFee(t *testing.T, sender *EOA, staker common.Address) (*types.Transaction, error) {
	return g.stakingContractTx(t, "executeChangeFee", sender, nil, staker)
}

func (g *GovWBFT) stakingContractTx(t *testing.T, method string, sender *EOA, value *big.Int, params ...interface{}) (*types.Transaction, error) {
	return g.stakingContract.Transact(NewTxOptsWithValue(t, sender, value), method, params...)
}

// NCP Contract
func (g *GovWBFT) NewProposalToAddNCP(t *testing.T, proposer *EOA, ncp common.Address) (*types.Transaction, error) {
	return g.ncpContractTx(t, "newProposalToAddNCP", proposer, nil, ncp)
}

func (g *GovWBFT) NewProposalToRemoveNCP(t *testing.T, proposer *EOA, ncp common.Address) (*types.Transaction, error) {
	return g.ncpContractTx(t, "newProposalToRemoveNCP", proposer, nil, ncp)
}

func (g *GovWBFT) ChangeNCP(t *testing.T, ncp *EOA, newNCP common.Address) (*types.Transaction, error) {
	return g.ncpContractTx(t, "changeNCP", ncp, nil, newNCP)
}

func (g *GovWBFT) Vote(t *testing.T, voter *EOA, proposalID *big.Int, accept bool) (*types.Transaction, error) {
	return g.ncpContractTx(t, "vote", voter, nil, proposalID, accept)
}

func (g *GovWBFT) CancelProposal(t *testing.T, sender *EOA, proposalID *big.Int) (*types.Transaction, error) {
	return g.ncpContractTx(t, "cancelProposal", sender, nil, proposalID)
}

func (g *GovWBFT) ncpContractTx(t *testing.T, method string, sender *EOA, value *big.Int, params ...interface{}) (*types.Transaction, error) {
	return g.ncpContract.Transact(NewTxOptsWithValue(t, sender, value), method, params...)
}

func (g *GovWBFT) balanceAt(t *testing.T, ctx context.Context, addr common.Address, num *big.Int) *big.Int {
	balance, err := g.backend.Client().BalanceAt(ctx, addr, num)
	require.NoError(t, err)

	return balance
}

func (g *GovWBFT) adjustTime(adjustment time.Duration) {
	g.backend.AdjustTime(adjustment)
	g.backend.AdjustTime(defaultBlockPeriod)
}

type TestStaker struct {
	Staker       *EOA
	Operator     *EOA
	FeeRecipient *EOA
}

func NewTestStaker() *TestStaker {
	return &TestStaker{
		Staker:       NewEOA(),
		Operator:     NewEOA(),
		FeeRecipient: NewEOA(),
	}
}

func (s *TestStaker) GetBLSSecretKey() (bls.SecretKey, error) {
	blsSecretKey, err := bls.DeriveFromECDSA(s.Staker.PrivateKey)
	if err != nil {
		return nil, err
	}
	return blsSecretKey, nil
}

func (s *TestStaker) GetBLSPublicKey() (bls.PublicKey, error) {
	blsSecretKey, err := s.GetBLSSecretKey()
	if err != nil {
		return nil, err
	}
	return blsSecretKey.PublicKey(), nil
}

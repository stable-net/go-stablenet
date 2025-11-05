// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package test

import (
	"context"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/bls"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	sc "github.com/ethereum/go-ethereum/systemcontracts"

	compile "github.com/ethereum/go-ethereum/systemcontracts/compile/compiler"
	"github.com/stretchr/testify/require"
)

// mustParseType is a helper function to parse ABI types
func mustParseType(typeName string) abi.Type {
	typ, err := abi.NewType(typeName, "", nil)
	if err != nil {
		panic(err)
	}
	return typ
}

var (
	TestGovValidatorAddress    = params.DefaultGovValidatorAddress
	TestCoinAdapterAddress     = params.DefaultNativeCoinAdapterAddress
	TestGovMinterAddress       = params.DefaultGovMinterAddress
	TestGovMasterMinterAddress = params.DefaultGovMasterMinterAddress
	TestMockFiatTokenAddress   = common.HexToAddress("0x1004")
)

var (
	compiledWBFT compiledContractWBFT
)

func init() {
	compiledWBFT.Compile("../solidity/v1", "../solidity/openzeppelin")
}

type compiledContractWBFT struct {
	GovValidator    *bindContract
	CoinAdapter     *bindContract
	GovMinter       *bindContract
	GovMasterMinter *bindContract
	MockFiatToken   *bindContract
}

func (c *compiledContractWBFT) Compile(root, openzeppelinPath string) {
	if contracts, err := compile.Compile(openzeppelinPath,
		filepath.Join(root, "GovValidator.sol"),
		filepath.Join(root, "NativeCoinAdapter.sol"),
		filepath.Join(root, "GovMinter.sol"),
		filepath.Join(root, "GovMasterMinter.sol"),
		filepath.Join(root, "../test", "MockFiatToken.sol"),
	); err != nil {
		panic(err)
	} else {
		if c.GovValidator, err = newBindContract(contracts["GovValidator"]); err != nil {
			panic(err)
		}
		if c.CoinAdapter, err = newBindContract(contracts["NativeCoinAdapter"]); err != nil {
			panic(err)
		}
		if c.GovMinter, err = newBindContract(contracts["GovMinter"]); err != nil {
			panic(err)
		}
		if c.GovMasterMinter, err = newBindContract(contracts["GovMasterMinter"]); err != nil {
			panic(err)
		}
		if c.MockFiatToken, err = newBindContract(contracts["MockFiatToken"]); err != nil {
			panic(err)
		}
	}
}

type GovWBFT struct {
	backend         *simulated.WBFTBackend
	owner           *bind.TransactOpts
	govValidator    *bind.BoundContract
	coinAdapter     *bind.BoundContract
	govMinter       *bind.BoundContract
	govMasterMinter *bind.BoundContract
	mockFiatToken   *bind.BoundContract
}

type TestCandidate struct {
	Validator *EOA
	Operator  *EOA
}

func NewTestCandidate() *TestCandidate {
	return &TestCandidate{
		Validator: NewEOA(),
		Operator:  NewEOA(),
	}
}

func (s *TestCandidate) GetBLSSecretKey() (bls.SecretKey, error) {
	blsSecretKey, err := bls.DeriveFromECDSA(s.Validator.PrivateKey)
	if err != nil {
		return nil, err
	}
	return blsSecretKey, nil
}

func (s *TestCandidate) BLSSign(t *testing.T, msg []byte) bls.Signature {
	sk, err := s.GetBLSSecretKey()
	if err != nil {
		t.Errorf("failed to get bls secret key: %v", err)
		return nil
	}
	return sk.Sign(msg)
}

func (s *TestCandidate) GetBLSPublicKey(t *testing.T) bls.PublicKey {
	blsSecretKey, err := s.GetBLSSecretKey()
	if err != nil {
		t.Errorf("failed to get bls secret key: %v", err)
		return nil
	}
	return blsSecretKey.PublicKey()
}

func (s *TestCandidate) GetBLSPoPSignature(t *testing.T) bls.Signature {
	pk := s.GetBLSPublicKey(t)
	return s.BLSSign(t, pk.Marshal())
}

var defaultBlockPeriod time.Duration

func NewGovWBFT(t *testing.T, alloc types.GenesisAlloc, validatorOption, adapterOption, minterOption, masterMinterOption, fiatTokenOption func(*params.SystemContract)) (*GovWBFT, error) {
	owner := getTxOpt(t, "owner")

	if alloc == nil {
		alloc = make(types.GenesisAlloc)
	}
	alloc[owner.From] = types.Account{Balance: MAX_UINT_128}

	// Deploy MockFiatToken at genesis BEFORE backend creation (NOT a system contract, just test helper)
	if fiatTokenOption != nil {
		// Get compiled MockFiatToken runtime bytecode (NOT creation bytecode)
		mockFiatTokenBytecode := compiledWBFT.MockFiatToken.RuntimeBin
		alloc[TestMockFiatTokenAddress] = types.Account{
			Balance: big.NewInt(0),
			Code:    mockFiatTokenBytecode,
		}
	}

	g := &GovWBFT{
		owner: owner,
		backend: simulated.NewWBFTBackend(alloc, func(nodeConf *node.Config, ethConf *ethconfig.Config) {
			anzeonConfig := ethConf.Genesis.Config.Anzeon
			defaultBlockPeriod = time.Duration(anzeonConfig.WBFT.BlockPeriodSeconds) * time.Second

			anzeonConfig.SystemContracts.GovValidator.Address = TestGovValidatorAddress
			if validatorOption != nil {
				validatorOption(anzeonConfig.SystemContracts.GovValidator)
			}

			anzeonConfig.SystemContracts.NativeCoinAdapter.Address = TestCoinAdapterAddress
			if adapterOption != nil {
				adapterOption(anzeonConfig.SystemContracts.NativeCoinAdapter)
			}

			if minterOption != nil {
				anzeonConfig.SystemContracts.GovMinter = &params.SystemContract{
					Address: TestGovMinterAddress,
					Version: "v1",
				}
				minterOption(anzeonConfig.SystemContracts.GovMinter)
			}

			if masterMinterOption != nil {
				anzeonConfig.SystemContracts.GovMasterMinter = &params.SystemContract{
					Address: TestGovMasterMinterAddress,
					Version: "v1",
				}
				masterMinterOption(anzeonConfig.SystemContracts.GovMasterMinter)
			}
		}),
	}

	g.govValidator = compiledWBFT.GovValidator.New(g.backend.Client(), TestGovValidatorAddress)
	g.coinAdapter = compiledWBFT.CoinAdapter.New(g.backend.Client(), TestCoinAdapterAddress)

	if minterOption != nil {
		g.govMinter = compiledWBFT.GovMinter.New(g.backend.Client(), TestGovMinterAddress)
	}
	if masterMinterOption != nil {
		g.govMasterMinter = compiledWBFT.GovMasterMinter.New(g.backend.Client(), TestGovMasterMinterAddress)
	}
	if fiatTokenOption != nil {
		g.mockFiatToken = compiledWBFT.MockFiatToken.New(g.backend.Client(), TestMockFiatTokenAddress)
	}

	return g, nil
}

func (g *GovWBFT) ExpectedOk(tx *types.Transaction, txErr error) (*types.Receipt, error) {
	receipt, err := commitTx(g.backend, tx, txErr)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (g *GovWBFT) ExpectedFail(tx *types.Transaction, txErr error) error {
	_, err := commitTx(g.backend, tx, txErr)
	if err != nil {
		return err
	}
	return nil
}

// GovBase Contract
func (g *GovWBFT) BaseQuorum(contract *bind.BoundContract, sender *EOA) (uint32, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "quorum")
	if err != nil {
		return 0, err
	}
	return result[0].(uint32), nil
}

func (g *GovWBFT) BaseProposalExpiry(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "proposalExpiry")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseMembers(contract *bind.BoundContract, sender *EOA, member common.Address) (sc.Member, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "members", member)
	if err != nil {
		return sc.Member{}, err
	}
	resultMember := sc.Member{
		IsActive: *abi.ConvertType(result[0], new(bool)).(*bool),
		JoinedAt: *abi.ConvertType(result[1], new(uint32)).(*uint32),
	}

	return resultMember, nil
}

func (g *GovWBFT) BaseVersionedMemberList(contract *bind.BoundContract, sender *EOA, version *big.Int, index *big.Int) (common.Address, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "versionedMemberList", version, index)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) BaseMemberVersion(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "memberVersion")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseCurrentProposalId(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "currentProposalId")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseGetProposal(contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (sc.Proposal, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, &result, "getProposal", proposalId)
	if err != nil {
		return sc.Proposal{}, err
	}

	// The result is a tuple, we need to manually parse the fields
	rawProposal := result[0].(struct {
		ActionType        [32]byte       `json:"actionType"`
		MemberVersion     *big.Int       `json:"memberVersion"`
		VotedBitmap       *big.Int       `json:"votedBitmap"`
		CreatedAt         *big.Int       `json:"createdAt"`
		ExecutedAt        *big.Int       `json:"executedAt"`
		Proposer          common.Address `json:"proposer"`
		RequiredApprovals uint32         `json:"requiredApprovals"`
		Approved          uint32         `json:"approved"`
		Rejected          uint32         `json:"rejected"`
		Status            uint8          `json:"status"` // Parse as uint8 first
		CallData          []byte         `json:"callData"`
	})

	proposal := sc.Proposal{
		ActionType:        rawProposal.ActionType,
		MemberVersion:     rawProposal.MemberVersion,
		VotedBitmap:       rawProposal.VotedBitmap,
		CreatedAt:         rawProposal.CreatedAt,
		ExecutedAt:        rawProposal.ExecutedAt,
		Proposer:          rawProposal.Proposer,
		RequiredApprovals: rawProposal.RequiredApprovals,
		Approved:          rawProposal.Approved,
		Rejected:          rawProposal.Rejected,
		Status:            sc.ProposalStatus(rawProposal.Status), // Convert uint8 to ProposalStatus
		CallData:          rawProposal.CallData,
	}

	return proposal, nil
}

func (g *GovWBFT) BaseTxProposeAddMember(t *testing.T, contract *bind.BoundContract, sender *EOA, newMember common.Address, newQuorum uint32) (*big.Int, *types.Transaction, error) {
	// Get current proposal ID before transaction
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address}, &result, "currentProposalId")
	if err != nil {
		return nil, nil, err
	}
	currentId := result[0].(*big.Int)
	nextId := new(big.Int).Add(currentId, big.NewInt(1))

	tx, err := contract.Transact(NewTxOptsWithValue(t, sender, nil), "proposeAddMember", newMember, newQuorum)
	return nextId, tx, err
}

func (g *GovWBFT) BaseTxProposeChangeQuorum(t *testing.T, contract *bind.BoundContract, sender *EOA, newQuorum uint32) (*big.Int, *types.Transaction, error) {
	// Get current proposal ID before transaction
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address}, &result, "currentProposalId")
	if err != nil {
		return nil, nil, err
	}
	currentId := result[0].(*big.Int)
	nextId := new(big.Int).Add(currentId, big.NewInt(1))

	tx, err := contract.Transact(NewTxOptsWithValue(t, sender, nil), "proposeChangeQuorum", newQuorum)
	return nextId, tx, err
}

func (g *GovWBFT) BaseMaxActiveProposalsPerMember(contract *bind.BoundContract, sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address}, &result, "maxActiveProposalsPerMember")
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) BaseTxProposeChangeMaxProposals(t *testing.T, contract *bind.BoundContract, sender *EOA, newMax *big.Int) (*big.Int, *types.Transaction, error) {
	// Get current proposal ID before transaction
	var result []interface{}
	err := contract.Call(&bind.CallOpts{From: sender.Address}, &result, "currentProposalId")
	if err != nil {
		return nil, nil, err
	}
	currentId := result[0].(*big.Int)
	nextId := new(big.Int).Add(currentId, big.NewInt(1))

	tx, err := contract.Transact(NewTxOptsWithValue(t, sender, nil), "proposeChangeMaxProposals", newMax)
	return nextId, tx, err
}

func (g *GovWBFT) BaseTxApproveProposal(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "approveProposal", proposalId)
}

func (g *GovWBFT) BaseTxExecuteProposal(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "executeProposal", proposalId)
}

func (g *GovWBFT) BaseTxChangeMember(t *testing.T, contract *bind.BoundContract, sender *EOA, newMember common.Address) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "changeMember", newMember)
}

// GovValidator Contract
func (g *GovWBFT) ConfigureValidator(t *testing.T, v *TestCandidate) (*types.Transaction, error) {
	blsPubKey := v.GetBLSPublicKey(t)
	blsPoPSig := v.GetBLSPoPSignature(t)
	return g.validatorContractTx(t, "configureValidator", v.Operator, v.Validator.Address, blsPubKey.Marshal(), blsPoPSig.Marshal())
}

func (g *GovWBFT) ValidatorList(sender *EOA) ([]common.Address, error) {
	var result []interface{}
	err := g.validatorCall("validatorList", sender, &result)
	if err != nil {
		return nil, err
	}
	validators := make([]common.Address, len(result))
	for i, v := range result {
		validators[i] = v.(common.Address)
	}
	return validators, nil
}

func (g *GovWBFT) IsValidator(sender *EOA, addr common.Address) (bool, error) {
	var result []interface{}
	err := g.validatorCall("isValidator", sender, &result, addr)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

func (g *GovWBFT) ValidatorCount(sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := g.validatorCall("validatorCount", sender, &result)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) ValidatorToOperator(sender *EOA, addr common.Address) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("validatorToOperator", sender, &result, addr)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) OperatorToValidator(sender *EOA, addr common.Address) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("operatorToValidator", sender, &result, addr)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) ValidatorToBlsKey(sender *EOA, addr common.Address) ([]byte, error) {
	var result []interface{}
	err := g.validatorCall("validatorToBlsKey", sender, &result, addr)
	if err != nil {
		return nil, err
	}
	return result[0].([]byte), nil
}

func (g *GovWBFT) BlsKeyToValidator(sender *EOA, blsKey []byte) (common.Address, error) {
	var result []interface{}
	err := g.validatorCall("blsKeyToValidator", sender, &result, blsKey)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) validatorContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.govValidator.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
}

func (g *GovWBFT) validatorCall(method string, sender *EOA, result *[]interface{}, params ...interface{}) error {
	return g.govValidator.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, result, method, params...)
}

// GovMinter Contract
func (g *GovWBFT) FiatToken(sender *EOA) (common.Address, error) {
	var result []interface{}
	err := g.minterCall("fiatToken", sender, &result)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) GetBurnBalance(sender *EOA, member common.Address) (*big.Int, error) {
	var result []interface{}
	err := g.minterCall("burnBalance", sender, &result, member)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) TxProposeMint(t *testing.T, sender *EOA, recipient common.Address, amount *big.Int) (*types.Transaction, error) {
	// Encode MintProof as bytes
	// MintProof: (address beneficiary, uint256 amount, uint256 timestamp, string depositId, string bankReference, string memo)
	// Note: Beneficiary validation is performed off-chain

	// Use recipient as beneficiary
	timestamp := big.NewInt(time.Now().Unix())
	// Include nanoseconds in depositId to ensure uniqueness (proposals may occur in same second)
	depositId := fmt.Sprintf("deposit-%s-%d", recipient.Hex()[:10], time.Now().UnixNano())
	bankReference := "bank-ref-001"
	memo := "test mint"

	proofData, err := abi.Arguments{
		{Type: mustParseType("address")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
	}.Pack(recipient, amount, timestamp, depositId, bankReference, memo)

	if err != nil {
		return nil, err
	}

	return g.minterContractTx(t, "proposeMint", sender, proofData)
}

func (g *GovWBFT) TxProposeBurn(t *testing.T, sender *EOA, from common.Address, amount *big.Int) (*types.Transaction, error) {
	// Encode BurnProof as bytes
	// BurnProof: (address from, uint256 amount, uint256 timestamp, string withdrawalId, string referenceId, string memo)
	// Note: from must be the sender's address (BurnFromMustBeProposer validation)
	// Note: proposeBurn is now payable, so amount must be sent as msg.value

	// Use sender.Address as the 'from' field to satisfy BurnFromMustBeProposer check
	actualFrom := sender.Address

	timestamp := big.NewInt(time.Now().Unix())
	// Include nanoseconds in withdrawalId to ensure uniqueness (proposals may occur in same second)
	withdrawalId := fmt.Sprintf("withdrawal-%s-%d", actualFrom.Hex()[:10], time.Now().UnixNano())
	referenceId := "ref-001"
	memo := "test burn"

	proofData, err := abi.Arguments{
		{Type: mustParseType("address")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
	}.Pack(actualFrom, amount, timestamp, withdrawalId, referenceId, memo)

	if err != nil {
		return nil, err
	}

	// Send amount as msg.value (proposeBurn is now payable)
	return g.govMinter.Transact(NewTxOptsWithValue(t, sender, amount), "proposeBurn", proofData)
}

func (g *GovWBFT) minterContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.govMinter.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
}

func (g *GovWBFT) minterCall(method string, sender *EOA, result *[]interface{}, params ...interface{}) error {
	return g.govMinter.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, result, method, params...)
}

// GovMasterMinter Contract
func (g *GovWBFT) MasterMinterFiatToken(sender *EOA) (common.Address, error) {
	var result []interface{}
	err := g.masterMinterCall("fiatToken", sender, &result)
	if err != nil {
		return common.Address{}, err
	}
	return result[0].(common.Address), nil
}

func (g *GovWBFT) MaxMinterAllowance(sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := g.masterMinterCall("maxMinterAllowance", sender, &result)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) GetIsMinter(sender *EOA, minter common.Address) (bool, error) {
	var result []interface{}
	err := g.masterMinterCall("getIsMinter", sender, &result, minter)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

func (g *GovWBFT) GetMinterAllowance(sender *EOA, minter common.Address) (*big.Int, error) {
	var result []interface{}
	err := g.masterMinterCall("getMinterAllowance", sender, &result, minter)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) TxProposeConfigureMinter(t *testing.T, sender *EOA, minter common.Address, allowance *big.Int) (*types.Transaction, error) {
	return g.masterMinterContractTx(t, "proposeConfigureMinter", sender, minter, allowance)
}

func (g *GovWBFT) TxProposeRemoveMinter(t *testing.T, sender *EOA, minter common.Address) (*types.Transaction, error) {
	return g.masterMinterContractTx(t, "proposeRemoveMinter", sender, minter)
}

func (g *GovWBFT) TxProposeUpdateMaxMinterAllowance(t *testing.T, sender *EOA, newMax *big.Int) (*types.Transaction, error) {
	return g.masterMinterContractTx(t, "proposeUpdateMaxMinterAllowance", sender, newMax)
}

func (g *GovWBFT) masterMinterContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.govMasterMinter.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
}

func (g *GovWBFT) masterMinterCall(method string, sender *EOA, result *[]interface{}, params ...interface{}) error {
	return g.govMasterMinter.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, result, method, params...)
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

// ========== MockFiatToken Helper Functions ==========

func (g *GovWBFT) GetMockFiatTokenBalance(sender *EOA, account common.Address) (*big.Int, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	var result []interface{}
	err := g.mockFiatTokenCall("balanceOf", sender, &result, account)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) GetMockFiatTokenTotalSupply(sender *EOA) (*big.Int, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	var result []interface{}
	err := g.mockFiatTokenCall("totalSupply", sender, &result)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) SetMockFiatTokenMintShouldFail(t *testing.T, sender *EOA, shouldFail bool) (*types.Transaction, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	return g.mockFiatTokenContractTx(t, "setFailMint", sender, shouldFail)
}

func (g *GovWBFT) SetMockFiatTokenBurnShouldFail(t *testing.T, sender *EOA, shouldFail bool) (*types.Transaction, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	return g.mockFiatTokenContractTx(t, "setFailBurn", sender, shouldFail)
}

func (g *GovWBFT) ConfigureMockFiatTokenMinter(t *testing.T, sender *EOA, minter common.Address, allowance *big.Int) (*types.Transaction, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	return g.mockFiatTokenContractTx(t, "configureMinter", sender, minter, allowance)
}

func (g *GovWBFT) GetMockFiatTokenMinterAllowance(sender *EOA, minter common.Address) (*big.Int, error) {
	if g.mockFiatToken == nil {
		return nil, fmt.Errorf("mockFiatToken not initialized")
	}
	var result []interface{}
	err := g.mockFiatTokenCall("minterAllowance", sender, &result, minter)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) mockFiatTokenContractTx(t *testing.T, method string, sender *EOA, params ...interface{}) (*types.Transaction, error) {
	return g.mockFiatToken.Transact(NewTxOptsWithValue(t, sender, nil), method, params...)
}

func (g *GovWBFT) mockFiatTokenCall(method string, sender *EOA, result *[]interface{}, params ...interface{}) error {
	return g.mockFiatToken.Call(&bind.CallOpts{From: sender.Address, Context: context.TODO()}, result, method, params...)
}

// ========== GovMinter State Getter Functions ==========

func (g *GovWBFT) GetReservedMintAmount(sender *EOA) (*big.Int, error) {
	var result []interface{}
	err := g.minterCall("reservedMintAmount", sender, &result)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

func (g *GovWBFT) GetMintProposalAmount(sender *EOA, proposalId *big.Int) (*big.Int, error) {
	var result []interface{}
	err := g.minterCall("mintProposalAmounts", sender, &result, proposalId)
	if err != nil {
		return nil, err
	}
	return result[0].(*big.Int), nil
}

// ========== GovBase Transaction Functions ==========

func (g *GovWBFT) BaseTxCancelProposal(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "cancelProposal", proposalId)
}

func (g *GovWBFT) BaseTxDisapproveProposal(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "disapproveProposal", proposalId)
}

func (g *GovWBFT) BaseTxExpireProposal(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "expireProposal", proposalId)
}

func (g *GovWBFT) BaseTxExecuteWithFailure(t *testing.T, contract *bind.BoundContract, sender *EOA, proposalId *big.Int) (*types.Transaction, error) {
	return contract.Transact(NewTxOptsWithValue(t, sender, nil), "executeWithFailure", proposalId)
}

// ========== Proof Generation Helper Functions ==========

// CreateMintProof creates ABI-encoded mint proof with custom parameters
func CreateMintProof(beneficiary common.Address, amount *big.Int, depositId, bankReference, memo string) ([]byte, error) {
	timestamp := big.NewInt(time.Now().Unix())

	return abi.Arguments{
		{Type: mustParseType("address")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
	}.Pack(beneficiary, amount, timestamp, depositId, bankReference, memo)
}

// CreateBurnProof creates ABI-encoded burn proof with custom parameters
func CreateBurnProof(from common.Address, amount *big.Int, withdrawalId, referenceId, memo string) ([]byte, error) {
	timestamp := big.NewInt(time.Now().Unix())

	return abi.Arguments{
		{Type: mustParseType("address")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
	}.Pack(from, amount, timestamp, withdrawalId, referenceId, memo)
}

// TxProposeMintWithProof proposes mint with raw proof data (for custom proof testing)
func (g *GovWBFT) TxProposeMintWithProof(t *testing.T, sender *EOA, proofData []byte) (*types.Transaction, error) {
	return g.minterContractTx(t, "proposeMint", sender, proofData)
}

// TxProposeBurnWithProof proposes burn with raw proof data (for custom proof testing)
// Note: amount must be extracted from proofData to send as msg.value for burn proposals
func (g *GovWBFT) TxProposeBurnWithProof(t *testing.T, sender *EOA, proofData []byte) (*types.Transaction, error) {
	// Parse proofData to extract amount (second field in BurnProof ABI encoding)
	// BurnProof: (address from, uint256 amount, uint256 timestamp, string withdrawalId, string referenceId, string memo)
	burnProofArgs := abi.Arguments{
		{Type: mustParseType("address")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("uint256")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
		{Type: mustParseType("string")},
	}

	values, err := burnProofArgs.Unpack(proofData)
	if err != nil {
		return nil, err
	}

	amount := values[1].(*big.Int) // Second field is amount

	// Send amount as msg.value (proposeBurn is payable)
	return g.govMinter.Transact(NewTxOptsWithValue(t, sender, amount), "proposeBurn", proofData)
}

// ========== Emergency Pause Helper Functions ==========

// TxProposePause proposes emergency pause
func (g *GovWBFT) TxProposePause(t *testing.T, sender *EOA) (*types.Transaction, error) {
	return g.minterContractTx(t, "proposePause", sender)
}

// TxProposeUnpause proposes emergency unpause
func (g *GovWBFT) TxProposeUnpause(t *testing.T, sender *EOA) (*types.Transaction, error) {
	return g.minterContractTx(t, "proposeUnpause", sender)
}

// IsEmergencyPaused returns whether contract is paused
func (g *GovWBFT) IsEmergencyPaused(sender *EOA) (bool, error) {
	var result []interface{}
	err := g.minterCall("emergencyPaused", sender, &result)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

// ========== State Query Helper Functions ==========

// IsDepositIdExecuted returns whether depositId has been permanently executed
func (g *GovWBFT) IsDepositIdExecuted(sender *EOA, depositId string) (bool, error) {
	var result []interface{}
	err := g.minterCall("executedDepositIds", sender, &result, depositId)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

// IsWithdrawalIdExecuted returns whether withdrawalId has been permanently executed
func (g *GovWBFT) IsWithdrawalIdExecuted(sender *EOA, withdrawalId string) (bool, error) {
	var result []interface{}
	err := g.minterCall("executedWithdrawalIds", sender, &result, withdrawalId)
	if err != nil {
		return false, err
	}
	return result[0].(bool), nil
}

// GetMaxActiveProposalsPerMember returns the maximum number of active proposals per member
func (g *GovWBFT) GetMaxActiveProposalsPerMember(sender *EOA) (*big.Int, error) {
	return g.BaseMaxActiveProposalsPerMember(g.govMinter, sender)
}

// ==================== Proposal Execution Workflow Helpers ====================

// ExecuteProposal - Generic proposal execution workflow with automatic approval
// This helper automates the entire approval workflow:
// 1. Gets proposal status and quorum requirement
// 2. Collects approvals from provided EOAs until quorum is reached
// 3. Executes the proposal automatically once approved
// Returns the execution receipt
func (g *GovWBFT) ExecuteProposal(
	t *testing.T,
	contract *bind.BoundContract,
	proposalId *big.Int,
	approvers []*EOA,
) (*types.Receipt, error) {
	// Get current proposal to check how many approvals it already has
	proposal, err := g.BaseGetProposal(contract, approvers[0], proposalId)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	// Calculate how many more approvals we need
	currentApprovals := proposal.Approved
	requiredApprovals := proposal.RequiredApprovals
	neededApprovals := int(requiredApprovals) - int(currentApprovals)

	if neededApprovals <= 0 {
		// Already has enough approvals, just execute
		tx, err := g.BaseTxExecuteProposal(t, contract, approvers[0], proposalId)
		receipt, err := g.ExpectedOk(tx, err)
		if err != nil {
			return nil, fmt.Errorf("execution failed: %w", err)
		}
		return receipt, nil
	}

	// Check we have enough approvers
	if len(approvers) < neededApprovals {
		return nil, fmt.Errorf("insufficient approvers: need %d more approvals but only %d approvers provided", neededApprovals, len(approvers))
	}

	// Collect approvals until we reach quorum
	for i := 0; i < neededApprovals; i++ {
		tx, err := g.BaseTxApproveProposal(t, contract, approvers[i], proposalId)
		_, err = g.ExpectedOk(tx, err)
		if err != nil {
			return nil, fmt.Errorf("approval %d/%d failed: %w", i+1, neededApprovals, err)
		}
	}

	// Verify proposal is now Approved or Executed (auto-execution on quorum)
	proposal, err = g.BaseGetProposal(contract, approvers[0], proposalId)
	if err != nil {
		return nil, fmt.Errorf("failed to verify approval status: %w", err)
	}

	if proposal.Status == sc.ProposalStatusExecuted {
		// Proposal was auto-executed when quorum was reached
		// Return the receipt from the last approval transaction
		return nil, nil // Receipt not available from approval, but proposal executed successfully
	}

	if proposal.Status != sc.ProposalStatusApproved {
		return nil, fmt.Errorf("proposal status is %v, expected Approved or Executed", proposal.Status)
	}

	// Execute the proposal (only if not auto-executed)
	tx, err := g.BaseTxExecuteProposal(t, contract, approvers[0], proposalId)
	receipt, err := g.ExpectedOk(tx, err)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return receipt, nil
}

// CompleteMintProposal - Full mint workflow helper
// This automates the entire mint workflow:
// 1. Creates a mint proposal
// 2. Gets the proposal ID
// 3. Collects approvals and executes via ExecuteProposal
// Returns the final execution receipt
func (g *GovWBFT) CompleteMintProposal(
	t *testing.T,
	proposer *EOA,
	recipient common.Address,
	amount *big.Int,
	approvers []*EOA,
) (*types.Receipt, error) {
	// Step 1: Create mint proposal
	tx, err := g.TxProposeMint(t, proposer, recipient, amount)
	_, err = g.ExpectedOk(tx, err)
	if err != nil {
		return nil, fmt.Errorf("propose mint failed: %w", err)
	}

	// Step 2: Get the proposal ID
	proposalId, err := g.BaseCurrentProposalId(g.govMinter, proposer)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal ID: %w", err)
	}

	// Step 3: Execute the proposal (approve + execute)
	receipt, err := g.ExecuteProposal(t, g.govMinter, proposalId, approvers)
	if err != nil {
		return nil, fmt.Errorf("execute proposal failed: %w", err)
	}

	return receipt, nil
}

// CompleteBurnProposal - Full burn workflow helper
// This automates the entire burn workflow:
// 1. Deposits tokens for burn
// 2. Creates a burn proposal
// 3. Gets the proposal ID
// 4. Collects approvals and executes via ExecuteProposal
// Returns the final execution receipt
func (g *GovWBFT) CompleteBurnProposal(
	t *testing.T,
	proposer *EOA,
	amount *big.Int,
	approvers []*EOA,
) (*types.Receipt, error) {
	// Step 1: Create burn proposal (deposits native coins via msg.value)
	tx, err := g.TxProposeBurn(t, proposer, proposer.Address, amount)
	_, err = g.ExpectedOk(tx, err)
	if err != nil {
		return nil, fmt.Errorf("propose burn failed: %w", err)
	}

	// Step 2: Get the proposal ID
	proposalId, err := g.BaseCurrentProposalId(g.govMinter, proposer)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal ID: %w", err)
	}

	// Step 3: Execute the proposal (approve + execute)
	receipt, err := g.ExecuteProposal(t, g.govMinter, proposalId, approvers)
	if err != nil {
		return nil, fmt.Errorf("execute proposal failed: %w", err)
	}

	return receipt, nil
}

// ========== NativeCoinAdapter Contract ==========

func (g *GovWBFT) ConfigureMinter(t *testing.T, masterMinter *EOA, minter common.Address, allowedAmount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, masterMinter), "configureMinter", minter, allowedAmount)
}

func (g *GovWBFT) RemoveMinter(t *testing.T, masterMinter *EOA, minter common.Address) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, masterMinter), "removeMinter", minter)
}
func (g *GovWBFT) BalanceOf(t *testing.T, address common.Address) *big.Int {
	return contractCall(t, g.coinAdapter, "balanceOf", address)[0].(*big.Int)
}

func (g *GovWBFT) TotalSupply(t *testing.T) *big.Int {
	return contractCall(t, g.coinAdapter, "totalSupply")[0].(*big.Int)
}

func (g *GovWBFT) IsMinter(t *testing.T, minter common.Address) bool {
	return contractCall(t, g.coinAdapter, "isMinter", minter)[0].(bool)
}

func (g *GovWBFT) MinterAllowance(t *testing.T, minter common.Address) *big.Int {
	return contractCall(t, g.coinAdapter, "minterAllowance", minter)[0].(*big.Int)
}

func (g *GovWBFT) Mint(t *testing.T, minter *EOA, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, minter), "mint", to, amount)
}

func (g *GovWBFT) Burn(t *testing.T, minter *EOA, amount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, minter), "burn", amount)
}

func (g *GovWBFT) Transfer(t *testing.T, sender *EOA, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, sender), "transfer", to, amount)
}

func (g *GovWBFT) TransferFrom(t *testing.T, sender *EOA, from, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, sender), "transferFrom", from, to, amount)
}

func (g *GovWBFT) Allowance(t *testing.T, owner, spender common.Address) *big.Int {
	return contractCall(t, g.coinAdapter, "allowance", owner, spender)[0].(*big.Int)
}

func (g *GovWBFT) Approve(t *testing.T, owner *EOA, spender common.Address, amount *big.Int) (*types.Transaction, error) {
	return g.coinAdapter.Transact(NewTxOpts(t, owner), "approve", spender, amount)
}

// signatureArgs must be either signature(bytes) or v, r, s(uint8,bytes32,byte32)
func (g *GovWBFT) Permit(t *testing.T, sender *EOA, owner, spender common.Address, amount, deadline *big.Int, signatureArgs ...interface{},
) (*types.Transaction, error) {
	// validate signature args
	require.NoError(t, CheckSignatureArgs(signatureArgs...))

	if deadline == nil {
		deadline = MAX_UINT_256
	}
	params := append([]interface{}{owner, spender, amount, deadline}, signatureArgs...)

	// Use fallback to handle ABI name mismatch from overloaded functions.
	tx, err := g.coinAdapter.Transact(NewTxOpts(t, sender), "permit", params...)
	if err != nil && strings.Contains(err.Error(), "argument count mismatch") {
		return g.coinAdapter.Transact(NewTxOpts(t, sender), "permit0", params...)
	}
	return tx, err
}

// signatureArgs must be either signature(bytes) or v, r, s(uint8,bytes32,byte32)
func (g *GovWBFT) TransferWithAuthorization(
	t *testing.T, sender *EOA, from, to common.Address, amount, validAfter, validBefore *big.Int, nonce common.Hash, signatureArgs ...interface{},
) (*types.Transaction, error) {
	// validate signature args
	require.NoError(t, CheckSignatureArgs(signatureArgs...))

	if validAfter == nil {
		validAfter = common.Big0
	}
	if validBefore == nil {
		validBefore = MAX_UINT_256
	}
	params := append([]interface{}{from, to, amount, validAfter, validBefore, nonce}, signatureArgs...)

	// Use fallback to handle ABI name mismatch from overloaded functions.
	tx, err := g.coinAdapter.Transact(NewTxOpts(t, sender), "transferWithAuthorization", params...)
	if err != nil && strings.Contains(err.Error(), "argument count mismatch") {
		return g.coinAdapter.Transact(NewTxOpts(t, sender), "transferWithAuthorization0", params...)
	}
	return tx, err
}

// signatureArgs must be either signature(bytes) or v, r, s(uint8,bytes32,byte32)
func (g *GovWBFT) ReceiveWithAuthorization(
	t *testing.T, sender *EOA, from common.Address, amount, validAfter, validBefore *big.Int, nonce common.Hash, signatureArgs ...interface{},
) (*types.Transaction, error) {
	// validate signature args
	require.NoError(t, CheckSignatureArgs(signatureArgs...))

	if validAfter == nil {
		validAfter = common.Big0
	}
	if validBefore == nil {
		validBefore = MAX_UINT_256
	}
	params := append([]interface{}{from, sender.Address, amount, validAfter, validBefore, nonce}, signatureArgs...)

	// Use fallback to handle ABI name mismatch from overloaded functions.
	tx, err := g.coinAdapter.Transact(NewTxOpts(t, sender), "receiveWithAuthorization", params...)
	if err != nil && strings.Contains(err.Error(), "argument count mismatch") {
		return g.coinAdapter.Transact(NewTxOpts(t, sender), "receiveWithAuthorization0", params...)
	}
	return tx, err
}

// signatureArgs must be either signature(bytes) or v, r, s(uint8,bytes32,byte32)
func (g *GovWBFT) CancelAuthorization(
	t *testing.T, sender *EOA, authorizer common.Address, nonce common.Hash, signatureArgs ...interface{}) (*types.Transaction, error) {
	// validate signature args
	require.NoError(t, CheckSignatureArgs(signatureArgs...))

	params := append([]interface{}{authorizer, nonce}, signatureArgs...)

	// Use fallback to handle ABI name mismatch from overloaded functions.
	tx, err := g.coinAdapter.Transact(NewTxOpts(t, sender), "cancelAuthorization", params...)
	if err != nil && strings.Contains(err.Error(), "argument count mismatch") {
		return g.coinAdapter.Transact(NewTxOpts(t, sender), "cancelAuthorization0", params...)
	}
	return tx, err
}

func (g *GovWBFT) PermitNonce(t *testing.T, owner common.Address) *big.Int {
	return contractCall(t, g.coinAdapter, "nonces", owner)[0].(*big.Int)
}

func (g *GovWBFT) DomainSeparator(t *testing.T) common.Hash {
	return common.Hash(contractCall(t, g.coinAdapter, "DOMAIN_SEPARATOR")[0].([32]byte))
}

func (g *GovWBFT) BuildPermitSig(t *testing.T, owner *EOA, spender common.Address, amount, deadline *big.Int) (sig []byte, r, s common.Hash, v uint8) {
	if deadline == nil {
		deadline = MAX_UINT_256
	}
	typeHash := contractCall(t, g.coinAdapter, "PERMIT_TYPEHASH")[0].([32]byte)
	permitHash := crypto.Keccak256Hash(concatBytes(
		typeHash[:],
		common.LeftPadBytes(owner.Address.Bytes(), 32),
		common.LeftPadBytes(spender.Bytes(), 32),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.LeftPadBytes(g.PermitNonce(t, owner.Address).Bytes(), 32),
		common.LeftPadBytes(deadline.Bytes(), 32),
	))
	return SignEIP712Hash(t, g.DomainSeparator(t), permitHash, owner)
}

func (g *GovWBFT) BuildTransferWithAuthSig(
	t *testing.T, from *EOA, to common.Address, amount, validAfter, validBefore *big.Int, nonce common.Hash,
) (sig []byte, r, s common.Hash, v uint8) {
	return g.buildAuthorizationSig(t, "TRANSFER_WITH_AUTHORIZATION_TYPEHASH", from, to, amount, validAfter, validBefore, nonce)
}

func (g *GovWBFT) BuildReceiveWithAuthSig(
	t *testing.T, from *EOA, to common.Address, amount, validAfter, validBefore *big.Int, nonce common.Hash,
) (sig []byte, r, s common.Hash, v uint8) {
	return g.buildAuthorizationSig(t, "RECEIVE_WITH_AUTHORIZATION_TYPEHASH", from, to, amount, validAfter, validBefore, nonce)
}

func (g *GovWBFT) buildAuthorizationSig(
	t *testing.T, methodType string, from *EOA, to common.Address, amount, validAfter, validBefore *big.Int, nonce common.Hash,
) (sig []byte, r, s common.Hash, v uint8) {
	if validAfter == nil {
		validAfter = common.Big0
	}
	if validBefore == nil {
		validBefore = MAX_UINT_256
	}
	typeHash := contractCall(t, g.coinAdapter, methodType)[0].([32]byte)
	permitHash := crypto.Keccak256Hash(concatBytes(
		typeHash[:],
		common.LeftPadBytes(from.Address.Bytes(), 32),
		common.LeftPadBytes(to.Bytes(), 32),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.LeftPadBytes(validAfter.Bytes(), 32),
		common.LeftPadBytes(validBefore.Bytes(), 32),
		nonce.Bytes(),
	))
	return SignEIP712Hash(t, g.DomainSeparator(t), permitHash, from)
}

func (g *GovWBFT) BuildCancelAuthSig(t *testing.T, authorizer *EOA, nonce common.Hash) (sig []byte, r, s common.Hash, v uint8) {
	typeHash := contractCall(t, g.coinAdapter, "CANCEL_AUTHORIZATION_TYPEHASH")[0].([32]byte)
	permitHash := crypto.Keccak256Hash(concatBytes(
		typeHash[:],
		common.LeftPadBytes(authorizer.Address.Bytes(), 32),
		nonce.Bytes(),
	))
	return SignEIP712Hash(t, g.DomainSeparator(t), permitHash, authorizer)
}

// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package systemcontracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// mockStateReader implements StateReader interface for testing
type mockStateReader struct {
	storage map[common.Address]map[common.Hash]common.Hash
}

func newMockStateReader() *mockStateReader {
	return &mockStateReader{
		storage: make(map[common.Address]map[common.Hash]common.Hash),
	}
}

func (m *mockStateReader) GetState(addr common.Address, key common.Hash) common.Hash {
	if addrStorage, ok := m.storage[addr]; ok {
		if value, ok := addrStorage[key]; ok {
			return value
		}
	}
	return common.Hash{}
}

func (m *mockStateReader) SetState(addr common.Address, key common.Hash, value common.Hash) {
	if _, ok := m.storage[addr]; !ok {
		m.storage[addr] = make(map[common.Hash]common.Hash)
	}
	m.storage[addr][key] = value
}

func TestInitializeMinter(t *testing.T) {
	govMinterAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	fiatTokenAddr := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	member1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	member2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	t.Run("initialize with fiatToken", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex() + "," + member2.Hex(),
			GOV_BASE_PARAM_QUORUM:         "2",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
			GOV_MINTER_PARAM_FIAT_TOKEN:   fiatTokenAddr.Hex(),
		}

		sp, err := initializeMinter(govMinterAddr, params)
		require.NoError(t, err)
		require.NotEmpty(t, sp)

		// Verify fiatToken is set
		foundFiatToken := false

		for _, param := range sp {
			if param.Address == govMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MINTER_fiatToken) {
				foundFiatToken = true
				require.Equal(t, common.BytesToHash(fiatTokenAddr.Bytes()), param.Value)
			}
		}

		require.True(t, foundFiatToken, "fiatToken should be initialized")
	})

	t.Run("initialize without fiatToken", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "1",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
		}

		sp, err := initializeMinter(govMinterAddr, params)
		require.NoError(t, err)
		require.NotEmpty(t, sp)

		// Verify fiatToken is NOT set
		for _, param := range sp {
			if param.Address == govMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MINTER_fiatToken) {
				t.Fatal("fiatToken should not be initialized")
			}
		}
	})

	t.Run("initialize with invalid fiatToken address", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "1",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
			GOV_MINTER_PARAM_FIAT_TOKEN:   "0x0000000000000000000000000000000000000000",
		}

		_, err := initializeMinter(govMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "fiatToken")
		require.Contains(t, err.Error(), "invalid address")
	})

	t.Run("initialize with base param error", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "invalid", // Invalid quorum
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
		}

		_, err := initializeMinter(govMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quorum")
	})
}

func TestGetReservedMintAmount(t *testing.T) {
	govMinterAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	reservedAmount := big.NewInt(5000000)

	t.Run("get existing reserved mint amount", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set reserved mint amount
		slot := common.HexToHash(SLOT_GOV_MINTER_reservedMintAmount)
		mockState.SetState(govMinterAddr, slot, common.BigToHash(reservedAmount))

		result := GetReservedMintAmount(govMinterAddr, mockState)
		require.Equal(t, reservedAmount, result)
	})

	t.Run("get zero reserved mint amount", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetReservedMintAmount(govMinterAddr, mockState)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestGetMintProposalAmount(t *testing.T) {
	govMinterAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	proposalId := big.NewInt(1)
	amount := big.NewInt(2000000)

	t.Run("get existing mint proposal amount", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set mint proposal amount
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MINTER_mintProposalAmounts), common.BigToHash(proposalId))
		mockState.SetState(govMinterAddr, key, common.BigToHash(amount))

		result := GetMintProposalAmount(govMinterAddr, mockState, proposalId)
		require.Equal(t, amount, result)
	})

	t.Run("get zero mint proposal amount", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetMintProposalAmount(govMinterAddr, mockState, proposalId)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestGetBurnBalance(t *testing.T) {
	govMinterAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	balance := big.NewInt(1000000)

	t.Run("get existing burn balance", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set burn balance
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MINTER_burnBalance), addr)
		mockState.SetState(govMinterAddr, key, common.BigToHash(balance))

		result := GetBurnBalance(govMinterAddr, mockState, addr)
		require.Equal(t, balance, result)
	})

	t.Run("get zero burn balance", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetBurnBalance(govMinterAddr, mockState, addr)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestGetRefundableBalance(t *testing.T) {
	govMinterAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	balance := big.NewInt(3000000)

	t.Run("get existing refundable balance", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set refundable balance
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MINTER_refundableBalance), addr)
		mockState.SetState(govMinterAddr, key, common.BigToHash(balance))

		result := GetRefundableBalance(govMinterAddr, mockState, addr)
		require.Equal(t, balance, result)
	})

	t.Run("get zero refundable balance", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetRefundableBalance(govMinterAddr, mockState, addr)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestMintProofStruct(t *testing.T) {
	t.Run("create MintProof", func(t *testing.T) {
		proof := MintProof{
			Beneficiary:   common.HexToAddress("0x1111111111111111111111111111111111111111"),
			Amount:        big.NewInt(1000000),
			Timestamp:     big.NewInt(1234567890),
			DepositId:     "DEPOSIT-001",
			BankReference: "BANK-REF-001",
			Memo:          "Test deposit",
		}

		require.NotNil(t, proof.Beneficiary)
		require.NotNil(t, proof.Amount)
		require.NotNil(t, proof.Timestamp)
		require.NotEmpty(t, proof.DepositId)
		require.NotEmpty(t, proof.BankReference)
		require.NotEmpty(t, proof.Memo)
	})
}

func TestBurnProofStruct(t *testing.T) {
	t.Run("create BurnProof", func(t *testing.T) {
		proof := BurnProof{
			From:         common.HexToAddress("0x1111111111111111111111111111111111111111"),
			Amount:       big.NewInt(500000),
			Timestamp:    big.NewInt(1234567890),
			WithdrawalId: "WITHDRAWAL-001",
			ReferenceId:  "REF-001",
			Memo:         "Test withdrawal",
		}

		require.NotNil(t, proof.From)
		require.NotNil(t, proof.Amount)
		require.NotNil(t, proof.Timestamp)
		require.NotEmpty(t, proof.WithdrawalId)
		require.NotEmpty(t, proof.ReferenceId)
		require.NotEmpty(t, proof.Memo)
	})
}

func TestBurnProposalDataStruct(t *testing.T) {
	t.Run("create BurnProposalData", func(t *testing.T) {
		data := BurnProposalData{
			Amount:    big.NewInt(750000),
			Requester: common.HexToAddress("0x2222222222222222222222222222222222222222"),
		}

		require.NotNil(t, data.Amount)
		require.NotNil(t, data.Requester)
	})
}

// --- upgradeMinter tests ---

func TestUpgradeMinter_EmptyParams(t *testing.T) {
	sp, err := upgradeMinter(common.HexToAddress("0x1003"), map[string]string{})
	require.NoError(t, err)
	require.Equal(t, 0, len(sp), "Empty params should produce no StateParams")
}

func TestUpgradeMinter_FiatTokenOnly(t *testing.T) {
	addr := common.HexToAddress("0x1003")
	sp, err := upgradeMinter(addr, map[string]string{
		"fiatToken": "0x0000000000000000000000000000000000002000",
	})
	require.NoError(t, err)

	foundFiatToken := false
	for _, p := range sp {
		if p.Key == common.HexToHash(SLOT_GOV_MINTER_fiatToken) {
			foundFiatToken = true
			require.Equal(t, common.BytesToHash(common.HexToAddress("0x2000").Bytes()), p.Value)
		}
	}
	require.True(t, foundFiatToken, "fiatToken should be present")

	// No base params -> no base slots
	for _, p := range sp {
		require.NotEqual(t, common.HexToHash(SLOT_GOV_BASE_quorum), p.Key,
			"quorum should NOT be present")
	}
}

func TestUpgradeMinter_InvalidFiatToken(t *testing.T) {
	_, err := upgradeMinter(common.HexToAddress("0x1003"), map[string]string{
		"fiatToken": "0x0000000000000000000000000000000000000000",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "fiatToken")
}

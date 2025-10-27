// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The stable-one Authors
// This file is part of the stable-one library.
//
// The stable-one library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The stable-one is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the stable-one library. If not, see <http://www.gnu.org/licenses/>.

package systemcontracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestInitializeMasterMinter(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	fiatTokenAddr := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	member1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	member2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	t.Run("initialize with fiatToken and default maxMinterAllowance", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex() + "," + member2.Hex(),
			GOV_BASE_PARAM_QUORUM:         "2",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
			GOV_MASTER_MINTER_PARAM_FIAT_TOKEN: fiatTokenAddr.Hex(),
		}

		sp, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.NoError(t, err)
		require.NotEmpty(t, sp)

		// Verify fiatToken is set
		foundFiatToken := false
		foundMaxMinterAllowance := false

		for _, param := range sp {
			if param.Address == govMasterMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MASTER_MINTER_fiatToken) {
				foundFiatToken = true
				require.Equal(t, common.BytesToHash(fiatTokenAddr.Bytes()), param.Value)
			}
			if param.Address == govMasterMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance) {
				foundMaxMinterAllowance = true
				require.Equal(t, common.BigToHash(DefaultMaxMinterAllowance), param.Value)
			}
		}

		require.True(t, foundFiatToken, "fiatToken should be initialized")
		require.True(t, foundMaxMinterAllowance, "maxMinterAllowance should be initialized with default value")
	})

	t.Run("initialize with custom maxMinterAllowance", func(t *testing.T) {
		customAllowance := "5000000000000000000000000000" // 5B tokens with 18 decimals
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:                      member1.Hex(),
			GOV_BASE_PARAM_QUORUM:                       "1",
			GOV_BASE_PARAM_EXPIRY:                       "604800",
			GOV_BASE_PARAM_MEMBER_VERSION:               "1",
			GOV_MASTER_MINTER_PARAM_FIAT_TOKEN:          fiatTokenAddr.Hex(),
			GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: customAllowance,
		}

		sp, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.NoError(t, err)
		require.NotEmpty(t, sp)

		// Verify custom maxMinterAllowance is set
		foundMaxMinterAllowance := false
		expectedAllowance, _ := new(big.Int).SetString(customAllowance, 10)

		for _, param := range sp {
			if param.Address == govMasterMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance) {
				foundMaxMinterAllowance = true
				require.Equal(t, common.BigToHash(expectedAllowance), param.Value)
			}
		}

		require.True(t, foundMaxMinterAllowance, "custom maxMinterAllowance should be initialized")
	})

	t.Run("initialize without fiatToken", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "1",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
		}

		sp, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.NoError(t, err)
		require.NotEmpty(t, sp)

		// Verify fiatToken is NOT set, but maxMinterAllowance should still be set
		foundFiatToken := false
		foundMaxMinterAllowance := false

		for _, param := range sp {
			if param.Address == govMasterMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MASTER_MINTER_fiatToken) {
				foundFiatToken = true
			}
			if param.Address == govMasterMinterAddr && param.Key == common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance) {
				foundMaxMinterAllowance = true
			}
		}

		require.False(t, foundFiatToken, "fiatToken should not be initialized")
		require.True(t, foundMaxMinterAllowance, "maxMinterAllowance should be initialized with default")
	})

	t.Run("initialize with invalid fiatToken address", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "1",
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
			GOV_MASTER_MINTER_PARAM_FIAT_TOKEN: "0x0000000000000000000000000000000000000000",
		}

		_, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "fiatToken")
		require.Contains(t, err.Error(), "invalid address")
	})

	t.Run("initialize with invalid maxMinterAllowance format", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:                      member1.Hex(),
			GOV_BASE_PARAM_QUORUM:                       "1",
			GOV_BASE_PARAM_EXPIRY:                       "604800",
			GOV_BASE_PARAM_MEMBER_VERSION:               "1",
			GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: "invalid_number",
		}

		_, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "maxMinterAllowance")
		require.Contains(t, err.Error(), "invalid number")
	})

	t.Run("initialize with zero maxMinterAllowance", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:                      member1.Hex(),
			GOV_BASE_PARAM_QUORUM:                       "1",
			GOV_BASE_PARAM_EXPIRY:                       "604800",
			GOV_BASE_PARAM_MEMBER_VERSION:               "1",
			GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: "0",
		}

		_, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "maxMinterAllowance")
		require.Contains(t, err.Error(), "must be positive")
	})

	t.Run("initialize with negative maxMinterAllowance", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:                      member1.Hex(),
			GOV_BASE_PARAM_QUORUM:                       "1",
			GOV_BASE_PARAM_EXPIRY:                       "604800",
			GOV_BASE_PARAM_MEMBER_VERSION:               "1",
			GOV_MASTER_MINTER_PARAM_MAX_MINTER_ALLOWANCE: "-1000",
		}

		_, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "maxMinterAllowance")
		require.Contains(t, err.Error(), "must be positive")
	})

	t.Run("initialize with base param error", func(t *testing.T) {
		params := map[string]string{
			GOV_BASE_PARAM_MEMBERS:        member1.Hex(),
			GOV_BASE_PARAM_QUORUM:         "invalid", // Invalid quorum
			GOV_BASE_PARAM_EXPIRY:         "604800",
			GOV_BASE_PARAM_MEMBER_VERSION: "1",
		}

		_, err := initializeMasterMinter(govMasterMinterAddr, params)
		require.Error(t, err)
		require.Contains(t, err.Error(), "quorum")
	})
}

func TestGetMinterAllowance(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	minter := common.HexToAddress("0x1111111111111111111111111111111111111111")
	allowance := big.NewInt(1000000000)

	t.Run("get existing allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter allowance
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MASTER_MINTER_minterAllowances), minter)
		mockState.SetState(govMasterMinterAddr, key, common.BigToHash(allowance))

		result := GetMinterAllowance(govMasterMinterAddr, mockState, minter)
		require.Equal(t, allowance, result)
	})

	t.Run("get zero allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetMinterAllowance(govMasterMinterAddr, mockState, minter)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestIsMinter(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	minter := common.HexToAddress("0x1111111111111111111111111111111111111111")

	t.Run("check existing minter", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter status
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MASTER_MINTER_isMinter), minter)
		mockState.SetState(govMasterMinterAddr, key, common.BigToHash(big.NewInt(1)))

		result := IsMinter(govMasterMinterAddr, mockState, minter)
		require.True(t, result)
	})

	t.Run("check non-existing minter", func(t *testing.T) {
		mockState := newMockStateReader()

		result := IsMinter(govMasterMinterAddr, mockState, minter)
		require.False(t, result)
	})

	t.Run("check explicitly false minter", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter status to false
		key := CalculateMappingSlot(common.HexToHash(SLOT_GOV_MASTER_MINTER_isMinter), minter)
		mockState.SetState(govMasterMinterAddr, key, common.BigToHash(big.NewInt(0)))

		result := IsMinter(govMasterMinterAddr, mockState, minter)
		require.False(t, result)
	})
}

func TestGetMinterCount(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")

	t.Run("get minter count", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter list length
		minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
		mockState.SetState(govMasterMinterAddr, minterListSlot, common.BigToHash(big.NewInt(5)))

		result := GetMinterCount(govMasterMinterAddr, mockState)
		require.Equal(t, uint64(5), result)
	})

	t.Run("get zero minter count", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetMinterCount(govMasterMinterAddr, mockState)
		require.Equal(t, uint64(0), result)
	})
}

func TestGetMinterAt(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	minter1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	minter2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	t.Run("get minter at index 0", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter at index 0
		minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
		key := CalculateDynamicSlot(minterListSlot, big.NewInt(0))
		mockState.SetState(govMasterMinterAddr, key, common.BytesToHash(minter1.Bytes()))

		result := GetMinterAt(govMasterMinterAddr, mockState, 0)
		require.Equal(t, minter1, result)
	})

	t.Run("get minter at index 1", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set minter at index 1
		minterListSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_minterList)
		key := CalculateDynamicSlot(minterListSlot, big.NewInt(1))
		mockState.SetState(govMasterMinterAddr, key, common.BytesToHash(minter2.Bytes()))

		result := GetMinterAt(govMasterMinterAddr, mockState, 1)
		require.Equal(t, minter2, result)
	})

	t.Run("get minter at non-existing index", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetMinterAt(govMasterMinterAddr, mockState, 999)
		require.Equal(t, common.Address{}, result)
	})
}

func TestGetTotalMinterAllowance(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")
	totalAllowance := big.NewInt(5000000000)

	t.Run("get total minter allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set total minter allowance
		totalMinterAllowanceSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_totalMinterAllowance)
		mockState.SetState(govMasterMinterAddr, totalMinterAllowanceSlot, common.BigToHash(totalAllowance))

		result := GetTotalMinterAllowance(govMasterMinterAddr, mockState)
		require.Equal(t, totalAllowance, result)
	})

	t.Run("get zero total allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetTotalMinterAllowance(govMasterMinterAddr, mockState)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})
}

func TestGetMaxMinterAllowance(t *testing.T) {
	govMasterMinterAddr := common.HexToAddress("0x9876543210987654321098765432109876543210")

	t.Run("get max minter allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		// Set max minter allowance
		maxMinterAllowanceSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance)
		mockState.SetState(govMasterMinterAddr, maxMinterAllowanceSlot, common.BigToHash(DefaultMaxMinterAllowance))

		result := GetMaxMinterAllowance(govMasterMinterAddr, mockState)
		require.Equal(t, DefaultMaxMinterAllowance, result)
	})

	t.Run("get zero max allowance", func(t *testing.T) {
		mockState := newMockStateReader()

		result := GetMaxMinterAllowance(govMasterMinterAddr, mockState)
		require.Equal(t, 0, result.Cmp(big.NewInt(0)))
	})

	t.Run("get custom max allowance", func(t *testing.T) {
		mockState := newMockStateReader()
		customAllowance := big.NewInt(999999999)

		// Set custom max minter allowance
		maxMinterAllowanceSlot := common.HexToHash(SLOT_GOV_MASTER_MINTER_maxMinterAllowance)
		mockState.SetState(govMasterMinterAddr, maxMinterAllowanceSlot, common.BigToHash(customAllowance))

		result := GetMaxMinterAllowance(govMasterMinterAddr, mockState)
		require.Equal(t, customAllowance, result)
	})
}

func TestDefaultMaxMinterAllowance(t *testing.T) {
	t.Run("verify default max minter allowance value", func(t *testing.T) {
		// Default should be 10B tokens (10000000000 * 10^18)
		expected := new(big.Int).Mul(
			big.NewInt(10000000000),
			new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		)

		require.Equal(t, expected, DefaultMaxMinterAllowance)
		require.Equal(t, "10000000000000000000000000000", DefaultMaxMinterAllowance.String())
	})
}

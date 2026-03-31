// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright 2025 The go-stablenet Authors

package systemcontracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUpgradeAddr = common.HexToAddress("0x1003")

// findStateParam finds a StateParam by Key in the result slice.
func findStateParam(sp []params.StateParam, key common.Hash) *params.StateParam {
	for i := range sp {
		if sp[i].Key == key {
			return &sp[i]
		}
	}
	return nil
}

// --- getContractCode tests ---

func TestGetContractCode_ValidVersions(t *testing.T) {
	tests := []struct {
		contractType string
		version      string
	}{
		{CONTRACT_GOV_VALIDATOR, "v1"},
		{CONTRACT_COIN_ADAPTER, "v1"},
		{CONTRACT_GOV_MINTER, "v1"},
		{CONTRACT_GOV_MINTER, "v2"},
		{CONTRACT_GOV_MASTER_MINTER, "v1"},
		{CONTRACT_GOV_COUNCIL, "v1"},
	}

	for _, tt := range tests {
		t.Run(tt.contractType+"/"+tt.version, func(t *testing.T) {
			code, err := getContractCode(tt.contractType, tt.version)
			require.NoError(t, err)
			assert.NotEmpty(t, code, "code should not be empty")
		})
	}
}

func TestGetContractCode_InvalidVersion(t *testing.T) {
	_, err := getContractCode(CONTRACT_GOV_MINTER, "v99")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version")
}

func TestGetContractCode_InvalidContractType(t *testing.T) {
	_, err := getContractCode("nonexistent", "v1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown contract type")
}

// --- GetSystemContractsTransition routing tests ---

func TestGetSystemContractsTransition_GenesisPath(t *testing.T) {
	// GetSystemContractsTransition -> initializeMinter (genesis path)
	sc := &params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: testUpgradeAddr,
			Version: "v1",
			Params: map[string]string{
				"quorum":        "1",
				"expiry":        "604800",
				"members":       "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
				"memberVersion": "1",
				"fiatToken":     "0x0000000000000000000000000000000000001000",
				"maxProposals":  "3",
			},
		},
	}

	st, err := GetSystemContractsTransition(sc, nil)
	require.NoError(t, err)
	require.NotNil(t, st)

	assert.Equal(t, 1, len(st.Codes), "Should have 1 Code (GovMinter)")
	assert.True(t, len(st.States) > 0, "Should have States from initializeMinter")

	maxProposalsParam := findStateParam(st.States, common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember))
	require.NotNil(t, maxProposalsParam)
	assert.Equal(t, common.BigToHash(big.NewInt(3)), maxProposalsParam.Value, "Genesis path should use default maxProposals=3")
}

// --- initializeCoinAdapter: partial Params upgrade scenario tests ---

// fullCoinAdapterParams returns a complete set of CoinAdapter params for genesis initialization.
func fullCoinAdapterParams() map[string]string {
	return map[string]string{
		"masterMinter":  "0x0000000000000000000000000000000000001002",
		"minters":       "0x0000000000000000000000000000000000001003",
		"minterAllowed": "10000000000000000000000000000",
		"name":          "WKRC",
		"symbol":        "WKRC",
		"decimals":      "18",
		"currency":      "KRW",
	}
}

func TestInitializeCoinAdapter_FullParams_V1(t *testing.T) {
	// v1 full initialization — all required params provided
	addr := common.HexToAddress("0x1000")
	sp, err := initializeCoinAdapter(addr, fullCoinAdapterParams(), nil)
	require.NoError(t, err)
	assert.True(t, len(sp) > 0, "Should produce StateParams for full initialization")

	// verify name slot
	nameSlots := EncodeBytesToSlots(common.HexToHash(SLOT_COIN_ADAPTER_NAME), []byte("WKRC"))
	for slot, expected := range nameSlots {
		p := findStateParam(sp, slot)
		require.NotNil(t, p, "name slot should exist")
		assert.Equal(t, expected, p.Value)
	}

	// verify decimals
	decParam := findStateParam(sp, common.HexToHash(SLOT_COIN_ADAPTER_DECIMALS))
	require.NotNil(t, decParam)
	assert.Equal(t, common.BytesToHash([]byte{18}), decParam.Value)
}

func TestInitializeCoinAdapter_PartialParams_FailsMissingRequired(t *testing.T) {
	// Upgrade scenario: want to change minterAllowed only,
	// but masterMinter is required so it errors
	addr := common.HexToAddress("0x1000")
	_, err := initializeCoinAdapter(addr, map[string]string{
		"minters":       "0x0000000000000000000000000000000000001003",
		"minterAllowed": "99999999999999999999999999999",
		// masterMinter missing -> error
	}, nil)
	assert.Error(t, err, "Should fail: masterMinter is required")
	assert.Contains(t, err.Error(), "masterMinter")
}

func TestInitializeCoinAdapter_PartialParams_FailsMissingName(t *testing.T) {
	// Even with masterMinter provided, missing name causes error
	addr := common.HexToAddress("0x1000")
	_, err := initializeCoinAdapter(addr, map[string]string{
		"masterMinter": "0x0000000000000000000000000000000000001002",
		// name, symbol, decimals, currency missing
	}, nil)
	assert.Error(t, err, "Should fail: name is required")
	assert.Contains(t, err.Error(), "name")
}

func TestInitializeCoinAdapter_HardcodedOverwrite(t *testing.T) {
	// coinManager and accountManager are always written with hardcoded values regardless of Params.
	// This applies to both initialization and upgrades — always overwritten with hardcoded constants.
	addr := common.HexToAddress("0x1000")
	sp, err := initializeCoinAdapter(addr, fullCoinAdapterParams(), nil)
	require.NoError(t, err)

	coinMgr := findStateParam(sp, common.HexToHash(SLOT_COIN_ADAPTER_COIN_MANAGER))
	require.NotNil(t, coinMgr)
	assert.Equal(t, common.BytesToHash(params.NativeCoinManagerAddress.Bytes()), coinMgr.Value,
		"coinManager is always hardcoded to NativeCoinManagerAddress")

	accMgr := findStateParam(sp, common.HexToHash(SLOT_COIN_ADAPTER_ACCOUNT_MANAGER))
	require.NotNil(t, accMgr)
	assert.Equal(t, common.BytesToHash(params.AccountManagerAddress.Bytes()), accMgr.Value,
		"accountManager is always hardcoded to AccountManagerAddress")
}

func TestInitializeCoinAdapter_TotalSupply_NilAlloc(t *testing.T) {
	// alloc=nil (runtime path) -> totalSupply is not written
	addr := common.HexToAddress("0x1000")
	sp, err := initializeCoinAdapter(addr, fullCoinAdapterParams(), nil)
	require.NoError(t, err)

	totalSupply := findStateParam(sp, common.HexToHash(SLOT_COIN_ADAPTER_TOTAL_SUPPLY))
	assert.Nil(t, totalSupply, "totalSupply should NOT be set when alloc is nil (runtime path)")
}

// --- GovMinter: initializeMinter P5 issue verification ---

func TestInitializeMinter_PartialParams_MaxProposalsOverwrite(t *testing.T) {
	// P5 issue: when partial Params are passed to initializeMinter,
	// maxProposals defaults to hardcoded value (3) if key is missing from Params
	addr := common.HexToAddress("0x1003")
	sp, err := initializeMinter(addr, map[string]string{
		"quorum":    "2",
		"expiry":    "604800",
		"fiatToken": "0x0000000000000000000000000000000000001000",
		// maxProposals missing -> initializeBase writes default value 3
	})
	require.NoError(t, err)

	maxProposals := findStateParam(sp, common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember))
	require.NotNil(t, maxProposals)
	assert.Equal(t, common.BigToHash(big.NewInt(3)), maxProposals.Value,
		"P5: initializeMinter always writes maxProposals=3 when key is missing from Params")
}

// --- initializeBase parameter behavior pattern tests ---

func TestInitializeBase_ParamPatterns(t *testing.T) {
	// initializeBase parameter behavior patterns:
	//   Pattern A (conditional write): quorum, expiry, members — not written if key missing
	//   Pattern B (unconditional write): maxProposals — written with default even if key missing
	addr := common.HexToAddress("0x1001")

	// empty param -> Pattern A not written, only Pattern B written
	sp, err := initializeBase(addr, map[string]string{})
	require.NoError(t, err)

	// Pattern A: not written
	assert.Nil(t, findStateParam(sp, common.HexToHash(SLOT_GOV_BASE_quorum)),
		"Pattern A: quorum should NOT be written when key is missing")
	assert.Nil(t, findStateParam(sp, common.HexToHash(SLOT_GOV_BASE_proposalExpiry)),
		"Pattern A: expiry should NOT be written when key is missing")

	// Pattern B: written with default value
	maxProposals := findStateParam(sp, common.HexToHash(SLOT_GOV_BASE_maxActiveProposalsPerMember))
	require.NotNil(t, maxProposals, "Pattern B: maxProposals is ALWAYS written")
	assert.Equal(t, common.BigToHash(big.NewInt(3)), maxProposals.Value,
		"Pattern B: maxProposals defaults to 3 when key is missing")
}

// =============================================================================
// Upgrade tests: Version != "v1" -> code-only deployment (no state changes).
// Hardfork upgrades only replace contract code. On-chain state is preserved.
// =============================================================================

func TestUpgrade_NonV1_ParamsIgnored(t *testing.T) {
	// Version="v2" + Params present -> Params are ignored, code only
	addr := common.HexToAddress("0x1003")
	sc := &params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: addr,
			Version: "v2",
			Params:  map[string]string{"quorum": "3"},
		},
	}

	st, err := GetSystemContractsTransition(sc, nil)
	require.NoError(t, err)
	require.NotNil(t, st)
	assert.Equal(t, 1, len(st.Codes), "Should have 1 Code (GovMinter v2)")
	assert.Equal(t, 0, len(st.States), "Non-v1 upgrade should produce no States even with Params")
}

func TestUpgrade_CodeOnly_ParamsNil(t *testing.T) {
	sc := &params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: common.HexToAddress("0x1003"),
			Version: "v2",
			// Params: nil -> code only
		},
	}

	st, err := GetSystemContractsTransition(sc, nil)
	require.NoError(t, err)
	require.NotNil(t, st)

	assert.Equal(t, 1, len(st.Codes), "Should have 1 Code")
	assert.Equal(t, 0, len(st.States), "Params=nil should produce no States")
}

// =============================================================================

func TestGetSystemContractsTransition_CodeOnlyPath(t *testing.T) {
	// Params=nil -> code only, no States (P4 behavior preserved)
	sc := &params.SystemContracts{
		GovMinter: &params.SystemContract{
			Address: testUpgradeAddr,
			Version: "v2",
			// Params: nil
		},
	}

	st, err := GetSystemContractsTransition(sc, nil)
	require.NoError(t, err)
	require.NotNil(t, st)

	assert.Equal(t, 1, len(st.Codes), "Should have 1 Code (GovMinter)")
	assert.Equal(t, 0, len(st.States), "Params=nil should produce no States")
}


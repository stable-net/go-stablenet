// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/systemcontracts"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/ethereum/go-ethereum/triedb/pathdb"
)

func TestInvalidCliqueConfig(t *testing.T) {
	block := DefaultGoerliGenesisBlock()
	block.ExtraData = []byte{}
	db := rawdb.NewMemoryDatabase()
	if _, err := block.Commit(db, triedb.NewDatabase(db, nil)); err == nil {
		t.Fatal("Expected error on invalid clique config")
	}
}

func TestSetupGenesis(t *testing.T) {
	testSetupGenesis(t, rawdb.HashScheme)
	testSetupGenesis(t, rawdb.PathScheme)
}

func TestLoadChainConfigWithOverride(t *testing.T) {
	t.Run("no block in DB, genesis with overrides", func(t *testing.T) {
		db := rawdb.NewMemoryDatabase()
		cancunTime := uint64(1234)

		config := *params.TestChainConfig
		genesis := &Genesis{
			Config:     &config,
			GasLimit:   4712388,
			Difficulty: big.NewInt(1),
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1)},
			},
		}
		got, err := LoadChainConfigWithOverride(db, genesis, &ChainOverrides{OverrideCancun: &cancunTime})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.CancunTime == nil || *got.CancunTime != cancunTime {
			t.Fatalf("override not applied, got=%v want=%d", got.CancunTime, cancunTime)
		}
	})

	t.Run("private network in DB, genesis nil uses stored config with overrides", func(t *testing.T) {
		db := rawdb.NewMemoryDatabase()
		verkleTime := uint64(5678)

		config := *params.TestChainConfig
		config.ChainID = big.NewInt(1337)
		genesis := &Genesis{
			Config:     &config,
			GasLimit:   4712388,
			Difficulty: big.NewInt(1),
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1)},
			},
		}
		block := genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults))

		got, err := LoadChainConfigWithOverride(db, nil, &ChainOverrides{OverrideVerkle: &verkleTime})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ChainID.Cmp(config.ChainID) != 0 {
			t.Fatalf("unexpected chain id, got=%v want=%v", got.ChainID, config.ChainID)
		}
		if got.VerkleTime == nil || *got.VerkleTime != verkleTime {
			t.Fatalf("override not applied, got=%v want=%d", got.VerkleTime, verkleTime)
		}
		storedcfg := rawdb.ReadChainConfig(db, block.Hash())
		if storedcfg == nil {
			t.Fatalf("stored chain config missing")
		}
		if storedcfg.ChainID.Cmp(got.ChainID) != 0 {
			t.Fatalf("unexpected stored chain id, got=%v want=%v", storedcfg.ChainID, got.ChainID)
		}
	})

	t.Run("stored genesis with mismatching provided genesis returns mismatch error", func(t *testing.T) {
		db := rawdb.NewMemoryDatabase()

		config := *params.TestChainConfig
		config.ChainID = big.NewInt(9999)
		storedGenesis := &Genesis{
			Config:     &config,
			GasLimit:   4712388,
			Difficulty: big.NewInt(1),
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1)},
			},
		}
		block := storedGenesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults))

		mismatchGenesis := &Genesis{
			Config:     &config,
			GasLimit:   4712388,
			Difficulty: big.NewInt(1),
			Nonce:      1, // force different genesis hash
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1)},
			},
		}

		_, err := LoadChainConfigWithOverride(db, mismatchGenesis, nil)
		if err == nil {
			t.Fatalf("expected mismatch error")
		}
		mismatchErr, ok := err.(*GenesisMismatchError)
		if !ok {
			t.Fatalf("unexpected error type: %T", err)
		}
		if mismatchErr.Stored != block.Hash() {
			t.Fatalf("unexpected stored hash, got=%s want=%s", mismatchErr.Stored, block.Hash())
		}
		if mismatchErr.New != mismatchGenesis.ToBlock().Hash() {
			t.Fatalf("unexpected new hash, got=%s want=%s", mismatchErr.New, mismatchGenesis.ToBlock().Hash())
		}
	})

	t.Run("genesis without config returns errGenesisNoConfig", func(t *testing.T) {
		db := rawdb.NewMemoryDatabase()
		got, err := LoadChainConfigWithOverride(db, new(Genesis), nil)
		if err != errGenesisNoConfig {
			t.Fatalf("unexpected error, got=%v want=%v", err, errGenesisNoConfig)
		}
		if got != params.AllEthashProtocolChanges {
			t.Fatalf("unexpected config return for error path")
		}
	})
}

func testSetupGenesis(t *testing.T, scheme string) {
	var (
		customghash = common.HexToHash("0x89c99d90b79719238d2645c7642f2c9295246e80775b38cfd162b696817fbd50")
		customg     = Genesis{
			Config: &params.ChainConfig{HomesteadBlock: big.NewInt(3)},
			Alloc: types.GenesisAlloc{
				{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
			},
		}
		oldcustomg = customg
	)
	oldcustomg.Config = &params.ChainConfig{HomesteadBlock: big.NewInt(2)}

	tests := []struct {
		name       string
		fn         func(ethdb.Database) (*params.ChainConfig, common.Hash, error)
		wantConfig *params.ChainConfig
		wantHash   common.Hash
		wantErr    error
	}{
		{
			name: "genesis without ChainConfig",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), new(Genesis))
			},
			wantErr:    errGenesisNoConfig,
			wantConfig: params.AllEthashProtocolChanges,
		},
		{
			name: "no block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), nil)
			},
			wantHash:   params.StableNetMainnetGenesisHash,
			wantConfig: params.StableNetMainnetChainConfig,
		},
		{
			name: "mainnet block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				genesis := DefaultStableNetMainnetGenesisBlock()
				initializeAnzeonGenesis(genesis)
				genesis.MustCommit(db, triedb.NewDatabase(db, newDbConfig(scheme)))
				return SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), nil)
			},
			wantHash:   params.StableNetMainnetGenesisHash,
			wantConfig: params.StableNetMainnetChainConfig,
		},
		{
			name: "custom block in DB, genesis == nil",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				customg.Commit(db, tdb)
				return SetupGenesisBlock(db, tdb, nil)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "custom block in DB, genesis == goerli",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				customg.Commit(db, tdb)
				return SetupGenesisBlock(db, tdb, DefaultGoerliGenesisBlock())
			},
			wantErr:    &GenesisMismatchError{Stored: customghash, New: params.GoerliGenesisHash},
			wantHash:   params.GoerliGenesisHash,
			wantConfig: params.GoerliChainConfig,
		},
		{
			name: "compatible config in DB",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				oldcustomg.Commit(db, tdb)
				return SetupGenesisBlock(db, tdb, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
		},
		{
			name: "incompatible config in DB",
			fn: func(db ethdb.Database) (*params.ChainConfig, common.Hash, error) {
				// Commit the 'old' genesis block with Homestead transition at #2.
				// Advance to block #4, past the homestead transition block of customg.
				tdb := triedb.NewDatabase(db, newDbConfig(scheme))
				oldcustomg.Commit(db, tdb)

				bc, _ := NewBlockChain(db, DefaultCacheConfigWithScheme(scheme), &oldcustomg, nil, ethash.NewFullFaker(), vm.Config{}, nil, nil)
				defer bc.Stop()

				_, blocks, _ := GenerateChainWithGenesis(&oldcustomg, ethash.NewFaker(), 4, nil)
				bc.InsertChain(blocks)

				// This should return a compatibility error.
				return SetupGenesisBlock(db, tdb, &customg)
			},
			wantHash:   customghash,
			wantConfig: customg.Config,
			wantErr: &params.ConfigCompatError{
				What:          "Homestead fork block",
				StoredBlock:   big.NewInt(2),
				NewBlock:      big.NewInt(3),
				RewindToBlock: 1,
			},
		},
	}

	for _, test := range tests {
		db := rawdb.NewMemoryDatabase()
		config, hash, err := test.fn(db)
		// Check the return values.
		if !reflect.DeepEqual(err, test.wantErr) {
			spew := spew.ConfigState{DisablePointerAddresses: true, DisableCapacities: true}
			t.Errorf("%s: returned error %#v, want %#v", test.name, spew.NewFormatter(err), spew.NewFormatter(test.wantErr))
		}
		if !reflect.DeepEqual(config, test.wantConfig) {
			t.Errorf("%s:\nreturned %v\nwant     %v", test.name, config, test.wantConfig)
		}
		if hash != test.wantHash {
			t.Errorf("%s: returned hash %s, want %s", test.name, hash.Hex(), test.wantHash.Hex())
		} else if err == nil {
			// Check database content.
			stored := rawdb.ReadBlock(db, test.wantHash, 0)
			if stored.Hash() != test.wantHash {
				t.Errorf("%s: block in DB has hash %s, want %s", test.name, stored.Hash(), test.wantHash)
			}
		}
	}
}

// TestGenesisHashes checks the congruity of default genesis data to
// corresponding hardcoded genesis hash values.
func TestGenesisHashes(t *testing.T) {
	for i, c := range []struct {
		genesis *Genesis
		want    common.Hash
	}{
		{DefaultGenesisBlock(), params.MainnetGenesisHash},
		{DefaultGoerliGenesisBlock(), params.GoerliGenesisHash},
		{DefaultSepoliaGenesisBlock(), params.SepoliaGenesisHash},
	} {
		// Test via MustCommit
		db := rawdb.NewMemoryDatabase()
		if have := c.genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults)).Hash(); have != c.want {
			t.Errorf("case: %d a), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
		// Test via ToBlock
		if have := c.genesis.ToBlock().Hash(); have != c.want {
			t.Errorf("case: %d a), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
	}
}

// TestStableNetGenesisHashes checks the congruity of default genesis data to
// corresponding hardcoded genesis hash values.
func TestStableNetGenesisHashes(t *testing.T) {
	for i, c := range []struct {
		genesis *Genesis
		want    common.Hash
	}{
		{DefaultStableNetMainnetGenesisBlock(), params.StableNetMainnetGenesisHash},
		{DefaultStableNetTestnetGenesisBlock(), params.StableNetTestnetGenesisHash},
	} {
		initializeAnzeonGenesis(c.genesis)

		// Test via MustCommit
		db := rawdb.NewMemoryDatabase()
		if have := c.genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults)).Hash(); have != c.want {
			t.Errorf("case: %d a), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
		// Test via ToBlock
		if have := c.genesis.ToBlock().Hash(); have != c.want {
			t.Errorf("case: %d a), want: %s, got: %s", i, c.want.Hex(), have.Hex())
		}
	}
}

func TestGenesis_Commit(t *testing.T) {
	genesis := &Genesis{
		BaseFee: big.NewInt(params.InitialBaseFee),
		Config:  params.TestChainConfig,
		// difficulty is nil
	}

	db := rawdb.NewMemoryDatabase()
	genesisBlock := genesis.MustCommit(db, triedb.NewDatabase(db, triedb.HashDefaults))

	if genesis.Difficulty != nil {
		t.Fatalf("assumption wrong")
	}

	// This value should have been set as default in the ToBlock method.
	if genesisBlock.Difficulty().Cmp(params.GenesisDifficulty) != 0 {
		t.Errorf("assumption wrong: want: %d, got: %v", params.GenesisDifficulty, genesisBlock.Difficulty())
	}

	// Expect the stored total difficulty to be the difficulty of the genesis block.
	stored := rawdb.ReadTd(db, genesisBlock.Hash(), genesisBlock.NumberU64())

	if stored.Cmp(genesisBlock.Difficulty()) != 0 {
		t.Errorf("inequal difficulty; stored: %v, genesisBlock: %v", stored, genesisBlock.Difficulty())
	}
}

func TestReadWriteGenesisAlloc(t *testing.T) {
	var (
		db    = rawdb.NewMemoryDatabase()
		alloc = &types.GenesisAlloc{
			{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
			{2}: {Balance: big.NewInt(2), Storage: map[common.Hash]common.Hash{{2}: {2}}},
		}
		hash, _ = hashAlloc(alloc, false)
	)
	blob, _ := json.Marshal(alloc)
	rawdb.WriteGenesisStateSpec(db, hash, blob)

	var reload types.GenesisAlloc
	err := reload.UnmarshalJSON(rawdb.ReadGenesisStateSpec(db, hash))
	if err != nil {
		t.Fatalf("Failed to load genesis state %v", err)
	}
	if len(reload) != len(*alloc) {
		t.Fatal("Unexpected genesis allocation")
	}
	for addr, account := range reload {
		want, ok := (*alloc)[addr]
		if !ok {
			t.Fatal("Account is not found")
		}
		if !reflect.DeepEqual(want, account) {
			t.Fatal("Unexpected account")
		}
	}
}

func newDbConfig(scheme string) *triedb.Config {
	if scheme == rawdb.HashScheme {
		return triedb.HashDefaults
	}
	return &triedb.Config{PathDB: pathdb.Defaults}
}

func TestVerkleGenesisCommit(t *testing.T) {
	var verkleTime uint64 = 0
	verkleConfig := &params.ChainConfig{
		ChainID:                       big.NewInt(1),
		HomesteadBlock:                big.NewInt(0),
		DAOForkBlock:                  nil,
		DAOForkSupport:                false,
		EIP150Block:                   big.NewInt(0),
		EIP155Block:                   big.NewInt(0),
		EIP158Block:                   big.NewInt(0),
		ByzantiumBlock:                big.NewInt(0),
		ConstantinopleBlock:           big.NewInt(0),
		PetersburgBlock:               big.NewInt(0),
		IstanbulBlock:                 big.NewInt(0),
		MuirGlacierBlock:              big.NewInt(0),
		BerlinBlock:                   big.NewInt(0),
		LondonBlock:                   big.NewInt(0),
		ArrowGlacierBlock:             big.NewInt(0),
		GrayGlacierBlock:              big.NewInt(0),
		MergeNetsplitBlock:            nil,
		ShanghaiTime:                  &verkleTime,
		CancunTime:                    &verkleTime,
		PragueTime:                    &verkleTime,
		VerkleTime:                    &verkleTime,
		TerminalTotalDifficulty:       big.NewInt(0),
		TerminalTotalDifficultyPassed: true,
		Ethash:                        nil,
		Clique:                        nil,
	}

	genesis := &Genesis{
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Config:     verkleConfig,
		Timestamp:  verkleTime,
		Difficulty: big.NewInt(0),
		Alloc: types.GenesisAlloc{
			{1}: {Balance: big.NewInt(1), Storage: map[common.Hash]common.Hash{{1}: {1}}},
		},
	}

	expected := common.Hex2Bytes("14398d42be3394ff8d50681816a4b7bf8d8283306f577faba2d5bc57498de23b")
	got := genesis.ToBlock().Root().Bytes()
	if !bytes.Equal(got, expected) {
		t.Fatalf("invalid genesis state root, expected %x, got %x", expected, got)
	}

	db := rawdb.NewMemoryDatabase()
	triedb := triedb.NewDatabase(db, &triedb.Config{IsVerkle: true, PathDB: pathdb.Defaults})
	block := genesis.MustCommit(db, triedb)
	if !bytes.Equal(block.Root().Bytes(), expected) {
		t.Fatalf("invalid genesis state root, expected %x, got %x", expected, got)
	}

	// Test that the trie is verkle
	if !triedb.IsVerkle() {
		t.Fatalf("expected trie to be verkle")
	}

	if !rawdb.ExistsAccountTrieNode(db, nil) {
		t.Fatal("could not find node")
	}
}

func TestGenerateStableNetGenesisJson(t *testing.T) {
	genesis := &Genesis{
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Config:     params.StableNetMainnetChainConfig,
		Timestamp:  0,
		Difficulty: big.NewInt(0),
		Alloc:      decodePrealloc(stablenetMainnetAllocData),
	}

	db := rawdb.NewMemoryDatabase()
	SetupGenesisBlock(db, triedb.NewDatabase(db, triedb.HashDefaults), genesis)
	genesisJson, _ := genesis.MarshalJSON()
	t.Logf("StableNet Mainnet genesis json: %s", genesisJson)

	genesis = &Genesis{
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Config:     params.StableNetTestnetChainConfig,
		Timestamp:  0,
		Difficulty: big.NewInt(0),
		Alloc:      decodePrealloc(stablenetTestnetAllocData),
	}
	db = rawdb.NewMemoryDatabase()
	SetupGenesisBlock(db, triedb.NewDatabase(db, triedb.HashDefaults), genesis)
	genesisJson, _ = genesis.MarshalJSON()
	t.Logf("StableNet Testnet genesis json: %s", genesisJson)
}

func TestSetGenesisBlockBaseFee(t *testing.T) {
	var tests = []struct {
		baseFee  *big.Int
		config   *params.ChainConfig
		expected *big.Int
	}{
		{
			baseFee:  big.NewInt(2 * params.InitialBaseFee),
			config:   params.TestChainConfig,
			expected: big.NewInt(2 * params.InitialBaseFee),
		},
		{
			baseFee:  nil,
			config:   params.TestChainConfig,
			expected: big.NewInt(params.InitialBaseFee),
		},
		{
			baseFee:  big.NewInt(params.InitialBaseFee),
			config:   params.StableNetMainnetChainConfig,
			expected: new(big.Int).SetUint64(params.MinBaseFee),
		},
		{
			baseFee:  nil,
			config:   params.StableNetTestnetChainConfig,
			expected: new(big.Int).SetUint64(params.MinBaseFee),
		},
	}

	for i, test := range tests {
		genesis := &Genesis{
			BaseFee: test.baseFee,
			Config:  test.config,
		}
		block := genesis.ToBlock()
		if block.BaseFee().Cmp(test.expected) != 0 {
			t.Fatalf("invalid genesis block BaseFee, test: %d expected %v got %v", i, test.expected, block.BaseFee())
		}
	}
}

// newTestAnzeonConfig creates a minimal Anzeon config with all 5 system contracts (v1).
func newTestAnzeonConfig() *params.AnzeonConfig {
	return &params.AnzeonConfig{
		SystemContracts: &params.SystemContracts{
			GovValidator: &params.SystemContract{
				Address: params.DefaultGovValidatorAddress,
				Version: "v1",
				Params: map[string]string{
					"quorum":        "1",
					"expiry":        "604800",
					"members":       "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
					"memberVersion": "1",
					"validators":    "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
					"blsPublicKeys": "0xaec493af8fa358a1c6f05499f2dd712721ade88c477d21b799d38e9b84582b6fbe4f4adc21e1e454bc37522eb3478b9b",
					"maxProposals":  "3",
					"gasTip":        "27600000000000",
				},
			},
			NativeCoinAdapter: &params.SystemContract{
				Address: params.DefaultNativeCoinAdapterAddress,
				Version: "v1",
				Params: map[string]string{
					"masterMinter":  "0x0000000000000000000000000000000000001002",
					"minters":       "0x0000000000000000000000000000000000001003",
					"minterAllowed": "10000000000000000000000000000",
					"name":          "WKRC",
					"symbol":        "WKRC",
					"decimals":      "18",
					"currency":      "KRW",
				},
			},
			GovMasterMinter: &params.SystemContract{
				Address: params.DefaultGovMasterMinterAddress,
				Version: "v1",
				Params: map[string]string{
					"quorum":             "1",
					"expiry":             "604800",
					"members":            "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
					"memberVersion":      "1",
					"fiatToken":          "0x0000000000000000000000000000000000001000",
					"minters":            "0x0000000000000000000000000000000000001003",
					"maxMinterAllowance": "10000000000000000000000000000",
					"maxProposals":       "3",
				},
			},
			GovMinter: &params.SystemContract{
				Address: params.DefaultGovMinterAddress,
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
			GovCouncil: &params.SystemContract{
				Address: params.DefaultGovCouncilAddress,
				Version: "v1",
				Params: map[string]string{
					"quorum":        "1",
					"expiry":        "604800",
					"members":       "0xaa5faa65e9cc0f74a85b6fdfb5f6991f5c094697",
					"memberVersion": "1",
					"maxProposals":  "3",
				},
			},
		},
	}
}

func govMinterV2Code() string {
	return systemcontracts.SystemContractCodes[systemcontracts.CONTRACT_GOV_MINTER][systemcontracts.SYSTEM_CONTRACT_VERSION_2]
}

func govMinterV1Code() string {
	return systemcontracts.SystemContractCodes[systemcontracts.CONTRACT_GOV_MINTER][systemcontracts.SYSTEM_CONTRACT_VERSION_1]
}

// TestInjectContracts_BohoBlock0 verifies that when BohoBlock=0,
// genesis alloc contains GovMinter v2 (not v1).
func TestInjectContracts_BohoBlock0(t *testing.T) {
	genesis := &Genesis{
		Alloc: make(types.GenesisAlloc),
		Config: &params.ChainConfig{
			ChainID:   big.NewInt(8282),
			BohoBlock: big.NewInt(0),
			Anzeon:    newTestAnzeonConfig(),
			Boho: &params.AnzeonConfig{
				SystemContracts: &params.SystemContracts{
					GovMinter: &params.SystemContract{
						Address: params.DefaultGovMinterAddress,
						Version: "v2",
					},
				},
			},
		},
	}

	err := InjectContracts(genesis, genesis.Config)
	if err != nil {
		t.Fatalf("InjectContracts failed: %v", err)
	}

	// GovMinter should have v2 code
	minterAccount, ok := genesis.Alloc[params.DefaultGovMinterAddress]
	if !ok {
		t.Fatal("GovMinter not found in genesis alloc")
	}

	expectedCode := govMinterV2Code()
	actualCode := common.Bytes2Hex(minterAccount.Code)
	v1Code := govMinterV1Code()

	if actualCode == common.Bytes2Hex(common.FromHex(v1Code)) {
		t.Error("GovMinter should be v2, but found v1 code")
	}
	if actualCode != common.Bytes2Hex(common.FromHex(expectedCode)) {
		t.Error("GovMinter code does not match v2")
	}

	// Other contracts should still be present
	if _, ok := genesis.Alloc[params.DefaultGovValidatorAddress]; !ok {
		t.Error("GovValidator should be in genesis alloc")
	}
	if _, ok := genesis.Alloc[params.DefaultGovCouncilAddress]; !ok {
		t.Error("GovCouncil should be in genesis alloc")
	}
}

// TestInjectContracts_BohoBlockN verifies that when BohoBlock=100,
// genesis alloc contains only GovMinter v1 (no overlay applied).
func TestInjectContracts_BohoBlockN(t *testing.T) {
	genesis := &Genesis{
		Alloc: make(types.GenesisAlloc),
		Config: &params.ChainConfig{
			ChainID:   big.NewInt(8282),
			BohoBlock: big.NewInt(100),
			Anzeon:    newTestAnzeonConfig(),
			Boho: &params.AnzeonConfig{
				SystemContracts: &params.SystemContracts{
					GovMinter: &params.SystemContract{
						Address: params.DefaultGovMinterAddress,
						Version: "v2",
					},
				},
			},
		},
	}

	err := InjectContracts(genesis, genesis.Config)
	if err != nil {
		t.Fatalf("InjectContracts failed: %v", err)
	}

	// GovMinter should have v1 code (no overlay at block 100)
	minterAccount, ok := genesis.Alloc[params.DefaultGovMinterAddress]
	if !ok {
		t.Fatal("GovMinter not found in genesis alloc")
	}

	v1Code := govMinterV1Code()
	actualCode := common.Bytes2Hex(minterAccount.Code)
	if actualCode != common.Bytes2Hex(common.FromHex(v1Code)) {
		t.Error("GovMinter should be v1 when BohoBlock > 0")
	}
}

// TestMultipleForksAtBlock0 verifies that when multiple hardforks are at block 0,
// overlays are applied in order and the final state is correct.
func TestMultipleForksAtBlock0(t *testing.T) {
	genesis := &Genesis{
		Alloc: make(types.GenesisAlloc),
		Config: &params.ChainConfig{
			ChainID:   big.NewInt(8282),
			BohoBlock: big.NewInt(0),
			Anzeon:    newTestAnzeonConfig(),
			Boho: &params.AnzeonConfig{
				SystemContracts: &params.SystemContracts{
					GovMinter: &params.SystemContract{
						Address: params.DefaultGovMinterAddress,
						Version: "v2",
					},
				},
			},
		},
	}

	err := InjectContracts(genesis, genesis.Config)
	if err != nil {
		t.Fatalf("InjectContracts failed: %v", err)
	}

	// GovMinter: v2 (from Boho overlay)
	minterAccount := genesis.Alloc[params.DefaultGovMinterAddress]
	expectedV2 := govMinterV2Code()
	if common.Bytes2Hex(minterAccount.Code) != common.Bytes2Hex(common.FromHex(expectedV2)) {
		t.Error("GovMinter should be v2 after Boho overlay at block 0")
	}

	// GovMinter storage should be preserved from Phase 1
	if minterAccount.Storage == nil {
		t.Error("GovMinter storage should be preserved after overlay")
	}
	if len(minterAccount.Storage) == 0 {
		t.Error("GovMinter storage should not be empty after overlay")
	}

	// GovValidator: should remain v1 (Boho doesn't touch it)
	validatorAccount := genesis.Alloc[params.DefaultGovValidatorAddress]
	v1ValidatorCode := systemcontracts.SystemContractCodes[systemcontracts.CONTRACT_GOV_VALIDATOR][systemcontracts.SYSTEM_CONTRACT_VERSION_1]
	if common.Bytes2Hex(validatorAccount.Code) != common.Bytes2Hex(common.FromHex(v1ValidatorCode)) {
		t.Error("GovValidator should remain v1 (Boho only upgrades GovMinter)")
	}
}

// TestSetupGenesis_MainnetWithAnzeonInit verifies that SetupGenesisBlock
// returns the correct config when mainnet genesis with Anzeon init is in the DB.
func TestSetupGenesis_MainnetWithAnzeonInit(t *testing.T) {
	for _, scheme := range []string{rawdb.HashScheme, rawdb.PathScheme} {
		t.Run(scheme, func(t *testing.T) {
			db := rawdb.NewMemoryDatabase()

			genesis := DefaultStableNetMainnetGenesisBlock()
			initializeAnzeonGenesis(genesis)
			genesis.MustCommit(db, triedb.NewDatabase(db, newDbConfig(scheme)))

			config, hash, err := SetupGenesisBlock(db, triedb.NewDatabase(db, newDbConfig(scheme)), nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hash != params.StableNetMainnetGenesisHash {
				t.Errorf("want hash %s, got %s", params.StableNetMainnetGenesisHash.Hex(), hash.Hex())
			}
			if config.ChainID.Cmp(params.StableNetMainnetChainConfig.ChainID) != 0 {
				t.Errorf("want chainID %v, got %v", params.StableNetMainnetChainConfig.ChainID, config.ChainID)
			}
		})
	}
}

// TestStableNetGenesisAllocConsistency verifies that the decodePrealloc-based
// genesis construction (with InjectContracts applied) produces the same alloc
// as the canonical JSON files (genesis_mainnet.json, genesis_testnet.json).
func TestStableNetGenesisAllocConsistency(t *testing.T) {
	tests := []struct {
		name     string
		jsonFile string
		genesis  func() *Genesis
	}{
		{
			name:     "mainnet",
			jsonFile: "genesis_mainnet.json",
			genesis:  DefaultStableNetMainnetGenesisBlock,
		},
		{
			name:     "testnet",
			jsonFile: "genesis_testnet.json",
			genesis:  DefaultStableNetTestnetGenesisBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Decode alloc from the canonical JSON file
			data, err := os.ReadFile(tt.jsonFile)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tt.jsonFile, err)
			}
			var jsonGenesis Genesis
			if err := json.Unmarshal(data, &jsonGenesis); err != nil {
				t.Fatalf("failed to decode genesis JSON: %v", err)
			}

			// Build alloc via decodePrealloc + InjectContracts (same path as SetupGenesisBlock)
			genesis := tt.genesis()
			if err := initializeAnzeonGenesis(genesis); err != nil {
				t.Fatalf("initializeAnzeonGenesis failed: %v", err)
			}
			builtAlloc := genesis.Alloc

			// Compare address sets
			if len(jsonGenesis.Alloc) != len(builtAlloc) {
				t.Errorf("alloc length mismatch: json=%d, built=%d", len(jsonGenesis.Alloc), len(builtAlloc))
				for addr := range jsonGenesis.Alloc {
					if _, ok := builtAlloc[addr]; !ok {
						t.Errorf("  json-only address: %s", addr.Hex())
					}
				}
				for addr := range builtAlloc {
					if _, ok := jsonGenesis.Alloc[addr]; !ok {
						t.Errorf("  built-only address: %s", addr.Hex())
					}
				}
				t.FailNow()
			}

			for addr, jsonAccount := range jsonGenesis.Alloc {
				builtAccount, ok := builtAlloc[addr]
				if !ok {
					t.Errorf("address %s present in JSON but missing from built alloc", addr.Hex())
					continue
				}

				if jsonAccount.Balance.Cmp(builtAccount.Balance) != 0 {
					t.Errorf("address %s balance mismatch: json=%s, built=%s",
						addr.Hex(), jsonAccount.Balance, builtAccount.Balance)
				}

				if !bytes.Equal(jsonAccount.Code, builtAccount.Code) {
					t.Errorf("address %s code mismatch: json=%d bytes, built=%d bytes",
						addr.Hex(), len(jsonAccount.Code), len(builtAccount.Code))
				}

				if len(jsonAccount.Storage) != len(builtAccount.Storage) {
					t.Errorf("address %s storage length mismatch: json=%d, built=%d",
						addr.Hex(), len(jsonAccount.Storage), len(builtAccount.Storage))
					continue
				}
				for key, jsonVal := range jsonAccount.Storage {
					if builtVal, exists := builtAccount.Storage[key]; !exists {
						t.Errorf("address %s storage key %s present in JSON but missing from built alloc",
							addr.Hex(), key.Hex())
					} else if jsonVal != builtVal {
						t.Errorf("address %s storage key %s value mismatch: json=%s, built=%s",
							addr.Hex(), key.Hex(), jsonVal.Hex(), builtVal.Hex())
					}
				}
			}
		})
	}
}

// TestInitializeAnzeonGenesis_InvalidExtraBit verifies that initializeAnzeonGenesis
// returns an error when an alloc entry contains an undefined Extra bit.
func TestInitializeAnzeonGenesis_InvalidExtraBit(t *testing.T) {
	addr := common.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	genesis := &Genesis{
		Alloc: types.GenesisAlloc{
			addr: {Balance: big.NewInt(0), Extra: 1}, // undefined bit
		},
		Config: params.TestWBFTChainConfig,
	}

	want := fmt.Sprintf("invalid account extra at %s: unknown bits set in account extra: 0x%016x", addr.Hex(), uint64(1))
	err := initializeAnzeonGenesis(genesis)
	if err == nil {
		t.Fatalf("want error %q, got nil", want)
	}
	if err.Error() != want {
		t.Fatalf("want %q, have %q", want, err.Error())
	}
}

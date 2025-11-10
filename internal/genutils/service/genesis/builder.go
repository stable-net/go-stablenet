package genesis

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/wbft"
	wbftcommon "github.com/ethereum/go-ethereum/consensus/wbft/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/params"
)

// GenesisBuilder provides functionality for building genesis files from GenTxCollection.
//
// The builder is responsible for:
//   - Extracting validators and BLS keys from GenTxCollection
//   - Creating ChainConfig with AnzeonConfig
//   - Configuring WBFT consensus parameters
//   - Setting up system contracts
//   - Generating genesis ExtraData
//   - Injecting system contract bytecode
//
// GenesisBuilder is NOT safe for concurrent use.
type GenesisBuilder struct {
	chainID         string
	timestamp       time.Time
	wbftConfig      *params.WBFTConfig
	systemContracts *params.SystemContracts
	initialAlloc    types.GenesisAlloc
}

// NewGenesisBuilder creates a new GenesisBuilder instance.
//
// Parameters:
//   - chainID: Chain identifier (e.g., "stable-testnet-1")
//   - timestamp: Genesis block timestamp
//
// Returns a new GenesisBuilder with default configurations.
func NewGenesisBuilder(chainID string, timestamp time.Time) *GenesisBuilder {
	return &GenesisBuilder{
		chainID:         chainID,
		timestamp:       timestamp,
		wbftConfig:      createDefaultWBFTConfig(),
		systemContracts: createDefaultSystemContracts(),
		initialAlloc:    make(types.GenesisAlloc),
	}
}

// WithWBFTConfig sets custom WBFT configuration.
//
// Parameters:
//   - config: Custom WBFT consensus configuration
//
// Returns the builder for method chaining.
func (b *GenesisBuilder) WithWBFTConfig(config *params.WBFTConfig) *GenesisBuilder {
	if config != nil {
		b.wbftConfig = config
	}
	return b
}

// WithSystemContracts sets custom system contracts configuration.
//
// Parameters:
//   - contracts: Custom system contracts configuration
//
// Returns the builder for method chaining.
func (b *GenesisBuilder) WithSystemContracts(contracts *params.SystemContracts) *GenesisBuilder {
	if contracts != nil {
		b.systemContracts = contracts
	}
	return b
}

// WithInitialAlloc sets initial account allocations.
//
// Parameters:
//   - alloc: Initial account balances and code
//
// Returns the builder for method chaining.
func (b *GenesisBuilder) WithInitialAlloc(alloc types.GenesisAlloc) *GenesisBuilder {
	if alloc != nil {
		b.initialAlloc = alloc
	}
	return b
}

// BuildFromCollection builds a genesis block from a GenTxCollection.
//
// The method performs the following operations:
//  1. Validates collection is not empty
//  2. Extracts validators and BLS keys from collection
//  3. Creates ChainConfig with AnzeonConfig
//  4. Generates WBFT ExtraData
//  5. Injects system contract bytecode
//  6. Builds final Genesis structure
//
// Parameters:
//   - collection: Collection of validated GenTx objects
//
// Returns:
//   - *core.Genesis: Complete genesis block ready for marshaling to JSON
//   - error: Error if collection is empty or build fails
//
// Example:
//
//	builder := NewGenesisBuilder("stable-testnet-1", time.Now().UTC())
//	genesis, err := builder.BuildFromCollection(collection)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Marshal to JSON
//	jsonBytes, _ := json.MarshalIndent(genesis, "", "  ")
//	os.WriteFile("genesis.json", jsonBytes, 0644)
func (b *GenesisBuilder) BuildFromCollection(collection *domain.GenTxCollection) (*core.Genesis, error) {
	// Validate collection
	if collection == nil || collection.Size() == 0 {
		return nil, errors.New("cannot build genesis from empty collection")
	}

	// Extract validators and BLS keys
	validators, blsKeys, err := b.extractValidatorsAndBLSKeys(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to extract validators: %w", err)
	}

	// Create ChainConfig
	chainConfig := b.createChainConfig(validators, blsKeys)

	// Create base genesis
	genesis := &core.Genesis{
		Config:     chainConfig,
		Timestamp:  uint64(b.timestamp.Unix()),
		GasLimit:   4700000,
		Difficulty: big.NewInt(524288),
		Alloc:      make(types.GenesisAlloc),
		Nonce:      wbftcommon.EmptyBlockNonce.Uint64(),
	}

	// Copy initial allocations
	for addr, account := range b.initialAlloc {
		genesis.Alloc[addr] = account
	}

	// Generate ExtraData for WBFT
	if genesis.Config.AnzeonEnabled() {
		extraData, err := wbft.CreateInitialExtraData(genesis.Config.Anzeon)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial extra data: %w", err)
		}
		genesis.ExtraData = extraData

		// Inject system contracts
		core.InjectContracts(genesis, genesis.Config)
	}

	return genesis, nil
}

// extractValidatorsAndBLSKeys extracts validator addresses and BLS public keys from collection.
//
// Returns validators and BLS keys in the same sorted order as the collection.
func (b *GenesisBuilder) extractValidatorsAndBLSKeys(collection *domain.GenTxCollection) ([]common.Address, []string, error) {
	gentxs := collection.GetAll()
	if len(gentxs) == 0 {
		return nil, nil, errors.New("collection is empty")
	}

	validators := make([]common.Address, len(gentxs))
	blsKeys := make([]string, len(gentxs))

	for i, gentx := range gentxs {
		// Convert domain.Address to common.Address
		validatorAddr := gentx.ValidatorAddress()
		validators[i] = common.HexToAddress(validatorAddr.String())
		blsKeys[i] = gentx.BLSPublicKey().String()
	}

	return validators, blsKeys, nil
}

// createChainConfig creates a ChainConfig with Anzeon configuration.
func (b *GenesisBuilder) createChainConfig(validators []common.Address, blsKeys []string) *params.ChainConfig {
	// Parse chain ID
	chainIDInt := new(big.Int)
	_, success := chainIDInt.SetString(b.chainID, 10)

	// If parsing fails, use hash of chain ID string
	if !success {
		// Hash the chain ID string to get a unique numeric ID
		hash := common.BytesToHash([]byte(b.chainID))
		chainIDInt = new(big.Int).SetBytes(hash.Bytes()[:8]) // Use first 8 bytes for reasonable size
	}

	return &params.ChainConfig{
		ChainID:             chainIDInt,
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   nil,
		GrayGlacierBlock:    nil,
		MergeNetsplitBlock:  nil,
		ShanghaiTime:        nil,
		CancunTime:          nil,
		PragueTime:          nil,
		VerkleTime:          nil,
		ApplepieBlock:       big.NewInt(0),
		Anzeon: &params.AnzeonConfig{
			WBFT: b.wbftConfig,
			Init: &params.WBFTInit{
				Validators:    validators,
				BLSPublicKeys: blsKeys,
			},
			SystemContracts: b.systemContracts,
		},
	}
}

// createDefaultWBFTConfig creates default WBFT consensus configuration.
func createDefaultWBFTConfig() *params.WBFTConfig {
	proposerPolicy := uint64(0)
	return &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}
}

// createDefaultSystemContracts creates default system contracts configuration.
func createDefaultSystemContracts() *params.SystemContracts {
	return &params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: params.DefaultGovValidatorAddress,
			Version: params.DefaultGovVersion,
			Params:  make(map[string]string),
		},
		NativeCoinAdapter: &params.SystemContract{
			Address: params.DefaultNativeCoinAdapterAddress,
			Version: params.DefaultNativeCoinAdapterVersion,
			Params:  make(map[string]string),
		},
		GovMasterMinter: &params.SystemContract{
			Address: params.DefaultGovMasterMinterAddress,
			Version: params.DefaultGovMasterMinterVersion,
			Params:  make(map[string]string),
		},
		GovMinter: &params.SystemContract{
			Address: params.DefaultGovMinterAddress,
			Version: params.DefaultGovMinterVersion,
			Params:  make(map[string]string),
		},
	}
}

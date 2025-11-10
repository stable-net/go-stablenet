package genesis_test

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/internal/genutils/domain"
	"github.com/ethereum/go-ethereum/internal/genutils/service/genesis"
	"github.com/ethereum/go-ethereum/params"
	ethcrypto "github.com/ethereum/go-ethereum/pkg/crypto/ethereum"
)

func TestNewGenesisBuilder(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()

	// Act
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	// Assert
	assert.NotNil(t, builder)
}

func TestGenesisBuilder_BuildFromCollection_EmptyCollection(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	collection := domain.NewGenTxCollection()

	// Act
	_, err := builder.BuildFromCollection(collection)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestGenesisBuilder_BuildFromCollection_SingleGenTx(t *testing.T) {
	// Arrange
	chainID := "8282" // Numeric chain ID
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	err := collection.Add(gentx)
	require.NoError(t, err)

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gen)

	// Verify ChainConfig
	require.NotNil(t, gen.Config)
	assert.Equal(t, chainID, gen.Config.ChainID.String())

	// Verify Anzeon config
	require.NotNil(t, gen.Config.Anzeon)
	require.NotNil(t, gen.Config.Anzeon.Init)

	// Verify validators
	assert.Equal(t, 1, len(gen.Config.Anzeon.Init.Validators))
	assert.Equal(t, common.HexToAddress(gentx.ValidatorAddress().String()), gen.Config.Anzeon.Init.Validators[0])

	// Verify BLS keys
	assert.Equal(t, 1, len(gen.Config.Anzeon.Init.BLSPublicKeys))
	expectedBLSKey := gentx.BLSPublicKey().String()
	assert.Equal(t, expectedBLSKey, gen.Config.Anzeon.Init.BLSPublicKeys[0])

	// Verify timestamp
	assert.Equal(t, uint64(timestamp.Unix()), gen.Timestamp)
}

func TestGenesisBuilder_BuildFromCollection_MultipleGenTxs(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()

	// Create multiple GenTxs
	gentx1 := createValidGenTx(t, cryptoProvider, chainID)
	gentx2 := createValidGenTx(t, cryptoProvider, chainID)
	gentx3 := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx1))
	require.NoError(t, collection.Add(gentx2))
	require.NoError(t, collection.Add(gentx3))

	// Get expected sorted order from collection
	expectedGenTxs := collection.GetAll()

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gen)

	// Verify all validators are included
	assert.Equal(t, 3, len(gen.Config.Anzeon.Init.Validators))
	assert.Equal(t, 3, len(gen.Config.Anzeon.Init.BLSPublicKeys))

	// Verify validators match expected order from collection
	for i := 0; i < len(expectedGenTxs); i++ {
		expectedAddr := common.HexToAddress(expectedGenTxs[i].ValidatorAddress().String())
		assert.Equal(t, expectedAddr, gen.Config.Anzeon.Init.Validators[i],
			"Validator at index %d should match collection order", i)
	}
}

func TestGenesisBuilder_WithWBFTConfig_DefaultConfig(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gen.Config.Anzeon.WBFT)

	// Verify default WBFT config
	assert.Equal(t, uint64(2), gen.Config.Anzeon.WBFT.RequestTimeoutSeconds)
	assert.Equal(t, uint64(1), gen.Config.Anzeon.WBFT.BlockPeriodSeconds)
	assert.Equal(t, uint64(10), gen.Config.Anzeon.WBFT.EpochLength)
	assert.NotNil(t, gen.Config.Anzeon.WBFT.ProposerPolicy)
	assert.Equal(t, uint64(0), *gen.Config.Anzeon.WBFT.ProposerPolicy)
}

func TestGenesisBuilder_WithWBFTConfig_CustomConfig(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Custom WBFT config
	proposerPolicy := uint64(1)
	customWBFT := &params.WBFTConfig{
		RequestTimeoutSeconds: 5,
		BlockPeriodSeconds:    2,
		EpochLength:           20,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	gen, err := builder.WithWBFTConfig(customWBFT).BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gen.Config.Anzeon.WBFT)

	// Verify custom WBFT config
	assert.Equal(t, uint64(5), gen.Config.Anzeon.WBFT.RequestTimeoutSeconds)
	assert.Equal(t, uint64(2), gen.Config.Anzeon.WBFT.BlockPeriodSeconds)
	assert.Equal(t, uint64(20), gen.Config.Anzeon.WBFT.EpochLength)
	assert.Equal(t, uint64(1), *gen.Config.Anzeon.WBFT.ProposerPolicy)
}

func TestGenesisBuilder_WithSystemContracts_DefaultContracts(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gen.Config.Anzeon.SystemContracts)

	// Verify default system contracts
	require.NotNil(t, gen.Config.Anzeon.SystemContracts.GovValidator)
	assert.Equal(t, params.DefaultGovValidatorAddress, gen.Config.Anzeon.SystemContracts.GovValidator.Address)
	assert.Equal(t, params.DefaultGovVersion, gen.Config.Anzeon.SystemContracts.GovValidator.Version)

	require.NotNil(t, gen.Config.Anzeon.SystemContracts.NativeCoinAdapter)
	assert.Equal(t, params.DefaultNativeCoinAdapterAddress, gen.Config.Anzeon.SystemContracts.NativeCoinAdapter.Address)

	require.NotNil(t, gen.Config.Anzeon.SystemContracts.GovMasterMinter)
	assert.Equal(t, params.DefaultGovMasterMinterAddress, gen.Config.Anzeon.SystemContracts.GovMasterMinter.Address)

	require.NotNil(t, gen.Config.Anzeon.SystemContracts.GovMinter)
	assert.Equal(t, params.DefaultGovMinterAddress, gen.Config.Anzeon.SystemContracts.GovMinter.Address)
}

func TestGenesisBuilder_WithSystemContracts_CustomContracts(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Custom system contracts
	customContracts := &params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: common.HexToAddress("0x2001"),
			Version: "v2",
			Params: map[string]string{
				"quorum": "2",
			},
		},
		NativeCoinAdapter: &params.SystemContract{
			Address: common.HexToAddress("0x2000"),
			Version: "v2",
		},
		GovMasterMinter: &params.SystemContract{
			Address: common.HexToAddress("0x2002"),
			Version: "v2",
		},
		GovMinter: &params.SystemContract{
			Address: common.HexToAddress("0x2003"),
			Version: "v2",
		},
	}

	// Act
	gen, err := builder.WithSystemContracts(customContracts).BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, gen.Config.Anzeon.SystemContracts)

	// Verify custom system contracts
	assert.Equal(t, common.HexToAddress("0x2001"), gen.Config.Anzeon.SystemContracts.GovValidator.Address)
	assert.Equal(t, "v2", gen.Config.Anzeon.SystemContracts.GovValidator.Version)
	assert.Equal(t, "2", gen.Config.Anzeon.SystemContracts.GovValidator.Params["quorum"])
}

func TestGenesisBuilder_WithInitialAlloc(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Initial allocations
	addr1 := common.HexToAddress("0x1234567890123456789012345678901234567890")
	addr2 := common.HexToAddress("0x2345678901234567890123456789012345678901")

	alloc := core.GenesisAlloc{
		addr1: {Balance: big.NewInt(1000000000000000000)}, // 1 ETH
		addr2: {Balance: big.NewInt(2000000000000000000)}, // 2 ETH
	}

	// Act
	gen, err := builder.WithInitialAlloc(alloc).BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, gen.Alloc)

	// Verify initial allocations are included
	assert.NotNil(t, gen.Alloc[addr1])
	assert.Equal(t, big.NewInt(1000000000000000000), gen.Alloc[addr1].Balance)

	assert.NotNil(t, gen.Alloc[addr2])
	assert.Equal(t, big.NewInt(2000000000000000000), gen.Alloc[addr2].Balance)
}

func TestGenesisBuilder_GenesisFields(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)

	// Verify genesis header fields
	assert.Equal(t, uint64(timestamp.Unix()), gen.Timestamp)
	assert.Equal(t, uint64(4700000), gen.GasLimit)
	assert.NotNil(t, gen.Difficulty)
	assert.NotNil(t, gen.ExtraData)
	assert.NotEmpty(t, gen.ExtraData)
	assert.NotNil(t, gen.Alloc)
}

func TestGenesisBuilder_JSONMarshaling(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)
	require.NoError(t, err)

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(gen, "", "  ")

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, jsonBytes)

	// Verify JSON can be unmarshaled
	var unmarshaledGen core.Genesis
	err = json.Unmarshal(jsonBytes, &unmarshaledGen)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, gen.Config.ChainID.String(), unmarshaledGen.Config.ChainID.String())
	assert.Equal(t, gen.Timestamp, unmarshaledGen.Timestamp)
}

func TestGenesisBuilder_ExtraDataGeneration(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)

	// ExtraData should be generated by wbft.CreateInitialExtraData()
	assert.NotNil(t, gen.ExtraData)
	assert.NotEmpty(t, gen.ExtraData)

	// ExtraData should contain validator information
	// (detailed verification would require decoding the WBFT extra data format)
	assert.True(t, len(gen.ExtraData) > 0)
}

func TestGenesisBuilder_SystemContractInjection(t *testing.T) {
	// Arrange
	chainID := "8282"
	timestamp := time.Now().UTC()
	builder := genesis.NewGenesisBuilder(chainID, timestamp)

	cryptoProvider := ethcrypto.NewEthereumProvider()
	gentx := createValidGenTx(t, cryptoProvider, chainID)

	collection := domain.NewGenTxCollection()
	require.NoError(t, collection.Add(gentx))

	// Act
	gen, err := builder.BuildFromCollection(collection)

	// Assert
	require.NoError(t, err)

	// Verify system contract addresses are in Alloc
	// (core.InjectContracts should populate Alloc with contract bytecode)
	assert.NotNil(t, gen.Alloc)

	// System contracts should be deployed at their addresses
	// Note: We can't verify exact bytecode without importing systemcontracts package
	// but we can verify the structure is correct
	assert.NotNil(t, gen.Config.Anzeon.SystemContracts)
}

// Helper function to create a valid GenTx
func createValidGenTx(t *testing.T, cryptoProvider domain.CryptoProvider, chainID string) domain.GenTx {
	t.Helper()

	privateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)

	// Derive validator address from private key
	tempMessage := []byte("temp")
	tempSig, err := cryptoProvider.Sign(privateKey, tempMessage)
	require.NoError(t, err)
	validatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, tempSig)
	require.NoError(t, err)

	// Generate unique operator address
	operatorPrivateKey, err := ethcrypto.GeneratePrivateKey()
	require.NoError(t, err)
	operatorSig, err := cryptoProvider.Sign(operatorPrivateKey, tempMessage)
	require.NoError(t, err)
	operatorAddr, err := cryptoProvider.RecoverAddress(tempMessage, operatorSig)
	require.NoError(t, err)

	metadata, _ := domain.NewValidatorMetadata("Test Validator", "Test Description", "https://test.com")

	// Derive BLS key
	blsKey, err := cryptoProvider.DeriveBLSPublicKey(privateKey)
	require.NoError(t, err)

	timestamp := time.Now().UTC().Add(-1 * time.Hour)

	// Create signature data
	sigData := domain.SignatureData{
		ValidatorAddress: validatorAddr,
		OperatorAddress:  operatorAddr,
		BLSPublicKey:     blsKey,
		ChainID:          chainID,
		Timestamp:        timestamp.Unix(),
	}

	// Sign the data
	message := sigData.Bytes()
	signature, err := cryptoProvider.Sign(privateKey, message)
	require.NoError(t, err)

	// Create GenTx
	gentx, err := domain.NewGenTx(validatorAddr, operatorAddr, blsKey, metadata, signature, chainID, timestamp)
	require.NoError(t, err)

	return gentx
}

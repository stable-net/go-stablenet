package genesis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/internal/genutils/service/genesis"
	"github.com/ethereum/go-ethereum/params"
)

func TestNewContractInjector(t *testing.T) {
	// Act
	injector := genesis.NewContractInjector()

	// Assert
	assert.NotNil(t, injector)
}

func TestContractInjector_ValidateSystemContracts_NilContracts(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	// Act
	err := injector.ValidateSystemContracts(nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestContractInjector_ValidateSystemContracts_MissingGovValidator(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator:      nil, // Missing
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GovValidator")
}

func TestContractInjector_ValidateSystemContracts_MissingNativeCoinAdapter(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: nil, // Missing
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NativeCoinAdapter")
}

func TestContractInjector_ValidateSystemContracts_MissingGovMasterMinter(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   nil, // Missing
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GovMasterMinter")
}

func TestContractInjector_ValidateSystemContracts_MissingGovMinter(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         nil, // Missing
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GovMinter")
}

func TestContractInjector_ValidateSystemContracts_ZeroAddress(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: common.Address{}, // Zero address
			Version: "v1",
		},
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "zero")
}

func TestContractInjector_ValidateSystemContracts_EmptyVersion(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: params.DefaultGovValidatorAddress,
			Version: "", // Empty version
		},
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestContractInjector_ValidateSystemContracts_ValidContracts(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	contracts := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	err := injector.ValidateSystemContracts(contracts)

	// Assert
	require.NoError(t, err)
}

func TestContractInjector_MergeSystemContracts_NilCustom(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	defaults := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	// Act
	merged := injector.MergeSystemContracts(nil, defaults)

	// Assert
	assert.Equal(t, defaults, merged)
}

func TestContractInjector_MergeSystemContracts_PartialCustom(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	defaults := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	custom := &params.SystemContracts{
		GovValidator: &params.SystemContract{
			Address: common.HexToAddress("0x2001"),
			Version: "v2",
		},
		NativeCoinAdapter: nil, // Missing - should use default
		GovMasterMinter:   nil, // Missing - should use default
		GovMinter:         nil, // Missing - should use default
	}

	// Act
	merged := injector.MergeSystemContracts(custom, defaults)

	// Assert
	assert.Equal(t, common.HexToAddress("0x2001"), merged.GovValidator.Address)
	assert.Equal(t, defaults.NativeCoinAdapter, merged.NativeCoinAdapter)
	assert.Equal(t, defaults.GovMasterMinter, merged.GovMasterMinter)
	assert.Equal(t, defaults.GovMinter, merged.GovMinter)
}

func TestContractInjector_MergeSystemContracts_FullCustom(t *testing.T) {
	// Arrange
	injector := genesis.NewContractInjector()

	defaults := &params.SystemContracts{
		GovValidator:      createValidContract(params.DefaultGovValidatorAddress),
		NativeCoinAdapter: createValidContract(params.DefaultNativeCoinAdapterAddress),
		GovMasterMinter:   createValidContract(params.DefaultGovMasterMinterAddress),
		GovMinter:         createValidContract(params.DefaultGovMinterAddress),
	}

	custom := &params.SystemContracts{
		GovValidator:      createValidContract(common.HexToAddress("0x2001")),
		NativeCoinAdapter: createValidContract(common.HexToAddress("0x2000")),
		GovMasterMinter:   createValidContract(common.HexToAddress("0x2002")),
		GovMinter:         createValidContract(common.HexToAddress("0x2003")),
	}

	// Act
	merged := injector.MergeSystemContracts(custom, defaults)

	// Assert
	assert.Equal(t, custom.GovValidator, merged.GovValidator)
	assert.Equal(t, custom.NativeCoinAdapter, merged.NativeCoinAdapter)
	assert.Equal(t, custom.GovMasterMinter, merged.GovMasterMinter)
	assert.Equal(t, custom.GovMinter, merged.GovMinter)
}

// Helper function to create a valid system contract
func createValidContract(address common.Address) *params.SystemContract {
	return &params.SystemContract{
		Address: address,
		Version: "v1",
		Params:  make(map[string]string),
	}
}

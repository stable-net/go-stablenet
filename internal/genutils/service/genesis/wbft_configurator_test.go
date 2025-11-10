package genesis_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/service/genesis"
	"github.com/ethereum/go-ethereum/params"
)

func TestNewWBFTConfigurator(t *testing.T) {
	// Act
	configurator := genesis.NewWBFTConfigurator()

	// Assert
	assert.NotNil(t, configurator)
}

func TestWBFTConfigurator_ValidateWBFTConfig_NilConfig(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	// Act
	err := configurator.ValidateWBFTConfig(nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestWBFTConfigurator_ValidateWBFTConfig_ZeroRequestTimeout(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(0)
	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 0, // Invalid
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RequestTimeoutSeconds")
}

func TestWBFTConfigurator_ValidateWBFTConfig_ZeroBlockPeriod(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(0)
	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    0, // Invalid
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "BlockPeriodSeconds")
}

func TestWBFTConfigurator_ValidateWBFTConfig_EpochLengthTooSmall(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(0)
	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           1, // Invalid (must be >= 2)
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "EpochLength")
}

func TestWBFTConfigurator_ValidateWBFTConfig_NilProposerPolicy(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        nil, // Invalid
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ProposerPolicy")
}

func TestWBFTConfigurator_ValidateWBFTConfig_InvalidProposerPolicy(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(2) // Invalid (must be 0 or 1)
	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ProposerPolicy")
}

func TestWBFTConfigurator_ValidateWBFTConfig_ValidConfig(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(0)
	config := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	err := configurator.ValidateWBFTConfig(config)

	// Assert
	require.NoError(t, err)
}

func TestWBFTConfigurator_MergeWBFTConfig_NilCustom(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	proposerPolicy := uint64(0)
	defaults := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}

	// Act
	merged := configurator.MergeWBFTConfig(nil, defaults)

	// Assert
	assert.Equal(t, defaults, merged)
}

func TestWBFTConfigurator_MergeWBFTConfig_PartialCustom(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	defaultProposerPolicy := uint64(0)
	defaults := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &defaultProposerPolicy,
	}

	customProposerPolicy := uint64(1)
	custom := &params.WBFTConfig{
		RequestTimeoutSeconds: 5,                     // Custom
		BlockPeriodSeconds:    0,                     // Should use default
		EpochLength:           0,                     // Should use default
		ProposerPolicy:        &customProposerPolicy, // Custom
	}

	// Act
	merged := configurator.MergeWBFTConfig(custom, defaults)

	// Assert
	assert.Equal(t, uint64(5), merged.RequestTimeoutSeconds)
	assert.Equal(t, uint64(1), merged.BlockPeriodSeconds) // From defaults
	assert.Equal(t, uint64(10), merged.EpochLength)       // From defaults
	assert.Equal(t, uint64(1), *merged.ProposerPolicy)
}

func TestWBFTConfigurator_MergeWBFTConfig_FullCustom(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	defaultProposerPolicy := uint64(0)
	defaults := &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &defaultProposerPolicy,
	}

	customProposerPolicy := uint64(1)
	custom := &params.WBFTConfig{
		RequestTimeoutSeconds: 5,
		BlockPeriodSeconds:    2,
		EpochLength:           20,
		ProposerPolicy:        &customProposerPolicy,
	}

	// Act
	merged := configurator.MergeWBFTConfig(custom, defaults)

	// Assert
	assert.Equal(t, custom.RequestTimeoutSeconds, merged.RequestTimeoutSeconds)
	assert.Equal(t, custom.BlockPeriodSeconds, merged.BlockPeriodSeconds)
	assert.Equal(t, custom.EpochLength, merged.EpochLength)
	assert.Equal(t, *custom.ProposerPolicy, *merged.ProposerPolicy)
}

func TestWBFTConfigurator_CreateDefaultWBFTConfig(t *testing.T) {
	// Arrange
	configurator := genesis.NewWBFTConfigurator()

	// Act
	config := configurator.CreateDefaultWBFTConfig()

	// Assert
	require.NotNil(t, config)
	assert.Equal(t, uint64(2), config.RequestTimeoutSeconds)
	assert.Equal(t, uint64(1), config.BlockPeriodSeconds)
	assert.Equal(t, uint64(10), config.EpochLength)
	require.NotNil(t, config.ProposerPolicy)
	assert.Equal(t, uint64(0), *config.ProposerPolicy)
}

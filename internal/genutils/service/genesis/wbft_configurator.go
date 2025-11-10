package genesis

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/params"
)

// WBFTConfigurator provides utilities for WBFT consensus configuration.
//
// This helper validates and prepares WBFT configurations for genesis blocks.
type WBFTConfigurator struct{}

// NewWBFTConfigurator creates a new WBFTConfigurator instance.
func NewWBFTConfigurator() *WBFTConfigurator {
	return &WBFTConfigurator{}
}

// ValidateWBFTConfig validates WBFT consensus configuration.
//
// Validates:
//   - RequestTimeoutSeconds > 0
//   - BlockPeriodSeconds > 0
//   - EpochLength >= 2
//   - ProposerPolicy is set
//
// Parameters:
//   - config: WBFT configuration to validate
//
// Returns:
//   - error: Validation error if configuration is invalid
func (wc *WBFTConfigurator) ValidateWBFTConfig(config *params.WBFTConfig) error {
	if config == nil {
		return errors.New("WBFT config is nil")
	}

	if config.RequestTimeoutSeconds == 0 {
		return errors.New("RequestTimeoutSeconds must be greater than 0")
	}

	if config.BlockPeriodSeconds == 0 {
		return errors.New("BlockPeriodSeconds must be greater than 0")
	}

	if config.EpochLength < 2 {
		return errors.New("EpochLength must be greater than or equal to 2")
	}

	if config.ProposerPolicy == nil {
		return errors.New("ProposerPolicy must be set")
	}

	// Validate ProposerPolicy value (0 = round-robin, 1 = sticky proposer)
	if *config.ProposerPolicy > 1 {
		return fmt.Errorf("ProposerPolicy must be 0 or 1, got %d", *config.ProposerPolicy)
	}

	return nil
}

// MergeWBFTConfig merges custom config with defaults.
//
// If custom config is provided, it overrides the defaults.
// Missing fields in custom config use default values.
//
// Parameters:
//   - custom: Custom WBFT config (may be partial)
//   - defaults: Default WBFT config
//
// Returns merged WBFT configuration.
func (wc *WBFTConfigurator) MergeWBFTConfig(custom, defaults *params.WBFTConfig) *params.WBFTConfig {
	if custom == nil {
		return defaults
	}

	merged := &params.WBFTConfig{
		RequestTimeoutSeconds:    custom.RequestTimeoutSeconds,
		BlockPeriodSeconds:       custom.BlockPeriodSeconds,
		EpochLength:              custom.EpochLength,
		ProposerPolicy:           custom.ProposerPolicy,
		AllowedFutureBlockTime:   custom.AllowedFutureBlockTime,
		MaxRequestTimeoutSeconds: custom.MaxRequestTimeoutSeconds,
	}

	// Use defaults for zero values
	if merged.RequestTimeoutSeconds == 0 && defaults != nil {
		merged.RequestTimeoutSeconds = defaults.RequestTimeoutSeconds
	}
	if merged.BlockPeriodSeconds == 0 && defaults != nil {
		merged.BlockPeriodSeconds = defaults.BlockPeriodSeconds
	}
	if merged.EpochLength == 0 && defaults != nil {
		merged.EpochLength = defaults.EpochLength
	}
	if merged.ProposerPolicy == nil && defaults != nil {
		merged.ProposerPolicy = defaults.ProposerPolicy
	}

	return merged
}

// CreateDefaultWBFTConfig creates a default WBFT configuration.
//
// Default values:
//   - RequestTimeoutSeconds: 2
//   - BlockPeriodSeconds: 1
//   - EpochLength: 10
//   - ProposerPolicy: 0 (round-robin)
//
// Returns default WBFT configuration.
func (wc *WBFTConfigurator) CreateDefaultWBFTConfig() *params.WBFTConfig {
	proposerPolicy := uint64(0)
	return &params.WBFTConfig{
		RequestTimeoutSeconds: 2,
		BlockPeriodSeconds:    1,
		EpochLength:           10,
		ProposerPolicy:        &proposerPolicy,
	}
}

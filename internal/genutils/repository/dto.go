package repository

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

// GenTxDTO is a Data Transfer Object for GenTx persistence.
// It contains public fields for JSON serialization/deserialization.
type GenTxDTO struct {
	ValidatorAddress string `json:"validator_address"`
	OperatorAddress  string `json:"operator_address"`
	BLSPublicKey     string `json:"bls_public_key"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Website          string `json:"website"`
	Signature        string `json:"signature"`
	ChainID          string `json:"chain_id"`
	Timestamp        int64  `json:"timestamp"` // Unix timestamp
}

// ToDTO converts a domain.GenTx to a GenTxDTO for persistence.
func ToDTO(gentx domain.GenTx) GenTxDTO {
	return GenTxDTO{
		ValidatorAddress: gentx.ValidatorAddress().String(),
		OperatorAddress:  gentx.OperatorAddress().String(),
		BLSPublicKey:     gentx.BLSPublicKey().String(),
		Name:             gentx.Metadata().Name(),
		Description:      gentx.Metadata().Description(),
		Website:          gentx.Metadata().Website(),
		Signature:        gentx.Signature().String(),
		ChainID:          gentx.ChainID(),
		Timestamp:        gentx.Timestamp().Unix(),
	}
}

// ToDomain converts a GenTxDTO back to a domain.GenTx.
// Returns an error if any validation fails during domain object reconstruction.
func (dto GenTxDTO) ToDomain() (domain.GenTx, error) {
	// Parse validator address
	validatorAddr, err := domain.NewAddress(dto.ValidatorAddress)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("invalid validator address: %w", err)
	}

	// Parse operator address
	operatorAddr, err := domain.NewAddress(dto.OperatorAddress)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("invalid operator address: %w", err)
	}

	// Parse BLS public key
	blsKey, err := domain.NewBLSPublicKey(dto.BLSPublicKey)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("invalid BLS public key: %w", err)
	}

	// Parse metadata
	metadata, err := domain.NewValidatorMetadata(dto.Name, dto.Description, dto.Website)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("invalid metadata: %w", err)
	}

	// Parse signature (hex string with 0x prefix)
	signature, err := parseSignature(dto.Signature)
	if err != nil {
		return domain.GenTx{}, fmt.Errorf("invalid signature: %w", err)
	}

	// Create GenTx
	return domain.NewGenTx(
		validatorAddr,
		operatorAddr,
		blsKey,
		metadata,
		signature,
		dto.ChainID,
		time.Unix(dto.Timestamp, 0).UTC(),
	)
}

// parseSignature converts a hex string signature (with 0x prefix) to domain.Signature.
func parseSignature(hexStr string) (domain.Signature, error) {
	// Remove 0x prefix if present
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		hexStr = hexStr[2:]
	}

	// Decode hex string
	sigBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return domain.Signature{}, fmt.Errorf("failed to decode hex: %w", err)
	}

	// Create signature
	return domain.NewSignature(sigBytes)
}

// MarshalJSON implements custom JSON marshaling for GenTxDTO.
func (dto GenTxDTO) MarshalJSON() ([]byte, error) {
	type Alias GenTxDTO
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&dto),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for GenTxDTO.
func (dto *GenTxDTO) UnmarshalJSON(data []byte) error {
	type Alias GenTxDTO
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(dto),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

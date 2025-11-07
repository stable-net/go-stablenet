package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/internal/genutils/domain"
)

func TestGenTxCreatedEvent_Creation(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")
	chainID := "stablenet-1"

	// Act
	event := domain.NewGenTxCreatedEvent(validatorAddr, operatorAddr, blsKey, chainID)

	// Assert
	assert.Equal(t, "GenTxCreated", event.EventType())
	assert.Equal(t, validatorAddr, event.ValidatorAddress)
	assert.Equal(t, operatorAddr, event.OperatorAddress)
	assert.Equal(t, blsKey, event.BLSPublicKey)
	assert.Equal(t, chainID, event.ChainID)
	assert.WithinDuration(t, time.Now().UTC(), event.OccurredAt(), time.Second)
}

func TestSignatureVerifiedEvent_Creation(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	sigBytes := make([]byte, 65)
	for i := 0; i < 65; i++ {
		sigBytes[i] = byte(i)
	}
	signature, _ := domain.NewSignature(sigBytes)

	// Act
	event := domain.NewSignatureVerifiedEvent(validatorAddr, signature)

	// Assert
	assert.Equal(t, "SignatureVerified", event.EventType())
	assert.Equal(t, validatorAddr, event.ValidatorAddress)
	assert.Equal(t, signature, event.Signature)
	assert.WithinDuration(t, time.Now().UTC(), event.OccurredAt(), time.Second)
}

func TestGenTxAddedToCollectionEvent_Creation(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	collectionSize := 5

	// Act
	event := domain.NewGenTxAddedToCollectionEvent(validatorAddr, collectionSize)

	// Assert
	assert.Equal(t, "GenTxAddedToCollection", event.EventType())
	assert.Equal(t, validatorAddr, event.ValidatorAddress)
	assert.Equal(t, collectionSize, event.CollectionSize)
	assert.WithinDuration(t, time.Now().UTC(), event.OccurredAt(), time.Second)
}

func TestCollectionValidatedEvent_Creation(t *testing.T) {
	// Arrange
	total := 10
	valid := 8
	invalid := 1
	duplicates := 1

	// Act
	event := domain.NewCollectionValidatedEvent(total, valid, invalid, duplicates)

	// Assert
	assert.Equal(t, "CollectionValidated", event.EventType())
	assert.Equal(t, total, event.TotalGenTxs)
	assert.Equal(t, valid, event.ValidGenTxs)
	assert.Equal(t, invalid, event.InvalidGenTxs)
	assert.Equal(t, duplicates, event.DuplicatesFound)
	assert.WithinDuration(t, time.Now().UTC(), event.OccurredAt(), time.Second)
}

func TestGenesisBuiltEvent_Creation(t *testing.T) {
	// Arrange
	chainID := "stablenet-1"
	validatorCount := 10
	filePath := "/path/to/genesis.json"

	// Act
	event := domain.NewGenesisBuiltEvent(chainID, validatorCount, filePath)

	// Assert
	assert.Equal(t, "GenesisBuilt", event.EventType())
	assert.Equal(t, chainID, event.ChainID)
	assert.Equal(t, validatorCount, event.ValidatorCount)
	assert.Equal(t, filePath, event.GenesisFilePath)
	assert.WithinDuration(t, time.Now().UTC(), event.OccurredAt(), time.Second)
}

func TestDomainEvent_Interface(t *testing.T) {
	// Arrange
	validatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	operatorAddr, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f01230")
	blsKey, _ := domain.NewBLSPublicKey("0x" +
		"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c" +
		"0d0b63eb6b6e4c6d0a1a2a3a4a5a6a7a8a9aaabacadaeafb0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c")

	// Act
	event := domain.NewGenTxCreatedEvent(validatorAddr, operatorAddr, blsKey, "stablenet-1")

	// Assert - verify it implements DomainEvent interface
	var domainEvent domain.DomainEvent = event
	require.NotNil(t, domainEvent)
	assert.NotEmpty(t, domainEvent.EventType())
	assert.NotZero(t, domainEvent.OccurredAt())
}

func TestBaseEvent_Timestamp(t *testing.T) {
	// Arrange
	before := time.Now().UTC()

	// Act
	event := domain.NewGenTxCreatedEvent(
		domain.Address{},
		domain.Address{},
		domain.BLSPublicKey{},
		"test-chain",
	)

	after := time.Now().UTC()

	// Assert
	occurredAt := event.OccurredAt()
	assert.True(t, occurredAt.After(before) || occurredAt.Equal(before))
	assert.True(t, occurredAt.Before(after) || occurredAt.Equal(after))
}

package domain

import (
	"time"
)

// DomainEvent represents a domain event interface.
// All domain events must implement this interface.
type DomainEvent interface {
	// OccurredAt returns the time when the event occurred.
	OccurredAt() time.Time
	// EventType returns the type identifier of the event.
	EventType() string
}

// BaseEvent provides common fields for all domain events.
type BaseEvent struct {
	occurredAt time.Time
}

// NewBaseEvent creates a new BaseEvent with the current timestamp.
func NewBaseEvent() BaseEvent {
	return BaseEvent{
		occurredAt: time.Now().UTC(),
	}
}

// OccurredAt returns the time when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// GenTxCreatedEvent is emitted when a new GenTx is created.
type GenTxCreatedEvent struct {
	BaseEvent
	ValidatorAddress Address
	OperatorAddress  Address
	BLSPublicKey     BLSPublicKey
	ChainID          string
}

// NewGenTxCreatedEvent creates a new GenTxCreatedEvent.
func NewGenTxCreatedEvent(validatorAddr, operatorAddr Address, blsKey BLSPublicKey, chainID string) GenTxCreatedEvent {
	return GenTxCreatedEvent{
		BaseEvent:        NewBaseEvent(),
		ValidatorAddress: validatorAddr,
		OperatorAddress:  operatorAddr,
		BLSPublicKey:     blsKey,
		ChainID:          chainID,
	}
}

// EventType returns the event type identifier.
func (e GenTxCreatedEvent) EventType() string {
	return "GenTxCreated"
}

// SignatureVerifiedEvent is emitted when a GenTx signature is successfully verified.
type SignatureVerifiedEvent struct {
	BaseEvent
	ValidatorAddress Address
	Signature        Signature
}

// NewSignatureVerifiedEvent creates a new SignatureVerifiedEvent.
func NewSignatureVerifiedEvent(validatorAddr Address, signature Signature) SignatureVerifiedEvent {
	return SignatureVerifiedEvent{
		BaseEvent:        NewBaseEvent(),
		ValidatorAddress: validatorAddr,
		Signature:        signature,
	}
}

// EventType returns the event type identifier.
func (e SignatureVerifiedEvent) EventType() string {
	return "SignatureVerified"
}

// GenTxAddedToCollectionEvent is emitted when a GenTx is added to a collection.
type GenTxAddedToCollectionEvent struct {
	BaseEvent
	ValidatorAddress Address
	CollectionSize   int
}

// NewGenTxAddedToCollectionEvent creates a new GenTxAddedToCollectionEvent.
func NewGenTxAddedToCollectionEvent(validatorAddr Address, collectionSize int) GenTxAddedToCollectionEvent {
	return GenTxAddedToCollectionEvent{
		BaseEvent:        NewBaseEvent(),
		ValidatorAddress: validatorAddr,
		CollectionSize:   collectionSize,
	}
}

// EventType returns the event type identifier.
func (e GenTxAddedToCollectionEvent) EventType() string {
	return "GenTxAddedToCollection"
}

// CollectionValidatedEvent is emitted when a GenTx collection is validated.
type CollectionValidatedEvent struct {
	BaseEvent
	TotalGenTxs      int
	ValidGenTxs      int
	InvalidGenTxs    int
	DuplicatesFound  int
}

// NewCollectionValidatedEvent creates a new CollectionValidatedEvent.
func NewCollectionValidatedEvent(total, valid, invalid, duplicates int) CollectionValidatedEvent {
	return CollectionValidatedEvent{
		BaseEvent:       NewBaseEvent(),
		TotalGenTxs:     total,
		ValidGenTxs:     valid,
		InvalidGenTxs:   invalid,
		DuplicatesFound: duplicates,
	}
}

// EventType returns the event type identifier.
func (e CollectionValidatedEvent) EventType() string {
	return "CollectionValidated"
}

// GenesisBuiltEvent is emitted when the genesis file is successfully built.
type GenesisBuiltEvent struct {
	BaseEvent
	ChainID         string
	ValidatorCount  int
	GenesisFilePath string
}

// NewGenesisBuiltEvent creates a new GenesisBuiltEvent.
func NewGenesisBuiltEvent(chainID string, validatorCount int, filePath string) GenesisBuiltEvent {
	return GenesisBuiltEvent{
		BaseEvent:       NewBaseEvent(),
		ChainID:         chainID,
		ValidatorCount:  validatorCount,
		GenesisFilePath: filePath,
	}
}

// EventType returns the event type identifier.
func (e GenesisBuiltEvent) EventType() string {
	return "GenesisBuilt"
}

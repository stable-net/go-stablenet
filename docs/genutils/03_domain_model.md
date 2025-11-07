# Domain Model (DDD)

## Domain-Driven Design Overview

This document defines the domain model for GenUtils using Domain-Driven Design principles:

- **Ubiquitous Language**: Common terminology used by domain experts and developers
- **Bounded Contexts**: Clear boundaries around related concepts
- **Aggregates**: Clusters of domain objects treated as a single unit
- **Entities**: Objects with distinct identity
- **Value Objects**: Immutable objects defined by their attributes
- **Domain Services**: Stateless operations that don't belong to entities
- **Domain Events**: Significant occurrences in the domain

## Ubiquitous Language

### Core Terms

| Term | Definition | Usage |
|------|------------|-------|
| **GenTx** | Genesis Transaction - declaration of validator participation | "Create a gentx for validator registration" |
| **Validator** | Entity that participates in consensus | "Each validator must submit a gentx" |
| **Operator** | Address that controls validator via governance | "Set operator address to multisig contract" |
| **Validator Key** | Private key used for block signing | "Derive validator address from node key" |
| **BLS Key** | BLS12-381 public key for WBFT consensus | "Generate BLS key from node key" |
| **Genesis** | Initial blockchain state | "Collect gentxs to build genesis file" |
| **Coordinator** | Entity that collects and merges gentxs | "Network coordinator validates all gentxs" |
| **Bootstrap** | Initialize blockchain from genesis | "Bootstrap validators from gentx collection" |

### Process Terms

| Term | Definition | Usage |
|------|------------|-------|
| **Create** | Generate new gentx from validator keys | "gstable gentx create" |
| **Sign** | Cryptographically sign gentx with validator key | "Sign message to prove key ownership" |
| **Validate** | Verify gentx format and signatures | "Validate all gentxs before collection" |
| **Collect** | Gather gentxs from multiple validators | "Collect gentxs into single directory" |
| **Merge** | Combine validated gentxs into genesis | "Merge gentxs into canonical genesis file" |
| **Bootstrap** | Initialize system from genesis | "Bootstrap network with genesis validators" |

## Bounded Contexts

### 1. GenTx Creation Context

**Purpose**: Create and sign genesis transactions

**Aggregates**:
- GenTx (root)
- Validator

**Services**:
- GenTxCreationService
- SigningService

**Invariants**:
- GenTx must be signed by validator key
- Validator address must be derivable from signature
- Operator address must be valid Ethereum address
- BLS public key must be valid G2 point

```
┌─────────────────────────────────────────┐
│     GenTx Creation Context              │
│                                          │
│  ┌────────────┐      ┌───────────┐     │
│  │   GenTx    │◄─────│ Validator │     │
│  └────────────┘      └───────────┘     │
│         │                                │
│         │ creates                        │
│         ▼                                │
│  ┌────────────┐                         │
│  │ Signature  │                         │
│  └────────────┘                         │
│                                          │
│  Services:                               │
│  - GenTxCreationService                 │
│  - SigningService                       │
└─────────────────────────────────────────┘
```

### 2. GenTx Validation Context

**Purpose**: Validate gentx correctness

**Aggregates**:
- GenTx (root)

**Services**:
- ValidationService
- SignatureVerificationService
- BLSKeyValidationService

**Invariants**:
- Signature must be valid for validator address
- BLS key must be on curve
- Timestamp must be within reasonable bounds
- No duplicate validator addresses

```
┌─────────────────────────────────────────┐
│     GenTx Validation Context            │
│                                          │
│  ┌────────────┐                         │
│  │   GenTx    │                         │
│  └─────┬──────┘                         │
│        │                                 │
│        │ validates                       │
│        ▼                                 │
│  ┌────────────────┐                     │
│  │  Validation    │                     │
│  │    Result      │                     │
│  └────────────────┘                     │
│                                          │
│  Services:                               │
│  - ValidationService                    │
│  - SignatureVerificationService         │
│  - BLSKeyValidationService              │
└─────────────────────────────────────────┘
```

### 3. GenTx Collection Context

**Purpose**: Collect and merge gentxs

**Aggregates**:
- GenTxCollection (root)
  - Contains multiple GenTx

**Services**:
- CollectionService
- MergingService
- DeduplicationService

**Invariants**:
- No duplicate validator addresses
- No duplicate operator addresses
- No duplicate BLS keys
- All gentxs must be valid

```
┌─────────────────────────────────────────┐
│     GenTx Collection Context            │
│                                          │
│  ┌───────────────────┐                  │
│  │ GenTxCollection   │                  │
│  │                   │                  │
│  │  ┌─────────┐     │                  │
│  │  │ GenTx 1 │     │                  │
│  │  ├─────────┤     │                  │
│  │  │ GenTx 2 │     │                  │
│  │  ├─────────┤     │                  │
│  │  │ GenTx N │     │                  │
│  │  └─────────┘     │                  │
│  └───────────────────┘                  │
│                                          │
│  Services:                               │
│  - CollectionService                    │
│  - MergingService                       │
│  - DeduplicationService                 │
└─────────────────────────────────────────┘
```

### 4. Genesis Building Context

**Purpose**: Build genesis file from gentxs

**Aggregates**:
- Genesis (root)
  - Contains GenTxCollection
  - Contains SystemContracts

**Services**:
- GenesisBuilderService
- ContractInjectionService
- WBFTConfigurationService

**Invariants**:
- Genesis must contain valid gentx collection
- System contracts must be initialized
- WBFT must be properly configured
- Validator set must be consistent

```
┌─────────────────────────────────────────┐
│     Genesis Building Context            │
│                                          │
│  ┌───────────────────┐                  │
│  │     Genesis       │                  │
│  │                   │                  │
│  │  ┌──────────────┐│                  │
│  │  │GenTxCollection││                  │
│  │  └──────────────┘│                  │
│  │  ┌──────────────┐│                  │
│  │  │   System     ││                  │
│  │  │  Contracts   ││                  │
│  │  └──────────────┘│                  │
│  │  ┌──────────────┐│                  │
│  │  │   WBFT       ││                  │
│  │  │   Config     ││                  │
│  │  └──────────────┘│                  │
│  └───────────────────┘                  │
│                                          │
│  Services:                               │
│  - GenesisBuilderService                │
│  - ContractInjectionService             │
│  - WBFTConfigurationService             │
└─────────────────────────────────────────┘
```

## Aggregates

### GenTx Aggregate

**Aggregate Root**: GenTx

**Responsibilities**:
- Enforce gentx invariants
- Coordinate validator and signature
- Emit domain events

**Structure**:

```go
// internal/genutils/domain/gentx.go

// GenTx is the aggregate root
type GenTx struct {
    // Identity (not exposed - internal)
    id string

    // Value Objects
    validatorAddress Address
    operatorAddress  Address
    blsPublicKey     BLSPublicKey
    signature        Signature

    // Metadata (Value Object)
    metadata ValidatorMetadata

    // Temporal attributes
    timestamp uint64
    chainID   string

    // Domain events
    events []DomainEvent
}

// Factory method - ensures invariants from creation
func NewGenTx(
    validatorKey []byte,
    operatorAddr Address,
    metadata ValidatorMetadata,
    chainID string,
    timestamp uint64,
) (*GenTx, error) {
    // Derive validator address from key
    validatorAddr, err := deriveAddress(validatorKey)
    if err != nil {
        return nil, fmt.Errorf("invalid validator key: %w", err)
    }

    // Derive BLS public key
    blsKey, err := deriveBLSPublicKey(validatorKey)
    if err != nil {
        return nil, fmt.Errorf("failed to derive BLS key: %w", err)
    }

    // Create signature
    message := constructSigningMessage(validatorAddr, operatorAddr, chainID, timestamp)
    sig, err := sign(message, validatorKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign: %w", err)
    }

    gentx := &GenTx{
        id:               generateID(),
        validatorAddress: validatorAddr,
        operatorAddress:  operatorAddr,
        blsPublicKey:     blsKey,
        signature:        sig,
        metadata:         metadata,
        timestamp:        timestamp,
        chainID:          chainID,
        events:           []DomainEvent{},
    }

    // Validate invariants
    if err := gentx.validate(); err != nil {
        return nil, err
    }

    // Emit creation event
    gentx.addEvent(GenTxCreatedEvent{
        ValidatorAddress: validatorAddr,
        Timestamp:        timestamp,
    })

    return gentx, nil
}

// Getters (read-only access to value objects)
func (g *GenTx) ValidatorAddress() Address {
    return g.validatorAddress
}

func (g *GenTx) OperatorAddress() Address {
    return g.operatorAddress
}

func (g *GenTx) BLSPublicKey() BLSPublicKey {
    return g.blsPublicKey
}

func (g *GenTx) Signature() Signature {
    return g.signature
}

func (g *GenTx) Metadata() ValidatorMetadata {
    return g.metadata
}

// Domain methods
func (g *GenTx) VerifySignature(crypto CryptoProvider) error {
    message := constructSigningMessage(
        g.validatorAddress,
        g.operatorAddress,
        g.chainID,
        g.timestamp,
    )

    recoveredAddr, err := crypto.RecoverAddress(message, g.signature.Bytes())
    if err != nil {
        return fmt.Errorf("signature recovery failed: %w", err)
    }

    if !recoveredAddr.Equals(g.validatorAddress) {
        return ErrInvalidSignature
    }

    g.addEvent(SignatureVerifiedEvent{
        ValidatorAddress: g.validatorAddress,
    })

    return nil
}

// Invariant validation
func (g *GenTx) validate() error {
    if g.validatorAddress.IsZero() {
        return ErrInvalidValidatorAddress
    }

    if g.operatorAddress.IsZero() {
        return ErrInvalidOperatorAddress
    }

    if !g.blsPublicKey.IsValid() {
        return ErrInvalidBLSKey
    }

    if g.signature.IsEmpty() {
        return ErrMissingSignature
    }

    if g.timestamp == 0 {
        return ErrInvalidTimestamp
    }

    if g.chainID == "" {
        return ErrMissingChainID
    }

    return nil
}

// Domain events
func (g *GenTx) DomainEvents() []DomainEvent {
    return g.events
}

func (g *GenTx) ClearEvents() {
    g.events = []DomainEvent{}
}

func (g *GenTx) addEvent(event DomainEvent) {
    g.events = append(g.events, event)
}
```

### GenTxCollection Aggregate

**Aggregate Root**: GenTxCollection

**Responsibilities**:
- Manage collection of gentxs
- Enforce uniqueness constraints
- Coordinate validation across collection

**Structure**:

```go
// internal/genutils/domain/gentx_collection.go

type GenTxCollection struct {
    gentxs   []*GenTx
    chainID  string
    events   []DomainEvent

    // Indexes for fast lookup
    byValidator map[string]*GenTx
    byOperator  map[string]*GenTx
    byBLSKey    map[string]*GenTx
}

// Factory
func NewGenTxCollection(chainID string) *GenTxCollection {
    return &GenTxCollection{
        gentxs:      make([]*GenTx, 0),
        chainID:     chainID,
        events:      []DomainEvent{},
        byValidator: make(map[string]*GenTx),
        byOperator:  make(map[string]*GenTx),
        byBLSKey:    make(map[string]*GenTx),
    }
}

// Add gentx with validation
func (c *GenTxCollection) Add(gentx *GenTx) error {
    // Verify chain ID matches
    if gentx.chainID != c.chainID {
        return ErrChainIDMismatch
    }

    // Check for duplicates
    validatorKey := gentx.ValidatorAddress().String()
    if _, exists := c.byValidator[validatorKey]; exists {
        return ErrDuplicateValidator
    }

    operatorKey := gentx.OperatorAddress().String()
    if _, exists := c.byOperator[operatorKey]; exists {
        return ErrDuplicateOperator
    }

    blsKey := gentx.BLSPublicKey().String()
    if _, exists := c.byBLSKey[blsKey]; exists {
        return ErrDuplicateBLSKey
    }

    // Add to collection
    c.gentxs = append(c.gentxs, gentx)
    c.byValidator[validatorKey] = gentx
    c.byOperator[operatorKey] = gentx
    c.byBLSKey[blsKey] = gentx

    c.addEvent(GenTxAddedToCollectionEvent{
        ValidatorAddress: gentx.ValidatorAddress(),
        Count:            len(c.gentxs),
    })

    return nil
}

// Query methods
func (c *GenTxCollection) FindByValidator(addr Address) (*GenTx, bool) {
    gentx, exists := c.byValidator[addr.String()]
    return gentx, exists
}

func (c *GenTxCollection) Count() int {
    return len(c.gentxs)
}

func (c *GenTxCollection) GetAll() []*GenTx {
    // Return copy to prevent external modification
    result := make([]*GenTx, len(c.gentxs))
    copy(result, c.gentxs)
    return result
}

// Sort by deterministic order (for genesis consistency)
func (c *GenTxCollection) Sort() {
    sort.Slice(c.gentxs, func(i, j int) bool {
        return c.gentxs[i].ValidatorAddress().String() <
               c.gentxs[j].ValidatorAddress().String()
    })

    c.addEvent(CollectionSortedEvent{
        Count: len(c.gentxs),
    })
}

// Validate entire collection
func (c *GenTxCollection) ValidateAll(validator GenTxValidator) error {
    for _, gentx := range c.gentxs {
        if err := validator.Validate(gentx); err != nil {
            return fmt.Errorf("invalid gentx for validator %s: %w",
                gentx.ValidatorAddress(), err)
        }
    }

    c.addEvent(CollectionValidatedEvent{
        Count: len(c.gentxs),
    })

    return nil
}

// Domain events
func (c *GenTxCollection) DomainEvents() []DomainEvent {
    return c.events
}

func (c *GenTxCollection) ClearEvents() {
    c.events = []DomainEvent{}
}

func (c *GenTxCollection) addEvent(event DomainEvent) {
    c.events = append(c.events, event)
}
```

## Entities

### Validator Entity

**Identity**: Ethereum address

**Structure**:

```go
// internal/genutils/domain/validator.go

type Validator struct {
    address  Address
    metadata ValidatorMetadata
}

func NewValidator(address Address, metadata ValidatorMetadata) *Validator {
    return &Validator{
        address:  address,
        metadata: metadata,
    }
}

// Identity comparison
func (v *Validator) Equals(other *Validator) bool {
    return v.address.Equals(other.address)
}

// Getters
func (v *Validator) Address() Address {
    return v.address
}

func (v *Validator) Metadata() ValidatorMetadata {
    return v.metadata
}
```

## Value Objects

### Address

```go
// internal/genutils/domain/address.go

type Address struct {
    value common.Address
}

func NewAddress(hex string) (Address, error) {
    if !common.IsHexAddress(hex) {
        return Address{}, ErrInvalidAddress
    }
    return Address{value: common.HexToAddress(hex)}, nil
}

func NewAddressFromBytes(bytes []byte) (Address, error) {
    if len(bytes) != 20 {
        return Address{}, ErrInvalidAddress
    }
    return Address{value: common.BytesToAddress(bytes)}, nil
}

// Immutable operations
func (a Address) String() string {
    return a.value.Hex()
}

func (a Address) Bytes() []byte {
    return a.value.Bytes()
}

func (a Address) Equals(other Address) bool {
    return a.value == other.value
}

func (a Address) IsZero() bool {
    return a.value == common.Address{}
}
```

### Signature

```go
// internal/genutils/domain/signature.go

type Signature struct {
    data [65]byte // r (32) + s (32) + v (1)
}

func NewSignature(bytes []byte) (Signature, error) {
    if len(bytes) != 65 {
        return Signature{}, ErrInvalidSignatureLength
    }

    var sig Signature
    copy(sig.data[:], bytes)
    return sig, nil
}

func (s Signature) Bytes() []byte {
    return s.data[:]
}

func (s Signature) R() *big.Int {
    return new(big.Int).SetBytes(s.data[:32])
}

func (s Signature) S() *big.Int {
    return new(big.Int).SetBytes(s.data[32:64])
}

func (s Signature) V() uint8 {
    return s.data[64]
}

func (s Signature) IsEmpty() bool {
    return s.data == [65]byte{}
}

func (s Signature) Equals(other Signature) bool {
    return s.data == other.data
}
```

### BLSPublicKey

```go
// internal/genutils/domain/bls_public_key.go

type BLSPublicKey struct {
    point *bls12381.G2
}

func NewBLSPublicKey(bytes []byte) (BLSPublicKey, error) {
    point, err := bls12381.NewG2().FromBytes(bytes)
    if err != nil {
        return BLSPublicKey{}, ErrInvalidBLSKey
    }

    return BLSPublicKey{point: point}, nil
}

func (b BLSPublicKey) Bytes() []byte {
    return bls12381.NewG2().ToBytes(b.point)
}

func (b BLSPublicKey) String() string {
    return hex.EncodeToString(b.Bytes())
}

func (b BLSPublicKey) IsValid() bool {
    return b.point != nil && bls12381.NewG2().IsOnCurve(b.point)
}

func (b BLSPublicKey) Equals(other BLSPublicKey) bool {
    return bytes.Equal(b.Bytes(), other.Bytes())
}

// Verify BLS signature (for future use)
func (b BLSPublicKey) VerifySignature(message, signature []byte) bool {
    // BLS signature verification logic
    return bls12381.Verify(b.point, message, signature)
}
```

### ValidatorMetadata

```go
// internal/genutils/domain/validator_metadata.go

type ValidatorMetadata struct {
    name        string
    description string
    website     string
    contact     string
}

func NewValidatorMetadata(
    name, description, website, contact string,
) (ValidatorMetadata, error) {
    // Validation
    if name == "" {
        return ValidatorMetadata{}, ErrMissingValidatorName
    }

    if len(name) > 70 {
        return ValidatorMetadata{}, ErrValidatorNameTooLong
    }

    if len(description) > 280 {
        return ValidatorMetadata{}, ErrDescriptionTooLong
    }

    return ValidatorMetadata{
        name:        name,
        description: description,
        website:     website,
        contact:     contact,
    }, nil
}

// Getters
func (m ValidatorMetadata) Name() string        { return m.name }
func (m ValidatorMetadata) Description() string { return m.description }
func (m ValidatorMetadata) Website() string     { return m.website }
func (m ValidatorMetadata) Contact() string     { return m.contact }

// Equality
func (m ValidatorMetadata) Equals(other ValidatorMetadata) bool {
    return m.name == other.name &&
           m.description == other.description &&
           m.website == other.website &&
           m.contact == other.contact
}
```

## Domain Services

### ValidationService

```go
// internal/genutils/service/validation/service.go

type ValidationService struct {
    signatureValidator *SignatureValidator
    formatValidator    *FormatValidator
    businessValidator  *BusinessRuleValidator
}

func NewValidationService(
    crypto CryptoProvider,
) *ValidationService {
    return &ValidationService{
        signatureValidator: NewSignatureValidator(crypto),
        formatValidator:    NewFormatValidator(),
        businessValidator:  NewBusinessRuleValidator(),
    }
}

func (s *ValidationService) Validate(gentx *GenTx) error {
    if err := s.formatValidator.Validate(gentx); err != nil {
        return fmt.Errorf("format validation failed: %w", err)
    }

    if err := s.signatureValidator.Validate(gentx); err != nil {
        return fmt.Errorf("signature validation failed: %w", err)
    }

    if err := s.businessValidator.Validate(gentx); err != nil {
        return fmt.Errorf("business rule validation failed: %w", err)
    }

    return nil
}
```

### CollectionService

```go
// internal/genutils/service/collection/service.go

type CollectionService struct {
    repository GenTxRepository
    validator  GenTxValidator
}

func NewCollectionService(
    repo GenTxRepository,
    validator GenTxValidator,
) *CollectionService {
    return &CollectionService{
        repository: repo,
        validator:  validator,
    }
}

func (s *CollectionService) CollectFromDirectory(
    dir string,
    chainID string,
) (*GenTxCollection, error) {
    // Load all gentxs from directory
    gentxs, err := s.repository.FindAll()
    if err != nil {
        return nil, fmt.Errorf("failed to load gentxs: %w", err)
    }

    // Create collection
    collection := NewGenTxCollection(chainID)

    // Validate and add each gentx
    for _, gentx := range gentxs {
        if err := s.validator.Validate(gentx); err != nil {
            return nil, fmt.Errorf(
                "invalid gentx for validator %s: %w",
                gentx.ValidatorAddress(), err,
            )
        }

        if err := collection.Add(gentx); err != nil {
            return nil, fmt.Errorf(
                "failed to add gentx for validator %s: %w",
                gentx.ValidatorAddress(), err,
            )
        }
    }

    // Sort for deterministic order
    collection.Sort()

    return collection, nil
}
```

### GenesisBuilderService

```go
// internal/genutils/service/genesis/builder_service.go

type GenesisBuilderService struct {
    contractInjector ContractInjector
    wbftConfigurator WBFTConfigurator
}

func NewGenesisBuilderService(
    injector ContractInjector,
    configurator WBFTConfigurator,
) *GenesisBuilderService {
    return &GenesisBuilderService{
        contractInjector: injector,
        wbftConfigurator: configurator,
    }
}

func (s *GenesisBuilderService) BuildGenesis(
    collection *GenTxCollection,
    baseGenesis *core.Genesis,
) (*core.Genesis, error) {
    // Start with base genesis
    genesis := baseGenesis

    // Extract validators and BLS keys
    validators := make([]common.Address, 0, collection.Count())
    blsKeys := make([]string, 0, collection.Count())

    for _, gentx := range collection.GetAll() {
        validators = append(validators, gentx.ValidatorAddress().value)
        blsKeys = append(blsKeys, gentx.BLSPublicKey().String())
    }

    // Configure WBFT
    if err := s.wbftConfigurator.Configure(
        genesis,
        validators,
        blsKeys,
    ); err != nil {
        return nil, fmt.Errorf("failed to configure WBFT: %w", err)
    }

    // Inject system contracts
    if err := s.contractInjector.InjectGovValidator(
        genesis,
        collection,
    ); err != nil {
        return nil, fmt.Errorf("failed to inject contracts: %w", err)
    }

    return genesis, nil
}
```

## Domain Events

```go
// internal/genutils/domain/events.go

// DomainEvent interface
type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
}

// Base event
type baseEvent struct {
    occurredAt time.Time
}

func (e baseEvent) OccurredAt() time.Time {
    return e.occurredAt
}

// Specific events
type GenTxCreatedEvent struct {
    baseEvent
    ValidatorAddress Address
    Timestamp        uint64
}

func (e GenTxCreatedEvent) EventName() string {
    return "GenTxCreated"
}

type SignatureVerifiedEvent struct {
    baseEvent
    ValidatorAddress Address
}

func (e SignatureVerifiedEvent) EventName() string {
    return "SignatureVerified"
}

type GenTxAddedToCollectionEvent struct {
    baseEvent
    ValidatorAddress Address
    Count            int
}

func (e GenTxAddedToCollectionEvent) EventName() string {
    return "GenTxAddedToCollection"
}

type CollectionValidatedEvent struct {
    baseEvent
    Count int
}

func (e CollectionValidatedEvent) EventName() string {
    return "CollectionValidated"
}

type GenesisBuiltEvent struct {
    baseEvent
    ValidatorCount int
    ChainID        string
}

func (e GenesisBuiltEvent) EventName() string {
    return "GenesisBuilt"
}
```

## Repository Interfaces

```go
// internal/genutils/repository/interface.go

// GenTxRepository - persistence abstraction
type GenTxRepository interface {
    // Save gentx to storage
    Save(gentx *GenTx) error

    // Find gentx by validator address
    FindByValidator(address Address) (*GenTx, error)

    // Find all gentxs
    FindAll() ([]*GenTx, error)

    // Check if gentx exists
    Exists(address Address) (bool, error)

    // Delete gentx
    Delete(address Address) error

    // Count gentxs
    Count() (int, error)
}

// GenesisRepository - genesis file persistence
type GenesisRepository interface {
    // Save genesis file
    Save(genesis *core.Genesis, path string) error

    // Load genesis file
    Load(path string) (*core.Genesis, error)

    // Validate genesis file
    Validate(genesis *core.Genesis) error
}
```

## Summary

This domain model provides:

1. **Clear Boundaries**: Bounded contexts separate concerns
2. **Rich Domain Logic**: Aggregates enforce invariants
3. **Immutability**: Value objects are immutable
4. **Type Safety**: Strong typing prevents invalid states
5. **Testability**: Pure domain logic easy to test
6. **Event Sourcing**: Domain events track state changes

Next: See [04_implementation_guide.md](./04_implementation_guide.md) for TDD/DDD implementation approach.

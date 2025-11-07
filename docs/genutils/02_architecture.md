# GenUtils Architecture Design

## Design Principles

This architecture follows **SOLID principles** and **Clean Architecture** to ensure maintainability, testability, and extensibility.

## Layered Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Presentation Layer                  │
│              (CLI Commands - cmd/gstable/)           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │  create  │  │ validate │  │ collect  │          │
│  └──────────┘  └──────────┘  └──────────┘          │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                 Application Layer                    │
│         (Use Cases - internal/genutils/application/) │
│  ┌──────────────┐  ┌──────────────┐                │
│  │CreateGenTx   │  │CollectGenTxs │                │
│  │UseCase       │  │UseCase       │                │
│  └──────────────┘  └──────────────┘                │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│                  Domain Layer                        │
│           (Business Logic - internal/genutils/domain/)│
│  ┌────────────┐  ┌───────────┐  ┌──────────┐      │
│  │ GenTx      │  │ Validator │  │  Value   │      │
│  │ (Aggregate)│  │ (Entity)  │  │  Objects │      │
│  └────────────┘  └───────────┘  └──────────┘      │
│                                                      │
│  ┌─────────────────────────────────────────┐       │
│  │         Domain Services                  │       │
│  │  - Validation  - Collection  - Genesis  │       │
│  └─────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────┐
│               Infrastructure Layer                   │
│     (Persistence, Crypto - internal/genutils/infra/) │
│  ┌────────────┐  ┌───────────┐  ┌──────────┐      │
│  │ File       │  │ Keystore  │  │  Crypto  │      │
│  │ Repository │  │ Manager   │  │  Service │      │
│  └────────────┘  └───────────┘  └──────────┘      │
└─────────────────────────────────────────────────────┘
```

## SOLID Principles Application

### 1. Single Responsibility Principle (SRP)

Each component has one reason to change:

#### **GenTx (Domain Aggregate)**
- **Responsibility**: Represent a genesis transaction with validation rules
- **Changes when**: GenTx structure or validation rules change

```go
// internal/genutils/domain/gentx.go
type GenTx struct {
    validatorAddress common.Address
    operatorAddress  common.Address
    blsPublicKey     BLSPublicKey
    signature        Signature
    metadata         ValidatorMetadata
    timestamp        uint64
}

// Single responsibility: validate self-consistency
func (g *GenTx) Validate() error {
    // Validation logic
}
```

#### **ValidationService**
- **Responsibility**: Validate gentx signatures and format
- **Changes when**: Validation rules change

```go
// internal/genutils/service/validation/service.go
type ValidationService struct {
    cryptoProvider CryptoProvider
}

// Single responsibility: validate signature
func (v *ValidationService) ValidateSignature(gentx *GenTx) error {
    // Signature validation logic
}
```

#### **GenTxRepository**
- **Responsibility**: Persist and retrieve gentx data
- **Changes when**: Storage mechanism changes

```go
// internal/genutils/repository/interface.go
type GenTxRepository interface {
    Save(gentx *GenTx) error
    FindAll() ([]*GenTx, error)
    FindByValidator(address common.Address) (*GenTx, error)
}
```

### 2. Open/Closed Principle (OCP)

Components are open for extension but closed for modification:

#### **Validator Interface**

```go
// internal/genutils/service/validation/interface.go
type Validator interface {
    Validate(gentx *GenTx) error
}

// Base validators - CLOSED for modification
type SignatureValidator struct{}
type FormatValidator struct{}
type BLSKeyValidator struct{}

// Composite validator - OPEN for extension
type CompositeValidator struct {
    validators []Validator
}

func (c *CompositeValidator) AddValidator(v Validator) {
    c.validators = append(c.validators, v)
}

// New validators can be added without modifying existing code
type CustomBusinessRuleValidator struct{}
func (c *CustomBusinessRuleValidator) Validate(gentx *GenTx) error {
    // Custom validation logic
}

// Usage
validator := NewCompositeValidator()
validator.AddValidator(&SignatureValidator{})
validator.AddValidator(&FormatValidator{})
validator.AddValidator(&CustomBusinessRuleValidator{}) // Extension
```

#### **Genesis Builder Strategy**

```go
// internal/genutils/service/genesis/builder.go
type GenesisBuilder interface {
    Build(gentxs []*GenTx) (*core.Genesis, error)
}

// Default implementation
type DefaultGenesisBuilder struct {
    injectors []ContractInjector
}

// Can extend with custom builders without modifying base
type CustomGenesisBuilder struct {
    DefaultGenesisBuilder
    customLogic func([]*GenTx) error
}
```

### 3. Liskov Substitution Principle (LSP)

Subtypes are substitutable for their base types:

#### **Repository Implementations**

```go
// internal/genutils/repository/interface.go
type GenTxRepository interface {
    Save(gentx *GenTx) error
    FindAll() ([]*GenTx, error)
}

// File-based implementation
type FileRepository struct {
    baseDir string
}

func (f *FileRepository) Save(gentx *GenTx) error {
    // File storage implementation
}

// Memory-based implementation (for testing)
type MemoryRepository struct {
    storage map[string]*GenTx
}

func (m *MemoryRepository) Save(gentx *GenTx) error {
    // In-memory storage implementation
}

// Both can be used interchangeably
func ProcessGenTxs(repo GenTxRepository, gentxs []*GenTx) error {
    for _, gentx := range gentxs {
        if err := repo.Save(gentx); err != nil {
            return err
        }
    }
    return nil
}

// LSP satisfied: both implementations work correctly
fileRepo := NewFileRepository("/path/to/gentxs")
memoryRepo := NewMemoryRepository()

ProcessGenTxs(fileRepo, gentxs)   // Works
ProcessGenTxs(memoryRepo, gentxs) // Works identically
```

#### **Crypto Provider**

```go
// internal/genutils/crypto/interface.go
type CryptoProvider interface {
    Sign(message []byte, key []byte) ([]byte, error)
    Verify(message, signature []byte, address common.Address) error
    RecoverAddress(message, signature []byte) (common.Address, error)
}

// Ethereum crypto implementation
type EthereumCryptoProvider struct{}

// Mock for testing
type MockCryptoProvider struct{}

// Both satisfy the interface contract
```

### 4. Interface Segregation Principle (ISP)

Clients should not depend on interfaces they don't use:

#### **Segregated Interfaces**

```go
// Instead of one large interface
type GenTxManager interface { // ❌ Bad
    Create(...) error
    Validate(...) error
    Save(...) error
    Load(...) error
    Collect(...) error
    Merge(...) error
    BuildGenesis(...) error
}

// Segregate into focused interfaces
type GenTxCreator interface { // ✅ Good
    Create(...) error
}

type GenTxValidator interface { // ✅ Good
    Validate(...) error
}

type GenTxRepository interface { // ✅ Good
    Save(...) error
    Load(...) error
}

type GenTxCollector interface { // ✅ Good
    Collect(...) error
}

type GenesisBuilder interface { // ✅ Good
    BuildGenesis(...) error
}

// Clients only depend on what they need
type CreateGenTxUseCase struct {
    creator    GenTxCreator    // Only needs creator
    repository GenTxRepository // Only needs repository
}

type CollectGenTxsUseCase struct {
    collector GenTxCollector  // Only needs collector
    validator GenTxValidator  // Only needs validator
    builder   GenesisBuilder  // Only needs builder
}
```

### 5. Dependency Inversion Principle (DIP)

Depend on abstractions, not concretions:

#### **Use Case with Dependency Injection**

```go
// internal/genutils/application/create_gentx.go

// High-level policy depends on abstractions
type CreateGenTxUseCase struct {
    cryptoProvider CryptoProvider    // Interface
    repository     GenTxRepository   // Interface
    validator      GenTxValidator    // Interface
    keyManager     KeyManager        // Interface
}

// Constructor injection
func NewCreateGenTxUseCase(
    crypto CryptoProvider,
    repo GenTxRepository,
    validator GenTxValidator,
    keyManager KeyManager,
) *CreateGenTxUseCase {
    return &CreateGenTxUseCase{
        cryptoProvider: crypto,
        repository:     repo,
        validator:      validator,
        keyManager:     keyManager,
    }
}

// Execute depends only on abstractions
func (uc *CreateGenTxUseCase) Execute(request CreateGenTxRequest) error {
    // Use interfaces, not concrete implementations
    key, err := uc.keyManager.GetValidatorKey(request.KeyFile)
    if err != nil {
        return err
    }

    signature, err := uc.cryptoProvider.Sign(message, key)
    if err != nil {
        return err
    }

    gentx := domain.NewGenTx(...)

    if err := uc.validator.Validate(gentx); err != nil {
        return err
    }

    return uc.repository.Save(gentx)
}

// Concrete implementations injected from outside
func main() {
    // Dependency injection container
    crypto := ethereum.NewCryptoProvider()
    repo := file.NewRepository("/path/to/gentxs")
    validator := validation.NewCompositeValidator()
    keyManager := keystore.NewManager()

    useCase := NewCreateGenTxUseCase(crypto, repo, validator, keyManager)

    err := useCase.Execute(request)
}
```

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│                                                                  │
│  ┌──────────────┐  ┌───────────────┐  ┌────────────────┐      │
│  │ gentx create │  │gentx validate │  │ gentx collect  │      │
│  └──────┬───────┘  └───────┬───────┘  └────────┬───────┘      │
│         │                   │                    │              │
└─────────┼───────────────────┼────────────────────┼──────────────┘
          │                   │                    │
          ▼                   ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Application Layer                           │
│                                                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │ CreateGenTxUseCase  │  │CollectGenTxsUseCase │              │
│  │                     │  │                     │              │
│  │ - cryptoProvider────┼─→│ - collector         │              │
│  │ - repository────────┼─→│ - validator         │              │
│  │ - validator         │  │ - builder           │              │
│  │ - keyManager        │  │ - repository        │              │
│  └─────────────────────┘  └─────────────────────┘              │
│            │                        │                            │
└────────────┼────────────────────────┼────────────────────────────┘
             │                        │
             ▼                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Domain Layer                              │
│                                                                  │
│  ┌────────────────┐         ┌──────────────────┐               │
│  │     GenTx      │◄────────│   Validator      │               │
│  │  (Aggregate)   │         │    (Entity)      │               │
│  │                │         │                  │               │
│  │ - Validate()   │         │ - GetAddress()   │               │
│  │ - Sign()       │         │ - GetBLSKey()    │               │
│  └────────────────┘         └──────────────────┘               │
│          │                           │                           │
│          │   ┌───────────────────────┴────────┐                │
│          │   │      Value Objects              │                │
│          │   │  - Address                      │                │
│          └───┤  - Signature                    │                │
│              │  - BLSPublicKey                 │                │
│              │  - ValidatorMetadata            │                │
│              └─────────────────────────────────┘                │
│                                                                  │
│  ┌─────────────────────────────────────────────────────┐       │
│  │              Domain Services                         │       │
│  │                                                       │       │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐│       │
│  │  │ValidationSvc │  │CollectionSvc │  │GenesisSvc ││       │
│  │  └──────────────┘  └──────────────┘  └───────────┘│       │
│  └─────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                           │
│                                                                  │
│  ┌─────────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ FileRepository  │  │  KeyManager  │  │CryptoProvider│      │
│  │                 │  │              │  │              │      │
│  │ - Save()        │  │ - LoadKey()  │  │ - Sign()     │      │
│  │ - FindAll()     │  │ - SaveKey()  │  │ - Verify()   │      │
│  └─────────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────────┘
```

## Domain Model (DDD)

### Bounded Contexts

```
┌────────────────────────────────────────────────────┐
│           GenTx Bounded Context                    │
│                                                     │
│  ┌──────────────────────────────────────┐         │
│  │     GenTx Creation Context           │         │
│  │  - Create GenTx                      │         │
│  │  - Sign GenTx                        │         │
│  │  - Validate Format                   │         │
│  └──────────────────────────────────────┘         │
│                                                     │
│  ┌──────────────────────────────────────┐         │
│  │     GenTx Collection Context         │         │
│  │  - Collect GenTxs                    │         │
│  │  - Validate Collection               │         │
│  │  - Merge GenTxs                      │         │
│  └──────────────────────────────────────┘         │
│                                                     │
│  ┌──────────────────────────────────────┐         │
│  │     Genesis Building Context         │         │
│  │  - Build Genesis                     │         │
│  │  - Inject Contracts                  │         │
│  │  - Configure WBFT                    │         │
│  └──────────────────────────────────────┘         │
└────────────────────────────────────────────────────┘
```

### Aggregates

**GenTx** is the aggregate root:

```go
// internal/genutils/domain/gentx.go
type GenTx struct {
    // Identity
    id               string

    // Value Objects
    validatorAddress Address
    operatorAddress  Address
    blsPublicKey     BLSPublicKey
    signature        Signature

    // Entity Reference
    validator        *Validator

    // Metadata
    metadata         ValidatorMetadata
    timestamp        uint64
    chainID          string
}

// Aggregate invariants enforced by methods
func (g *GenTx) ChangeOperator(newOperator Address, signature Signature) error {
    // Enforce business rules
    if !g.validator.CanChangeOperator() {
        return ErrOperatorChangeForbidden
    }

    // Validate signature
    if !signature.IsValidFor(newOperator) {
        return ErrInvalidSignature
    }

    g.operatorAddress = newOperator
    g.EmitEvent(OperatorChangedEvent{...})
    return nil
}
```

### Entities

**Validator** entity:

```go
// internal/genutils/domain/validator.go
type Validator struct {
    address      Address
    name         string
    blsPublicKey BLSPublicKey
    metadata     ValidatorMetadata
}

// Entity has identity and lifecycle
func (v *Validator) Equals(other *Validator) bool {
    return v.address.Equals(other.address)
}
```

### Value Objects

```go
// internal/genutils/domain/value_objects.go

// Address - immutable value object
type Address struct {
    value common.Address
}

func NewAddress(hex string) (Address, error) {
    if !common.IsHexAddress(hex) {
        return Address{}, ErrInvalidAddress
    }
    return Address{value: common.HexToAddress(hex)}, nil
}

func (a Address) Equals(other Address) bool {
    return a.value == other.value
}

// Signature - immutable value object
type Signature struct {
    r, s, v *big.Int
}

func (s Signature) Bytes() []byte {
    return crypto.FromECDSA(s.r, s.s, s.v)
}

// BLSPublicKey - immutable value object
type BLSPublicKey struct {
    point *bls.G2Point
}

func (b BLSPublicKey) Verify(message []byte, signature []byte) bool {
    return bls.Verify(b.point, message, signature)
}
```

### Domain Services

```go
// internal/genutils/service/validation/service.go
type ValidationService struct {
    signatureValidator *SignatureValidator
    formatValidator    *FormatValidator
    businessValidator  *BusinessRuleValidator
}

func (s *ValidationService) ValidateGenTx(gentx *GenTx) error {
    if err := s.signatureValidator.Validate(gentx); err != nil {
        return err
    }
    if err := s.formatValidator.Validate(gentx); err != nil {
        return err
    }
    return s.businessValidator.Validate(gentx)
}

// internal/genutils/service/collection/service.go
type CollectionService struct {
    repository GenTxRepository
    validator  GenTxValidator
}

func (s *CollectionService) CollectGenTxs(dir string) ([]*GenTx, error) {
    gentxs, err := s.repository.FindAll()
    if err != nil {
        return nil, err
    }

    // Domain logic: validate and deduplicate
    validated := make([]*GenTx, 0)
    seen := make(map[string]bool)

    for _, gentx := range gentxs {
        if err := s.validator.Validate(gentx); err != nil {
            return nil, fmt.Errorf("invalid gentx: %w", err)
        }

        key := gentx.ValidatorAddress().String()
        if seen[key] {
            return nil, ErrDuplicateValidator
        }
        seen[key] = true

        validated = append(validated, gentx)
    }

    return validated, nil
}
```

## Data Flow

### Create GenTx Flow

```
User Input
    │
    ▼
┌─────────────────────┐
│  CLI Command        │
│  (gentx create)     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ CreateGenTxUseCase  │
│                     │
│ 1. Load key         │───→ KeyManager
│ 2. Create GenTx     │───→ GenTx.New()
│ 3. Sign message     │───→ CryptoProvider
│ 4. Validate         │───→ Validator
│ 5. Save             │───→ Repository
└──────────┬──────────┘
           │
           ▼
    gentx-xxx.json
```

### Collect GenTxs Flow

```
GenTx Files
    │
    ▼
┌─────────────────────┐
│  CLI Command        │
│  (gentx collect)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│CollectGenTxsUseCase │
│                     │
│ 1. Load all         │───→ Repository
│ 2. Validate each    │───→ Validator
│ 3. Deduplicate      │───→ CollectionService
│ 4. Sort             │───→ CollectionService
│ 5. Build genesis    │───→ GenesisBuilder
│ 6. Inject contracts │───→ ContractInjector
│ 7. Save genesis     │───→ GenesisWriter
└──────────┬──────────┘
           │
           ▼
      genesis.json
```

## Error Handling Strategy

### Domain Errors

```go
// internal/genutils/domain/errors.go
type DomainError struct {
    Code    string
    Message string
    Cause   error
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}

// Predefined domain errors
var (
    ErrInvalidAddress       = &DomainError{Code: "INVALID_ADDRESS", Message: "invalid ethereum address"}
    ErrInvalidSignature     = &DomainError{Code: "INVALID_SIGNATURE", Message: "signature verification failed"}
    ErrDuplicateValidator   = &DomainError{Code: "DUPLICATE_VALIDATOR", Message: "validator already exists"}
    ErrInvalidBLSKey        = &DomainError{Code: "INVALID_BLS_KEY", Message: "invalid BLS public key"}
)
```

### Error Propagation

```go
// Use case handles errors appropriately
func (uc *CreateGenTxUseCase) Execute(req CreateGenTxRequest) error {
    gentx, err := domain.NewGenTx(...)
    if err != nil {
        // Domain error - user input problem
        return fmt.Errorf("failed to create gentx: %w", err)
    }

    if err := uc.validator.Validate(gentx); err != nil {
        // Validation error - user input problem
        return fmt.Errorf("validation failed: %w", err)
    }

    if err := uc.repository.Save(gentx); err != nil {
        // Infrastructure error - system problem
        return fmt.Errorf("failed to save gentx: %w", err)
    }

    return nil
}
```

## Testing Strategy

### Unit Tests

Each component tested in isolation:

```go
// internal/genutils/domain/gentx_test.go
func TestGenTx_Validate(t *testing.T) {
    tests := []struct{
        name    string
        gentx   *GenTx
        wantErr error
    }{
        {
            name: "valid gentx",
            gentx: validGenTx(),
            wantErr: nil,
        },
        {
            name: "invalid address",
            gentx: gentxWithInvalidAddress(),
            wantErr: ErrInvalidAddress,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.gentx.Validate()
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("got %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

Test component interactions:

```go
// internal/genutils/application/create_gentx_test.go
func TestCreateGenTxUseCase_Execute(t *testing.T) {
    // Setup dependencies
    crypto := mock.NewCryptoProvider()
    repo := memory.NewRepository()
    validator := validation.NewCompositeValidator()
    keyManager := mock.NewKeyManager()

    useCase := NewCreateGenTxUseCase(crypto, repo, validator, keyManager)

    // Execute use case
    req := CreateGenTxRequest{...}
    err := useCase.Execute(req)

    // Verify results
    assert.NoError(t, err)

    gentxs, _ := repo.FindAll()
    assert.Len(t, gentxs, 1)
}
```

### End-to-End Tests

Test complete workflows:

```go
// cmd/gstable/gentx_e2e_test.go
func TestGenTxWorkflow_CreateAndCollect(t *testing.T) {
    // Setup test environment
    tmpDir := setupTestDir(t)
    defer os.RemoveAll(tmpDir)

    // Create 3 validators
    for i := 0; i < 3; i++ {
        cmd := exec.Command("gstable", "gentx", "create",
            "--validator-key", fmt.Sprintf("key%d", i),
            "--output", filepath.Join(tmpDir, fmt.Sprintf("gentx-%d.json", i)),
        )
        err := cmd.Run()
        assert.NoError(t, err)
    }

    // Collect gentxs
    cmd := exec.Command("gstable", "gentx", "collect",
        "--gentx-dir", tmpDir,
        "--output", filepath.Join(tmpDir, "genesis.json"),
    )
    err := cmd.Run()
    assert.NoError(t, err)

    // Verify genesis file
    genesis := loadGenesis(t, filepath.Join(tmpDir, "genesis.json"))
    assert.Len(t, genesis.Anzeon.Init.Validators, 3)
}
```

## Dependency Injection

### Manual DI (for simplicity)

```go
// cmd/gstable/main.go
func buildCreateGenTxUseCase(ctx *cli.Context) *application.CreateGenTxUseCase {
    // Infrastructure dependencies
    crypto := ethereum.NewCryptoProvider()
    keyManager := keystore.NewManager(ctx.String("keystore"))
    repo := file.NewRepository(ctx.String("gentx-dir"))

    // Domain services
    sigValidator := validation.NewSignatureValidator(crypto)
    fmtValidator := validation.NewFormatValidator()
    blsValidator := validation.NewBLSValidator()

    validator := validation.NewCompositeValidator(
        sigValidator,
        fmtValidator,
        blsValidator,
    )

    // Application use case
    return application.NewCreateGenTxUseCase(
        crypto,
        repo,
        validator,
        keyManager,
    )
}
```

### Future: DI Container (wire, dig, fx)

```go
// internal/genutils/wire.go (using google/wire)
//+build wireinject

func InitializeCreateGenTxUseCase(keystoreDir, gentxDir string) *application.CreateGenTxUseCase {
    wire.Build(
        // Infrastructure
        ethereum.NewCryptoProvider,
        keystore.NewManager,
        file.NewRepository,

        // Domain services
        validation.NewSignatureValidator,
        validation.NewFormatValidator,
        validation.NewBLSValidator,
        validation.NewCompositeValidator,

        // Application
        application.NewCreateGenTxUseCase,
    )
    return nil
}
```

## Summary

This architecture provides:

1. **Maintainability**: SOLID principles ensure each component is focused and loosely coupled
2. **Testability**: Dependency injection and interface-based design enable easy testing
3. **Extensibility**: Open/Closed principle allows adding features without breaking existing code
4. **Domain Focus**: DDD patterns keep business logic separate from infrastructure concerns
5. **Type Safety**: Strong typing and value objects prevent invalid states

Next: See [03_domain_model.md](./03_domain_model.md) for detailed domain model design.

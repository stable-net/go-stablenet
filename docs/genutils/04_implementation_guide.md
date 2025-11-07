# Implementation Guide (TDD + DDD)

## Overview

This guide demonstrates how to implement GenUtils using Test-Driven Development (TDD) combined with Domain-Driven Design (DDD).

## TDD Cycle: Red → Green → Refactor

```
┌─────────────┐
│  1. RED     │  Write failing test
│  ❌ Test    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  2. GREEN   │  Write minimal code to pass
│  ✅ Test    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 3. REFACTOR │  Clean up code
│  ♻️ Code    │
└──────┬──────┘
       │
       └──────┐
              │
              ▼
       (Repeat)
```

## Implementation Order

Following DDD inside-out approach:

1. **Domain Layer** (Core business logic)
   - Value Objects
   - Entities
   - Aggregates
   - Domain Services
   - Domain Events

2. **Application Layer** (Use cases)
   - Use case implementations
   - DTO definitions
   - Application services

3. **Infrastructure Layer** (Technical concerns)
   - Repository implementations
   - Crypto providers
   - File system operations

4. **Presentation Layer** (User interface)
   - CLI commands
   - Input validation
   - Output formatting

## Phase 1: Value Objects (Foundation)

### Step 1: Address Value Object

#### 1.1 RED - Write failing test

```go
// internal/genutils/domain/address_test.go
package domain_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "go-stablenet/internal/genutils/domain"
)

func TestAddress_NewAddress_ValidHex(t *testing.T) {
    // Arrange
    validHex := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"

    // Act
    addr, err := domain.NewAddress(validHex)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, validHex, addr.String())
}

func TestAddress_NewAddress_InvalidHex(t *testing.T) {
    // Arrange
    invalidHex := "not-a-hex-address"

    // Act
    _, err := domain.NewAddress(invalidHex)

    // Assert
    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidAddress)
}

func TestAddress_Equals(t *testing.T) {
    addr1, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    addr2, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    addr3, _ := domain.NewAddress("0x123d35Cc6634C0532925a3b844Bc9e7595f0123")

    assert.True(t, addr1.Equals(addr2))
    assert.False(t, addr1.Equals(addr3))
}
```

Run test: `go test ./internal/genutils/domain/` → **FAIL** ❌

#### 1.2 GREEN - Implement minimal code

```go
// internal/genutils/domain/address.go
package domain

import (
    "errors"
    "github.com/ethereum/go-ethereum/common"
)

var ErrInvalidAddress = errors.New("invalid ethereum address")

type Address struct {
    value common.Address
}

func NewAddress(hex string) (Address, error) {
    if !common.IsHexAddress(hex) {
        return Address{}, ErrInvalidAddress
    }
    return Address{value: common.HexToAddress(hex)}, nil
}

func (a Address) String() string {
    return a.value.Hex()
}

func (a Address) Equals(other Address) bool {
    return a.value == other.value
}

func (a Address) IsZero() bool {
    return a.value == common.Address{}
}
```

Run test: `go test ./internal/genutils/domain/` → **PASS** ✅

#### 1.3 REFACTOR - Add more functionality

```go
// Add more tests
func TestAddress_Bytes(t *testing.T) {
    addr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    bytes := addr.Bytes()
    assert.Len(t, bytes, 20)
}

func TestAddress_IsZero(t *testing.T) {
    zero := domain.Address{}
    assert.True(t, zero.IsZero())

    nonZero, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    assert.False(t, nonZero.IsZero())
}
```

Add implementation:

```go
func (a Address) Bytes() []byte {
    return a.value.Bytes()
}
```

### Step 2: Signature Value Object

#### 2.1 RED - Write test

```go
// internal/genutils/domain/signature_test.go
func TestSignature_NewSignature_Valid(t *testing.T) {
    validSig := make([]byte, 65)
    sig, err := domain.NewSignature(validSig)

    assert.NoError(t, err)
    assert.Equal(t, validSig, sig.Bytes())
}

func TestSignature_NewSignature_InvalidLength(t *testing.T) {
    invalidSig := make([]byte, 32) // Wrong length
    _, err := domain.NewSignature(invalidSig)

    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidSignatureLength)
}

func TestSignature_Components(t *testing.T) {
    sigBytes := make([]byte, 65)
    // Set known values
    copy(sigBytes[:32], bytes.Repeat([]byte{0x01}, 32))  // R
    copy(sigBytes[32:64], bytes.Repeat([]byte{0x02}, 32)) // S
    sigBytes[64] = 0x1b // V

    sig, _ := domain.NewSignature(sigBytes)

    assert.Equal(t, big.NewInt(0).SetBytes(bytes.Repeat([]byte{0x01}, 32)), sig.R())
    assert.Equal(t, big.NewInt(0).SetBytes(bytes.Repeat([]byte{0x02}, 32)), sig.S())
    assert.Equal(t, uint8(0x1b), sig.V())
}
```

#### 2.2 GREEN - Implement

```go
// internal/genutils/domain/signature.go
package domain

import (
    "errors"
    "math/big"
)

var ErrInvalidSignatureLength = errors.New("signature must be 65 bytes")

type Signature struct {
    data [65]byte
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
```

### Step 3: BLSPublicKey Value Object

#### 3.1 RED - Write test

```go
// internal/genutils/domain/bls_public_key_test.go
func TestBLSPublicKey_NewBLSPublicKey_Valid(t *testing.T) {
    // Generate valid BLS key bytes (G2 point)
    validBytes := generateValidBLSKeyBytes(t)

    key, err := domain.NewBLSPublicKey(validBytes)

    assert.NoError(t, err)
    assert.True(t, key.IsValid())
}

func TestBLSPublicKey_NewBLSPublicKey_Invalid(t *testing.T) {
    invalidBytes := []byte{0x00, 0x01, 0x02} // Invalid G2 point

    _, err := domain.NewBLSPublicKey(invalidBytes)

    assert.Error(t, err)
    assert.ErrorIs(t, err, domain.ErrInvalidBLSKey)
}
```

#### 3.2 GREEN - Implement

```go
// internal/genutils/domain/bls_public_key.go
package domain

import (
    "bytes"
    "encoding/hex"
    "errors"
    "github.com/ethereum/go-ethereum/crypto/bls12381"
)

var ErrInvalidBLSKey = errors.New("invalid BLS public key")

type BLSPublicKey struct {
    point *bls12381.PointG2
}

func NewBLSPublicKey(data []byte) (BLSPublicKey, error) {
    point := &bls12381.PointG2{}
    if err := point.DecodePoint(data); err != nil {
        return BLSPublicKey{}, ErrInvalidBLSKey
    }

    return BLSPublicKey{point: point}, nil
}

func (b BLSPublicKey) Bytes() []byte {
    return b.point.EncodePoint()
}

func (b BLSPublicKey) String() string {
    return "0x" + hex.EncodeToString(b.Bytes())
}

func (b BLSPublicKey) IsValid() bool {
    return b.point != nil && b.point.IsOnCurve()
}

func (b BLSPublicKey) Equals(other BLSPublicKey) bool {
    return bytes.Equal(b.Bytes(), other.Bytes())
}
```

## Phase 2: Aggregates

### Step 4: GenTx Aggregate

#### 4.1 RED - Write test for factory method

```go
// internal/genutils/domain/gentx_test.go
func TestGenTx_NewGenTx_Valid(t *testing.T) {
    // Arrange
    validatorKey := generateTestKey(t)
    operatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    metadata, _ := domain.NewValidatorMetadata("Test Validator", "", "", "")
    chainID := "stablenet-1"
    timestamp := uint64(time.Now().Unix())

    // Act
    gentx, err := domain.NewGenTx(
        validatorKey,
        operatorAddr,
        metadata,
        chainID,
        timestamp,
    )

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, gentx)
    assert.Equal(t, operatorAddr, gentx.OperatorAddress())
    assert.Equal(t, chainID, gentx.ChainID())
    assert.False(t, gentx.Signature().IsEmpty())
    assert.True(t, gentx.BLSPublicKey().IsValid())
}

func TestGenTx_NewGenTx_InvalidKey(t *testing.T) {
    invalidKey := []byte{0x00} // Invalid key
    operatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    metadata, _ := domain.NewValidatorMetadata("Test", "", "", "")

    _, err := domain.NewGenTx(invalidKey, operatorAddr, metadata, "chain-1", 123)

    assert.Error(t, err)
}
```

#### 4.2 GREEN - Implement aggregate

```go
// internal/genutils/domain/gentx.go
package domain

import (
    "crypto/ecdsa"
    "errors"
    "fmt"
    "time"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/google/uuid"
)

type GenTx struct {
    id               string
    validatorAddress Address
    operatorAddress  Address
    blsPublicKey     BLSPublicKey
    signature        Signature
    metadata         ValidatorMetadata
    timestamp        uint64
    chainID          string
    events           []DomainEvent
}

func NewGenTx(
    validatorKey []byte,
    operatorAddr Address,
    metadata ValidatorMetadata,
    chainID string,
    timestamp uint64,
) (*GenTx, error) {
    // Parse private key
    privKey, err := crypto.ToECDSA(validatorKey)
    if err != nil {
        return nil, fmt.Errorf("invalid validator key: %w", err)
    }

    // Derive validator address
    validatorAddr := crypto.PubkeyToAddress(privKey.PublicKey)

    // Derive BLS public key
    blsKey, err := deriveBLSPublicKey(validatorKey)
    if err != nil {
        return nil, fmt.Errorf("failed to derive BLS key: %w", err)
    }

    // Create signing message
    message := constructSigningMessage(
        Address{value: validatorAddr},
        operatorAddr,
        chainID,
        timestamp,
    )

    // Sign message
    sigBytes, err := crypto.Sign(message, privKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign: %w", err)
    }

    signature, err := NewSignature(sigBytes)
    if err != nil {
        return nil, err
    }

    gentx := &GenTx{
        id:               uuid.New().String(),
        validatorAddress: Address{value: validatorAddr},
        operatorAddress:  operatorAddr,
        blsPublicKey:     blsKey,
        signature:        signature,
        metadata:         metadata,
        timestamp:        timestamp,
        chainID:          chainID,
        events:           []DomainEvent{},
    }

    // Validate
    if err := gentx.validate(); err != nil {
        return nil, err
    }

    // Emit event
    gentx.addEvent(GenTxCreatedEvent{
        baseEvent:        baseEvent{occurredAt: time.Now()},
        ValidatorAddress: gentx.validatorAddress,
        Timestamp:        timestamp,
    })

    return gentx, nil
}

// Getters
func (g *GenTx) ID() string                       { return g.id }
func (g *GenTx) ValidatorAddress() Address        { return g.validatorAddress }
func (g *GenTx) OperatorAddress() Address         { return g.operatorAddress }
func (g *GenTx) BLSPublicKey() BLSPublicKey       { return g.blsPublicKey }
func (g *GenTx) Signature() Signature             { return g.signature }
func (g *GenTx) Metadata() ValidatorMetadata      { return g.metadata }
func (g *GenTx) Timestamp() uint64                { return g.timestamp }
func (g *GenTx) ChainID() string                  { return g.chainID }
func (g *GenTx) DomainEvents() []DomainEvent      { return g.events }

func (g *GenTx) validate() error {
    if g.validatorAddress.IsZero() {
        return errors.New("validator address is zero")
    }
    if g.operatorAddress.IsZero() {
        return errors.New("operator address is zero")
    }
    if !g.blsPublicKey.IsValid() {
        return errors.New("invalid BLS public key")
    }
    if g.signature.IsEmpty() {
        return errors.New("signature is empty")
    }
    if g.chainID == "" {
        return errors.New("chain ID is empty")
    }
    if g.timestamp == 0 {
        return errors.New("timestamp is zero")
    }
    return nil
}

func (g *GenTx) addEvent(event DomainEvent) {
    g.events = append(g.events, event)
}

// Helper functions
func deriveBLSPublicKey(validatorKey []byte) (BLSPublicKey, error) {
    // Derive BLS key from validator key
    // This is simplified - real implementation would use proper BLS key derivation
    hash := crypto.Keccak256(validatorKey)

    // Convert to BLS G2 point (this is a placeholder)
    // Real implementation would use proper BLS key generation
    point := &bls12381.PointG2{}
    // ... BLS key generation logic ...

    return BLSPublicKey{point: point}, nil
}

func constructSigningMessage(
    validatorAddr, operatorAddr Address,
    chainID string,
    timestamp uint64,
) []byte {
    message := fmt.Sprintf(
        "GenTx Registration\nChain: %s\nValidator: %s\nOperator: %s\nTimestamp: %d",
        chainID,
        validatorAddr.String(),
        operatorAddr.String(),
        timestamp,
    )
    return crypto.Keccak256([]byte(message))
}
```

#### 4.3 RED - Write test for signature verification

```go
func TestGenTx_VerifySignature(t *testing.T) {
    // Arrange
    crypto := ethereum.NewCryptoProvider()
    validatorKey := generateTestKey(t)
    operatorAddr, _ := domain.NewAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    metadata, _ := domain.NewValidatorMetadata("Test", "", "", "")

    gentx, _ := domain.NewGenTx(
        validatorKey,
        operatorAddr,
        metadata,
        "chain-1",
        uint64(time.Now().Unix()),
    )

    // Act
    err := gentx.VerifySignature(crypto)

    // Assert
    assert.NoError(t, err)
}
```

#### 4.4 GREEN - Implement verification

```go
// Add to gentx.go
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
        baseEvent:        baseEvent{occurredAt: time.Now()},
        ValidatorAddress: g.validatorAddress,
    })

    return nil
}
```

## Phase 3: Domain Services

### Step 5: Validation Service

#### 5.1 RED - Write test

```go
// internal/genutils/service/validation/service_test.go
func TestValidationService_Validate_ValidGenTx(t *testing.T) {
    // Arrange
    crypto := mock.NewCryptoProvider()
    service := validation.NewValidationService(crypto)
    gentx := createValidGenTx(t)

    // Act
    err := service.Validate(gentx)

    // Assert
    assert.NoError(t, err)
}

func TestValidationService_Validate_InvalidSignature(t *testing.T) {
    // Arrange
    crypto := mock.NewCryptoProvider()
    crypto.SetShouldFailSignature(true) // Mock will fail signature check
    service := validation.NewValidationService(crypto)
    gentx := createValidGenTx(t)

    // Act
    err := service.Validate(gentx)

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "signature")
}
```

#### 5.2 GREEN - Implement service

```go
// internal/genutils/service/validation/service.go
package validation

import (
    "fmt"
    "go-stablenet/internal/genutils/domain"
)

type ValidationService struct {
    signatureValidator *SignatureValidator
    formatValidator    *FormatValidator
    businessValidator  *BusinessRuleValidator
}

func NewValidationService(crypto domain.CryptoProvider) *ValidationService {
    return &ValidationService{
        signatureValidator: NewSignatureValidator(crypto),
        formatValidator:    NewFormatValidator(),
        businessValidator:  NewBusinessRuleValidator(),
    }
}

func (s *ValidationService) Validate(gentx *domain.GenTx) error {
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

#### 5.3 Implement individual validators

```go
// internal/genutils/service/validation/signature_validator.go
type SignatureValidator struct {
    crypto domain.CryptoProvider
}

func NewSignatureValidator(crypto domain.CryptoProvider) *SignatureValidator {
    return &SignatureValidator{crypto: crypto}
}

func (v *SignatureValidator) Validate(gentx *domain.GenTx) error {
    return gentx.VerifySignature(v.crypto)
}

// internal/genutils/service/validation/format_validator.go
type FormatValidator struct{}

func NewFormatValidator() *FormatValidator {
    return &FormatValidator{}
}

func (v *FormatValidator) Validate(gentx *domain.GenTx) error {
    if gentx.ValidatorAddress().IsZero() {
        return domain.ErrInvalidValidatorAddress
    }
    if gentx.OperatorAddress().IsZero() {
        return domain.ErrInvalidOperatorAddress
    }
    if !gentx.BLSPublicKey().IsValid() {
        return domain.ErrInvalidBLSKey
    }
    return nil
}

// internal/genutils/service/validation/business_validator.go
type BusinessRuleValidator struct{}

func NewBusinessRuleValidator() *BusinessRuleValidator {
    return &BusinessRuleValidator{}
}

func (v *BusinessRuleValidator) Validate(gentx *domain.GenTx) error {
    // Check timestamp is not too far in the past or future
    now := uint64(time.Now().Unix())
    if gentx.Timestamp() > now+3600 { // 1 hour in future
        return errors.New("timestamp is too far in the future")
    }
    if gentx.Timestamp() < now-86400*7 { // 7 days in past
        return errors.New("timestamp is too old")
    }

    // Validate metadata
    if gentx.Metadata().Name() == "" {
        return errors.New("validator name is required")
    }

    return nil
}
```

## Phase 4: Application Layer (Use Cases)

### Step 6: Create GenTx Use Case

#### 6.1 RED - Write test

```go
// internal/genutils/application/create_gentx_test.go
func TestCreateGenTxUseCase_Execute_Success(t *testing.T) {
    // Arrange
    crypto := mock.NewCryptoProvider()
    repo := memory.NewRepository()
    validator := mock.NewValidator()
    keyManager := mock.NewKeyManager()

    useCase := application.NewCreateGenTxUseCase(
        crypto,
        repo,
        validator,
        keyManager,
    )

    request := application.CreateGenTxRequest{
        ValidatorKeyFile: "validator.key",
        OperatorAddress:  "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        ValidatorName:    "Test Validator",
        ChainID:          "stablenet-1",
    }

    // Act
    response, err := useCase.Execute(request)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, response.GenTxID)
    assert.NotEmpty(t, response.ValidatorAddress)

    // Verify saved
    gentxs, _ := repo.FindAll()
    assert.Len(t, gentxs, 1)
}
```

#### 6.2 GREEN - Implement use case

```go
// internal/genutils/application/create_gentx.go
package application

import (
    "fmt"
    "time"
    "go-stablenet/internal/genutils/domain"
)

type CreateGenTxRequest struct {
    ValidatorKeyFile string
    OperatorAddress  string
    ValidatorName    string
    Description      string
    Website          string
    Contact          string
    ChainID          string
}

type CreateGenTxResponse struct {
    GenTxID          string
    ValidatorAddress string
    OperatorAddress  string
    BLSPublicKey     string
}

type CreateGenTxUseCase struct {
    cryptoProvider domain.CryptoProvider
    repository     domain.GenTxRepository
    validator      domain.GenTxValidator
    keyManager     domain.KeyManager
}

func NewCreateGenTxUseCase(
    crypto domain.CryptoProvider,
    repo domain.GenTxRepository,
    validator domain.GenTxValidator,
    keyManager domain.KeyManager,
) *CreateGenTxUseCase {
    return &CreateGenTxUseCase{
        cryptoProvider: crypto,
        repository:     repo,
        validator:      validator,
        keyManager:     keyManager,
    }
}

func (uc *CreateGenTxUseCase) Execute(
    req CreateGenTxRequest,
) (*CreateGenTxResponse, error) {
    // Load validator key
    validatorKey, err := uc.keyManager.LoadKey(req.ValidatorKeyFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load validator key: %w", err)
    }

    // Parse operator address
    operatorAddr, err := domain.NewAddress(req.OperatorAddress)
    if err != nil {
        return nil, fmt.Errorf("invalid operator address: %w", err)
    }

    // Create metadata
    metadata, err := domain.NewValidatorMetadata(
        req.ValidatorName,
        req.Description,
        req.Website,
        req.Contact,
    )
    if err != nil {
        return nil, fmt.Errorf("invalid metadata: %w", err)
    }

    // Create GenTx
    gentx, err := domain.NewGenTx(
        validatorKey,
        operatorAddr,
        metadata,
        req.ChainID,
        uint64(time.Now().Unix()),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create gentx: %w", err)
    }

    // Validate
    if err := uc.validator.Validate(gentx); err != nil {
        return nil, fmt.Errorf("gentx validation failed: %w", err)
    }

    // Save
    if err := uc.repository.Save(gentx); err != nil {
        return nil, fmt.Errorf("failed to save gentx: %w", err)
    }

    // Return response
    return &CreateGenTxResponse{
        GenTxID:          gentx.ID(),
        ValidatorAddress: gentx.ValidatorAddress().String(),
        OperatorAddress:  gentx.OperatorAddress().String(),
        BLSPublicKey:     gentx.BLSPublicKey().String(),
    }, nil
}
```

## Phase 5: Infrastructure Layer

### Step 7: File Repository

#### 7.1 RED - Write test

```go
// internal/genutils/repository/file_repository_test.go
func TestFileRepository_Save_Success(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    repo := repository.NewFileRepository(tmpDir)
    gentx := createValidGenTx(t)

    // Act
    err := repo.Save(gentx)

    // Assert
    assert.NoError(t, err)

    // Verify file exists
    expectedPath := filepath.Join(tmpDir, fmt.Sprintf("gentx-%s.json", gentx.ValidatorAddress()))
    assert.FileExists(t, expectedPath)
}

func TestFileRepository_FindAll_Success(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    repo := repository.NewFileRepository(tmpDir)
    gentx1 := createValidGenTx(t)
    gentx2 := createValidGenTx(t)
    repo.Save(gentx1)
    repo.Save(gentx2)

    // Act
    gentxs, err := repo.FindAll()

    // Assert
    assert.NoError(t, err)
    assert.Len(t, gentxs, 2)
}
```

#### 7.2 GREEN - Implement repository

```go
// internal/genutils/repository/file_repository.go
package repository

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "go-stablenet/internal/genutils/domain"
)

type FileRepository struct {
    baseDir string
}

func NewFileRepository(baseDir string) *FileRepository {
    return &FileRepository{baseDir: baseDir}
}

func (r *FileRepository) Save(gentx *domain.GenTx) error {
    // Create directory if not exists
    if err := os.MkdirAll(r.baseDir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // Serialize to JSON
    data, err := r.serializeGenTx(gentx)
    if err != nil {
        return fmt.Errorf("failed to serialize: %w", err)
    }

    // Write to file
    filename := fmt.Sprintf("gentx-%s.json", gentx.ValidatorAddress())
    path := filepath.Join(r.baseDir, filename)

    if err := ioutil.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }

    return nil
}

func (r *FileRepository) FindAll() ([]*domain.GenTx, error) {
    files, err := ioutil.ReadDir(r.baseDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory: %w", err)
    }

    gentxs := make([]*domain.GenTx, 0)

    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
            path := filepath.Join(r.baseDir, file.Name())
            data, err := ioutil.ReadFile(path)
            if err != nil {
                return nil, fmt.Errorf("failed to read file %s: %w", file.Name(), err)
            }

            gentx, err := r.deserializeGenTx(data)
            if err != nil {
                return nil, fmt.Errorf("failed to deserialize %s: %w", file.Name(), err)
            }

            gentxs = append(gentxs, gentx)
        }
    }

    return gentxs, nil
}

func (r *FileRepository) serializeGenTx(gentx *domain.GenTx) ([]byte, error) {
    dto := GenTxDTO{
        ValidatorAddress: gentx.ValidatorAddress().String(),
        OperatorAddress:  gentx.OperatorAddress().String(),
        BLSPublicKey:     gentx.BLSPublicKey().String(),
        Signature:        fmt.Sprintf("0x%x", gentx.Signature().Bytes()),
        Metadata: MetadataDTO{
            Name:        gentx.Metadata().Name(),
            Description: gentx.Metadata().Description(),
            Website:     gentx.Metadata().Website(),
            Contact:     gentx.Metadata().Contact(),
        },
        Timestamp: gentx.Timestamp(),
        ChainID:   gentx.ChainID(),
    }

    return json.MarshalIndent(dto, "", "  ")
}

func (r *FileRepository) deserializeGenTx(data []byte) (*domain.GenTx, error) {
    var dto GenTxDTO
    if err := json.Unmarshal(data, &dto); err != nil {
        return nil, err
    }

    // Reconstruct domain object
    // Note: This is simplified - real implementation would need to
    // reconstruct all value objects properly
    // ...

    return gentx, nil
}

// DTOs for serialization
type GenTxDTO struct {
    ValidatorAddress string       `json:"validator_address"`
    OperatorAddress  string       `json:"operator_address"`
    BLSPublicKey     string       `json:"bls_public_key"`
    Signature        string       `json:"signature"`
    Metadata         MetadataDTO  `json:"metadata"`
    Timestamp        uint64       `json:"timestamp"`
    ChainID          string       `json:"chain_id"`
}

type MetadataDTO struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Website     string `json:"website"`
    Contact     string `json:"contact"`
}
```

## Phase 6: Presentation Layer (CLI)

### Step 8: CLI Command

#### 8.1 RED - Write integration test

```go
// cmd/gstable/gentx_create_test.go
func TestGenTxCreateCommand_Success(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    keyFile := filepath.Join(tmpDir, "validator.key")
    gentxDir := filepath.Join(tmpDir, "gentxs")

    // Generate test key
    generateTestKeyFile(t, keyFile)

    // Act
    cmd := exec.Command("gstable", "gentx", "create",
        "--validator-key", keyFile,
        "--operator", "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        "--name", "Test Validator",
        "--chain-id", "stablenet-1",
        "--output-dir", gentxDir,
    )

    output, err := cmd.CombinedOutput()

    // Assert
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Genesis transaction created successfully")

    // Verify file created
    files, _ := ioutil.ReadDir(gentxDir)
    assert.Len(t, files, 1)
}
```

#### 8.2 GREEN - Implement CLI

```go
// cmd/gstable/gentx_create.go
package main

import (
    "fmt"
    "github.com/urfave/cli/v2"
    "go-stablenet/internal/genutils/application"
    "go-stablenet/internal/genutils/repository"
    "go-stablenet/internal/genutils/service/validation"
    "go-stablenet/pkg/crypto/ethereum"
    "go-stablenet/pkg/keystore"
)

var gentxCreateCommand = &cli.Command{
    Name:  "create",
    Usage: "Create a genesis transaction",
    Flags: []cli.Flag{
        &cli.StringFlag{
            Name:     "validator-key",
            Usage:    "Path to validator private key file",
            Required: true,
        },
        &cli.StringFlag{
            Name:     "operator",
            Usage:    "Operator address (can be multisig)",
            Required: true,
        },
        &cli.StringFlag{
            Name:     "name",
            Usage:    "Validator name",
            Required: true,
        },
        &cli.StringFlag{
            Name:  "description",
            Usage: "Validator description",
        },
        &cli.StringFlag{
            Name:  "website",
            Usage: "Validator website",
        },
        &cli.StringFlag{
            Name:  "contact",
            Usage: "Contact information",
        },
        &cli.StringFlag{
            Name:     "chain-id",
            Usage:    "Chain ID",
            Required: true,
        },
        &cli.StringFlag{
            Name:  "output-dir",
            Usage: "Output directory for gentx file",
            Value: "./gentxs",
        },
    },
    Action: func(ctx *cli.Context) error {
        return executeCreateGenTx(ctx)
    },
}

func executeCreateGenTx(ctx *cli.Context) error {
    // Build dependencies
    crypto := ethereum.NewCryptoProvider()
    repo := repository.NewFileRepository(ctx.String("output-dir"))
    validator := validation.NewValidationService(crypto)
    keyManager := keystore.NewManager()

    // Create use case
    useCase := application.NewCreateGenTxUseCase(
        crypto,
        repo,
        validator,
        keyManager,
    )

    // Build request
    request := application.CreateGenTxRequest{
        ValidatorKeyFile: ctx.String("validator-key"),
        OperatorAddress:  ctx.String("operator"),
        ValidatorName:    ctx.String("name"),
        Description:      ctx.String("description"),
        Website:          ctx.String("website"),
        Contact:          ctx.String("contact"),
        ChainID:          ctx.String("chain-id"),
    }

    // Execute
    response, err := useCase.Execute(request)
    if err != nil {
        return fmt.Errorf("failed to create gentx: %w", err)
    }

    // Print result
    fmt.Printf("Genesis transaction created successfully!\n")
    fmt.Printf("  GenTx ID: %s\n", response.GenTxID)
    fmt.Printf("  Validator Address: %s\n", response.ValidatorAddress)
    fmt.Printf("  Operator Address: %s\n", response.OperatorAddress)
    fmt.Printf("  BLS Public Key: %s\n", response.BLSPublicKey)

    return nil
}
```

## Testing Strategy Summary

### Unit Tests
- Test each component in isolation
- Use mocks for dependencies
- Fast execution (<1s)

### Integration Tests
- Test component interactions
- Use real implementations where possible
- Test repository with file system

### End-to-End Tests
- Test complete workflows
- Use actual CLI commands
- Verify file outputs

### Test Coverage Goals
- Domain layer: 95%+
- Application layer: 90%+
- Infrastructure layer: 80%+
- Overall: 85%+

## Summary

This implementation guide demonstrates:

1. **TDD Discipline**: Write tests first, implement minimum code, refactor
2. **DDD Focus**: Build from domain outward (inside-out)
3. **SOLID Compliance**: Each component has single responsibility
4. **Type Safety**: Strong typing prevents invalid states
5. **Testability**: Clean architecture enables easy testing

Next: See [05_data_structures.md](./05_data_structures.md) for file format specifications.

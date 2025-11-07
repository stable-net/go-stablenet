# Integration Plan

## Overview

This document outlines the step-by-step integration plan for implementing GenUtils in the go-stablenet codebase. The plan follows an iterative approach with clear milestones and deliverables.

## Prerequisites

### Development Environment

- Go 1.21+
- Git
- Make
- Docker (for E2E testing)

### Dependencies

```go
// Required dependencies
github.com/ethereum/go-ethereum v1.13.0+
github.com/urfave/cli/v2 v2.25.0+
github.com/stretchr/testify v1.8.4+
github.com/google/uuid v1.5.0+

// Testing
github.com/golang/mock v1.6.0+
```

## Phase 1: Foundation (Week 1-2)

### Milestone 1.1: Project Structure Setup

**Goal**: Create directory structure and basic scaffolding

**Tasks**:

1. Create directory structure:
```bash
mkdir -p internal/genutils/{domain,repository,service,application}
mkdir -p internal/genutils/service/{validation,collection,genesis}
mkdir -p pkg/crypto/ethereum
mkdir -p pkg/keystore
mkdir -p cmd/gstable/gentx
mkdir -p docs/genutils
```

2. Initialize Go modules:
```bash
cd internal/genutils
go mod init go-stablenet/internal/genutils
```

3. Setup testing infrastructure:
```bash
mkdir -p internal/genutils/testutil
mkdir -p internal/genutils/mock
```

**Deliverables**:
- ✅ Directory structure created
- ✅ Go modules initialized
- ✅ CI/CD pipeline configured for tests

**Validation**:
```bash
go test ./internal/genutils/... # Should pass (no tests yet)
```

### Milestone 1.2: Domain Layer - Value Objects

**Goal**: Implement core value objects with TDD

**Implementation Order**:

1. **Address** (2 hours)
   - Test: `internal/genutils/domain/address_test.go`
   - Impl: `internal/genutils/domain/address.go`
   - Coverage target: 95%

2. **Signature** (2 hours)
   - Test: `internal/genutils/domain/signature_test.go`
   - Impl: `internal/genutils/domain/signature.go`
   - Coverage target: 95%

3. **BLSPublicKey** (3 hours)
   - Test: `internal/genutils/domain/bls_public_key_test.go`
   - Impl: `internal/genutils/domain/bls_public_key.go`
   - Coverage target: 95%

4. **ValidatorMetadata** (2 hours)
   - Test: `internal/genutils/domain/validator_metadata_test.go`
   - Impl: `internal/genutils/domain/validator_metadata.go`
   - Coverage target: 95%

**Deliverables**:
- ✅ All value objects implemented
- ✅ Unit tests passing
- ✅ Coverage ≥95%

**Validation**:
```bash
go test -v -cover ./internal/genutils/domain/
```

### Milestone 1.3: Domain Layer - Aggregates

**Goal**: Implement GenTx and GenTxCollection aggregates

**Implementation Order**:

1. **Domain Events** (1 hour)
   - Impl: `internal/genutils/domain/events.go`
   - Define all event types

2. **GenTx Aggregate** (8 hours)
   - Test: `internal/genutils/domain/gentx_test.go`
   - Impl: `internal/genutils/domain/gentx.go`
   - Factory method
   - Signature verification
   - Validation logic
   - Coverage target: 95%

3. **GenTxCollection Aggregate** (6 hours)
   - Test: `internal/genutils/domain/gentx_collection_test.go`
   - Impl: `internal/genutils/domain/gentx_collection.go`
   - Add/remove gentxs
   - Uniqueness constraints
   - Sorting logic
   - Coverage target: 95%

**Deliverables**:
- ✅ GenTx aggregate fully tested
- ✅ GenTxCollection aggregate fully tested
- ✅ Domain events defined
- ✅ Coverage ≥95%

**Validation**:
```bash
go test -v -cover ./internal/genutils/domain/
```

## Phase 2: Services (Week 3-4)

### Milestone 2.1: Infrastructure - Crypto Provider

**Goal**: Implement cryptographic operations

**Tasks**:

1. **Crypto Interface** (2 hours)
   - Impl: `internal/genutils/domain/crypto.go`
   - Define CryptoProvider interface

2. **Ethereum Crypto Implementation** (6 hours)
   - Test: `pkg/crypto/ethereum/provider_test.go`
   - Impl: `pkg/crypto/ethereum/provider.go`
   - Sign/Verify/RecoverAddress
   - BLS key derivation
   - Coverage target: 90%

3. **Mock Crypto** (2 hours)
   - Impl: `internal/genutils/mock/crypto.go`
   - For testing

**Deliverables**:
- ✅ CryptoProvider interface
- ✅ Ethereum implementation
- ✅ Mock implementation
- ✅ Coverage ≥90%

**Validation**:
```bash
go test -v -cover ./pkg/crypto/ethereum/
```

### Milestone 2.2: Infrastructure - Repository

**Goal**: Implement persistence layer

**Tasks**:

1. **Repository Interface** (1 hour)
   - Impl: `internal/genutils/repository/interface.go`

2. **File Repository** (8 hours)
   - Test: `internal/genutils/repository/file_repository_test.go`
   - Impl: `internal/genutils/repository/file_repository.go`
   - Save/Load/FindAll
   - JSON serialization
   - Error handling
   - Coverage target: 85%

3. **Memory Repository** (4 hours)
   - Test: `internal/genutils/repository/memory_repository_test.go`
   - Impl: `internal/genutils/repository/memory_repository.go`
   - For testing
   - Coverage target: 90%

**Deliverables**:
- ✅ Repository interface
- ✅ File implementation
- ✅ Memory implementation
- ✅ Coverage ≥85%

**Validation**:
```bash
go test -v -cover ./internal/genutils/repository/
```

### Milestone 2.3: Domain Services

**Goal**: Implement validation and collection services

**Tasks**:

1. **Validation Service** (10 hours)
   - Test: `internal/genutils/service/validation/service_test.go`
   - Impl: `internal/genutils/service/validation/service.go`
   - Impl: `internal/genutils/service/validation/signature_validator.go`
   - Impl: `internal/genutils/service/validation/format_validator.go`
   - Impl: `internal/genutils/service/validation/business_validator.go`
   - Coverage target: 90%

2. **Collection Service** (8 hours)
   - Test: `internal/genutils/service/collection/service_test.go`
   - Impl: `internal/genutils/service/collection/service.go`
   - Impl: `internal/genutils/service/collection/merger.go`
   - Coverage target: 90%

3. **Genesis Builder Service** (10 hours)
   - Test: `internal/genutils/service/genesis/builder_test.go`
   - Impl: `internal/genutils/service/genesis/builder.go`
   - Impl: `internal/genutils/service/genesis/contract_injector.go`
   - Impl: `internal/genutils/service/genesis/wbft_configurator.go`
   - Coverage target: 85%

**Deliverables**:
- ✅ Validation service implemented
- ✅ Collection service implemented
- ✅ Genesis builder service implemented
- ✅ Coverage ≥85%

**Validation**:
```bash
go test -v -cover ./internal/genutils/service/...
```

## Phase 3: Application Layer (Week 5)

### Milestone 3.1: Use Cases

**Goal**: Implement application use cases

**Tasks**:

1. **Create GenTx Use Case** (8 hours)
   - Test: `internal/genutils/application/create_gentx_test.go`
   - Impl: `internal/genutils/application/create_gentx.go`
   - Coverage target: 90%

2. **Validate GenTx Use Case** (4 hours)
   - Test: `internal/genutils/application/validate_gentx_test.go`
   - Impl: `internal/genutils/application/validate_gentx.go`
   - Coverage target: 90%

3. **Collect GenTxs Use Case** (8 hours)
   - Test: `internal/genutils/application/collect_gentxs_test.go`
   - Impl: `internal/genutils/application/collect_gentxs.go`
   - Coverage target: 90%

**Deliverables**:
- ✅ All use cases implemented
- ✅ Integration tests passing
- ✅ Coverage ≥90%

**Validation**:
```bash
go test -v -cover ./internal/genutils/application/
```

## Phase 4: CLI Integration (Week 6)

### Milestone 4.1: CLI Commands

**Goal**: Implement CLI commands

**Tasks**:

1. **GenTx Create Command** (8 hours)
   - Test: `cmd/gstable/gentx/create_test.go`
   - Impl: `cmd/gstable/gentx/create.go`
   - Impl: `cmd/gstable/gentx/flags.go`

2. **GenTx Validate Command** (4 hours)
   - Test: `cmd/gstable/gentx/validate_test.go`
   - Impl: `cmd/gstable/gentx/validate.go`

3. **GenTx Collect Command** (8 hours)
   - Test: `cmd/gstable/gentx/collect_test.go`
   - Impl: `cmd/gstable/gentx/collect.go`

4. **GenTx Inspect Command** (4 hours)
   - Test: `cmd/gstable/gentx/inspect_test.go`
   - Impl: `cmd/gstable/gentx/inspect.go`

**Deliverables**:
- ✅ All CLI commands implemented
- ✅ Help text and examples
- ✅ Integration tests passing

**Validation**:
```bash
go test -v ./cmd/gstable/gentx/...
gstable gentx --help
```

### Milestone 4.2: Main Binary Integration

**Goal**: Integrate gentx commands into main gstable binary

**Tasks**:

1. Register commands in `cmd/gstable/main.go`:
```go
app.Commands = append(app.Commands, &cli.Command{
    Name:  "gentx",
    Usage: "Genesis transaction operations",
    Subcommands: []*cli.Command{
        gentx.CreateCommand,
        gentx.ValidateCommand,
        gentx.CollectCommand,
        gentx.InspectCommand,
    },
})
```

2. Update Makefile:
```makefile
.PHONY: genutils-test
genutils-test:
	go test -v -cover ./internal/genutils/...

.PHONY: gentx-e2e
gentx-e2e:
	./scripts/test-gentx-e2e.sh
```

**Deliverables**:
- ✅ Commands registered
- ✅ Build succeeds
- ✅ Commands accessible via CLI

**Validation**:
```bash
make build
./build/bin/gstable gentx --help
```

## Phase 5: Core Integration (Week 7-8)

### Milestone 5.1: Genesis Integration

**Goal**: Integrate gentx into genesis block creation

**Tasks**:

1. **Extend Genesis Structure** (4 hours)
   - Modify: `core/genesis.go`
   - Add `AnzeonGenTxs` field
   - Add parsing logic

2. **GenTx Processing** (8 hours)
   - Impl: `core/genesis_gentx.go`
   - `ProcessGenTxs()` function
   - Extract validators/BLS keys
   - Update anzeon.init

3. **System Contract Bootstrap** (10 hours)
   - Modify: `systemcontracts/gov_validator.go`
   - Add `BootstrapFromGenTxs()` function
   - Initialize validator state from gentxs

4. **Genesis Setup Integration** (8 hours)
   - Modify: `core/genesis.go::SetupGenesisBlock()`
   - Call `ProcessGenTxs()` before contract injection
   - Validate gentx data

**Integration Points**:

```go
// core/genesis.go
type Genesis struct {
    // ... existing fields ...
    AnzeonGenTxs []GenTxData `json:"anzeonGenTxs,omitempty"`
}

func (g *Genesis) ProcessGenTxs() error {
    // Extract validators
    validators := make([]common.Address, len(g.AnzeonGenTxs))
    blsKeys := make([]string, len(g.AnzeonGenTxs))

    for i, gentx := range g.AnzeonGenTxs {
        validators[i] = common.HexToAddress(gentx.ValidatorAddress)
        blsKeys[i] = gentx.BLSPublicKey
    }

    // Update anzeon config
    g.Config.Anzeon.Init.Validators = validators
    g.Config.Anzeon.Init.BLSPublicKeys = blsKeys

    return nil
}

func SetupGenesisBlock(...) {
    // ... existing code ...

    // Process gentxs if present
    if len(genesis.AnzeonGenTxs) > 0 {
        if err := genesis.ProcessGenTxs(); err != nil {
            return nil, common.Hash{}, err
        }
    }

    // ... continue with existing logic ...
}
```

**Deliverables**:
- ✅ Genesis extended with gentx support
- ✅ GenTx processing implemented
- ✅ System contracts bootstrapped from gentxs
- ✅ Integration tests passing

**Validation**:
```bash
# Unit tests
go test -v ./core/

# Integration test
gstable init genesis-with-gentxs.json --datadir /tmp/test-node
```

### Milestone 5.2: WBFT Integration

**Goal**: Integrate gentx-derived validators into WBFT consensus

**Tasks**:

1. **Validator Set Initialization** (6 hours)
   - Modify: `consensus/wbft/engine/engine.go`
   - Ensure genesis validators recognized

2. **BLS Key Validation** (4 hours)
   - Modify: `consensus/wbft/core/validator_set.go`
   - Validate BLS keys at startup

3. **Testing** (8 hours)
   - Test validator set initialization
   - Test BLS key verification
   - Test consensus startup

**Deliverables**:
- ✅ WBFT recognizes gentx validators
- ✅ BLS keys properly validated
- ✅ Consensus starts correctly

**Validation**:
```bash
# Start 3-node network with gentx-generated genesis
./scripts/test-multi-node-gentx.sh
```

## Phase 6: Testing & Documentation (Week 9-10)

### Milestone 6.1: End-to-End Testing

**Goal**: Comprehensive E2E test suite

**Tasks**:

1. **Single Validator Test** (4 hours)
   - Script: `scripts/test-gentx-single-validator.sh`
   - Create gentx → collect → init → start

2. **Multi-Validator Test** (8 hours)
   - Script: `scripts/test-gentx-multi-validator.sh`
   - 3+ validators
   - Distributed gentx creation
   - Centralized collection
   - Network startup

3. **Error Scenarios** (8 hours)
   - Invalid signatures
   - Duplicate validators
   - Mismatched chain IDs
   - Corrupted files

4. **Performance Testing** (4 hours)
   - Large validator sets (100+)
   - Collection time
   - Genesis file size

**Deliverables**:
- ✅ E2E test suite
- ✅ All scenarios covered
- ✅ Performance benchmarks
- ✅ CI/CD integration

**Validation**:
```bash
make gentx-e2e
```

### Milestone 6.2: Documentation

**Goal**: Complete user and developer documentation

**Tasks**:

1. **User Guide** (8 hours)
   - `docs/gentx/user-guide.md`
   - Step-by-step tutorials
   - Common workflows
   - Troubleshooting

2. **Developer Guide** (6 hours)
   - `docs/gentx/developer-guide.md`
   - Architecture overview
   - API reference
   - Extension points

3. **CLI Reference** (4 hours)
   - `docs/gentx/cli-reference.md`
   - All commands
   - Flags and options
   - Examples

4. **Migration Guide** (4 hours)
   - `docs/gentx/migration-guide.md`
   - From manual setup
   - Backward compatibility

**Deliverables**:
- ✅ Complete documentation
- ✅ Examples and tutorials
- ✅ API reference
- ✅ Migration guide

## Phase 7: Polish & Release (Week 11-12)

### Milestone 7.1: Code Review & Refactoring

**Goal**: Ensure code quality

**Tasks**:

1. **Code Review** (16 hours)
   - SOLID principles adherence
   - DDD patterns correctness
   - Test coverage gaps
   - Performance optimization

2. **Refactoring** (16 hours)
   - Address code review feedback
   - Simplify complex logic
   - Improve error messages
   - Optimize performance

3. **Security Audit** (8 hours)
   - Private key handling
   - Signature verification
   - Input validation
   - File permissions

**Deliverables**:
- ✅ Code reviewed
- ✅ Issues addressed
- ✅ Security validated
- ✅ Performance optimized

### Milestone 7.2: Release Preparation

**Goal**: Prepare for production release

**Tasks**:

1. **Release Notes** (4 hours)
   - Changelog
   - Breaking changes
   - Migration instructions

2. **Version Tagging** (2 hours)
   - Git tag: v1.0.0-gentx
   - Release branch

3. **Binary Distribution** (4 hours)
   - Build artifacts
   - Docker images
   - Installation instructions

**Deliverables**:
- ✅ Release notes
- ✅ Tagged version
- ✅ Binaries available
- ✅ Installation guide

## Rollout Strategy

### Stage 1: Internal Testing (Week 13)
- Deploy to internal testnet
- 5 internal validators
- Monitor for issues
- Gather feedback

### Stage 2: Public Testnet (Week 14-15)
- Deploy to public testnet
- 20+ external validators
- Community testing
- Bug fixes

### Stage 3: Mainnet Preparation (Week 16)
- Finalize documentation
- Security audit completion
- Mainnet genesis coordination
- Go-live plan

### Stage 4: Mainnet Launch (Week 17+)
- Coordinate validator gentx submissions
- Collect and validate all gentxs
- Build final genesis
- Network launch

## Success Criteria

### Functional Requirements
- ✅ Validators can independently create gentxs
- ✅ Coordinator can collect and merge gentxs
- ✅ Genesis file correctly initializes validators
- ✅ Network starts with all validators participating
- ✅ No manual validator configuration required

### Non-Functional Requirements
- ✅ Test coverage ≥85%
- ✅ Zero critical bugs
- ✅ Documentation complete
- ✅ CI/CD pipeline green
- ✅ Performance acceptable (100 validators in <5min)

### Quality Gates
- ✅ Code review approval (2+ reviewers)
- ✅ Security audit pass
- ✅ All E2E tests passing
- ✅ Community feedback positive

## Risk Mitigation

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Genesis compatibility issues | High | Medium | Extensive testing, backward compatibility |
| BLS key derivation bugs | High | Low | Unit tests, crypto audit |
| Performance issues | Medium | Medium | Benchmarking, optimization |
| Data corruption | High | Low | Validation, backups |

### Schedule Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Underestimated complexity | High | Buffer time, iterative approach |
| Dependency delays | Medium | Parallel work, early integration |
| Testing coverage gaps | High | TDD from start |

## Maintenance Plan

### Post-Launch Support
- Bug fix releases: Within 24 hours for critical
- Minor releases: Monthly for 3 months
- Documentation updates: Ongoing
- Community support: Discord/Telegram

### Long-Term Evolution
- Version 2.0 planning (6 months post-launch)
- Additional features based on feedback
- Performance improvements
- Security enhancements

## Summary

This integration plan provides:

1. **Clear Roadmap**: 12-week implementation schedule
2. **Incremental Delivery**: Working software at each milestone
3. **Quality Focus**: TDD, code review, testing
4. **Risk Management**: Identified risks and mitigations
5. **Success Criteria**: Measurable objectives

**Total Estimated Effort**: ~400 hours over 12 weeks

**Team Size**: 2-3 developers (full-time equivalent)

**Start Date**: TBD
**Target Launch**: 17 weeks from start

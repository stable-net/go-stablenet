# GenUtils Documentation

## Overview

GenUtils is a comprehensive genesis transaction system for Go-StableNet, inspired by Cosmos SDK's genutil module but adapted for Ethereum-based blockchain with WBFT consensus.

This documentation suite provides complete design specifications, implementation guides, and integration plans following SOLID principles, Domain-Driven Design (DDD), and Test-Driven Development (TDD) methodologies.

## Documentation Index

### 1. [Overview](./01_overview.md)
**Purpose**: Introduction and high-level architecture

**Topics**:
- What is GenUtils and why it's needed
- Key differences from Cosmos SDK
- Core workflows (create → collect → genesis)
- Component overview
- Security considerations

**Audience**: Everyone - start here!

**Estimated Reading Time**: 15-20 minutes

---

### 2. [Architecture](./02_architecture.md)
**Purpose**: Detailed architectural design following SOLID principles

**Topics**:
- Layered architecture (Domain → Application → Infrastructure → Presentation)
- SOLID principles application with code examples
- Component diagram and dependencies
- Domain-Driven Design patterns
- Clean Architecture principles
- Dependency injection strategies

**Audience**: Developers, architects

**Estimated Reading Time**: 30-40 minutes

**Key Takeaways**:
- How SOLID principles are applied
- Why each architectural decision was made
- Component responsibilities and interactions

---

### 3. [Domain Model](./03_domain_model.md)
**Purpose**: Comprehensive domain model using DDD

**Topics**:
- Bounded contexts (GenTx Creation, Validation, Collection, Genesis Building)
- Aggregates (GenTx, GenTxCollection)
- Entities (Validator)
- Value objects (Address, Signature, BLSPublicKey, ValidatorMetadata)
- Domain services (Validation, Collection, Genesis Builder)
- Domain events
- Repository interfaces

**Audience**: Developers implementing domain logic

**Estimated Reading Time**: 40-50 minutes

**Key Takeaways**:
- What are the core domain concepts
- How domain invariants are enforced
- Why aggregates are structured this way

---

### 4. [Implementation Guide](./04_implementation_guide.md)
**Purpose**: Step-by-step TDD + DDD implementation approach

**Topics**:
- TDD cycle: Red → Green → Refactor
- Implementation order (inside-out: Domain → Application → Infrastructure → Presentation)
- Detailed code examples with tests
- Phase 1: Value Objects
- Phase 2: Aggregates
- Phase 3: Domain Services
- Phase 4: Application Layer (Use Cases)
- Phase 5: Infrastructure Layer (Repository, Crypto)
- Phase 6: Presentation Layer (CLI)

**Audience**: Developers actively coding the system

**Estimated Reading Time**: 60-90 minutes

**Key Takeaways**:
- How to write tests first
- How to implement minimum viable code
- How to refactor with confidence

---

### 5. [Data Structures](./05_data_structures.md)
**Purpose**: File formats, schemas, and serialization

**Topics**:
- GenTx JSON format specification
- Genesis JSON format extensions
- Go data structures
- Serialization/Deserialization
- File system layout
- JSON schemas for validation
- Version migration strategy

**Audience**: Developers working on persistence, integration, or debugging

**Estimated Reading Time**: 30-40 minutes

**Key Takeaways**:
- Exact file formats
- How to validate gentx files
- How to extend for future versions

---

### 6. [Integration Plan](./06_integration_plan.md)
**Purpose**: Detailed project plan with timeline and milestones

**Topics**:
- 12-week implementation schedule
- Phase-by-phase breakdown
- Dependencies and prerequisites
- Integration points with existing codebase
- Testing strategy
- Rollout plan (testnet → mainnet)
- Risk mitigation
- Success criteria

**Audience**: Project managers, technical leads, developers

**Estimated Reading Time**: 40-50 minutes

**Key Takeaways**:
- When will features be ready
- What are the dependencies
- How to validate each milestone

---

## Quick Start Guide

### For Developers New to the Project

**Step 1**: Read [01_overview.md](./01_overview.md)
- Understand the problem being solved
- Learn the core workflow

**Step 2**: Read [02_architecture.md](./02_architecture.md)
- Understand architectural decisions
- Learn component structure

**Step 3**: Read [03_domain_model.md](./03_domain_model.md)
- Understand business logic
- Learn domain concepts

**Step 4**: Follow [04_implementation_guide.md](./04_implementation_guide.md)
- Write tests first
- Implement features
- Refactor continuously

**Step 5**: Reference [05_data_structures.md](./05_data_structures.md)
- When working on persistence
- When debugging file formats

**Step 6**: Track progress with [06_integration_plan.md](./06_integration_plan.md)
- Check current milestone
- Plan next tasks

### For Project Managers

**Must Read**:
1. [01_overview.md](./01_overview.md) - Understand what is being built
2. [06_integration_plan.md](./06_integration_plan.md) - Timeline and milestones

**Optional**:
- [02_architecture.md](./02_architecture.md) - High-level technical approach

### For Technical Reviewers

**Must Read**:
1. [02_architecture.md](./02_architecture.md) - Design decisions
2. [03_domain_model.md](./03_domain_model.md) - Business logic
3. [04_implementation_guide.md](./04_implementation_guide.md) - Code examples

**Optional**:
- [05_data_structures.md](./05_data_structures.md) - Data formats
- [06_integration_plan.md](./06_integration_plan.md) - Integration approach

### For Operations/DevOps

**Must Read**:
1. [01_overview.md](./01_overview.md) - What GenUtils does
2. [05_data_structures.md](./05_data_structures.md) - File formats and structure
3. [06_integration_plan](./06_integration_plan.md) - Rollout strategy

## Design Philosophy

### Why SOLID?

This system follows SOLID principles to ensure:
- **Single Responsibility**: Each component has one reason to change
- **Open/Closed**: Easy to extend without modifying existing code
- **Liskov Substitution**: Components are truly interchangeable
- **Interface Segregation**: Clients only depend on what they need
- **Dependency Inversion**: Depend on abstractions, not concretions

### Why DDD?

Domain-Driven Design is used to:
- Keep business logic separate from infrastructure concerns
- Model the real-world problem domain accurately
- Use ubiquitous language shared by developers and domain experts
- Enforce domain invariants through aggregates
- Enable evolutionary architecture

### Why TDD?

Test-Driven Development ensures:
- Tests are written before implementation
- Only necessary code is written
- Comprehensive test coverage (target: 85%+)
- Confidence to refactor
- Documentation through tests

## Key Features

### Decentralized Genesis Creation
- Validators independently create gentxs
- No central authority needed for key generation
- Each validator controls their own keys

### Security by Design
- Private keys never leave validator machines
- Signature verification proves key ownership
- BLS keys for WBFT consensus security

### Type Safety
- Strong typing prevents invalid states
- Value objects ensure immutability
- Aggregates enforce invariants

### Extensibility
- Open/Closed principle allows adding features
- Version field supports format evolution
- Plugin architecture for custom validators

### Testability
- Clean architecture enables isolated testing
- Dependency injection simplifies mocking
- TDD ensures comprehensive coverage

## Architecture Highlights

### Layered Architecture

```
┌─────────────────────────────────────┐
│      Presentation Layer             │  CLI Commands
│         (cmd/gstable/)              │
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│      Application Layer              │  Use Cases
│   (internal/genutils/application/)  │
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│         Domain Layer                │  Business Logic
│     (internal/genutils/domain/)     │  (Pure, framework-agnostic)
└───────────────┬─────────────────────┘
                │
┌───────────────▼─────────────────────┐
│     Infrastructure Layer            │  Technical Concerns
│   (internal/genutils/repository/)   │  (Files, Crypto, etc.)
└─────────────────────────────────────┘
```

### Domain Model Overview

```
GenTx (Aggregate Root)
  ├── validator_address (Value Object)
  ├── operator_address (Value Object)
  ├── bls_public_key (Value Object)
  ├── signature (Value Object)
  └── metadata (Value Object)

GenTxCollection (Aggregate Root)
  ├── gentxs []GenTx
  ├── byValidator map
  ├── byOperator map
  └── byBLSKey map
```

## Workflows

### 1. Individual Validator Setup

```bash
# Generate validator keys
bootnode --genkey validator.key

# Create genesis transaction
gstable gentx create \
  --validator-key validator.key \
  --operator 0x... \
  --name "My Validator" \
  --chain-id stablenet-1

# Output: gentx-0x{validator_address}.json
```

### 2. Coordinator Genesis Assembly

```bash
# Collect all gentx files
mkdir gentxs/
# (receive gentx files from all validators)

# Validate collected gentxs
gstable gentx validate --gentx-dir gentxs/

# Generate final genesis
gstable gentx collect \
  --gentx-dir gentxs/ \
  --chain-id stablenet-1 \
  --output genesis.json
```

### 3. Network Launch

```bash
# All validators initialize with same genesis
gstable init genesis.json --datadir ./node

# Start network
gstable --datadir ./node ...
```

## Development Workflow

### Setting Up Development Environment

```bash
# Clone repository
git clone https://github.com/stable-net/go-stablenet.git
cd go-stablenet

# Install dependencies
go mod download

# Run tests
make test

# Run genutils tests specifically
go test -v -cover ./internal/genutils/...
```

### Making Changes

1. **Write Test First** (RED)
```bash
# Create test file
touch internal/genutils/domain/my_feature_test.go

# Write failing test
# Run: go test ./internal/genutils/domain/
# Should fail ❌
```

2. **Implement Feature** (GREEN)
```bash
# Create implementation file
touch internal/genutils/domain/my_feature.go

# Write minimal code to pass test
# Run: go test ./internal/genutils/domain/
# Should pass ✅
```

3. **Refactor** (REFACTOR)
```bash
# Improve code quality
# Run: go test ./internal/genutils/domain/
# Should still pass ✅
```

### Running Tests

```bash
# All tests
make test

# Genutils tests only
go test -v -cover ./internal/genutils/...

# Specific package
go test -v -cover ./internal/genutils/domain/

# With race detection
go test -race ./internal/genutils/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/genutils/...
go tool cover -html=coverage.out
```

## Contributing

### Code Style

Follow Go standard practices:
- `gofmt` for formatting
- `golint` for style
- `go vet` for correctness
- Comments for exported functions

### Pull Request Process

1. Fork repository
2. Create feature branch
3. Write tests (TDD)
4. Implement feature
5. Ensure tests pass
6. Update documentation
7. Submit pull request

### Review Checklist

- [ ] Tests written first
- [ ] All tests passing
- [ ] Coverage ≥85%
- [ ] SOLID principles followed
- [ ] DDD patterns correct
- [ ] Documentation updated
- [ ] No breaking changes (or documented)

## Support and Resources

### Documentation
- This documentation suite (you are here!)
- Code comments in implementation
- Test cases as examples

### Community
- GitHub Issues: Bug reports and feature requests
- Discord: Real-time discussion
- Telegram: Community chat

### Related Projects
- **AtomOne**: Inspiration for gentx concept
  - https://github.com/atomone-hub/atomone
- **Go-Ethereum**: Base blockchain implementation
  - https://github.com/ethereum/go-ethereum
- **WBFT Consensus**: Consensus engine
  - https://github.com/node-a-team/go-wemix

## Glossary

| Term | Definition |
|------|------------|
| **GenTx** | Genesis Transaction - declaration of validator intent to participate |
| **Aggregate** | DDD pattern - cluster of objects treated as a unit |
| **Value Object** | DDD pattern - immutable object defined by attributes |
| **Use Case** | Application layer - orchestrates domain objects for a specific task |
| **Repository** | Persistence abstraction - hides storage implementation details |
| **Domain Event** | Significant occurrence in the domain |
| **BLS Key** | BLS12-381 public key for WBFT consensus |
| **WBFT** | Weighted Byzantine Fault Tolerance consensus algorithm |

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.0 | 2024-01 | Initial design documentation |

## License

This documentation and the GenUtils implementation are part of the go-stablenet project.

See LICENSE file in the root directory for details.

---

## Next Steps

**New to the project?** → Start with [01_overview.md](./01_overview.md)

**Ready to implement?** → Follow [04_implementation_guide.md](./04_implementation_guide.md)

**Planning the project?** → Review [06_integration_plan.md](./06_integration_plan.md)

**Questions?** → Open an issue or ask in Discord/Telegram

---

*Last Updated: 2024-01*
*Maintained by: StableNet Core Team*

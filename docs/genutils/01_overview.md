# GenUtils - Genesis Transaction Utilities for Go-StableNet

## Overview

GenUtils is a comprehensive system for decentralized genesis block creation in Go-StableNet, inspired by Cosmos SDK's genutil module but adapted for Ethereum-based blockchain architecture with WBFT consensus.

## Purpose

Enable validators to independently create and sign genesis transactions (gentx) that:
- Declare their intention to participate in the network
- Prove ownership of validator keys
- Register BLS public keys for WBFT consensus
- Provide metadata and configuration

These gentxs are then collected and merged into a canonical genesis file that all validators use to bootstrap the network.

## Key Differences: Cosmos SDK vs Go-StableNet

| Aspect | Cosmos SDK (atomone) | Go-StableNet (this implementation) |
|--------|---------------------|-----------------------------------|
| **Consensus** | CometBFT (Tendermint) | WBFT (Anzeon) |
| **Account Format** | Bech32 (cosmos1...) | Ethereum hex (0x...) |
| **Key Type** | secp256k1, ed25519 | secp256k1 |
| **Validator Registration** | x/staking module | System contracts (GovValidator) |
| **Genesis TX Format** | Protobuf transaction | JSON metadata (no TX before genesis) |
| **Key Management** | Cosmos SDK keyring | Ethereum keystore |
| **BLS Keys** | Not required | Required for WBFT consensus |

## Architecture Principles

This implementation follows:

### SOLID Principles

- **Single Responsibility**: Each component has one reason to change
- **Open/Closed**: Extensible without modifying core logic
- **Liskov Substitution**: Interfaces are properly abstracted
- **Interface Segregation**: Clients depend only on needed interfaces
- **Dependency Inversion**: Depend on abstractions, not concretions

### Domain-Driven Design (DDD)

- **Bounded Contexts**: Clear separation between genesis, validation, collection
- **Aggregates**: GenTx as root aggregate with value objects
- **Entities**: Validator, Account as distinct entities
- **Value Objects**: Address, Signature, BLSPublicKey as immutable values
- **Domain Services**: Validation, Collection, Merging as stateless services

### Test-Driven Development (TDD)

- **Red-Green-Refactor**: Write failing test → implement → refactor
- **Unit Tests**: Test each component in isolation
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete workflows

## Core Workflows

### 1. Individual Validator Setup

```bash
# Step 1: Generate validator keys
bootnode --genkey validator.key
bootnode --nodekey validator.key --writeaddress > validator.info

# Step 2: Create account and fund it
gstable account new --password password.txt

# Step 3: Generate genesis transaction
gstable gentx create \
  --validator-key validator.key \
  --operator-address 0x... \
  --name "My Validator" \
  --output gentx-myvalidator.json
```

### 2. Coordinator Genesis Assembly

```bash
# Step 1: Collect all gentx files
mkdir gentxs/
cp validator1/gentx-*.json gentxs/
cp validator2/gentx-*.json gentxs/
# ... collect from all validators

# Step 2: Validate gentxs
gstable gentx validate --gentx-dir gentxs/

# Step 3: Generate genesis file
gstable gentx collect \
  --gentx-dir gentxs/ \
  --chain-id stablenet-1 \
  --output genesis.json
```

### 3. Network Launch

```bash
# All validators receive genesis.json
gstable init genesis.json

# Start nodes
gstable --datadir ./node1 ...
```

## Component Overview

### Core Components

1. **GenTx Domain** (`/internal/genutils/domain/`)
   - GenTx aggregate root
   - Validator entity
   - Value objects (Address, Signature, BLSKey)
   - Domain events

2. **GenTx Repository** (`/internal/genutils/repository/`)
   - File-based storage
   - GenTx serialization/deserialization
   - Query interface

3. **Validation Service** (`/internal/genutils/service/validation/`)
   - Signature verification
   - Format validation
   - Business rule validation

4. **Collection Service** (`/internal/genutils/service/collection/`)
   - GenTx aggregation
   - Duplicate detection
   - Ordering logic

5. **Genesis Builder** (`/internal/genutils/service/genesis/`)
   - Genesis.json construction
   - System contract initialization
   - WBFT configuration

6. **CLI Commands** (`/cmd/gstable/`)
   - `gentx create` - Create genesis transaction
   - `gentx validate` - Validate gentx files
   - `gentx collect` - Collect and merge gentxs
   - `gentx inspect` - Inspect gentx content

## File Structure

```
go-stablenet/
├── cmd/gstable/
│   ├── gentx_create.go          # Create command
│   ├── gentx_collect.go         # Collect command
│   ├── gentx_validate.go        # Validate command
│   └── gentx_inspect.go         # Inspect command
├── internal/genutils/
│   ├── domain/
│   │   ├── gentx.go             # GenTx aggregate
│   │   ├── validator.go         # Validator entity
│   │   ├── value_objects.go    # Address, Signature, etc.
│   │   └── events.go            # Domain events
│   ├── repository/
│   │   ├── interface.go         # Repository interface
│   │   ├── file.go              # File-based implementation
│   │   └── memory.go            # In-memory (for testing)
│   ├── service/
│   │   ├── validation/
│   │   │   ├── signature.go    # Signature validation
│   │   │   ├── format.go       # Format validation
│   │   │   └── business.go     # Business rules
│   │   ├── collection/
│   │   │   ├── collector.go    # Collection logic
│   │   │   └── merger.go       # Merge gentxs
│   │   └── genesis/
│   │       ├── builder.go      # Genesis construction
│   │       └── injector.go     # System contract injection
│   └── application/
│       ├── create_gentx.go      # Create use case
│       ├── validate_gentx.go    # Validate use case
│       └── collect_gentxs.go    # Collect use case
├── core/
│   ├── gentx.go                 # GenTx core types
│   └── genesis_gentx.go         # Genesis integration
└── systemcontracts/
    └── gentx_bootstrap.go       # GovValidator bootstrap

# Documentation
docs/genutils/
├── 01_overview.md               # This file
├── 02_architecture.md           # SOLID architecture design
├── 03_domain_model.md           # DDD domain model
├── 04_implementation_guide.md   # TDD/DDD implementation guide
├── 05_data_structures.md        # Data structures and formats
└── 06_integration_plan.md       # Integration with existing codebase
```

## Key Design Decisions

### 1. No Transaction Before Genesis

**Challenge**: Ethereum can't process transactions before genesis block is committed.

**Solution**: GenTx is metadata (JSON) stored in genesis.json, not a blockchain transaction.

```json
{
  "alloc": {...},
  "config": {...},
  "anzeonGenTxs": [
    {
      "validatorAddress": "0x...",
      "operatorAddress": "0x...",
      "blsPublicKey": "0x...",
      "signature": "0x...",
      "metadata": {...}
    }
  ]
}
```

### 2. Three-Key System

Each validator requires:

1. **Operator Address**: Controls validator via GovValidator contract (can be multisig)
2. **Validator Address**: Block signing key (node's coinbase)
3. **BLS Public Key**: WBFT consensus participation

```
Node Key (secp256k1)
  ├─→ Validator Address (0x...)
  └─→ BLS Private Key → BLS Public Key (0x...)
```

### 3. Signature Scheme

GenTx creator proves key ownership by signing:

```
message = keccak256(
  "GenTx Registration\n" +
  "Chain: stablenet-1\n" +
  "Validator: 0x...\n" +
  "Operator: 0x...\n" +
  "Timestamp: 1234567890"
)

signature = secp256k1_sign(message, validator_private_key)
```

Verification:
```go
recoveredAddress := crypto.Ecrecover(message, signature)
require(recoveredAddress == gentx.ValidatorAddress)
```

### 4. System Contract Bootstrap

GenTxs initialize `GovValidator` contract state:

```go
// During genesis block commit
for _, gentx := range genesis.AnzeonGenTxs {
    govValidator.AddValidator(
        gentx.ValidatorAddress,
        gentx.OperatorAddress,
        gentx.Metadata,
    )
}
```

## Security Considerations

1. **Private Key Protection**
   - Validator keys never leave local machine
   - GenTx contains only public keys and signatures

2. **Signature Verification**
   - All gentxs must have valid signatures
   - Timestamp prevents replay attacks

3. **Duplicate Prevention**
   - Each validator address can appear only once
   - Each operator address can appear only once

4. **BLS Key Validation**
   - BLS public keys must be valid G2 points
   - No duplicate BLS keys allowed

5. **Genesis File Integrity**
   - Final genesis.json hash is verified by all validators
   - Any tampering will cause consensus failure

## Migration from Manual Setup

Current manual setup:
```go
// params/config_wbft.go
anzeon := &AnzeonConfig{
    Init: &WBFTInit{
        Validators: []common.Address{
            common.HexToAddress("0x..."),
            common.HexToAddress("0x..."),
        },
        BLSPublicKeys: []string{
            "0x...",
            "0x...",
        },
    },
}
```

New gentx-based setup:
```bash
# Each validator creates gentx
gstable gentx create --validator-key node.key ...

# Coordinator collects and generates genesis
gstable gentx collect --gentx-dir ./gentxs --output genesis.json
```

The resulting genesis.json contains the same validator configuration, but generated in a decentralized manner.

## Next Steps

1. Review architecture design (02_architecture.md)
2. Study domain model (03_domain_model.md)
3. Follow implementation guide (04_implementation_guide.md)
4. Understand data structures (05_data_structures.md)
5. Execute integration plan (06_integration_plan.md)

## References

- Cosmos SDK genutil: https://github.com/cosmos/cosmos-sdk/tree/main/x/genutil
- Ethereum Genesis: https://github.com/ethereum/go-ethereum/blob/master/core/genesis.go
- WBFT Consensus: https://github.com/node-a-team/go-wemix/tree/main/consensus/wbft
- BLS Signatures: https://github.com/ethereum/go-ethereum/tree/master/crypto/bls12381

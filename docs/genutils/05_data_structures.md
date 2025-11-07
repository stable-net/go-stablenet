# Data Structures and File Formats

## Overview

This document defines the data structures and file formats used by GenUtils, including JSON schemas, serialization formats, and on-disk representations.

## GenTx JSON Format

### Schema Version 1.0

```json
{
  "$schema": "https://stablenet.io/schemas/gentx/v1.json",
  "version": "1.0",
  "chain_id": "stablenet-1",
  "validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "operator_address": "0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B",
  "bls_public_key": "0x1234567890abcdef...",
  "signature": "0xabcdef1234567890...",
  "metadata": {
    "name": "Validator One",
    "description": "Professional validator service",
    "website": "https://validator-one.example.com",
    "contact": "admin@validator-one.example.com"
  },
  "timestamp": 1704067200,
  "node_info": {
    "node_id": "e44b6d5e1e32d1c0f4f8c7e3a2b9c6d5",
    "listen_addr": "tcp://0.0.0.0:26656",
    "network": "stablenet"
  }
}
```

### Field Specifications

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | GenTx format version (currently "1.0") |
| `chain_id` | string | Yes | Chain identifier (e.g., "stablenet-1") |
| `validator_address` | string | Yes | Ethereum address (0x-prefixed hex) of validator key |
| `operator_address` | string | Yes | Ethereum address of operator (can be multisig) |
| `bls_public_key` | string | Yes | BLS12-381 G2 public key (0x-prefixed hex) |
| `signature` | string | Yes | ECDSA signature proving validator key ownership |
| `metadata` | object | Yes | Validator metadata |
| `timestamp` | uint64 | Yes | Unix timestamp when gentx was created |
| `node_info` | object | No | Optional node information |

### Validation Rules

#### validator_address
- Must be valid Ethereum address (0x + 40 hex chars)
- Must be checksummed (EIP-55)
- Cannot be zero address
- Must match address recovered from signature

#### operator_address
- Must be valid Ethereum address
- Can be same as validator_address
- Can be contract address (multisig)
- Cannot be zero address

#### bls_public_key
- Must be valid BLS12-381 G2 point
- Length: 192 bytes (384 hex chars + 0x prefix)
- Must be on curve
- Must be derived from validator key

#### signature
- Length: 65 bytes (130 hex chars + 0x prefix)
- Format: R (32 bytes) || S (32 bytes) || V (1 byte)
- Must verify against signing message
- Recovery ID (V) must be 27 or 28

#### metadata
- `name`: 1-70 characters, required
- `description`: 0-280 characters, optional
- `website`: Valid URL or empty
- `contact`: Email or other contact, optional

#### timestamp
- Unix timestamp (seconds since epoch)
- Must be within 7 days of current time (past)
- Must not be more than 1 hour in future

### Signing Message Format

The message signed by the validator key:

```
GenTx Registration
Chain: {chain_id}
Validator: {validator_address}
Operator: {operator_address}
Timestamp: {timestamp}
```

**Hashing**: Keccak256 of UTF-8 encoded message

**Example**:
```
GenTx Registration
Chain: stablenet-1
Validator: 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb
Operator: 0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B
Timestamp: 1704067200
```

### File Naming Convention

GenTx files are named using the validator address:

```
gentx-{validator_address}.json
```

**Example**:
```
gentx-0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb.json
```

### Complete Example

```json
{
  "version": "1.0",
  "chain_id": "stablenet-1",
  "validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "operator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "bls_public_key": "0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
  "signature": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd1b",
  "metadata": {
    "name": "StableNet Validator One",
    "description": "Enterprise-grade validator service with 24/7 monitoring and redundant infrastructure",
    "website": "https://validator.stablenet.io",
    "contact": "ops@validator.stablenet.io"
  },
  "timestamp": 1704067200,
  "node_info": {
    "node_id": "e44b6d5e1e32d1c0f4f8c7e3a2b9c6d5",
    "listen_addr": "tcp://0.0.0.0:26656",
    "network": "stablenet"
  }
}
```

## Genesis JSON Format (Extended)

### Standard Ethereum Genesis

Go-StableNet uses standard Ethereum genesis format with extensions:

```json
{
  "config": {
    "chainId": 1001,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "anzeon": {
      "epoch": 3600,
      "policy": 0,
      "sub": 22,
      "wbft": {
        "timeout": 10000,
        "blockPeriod": 1,
        "requestTimeout": 3000,
        "blockReward": "0"
      },
      "init": {
        "validators": [
          "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
          "0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B"
        ],
        "blsPublicKeys": [
          "0xa1b2c3d4...",
          "0xb2c3d4e5..."
        ]
      },
      "systemContracts": {
        "govValidator": {
          "addr": "0x0000000000000000000000000000000000000400",
          "params": {
            "operators": [
              "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
              "0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B"
            ],
            "metadata": [
              {
                "name": "Validator One",
                "description": "...",
                "website": "...",
                "contact": "..."
              },
              {
                "name": "Validator Two",
                "description": "...",
                "website": "...",
                "contact": "..."
              }
            ]
          }
        }
      }
    }
  },
  "nonce": "0x0",
  "timestamp": "0x659a8c00",
  "extraData": "0x",
  "gasLimit": "0xe8d4a51000",
  "difficulty": "0x1",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "alloc": {
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb": {
      "balance": "0x52b7d2dcc80cd2e4000000"
    },
    "0x8f0a5E3d2F1b4C9a8B7E6D5C4A3B2C1D0E9F8A7B": {
      "balance": "0x52b7d2dcc80cd2e4000000"
    }
  },
  "anzeonGenTxs": [
    {
      "validator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "operator_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "bls_public_key": "0xa1b2c3d4...",
      "signature": "0x1234567890...",
      "metadata": {
        "name": "Validator One",
        "description": "...",
        "website": "...",
        "contact": "..."
      },
      "timestamp": 1704067200
    }
  ]
}
```

### Extended Fields for GenTx

The `anzeonGenTxs` field contains the full gentx data:

```json
{
  "anzeonGenTxs": [
    {
      "validator_address": "0x...",
      "operator_address": "0x...",
      "bls_public_key": "0x...",
      "signature": "0x...",
      "metadata": {...},
      "timestamp": 1234567890
    }
  ]
}
```

### Genesis Building Process

1. **Load Base Genesis**: Start with template genesis.json
2. **Extract Validators**: Parse all gentx files
3. **Build Validator Lists**:
   - `config.anzeon.init.validators` ← validator addresses
   - `config.anzeon.init.blsPublicKeys` ← BLS public keys
4. **Build System Contract Params**:
   - `config.anzeon.systemContracts.govValidator.params.operators` ← operator addresses
   - `config.anzeon.systemContracts.govValidator.params.metadata` ← validator metadata
5. **Add GenTxs**: Store full gentx data in `anzeonGenTxs`
6. **Set Genesis Time**: Use canonical time for all validators
7. **Set Chain ID**: Ensure consistency across all components

## Go Data Structures

### Core Domain Types

```go
// Address value object
type Address struct {
    value common.Address
}

// Signature value object
type Signature struct {
    data [65]byte // R (32) + S (32) + V (1)
}

// BLSPublicKey value object
type BLSPublicKey struct {
    point *bls12381.PointG2
}

// ValidatorMetadata value object
type ValidatorMetadata struct {
    name        string
    description string
    website     string
    contact     string
}

// GenTx aggregate root
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

// GenTxCollection aggregate
type GenTxCollection struct {
    gentxs      []*GenTx
    chainID     string
    byValidator map[string]*GenTx
    byOperator  map[string]*GenTx
    byBLSKey    map[string]*GenTx
    events      []DomainEvent
}
```

### Serialization DTOs

```go
// GenTxDTO for JSON serialization
type GenTxDTO struct {
    Version          string       `json:"version"`
    ChainID          string       `json:"chain_id"`
    ValidatorAddress string       `json:"validator_address"`
    OperatorAddress  string       `json:"operator_address"`
    BLSPublicKey     string       `json:"bls_public_key"`
    Signature        string       `json:"signature"`
    Metadata         MetadataDTO  `json:"metadata"`
    Timestamp        uint64       `json:"timestamp"`
    NodeInfo         *NodeInfoDTO `json:"node_info,omitempty"`
}

// MetadataDTO for validator metadata
type MetadataDTO struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Website     string `json:"website,omitempty"`
    Contact     string `json:"contact,omitempty"`
}

// NodeInfoDTO for optional node information
type NodeInfoDTO struct {
    NodeID     string `json:"node_id"`
    ListenAddr string `json:"listen_addr"`
    Network    string `json:"network"`
}
```

### Genesis Extensions

```go
// Extended genesis structure
type GenesisWithGenTxs struct {
    *core.Genesis
    AnzeonGenTxs []GenTxDTO `json:"anzeonGenTxs"`
}

// System contract initialization parameters
type GovValidatorParams struct {
    Operators []common.Address     `json:"operators"`
    Metadata  []ValidatorMetadata  `json:"metadata"`
}

// Anzeon configuration with gentx-derived data
type AnzeonConfig struct {
    Epoch           uint64           `json:"epoch"`
    Policy          uint64           `json:"policy"`
    Sub             uint64           `json:"sub"`
    WBFT            *WBFTConfig      `json:"wbft"`
    Init            *WBFTInit        `json:"init"`
    SystemContracts *SystemContracts `json:"systemContracts"`
}

type WBFTInit struct {
    Validators    []common.Address `json:"validators"`
    BLSPublicKeys []string         `json:"blsPublicKeys"`
}
```

## Serialization/Deserialization

### GenTx to JSON

```go
func SerializeGenTx(gentx *domain.GenTx) ([]byte, error) {
    dto := GenTxDTO{
        Version:          "1.0",
        ChainID:          gentx.ChainID(),
        ValidatorAddress: gentx.ValidatorAddress().String(),
        OperatorAddress:  gentx.OperatorAddress().String(),
        BLSPublicKey:     gentx.BLSPublicKey().String(),
        Signature:        formatSignature(gentx.Signature()),
        Metadata: MetadataDTO{
            Name:        gentx.Metadata().Name(),
            Description: gentx.Metadata().Description(),
            Website:     gentx.Metadata().Website(),
            Contact:     gentx.Metadata().Contact(),
        },
        Timestamp: gentx.Timestamp(),
    }

    return json.MarshalIndent(dto, "", "  ")
}

func formatSignature(sig domain.Signature) string {
    return fmt.Sprintf("0x%x", sig.Bytes())
}
```

### JSON to GenTx

```go
func DeserializeGenTx(data []byte) (*domain.GenTx, error) {
    var dto GenTxDTO
    if err := json.Unmarshal(data, &dto); err != nil {
        return nil, fmt.Errorf("failed to unmarshal: %w", err)
    }

    // Validate version
    if dto.Version != "1.0" {
        return nil, fmt.Errorf("unsupported version: %s", dto.Version)
    }

    // Parse addresses
    validatorAddr, err := domain.NewAddress(dto.ValidatorAddress)
    if err != nil {
        return nil, fmt.Errorf("invalid validator address: %w", err)
    }

    operatorAddr, err := domain.NewAddress(dto.OperatorAddress)
    if err != nil {
        return nil, fmt.Errorf("invalid operator address: %w", err)
    }

    // Parse BLS key
    blsKeyBytes, err := hex.DecodeString(strings.TrimPrefix(dto.BLSPublicKey, "0x"))
    if err != nil {
        return nil, fmt.Errorf("invalid BLS key: %w", err)
    }
    blsKey, err := domain.NewBLSPublicKey(blsKeyBytes)
    if err != nil {
        return nil, err
    }

    // Parse signature
    sigBytes, err := hex.DecodeString(strings.TrimPrefix(dto.Signature, "0x"))
    if err != nil {
        return nil, fmt.Errorf("invalid signature: %w", err)
    }
    signature, err := domain.NewSignature(sigBytes)
    if err != nil {
        return nil, err
    }

    // Parse metadata
    metadata, err := domain.NewValidatorMetadata(
        dto.Metadata.Name,
        dto.Metadata.Description,
        dto.Metadata.Website,
        dto.Metadata.Contact,
    )
    if err != nil {
        return nil, err
    }

    // Reconstruct GenTx
    // Note: Cannot use normal constructor as we don't have private key
    gentx := domain.ReconstructGenTx(
        validatorAddr,
        operatorAddr,
        blsKey,
        signature,
        metadata,
        dto.Timestamp,
        dto.ChainID,
    )

    return gentx, nil
}
```

## File System Layout

### Individual Validator Setup

```
validator-node/
├── keystore/
│   └── validator.key           # Validator private key
├── config/
│   └── node_key.json           # Node P2P identity
└── gentxs/
    └── gentx-0x742d35...json   # Generated gentx
```

### Coordinator Genesis Assembly

```
genesis-coordinator/
├── gentxs/                     # Collected gentx files
│   ├── gentx-0x742d35...json
│   ├── gentx-0x8f0a5e...json
│   └── gentx-0x1a2b3c...json
├── base-genesis.json           # Template genesis
└── genesis.json                # Final genesis with gentxs
```

## Validation Schemas

### JSON Schema for GenTx

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "GenTx",
  "type": "object",
  "required": [
    "version",
    "chain_id",
    "validator_address",
    "operator_address",
    "bls_public_key",
    "signature",
    "metadata",
    "timestamp"
  ],
  "properties": {
    "version": {
      "type": "string",
      "enum": ["1.0"]
    },
    "chain_id": {
      "type": "string",
      "minLength": 1
    },
    "validator_address": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{40}$"
    },
    "operator_address": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{40}$"
    },
    "bls_public_key": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{384}$"
    },
    "signature": {
      "type": "string",
      "pattern": "^0x[0-9a-fA-F]{130}$"
    },
    "metadata": {
      "type": "object",
      "required": ["name"],
      "properties": {
        "name": {
          "type": "string",
          "minLength": 1,
          "maxLength": 70
        },
        "description": {
          "type": "string",
          "maxLength": 280
        },
        "website": {
          "type": "string",
          "format": "uri"
        },
        "contact": {
          "type": "string"
        }
      }
    },
    "timestamp": {
      "type": "integer",
      "minimum": 0
    },
    "node_info": {
      "type": "object",
      "properties": {
        "node_id": {
          "type": "string",
          "pattern": "^[0-9a-f]{32}$"
        },
        "listen_addr": {
          "type": "string"
        },
        "network": {
          "type": "string"
        }
      }
    }
  }
}
```

## Migration and Versioning

### Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024-01 | Initial format specification |

### Future Versions

Version 2.0 may include:
- Additional signature types (multisig, threshold)
- Enhanced metadata (social links, identity proofs)
- Stake amount declaration
- Commission rate preferences
- Delegation policies

### Backward Compatibility

Newer versions must support reading older formats:

```go
func DeserializeGenTx(data []byte) (*domain.GenTx, error) {
    var dto GenTxDTO
    if err := json.Unmarshal(data, &dto); err != nil {
        return nil, err
    }

    switch dto.Version {
    case "1.0":
        return deserializeV1(dto)
    case "2.0":
        return deserializeV2(dto)
    default:
        return nil, fmt.Errorf("unsupported version: %s", dto.Version)
    }
}
```

## Summary

This data structure specification provides:

1. **Clear Format**: Well-defined JSON schema for gentx
2. **Validation Rules**: Comprehensive validation requirements
3. **Type Safety**: Strong typing in Go domain model
4. **Extensibility**: Version field allows future evolution
5. **Compatibility**: Standard Ethereum genesis format compatibility

Next: See [06_integration_plan.md](./06_integration_plan.md) for integration roadmap.

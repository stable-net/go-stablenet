# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **go-stablenet**, the official Go implementation of the StableNet protocol - a BFT-based blockchain designed specifically for stablecoins. It's a fork of WBFT (go-wbft) with extensive modifications to enable stablecoins as the base coin (gas token).

Key innovations:
- Stablecoins as native gas tokens with mint/burn governance
- NativeCoinAdapter: ERC20-compatible wrapper for the base coin
- Anzeon WBFT consensus engine (modified from WBFT/QBFT)
- Three-tier governance: GovValidator, GovMinter, GovMasterMinter
- Gas fee policy optimized for stable tokens (pegged to KRW)

## Build Commands

```bash
# Build main client (gstable)
make gstable

# Build genesis generator
make genesis_generator

# Build all executables
make all

# Run tests (only success test suite)
make test

# Run short tests
make test-short

# Run linter
make lint

# Clean build artifacts
make clean

# Install developer tools
make devtools
```

Main executable: `./build/bin/gstable`

## Running Tests

Tests are split into two categories (see Makefile):
- **SUCCESS_TESTS**: Currently passing test packages
- **FAILURE_TESTS**: Tests that need to be migrated to SUCCESS_TESTS

To run a single test package:
```bash
go test github.com/ethereum/go-ethereum/consensus/wbft/...
```

To run a specific test:
```bash
go test -run TestValidatorSelection github.com/ethereum/go-ethereum/consensus/wbft
```

## Architecture Overview

### Consensus Layer (Anzeon WBFT)

Located in `consensus/wbft/`:
- **core/**: Core BFT consensus logic
- **backend/**: Blockchain backend integration
- **engine/**: Consensus engine implementation
- **validator/**: Validator set management
- **messages/**: BFT message types

Key differences from WBFT:
- No staking mechanism (PoA only)
- No block rewards
- Validator management through GovValidator contract
- BLS keys stored in GovValidator (not GovStaking)

### System Contracts

Located in `systemcontracts/`:
- **GovValidator** (0x...1001): Manages validators (operator keys, validator keys, BLS keys)
- **NativeCoinAdapter** (0x...1000): ERC20 wrapper for the base coin
- **GovMinter** (0x...1003): Manages minters and mint/burn operations
- **GovMasterMinter** (0x...1002): Manages minter registration/removal

These are genesis-deployed contracts without owners. Upgrades require hard forks.

Contract compilation: `systemcontracts/compile` script generates Go bindings in `systemcontracts/artifacts/`

### Validator Composition

Each validator consists of three components:
1. **Operator key**: Governance and voting (can be EOA or multisig)
2. **Validator key**: Block signing address (coinbase)
3. **BLS key**: WBFT consensus message signing

### Gas Fee Policy

Constants in `params/protocol_params.go`:
- `GasTargetPercentage`: 90% (base fee rises when block >90% full)
- `BaseFeeChangeRate`: 15%
- `MinBaseFee`: 5000000000000 (pegged to KRW)
- `MaxBaseFee`: 5000000000000000
- `InitialGasTip`: 5000000000000 (set by validator governance, not user)

Unlike Ethereum:
- Base fee paid to validators (not burned, since stablecoins can't be burned without fiat redemption)
- Priority fee set by governance consensus, not per-transaction
- No tip-based transaction prioritization

### Core Packages

- `core/`: Blockchain core (state processor, tx pool, block validator)
- `core/types/`: Core data structures (Block, Transaction, Receipt)
- `eth/`: Ethereum protocol implementation
- `miner/`: Block production
- `params/`: Chain configuration and protocol parameters
- `accounts/`: Account management and signing
- `crypto/`: Cryptographic utilities
- `ethdb/`: Database abstractions
- `trie/`: Merkle Patricia Trie implementation
- `pkg/crypto/ethereum/`: Custom Ethereum crypto provider

### Command Line Tools

Located in `cmd/`:
- **gstable**: Main client (equivalent to geth)
- **genesis_generator**: Interactive genesis file creator
- **bootnode**: Network bootstrap node
- **clef**: External signer
- **abigen**: Contract binding generator
- **ethkey**: Key utilities
- **evm**: Standalone EVM executor

## Genesis Configuration

Anzeon config structure (in genesis.json):
```json
{
  "config": {
    "anzeon": {
      "wbft": {
        "requestTimeoutSeconds": 2,
        "blockPeriodSeconds": 1,
        "epochLength": 10,
        "proposerPolicy": 0
      },
      "init": {
        "validators": ["0x..."],
        "blsPublicKeys": ["0x..."]
      },
      "systemContracts": {
        "govValidator": {...},
        "nativeCoinAdapter": {...},
        "govMinter": {...},
        "govMasterMinter": {...}
      }
    }
  }
}
```

**Important**: The `init.validators` apply to epoch 0 only. From epoch 1 onwards, validators come from `govValidator` configuration.

## Key Architectural Patterns

### NativeCoinAdapter Design

The NativeCoinAdapter is a precompile-backed system contract that:
- Wraps the base coin as an ERC20 token
- Uses precompiles (NativeCoinManagerAddress 0xb00002) to manage allowances
- Emits Transfer events for ALL base coin movements (including gas payments)
- References account native balance directly (no separate storage)
- Supports full FiatTokenV2_2 interface

This enables legacy ERC20-based services to work with the base coin without modification.

### Mint/Burn Protocol

Minting requires:
1. Fiat deposit to collateral account
2. Minter creates "minting proof" (deposit_id, amount, beneficiary, timestamp)
3. Proposal submitted to GovMinter
4. Quorum approval from other minters
5. Mint executes via NativeCoinAdapter.mint()

Burning follows similar governance process with "burn proof" (withdrawal_id, amount, from, timestamp)

### Validator Management

Validator changes:
1. Proposed via GovValidator by existing validator operator
2. Voted on by validator operators (quorum required)
3. Changes take effect at next epoch boundary
4. Each validator must provide: operator address, validator address, BLS public key

## Development Notes

- Go version: 1.20+
- Module name: `github.com/ethereum/go-ethereum` (inherited from go-ethereum)
- Branch strategy: `dev` is main development branch (see git status)
- Code must follow Go formatting guidelines (`gofmt`)
- Commit messages should be prefixed with package names (e.g., "eth, rpc: fix trace configs")

## Common Tasks

Generate nodekey and addresses:
```bash
./build/bin/bootnode -genkey mynodekey
./build/bin/bootnode -nodekey mynodekey -writeaddress
# Outputs: public key, address, derived BLS public key
```

Initialize private chain:
```bash
./build/bin/genesis_generator  # Interactive genesis creation
./build/bin/gstable init genesis.json
./build/bin/gstable --datadir ./data --mine
```

Run with RPC enabled:
```bash
./build/bin/gstable --http --http.addr "0.0.0.0" --http.port 8545 --mine
```

## Testing System Contracts

System contract tests are in `systemcontracts/*_test.go`. They use:
- Go testing framework
- Simulated backends for contract interaction
- Test utilities in `systemcontracts/test/`

Recompile contracts after Solidity changes:
```bash
cd systemcontracts
./compile
```

## Important Constraints

1. **No Block Rewards**: Validators earn only transaction fees
2. **No Staking**: Pure PoA - validator admission by governance vote only
3. **No Token Burning**: Base fee goes to validators, never burned (stablecoins require fiat redemption)
4. **Immutable System Contracts**: No upgradability - changes require hard fork
5. **Governance Quorum**: All governance actions require quorum approval
6. **Epoch Boundaries**: Validator set changes apply only at epoch boundaries

## WBFT Engine Specifics

The Anzeon WBFT engine removed from original WBFT:
- WPoA (legacy WEMIX 3.0 engine)
- GovStaking, GovConfig, GovNCP, GovRewardeeImp contracts
- Staking/slashing/diligence mechanisms
- Block reward distribution (Brioche fork logic)
- Croissant config (replaced with Anzeon config)
- Properties: stabilizingStakersThreshold, targetValidators, useNCP

Retained:
- BFT consensus algorithm
- Governance framework
- Validator management
- BLS signature aggregation

# Design Document

## Overview

This design outlines the consolidation of the PoH Blockchain project's numerous scripts and binaries into a clean, focused set of executables. The current project has multiple overlapping scripts (`audit.sh`, `devnet.sh`, `demo-dpos.sh`, `test-launcher.sh`, etc.) and binaries (`poh-node`, `neon-wallet`, `validator-tui`) with unclear responsibilities. This design establishes a clear separation of concerns and eliminates redundancy.

## Architecture

The consolidated architecture will have 5 main binaries, each with a single, well-defined purpose:

1. **audit** - Security verification and network simulation
2. **neon-wallet** - Account management and token transfers
3. **validator** - Staking, unstaking, delegation, and node operations
4. **devnet** - Local validator network creation and management
5. **build** - Project compilation and dependency management

### Current State Analysis

**Existing Scripts to be Consolidated:**
- `audit.sh` - Comprehensive blockchain testing (BFT, DPoS, consensus)
- `devnet.sh` - Local validator network management
- `demo-dpos.sh` - DPoS demonstration with stake-weighted scheduling
- `test-launcher.sh` - BFT test suite runner
- `analyze-results.sh` - Test result analysis
- `test_runner.sh` - Test execution
- `run_fibonacci.sh` - QuanticScript example runner
- `stop-demo.sh` - Demo cleanup
- `build.sh` - Binary compilation

**Existing Binaries:**
- `poh-node` - Main blockchain node (multi-purpose)
- `neon-wallet` - Wallet management (already properly scoped)
- `validator-tui` - Validator terminal UI

### Proposed Architecture

```
.
├── bin/
│   ├── audit           # Consolidated from audit.sh, test-launcher.sh, analyze-results.sh
│   ├── neon-wallet     # Keep existing functionality
│   ├── validator       # Consolidated from poh-node (validator features) + validator-tui
│   ├── devnet          # Consolidated from devnet.sh + demo-dpos.sh
│   └── build           # Consolidated from build.sh
├── cmd/
│   ├── audit/          # New package for audit binary
│   ├── neon-wallet/    # Existing wallet package
│   ├── validator/      # New package for validator binary
│   ├── devnet/         # New package for devnet binary
│   └── build/          # New package for build binary
└── scripts/            # Remaining essential scripts (if any)
```

## Components and Interfaces

### 1. Audit Binary (`bin/audit`)

**Responsibilities:**
- Run network simulations and verify security properties
- Test Byzantine Fault Tolerance (BFT) scenarios
- Validate DPoS mechanics and leader elections
- Generate comprehensive test reports

**Interface:**
```bash
# Basic usage
audit --duration 30 --validators 3

# BFT testing
audit bft --honest 4 --malicious 1 --duration 60

# DPoS testing
audit dpos --validators 5 --duration 120

# Generate report
audit report --db-dir ./test-results
```

**Source Consolidation:**
- `audit.sh` → core functionality
- `test-launcher.sh` → test orchestration
- `analyze-results.sh` → result analysis
- `test_runner.sh` → test execution

### 2. Neon-Wallet Binary (`bin/neon-wallet`)

**Responsibilities:**
- Create and manage accounts
- Store accounts in dedicated storage location (`~/.config/poh-blockchain/wallets/`)
- Send Neon tokens by address
- Load accounts for validator operations

**Interface:**
```bash
# Account management
neon-wallet create --name my-wallet
neon-wallet list
neon-wallet show --name my-wallet

# Token transfers
neon-wallet transfer --from my-wallet --to <address> --amount 1000

# Export/import
neon-wallet export --name my-wallet --output wallet.json
neon-wallet import --input wallet.json --name my-wallet
```

**Source:**
- `cmd/wallet/main.go` → keep existing implementation
- Add dedicated storage path enforcement

### 3. Validator Binary (`bin/validator`)

**Responsibilities:**
- Staking and unstaking operations
- Delegation to other validators
- Spinning up nodes that connect to devnet
- Loading Neon accounts from neon-wallet storage location
- Terminal UI for validator operations

**Interface:**
```bash
# Staking operations
validator stake --wallet my-wallet --amount 1000000
validator unstake --wallet my-wallet --amount 500000
validator delegate --wallet my-wallet --to <validator-address> --amount 1000000

# Node operations
validator node start --wallet my-wallet --port 8000
validator node status
validator node stop

# TUI mode
validator tui --wallet my-wallet
```

**Source Consolidation:**
- `poh-node` validator features → core node operations
- `validator-tui` → terminal UI
- Remove non-validator features from `poh-node`

### 4. Devnet Binary (`bin/devnet`)

**Responsibilities:**
- Create local set of validators with leadership switching
- Initialize developer/tester account with coins
- Store seed phrase in `devnet_tester.txt`
- Manage local validator networks

**Interface:**
```bash
# Basic usage
devnet start --validators 3

# With developer account
devnet start --validators 3 --developer-account

# Management
devnet status
devnet logs --validator 1
devnet stop
devnet clean
```

**Source Consolidation:**
- `devnet.sh` → core devnet management
- `demo-dpos.sh` → DPoS demonstration features
- `stop-demo.sh` → cleanup functionality

### 5. Build Binary (`bin/build`)

**Responsibilities:**
- Compile all project binaries
- Download dependencies
- Create minimal, optimized builds

**Interface:**
```bash
# Build all binaries
build all

# Build specific binary
build audit
build neon-wallet
build validator
build devnet

# Clean build
build clean
```

**Source:**
- `build.sh` → core compilation logic

## Data Models

### Wallet Storage Location
```
~/.config/poh-blockchain/wallets/
├── my-wallet.wallet
├── validator1.wallet
└── devnet-tester.wallet
```

### Devnet Tester Account
```json
{
  "seed_phrase": "word1 word2 word3 ... word24",
  "public_key": "0x...",
  "private_key": "0x...",
  "initial_balance": 10000000,
  "created_at": "2026-04-30T21:57:02Z"
}
```

### Configuration Files
- `devnet_tester.txt` - Seed phrase for developer/tester account
- `devnet-config.json` - Devnet configuration
- `audit-config.json` - Audit test configuration

## Error Handling

### Binary-Specific Errors
Each binary will have its own error domain:
- `audit`: Network simulation errors, BFT verification failures
- `neon-wallet`: Account creation errors, insufficient balance, invalid addresses
- `validator`: Staking errors, node connection failures, delegation errors
- `devnet`: Validator startup failures, network initialization errors
- `build`: Compilation errors, dependency resolution failures

### Common Error Patterns
- Clear error messages with suggested fixes
- Proper exit codes (0 = success, 1 = user error, 2 = system error)
- Logging to appropriate locations (`~/.local/share/poh-blockchain/logs/`)

## Testing Strategy

### Unit Testing
- Each binary package has its own `*_test.go` files
- Mock external dependencies (network, file system)
- Test command-line interfaces

### Integration Testing
- `audit` → Test BFT scenarios end-to-end
- `neon-wallet` → Test account creation and transfers
- `validator` → Test staking and node operations
- `devnet` → Test local network creation
- Cross-binary integration (e.g., wallet → validator → devnet)

### Acceptance Testing
- Scripts to verify binary consolidation
- Verify no functionality loss during consolidation
- Test backward compatibility where applicable

## Implementation Plan

### Phase 1: Analysis and Planning
1. Document all existing scripts and their functionality
2. Map functionality to new binary structure
3. Identify dependencies and integration points

### Phase 2: Core Binary Implementation
1. Create `audit` binary package
2. Enhance `neon-wallet` with dedicated storage
3. Create `validator` binary package
4. Create `devnet` binary package
5. Create `build` binary package

### Phase 3: Feature Migration
1. Migrate functionality from scripts to binaries
2. Implement developer/tester account creation
3. Add dedicated wallet storage path
4. Implement devnet tester seed phrase storage

### Phase 4: Testing and Validation
1. Unit tests for each binary
2. Integration tests for cross-binary workflows
3. Acceptance tests for full functionality
4. Performance testing

### Phase 5: Cleanup
1. Remove deprecated scripts
2. Update documentation
3. Update CI/CD pipelines
4. Release new binary structure
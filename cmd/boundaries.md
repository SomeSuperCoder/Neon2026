# Binary Boundaries

This document defines the boundaries and responsibilities for each binary in the consolidated architecture.

## 1. Audit Binary (`bin/audit`)

**Responsibilities:**
- Network simulations and security verification
- Byzantine Fault Tolerance (BFT) testing
- DPoS mechanics and leader election validation
- Test report generation

**Boundaries:**
- SHALL NOT create or manage wallets
- SHALL NOT perform staking/unstaking operations
- SHALL NOT manage validator nodes (except for testing)
- SHALL NOT compile or build binaries

**Interface:**
- Command-line interface for test configuration
- JSON report output format
- Network simulation control

## 2. Neon-Wallet Binary (`bin/neon-wallet`)

**Responsibilities:**
- Account creation and management
- Token transfers by address
- Secure wallet storage in dedicated location (`~/.config/poh-blockchain/wallets/`)
- Wallet encryption and decryption

**Boundaries:**
- SHALL NOT run network simulations
- SHALL NOT perform staking/unstaking operations
- SHALL NOT manage validator nodes
- SHALL NOT compile or build binaries

**Interface:**
- Command-line interface for wallet operations
- File-based wallet storage
- RPC client for blockchain interaction

## 3. Validator Binary (`bin/validator`)

**Responsibilities:**
- Staking and unstaking operations
- Delegation to other validators
- Validator node management
- Loading wallets from neon-wallet storage location
- Terminal UI for validator operations

**Boundaries:**
- SHALL NOT create new wallets (uses neon-wallet storage)
- SHALL NOT run network simulations (except for testing)
- SHALL NOT compile or build binaries
- SHALL NOT manage devnet networks (except for connecting to them)

**Interface:**
- Command-line interface for staking operations
- Node management commands
- TUI for interactive operations
- RPC client for blockchain interaction

## 4. Devnet Binary (`bin/devnet`)

**Responsibilities:**
- Local validator network creation and management
- Developer/tester account creation with initial coins
- Seed phrase storage in `devnet_tester.txt`
- Network lifecycle management (start, stop, status, logs)

**Boundaries:**
- SHALL NOT create general-purpose wallets (except developer account)
- SHALL NOT perform staking/unstaking operations
- SHALL NOT run security audits
- SHALL NOT compile or build binaries

**Interface:**
- Command-line interface for network management
- Log viewing and monitoring
- Network status reporting

## 5. Build Binary (`bin/build`)

**Responsibilities:**
- Compilation of all project binaries
- Dependency downloading and management
- Clean build functionality

**Boundaries:**
- SHALL NOT run any blockchain operations
- SHALL NOT create or manage wallets
- SHALL NOT perform staking/unstaking operations
- SHALL NOT run network simulations

**Interface:**
- Command-line interface for compilation
- Binary-specific build options
- Cleanup operations

## Cross-Binary Integration Points

1. **Wallet Storage:** `neon-wallet` stores wallets → `validator` loads wallets
2. **Devnet Connection:** `validator` nodes connect to networks created by `devnet`
3. **Test Accounts:** `devnet` creates test accounts → used by `audit` for testing
4. **Binary Production:** `build` compiles binaries → all other binaries use them

## Enforcement Mechanisms

1. **Compile-time checks:** Each binary has minimal dependencies
2. **Runtime checks:** Binaries validate operations against their boundaries
3. **Documentation:** Clear separation documented in code and user guides
4. **Testing:** Integration tests verify boundary compliance
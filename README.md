# PoH Blockchain

A Proof of History (PoH) blockchain implementation inspired by Solana's architecture, built in Go.

## Quick Start

```bash
# Run comprehensive test suite
./audit.sh

# Start local development network
./devnet.sh start

# Check network status
./devnet.sh status

# Stop the network
./devnet.sh stop
```

See [docs/guides/quickstart.md](docs/guides/quickstart.md) for a 30-second introduction.

## Overview

This project implements a verifiable delay function using sequential SHA-256 hashing to create a cryptographic clock, enabling high-throughput transaction ordering without traditional consensus overhead.

## Features

- Proof of History clock generator with verifiable sequential hashing
- Leader-based consensus protocol with 400ms slots
- P2P network communication for block distribution
- SQLite-based persistent ledger storage
- Full chain verification and integrity checking
- Transaction integration with cryptographic timestamping
- **Delegated Proof of Stake (DPoS) genesis initialization**
- **Byzantine Fault Tolerance testing with malicious nodes**
- **Comprehensive audit script for automated testing**
- **Local development network (devnet) for testing and development**
- **CI/CD-friendly testing with JSON reports**

## Requirements

- Go 1.23.0 or higher
- SQLite3

## Installation

```bash
# Clone the repository
git clone https://github.com/poh-blockchain
cd poh-blockchain

# Install dependencies
go mod download
```

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── main.go            # Main CLI with node operations and subcommands
│   ├── validator-tui/     # Validator TUI dashboard
│   │   └── main.go
│   └── wallet/            # Wallet implementation
│       └── core/          # Wallet core functionality
│           ├── config.go       # Configuration management
│           ├── mnemonic.go     # BIP39 seed phrase generation
│           ├── mnemonic_test.go
│           ├── derivation.go   # BIP44 key derivation
│           ├── derivation_test.go
│           ├── encryption.go   # AES-256-GCM encryption
│           ├── encryption_test.go
│           ├── types.go        # Core wallet types
│           ├── wallet.go       # Wallet management
│           └── wallet_test.go
├── internal/
│   ├── poh/               # PoH clock generator
│   ├── blockchain/        # Core blockchain data structures and serialization
│   ├── network/           # P2P networking layer
│   ├── consensus/         # Consensus protocol
│   ├── storage/           # SQLite ledger persistence
│   ├── verification/      # Chain integrity verification
│   ├── filestore/         # File-based state model (BadgerDB)
│   ├── transaction/       # Transaction and instruction types
│   ├── access/            # Access control and permission validation
│   ├── processor/         # Atomic transaction processing with rollback
│   ├── runtime/           # Program execution runtime and builtin registry
│   ├── system/            # Go-side system program
│   ├── wallet/            # Validator wallet (encrypted Ed25519 keypair storage)
│   ├── genesis/           # Genesis bootstrap (loads builtin programs, initializes DPoS)
│   ├── parallel/          # Parallel execution conflict analysis
│   ├── rpc/               # JSON-RPC 2.0 server and query engine
│   │   ├── server.go      # HTTP server with CORS support
│   │   ├── handler.go     # Request handler and method routing
│   │   ├── query.go       # Query engine with caching
│   │   └── types.go       # JSON-RPC types
│   └── quanticscript/     # QuanticScript language implementation
│       ├── lexer.go        # Tokenization with source locations
│       ├── parser.go       # AST construction
│       ├── typechecker.go  # Type inference and validation
│       ├── codegen.go      # AST to bytecode compilation
│       ├── interpreter.go  # Bytecode execution
│       ├── assembler.go    # Assembly to bytecode
│       ├── disassembler.go # Bytecode to assembly
│       ├── stdlib.go       # Standard library functions
│       └── stdlib_programs.go # Builtin program helpers
├── programs/               # Smart contract programs
│   ├── system/             # System_Program (.qs, .qsa, .qsb)
│   ├── token/              # Token_Program (.qs, .qsa, .qsb)
│   └── staking/            # Staking_Program (.qs, .qsa, .qsb)
├── examples/               # Example QuanticScript programs
├── docs/                   # Documentation
├── go.mod
└── README.md
```

## Architecture

The system uses a layered architecture:

- **Application Layer**: CLI and node management
- **Consensus Layer**: Leader selection and block production
- **Network Layer**: P2P communication and serialization
- **Core Blockchain Layer**: PoH clock, entries, and blocks
- **Storage Layer**: Persistent ledger with SQLite
- **Verification Layer**: Chain integrity and validity verification
- **File-Based State Layer**: File store, transaction processing, and smart contract runtime
- **Transaction Layer**: Transaction builder with input validation and permission declarations

## Key Concepts

- **PoH Clock**: Sequential SHA-256 hash chain serving as a cryptographic timeline
- **Tick**: Output of 12,500 hash operations
- **Entry**: Ledger record containing hash link, tick count, and transaction data
- **Block**: Collection of entries produced during a 400ms slot, with versioning support for backwards compatibility
- **Block Version**: Format version identifier (currently v1) enabling future protocol upgrades
- **Block Header**: Metadata including previous block hash, Merkle root, state root, slot, timestamp, and block height
- **State Root**: Merkle root hash of all file state, enabling verification of file-based state transitions
- **Slot**: Time window for block production (minimum 64 ticks)
- **Slot Tolerance**: 100-slot (~40 second) window for accepting blocks to handle clock skew and network delays

## Usage

### CLI Commands for Account Management

The blockchain includes CLI commands for managing accounts and transactions. For programmatic transaction construction, see the [Transaction Builder API](docs/reference/transaction-builder.md).

```bash
# Create a new account (production-ready via CREATE_FILE transaction)
go run cmd/main.go account create --balance 1000000 --output keypair.json --state state.db

# Transfer balance between accounts
go run cmd/main.go transfer --from keypair.json --to <address> --amount 1000 --state state.db

# Query account information
go run cmd/main.go query --address <address> --state state.db

# Submit a transaction from JSON file
go run cmd/main.go submit --tx transaction.json --state state.db

# Check transaction status
go run cmd/main.go status --tx <tx-id> --state state.db

# Start RPC node for blockchain queries and transaction submission
go run cmd/main.go rpc --ledger-path ./validator1.db --state-path ./validator1_state.db

# Show help
go run cmd/main.go help
```

See [docs/guides/cli-usage.md](docs/guides/cli-usage.md) for detailed CLI documentation.

### Neon Wallet (TUI)

The Neon Wallet provides a modern Terminal User Interface for managing accounts, viewing transaction history, and executing transfers:

```bash
# Build the wallet
go build -o bin/neon-wallet ./cmd/wallet

# Start with default settings
./bin/neon-wallet

# Custom wallet path and RPC endpoint
./bin/neon-wallet --wallet-path ./my-wallet.dat --rpc-url http://localhost:9000
```

**Features:**
- BIP39 seed phrase generation (12 or 24 words)
- BIP44 key derivation for Ed25519
- AES-256-GCM wallet encryption
- Multi-seed phrase management
- Transaction building and signing
- RPC client integration
- Modern TUI with Bubble Tea framework

**First-Time Setup:**
When you run the wallet for the first time, the Wallet Creation Wizard guides you through:
1. Creating a new wallet or restoring from seed phrase
2. Choosing seed phrase length (12 or 24 words)
3. Securely displaying your seed phrase (only shown once!)
4. Setting a strong password to encrypt your wallet

See [docs/guides/wallet-usage.md](docs/guides/wallet-usage.md) for complete wallet documentation.

### Validator Wallet

The validator wallet system (`internal/wallet`) provides secure, encrypted storage for Ed25519 keypairs used by validator nodes for identity and block signing. This is separate from the user-facing Neon Wallet.

**Features:**
- AES-256-GCM encryption with Argon2id key derivation
- Platform-specific secure storage (Linux/macOS/Windows)
- Ed25519 keypair generation and management
- Export/import for backup and migration
- Password-protected with minimum 8-character requirement

**Usage:**

```go
import "github.com/poh-blockchain/internal/wallet"

// Create new validator wallet
w, err := wallet.Create("validator1", "strong-password")

// Open existing wallet
w, err := wallet.Open("validator1", "strong-password")

// List all wallets
wallets, err := wallet.List()

// Export for backup
err := w.Export("/secure/backup.json")

// Import from backup
w, err := wallet.Import("/secure/backup.json", "validator2", "password")
```

**Node Integration:**

```bash
# Start validator with wallet identity
./poh-blockchain --wallet validator1

# Start in observer mode (no wallet)
./poh-blockchain
```

See [docs/reference/validator-wallet.md](docs/reference/validator-wallet.md) for complete validator wallet documentation.

### Storage Cost Calculator

Calculate the storage cost for any file before deploying it to the blockchain:

```bash
go run calculate_storage_cost.go <file_path>
```

**Example:**
```bash
go run calculate_storage_cost.go programs/token/token.qsb
```

This utility helps estimate the Neon balance required to store program bytecode or data on-chain using the exponential storage cost formula. See [docs/reference/cost-model.md](docs/reference/cost-model.md) for details on storage costs.

### Quick Demo with devnet.sh

The easiest way to see the blockchain in action is to use the devnet script, which starts a local development network:

```bash
# Start with default 3 validators
./devnet.sh start

# Or specify the number of validators
./devnet.sh start 5

# Check network status
./devnet.sh status

# View logs
./devnet.sh logs

# Stop the network
./devnet.sh stop
```

This will:
- Build the project
- Start a leader node on port 8000
- Start the specified number of replica nodes on ports 8001, 8002, etc.
- Run nodes as background processes
- Maintain state across restarts

**devnet.sh Commands:**
- `start [N]`: Start network with N validators (default: 3)
- `stop`: Stop all validators
- `restart [N]`: Restart network
- `status`: Show validator status and block counts
- `logs [ID]`: View logs for specific validator or all
- `clean`: Stop network and remove all data

### Comprehensive Testing with audit.sh

Run the audit script to validate all blockchain functionality:

```bash
# Run with default settings (30s per phase, 3 validators)
./audit.sh

# Run with custom settings
./audit.sh --duration 60 --validators 5

# Run in CI mode (no colors, no prompts)
./audit.sh --ci
```

The audit script tests:
- **Phase 1**: Basic consensus with honest nodes
- **Phase 2**: BFT with tolerance (4 honest + 1 malicious)
- **Phase 3**: BFT without tolerance (2 honest + 2 malicious)
- **Phase 4**: DPoS lifecycle (delegation, epochs, rewards, slashing)

Results are saved to `logs/audit-TIMESTAMP.json` for analysis.

### Command-Line Options

- `--type`: Node type, either "leader" or "replica" (default: "replica")
- `--port`: Port to listen on for P2P connections (default: 8080)
- `--peers`: Comma-separated list of peer addresses in format "host:port"
- `--db`: Path to the SQLite database file (default: "blockchain.db")
- `--malicious`: Run node in malicious mode for BFT testing (default: false)
  - Enables Byzantine fault behaviors for testing network resilience
  - See BFT-TESTING.md for detailed malicious behaviors

### Running a Leader Node

```bash
go run cmd/main.go --type=leader --port=8000 --db=./leader.db
```

The leader node will:
- Initialize the PoH clock and start continuous block production
- Initialize DPoS genesis with default 2-validator configuration (10 Neon and 5 Neon stakes)
- Configure epoch length to 432,000 slots (~2 days at 400ms/slot)
- Produce blocks every 400ms slot (minimum 64 ticks per block)
- Store blocks to the local ledger
- Broadcast blocks to all connected replica nodes

**Note:** In production, the genesis validator configuration would be loaded from a config file. The current implementation uses hardcoded validators for development purposes.

### Running a Replica Node

```bash
go run cmd/main.go --type=replica --port=8001 --peers=localhost:8000 --db=./replica.db
```

The replica node will:
- Connect to the specified leader/peer nodes
- Receive and validate blocks from the network
- Verify block integrity and chain linkage
- Store valid blocks to the local ledger

### Running a Malicious Node (for BFT Testing)

```bash
# Malicious leader
go run cmd/main.go --type=leader --port=8000 --db=./leader.db --malicious

# Malicious replica
go run cmd/main.go --type=replica --port=8002 --peers=localhost:8000 --db=./malicious.db --malicious
```

Malicious nodes exhibit Byzantine fault behaviors for testing network resilience. See [docs/testing/bft-testing.md](docs/testing/bft-testing.md) for details on malicious behaviors and testing scenarios.

### Running a Multi-Node Network

Start a leader node:
```bash
go run cmd/main.go --type=leader --port=8000 --db=./leader.db
```

Start replica nodes connecting to the leader:
```bash
go run cmd/main.go --type=replica --port=8001 --peers=localhost:8000 --db=./replica1.db
go run cmd/main.go --type=replica --port=8002 --peers=localhost:8000 --db=./replica2.db
```

### Validator TUI Dashboard

Monitor validator and staking information in real-time with the validator TUI:

```bash
# Build the validator TUI
go build -o validator-tui ./cmd/validator-tui/main.go

# Run the TUI pointing to a FileStore database
./validator-tui --state ./state.db
```

**Features:**
- Built with Bubble Tea framework for rich terminal UI
- Real-time epoch and slot information
- Active validator count and status
- Validator records with commission, blocks produced, and missed blocks
- Stake account information
- Reward pool balance and estimated APY
- Slashing status indicators with color-coded display
- 1-second refresh interval
- Press 'q' or Ctrl+C to exit
- Automatic state initialization if directory doesn't exist

**Dashboard Layout:**
```
╔════════════════════════════════════════════════════════════════════════════════╗
║                         VALIDATOR TUI DASHBOARD                                ║
╚════════════════════════════════════════════════════════════════════════════════╝

Epoch: 0 | Slot: 0 | Local Validator: not_set | Active Validators: 0 | Local Stake: 0 electrons

┌──────────────────┬──────────┬──────────────────┬────────┬────────┬──────┬────────┐
│ Pubkey (16 hex)  │ Status   │ Total Stake      │ Comm % │ Blocks │ Miss │ Slashed│
├──────────────────┼──────────┼──────────────────┼────────┼────────┼──────┼────────┤
│ (validator rows)                                                                 │
└──────────────────┴──────────┴──────────────────┴────────┴────────┴──────┴────────┘

┌──────────────────────────────────────────────────────────────────────────────────┐
│ Total Staked: 0 electrons | Reward Pool: 0 electrons | Est. APY: 0.00% │
└──────────────────────────────────────────────────────────────────────────────────┘

Press 'q' or Ctrl+C to exit
Last refresh: 15:04:05
```

See [docs/guides/validator-tui.md](docs/guides/validator-tui.md) for detailed usage.

### Graceful Shutdown

Press `Ctrl+C` or send `SIGINT`/`SIGTERM` to gracefully shutdown the node. The node will:
- Stop block production/reception
- Close all network connections
- Flush and close the database
- Exit cleanly

## Development Status

This project is under active development. See `.kiro/specs/` for feature specifications and roadmaps.

### In Development

- **RPC Node and Wallet** — JSON-RPC 2.0 API for blockchain queries and transaction submission, TUI wallet with BIP39/44 support (see `.kiro/specs/rpc-node-and-wallet/`)
  - ✅ RPC server with HTTP transport and CORS support
  - ✅ Query engine with caching for blockchain data
  - ✅ Transaction submission with signature verification
  - ✅ All RPC methods implemented (getBalance, getAccountInfo, getBlockHeight, getRecentBlockhash, getTransactionHistory, sendTransaction, getTransactionStatus)
  - ✅ Wallet configuration management
  - ✅ BIP39 mnemonic generation and validation
  - ✅ BIP44 key derivation for Ed25519
  - ✅ AES-256-GCM wallet encryption
  - ✅ Multi-seed phrase management (import multiple independent seed phrases)
  - ✅ RPC client for wallet (JSON-RPC 2.0 client with comprehensive error handling)
  - 🚧 Transaction building and signing (in progress)
  - 🚧 TUI with Bubble Tea framework (in progress)

- **Delegated Proof of Stake (DPoS)** — validator registration, stake delegation, epoch scheduling, reward distribution, and slashing (see `.kiro/specs/delegated-proof-of-stake/`)
  - ✅ Genesis initialization with configurable validators
  - ✅ ConsensusManager integration with FileStore and Runtime
  - ✅ Automatic DPoS state initialization on node startup
  - ✅ Validator TUI dashboard
  - 🚧 Demo script implementation (in progress)

### Completed

- [x] Project initialization and Go module setup
- [x] PoH Clock Generator
- [x] Core blockchain data structures (Transaction, Entry, Block, BlockHeader with versioning)
- [x] JSON serialization with hex encoding
- [x] Block Producer with Merkle root calculation
- [x] Network Layer — TCP P2P with block broadcasting
- [x] Consensus Manager — leader-based, 400ms slots
- [x] Ledger Storage — SQLite with full CRUD and chain recovery
- [x] Verification Engine — entry hash chain, block linkage, full chain verification
- [x] Main application and CLI — node flags, leader/replica logic, graceful shutdown
- [x] Integration tests — block production, leader-replica communication, ledger persistence
- [x] File-based state model — BadgerDB persistence, storage cost with exponential growth
- [x] Transaction and Instruction structures — serialization, signature verification, fee calculation
- [x] Access Control System — permission validation, access logging, concurrent safety
- [x] Blockchain integration — StateRoot in BlockHeader, FileTransactions in Entry
- [x] Transaction Processor — atomic execution with automatic rollback
- [x] System Program (Go-side) — CreateAccount, Transfer, CloseAccount, AllocateData
- [x] Runtime System — builtin program registry with compute limits
- [x] Genesis Loader — idempotent bootstrap of System_Program and Token_Program at startup
- [x] Parallel Execution Analyzer — conflict detection for transaction scheduling
- [x] QuanticScript Lexer, Parser, Type Checker, Code Generator
- [x] QuanticScript Interpreter with full instruction set and cost metering
- [x] QuanticScript Assembler and Disassembler
- [x] QuanticScript Standard Library (string, math, crypto, blockchain, collections, invoke)
- [x] DISPATCH opcode and instruction registry for smart contract routing
- [x] System_Program and Token_Program written in QuanticScript
- [x] Cross-Program Invocation — INVOKE/INVOKERET with depth tracking (max 4 levels)

**Token System:**

The lexer tokenizes QuanticScript source code into a stream of tokens for parsing:
- Keywords: `function`, `export`, `import`, `let`, `const`, `if`, `else`, `while`, `for`, `return`, `__asm__`
- Type annotations: `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`, `bool`, `string`, `bytes`, `void`
- Operators: arithmetic (`+`, `-`, `*`, `/`, `%`), comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical (`&&`, `||`, `!`)
- Bitwise operators: `&`, `|`, `^`, `~`, `<<`, `>>`
- Source location tracking for error reporting with filename, line, and column information

**Code Generator:**

The code generator compiles QuanticScript source code (AST) to bytecode:

```go
import "github.com/poh-blockchain/internal/quanticscript"

// Parse source code
lexer := quanticscript.NewLexer(source, "program.qs")
parser := quanticscript.NewParser(lexer)
program := parser.ParseProgram()

// Type check
typeChecker := quanticscript.NewTypeChecker()
typeChecker.CheckProgram(program)

// Generate bytecode
codeGen := quanticscript.NewCodeGenerator()
bytecode, err := codeGen.Generate(program)
```

**Code Generator Features:**
- Translates high-level expressions to stack-based bytecode operations
- Generates control flow instructions (if/else, while, for loops)
- Allocates local memory slots for variables
- Handles function calls with automatic label resolution
- Supports inline assembly blocks with variable binding
- Compound assignment operators (+=, -=, *=, /=, %=)
- Two-pass compilation with forward reference resolution
- Comprehensive error reporting with source locations

**Assembly Language:**

QuanticScript includes a human-readable assembly language for low-level programming and debugging:

```assembly
; Simple addition program
entry:
    PUSH i64 42      ; Push integer 42
    PUSH i64 10      ; Push integer 10
    ADD              ; Add top two values
    RET              ; Return result

; Function with labels
loop_example:
    PUSH i64 0
    STORE 0          ; counter = 0
loop_start:
    LOAD 0           ; Load counter
    PUSH i64 10
    LT               ; counter < 10?
    JMPIF loop_body
    RET
loop_body:
    LOAD 0
    PUSH i64 1
    ADD
    STORE 0          ; counter++
    JMP loop_start
```

**Assembler Features:**
- Human-readable mnemonics for all bytecode instructions
- Label support for jumps and function calls
- Comment support (semicolon-prefixed)
- Type annotations for PUSH instructions
- Automatic label resolution with relative/absolute offsets
- Error reporting with line numbers

**High-Level Syntax:**

QuanticScript provides TypeScript-like syntax for writing smart contracts:

```typescript
// Simple addition function
export function add(a: i64, b: i64): i64 {
    return a + b;
}

// Function with control flow
function factorial(n: i64): i64 {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

// Function with loops
function sum_range(start: i64, end: i64): i64 {
    let total: i64 = 0;
    for (let i = start; i < end; i += 1) {
        total += i;
    }
    return total;
}

// Function with inline assembly
function optimized_multiply(a: i64, b: i64): i64 {
    let result: i64;
    __asm__ {
        LOAD a
        LOAD b
        MUL
        STORE result
    }
    return result;
}
```

**Usage:**

```go
import "github.com/poh-blockchain/internal/quanticscript"

// Assemble to bytecode file (with header)
bytecode, err := quanticscript.AssembleToFile(assemblySource)

// Assemble to bytecode body only (no header)
body, err := quanticscript.AssembleToBody(assemblySource)

// Disassemble bytecode to assembly text
assembly, err := quanticscript.Disassemble(bytecode)
```

See `.kiro/specs/` for detailed specifications on each subsystem.

## Testing

The project includes comprehensive integration tests that verify the complete blockchain functionality.

### Running Tests

```bash
# Run all tests including integration tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run only integration tests
go test -v ./internal -run Integration

# Run QuanticScript tests
go test -v ./internal/quanticscript
```

### Integration Tests

Three main integration tests are included in `internal/integration_test.go`:

1. **TestFullNodeBlockProductionAndVerification**
   - Tests complete node initialization and block production
   - Produces multiple blocks with proper linkage
   - Verifies chain integrity and block height
   - Validates the entire blockchain from genesis

2. **TestLeaderReplicaCommunication**
   - Tests P2P communication between leader and replica nodes
   - Verifies block broadcasting and reception
   - Validates consensus and verification on replica
   - Ensures both nodes maintain identical blockchain state

3. **TestLedgerPersistenceAndRecovery**
   - Tests database persistence across restarts
   - Creates blockchain, closes database, then reopens
   - Verifies all blocks are correctly recovered
   - Validates chain integrity after recovery

### QuanticScript Tests

The QuanticScript language implementation includes comprehensive unit tests in `internal/quanticscript/`:

- **Lexer Tests** (`lexer_test.go`) - Tokenization and source location tracking
- **Parser Tests** (`parser_test.go`) - AST construction and error handling
- **Type Checker Tests** (`typechecker_test.go`) - Type inference and validation
- **Code Generator Tests** (`codegen_test.go`) - Bytecode emission and optimization
- **Interpreter Tests** (`interpreter_test.go`) - Bytecode execution and stack operations
- **Assembler Tests** (`assembler_test.go`) - Assembly to bytecode conversion
- **Disassembler Tests** (`disassembler_test.go`) - Bytecode to assembly conversion
- **Standard Library Tests** (`stdlib_test.go`) - Built-in functions and modules
- **Comprehensive Tests** (`comprehensive_test.go`) - End-to-end language features

The comprehensive test file includes bytecode helper functions for building test cases:
- `buildPushI64`, `buildPushU64`, `buildPushBool` - Push value instructions
- `buildPushString`, `buildPushBytes` - Push complex types
- `buildJump`, `buildJumpIf` - Control flow instructions

These helpers simplify testing of control flow, data structures, arithmetic operations, blockchain state access, cross-program invocation, string operations, cryptographic functions, and function calls.

### File-Based State Integration Tests

Eight integration tests for the file-based state model are included in `internal/e2e_transaction_test.go` and `internal/access_control_integration_test.go`:

**E2E Transaction Tests:**
1. **TestEndToEndAccountCreationAndTransfer** - Full account lifecycle with balance transfers
2. **TestMultiInstructionTransactionAtomicity** - Atomic multi-instruction execution with rollback
3. **TestTransactionRevertOnInstructionFailure** - State rollback on instruction errors
4. **TestFeePaymentAndBalanceUpdates** - Fee deduction and balance management

**Access Control Integration Tests:**
5. **TestReadPermissionEnforcement** - Read-only permission validation and enforcement
6. **TestWritePermissionEnforcement** - Write permissions allow both read and write operations
7. **TestUndeclaredFileAccessDetection** - Undeclared file access rejection with state preservation
8. **TestPermissionViolationHandling** - Mid-execution permission violations with automatic state revert

### Test Database Cleanup

Integration tests create temporary SQLite databases that are automatically cleaned up after each test run.

## Documentation

Complete documentation is available in the [docs/](docs/) directory:

### Getting Started
- **[Quick Start Guide](docs/guides/quickstart.md)** - Get up and running in 30 seconds
- **[Demo Guide](docs/guides/demo.md)** - Interactive demos with tmux
- **[DPoS Demo Guide](docs/guides/dpos-demo.md)** - Delegated Proof of Stake demonstration
- **[Validator TUI Guide](docs/guides/validator-tui.md)** - Terminal dashboard for validators
- **[CLI Usage Guide](docs/guides/cli-usage.md)** - Command-line interface reference
- **[QuanticScript Guide](docs/guides/quanticscript.md)** - Smart contract language overview

### Testing
- **[BFT Testing Guide](docs/testing/bft-testing.md)** - Byzantine Fault Tolerance testing
- **[Automated Testing Guide](docs/testing/automated-testing.md)** - Testing without tmux
- **[Testing Summary](docs/testing/testing-summary.md)** - Quick reference for all testing features

### Reference
- **[Language Reference](docs/reference/language-reference.md)** - QuanticScript syntax and semantics
- **[Standard Library Reference](docs/reference/stdlib-reference.md)** - Built-in functions and modules
- **[Security Model](docs/reference/security-model.md)** - Security mechanisms and privilege restrictions
- **[RPC API Reference](docs/reference/rpc-api.md)** - JSON-RPC 2.0 API documentation
- **[Inline Assembly Guide](docs/reference/inline-assembly.md)** - Low-level assembly programming
- **[Bytecode Reference](docs/reference/bytecode-reference.md)** - Bytecode format and opcodes
- **[Cost Model Guide](docs/reference/cost-model.md)** - Understanding computational costs
- **[Implementation Summary](docs/reference/implementation-summary.md)** - Architecture and features
- **[BFT Fix Summary](docs/reference/bft-fix-summary.md)** - Technical details of BFT fixes

### Specifications
Detailed specifications are available in the `.kiro/specs/` directory:
- `poh-blockchain/`: Core blockchain implementation specs
- `quanticscript-language/`: QuanticScript language specs
- `file-based-state/`: State model and transaction processing specs

## License

[Add license information]

## Contributing

[Add contribution guidelines]

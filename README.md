# PoH Blockchain

A Proof of History (PoH) blockchain implementation inspired by Solana's architecture, built in Go.

## Quick Start

```bash
# Run a basic demo (with tmux)
./demo.sh

# Test Byzantine Fault Tolerance (with tmux)
./demo-bft.sh 3 1

# Automated testing (no tmux, AI-friendly)
./demo-automated.sh 3 1 15
./analyze-results.sh
```

See [QUICKSTART.md](QUICKSTART.md) for a 30-second introduction.

## Overview

This project implements a verifiable delay function using sequential SHA-256 hashing to create a cryptographic clock, enabling high-throughput transaction ordering without traditional consensus overhead.

## Features

- Proof of History clock generator with verifiable sequential hashing
- Leader-based consensus protocol with 400ms slots
- P2P network communication for block distribution
- SQLite-based persistent ledger storage
- Full chain verification and integrity checking
- Transaction integration with cryptographic timestamping
- **Byzantine Fault Tolerance testing with malicious nodes**
- **Automated demo scripts with tmux visualization**
- **AI-friendly automated testing tools (no tmux required)**
- **Comprehensive test launcher with reporting**

## Requirements

- Go 1.21 or higher
- tmux (for demo scripts)
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
├── internal/
│   ├── poh/               # PoH clock generator
│   ├── blockchain/        # Core blockchain data structures and serialization
│   │   ├── types.go       # Transaction, Entry, Block, BlockHeader definitions with versioning
│   │   ├── serialization.go # JSON marshaling with hex encoding
│   │   └── block_producer.go # Block production and transaction integration
│   ├── network/           # P2P networking layer
│   │   └── network_node.go # TCP-based node communication
│   ├── consensus/         # Consensus protocol
│   │   └── consensus_manager.go # Leader selection, slot timing, block validation
│   ├── storage/           # Ledger persistence
│   │   └── ledger.go      # SQLite-based blockchain storage with CRUD operations
│   ├── verification/      # Chain verification
│   │   └── verifier.go    # Entry, block, and full chain integrity verification
│   ├── filestore/         # File-based state model (in development)
│   │   ├── filestore.go   # File data structures, storage cost calculation
│   │   └── filestore_test.go # Comprehensive unit tests
│   ├── runtime/           # Program execution runtime (planned)
│   └── system/            # System program for account management (planned)
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
- **File-Based State Layer**: File store, transaction processing, and smart contract runtime (in development)

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

The blockchain now includes CLI commands for managing accounts and transactions:

```bash
# Create a new account
go run cmd/main.go account create --balance 1000000 --output keypair.json --state state.db

# Transfer balance between accounts
go run cmd/main.go transfer --from keypair.json --to <address> --amount 1000 --state state.db

# Query account information
go run cmd/main.go query --address <address> --state state.db

# Submit a transaction from JSON file
go run cmd/main.go submit --tx transaction.json --state state.db

# Check transaction status
go run cmd/main.go status --tx <tx-id> --state state.db

# Show help
go run cmd/main.go help
```

See [CLI-USAGE.md](CLI-USAGE.md) for detailed CLI documentation.

### Quick Demo with tmux

The easiest way to see the blockchain in action is to use the demo script, which starts a leader and multiple replicas in tmux panes:

```bash
# Start with default 2 replicas
./demo.sh

# Or specify the number of replicas (1-9)
./demo.sh 3
./demo.sh 5
```

This will:
- Build the project
- Clean up old database files
- Start a leader node on port 8000
- Start the specified number of replica nodes on ports 8001, 8002, etc.
- Display all nodes in a tiled tmux layout

**Example with 2 replicas:**
```
┌─────────────────────────────┐
│        Leader Node          │
├─────────────┬───────────────┤
│  Replica 1  │   Replica 2   │
└─────────────┴───────────────┘
```

**Example with 4 replicas:**
```
┌─────────────┬───────────────┐
│   Leader    │   Replica 1   │
├─────────────┼───────────────┤
│  Replica 2  │   Replica 3   │
├─────────────┴───────────────┤
│        Replica 4            │
└─────────────────────────────┘
```

**tmux Commands:**
- Detach from session: `Ctrl+B`, then `D`
- Navigate between panes: `Ctrl+B`, then arrow keys
- Stop all nodes: `Ctrl+C` in each pane, or run `./stop-demo.sh`

### BFT Testing Demo

Test Byzantine Fault Tolerance with malicious nodes:

```bash
# Run with honest and malicious replicas
./demo-bft.sh 3 1    # 3 honest + 1 malicious (has BFT)
./demo-bft.sh 4 2    # 4 honest + 2 malicious (has BFT)
./demo-bft.sh 2 2    # 2 honest + 2 malicious (NO BFT)
```

The script automatically calculates BFT status:
- **BFT Requirement**: Honest nodes > 2 × Malicious nodes
- **With BFT**: Network rejects invalid blocks, maintains integrity
- **Without BFT**: Network may accept corrupted blocks

**Malicious behaviors include**:
- Sending blocks with invalid hash counts
- Broadcasting blocks with wrong previous hashes
- Skipping validation and accepting invalid blocks
- Storing unvalidated blocks

See `DEMO.md` for detailed BFT testing scenarios and expected outcomes.

### Automated Testing (No tmux)

For AI agents and automated testing without tmux:

```bash
# Run a single test scenario
./demo-automated.sh 3 1 15    # 3 honest + 1 malicious, 15 seconds

# Analyze results
./analyze-results.sh

# Run comprehensive test suite
./test-launcher.sh 20         # Run all scenarios, 20s each
```

**Features**:
- Clear log prefixes: `[LEADER]`, `[HONEST-1]`, `[MALICIOUS-1]`
- Automatic results analysis
- Generates markdown reports
- Saves logs and databases for inspection

See [AUTOMATED-TESTING.md](AUTOMATED-TESTING.md) for complete guide.

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
- Produce blocks every 400ms slot (minimum 64 ticks per block)
- Store blocks to the local ledger
- Broadcast blocks to all connected replica nodes

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

Malicious nodes exhibit Byzantine fault behaviors for testing network resilience. See [BFT-TESTING.md](BFT-TESTING.md) for details on malicious behaviors and testing scenarios.

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

### Graceful Shutdown

Press `Ctrl+C` or send `SIGINT`/`SIGTERM` to gracefully shutdown the node. The node will:
- Stop block production/reception
- Close all network connections
- Flush and close the database
- Exit cleanly

## Development Status

This project is currently under active development. See `.kiro/specs/poh-blockchain/tasks.md` for the PoH blockchain implementation roadmap.

### File-Based State Model (Production Ready)

The file-based state model enables smart contract functionality and account management. This architecture treats all on-chain state (user accounts, programs, data) as uniform file objects, inspired by Unix's "everything is a file" philosophy and Solana's account model.

**Key Features:**
- Uniform file abstraction for accounts, programs, and data
- Explicit transaction access patterns for future parallel execution
- Storage cost mechanics with exponential growth
- Built-in System Program for account management
- Runtime system for program execution with builtin registry
- Access control with read/write permission validation
- CLI tools for account creation, transfers, and queries

**Completed Components:**
- ✅ File data structures with BadgerDB persistence
- ✅ Storage cost calculation and validation
- ✅ Transaction and Instruction types with serialization
- ✅ Fee calculation system
- ✅ AccessController for permission validation and access logging
- ✅ ExecutionContext for program execution with file access API
- ✅ Runtime system with builtin program registry and validation
- ✅ System Program with account management operations (CreateAccount, Transfer, CloseAccount, AllocateData)
- ✅ Transaction Processor with atomic execution and automatic rollback
- ✅ CLI commands for account management and transaction operations
- ✅ End-to-end integration tests (12 tests covering transaction flow, access control, and storage cost enforcement)

See `.kiro/specs/file-based-state/` for detailed specifications:
- `requirements.md`: Functional requirements
- `design.md`: Architecture and component design
- `tasks.md`: Implementation progress
- `IMPLEMENTATION-STATUS.md`: Current status and test coverage

### Completed

- [x] Project initialization and Go module setup
- [x] PoH Clock Generator - Full implementation with thread-safe operations
- [x] Core blockchain data structures - Transaction, Entry, Block, and BlockHeader types with version support for backwards compatibility
- [x] JSON serialization - Custom marshaling/unmarshaling with hex encoding for byte slices
- [x] Block Producer - Transaction integration, entry production, and block generation with Merkle root calculation
- [x] Network Layer - P2P communication with TCP, block broadcasting/receiving, message framing
- [x] Consensus Manager - Leader-based consensus with slot timing, block validation, and 400ms slot duration
- [x] Ledger Storage - SQLite-based persistent storage with full CRUD operations and blockchain state recovery
- [x] Verification Engine - Complete chain integrity verification with entry hash chain validation, block verification, block linkage checking, and full chain verification from genesis

### Completed (Continued)

- [x] Main application and CLI - Full node initialization with command-line flags, leader/replica logic, graceful shutdown handling
- [x] Integration tests - Complete end-to-end testing including full node block production, leader-replica communication, and ledger persistence/recovery
- [x] File-based state model foundation - File data structures, storage cost calculation with exponential growth, BadgerDB persistence
- [x] Transaction and Instruction structures - Complete implementation with JSON serialization, signature verification, and fee calculation
- [x] Access Control System - Full AccessController implementation with permission validation, access logging, and concurrent safety
- [x] Blockchain integration - StateRoot field added to BlockHeader for file state verification, Entry structure supports FileTransactions
- [x] Transaction Processor - Complete implementation with atomic execution, automatic rollback, and fee management
- [x] System Program - Built-in program for account management (CreateAccount, Transfer, CloseAccount, AllocateData)
- [x] Runtime System - Builtin program registry with execution validation and compute limits
- [x] End-to-End Integration Tests - 8 comprehensive tests covering transaction flow and access control validation

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

### Quick References
- **[QUICKSTART.md](QUICKSTART.md)** - 30-second introduction to running demos
- **[DEMO.md](DEMO.md)** - Complete demo guide with tmux commands and troubleshooting
- **[BFT-TESTING.md](BFT-TESTING.md)** - In-depth Byzantine Fault Tolerance testing guide
- **[AUTOMATED-TESTING.md](AUTOMATED-TESTING.md)** - Automated testing guide for AI agents (no tmux)
- **[TESTING-SUMMARY.md](TESTING-SUMMARY.md)** - Comprehensive testing reference

### Specifications
Detailed specifications are available in the `.kiro/specs/poh-blockchain/` directory:
- `requirements.md`: Functional requirements and acceptance criteria
- `design.md`: Architecture and component design
- `tasks.md`: Implementation plan and task breakdown

## License

[Add license information]

## Contributing

[Add contribution guidelines]

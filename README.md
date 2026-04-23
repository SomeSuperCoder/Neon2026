# PoH Blockchain

A Proof of History (PoH) blockchain implementation inspired by Solana's architecture, built in Go.

## Overview

This project implements a verifiable delay function using sequential SHA-256 hashing to create a cryptographic clock, enabling high-throughput transaction ordering without traditional consensus overhead.

## Features

- Proof of History clock generator with verifiable sequential hashing
- Leader-based consensus protocol with 400ms slots
- P2P network communication for block distribution
- SQLite-based persistent ledger storage
- Full chain verification and integrity checking
- Transaction integration with cryptographic timestamping

## Requirements

- Go 1.21 or higher
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
│   │   ├── types.go       # Transaction, Entry, Block, BlockHeader definitions
│   │   ├── serialization.go # JSON marshaling with hex encoding
│   │   └── block_producer.go # Block production and transaction integration
│   ├── network/           # P2P networking layer
│   │   └── network_node.go # TCP-based node communication
│   ├── consensus/         # Consensus protocol
│   │   └── consensus_manager.go # Leader selection, slot timing, block validation
│   ├── storage/           # Ledger persistence
│   │   └── ledger.go      # SQLite-based blockchain storage with CRUD operations
│   └── verification/      # Chain verification
│       └── verifier.go    # Entry, block, and full chain integrity verification
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

## Key Concepts

- **PoH Clock**: Sequential SHA-256 hash chain serving as a cryptographic timeline
- **Tick**: Output of 12,500 hash operations
- **Entry**: Ledger record containing hash link, tick count, and transaction data
- **Block**: Collection of entries produced during a 400ms slot
- **Slot**: Time window for block production (minimum 64 ticks)

## Usage

### Command-Line Options

- `--type`: Node type, either "leader" or "replica" (default: "replica")
- `--port`: Port to listen on for P2P connections (default: 8080)
- `--peers`: Comma-separated list of peer addresses in format "host:port"
- `--db`: Path to the SQLite database file (default: "blockchain.db")

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

This project is currently under active development. See `.kiro/specs/poh-blockchain/tasks.md` for the implementation roadmap.

### Completed

- [x] Project initialization and Go module setup
- [x] PoH Clock Generator - Full implementation with thread-safe operations
- [x] Core blockchain data structures - Transaction, Entry, Block, and BlockHeader types
- [x] JSON serialization - Custom marshaling/unmarshaling with hex encoding for byte slices
- [x] Block Producer - Transaction integration, entry production, and block generation with Merkle root calculation
- [x] Network Layer - P2P communication with TCP, block broadcasting/receiving, message framing
- [x] Consensus Manager - Leader-based consensus with slot timing, block validation, and 400ms slot duration
- [x] Ledger Storage - SQLite-based persistent storage with full CRUD operations and blockchain state recovery
- [x] Verification Engine - Complete chain integrity verification with entry hash chain validation, block verification, block linkage checking, and full chain verification from genesis

### Completed (Continued)

- [x] Main application and CLI - Full node initialization with command-line flags, leader/replica logic, graceful shutdown handling
- [x] Integration tests - Complete end-to-end testing including full node block production, leader-replica communication, and ledger persistence/recovery

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

### Test Database Cleanup

Integration tests create temporary SQLite databases that are automatically cleaned up after each test run.

## Documentation

Detailed specifications are available in the `.kiro/specs/poh-blockchain/` directory:

- `requirements.md`: Functional requirements and acceptance criteria
- `design.md`: Architecture and component design
- `tasks.md`: Implementation plan and task breakdown

## License

[Add license information]

## Contributing

[Add contribution guidelines]

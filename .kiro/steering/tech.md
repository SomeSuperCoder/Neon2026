---
inclusion: always
---

# Tech Stack

## Language & Runtime

- **Go 1.23.0+** - Primary implementation language
- Standard library for core functionality
- No external frameworks for blockchain logic

## Dependencies

- **SQLite3** (`github.com/mattn/go-sqlite3`) - Blockchain ledger persistence
- **BadgerDB v4** (`github.com/dgraph-io/badger/v4`) - File-based state storage
- **Ed25519** (crypto/ed25519) - Cryptographic signatures
- **SHA-256** (crypto/sha256) - PoH clock hashing

## Build System

Standard Go toolchain with modules:

```bash
# Build the main binary
go build -o poh-blockchain ./cmd/main.go

# Install dependencies
go mod download

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific integration tests
go test -v ./internal -run Integration
```

## Common Commands

### Node Operations

```bash
# Run as leader node
go run cmd/main.go --type=leader --port=8000 --db=./leader.db

# Run as replica node
go run cmd/main.go --type=replica --port=8001 --peers=localhost:8000 --db=./replica.db

# Run malicious node (BFT testing)
go run cmd/main.go --type=leader --port=8000 --db=./leader.db --malicious
```

### CLI Commands

```bash
# Create account
go run cmd/main.go account create --balance 1000000 --output keypair.json --state state.db

# Transfer balance
go run cmd/main.go transfer --from keypair.json --to <address> --amount 1000 --state state.db

# Query account
go run cmd/main.go query --address <address> --state state.db

# Calculate storage cost
go run calculate_storage_cost.go <file_path>
```

### QuanticScript Compiler

```bash
# Compile source to bytecode
go run cmd/main.go qsc compile -i program.qs -o program.qsb

# Assemble assembly to bytecode
go run cmd/main.go qsc assemble -i program.qsa -o program.qsb

# Disassemble bytecode to assembly
go run cmd/main.go qsc disassemble -i program.qsb -o program.qsa
```

### Demo Scripts

```bash
# Interactive demo with tmux (2-9 replicas)
./demo.sh 3

# BFT testing demo (honest + malicious nodes)
./demo-bft.sh 3 1

# Automated testing (no tmux, AI-friendly)
./demo-automated.sh 3 1 15
./analyze-results.sh

# Comprehensive test suite
./test-launcher.sh 20
```

## Testing

- Use standard Go testing package (`testing`)
- Integration tests in `internal/*_integration_test.go`
- Unit tests alongside implementation files (`*_test.go`)
- Test helper: `checkParserErrors(t, parser)` pattern for compiler tests
- No external testing frameworks required

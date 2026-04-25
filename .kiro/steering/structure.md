---
inclusion: always
---

# Project Structure

## Directory Organization

```
.
├── cmd/                          # Application entry points
│   └── main.go                   # Main CLI with node operations and subcommands
├── internal/                     # Internal packages (not importable externally)
│   ├── poh/                      # PoH clock generator
│   ├── blockchain/               # Core blockchain structures
│   │   ├── types.go              # Transaction, Entry, Block, BlockHeader
│   │   ├── serialization.go     # JSON marshaling with hex encoding
│   │   └── block_producer.go    # Block production and Merkle roots
│   ├── network/                  # P2P networking
│   │   └── network_node.go      # TCP-based node communication
│   ├── consensus/                # Consensus protocol
│   │   └── consensus_manager.go # Leader selection, slot timing, validation
│   ├── storage/                  # Ledger persistence
│   │   └── ledger.go            # SQLite blockchain storage
│   ├── verification/             # Chain verification
│   │   └── verifier.go          # Entry, block, and chain integrity
│   ├── filestore/                # File-based state model
│   │   ├── filestore.go         # File structures and storage costs
│   │   └── filestore_test.go    # Unit tests
│   ├── transaction/              # Transaction structures
│   │   ├── transaction.go       # Transaction, Instruction, Signature
│   │   └── transaction_test.go  # Unit tests
│   ├── access/                   # Access control
│   │   ├── access_controller.go # Permission validation
│   │   └── access_controller_test.go
│   ├── processor/                # Transaction processing
│   │   ├── tx_processor.go      # Atomic execution with rollback
│   │   └── tx_processor_test.go
│   ├── runtime/                  # Program execution
│   │   ├── runtime.go           # Builtin program registry
│   │   └── runtime_test.go
│   ├── system/                   # System program
│   │   ├── system.go            # Account management operations
│   │   └── system_test.go
│   ├── quanticscript/            # QuanticScript language implementation
│   │   ├── types.go             # Core type system and runtime values
│   │   ├── token.go             # Token definitions
│   │   ├── lexer.go             # Tokenization with source locations
│   │   ├── lexer_test.go
│   │   ├── parser.go            # AST construction
│   │   ├── parser_test.go
│   │   ├── ast.go               # AST node types
│   │   ├── typechecker.go       # Type inference and validation
│   │   ├── typechecker_test.go
│   │   ├── codegen.go           # AST to bytecode compilation
│   │   ├── codegen_test.go
│   │   ├── opcodes.go           # Bytecode instruction opcodes
│   │   ├── bytecode.go          # Bytecode format specification
│   │   ├── costs.go             # Instruction cost table
│   │   ├── assembler.go         # Assembly to bytecode
│   │   ├── assembler_test.go
│   │   ├── disassembler.go      # Bytecode to assembly
│   │   ├── disassembler_test.go
│   │   ├── interpreter.go       # Bytecode execution
│   │   ├── interpreter_test.go
│   │   ├── stdlib.go            # Standard library functions
│   │   ├── stdlib_test.go
│   │   ├── stdlib_programs.go   # Builtin programs (System, Token)
│   │   └── stdlib_programs_test.go
│   ├── parallel/                 # Parallel execution analysis
│   │   ├── analyzer.go
│   │   └── analyzer_test.go
│   └── *_integration_test.go     # Integration tests
├── programs/                     # Smart contract programs
│   ├── system/                   # System program (account management)
│   │   ├── system.qs            # Source code
│   │   ├── system.qsa           # Assembly
│   │   └── system.qsb           # Bytecode
│   └── token/                    # Token program
│       ├── token.qs
│       ├── token.qsa
│       └── token.qsb
├── examples/                     # Example programs
│   ├── simple_transfer.qs
│   ├── token_transfer.qs
│   ├── advanced_token.qs
│   └── assembly_example.qsa
├── docs/                         # Documentation
│   ├── guides/                   # User guides
│   ├── reference/                # Technical reference
│   └── testing/                  # Testing documentation
├── .kiro/                        # Kiro IDE configuration
│   ├── specs/                    # Feature specifications
│   └── steering/                 # AI steering rules
├── logs/                         # Demo and test logs
├── go.mod                        # Go module definition
└── go.sum                        # Dependency checksums
```

## Code Organization Patterns

### Package Structure

- Each `internal/` subdirectory is a Go package
- Package name matches directory name
- Tests live alongside implementation (`*_test.go`)
- Integration tests use `*_integration_test.go` naming

### File Naming Conventions

- Implementation: `component_name.go`
- Tests: `component_name_test.go`
- QuanticScript source: `*.qs`
- QuanticScript assembly: `*.qsa`
- QuanticScript bytecode: `*.qsb`

### Import Paths

All internal packages use the module path prefix:
```go
import "github.com/poh-blockchain/internal/quanticscript"
import "github.com/poh-blockchain/internal/filestore"
```

### Type Definitions

- Core types in dedicated files (e.g., `types.go`, `ast.go`)
- Related functionality grouped in same package
- Public types use PascalCase
- Private types use camelCase

### Error Handling

- Return errors explicitly, don't panic
- Use `fmt.Errorf` for error wrapping
- Custom error types for structured errors (e.g., `ParserError`, `CodeGenError`)
- Include source location in compiler errors

### Testing Patterns

- Table-driven tests for multiple cases
- Helper functions like `checkParserErrors(t, parser)`
- Integration tests verify end-to-end flows
- Use `t.Fatalf` for setup failures, `t.Errorf` for assertion failures

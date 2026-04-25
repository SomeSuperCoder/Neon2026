# Requirements Document

## Introduction

The project currently has two implementations of the System_Program: a native Go implementation in `internal/system/` and a QuanticScript implementation in `programs/system/system.qs` (compiled to `programs/system/system.qsb`). The goal is to remove the Go implementation entirely so that the QuanticScript bytecode version is the sole System_Program. This requires wiring the QuanticScript bytecode through the existing `genesis.LoadBuiltinPrograms` path (already in place), removing the Go `system` package, and updating all callers — including `cmd/main.go`, integration tests, and the runtime — to no longer depend on the Go builtin.

## Glossary

- **System_Program**: The built-in program responsible for account creation, balance transfers, and storage allocation. Identified by the well-known FileID `0x00...01`.
- **QuanticScript**: The TypeScript-like smart contract language used in this project, compiled to bytecode (`.qsb`).
- **Go system package**: The `internal/system/` Go package that currently implements System_Program logic natively.
- **Genesis loader**: The `genesis.LoadBuiltinPrograms` function that stores compiled `.qsb` bytecode into the FileStore at startup.
- **Runtime**: The `internal/runtime` package that dispatches program execution — either to a registered `BuiltinProgram` (Go) or to the QuanticScript bytecode interpreter.
- **BuiltinProgram interface**: The Go interface in `internal/runtime` that native Go programs implement (`Execute`, `GetProgramID`).
- **FileStore**: The BadgerDB-backed state store (`internal/filestore`) that holds all files including programs and accounts.
- **TxProcessor**: The `internal/processor` package that orchestrates transaction validation and instruction execution.
- **SystemProgramID**: The canonical FileID `0x00...01` identifying the System_Program, currently defined in both `internal/system` and `internal/genesis`.

## Requirements

### Requirement 1

**User Story:** As a blockchain node operator, I want the System_Program to run as QuanticScript bytecode so that all program logic is expressed uniformly in the same language and execution model.

#### Acceptance Criteria

1. WHEN the node initialises, THE System_Program SHALL be loaded exclusively from the compiled QuanticScript bytecode at `programs/system/system.qsb` via `genesis.LoadBuiltinPrograms`.
2. WHEN the Runtime receives an instruction targeting the System_Program FileID, THE Runtime SHALL execute the QuanticScript bytecode interpreter rather than a native Go handler.
3. THE codebase SHALL contain no import of the `github.com/poh-blockchain/internal/system` package outside of the `internal/system` directory itself.
4. WHEN the Go `internal/system` package is deleted, THE project SHALL compile without errors.

### Requirement 2

**User Story:** As a developer, I want a single canonical definition of SystemProgramID so that there is no risk of the two definitions diverging.

#### Acceptance Criteria

1. THE codebase SHALL define `SystemProgramID` (`0x00...01`) in exactly one location.
2. WHEN any package needs to reference the System_Program FileID, THE package SHALL import it from `internal/genesis`.
3. IF a file references `system.SystemProgramID`, THEN THE file SHALL be updated to reference `genesis.SystemProgramID` instead.

### Requirement 3

**User Story:** As a developer, I want the CLI commands (`account create`, `transfer`, `submit`) to continue working after the Go system package is removed.

#### Acceptance Criteria

1. WHEN `account create` is invoked, THE CLI SHALL create an account file with `TxManager` set to `genesis.SystemProgramID` without importing the Go `system` package.
2. WHEN `transfer` is invoked, THE CLI SHALL build and submit a transfer instruction targeting `genesis.SystemProgramID` without importing the Go `system` package.
3. WHEN `submit` is invoked, THE CLI SHALL process the transaction without registering a Go builtin for the System_Program.
4. THE CLI SHALL encode transfer instruction data using a helper that does not depend on the `internal/system` package.

### Requirement 4

**User Story:** As a developer, I want all existing tests to remain valid after the migration so that correctness is preserved.

#### Acceptance Criteria

1. WHEN the test suite is run after the migration, THE test suite SHALL pass with no new failures introduced by this change.
2. THE integration tests in `internal/builtin_programs_integration_test.go` and `internal/e2e_transaction_test.go` SHALL be updated to reference `genesis.SystemProgramID` and inline any previously imported helpers from the `system` package.
3. IF a test file imports `github.com/poh-blockchain/internal/system`, THEN THE test file SHALL be updated to remove that import.
4. THE `internal/system/system_test.go` file SHALL be deleted along with the rest of the `internal/system` package.

### Requirement 5

**User Story:** As a developer, I want the `ensureSystemProgram` stub in `cmd/main.go` removed so that the legacy Go-based stub no longer conflicts with the QuanticScript program loaded by genesis.

#### Acceptance Criteria

1. WHEN `initFileStore` is called, THE function SHALL NOT call `ensureSystemProgram` or create a stub file for the System_Program.
2. THE `ensureSystemProgram` function SHALL be deleted from `cmd/main.go`.
3. WHILE the node is running, THE System_Program file in the FileStore SHALL be the QuanticScript bytecode version loaded by `genesis.LoadBuiltinPrograms`.

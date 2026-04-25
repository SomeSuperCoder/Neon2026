# Design Document: Remove Go System Program

## Overview

The project has two System_Program implementations: a native Go package (`internal/system/`) and a QuanticScript bytecode program (`programs/system/system.qsb`). The QuanticScript version is already loaded at genesis via `genesis.LoadBuiltinPrograms` and executed by the bytecode interpreter. The Go version is registered as a `BuiltinProgram` in the Runtime and takes precedence over the bytecode path.

This design describes how to remove the Go implementation so the QuanticScript bytecode becomes the sole System_Program, with no loss of functionality and no new test failures.

## Architecture

### Current State

```
cmd/main.go
  └─ imports internal/system
  └─ rt.RegisterBuiltinProgram(system.NewSystemProgram())   ← Go builtin wins
  └─ ensureSystemProgram(fs)                                ← writes stub file
  └─ genesis.LoadBuiltinPrograms(...)                       ← writes QS bytecode (may be shadowed)

internal/builtin_programs_integration_test.go
  └─ imports internal/system
  └─ rt.RegisterBuiltinProgram(system.NewSystemProgram())

internal/e2e_transaction_test.go
  └─ imports internal/system
  └─ rt.RegisterBuiltinProgram(system.NewSystemProgram())

Runtime.ExecuteProgram()
  1. Check builtinPrograms map  ← Go handler fires here for System_Program
  2. Check isQuanticScriptBytecode
  3. Execute bytecode interpreter
```

### Target State

```
cmd/main.go
  └─ imports internal/genesis (for SystemProgramID constant)
  └─ NO rt.RegisterBuiltinProgram for System_Program
  └─ NO ensureSystemProgram
  └─ genesis.LoadBuiltinPrograms(...)  ← sole source of System_Program

Runtime.ExecuteProgram()
  1. Check builtinPrograms map  ← empty for System_Program
  2. Check isQuanticScriptBytecode  ← fires for System_Program
  3. Execute bytecode interpreter
```

## Components and Interfaces

### Deleted: `internal/system/` package

The entire package is removed:
- `internal/system/system.go`
- `internal/system/system_test.go`

No other package will import it after the migration.

### Modified: `cmd/main.go`

| Change | Detail |
|--------|--------|
| Remove import | `"github.com/poh-blockchain/internal/system"` |
| Add import | `"github.com/poh-blockchain/internal/genesis"` (already present) |
| Remove calls | `rt.RegisterBuiltinProgram(system.NewSystemProgram())` in `handleAccountCommand`, `handleTransferCommand`, `handleSubmitCommand` |
| Remove function | `ensureSystemProgram` |
| Remove call | `ensureSystemProgram(fs)` inside `initFileStore` |
| Replace constant | `system.SystemProgramID` → `genesis.SystemProgramID` |
| Inline helper | `system.EncodeTransferInstruction(amount)` → local inline encoding (1 byte type + 8 bytes LE amount) |

The transfer instruction encoding is trivial (9 bytes) and can be inlined directly in `handleTransferCommand` without a shared helper.

### Modified: `internal/builtin_programs_integration_test.go`

| Change | Detail |
|--------|--------|
| Remove import | `"github.com/poh-blockchain/internal/system"` |
| Remove call | `rt.RegisterBuiltinProgram(sysProg)` in `setupBuiltinEnv` |
| Replace constant | `system.SystemProgramID` → `genesis.SystemProgramID` |
| Replace helper | `system.EncodeTransferInstruction(n)` → inline encoding |

### Modified: `internal/e2e_transaction_test.go`

| Change | Detail |
|--------|--------|
| Remove import | `"github.com/poh-blockchain/internal/system"` |
| Remove call | `rt.RegisterBuiltinProgram(sysProg)` |
| Remove stub creation | The block that creates a `sysProgramFile` with `Data: []byte{}` and `Executable: true` at `system.SystemProgramID` — this stub is no longer needed because `genesis.LoadBuiltinPrograms` is called in `setupBuiltinEnv` (or must be added to the e2e setup) |
| Replace constant | `system.SystemProgramID` → `genesis.SystemProgramID` |
| Replace helper | `system.EncodeTransferInstruction(n)` → inline encoding |

### Unchanged: `internal/runtime/runtime.go`

The Runtime already supports QuanticScript bytecode execution. No changes are needed — removing the registered builtin simply causes `ExecuteProgram` to fall through to the bytecode interpreter path, which is already correct.

### Unchanged: `internal/genesis/programs.go`

`genesis.SystemProgramID` is already defined here. It becomes the single canonical definition.

### Unchanged: `programs/system/system.qsb`

The compiled bytecode is already embedded via `programs/embed.go` and loaded by `genesis.LoadBuiltinPrograms`.

## Data Models

No data model changes. The `filestore.File` structure for the System_Program remains identical — the only difference is that `Data` now contains real QuanticScript bytecode instead of the stub `[]byte("builtin-system-program")`.

## Instruction Encoding

The `system.EncodeTransferInstruction` helper produces:

```
byte[0]   = 0x01  (InstructionTransfer)
byte[1-8] = amount as uint64 little-endian
```

This is inlined wherever needed:

```go
func encodeTransferInstruction(amount int64) []byte {
    data := make([]byte, 9)
    data[0] = 1 // Transfer
    binary.LittleEndian.PutUint64(data[1:], uint64(amount))
    return data
}
```

## Error Handling

- If `genesis.LoadBuiltinPrograms` fails at startup, the node already logs a warning and continues. No change to this behaviour.
- If the QuanticScript System_Program bytecode fails to execute an instruction, the Runtime returns an error that propagates through `TxProcessor.ExecuteInstruction` and triggers the existing rollback path. No change needed.

## Testing Strategy

1. Delete `internal/system/system_test.go` along with the package — those tests are no longer relevant.
2. Update `internal/builtin_programs_integration_test.go` and `internal/e2e_transaction_test.go` to remove the `system` import and inline the encoding helper.
3. Ensure `setupBuiltinEnv` (or equivalent setup in e2e tests) calls `genesis.LoadBuiltinPrograms` so the QuanticScript bytecode is present before any instruction targeting the System_Program is executed.
4. Run `go test ./...` to confirm no regressions.

The existing integration tests (`TestSystemProgramCreateAccountAndTransfer`, `TestTransactionProcessingThroughRuntime`, etc.) will continue to exercise the same behaviour — the only difference is that execution now goes through the bytecode interpreter rather than the Go handler.

# Requirements Document

## Introduction

The QuanticScript programs (System_Program and Token_Program) have stub `.qs` source files and stub assembly (`.qsa`) and bytecode (`.qsb`) files. The goal is to produce production-ready bytecode by:

1. Adding a `DISPATCH` opcode to the interpreter that maps an instruction type code to a typed argument struct — all parsing logic lives in Go, not in QuanticScript source
2. Writing complete `.qs` source files that use `DISPATCH` as their entry point, then call handler functions with pre-parsed args on the stack
3. Compiling `.qs` → `.qsa` → `.qsb` using the QuanticScript compiler toolchain and embedding the bytecode at genesis

**Core constraint: the `.qs` source is the single source of truth. Assembly and bytecode are always compiler outputs — never hand-written. The compiler, stdlib, and interpreter own all complexity.**

## Glossary

- **System_Program**: Built-in program managing Neon accounts (create, transfer, allocate)
- **Token_Program**: Built-in program managing fungible tokens (mint, burn, transfer, freeze, delegate)
- **QuanticScript**: TypeScript-like smart contract language with deterministic execution
- **QuanticScript Source (.qs)**: High-level TypeScript-like source — the single source of truth for program logic
- **Assembly (.qsa)**: Human-readable bytecode representation — always a compiler output, never hand-written
- **Bytecode (.qsb)**: Compiled binary executed by the interpreter — always produced by the toolchain
- **InstructionDef**: Go struct describing an instruction's type code, name, and typed arg schema
- **DISPATCH opcode**: Single interpreter opcode that reads raw instruction bytes, looks up the registry, and pushes parsed args onto the stack
- **ArgDef**: A single argument definition: `{Name, Type, Offset, Length}`

## Requirements

### Requirement 1: Instruction Dispatch Registry

**User Story:** As a program author, I want to declare instruction schemas once in Go and have the interpreter handle all parsing, so that assembly programs contain zero manual byte manipulation.

#### Acceptance Criteria

1. THE interpreter SHALL expose a `DISPATCH` opcode that pops raw instruction bytes, looks up the registered `InstructionDef` by type code, parses all args according to their `ArgDef` schemas, and pushes the parsed args onto the stack
2. WHEN an instruction type code has no registered `InstructionDef`, THE interpreter SHALL return `ERROR_INVALID_INSTRUCTION` without modifying state
3. WHEN instruction data is shorter than required by the arg schema, THE interpreter SHALL return `ERROR_INVALID_INSTRUCTION` without modifying state
4. THE System_Program SHALL register `InstructionDef` schemas for `CREATE_ACCOUNT(0)`, `TRANSFER(1)`, and `ALLOCATE_SPACE(2)` with fully typed arg offsets
5. THE Token_Program SHALL register `InstructionDef` schemas for all 11 instruction types with fully typed arg offsets

### Requirement 2: QuanticScript Source Programs

**User Story:** As a developer reading the programs, I want each program written in `.qs` source as a thin dispatch shell, so that the logic is easy to audit and the bytecode stays small.

#### Acceptance Criteria

1. THE System_Program `.qs` source SHALL contain no manual byte-offset arithmetic — all arg extraction SHALL be performed by `DISPATCH`
2. THE Token_Program `.qs` source SHALL contain no manual byte-offset arithmetic — all arg extraction SHALL be performed by `DISPATCH`
3. WHEN `DISPATCH` is called, THE `.qs` source SHALL branch directly to the correct handler function using the instruction name returned
4. THE `.qs` source files SHALL contain only: the entry point, handler functions, blockchain stdlib calls (getFile, updateBalance, hasSigner, etc.), and control flow
5. THE `.qs` source files SHALL NOT contain helper functions for serialization, byte slicing, or little-endian parsing — those belong in the interpreter or stdlib

### Requirement 3: Production-Ready Bytecode via Compiler Toolchain

**User Story:** As a node operator, I want the embedded bytecode to correctly execute all documented operations, so that the chain behaves as specified.

#### Acceptance Criteria

1. WHEN the `.qs` source files are compiled, THE resulting `.qsa` assembly and `.qsb` bytecode SHALL be valid and loadable by the interpreter
2. WHEN the bytecode is executed with valid instruction data, THE programs SHALL produce correct results for all operations
3. WHEN invalid or malformed instruction data is provided, THE programs SHALL return appropriate error codes without corrupting state
4. WHEN the bytecode is executed multiple times with identical inputs, THE results SHALL be identical (determinism)
5. WHEN the bytecode is embedded and loaded at genesis, THE programs SHALL be available for transaction execution

### Requirement 4: Testing

**User Story:** As a developer, I want tests that verify the dispatch registry and program execution, so that regressions are caught immediately.

#### Acceptance Criteria

1. WHEN `ParseArgs` is called with valid instruction data, THE function SHALL return correctly typed values for all arg types (i64, u64, bytes, bool)
2. WHEN `Dispatch` is called with a known instruction code, THE function SHALL return the matching `InstructionDef` and parsed args
3. WHEN the System_Program bytecode is executed, THE program SHALL produce correct results for all three instruction types
4. WHEN the Token_Program bytecode is executed, THE program SHALL produce correct results for all eleven instruction types
5. WHEN error conditions occur, THE programs SHALL return the correct error codes defined in the error code table

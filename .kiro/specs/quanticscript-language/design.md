# QuanticScript Language Design

## Overview

QuanticScript is a deterministic, developer-friendly programming language designed for blockchain smart contract execution. It features TypeScript-like syntax, functional programming principles with pragmatic imperative constructs, and compiles to cost-metered bytecode. The language integrates seamlessly with the existing POH blockchain runtime, providing inline assembly support, cross-program invocation, and a rich standard library while maintaining 100% deterministic execution through an unescapable sandbox.

## Architecture

### Compilation Pipeline

```
QuanticScript Source (.qs)
         ↓
    [Lexer/Parser]
         ↓
    Abstract Syntax Tree (AST)
         ↓
   [Type Checker]
         ↓
   [Optimizer]
         ↓
   [Code Generator]
         ↓
    Bytecode (.qsb)
         ↓
    [Assembler] ←→ Assembly (.qsa)
         ↓
   [Runtime Executor]
```

### Integration with Existing Runtime

QuanticScript programs will integrate with the existing `internal/runtime` package:

- Programs are stored as executable files in the FileStore with bytecode in the `Data` field
- The Runtime's `ExecuteProgram` method will be extended to interpret QuanticScript bytecode
- ExecutionContext provides access to instruction data, signers, and file operations
- Cost metering will be implemented at the bytecode instruction level

## Components and Interfaces

### 1. Compiler Components

#### Lexer
- Tokenizes QuanticScript source code
- Handles TypeScript-like syntax including type annotations
- Supports inline assembly blocks with `__asm__` keyword

#### Parser
- Builds AST from token stream
- Validates syntax structure
- Preserves source location information for error reporting

#### Type Checker
- Enforces static type safety
- Validates function signatures and variable types
- Ensures type consistency across assembly boundaries
- Detects non-deterministic operations at compile time

#### Optimizer
- Performs constant folding and dead code elimination
- Inlines small functions where beneficial
- Optimizes common patterns without changing semantics

#### Code Generator
- Translates AST to bytecode instructions
- Assigns cost values to each instruction
- Generates assembly representation alongside bytecode

### 2. Bytecode Specification

#### Instruction Format

```
[Opcode: 1 byte][Operands: variable length]
```

#### Core Instruction Set (with costs)

**Stack Operations:**
- `PUSH <value>` (cost: 1) - Push value onto stack
- `POP` (cost: 1) - Pop value from stack
- `DUP` (cost: 1) - Duplicate top stack value
- `SWAP` (cost: 1) - Swap top two stack values

**Arithmetic Operations:**
- `ADD` (cost: 2) - Add two integers
- `SUB` (cost: 2) - Subtract two integers
- `MUL` (cost: 3) - Multiply two integers
- `DIV` (cost: 4) - Divide two integers
- `MOD` (cost: 4) - Modulo operation

**Comparison Operations:**
- `EQ` (cost: 2) - Equality comparison
- `LT` (cost: 2) - Less than comparison
- `GT` (cost: 2) - Greater than comparison
- `LTE` (cost: 2) - Less than or equal
- `GTE` (cost: 2) - Greater than or equal

**Logical Operations:**
- `AND` (cost: 2) - Logical AND
- `OR` (cost: 2) - Logical OR
- `NOT` (cost: 1) - Logical NOT

**Bitwise Operations:**
- `BAND` (cost: 2) - Bitwise AND
- `BOR` (cost: 2) - Bitwise OR
- `BXOR` (cost: 2) - Bitwise XOR
- `BNOT` (cost: 1) - Bitwise NOT
- `SHL` (cost: 2) - Shift left
- `SHR` (cost: 2) - Shift right

**Control Flow:**
- `JMP <offset>` (cost: 2) - Unconditional jump
- `JMPIF <offset>` (cost: 2) - Conditional jump if true
- `CALL <function_id>` (cost: 5) - Call function
- `RET` (cost: 2) - Return from function

**Memory Operations:**
- `LOAD <offset>` (cost: 3) - Load from local memory
- `STORE <offset>` (cost: 3) - Store to local memory
- `LOADG <key>` (cost: 10) - Load from global state
- `STOREG <key>` (cost: 15) - Store to global state

**Blockchain Operations:**
- `GETFILE <file_id>` (cost: 50) - Get file from FileStore
- `GETFILEMUT <file_id>` (cost: 50) - Get mutable file reference
- `UPDATEFILE` (cost: 100) - Update file in FileStore
- `GETBALANCE <file_id>` (cost: 30) - Get file balance
- `UPDATEBALANCE <file_id> <delta>` (cost: 80) - Update file balance
- `GETSIGNER <index>` (cost: 10) - Get transaction signer
- `HASSIGNER <pubkey>` (cost: 15) - Check if pubkey is signer
- `GETINSTRDATA` (cost: 5) - Get instruction data
- `GETPROGRAMID` (cost: 5) - Get current program ID

**Cross-Program Invocation:**
- `INVOKE <program_id>` (cost: 200) - Invoke another program
- `INVOKERET` (cost: 10) - Return from cross-program call

**Cryptographic Operations:**
- `SHA256` (cost: 50) - SHA-256 hash
- `VERIFYSIG` (cost: 100) - Verify Ed25519 signature
- `DERIVEPUBKEY` (cost: 80) - Derive public key

**Query Operations (Finalized Data):**
- `QUERYBLOCK <hash>` (cost: 100) - Query finalized block
- `QUERYTX <tx_id>` (cost: 80) - Query finalized transaction
- `QUERYINSTR <instr_ref>` (cost: 80) - Query finalized instruction

### 3. Assembly Language

#### Syntax

```assembly
; Comments start with semicolon
; Labels end with colon

entry:
    PUSH 42
    PUSH 10
    ADD
    STORE 0
    LOAD 0
    RET

transfer:
    ; Get source and destination from instruction data
    GETINSTRDATA
    ; Parse and validate
    CALL parse_transfer_data
    ; Execute transfer logic
    GETFILE source_id
    GETFILEMUT dest_id
    ; ... transfer logic
    RET
```

#### Assembler/Disassembler

- Assembler: Converts `.qsa` assembly text to `.qsb` bytecode
- Disassembler: Converts `.qsb` bytecode to `.qsa` assembly text
- Both maintain instruction cost annotations

### 4. Runtime Executor

#### Bytecode Interpreter

Extends the existing `Runtime` struct in `internal/runtime/runtime.go`:

```go
type BytecodeInterpreter struct {
    stack         []Value
    memory        []Value
    programCounter int
    computeBudget  int64
    ctx           *ExecutionContext
}

func (bi *BytecodeInterpreter) Execute(bytecode []byte, ctx *ExecutionContext) error {
    // Initialize interpreter state
    // Execute instructions until RET or budget exhausted
    // Track cost for each instruction
}
```

#### Cost Metering

- Each instruction deducts its cost from the compute budget
- Budget is set per instruction execution (from transaction fee)
- Execution halts with error if budget reaches zero
- Unused budget is not refunded (prevents timing attacks)

### 5. Standard Library

#### Module Organization

**core** - Core language utilities
- Type conversion functions
- Error handling utilities
- Assertion functions

**crypto** - Cryptographic operations
- `sha256(data: bytes): bytes` - SHA-256 hashing
- `verifySignature(pubkey: PublicKey, message: bytes, signature: bytes): bool`
- `derivePublicKey(seed: bytes): PublicKey`

**blockchain** - Blockchain-specific operations
- `getFile(fileId: FileID): File` - Get file (read-only)
- `getFileMut(fileId: FileID): MutableFile` - Get mutable file
- `updateFile(file: MutableFile): void` - Update file
- `getBalance(fileId: FileID): i64` - Get file balance
- `updateBalance(fileId: FileID, delta: i64): void` - Update balance
- `hasSigner(pubkey: PublicKey): bool` - Check if pubkey signed transaction
- `getInstructionData(): bytes` - Get current instruction data
- `getProgramId(): FileID` - Get current program ID

**query** - Query finalized blockchain data
- `queryBlock(hash: bytes): Block | null` - Query finalized block
- `queryTransaction(txId: TxID): Transaction | null` - Query finalized transaction
- `queryInstruction(ref: InstrRef): Instruction | null` - Query finalized instruction

**invoke** - Cross-program invocation
- `invoke(programId: FileID, data: bytes): bytes` - Invoke another program
- `getInvokeDepth(): u32` - Get current call depth

**collections** - Data structures
- Array operations (map, filter, reduce, sort)
- Map/dictionary operations
- Set operations

**string** - String manipulation
- `concat(a: string, b: string): string`
- `substring(s: string, start: u32, end: u32): string`
- `toBytes(s: string): bytes`
- `fromBytes(b: bytes): string`

**math** - Mathematical operations
- Basic arithmetic (already in bytecode)
- Min/max functions
- Absolute value
- Power and logarithm (deterministic implementations)

## Data Models

### Type System

#### Primitive Types
- `i8`, `i16`, `i32`, `i64` - Signed integers
- `u8`, `u16`, `u32`, `u64` - Unsigned integers
- `bool` - Boolean
- `bytes` - Byte array
- `string` - UTF-8 string

#### Blockchain Types
- `FileID` - 32-byte file identifier
- `PublicKey` - 32-byte Ed25519 public key
- `TxID` - 32-byte transaction identifier
- `File` - File structure from FileStore
- `MutableFile` - Mutable file reference

#### Composite Types
- Arrays: `T[]`
- Tuples: `[T1, T2, ...]`
- Structs: `struct { field: Type, ... }`
- Enums: `enum { Variant1, Variant2, ... }`
- Options: `T | null`

### Memory Model

- Stack-based execution for local variables
- Heap allocation for dynamic data structures
- Automatic memory management within instruction execution
- Memory is isolated per instruction (no cross-instruction state)

### Entry Function Signature

Every QuanticScript program must export an `entry` function:

```typescript
export function entry(ctx: InstructionContext): Result<void, Error> {
    // Program logic here
}
```

Where `InstructionContext` provides:
```typescript
interface InstructionContext {
    readonly instructionData: bytes;
    readonly programId: FileID;
    readonly signers: PublicKey[];
    readonly inputs: Map<string, FileAccess>;
}
```

## Error Handling

### Error Types

- `CompileError` - Errors during compilation
- `RuntimeError` - Errors during execution
- `OutOfComputeError` - Compute budget exhausted
- `AccessViolationError` - Unauthorized file access
- `InvocationError` - Cross-program invocation failed
- `DeterminismError` - Non-deterministic operation detected

### Error Propagation

- Functions return `Result<T, Error>` type
- Use `?` operator for error propagation (similar to Rust)
- Errors cause instruction to fail and transaction to rollback

## Testing Strategy

### Unit Testing

- Test individual compiler components (lexer, parser, type checker)
- Test bytecode instruction execution
- Test standard library functions
- Test cost metering accuracy

### Integration Testing

- Test full compilation pipeline
- Test program execution in runtime
- Test cross-program invocation
- Test error handling and rollback

### Determinism Testing

- Execute same program with same inputs on multiple nodes
- Verify identical results and state changes
- Test edge cases (overflow, division by zero, etc.)

### Security Testing

- Attempt sandbox escapes
- Test resource exhaustion attacks
- Verify access control enforcement
- Test malicious bytecode rejection

## Inline Assembly Support

### Syntax

```typescript
export function optimizedTransfer(amount: u64): void {
    let result: u64;
    
    __asm__ {
        // Load amount from local variable
        LOAD amount
        // Multiply by 2 (optimized)
        DUP
        ADD
        // Store result
        STORE result
    }
    
    // Continue with high-level code
    updateBalance(destinationId, result);
}
```

### Constraints

- Assembly blocks must be type-safe at boundaries
- Variables used in assembly must be declared in QuanticScript
- Assembly cannot access variables not explicitly bound
- All assembly instructions are cost-metered

## Cross-Program Invocation

### Invocation Model

```typescript
import { invoke } from "std/invoke";

export function entry(ctx: InstructionContext): Result<void, Error> {
    // Get target program ID from instruction data
    let targetProgram = parseTargetProgram(ctx.instructionData);
    
    // Prepare invocation data
    let invokeData = encodeInvokeData({ action: "transfer", amount: 100 });
    
    // Invoke target program
    let result = invoke(targetProgram, invokeData)?;
    
    // Process result
    handleInvokeResult(result);
    
    return Ok(());
}
```

### Security Constraints

- Target program must be in instruction's declared program list
- Maximum call depth of 4 to prevent stack overflow
- Each invocation has its own compute budget (deducted from parent)
- Failed invocations rollback all state changes

## Determinism Guarantees

### Prohibited Operations

The compiler rejects code that attempts:
- Random number generation (use deterministic PRNGs with seeds)
- System time access (use block timestamp from finalized blocks)
- File I/O (use FileStore only)
- Network access
- Threading or concurrency
- Floating-point operations (use fixed-point arithmetic)

### Deterministic Alternatives

- Use block hashes as entropy sources (from finalized blocks)
- Use instruction data for randomness seeds
- Use deterministic sorting algorithms
- Use fixed-point arithmetic for decimal calculations

## Developer Experience Features

### Rich Error Messages

```
Error: Type mismatch in function call
  --> src/transfer.qs:15:20
   |
15 |     let amount = transfer(source, "100");
   |                                   ^^^^^ expected u64, found string
   |
help: Convert string to integer using parseInt()
```

### IDE Support

- Language server protocol (LSP) implementation
- Syntax highlighting
- Auto-completion
- Inline documentation
- Go-to-definition
- Find references

### Debugging Tools

- Bytecode disassembler with source mapping
- Execution trace viewer
- Cost profiler
- State inspector

### Documentation

- Comprehensive language reference
- Standard library API documentation
- Example programs and tutorials
- Best practices guide
- Migration guide from other languages

## Implementation Phases

### Phase 1: Core Language (MVP)
- Basic lexer and parser
- Simple type system
- Core bytecode instructions
- Stack-based interpreter
- Basic standard library

### Phase 2: Advanced Features
- Inline assembly support
- Cross-program invocation
- Full standard library
- Optimizer

### Phase 3: Developer Tools
- Assembler/disassembler
- Debugger
- IDE integration
- Documentation site

### Phase 4: Production Readiness
- Security audits
- Performance optimization
- Comprehensive test suite
- Production deployment

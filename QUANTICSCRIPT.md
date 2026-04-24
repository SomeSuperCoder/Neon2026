# QuanticScript Compiler

QuanticScript is a TypeScript-like programming language designed for blockchain smart contract execution with deterministic guarantees and cost metering.

## Features

- **TypeScript-like Syntax**: Familiar syntax for developers
- **Deterministic Execution**: 100% deterministic across all nodes
- **Cost Metering**: Every instruction has a fixed computational cost
- **Inline Assembly**: Write optimized assembly code when needed
- **Rich Type System**: Static type checking with type inference
- **Standard Library**: Built-in functions for crypto, blockchain operations, and more

## Installation

Build the compiler from source:

```bash
go build -o poh-blockchain ./cmd/main.go
```

## Usage

The QuanticScript compiler is accessed through the `qsc` subcommand:

```bash
./poh-blockchain qsc <subcommand> [options]
```

### Compile Source to Bytecode

Compile QuanticScript source code (`.qs`) to bytecode (`.qsb`):

```bash
./poh-blockchain qsc compile -i program.qs -o program.qsb
```

With verbose output:

```bash
./poh-blockchain qsc compile -i program.qs -o program.qsb -v
```

### Assemble Assembly to Bytecode

Assemble QuanticScript assembly (`.qsa`) to bytecode (`.qsb`):

```bash
./poh-blockchain qsc assemble -i program.qsa -o program.qsb
```

### Disassemble Bytecode to Assembly

Disassemble bytecode (`.qsb`) to assembly (`.qsa`):

```bash
./poh-blockchain qsc disassemble -i program.qsb -o program.qsa
```

## Examples

### Simple QuanticScript Program

```typescript
// simple_transfer.qs
export function entry(ctx: InstructionContext): i64 {
    let amount: i64 = 1000;
    let fee: i64 = 10;
    let total: i64 = amount + fee;
    return total;
}
```

Compile it:

```bash
./poh-blockchain qsc compile -i examples/simple_transfer.qs -o simple_transfer.qsb -v
```

Output:
```
Compiling examples/simple_transfer.qs to simple_transfer.qsb...
Phase 1: Lexical analysis...
Phase 2: Parsing...
  Parsed 1 declarations
Phase 3: Type checking...
  Type checking passed
Phase 4: Code generation...
  Generated 47 bytes of bytecode
Compilation complete!
Successfully compiled to simple_transfer.qsb (55 bytes)
```

### Assembly Programming

```assembly
; assembly_example.qsa
entry:
    PUSH i64 100
    PUSH i64 50
    ADD
    STORE 0
    LOAD 0
    RET
```

Assemble it:

```bash
./poh-blockchain qsc assemble -i examples/assembly_example.qsa -o assembly_example.qsb
```

Disassemble to verify:

```bash
./poh-blockchain qsc disassemble -i assembly_example.qsb -o output.qsa
```

## Error Reporting

The compiler provides helpful error messages with suggestions:

```bash
./poh-blockchain qsc compile -i program_with_error.qs -o output.qsb -v
```

Example error output:
```
Compilation failed with 1 error(s):

Error: <input>:4:18: undefined variable 'z'

Help: Make sure the variable is declared before use.
      Use 'let' or 'const' to declare variables.
```

## Documentation

For complete language documentation, see the [docs](docs/) directory:

- **[Language Reference](docs/language-reference.md)** - Complete syntax, types, operators, and control flow
- **[Standard Library Reference](docs/stdlib-reference.md)** - All built-in functions and modules
- **[Inline Assembly Guide](docs/inline-assembly.md)** - Writing assembly for performance
- **[Bytecode Reference](docs/bytecode-reference.md)** - Low-level bytecode format and opcodes
- **[Cost Model Guide](docs/cost-model.md)** - Understanding costs and optimization
- **[Examples](examples/README.md)** - Sample programs and tutorials

## Quick Language Overview

### Type System

QuanticScript supports:
- **Integers**: `i8`, `i16`, `i32`, `i64`, `u8`, `u16`, `u32`, `u64`
- **Boolean**: `bool`
- **Bytes**: `bytes`
- **String**: `string`
- **Arrays**: `T[]`

### Basic Syntax

```typescript
// Variables and types
let x: i64 = 42;
let name: string = "Alice";

// Control flow
if (x > 0) {
    // then block
} else {
    // else block
}

// Loops
for (let i: i64 = 0; i < 10; i = i + 1) {
    // loop body
}

// Functions
function add(a: i64, b: i64): i64 {
    return a + b;
}

// Entry function (required)
export function entry(ctx: InstructionContext): i64 {
    return 0;
}
```

### Inline Assembly

```typescript
export function optimized(x: i64): i64 {
    let result: i64;
    
    __asm__ {
        LOAD x
        DUP
        ADD
        STORE result
    }
    
    return result;
}
```

### Blockchain Operations

```typescript
// Get and update balances
let balance: i64 = getBalance(accountId);
updateBalance(accountId, amount);

// Access instruction data
let data: bytes = getInstructionData();
let programId: i64 = getProgramId();
```

## Key Features

### Deterministic Execution

QuanticScript enforces 100% determinism by:
- Rejecting random number generation
- Rejecting system time access
- Rejecting file I/O and network operations
- Rejecting floating-point operations
- Using fixed-cost instructions

### Cost Metering

Every instruction has a fixed computational cost:
- Stack operations: 1 unit
- Arithmetic: 2-4 units
- Memory operations: 3 units
- Blockchain operations: 30-100 units
- Cryptographic operations: 50-100 units

See the [Cost Model Guide](docs/cost-model.md) for details.

### Standard Library

Built-in modules for common operations:
- **crypto**: Hashing, signatures, key derivation
- **blockchain**: File operations, balances, signers
- **query**: Query finalized blockchain data
- **invoke**: Cross-program invocation
- **collections**: Array operations
- **string**: String manipulation
- **math**: Mathematical operations

See the [Standard Library Reference](docs/stdlib-reference.md) for complete API documentation.

## Development Workflow

1. Write QuanticScript source code (`.qs`)
2. Compile to bytecode (`.qsb`)
3. Optionally disassemble to review generated code (`.qsa`)
4. Deploy bytecode to blockchain
5. Execute through transaction instructions

## Troubleshooting

### Common Errors

**"undefined variable"**
- Make sure variables are declared with `let` or `const` before use

**"expected ;"**
- Check for missing semicolons at the end of statements

**"type mismatch"**
- Ensure operands have compatible types
- Add explicit type conversions if needed

**"undefined function"**
- Make sure functions are declared or imported
- Check function names for typos

### Verbose Mode

Use `-v` or `--verbose` flag to see detailed compilation phases and diagnostics:

```bash
./poh-blockchain qsc compile -i program.qs -o program.qsb -v
```

## Contributing

See the main project README for contribution guidelines.

## License

See the main project LICENSE file.

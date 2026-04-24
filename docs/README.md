# QuanticScript Documentation

Complete documentation for the QuanticScript programming language.

## Getting Started

- [Quick Start Guide](../QUICKSTART.md) - Get up and running quickly
- [QuanticScript Overview](../QUANTICSCRIPT.md) - Compiler usage and basic examples
- [Examples](../examples/README.md) - Example programs

## Language Documentation

### [Language Reference](language-reference.md)
Complete reference for QuanticScript syntax, types, operators, and control flow.

**Topics covered:**
- Type system (integers, booleans, strings, bytes)
- Variables and constants
- Operators (arithmetic, comparison, logical, bitwise)
- Control flow (if, while, for)
- Functions
- Entry function requirements
- Comments and documentation

### [Standard Library Reference](stdlib-reference.md)
Documentation for all built-in functions and modules.

**Modules:**
- **Blockchain**: File operations, balance management, signers
- **Crypto**: Hashing, signature verification, key derivation
- **Collections**: Array operations (map, filter, reduce)
- **String**: String manipulation and conversion
- **Math**: Mathematical operations
- **Query**: Query finalized blockchain data
- **Invoke**: Cross-program invocation

### [Inline Assembly Guide](inline-assembly.md)
Guide to writing inline assembly for performance-critical code.

**Topics covered:**
- Assembly syntax and `__asm__` blocks
- Stack-based execution model
- Variable binding between high-level and assembly
- Complete instruction reference
- Common patterns and examples
- Best practices

### [Bytecode Reference](bytecode-reference.md)
Low-level bytecode format and instruction set reference.

**Topics covered:**
- Bytecode file format (.qsb)
- Complete opcode table
- Instruction encoding
- Execution model
- Error conditions

### [Cost Model and Optimization Guide](cost-model.md)
Understanding computational costs and optimizing programs.

**Topics covered:**
- How cost metering works
- Instruction costs by category
- Compute budget management
- Optimization strategies
- Cost analysis techniques
- Best practices for efficient code

## Quick Reference

### Common Operations

```typescript
// Variables
let x: i64 = 42;
let name: string = "Alice";

// Arithmetic
let sum: i64 = a + b;
let product: i64 = x * y;

// Comparisons
if (x > 0 && x < 100) {
    // ...
}

// Loops
for (let i: i64 = 0; i < 10; i = i + 1) {
    // ...
}

// Functions
function add(a: i64, b: i64): i64 {
    return a + b;
}

// Entry function
export function entry(ctx: InstructionContext): i64 {
    return 0;
}
```

### Blockchain Operations

```typescript
// Get balance
let balance: i64 = getBalance(accountId);

// Update balance
updateBalance(accountId, amount);

// Get instruction data
let data: bytes = getInstructionData();

// Get program ID
let programId: i64 = getProgramId();
```

### Inline Assembly

```typescript
let result: i64;
__asm__ {
    LOAD x
    LOAD y
    ADD
    STORE result
}
```

## Instruction Costs Quick Reference

| Category | Operation | Cost |
|----------|-----------|------|
| Stack | PUSH, POP, DUP, SWAP | 1 |
| Arithmetic | ADD, SUB | 2 |
| Arithmetic | MUL | 3 |
| Arithmetic | DIV, MOD | 4 |
| Comparison | EQ, LT, GT, LTE, GTE | 2 |
| Logical | AND, OR | 2 |
| Logical | NOT | 1 |
| Control Flow | JMP, JMPIF, RET | 2 |
| Control Flow | CALL | 5 |
| Memory | LOAD, STORE | 3 |
| Blockchain | GETBALANCE | 30 |
| Blockchain | UPDATEBALANCE | 80 |
| Blockchain | GETFILE | 50 |
| Blockchain | UPDATEFILE | 100 |
| Crypto | SHA256 | 50 |
| Crypto | VERIFYSIG | 100 |
| Invoke | INVOKE | 200+ |

## Development Workflow

1. **Write** QuanticScript source code (`.qs`)
2. **Compile** to bytecode (`.qsb`)
   ```bash
   ./poh-blockchain qsc compile -i program.qs -o program.qsb
   ```
3. **Review** generated assembly (optional)
   ```bash
   ./poh-blockchain qsc disassemble -i program.qsb -o program.qsa
   ```
4. **Deploy** bytecode to blockchain
5. **Execute** through transaction instructions

## Error Messages

### Common Compilation Errors

- **"undefined variable"**: Variable not declared before use
- **"type mismatch"**: Incompatible types in operation
- **"expected ;"**: Missing semicolon
- **"undefined function"**: Function not declared or imported

### Common Runtime Errors

- **OutOfComputeError**: Compute budget exhausted
- **InsufficientBalanceError**: Balance too low for operation
- **AccessViolationError**: Unauthorized file access
- **DivisionByZeroError**: Division or modulo by zero

## Best Practices

1. **Check errors early**: Validate inputs at the start of functions
2. **Use descriptive names**: Make code self-documenting
3. **Avoid magic numbers**: Use named constants
4. **Minimize expensive operations**: Cache results when possible
5. **Document costs**: Add comments showing expected costs
6. **Test edge cases**: Verify behavior with minimal compute budgets
7. **Use inline assembly sparingly**: Only for proven bottlenecks

## Language Limitations

QuanticScript enforces determinism by prohibiting:

- ❌ Random number generation
- ❌ System time access (use block timestamps)
- ❌ File I/O (except FileStore)
- ❌ Network access
- ❌ Floating-point operations (use fixed-point)
- ❌ Threading or concurrency

## Additional Resources

- [Main README](../README.md) - Project overview
- [Examples Directory](../examples/) - Sample programs
- [CLI Usage Guide](../CLI-USAGE.md) - Command-line interface

## Contributing

See the main project README for contribution guidelines.

## Support

For questions and issues:
- Review the documentation
- Check the examples
- Examine compiler error messages (they include helpful suggestions)
- Use verbose mode (`-v`) for detailed diagnostics

## Version

This documentation is for QuanticScript version 1.0.

Last updated: 2026-04-24

# QuanticScript Examples

This directory contains example QuanticScript programs demonstrating various language features and use cases.

## Examples Overview

### Currently Compilable Examples

These examples use only the currently implemented language features and will compile successfully:

#### 1. simple_transfer.qs
**Difficulty**: Beginner  
**Features**: Basic syntax, variables, arithmetic

A minimal QuanticScript program showing basic variable declarations and arithmetic operations.

```bash
./poh-blockchain qsc compile -i examples/simple_transfer.qs -o simple_transfer.qsb -v
```

#### 2. arithmetic_demo.qs
**Difficulty**: Beginner  
**Features**: Arithmetic operators, variables, expressions

Demonstrates all arithmetic operations and complex expressions.

```bash
./poh-blockchain qsc compile -i examples/arithmetic_demo.qs -o arithmetic_demo.qsb -v
```

#### 3. control_flow_demo.qs
**Difficulty**: Intermediate  
**Features**: Functions, if/else, while loops, for loops

Shows control flow constructs and function definitions.

```bash
./poh-blockchain qsc compile -i examples/control_flow_demo.qs -o control_flow_demo.qsb -v
```

#### 4. optimized_math.qs
**Difficulty**: Advanced  
**Features**: Inline assembly, stack manipulation, bit operations

Demonstrates using inline assembly (`__asm__`) for performance-critical operations. Shows stack manipulation, bit shifting, and conditional logic at the assembly level.

```bash
./poh-blockchain qsc compile -i examples/optimized_math.qs -o optimized_math.qsb -v
```

### Standard Library API Examples

These examples demonstrate the intended standard library API. Some functions may not be fully implemented yet, but they show how blockchain programs should be written:

#### 5. token_transfer.qs
**Difficulty**: Intermediate  
**Features**: Blockchain operations, balance management, conditionals

Demonstrates a basic token transfer program using blockchain-specific operations like `getBalance()` and `updateBalance()`.

**Note**: Requires full standard library implementation.

#### 6. token_with_invoke.qs
**Difficulty**: Advanced  
**Features**: Cross-program invocation, fee calculation, program ID access

Shows how to implement a token program that can invoke other programs for validation or additional logic. Includes fee calculation and distribution.

**Note**: Requires full standard library implementation.

#### 7. advanced_token.qs
**Difficulty**: Advanced  
**Features**: Multiple operations, error handling, helper functions, validation

A comprehensive token program supporting transfer, mint, and burn operations with proper error handling and validation.

**Note**: Requires full standard library implementation.

## Assembly Examples

### assembly_example.qsa
A simple assembly program demonstrating basic stack operations and arithmetic.

```bash
./poh-blockchain qsc assemble -i examples/assembly_example.qsa -o assembly_example.qsb
./poh-blockchain qsc disassemble -i assembly_example.qsb -o output.qsa
```

## Learning Path

1. Start with `simple_transfer.qs` to understand basic syntax
2. Move to `arithmetic_demo.qs` to see all arithmetic operations
3. Study `control_flow_demo.qs` for functions and control flow
4. Explore `optimized_math.qs` for inline assembly
5. Review the standard library examples (`token_transfer.qs`, `token_with_invoke.qs`, `advanced_token.qs`) to understand the intended blockchain API

## Common Patterns

### Error Handling
```typescript
if (condition) {
    return -1;  // Error code
}
return 0;  // Success
```

### Balance Checks
```typescript
let balance: i64 = getBalance(accountId);
if (balance < amount) {
    return -2;  // Insufficient balance
}
```

### Overflow Protection
```typescript
let maxBalance: i64 = 9223372036854775807;
if (currentBalance > maxBalance - amount) {
    return -3;  // Overflow error
}
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

## Compiling Examples

To compile the currently working examples:

```bash
# Compile all working examples
for file in examples/simple_transfer.qs examples/arithmetic_demo.qs examples/control_flow_demo.qs examples/optimized_math.qs; do
    output="${file%.qs}.qsb"
    echo "Compiling $file..."
    ./poh-blockchain qsc compile -i "$file" -o "$output" -v
done
```

Note: The standard library examples (token_transfer.qs, token_with_invoke.qs, advanced_token.qs) demonstrate the intended API but may not compile until the standard library is fully implemented.

## Next Steps

- Read the [QuanticScript Language Reference](../docs/language-reference.md)
- Review the [Standard Library Documentation](../docs/stdlib-reference.md)
- Check the [Cost Model Guide](../docs/cost-model.md)
- Learn about [Inline Assembly](../docs/inline-assembly.md)

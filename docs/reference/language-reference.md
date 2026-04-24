# QuanticScript Language Reference

## Table of Contents

1. [Introduction](#introduction)
2. [Syntax Overview](#syntax-overview)
3. [Type System](#type-system)
4. [Variables and Constants](#variables-and-constants)
5. [Operators](#operators)
6. [Control Flow](#control-flow)
7. [Functions](#functions)
8. [Inline Assembly](#inline-assembly)
9. [Entry Function](#entry-function)
10. [Comments](#comments)

## Introduction

QuanticScript is a statically-typed, deterministic programming language designed for blockchain smart contract execution. It features TypeScript-like syntax with functional programming principles and compiles to cost-metered bytecode.

### Design Principles

- **Deterministic**: Identical inputs always produce identical outputs
- **Type-Safe**: Static type checking prevents runtime type errors
- **Cost-Metered**: Every operation has a fixed computational cost
- **Sandboxed**: Programs cannot access unauthorized resources
- **Developer-Friendly**: Familiar syntax for TypeScript/JavaScript developers

## Syntax Overview

QuanticScript uses TypeScript-like syntax with some restrictions to ensure determinism:

```typescript
// Variable declaration
let x: i64 = 42;

// Function definition
function add(a: i64, b: i64): i64 {
    return a + b;
}

// Entry function (required)
export function entry(ctx: InstructionContext): i64 {
    let result: i64 = add(10, 20);
    return result;
}
```

## Type System

### Primitive Types

#### Integer Types

| Type | Size | Range | Description |
|------|------|-------|-------------|
| `i8` | 8-bit | -128 to 127 | Signed 8-bit integer |
| `i16` | 16-bit | -32,768 to 32,767 | Signed 16-bit integer |
| `i32` | 32-bit | -2,147,483,648 to 2,147,483,647 | Signed 32-bit integer |
| `i64` | 64-bit | -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807 | Signed 64-bit integer |
| `u8` | 8-bit | 0 to 255 | Unsigned 8-bit integer |
| `u16` | 16-bit | 0 to 65,535 | Unsigned 16-bit integer |
| `u32` | 32-bit | 0 to 4,294,967,295 | Unsigned 32-bit integer |
| `u64` | 64-bit | 0 to 18,446,744,073,709,551,615 | Unsigned 64-bit integer |

#### Other Primitive Types

- `bool`: Boolean type (`true` or `false`)
- `bytes`: Byte array for binary data
- `string`: UTF-8 encoded string

### Blockchain Types

- `FileID`: 32-byte file identifier (represented as `i64` in current implementation)
- `PublicKey`: 32-byte Ed25519 public key
- `InstructionContext`: Context passed to entry function

### Type Annotations

Type annotations are required for variable declarations and function parameters:

```typescript
let age: i64 = 25;
let name: string = "Alice";
let isActive: bool = true;
```

### Type Inference

The compiler can infer types in some contexts:

```typescript
let x: i64 = 10;
let y = x + 5;  // y is inferred as i64
```

## Variables and Constants

### Variable Declaration

Variables are declared using the `let` keyword:

```typescript
let x: i64 = 42;
let name: string = "Bob";
let isValid: bool = false;
```

### Variable Assignment

Variables can be reassigned after declaration:

```typescript
let counter: i64 = 0;
counter = counter + 1;
counter = 10;
```

### Constants

Constants are declared using the `const` keyword (currently treated as immutable variables):

```typescript
const MAX_SUPPLY: i64 = 1000000;
const PROGRAM_VERSION: i64 = 1;
```

### Scope

Variables are block-scoped:

```typescript
let x: i64 = 10;

if (x > 5) {
    let y: i64 = 20;  // y is only visible in this block
    x = x + y;
}

// y is not accessible here
```

## Operators

### Arithmetic Operators

| Operator | Description | Example | Cost |
|----------|-------------|---------|------|
| `+` | Addition | `a + b` | 2 |
| `-` | Subtraction | `a - b` | 2 |
| `*` | Multiplication | `a * b` | 3 |
| `/` | Division | `a / b` | 4 |
| `%` | Modulo | `a % b` | 4 |

```typescript
let sum: i64 = 10 + 20;        // 30
let diff: i64 = 50 - 15;       // 35
let product: i64 = 6 * 7;      // 42
let quotient: i64 = 100 / 4;   // 25
let remainder: i64 = 17 % 5;   // 2
```

### Comparison Operators

| Operator | Description | Example | Cost |
|----------|-------------|---------|------|
| `==` | Equal to | `a == b` | 2 |
| `!=` | Not equal to | `a != b` | 2 |
| `<` | Less than | `a < b` | 2 |
| `>` | Greater than | `a > b` | 2 |
| `<=` | Less than or equal | `a <= b` | 2 |
| `>=` | Greater than or equal | `a >= b` | 2 |

```typescript
let isEqual: bool = (10 == 10);      // true
let isGreater: bool = (20 > 15);     // true
let isLessOrEqual: bool = (5 <= 5);  // true
```

### Logical Operators

| Operator | Description | Example | Cost |
|----------|-------------|---------|------|
| `&&` | Logical AND | `a && b` | 2 |
| `\|\|` | Logical OR | `a \|\| b` | 2 |
| `!` | Logical NOT | `!a` | 1 |

```typescript
let result: bool = (x > 0) && (x < 100);
let isValid: bool = (status == 1) || (status == 2);
let isInvalid: bool = !isValid;
```

### Bitwise Operators

| Operator | Description | Example | Cost |
|----------|-------------|---------|------|
| `&` | Bitwise AND | `a & b` | 2 |
| `\|` | Bitwise OR | `a \| b` | 2 |
| `^` | Bitwise XOR | `a ^ b` | 2 |
| `~` | Bitwise NOT | `~a` | 1 |
| `<<` | Left shift | `a << b` | 2 |
| `>>` | Right shift | `a >> b` | 2 |

```typescript
let masked: i64 = value & 0xFF;
let flags: i64 = flag1 | flag2;
let shifted: i64 = x << 2;  // Multiply by 4
```

### Operator Precedence

From highest to lowest:

1. Unary operators: `!`, `~`, `-` (negation)
2. Multiplicative: `*`, `/`, `%`
3. Additive: `+`, `-`
4. Shift: `<<`, `>>`
5. Relational: `<`, `>`, `<=`, `>=`
6. Equality: `==`, `!=`
7. Bitwise AND: `&`
8. Bitwise XOR: `^`
9. Bitwise OR: `|`
10. Logical AND: `&&`
11. Logical OR: `||`

Use parentheses to override precedence:

```typescript
let result: i64 = (a + b) * c;
```

## Control Flow

### If Statements

```typescript
if (condition) {
    // executed if condition is true
}

if (x > 0) {
    // positive
} else {
    // zero or negative
}

if (x > 0) {
    // positive
} else if (x < 0) {
    // negative
} else {
    // zero
}
```

### While Loops

```typescript
let i: i64 = 0;
while (i < 10) {
    // loop body
    i = i + 1;
}
```

### For Loops

```typescript
for (let i: i64 = 0; i < 10; i = i + 1) {
    // loop body
}
```

### Early Return

Use `return` to exit a function early:

```typescript
function validate(x: i64): i64 {
    if (x < 0) {
        return -1;  // Error: negative value
    }
    
    if (x > 1000) {
        return -2;  // Error: value too large
    }
    
    return 0;  // Success
}
```

## Functions

### Function Declaration

```typescript
function functionName(param1: Type1, param2: Type2): ReturnType {
    // function body
    return value;
}
```

### Example Functions

```typescript
function add(a: i64, b: i64): i64 {
    return a + b;
}

function isEven(n: i64): bool {
    return (n % 2) == 0;
}

function max(a: i64, b: i64): i64 {
    if (a > b) {
        return a;
    } else {
        return b;
    }
}
```

### Function Calls

```typescript
let sum: i64 = add(10, 20);
let result: bool = isEven(42);
let maximum: i64 = max(100, 200);
```

### Void Functions

Functions that don't return a value use `void` as return type:

```typescript
function logValue(x: i64): void {
    // In a real implementation, this might write to a log
    // For now, it just performs operations without returning
}
```

## Inline Assembly

### Assembly Blocks

Use `__asm__` keyword to write inline assembly:

```typescript
let result: i64;
__asm__ {
    LOAD x
    LOAD y
    ADD
    STORE result
}
```

### Variable Binding

Variables used in assembly must be declared in QuanticScript:

```typescript
function double(x: i64): i64 {
    let result: i64;
    
    __asm__ {
        LOAD x      // Load variable from QuanticScript scope
        DUP         // Duplicate top of stack
        ADD         // Add: x + x
        STORE result // Store to QuanticScript variable
    }
    
    return result;
}
```

### Assembly Instructions

See [Bytecode Reference](bytecode-reference.md) for complete instruction set.

Common instructions:
- `PUSH <type> <value>`: Push value onto stack
- `POP`: Remove top of stack
- `DUP`: Duplicate top of stack
- `SWAP`: Swap top two stack values
- `LOAD <var>`: Load variable onto stack
- `STORE <var>`: Store top of stack to variable
- `ADD`, `SUB`, `MUL`, `DIV`: Arithmetic operations
- `JMP <label>`, `JMPIF <label>`: Control flow
- `RET`: Return from function

### Type Safety

Assembly blocks must maintain type safety at boundaries:

```typescript
let x: i64 = 10;
let result: i64;  // Must match the type of value stored

__asm__ {
    LOAD x
    PUSH i64 5
    ADD
    STORE result  // Type must match
}
```

## Entry Function

Every QuanticScript program must have an `entry` function:

```typescript
export function entry(ctx: InstructionContext): i64 {
    // Program logic here
    return 0;  // Return code (0 = success)
}
```

### InstructionContext

The `InstructionContext` provides access to:
- Instruction data
- Program ID
- Transaction signers
- File access permissions

### Return Codes

By convention:
- `0`: Success
- Negative values: Error codes
- Positive values: Application-specific results

```typescript
export function entry(ctx: InstructionContext): i64 {
    if (errorCondition) {
        return -1;  // Error
    }
    return 0;  // Success
}
```

## Comments

### Single-Line Comments

```typescript
// This is a single-line comment
let x: i64 = 42;  // Comment after code
```

### Multi-Line Comments

```typescript
/*
 * This is a multi-line comment
 * It can span multiple lines
 */
let y: i64 = 100;
```

### Documentation Comments

Use comments to document functions:

```typescript
// Calculate the sum of two numbers
// Parameters:
//   a: First number
//   b: Second number
// Returns: Sum of a and b
function add(a: i64, b: i64): i64 {
    return a + b;
}
```

## Best Practices

### 1. Use Descriptive Names

```typescript
// Good
let accountBalance: i64 = 1000;
let isAuthorized: bool = true;

// Avoid
let x: i64 = 1000;
let f: bool = true;
```

### 2. Check for Errors Early

```typescript
function transfer(amount: i64): i64 {
    if (amount <= 0) {
        return -1;  // Invalid amount
    }
    
    if (balance < amount) {
        return -2;  // Insufficient balance
    }
    
    // Proceed with transfer
    return 0;
}
```

### 3. Avoid Magic Numbers

```typescript
// Good
const MAX_TRANSFER: i64 = 10000;
if (amount > MAX_TRANSFER) {
    return -1;
}

// Avoid
if (amount > 10000) {
    return -1;
}
```

### 4. Use Inline Assembly Sparingly

Only use inline assembly for performance-critical code:

```typescript
// Use high-level code for clarity
let result: i64 = x * 2;

// Use assembly only when necessary
let optimized: i64;
__asm__ {
    LOAD x
    DUP
    ADD
    STORE optimized
}
```

### 5. Handle Overflow

```typescript
const MAX_I64: i64 = 9223372036854775807;

function safeAdd(a: i64, b: i64): i64 {
    if (a > MAX_I64 - b) {
        return -1;  // Overflow would occur
    }
    return a + b;
}
```

## Limitations

### No Floating-Point

QuanticScript does not support floating-point operations to ensure determinism. Use fixed-point arithmetic instead:

```typescript
// Represent 12.34 as 1234 with 2 decimal places
let price: i64 = 1234;
let quantity: i64 = 5;
let total: i64 = (price * quantity) / 100;  // 61.70
```

### No Random Numbers

Random number generation is not allowed. Use deterministic sources:

```typescript
// Use block hash or instruction data as entropy source
let seed: i64 = getInstructionData()[0];
```

### No System Calls

Programs cannot access system resources:
- No file I/O (except FileStore)
- No network access
- No system time (use block timestamps)
- No threading

## Next Steps

- [Standard Library Reference](stdlib-reference.md)
- [Bytecode Reference](bytecode-reference.md)
- [Cost Model Guide](cost-model.md)
- [Examples](../examples/README.md)

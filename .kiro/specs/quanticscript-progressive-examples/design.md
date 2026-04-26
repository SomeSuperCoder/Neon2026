# Design Document

## Overview

This design outlines a progressive series of QuanticScript example programs that will be created, compiled, and verified to refine the language and compiler. Each example builds on previous ones, increasing in complexity while avoiding inline assembly to keep the language accessible. The examples will test the complete compiler pipeline: lexer → parser → type checker → code generator → bytecode.

## Architecture

### Example Progression Strategy

The examples follow a carefully designed progression:

1. **Level 1: Literals and Basic Operations** - Test fundamental language features
2. **Level 2: Variables and Expressions** - Test variable declarations and complex expressions
3. **Level 3: Control Flow** - Test conditionals and loops
4. **Level 4: Functions** - Test function declarations, calls, and returns
5. **Level 5: Blockchain Operations** - Test standard library blockchain functions
6. **Level 6: Complex Programs** - Test multiple features working together

### Compilation Verification

Each example will be:
1. Written as a `.qs` source file
2. Compiled using `go run cmd/main.go qsc compile -i <input>.qs -o <output>.qsb`
3. Verified by checking the output bytecode file exists and is non-empty
4. Documented with the features it tests

## Components and Interfaces

### Example Files Structure

```
examples/
├── 01_literals.qs           # Integer literals, basic arithmetic
├── 02_variables.qs          # Variable declarations, assignments
├── 03_expressions.qs        # Complex expressions, operator precedence
├── 04_conditionals.qs       # If-else statements
├── 05_while_loop.qs         # While loops
├── 06_for_loop.qs           # For loops
├── 07_functions.qs          # Function declarations and calls
├── 08_recursion.qs          # Recursive functions
├── 09_balance_ops.qs        # Balance queries and updates
├── 10_file_ops.qs           # File operations
├── 11_crypto_ops.qs         # Cryptographic operations
├── 12_system_program.qs     # Complete system program (FIRST)
└── 13_token_program.qs      # Complete token program (SECOND)
```

### Compilation Command Pattern

```bash
go run cmd/main.go qsc compile -i examples/<name>.qs -o examples/<name>.qsb
```

## Data Models

### Example Metadata

Each example will be documented with:
- **Name**: Descriptive name
- **Level**: Complexity level (1-6)
- **Features Tested**: List of language features
- **Dependencies**: Previous examples that must work first
- **Expected Outcome**: What should happen when compiled

### Example Template

```typescript
// Example: <name>
// Level: <level>
// Features: <comma-separated list>
// Description: <what this example demonstrates>

export function entry(ctx: InstructionContext): i64 {
    // Example code here
    return 0;
}
```

## Error Handling

### Compilation Failure Strategy

When an example fails to compile:
1. Stop the progression
2. Report the compilation error with full details
3. Identify which language feature is failing
4. Allow manual review and compiler fixes
5. Resume from the failed example after fixes

### Error Categories

- **Lexer Errors**: Invalid tokens, unexpected characters
- **Parser Errors**: Syntax errors, unexpected tokens
- **Type Checker Errors**: Type mismatches, undefined variables
- **Code Generator Errors**: Unsupported constructs, bytecode generation failures

## Testing Strategy

### Compilation Testing

For each example:
1. Verify the source file is valid QuanticScript
2. Run the compiler
3. Check for compilation errors
4. Verify bytecode output exists
5. Verify bytecode is non-empty (> 0 bytes)

### Feature Coverage

Track which language features are tested:
- ✓ Integer literals (i64)
- ✓ Arithmetic operators (+, -, *, /, %)
- ✓ Variable declarations (let, const)
- ✓ Assignments (=, +=, -=, *=, /=)
- ✓ Comparison operators (==, !=, <, >, <=, >=)
- ✓ Logical operators (&&, ||, !)
- ✓ If-else statements
- ✓ While loops
- ✓ For loops
- ✓ Function declarations
- ✓ Function calls
- ✓ Return statements
- ✓ Standard library functions

### Progressive Validation

Each example must compile successfully before moving to the next. This ensures:
- Incremental validation of compiler features
- Early detection of issues
- Clear identification of which feature is broken

## Example Specifications

### Level 1: Literals and Basic Operations

**01_literals.qs**
```typescript
// Tests: integer literals, basic arithmetic, return statement
export function entry(ctx: InstructionContext): i64 {
    return 42;
}
```

**02_variables.qs**
```typescript
// Tests: variable declarations, assignments, arithmetic
export function entry(ctx: InstructionContext): i64 {
    let x: i64 = 10;
    let y: i64 = 20;
    let sum: i64 = x + y;
    return sum;
}
```

**03_expressions.qs**
```typescript
// Tests: complex expressions, operator precedence, multiple operations
export function entry(ctx: InstructionContext): i64 {
    let a: i64 = 5;
    let b: i64 = 3;
    let c: i64 = 2;
    let result: i64 = (a + b) * c - a / b;
    return result;
}
```

### Level 2: Control Flow

**04_conditionals.qs**
```typescript
// Tests: if-else statements, comparison operators, boolean logic
export function entry(ctx: InstructionContext): i64 {
    let x: i64 = 10;
    if (x > 5) {
        return 1;
    } else {
        return 0;
    }
}
```

**05_while_loop.qs**
```typescript
// Tests: while loops, loop conditions, increment operators
export function entry(ctx: InstructionContext): i64 {
    let sum: i64 = 0;
    let i: i64 = 1;
    while (i <= 10) {
        sum = sum + i;
        i = i + 1;
    }
    return sum;
}
```

**06_for_loop.qs**
```typescript
// Tests: for loops, loop initialization, conditions, updates
export function entry(ctx: InstructionContext): i64 {
    let sum: i64 = 0;
    for (let i: i64 = 1; i <= 10; i = i + 1) {
        sum = sum + i;
    }
    return sum;
}
```

### Level 3: Functions

**07_functions.qs**
```typescript
// Tests: function declarations, function calls, parameters, return values
function add(a: i64, b: i64): i64 {
    return a + b;
}

export function entry(ctx: InstructionContext): i64 {
    let result: i64 = add(10, 20);
    return result;
}
```

**08_recursion.qs**
```typescript
// Tests: recursive function calls, base cases
function factorial(n: i64): i64 {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

export function entry(ctx: InstructionContext): i64 {
    return factorial(5);
}
```

### Level 4: Blockchain Operations

**09_balance_ops.qs**
```typescript
// Tests: getBalance, updateBalance standard library functions
export function entry(ctx: InstructionContext): i64 {
    let currentBalance: i64 = getBalance(0);
    updateBalance(0, currentBalance + 100);
    return getBalance(0);
}
```

**10_file_ops.qs**
```typescript
// Tests: getFile, updateFile, file operations
export function entry(ctx: InstructionContext): i64 {
    let fileId: FileID = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1];
    let data: bytes = getFile(fileId);
    return len(data);
}
```

**11_crypto_ops.qs**
```typescript
// Tests: sha256, cryptographic operations
export function entry(ctx: InstructionContext): i64 {
    let data: bytes = [1, 2, 3, 4, 5];
    let hash: bytes = sha256(data);
    return len(hash);
}
```

### Level 5: Complex Programs

**12_system_program.qs**
```typescript
// Tests: complete system program with account management (FIRST COMPLEX PROGRAM)
function createAccount(data: bytes): i64 {
    // Create new account logic
    updateBalance(0, 1000000);
    return 0;
}

function closeAccount(data: bytes): i64 {
    // Close account logic
    let balance: i64 = getBalance(0);
    updateBalance(0, 0);
    return balance;
}

export function entry(ctx: InstructionContext): i64 {
    let instructionData: bytes = getInstructionData();
    if (len(instructionData) == 0) {
        return 1;
    }
    
    let instruction: i64 = instructionData[0];
    if (instruction == 0) {
        return createAccount(instructionData);
    } else if (instruction == 1) {
        return closeAccount(instructionData);
    }
    
    return 1;
}
```

**13_token_program.qs**
```typescript
// Tests: multiple functions, dispatch pattern, instruction parsing (SECOND COMPLEX PROGRAM)
function transfer(data: bytes): i64 {
    let amount: i64 = 1000;
    updateBalance(0, getBalance(0) - amount);
    updateBalance(1, getBalance(1) + amount);
    return 0;
}

function mint(data: bytes): i64 {
    let amount: i64 = 1000;
    updateBalance(0, getBalance(0) + amount);
    return 0;
}

export function entry(ctx: InstructionContext): i64 {
    let instructionData: bytes = getInstructionData();
    if (len(instructionData) == 0) {
        return 1; // Error: no instruction data
    }
    
    let instruction: i64 = instructionData[0];
    if (instruction == 0) {
        return transfer(instructionData);
    } else if (instruction == 1) {
        return mint(instructionData);
    }
    
    return 1; // Error: unknown instruction
}
```

## Implementation Notes

### Compiler Requirements

The examples assume the following compiler features are working:
- Lexer: All token types (keywords, operators, literals, identifiers)
- Parser: All statement and expression types
- Type Checker: Type inference, type validation, scope management
- Code Generator: All opcodes, function calls, control flow

### Standard Library Requirements

The examples use these standard library functions:
- `getBalance(index: i64): i64`
- `updateBalance(index: i64, amount: i64): void`
- `getFile(fileId: FileID): bytes`
- `updateFile(fileId: FileID, data: bytes): void`
- `getInstructionData(): bytes`
- `sha256(data: bytes): bytes`
- `len(data: bytes): i64`

### No Inline Assembly

All examples use pure QuanticScript without `__asm__` blocks to ensure the language is accessible and the compiler handles all code generation.

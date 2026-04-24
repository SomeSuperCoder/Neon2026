# Inline Assembly Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Syntax](#basic-syntax)
3. [Stack Model](#stack-model)
4. [Variable Binding](#variable-binding)
5. [Instruction Reference](#instruction-reference)
6. [Common Patterns](#common-patterns)
7. [Best Practices](#best-practices)
8. [Examples](#examples)

## Introduction

Inline assembly allows you to write low-level bytecode instructions directly in your QuanticScript programs. This is useful for:

- Performance-critical code sections
- Fine-grained control over execution
- Optimizing specific operations
- Learning how the compiler generates code

### When to Use Inline Assembly

**Use inline assembly when:**
- You need maximum performance for hot code paths
- You're implementing low-level algorithms
- You want precise control over instruction costs
- Standard operations are insufficient

**Avoid inline assembly when:**
- High-level code is clear and sufficient
- The performance gain is negligible
- It makes code harder to maintain
- You're not familiar with the bytecode model

## Basic Syntax

Inline assembly blocks use the `__asm__` keyword:

```typescript
__asm__ {
    // Assembly instructions here
    PUSH i64 42
    PUSH i64 10
    ADD
}
```

### Integrating with QuanticScript

```typescript
export function entry(ctx: InstructionContext): i64 {
    let x: i64 = 10;
    let result: i64;
    
    __asm__ {
        LOAD x
        PUSH i64 5
        ADD
        STORE result
    }
    
    return result;  // Returns 15
}
```

## Stack Model

QuanticScript bytecode uses a stack-based execution model.

### Stack Operations

The stack grows upward. The "top" is the most recently pushed value:

```
Stack visualization:
[bottom] ... [value3] [value2] [value1] [top]
```

### Example Stack Execution

```typescript
__asm__ {
    PUSH i64 10    // Stack: [10]
    PUSH i64 20    // Stack: [10, 20]
    ADD            // Stack: [30]
    PUSH i64 5     // Stack: [30, 5]
    MUL            // Stack: [150]
}
```

### Stack Manipulation

```typescript
__asm__ {
    PUSH i64 1     // Stack: [1]
    PUSH i64 2     // Stack: [1, 2]
    PUSH i64 3     // Stack: [1, 2, 3]
    
    DUP            // Stack: [1, 2, 3, 3]  (duplicate top)
    SWAP           // Stack: [1, 2, 3, 3]  (swap top two)
    POP            // Stack: [1, 2, 3]     (remove top)
}
```

## Variable Binding

### Loading Variables

Use `LOAD` to push a QuanticScript variable onto the stack:

```typescript
let x: i64 = 42;
let y: i64 = 10;

__asm__ {
    LOAD x    // Push x onto stack
    LOAD y    // Push y onto stack
    ADD       // Add them
}
```

### Storing Variables

Use `STORE` to pop the stack top into a variable:

```typescript
let result: i64;

__asm__ {
    PUSH i64 100
    PUSH i64 50
    ADD
    STORE result  // result = 150
}
```

### Type Safety

Variables maintain their types across assembly boundaries:

```typescript
let x: i64 = 10;
let y: bool;

__asm__ {
    LOAD x
    PUSH i64 0
    GT
    STORE y  // y is bool, stores comparison result
}
```

## Instruction Reference

### Stack Operations

#### PUSH
Push a value onto the stack.

```typescript
__asm__ {
    PUSH i64 42
    PUSH i64 100
    PUSH bool true
}
```

**Cost:** 1

---

#### POP
Remove the top value from the stack.

```typescript
__asm__ {
    PUSH i64 10
    PUSH i64 20
    POP          // Remove 20, stack now has [10]
}
```

**Cost:** 1

---

#### DUP
Duplicate the top stack value.

```typescript
__asm__ {
    PUSH i64 42
    DUP          // Stack: [42, 42]
}
```

**Cost:** 1

---

#### SWAP
Swap the top two stack values.

```typescript
__asm__ {
    PUSH i64 10
    PUSH i64 20
    SWAP         // Stack: [20, 10]
}
```

**Cost:** 1

### Arithmetic Operations

#### ADD
Add two integers.

```typescript
__asm__ {
    PUSH i64 10
    PUSH i64 20
    ADD          // Stack: [30]
}
```

**Cost:** 2

---

#### SUB
Subtract two integers (top from second).

```typescript
__asm__ {
    PUSH i64 50
    PUSH i64 20
    SUB          // Stack: [30] (50 - 20)
}
```

**Cost:** 2

---

#### MUL
Multiply two integers.

```typescript
__asm__ {
    PUSH i64 6
    PUSH i64 7
    MUL          // Stack: [42]
}
```

**Cost:** 3

---

#### DIV
Divide two integers (second by top).

```typescript
__asm__ {
    PUSH i64 100
    PUSH i64 4
    DIV          // Stack: [25] (100 / 4)
}
```

**Cost:** 4

---

#### MOD
Modulo operation.

```typescript
__asm__ {
    PUSH i64 17
    PUSH i64 5
    MOD          // Stack: [2] (17 % 5)
}
```

**Cost:** 4

### Comparison Operations

#### EQ, LT, GT, LTE, GTE
Comparison operations.

```typescript
__asm__ {
    PUSH i64 10
    PUSH i64 20
    LT           // Stack: [true] (10 < 20)
    
    PUSH i64 30
    PUSH i64 30
    EQ           // Stack: [true, true] (30 == 30)
}
```

**Cost:** 2 each

### Logical Operations

#### AND, OR, NOT
Logical operations on booleans.

```typescript
__asm__ {
    PUSH bool true
    PUSH bool false
    AND          // Stack: [false]
    
    PUSH bool true
    OR           // Stack: [true]
    
    NOT          // Stack: [false]
}
```

**Cost:** 1-2

### Bitwise Operations

#### BAND, BOR, BXOR, BNOT
Bitwise operations on integers.

```typescript
__asm__ {
    PUSH i64 0xFF
    PUSH i64 0x0F
    BAND         // Stack: [0x0F]
}
```

**Cost:** 1-2

---

#### SHL, SHR
Bit shift operations.

```typescript
__asm__ {
    PUSH i64 1
    PUSH i64 3
    SHL          // Stack: [8] (1 << 3)
    
    PUSH i64 2
    SHR          // Stack: [2] (8 >> 2)
}
```

**Cost:** 2

### Control Flow

#### JMP
Unconditional jump to label.

```typescript
__asm__ {
    JMP target
    PUSH i64 1   // Skipped
    
    target:
    PUSH i64 2   // Executed
}
```

**Cost:** 2

---

#### JMPIF
Conditional jump if top of stack is true.

```typescript
__asm__ {
    PUSH i64 10
    PUSH i64 20
    LT           // Stack: [true]
    JMPIF is_less
    
    PUSH i64 0   // Skipped
    JMP end
    
    is_less:
    PUSH i64 1   // Executed
    
    end:
}
```

**Cost:** 2

---

#### RET
Return from function.

```typescript
__asm__ {
    PUSH i64 42
    RET          // Return with 42 on stack
}
```

**Cost:** 2

### Memory Operations

#### LOAD
Load variable onto stack.

```typescript
let x: i64 = 100;

__asm__ {
    LOAD x       // Stack: [100]
}
```

**Cost:** 3

---

#### STORE
Store top of stack to variable.

```typescript
let result: i64;

__asm__ {
    PUSH i64 42
    STORE result // result = 42
}
```

**Cost:** 3

## Common Patterns

### Pattern 1: Double a Value

```typescript
function double(x: i64): i64 {
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

### Pattern 2: Swap Two Variables

```typescript
function swap(a: i64, b: i64): void {
    let temp: i64;
    
    __asm__ {
        LOAD a
        LOAD b
        STORE a    // a = b
        STORE temp // temp = original a
        LOAD temp
        STORE b    // b = original a
    }
}
```

### Pattern 3: Conditional Assignment

```typescript
function max(a: i64, b: i64): i64 {
    let result: i64;
    
    __asm__ {
        LOAD a
        LOAD b
        DUP
        LOAD a
        GT
        JMPIF use_b
        
        // Use a
        POP
        STORE result
        JMP end
        
        use_b:
        SWAP
        POP
        STORE result
        
        end:
    }
    
    return result;
}
```

### Pattern 4: Loop Counter

```typescript
function sumToN(n: i64): i64 {
    let sum: i64 = 0;
    let i: i64 = 1;
    
    __asm__ {
        loop_start:
        LOAD i
        LOAD n
        GT
        JMPIF loop_end
        
        // sum += i
        LOAD sum
        LOAD i
        ADD
        STORE sum
        
        // i++
        LOAD i
        PUSH i64 1
        ADD
        STORE i
        
        JMP loop_start
        
        loop_end:
    }
    
    return sum;
}
```

### Pattern 5: Bit Manipulation

```typescript
function isPowerOfTwo(x: i64): bool {
    let result: bool;
    
    __asm__ {
        LOAD x
        PUSH i64 0
        GT
        JMPIF check_bits
        
        PUSH bool false
        STORE result
        JMP end
        
        check_bits:
        LOAD x
        LOAD x
        PUSH i64 1
        SUB
        BAND
        PUSH i64 0
        EQ
        STORE result
        
        end:
    }
    
    return result;
}
```

## Best Practices

### 1. Document Assembly Blocks

```typescript
__asm__ {
    // Calculate (a + b) * c
    LOAD a
    LOAD b
    ADD
    LOAD c
    MUL
    STORE result
}
```

### 2. Keep Assembly Blocks Small

```typescript
// Good: Small, focused assembly block
function optimizedAdd(x: i64, y: i64): i64 {
    let result: i64;
    __asm__ {
        LOAD x
        LOAD y
        ADD
        STORE result
    }
    return result;
}

// Avoid: Large, complex assembly blocks
// Use high-level code for complex logic
```

### 3. Verify Stack Balance

Ensure the stack is balanced at the end of assembly blocks:

```typescript
// Good: Stack is balanced
__asm__ {
    PUSH i64 10
    PUSH i64 20
    ADD          // Stack: [30]
    STORE result // Stack: []
}

// Bad: Stack not balanced
__asm__ {
    PUSH i64 10
    PUSH i64 20
    // Missing ADD or STORE
}
```

### 4. Use Labels Descriptively

```typescript
__asm__ {
    LOAD x
    PUSH i64 0
    LT
    JMPIF is_negative
    
    // Handle positive case
    JMP end
    
    is_negative:
    // Handle negative case
    
    end:
}
```

### 5. Consider Readability

```typescript
// Sometimes high-level code is better
let result: i64 = (a + b) * c;

// Only use assembly when there's a clear benefit
let optimized: i64;
__asm__ {
    LOAD a
    LOAD b
    ADD
    LOAD c
    MUL
    STORE optimized
}
```

## Examples

### Example 1: Absolute Value

```typescript
function abs(x: i64): i64 {
    let result: i64;
    
    __asm__ {
        LOAD x
        DUP
        PUSH i64 0
        LT
        JMPIF is_negative
        
        // Positive: just store x
        STORE result
        JMP end
        
        is_negative:
        // Negative: negate x
        PUSH i64 0
        SWAP
        SUB
        STORE result
        
        end:
    }
    
    return result;
}
```

### Example 2: Clamp Value

```typescript
function clamp(value: i64, min: i64, max: i64): i64 {
    let result: i64;
    
    __asm__ {
        // Check if value < min
        LOAD value
        LOAD min
        LT
        JMPIF use_min
        
        // Check if value > max
        LOAD value
        LOAD max
        GT
        JMPIF use_max
        
        // Value is in range
        LOAD value
        STORE result
        JMP end
        
        use_min:
        LOAD min
        STORE result
        JMP end
        
        use_max:
        LOAD max
        STORE result
        
        end:
    }
    
    return result;
}
```

### Example 3: Fast Exponentiation

```typescript
function fastPow2(exponent: i64): i64 {
    let result: i64;
    
    __asm__ {
        PUSH i64 1
        LOAD exponent
        SHL
        STORE result
    }
    
    return result;  // Returns 2^exponent
}
```

## Debugging Assembly

### Disassemble Compiled Code

Use the disassembler to see what the compiler generates:

```bash
./poh-blockchain qsc compile -i program.qs -o program.qsb
./poh-blockchain qsc disassemble -i program.qsb -o program.qsa
```

### Compare with Your Assembly

Review the disassembled output to understand:
- How the compiler translates high-level code
- Whether your assembly is more efficient
- If there are optimization opportunities

### Verbose Compilation

Use verbose mode to see compilation details:

```bash
./poh-blockchain qsc compile -i program.qs -o program.qsb -v
```

## Next Steps

- [Bytecode Reference](bytecode-reference.md) - Complete instruction set
- [Cost Model Guide](cost-model.md) - Understanding instruction costs
- [Language Reference](language-reference.md) - High-level language features
- [Examples](../examples/README.md) - More code examples

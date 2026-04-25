# Cost Model and Optimization Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Cost Model Overview](#cost-model-overview)
3. [Instruction Costs](#instruction-costs)
4. [Compute Budget](#compute-budget)
5. [Optimization Strategies](#optimization-strategies)
6. [Cost Analysis](#cost-analysis)
7. [Best Practices](#best-practices)

## Introduction

QuanticScript uses a deterministic cost model where every bytecode instruction has a fixed computational cost. This prevents resource exhaustion attacks and ensures fair resource allocation across all programs.

### Why Cost Metering?

- **Prevent DoS attacks**: Malicious programs can't consume unlimited resources
- **Fair resource allocation**: All programs pay proportionally for computation
- **Predictable performance**: Developers can estimate execution costs
- **Consensus safety**: All nodes agree on execution costs

## Cost Model Overview

### How Costs Work

1. Each instruction has a fixed cost based on its computational complexity
2. Before execution, a compute budget is allocated (based on transaction fee)
3. Each instruction deducts its cost from the budget
4. If budget reaches zero, execution halts with an error
5. Unused budget is not refunded (prevents timing attacks)

### Cost Units

Costs are measured in abstract "compute units":
- Simple operations (stack manipulation): 1-2 units
- Arithmetic operations: 2-4 units
- Memory operations: 3-5 units
- Blockchain operations: 30-100 units
- Cryptographic operations: 50-100 units
- Cross-program invocation: 200+ units

## Instruction Costs

### Stack Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `PUSH` | 1 | Push value onto stack |
| `POP` | 1 | Remove top of stack |
| `DUP` | 1 | Duplicate top value |
| `SWAP` | 1 | Swap top two values |

**Example:**
```typescript
__asm__ {
    PUSH i64 42    // Cost: 1
    DUP            // Cost: 1
    SWAP           // Cost: 1
}
// Total: 3 units
```

### Arithmetic Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `ADD` | 2 | Addition |
| `SUB` | 2 | Subtraction |
| `MUL` | 3 | Multiplication |
| `DIV` | 4 | Division |
| `MOD` | 4 | Modulo |

**Example:**
```typescript
let result: i64 = (a + b) * c / d;
// ADD: 2, MUL: 3, DIV: 4
// Total: 9 units
```

### Comparison Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `EQ` | 2 | Equal |
| `LT` | 2 | Less than |
| `GT` | 2 | Greater than |
| `LTE` | 2 | Less than or equal |
| `GTE` | 2 | Greater than or equal |

**Example:**
```typescript
if (x > 0 && x < 100) {
    // GT: 2, LT: 2, AND: 2
    // Total: 6 units
}
```

### Logical Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `AND` | 2 | Logical AND |
| `OR` | 2 | Logical OR |
| `NOT` | 1 | Logical NOT |

### Bitwise Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `BAND` | 2 | Bitwise AND |
| `BOR` | 2 | Bitwise OR |
| `BXOR` | 2 | Bitwise XOR |
| `BNOT` | 1 | Bitwise NOT |
| `SHL` | 2 | Shift left |
| `SHR` | 2 | Shift right |

### Control Flow

| Instruction | Cost | Description |
|-------------|------|-------------|
| `JMP` | 2 | Unconditional jump |
| `JMPIF` | 2 | Conditional jump |
| `CALL` | 5 | Function call |
| `RET` | 2 | Return from function |

**Example:**
```typescript
function helper(): i64 {
    return 42;
}

let x: i64 = helper();
// CALL: 5, RET: 2
// Total: 7 units (plus function body)
```

### Memory Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `LOAD` | 3 | Load local variable |
| `STORE` | 3 | Store local variable |
| `LOADG` | 10 | Load global state |
| `STOREG` | 15 | Store global state |

**Example:**
```typescript
let x: i64 = 10;  // STORE: 3
let y: i64 = x;   // LOAD: 3, STORE: 3
// Total: 9 units
```

### Blockchain Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `GETFILE` | 50 | Get file from FileStore |
| `GETFILEMUT` | 50 | Get mutable file reference |
| `UPDATEFILE` | 100 | Update file in FileStore |
| `GETBALANCE` | 30 | Get file balance |
| `UPDATEBALANCE` | 80 | Update file balance |
| `GETSIGNER` | 10 | Get transaction signer |
| `HASSIGNER` | 15 | Check if pubkey is signer |
| `GETINSTRDATA` | 5 | Get instruction data |
| `GETPROGRAMID` | 5 | Get current program ID |

**Example:**
```typescript
let balance: i64 = getBalance(accountId);  // Cost: 30
updateBalance(accountId, -100);            // Cost: 80
// Total: 110 units
```

### Cryptographic Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `SHA256` | 50 | SHA-256 hash |
| `VERIFYSIG` | 100 | Verify Ed25519 signature |
| `DERIVEPUBKEY` | 80 | Derive public key |

**Example:**
```typescript
let hash: bytes = sha256(data);                           // Cost: 50
let valid: bool = verifySignature(pubkey, msg, sig);      // Cost: 100
// Total: 150 units
```

### Cross-Program Invocation

| Instruction | Cost | Description |
|-------------|------|-------------|
| `INVOKE` | 200 | Invoke another program |
| `INVOKERET` | 10 | Return from invocation |

**Note:** The invoked program's execution cost is added to the total.

**Example:**
```typescript
// let result: bytes = invoke(programId, data);
// Cost: 200 + (cost of invoked program)
```

### Query Operations

| Instruction | Cost | Description |
|-------------|------|-------------|
| `QUERYBLOCK` | 100 | Query finalized block |
| `QUERYTX` | 80 | Query finalized transaction |
| `QUERYINSTR` | 80 | Query finalized instruction |

## Compute Budget

### Budget Allocation

The compute budget is determined by the transaction fee:

```
Compute Budget = Transaction Fee * Units Per Token
```

### Budget Exhaustion

If the budget reaches zero during execution:

```typescript
export function entry(ctx: InstructionContext): i64 {
    // If this loop runs too long, budget will be exhausted
    for (let i: i64 = 0; i < 1000000; i = i + 1) {
        // Each iteration costs units
    }
    return 0;  // May never reach here if budget exhausted
}
```

**Error:** `OutOfComputeError: Compute budget exhausted`

### Estimating Budget Needs

Calculate the worst-case cost of your program:

```typescript
function transfer(amount: i64): i64 {
    // getBalance: 30
    let balance: i64 = getBalance(sourceId);
    
    // LT comparison: 2
    if (balance < amount) {
        return -1;
    }
    
    // updateBalance: 80 (x2)
    updateBalance(sourceId, -amount);
    updateBalance(destId, amount);
    
    return 0;
}

// Worst case: 30 + 2 + 80 + 80 = 192 units
```

## Optimization Strategies

### 1. Minimize Expensive Operations

**Avoid:**
```typescript
// Multiple balance queries
for (let i: i64 = 0; i < 10; i = i + 1) {
    let balance: i64 = getBalance(accountId);  // 30 units each
    // Total: 300 units
}
```

**Better:**
```typescript
// Query once
let balance: i64 = getBalance(accountId);  // 30 units
for (let i: i64 = 0; i < 10; i = i + 1) {
    // Use cached balance
}
```

### 2. Use Efficient Algorithms

**Avoid:**
```typescript
// Inefficient: O(n²)
function sumPairs(arr: i64[]): i64 {
    let sum: i64 = 0;
    for (let i: i64 = 0; i < len; i = i + 1) {
        for (let j: i64 = 0; j < len; j = j + 1) {
            sum = sum + arr[i] * arr[j];  // Many operations
        }
    }
    return sum;
}
```

**Better:**
```typescript
// Efficient: O(n)
function sumPairs(arr: i64[]): i64 {
    let sum: i64 = 0;
    let total: i64 = 0;
    
    // First pass: sum all elements
    for (let i: i64 = 0; i < len; i = i + 1) {
        total = total + arr[i];
    }
    
    // Result is total²
    return total * total;
}
```

### 3. Optimize Arithmetic

**Avoid:**
```typescript
let doubled: i64 = x * 2;  // MUL: 3 units
```

**Better:**
```typescript
let doubled: i64 = x + x;  // ADD: 2 units
```

**Even Better (with assembly):**
```typescript
let doubled: i64;
__asm__ {
    LOAD x
    DUP     // 1 unit
    ADD     // 2 units
    STORE doubled
}
// Total: 3 units (vs 3 for MUL, but more explicit)
```

### 4. Use Bit Operations

**Avoid:**
```typescript
let powerOf2: i64 = pow(2, n);  // Expensive
```

**Better:**
```typescript
let powerOf2: i64 = 1 << n;  // SHL: 2 units
```

**Avoid:**
```typescript
let divided: i64 = x / 4;  // DIV: 4 units
```

**Better:**
```typescript
let divided: i64 = x >> 2;  // SHR: 2 units
```

### 5. Early Exit

**Avoid:**
```typescript
function validate(x: i64, y: i64, z: i64): i64 {
    let valid: bool = true;
    
    if (x < 0) {
        valid = false;
    }
    if (y < 0) {
        valid = false;
    }
    if (z < 0) {
        valid = false;
    }
    
    if (!valid) {
        return -1;
    }
    return 0;
}
```

**Better:**
```typescript
function validate(x: i64, y: i64, z: i64): i64 {
    if (x < 0) {
        return -1;  // Exit early
    }
    if (y < 0) {
        return -1;
    }
    if (z < 0) {
        return -1;
    }
    return 0;
}
```

### 6. Batch Operations

**Avoid:**
```typescript
updateBalance(account, 10);   // 80 units
updateBalance(account, 20);   // 80 units
updateBalance(account, 30);   // 80 units
// Total: 240 units
```

**Better:**
```typescript
let total: i64 = 10 + 20 + 30;  // 4 units
updateBalance(account, total);   // 80 units
// Total: 84 units
```

### 7. Avoid Redundant Checks

**Avoid:**
```typescript
if (x > 0) {
    if (x > 0) {  // Redundant
        // ...
    }
}
```

**Better:**
```typescript
if (x > 0) {
    // ...
}
```

### 8. Use Inline Assembly for Hot Paths

For frequently executed code, inline assembly can save costs:

```typescript
// High-level: ~10 units
function calculate(a: i64, b: i64, c: i64): i64 {
    return (a + b) * c;
}

// Assembly: ~8 units
function calculateOptimized(a: i64, b: i64, c: i64): i64 {
    let result: i64;
    __asm__ {
        LOAD a      // 3
        LOAD b      // 3
        ADD         // 2
        LOAD c      // 3
        MUL         // 3
        STORE result // 3
    }
    return result;
}
// Total: 17 units (but more explicit control)
```

## Cost Analysis

### Analyzing Your Program

1. **Identify hot paths**: Which code runs most frequently?
2. **Count operations**: Sum up instruction costs
3. **Find bottlenecks**: Which operations are most expensive?
4. **Optimize**: Apply strategies to reduce costs

### Example Analysis

```typescript
export function entry(ctx: InstructionContext): i64 {
    // Parse data: ~10 units
    let data: bytes = getInstructionData();  // 5 units
    
    // Validation: ~10 units
    if (amount <= 0) {  // 2 units
        return -1;
    }
    
    // Balance check: ~35 units
    let balance: i64 = getBalance(sourceId);  // 30 units
    if (balance < amount) {  // 2 units
        return -2;
    }
    
    // Transfer: ~160 units
    updateBalance(sourceId, -amount);   // 80 units
    updateBalance(destId, amount);      // 80 units
    
    return 0;
}

// Total worst case: ~215 units
```

### Profiling Tips

1. **Use verbose compilation**: See generated bytecode
2. **Disassemble output**: Count actual instructions
3. **Test with limits**: Run with low compute budgets
4. **Measure iterations**: Count loop iterations

## Best Practices

### 1. Design for Efficiency

Plan your program structure to minimize expensive operations:

```typescript
// Good: Single balance query
let balance: i64 = getBalance(accountId);
let canTransfer: bool = balance >= amount;

// Avoid: Multiple queries
if (getBalance(accountId) >= amount) {
    if (getBalance(accountId) >= fee) {
        // ...
    }
}
```

### 2. Set Compute Budgets Appropriately

Ensure transactions have sufficient budget:

```
Minimum Budget = Worst Case Cost * Safety Factor
```

Example: If worst case is 215 units, set budget to 300-400 units.

### 3. Document Costs

Add comments showing expected costs:

```typescript
function transfer(amount: i64): i64 {
    // Cost: ~215 units worst case
    // - getBalance: 30
    // - comparisons: 4
    // - updateBalance: 160
    // - overhead: 21
    
    let balance: i64 = getBalance(sourceId);
    // ...
}
```

### 4. Test Edge Cases

Test with minimal compute budgets to ensure your program handles exhaustion gracefully:

```typescript
// Test with budget just below requirement
// Should fail gracefully, not corrupt state
```

### 5. Monitor Production Costs

Track actual execution costs in production:
- Average cost per transaction
- Maximum observed cost
- Failed transactions due to budget exhaustion

### 6. Optimize Incrementally

1. Profile to find bottlenecks
2. Optimize the most expensive operations first
3. Measure improvement
4. Repeat

### 7. Balance Readability and Performance

Don't sacrifice code clarity for minor optimizations:

```typescript
// Clear and good enough
let result: i64 = x * 2;

// Only optimize if this is a proven bottleneck
let result: i64 = x + x;
```

## Cost Comparison Examples

### Example 1: Loop Optimization

**Unoptimized:**
```typescript
let sum: i64 = 0;
for (let i: i64 = 0; i < 100; i = i + 1) {
    sum = sum + i;
}
// Cost: ~600 units (100 iterations * ~6 units each)
```

**Optimized:**
```typescript
// Use formula: sum = n * (n + 1) / 2
let n: i64 = 100;
let sum: i64 = (n * (n + 1)) / 2;
// Cost: ~12 units (MUL + ADD + DIV)
```

### Example 2: Conditional Logic

**Unoptimized:**
```typescript
let result: i64;
if (x > 0) {
    result = 1;
} else if (x < 0) {
    result = -1;
} else {
    result = 0;
}
// Cost: ~12 units worst case
```

**Optimized:**
```typescript
// Use arithmetic trick
let result: i64 = (x > 0) - (x < 0);
// Cost: ~6 units
```

### Example 3: String Operations

**Unoptimized:**
```typescript
let result: string = "";
for (let i: i64 = 0; i < 10; i = i + 1) {
    result = stringConcat(result, "x");
}
// Cost: ~500+ units (repeated concatenation)
```

**Optimized:**
```typescript
// Build once if possible, or use array
// Cost depends on implementation
```

## Storage Cost Model

In addition to compute costs, the blockchain enforces storage costs for on-chain data. Files require Neon balance proportional to their data size using an exponential growth formula.

### Storage Cost Formula

```
cost = base_cost_per_kb * size_in_kb * (1.1 ^ size_in_mb)
```

Where:
- `base_cost_per_kb = 1000` Neon units
- Size is rounded up to the nearest KB
- Exponential factor grows with file size to discourage state bloat

### Storage Cost Calculator

Use the storage cost calculator utility to estimate costs before deployment:

```bash
go run calculate_storage_cost.go <file_path>
```

**Example:**
```bash
go run calculate_storage_cost.go programs/token/token.qsb
```

**Output:**
```
File: programs/token/token.qsb
Size: 2048 bytes
Storage Cost: 2000 Neon units
```

### Storage Cost Examples

| File Size | Size (KB) | Size (MB) | Multiplier | Cost (Neon) |
|-----------|-----------|-----------|------------|-------------|
| 512 bytes | 1 KB | 0.0005 MB | 1.00005 | 1,000 |
| 1 KB | 1 KB | 0.001 MB | 1.0001 | 1,000 |
| 10 KB | 10 KB | 0.01 MB | 1.001 | 10,010 |
| 100 KB | 100 KB | 0.1 MB | 1.01 | 101,000 |
| 1 MB | 1024 KB | 1 MB | 1.1 | 1,126,400 |
| 10 MB | 10240 KB | 10 MB | 2.59 | 26,542,080 |

### Storage Rent Enforcement

The FileStore validates storage costs at multiple points:

1. **File Creation**: Balance must cover initial data size
2. **Balance Decrease**: Remaining balance must still cover data size
3. **Data Size Increase**: Balance must cover new data size

**Example validation:**
```typescript
// Creating a file with 2 KB of data
let dataSize: i64 = 2048;
let storageCost: i64 = calculateStorageCost(dataSize);  // ~2000 Neon
let requiredBalance: i64 = storageCost;

// File creation will fail if balance < requiredBalance
```

### Optimizing Storage Costs

1. **Minimize data size**: Store only essential data on-chain
2. **Use efficient serialization**: Compact binary formats over JSON
3. **Compress data**: Use compression for large data structures
4. **Off-chain storage**: Store large data off-chain, only hashes on-chain
5. **Close unused accounts**: Reclaim Neon by closing empty accounts

**Example:**
```typescript
// Inefficient: Store full string
let data: string = "This is a very long description...";  // 100+ bytes

// Efficient: Store hash reference
let dataHash: bytes = sha256(data);  // 32 bytes
```

## Next Steps

- [Language Reference](language-reference.md) - Language syntax
- [Standard Library Reference](stdlib-reference.md) - Built-in functions
- [Inline Assembly Guide](inline-assembly.md) - Low-level optimization
- [CLI Usage Guide](../guides/cli-usage.md) - Storage cost calculator
- [Examples](../examples/README.md) - Practical examples

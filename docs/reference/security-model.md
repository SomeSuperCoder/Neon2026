# QuanticScript Security Model

## Overview

QuanticScript implements multiple layers of security to ensure safe execution of smart contracts on the blockchain. This document describes the security mechanisms and restrictions in place.

## Privilege Levels

### System Program

The system program is a special built-in program with elevated privileges. It has the unique identifier:

```
FileID: 0x0000000000000000000000000000000000000000000000000000000000000001
```

The system program can:
- Modify account balances directly via `UPDATEBALANCE`
- Create and close accounts
- Allocate and deallocate storage
- Perform privileged operations

### Regular Programs

Regular programs (user-deployed smart contracts) have restricted privileges:
- Cannot modify balances directly
- Must invoke the system program for balance transfers
- Subject to compute budget limits
- Cannot escape the sandbox

## Privileged Instructions

### UPDATEBALANCE (0x74)

**Restriction:** System program only

The `UPDATEBALANCE` instruction can only be executed by the system program. Any attempt by a regular program to execute this instruction will result in a security error:

```
Error: UPDATEBALANCE can only be called by the system program
```

**Rationale:** Direct balance manipulation must be controlled to prevent:
- Unauthorized minting of tokens
- Balance theft
- Double-spending attacks
- Accounting inconsistencies

**For Regular Programs:**

To transfer balances, regular programs must use cross-program invocation to call the system program:

```typescript
// Example: Transfer tokens between accounts
export function transfer(from: i64, to: i64, amount: i64): i64 {
    // Verify authorization
    if (!hasSigner(getAccountOwner(from))) {
        return -4;  // Unauthorized
    }
    
    // Invoke system program to perform transfer
    let systemProgramId: i64 = 0x01;
    let transferData: bytes = encodeTransferInstruction(from, to, amount);
    
    let result: bytes = invoke(systemProgramId, transferData, 10000);
    
    return 0;  // Success
}
```

## Sandbox Isolation

### Memory Isolation

Programs execute in isolated memory spaces:
- Stack: 256 values maximum
- Local memory: 256 slots
- No direct memory pointer access
- No access to other programs' memory

### Compute Budget

Every instruction consumes compute units from a fixed budget:
- Budget specified at invocation time
- Execution halts when budget exhausted
- Prevents infinite loops and DoS attacks

### File Access Control

Programs can only access files explicitly declared in the transaction:
- Read-only access: `GETFILE`
- Mutable access: `GETFILEMUT`
- Undeclared file access results in error

## Cross-Program Invocation Security

### Call Depth Limit

Maximum invocation depth: **4 levels**

```
Program A → Program B → Program C → Program D → Program E (REJECTED)
```

**Rationale:** Prevents stack overflow and excessive resource consumption.

### Program Declaration

Invoked programs must be declared in the transaction's program list:

```typescript
// This will fail if targetProgram is not declared
let result: bytes = invoke(targetProgram, data, budget);
```

**Rationale:** Prevents unexpected program execution and enables static analysis.

### Budget Allocation

Compute budget is allocated from the caller's budget:

```typescript
// Caller has 10000 units
// Allocates 5000 to invoked program
let result: bytes = invoke(program, data, 5000);
// Caller now has 5000 units remaining
```

### State Rollback

Failed invocations automatically rollback all state changes:

```typescript
// If this invocation fails, all state changes are reverted
try {
    let result: bytes = invoke(program, data, budget);
} catch {
    // State is rolled back to before invocation
}
```

## Type Safety

### Type Confusion Prevention

The interpreter enforces strict type checking:
- Cannot use integers as booleans
- Cannot use booleans in arithmetic
- Type mismatches cause immediate errors

### Bounds Checking

All array and memory accesses are bounds-checked:
- Stack underflow/overflow detection
- Memory access validation
- Array index validation

## Determinism Enforcement

### Prohibited Operations

The following operations are rejected at compile time:

1. **Random Number Generation**
   ```typescript
   let x: i64 = random();  // REJECTED
   ```

2. **System Time Access**
   ```typescript
   let now: i64 = getTime();  // REJECTED
   ```

3. **File I/O**
   ```typescript
   let data: bytes = readFile("file.txt");  // REJECTED
   ```

4. **Network Operations**
   ```typescript
   let response: bytes = httpGet("url");  // REJECTED
   ```

5. **Floating-Point Operations**
   ```typescript
   let x: f64 = 3.14;  // REJECTED
   ```

**Rationale:** All operations must be deterministic to ensure consensus across nodes.

## Resource Limits

### Stack Size

Maximum stack depth: **256 values**

Prevents stack overflow attacks.

### Memory Size

Maximum local memory: **256 slots**

Each slot can hold one value of any type.

### Call Stack Depth

Maximum function call depth: **64 levels**

Prevents stack overflow from recursive functions.

### Bytecode Size

Maximum bytecode size: **1 MB**

Prevents excessively large programs.

## Error Handling

### Security Errors

Security violations result in immediate execution halt:

```
Error: UPDATEBALANCE can only be called by the system program
Error: maximum cross-program invocation depth exceeded
Error: program not in declared program list
Error: insufficient compute budget for invocation
```

### Error Propagation

Errors propagate up the call stack:
- State changes are rolled back
- Compute budget is not refunded
- Error message is returned to caller

## Best Practices

### 1. Validate All Inputs

```typescript
export function entry(ctx: InstructionContext): i64 {
    let data: bytes = getInstructionData();
    
    // Validate data length
    if (data.length < 8) {
        return -2;  // Invalid input
    }
    
    // Parse and validate parameters
    let amount: i64 = parseI64(data, 0);
    if (amount <= 0) {
        return -2;  // Invalid amount
    }
    
    // Continue processing...
    return 0;
}
```

### 2. Check Authorization

```typescript
function requireSigner(pubkey: PublicKey): i64 {
    if (!hasSigner(pubkey)) {
        return -4;  // Unauthorized
    }
    return 0;
}
```

### 3. Use System Program for Balance Operations

```typescript
// WRONG: Direct balance manipulation (will fail)
updateBalance(accountId, amount);

// CORRECT: Invoke system program
let systemProgramId: i64 = 0x01;
let transferData: bytes = encodeTransfer(from, to, amount);
invoke(systemProgramId, transferData, 10000);
```

### 4. Handle Invocation Failures

```typescript
// Check invocation result
let result: bytes = invoke(program, data, budget);
if (result.length == 0) {
    return -1;  // Invocation failed
}
```

### 5. Allocate Sufficient Compute Budget

```typescript
// Allocate enough budget for complex operations
let budget: i64 = 50000;  // Adjust based on expected cost
let result: bytes = invoke(program, data, budget);
```

## Security Audit Checklist

When reviewing QuanticScript programs, check for:

- [ ] Input validation on all external data
- [ ] Authorization checks before privileged operations
- [ ] Proper error handling
- [ ] No direct balance manipulation (use system program)
- [ ] Sufficient compute budget allocation
- [ ] Bounds checking on array access
- [ ] No integer overflow/underflow vulnerabilities
- [ ] Proper use of cross-program invocation
- [ ] Declared programs list includes all invoked programs

## Related Documentation

- [Bytecode Reference](bytecode-reference.md) - Instruction set details
- [Standard Library Reference](stdlib-reference.md) - Built-in functions
- [Cost Model Guide](cost-model.md) - Compute costs
- [Language Reference](language-reference.md) - Language syntax

## Version History

- **v1.0** (2026-04-24): Initial security model documentation
  - Added UPDATEBALANCE privilege restriction
  - Documented system program privileges
  - Defined sandbox isolation mechanisms

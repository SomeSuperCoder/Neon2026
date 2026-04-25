# QuanticScript Standard Library Reference

## Table of Contents

1. [Overview](#overview)
2. [Blockchain Module](#blockchain-module)
3. [Crypto Module](#crypto-module)
4. [Collections Module](#collections-module)
5. [String Module](#string-module)
6. [Math Module](#math-module)
7. [Query Module](#query-module)
8. [Invoke Module](#invoke-module)

## Overview

The QuanticScript standard library provides built-in functions for common operations. All functions are deterministic and cost-metered.

### Importing Modules

```typescript
// Functions are available globally, no import needed
let balance: i64 = getBalance(accountId);
```

## Blockchain Module

Functions for interacting with blockchain state and files.

### getBalance

Get the balance of a file/account.

```typescript
function getBalance(fileId: i64): i64
```

**Parameters:**
- `fileId`: The file identifier

**Returns:** The balance in the smallest unit

**Cost:** 30

**Example:**
```typescript
let accountId: i64 = 100;
let balance: i64 = getBalance(accountId);

if (balance < 1000) {
    return -1;  // Insufficient balance
}
```

---

### updateBalance

Update the balance of a file/account.

**SECURITY: This function can only be called by the system program.** Regular programs must use `invoke` to call the system program's Transfer instruction to move balances between accounts.

```typescript
function updateBalance(fileId: i64, delta: i64): void
```

**Parameters:**
- `fileId`: The file identifier
- `delta`: Amount to add (positive) or subtract (negative)

**Cost:** 80

**Example (System Program Only):**
```typescript
// This only works inside the system program
updateBalance(sourceId, -100);
updateBalance(destId, 100);
```

**For Regular Programs:**
```typescript
// Regular programs must invoke the system program to transfer balances
let systemProgramId: i64 = 0x01;  // System program ID
let transferData: bytes = encodeTransfer(amount);
invoke(systemProgramId, transferData, 10000);
```

**Note:** This function will fail if:
- Called by a non-system program (security error)
- The resulting balance would be negative

---

### getInstructionData

Get the data passed to the current instruction.

```typescript
function getInstructionData(): bytes
```

**Returns:** Byte array containing instruction data

**Cost:** 5

**Example:**
```typescript
export function entry(ctx: InstructionContext): i64 {
    let data: bytes = getInstructionData();
    // Parse data to extract parameters
    return 0;
}
```

---

### getProgramId

Get the ID of the currently executing program.

```typescript
function getProgramId(): i64
```

**Returns:** The program's file ID

**Cost:** 5

**Example:**
```typescript
let programId: i64 = getProgramId();
// Use program ID for fee collection, etc.
updateBalance(programId, feeAmount);
```

---

### hasSigner

Check if a specific public key signed the transaction.

```typescript
function hasSigner(pubkey: PublicKey): bool
```

**Parameters:**
- `pubkey`: Public key to check

**Returns:** `true` if the public key signed the transaction

**Cost:** 15

**Example:**
```typescript
// In a real implementation
// let ownerKey: PublicKey = getFileOwner(accountId);
// if (!hasSigner(ownerKey)) {
//     return -1;  // Unauthorized
// }
```

## Crypto Module

Cryptographic functions for hashing and signature verification.

### sha256

Compute SHA-256 hash of data.

```typescript
function sha256(data: bytes): bytes
```

**Parameters:**
- `data`: Data to hash

**Returns:** 32-byte hash

**Cost:** 50

**Example:**
```typescript
let message: bytes = [1, 2, 3, 4, 5];
let hash: bytes = sha256(message);
```

---

### verifySignature

Verify an Ed25519 signature.

```typescript
function verifySignature(pubkey: PublicKey, message: bytes, signature: bytes): bool
```

**Parameters:**
- `pubkey`: Public key (32 bytes)
- `message`: Message that was signed
- `signature`: Signature to verify (64 bytes)

**Returns:** `true` if signature is valid

**Cost:** 100

**Example:**
```typescript
let isValid: bool = verifySignature(publicKey, message, signature);
if (!isValid) {
    return -1;  // Invalid signature
}
```

---

### derivePublicKey

Derive a public key from a seed.

```typescript
function derivePublicKey(seed: bytes): PublicKey
```

**Parameters:**
- `seed`: Seed bytes (32 bytes)

**Returns:** Derived public key

**Cost:** 80

**Example:**
```typescript
let seed: bytes = sha256(baseData);
let derivedKey: PublicKey = derivePublicKey(seed);
```

## Collections Module

Functions for working with arrays and data structures.

### len

Get the length of a byte array or array.

```typescript
function len(data: bytes): i64
```

**Parameters:**
- `data`: Byte array

**Returns:** Number of bytes

**Cost:** 2

**Example:**
```typescript
let data: bytes = [1, 2, 3, 4, 5];
let length: i64 = len(data);  // 5

let instrData: bytes = getInstructionData();
if (len(instrData) < 32) {
    return -1;  // Insufficient data
}
```

**Note:** This is the preferred function for getting the length of byte arrays. For typed arrays, use `arrayLength`.

---

### arrayLength

Get the length of an array.

```typescript
function arrayLength<T>(arr: T[]): i64
```

**Parameters:**
- `arr`: Array

**Returns:** Number of elements

**Cost:** 2

**Example:**
```typescript
let numbers: i64[] = [1, 2, 3, 4, 5];
let len: i64 = arrayLength(numbers);  // 5
```

---

### arrayGet

Get an element from an array by index.

```typescript
function arrayGet<T>(arr: T[], index: i64): T
```

**Parameters:**
- `arr`: Array
- `index`: Zero-based index

**Returns:** Element at index

**Cost:** 3

**Example:**
```typescript
let numbers: i64[] = [10, 20, 30];
let second: i64 = arrayGet(numbers, 1);  // 20
```

---

### arraySet

Set an element in an array by index.

```typescript
function arraySet<T>(arr: T[], index: i64, value: T): void
```

**Parameters:**
- `arr`: Array
- `index`: Zero-based index
- `value`: Value to set

**Cost:** 5

**Example:**
```typescript
let numbers: i64[] = [10, 20, 30];
arraySet(numbers, 1, 25);  // [10, 25, 30]
```

---

### arrayPush

Add an element to the end of an array.

```typescript
function arrayPush<T>(arr: T[], value: T): void
```

**Parameters:**
- `arr`: Array
- `value`: Value to add

**Cost:** 5

**Example:**
```typescript
let numbers: i64[] = [1, 2, 3];
arrayPush(numbers, 4);  // [1, 2, 3, 4]
```

---

### arrayMap

Transform each element of an array.

```typescript
function arrayMap<T, U>(arr: T[], fn: (T) => U): U[]
```

**Parameters:**
- `arr`: Input array
- `fn`: Transformation function

**Returns:** New array with transformed elements

**Cost:** 10 + (5 * array length)

**Example:**
```typescript
function double(x: i64): i64 {
    return x * 2;
}

let numbers: i64[] = [1, 2, 3];
let doubled: i64[] = arrayMap(numbers, double);  // [2, 4, 6]
```

---

### arrayFilter

Filter array elements based on a predicate.

```typescript
function arrayFilter<T>(arr: T[], fn: (T) => bool): T[]
```

**Parameters:**
- `arr`: Input array
- `fn`: Predicate function

**Returns:** New array with elements that pass the predicate

**Cost:** 10 + (5 * array length)

**Example:**
```typescript
function isPositive(x: i64): bool {
    return x > 0;
}

let numbers: i64[] = [-1, 2, -3, 4];
let positive: i64[] = arrayFilter(numbers, isPositive);  // [2, 4]
```

---

### arrayReduce

Reduce an array to a single value.

```typescript
function arrayReduce<T, U>(arr: T[], fn: (U, T) => U, initial: U): U
```

**Parameters:**
- `arr`: Input array
- `fn`: Reducer function
- `initial`: Initial accumulator value

**Returns:** Final accumulated value

**Cost:** 10 + (5 * array length)

**Example:**
```typescript
function sum(acc: i64, x: i64): i64 {
    return acc + x;
}

let numbers: i64[] = [1, 2, 3, 4];
let total: i64 = arrayReduce(numbers, sum, 0);  // 10
```

## String Module

Functions for string manipulation.

### stringLength

Get the length of a string.

```typescript
function stringLength(s: string): i64
```

**Parameters:**
- `s`: String

**Returns:** Number of characters

**Cost:** 2

**Example:**
```typescript
let text: string = "Hello";
let len: i64 = stringLength(text);  // 5
```

---

### stringConcat

Concatenate two strings.

```typescript
function stringConcat(a: string, b: string): string
```

**Parameters:**
- `a`: First string
- `b`: Second string

**Returns:** Concatenated string

**Cost:** 5 + (length of result)

**Example:**
```typescript
let greeting: string = stringConcat("Hello, ", "World!");
// "Hello, World!"
```

---

### stringSubstring

Extract a substring.

```typescript
function stringSubstring(s: string, start: i64, end: i64): string
```

**Parameters:**
- `s`: Source string
- `start`: Start index (inclusive)
- `end`: End index (exclusive)

**Returns:** Substring

**Cost:** 5 + (length of result)

**Example:**
```typescript
let text: string = "Hello, World!";
let sub: string = stringSubstring(text, 0, 5);  // "Hello"
```

---

### stringToBytes

Convert string to byte array.

```typescript
function stringToBytes(s: string): bytes
```

**Parameters:**
- `s`: String to convert

**Returns:** UTF-8 encoded bytes

**Cost:** 5 + (length of string)

**Example:**
```typescript
let text: string = "Hello";
let data: bytes = stringToBytes(text);
```

---

### bytesToString

Convert byte array to string.

```typescript
function bytesToString(b: bytes): string
```

**Parameters:**
- `b`: UTF-8 encoded bytes

**Returns:** String

**Cost:** 5 + (length of bytes)

**Example:**
```typescript
let data: bytes = [72, 101, 108, 108, 111];
let text: string = bytesToString(data);  // "Hello"
```

## Math Module

Mathematical functions (all operations are deterministic).

### abs

Absolute value.

```typescript
function abs(x: i64): i64
```

**Parameters:**
- `x`: Input value

**Returns:** Absolute value

**Cost:** 2

**Example:**
```typescript
let a: i64 = abs(-42);   // 42
let b: i64 = abs(100);   // 100
```

---

### min

Minimum of two values.

```typescript
function min(a: i64, b: i64): i64
```

**Parameters:**
- `a`: First value
- `b`: Second value

**Returns:** Smaller value

**Cost:** 2

**Example:**
```typescript
let smaller: i64 = min(10, 20);  // 10
```

---

### max

Maximum of two values.

```typescript
function max(a: i64, b: i64): i64
```

**Parameters:**
- `a`: First value
- `b`: Second value

**Returns:** Larger value

**Cost:** 2

**Example:**
```typescript
let larger: i64 = max(10, 20);  // 20
```

---

### pow

Power function (integer exponentiation).

```typescript
function pow(base: i64, exponent: i64): i64
```

**Parameters:**
- `base`: Base value
- `exponent`: Exponent (must be non-negative)

**Returns:** base^exponent

**Cost:** 10 + (2 * exponent)

**Example:**
```typescript
let result: i64 = pow(2, 10);  // 1024
```

**Note:** Be careful of overflow with large exponents.

---

### sqrt

Integer square root (floor).

```typescript
function sqrt(x: i64): i64
```

**Parameters:**
- `x`: Input value (must be non-negative)

**Returns:** Floor of square root

**Cost:** 20

**Example:**
```typescript
let result: i64 = sqrt(100);  // 10
let result2: i64 = sqrt(101); // 10 (floor)
```

## Query Module

Functions for querying finalized blockchain data.

### queryBlock

Query a finalized block by hash.

```typescript
function queryBlock(hash: bytes): Block | null
```

**Parameters:**
- `hash`: Block hash (32 bytes)

**Returns:** Block data or null if not found/not finalized

**Cost:** 100

**Example:**
```typescript
let blockHash: bytes = [/* 32 bytes */];
// let block: Block | null = queryBlock(blockHash);
// if (block == null) {
//     return -1;  // Block not found
// }
```

---

### queryTransaction

Query a finalized transaction by ID.

```typescript
function queryTransaction(txId: bytes): Transaction | null
```

**Parameters:**
- `txId`: Transaction ID (32 bytes)

**Returns:** Transaction data or null if not found/not finalized

**Cost:** 80

**Example:**
```typescript
let txId: bytes = [/* 32 bytes */];
// let tx: Transaction | null = queryTransaction(txId);
```

---

### queryInstruction

Query a finalized instruction.

```typescript
function queryInstruction(ref: InstructionRef): Instruction | null
```

**Parameters:**
- `ref`: Instruction reference

**Returns:** Instruction data or null if not found/not finalized

**Cost:** 80

## Invoke Module

Functions for cross-program invocation.

### invoke

Invoke another program.

```typescript
function invoke(programId: i64, data: bytes): bytes
```

**Parameters:**
- `programId`: Target program file ID
- `data`: Instruction data to pass

**Returns:** Result data from invoked program

**Cost:** 200 + (cost of invoked program)

**Example:**
```typescript
export function entry(ctx: InstructionContext): i64 {
    let validatorProgram: i64 = 999;
    let validationData: bytes = [1, 2, 3];
    
    // Invoke validator program
    // let result: bytes = invoke(validatorProgram, validationData);
    
    return 0;
}
```

**Notes:**
- Target program must be in the instruction's declared program list
- Maximum call depth is 4
- Failed invocations rollback all state changes

---

### getInvokeDepth

Get the current cross-program invocation depth.

```typescript
function getInvokeDepth(): i64
```

**Returns:** Current call depth (0 for top-level)

**Cost:** 2

**Example:**
```typescript
let depth: i64 = getInvokeDepth();
if (depth >= 4) {
    return -1;  // Max depth reached
}
```

## Error Handling

Standard library functions may fail in various ways:

### Common Error Patterns

```typescript
// Check return values
let balance: i64 = getBalance(accountId);
if (balance < 0) {
    return -1;  // Error occurred
}

// Validate inputs before calling
if (amount <= 0) {
    return -2;  // Invalid amount
}
updateBalance(accountId, amount);
```

### Error Codes

By convention, negative return values indicate errors:

- `-1`: General error
- `-2`: Invalid parameter
- `-3`: Insufficient balance
- `-4`: Unauthorized
- `-5`: Not found

## Performance Tips

### 1. Minimize Expensive Operations

```typescript
// Expensive: Multiple balance queries
let b1: i64 = getBalance(id);
let b2: i64 = getBalance(id);

// Better: Query once
let balance: i64 = getBalance(id);
```

### 2. Use Inline Assembly for Hot Paths

```typescript
// For simple operations, assembly can be faster
let doubled: i64;
__asm__ {
    LOAD x
    DUP
    ADD
    STORE doubled
}
```

### 3. Batch Operations

```typescript
// Better: Calculate total first, then update once
let total: i64 = amount1 + amount2 + amount3;
updateBalance(accountId, total);

// Avoid: Multiple updates
updateBalance(accountId, amount1);
updateBalance(accountId, amount2);
updateBalance(accountId, amount3);
```

## Next Steps

- [Language Reference](language-reference.md)
- [Bytecode Reference](bytecode-reference.md)
- [Cost Model Guide](cost-model.md)
- [Examples](../examples/README.md)

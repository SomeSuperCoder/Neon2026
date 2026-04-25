# QuanticScript Bytecode Reference

## Table of Contents

1. [Introduction](#introduction)
2. [Bytecode Format](#bytecode-format)
3. [Instruction Set](#instruction-set)
4. [Opcode Table](#opcode-table)
5. [Execution Model](#execution-model)
6. [Examples](#examples)

## Introduction

This document provides a complete reference for QuanticScript bytecode format and instruction set. It's useful for:

- Understanding compiler output
- Writing assembly programs
- Debugging bytecode execution
- Implementing bytecode tools

## Bytecode Format

### File Structure

QuanticScript bytecode files (`.qsb`) have the following structure:

```
Offset | Size | Field        | Description
-------|------|--------------|----------------------------------
0x00   | 4    | Magic        | "QSCB" (0x51534342)
0x04   | 2    | Version      | Format version (0x0100 = v1.0)
0x06   | 2    | Entry Offset | Offset to entry function
0x08   | N    | Bytecode     | Instruction bytes
```

### Instruction Format

Each instruction consists of:

```
[Opcode: 1 byte][Operands: 0-N bytes]
```

Operand encoding:
- Immediate values: Type-specific encoding
- Labels/offsets: 2-byte signed offset
- Variable references: Variable index

## Instruction Set

### Stack Operations

#### PUSH (0x01)
Push a value onto the stack.

**Format:** `PUSH <type> <value>`

**Operands:**
- Type byte (i64=0x01, bool=0x02, etc.)
- Value bytes (type-dependent)

**Stack:** `[] → [value]`

**Cost:** 1

**Example:**
```assembly
PUSH i64 42
PUSH bool true
```

---

#### POP (0x02)
Remove the top value from the stack.

**Format:** `POP`

**Operands:** None

**Stack:** `[value] → []`

**Cost:** 1

**Example:**
```assembly
PUSH i64 10
POP
```

---

#### DUP (0x03)
Duplicate the top stack value.

**Format:** `DUP`

**Operands:** None

**Stack:** `[value] → [value, value]`

**Cost:** 1

**Example:**
```assembly
PUSH i64 42
DUP
; Stack now has [42, 42]
```

---

#### SWAP (0x04)
Swap the top two stack values.

**Format:** `SWAP`

**Operands:** None

**Stack:** `[a, b] → [b, a]`

**Cost:** 1

**Example:**
```assembly
PUSH i64 10
PUSH i64 20
SWAP
; Stack: [20, 10]
```

### Arithmetic Operations

#### ADD (0x10)
Add two integers.

**Format:** `ADD`

**Operands:** None

**Stack:** `[a, b] → [a + b]`

**Cost:** 2

**Example:**
```assembly
PUSH i64 10
PUSH i64 20
ADD
; Stack: [30]
```

---

#### SUB (0x11)
Subtract two integers.

**Format:** `SUB`

**Operands:** None

**Stack:** `[a, b] → [a - b]`

**Cost:** 2

**Example:**
```assembly
PUSH i64 50
PUSH i64 20
SUB
; Stack: [30]
```

---

#### MUL (0x12)
Multiply two integers.

**Format:** `MUL`

**Operands:** None

**Stack:** `[a, b] → [a * b]`

**Cost:** 3

**Example:**
```assembly
PUSH i64 6
PUSH i64 7
MUL
; Stack: [42]
```

---

#### DIV (0x13)
Divide two integers.

**Format:** `DIV`

**Operands:** None

**Stack:** `[a, b] → [a / b]`

**Cost:** 4

**Note:** Division by zero causes runtime error.

**Example:**
```assembly
PUSH i64 100
PUSH i64 4
DIV
; Stack: [25]
```

---

#### MOD (0x14)
Modulo operation.

**Format:** `MOD`

**Operands:** None

**Stack:** `[a, b] → [a % b]`

**Cost:** 4

**Example:**
```assembly
PUSH i64 17
PUSH i64 5
MOD
; Stack: [2]
```

### Comparison Operations

#### EQ (0x20)
Test equality.

**Format:** `EQ`

**Operands:** None

**Stack:** `[a, b] → [a == b]`

**Cost:** 2

**Example:**
```assembly
PUSH i64 10
PUSH i64 10
EQ
; Stack: [true]
```

---

#### LT (0x21)
Test less than.

**Format:** `LT`

**Operands:** None

**Stack:** `[a, b] → [a < b]`

**Cost:** 2

---

#### GT (0x22)
Test greater than.

**Format:** `GT`

**Operands:** None

**Stack:** `[a, b] → [a > b]`

**Cost:** 2

---

#### LTE (0x23)
Test less than or equal.

**Format:** `LTE`

**Operands:** None

**Stack:** `[a, b] → [a <= b]`

**Cost:** 2

---

#### GTE (0x24)
Test greater than or equal.

**Format:** `GTE`

**Operands:** None

**Stack:** `[a, b] → [a >= b]`

**Cost:** 2

### Logical Operations

#### AND (0x30)
Logical AND.

**Format:** `AND`

**Operands:** None

**Stack:** `[a, b] → [a && b]`

**Cost:** 2

---

#### OR (0x31)
Logical OR.

**Format:** `OR`

**Operands:** None

**Stack:** `[a, b] → [a || b]`

**Cost:** 2

---

#### NOT (0x32)
Logical NOT.

**Format:** `NOT`

**Operands:** None

**Stack:** `[a] → [!a]`

**Cost:** 1

### Bitwise Operations

#### BAND (0x40)
Bitwise AND.

**Format:** `BAND`

**Operands:** None

**Stack:** `[a, b] → [a & b]`

**Cost:** 2

---

#### BOR (0x41)
Bitwise OR.

**Format:** `BOR`

**Operands:** None

**Stack:** `[a, b] → [a | b]`

**Cost:** 2

---

#### BXOR (0x42)
Bitwise XOR.

**Format:** `BXOR`

**Operands:** None

**Stack:** `[a, b] → [a ^ b]`

**Cost:** 2

---

#### BNOT (0x43)
Bitwise NOT.

**Format:** `BNOT`

**Operands:** None

**Stack:** `[a] → [~a]`

**Cost:** 1

---

#### SHL (0x44)
Shift left.

**Format:** `SHL`

**Operands:** None

**Stack:** `[a, b] → [a << b]`

**Cost:** 2

---

#### SHR (0x45)
Shift right.

**Format:** `SHR`

**Operands:** None

**Stack:** `[a, b] → [a >> b]`

**Cost:** 2

### Control Flow

#### JMP (0x50)
Unconditional jump.

**Format:** `JMP <offset>`

**Operands:** 2-byte signed offset

**Stack:** No change

**Cost:** 2

**Example:**
```assembly
JMP target
PUSH i64 1  ; Skipped

target:
PUSH i64 2  ; Executed
```

---

#### JMPIF (0x51)
Conditional jump if true.

**Format:** `JMPIF <offset>`

**Operands:** 2-byte signed offset

**Stack:** `[condition] → []`

**Cost:** 2

**Example:**
```assembly
PUSH bool true
JMPIF target
PUSH i64 1  ; Skipped

target:
PUSH i64 2  ; Executed
```

---

#### CALL (0x52)
Call a function.

**Format:** `CALL <function_id>`

**Operands:** 2-byte function ID

**Stack:** Depends on function

**Cost:** 5

---

#### RET (0x53)
Return from function.

**Format:** `RET`

**Operands:** None

**Stack:** `[return_value] → []` (pops to caller)

**Cost:** 2

### Memory Operations

#### LOAD (0x60)
Load local variable onto stack.

**Format:** `LOAD <var_index>`

**Operands:** 2-byte variable index

**Stack:** `[] → [value]`

**Cost:** 3

**Example:**
```assembly
; Assuming variable 0 contains 42
LOAD 0
; Stack: [42]
```

---

#### STORE (0x61)
Store top of stack to local variable.

**Format:** `STORE <var_index>`

**Operands:** 2-byte variable index

**Stack:** `[value] → []`

**Cost:** 3

**Example:**
```assembly
PUSH i64 42
STORE 0
; Variable 0 now contains 42
```

---

#### LOADG (0x62)
Load from global state.

**Format:** `LOADG <key>`

**Operands:** Variable-length key

**Stack:** `[] → [value]`

**Cost:** 10

---

#### STOREG (0x63)
Store to global state.

**Format:** `STOREG <key>`

**Operands:** Variable-length key

**Stack:** `[value] → []`

**Cost:** 15

### Blockchain Operations

#### GETFILE (0x70)
Get file from FileStore.

**Format:** `GETFILE`

**Operands:** None

**Stack:** `[file_id] → [file_data]`

**Cost:** 50

---

#### GETFILEMUT (0x71)
Get mutable file reference.

**Format:** `GETFILEMUT`

**Operands:** None

**Stack:** `[file_id] → [file_ref]`

**Cost:** 50

---

#### UPDATEFILE (0x72)
Update file in FileStore.

**Format:** `UPDATEFILE`

**Operands:** None

**Stack:** `[file_ref, data] → []`

**Cost:** 100

---

#### GETBALANCE (0x73)
Get file balance.

**Format:** `GETBALANCE`

**Operands:** None

**Stack:** `[file_id] → [balance]`

**Cost:** 30

---

#### UPDATEBALANCE (0x74)
Update file balance.

**SECURITY: This instruction can only be executed by the system program.** Regular programs attempting to call this will receive a security error.

**Format:** `UPDATEBALANCE`

**Operands:** None

**Stack:** `[file_id, delta] → []`

**Cost:** 80

**Note:** The `file_id` parameter accepts both `FileID` (32-byte identifier) and `i64` (converted to FileID by placing the value in the last 8 bytes, big-endian). This allows convenient use of integer account identifiers.

**Errors:**
- Security error if called by non-system program
- Type error if file_id is not FileID or i64, or if delta is not i64
- Runtime error if resulting balance would be negative

---

#### GETSIGNER (0x75)
Get transaction signer by index.

**Format:** `GETSIGNER`

**Operands:** None

**Stack:** `[index] → [pubkey]`

**Cost:** 10

---

#### HASSIGNER (0x76)
Check if pubkey is signer.

**Format:** `HASSIGNER`

**Operands:** None

**Stack:** `[pubkey] → [bool]`

**Cost:** 15

---

#### GETINSTRDATA (0x77)
Get instruction data.

**Format:** `GETINSTRDATA`

**Operands:** None

**Stack:** `[] → [data]`

**Cost:** 5

---

#### GETPROGRAMID (0x78)
Get current program ID.

**Format:** `GETPROGRAMID`

**Operands:** None

**Stack:** `[] → [program_id]`

**Cost:** 5

### Cryptographic Operations

#### SHA256 (0x80)
Compute SHA-256 hash.

**Format:** `SHA256`

**Operands:** None

**Stack:** `[data] → [hash]`

**Cost:** 50

---

#### VERIFYSIG (0x81)
Verify Ed25519 signature.

**Format:** `VERIFYSIG`

**Operands:** None

**Stack:** `[pubkey, message, signature] → [bool]`

**Cost:** 100

---

#### DERIVEPUBKEY (0x82)
Derive public key from seed.

**Format:** `DERIVEPUBKEY`

**Operands:** None

**Stack:** `[seed] → [pubkey]`

**Cost:** 80

### Cross-Program Invocation

#### INVOKE (0x90)
Invoke another program.

**Format:** `INVOKE`

**Operands:** None

**Stack:** `[program_id, data] → [result]`

**Cost:** 200 + invoked program cost

---

#### INVOKERET (0x91)
Return from cross-program call.

**Format:** `INVOKERET`

**Operands:** None

**Stack:** `[result] → []` (returns to caller)

**Cost:** 10

### Query Operations

#### QUERYBLOCK (0xA0)
Query finalized block.

**Format:** `QUERYBLOCK`

**Operands:** None

**Stack:** `[hash] → [block_data]`

**Cost:** 100

---

#### QUERYTX (0xA1)
Query finalized transaction.

**Format:** `QUERYTX`

**Operands:** None

**Stack:** `[tx_id] → [tx_data]`

**Cost:** 80

---

#### QUERYINSTR (0xA2)
Query finalized instruction.

**Format:** `QUERYINSTR`

**Operands:** None

**Stack:** `[instr_ref] → [instr_data]`

**Cost:** 80

## Opcode Table

| Opcode | Mnemonic | Category | Cost |
|--------|----------|----------|------|
| 0x01 | PUSH | Stack | 1 |
| 0x02 | POP | Stack | 1 |
| 0x03 | DUP | Stack | 1 |
| 0x04 | SWAP | Stack | 1 |
| 0x10 | ADD | Arithmetic | 2 |
| 0x11 | SUB | Arithmetic | 2 |
| 0x12 | MUL | Arithmetic | 3 |
| 0x13 | DIV | Arithmetic | 4 |
| 0x14 | MOD | Arithmetic | 4 |
| 0x20 | EQ | Comparison | 2 |
| 0x21 | LT | Comparison | 2 |
| 0x22 | GT | Comparison | 2 |
| 0x23 | LTE | Comparison | 2 |
| 0x24 | GTE | Comparison | 2 |
| 0x30 | AND | Logical | 2 |
| 0x31 | OR | Logical | 2 |
| 0x32 | NOT | Logical | 1 |
| 0x40 | BAND | Bitwise | 2 |
| 0x41 | BOR | Bitwise | 2 |
| 0x42 | BXOR | Bitwise | 2 |
| 0x43 | BNOT | Bitwise | 1 |
| 0x44 | SHL | Bitwise | 2 |
| 0x45 | SHR | Bitwise | 2 |
| 0x50 | JMP | Control Flow | 2 |
| 0x51 | JMPIF | Control Flow | 2 |
| 0x52 | CALL | Control Flow | 5 |
| 0x53 | RET | Control Flow | 2 |
| 0x60 | LOAD | Memory | 3 |
| 0x61 | STORE | Memory | 3 |
| 0x62 | LOADG | Memory | 10 |
| 0x63 | STOREG | Memory | 15 |
| 0x70 | GETFILE | Blockchain | 50 |
| 0x71 | GETFILEMUT | Blockchain | 50 |
| 0x72 | UPDATEFILE | Blockchain | 100 |
| 0x73 | GETBALANCE | Blockchain | 30 |
| 0x74 | UPDATEBALANCE | Blockchain | 80 |
| 0x75 | GETSIGNER | Blockchain | 10 |
| 0x76 | HASSIGNER | Blockchain | 15 |
| 0x77 | GETINSTRDATA | Blockchain | 5 |
| 0x78 | GETPROGRAMID | Blockchain | 5 |
| 0x80 | SHA256 | Crypto | 50 |
| 0x81 | VERIFYSIG | Crypto | 100 |
| 0x82 | DERIVEPUBKEY | Crypto | 80 |
| 0x90 | INVOKE | Invocation | 200+ |
| 0x91 | INVOKERET | Invocation | 10 |
| 0xA0 | QUERYBLOCK | Query | 100 |
| 0xA1 | QUERYTX | Query | 80 |
| 0xA2 | QUERYINSTR | Query | 80 |

## Execution Model

### Stack Machine

QuanticScript uses a stack-based execution model:

1. Instructions operate on a value stack
2. Operands are popped from the stack
3. Results are pushed onto the stack
4. Local variables are stored separately

### Execution Flow

```
1. Load bytecode
2. Initialize stack and memory
3. Set program counter to entry point
4. While budget > 0 and PC < bytecode length:
   a. Fetch instruction at PC
   b. Deduct instruction cost from budget
   c. Execute instruction
   d. Update PC
   e. Check safety limit (max 1000 steps)
5. Return result or error
```

### Safety Limits

The interpreter enforces a safety limit of 1000 execution steps to prevent infinite loops and ensure timely termination. If execution exceeds this limit, an error is returned:

```
execution exceeded safety limit of 1000 steps
```

This limit is independent of the compute budget and serves as a hard cap on execution time.

### Debug Logging

The interpreter supports optional debug logging for troubleshooting and development:

```go
interpreter := NewBytecodeInterpreter(bytecode, ctx, budget)
interpreter.SetDebugLogger(func(format string, args ...interface{}) {
    log.Printf(format, args...)
})
```

Debug logs include:
- Execution start/completion
- Each instruction step with PC, opcode name, and call stack state
- Error conditions with step number and PC location

Example debug output:
```
Execute: Starting execution, bytecode length=256
Execute: step=1 PC=0 opcode=PUSH callStack=[]
Execute: step=2 PC=10 opcode=CALL callStack=[]
Execute: step=3 PC=50 opcode=UPDATEBALANCE callStack=[14]
Execute: Error at step=3 PC=50: UPDATEBALANCE can only be called by the system program
```

### Error Conditions

- **OutOfComputeError**: Compute budget exhausted
- **StackUnderflowError**: Pop from empty stack
- **DivisionByZeroError**: Division or modulo by zero
- **InvalidOpcodeError**: Unknown opcode
- **InvalidOperandError**: Malformed operand
- **AccessViolationError**: Unauthorized file access
- **SafetyLimitError**: Execution exceeded 1000 steps
- **SecurityError**: Privileged instruction called by unauthorized program

## Examples

### Example 1: Simple Addition

**Assembly:**
```assembly
entry:
    PUSH i64 10
    PUSH i64 20
    ADD
    RET
```

**Bytecode (hex):**
```
51 53 43 42  ; Magic "QSCB"
01 00        ; Version 1.0
08 00        ; Entry at offset 8
01 01 0A 00 00 00 00 00 00 00  ; PUSH i64 10
01 01 14 00 00 00 00 00 00 00  ; PUSH i64 20
10           ; ADD
53           ; RET
```

### Example 2: Conditional Logic

**Assembly:**
```assembly
entry:
    PUSH i64 10
    PUSH i64 20
    LT
    JMPIF is_less
    
    PUSH i64 0
    RET
    
is_less:
    PUSH i64 1
    RET
```

### Example 3: Loop

**Assembly:**
```assembly
entry:
    PUSH i64 0      ; sum
    PUSH i64 0      ; i
    
loop_start:
    DUP             ; Duplicate i
    PUSH i64 10
    LT
    JMPIF loop_body
    
    SWAP            ; Get sum
    POP             ; Remove i
    RET
    
loop_body:
    SWAP            ; Get sum
    DUP             ; Duplicate i
    ADD             ; sum += i
    SWAP            ; Put sum back
    
    PUSH i64 1
    ADD             ; i++
    
    JMP loop_start
```

## Next Steps

- [Language Reference](language-reference.md) - High-level language
- [Inline Assembly Guide](inline-assembly.md) - Writing assembly
- [Cost Model Guide](cost-model.md) - Understanding costs
- [Examples](../examples/README.md) - Practical examples

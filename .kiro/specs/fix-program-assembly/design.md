# Design Document: Minimal Program Assembly via Dispatch Registry

## Overview

Replace the stub `.qsa`/`.qsb` files with production-ready bytecode. The guiding principle is **minimalism**: all parsing and dispatch complexity lives in Go (interpreter + stdlib); the assembly files are thin shells that call `DISPATCH` once and branch to handler labels.

## Architecture

```
Raw instruction bytes
    ↓
DISPATCH opcode (interpreter)
    ├── looks up InstructionDef by type code
    ├── validates data length against arg schema
    ├── parses all args (i64, u64, bytes, bool) with LE conversion
    └── pushes parsed args onto stack
    ↓
Assembly handler label
    ├── pops args from stack
    ├── calls blockchain opcodes (HASSIGNER, GETFILE, UPDATEBALANCE, etc.)
    └── RET with result code
    ↓
Assembler → .qsb → embedded in programs/embed.go → loaded at genesis
```

## Components

### 1. Instruction Dispatch Registry (`internal/quanticscript/instructions.go`)

```go
type ArgType int
const (
    ArgTypeI64   ArgType = iota
    ArgTypeU64
    ArgTypeBytes
    ArgTypeBool
)

type ArgDef struct {
    Name   string
    Type   ArgType
    Offset int    // byte offset in instruction data (after type byte)
    Length int    // only used for ArgTypeBytes
}

type InstructionDef struct {
    Code    int
    Name    string
    Args    []ArgDef
    Handler string // assembly label to jump to
}
```

**`ParseArgs(data []byte, def InstructionDef) (map[string]Value, error)`**
- Validates `len(data)` covers all arg offsets
- Reads each arg according to its `ArgDef` with little-endian conversion
- Returns typed `Value` map keyed by arg name

**`Dispatch(data []byte, registry map[int]InstructionDef) (InstructionDef, map[string]Value, error)`**
- Reads `data[0]` as instruction type code
- Looks up `registry[code]` — returns `ERROR_INVALID_INSTRUCTION` if missing
- Calls `ParseArgs` and returns result

### 2. `OpDispatch` Opcode

Added to `opcodes.go`:
```go
OpDispatch Opcode = 0xF0  // DISPATCH
```

Interpreter implementation (`interpreter.go`):
- Pops `bytes` from stack (raw instruction data)
- Calls `Dispatch` with the program's registered registry
- On success: pushes each parsed arg value onto the stack in schema order, then pushes handler name as string
- On error: pushes error code as i64 and returns

Cost (`costs.go`): base cost 10 + 2 per arg.

### 3. System_Program Registry

```go
var SystemProgramRegistry = map[int]InstructionDef{
    0: {Code: 0, Name: "CREATE_ACCOUNT", Handler: "handle_create_account", Args: []ArgDef{
        {Name: "owner",   Type: ArgTypeBytes, Offset: 1, Length: 32},
        {Name: "balance", Type: ArgTypeI64,   Offset: 33},
    }},
    1: {Code: 1, Name: "TRANSFER", Handler: "handle_transfer", Args: []ArgDef{
        {Name: "from",   Type: ArgTypeBytes, Offset: 1,  Length: 32},
        {Name: "to",     Type: ArgTypeBytes, Offset: 33, Length: 32},
        {Name: "amount", Type: ArgTypeI64,   Offset: 65},
    }},
    2: {Code: 2, Name: "ALLOCATE_SPACE", Handler: "handle_allocate_space", Args: []ArgDef{
        {Name: "account",    Type: ArgTypeBytes, Offset: 1, Length: 32},
        {Name: "extra_balance", Type: ArgTypeI64, Offset: 33},
    }},
}
```

### 4. Token_Program Registry

All 11 instruction types registered with their arg schemas:

| Code | Name | Args |
|------|------|------|
| 0 | INITIALIZE_MINT | decimals(u8@1), has_mint_auth(bool@2), mint_auth(bytes@3,32), has_freeze_auth(bool@35), freeze_auth(bytes@36,32) |
| 1 | INITIALIZE_ACCOUNT | mint(bytes@1,32), owner(bytes@33,32) |
| 2 | TRANSFER | from(bytes@1,32), to(bytes@33,32), amount(u64@65) |
| 3 | MINT_TO | mint(bytes@1,32), dest(bytes@33,32), amount(u64@65) |
| 4 | BURN | account(bytes@1,32), amount(u64@33) |
| 5 | CLOSE_ACCOUNT | account(bytes@1,32), dest(bytes@33,32) |
| 6 | FREEZE_ACCOUNT | account(bytes@1,32) |
| 7 | THAW_ACCOUNT | account(bytes@1,32) |
| 8 | APPROVE | account(bytes@1,32), delegate(bytes@33,32), amount(u64@65) |
| 9 | REVOKE | account(bytes@1,32) |
| 10 | CREATE_ASSOCIATED_TOKEN_ACCOUNT | owner(bytes@1,32), mint(bytes@33,32) |

### 5. Assembly Structure (minimal)

**system.qsa** — illustrative skeleton:
```asm
; System_Program entry point
entry:
    GETINSTRDATA
    DISPATCH                    ; pops bytes, pushes (handler_name, arg0, arg1, ...)
    JMPTABLE                    ; jumps to handler label from stack

handle_create_account:
    ; stack: [owner: bytes, balance: i64]
    STORE 0                     ; balance
    STORE 1                     ; owner
    LOAD 1
    HASSIGNER
    JMPIF auth_ok
    PUSH i64 0x1004             ; ERROR_UNAUTHORIZED_SIGNER
    RET
auth_ok:
    LOAD 1
    LOAD 0
    CREATEFILE                  ; (owner, balance) → fileID
    PUSH i64 0
    RET

handle_transfer:
    ; stack: [from: bytes, to: bytes, amount: i64]
    ; ... balance checks, HASSIGNER, UPDATEBALANCE calls
    PUSH i64 0
    RET

handle_allocate_space:
    ; stack: [account: bytes, extra_balance: i64]
    ; ... storage cost check, UPDATEBALANCE
    PUSH i64 0
    RET
```

**token.qsa** follows the same pattern — `DISPATCH` at entry, one label per instruction, blockchain opcodes only.

## Data Models

**MintAccount** (FileStore data field):
```
[0:8]   supply        i64 LE
[8]     decimals      u8
[9]     has_mint_auth bool
[10:42] mint_auth     bytes (if has_mint_auth)
[42]    has_freeze_auth bool
[43:75] freeze_auth   bytes (if has_freeze_auth)
```

**TokenAccount** (FileStore data field):
```
[0:32]  mint           bytes
[32:64] owner          bytes
[64:72] token_balance  i64 LE
[72]    has_delegate   bool
[73:105] delegate      bytes (if has_delegate)
[105:113] delegated_amount i64 LE
[113]   frozen         bool
```

## Error Codes

| Code | Name |
|------|------|
| 0x1000 | ERROR_INSUFFICIENT_BALANCE |
| 0x1001 | ERROR_INVALID_ACCOUNT |
| 0x1002 | ERROR_BALANCE_OVERFLOW |
| 0x1003 | ERROR_STORAGE_RENT_VIOLATION |
| 0x1004 | ERROR_UNAUTHORIZED_SIGNER |
| 0x1FFF | ERROR_INVALID_INSTRUCTION (System) |
| 0x2000 | ERROR_INSUFFICIENT_TOKEN_BALANCE |
| 0x2001 | ERROR_INVALID_MINT |
| 0x2002 | ERROR_INVALID_TOKEN_ACCOUNT |
| 0x2003 | ERROR_MINT_MISMATCH |
| 0x2004 | ERROR_UNAUTHORIZED_MINT_AUTHORITY |
| 0x2005 | ERROR_UNAUTHORIZED_FREEZE_AUTHORITY |
| 0x2006 | ERROR_UNAUTHORIZED_OWNER |
| 0x2007 | ERROR_ACCOUNT_FROZEN |
| 0x2008 | ERROR_ACCOUNT_NOT_EMPTY |
| 0x2009 | ERROR_DELEGATE_NOT_SET |
| 0x200A | ERROR_INSUFFICIENT_DELEGATED_AMOUNT |
| 0x200B | ERROR_FIXED_SUPPLY_MINT |
| 0x2FFF | ERROR_INVALID_INSTRUCTION (Token) |

## Error Handling

- All validation happens before any state change
- `DISPATCH` returns error code on bad data — assembly never sees malformed args
- Handler labels only run with guaranteed-valid, typed args on the stack
- No panics, no state corruption on any error path

## Testing Strategy

- Unit tests for `ParseArgs` and `Dispatch` in Go covering all arg types and error paths
- Bytecode execution tests for each System_Program and Token_Program instruction
- Determinism test: same inputs → identical outputs across multiple runs
- Error code tests: every error condition returns the correct code
- Genesis loading test: bytecode embeds and loads without errors

## Bytecode Generation

```bash
go run cmd/main.go qsc assemble -i programs/system/system.qsa -o programs/system/system.qsb
go run cmd/main.go qsc assemble -i programs/token/token.qsa -o programs/token/token.qsb
```

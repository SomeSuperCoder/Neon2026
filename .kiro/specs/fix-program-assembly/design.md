# Design Document: Minimal Program Assembly via Dispatch Registry

## Overview

Replace the stub `.qs`/`.qsa`/`.qsb` files with production-ready programs. The guiding principle is **source-first**: the `.qs` file is the single source of truth. All parsing and dispatch complexity lives in Go (interpreter + stdlib); the `.qs` source files are thin shells that call `DISPATCH` once and branch to handler functions. Assembly and bytecode are always compiler outputs.

## Architecture

```
programs/system/system.qs   (source of truth)
programs/token/token.qs     (source of truth)
    ↓  go run cmd/main.go qsc compile
programs/system/system.qsa  (compiler output)
programs/token/token.qsa    (compiler output)
    ↓  go run cmd/main.go qsc assemble
programs/system/system.qsb  (compiler output)
programs/token/token.qsb    (compiler output)
    ↓  programs/embed.go
Loaded at genesis

Runtime flow:
Raw instruction bytes
    ↓
DISPATCH opcode (interpreter)
    ├── looks up InstructionDef by type code
    ├── validates data length against arg schema
    ├── parses all args (i64, u64, bytes, bool) with LE conversion
    └── pushes parsed args onto stack
    ↓
.qs handler function
    ├── receives args as typed parameters
    ├── calls blockchain stdlib (hasSigner, getFile, updateBalance, etc.)
    └── returns result code
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

### 5. QuanticScript Source Structure

**system.qs** — illustrative skeleton:
```typescript
// System_Program entry point
fn main(): i64 {
    let instrData = getInstrData();
    let (handlerName, args) = dispatch(instrData);
    if handlerName == "CREATE_ACCOUNT" {
        return handleCreateAccount(args["owner"], args["balance"]);
    } else if handlerName == "TRANSFER" {
        return handleTransfer(args["from"], args["to"], args["amount"]);
    } else if handlerName == "ALLOCATE_SPACE" {
        return handleAllocateSpace(args["account"], args["extra_balance"]);
    }
    return 0x1FFF; // ERROR_INVALID_INSTRUCTION
}

fn handleCreateAccount(owner: bytes, balance: i64): i64 {
    if !hasSigner(owner) { return 0x1004; }
    createFile(owner, balance);
    return 0;
}

fn handleTransfer(from: bytes, to: bytes, amount: i64): i64 {
    if !hasSigner(from) { return 0x1004; }
    // balance checks, updateBalance calls
    return 0;
}

fn handleAllocateSpace(account: bytes, extraBalance: i64): i64 {
    if !hasSigner(account) { return 0x1004; }
    // storage cost check, updateBalance
    return 0;
}
```

**token.qs** follows the same pattern — `dispatch()` at entry, one handler function per instruction, stdlib calls only.

The `.qsa` and `.qsb` files are always compiler outputs — never hand-written.

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
- `DISPATCH` returns error code on bad data — `.qs` handler functions never see malformed args
- Handler functions only run with guaranteed-valid, typed args
- No panics, no state corruption on any error path

## Testing Strategy

- Unit tests for `ParseArgs` and `Dispatch` in Go covering all arg types and error paths
- Bytecode execution tests for each System_Program and Token_Program instruction
- Determinism test: same inputs → identical outputs across multiple runs
- Error code tests: every error condition returns the correct code
- Genesis loading test: bytecode embeds and loads without errors

## Bytecode Generation

```bash
# Compile .qs source to assembly
go run cmd/main.go qsc compile -i programs/system/system.qs -o programs/system/system.qsa
go run cmd/main.go qsc compile -i programs/token/token.qs -o programs/token/token.qsa

# Assemble to bytecode
go run cmd/main.go qsc assemble -i programs/system/system.qsa -o programs/system/system.qsb
go run cmd/main.go qsc assemble -i programs/token/token.qsa -o programs/token/token.qsb
```

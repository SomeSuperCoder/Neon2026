# Design Document

## Overview

This design implements a secure, general-purpose balance transfer system for the PoH blockchain. The core innovation is replacing the restricted `updateBalance` opcode with a new `transfer` opcode that any program can use to move balance from files it owns to any destination file, while maintaining storage cost invariants.

## Architecture

### Component Interaction

```
┌─────────────────┐
│  Smart Contract │
│   (any program) │
└────────┬────────┘
         │ calls transfer opcode
         ▼
┌─────────────────┐
│   Interpreter   │◄──────────────┐
│  (execTransfer) │               │
└────────┬────────┘               │
         │                        │
         │ validates ownership    │
         │ checks storage costs   │
         │                        │
         ▼                        │
┌─────────────────┐               │
│ ExecutionContext│               │
│  (GetFile,      │               │
│   UpdateFile)   │               │
└────────┬────────┘               │
         │                        │
         ▼                        │
┌─────────────────┐               │
│   FileStore     │               │
│  (balance mgmt) │               │
└─────────────────┘               │
                                  │
┌─────────────────┐               │
│ System Program  │───────────────┘
│ (uses transfer  │  uses transfer for
│  for account    │  account operations
│  operations)    │
└─────────────────┘
```

### Key Design Decisions

1. **Ownership-Based Security**: Transfer validation is based on the TxManager field of the source file. Only the program that owns a file (TxManager == ProgramID) can transfer from it.

2. **Storage Cost Enforcement**: Before any transfer, we calculate the minimum balance required to sustain the source file's data. The transfer is only allowed if: `source.Balance - amount >= CalculateStorageCost(len(source.Data))`

3. **Atomic Operations**: All balance updates happen atomically within a single transaction. If any validation fails, no state changes occur.

4. **Language Extension**: The QuanticScript language will be extended with a `transfer()` builtin function that maps to the new opcode.

## Components and Interfaces

### 1. Transfer Opcode (OpTransfer)

**Location**: `internal/quanticscript/opcodes.go`

```go
const (
    OpTransfer Opcode = 0x7A // Transfer balance between files
)
```

**Stack Behavior**:
- Input: `[sourceFileID (FileID), destFileID (FileID), amount (i64)]`
- Output: `[]` (no return value, errors on failure)

**Validation Steps**:
1. Pop amount, destFileID, sourceFileID from stack
2. Validate amount > 0
3. Get source file from context
4. Verify ctx.GetProgramID() == source.TxManager (ownership check)
5. Calculate required storage cost for source file
6. Verify source.Balance - amount >= storageCost
7. Get destination file from context
8. Check for overflow: dest.Balance + amount <= MAX_I64
9. Update source.Balance -= amount
10. Update dest.Balance += amount
11. Persist both files atomically

### 2. Interpreter Implementation

**Location**: `internal/quanticscript/interpreter.go`

**New Method**: `execTransfer()`

```go
func (bi *BytecodeInterpreter) execTransfer() error {
    // Pop amount from stack
    amountValue, err := bi.pop()
    if err != nil {
        return err
    }
    if amountValue.Type != TypeI64 {
        return fmt.Errorf("TRANSFER requires i64 for amount")
    }
    amount, _ := amountValue.AsI64()
    
    // Validate amount is positive
    if amount <= 0 {
        return fmt.Errorf("TRANSFER amount must be positive, got %d", amount)
    }
    
    // Pop destination FileID
    destValue, err := bi.pop()
    if err != nil {
        return err
    }
    destFileID, err := valueToFileID(destValue)
    if err != nil {
        return fmt.Errorf("invalid destination FileID: %w", err)
    }
    
    // Pop source FileID
    sourceValue, err := bi.pop()
    if err != nil {
        return err
    }
    sourceFileID, err := valueToFileID(sourceValue)
    if err != nil {
        return fmt.Errorf("invalid source FileID: %w", err)
    }
    
    // Get source file with write permission
    sourceFile, err := bi.ctx.GetFileMut(sourceFileID)
    if err != nil {
        return fmt.Errorf("failed to get source file: %w", err)
    }
    
    // Ownership check: only the program that owns the file can transfer from it
    currentProgramID := bi.ctx.GetProgramID()
    if sourceFile.TxManager != currentProgramID {
        return fmt.Errorf("TRANSFER denied: program %s does not own file %s (owned by %s)",
            currentProgramID.String(), sourceFileID.String(), sourceFile.TxManager.String())
    }
    
    // Calculate storage cost for source file
    storageCost := filestore.CalculateStorageCost(int64(len(sourceFile.Data)))
    
    // Check if transfer would violate storage cost constraint
    newSourceBalance := sourceFile.Balance - amount
    if newSourceBalance < storageCost {
        return fmt.Errorf("TRANSFER would violate storage cost: source balance would be %d, required %d",
            newSourceBalance, storageCost)
    }
    
    // Get destination file with write permission
    destFile, err := bi.ctx.GetFileMut(destFileID)
    if err != nil {
        return fmt.Errorf("failed to get destination file: %w", err)
    }
    
    // Check for overflow on destination
    const MAX_I64 = 9223372036854775807
    if destFile.Balance > MAX_I64 - amount {
        return fmt.Errorf("TRANSFER would cause overflow on destination")
    }
    
    // Perform the transfer atomically
    sourceFile.Balance -= amount
    destFile.Balance += amount
    
    // Update both files
    if err := bi.ctx.UpdateFile(sourceFile); err != nil {
        return fmt.Errorf("failed to update source file: %w", err)
    }
    if err := bi.ctx.UpdateFile(destFile); err != nil {
        return fmt.Errorf("failed to update destination file: %w", err)
    }
    
    return nil
}
```

### 3. Standard Library Extension

**Location**: `internal/quanticscript/stdlib.go`

**New Builtin Function**: `transfer()`

```go
// transfer moves balance from source file to destination file
// Only works if the calling program owns the source file
// Signature: transfer(sourceFileID: FileID, destFileID: FileID, amount: i64) -> void
func builtinTransfer(args []Value) (Value, error) {
    if len(args) != 3 {
        return Value{}, fmt.Errorf("transfer requires 3 arguments")
    }
    
    // Validation happens in the interpreter's execTransfer
    // This is just a wrapper that emits the opcode
    
    return NewVoid(), nil
}
```

**Compiler Integration**: Update `codegen.go` to recognize `transfer()` calls and emit `OpTransfer` opcode.

### 4. Language Extension: Transfer Statement

**Location**: `internal/quanticscript/lexer.go`, `parser.go`, `ast.go`, `codegen.go`

**New Syntax**:
```typescript
transfer(sourceFileID, destFileID, amount);
```

**Implementation Steps**:

1. **Lexer**: Already supports function call syntax, no changes needed

2. **Parser**: Already supports function calls, no changes needed

3. **AST**: Use existing `CallExpression` node

4. **Type Checker**: Add validation for `transfer()` builtin
   - Verify 3 arguments
   - Verify types: (FileID, FileID, i64)
   - Return type: void

5. **Code Generator**: Recognize `transfer()` and emit:
   ```
   PUSH sourceFileID
   PUSH destFileID  
   PUSH amount
   TRANSFER
   ```

### 5. System Program Rewrite

**Location**: `programs/system/system.qs`

**Updated Implementation**:

```typescript
// System_Program: Built-in program for account management
// Uses the new transfer() instruction for all balance operations

// Error codes
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_INVALID_ACCOUNT: i64 = 0x1001;
const ERROR_BALANCE_OVERFLOW: i64 = 0x1002;
const ERROR_STORAGE_RENT_VIOLATION: i64 = 0x1003;
const ERROR_UNAUTHORIZED_SIGNER: i64 = 0x1004;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;
const ERROR_INVALID_AMOUNT: i64 = 0x1005;

const SUCCESS: i64 = 0;
const MAX_I64: i64 = 9223372036854775807;

// Instruction type codes
const INSTR_CREATE_ACCOUNT: u8 = 0;
const INSTR_TRANSFER: u8 = 1;
const INSTR_ALLOCATE_SPACE: u8 = 2;

// Entry point with instruction dispatch
export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    
    if (len(instrData) == 0) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let instrType: u8 = instrData[0];
    
    if (instrType == INSTR_CREATE_ACCOUNT) {
        return handleCreateAccount(instrData);
    } else if (instrType == INSTR_TRANSFER) {
        return handleTransfer(instrData);
    } else if (instrType == INSTR_ALLOCATE_SPACE) {
        return handleAllocateSpace(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// CREATE_ACCOUNT: Creates a new account file
// Format: [type:u8][owner:FileID(32)][balance:i64(8)]
function handleCreateAccount(instrData: bytes): i64 {
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Parse owner FileID (bytes 1-32)
    let owner: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Parse initial balance (bytes 33-40, little-endian)
    let balance: i64 = bytesToI64LE(slice(instrData, 33, 41));
    
    if (balance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Create the file by transferring balance from system program's file
    // to the new account file
    let systemFileID: FileID = getProgramID();
    transfer(systemFileID, owner, balance);
    
    return SUCCESS;
}

// TRANSFER: Transfers balance between two accounts
// Format: [type:u8][from:FileID(32)][to:FileID(32)][amount:i64(8)]
function handleTransfer(instrData: bytes): i64 {
    if (len(instrData) != 73) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Parse from FileID (bytes 1-32)
    let from: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Parse to FileID (bytes 33-64)
    let to: FileID = bytesToFileID(slice(instrData, 33, 65));
    
    // Parse amount (bytes 65-72, little-endian)
    let amount: i64 = bytesToI64LE(slice(instrData, 65, 73));
    
    if (amount <= 0) {
        return ERROR_INVALID_AMOUNT;
    }
    
    // Use the transfer instruction
    // This will automatically validate ownership and storage costs
    transfer(from, to, amount);
    
    return SUCCESS;
}

// ALLOCATE_SPACE: Adds balance to an account for storage rent
// Format: [type:u8][account:FileID(32)][extraBalance:i64(8)]
function handleAllocateSpace(instrData: bytes): i64 {
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Parse account FileID (bytes 1-32)
    let account: FileID = bytesToFileID(slice(instrData, 1, 33));
    
    // Parse extra balance (bytes 33-40, little-endian)
    let extraBalance: i64 = bytesToI64LE(slice(instrData, 33, 41));
    
    if (extraBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Transfer from system program to the account
    let systemFileID: FileID = getProgramID();
    transfer(systemFileID, account, extraBalance);
    
    return SUCCESS;
}
```

**Language Extensions Needed**:

1. **Array indexing**: `instrData[0]` - Already supported
2. **Slice function**: `slice(bytes, start, end)` - Need to add
3. **len function**: `len(bytes)` - Already supported  
4. **bytesToFileID**: Convert bytes to FileID - Need to add
5. **bytesToI64LE**: Convert bytes to i64 little-endian - Already have `OpBytesToI64LE`
6. **Comparison operators**: Already supported

### 6. Helper Functions to Add

**Location**: `internal/quanticscript/stdlib.go`

```go
// slice extracts a byte slice from start (inclusive) to end (exclusive)
func builtinSlice(args []Value) (Value, error) {
    if len(args) != 3 {
        return Value{}, fmt.Errorf("slice requires 3 arguments")
    }
    
    data, err := args[0].AsBytes()
    if err != nil {
        return Value{}, fmt.Errorf("slice: first argument must be bytes")
    }
    
    start, err := args[1].AsI64()
    if err != nil {
        return Value{}, fmt.Errorf("slice: second argument must be i64")
    }
    
    end, err := args[2].AsI64()
    if err != nil {
        return Value{}, fmt.Errorf("slice: third argument must be i64")
    }
    
    if start < 0 || end < 0 || start > end || end > int64(len(data)) {
        return Value{}, fmt.Errorf("slice: invalid range [%d:%d] for length %d", start, end, len(data))
    }
    
    result := data[start:end]
    return NewBytes(result), nil
}

// bytesToFileID converts a 32-byte slice to FileID
func builtinBytesToFileID(args []Value) (Value, error) {
    if len(args) != 1 {
        return Value{}, fmt.Errorf("bytesToFileID requires 1 argument")
    }
    
    data, err := args[0].AsBytes()
    if err != nil {
        return Value{}, fmt.Errorf("bytesToFileID: argument must be bytes")
    }
    
    if len(data) != 32 {
        return Value{}, fmt.Errorf("bytesToFileID: requires exactly 32 bytes, got %d", len(data))
    }
    
    return NewFileID(data), nil
}
```

## Data Models

### File Structure (Existing)

```go
type File struct {
    ID         FileID    // 32-byte unique identifier
    Balance    int64     // Economic units held by this file
    TxManager  FileID    // Program that owns/manages this file
    Data       []byte    // Arbitrary data stored in the file
    Executable bool      // Whether this file contains executable code
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

**Key Fields for Transfer**:
- `Balance`: Modified by transfer operations
- `TxManager`: Used for ownership validation
- `Data`: Length determines storage cost requirement

### Transfer Validation Model

```
Transfer Request:
  source: FileID
  dest: FileID
  amount: i64

Validation:
  1. amount > 0
  2. source.TxManager == currentProgramID
  3. source.Balance - amount >= CalculateStorageCost(len(source.Data))
  4. dest.Balance + amount <= MAX_I64

Execution:
  source.Balance -= amount
  dest.Balance += amount
  UpdateFile(source)
  UpdateFile(dest)
```

## Error Handling

### Transfer Opcode Errors

| Error | Condition | Message |
|-------|-----------|---------|
| Invalid Amount | amount <= 0 | "TRANSFER amount must be positive, got {amount}" |
| Ownership Violation | source.TxManager != programID | "TRANSFER denied: program {pid} does not own file {fid}" |
| Storage Cost Violation | newBalance < storageCost | "TRANSFER would violate storage cost: balance would be {bal}, required {cost}" |
| Overflow | dest.Balance + amount > MAX_I64 | "TRANSFER would cause overflow on destination" |
| File Not Found | GetFile fails | "failed to get source/destination file: {error}" |

### System Program Errors

| Error Code | Name | Condition |
|------------|------|-----------|
| 0x1000 | ERROR_INSUFFICIENT_BALANCE | Negative balance or amount |
| 0x1001 | ERROR_INVALID_ACCOUNT | File not found |
| 0x1002 | ERROR_BALANCE_OVERFLOW | Overflow detected |
| 0x1003 | ERROR_STORAGE_RENT_VIOLATION | Storage cost not covered |
| 0x1004 | ERROR_UNAUTHORIZED_SIGNER | Signer validation failed |
| 0x1005 | ERROR_INVALID_AMOUNT | Amount <= 0 |
| 0x1FFF | ERROR_INVALID_INSTRUCTION | Unknown instruction type |

## Testing Strategy

### Unit Tests

1. **Interpreter Tests** (`internal/quanticscript/interpreter_test.go`)
   - Test `execTransfer()` with valid ownership
   - Test ownership violation rejection
   - Test storage cost enforcement
   - Test negative amount rejection
   - Test overflow protection
   - Test FileID type conversions (i64, bytes, FileID)

2. **System Program Tests** (`internal/quanticscript/system_program_test.go`)
   - Test CREATE_ACCOUNT instruction
   - Test TRANSFER instruction
   - Test ALLOCATE_SPACE instruction
   - Test error code returns
   - Test instruction parsing

3. **Stdlib Tests** (`internal/quanticscript/stdlib_test.go`)
   - Test `slice()` function
   - Test `bytesToFileID()` function
   - Test edge cases and error conditions

### Integration Tests

1. **End-to-End Transfer** (`internal/quanticscript/transfer_integration_test.go`)
   - Create two files owned by a program
   - Transfer balance between them
   - Verify final balances
   - Verify storage costs are respected

2. **Cross-Program Scenarios**
   - Program A owns File X
   - Program B tries to transfer from File X (should fail)
   - Program A transfers from File X to File Y (should succeed)

3. **System Program Integration** (`internal/system_program_integration_test.go`)
   - Test account creation flow
   - Test transfer flow
   - Test space allocation flow
   - Verify all operations use the new transfer opcode

### Security Tests

1. **Negative Transfer Attack**
   - Attempt transfer with amount = -1000
   - Verify rejection

2. **Ownership Bypass Attack**
   - Program A tries to transfer from file owned by Program B
   - Verify rejection

3. **Storage Cost Bypass Attack**
   - File has 1MB data (requires X balance)
   - Attempt to transfer balance below X
   - Verify rejection

4. **Overflow Attack**
   - Destination has balance = MAX_I64 - 100
   - Attempt to transfer 200
   - Verify rejection

## Migration Plan

### Phase 1: Add Transfer Opcode
1. Add `OpTransfer` to opcodes.go
2. Implement `execTransfer()` in interpreter.go
3. Add unit tests for transfer opcode
4. Verify all tests pass

### Phase 2: Extend Language
1. Add `slice()` builtin function
2. Add `bytesToFileID()` builtin function
3. Update stdlib.go with new functions
4. Add tests for new functions
5. Update code generator to recognize `transfer()` calls

### Phase 3: Rewrite System Program
1. Update system.qs with new implementation
2. Compile to system.qsa and system.qsb
3. Add comprehensive tests
4. Verify all System Program operations work

### Phase 4: Deprecate UpdateBalance
1. Mark `OpUpdateBalance` as deprecated
2. Update error message to suggest using transfer
3. Remove all uses of updateBalance from codebase
4. Eventually remove the opcode entirely

### Phase 5: Integration Testing
1. Run full test suite
2. Test with demo scripts
3. Verify BFT scenarios still work
4. Performance testing

## Performance Considerations

### Transfer Operation Cost

```
Cost Breakdown:
- Stack operations (3 pops): 3 * 10 = 30 units
- File reads (2 GetFileMut): 2 * 1000 = 2000 units
- Ownership check: 10 units
- Storage cost calculation: 50 units
- Balance arithmetic: 20 units
- File writes (2 UpdateFile): 2 * 2000 = 4000 units
Total: ~6110 units
```

This is reasonable for a critical operation and similar to other blockchain operations.

### Optimization Opportunities

1. **Cache Storage Costs**: If file data doesn't change, cache the storage cost calculation
2. **Batch Transfers**: Future enhancement to transfer to multiple destinations in one operation
3. **Balance-Only Updates**: Optimize UpdateFile to only persist balance changes when data is unchanged

## Security Model

### Trust Boundaries

1. **Program Ownership**: Programs can only transfer from files they own (TxManager check)
2. **Storage Invariant**: Balance must always cover storage costs
3. **Arithmetic Safety**: All operations check for overflow/underflow
4. **Atomicity**: Transfers are all-or-nothing (both files updated or neither)

### Attack Vectors Mitigated

1. **Unauthorized Transfer**: Blocked by TxManager validation
2. **Negative Transfer**: Blocked by amount > 0 check
3. **Storage Cost Bypass**: Blocked by storage cost validation
4. **Integer Overflow**: Blocked by MAX_I64 check
5. **Reentrancy**: Not applicable (no callbacks during transfer)

## Backward Compatibility

### Breaking Changes

1. `OpUpdateBalance` is deprecated and will be removed
2. System Program bytecode changes (new instruction format)
3. Programs using `updateBalance()` must be rewritten to use `transfer()`

### Migration Path

1. Update all builtin programs to use `transfer()`
2. Provide clear error messages for deprecated opcodes
3. Document migration guide for smart contract developers
4. Maintain deprecated opcode for one release cycle with warnings

## Future Enhancements

1. **Batch Transfers**: Transfer to multiple destinations atomically
2. **Conditional Transfers**: Transfer only if certain conditions are met
3. **Transfer Hooks**: Allow programs to register callbacks on balance changes
4. **Gas Refunds**: Return unused compute budget as balance
5. **Transfer Limits**: Rate limiting or maximum transfer amounts per transaction

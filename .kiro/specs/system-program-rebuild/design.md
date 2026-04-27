# Design Document

## Overview

The System Program will be rebuilt as a clean, production-ready QuanticScript program focused on four core operations: creating files, transferring Neon, burning Neon, and closing files. The design leverages existing blockchain opcodes (OpTransfer, OpCreateFile, OpDeleteFile) and follows the established instruction dispatch pattern used in other programs.

The System Program will be the sole owner of files it creates, enforcing strict ownership validation before any operation. All operations will respect storage cost constraints to maintain blockchain integrity.

## Architecture

### Program Structure

```
System Program (FileID: 0x00...01)
├── Entry Point (entry function)
├── Instruction Dispatcher
├── CREATE_FILE Handler
├── TRANSFER Handler
├── BURN Handler
└── CLOSE_FILE Handler
```

### Instruction Format

All instructions follow a consistent format:
- First byte: instruction type (u8)
- Remaining bytes: instruction-specific parameters

### Error Handling

The program returns i64 error codes:
- `0x00`: SUCCESS
- `0x1000`: ERROR_INSUFFICIENT_BALANCE
- `0x1004`: ERROR_UNAUTHORIZED
- `0x1005`: ERROR_INVALID_AMOUNT
- `0x1006`: ERROR_FILE_HAS_DATA
- `0x1FFF`: ERROR_INVALID_INSTRUCTION

## Components and Interfaces

### 1. Entry Point

**Function**: `entry(): i64`

**Responsibilities**:
- Retrieve instruction data using `getInstructionData()`
- Validate minimum instruction length (at least 1 byte for type)
- Extract instruction type from first byte
- Dispatch to appropriate handler
- Return error code or SUCCESS

**Pseudocode**:
```typescript
export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    
    if (len(instrData) < 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let instrType: i64 = instrData[0];
    
    if (instrType == INSTR_CREATE_FILE) {
        return handleCreateFile(instrData);
    }
    if (instrType == INSTR_TRANSFER) {
        return handleTransfer(instrData);
    }
    if (instrType == INSTR_BURN) {
        return handleBurn(instrData);
    }
    if (instrType == INSTR_CLOSE_FILE) {
        return handleCloseFile(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}
```

### 2. CREATE_FILE Handler

**Function**: `handleCreateFile(instrData: bytes): i64`

**Instruction Format**: `[type:u8(1)][owner:FileID(32)][initial_balance:i64(8)]` = 41 bytes

**Responsibilities**:
- Validate instruction length is exactly 41 bytes
- Extract owner FileID (bytes 1-33)
- Extract initial balance (bytes 33-41, little-endian)
- Validate initial balance is non-negative
- Get System Program's FileID using `getProgramID()`
- Check System Program has sufficient balance (including storage cost)
- Create new file using `OpCreateFile` with TxManager set to System Program
- Transfer initial balance from System Program to new file using `transfer()`
- Return SUCCESS or error code

**Key Operations**:
- Uses `bytesToFileID()` to convert owner bytes to FileID
- Uses `bytesToI64LE()` to decode balance
- Uses `transfer(systemFileID, ownerFileID, initial_balance)` for balance transfer
- Relies on OpTransfer's built-in storage cost validation

### 3. TRANSFER Handler

**Function**: `handleTransfer(instrData: bytes): i64`

**Instruction Format**: `[type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]` = 73 bytes

**Responsibilities**:
- Validate instruction length is exactly 73 bytes
- Extract from FileID (bytes 1-33)
- Extract to FileID (bytes 33-65)
- Extract amount (bytes 65-73, little-endian)
- Validate amount is positive (> 0)
- Verify from file is owned by System Program using `getFile()` and checking TxManager
- Use `transfer(fromFileID, toFileID, amount)` to perform transfer
- Return SUCCESS or error code

**Ownership Validation**:
```typescript
let fromFile: File = getFile(fromFileID);
let systemID: FileID = getProgramID();

if (fromFile.txManager != systemID) {
    return ERROR_UNAUTHORIZED;
}
```

**Key Operations**:
- OpTransfer automatically validates storage costs
- OpTransfer automatically validates ownership at the interpreter level
- Double validation (QuanticScript + interpreter) provides defense in depth

### 4. BURN Handler

**Function**: `handleBurn(instrData: bytes): i64`

**Instruction Format**: `[type:u8(1)][file:FileID(32)][amount:i64(8)]` = 41 bytes

**Responsibilities**:
- Validate instruction length is exactly 41 bytes
- Extract file FileID (bytes 1-33)
- Extract burn amount (bytes 33-41, little-endian)
- Validate amount is positive (> 0)
- Verify file is owned by System Program
- Get file using `getFile()` to check current balance and data size
- Calculate storage cost using file's data size
- Verify remaining balance after burn covers storage cost
- Reduce file balance by burn amount (balance -= amount)
- Update file using `updateFile()`
- Return SUCCESS or error code

**Storage Cost Validation**:
```typescript
let file: File = getFile(fileID);
let storageCost: i64 = calculateStorageCost(len(file.data));
let newBalance: i64 = file.balance - amount;

if (newBalance < storageCost) {
    return ERROR_INSUFFICIENT_BALANCE;
}
```

**Key Operations**:
- Manual balance reduction (not a transfer)
- Explicit storage cost check before modification
- Uses `updateFile()` to persist changes

### 5. CLOSE_FILE Handler

**Function**: `handleCloseFile(instrData: bytes): i64`

**Instruction Format**: `[type:u8(1)][file_to_close:FileID(32)][destination:FileID(32)]` = 65 bytes

**Responsibilities**:
- Validate instruction length is exactly 65 bytes
- Extract file_to_close FileID (bytes 1-33)
- Extract destination FileID (bytes 33-65)
- Verify file is owned by System Program
- Get file using `getFile()`
- Verify file has zero data length
- Transfer all remaining balance to destination using `transfer()`
- Delete file using `OpDeleteFile`
- Return SUCCESS or error code

**Data Length Validation**:
```typescript
let file: File = getFile(fileToClose);

if (len(file.data) > 0) {
    return ERROR_FILE_HAS_DATA;
}
```

**Key Operations**:
- Uses `transfer()` to move remaining balance
- Uses `OpDeleteFile` to remove file from FileStore
- Two-step process ensures balance is preserved before deletion

## Data Models

### File Structure (from filestore)

```go
type File struct {
    ID         FileID    // 32-byte unique identifier
    Balance    int64     // Neon balance
    TxManager  FileID    // Owner program ID
    Data       []byte    // File data
    Executable bool      // Whether file is executable
    CreatedAt  time.Time // Creation timestamp
    UpdatedAt  time.Time // Last update timestamp
}
```

### Instruction Types

```typescript
const INSTR_CREATE_FILE: i64 = 0;
const INSTR_TRANSFER: i64 = 1;
const INSTR_BURN: i64 = 2;
const INSTR_CLOSE_FILE: i64 = 3;
```

### Error Codes

```typescript
const SUCCESS: i64 = 0x00;
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_UNAUTHORIZED: i64 = 0x1004;
const ERROR_INVALID_AMOUNT: i64 = 0x1005;
const ERROR_FILE_HAS_DATA: i64 = 0x1006;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;
```

## Error Handling

### Validation Layers

1. **Instruction Format Validation**
   - Length checks for each instruction type
   - Performed before any state reads

2. **Parameter Validation**
   - Amount positivity checks
   - Balance non-negativity checks
   - Performed after parsing, before state modifications

3. **Ownership Validation**
   - TxManager checks against System Program ID
   - Performed before any file modifications
   - Enforced at both QuanticScript and interpreter levels

4. **Storage Cost Validation**
   - Automatic validation by OpTransfer
   - Manual validation for burn operations
   - Ensures files never become invalid

5. **Data Length Validation**
   - Required for close operations
   - Prevents closing files with data

### Error Propagation

All errors return immediately with appropriate error codes. No partial state modifications occur on error due to:
- Validation before modification
- Atomic operations (OpTransfer, OpCreateFile, OpDeleteFile)
- Transaction-level rollback on any error

## Testing Strategy

### Unit Tests (QuanticScript Level)

1. **CREATE_FILE Tests**
   - Valid creation with positive balance
   - Creation with zero balance
   - Invalid instruction length
   - Negative balance rejection
   - System Program balance validation

2. **TRANSFER Tests**
   - Valid transfer between owned files
   - Transfer from non-owned file (should fail)
   - Zero amount rejection
   - Negative amount rejection
   - Storage cost violation (should fail)
   - Overflow protection

3. **BURN Tests**
   - Valid burn with sufficient remaining balance
   - Burn from non-owned file (should fail)
   - Zero amount rejection
   - Negative amount rejection
   - Storage cost violation (should fail)

4. **CLOSE_FILE Tests**
   - Valid close with zero data
   - Close non-owned file (should fail)
   - Close file with data (should fail)
   - Balance transfer to destination
   - File deletion verification

### Integration Tests (Go Level)

1. **End-to-End Workflow**
   - Create file → Transfer → Burn → Close
   - Multiple files managed by System Program
   - Cross-program interaction scenarios

2. **Error Recovery**
   - Transaction rollback on errors
   - State consistency after failures

3. **Storage Cost Scenarios**
   - Files with varying data sizes
   - Minimum balance enforcement
   - Cost calculation accuracy

### Security Tests

1. **Ownership Enforcement**
   - Attempt operations on files owned by other programs
   - Verify all operations reject unauthorized access

2. **Integer Overflow**
   - Large balance transfers
   - Maximum i64 values

3. **Malformed Instructions**
   - Invalid lengths
   - Corrupted data
   - Missing parameters

## Implementation Notes

### Leveraging Existing Opcodes

The design relies heavily on existing, tested opcodes:

- **OpTransfer**: Handles balance transfers with automatic ownership and storage cost validation
- **OpCreateFile**: Creates new files with specified owner (TxManager)
- **OpDeleteFile**: Removes files with ownership validation
- **OpGetFile**: Reads file state for validation
- **OpUpdateFile**: Persists file modifications

This minimizes new code and leverages battle-tested implementations.

### Storage Cost Calculation

Storage costs are calculated using the exponential formula:
```
cost = base * size_in_kb * (1.1 ^ size_in_mb)
```

Where:
- `base = 1000 units per KB`
- `growth_rate = 1.1`

This is handled by `filestore.CalculateStorageCost()` and enforced by OpTransfer automatically.

### Ownership Model

The System Program owns all files it creates by setting `TxManager = SystemProgramID` during creation. This ownership is immutable and enforced at the interpreter level, providing strong security guarantees.

### Balance Management

All balance operations go through OpTransfer except for burn operations, which manually reduce balance and update the file. This ensures:
- Atomic balance updates
- Automatic storage cost validation (for transfers)
- Overflow protection
- Consistent error handling

## Migration from Legacy System

### Removal of Old Implementation

1. Delete `internal/system/` directory
2. Remove Go-based System Program registration
3. Update all references to use `genesis.SystemProgramID`
4. Ensure QuanticScript bytecode is loaded via `genesis.LoadBuiltinPrograms`

### Compatibility

The new System Program maintains the same instruction interface as the legacy implementation, ensuring existing clients continue to work without modification.

### Testing Migration

All existing System Program tests will be updated to:
- Use QuanticScript bytecode execution
- Remove Go package imports
- Verify identical behavior to legacy implementation

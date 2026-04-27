# Design Document

## Overview

This design establishes a standardized pattern for constructing transactions with properly declared instruction inputs. The focus is on ensuring that all file accesses are explicitly declared with correct permissions before execution, enabling both access control validation and parallel execution analysis.

The design leverages the existing `AccessController` and `TxProcessor` infrastructure, adding validation layers and helper functions to enforce the transaction structure requirements. For System Program operations like transfers, the pattern requires declaring the System Program as readable, and sender/receiver files as writable.

## Architecture

### Transaction Structure Flow

```
Transaction
├── LastSeen (TxID)
├── Signatures[] (fee payer first)
│   └── Signature { PublicKey, Signature[64] }
└── Instructions[]
    └── Instruction
        ├── ProgramID (FileID)
        ├── Inputs (map[string]FileAccess)
        │   ├── "program" → { FileID, Read }
        │   ├── "sender" → { FileID, Write }
        │   └── "receiver" → { FileID, Write }
        └── Data ([]byte - instruction-specific)
```

### Validation Layers

```
1. Transaction Validation (TxProcessor)
   ↓
2. Signature Verification
   ↓
3. Fee Deduction
   ↓
4. Instruction Input Validation (NEW)
   ↓
5. Access Controller Setup
   ↓
6. Program Execution
   ↓
7. Final Access Validation
```

## Components and Interfaces

### 1. Transaction Builder Helper

**Purpose**: Provide a standardized way to construct transactions with proper input declarations.

**Location**: `internal/transaction/builder.go` (new file)

**Interface**:

```go
type TransactionBuilder struct {
    lastSeen     TxID
    instructions []Instruction
    signatures   []Signature
}

// NewTransactionBuilder creates a new builder
func NewTransactionBuilder(lastSeen TxID) *TransactionBuilder

// AddTransferInstruction adds a System Program transfer instruction
// with proper input declarations
func (tb *TransactionBuilder) AddTransferInstruction(
    systemProgramID FileID,
    senderID FileID,
    receiverID FileID,
    amount int64,
) error

// AddSignature adds a signature to the transaction
func (tb *TransactionBuilder) AddSignature(sig Signature) error

// Build constructs the final transaction
func (tb *TransactionBuilder) Build() (*Transaction, error)
```

**Key Operations**:
- Automatically declares System Program with Read permission
- Automatically declares sender and receiver with Write permission
- Encodes transfer instruction data in correct format
- Validates all required fields before building

**Example Usage**:

```go
builder := transaction.NewTransactionBuilder(lastSeenTxID)

// Add transfer instruction
err := builder.AddTransferInstruction(
    genesis.SystemProgramID,
    senderFileID,
    receiverFileID,
    1000, // amount
)

// Add sender signature
builder.AddSignature(senderSignature)

// Build transaction
tx, err := builder.Build()
```

### 2. Instruction Input Validator

**Purpose**: Validate that instruction inputs are properly declared before execution.

**Location**: `internal/processor/input_validator.go` (new file)

**Interface**:

```go
type InputValidator struct {
    fileStore *filestore.FileStore
}

// NewInputValidator creates a new input validator
func NewInputValidator(fs *filestore.FileStore) *InputValidator

// ValidateInstructionInputs validates all input declarations
func (iv *InputValidator) ValidateInstructionInputs(instr *Instruction) error

// ValidateExecutableProgram validates program is executable
func (iv *InputValidator) ValidateExecutableProgram(programID FileID) error

// ValidateFileExists validates file exists in store
func (iv *InputValidator) ValidateFileExists(fileID FileID) error
```

**Validation Rules**:

1. All files in Inputs map must exist in FileStore
2. ProgramID must be in Inputs map (or implicitly readable)
3. Program file must have Executable flag set to true
4. Files declared with Write permission must be writable
5. Inputs map must not be empty for instructions that access files

**Error Messages**:

```go
// Missing program declaration
"program %s not declared in instruction inputs"

// Program not executable
"program %s is not marked as executable"

// File not found
"input file %s not found in file store"

// Invalid permission
"file %s declared with invalid permission: %d"
```

### 3. Enhanced TxProcessor Integration

**Purpose**: Integrate input validation into transaction processing flow.

**Location**: `internal/processor/tx_processor.go` (modifications)

**Changes**:

```go
// Add input validator field
type TxProcessor struct {
    fileStore      *filestore.FileStore
    runtime        *runtime.Runtime
    feeConfig      FeeConfig
    stateCache     map[filestore.FileID]*filestore.File
    inputValidator *InputValidator  // NEW
    mu             sync.Mutex
}

// Update ExecuteInstruction to validate inputs first
func (tp *TxProcessor) ExecuteInstruction(
    instr *transaction.Instruction,
    signers []transaction.PublicKey,
) error {
    // NEW: Validate instruction inputs before execution
    if err := tp.inputValidator.ValidateInstructionInputs(instr); err != nil {
        return fmt.Errorf("input validation failed: %w", err)
    }
    
    // Existing validation and execution logic...
}
```

**Validation Order**:

1. Validate instruction inputs (new)
2. Load program file
3. Validate program is executable (existing)
4. Validate all input files exist (enhanced)
5. Create access controller
6. Execute program
7. Validate final access

### 4. System Program Transfer Helper

**Purpose**: Provide a convenient function for creating transfer instructions.

**Location**: `internal/transaction/system_helpers.go` (new file)

**Interface**:

```go
// CreateTransferInstruction creates a properly formatted transfer instruction
func CreateTransferInstruction(
    systemProgramID filestore.FileID,
    senderID filestore.FileID,
    receiverID filestore.FileID,
    amount int64,
) (*Instruction, error)

// EncodeTransferData encodes transfer parameters into instruction data
func EncodeTransferData(
    fromID filestore.FileID,
    toID filestore.FileID,
    amount int64,
) []byte
```

**Implementation**:

```go
func CreateTransferInstruction(
    systemProgramID filestore.FileID,
    senderID filestore.FileID,
    receiverID filestore.FileID,
    amount int64,
) (*Instruction, error) {
    // Validate amount
    if amount <= 0 {
        return nil, fmt.Errorf("amount must be positive")
    }
    
    // Create inputs map
    inputs := map[string]FileAccess{
        "program": {
            FileID:     systemProgramID,
            Permission: Read,
        },
        "sender": {
            FileID:     senderID,
            Permission: Write,
        },
        "receiver": {
            FileID:     receiverID,
            Permission: Write,
        },
    }
    
    // Encode instruction data
    data := EncodeTransferData(senderID, receiverID, amount)
    
    return &Instruction{
        ProgramID: systemProgramID,
        Inputs:    inputs,
        Data:      data,
    }, nil
}

func EncodeTransferData(
    fromID filestore.FileID,
    toID filestore.FileID,
    amount int64,
) []byte {
    // Format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
    data := make([]byte, 73)
    
    // Instruction type (TRANSFER = 1)
    data[0] = 1
    
    // From FileID
    copy(data[1:33], fromID[:])
    
    // To FileID
    copy(data[33:65], toID[:])
    
    // Amount (little-endian)
    binary.LittleEndian.PutUint64(data[65:73], uint64(amount))
    
    return data
}
```

## Data Models

### FileAccess Structure (existing)

```go
type FileAccess struct {
    FileID     filestore.FileID
    Permission AccessPermission
}
```

### AccessPermission Constants (existing)

```go
const (
    Read  AccessPermission = 1
    Write AccessPermission = 2
)
```

### Instruction Structure (existing)

```go
type Instruction struct {
    ProgramID filestore.FileID
    Inputs    map[string]FileAccess
    Data      []byte
}
```

### Standard Input Keys

For consistency across the codebase, we define standard keys for common file types:

```go
const (
    InputKeyProgram  = "program"   // The program being invoked
    InputKeySender   = "sender"    // Source account for transfers
    InputKeyReceiver = "receiver"  // Destination account for transfers
    InputKeyPayer    = "payer"     // Fee payer or balance source
    InputKeyOwner    = "owner"     // File owner for ownership checks
    InputKeyTarget   = "target"    // Generic target file
)
```

## Error Handling

### Validation Error Types

```go
// ErrMissingInput indicates a required file is not declared in inputs
type ErrMissingInput struct {
    FileID   filestore.FileID
    Required string // Description of why it's required
}

// ErrInvalidPermission indicates incorrect permission declaration
type ErrInvalidPermission struct {
    FileID   filestore.FileID
    Declared AccessPermission
    Required AccessPermission
}

// ErrProgramNotExecutable indicates program file lacks executable flag
type ErrProgramNotExecutable struct {
    ProgramID filestore.FileID
}

// ErrFileNotFound indicates input file doesn't exist
type ErrFileNotFound struct {
    FileID filestore.FileID
}
```

### Error Messages

All validation errors should include:
1. The specific file ID that caused the error
2. What was expected vs what was found
3. How to fix the issue

Example error messages:

```
"program 0x00...01 not declared in instruction inputs; add to Inputs map with Read permission"

"file 0xab...cd declared with Read permission but Write required for balance modification"

"program 0x00...01 is not marked as executable; ensure Executable flag is true"

"input file 0xab...cd not found in file store; ensure file exists before transaction"
```

## Testing Strategy

### Unit Tests

1. **TransactionBuilder Tests** (`transaction/builder_test.go`)
   - Test building transfer transaction with correct inputs
   - Test validation of missing fields
   - Test encoding of instruction data
   - Test multiple instructions in one transaction

2. **InputValidator Tests** (`processor/input_validator_test.go`)
   - Test validation of properly declared inputs
   - Test detection of missing program declaration
   - Test detection of non-executable program
   - Test detection of missing files
   - Test detection of incorrect permissions
   - Test validation of empty inputs map

3. **System Helpers Tests** (`transaction/system_helpers_test.go`)
   - Test CreateTransferInstruction with valid parameters
   - Test rejection of zero/negative amounts
   - Test correct encoding of transfer data
   - Test correct permission declarations

### Integration Tests

1. **End-to-End Transfer Test** (`internal/transfer_integration_test.go`)
   - Create sender and receiver accounts
   - Build transfer transaction using builder
   - Submit transaction to processor
   - Verify balances updated correctly
   - Verify access log shows correct permissions

2. **Validation Failure Tests** (`internal/validation_integration_test.go`)
   - Test transaction with missing program declaration
   - Test transaction with incorrect permissions
   - Test transaction with non-existent files
   - Test transaction with non-executable program
   - Verify all fail before execution
   - Verify no state changes on validation failure

3. **Multi-Instruction Tests** (`internal/multi_instruction_test.go`)
   - Test transaction with multiple transfer instructions
   - Test transaction with mixed instruction types
   - Verify each instruction validated independently
   - Verify rollback on any instruction failure

### Example Test Cases

```go
func TestTransferTransactionBuilder(t *testing.T) {
    // Setup
    lastSeen := TxID{}
    systemID := genesis.SystemProgramID
    senderID := filestore.NewFileID()
    receiverID := filestore.NewFileID()
    
    // Build transaction
    builder := transaction.NewTransactionBuilder(lastSeen)
    err := builder.AddTransferInstruction(systemID, senderID, receiverID, 1000)
    require.NoError(t, err)
    
    builder.AddSignature(senderSignature)
    
    tx, err := builder.Build()
    require.NoError(t, err)
    
    // Verify structure
    require.Len(t, tx.Instructions, 1)
    instr := tx.Instructions[0]
    
    // Verify program declared with Read
    programAccess, ok := instr.Inputs["program"]
    require.True(t, ok)
    require.Equal(t, systemID, programAccess.FileID)
    require.Equal(t, transaction.Read, programAccess.Permission)
    
    // Verify sender declared with Write
    senderAccess, ok := instr.Inputs["sender"]
    require.True(t, ok)
    require.Equal(t, senderID, senderAccess.FileID)
    require.Equal(t, transaction.Write, senderAccess.Permission)
    
    // Verify receiver declared with Write
    receiverAccess, ok := instr.Inputs["receiver"]
    require.True(t, ok)
    require.Equal(t, receiverID, receiverAccess.FileID)
    require.Equal(t, transaction.Write, receiverAccess.Permission)
}

func TestInputValidatorRejectsMissingProgram(t *testing.T) {
    // Setup
    fs := filestore.NewFileStore()
    validator := processor.NewInputValidator(fs)
    
    // Create instruction without program in inputs
    instr := &transaction.Instruction{
        ProgramID: genesis.SystemProgramID,
        Inputs:    map[string]transaction.FileAccess{}, // Empty!
        Data:      []byte{1, 2, 3},
    }
    
    // Validate
    err := validator.ValidateInstructionInputs(instr)
    
    // Should fail
    require.Error(t, err)
    require.Contains(t, err.Error(), "not declared in instruction inputs")
}
```

## Implementation Notes

### Backward Compatibility

The existing transaction structure already supports the Inputs map, so this design is backward compatible. However, we're adding stricter validation:

- Old code that doesn't declare inputs will now fail validation
- Old code must be updated to use the builder or manually declare inputs
- Tests must be updated to include proper input declarations

### Performance Considerations

Input validation adds minimal overhead:
- Map lookups: O(1) for each file
- File existence checks: Already cached in FileStore
- Permission checks: Simple integer comparisons

The validation happens once per instruction before execution, so the performance impact is negligible compared to program execution.

### Standard Pattern Documentation

We should document the standard pattern for each System Program operation:

**CREATE_FILE**:
```go
Inputs: {
    "program": { SystemProgramID, Read },
    "payer": { PayerFileID, Write },
    "owner": { NewFileID, Write },
}
```

**TRANSFER**:
```go
Inputs: {
    "program": { SystemProgramID, Read },
    "sender": { SenderFileID, Write },
    "receiver": { ReceiverFileID, Write },
}
```

**BURN**:
```go
Inputs: {
    "program": { SystemProgramID, Read },
    "target": { TargetFileID, Write },
}
```

**CLOSE_FILE**:
```go
Inputs: {
    "program": { SystemProgramID, Read },
    "target": { FileToClose, Write },
    "destination": { DestinationFileID, Write },
}
```

### CLI Integration

The CLI commands should be updated to use the transaction builder:

```go
// In cmd/main.go transfer command
func executeTransfer(from, to FileID, amount int64) error {
    // Load last seen transaction
    lastSeen := getLastSeenTx()
    
    // Build transaction using helper
    builder := transaction.NewTransactionBuilder(lastSeen)
    err := builder.AddTransferInstruction(
        genesis.SystemProgramID,
        from,
        to,
        amount,
    )
    if err != nil {
        return err
    }
    
    // Sign with sender's keypair
    sig := signTransaction(builder, senderKeypair)
    builder.AddSignature(sig)
    
    // Build and submit
    tx, err := builder.Build()
    if err != nil {
        return err
    }
    
    return submitTransaction(tx)
}
```

## Migration Path

### Phase 1: Add New Components (Non-Breaking)
1. Add TransactionBuilder
2. Add InputValidator
3. Add system_helpers.go
4. Add comprehensive tests

### Phase 2: Integrate Validation (Breaking)
1. Update TxProcessor to use InputValidator
2. Update all tests to use proper input declarations
3. Update CLI to use TransactionBuilder

### Phase 3: Documentation and Examples
1. Document standard patterns for each operation
2. Update examples to show proper usage
3. Add migration guide for existing code

### Phase 4: Enforcement
1. Enable strict validation in production
2. Reject transactions with missing/incorrect inputs
3. Monitor for validation failures

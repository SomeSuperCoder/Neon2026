# Design Document: File-Based State Model

## Overview

This design implements a file-based state model for the blockchain where all on-chain state is represented as uniform file objects. The architecture draws inspiration from Unix's "everything is a file" philosophy and Solana's account model, creating a flexible system where user accounts, programs, and arbitrary data are all treated as files with associated metadata.

The system enables explicit declaration of file access patterns in transactions, laying the groundwork for future parallel execution of non-conflicting transactions. Programs are themselves files containing bytecode, executed by a built-in Runtime program.

## Architecture

The system follows a layered architecture that integrates with the existing PoH blockchain:

```
┌─────────────────────────────────────────┐
│         Application Layer               │
│  (Wallet, CLI, Program Deployment)      │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│      Transaction Processing Layer       │
│  (Instruction Execution, Fee Payment)   │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         Runtime Layer                   │
│  (Bytecode Interpreter, Program Exec)   │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│         State Management Layer          │
│  (File Store, Access Control)           │
└─────────────────────────────────────────┘
                  │
┌─────────────────────────────────────────┐
│      PoH Blockchain Layer               │
│  (Transaction Ordering, Blocks)         │
└─────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go (Golang)
- **Storage**: Embedded key-value store (BadgerDB) for file state
- **Serialization**: Protocol Buffers for efficient file and transaction encoding (implemented)
- **Hashing**: SHA-256 for file IDs and content addressing
- **Cryptography**: Ed25519 for signatures
- **Future**: WebAssembly (WASM) or custom bytecode interpreter for program execution

## Components and Interfaces

### 1. File Store

**Purpose**: Manage all on-chain file state with efficient storage and retrieval.

**Struct**: `FileStore`

**Key Methods**:
- `NewFileStore(dbPath string) (*FileStore, error)`: Initialize file store
- `CreateFile(file *File) (FileID, error)`: Create a new file and return its ID
- `GetFile(id FileID) (*File, error)`: Retrieve a file by ID
- `UpdateFile(id FileID, file *File) error`: Update an existing file
- `DeleteFile(id FileID) error`: Remove a file from state
- `GetFileBalance(id FileID) (int64, error)`: Get file balance
- `UpdateFileBalance(id FileID, delta int64) error`: Modify file balance
- `ValidateStorageCost(file *File) error`: Check if balance covers storage cost

**Fields**:
- `db *badger.DB`: BadgerDB instance for persistent storage
- `cache map[FileID]*File`: In-memory cache for hot files
- `mu sync.RWMutex`: Mutex for thread-safe operations

### 2. File Data Structure

**Struct**: `File`

```go
type File struct {
    ID          FileID
    Balance     int64
    TxManager   FileID
    Data        []byte
    Executable  bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type FileID [32]byte  // SHA-256 hash
```

**Note**: File ownership and access control is managed by the TxManager program through the file's Data field. Each TxManager defines its own ownership model and stores ownership information within the file data it manages.

**Storage Cost Calculation**:
```go
func CalculateStorageCost(dataSize int64) int64 {
    // Base cost: 1000 units per KB
    // Exponential growth: cost = base * (1.1 ^ (size_in_mb))
    baseCostPerKB := int64(1000)
    sizeInKB := dataSize / 1024
    if sizeInKB == 0 {
        sizeInKB = 1
    }
    sizeInMB := float64(sizeInKB) / 1024.0
    multiplier := math.Pow(1.1, sizeInMB)
    return int64(float64(baseCostPerKB*sizeInKB) * multiplier)
}
```

### 3. Transaction Structure

**Struct**: `Transaction`

```go
type Transaction struct {
    LastSeen     TxID
    Instructions []Instruction
    Signatures   []Signature
}

type TxID [32]byte

type Signature struct {
    PublicKey PublicKey
    Signature [64]byte
}
```

**Transaction Processing Flow**:
1. Validate signatures
2. Identify fee payer (first signature)
3. Deduct transaction fee from fee payer
4. Execute instructions sequentially
5. Commit state changes or revert on error

### 4. Instruction Structure

**Struct**: `Instruction`

```go
type Instruction struct {
    ProgramID  FileID
    Inputs     map[string]FileAccess
    Data       []byte
}

type FileAccess struct {
    FileID     FileID
    Permission AccessPermission
}

type AccessPermission uint8

const (
    Read  AccessPermission = 1
    Write AccessPermission = 2
)
```

**Instruction Execution**:
1. Load program file by ProgramID
2. Validate all input files exist
3. Check access permissions match actual usage
4. Load program's transaction manager
5. Execute transaction manager with instruction data
6. Validate no unauthorized file access occurred
7. Commit file state changes

### 5. Transaction Processor

**Purpose**: Execute transactions and manage state transitions.

**Struct**: `TxProcessor`

**Key Methods**:
- `NewTxProcessor(fileStore *FileStore, runtime *Runtime) *TxProcessor`: Initialize processor
- `ProcessTransaction(tx *Transaction) (*TxResult, error)`: Execute a transaction
- `ValidateTransaction(tx *Transaction) error`: Pre-execution validation
- `ExecuteInstruction(instr *Instruction, signers []PublicKey) error`: Execute single instruction
- `DeductFee(feePayer FileID, fee int64) error`: Charge transaction fee
- `CalculateFee(tx *Transaction) int64`: Compute transaction cost
- `RevertState()`: Rollback changes on error

**Fields**:
- `fileStore *FileStore`: Reference to file store
- `runtime *Runtime`: Reference to runtime for program execution
- `stateCache map[FileID]*File`: Transaction-local state cache
- `accessLog map[FileID]AccessPermission`: Track actual file accesses
- `mu sync.Mutex`: Mutex for transaction execution

**Fee Calculation**:
```go
func (tp *TxProcessor) CalculateFee(tx *Transaction) int64 {
    baseFee := int64(5000)  // Base fee per transaction
    instructionFee := int64(1000) * int64(len(tx.Instructions))
    signatureFee := int64(500) * int64(len(tx.Signatures))
    return baseFee + instructionFee + signatureFee
}
```

### 6. Runtime System

**Purpose**: Execute program bytecode (placeholder for future bytecode interpreter).

**Struct**: `Runtime`

**Key Methods**:
- `NewRuntime() *Runtime`: Initialize runtime
- `ExecuteProgram(program *File, instruction *Instruction, ctx *ExecutionContext) error`: Execute program bytecode
- `ValidateProgram(program *File) error`: Verify program bytecode is valid
- `GetProgramInterface(programID FileID) (*ProgramInterface, error)`: Get program metadata

**Fields**:
- `builtinPrograms map[FileID]BuiltinProgram`: Registry of built-in programs
- `executionLimit int64`: Maximum computation units per instruction

**Execution Context**:
```go
type ExecutionContext struct {
    Instruction  *Instruction
    Signers      []PublicKey
    FileStore    *FileStore
    AccessLog    map[FileID]AccessPermission
}
```

**Note**: Initial implementation will use native Go functions for built-in programs. Future versions will implement a bytecode interpreter (WASM or custom VM).

### 7. System Program

**Purpose**: Provide built-in operations for account management.

**Struct**: `SystemProgram`

**Key Methods**:
- `CreateAccount(ctx *ExecutionContext, owner PublicKey, initialBalance int64) (FileID, error)`: Create new user account
- `Transfer(ctx *ExecutionContext, from FileID, to FileID, amount int64) error`: Transfer balance between accounts
- `CloseAccount(ctx *ExecutionContext, account FileID, destination FileID) error`: Close account and reclaim balance
- `AllocateData(ctx *ExecutionContext, account FileID, size int64) error`: Allocate data space for account

**System Program ID**: `0x0000000000000000000000000000000000000000000000000000000000000001`

**Instruction Data Format**:
```go
type SystemInstruction struct {
    InstructionType uint8
    Params          []byte  // Encoded parameters specific to instruction type
}

const (
    CreateAccountInstruction uint8 = 0
    TransferInstruction      uint8 = 1
    CloseAccountInstruction  uint8 = 2
    AllocateDataInstruction  uint8 = 3
)
```

### 8. Access Control Manager

**Purpose**: Validate file access permissions during instruction execution.

**Struct**: `AccessController`

**Key Methods**:
- `NewAccessController() *AccessController`: Initialize access controller
- `ValidateAccess(fileID FileID, permission AccessPermission, declared map[string]FileAccess) error`: Check if access is allowed
- `RecordAccess(fileID FileID, permission AccessPermission)`: Log file access
- `GetAccessLog() map[FileID]AccessPermission`: Retrieve access log for validation

**Note**: Signature authorization is delegated to the TxManager program, which implements its own authorization logic based on the file's data.

**Fields**:
- `accessLog map[FileID]AccessPermission`: Track actual accesses during execution
- `declaredAccess map[FileID]AccessPermission`: Declared access from instruction

### 9. Parallel Execution Analyzer (Future)

**Purpose**: Analyze transactions for parallel execution opportunities.

**Struct**: `ParallelAnalyzer`

**Key Methods**:
- `AnalyzeDependencies(txs []*Transaction) *DependencyGraph`: Build transaction dependency graph
- `FindParallelBatches(graph *DependencyGraph) [][]int`: Identify non-conflicting transaction sets
- `HasConflict(tx1 *Transaction, tx2 *Transaction) bool`: Check if two transactions conflict

**Conflict Detection**:
- Two transactions conflict if they both write to the same file
- Two transactions conflict if one writes and another reads the same file
- Two transactions do not conflict if they only read the same files

## Data Models

### File ID Generation

```
File ID = SHA-256(owner_pubkey || creation_timestamp || nonce)
```

**Note**: The specific File ID generation scheme may be defined by the creating program. The above is a suggested approach for the System Program.

### File State Lifecycle

```
[Create] -> [Active] -> [Modified] -> [Active] -> [Closed]
              │            │
              └────────────┘
           (balance updates, data changes)
```

### Transaction Execution Flow

```
Transaction Submitted
    │
    ├─> Validate Signatures
    ├─> Validate Fee Payer Balance
    ├─> Deduct Transaction Fee
    │
    ├─> For Each Instruction:
    │   ├─> Load Program File
    │   ├─> Validate Input Files Exist
    │   ├─> Check Declared Permissions
    │   ├─> Execute Program via Runtime
    │   ├─> Validate Actual Access Matches Declared
    │   └─> Commit or Revert
    │
    └─> Return Transaction Result
```

### Instruction Execution Example

```go
// Transfer 1000 tokens from Alice to Bob
instruction := Instruction{
    ProgramID: SystemProgramID,
    Inputs: map[string]FileAccess{
        "from": {FileID: aliceAccountID, Permission: Write},
        "to":   {FileID: bobAccountID, Permission: Write},
    },
    Data: EncodeTransferParams(1000),
}
```

### Storage Cost Example

```
File with 1 KB data:
  Cost = 1000 * 1 * (1.1^0.001) ≈ 1000 units

File with 1 MB data:
  Cost = 1000 * 1024 * (1.1^1) ≈ 1,126,400 units

File with 10 MB data:
  Cost = 1000 * 10240 * (1.1^10) ≈ 26,577,000 units
```

## Error Handling

### File Store Errors
- **File Not Found**: Return error, do not create implicitly
- **Insufficient Balance**: Reject operation, revert transaction
- **Storage Cost Exceeded**: Reject data allocation, maintain current state
- **Database Corruption**: Attempt recovery, fallback to blockchain resync

### Transaction Processing Errors
- **Invalid Signature**: Reject transaction immediately
- **Insufficient Fee**: Reject transaction before execution
- **Unauthorized Access**: Halt execution, revert all changes
- **Program Execution Failure**: Revert transaction, return error to user

### Access Control Errors
- **Undeclared File Access**: Halt execution, revert transaction
- **Permission Violation**: Halt execution, revert transaction
- **Missing Signature**: Reject instruction before execution

### Runtime Errors
- **Invalid Bytecode**: Reject program deployment
- **Execution Timeout**: Halt program, revert transaction
- **Out of Compute Units**: Halt program, revert transaction
- **Stack Overflow**: Halt program, revert transaction

## Testing Strategy

### Unit Tests

1. **File Store Tests**
   - Test file creation with valid parameters
   - Test file retrieval by ID
   - Test balance updates
   - Test storage cost validation
   - Test file deletion

2. **Transaction Processing Tests**
   - Test fee calculation for various transaction sizes
   - Test fee deduction from fee payer
   - Test instruction execution order
   - Test state revert on error
   - Test signature validation

3. **Access Control Tests**
   - Test read permission validation
   - Test write permission validation
   - Test undeclared access detection
   - Test signature authorization

4. **System Program Tests**
   - Test account creation
   - Test balance transfer
   - Test account closure
   - Test data allocation

### Integration Tests

1. **End-to-End Transaction Flow**
   - Create accounts, transfer balance, verify final state

2. **Multi-Instruction Transactions**
   - Execute transaction with multiple instructions, verify atomicity

3. **Program Deployment and Execution**
   - Deploy program file, execute instruction targeting program

4. **Storage Cost Enforcement**
   - Create file with data, reduce balance, verify rejection

### Future Parallel Execution Tests

1. **Dependency Analysis**
   - Analyze non-conflicting transactions, verify correct batching

2. **Conflict Detection**
   - Test read-read (no conflict), read-write (conflict), write-write (conflict)

## Integration with PoH Blockchain

The file-based state model integrates with the existing PoH blockchain as follows:

1. **Transaction Inclusion**: File-based transactions are included in PoH entries
2. **State Commitment**: File state root hash is included in block headers
3. **Verification**: Block verification includes file state transition validation
4. **Persistence**: File state is persisted alongside blockchain data

### Modified Block Structure

```go
type Block struct {
    Header  BlockHeader
    Entries []Entry
}

type BlockHeader struct {
    PreviousBlockHash []byte
    MerkleRoot        []byte
    StateRoot         []byte  // NEW: Root hash of file state tree
    Slot              int64
    Timestamp         time.Time
    BlockHeight       int64
}
```

### State Root Calculation

```
State Root = MerkleRoot(sorted_file_ids, file_hashes)

Where file_hash = SHA-256(file.ID || file.Balance || file.TxManager || 
                          file.Data || file.Executable || file.UpdatedAt)
```

## Implementation Phases

1. **Phase 1**: File Store and Data Structures
   - Implement File struct and FileStore
   - Implement storage cost calculation
   - Add persistence with BadgerDB

2. **Phase 2**: Transaction and Instruction Processing
   - Implement Transaction and Instruction structs
   - Implement TxProcessor with fee handling
   - Add signature validation

3. **Phase 3**: Access Control System
   - Implement AccessController
   - Add permission validation
   - Implement access logging

4. **Phase 4**: System Program
   - Implement built-in System Program
   - Add account creation, transfer, close operations
   - Integrate with TxProcessor

5. **Phase 5**: Runtime System (Stub)
   - Create Runtime interface
   - Implement native Go execution for built-in programs
   - Add execution context and limits

6. **Phase 6**: Integration with PoH Blockchain
   - Modify block structure to include state root
   - Integrate file-based transactions into PoH entries
   - Add state verification to block validation

7. **Phase 7**: Testing and Optimization
   - Comprehensive unit and integration tests
   - Performance optimization
   - Documentation

## Future Enhancements

1. **Bytecode Interpreter**: Implement WASM or custom VM for program execution
2. **Parallel Execution**: Implement transaction scheduler using dependency analysis
3. **State Compression**: Add state tree compression for efficient storage
4. **Cross-Program Invocation**: Allow programs to call other programs
5. **Rent System**: Implement periodic rent collection for file storage
6. **File Versioning**: Add support for file history and rollback


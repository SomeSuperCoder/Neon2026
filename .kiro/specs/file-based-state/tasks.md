# Implementation Plan

- [x] 1. Set up project structure for file-based state model
  - Create directory structure: `filestore/`, `transaction/`, `runtime/`, `system/`
  - Define Go module and dependencies (BadgerDB, protobuf)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  - _Status: Package created at `internal/filestore/filestore.go`_

- [x] 2. Implement core File data structure and FileID type
  - Create `File` struct with all required fields (Balance, TxManager, Data, Executable, timestamps)
  - Implement `FileID` as 32-byte array type
  - Add file serialization/deserialization methods using protobuf
  - Implement FileID helper functions (String, FromString, FromBytes, Generate)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  - _Status: Completed with protobuf serialization in `internal/filestore/filestore.go`_

- [x] 3. Implement storage cost calculation
  - Create `CalculateStorageCost(dataSize int64) int64` function with exponential growth formula
  - Implement cost validation logic
  - Add helper functions for KB/MB conversions
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 4. Implement FileStore with BadgerDB persistence
  - Create `FileStore` struct with BadgerDB instance and in-memory cache
  - Implement `NewFileStore(dbPath string)` initialization
  - Implement `CreateFile(file *File) (FileID, error)` with storage cost validation
  - Implement `GetFile(id FileID) (*File, error)` with caching
  - Implement `UpdateFile(id FileID, file *File)` with validation
  - Implement `DeleteFile(id FileID)` 
  - Implement `GetFileBalance()` and `UpdateFileBalance()` methods
  - Add thread-safe operations with RWMutex
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 5. Implement Transaction and Instruction data structures
  - Create `Transaction` struct with LastSeen, Instructions, Signatures fields
  - Create `Instruction` struct with ProgramID, Inputs map, Data fields
  - Create `FileAccess` struct with FileID and Permission fields
  - Define `AccessPermission` type with Read/Write constants
  - Create `Signature` struct with PublicKey and signature bytes
  - Add serialization/deserialization for all structures
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_
  - _Status: Completed with JSON serialization, TxID/PublicKey helpers, signature verification, and GetFeePayer/GetSigners methods in `internal/transaction/transaction.go`_

- [x] 6. Implement transaction fee calculation
  - Create `CalculateFee(tx *Transaction) int64` function
  - Implement base fee + per-instruction + per-signature cost model
  - Add configurable fee parameters
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  - _Status: Completed with FeeConfig struct, DefaultFeeConfig(), CalculateFee(), and CalculateFeeWithDefaults() in `internal/transaction/transaction.go`_

- [x] 7. Implement AccessController for permission validation
  - Create `AccessController` struct with access logging
  - Implement `ValidateAccess()` to check declared vs actual access
  - Implement `RecordAccess()` to log file accesses during execution
  - Implement `GetAccessLog()` to retrieve access history
  - Add permission checking logic (Read vs Write)
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Status: Completed in `internal/access/access_controller.go` with full implementation including SetDeclaredAccess, ValidateAndRecord, Reset, GetDeclaredAccess, and ValidateFinalAccess methods_

- [x] 8. Implement ExecutionContext for program execution
  - Create `ExecutionContext` struct with Instruction, Signers, FileStore, AccessController
  - Add helper methods for file access within context (GetFile, GetFileMut, UpdateFile)
  - Implement context-aware file loading and modification with permission validation
  - Add convenience methods: GetFileBalance, UpdateFileBalance, HasSigner, GetInstructionData, GetProgramID, GetInputFileID
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Status: Completed in `internal/runtime/runtime.go` with 12 comprehensive unit tests_

- [x] 9. Implement Runtime system
  - Create `Runtime` struct with builtin program registry and execution limit
  - Implement `NewRuntime()` initialization with default 1M compute units limit
  - Implement `RegisterBuiltinProgram()` for adding native Go program implementations
  - Implement `ExecuteProgram()` that dispatches to builtin programs or bytecode interpreter
  - Implement `ValidateProgram()` for program validation (executable flag, bytecode presence)
  - Add helper methods: GetBuiltinProgram, IsBuiltinProgram, SetExecutionLimit, GetExecutionLimit
  - Create `BuiltinProgram` interface with Execute() and GetProgramID() methods
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  - _Status: Completed in `internal/runtime/runtime.go` with 6 comprehensive unit tests_

- [x] 10. Implement System Program builtin
  - Define System Program ID constant
  - Create `SystemProgram` struct implementing builtin program interface
  - Implement `CreateAccount()` instruction handler
  - Implement `Transfer()` instruction handler  
  - Implement `CloseAccount()` instruction handler
  - Implement `AllocateData()` instruction handler
  - Define instruction data encoding/decoding for each operation
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3, 10.4, 10.5_
  - _Status: Completed in `internal/system/system.go` with all four instruction handlers (CreateAccount, Transfer, CloseAccount, AllocateData) and encoding helper functions_

- [x] 11. Implement TxProcessor for transaction execution
  - Create `TxProcessor` struct with FileStore and Runtime references
  - Implement `NewTxProcessor()` initialization
  - Implement `ValidateTransaction()` for pre-execution checks (signatures, fee payer balance)
  - Implement `DeductFee()` to charge fee payer before execution
  - Implement `ExecuteInstruction()` to process single instruction with access control
  - Implement `ProcessTransaction()` orchestrating full transaction flow
  - Implement `RevertState()` for rollback on errors
  - Add transaction-local state cache for atomicity
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 7.1, 7.2, 7.3, 7.4, 7.5, 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Status: Completed in `internal/processor/tx_processor.go` with full transaction orchestration, atomic state management via transaction-local cache, cachedFileStore wrapper for isolated execution, and automatic rollback on errors_

- [x] 12. Integrate file-based state with PoH blockchain
  - Modify `BlockHeader` struct to include StateRoot field
  - Implement state root calculation from file state
  - Integrate file-based transactions into PoH Entry structure
  - Update block validation to verify state transitions
  - Add state root verification to chain verification logic
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  - _Status: StateRoot field added to BlockHeader in `internal/blockchain/types.go`. FileStore includes CalculateStateRoot() method. Entry structure already supports FileTransactions field._

- [x] 13. Implement end-to-end transaction flow
  - Create integration test: account creation → balance transfer → verification
  - Test multi-instruction transaction atomicity
  - Test transaction revert on instruction failure
  - Test fee payment and balance updates
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5, 10.1, 10.2, 10.3, 10.4, 10.5_
  - _Status: Completed in `internal/e2e_transaction_test.go` with 4 comprehensive tests_

- [x] 14. Implement access control validation tests
  - Test read permission enforcement
  - Test write permission enforcement
  - Test undeclared file access detection and rejection
  - Test permission violation handling
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Status: Completed in `internal/access_control_integration_test.go` with 4 comprehensive tests_

- [x] 15. Implement storage cost enforcement tests
  - Test file creation with insufficient balance
  - Test data allocation exceeding balance
  - Test balance reduction below storage cost requirement
  - Test exponential cost growth for large files
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_
  - _Status: Completed in `internal/storage_cost_integration_test.go` with 4 comprehensive tests covering all storage cost scenarios_

- [x] 16. Add parallel execution preparation
  - Implement `ParallelAnalyzer` struct with dependency analysis methods
  - Implement `AnalyzeDependencies()` to build transaction dependency graph
  - Implement `HasConflict()` to detect read/write conflicts between transactions
  - Implement `FindParallelBatches()` to identify non-conflicting transaction sets
  - Add documentation for future parallel execution implementation
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [x] 17. Create CLI tool for file-based operations
  - Implement account creation command
  - Implement balance transfer command
  - Implement file query command
  - Implement transaction submission command
  - Add transaction status checking
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3, 10.4, 10.5_


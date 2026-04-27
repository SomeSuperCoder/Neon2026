# Implementation Plan

- [x] 1. Create transaction builder helper
  - Implement TransactionBuilder for constructing transactions with proper input declarations
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 1.1 Create internal/transaction/builder.go
  - Define TransactionBuilder struct with lastSeen, instructions, signatures fields
  - Implement NewTransactionBuilder(lastSeen TxID) constructor
  - Implement AddTransferInstruction() to create transfer instructions with proper inputs
  - Implement AddSignature() to add signatures to the transaction
  - Implement Build() to construct the final Transaction
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 1.2 Write tests for transaction builder
  - Test building transfer transaction with correct input declarations
  - Test validation of missing fields
  - Test multiple instructions in one transaction
  - Test signature handling
  - Document TransactionBuilder API in docs/reference/transaction-builder.md
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 2. Create System Program helper functions
  - Implement helper functions for creating System Program instructions
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 2.1 Create internal/transaction/system_helpers.go
  - Implement CreateTransferInstruction() with proper input declarations
  - Implement EncodeTransferData() for instruction data encoding
  - Define standard input key constants (InputKeyProgram, InputKeySender, etc.)
  - Validate amount is positive before creating instruction
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 2.2 Write tests for system helpers
  - Test CreateTransferInstruction with valid parameters
  - Test rejection of zero/negative amounts
  - Test correct encoding of transfer data
  - Test correct permission declarations (program=Read, sender/receiver=Write)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 3. Create instruction input validator
  - Implement validation logic for instruction inputs before execution
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 3.1 Create internal/processor/input_validator.go
  - Define InputValidator struct with fileStore field
  - Implement NewInputValidator(fs *filestore.FileStore) constructor
  - Implement ValidateInstructionInputs() to check all input declarations
  - Implement ValidateExecutableProgram() to verify program has Executable flag
  - Implement ValidateFileExists() to check file existence
  - Define custom error types (ErrMissingInput, ErrInvalidPermission, etc.)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 3.2 Write tests for input validator
  - Test validation of properly declared inputs
  - Test detection of missing program declaration
  - Test detection of non-executable program
  - Test detection of missing files
  - Test detection of incorrect permissions
  - Test validation of empty inputs map
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 4. Integrate input validation into TxProcessor
  - Update transaction processor to validate inputs before execution
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 4.1 Update internal/processor/tx_processor.go
  - Add inputValidator field to TxProcessor struct
  - Initialize InputValidator in NewTxProcessor()
  - Update ExecuteInstruction() to call ValidateInstructionInputs() before execution
  - Ensure validation happens before any state modifications
  - Return validation errors immediately without executing program
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 4.2 Update processor tests
  - Update existing tests to include proper input declarations
  - Add tests for validation failure scenarios
  - Verify no state changes on validation failure
  - Test that validation happens before program execution
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 5. Update existing tests to use proper input declarations
  - Fix all existing tests to declare inputs correctly
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5.1 Update internal/quanticscript/system_program_test.go ✓ COMPLETED
  - Add proper input declarations to all test instructions
  - Declare System Program with Read permission
  - Declare sender/receiver/payer files with Write permission
  - Use CreateTransferInstruction helper where applicable
  - All test cases now include proper "program" input declaration
  - Tests passing successfully
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5.2 Update internal/builtin_programs_integration_test.go
  - Add proper input declarations to all test instructions
  - Update helper functions to include input declarations
  - Verify tests pass with new validation
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5.3 Update internal/e2e_transaction_test.go
  - Add proper input declarations to all test transactions
  - Use TransactionBuilder where applicable
  - Verify end-to-end flow works with validation
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5.4 Update internal/processor/tx_processor_test.go
  - Add proper input declarations to all test instructions
  - Add tests for validation error scenarios
  - Verify rollback works when validation fails
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 6. Add integration tests for validation
  - Create comprehensive integration tests for the validation flow
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 6.1 Create internal/validation_integration_test.go
  - Test end-to-end transfer with proper input declarations
  - Test transaction rejection with missing program declaration
  - Test transaction rejection with incorrect permissions
  - Test transaction rejection with non-existent files
  - Test transaction rejection with non-executable program
  - Verify no state changes on validation failure
  - Test multi-instruction transactions with validation
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 7. Update CLI to use transaction builder
  - Modify CLI commands to use the new builder pattern
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [ ] 7.1 Update cmd/main.go transfer command
  - Replace manual transaction construction with TransactionBuilder
  - Use CreateTransferInstruction helper
  - Ensure proper input declarations in all CLI operations
  - Update error messages to guide users on proper usage
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [ ] 7.2 Update cmd/main.go account create command
  - Use proper input declarations for CREATE_FILE instructions
  - Declare System Program, payer, and new file with correct permissions
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 8. Run full test suite and verify
  - Ensure all tests pass with new validation
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 8.1 Run go test ./...
  - Execute full test suite
  - Verify all tests pass
  - Check for any compilation errors
  - Fix any remaining test failures
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ] 8.2 Test CLI operations manually
  - Test account create command
  - Test transfer command
  - Test query command
  - Verify proper error messages on validation failures
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 5.1, 5.2, 5.3, 5.4, 5.5_

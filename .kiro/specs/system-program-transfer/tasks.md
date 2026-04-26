# Implementation Plan

- [x] 1. Add transfer opcode to the interpreter
  - Add OpTransfer constant (0x7A) to opcodes.go
  - Add "TRANSFER" to OpcodeNames map
  - Implement execTransfer() method in interpreter.go with full validation
  - Add transfer opcode to executeInstruction() switch statement
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 2. Extend QuanticScript language with helper functions
  - [x] 2.1 Implement slice() builtin function
    - Add builtinSlice to stdlib.go
    - Register slice in builtin function registry
    - Handle bounds checking and error cases
    - _Requirements: 4.1, 4.4_
  
  - [x] 2.2 Implement bytesToFileID() builtin function
    - Add builtinBytesToFileID to stdlib.go
    - Register bytesToFileID in builtin function registry
    - Validate 32-byte length requirement
    - _Requirements: 4.1, 4.4_
  
  - [x] 2.3 Add transfer() builtin function
    - Add builtinTransfer to stdlib.go
    - Register transfer in builtin function registry
    - Map to OpTransfer opcode in code generator
    - _Requirements: 1.1, 4.2_

- [x] 3. Rewrite System Program with new transfer instruction
  - Update programs/system/system.qs with new implementation
  - Implement instruction dispatch logic (INSTR_CREATE_ACCOUNT, INSTR_TRANSFER, INSTR_ALLOCATE_SPACE)
  - Implement handleCreateAccount using transfer()
  - Implement handleTransfer using transfer()
  - Implement handleAllocateSpace using transfer()
  - Add proper error codes and validation
  - Compile to system.qsa and system.qsb
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 4. Write comprehensive tests for transfer opcode
  - [x] 4.1 Write interpreter unit tests
    - Test successful transfer between owned files
    - Test ownership violation rejection
    - Test storage cost enforcement
    - Test negative amount rejection
    - Test zero amount rejection
    - Test overflow protection
    - Test FileID type conversions (i64, bytes, FileID)
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 6.1, 6.2, 6.3, 6.4_
  
  - [x] 4.2 Write stdlib helper function tests
    - Test slice() with valid ranges
    - Test slice() with invalid ranges
    - Test bytesToFileID() with valid input
    - Test bytesToFileID() with invalid length
    - _Requirements: 6.1, 6.4_
  
  - [x] 4.3 Write System Program integration tests
    - Test CREATE_ACCOUNT instruction
    - Test TRANSFER instruction
    - Test ALLOCATE_SPACE instruction
    - Test error code returns
    - Test instruction parsing and validation
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 6.5_
  
  - [x] 4.4 Write end-to-end transfer integration tests
    - Test transfer between files owned by same program
    - Test transfer rejection from non-owned file
    - Test storage cost constraint enforcement
    - Test cross-program scenarios
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 6.1, 6.2, 6.3, 6.4_

- [x] 5. Deprecate and remove updateBalance opcode
  - Mark OpUpdateBalance as deprecated in opcodes.go
  - Update execUpdateBalance to return deprecation error
  - Remove all uses of updateBalance from codebase
  - Update documentation to reflect deprecation
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 6. Update instruction cost table
  - Add OpTransfer cost to costs.go (6110 units based on design)
  - Verify cost is appropriate for operation complexity
  - _Requirements: 1.3_

- [x] 7. Run full test suite and verify all tests pass
  - Run go test ./internal/quanticscript/...
  - Run integration tests
  - Verify no regressions in existing functionality
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

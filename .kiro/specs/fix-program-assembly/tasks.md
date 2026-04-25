# Implementation Plan

## Phase 1: Instruction Dispatch Registry (DX Improvement)

- [x] 1. Implement instruction dispatch registry in the stdlib and compiler
  - Define `InstructionDef` struct with `Code int`, `Name string`, `Args []ArgDef{Name, Type, Offset, Length}`, `Handler string`
  - Define `ArgType` enum: `ArgTypeI64`, `ArgTypeU64`, `ArgTypeBytes`, `ArgTypeBool`
  - Implement `ParseArgs(instrData []byte, def InstructionDef) (map[string]Value, error)` with full bounds checking
  - Implement `Dispatch(instrData []byte, registry map[int]InstructionDef) (InstructionDef, map[string]Value, error)`
  - Register all System_Program instruction schemas (CREATE_ACCOUNT, TRANSFER, ALLOCATE_SPACE) with arg offsets and types
  - Register all Token_Program instruction schemas (all 11 types) with arg offsets and types
  - Add `OpDispatch` opcode (`DISPATCH`) to `opcodes.go` — pops raw instruction bytes, looks up registry, pushes parsed args as structured value
  - Implement `OpDispatch` in `interpreter.go` — all parsing logic lives here, assembly emits a single `DISPATCH` call
  - Add cost entry in `costs.go` and update assembler to recognize `DISPATCH` mnemonic
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 1.1 Write unit tests for the instruction dispatch registry
  - Test `ParseArgs` correctly extracts all arg types from known byte sequences
  - Test `Dispatch` returns correct `InstructionDef` for each registered instruction code
  - Test `Dispatch` returns error for unknown instruction codes
  - Test bounds checking catches truncated instruction data
  - Test all System_Program instruction schemas parse correctly
  - Test all Token_Program instruction schemas parse correctly
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

## Phase 2: System_Program Assembly

- [ ] 2. Write minimal System_Program assembly
  - Entry point calls `GETINSTRDATA` then `DISPATCH`, then branches to handler label
  - Write `handle_create_account` label: pop args, call `HASSIGNER`, `CREATEFILE`, return 0 or error code
  - Write `handle_transfer` label: pop args, call `HASSIGNER`, `GETBALANCE`, `UPDATEBALANCE` x2, return 0 or error code
  - Write `handle_allocate_space` label: pop args, call `HASSIGNER`, `GETBALANCE`, `GETFILE`, `UPDATEBALANCE`, return 0 or error code
  - No helper functions, no byte arithmetic — only blockchain opcodes and control flow
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 2.1 Assemble System_Program to bytecode
  - Run assembler on system.qsa to generate system.qsb
  - Verify bytecode is valid and loadable
  - Verify bytecode size is significantly larger than stub version
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 2.2 Write unit tests for System_Program assembly
  - Test instruction parsing and validation
  - Test handleCreateAccount with valid and invalid inputs
  - Test handleTransfer with balance checks and authorization
  - Test handleAllocateSpace with storage cost calculation
  - Test error codes for all error conditions
  - Test edge cases (zero amounts, max values, overflow)
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

## Phase 3: Token_Program Assembly

- [ ] 3. Write minimal Token_Program assembly
  - Entry point calls `GETINSTRDATA` then `DISPATCH`, then branches to handler label
  - Write `handle_initialize_mint`: pop args, call `CREATEFILE` with serialized mint data, return 0 or error
  - Write `handle_mint_to`: pop args, validate mint authority via `HASSIGNER`, update token balance and supply, return 0 or error
  - Write `handle_initialize_account`: pop args, verify mint exists via `GETFILE`, call `CREATEFILE`, return 0 or error
  - Write `handle_transfer`: pop args, check frozen flag, validate owner/delegate via `HASSIGNER`, update balances, return 0 or error
  - Write `handle_burn`: pop args, validate owner, decrease balance and supply, return 0 or error
  - Write `handle_freeze_account` and `handle_thaw_account`: pop args, validate freeze authority, update frozen flag, return 0 or error
  - Write `handle_approve` and `handle_revoke`: pop args, validate owner, update delegate fields, return 0 or error
  - Write `handle_create_associated_token_account`: pop args, derive address via `SHA256`, call `CREATEFILE`, return 0 or error
  - Write `handle_close_account`: pop args, validate owner, verify zero balance, call `TRANSFERBALANCE` then `DELETEFILE`, return 0 or error
  - No helper functions, no byte arithmetic — only blockchain opcodes and control flow
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 3.1 Assemble Token_Program to bytecode
  - Run assembler on token.qsa to generate token.qsb
  - Verify bytecode is valid and loadable
  - Verify bytecode size is significantly larger than stub version
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 3.2 Write unit tests for Token_Program assembly
  - Test instruction parsing and validation
  - Test handleInitializeMint with various authority configurations
  - Test handleMintTo with supply tracking
  - Test handleInitializeAccount with mint verification
  - Test handleTransfer with balance checks and delegate support
  - Test handleBurn with supply reduction
  - Test freeze/thaw state management
  - Test approve/revoke delegate operations
  - Test associated token account creation and derivation
  - Test handleCloseAccount with balance reclamation
  - Test error codes for all error conditions
  - Test edge cases (overflow, underflow, authorization failures)
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

## Phase 4: Integration and Validation

- [ ] 4. Verify bytecode loads at genesis
  - Confirm system.qsb and token.qsb are embedded in programs/embed.go
  - Verify bytecode loads correctly via genesis.LoadBuiltinPrograms
  - Verify program IDs are correctly registered
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 4.1 Run integration tests
  - Execute System_Program bytecode with test transactions
  - Execute Token_Program bytecode with test transactions
  - Verify correct results for all operations
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

- [ ] 4.2 Verify determinism
  - Execute bytecode multiple times with same inputs
  - Verify results are identical each time
  - Verify cost calculations are consistent
  - _Requirements: 5.4_

- [ ] 4.3 Validate error handling
  - Test all error conditions return correct error codes
  - Verify no state corruption on errors
  - Verify authorization checks cannot be bypassed
  - Verify balance/supply consistency maintained
  - _Requirements: 5.2, 5.3, 5.6, 5.7_

- [ ] 4.4 Run bytecode verification tests
  - Verify bytecode structure and instruction validity
  - Verify bytecode disassembles to readable assembly
  - Verify round-trip assembly → bytecode → disassembly
  - Verify cost annotations are accurate
  - _Requirements: 3.1, 3.2, 3.3_

- [ ] 4.5 Production readiness review
  - Confirm all source code functionality is implemented
  - Confirm all error conditions are handled
  - Confirm all edge cases are covered
  - Confirm determinism is verified
  - Confirm bytecode size reflects complete implementation
  - Confirm all tests passing
  - Confirm no state corruption or balance inconsistencies
  - Confirm authorization checks cannot be bypassed
  - Mark as ready for mainnet deployment
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

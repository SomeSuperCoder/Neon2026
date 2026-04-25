# Implementation Plan

- [x] 1. Delete the Go system package
  - Remove `internal/system/system.go` and `internal/system/system_test.go`
  - _Requirements: 1.3, 1.4, 4.4_

- [x] 2. Update `cmd/main.go` to remove all system package dependencies
- [x] 2.1 Remove the `internal/system` import and `rt.RegisterBuiltinProgram` calls
  - Delete the `"github.com/poh-blockchain/internal/system"` import
  - Remove the three `rt.RegisterBuiltinProgram(system.NewSystemProgram())` calls in `handleAccountCommand`, `handleTransferCommand`, and `handleSubmitCommand`
  - _Requirements: 1.3, 3.1, 3.2, 3.3_

- [x] 2.2 Replace `system.SystemProgramID` references with `genesis.SystemProgramID`
  - Update `handleAccountCommand` where `TxManager: system.SystemProgramID` is set
  - _Requirements: 2.1, 2.2, 2.3, 3.1_

- [x] 2.3 Inline the transfer instruction encoder and remove `system.EncodeTransferInstruction`
  - Add a local `encodeTransferInstruction(amount int64) []byte` helper in `cmd/main.go`
  - Replace the call to `system.EncodeTransferInstruction` in `handleTransferCommand`
  - Update the `ProgramID` in the transfer instruction to use `genesis.SystemProgramID`
  - _Requirements: 3.2, 3.4_

- [x] 2.4 Remove `ensureSystemProgram` and its call from `initFileStore`
  - Delete the `ensureSystemProgram` function body
  - Remove the `ensureSystemProgram(fs)` call inside `initFileStore`
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 3. Update integration tests to remove system package dependencies
- [x] 3.1 Update `internal/builtin_programs_integration_test.go`
  - Remove the `"github.com/poh-blockchain/internal/system"` import
  - Remove `rt.RegisterBuiltinProgram(sysProg)` from `setupBuiltinEnv`
  - Replace `system.SystemProgramID` with `genesis.SystemProgramID` throughout
  - Replace `system.EncodeTransferInstruction(n)` calls with an inline helper
  - _Requirements: 4.2, 4.3_

- [x] 3.2 Update `internal/e2e_transaction_test.go`
  - Remove the `"github.com/poh-blockchain/internal/system"` import
  - Remove `rt.RegisterBuiltinProgram(sysProg)` calls
  - Remove the manual stub `sysProgramFile` creation blocks; add `genesis.LoadBuiltinPrograms` call to each test setup instead
  - Replace `system.SystemProgramID` with `genesis.SystemProgramID` throughout
  - Replace `system.EncodeTransferInstruction(n)` calls with an inline helper
  - Add the `"github.com/poh-blockchain/programs"` import needed for `genesis.LoadBuiltinPrograms`
  - _Requirements: 4.2, 4.3_

- [x] 3.3 Verify the full test suite passes
  - Run `go test ./...` and confirm no failures introduced by this change
  - _Requirements: 4.1_

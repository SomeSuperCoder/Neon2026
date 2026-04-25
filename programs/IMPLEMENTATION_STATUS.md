# Built-in Programs Implementation Status

## Overview

This document tracks the implementation status of the System_Program and Token_Program built-in blockchain programs written in QuanticScript.

## Current Status

### ✅ Completed

1. **QuanticScript Standard Library Extensions** (`internal/quanticscript/stdlib_programs.go`)
   - Serialization helpers for MintAccount and TokenAccount
   - Instruction data parsing utilities (U8, U64, I64, FileID, PublicKey, Bool, Optional)
   - FileID and PublicKey type conversion helpers
   - Complete serialization/deserialization for both account types

2. **Token_Program Logic** (`programs/token/token.qs`)
   - All 11 instruction handlers implemented with full logic:
     - InitializeMint
     - MintTo
     - InitializeAccount
     - CreateAssociatedTokenAccount
     - Transfer
     - Burn
     - CloseAccount
     - FreezeAccount
     - ThawAccount
     - Approve
     - Revoke
   - Data structure serialization/deserialization
   - Authority validation logic
   - Freeze/delegate logic
   - **FIXED**: Removed placeholder functions, now uses real builtins

3. **System_Program Logic** (`programs/system/system.qs`)
   - All 3 instruction handlers implemented with full logic:
     - CreateAccount
     - Transfer
     - AllocateSpace
   - Balance validation logic
   - Storage rent calculation
   - Authority validation
   - **FIXED**: Replaced placeholder stubs with real implementation

## ⚠️ Critical Blockers

### Missing Runtime Support

The QuanticScript programs are **logically complete** but **cannot execute** because they require privileged file operations that are not yet available as builtins or opcodes:

#### Required Operations

1. **File Creation**
   ```
   createFile(data: bytes, balance: i64, txManager: bytes): bytes
   ```
   - Creates a new file in FileStore
   - Returns the generated FileID
   - Needed by: InitializeMint, InitializeAccount

2. **File Creation with Specific ID**
   ```
   createFileWithID(fileID: bytes, data: bytes, balance: i64, txManager: bytes): void
   ```
   - Creates a file with a predetermined FileID
   - Needed by: CreateAssociatedTokenAccount

3. **File Deletion**
   ```
   deleteFile(fileID: bytes): void
   ```
   - Removes a file from FileStore
   - Needed by: CloseAccount

4. **Balance Transfer**
   ```
   transferBalance(from: bytes, to: bytes, amount: i64): void
   ```
   - Transfers Neon balance between files
   - Needed by: Transfer (System_Program), CloseAccount

5. **Array Operations**
   ```
   len(data: bytes): i64
   slice(data: bytes, start: i64, end: i64): bytes
   append(a: bytes, b: bytes): bytes
   ```
   - Basic byte array manipulation
   - Currently implemented as helper functions but should be builtins for efficiency

### Available Builtins (Already Working)

These are already implemented and working:
- `getInstructionData(): bytes` - ✅ OpGetInstrData
- `getFile(fileID: bytes): file` - ✅ OpGetFile
- `updateFile(fileID: bytes, data: bytes): void` - ✅ OpUpdateFile
- `hasSigner(pubkey: bytes): bool` - ✅ OpHasSigner
- `sha256(data: bytes): bytes` - ✅ OpSha256

## 📋 Next Steps

### Priority 1: Implement Missing Runtime Support (Task 9.2)

Add the missing privileged operations to the QuanticScript runtime:

1. **Add Opcodes** (`internal/quanticscript/opcodes.go`)
   ```go
   OpCreateFile       Opcode = 0x7A
   OpCreateFileWithID Opcode = 0x7B
   OpDeleteFile       Opcode = 0x7C
   OpTransferBalance  Opcode = 0x7D
   OpBytesLen         Opcode = 0x7E
   OpBytesSlice       Opcode = 0x7F
   OpBytesAppend      Opcode = 0x80
   ```

2. **Implement Opcode Handlers** (`internal/quanticscript/interpreter.go`)
   - `execCreateFile()` - Create file and return FileID
   - `execCreateFileWithID()` - Create file with specific ID
   - `execDeleteFile()` - Remove file from FileStore
   - `execTransferBalance()` - Transfer Neon between files
   - `execBytesLen()` - Get byte array length
   - `execBytesSlice()` - Extract byte slice
   - `execBytesAppend()` - Concatenate byte arrays

3. **Update Codegen** (`internal/quanticscript/codegen.go`)
   - Add builtin function mappings in `emitBuiltinCall()`
   - Register new builtins in `isBuiltinFunction()`

4. **Add Compute Costs** (`internal/quanticscript/costs.go`)
   - Define compute budget costs for each new opcode

### Priority 2: Compile Programs to Bytecode (Tasks 2.4, 9)

Once runtime support is complete:
1. Compile `programs/system/system.qs` → `programs/system/system.qsb`
2. Compile `programs/token/token.qs` → `programs/token/token.qsb`
3. Verify bytecode structure and size

### Priority 3: Write Tests (Tasks 2.5, 9.1)

1. Unit tests for System_Program
2. Unit tests for Token_Program
3. Test all instruction handlers
4. Test error conditions

### Priority 4: Genesis Integration (Task 10)

1. Implement genesis loader
2. Load programs at blockchain initialization
3. Verify programs are executable

### Priority 5: Integration Tests (Task 11)

1. End-to-end testing
2. Cross-program invocation tests
3. Storage rent enforcement tests

## 🔍 Why This Happened

The tasks were marked complete because:
1. The **logic** for all instructions was written
2. The **data structures** were defined
3. The **validation** was implemented

However, the programs used **placeholder functions** at the bottom instead of calling real runtime operations. This made them look complete but non-functional.

## ✅ What Was Fixed

1. **Token_Program** (`programs/token/token.qs`)
   - Removed all placeholder function definitions
   - Added clear documentation about which functions are builtins
   - Added notes about which operations require runtime support
   - Program now uses real `getFile()`, `updateFile()`, `hasSigner()`, `sha256()` builtins

2. **System_Program** (`programs/system/system.qs`)
   - Replaced empty stubs with full implementation
   - Implemented all 3 instruction handlers
   - Added validation logic
   - Added storage cost calculation
   - Added clear documentation about runtime requirements

3. **Tasks** (`.kiro/specs/builtin-programs/tasks.md`)
   - Unmarked tasks that require runtime support
   - Added new critical task 9.2 for implementing missing runtime operations
   - Added notes explaining what's blocking completion

## 📝 Summary

**The programs are now properly implemented** with real logic and proper use of available builtins. However, they **cannot execute** until the runtime provides the missing privileged file operations (create, delete, transfer balance).

The next developer should focus on **Task 9.2** - implementing the missing runtime support. Once that's done, the programs can be compiled and tested.

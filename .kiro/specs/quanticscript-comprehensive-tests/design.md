# Design Document

## Overview

This design outlines a comprehensive test suite for QuanticScript that validates both general-purpose programming capabilities and blockchain-specific features. The test suite consists of small, isolated Go test files that directly test the interpreter with hand-crafted bytecode, ensuring we can identify issues at the lowest level of the language implementation.

The tests are designed to be:
- Small and focused on specific features
- Isolated from each other to prevent cascading failures
- Practical and representative of real-world usage
- Easy to debug when failures occur

## Architecture

### Test Organization

Tests will be organized in a single new file: `internal/quanticscript/comprehensive_test.go`

This file will contain multiple test functions, each focused on a specific capability area:
- `TestControlFlow` - if/else, loops, early returns
- `TestDataStructures` - arrays, maps, nested operations
- `TestArithmetic` - integer operations, comparisons, logical operations
- `TestBlockchainState` - file operations, balance queries
- `TestCrossProgram` - program invocation, depth tracking
- `TestStringOps` - concatenation, substring, conversions
- `TestCryptoOps` - hashing, signature verification
- `TestFunctions` - function calls, recursion, parameter passing

### Test Approach

Each test will:
1. Create a `MockExecutionContext` with necessary test data
2. Hand-craft minimal bytecode that exercises the feature
3. Create a `BytecodeInterpreter` instance
4. Execute the bytecode
5. Verify the results by inspecting the stack or context state

This approach allows us to test the interpreter directly without depending on the compiler, lexer, or parser, isolating issues to the execution layer.

## Components and Interfaces

### MockExecutionContext Enhancement

The existing `MockExecutionContext` in `interpreter_test.go` will be extended to support:
- Cross-program invocation simulation
- Declared program list management
- More realistic file and balance operations

### Bytecode Helper Functions

New helper functions will be created to simplify bytecode construction:
- `buildPushI64(value int64) []byte` - Build PUSH instruction for i64
- `buildPushU64(value uint64) []byte` - Build PUSH instruction for u64
- `buildPushBool(value bool) []byte` - Build PUSH instruction for bool
- `buildPushString(value string) []byte` - Build PUSH instruction for string
- `buildPushBytes(value []byte) []byte` - Build PUSH instruction for bytes
- `buildJump(offset int32) []byte` - Build JMP instruction
- `buildJumpIf(offset int32) []byte` - Build JMPIF instruction

These helpers will make test bytecode more readable and maintainable.

## Data Models

### Test Case Structure

Each test function will follow this pattern:

```go
func TestFeatureName(t *testing.T) {
    // Setup
    ctx := NewMockExecutionContext()
    // ... configure context ...
    
    // Build bytecode
    bytecode := []byte{
        // ... instructions ...
    }
    
    // Execute
    interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
    err := interp.Execute()
    
    // Verify
    if err != nil {
        t.Fatalf("execution failed: %v", err)
    }
    
    // Check results
    result, err := interp.pop()
    if err != nil {
        t.Fatalf("failed to get result: %v", err)
    }
    
    // Assert expected values
    // ...
}
```

### Test Scenarios

#### 1. Control Flow Tests

**If-Else with Multiple Conditions:**
- Test nested if-else statements
- Test early returns from conditional branches
- Verify correct branch execution based on runtime values

**While Loop with Break/Continue:**
- Test loop iteration with counter
- Test break statement exits loop correctly
- Test continue statement skips to next iteration

#### 2. Data Structure Tests

**Array Operations:**
- Create array, push elements, get by index
- Test array length operation
- Test array pop operation
- Verify element ordering

**Map Operations:**
- Create map, set key-value pairs
- Test map get operation
- Test map has operation
- Test map delete operation

**Nested Structures:**
- Create array of maps
- Access nested elements
- Modify nested values

#### 3. Arithmetic Tests

**Integer Arithmetic:**
- Test addition, subtraction, multiplication
- Test division with non-zero divisor
- Test modulo operation
- Verify overflow behavior

**Comparisons:**
- Test equality (==) for numbers and booleans
- Test less than (<), greater than (>)
- Test less than or equal (<=), greater than or equal (>=)

**Logical Operations:**
- Test AND, OR, NOT operations
- Verify short-circuit evaluation (if applicable)
- Test compound boolean expressions

#### 4. Blockchain State Tests

**File Operations:**
- Create test file in context
- Read file data with GETFILE
- Modify file data with UPDATEFILE
- Verify changes persisted

**Balance Operations:**
- Query file balance with GETBALANCE
- Test balance queries for non-existent files

**Signer Operations:**
- Test GETSIGNER with valid index
- Test HASSIGNER with present/absent keys
- Test GETINSTRDATA returns correct data

#### 5. Cross-Program Invocation Tests

**Basic Invocation:**
- Mock a target program in context
- Invoke with valid program ID and data
- Verify result returned correctly

**Depth Tracking:**
- Test invocation depth increments
- Test maximum depth limit enforcement
- Verify depth resets after return

**Permission Validation:**
- Test invocation fails for undeclared programs
- Test compute budget deduction
- Test insufficient budget handling

#### 6. String Operation Tests

**Concatenation:**
- Concatenate two strings
- Verify result is correct combination

**Substring:**
- Extract substring with start and end indices
- Test boundary conditions
- Test invalid indices handling

**Length and Conversions:**
- Test string length operation
- Test string to bytes conversion
- Test bytes to string conversion

#### 7. Cryptographic Operation Tests

**Hashing:**
- Hash known input data
- Verify deterministic output
- Test with empty input

**Signature Verification:**
- Create valid signature for test data
- Verify signature passes validation
- Test invalid signature fails validation

**Public Key Derivation:**
- Derive public key from seed
- Verify deterministic derivation

#### 8. Function Call Tests

**Basic Function Calls:**
- Define function at specific offset
- Call function with CALL instruction
- Return from function with RET
- Verify return address handling

**Recursive Calls:**
- Implement simple recursive function (e.g., factorial)
- Test recursion depth tracking
- Verify stack overflow protection

**Parameter Passing:**
- Pass parameters via stack
- Access parameters in function
- Return values via stack

## Error Handling

Each test will verify proper error handling:
- Stack underflow/overflow conditions
- Type mismatches in operations
- Out-of-bounds access for arrays and memory
- Invalid file IDs or program IDs
- Compute budget exhaustion
- Maximum invocation depth exceeded

Tests will use `t.Fatalf()` for unexpected errors and verify expected errors are returned correctly.

## Testing Strategy

### Unit Testing Approach

Each test function is a unit test that exercises a specific feature in isolation. Tests do not depend on each other and can run in any order.

### Test Data

Test data will be minimal and embedded directly in test functions:
- Small integers for arithmetic tests
- Short strings for string operation tests
- Mock file IDs using simple patterns (e.g., all zeros except last byte)
- Test public keys and signatures using known test vectors

### Verification Methods

Tests will verify correctness by:
1. Checking the interpreter's stack state after execution
2. Inspecting the mock context for state changes
3. Verifying error messages for negative test cases
4. Checking compute budget consumption

### Coverage Goals

The test suite aims to cover:
- All bytecode opcodes at least once
- Common operation combinations
- Edge cases (empty arrays, zero values, boundary conditions)
- Error conditions for each operation type

### Test Execution

Tests will be run using standard Go testing:
```bash
go test -v ./internal/quanticscript -run Comprehensive
```

This allows running just the comprehensive tests without running all other QuanticScript tests.

## Implementation Notes

### Bytecode Construction

Bytecode will be constructed manually using byte slices. While verbose, this approach:
- Doesn't depend on the compiler being correct
- Makes it obvious what instructions are being tested
- Allows testing of edge cases and invalid bytecode

### Mock Context Behavior

The `MockExecutionContext` will be enhanced to:
- Support registering "programs" that can be invoked
- Track invocation history for verification
- Simulate realistic file storage behavior
- Provide deterministic responses for queries

### Test Isolation

Each test function will:
- Create its own context and interpreter
- Not share state with other tests
- Clean up resources (though Go's GC handles this)
- Be independent and self-contained

## Dependencies

The test suite depends on:
- `internal/quanticscript` package (interpreter, types, opcodes)
- `internal/filestore` package (FileID, File types)
- `internal/transaction` package (PublicKey, TxID types)
- Standard Go `testing` package
- Existing `MockExecutionContext` from `interpreter_test.go`

No external dependencies are required.

## Future Enhancements

Potential future additions to the test suite:
- Performance benchmarks for each operation type
- Fuzz testing with random bytecode
- Property-based testing for arithmetic operations
- Integration tests that combine multiple features
- Tests for the full compilation pipeline (lexer → parser → codegen → interpreter)

These enhancements are out of scope for the initial implementation but provide a roadmap for more comprehensive testing.

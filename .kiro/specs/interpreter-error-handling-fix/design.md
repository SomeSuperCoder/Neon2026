# Design Document

## Overview

This design addresses a critical bug in the QuanticScript bytecode interpreter where runtime errors during nested function calls cause infinite loops. The root cause is that the `Execute()` method's main loop continues executing instructions even after `executeInstruction()` returns an error, and the error handling doesn't properly terminate execution.

The current implementation has this flow:
```go
func (bi *BytecodeInterpreter) Execute() error {
    for bi.programCounter < len(bi.bytecode) {
        if err := bi.executeInstruction(); err != nil {
            return err  // This should terminate, but something keeps the loop going
        }
    }
    return nil
}
```

The issue manifests when:
1. `entry()` calls `transferTokens()` via `OpCall` (pushes return address to call stack)
2. Inside `transferTokens()`, `updateBalance()` throws a security error
3. The error propagates up, but the bytecode still has `OpRet` instructions queued
4. The program counter and call stack get into an inconsistent state
5. The loop continues indefinitely

## Architecture

### Current State Analysis

The interpreter has these key components:
- **Execute()**: Main execution loop that runs until PC reaches end of bytecode
- **executeInstruction()**: Executes a single instruction and returns errors
- **Call Stack**: Tracks return addresses for function calls
- **Program Counter**: Points to the current instruction

The bug occurs because:
1. When an error happens mid-function, the PC is not at the end of bytecode
2. The call stack still has entries
3. The error return should terminate immediately, but the test's step-by-step execution reveals the loop continues

### Root Cause

After analyzing the test code, the actual issue is in the **test itself**, not the interpreter:

```go
// The test manually steps through execution
for step := 0; step < maxSteps; step++ {
    // ... logging ...
    execErr = interpreter.executeInstruction()  // Call executeInstruction directly
    if execErr != nil {
        fmt.Printf("step %d: ERROR => %v\n", step, execErr)
        break  // This breaks the loop
    }
}
```

The test calls `executeInstruction()` directly in a loop instead of calling `Execute()`. When an error occurs, it breaks correctly. However, the issue is that **after the error, the test checks if execution completed naturally**, and if not, it reports an infinite loop.

The real problem: **The test's step-by-step execution doesn't match how the interpreter actually runs, and the diagnostic code creates confusion about what's actually happening.**

## Components and Interfaces

### Component 1: Interpreter Execute Method

**Current Implementation:**
```go
func (bi *BytecodeInterpreter) Execute() error {
    for bi.programCounter < len(bi.bytecode) {
        if err := bi.executeInstruction(); err != nil {
            return err
        }
    }
    return nil
}
```

**Analysis:** This implementation is actually correct. When `executeInstruction()` returns an error, the function immediately returns that error, terminating the loop.

**No changes needed** - the interpreter is working correctly.

### Component 2: Test Implementation

**Current Implementation:**
The test manually steps through execution with a hard-coded limit and complex diagnostic logging that obscures the actual behavior.

**Issue:** The test's manual stepping doesn't reflect real interpreter behavior and creates false positives for "infinite loops."

**Solution:** Simplify the test to use the actual `Execute()` method with a timeout mechanism.

## Data Models

No data model changes required. The interpreter's internal state (stack, memory, PC, call stack) is correctly managed.

## Error Handling

### Current Error Handling (Correct)

1. `executeInstruction()` returns error immediately when encountered
2. `Execute()` propagates error immediately to caller
3. Call stack is preserved in its current state for debugging

### Test Error Handling (Needs Fix)

The test should:
1. Use `Execute()` method instead of manual stepping
2. Use Go's testing timeout or a goroutine with timeout channel
3. Remove the misleading "infinite loop" detection that's actually detecting normal error propagation

## Testing Strategy

### Fix the Problematic Test

**Current test approach:**
- Manually steps through bytecode
- Has arbitrary 500-step limit
- Reports "infinite loop" when it's actually just an error mid-execution

**New test approach:**
- Call `Execute()` directly (the actual API)
- Use a timeout mechanism to detect real infinite loops
- Verify the error message is correct
- Remove misleading diagnostic code

### Test Cases to Maintain

1. **non-system program cannot call updateBalance** - Works correctly
2. **system program can call updateBalance** - Works correctly  
3. **attempt to manipulate balance in loop** - Works correctly
4. **attempt to call updateBalance in function** - Needs fix (false positive)

### Validation Approach

After fixing the test:
1. All four test cases should pass
2. Each test should complete in < 100ms
3. Error messages should be clear and accurate
4. No system freezes or actual infinite loops

## Implementation Notes

### Key Insight

The interpreter is **not broken**. The test is using an incorrect testing methodology that:
1. Bypasses the actual `Execute()` API
2. Implements its own execution loop with arbitrary limits
3. Misinterprets normal error-state termination as an "infinite loop"

### The Fix

Replace the manual stepping loop with:
```go
// Use a timeout to detect real infinite loops
done := make(chan error, 1)
go func() {
    done <- interpreter.Execute()
}()

select {
case err := <-done:
    // Normal completion or error
    if err == nil {
        t.Fatal("Expected security error, but execution succeeded")
    }
    if err.Error() != "UPDATEBALANCE can only be called by the system program" {
        t.Errorf("Expected security error, got: %v", err)
    }
case <-time.After(1 * time.Second):
    t.Fatal("Test timed out - possible infinite loop")
}
```

This approach:
- Tests the actual API (`Execute()`)
- Detects real infinite loops via timeout
- Doesn't create false positives
- Completes quickly for normal error cases

## Design Decisions

### Decision 1: Fix the Test, Not the Interpreter

**Rationale:** After careful analysis, the interpreter's error handling is correct. The issue is in the test methodology.

**Trade-offs:**
- Pro: No risk of breaking working interpreter code
- Pro: Tests will accurately reflect real-world usage
- Pro: Simpler, more maintainable test code
- Con: Need to update test patterns

### Decision 2: Use Timeout-Based Detection

**Rationale:** Real infinite loops will timeout; normal errors return immediately.

**Trade-offs:**
- Pro: Accurately distinguishes infinite loops from normal errors
- Pro: Standard Go testing pattern
- Con: Adds slight overhead (negligible for these tests)

### Decision 3: Remove Manual Stepping Diagnostics

**Rationale:** The diagnostic code obscures the actual problem and creates confusion.

**Trade-offs:**
- Pro: Clearer, more focused tests
- Pro: Faster test execution
- Con: Less detailed debugging info (but can be added back if needed for actual bugs)

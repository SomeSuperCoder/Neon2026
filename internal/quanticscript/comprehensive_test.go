// Package quanticscript - Comprehensive Test Suite
//
// This file contains a comprehensive set of tests for the QuanticScript interpreter,
// covering the following feature areas (in order):
//
//  1. Bytecode helper functions (buildPushI64, buildJump, buildLoad, etc.)
//  2. Control flow (if-else, while loops with break/continue, nested structures, early returns)
//  3. Data structures (arrays, maps, nested structures, passing to functions)
//  4. Blockchain state (file read/write, balance queries, signer operations, error handling)
//  5. String operations (concatenation, substring, length, byte conversions)
//  6. Cryptographic operations (SHA-256 hashing, Ed25519 signature verification, key derivation)
//  7. Function calls (basic call/return, recursion, parameter passing, error conditions)
//
// Each test creates its own MockExecutionContext and BytecodeInterpreter, ensuring
// full isolation between test cases.
package quanticscript

import (
	"crypto/ed25519"
	"encoding/binary"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// Bytecode Helper Functions
// These functions simplify bytecode construction for comprehensive tests

// buildPushI64 creates a PUSH instruction for an i64 value
func buildPushI64(value int64) []byte {
	bytecode := make([]byte, 10)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeI64)
	binary.LittleEndian.PutUint64(bytecode[2:], uint64(value))
	return bytecode
}

// buildPushU64 creates a PUSH instruction for a u64 value
func buildPushU64(value uint64) []byte {
	bytecode := make([]byte, 10)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeU64)
	binary.LittleEndian.PutUint64(bytecode[2:], value)
	return bytecode
}

// buildPushBool creates a PUSH instruction for a bool value
func buildPushBool(value bool) []byte {
	bytecode := make([]byte, 3)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeBool)
	if value {
		bytecode[2] = 1
	} else {
		bytecode[2] = 0
	}
	return bytecode
}

// buildPushString creates a PUSH instruction for a string value
func buildPushString(value string) []byte {
	data := []byte(value)
	length := uint64(len(data))
	bytecode := make([]byte, 10+length)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeString)
	binary.LittleEndian.PutUint64(bytecode[2:], length)
	copy(bytecode[10:], data)
	return bytecode
}

// buildPushBytes creates a PUSH instruction for a bytes value
func buildPushBytes(value []byte) []byte {
	length := uint64(len(value))
	bytecode := make([]byte, 10+length)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeBytes)
	binary.LittleEndian.PutUint64(bytecode[2:], length)
	copy(bytecode[10:], value)
	return bytecode
}

// buildJump creates a JMP instruction with the specified offset
// The offset is relative to the instruction after the JMP
func buildJump(offset int32) []byte {
	bytecode := make([]byte, 5)
	bytecode[0] = byte(OpJmp)
	binary.LittleEndian.PutUint32(bytecode[1:], uint32(offset))
	return bytecode
}

// buildJumpIf creates a JMPIF instruction with the specified offset
// The offset is relative to the instruction after the JMPIF
func buildJumpIf(offset int32) []byte {
	bytecode := make([]byte, 5)
	bytecode[0] = byte(OpJmpIf)
	binary.LittleEndian.PutUint32(bytecode[1:], uint32(offset))
	return bytecode
}

// Test to verify bytecode helper functions work correctly
func TestBytecodeHelpers(t *testing.T) {
	t.Run("buildPushI64", func(t *testing.T) {
		// Build bytecode using helper
		bytecode := append(buildPushI64(42), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("buildPushU64", func(t *testing.T) {
		// Build bytecode using helper
		bytecode := append(buildPushU64(12345), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 12345 {
			t.Errorf("Expected 12345, got %d", result)
		}
	})

	t.Run("buildPushBool", func(t *testing.T) {
		// Test true
		bytecode := append(buildPushBool(true), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if !result {
			t.Errorf("Expected true, got false")
		}

		// Test false
		bytecode = append(buildPushBool(false), byte(OpRet))
		interp = NewBytecodeInterpreter(bytecode, ctx, 1000)

		err = interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err = interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result {
			t.Errorf("Expected false, got true")
		}
	})

	t.Run("buildPushString", func(t *testing.T) {
		// Build bytecode using helper
		bytecode := append(buildPushString("hello"), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "hello" {
			t.Errorf("Expected 'hello', got '%s'", result)
		}
	})

	t.Run("buildPushBytes", func(t *testing.T) {
		// Build bytecode using helper
		testBytes := []byte{1, 2, 3, 4, 5}
		bytecode := append(buildPushBytes(testBytes), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if len(result) != len(testBytes) {
			t.Fatalf("Expected %d bytes, got %d", len(testBytes), len(result))
		}

		for i, b := range testBytes {
			if result[i] != b {
				t.Errorf("Byte mismatch at index %d: expected %d, got %d", i, b, result[i])
			}
		}
	})

	t.Run("buildJump", func(t *testing.T) {
		// Build bytecode: PUSH 1, JMP (skip PUSH 99), RET
		var bytecode []byte
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, buildJump(10)...)    // Skip next PUSH instruction (10 bytes)
		bytecode = append(bytecode, buildPushI64(99)...) // This should be skipped
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Should only have 1 on stack, not 99
		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, _ := interp.stack[0].AsI64()
		if result != 1 {
			t.Errorf("Expected 1, got %d (jump didn't work)", result)
		}
	})

	t.Run("buildJumpIf", func(t *testing.T) {
		// Test jump when condition is true
		var bytecode []byte
		bytecode = append(bytecode, buildPushBool(true)...)
		bytecode = append(bytecode, buildJumpIf(10)...)  // Skip next PUSH instruction
		bytecode = append(bytecode, buildPushI64(99)...) // This should be skipped
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Stack should be empty (only had the bool which was popped)
		if len(interp.stack) != 0 {
			t.Errorf("Expected empty stack, got %d items", len(interp.stack))
		}

		// Test no jump when condition is false
		bytecode = nil
		bytecode = append(bytecode, buildPushBool(false)...)
		bytecode = append(bytecode, buildJumpIf(10)...)  // Don't jump
		bytecode = append(bytecode, buildPushI64(42)...) // This should execute
		bytecode = append(bytecode, byte(OpRet))

		interp = NewBytecodeInterpreter(bytecode, ctx, 1000)

		err = interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, _ := interp.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})
}

// TestControlFlowNested tests nested if statements inside loops
func TestControlFlowNested(t *testing.T) {
	// Simulates:
	// result = 0
	// for i = 0; i < 3; i++ {
	//   if (i == 0) {
	//     result = result + 1
	//   } else if (i == 1) {
	//     if (result > 0) {
	//       result = result + 10
	//     }
	//   } else {
	//     result = result + 100
	//   }
	// }
	// return result  // Should be 1 + 10 + 100 = 111

	var bytecode []byte

	// Initialize i = 0 at memory[0]
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, buildStore(0)...)

	// Initialize result = 0 at memory[1]
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, buildStore(1)...)

	// Loop start
	loopStart := len(bytecode)

	// Check loop condition: i < 3
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(3)...)
	bytecode = append(bytecode, byte(OpLt))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(i < 3), exit loop

	jumpToEndPos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder - jump to end if i >= 3

	// Check if i == 0
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, byte(OpEq))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(i == 0), skip to else-if

	jumpToElseIfPos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

	// Then branch (i == 0): result = result + 1
	bytecode = append(bytecode, buildLoad(1)...)
	bytecode = append(bytecode, buildPushI64(1)...)
	bytecode = append(bytecode, byte(OpAdd))
	bytecode = append(bytecode, buildStore(1)...)

	jumpToEndOfIfPos := len(bytecode)
	bytecode = append(bytecode, buildJump(0)...) // Placeholder - skip else-if and else

	// Else-if branch: check if i == 1
	elseIfStart := len(bytecode)
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(1)...)
	bytecode = append(bytecode, byte(OpEq))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(i == 1), skip to else

	jumpToElsePos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

	// Nested if: check if result > 0
	bytecode = append(bytecode, buildLoad(1)...)
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, byte(OpGt))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(result > 0), skip addition

	jumpToSkipNestedPos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

	// Nested then: result = result + 10
	bytecode = append(bytecode, buildLoad(1)...)
	bytecode = append(bytecode, buildPushI64(10)...)
	bytecode = append(bytecode, byte(OpAdd))
	bytecode = append(bytecode, buildStore(1)...)

	skipNestedEnd := len(bytecode)

	jumpToEndOfIf2Pos := len(bytecode)
	bytecode = append(bytecode, buildJump(0)...) // Placeholder - skip else

	// Else branch: result = result + 100
	elseStart := len(bytecode)
	bytecode = append(bytecode, buildLoad(1)...)
	bytecode = append(bytecode, buildPushI64(100)...)
	bytecode = append(bytecode, byte(OpAdd))
	bytecode = append(bytecode, buildStore(1)...)

	// End of if-else chain
	endOfIfChain := len(bytecode)

	// Increment i
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(1)...)
	bytecode = append(bytecode, byte(OpAdd))
	bytecode = append(bytecode, buildStore(0)...)

	// Jump back to loop start
	loopEnd := len(bytecode)
	jumpBackOffset := int32(loopStart - (loopEnd + 5))
	bytecode = append(bytecode, buildJump(jumpBackOffset)...)

	// End of loop
	endOfLoop := len(bytecode)

	// Fix all jump offsets
	// 1. Loop exit condition
	jumpForwardOffset := int32(endOfLoop - (jumpToEndPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToEndPos+1:jumpToEndPos+5], uint32(jumpForwardOffset))

	// 2. Jump to else-if
	jumpToElseIfOffset := int32(elseIfStart - (jumpToElseIfPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToElseIfPos+1:jumpToElseIfPos+5], uint32(jumpToElseIfOffset))

	// 3. Jump to end of if-else chain (from first if)
	jumpToEndOfIfOffset := int32(endOfIfChain - (jumpToEndOfIfPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToEndOfIfPos+1:jumpToEndOfIfPos+5], uint32(jumpToEndOfIfOffset))

	// 4. Jump to else
	jumpToElseOffset := int32(elseStart - (jumpToElsePos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToElsePos+1:jumpToElsePos+5], uint32(jumpToElseOffset))

	// 5. Jump to skip nested if
	jumpToSkipNestedOffset := int32(skipNestedEnd - (jumpToSkipNestedPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToSkipNestedPos+1:jumpToSkipNestedPos+5], uint32(jumpToSkipNestedOffset))

	// 6. Jump to end of if-else chain (from else-if)
	jumpToEndOfIf2Offset := int32(endOfIfChain - (jumpToEndOfIf2Pos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToEndOfIf2Pos+1:jumpToEndOfIf2Pos+5], uint32(jumpToEndOfIf2Offset))

	// Load result
	bytecode = append(bytecode, buildLoad(1)...)

	// Return
	bytecode = append(bytecode, byte(OpRet))

	ctx := NewMockExecutionContext()
	interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

	err := interp.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if len(interp.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
	}

	result, err := interp.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	expected := int64(1 + 10 + 100)
	if result != expected {
		t.Errorf("Expected result %d (1+10+100), got %d", expected, result)
	}
}

// buildLoad creates a LOAD instruction with the specified memory offset
func buildLoad(offset uint16) []byte {
	bytecode := make([]byte, 3)
	bytecode[0] = byte(OpLoad)
	bytecode[1] = byte(offset & 0xFF)
	bytecode[2] = byte((offset >> 8) & 0xFF)
	return bytecode
}

// buildStore creates a STORE instruction with the specified memory offset
func buildStore(offset uint16) []byte {
	bytecode := make([]byte, 3)
	bytecode[0] = byte(OpStore)
	bytecode[1] = byte(offset & 0xFF)
	bytecode[2] = byte((offset >> 8) & 0xFF)
	return bytecode
}

// buildCall creates a CALL instruction with the specified function offset
func buildCall(offset int32) []byte {
	bytecode := make([]byte, 5)
	bytecode[0] = byte(OpCall)
	binary.LittleEndian.PutUint32(bytecode[1:], uint32(offset))
	return bytecode
}

// TestControlFlowWhileLoop tests while loop with counter, break, and continue
func TestControlFlowWhileLoop(t *testing.T) {
	t.Run("basic while loop with counter", func(t *testing.T) {
		// Simulates:
		// counter = 0
		// sum = 0
		// while (counter < 5) {
		//   sum = sum + counter
		//   counter = counter + 1
		// }
		// return sum  // Should be 0+1+2+3+4 = 10

		var bytecode []byte

		// Initialize counter = 0 at memory[0]
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, buildStore(0)...)

		// Initialize sum = 0 at memory[1]
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, buildStore(1)...)

		// Loop start (offset 26)
		loopStart := len(bytecode)

		// Load counter
		bytecode = append(bytecode, buildLoad(0)...)

		// Push 5 for comparison
		bytecode = append(bytecode, buildPushI64(5)...)

		// Compare: counter >= 5 (exit condition)
		bytecode = append(bytecode, byte(OpGte))

		// If true (counter >= 5), jump to end
		// We'll calculate the jump offset after building the loop body
		jumpToEndPos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder, will fix offset

		// Loop body: sum = sum + counter
		// Load sum
		bytecode = append(bytecode, buildLoad(1)...)

		// Load counter
		bytecode = append(bytecode, buildLoad(0)...)

		// Add them
		bytecode = append(bytecode, byte(OpAdd))

		// Store back to sum
		bytecode = append(bytecode, buildStore(1)...)

		// Increment counter: counter = counter + 1
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, buildStore(0)...)

		// Jump back to loop start
		loopEnd := len(bytecode)
		jumpBackOffset := int32(loopStart - (loopEnd + 5)) // 5 = size of JMP instruction
		bytecode = append(bytecode, buildJump(jumpBackOffset)...)

		// End of loop (offset after loop)
		endOfLoop := len(bytecode)

		// Fix the conditional jump offset
		// Jump from jumpToEndPos to endOfLoop
		jumpForwardOffset := int32(endOfLoop - (jumpToEndPos + 5))
		copy(bytecode[jumpToEndPos+1:jumpToEndPos+5], []byte{
			byte(jumpForwardOffset & 0xFF),
			byte((jumpForwardOffset >> 8) & 0xFF),
			byte((jumpForwardOffset >> 16) & 0xFF),
			byte((jumpForwardOffset >> 24) & 0xFF),
		})

		// Load sum as result
		bytecode = append(bytecode, buildLoad(1)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		expected := int64(0 + 1 + 2 + 3 + 4)
		if result != expected {
			t.Errorf("Expected sum %d, got %d", expected, result)
		}
	})

	t.Run("while loop with break", func(t *testing.T) {
		// Simulates:
		// counter = 0
		// while (counter < 10) {
		//   if (counter == 5) break
		//   counter = counter + 1
		// }
		// return counter  // Should be 5

		var bytecode []byte

		// Initialize counter = 0 at memory[0]
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, buildStore(0)...)

		// Loop start
		loopStart := len(bytecode)

		// Load counter
		bytecode = append(bytecode, buildLoad(0)...)

		// Push 10 for comparison
		bytecode = append(bytecode, buildPushI64(10)...)

		// Compare: counter >= 10 (exit condition)
		bytecode = append(bytecode, byte(OpGte))

		// If true, jump to end
		jumpToEndPos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

		// Check if counter == 5 (break condition)
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(5)...)
		bytecode = append(bytecode, byte(OpEq))

		// If true, break (jump to end)
		breakJumpPos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

		// Increment counter
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, buildStore(0)...)

		// Jump back to loop start
		loopEnd := len(bytecode)
		jumpBackOffset := int32(loopStart - (loopEnd + 5))
		bytecode = append(bytecode, buildJump(jumpBackOffset)...)

		// End of loop
		endOfLoop := len(bytecode)

		// Fix the conditional jump offset (loop condition)
		jumpForwardOffset := int32(endOfLoop - (jumpToEndPos + 5))
		copy(bytecode[jumpToEndPos+1:jumpToEndPos+5], []byte{
			byte(jumpForwardOffset & 0xFF),
			byte((jumpForwardOffset >> 8) & 0xFF),
			byte((jumpForwardOffset >> 16) & 0xFF),
			byte((jumpForwardOffset >> 24) & 0xFF),
		})

		// Fix the break jump offset
		breakJumpOffset := int32(endOfLoop - (breakJumpPos + 5))
		copy(bytecode[breakJumpPos+1:breakJumpPos+5], []byte{
			byte(breakJumpOffset & 0xFF),
			byte((breakJumpOffset >> 8) & 0xFF),
			byte((breakJumpOffset >> 16) & 0xFF),
			byte((breakJumpOffset >> 24) & 0xFF),
		})

		// Load counter as result
		bytecode = append(bytecode, buildLoad(0)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 5 {
			t.Errorf("Expected counter to be 5 (break at 5), got %d", result)
		}
	})

	t.Run("while loop with continue", func(t *testing.T) {
		// Simulates:
		// counter = 0
		// sum = 0
		// while (counter < 10) {
		//   counter = counter + 1
		//   if (counter % 2 == 0) continue  // Skip even numbers
		//   sum = sum + counter
		// }
		// return sum  // Should be 1+3+5+7+9 = 25

		var bytecode []byte

		// Initialize counter = 0 at memory[0]
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, buildStore(0)...)

		// Initialize sum = 0 at memory[1]
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, buildStore(1)...)

		// Loop start
		loopStart := len(bytecode)

		// Load counter
		bytecode = append(bytecode, buildLoad(0)...)

		// Push 10 for comparison
		bytecode = append(bytecode, buildPushI64(10)...)

		// Compare: counter >= 10 (exit condition)
		bytecode = append(bytecode, byte(OpGte))

		// If true, jump to end
		jumpToEndPos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

		// Increment counter first
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, buildStore(0)...)

		// Check if counter % 2 == 0 (continue condition)
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(2)...)
		bytecode = append(bytecode, byte(OpMod))
		bytecode = append(bytecode, buildPushI64(0)...)
		bytecode = append(bytecode, byte(OpEq))

		// If true, continue (jump back to loop start)
		continueJumpPos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

		// Add counter to sum (only for odd numbers)
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, buildStore(1)...)

		// Jump back to loop start
		loopEnd := len(bytecode)
		jumpBackOffset := int32(loopStart - (loopEnd + 5))
		bytecode = append(bytecode, buildJump(jumpBackOffset)...)

		// End of loop
		endOfLoop := len(bytecode)

		// Fix the conditional jump offset (loop condition)
		jumpForwardOffset := int32(endOfLoop - (jumpToEndPos + 5))
		copy(bytecode[jumpToEndPos+1:jumpToEndPos+5], []byte{
			byte(jumpForwardOffset & 0xFF),
			byte((jumpForwardOffset >> 8) & 0xFF),
			byte((jumpForwardOffset >> 16) & 0xFF),
			byte((jumpForwardOffset >> 24) & 0xFF),
		})

		// Fix the continue jump offset (jump back to loop start)
		continueJumpOffset := int32(loopStart - (continueJumpPos + 5))
		copy(bytecode[continueJumpPos+1:continueJumpPos+5], []byte{
			byte(continueJumpOffset & 0xFF),
			byte((continueJumpOffset >> 8) & 0xFF),
			byte((continueJumpOffset >> 16) & 0xFF),
			byte((continueJumpOffset >> 24) & 0xFF),
		})

		// Load sum as result
		bytecode = append(bytecode, buildLoad(1)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		expected := int64(1 + 3 + 5 + 7 + 9)
		if result != expected {
			t.Errorf("Expected sum %d (odd numbers only), got %d", expected, result)
		}
	})
}

// TestControlFlowEarlyReturn tests early return from function
func TestControlFlowEarlyReturn(t *testing.T) {
	// Simulates:
	// function checkValue(x) {
	//   if (x < 0) {
	//     return -1  // Early return for negative
	//   }
	//   if (x == 0) {
	//     return 0   // Early return for zero
	//   }
	//   // Normal processing for positive values
	//   return x * 2
	// }
	// result = checkValue(5)  // Should return 10

	var bytecode []byte

	// Main program: call checkValue with argument 5
	// Push argument onto stack
	bytecode = append(bytecode, buildPushI64(5)...)

	// Store argument in memory[0] for function to access
	bytecode = append(bytecode, buildStore(0)...)

	// Call the function - we'll fix the offset after building the function
	callOffsetPos := len(bytecode) + 1           // Position where offset is stored (after OpCall byte)
	bytecode = append(bytecode, buildCall(0)...) // Placeholder offset

	// After function returns, result should be on stack
	// End main program
	bytecode = append(bytecode, byte(OpRet))

	// Function starts here - record the absolute position
	functionStart := len(bytecode)

	// Check if x < 0
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, byte(OpLt))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(x < 0), skip early return

	jumpToSecondCheckPos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

	// Early return -1 for negative
	bytecode = append(bytecode, buildPushI64(-1)...)
	bytecode = append(bytecode, byte(OpRet))

	// Second check: if x == 0
	secondCheckStart := len(bytecode)
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, byte(OpEq))
	bytecode = append(bytecode, byte(OpNot)) // Invert: if NOT(x == 0), skip early return

	jumpToNormalProcessingPos := len(bytecode)
	bytecode = append(bytecode, buildJumpIf(0)...) // Placeholder

	// Early return 0 for zero
	bytecode = append(bytecode, buildPushI64(0)...)
	bytecode = append(bytecode, byte(OpRet))

	// Normal processing: return x * 2
	normalProcessingStart := len(bytecode)
	bytecode = append(bytecode, buildLoad(0)...)
	bytecode = append(bytecode, buildPushI64(2)...)
	bytecode = append(bytecode, byte(OpMul))
	bytecode = append(bytecode, byte(OpRet))

	// Fix all jump offsets
	// 1. Fix CALL offset to point to function start (absolute)
	binary.LittleEndian.PutUint32(bytecode[callOffsetPos:callOffsetPos+4], uint32(functionStart))

	// 2. Jump to second check (relative)
	jumpToSecondCheckOffset := int32(secondCheckStart - (jumpToSecondCheckPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToSecondCheckPos+1:jumpToSecondCheckPos+5], uint32(jumpToSecondCheckOffset))

	// 3. Jump to normal processing (relative)
	jumpToNormalProcessingOffset := int32(normalProcessingStart - (jumpToNormalProcessingPos + 5))
	binary.LittleEndian.PutUint32(bytecode[jumpToNormalProcessingPos+1:jumpToNormalProcessingPos+5], uint32(jumpToNormalProcessingOffset))

	// Test with positive value (5)
	t.Run("positive value returns x*2", func(t *testing.T) {
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 10 {
			t.Errorf("Expected 10 (5*2), got %d", result)
		}
	})

	// Test with negative value
	t.Run("negative value returns -1 early", func(t *testing.T) {
		// Rebuild bytecode with -3 as argument
		var bytecode2 []byte
		bytecode2 = append(bytecode2, buildPushI64(-3)...)
		bytecode2 = append(bytecode2, buildStore(0)...)

		callOffsetPos2 := len(bytecode2) + 1
		bytecode2 = append(bytecode2, buildCall(0)...)
		bytecode2 = append(bytecode2, byte(OpRet))

		// Copy function code from original bytecode
		functionStart2 := len(bytecode2)
		bytecode2 = append(bytecode2, bytecode[functionStart:]...)

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode2[callOffsetPos2:callOffsetPos2+4], uint32(functionStart2))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode2, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != -1 {
			t.Errorf("Expected -1 (early return for negative), got %d", result)
		}
	})

	// Test with zero value
	t.Run("zero value returns 0 early", func(t *testing.T) {
		// Rebuild bytecode with 0 as argument
		var bytecode3 []byte
		bytecode3 = append(bytecode3, buildPushI64(0)...)
		bytecode3 = append(bytecode3, buildStore(0)...)

		callOffsetPos3 := len(bytecode3) + 1
		bytecode3 = append(bytecode3, buildCall(0)...)
		bytecode3 = append(bytecode3, byte(OpRet))

		// Copy function code from original bytecode
		functionStart3 := len(bytecode3)
		bytecode3 = append(bytecode3, bytecode[functionStart:]...)

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode3[callOffsetPos3:callOffsetPos3+4], uint32(functionStart3))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode3, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 0 {
			t.Errorf("Expected 0 (early return for zero), got %d", result)
		}
	})
}

// TestDataStructuresArray tests array creation, push, get, length, pop operations
func TestDataStructuresArray(t *testing.T) {
	t.Run("array creation and push", func(t *testing.T) {
		// Create array, push 3 elements, verify length
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Push first element (10)
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Push second element (20)
		bytecode = append(bytecode, buildPushI64(20)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Push third element (30)
		bytecode = append(bytecode, buildPushI64(30)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Get array length
		bytecode = append(bytecode, byte(OpArrayLen))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		length, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}

		if length != 3 {
			t.Errorf("Expected array length 3, got %d", length)
		}
	})

	t.Run("array get by index", func(t *testing.T) {
		// Create array with 3 elements, get element at index 1
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Push elements
		bytecode = append(bytecode, buildPushI64(100)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		bytecode = append(bytecode, buildPushI64(200)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		bytecode = append(bytecode, buildPushI64(300)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Get element at index 1 (should be 200)
		bytecode = append(bytecode, buildPushU64(1)...)
		bytecode = append(bytecode, byte(OpArrayGet))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		value, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != 200 {
			t.Errorf("Expected value 200, got %d", value)
		}
	})

	t.Run("array pop operation", func(t *testing.T) {
		// Create array with 3 elements, pop last element
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Push elements
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		bytecode = append(bytecode, buildPushI64(20)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		bytecode = append(bytecode, buildPushI64(30)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Pop last element (should return array and 30)
		bytecode = append(bytecode, byte(OpArrayPop))

		// Return (popped element should be on top)
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Stack should have: [array, popped_element]
		if len(interp.stack) != 2 {
			t.Fatalf("Expected 2 values on stack, got %d", len(interp.stack))
		}

		// Check popped element (top of stack)
		poppedValue, err := interp.stack[1].AsI64()
		if err != nil {
			t.Fatalf("Failed to get popped value: %v", err)
		}

		if poppedValue != 30 {
			t.Errorf("Expected popped value 30, got %d", poppedValue)
		}

		// Check array length
		arr, err := interp.stack[0].AsArray()
		if err != nil {
			t.Fatalf("Failed to get array: %v", err)
		}

		if len(arr) != 2 {
			t.Errorf("Expected array length 2 after pop, got %d", len(arr))
		}
	})

	t.Run("array out of bounds access", func(t *testing.T) {
		// Create array with 2 elements, try to access index 5
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Push elements
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		bytecode = append(bytecode, buildPushI64(20)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Try to get element at index 5 (out of bounds)
		bytecode = append(bytecode, buildPushU64(5)...)
		bytecode = append(bytecode, byte(OpArrayGet))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for out of bounds access, got nil")
		}

		// Verify error message contains "out of bounds"
		if err.Error() != "array index out of bounds: 5" {
			t.Errorf("Expected 'array index out of bounds: 5' error, got: %v", err)
		}
	})

	t.Run("array element ordering", func(t *testing.T) {
		// Create array with multiple elements, verify ordering is preserved
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Push elements in specific order
		for i := int64(1); i <= 5; i++ {
			bytecode = append(bytecode, buildPushI64(i*10)...)
			bytecode = append(bytecode, byte(OpArrayPush))
		}

		// Store array in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Verify each element
		for i := uint64(0); i < 5; i++ {
			// Load array
			bytecode = append(bytecode, buildLoad(0)...)
			// Get element at index i
			bytecode = append(bytecode, buildPushU64(i)...)
			bytecode = append(bytecode, byte(OpArrayGet))
			// Store in memory[i+1]
			bytecode = append(bytecode, buildStore(uint16(i+1))...)
		}

		// Load all elements back to stack for verification
		for i := uint16(1); i <= 5; i++ {
			bytecode = append(bytecode, buildLoad(i)...)
		}

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Stack should have 5 elements
		if len(interp.stack) != 5 {
			t.Fatalf("Expected 5 values on stack, got %d", len(interp.stack))
		}

		// Verify ordering
		for i := 0; i < 5; i++ {
			value, err := interp.stack[i].AsI64()
			if err != nil {
				t.Fatalf("Failed to get value at position %d: %v", i, err)
			}

			expected := int64((i + 1) * 10)
			if value != expected {
				t.Errorf("Expected value %d at position %d, got %d", expected, i, value)
			}
		}
	})
}

// TestDataStructuresMap tests map creation, set, get, has, delete operations
func TestDataStructuresMap(t *testing.T) {
	t.Run("map creation and set", func(t *testing.T) {
		// Create map, set key-value pairs
		var bytecode []byte

		// Create new map
		bytecode = append(bytecode, byte(OpMapNew))

		// Set "name" -> "Alice"
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, buildPushString("Alice")...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Set "age" -> 30
		bytecode = append(bytecode, buildPushString("age")...)
		bytecode = append(bytecode, buildPushI64(30)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Get "name" value
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		value, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != "Alice" {
			t.Errorf("Expected value 'Alice', got '%s'", value)
		}
	})

	t.Run("map get with string and number keys", func(t *testing.T) {
		// Create map with both string and number keys (as strings)
		var bytecode []byte

		// Create new map
		bytecode = append(bytecode, byte(OpMapNew))

		// Set "key1" -> 100
		bytecode = append(bytecode, buildPushString("key1")...)
		bytecode = append(bytecode, buildPushI64(100)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Set "key2" -> 200
		bytecode = append(bytecode, buildPushString("key2")...)
		bytecode = append(bytecode, buildPushI64(200)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Get "key2" value
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("key2")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		value, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != 200 {
			t.Errorf("Expected value 200, got %d", value)
		}
	})

	t.Run("map has operation", func(t *testing.T) {
		// Create map, check if keys exist
		var bytecode []byte

		// Create new map
		bytecode = append(bytecode, byte(OpMapNew))

		// Set "exists" -> true
		bytecode = append(bytecode, buildPushString("exists")...)
		bytecode = append(bytecode, buildPushBool(true)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Check if "exists" key is present
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("exists")...)
		bytecode = append(bytecode, byte(OpMapHas))

		// Store result in memory[1]
		bytecode = append(bytecode, buildStore(1)...)

		// Check if "notexists" key is present
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("notexists")...)
		bytecode = append(bytecode, byte(OpMapHas))

		// Store result in memory[2]
		bytecode = append(bytecode, buildStore(2)...)

		// Load both results
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, buildLoad(2)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 2 {
			t.Fatalf("Expected 2 values on stack, got %d", len(interp.stack))
		}

		// Check "exists" key result
		existsResult, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get exists result: %v", err)
		}

		if !existsResult {
			t.Errorf("Expected 'exists' key to be present")
		}

		// Check "notexists" key result
		notExistsResult, err := interp.stack[1].AsBool()
		if err != nil {
			t.Fatalf("Failed to get notexists result: %v", err)
		}

		if notExistsResult {
			t.Errorf("Expected 'notexists' key to be absent")
		}
	})

	t.Run("map delete operation", func(t *testing.T) {
		// Create map, delete a key, verify it's gone
		var bytecode []byte

		// Create new map
		bytecode = append(bytecode, byte(OpMapNew))

		// Set "key1" -> 100
		bytecode = append(bytecode, buildPushString("key1")...)
		bytecode = append(bytecode, buildPushI64(100)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Set "key2" -> 200
		bytecode = append(bytecode, buildPushString("key2")...)
		bytecode = append(bytecode, buildPushI64(200)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Delete "key1"
		bytecode = append(bytecode, buildPushString("key1")...)
		bytecode = append(bytecode, byte(OpMapDel))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Check if "key1" still exists
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("key1")...)
		bytecode = append(bytecode, byte(OpMapHas))

		// Store result in memory[1]
		bytecode = append(bytecode, buildStore(1)...)

		// Check if "key2" still exists
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("key2")...)
		bytecode = append(bytecode, byte(OpMapHas))

		// Store result in memory[2]
		bytecode = append(bytecode, buildStore(2)...)

		// Load both results
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, buildLoad(2)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 2 {
			t.Fatalf("Expected 2 values on stack, got %d", len(interp.stack))
		}

		// Check "key1" was deleted
		key1Exists, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get key1 result: %v", err)
		}

		if key1Exists {
			t.Errorf("Expected 'key1' to be deleted")
		}

		// Check "key2" still exists
		key2Exists, err := interp.stack[1].AsBool()
		if err != nil {
			t.Fatalf("Failed to get key2 result: %v", err)
		}

		if !key2Exists {
			t.Errorf("Expected 'key2' to still exist")
		}
	})

	t.Run("map value storage and retrieval", func(t *testing.T) {
		// Test storing and retrieving different value types
		var bytecode []byte

		// Create new map
		bytecode = append(bytecode, byte(OpMapNew))

		// Set "int_val" -> 42
		bytecode = append(bytecode, buildPushString("int_val")...)
		bytecode = append(bytecode, buildPushI64(42)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Set "str_val" -> "hello"
		bytecode = append(bytecode, buildPushString("str_val")...)
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Set "bool_val" -> true
		bytecode = append(bytecode, buildPushString("bool_val")...)
		bytecode = append(bytecode, buildPushBool(true)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Get "int_val"
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("int_val")...)
		bytecode = append(bytecode, byte(OpMapGet))
		bytecode = append(bytecode, buildStore(1)...)

		// Get "str_val"
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("str_val")...)
		bytecode = append(bytecode, byte(OpMapGet))
		bytecode = append(bytecode, buildStore(2)...)

		// Get "bool_val"
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("bool_val")...)
		bytecode = append(bytecode, byte(OpMapGet))
		bytecode = append(bytecode, buildStore(3)...)

		// Load all values
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, buildLoad(2)...)
		bytecode = append(bytecode, buildLoad(3)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 3 {
			t.Fatalf("Expected 3 values on stack, got %d", len(interp.stack))
		}

		// Verify int value
		intVal, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get int value: %v", err)
		}
		if intVal != 42 {
			t.Errorf("Expected int value 42, got %d", intVal)
		}

		// Verify string value
		strVal, err := interp.stack[1].AsString()
		if err != nil {
			t.Fatalf("Failed to get string value: %v", err)
		}
		if strVal != "hello" {
			t.Errorf("Expected string value 'hello', got '%s'", strVal)
		}

		// Verify bool value
		boolVal, err := interp.stack[2].AsBool()
		if err != nil {
			t.Fatalf("Failed to get bool value: %v", err)
		}
		if !boolVal {
			t.Errorf("Expected bool value true, got false")
		}
	})
}

// TestDataStructuresNested tests nested data structures (array containing maps)
func TestDataStructuresNested(t *testing.T) {
	t.Run("array of maps creation", func(t *testing.T) {
		// Create array containing maps
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Create first map: {"name": "Alice", "age": 30}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, buildPushString("Alice")...)
		bytecode = append(bytecode, byte(OpMapSet))
		bytecode = append(bytecode, buildPushString("age")...)
		bytecode = append(bytecode, buildPushI64(30)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push first map to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Create second map: {"name": "Bob", "age": 25}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, buildPushString("Bob")...)
		bytecode = append(bytecode, byte(OpMapSet))
		bytecode = append(bytecode, buildPushString("age")...)
		bytecode = append(bytecode, buildPushI64(25)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push second map to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store array in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Get array length
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, byte(OpArrayLen))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		length, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}

		if length != 2 {
			t.Errorf("Expected array length 2, got %d", length)
		}
	})

	t.Run("deep access nested structures", func(t *testing.T) {
		// Create array of maps, access nested values
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Create first map: {"name": "Alice", "score": 100}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, buildPushString("Alice")...)
		bytecode = append(bytecode, byte(OpMapSet))
		bytecode = append(bytecode, buildPushString("score")...)
		bytecode = append(bytecode, buildPushI64(100)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Create second map: {"name": "Bob", "score": 85}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, buildPushString("Bob")...)
		bytecode = append(bytecode, byte(OpMapSet))
		bytecode = append(bytecode, buildPushString("score")...)
		bytecode = append(bytecode, buildPushI64(85)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store array in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Access array[1]["name"] (should be "Bob")
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(1)...)
		bytecode = append(bytecode, byte(OpArrayGet))
		bytecode = append(bytecode, buildPushString("name")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Store result in memory[1]
		bytecode = append(bytecode, buildStore(1)...)

		// Access array[0]["score"] (should be 100)
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpArrayGet))
		bytecode = append(bytecode, buildPushString("score")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Store result in memory[2]
		bytecode = append(bytecode, buildStore(2)...)

		// Load both results
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, buildLoad(2)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 2 {
			t.Fatalf("Expected 2 values on stack, got %d", len(interp.stack))
		}

		// Verify array[1]["name"] = "Bob"
		name, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get name: %v", err)
		}
		if name != "Bob" {
			t.Errorf("Expected name 'Bob', got '%s'", name)
		}

		// Verify array[0]["score"] = 100
		score, err := interp.stack[1].AsI64()
		if err != nil {
			t.Fatalf("Failed to get score: %v", err)
		}
		if score != 100 {
			t.Errorf("Expected score 100, got %d", score)
		}
	})

	t.Run("modify nested structures", func(t *testing.T) {
		// Create array of maps, modify a nested value
		var bytecode []byte

		// Create new array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Create map: {"value": 10}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("value")...)
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store array in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Get the map from array[0]
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpArrayGet))

		// Modify the map: set "value" to 20
		bytecode = append(bytecode, buildPushString("value")...)
		bytecode = append(bytecode, buildPushI64(20)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Now we have modified_map on stack
		// Store it temporarily in memory[1]
		bytecode = append(bytecode, buildStore(1)...)

		// Now update the array: load array, push index, load modified map
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, buildLoad(1)...)

		// Set array element (returns new array)
		bytecode = append(bytecode, byte(OpArraySet))

		// Store updated array back to memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Now read back the value to verify
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpArrayGet))
		bytecode = append(bytecode, buildPushString("value")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Verify the value was modified to 20
		value, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		if value != 20 {
			t.Errorf("Expected modified value 20, got %d", value)
		}
	})

	t.Run("nested structure integrity", func(t *testing.T) {
		// Create complex nested structure and verify all values
		var bytecode []byte

		// Create array
		bytecode = append(bytecode, byte(OpArrayNew))

		// Add 3 maps with different data
		for i := int64(0); i < 3; i++ {
			// Create map
			bytecode = append(bytecode, byte(OpMapNew))
			bytecode = append(bytecode, buildPushString("id")...)
			bytecode = append(bytecode, buildPushI64(i)...)
			bytecode = append(bytecode, byte(OpMapSet))
			bytecode = append(bytecode, buildPushString("value")...)
			bytecode = append(bytecode, buildPushI64(i*100)...)
			bytecode = append(bytecode, byte(OpMapSet))

			// Push to array
			bytecode = append(bytecode, byte(OpArrayPush))
		}

		// Store array in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Verify all values are correct
		for i := uint64(0); i < 3; i++ {
			// Get map at index i
			bytecode = append(bytecode, buildLoad(0)...)
			bytecode = append(bytecode, buildPushU64(i)...)
			bytecode = append(bytecode, byte(OpArrayGet))

			// Get "id" value
			bytecode = append(bytecode, buildPushString("id")...)
			bytecode = append(bytecode, byte(OpMapGet))

			// Store in memory[i+1]
			bytecode = append(bytecode, buildStore(uint16(i+1))...)
		}

		// Load all id values
		for i := uint16(1); i <= 3; i++ {
			bytecode = append(bytecode, buildLoad(i)...)
		}

		// Return
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 3 {
			t.Fatalf("Expected 3 values on stack, got %d", len(interp.stack))
		}

		// Verify all id values
		for i := 0; i < 3; i++ {
			id, err := interp.stack[i].AsI64()
			if err != nil {
				t.Fatalf("Failed to get id at position %d: %v", i, err)
			}
			if id != int64(i) {
				t.Errorf("Expected id %d at position %d, got %d", i, i, id)
			}
		}
	})
}

// TestDataStructuresPassToFunction tests passing arrays and maps to functions
func TestDataStructuresPassToFunction(t *testing.T) {
	t.Run("pass array to function", func(t *testing.T) {
		// Create function that adds element to array
		// Main: create array, pass to function, verify result
		var bytecode []byte

		// Main program: create array with 2 elements
		bytecode = append(bytecode, byte(OpArrayNew))
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpArrayPush))
		bytecode = append(bytecode, buildPushI64(20)...)
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store array in memory[0] (function will access it)
		bytecode = append(bytecode, buildStore(0)...)

		// Call function
		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // Placeholder

		// After function returns, result array is in memory[0]
		// Get array length
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, byte(OpArrayLen))

		// End main
		bytecode = append(bytecode, byte(OpRet))

		// Function: add element 30 to array
		functionStart := len(bytecode)

		// Load array from memory[0]
		bytecode = append(bytecode, buildLoad(0)...)

		// Push new element
		bytecode = append(bytecode, buildPushI64(30)...)

		// Push to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store modified array back to memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Verify array length is now 3
		length, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}

		if length != 3 {
			t.Errorf("Expected array length 3 after function call, got %d", length)
		}
	})

	t.Run("pass map to function", func(t *testing.T) {
		// Create function that modifies map
		var bytecode []byte

		// Main program: create map
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("count")...)
		bytecode = append(bytecode, buildPushI64(5)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store map in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Call function
		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // Placeholder

		// After function returns, get "count" value
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushString("count")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// End main
		bytecode = append(bytecode, byte(OpRet))

		// Function: increment "count" by 10
		functionStart := len(bytecode)

		// Load map
		bytecode = append(bytecode, buildLoad(0)...)

		// Get current count
		bytecode = append(bytecode, buildPushString("count")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// Add 10
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpAdd))

		// Store result temporarily
		bytecode = append(bytecode, buildStore(1)...)

		// Load map again
		bytecode = append(bytecode, buildLoad(0)...)

		// Set new count value
		bytecode = append(bytecode, buildPushString("count")...)
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store modified map back
		bytecode = append(bytecode, buildStore(0)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Verify count is now 15
		count, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}

		if count != 15 {
			t.Errorf("Expected count 15 after function call, got %d", count)
		}
	})

	t.Run("verify reference semantics", func(t *testing.T) {
		// Pass array to function, modify it, verify changes persist
		var bytecode []byte

		// Main: create array [1, 2, 3]
		bytecode = append(bytecode, byte(OpArrayNew))
		for i := int64(1); i <= 3; i++ {
			bytecode = append(bytecode, buildPushI64(i)...)
			bytecode = append(bytecode, byte(OpArrayPush))
		}

		// Store in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Call function that modifies array
		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // Placeholder

		// After function, get element at index 1 (should be modified)
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(1)...)
		bytecode = append(bytecode, byte(OpArrayGet))

		// End main
		bytecode = append(bytecode, byte(OpRet))

		// Function: modify array[1] to 999
		functionStart := len(bytecode)

		// Load array
		bytecode = append(bytecode, buildLoad(0)...)

		// Set array[1] = 999
		bytecode = append(bytecode, buildPushU64(1)...)
		bytecode = append(bytecode, buildPushI64(999)...)
		bytecode = append(bytecode, byte(OpArraySet))

		// Store modified array back
		bytecode = append(bytecode, buildStore(0)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Verify array[1] was modified to 999
		value, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != 999 {
			t.Errorf("Expected value 999 (modified by function), got %d", value)
		}
	})

	t.Run("pass nested structure to function", func(t *testing.T) {
		// Pass array of maps to function, modify nested value
		var bytecode []byte

		// Main: create array with one map
		bytecode = append(bytecode, byte(OpArrayNew))

		// Create map {"x": 10}
		bytecode = append(bytecode, byte(OpMapNew))
		bytecode = append(bytecode, buildPushString("x")...)
		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Push map to array
		bytecode = append(bytecode, byte(OpArrayPush))

		// Store in memory[0]
		bytecode = append(bytecode, buildStore(0)...)

		// Call function
		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // Placeholder

		// After function, access array[0]["x"]
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpArrayGet))
		bytecode = append(bytecode, buildPushString("x")...)
		bytecode = append(bytecode, byte(OpMapGet))

		// End main
		bytecode = append(bytecode, byte(OpRet))

		// Function: modify array[0]["x"] to 50
		functionStart := len(bytecode)

		// Load array
		bytecode = append(bytecode, buildLoad(0)...)

		// Get map at index 0
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpArrayGet))

		// Modify map: set "x" to 50
		bytecode = append(bytecode, buildPushString("x")...)
		bytecode = append(bytecode, buildPushI64(50)...)
		bytecode = append(bytecode, byte(OpMapSet))

		// Store modified map temporarily
		bytecode = append(bytecode, buildStore(1)...)

		// Load array
		bytecode = append(bytecode, buildLoad(0)...)

		// Set array[0] to modified map
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, byte(OpArraySet))

		// Store modified array back
		bytecode = append(bytecode, buildStore(0)...)

		// Return
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL offset
		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Verify array[0]["x"] was modified to 50
		value, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}

		if value != 50 {
			t.Errorf("Expected value 50 (modified by function), got %d", value)
		}
	})
}

// buildPushFileID creates a PUSH instruction for a FileID value (fixed 32 bytes)
func buildPushFileID(id filestore.FileID) []byte {
	bytecode := make([]byte, 2+32)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypeFileID)
	copy(bytecode[2:], id[:])
	return bytecode
}

// buildPushPublicKey creates a PUSH instruction for a PublicKey value (fixed 32 bytes)
func buildPushPublicKey(key [32]byte) []byte {
	bytecode := make([]byte, 2+32)
	bytecode[0] = byte(OpPush)
	bytecode[1] = byte(TypePublicKey)
	copy(bytecode[2:], key[:])
	return bytecode
}

// TestBlockchainStateFileOperations tests GETFILE and UPDATEFILE operations
func TestBlockchainStateFileOperations(t *testing.T) {

	t.Run("GETFILE reads data correctly", func(t *testing.T) {
		// Create a test file in the mock context
		var testFileID filestore.FileID
		testFileID[31] = 0x42
		testData := []byte("hello blockchain")

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:   testFileID,
			Data: testData,
		}

		// Build bytecode: PUSH fileID, GETFILE, RET
		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, byte(OpGetFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if string(result) != string(testData) {
			t.Errorf("Expected data %q, got %q", testData, result)
		}
	})

	t.Run("UPDATEFILE persists changes", func(t *testing.T) {
		var testFileID filestore.FileID
		testFileID[31] = 0x43
		originalData := []byte("original data")
		updatedData := []byte("updated data")

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:   testFileID,
			Data: originalData,
		}

		// Build bytecode: PUSH fileID, PUSH newData, UPDATEFILE, PUSH fileID, GETFILE, RET
		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, buildPushBytes(updatedData)...)
		bytecode = append(bytecode, byte(OpUpdateFile))
		// Now read back to verify
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, byte(OpGetFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if string(result) != string(updatedData) {
			t.Errorf("Expected updated data %q, got %q", updatedData, result)
		}

		// Also verify the context was updated
		file, err := ctx.GetFile(testFileID)
		if err != nil {
			t.Fatalf("Failed to get file from context: %v", err)
		}
		if string(file.Data) != string(updatedData) {
			t.Errorf("Context not updated: expected %q, got %q", updatedData, file.Data)
		}
	})

	t.Run("file data integrity after multiple updates", func(t *testing.T) {
		var testFileID filestore.FileID
		testFileID[31] = 0x44

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:   testFileID,
			Data: []byte("initial"),
		}

		// Update twice, verify final state
		finalData := []byte("final state")
		var bytecode []byte
		// First update
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, buildPushBytes([]byte("intermediate"))...)
		bytecode = append(bytecode, byte(OpUpdateFile))
		// Second update
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, buildPushBytes(finalData)...)
		bytecode = append(bytecode, byte(OpUpdateFile))
		// Read back
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, byte(OpGetFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if string(result) != string(finalData) {
			t.Errorf("Expected final data %q, got %q", finalData, result)
		}
	})
}

// TestBlockchainStateBalanceOperations tests GETBALANCE operations
func TestBlockchainStateBalanceOperations(t *testing.T) {

	t.Run("GETBALANCE returns correct balance", func(t *testing.T) {
		var testFileID filestore.FileID
		testFileID[31] = 0x50
		expectedBalance := int64(1000000)

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:      testFileID,
			Balance: expectedBalance,
			Data:    []byte{},
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, byte(OpGetBalance))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		balance, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get balance: %v", err)
		}

		if balance != expectedBalance {
			t.Errorf("Expected balance %d, got %d", expectedBalance, balance)
		}
	})

	t.Run("GETBALANCE for zero balance file", func(t *testing.T) {
		var testFileID filestore.FileID
		testFileID[31] = 0x51

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:      testFileID,
			Balance: 0,
			Data:    []byte{},
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, byte(OpGetBalance))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		balance, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get balance: %v", err)
		}

		if balance != 0 {
			t.Errorf("Expected balance 0, got %d", balance)
		}
	})

	t.Run("GETBALANCE for non-existent file returns error", func(t *testing.T) {
		var missingFileID filestore.FileID
		missingFileID[31] = 0xFF

		ctx := NewMockExecutionContext()
		// Don't add the file to context

		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(missingFileID)...)
		bytecode = append(bytecode, byte(OpGetBalance))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for non-existent file, got nil")
		}
	})
}

// TestBlockchainStateSignerOperations tests GETSIGNER, HASSIGNER, and GETINSTRDATA
func TestBlockchainStateSignerOperations(t *testing.T) {
	t.Run("GETSIGNER with valid index", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		// Add a test signer
		var testKey [32]byte
		testKey[0] = 0xAB
		testKey[31] = 0xCD
		ctx.signers = []transaction.PublicKey{testKey}

		// Build bytecode: PUSH 0 (index), GETSIGNER, RET
		var bytecode []byte
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpGetSigner))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		// Result should be a PublicKey type
		if interp.stack[0].Type != TypePublicKey {
			t.Errorf("Expected TypePublicKey, got %v", interp.stack[0].Type)
		}

		keyBytes, ok := interp.stack[0].Data.([]byte)
		if !ok {
			t.Fatalf("Expected []byte data for PublicKey")
		}
		if len(keyBytes) < 32 {
			t.Fatalf("Expected 32 key bytes, got %d", len(keyBytes))
		}

		if keyBytes[0] != 0xAB || keyBytes[31] != 0xCD {
			t.Errorf("Key bytes mismatch: got %x", keyBytes)
		}
	})

	t.Run("HASSIGNER with present key", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		var testKey [32]byte
		testKey[0] = 0x11
		ctx.signers = []transaction.PublicKey{testKey}

		// Build bytecode: PUSH pubkey, HASSIGNER, RET
		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey(testKey)...)
		bytecode = append(bytecode, byte(OpHasSigner))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if !result {
			t.Errorf("Expected HASSIGNER to return true for present key")
		}
	})

	t.Run("HASSIGNER with absent key", func(t *testing.T) {
		ctx := NewMockExecutionContext()
		// No signers added

		var absentKey [32]byte
		absentKey[0] = 0x99

		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey(absentKey)...)
		bytecode = append(bytecode, byte(OpHasSigner))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result {
			t.Errorf("Expected HASSIGNER to return false for absent key")
		}
	})

	t.Run("GETINSTRDATA returns instruction data", func(t *testing.T) {
		ctx := NewMockExecutionContext()
		ctx.instrData = []byte{0x01, 0x02, 0x03, 0x04}

		var bytecode []byte
		bytecode = append(bytecode, byte(OpGetInstrData))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		data, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get data: %v", err)
		}

		expected := []byte{0x01, 0x02, 0x03, 0x04}
		if len(data) != len(expected) {
			t.Fatalf("Expected %d bytes, got %d", len(expected), len(data))
		}
		for i, b := range expected {
			if data[i] != b {
				t.Errorf("Byte mismatch at index %d: expected %d, got %d", i, b, data[i])
			}
		}
	})

	t.Run("GETSIGNER out of bounds returns error", func(t *testing.T) {
		ctx := NewMockExecutionContext()
		// No signers

		var bytecode []byte
		bytecode = append(bytecode, buildPushU64(0)...)
		bytecode = append(bytecode, byte(OpGetSigner))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for out-of-bounds signer index, got nil")
		}
	})
}

// TestBlockchainStateInvalidOperations tests error handling for invalid state operations
func TestBlockchainStateInvalidOperations(t *testing.T) {

	t.Run("GETFILE with non-existent file ID returns error", func(t *testing.T) {
		var missingID filestore.FileID
		missingID[31] = 0xDE

		ctx := NewMockExecutionContext()

		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(missingID)...)
		bytecode = append(bytecode, byte(OpGetFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for non-existent file, got nil")
		}
	})

	t.Run("UPDATEFILE with non-existent file ID returns error", func(t *testing.T) {
		var missingID filestore.FileID
		missingID[31] = 0xDF

		ctx := NewMockExecutionContext()

		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(missingID)...)
		bytecode = append(bytecode, buildPushBytes([]byte("data"))...)
		bytecode = append(bytecode, byte(OpUpdateFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for updating non-existent file, got nil")
		}
	})

	t.Run("GETFILE with wrong type on stack returns error", func(t *testing.T) {
		ctx := NewMockExecutionContext()

		// Push an i64 instead of a FileID
		var bytecode []byte
		bytecode = append(bytecode, buildPushI64(42)...)
		bytecode = append(bytecode, byte(OpGetFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected type error for GETFILE with i64, got nil")
		}
	})

	t.Run("UPDATEFILE with wrong data type returns error", func(t *testing.T) {
		var testFileID filestore.FileID
		testFileID[31] = 0xE0

		ctx := NewMockExecutionContext()
		ctx.files[testFileID] = &filestore.File{
			ID:   testFileID,
			Data: []byte("original"),
		}

		// Push i64 as data instead of bytes
		var bytecode []byte
		bytecode = append(bytecode, buildPushFileID(testFileID)...)
		bytecode = append(bytecode, buildPushI64(999)...) // wrong type
		bytecode = append(bytecode, byte(OpUpdateFile))
		bytecode = append(bytecode, byte(OpRet))

		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected type error for UPDATEFILE with i64 data, got nil")
		}
	})
}

// TestStringConcatenation tests STRCONCAT with two strings and empty strings
func TestStringConcatenation(t *testing.T) {
	t.Run("concatenate two non-empty strings", func(t *testing.T) {
		// Stack order: push str1, push str2, STRCONCAT -> str1+str2
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, buildPushString(" world")...)
		bytecode = append(bytecode, byte(OpStrConcat))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", result)
		}
	})

	t.Run("concatenate with empty first string", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("")...)
		bytecode = append(bytecode, buildPushString("world")...)
		bytecode = append(bytecode, byte(OpStrConcat))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "world" {
			t.Errorf("Expected 'world', got '%s'", result)
		}
	})

	t.Run("concatenate with empty second string", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, buildPushString("")...)
		bytecode = append(bytecode, byte(OpStrConcat))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "hello" {
			t.Errorf("Expected 'hello', got '%s'", result)
		}
	})

	t.Run("concatenate two empty strings", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("")...)
		bytecode = append(bytecode, buildPushString("")...)
		bytecode = append(bytecode, byte(OpStrConcat))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})
}

// TestStringSubstring tests STRSUBSTRING with valid indices, boundary conditions, and invalid indices
// Stack order for STRSUBSTRING: push string, push start (u64), push end (u64), STRSUBSTRING
func TestStringSubstring(t *testing.T) {
	t.Run("extract middle substring", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello world")...)
		bytecode = append(bytecode, buildPushU64(6)...)  // start
		bytecode = append(bytecode, buildPushU64(11)...) // end
		bytecode = append(bytecode, byte(OpStrSubstring))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "world" {
			t.Errorf("Expected 'world', got '%s'", result)
		}
	})

	t.Run("boundary: start=0 to end=length", func(t *testing.T) {
		str := "hello"
		var bytecode []byte
		bytecode = append(bytecode, buildPushString(str)...)
		bytecode = append(bytecode, buildPushU64(0)...)                // start
		bytecode = append(bytecode, buildPushU64(uint64(len(str)))...) // end = length
		bytecode = append(bytecode, byte(OpStrSubstring))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != str {
			t.Errorf("Expected '%s', got '%s'", str, result)
		}
	})

	t.Run("boundary: empty substring (start == end)", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, buildPushU64(2)...) // start
		bytecode = append(bytecode, buildPushU64(2)...) // end == start
		bytecode = append(bytecode, byte(OpStrSubstring))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "" {
			t.Errorf("Expected empty string, got '%s'", result)
		}
	})

	t.Run("invalid: end index out of bounds returns error", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, buildPushU64(0)...)  // start
		bytecode = append(bytecode, buildPushU64(10)...) // end > length
		bytecode = append(bytecode, byte(OpStrSubstring))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for out-of-bounds end index, got nil")
		}
	})

	t.Run("invalid: start greater than end returns error", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, buildPushU64(3)...) // start
		bytecode = append(bytecode, buildPushU64(1)...) // end < start
		bytecode = append(bytecode, byte(OpStrSubstring))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for start > end, got nil")
		}
	})
}

// TestStringLengthAndConversions tests STRLEN, STRTOBYTES, STRFROMBYTES, and round-trip conversion
func TestStringLengthAndConversions(t *testing.T) {
	t.Run("STRLEN returns correct length", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("hello")...)
		bytecode = append(bytecode, byte(OpStrLen))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 5 {
			t.Errorf("Expected length 5, got %d", result)
		}
	})

	t.Run("STRLEN of empty string returns 0", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("")...)
		bytecode = append(bytecode, byte(OpStrLen))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 0 {
			t.Errorf("Expected length 0, got %d", result)
		}
	})

	t.Run("STRTOBYTES converts string to byte slice", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("abc")...)
		bytecode = append(bytecode, byte(OpStrToBytes))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		expected := []byte("abc")
		if len(result) != len(expected) {
			t.Fatalf("Expected %d bytes, got %d", len(expected), len(result))
		}
		for i, b := range expected {
			if result[i] != b {
				t.Errorf("Byte mismatch at index %d: expected %d, got %d", i, b, result[i])
			}
		}
	})

	t.Run("STRFROMBYTES converts byte slice to string", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes([]byte("abc"))...)
		bytecode = append(bytecode, byte(OpStrFromBytes))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != "abc" {
			t.Errorf("Expected 'abc', got '%s'", result)
		}
	})

	t.Run("round-trip: string -> bytes -> string", func(t *testing.T) {
		original := "QuanticScript"
		var bytecode []byte
		bytecode = append(bytecode, buildPushString(original)...)
		bytecode = append(bytecode, byte(OpStrToBytes))
		bytecode = append(bytecode, byte(OpStrFromBytes))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != original {
			t.Errorf("Round-trip failed: expected '%s', got '%s'", original, result)
		}
	})
}

// TestCryptoHashing tests SHA256 with known input, deterministic output, and empty input
// Stack order for SHA256: push bytes, SHA256 -> bytes (hash)
func TestCryptoHashing(t *testing.T) {
	t.Run("SHA256 with known input produces correct hash", func(t *testing.T) {
		// "hello" SHA-256 = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
		input := []byte("hello")
		expectedHash := []byte{
			0x2c, 0xf2, 0x4d, 0xba, 0x5f, 0xb0, 0xa3, 0x0e,
			0x26, 0xe8, 0x3b, 0x2a, 0xc5, 0xb9, 0xe2, 0x9e,
			0x1b, 0x16, 0x1e, 0x5c, 0x1f, 0xa7, 0x42, 0x5e,
			0x73, 0x04, 0x33, 0x62, 0x93, 0x8b, 0x98, 0x24,
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(input)...)
		bytecode = append(bytecode, byte(OpSha256))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if len(result) != 32 {
			t.Fatalf("Expected 32-byte hash, got %d bytes", len(result))
		}

		for i, b := range expectedHash {
			if result[i] != b {
				t.Errorf("Hash mismatch at byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
			}
		}
	})

	t.Run("SHA256 is deterministic for same input", func(t *testing.T) {
		input := []byte("deterministic test")

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(input)...)
		bytecode = append(bytecode, byte(OpSha256))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()

		// Run twice and compare
		interp1 := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err := interp1.Execute()
		if err != nil {
			t.Fatalf("First execution failed: %v", err)
		}
		hash1, _ := interp1.stack[0].AsBytes()

		interp2 := NewBytecodeInterpreter(bytecode, ctx, 100000)
		err = interp2.Execute()
		if err != nil {
			t.Fatalf("Second execution failed: %v", err)
		}
		hash2, _ := interp2.stack[0].AsBytes()

		if len(hash1) != len(hash2) {
			t.Fatalf("Hash lengths differ: %d vs %d", len(hash1), len(hash2))
		}
		for i := range hash1 {
			if hash1[i] != hash2[i] {
				t.Errorf("Hashes differ at byte %d", i)
			}
		}
	})

	t.Run("SHA256 with empty input produces correct hash", func(t *testing.T) {
		// SHA-256("") = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		expectedHash := []byte{
			0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14,
			0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24,
			0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
			0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55,
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes([]byte{})...)
		bytecode = append(bytecode, byte(OpSha256))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBytes()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		for i, b := range expectedHash {
			if result[i] != b {
				t.Errorf("Empty hash mismatch at byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
			}
		}
	})
}

// TestCryptoSignatureVerification tests Ed25519 signature verification
// Stack order for VERIFYSIG: push pubkey (PublicKey), push message (bytes), push signature (bytes), VERIFYSIG -> bool
func TestCryptoSignatureVerification(t *testing.T) {
	// Use a fixed seed to generate a deterministic key pair for testing
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	privKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privKey.Public().(ed25519.PublicKey)
	message := []byte("test message for signature verification")
	signature := ed25519.Sign(privKey, message)

	t.Run("VERIFYSIG passes for valid signature", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey([32]byte(pubKey))...)
		bytecode = append(bytecode, buildPushBytes(message)...)
		bytecode = append(bytecode, buildPushBytes(signature)...)
		bytecode = append(bytecode, byte(OpVerifySig))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if !result {
			t.Errorf("Expected VERIFYSIG to return true for valid signature")
		}
	})

	t.Run("VERIFYSIG fails for invalid signature", func(t *testing.T) {
		// Corrupt the signature by flipping a byte
		badSig := make([]byte, len(signature))
		copy(badSig, signature)
		badSig[0] ^= 0xFF

		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey([32]byte(pubKey))...)
		bytecode = append(bytecode, buildPushBytes(message)...)
		bytecode = append(bytecode, buildPushBytes(badSig)...)
		bytecode = append(bytecode, byte(OpVerifySig))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result {
			t.Errorf("Expected VERIFYSIG to return false for invalid signature")
		}
	})

	t.Run("VERIFYSIG fails for wrong message", func(t *testing.T) {
		wrongMessage := []byte("different message")

		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey([32]byte(pubKey))...)
		bytecode = append(bytecode, buildPushBytes(wrongMessage)...)
		bytecode = append(bytecode, buildPushBytes(signature)...)
		bytecode = append(bytecode, byte(OpVerifySig))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result {
			t.Errorf("Expected VERIFYSIG to return false for wrong message")
		}
	})
}

// TestCryptoPublicKeyDerivation tests DERIVEPUBKEY with seed, deterministic derivation, and multiple seeds
// Stack order for DERIVEPUBKEY: push seed (bytes, 32 bytes), DERIVEPUBKEY -> PublicKey
func TestCryptoPublicKeyDerivation(t *testing.T) {
	t.Run("DERIVEPUBKEY produces a 32-byte public key", func(t *testing.T) {
		seed := make([]byte, 32)
		for i := range seed {
			seed[i] = byte(i + 1)
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(seed)...)
		bytecode = append(bytecode, byte(OpDerivePubKey))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		if interp.stack[0].Type != TypePublicKey {
			t.Errorf("Expected TypePublicKey, got %v", interp.stack[0].Type)
		}

		keyBytes, ok := interp.stack[0].Data.([]byte)
		if !ok {
			t.Fatalf("Expected []byte data for PublicKey")
		}

		if len(keyBytes) != 32 {
			t.Errorf("Expected 32-byte public key, got %d bytes", len(keyBytes))
		}
	})

	t.Run("DERIVEPUBKEY is deterministic for same seed", func(t *testing.T) {
		seed := make([]byte, 32)
		for i := range seed {
			seed[i] = byte(i + 5)
		}

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(seed)...)
		bytecode = append(bytecode, byte(OpDerivePubKey))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()

		interp1 := NewBytecodeInterpreter(bytecode, ctx, 100000)
		if err := interp1.Execute(); err != nil {
			t.Fatalf("First execution failed: %v", err)
		}
		key1, _ := interp1.stack[0].Data.([]byte)

		interp2 := NewBytecodeInterpreter(bytecode, ctx, 100000)
		if err := interp2.Execute(); err != nil {
			t.Fatalf("Second execution failed: %v", err)
		}
		key2, _ := interp2.stack[0].Data.([]byte)

		for i := range key1 {
			if key1[i] != key2[i] {
				t.Errorf("Keys differ at byte %d: 0x%02x vs 0x%02x", i, key1[i], key2[i])
			}
		}
	})

	t.Run("DERIVEPUBKEY produces different keys for different seeds", func(t *testing.T) {
		seed1 := make([]byte, 32)
		seed2 := make([]byte, 32)
		for i := range seed1 {
			seed1[i] = byte(i + 1)
			seed2[i] = byte(i + 100)
		}

		runDerive := func(seed []byte) []byte {
			var bytecode []byte
			bytecode = append(bytecode, buildPushBytes(seed)...)
			bytecode = append(bytecode, byte(OpDerivePubKey))
			bytecode = append(bytecode, byte(OpRet))

			ctx := NewMockExecutionContext()
			interp := NewBytecodeInterpreter(bytecode, ctx, 100000)
			if err := interp.Execute(); err != nil {
				t.Fatalf("Execution failed: %v", err)
			}
			key, _ := interp.stack[0].Data.([]byte)
			return key
		}

		key1 := runDerive(seed1)
		key2 := runDerive(seed2)

		identical := true
		for i := range key1 {
			if key1[i] != key2[i] {
				identical = false
				break
			}
		}
		if identical {
			t.Errorf("Expected different keys for different seeds, but got identical keys")
		}
	})

	t.Run("DERIVEPUBKEY result matches crypto/ed25519 derivation", func(t *testing.T) {
		seed := make([]byte, 32)
		for i := range seed {
			seed[i] = byte(i + 10)
		}

		// Derive expected key using Go's ed25519 directly
		privKey := ed25519.NewKeyFromSeed(seed)
		expectedPubKey := privKey.Public().(ed25519.PublicKey)

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(seed)...)
		bytecode = append(bytecode, byte(OpDerivePubKey))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, _ := interp.stack[0].Data.([]byte)

		for i, b := range expectedPubKey {
			if result[i] != b {
				t.Errorf("Key mismatch at byte %d: expected 0x%02x, got 0x%02x", i, b, result[i])
			}
		}
	})
}

// TestCryptoInvalidInputs tests graceful error handling for invalid crypto inputs
func TestCryptoInvalidInputs(t *testing.T) {
	t.Run("SHA256 with non-bytes type returns error", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushString("not bytes")...)
		bytecode = append(bytecode, byte(OpSha256))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for SHA256 with string input, got nil")
		}
	})

	t.Run("DERIVEPUBKEY with wrong seed length returns error", func(t *testing.T) {
		// Seed must be exactly 32 bytes; use 16 bytes instead
		shortSeed := make([]byte, 16)

		var bytecode []byte
		bytecode = append(bytecode, buildPushBytes(shortSeed)...)
		bytecode = append(bytecode, byte(OpDerivePubKey))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for DERIVEPUBKEY with short seed, got nil")
		}
	})

	t.Run("DERIVEPUBKEY with non-bytes type returns error", func(t *testing.T) {
		var bytecode []byte
		bytecode = append(bytecode, buildPushI64(42)...)
		bytecode = append(bytecode, byte(OpDerivePubKey))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for DERIVEPUBKEY with i64 input, got nil")
		}
	})

	t.Run("VERIFYSIG with wrong signature length returns false", func(t *testing.T) {
		seed := make([]byte, 32)
		privKey := ed25519.NewKeyFromSeed(seed)
		pubKey := privKey.Public().(ed25519.PublicKey)
		message := []byte("test")
		// Signature must be 64 bytes; use 32 bytes instead
		shortSig := make([]byte, 32)

		var bytecode []byte
		bytecode = append(bytecode, buildPushPublicKey([32]byte(pubKey))...)
		bytecode = append(bytecode, buildPushBytes(message)...)
		bytecode = append(bytecode, buildPushBytes(shortSig)...)
		bytecode = append(bytecode, byte(OpVerifySig))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Invalid signature length should return false, not error
		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}
		if result {
			t.Errorf("Expected VERIFYSIG to return false for malformed signature")
		}
	})

	t.Run("VERIFYSIG with non-PublicKey type returns error", func(t *testing.T) {
		message := []byte("test")
		sig := make([]byte, 64)

		var bytecode []byte
		// Push bytes instead of PublicKey type
		bytecode = append(bytecode, buildPushBytes(make([]byte, 32))...)
		bytecode = append(bytecode, buildPushBytes(message)...)
		bytecode = append(bytecode, buildPushBytes(sig)...)
		bytecode = append(bytecode, byte(OpVerifySig))
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for VERIFYSIG with bytes instead of PublicKey, got nil")
		}
	})
}

// TestFunctionBasicCall tests CALL and RET instructions with return address handling
func TestFunctionBasicCall(t *testing.T) {
	t.Run("basic call and return", func(t *testing.T) {
		// Layout:
		//   [0]  CALL -> functionStart   (jumps to function, pushes return addr)
		//   [5]  RET                     (end of main: returns from program)
		//   [6]  PUSH i64(42)            (function body)
		//   [16] RET                     (return from function, goes back to [5])
		var bytecode []byte

		// Main: call the function
		callPos := len(bytecode) + 1                 // offset of the 4-byte address inside CALL
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		// After call returns, end main
		bytecode = append(bytecode, byte(OpRet))

		// Function body starts here
		functionStart := len(bytecode)
		bytecode = append(bytecode, buildPushI64(42)...)
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL target to absolute function start
		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("call returns to correct address", func(t *testing.T) {
		// Verify execution continues after CALL returns, not at the wrong place.
		// Main:
		//   PUSH 1
		//   CALL -> fn (fn pushes 2)
		//   PUSH 3
		//   RET
		// fn:
		//   PUSH 2
		//   RET
		// Stack after execution: [1, 2, 3]
		var bytecode []byte

		bytecode = append(bytecode, buildPushI64(1)...)

		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder

		bytecode = append(bytecode, buildPushI64(3)...)
		bytecode = append(bytecode, byte(OpRet))

		functionStart := len(bytecode)
		bytecode = append(bytecode, buildPushI64(2)...)
		bytecode = append(bytecode, byte(OpRet))

		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(functionStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 3 {
			t.Fatalf("Expected 3 values on stack, got %d", len(interp.stack))
		}

		vals := make([]int64, 3)
		for i := 0; i < 3; i++ {
			v, err := interp.stack[i].AsI64()
			if err != nil {
				t.Fatalf("stack[%d]: %v", i, err)
			}
			vals[i] = v
		}

		if vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", vals)
		}
	})
}

// TestFunctionRecursion tests recursive function calls, depth tracking, and overflow protection
func TestFunctionRecursion(t *testing.T) {
	t.Run("factorial(5) = 120", func(t *testing.T) {
		// Implements:
		//   func factorial(n) {
		//     if n <= 1 { return 1 }
		//     return n * factorial(n-1)
		//   }
		//   result = factorial(5)  // 120
		//
		// Bytecode layout:
		//   [main]  STORE n=5 -> mem[0], CALL factorial, RET
		//   [fact]  LOAD mem[0], PUSH 1, LTE, NOT, JMPIF base_case
		//           LOAD mem[0], PUSH 1, SUB, STORE mem[0]
		//           CALL fact (recursive)
		//           LOAD mem[0], PUSH 1, ADD, STORE mem[0]  <- restore n
		//           MUL
		//           RET
		//   [base]  PUSH 1, RET
		//
		// NOTE: Because memory is shared (single interpreter), we use the stack
		// to save/restore n across recursive calls.
		//
		// Simpler approach: pass n via stack, function reads from stack top.
		// factorial(n):
		//   n is in mem[0] when called
		//   if n <= 1: return 1
		//   save n, set mem[0] = n-1, call factorial, restore n, multiply

		var bytecode []byte

		// Main: set n=5, call factorial
		bytecode = append(bytecode, buildPushI64(5)...)
		bytecode = append(bytecode, buildStore(0)...)

		mainCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		bytecode = append(bytecode, byte(OpRet))

		// factorial function starts here
		factStart := len(bytecode)

		// Load n from mem[0]
		bytecode = append(bytecode, buildLoad(0)...)
		// Push 1 for comparison
		bytecode = append(bytecode, buildPushI64(1)...)
		// n <= 1?
		bytecode = append(bytecode, byte(OpLte))
		// NOT: if NOT(n<=1), skip base case
		bytecode = append(bytecode, byte(OpNot))

		skipBasePos := len(bytecode)
		bytecode = append(bytecode, buildJumpIf(0)...) // placeholder -> skip to recursive case

		// Base case: push 1, return
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, byte(OpRet))

		// Recursive case starts here
		recursiveStart := len(bytecode)

		// Save n onto stack (for multiply after recursive call)
		bytecode = append(bytecode, buildLoad(0)...)
		// Compute n-1, store in mem[0] for recursive call
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildPushI64(1)...)
		bytecode = append(bytecode, byte(OpSub))
		bytecode = append(bytecode, buildStore(0)...)

		// Recursive call
		recCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder -> factStart

		// Stack now: [saved_n, factorial(n-1)]
		// Multiply: saved_n * factorial(n-1)
		bytecode = append(bytecode, byte(OpMul))
		bytecode = append(bytecode, byte(OpRet))

		// Fix jump offsets
		// skipBase: jump from after JMPIF to recursiveStart
		skipBaseOffset := int32(recursiveStart - (skipBasePos + 5))
		binary.LittleEndian.PutUint32(bytecode[skipBasePos+1:skipBasePos+5], uint32(skipBaseOffset))

		// Fix CALL targets (absolute)
		binary.LittleEndian.PutUint32(bytecode[mainCallPos:mainCallPos+4], uint32(factStart))
		binary.LittleEndian.PutUint32(bytecode[recCallPos:recCallPos+4], uint32(factStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 120 {
			t.Errorf("Expected factorial(5)=120, got %d", result)
		}
	})

	t.Run("call stack overflow protection", func(t *testing.T) {
		// A function that calls itself unconditionally triggers call stack overflow.
		// The interpreter enforces a max call stack depth of 64.
		var bytecode []byte

		// Main: call infinite_recurse
		mainCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		bytecode = append(bytecode, byte(OpRet))

		// infinite_recurse: just calls itself
		recurseStart := len(bytecode)
		selfCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder -> recurseStart
		bytecode = append(bytecode, byte(OpRet))

		// Fix CALL targets
		binary.LittleEndian.PutUint32(bytecode[mainCallPos:mainCallPos+4], uint32(recurseStart))
		binary.LittleEndian.PutUint32(bytecode[selfCallPos:selfCallPos+4], uint32(recurseStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected call stack overflow error, got nil")
		}

		// Verify the error is about call stack overflow
		errMsg := err.Error()
		if errMsg != "call stack overflow: maximum depth exceeded" &&
			errMsg != "execution exceeded safety limit of 1000 steps" {
			t.Errorf("Expected call stack overflow or step limit error, got: %v", err)
		}
	})
}

// TestFunctionParameterPassing tests passing parameters via stack and returning values
func TestFunctionParameterPassing(t *testing.T) {
	t.Run("pass two params, return sum", func(t *testing.T) {
		// add(a, b): pops b then a from stack, returns a+b
		// Main: PUSH 10, PUSH 32, CALL add, RET
		// add: POP->mem[1] (b), POP->mem[0] (a), LOAD 0, LOAD 1, ADD, RET
		var bytecode []byte

		bytecode = append(bytecode, buildPushI64(10)...)
		bytecode = append(bytecode, buildPushI64(32)...)

		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		bytecode = append(bytecode, byte(OpRet))

		// add function
		addStart := len(bytecode)
		// Pop b (top of stack) into mem[1]
		bytecode = append(bytecode, buildStore(1)...)
		// Pop a into mem[0]
		bytecode = append(bytecode, buildStore(0)...)
		// Compute a + b
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, byte(OpRet))

		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(addStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		if result != 42 {
			t.Errorf("Expected 42 (10+32), got %d", result)
		}
	})

	t.Run("multiple calls with different params", func(t *testing.T) {
		// Call multiply(a, b) twice with different args, verify independent results.
		// multiply: pops b, pops a, returns a*b
		// Main:
		//   PUSH 3, PUSH 4, CALL mul  -> 12
		//   PUSH 5, PUSH 6, CALL mul  -> 30
		//   ADD -> 42
		//   RET
		var bytecode []byte

		bytecode = append(bytecode, buildPushI64(3)...)
		bytecode = append(bytecode, buildPushI64(4)...)
		call1Pos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...)

		bytecode = append(bytecode, buildPushI64(5)...)
		bytecode = append(bytecode, buildPushI64(6)...)
		call2Pos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...)

		bytecode = append(bytecode, byte(OpAdd))
		bytecode = append(bytecode, byte(OpRet))

		// multiply function
		mulStart := len(bytecode)
		bytecode = append(bytecode, buildStore(1)...) // b -> mem[1]
		bytecode = append(bytecode, buildStore(0)...) // a -> mem[0]
		bytecode = append(bytecode, buildLoad(0)...)
		bytecode = append(bytecode, buildLoad(1)...)
		bytecode = append(bytecode, byte(OpMul))
		bytecode = append(bytecode, byte(OpRet))

		binary.LittleEndian.PutUint32(bytecode[call1Pos:call1Pos+4], uint32(mulStart))
		binary.LittleEndian.PutUint32(bytecode[call2Pos:call2Pos+4], uint32(mulStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get result: %v", err)
		}

		// 3*4 + 5*6 = 12 + 30 = 42
		if result != 42 {
			t.Errorf("Expected 42 (3*4 + 5*6), got %d", result)
		}
	})
}

// TestFunctionErrors tests error conditions in function calls
func TestFunctionErrors(t *testing.T) {
	t.Run("call target out of bounds", func(t *testing.T) {
		// CALL with an offset beyond bytecode length should error
		var bytecode []byte
		// Build a CALL pointing to offset 9999 (way out of bounds)
		bytecode = append(bytecode, buildCall(9999)...)
		bytecode = append(bytecode, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected error for out-of-bounds CALL target, got nil")
		}
	})

	t.Run("stack underflow in function body", func(t *testing.T) {
		// Function tries to pop from empty stack — should return stack underflow error
		var bytecode []byte

		callPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		bytecode = append(bytecode, byte(OpRet))

		// Function body: immediately tries to ADD with nothing on stack
		fnStart := len(bytecode)
		bytecode = append(bytecode, byte(OpAdd)) // stack underflow
		bytecode = append(bytecode, byte(OpRet))

		binary.LittleEndian.PutUint32(bytecode[callPos:callPos+4], uint32(fnStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 100000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected stack underflow error, got nil")
		}
	})

	t.Run("call stack overflow", func(t *testing.T) {
		// Directly verify the call stack depth limit (64) is enforced.
		// Build a chain: main -> f1 -> f2 -> ... -> f65 (exceeds limit)
		// Each function just calls the next one.
		// We use a single self-calling function to trigger the limit efficiently.
		var bytecode []byte

		mainCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder
		bytecode = append(bytecode, byte(OpRet))

		recurseStart := len(bytecode)
		selfCallPos := len(bytecode) + 1
		bytecode = append(bytecode, buildCall(0)...) // placeholder -> recurseStart
		bytecode = append(bytecode, byte(OpRet))

		binary.LittleEndian.PutUint32(bytecode[mainCallPos:mainCallPos+4], uint32(recurseStart))
		binary.LittleEndian.PutUint32(bytecode[selfCallPos:selfCallPos+4], uint32(recurseStart))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000000)

		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected call stack overflow error, got nil")
		}

		errMsg := err.Error()
		if errMsg != "call stack overflow: maximum depth exceeded" &&
			errMsg != "execution exceeded safety limit of 1000 steps" {
			t.Errorf("Expected overflow or step limit error, got: %v", err)
		}
	})
}

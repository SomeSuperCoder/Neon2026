package quanticscript

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// TestCrossProgramInvocation tests the INVOKE and INVOKERET instructions
func TestCrossProgramInvocation(t *testing.T) {
	t.Run("invoke depth tracking", func(t *testing.T) {
		// Create a simple bytecode program
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()

		// Test depth 0 (top-level)
		interp := NewBytecodeInterpreter(bytecode, ctx, 10000)
		if interp.GetInvokeDepth() != 0 {
			t.Errorf("Expected invoke depth 0, got %d", interp.GetInvokeDepth())
		}

		// Test depth 2 (nested invocation)
		interp2 := NewBytecodeInterpreterWithDepth(bytecode, ctx, 10000, 2)
		if interp2.GetInvokeDepth() != 2 {
			t.Errorf("Expected invoke depth 2, got %d", interp2.GetInvokeDepth())
		}
	})

	t.Run("invoke depth limit", func(t *testing.T) {
		// Create bytecode that attempts to invoke another program
		programID := filestore.GenerateFileID([]byte("test_program"))

		bytecode := []byte{byte(OpPush), byte(TypeFileID)}
		bytecode = append(bytecode, programID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeBytes), 0, 0, 0, 0, 0, 0, 0, 0)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64), 100, 0, 0, 0, 0, 0, 0, 0)
		bytecode = append(bytecode, byte(OpInvoke), byte(OpRet))

		ctx := NewMockExecutionContext()

		// Test at max depth - should fail
		interp := NewBytecodeInterpreterWithDepth(bytecode, ctx, 10000, MaxInvokeDepth)
		err := interp.Execute()
		if err == nil {
			t.Error("Expected error when invoking at max depth")
		}
	})

	t.Run("invoke requires declared program", func(t *testing.T) {
		programID := filestore.GenerateFileID([]byte("undeclared_program"))

		bytecode := []byte{byte(OpPush), byte(TypeFileID)}
		bytecode = append(bytecode, programID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeBytes), 0, 0, 0, 0, 0, 0, 0, 0)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64), 100, 0, 0, 0, 0, 0, 0, 0)
		bytecode = append(bytecode, byte(OpInvoke), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interp.Execute()
		if err == nil {
			t.Error("Expected error when invoking undeclared program")
		}
	})

	t.Run("invoke budget validation", func(t *testing.T) {
		programID := filestore.GenerateFileID([]byte("test_program"))

		bytecode := []byte{byte(OpPush), byte(TypeFileID)}
		bytecode = append(bytecode, programID[:]...)
		bytecode = append(bytecode, byte(OpPush), byte(TypeBytes), 0, 0, 0, 0, 0, 0, 0, 0)
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64), 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x7F)
		bytecode = append(bytecode, byte(OpInvoke), byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interp.Execute()
		if err == nil {
			t.Error("Expected error when invoking with insufficient budget")
		}
	})

	t.Run("invokeret returns result", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bytecode, ctx, 10000)

		err := interp.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interp.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interp.stack))
		}

		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("Failed to get i64: %v", err)
		}

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})
}

// TestInvokeInstructions tests the INVOKE and INVOKERET opcodes are defined
func TestInvokeInstructions(t *testing.T) {
	if OpInvoke == 0 {
		t.Error("OpInvoke should be defined")
	}
	if OpInvokeRet == 0 {
		t.Error("OpInvokeRet should be defined")
	}

	if OpcodeNames[OpInvoke] != "INVOKE" {
		t.Errorf("Expected OpInvoke name to be 'INVOKE', got '%s'", OpcodeNames[OpInvoke])
	}
	if OpcodeNames[OpInvokeRet] != "INVOKERET" {
		t.Errorf("Expected OpInvokeRet name to be 'INVOKERET', got '%s'", OpcodeNames[OpInvokeRet])
	}

	invokeCost := GetInstructionCost(OpInvoke)
	if invokeCost != 200 {
		t.Errorf("Expected OpInvoke cost to be 200, got %d", invokeCost)
	}

	invokeRetCost := GetInstructionCost(OpInvokeRet)
	if invokeRetCost != 10 {
		t.Errorf("Expected OpInvokeRet cost to be 10, got %d", invokeRetCost)
	}
}

package quanticscript

import (
	"fmt"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// MockExecutionContext implements ExecutionContext for testing
type MockExecutionContext struct {
	files     map[filestore.FileID]*filestore.File
	signers   []transaction.PublicKey
	instrData []byte
	programID filestore.FileID
}

func NewMockExecutionContext() *MockExecutionContext {
	return &MockExecutionContext{
		files:   make(map[filestore.FileID]*filestore.File),
		signers: []transaction.PublicKey{},
	}
}

func (m *MockExecutionContext) GetFile(fileID filestore.FileID) (*filestore.File, error) {
	if file, ok := m.files[fileID]; ok {
		return file, nil
	}
	return nil, fmt.Errorf("file not found")
}

func (m *MockExecutionContext) GetFileMut(fileID filestore.FileID) (*filestore.File, error) {
	return m.GetFile(fileID)
}

func (m *MockExecutionContext) UpdateFile(file *filestore.File) error {
	m.files[file.ID] = file
	return nil
}

func (m *MockExecutionContext) GetFileBalance(fileID filestore.FileID) (int64, error) {
	file, err := m.GetFile(fileID)
	if err != nil {
		return 0, err
	}
	return file.Balance, nil
}

func (m *MockExecutionContext) UpdateFileBalance(fileID filestore.FileID, delta int64) error {
	file, err := m.GetFile(fileID)
	if err != nil {
		return err
	}
	file.Balance += delta
	return m.UpdateFile(file)
}

func (m *MockExecutionContext) HasSigner(pubkey transaction.PublicKey) bool {
	for _, signer := range m.signers {
		if signer == pubkey {
			return true
		}
	}
	return false
}

func (m *MockExecutionContext) GetInstructionData() []byte {
	return m.instrData
}

func (m *MockExecutionContext) GetProgramID() filestore.FileID {
	return m.programID
}

func (m *MockExecutionContext) GetSigners() []transaction.PublicKey {
	return m.signers
}

func (m *MockExecutionContext) QueryBlock(blockHash []byte) ([]byte, error) {
	return nil, fmt.Errorf("block query not implemented in mock")
}

func (m *MockExecutionContext) QueryTransaction(txID transaction.TxID) ([]byte, error) {
	return nil, fmt.Errorf("transaction query not implemented in mock")
}

func (m *MockExecutionContext) QueryInstruction(txID transaction.TxID, instrIndex uint32) ([]byte, error) {
	return nil, fmt.Errorf("instruction query not implemented in mock")
}

func (m *MockExecutionContext) InvokeProgram(programID filestore.FileID, invokeData []byte, computeBudget int64, depth int) ([]byte, error) {
	return nil, fmt.Errorf("cross-program invocation not implemented in mock")
}

func (m *MockExecutionContext) GetDeclaredPrograms() []filestore.FileID {
	return []filestore.FileID{}
}

// TestStackOperations tests basic stack operations
func TestStackOperations(t *testing.T) {
	// Create bytecode: PUSH 42, PUSH 10, ADD, RET
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0, // PUSH 42
		byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0, // PUSH 10
		byte(OpAdd), // ADD
		byte(OpRet), // RET
	}

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

	err := interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check that we have one value on the stack (52)
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	if result != 52 {
		t.Errorf("Expected result 52, got %d", result)
	}
}

// TestMemoryOperations tests load and store operations
func TestMemoryOperations(t *testing.T) {
	// Create bytecode: PUSH 100, STORE 0, LOAD 0, RET
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 100, 0, 0, 0, 0, 0, 0, 0, // PUSH 100
		byte(OpStore), 0, 0, // STORE offset 0
		byte(OpLoad), 0, 0, // LOAD offset 0
		byte(OpRet), // RET
	}

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

	err := interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check that we have one value on the stack (100)
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	if result != 100 {
		t.Errorf("Expected result 100, got %d", result)
	}
}

// TestComputeBudget tests that compute budget is enforced
func TestComputeBudget(t *testing.T) {
	// Create bytecode with many operations
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 1, 0, 0, 0, 0, 0, 0, 0, // PUSH 1 (cost 1)
		byte(OpPush), byte(TypeI64), 2, 0, 0, 0, 0, 0, 0, 0, // PUSH 2 (cost 1)
		byte(OpAdd), // ADD (cost 2)
		byte(OpRet), // RET (cost 2)
	}

	ctx := NewMockExecutionContext()

	// Set budget too low (need at least 6 units)
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 5)

	err := interpreter.Execute()
	if err == nil {
		t.Fatal("Expected out of compute budget error")
	}

	if err.Error() != "out of compute budget" {
		t.Errorf("Expected 'out of compute budget' error, got: %v", err)
	}
}

// TestComparisonOperations tests comparison operations
func TestComparisonOperations(t *testing.T) {
	// Create bytecode: PUSH 10, PUSH 20, LT, RET
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0, // PUSH 10
		byte(OpPush), byte(TypeI64), 20, 0, 0, 0, 0, 0, 0, 0, // PUSH 20
		byte(OpLt),  // LT (10 < 20 = true)
		byte(OpRet), // RET
	}

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

	err := interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check result
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsBool()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	if !result {
		t.Errorf("Expected true, got false")
	}
}

// TestAllArithmeticOperations tests all arithmetic operations
func TestAllArithmeticOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		a, b     int64
		expected int64
	}{
		{"SUB", OpSub, 50, 8, 42},
		{"MUL", OpMul, 6, 7, 42},
		{"DIV", OpDiv, 84, 2, 42},
		{"MOD", OpMod, 47, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := []byte{
				byte(OpPush), byte(TypeI64), byte(tt.a), byte(tt.a >> 8), byte(tt.a >> 16), byte(tt.a >> 24),
				byte(tt.a >> 32), byte(tt.a >> 40), byte(tt.a >> 48), byte(tt.a >> 56),
				byte(OpPush), byte(TypeI64), byte(tt.b), byte(tt.b >> 8), byte(tt.b >> 16), byte(tt.b >> 24),
				byte(tt.b >> 32), byte(tt.b >> 40), byte(tt.b >> 48), byte(tt.b >> 56),
				byte(tt.op),
				byte(OpRet),
			}

			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

			err := interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsI64()
			if err != nil {
				t.Fatalf("Failed to get result: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestDivisionByZero tests error handling for division by zero
func TestDivisionByZero(t *testing.T) {
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
		byte(OpPush), byte(TypeI64), 0, 0, 0, 0, 0, 0, 0, 0,
		byte(OpDiv),
		byte(OpRet),
	}

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

	err := interpreter.Execute()
	if err == nil {
		t.Fatal("Expected division by zero error")
	}

	if err.Error() != "division by zero" {
		t.Errorf("Expected 'division by zero' error, got: %v", err)
	}
}

// TestAllComparisonOperations tests all comparison operations
func TestAllComparisonOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		a, b     int64
		expected bool
	}{
		{"EQ_true", OpEq, 42, 42, true},
		{"EQ_false", OpEq, 42, 43, false},
		{"LT_true", OpLt, 10, 20, true},
		{"LT_false", OpLt, 20, 10, false},
		{"GT_true", OpGt, 20, 10, true},
		{"GT_false", OpGt, 10, 20, false},
		{"LTE_true_less", OpLte, 10, 20, true},
		{"LTE_true_equal", OpLte, 20, 20, true},
		{"LTE_false", OpLte, 30, 20, false},
		{"GTE_true_greater", OpGte, 30, 20, true},
		{"GTE_true_equal", OpGte, 20, 20, true},
		{"GTE_false", OpGte, 10, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytecode := []byte{
				byte(OpPush), byte(TypeI64), byte(tt.a), byte(tt.a >> 8), byte(tt.a >> 16), byte(tt.a >> 24),
				byte(tt.a >> 32), byte(tt.a >> 40), byte(tt.a >> 48), byte(tt.a >> 56),
				byte(OpPush), byte(TypeI64), byte(tt.b), byte(tt.b >> 8), byte(tt.b >> 16), byte(tt.b >> 24),
				byte(tt.b >> 32), byte(tt.b >> 40), byte(tt.b >> 48), byte(tt.b >> 56),
				byte(tt.op),
				byte(OpRet),
			}

			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

			err := interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsBool()
			if err != nil {
				t.Fatalf("Failed to get result: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestLogicalOperations tests logical operations
func TestLogicalOperations(t *testing.T) {
	tests := []struct {
		name     string
		op       Opcode
		a, b     bool
		expected bool
		unary    bool
	}{
		{"AND_true", OpAnd, true, true, true, false},
		{"AND_false", OpAnd, true, false, false, false},
		{"OR_true", OpOr, true, false, true, false},
		{"OR_false", OpOr, false, false, false, false},
		{"NOT_true", OpNot, true, false, false, true},
		{"NOT_false", OpNot, false, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bytecode []byte

			if tt.unary {
				// Unary operation (NOT)
				aVal := byte(0)
				if tt.a {
					aVal = 1
				}
				bytecode = []byte{
					byte(OpPush), byte(TypeBool), aVal,
					byte(tt.op),
					byte(OpRet),
				}
			} else {
				// Binary operation (AND, OR)
				aVal := byte(0)
				if tt.a {
					aVal = 1
				}
				bVal := byte(0)
				if tt.b {
					bVal = 1
				}
				bytecode = []byte{
					byte(OpPush), byte(TypeBool), aVal,
					byte(OpPush), byte(TypeBool), bVal,
					byte(tt.op),
					byte(OpRet),
				}
			}

			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

			err := interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsBool()
			if err != nil {
				t.Fatalf("Failed to get result: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestStackOperations tests DUP and SWAP operations
func TestStackOperationsDupSwap(t *testing.T) {
	// Test DUP: PUSH 42, DUP, ADD
	t.Run("DUP", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpDup),
			byte(OpAdd),
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, _ := interpreter.stack[0].AsI64()
		if result != 84 {
			t.Errorf("Expected 84, got %d", result)
		}
	})

	// Test SWAP: PUSH 10, PUSH 20, SWAP, SUB (should be 20-10=10)
	t.Run("SWAP", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0,
			byte(OpPush), byte(TypeI64), 20, 0, 0, 0, 0, 0, 0, 0,
			byte(OpSwap),
			byte(OpSub),
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, _ := interpreter.stack[0].AsI64()
		if result != 10 {
			t.Errorf("Expected 10, got %d", result)
		}
	})
}

// TestControlFlow tests jump and call/return operations
func TestControlFlow(t *testing.T) {
	// Test JMP: skip over an instruction
	t.Run("JMP", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 1, 0, 0, 0, 0, 0, 0, 0,
			byte(OpJmp), 11, 0, 0, 0, // Jump forward 11 bytes (skip next PUSH)
			byte(OpPush), byte(TypeI64), 99, 0, 0, 0, 0, 0, 0, 0, // This should be skipped
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Should only have 1 on stack, not 99
		result, _ := interpreter.stack[0].AsI64()
		if result != 1 {
			t.Errorf("Expected 1, got %d", result)
		}
	})

	// Test JMPIF: conditional jump
	t.Run("JMPIF_true", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeBool), 1, // Push true
			byte(OpJmpIf), 11, 0, 0, 0, // Jump if true
			byte(OpPush), byte(TypeI64), 99, 0, 0, 0, 0, 0, 0, 0, // Should be skipped
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Stack should be empty (only had the bool which was popped)
		if len(interpreter.stack) != 0 {
			t.Errorf("Expected empty stack, got %d items", len(interpreter.stack))
		}
	})

	t.Run("JMPIF_false", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeBool), 0, // Push false
			byte(OpJmpIf), 11, 0, 0, 0, // Don't jump
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0, // Should execute
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		result, _ := interpreter.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	t.Run("StackUnderflow", func(t *testing.T) {
		bytecode := []byte{
			byte(OpAdd), // No values on stack
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err == nil {
			t.Fatal("Expected stack underflow error")
		}
	})

	t.Run("InvalidOpcode", func(t *testing.T) {
		bytecode := []byte{
			0xFF, // Invalid opcode
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err == nil {
			t.Fatal("Expected unknown opcode error")
		}
	})

	t.Run("MemoryOutOfBounds", func(t *testing.T) {
		bytecode := []byte{
			byte(OpLoad), 255, 255, // Load from offset 65535 (out of bounds)
			byte(OpRet),
		}

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)

		err := interpreter.Execute()
		if err == nil {
			t.Fatal("Expected memory out of bounds error")
		}
	})
}

// TestCostMeteringAccuracy tests that costs are accurately deducted
func TestCostMeteringAccuracy(t *testing.T) {
	// Create bytecode with known costs
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0, // Cost: 1
		byte(OpPush), byte(TypeI64), 20, 0, 0, 0, 0, 0, 0, 0, // Cost: 1
		byte(OpAdd), // Cost: 2
		byte(OpRet), // Cost: 2
	}

	initialBudget := int64(1000)
	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, initialBudget)

	err := interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Calculate expected remaining budget
	expectedRemaining := initialBudget - 1 - 1 - 2 - 2 // 994
	actualRemaining := interpreter.GetComputeBudget()

	if actualRemaining != expectedRemaining {
		t.Errorf("Expected remaining budget %d, got %d", expectedRemaining, actualRemaining)
	}
}

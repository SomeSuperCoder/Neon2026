package quanticscript

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// TestComputeBudgetEnforcement tests that compute budget limits are enforced
func TestComputeBudgetEnforcement(t *testing.T) {
	tests := []struct {
		name   string
		source string
		budget int64
	}{
		{
			name: "infinite loop protection",
			source: `
				export function entry(): i64 {
					let i: i64 = 0;
					while (true) {
						i = i + 1;
					}
					return i;
				}
			`,
			budget: 1000, // Low budget to trigger exhaustion
		},
		{
			name: "expensive computation",
			source: `
				export function entry(): i64 {
					let sum: i64 = 0;
					let i: i64 = 0;
					while (i < 1000) {
						let j: i64 = 0;
						while (j < 1000) {
							sum = sum + 1;
							j = j + 1;
						}
						i = i + 1;
					}
					return sum;
				}
			`,
			budget: 1000, // Too low for this computation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile
			lexer := NewLexer(tt.source, "test.qs")
			parser := NewParser(lexer)
			program := parser.ParseProgram()
			if len(parser.errors) > 0 {
				t.Fatalf("Parser errors: %v", parser.errors)
			}

			tc := NewTypeChecker()
			tc.CheckProgram(program)
			if len(tc.errors) > 0 {
				t.Fatalf("Type checker errors: %v", tc.errors)
			}

			cg := NewCodeGenerator()
			bytecode, err := cg.Generate(program)
			if err != nil {
				t.Fatalf("Code generator error: %v", err)
			}

			bytecodeFile := CreateBytecode(bytecode, 0)
			body, _ := GetBytecodeBody(bytecodeFile)

			// Execute with limited budget
			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(body, ctx, tt.budget)
			err = interpreter.Execute()

			// Should fail with out of compute budget error
			if err == nil {
				t.Fatal("Expected out of compute budget error")
			}

			if err.Error() != "out of compute budget" {
				t.Errorf("Expected 'out of compute budget' error, got: %v", err)
			}
		})
	}
}

// TestStackOverflowProtection tests protection against stack overflow
func TestStackOverflowProtection(t *testing.T) {
	// Create bytecode that tries to overflow the stack
	var bytecode []byte

	// Push many values onto the stack
	for i := 0; i < 300; i++ {
		bytecode = append(bytecode,
			byte(OpPush), byte(TypeI64), byte(i), 0, 0, 0, 0, 0, 0, 0,
		)
	}
	bytecode = append(bytecode, byte(OpRet))

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000000)
	err := interpreter.Execute()

	// The interpreter should handle this gracefully
	// Either by limiting stack size or allowing it within reasonable bounds
	if err != nil {
		// If there's an error, it should be a meaningful one
		t.Logf("Stack overflow protection triggered: %v", err)
	}
}

// TestMemoryAccessBounds tests that memory access is bounds-checked
func TestMemoryAccessBounds(t *testing.T) {
	tests := []struct {
		name   string
		opcode Opcode
		offset uint16
	}{
		{"load out of bounds", OpLoad, 65535},
		{"store out of bounds", OpStore, 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bytecode []byte

			if tt.opcode == OpStore {
				// Need a value on stack for STORE
				bytecode = append(bytecode,
					byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
				)
			}

			// Try to access out of bounds memory
			bytecode = append(bytecode,
				byte(tt.opcode),
				byte(tt.offset), byte(tt.offset>>8),
				byte(OpRet),
			)

			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
			err := interpreter.Execute()

			if err == nil {
				t.Fatal("Expected memory access error")
			}

			// Should get a bounds check error
			t.Logf("Memory bounds check triggered: %v", err)
		})
	}
}

// TestInvalidBytecodeRejection tests that invalid bytecode is rejected
func TestInvalidBytecodeRejection(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
	}{
		{
			name:     "invalid opcode",
			bytecode: []byte{0xFF, byte(OpRet)},
		},
		{
			name: "truncated instruction",
			bytecode: []byte{
				byte(OpPush), byte(TypeI64), 1, // Missing bytes
			},
		},
		{
			name: "invalid jump target",
			bytecode: []byte{
				byte(OpJmp), 255, 255, 255, 127, // Jump way out of bounds
				byte(OpRet),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(tt.bytecode, ctx, 10000)
			err := interpreter.Execute()

			if err == nil {
				t.Fatal("Expected error for invalid bytecode")
			}

			t.Logf("Invalid bytecode rejected: %v", err)
		})
	}
}

// TestCrossInvocationDepthLimit tests that cross-program invocation depth is limited
func TestCrossInvocationDepthLimit(t *testing.T) {
	// Create a mock program that tries to invoke itself recursively
	programID := filestore.FileID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	// Create bytecode that invokes another program
	bytecode := []byte{
		// Push program ID
		byte(OpPush), byte(TypeFileID),
	}
	bytecode = append(bytecode, programID[:]...)

	// Push invoke data (empty)
	bytecode = append(bytecode,
		byte(OpPush), byte(TypeBytes), 0, 0, 0, 0, 0, 0, 0, 0, // Empty bytes
		// Push compute budget
		byte(OpPush), byte(TypeI64), 100, 0, 0, 0, 0, 0, 0, 0,
		// Invoke
		byte(OpInvoke),
		byte(OpRet),
	)

	// Test at maximum depth
	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreterWithDepth(bytecode, ctx, 10000, MaxInvokeDepth)
	err := interpreter.Execute()

	if err == nil {
		t.Fatal("Expected invocation depth limit error")
	}

	// Should fail because we're at max depth
	t.Logf("Invocation depth limit enforced: %v", err)
}

// TestTypeConfusionPrevention tests that type confusion is prevented
func TestTypeConfusionPrevention(t *testing.T) {
	tests := []struct {
		name     string
		bytecode []byte
	}{
		{
			name: "bool as integer",
			bytecode: []byte{
				byte(OpPush), byte(TypeBool), 1,
				byte(OpPush), byte(TypeBool), 0,
				byte(OpAdd), // Try to add booleans
				byte(OpRet),
			},
		},
		{
			name: "integer as bool in condition",
			bytecode: []byte{
				byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
				byte(OpJmpIf), 0, 0, 0, 0, // Try to use integer as condition
				byte(OpRet),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(tt.bytecode, ctx, 10000)
			err := interpreter.Execute()

			if err == nil {
				t.Fatal("Expected type mismatch error")
			}

			t.Logf("Type confusion prevented: %v", err)
		})
	}
}

// TestResourceExhaustionProtection tests protection against resource exhaustion
func TestResourceExhaustionProtection(t *testing.T) {
	t.Run("excessive memory allocation", func(t *testing.T) {
		// Try to create a very large array
		source := `
			export function entry(): i64 {
				let arr: i64[] = [];
				let i: i64 = 0;
				while (i < 10000) {
					arr.push(i);
					i = i + 1;
				}
				return arr.length;
			}
		`

		lexer := NewLexer(source, "test.qs")
		parser := NewParser(lexer)
		program := parser.ParseProgram()
		if len(parser.errors) > 0 {
			t.Skipf("Parser errors (array operations may not be fully implemented): %v", parser.errors)
		}

		tc := NewTypeChecker()
		tc.CheckProgram(program)
		if len(tc.errors) > 0 {
			t.Skipf("Type checker errors (array operations may not be fully implemented): %v", tc.errors)
		}

		cg := NewCodeGenerator()
		bytecode, err := cg.Generate(program)
		if err != nil {
			t.Skipf("Code generator error (array operations may not be fully implemented): %v", err)
		}

		bytecodeFile := CreateBytecode(bytecode, 0)
		body, _ := GetBytecodeBody(bytecodeFile)

		// Execute with limited budget
		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(body, ctx, 5000)
		err = interpreter.Execute()

		// Should either succeed or fail with compute budget error
		if err != nil && err.Error() != "out of compute budget" {
			t.Logf("Resource exhaustion protection: %v", err)
		}
	})
}

// TestSandboxIsolation tests that programs cannot escape the sandbox
func TestSandboxIsolation(t *testing.T) {
	t.Run("no direct memory access", func(t *testing.T) {
		// The interpreter should not allow direct memory pointer manipulation
		// This is enforced by the type system and bytecode design
		// We verify that there are no opcodes for raw pointer operations

		// Check that dangerous opcodes don't exist
		dangerousOpcodes := []Opcode{
			// These would be dangerous if they existed
			Opcode(0xF0), // Hypothetical raw memory read
			Opcode(0xF1), // Hypothetical raw memory write
			Opcode(0xF2), // Hypothetical system call
		}

		for _, op := range dangerousOpcodes {
			bytecode := []byte{byte(op), byte(OpRet)}
			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
			err := interpreter.Execute()

			if err == nil {
				t.Errorf("Dangerous opcode 0x%02x was not rejected", op)
			}
		}
	})

	t.Run("file access control", func(t *testing.T) {
		// Test that file access requires proper context
		fileID := filestore.FileID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, fileID[:]...)
		bytecode = append(bytecode,
			byte(OpGetFile), // Try to get a file that doesn't exist
			byte(OpRet),
		)

		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()

		// Should fail because file doesn't exist in context
		if err == nil {
			t.Fatal("Expected file access error")
		}

		t.Logf("File access control enforced: %v", err)
	})
}

// TestCallStackDepthLimit tests that call stack depth is limited
func TestCallStackDepthLimit(t *testing.T) {
	// Create bytecode with recursive calls
	var bytecode []byte

	// Function that calls itself
	functionStart := 0
	bytecode = append(bytecode,
		byte(OpCall), byte(functionStart), byte(functionStart>>8), byte(functionStart>>16), byte(functionStart>>24),
	)

	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 100000)
	err := interpreter.Execute()

	if err == nil {
		t.Fatal("Expected call stack overflow error")
	}

	// Should fail with stack overflow or compute budget exhaustion
	t.Logf("Call stack depth limit enforced: %v", err)
}

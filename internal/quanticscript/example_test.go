package quanticscript

import (
	"os"
	"testing"
)

// TestLevel1Examples tests the Level 1 example programs
func TestLevel1Examples(t *testing.T) {
	tests := []struct {
		name           string
		bytecodeFile   string
		expectedResult int64
	}{
		{
			name:           "01_literals",
			bytecodeFile:   "../../examples/01_literals.qsb",
			expectedResult: 42,
		},
		{
			name:           "02_variables",
			bytecodeFile:   "../../examples/02_variables.qsb",
			expectedResult: 30, // 10 + 20
		},
		{
			name:           "03_expressions",
			bytecodeFile:   "../../examples/03_expressions.qsb",
			expectedResult: 15, // (5 + 3) * 2 - 5 / 3 = 8 * 2 - 1 = 16 - 1 = 15
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read bytecode file
			bytecode, err := os.ReadFile(tt.bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to read bytecode file %s: %v", tt.bytecodeFile, err)
			}

			// Extract body (skip header)
			body, err := GetBytecodeBody(bytecode)
			if err != nil {
				t.Fatalf("Failed to extract bytecode body: %v", err)
			}

			// Create mock execution context
			ctx := NewMockExecutionContext()

			// Create interpreter with sufficient budget
			interpreter := NewBytecodeInterpreter(body, ctx, 1000000)

			// Execute
			err = interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			// Check result (should be on top of stack)
			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsI64()
			if err != nil {
				t.Fatalf("Failed to get i64 result: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, result)
			}
		})
	}
}

// TestLevel2Examples tests the Level 2 example programs (control flow)
func TestLevel2Examples(t *testing.T) {
	tests := []struct {
		name           string
		bytecodeFile   string
		expectedResult int64
	}{
		{
			name:           "04_conditionals",
			bytecodeFile:   "../../examples/04_conditionals.qsb",
			expectedResult: 1, // x=10 > 5, so returns 1
		},
		{
			name:           "05_while_loop",
			bytecodeFile:   "../../examples/05_while_loop.qsb",
			expectedResult: 55, // sum of 1 to 10
		},
		{
			name:           "06_for_loop",
			bytecodeFile:   "../../examples/06_for_loop.qsb",
			expectedResult: 55, // sum of 1 to 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read bytecode file
			bytecode, err := os.ReadFile(tt.bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to read bytecode file %s: %v", tt.bytecodeFile, err)
			}

			// Extract body (skip header)
			body, err := GetBytecodeBody(bytecode)
			if err != nil {
				t.Fatalf("Failed to extract bytecode body: %v", err)
			}

			// Create mock execution context
			ctx := NewMockExecutionContext()

			// Create interpreter with sufficient budget
			interpreter := NewBytecodeInterpreter(body, ctx, 1000000)

			// Execute
			err = interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			// Check result (should be on top of stack)
			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsI64()
			if err != nil {
				t.Fatalf("Failed to get i64 result: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, result)
			}
		})
	}
}

// TestLevel3Examples tests the Level 3 example programs (functions)
func TestLevel3Examples(t *testing.T) {
	tests := []struct {
		name           string
		bytecodeFile   string
		expectedResult int64
	}{
		{
			name:           "07_functions",
			bytecodeFile:   "../../examples/07_functions.qsb",
			expectedResult: 30, // add(10, 20) = 30
		},
		{
			name:           "08_recursion",
			bytecodeFile:   "../../examples/08_recursion.qsb",
			expectedResult: 120, // factorial(5) = 5 * 4 * 3 * 2 * 1 = 120
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read bytecode file
			bytecode, err := os.ReadFile(tt.bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to read bytecode file %s: %v", tt.bytecodeFile, err)
			}

			// Extract body (skip header)
			body, err := GetBytecodeBody(bytecode)
			if err != nil {
				t.Fatalf("Failed to extract bytecode body: %v", err)
			}

			// Create mock execution context
			ctx := NewMockExecutionContext()

			// Create interpreter with sufficient budget
			interpreter := NewBytecodeInterpreter(body, ctx, 1000000)

			// Execute
			err = interpreter.Execute()
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			// Check result (should be on top of stack)
			if len(interpreter.stack) != 1 {
				t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
			}

			result, err := interpreter.stack[0].AsI64()
			if err != nil {
				t.Fatalf("Failed to get i64 result: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, result)
			}
		})
	}
}

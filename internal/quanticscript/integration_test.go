package quanticscript

import (
	"testing"
)

// TestAssemblerInterpreterIntegration tests that assembled code can be executed by the interpreter
func TestAssemblerInterpreterIntegration(t *testing.T) {
	tests := []struct {
		name           string
		assembly       string
		expectedStack  []Value
		expectedBudget int64
	}{
		{
			name: "simple arithmetic",
			assembly: `
				PUSH i64 10
				PUSH i64 20
				ADD
				RET
			`,
			expectedStack:  []Value{NewI64(30)},
			expectedBudget: 10000 - 1 - 1 - 2 - 2, // budget - PUSH - PUSH - ADD - RET
		},
		{
			name: "memory operations",
			assembly: `
				PUSH i64 42
				STORE 0
				LOAD 0
				PUSH i64 8
				ADD
				RET
			`,
			expectedStack:  []Value{NewI64(50)},
			expectedBudget: 10000 - 1 - 3 - 3 - 1 - 2 - 2, // PUSH - STORE - LOAD - PUSH - ADD - RET
		},
		{
			name: "conditional jump",
			assembly: `
				PUSH i64 5
				PUSH i64 5
				EQ
				JMPIF equal
				PUSH i64 0
				RET
			equal:
				PUSH i64 1
				RET
			`,
			expectedStack:  []Value{NewI64(1)},
			expectedBudget: 10000 - 1 - 1 - 2 - 2 - 1 - 2, // PUSH - PUSH - EQ - JMPIF - PUSH - RET
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assemble the code
			bytecode, err := AssembleToBody(tt.assembly)
			if err != nil {
				t.Fatalf("Assemble() error = %v", err)
			}

			// Create a mock execution context
			ctx := &MockExecutionContext{}

			// Create interpreter and execute
			interp := NewBytecodeInterpreter(bytecode, ctx, 10000)
			err = interp.Execute()
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			// Check stack
			if len(interp.stack) != len(tt.expectedStack) {
				t.Errorf("Stack length = %d, want %d", len(interp.stack), len(tt.expectedStack))
			}

			for i, expected := range tt.expectedStack {
				if i >= len(interp.stack) {
					break
				}
				actual := interp.stack[i]
				if actual.Type != expected.Type {
					t.Errorf("Stack[%d] type = %v, want %v", i, actual.Type, expected.Type)
				}
				if actual.Type == TypeI64 {
					actualVal, _ := actual.AsI64()
					expectedVal, _ := expected.AsI64()
					if actualVal != expectedVal {
						t.Errorf("Stack[%d] value = %d, want %d", i, actualVal, expectedVal)
					}
				}
			}

			// Check compute budget
			if interp.GetComputeBudget() != tt.expectedBudget {
				t.Errorf("Compute budget = %d, want %d", interp.GetComputeBudget(), tt.expectedBudget)
			}
		})
	}
}

// TestDisassemblerRoundTrip tests that disassembled code can be reassembled and executed
func TestDisassemblerRoundTrip(t *testing.T) {
	original := `
		PUSH i64 100
		PUSH i64 50
		SUB
		STORE 0
		LOAD 0
		PUSH i64 2
		MUL
		RET
	`

	// Assemble original
	bytecode1, err := AssembleToBody(original)
	if err != nil {
		t.Fatalf("First Assemble() error = %v", err)
	}

	// Execute original
	ctx1 := &MockExecutionContext{}
	interp1 := NewBytecodeInterpreter(bytecode1, ctx1, 10000)
	err = interp1.Execute()
	if err != nil {
		t.Fatalf("First Execute() error = %v", err)
	}

	// Disassemble
	disassembled, err := DisassembleBody(bytecode1)
	if err != nil {
		t.Fatalf("Disassemble() error = %v", err)
	}

	// Reassemble
	bytecode2, err := AssembleToBody(disassembled)
	if err != nil {
		t.Fatalf("Second Assemble() error = %v", err)
	}

	// Execute reassembled
	ctx2 := &MockExecutionContext{}
	interp2 := NewBytecodeInterpreter(bytecode2, ctx2, 10000)
	err = interp2.Execute()
	if err != nil {
		t.Fatalf("Second Execute() error = %v", err)
	}

	// Compare results
	if len(interp1.stack) != len(interp2.stack) {
		t.Errorf("Stack length mismatch: %d vs %d", len(interp1.stack), len(interp2.stack))
	}

	for i := range interp1.stack {
		if i >= len(interp2.stack) {
			break
		}
		if interp1.stack[i].Type != interp2.stack[i].Type {
			t.Errorf("Stack[%d] type mismatch: %v vs %v", i, interp1.stack[i].Type, interp2.stack[i].Type)
		}
		if interp1.stack[i].Type == TypeI64 {
			val1, _ := interp1.stack[i].AsI64()
			val2, _ := interp2.stack[i].AsI64()
			if val1 != val2 {
				t.Errorf("Stack[%d] value mismatch: %d vs %d", i, val1, val2)
			}
		}
	}

	if interp1.GetComputeBudget() != interp2.GetComputeBudget() {
		t.Errorf("Compute budget mismatch: %d vs %d", interp1.GetComputeBudget(), interp2.GetComputeBudget())
	}
}

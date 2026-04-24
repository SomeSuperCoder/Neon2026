package quanticscript

import (
	"testing"
)

// TestDeterministicExecution tests that the same program produces identical results
func TestDeterministicExecution(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		runCount int
	}{
		{
			name: "simple arithmetic",
			source: `
				export function entry(): i64 {
					let a: i64 = 100;
					let b: i64 = 50;
					return a + b * 2;
				}
			`,
			runCount: 10,
		},
		{
			name: "complex control flow",
			source: `
				export function entry(): i64 {
					let sum: i64 = 0;
					let i: i64 = 0;
					while (i < 10) {
						if (i % 2 == 0) {
							sum = sum + i;
						}
						i = i + 1;
					}
					return sum;
				}
			`,
			runCount: 10,
		},
		{
			name: "nested loops",
			source: `
				export function entry(): i64 {
					let result: i64 = 0;
					let i: i64 = 0;
					while (i < 5) {
						let j: i64 = 0;
						while (j < 5) {
							result = result + 1;
							j = j + 1;
						}
						i = i + 1;
					}
					return result;
				}
			`,
			runCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compile once
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
			body, err := GetBytecodeBody(bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to get bytecode body: %v", err)
			}

			// Execute multiple times and collect results
			var results []int64
			var budgets []int64

			for i := 0; i < tt.runCount; i++ {
				ctx := NewMockExecutionContext()
				interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
				err = interpreter.Execute()
				if err != nil {
					t.Fatalf("Execution %d failed: %v", i, err)
				}

				if len(interpreter.stack) == 0 {
					t.Fatalf("Execution %d: expected result on stack", i)
				}

				result := interpreter.stack[len(interpreter.stack)-1]
				val, err := result.AsI64()
				if err != nil {
					t.Fatalf("Execution %d: failed to get result: %v", i, err)
				}

				results = append(results, val)
				budgets = append(budgets, interpreter.GetComputeBudget())
			}

			// Verify all results are identical
			firstResult := results[0]
			firstBudget := budgets[0]

			for i := 1; i < len(results); i++ {
				if results[i] != firstResult {
					t.Errorf("Execution %d: result mismatch: got %d, expected %d", i, results[i], firstResult)
				}
				if budgets[i] != firstBudget {
					t.Errorf("Execution %d: budget mismatch: got %d, expected %d", i, budgets[i], firstBudget)
				}
			}
		})
	}
}

// TestDeterministicOverflow tests deterministic behavior with overflow
func TestDeterministicOverflow(t *testing.T) {
	source := `
		export function entry(): i64 {
			let max: i64 = 9223372036854775807;
			return max + 1;
		}
	`

	// Compile
	lexer := NewLexer(source, "test.qs")
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

	// Execute multiple times
	var results []int64
	for i := 0; i < 5; i++ {
		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
		err = interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}

		result := interpreter.stack[len(interpreter.stack)-1]
		val, _ := result.AsI64()
		results = append(results, val)
	}

	// Verify all results are identical (even if overflowed)
	firstResult := results[0]
	for i := 1; i < len(results); i++ {
		if results[i] != firstResult {
			t.Errorf("Execution %d: overflow result mismatch: got %d, expected %d", i, results[i], firstResult)
		}
	}
}

// TestDeterministicDivisionByZero tests deterministic error handling
func TestDeterministicDivisionByZero(t *testing.T) {
	source := `
		export function entry(): i64 {
			let a: i64 = 100;
			let b: i64 = 0;
			return a / b;
		}
	`

	// Compile
	lexer := NewLexer(source, "test.qs")
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

	// Execute multiple times - should always fail with same error
	var errors []string
	for i := 0; i < 5; i++ {
		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
		err = interpreter.Execute()
		if err == nil {
			t.Fatalf("Execution %d: expected division by zero error", i)
		}
		errors = append(errors, err.Error())
	}

	// Verify all errors are identical
	firstError := errors[0]
	for i := 1; i < len(errors); i++ {
		if errors[i] != firstError {
			t.Errorf("Execution %d: error mismatch: got '%s', expected '%s'", i, errors[i], firstError)
		}
	}
}

// TestDeterministicModuloByZero tests deterministic error handling for modulo
func TestDeterministicModuloByZero(t *testing.T) {
	source := `
		export function entry(): i64 {
			let a: i64 = 100;
			let b: i64 = 0;
			return a % b;
		}
	`

	// Compile
	lexer := NewLexer(source, "test.qs")
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

	// Execute multiple times - should always fail with same error
	var errors []string
	for i := 0; i < 5; i++ {
		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
		err = interpreter.Execute()
		if err == nil {
			t.Fatalf("Execution %d: expected modulo by zero error", i)
		}
		errors = append(errors, err.Error())
	}

	// Verify all errors are identical
	firstError := errors[0]
	for i := 1; i < len(errors); i++ {
		if errors[i] != firstError {
			t.Errorf("Execution %d: error mismatch: got '%s', expected '%s'", i, errors[i], firstError)
		}
	}
}

// TestDeterministicBytecodeExecution tests determinism at bytecode level
func TestDeterministicBytecodeExecution(t *testing.T) {
	// Test with raw bytecode to ensure interpreter determinism
	bytecode := []byte{
		byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0,
		byte(OpPush), byte(TypeI64), 20, 0, 0, 0, 0, 0, 0, 0,
		byte(OpAdd),
		byte(OpPush), byte(TypeI64), 3, 0, 0, 0, 0, 0, 0, 0,
		byte(OpMul),
		byte(OpRet),
	}

	var results []int64
	var budgets []int64

	for i := 0; i < 10; i++ {
		ctx := NewMockExecutionContext()
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}

		result := interpreter.stack[len(interpreter.stack)-1]
		val, _ := result.AsI64()
		results = append(results, val)
		budgets = append(budgets, interpreter.GetComputeBudget())
	}

	// Verify all results and budgets are identical
	firstResult := results[0]
	firstBudget := budgets[0]

	for i := 1; i < len(results); i++ {
		if results[i] != firstResult {
			t.Errorf("Execution %d: result mismatch: got %d, expected %d", i, results[i], firstResult)
		}
		if budgets[i] != firstBudget {
			t.Errorf("Execution %d: budget mismatch: got %d, expected %d", i, budgets[i], firstBudget)
		}
	}
}

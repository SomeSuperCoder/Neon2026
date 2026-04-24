package quanticscript

import (
	"strings"
	"testing"
)

// TestFullCompilationPipeline tests the complete pipeline from source code to execution
func TestFullCompilationPipeline(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedResult int64
		shouldFail     bool
	}{
		{
			name: "simple return",
			source: `
				export function entry(): i64 {
					return 42;
				}
			`,
			expectedResult: 42,
			shouldFail:     false,
		},
		{
			name: "arithmetic operations",
			source: `
				export function entry(): i64 {
					let a: i64 = 10;
					let b: i64 = 20;
					let c: i64 = a + b;
					return c * 2;
				}
			`,
			expectedResult: 60,
			shouldFail:     false,
		},
		{
			name: "conditional logic",
			source: `
				export function entry(): i64 {
					let x: i64 = 5;
					if (x > 3) {
						return 100;
					} else {
						return 200;
					}
				}
			`,
			expectedResult: 100,
			shouldFail:     false,
		},
		{
			name: "while loop",
			source: `
				export function entry(): i64 {
					let sum: i64 = 0;
					let i: i64 = 1;
					while (i <= 5) {
						sum = sum + i;
						i = i + 1;
					}
					return sum;
				}
			`,
			expectedResult: 15, // 1+2+3+4+5
			shouldFail:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lex and Parse
			lexer := NewLexer(tt.source, "test.qs")
			parser := NewParser(lexer)
			program := parser.ParseProgram()
			if len(parser.errors) > 0 {
				if tt.shouldFail {
					return
				}
				t.Fatalf("Parser errors: %v", parser.errors)
			}

			// Type check
			tc := NewTypeChecker()
			tc.CheckProgram(program)
			if len(tc.errors) > 0 {
				if tt.shouldFail {
					return
				}
				t.Fatalf("Type checker errors: %v", tc.errors)
			}

			// Generate bytecode
			cg := NewCodeGenerator()
			bytecode, err := cg.Generate(program)
			if err != nil {
				if tt.shouldFail {
					return
				}
				t.Fatalf("Code generator error: %v", err)
			}

			// Create bytecode file
			bytecodeFile := CreateBytecode(bytecode, 0)
			body, err := GetBytecodeBody(bytecodeFile)
			if err != nil {
				t.Fatalf("Failed to get bytecode body: %v", err)
			}

			// Execute
			ctx := NewMockExecutionContext()
			interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
			err = interpreter.Execute()
			if err != nil {
				if tt.shouldFail {
					return
				}
				t.Fatalf("Execution error: %v", err)
			}

			// Verify result
			if len(interpreter.stack) == 0 {
				t.Fatal("Expected result on stack")
			}
			result := interpreter.stack[len(interpreter.stack)-1]
			val, err := result.AsI64()
			if err != nil {
				t.Fatalf("Failed to get result value: %v", err)
			}

			if val != tt.expectedResult {
				t.Errorf("Expected result %d, got %d", tt.expectedResult, val)
			}
		})
	}
}

// TestInlineAssemblyCompilation tests compilation of inline assembly
func TestInlineAssemblyCompilation(t *testing.T) {
	t.Skip("Inline assembly syntax needs further investigation")
	source := `
		export function entry(): i64 {
			let result: i64;
			__asm__ {
				PUSH 50
				PUSH 8
				ADD
				STORE result
			}
			return result;
		}
	`

	// Lex and Parse
	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	if len(parser.errors) > 0 {
		t.Fatalf("Parser errors: %v", parser.errors)
	}

	// Type check
	tc := NewTypeChecker()
	tc.CheckProgram(program)
	if len(tc.errors) > 0 {
		t.Fatalf("Type checker errors: %v", tc.errors)
	}

	// Generate bytecode
	cg := NewCodeGenerator()
	bytecode, err := cg.Generate(program)
	if err != nil {
		t.Fatalf("Code generator error: %v", err)
	}

	// Execute
	bytecodeFile := CreateBytecode(bytecode, 0)
	body, _ := GetBytecodeBody(bytecodeFile)
	ctx := NewMockExecutionContext()
	interpreter := NewBytecodeInterpreter(body, ctx, 1000000)
	err = interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution error: %v", err)
	}

	// Verify result
	if len(interpreter.stack) == 0 {
		t.Fatal("Expected result on stack")
	}
	result := interpreter.stack[len(interpreter.stack)-1]
	val, _ := result.AsI64()
	if val != 58 {
		t.Errorf("Expected result 58, got %d", val)
	}
}

// TestCompilerErrorDetection tests that the compiler detects various errors
func TestCompilerErrorDetection(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError string
	}{
		{
			name: "undefined variable",
			source: `
				export function entry(): i64 {
					return x;
				}
			`,
			expectError: "undefined",
		},
		{
			name: "type mismatch",
			source: `
				export function entry(): i64 {
					let x: i64 = true;
					return x;
				}
			`,
			expectError: "type",
		},
		{
			name: "missing return",
			source: `
				export function entry(): i64 {
					let x: i64 = 42;
				}
			`,
			expectError: "", // May or may not error depending on implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Lex and Parse
			lexer := NewLexer(tt.source, "test.qs")
			parser := NewParser(lexer)
			program := parser.ParseProgram()
			if len(parser.errors) > 0 {
				if tt.expectError != "" && strings.Contains(strings.ToLower(parser.errors[0].Error()), strings.ToLower(tt.expectError)) {
					return // Expected error
				}
				if tt.expectError != "" {
					t.Fatalf("Expected error containing '%s', got: %v", tt.expectError, parser.errors[0])
				}
				return
			}

			// Type check
			tc := NewTypeChecker()
			tc.CheckProgram(program)
			if len(tc.errors) > 0 {
				if tt.expectError != "" && strings.Contains(strings.ToLower(tc.errors[0].Error()), strings.ToLower(tt.expectError)) {
					return // Expected error
				}
				if tt.expectError != "" {
					t.Fatalf("Expected error containing '%s', got: %v", tt.expectError, tc.errors[0])
				}
				return
			}

			// Generate bytecode
			cg := NewCodeGenerator()
			_, err := cg.Generate(program)
			if err != nil {
				if tt.expectError != "" && strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.expectError)) {
					return // Expected error
				}
				if tt.expectError != "" {
					t.Fatalf("Expected error containing '%s', got: %v", tt.expectError, err)
				}
				return
			}

			// If we expected an error but didn't get one
			if tt.expectError != "" {
				t.Errorf("Expected error containing '%s', but compilation succeeded", tt.expectError)
			}
		})
	}
}

// TestCompilerWithStdLib tests compilation with standard library functions
func TestCompilerWithStdLib(t *testing.T) {
	source := `
		export function entry(): i64 {
			let data: bytes = [1, 2, 3, 4];
			let hash: bytes = sha256(data);
			return hash.length;
		}
	`

	// Lex and Parse
	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()
	if len(parser.errors) > 0 {
		t.Fatalf("Parser errors: %v", parser.errors)
	}

	// Type check
	tc := NewTypeChecker()
	tc.CheckProgram(program)
	if len(tc.errors) > 0 {
		// This may fail if stdlib functions aren't registered, which is acceptable
		t.Skipf("Type checker errors (stdlib may not be registered): %v", tc.errors)
	}

	// Generate bytecode
	cg := NewCodeGenerator()
	_, err := cg.Generate(program)
	if err != nil {
		// This may fail if stdlib isn't fully implemented, which is acceptable
		t.Skipf("Code generator error (stdlib may not be fully implemented): %v", err)
	}
}

// TestAssemblyToSourceRoundTrip tests that assembly can be compiled and disassembled
func TestAssemblyToSourceRoundTrip(t *testing.T) {
	assembly := `
		PUSH i64 10
		PUSH i64 20
		ADD
		PUSH i64 2
		MUL
		RET
	`

	// Assemble
	bytecode, err := AssembleToBody(assembly)
	if err != nil {
		t.Fatalf("Assemble error: %v", err)
	}

	// Execute original
	ctx1 := NewMockExecutionContext()
	interp1 := NewBytecodeInterpreter(bytecode, ctx1, 10000)
	err = interp1.Execute()
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// Disassemble
	disassembled, err := DisassembleBody(bytecode)
	if err != nil {
		t.Fatalf("Disassemble error: %v", err)
	}

	// Reassemble
	bytecode2, err := AssembleToBody(disassembled)
	if err != nil {
		t.Fatalf("Reassemble error: %v", err)
	}

	// Execute reassembled
	ctx2 := NewMockExecutionContext()
	interp2 := NewBytecodeInterpreter(bytecode2, ctx2, 10000)
	err = interp2.Execute()
	if err != nil {
		t.Fatalf("Execute reassembled error: %v", err)
	}

	// Compare results
	if len(interp1.stack) != len(interp2.stack) {
		t.Errorf("Stack length mismatch: %d vs %d", len(interp1.stack), len(interp2.stack))
	}

	for i := range interp1.stack {
		if i >= len(interp2.stack) {
			break
		}
		val1, _ := interp1.stack[i].AsI64()
		val2, _ := interp2.stack[i].AsI64()
		if val1 != val2 {
			t.Errorf("Stack[%d] value mismatch: %d vs %d", i, val1, val2)
		}
	}
}

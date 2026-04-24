package quanticscript

import (
	"strings"
	"testing"
)

func TestTypeChecker_BasicTypes(t *testing.T) {
	source := `
		function test(): i64 {
			let x: i64 = 42;
			let y: i64 = 10;
			return x + y;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", parser.Errors())
	}

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) > 0 {
		t.Errorf("Expected no type errors, got: %v", tc.Errors())
	}
}

func TestTypeChecker_TypeMismatch(t *testing.T) {
	source := `
		function test(): i64 {
			let x: i64 = 42;
			let y: bool = true;
			return x + y;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected type error for adding i64 and bool")
	}
}

func TestTypeChecker_UndefinedVariable(t *testing.T) {
	source := `
		function test(): i64 {
			return x;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for undefined variable")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "undefined variable") {
		t.Errorf("Expected 'undefined variable' error, got: %s", errMsg)
	}
}

func TestTypeChecker_ConstAssignment(t *testing.T) {
	source := `
		function test(): void {
			const x: i64 = 42;
			x = 10;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for assigning to const variable")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "const") {
		t.Errorf("Expected 'const' error, got: %s", errMsg)
	}
}

func TestTypeChecker_FunctionCall(t *testing.T) {
	source := `
		function add(a: i64, b: i64): i64 {
			return a + b;
		}

		function test(): i64 {
			return add(10, 20);
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) > 0 {
		t.Errorf("Expected no type errors, got: %v", tc.Errors())
	}
}

func TestTypeChecker_FunctionCallWrongArgCount(t *testing.T) {
	source := `
		function add(a: i64, b: i64): i64 {
			return a + b;
		}

		function test(): i64 {
			return add(10);
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for wrong argument count")
	}
}

func TestTypeChecker_ArrayType(t *testing.T) {
	source := `
		function test(): i64 {
			let arr: i64[] = [1, 2, 3];
			return arr[0];
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) > 0 {
		t.Errorf("Expected no type errors, got: %v", tc.Errors())
	}
}

func TestTypeChecker_ArrayTypeMismatch(t *testing.T) {
	source := `
		function test(): void {
			let arr: i64[] = [1, 2, true];
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for array element type mismatch")
	}
}

func TestTypeChecker_NonDeterministicFunction(t *testing.T) {
	source := `
		function test(): i64 {
			return random();
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for non-deterministic function call")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "non-deterministic") {
		t.Errorf("Expected 'non-deterministic' error, got: %s", errMsg)
	}
}

func TestTypeChecker_FloatingPointType(t *testing.T) {
	source := `
		function test(): void {
			let x: f32;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for floating-point type")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "non-deterministic") {
		t.Errorf("Expected 'non-deterministic' error, got: %s", errMsg)
	}
}

func TestTypeChecker_AssemblyUndefinedVariable(t *testing.T) {
	source := `
		function test(): void {
			__asm__ {
				LOAD undefined_var
			}
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		t.Logf("Parser errors (expected): %v", parser.Errors())
	}

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for undefined variable in assembly")
		return
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "undefined variable") {
		t.Errorf("Expected 'undefined variable' error, got: %s", errMsg)
	}
}

func TestTypeChecker_AssemblyTypeSafety(t *testing.T) {
	source := `
		function test(): void {
			let x: bool = true;
			__asm__ {
				ADD x
			}
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		t.Logf("Parser errors (expected): %v", parser.Errors())
	}

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for using bool in arithmetic operation")
		return
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "numeric type") {
		t.Errorf("Expected 'numeric type' error, got: %s", errMsg)
	}
}

func TestTypeChecker_AssemblyNonDeterministic(t *testing.T) {
	source := `
		function test(): void {
			__asm__ {
				RANDOM
			}
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		t.Logf("Parser errors (expected): %v", parser.Errors())
	}

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for non-deterministic assembly instruction")
		return
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "non-deterministic") {
		t.Errorf("Expected 'non-deterministic' error, got: %s", errMsg)
	}
}

func TestTypeChecker_AssemblyUnsafeOperation(t *testing.T) {
	source := `
		function test(): void {
			__asm__ {
				SYSCALL
			}
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		t.Logf("Parser errors (expected): %v", parser.Errors())
	}

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for unsafe assembly operation")
		return
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "not allowed") {
		t.Errorf("Expected 'not allowed' error, got: %s", errMsg)
	}
}

func TestTypeChecker_IfConditionType(t *testing.T) {
	source := `
		function test(): void {
			if (42) {
				let x: i64 = 1;
			}
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for non-boolean if condition")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "boolean") {
		t.Errorf("Expected 'boolean' error, got: %s", errMsg)
	}
}

func TestTypeChecker_ReturnTypeMismatch(t *testing.T) {
	source := `
		function test(): i64 {
			return true;
		}
	`

	lexer := NewLexer(source, "test.qs")
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	tc := NewTypeChecker()
	tc.CheckProgram(program)

	if len(tc.Errors()) == 0 {
		t.Error("Expected error for return type mismatch")
	}

	errMsg := tc.Errors()[0].Error()
	if !strings.Contains(errMsg, "type mismatch") {
		t.Errorf("Expected 'type mismatch' error, got: %s", errMsg)
	}
}

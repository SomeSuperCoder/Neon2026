package quanticscript

import (
	"testing"
)

// TestComparisonOperationsExtended tests EQ, LT, GT, LTE, GTE for numbers and string equality
func TestComparisonOperationsExtended(t *testing.T) {
	runCmp := func(t *testing.T, a, b int64, op Opcode) bool {
		t.Helper()
		var bc []byte
		bc = append(bc, buildPushI64(a)...)
		bc = append(bc, buildPushI64(b)...)
		bc = append(bc, byte(op))
		bc = append(bc, byte(OpRet))
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("AsBool failed: %v", err)
		}
		return result
	}

	t.Run("EQ true", func(t *testing.T) {
		if !runCmp(t, 42, 42, OpEq) {
			t.Error("Expected 42 == 42 to be true")
		}
	})

	t.Run("EQ false", func(t *testing.T) {
		if runCmp(t, 42, 43, OpEq) {
			t.Error("Expected 42 == 43 to be false")
		}
	})

	t.Run("LT true", func(t *testing.T) {
		if !runCmp(t, 10, 20, OpLt) {
			t.Error("Expected 10 < 20 to be true")
		}
	})

	t.Run("LT false", func(t *testing.T) {
		if runCmp(t, 20, 10, OpLt) {
			t.Error("Expected 20 < 10 to be false")
		}
	})

	t.Run("LT equal values", func(t *testing.T) {
		if runCmp(t, 10, 10, OpLt) {
			t.Error("Expected 10 < 10 to be false")
		}
	})

	t.Run("GT true", func(t *testing.T) {
		if !runCmp(t, 20, 10, OpGt) {
			t.Error("Expected 20 > 10 to be true")
		}
	})

	t.Run("GT false", func(t *testing.T) {
		if runCmp(t, 10, 20, OpGt) {
			t.Error("Expected 10 > 20 to be false")
		}
	})

	t.Run("GT equal values", func(t *testing.T) {
		if runCmp(t, 10, 10, OpGt) {
			t.Error("Expected 10 > 10 to be false")
		}
	})

	t.Run("LTE less than", func(t *testing.T) {
		if !runCmp(t, 10, 20, OpLte) {
			t.Error("Expected 10 <= 20 to be true")
		}
	})

	t.Run("LTE equal", func(t *testing.T) {
		if !runCmp(t, 10, 10, OpLte) {
			t.Error("Expected 10 <= 10 to be true")
		}
	})

	t.Run("LTE greater", func(t *testing.T) {
		if runCmp(t, 20, 10, OpLte) {
			t.Error("Expected 20 <= 10 to be false")
		}
	})

	t.Run("GTE greater than", func(t *testing.T) {
		if !runCmp(t, 20, 10, OpGte) {
			t.Error("Expected 20 >= 10 to be true")
		}
	})

	t.Run("GTE equal", func(t *testing.T) {
		if !runCmp(t, 10, 10, OpGte) {
			t.Error("Expected 10 >= 10 to be true")
		}
	})

	t.Run("GTE less than", func(t *testing.T) {
		if runCmp(t, 10, 20, OpGte) {
			t.Error("Expected 10 >= 20 to be false")
		}
	})

	// String equality: different types return false (not an error)
	t.Run("string vs string different types returns false", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushString("hello")...)
		bc = append(bc, buildPushI64(42)...)
		bc = append(bc, byte(OpEq))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if result {
			t.Error("Expected string == i64 to be false")
		}
	})
}

// TestLogicalOperationsExtended tests AND, OR, NOT and compound boolean expressions
func TestLogicalOperationsExtended(t *testing.T) {
	runLogic := func(t *testing.T, a, b bool, op Opcode) bool {
		t.Helper()
		var bc []byte
		bc = append(bc, buildPushBool(a)...)
		bc = append(bc, buildPushBool(b)...)
		bc = append(bc, byte(op))
		bc = append(bc, byte(OpRet))
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, err := interp.stack[0].AsBool()
		if err != nil {
			t.Fatalf("AsBool failed: %v", err)
		}
		return result
	}

	// AND truth table
	t.Run("AND true true", func(t *testing.T) {
		if !runLogic(t, true, true, OpAnd) {
			t.Error("Expected true AND true = true")
		}
	})
	t.Run("AND true false", func(t *testing.T) {
		if runLogic(t, true, false, OpAnd) {
			t.Error("Expected true AND false = false")
		}
	})
	t.Run("AND false true", func(t *testing.T) {
		if runLogic(t, false, true, OpAnd) {
			t.Error("Expected false AND true = false")
		}
	})
	t.Run("AND false false", func(t *testing.T) {
		if runLogic(t, false, false, OpAnd) {
			t.Error("Expected false AND false = false")
		}
	})

	// OR truth table
	t.Run("OR true true", func(t *testing.T) {
		if !runLogic(t, true, true, OpOr) {
			t.Error("Expected true OR true = true")
		}
	})
	t.Run("OR true false", func(t *testing.T) {
		if !runLogic(t, true, false, OpOr) {
			t.Error("Expected true OR false = true")
		}
	})
	t.Run("OR false true", func(t *testing.T) {
		if !runLogic(t, false, true, OpOr) {
			t.Error("Expected false OR true = true")
		}
	})
	t.Run("OR false false", func(t *testing.T) {
		if runLogic(t, false, false, OpOr) {
			t.Error("Expected false OR false = false")
		}
	})

	// NOT
	t.Run("NOT true", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushBool(true)...)
		bc = append(bc, byte(OpNot))
		bc = append(bc, byte(OpRet))
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if result {
			t.Error("Expected NOT true = false")
		}
	})

	t.Run("NOT false", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushBool(false)...)
		bc = append(bc, byte(OpNot))
		bc = append(bc, byte(OpRet))
		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected NOT false = true")
		}
	})

	// Compound: (true AND false) OR true
	t.Run("compound (true AND false) OR true", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushBool(true)...)
		bc = append(bc, buildPushBool(false)...)
		bc = append(bc, byte(OpAnd))
		bc = append(bc, buildPushBool(true)...)
		bc = append(bc, byte(OpOr))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected (true AND false) OR true = true")
		}
	})

	// Compound: NOT (false OR false)
	t.Run("compound NOT (false OR false)", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushBool(false)...)
		bc = append(bc, buildPushBool(false)...)
		bc = append(bc, byte(OpOr))
		bc = append(bc, byte(OpNot))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected NOT (false OR false) = true")
		}
	})
}

// TestOperatorPrecedence tests combined arithmetic and comparison expressions
func TestOperatorPrecedence(t *testing.T) {
	// (3 + 4) * 2 == 14
	t.Run("arithmetic then comparison", func(t *testing.T) {
		var bc []byte
		// compute (3 + 4) * 2
		bc = append(bc, buildPushI64(3)...)
		bc = append(bc, buildPushI64(4)...)
		bc = append(bc, byte(OpAdd))
		bc = append(bc, buildPushI64(2)...)
		bc = append(bc, byte(OpMul))
		// compare result == 14
		bc = append(bc, buildPushI64(14)...)
		bc = append(bc, byte(OpEq))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected (3+4)*2 == 14 to be true")
		}
	})

	// (10 - 3) > (2 * 3)  =>  7 > 6  => true
	t.Run("subtraction vs multiplication comparison", func(t *testing.T) {
		var bc []byte
		// 10 - 3
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(3)...)
		bc = append(bc, byte(OpSub))
		// 2 * 3
		bc = append(bc, buildPushI64(2)...)
		bc = append(bc, buildPushI64(3)...)
		bc = append(bc, byte(OpMul))
		// (10-3) > (2*3)
		bc = append(bc, byte(OpGt))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected (10-3) > (2*3) to be true")
		}
	})

	// (100 / 5) <= (10 * 2)  =>  20 <= 20  => true
	t.Run("division vs multiplication LTE", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(100)...)
		bc = append(bc, buildPushI64(5)...)
		bc = append(bc, byte(OpDiv))
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(2)...)
		bc = append(bc, byte(OpMul))
		bc = append(bc, byte(OpLte))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected (100/5) <= (10*2) to be true")
		}
	})

	// (7 % 3 == 1) AND (10 > 5)  =>  true AND true  => true
	t.Run("mod comparison combined with AND", func(t *testing.T) {
		var bc []byte
		// 7 % 3 == 1
		bc = append(bc, buildPushI64(7)...)
		bc = append(bc, buildPushI64(3)...)
		bc = append(bc, byte(OpMod))
		bc = append(bc, buildPushI64(1)...)
		bc = append(bc, byte(OpEq))
		// 10 > 5
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(5)...)
		bc = append(bc, byte(OpGt))
		// AND
		bc = append(bc, byte(OpAnd))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsBool()
		if !result {
			t.Error("Expected (7%3==1) AND (10>5) to be true")
		}
	})
}

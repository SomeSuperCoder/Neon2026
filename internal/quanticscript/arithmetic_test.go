package quanticscript

import (
	"testing"
)

// TestArithmeticOperations tests ADD, SUB, MUL, DIV, MOD with i64 and u64
func TestArithmeticOperations(t *testing.T) {
	t.Run("ADD i64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(32)...)
		bc = append(bc, byte(OpAdd))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, err := interp.stack[0].AsI64()
		if err != nil {
			t.Fatalf("AsI64 failed: %v", err)
		}
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("SUB i64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(100)...)
		bc = append(bc, buildPushI64(58)...)
		bc = append(bc, byte(OpSub))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("MUL i64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(6)...)
		bc = append(bc, buildPushI64(7)...)
		bc = append(bc, byte(OpMul))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("DIV i64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(84)...)
		bc = append(bc, buildPushI64(2)...)
		bc = append(bc, byte(OpDiv))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("MOD i64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(100)...)
		bc = append(bc, buildPushI64(58)...)
		bc = append(bc, byte(OpMod))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsI64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("ADD u64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushU64(20)...)
		bc = append(bc, buildPushU64(22)...)
		bc = append(bc, byte(OpAdd))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsU64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("SUB u64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushU64(50)...)
		bc = append(bc, buildPushU64(8)...)
		bc = append(bc, byte(OpSub))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsU64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("MUL u64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushU64(7)...)
		bc = append(bc, buildPushU64(6)...)
		bc = append(bc, byte(OpMul))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsU64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("DIV u64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushU64(126)...)
		bc = append(bc, buildPushU64(3)...)
		bc = append(bc, byte(OpDiv))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsU64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("MOD u64", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushU64(142)...)
		bc = append(bc, buildPushU64(100)...)
		bc = append(bc, byte(OpMod))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsU64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("SUB i64 negative result", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(50)...)
		bc = append(bc, byte(OpSub))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		if err := interp.Execute(); err != nil {
			t.Fatalf("Execution failed: %v", err)
		}
		result, _ := interp.stack[0].AsI64()
		if result != -40 {
			t.Errorf("Expected -40, got %d", result)
		}
	})

	t.Run("DIV by zero error", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(0)...)
		bc = append(bc, byte(OpDiv))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected division by zero error, got nil")
		}
	})

	t.Run("MOD by zero error", func(t *testing.T) {
		var bc []byte
		bc = append(bc, buildPushI64(10)...)
		bc = append(bc, buildPushI64(0)...)
		bc = append(bc, byte(OpMod))
		bc = append(bc, byte(OpRet))

		ctx := NewMockExecutionContext()
		interp := NewBytecodeInterpreter(bc, ctx, 1000)
		err := interp.Execute()
		if err == nil {
			t.Fatal("Expected modulo by zero error, got nil")
		}
	})
}

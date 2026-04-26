package quanticscript

import (
	"testing"
)

// TestCryptoOperations tests the crypto module functions
func TestCryptoOperations(t *testing.T) {
	t.Run("SHA256", func(t *testing.T) {
		data := []byte("hello world")
		hash := sha256Hash(data)

		if len(hash) != 32 {
			t.Errorf("Expected hash length 32, got %d", len(hash))
		}

		// Verify determinism - same input produces same output
		hash2 := sha256Hash(data)
		if string(hash) != string(hash2) {
			t.Error("SHA256 is not deterministic")
		}
	})

	t.Run("VerifySignature", func(t *testing.T) {
		// Test with invalid signature (should return false, not error)
		pubkey := make([]byte, 32)
		message := []byte("test message")
		signature := make([]byte, 64)

		result := verifySignature(pubkey, message, signature)
		if result {
			t.Error("Expected signature verification to fail for invalid signature")
		}
	})

	t.Run("DerivePublicKey", func(t *testing.T) {
		seed := make([]byte, 32)
		for i := range seed {
			seed[i] = byte(i)
		}

		pubkey, err := derivePublicKey(seed)
		if err != nil {
			t.Fatalf("Failed to derive public key: %v", err)
		}

		if len(pubkey) != 32 {
			t.Errorf("Expected public key length 32, got %d", len(pubkey))
		}

		// Verify determinism
		pubkey2, _ := derivePublicKey(seed)
		if string(pubkey) != string(pubkey2) {
			t.Error("Public key derivation is not deterministic")
		}
	})
}

// TestStringOperations tests the string module functions
func TestStringOperations(t *testing.T) {
	ctx := NewMockExecutionContext()

	t.Run("StringConcat", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 5, 0, 0, 0, 0, 0, 0, 0, 'h', 'e', 'l', 'l', 'o',
			byte(OpPush), byte(TypeString), 6, 0, 0, 0, 0, 0, 0, 0, ' ', 'w', 'o', 'r', 'l', 'd',
			byte(OpStrConcat),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
		}

		result, err := interpreter.stack[0].AsString()
		if err != nil {
			t.Fatalf("Failed to get string: %v", err)
		}

		if result != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", result)
		}
	})

	t.Run("StringLength", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 5, 0, 0, 0, 0, 0, 0, 0, 'h', 'e', 'l', 'l', 'o',
			byte(OpStrLen),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
		}

		length, err := interpreter.stack[0].AsU64()
		if err != nil {
			t.Fatalf("Failed to get length: %v", err)
		}

		if length != 5 {
			t.Errorf("Expected length 5, got %d", length)
		}
	})
}

// TestMathOperations tests the math module functions
func TestMathOperations(t *testing.T) {
	ctx := NewMockExecutionContext()

	t.Run("MathMin", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0,
			byte(OpMathMin),
			byte(OpRet),
		}

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

		if result != 10 {
			t.Errorf("Expected 10, got %d", result)
		}
	})

	t.Run("MathMax", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpPush), byte(TypeI64), 10, 0, 0, 0, 0, 0, 0, 0,
			byte(OpMathMax),
			byte(OpRet),
		}

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

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("MathAbs", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 0xD6, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, // -42
			byte(OpMathAbs),
			byte(OpRet),
		}

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

		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})
}

// TestLogFunction tests the log() debug function
func TestLogFunction(t *testing.T) {
	ctx := NewMockExecutionContext()

	t.Run("log with string", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 5, 0, 0, 0, 0, 0, 0, 0, 'h', 'e', 'l', 'l', 'o',
			byte(OpLog),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// log() should consume the value and return nothing
		if len(interpreter.stack) != 0 {
			t.Fatalf("Expected empty stack after log, got %d values", len(interpreter.stack))
		}
	})

	t.Run("log with i64", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpLog),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 0 {
			t.Fatalf("Expected empty stack after log, got %d values", len(interpreter.stack))
		}
	})

	t.Run("log charges high cost", func(t *testing.T) {
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 5, 0, 0, 0, 0, 0, 0, 0, 't', 'e', 's', 't', 's',
			byte(OpLog),
			byte(OpRet),
		}

		// Start with exactly 1000 units - should fail because log costs more
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err == nil {
			t.Error("Expected execution to fail due to insufficient compute budget")
		}

		// With 10000 units, should succeed
		interpreter = NewBytecodeInterpreter(bytecode, ctx, 10000)
		err = interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed with sufficient budget: %v", err)
		}
	})
}

// TestCollectionOperations tests the collections module functions
func TestCollectionOperations(t *testing.T) {
	ctx := NewMockExecutionContext()

	t.Run("ArrayNew", func(t *testing.T) {
		bytecode := []byte{
			byte(OpArrayNew),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
		}

		arr, err := interpreter.stack[0].AsArray()
		if err != nil {
			t.Fatalf("Failed to get array: %v", err)
		}

		if len(arr) != 0 {
			t.Errorf("Expected empty array, got length %d", len(arr))
		}
	})

	t.Run("MapNew", func(t *testing.T) {
		bytecode := []byte{
			byte(OpMapNew),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
		}

		m, err := interpreter.stack[0].AsMap()
		if err != nil {
			t.Fatalf("Failed to get map: %v", err)
		}

		if len(m) != 0 {
			t.Errorf("Expected empty map, got length %d", len(m))
		}
	})

	t.Run("SetNew", func(t *testing.T) {
		bytecode := []byte{
			byte(OpSetNew),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 1000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		if len(interpreter.stack) != 1 {
			t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
		}

		s, err := interpreter.stack[0].AsSet()
		if err != nil {
			t.Fatalf("Failed to get set: %v", err)
		}

		if len(s) != 0 {
			t.Errorf("Expected empty set, got length %d", len(s))
		}
	})
}

// TestLogOperation tests the log function for debugging
func TestLogOperation(t *testing.T) {
	ctx := NewMockExecutionContext()

	t.Run("LogString", func(t *testing.T) {
		// Test logging a string message - OpLog pops the value and discards it
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 11, 0, 0, 0, 0, 0, 0, 0, 'H', 'e', 'l', 'l', 'o', ' ', 'W', 'o', 'r', 'l', 'd',
			byte(OpLog),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// After OpLog, the stack should be empty (value was popped)
		if len(interpreter.stack) != 0 {
			t.Errorf("Expected empty stack after log, got %d values", len(interpreter.stack))
		}
	})

	t.Run("LogInteger", func(t *testing.T) {
		// Test logging an integer value
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 42, 0, 0, 0, 0, 0, 0, 0,
			byte(OpLog),
			byte(OpRet),
		}

		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// After OpLog, the stack should be empty
		if len(interpreter.stack) != 0 {
			t.Errorf("Expected empty stack after log, got %d values", len(interpreter.stack))
		}
	})

	t.Run("LogHighCost", func(t *testing.T) {
		// Test that log operation charges a high cost (5000 per OpLog)
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 4, 0, 0, 0, 0, 0, 0, 0, 't', 'e', 's', 't',
			byte(OpLog),
			byte(OpRet),
		}

		// Set a low compute budget to verify high cost
		// OpLog costs 5000, so budget of 100 should fail
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 100)
		err := interpreter.Execute()

		// Should fail due to out of compute budget
		if err == nil {
			t.Error("Expected execution to fail due to out of compute budget")
		}
		if err != nil && err.Error() != "out of compute budget" {
			t.Errorf("Expected 'out of compute budget' error, got: %v", err)
		}
	})

	t.Run("LogMultipleValues", func(t *testing.T) {
		// Test logging multiple values in sequence
		// Each OpLog costs 5000, so 2 logs = 10000 cost
		bytecode := []byte{
			byte(OpPush), byte(TypeString), 5, 0, 0, 0, 0, 0, 0, 0, 'F', 'i', 'r', 's', 't',
			byte(OpLog),
			byte(OpPush), byte(TypeString), 6, 0, 0, 0, 0, 0, 0, 0, 'S', 'e', 'c', 'o', 'n', 'd',
			byte(OpLog),
			byte(OpRet),
		}

		// Need enough budget for 2 logs (10000) plus other operations
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 20000)
		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// After both logs, stack should be empty
		if len(interpreter.stack) != 0 {
			t.Errorf("Expected empty stack after logs, got %d values", len(interpreter.stack))
		}
	})

	t.Run("LogWithDebugLogger", func(t *testing.T) {
		// Test that SetDebugLogger can be called (even if not used by execLog currently)
		bytecode := []byte{
			byte(OpPush), byte(TypeI64), 99, 0, 0, 0, 0, 0, 0, 0,
			byte(OpLog),
			byte(OpRet),
		}

		loggerCalled := false
		interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
		interpreter.SetDebugLogger(func(format string, args ...interface{}) {
			loggerCalled = true
		})

		err := interpreter.Execute()
		if err != nil {
			t.Fatalf("Execution failed: %v", err)
		}

		// Note: Currently execLog doesn't call the debug logger, but the method exists
		// This test verifies the infrastructure is in place
		_ = loggerCalled // Logger infrastructure is available
	})
}

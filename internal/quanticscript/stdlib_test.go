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

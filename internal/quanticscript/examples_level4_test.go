package quanticscript

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
)

// TestLevel4BalanceOps tests the balance operations example
func TestLevel4BalanceOps(t *testing.T) {
	// Read the compiled bytecode
	bytecodeFile, err := os.ReadFile("../../examples/10_balance_ops.qsb")
	if err != nil {
		t.Fatalf("Failed to read bytecode: %v", err)
	}

	// Strip the bytecode header
	bytecode, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to parse bytecode: %v", err)
	}

	// Create mock context with a file that has balance
	ctx := NewMockExecutionContext()
	ctx.programID = genesis.SystemProgramID // Set to system program so updateBalance works
	fileID := filestore.FileID{}
	ctx.files[fileID] = &filestore.File{
		ID:      fileID,
		Balance: 1000,
		Data:    []byte{},
	}

	// Execute the bytecode
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
	err = interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check result - should be original balance (1000) + 100 = 1100
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	if result != 1100 {
		t.Errorf("Expected result 1100, got %d", result)
	}
}

// TestLevel4FileOps tests the file operations example
func TestLevel4FileOps(t *testing.T) {
	// Read the compiled bytecode
	bytecodeFile, err := os.ReadFile("../../examples/11_file_ops.qsb")
	if err != nil {
		t.Fatalf("Failed to read bytecode: %v", err)
	}

	// Strip the bytecode header
	bytecode, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to parse bytecode: %v", err)
	}

	// Create mock context with a file that has some data
	ctx := NewMockExecutionContext()
	fileID := filestore.FileID{}
	testData := []byte("Hello, World!")
	ctx.files[fileID] = &filestore.File{
		ID:      fileID,
		Balance: 0,
		Data:    testData,
	}

	// Execute the bytecode
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
	err = interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check result - should be length of test data
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	expectedLen := int64(len(testData))
	if result != expectedLen {
		t.Errorf("Expected result %d, got %d", expectedLen, result)
	}
}

// TestLevel4CryptoOps tests the cryptographic operations example
func TestLevel4CryptoOps(t *testing.T) {
	// Read the compiled bytecode
	bytecodeFile, err := os.ReadFile("../../examples/12_crypto_ops.qsb")
	if err != nil {
		t.Fatalf("Failed to read bytecode: %v", err)
	}

	// Strip the bytecode header
	bytecode, err := GetBytecodeBody(bytecodeFile)
	if err != nil {
		t.Fatalf("Failed to parse bytecode: %v", err)
	}

	// Create mock context with instruction data
	ctx := NewMockExecutionContext()
	ctx.instrData = []byte("test data for hashing")

	// Execute the bytecode
	interpreter := NewBytecodeInterpreter(bytecode, ctx, 10000)
	err = interpreter.Execute()
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Check result - should be 32 (SHA-256 hash length)
	if len(interpreter.stack) != 1 {
		t.Fatalf("Expected 1 value on stack, got %d", len(interpreter.stack))
	}

	result, err := interpreter.stack[0].AsI64()
	if err != nil {
		t.Fatalf("Failed to get result: %v", err)
	}

	if result != 32 {
		t.Errorf("Expected result 32 (SHA-256 hash length), got %d", result)
	}
}

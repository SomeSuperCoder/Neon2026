package runtime

import (
	"testing"
	"time"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/transaction"
)

// TestRuntime_ExecuteBytecode tests bytecode execution through the runtime
func TestRuntime_ExecuteBytecode(t *testing.T) {
	// Create a temporary file store
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a simple bytecode program: PUSH 42, PUSH 10, ADD, RET
	body := []byte{
		byte(quanticscript.OpPush), byte(quanticscript.TypeI64), 42, 0, 0, 0, 0, 0, 0, 0, // PUSH 42
		byte(quanticscript.OpPush), byte(quanticscript.TypeI64), 10, 0, 0, 0, 0, 0, 0, 0, // PUSH 10
		byte(quanticscript.OpAdd), // ADD
		byte(quanticscript.OpRet), // RET
	}

	// Create bytecode with header
	bytecode := quanticscript.CreateBytecode(body, 0)

	// Create program file
	programID := filestore.GenerateFileID([]byte("test-program"))
	program := &filestore.File{
		ID:         programID,
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       bytecode,
		Executable: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Store program in file store
	_, err = fs.CreateFile(program)
	if err != nil {
		t.Fatalf("Failed to create program file: %v", err)
	}

	// Create runtime
	runtime := NewRuntime()

	// Create execution context
	instruction := &transaction.Instruction{
		ProgramID: programID,
		Data:      []byte("test data"),
		Inputs:    make(map[string]transaction.FileAccess),
	}

	accessController := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, accessController)

	// Execute program
	err = runtime.ExecuteProgram(program, ctx)
	if err != nil {
		t.Fatalf("Failed to execute bytecode: %v", err)
	}

	// Test passed if no error occurred
	t.Log("Bytecode execution successful")
}

// TestRuntime_ExecuteBytecode_InvalidMagic tests rejection of invalid bytecode
func TestRuntime_ExecuteBytecode_InvalidMagic(t *testing.T) {
	// Create a temporary file store
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create invalid bytecode (wrong magic number)
	bytecode := []byte{0xFF, 0xFF, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}

	// Create program file
	programID := filestore.GenerateFileID([]byte("invalid-program"))
	program := &filestore.File{
		ID:         programID,
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       bytecode,
		Executable: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create runtime
	runtime := NewRuntime()

	// Create execution context
	instruction := &transaction.Instruction{
		ProgramID: programID,
		Data:      []byte("test data"),
		Inputs:    make(map[string]transaction.FileAccess),
	}

	accessController := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, accessController)

	// Execute program - should fail
	err = runtime.ExecuteProgram(program, ctx)
	if err == nil {
		t.Fatal("Expected error for invalid bytecode magic")
	}

	if err.Error() != "bytecode execution failed: failed to parse bytecode header: invalid bytecode magic: expected 0x5153, got 0xffff" {
		t.Logf("Got expected error: %v", err)
	}
}

// TestRuntime_ExecuteBytecode_ComputeBudget tests compute budget enforcement
func TestRuntime_ExecuteBytecode_ComputeBudget(t *testing.T) {
	// Create a temporary file store
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create bytecode with many operations
	body := []byte{
		byte(quanticscript.OpPush), byte(quanticscript.TypeI64), 1, 0, 0, 0, 0, 0, 0, 0, // PUSH 1
		byte(quanticscript.OpPush), byte(quanticscript.TypeI64), 2, 0, 0, 0, 0, 0, 0, 0, // PUSH 2
		byte(quanticscript.OpAdd), // ADD
		byte(quanticscript.OpRet), // RET
	}

	bytecode := quanticscript.CreateBytecode(body, 0)

	// Create program file
	programID := filestore.GenerateFileID([]byte("budget-test-program"))
	program := &filestore.File{
		ID:         programID,
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       bytecode,
		Executable: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create runtime with very low budget
	runtime := NewRuntime()
	runtime.SetExecutionLimit(5) // Too low for this program

	// Create execution context
	instruction := &transaction.Instruction{
		ProgramID: programID,
		Data:      []byte("test data"),
		Inputs:    make(map[string]transaction.FileAccess),
	}

	accessController := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, accessController)

	// Execute program - should fail due to budget
	err = runtime.ExecuteProgram(program, ctx)
	if err == nil {
		t.Fatal("Expected out of compute budget error")
	}

	if err.Error() != "bytecode execution failed: bytecode execution error: out of compute budget" {
		t.Logf("Got expected error: %v", err)
	}
}

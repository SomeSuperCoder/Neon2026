package runtime

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// TestExecutionContext_GetFile tests reading a file with proper permissions
func TestExecutionContext_GetFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with read permission
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Read},
		},
		Data: []byte{},
	}

	// Create access controller and set declared access
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Get file with read permission
	file, err := ctx.GetFile(fileID)
	if err != nil {
		t.Errorf("GetFile failed: %v", err)
	}
	if file == nil {
		t.Error("GetFile returned nil file")
	}
	if string(file.Data) != "test data" {
		t.Errorf("Expected data 'test data', got '%s'", string(file.Data))
	}

	// Verify access was logged
	accessLog := ac.GetAccessLog()
	if perm, ok := accessLog[fileID]; !ok || perm != transaction.Read {
		t.Errorf("Expected read access to be logged for file %s", fileID.String())
	}
}

// TestExecutionContext_GetFile_UndeclaredAccess tests that undeclared access is rejected
func TestExecutionContext_GetFile_UndeclaredAccess(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction WITHOUT declaring the file
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs:    map[string]transaction.FileAccess{},
		Data:      []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Attempt to get file without declaring it
	_, err = ctx.GetFile(fileID)
	if err == nil {
		t.Error("Expected error for undeclared file access, got nil")
	}
}

// TestExecutionContext_GetFileMut tests getting a file with write permission
func TestExecutionContext_GetFileMut(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with write permission
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Write},
		},
		Data: []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Get file with write permission
	file, err := ctx.GetFileMut(fileID)
	if err != nil {
		t.Errorf("GetFileMut failed: %v", err)
	}
	if file == nil {
		t.Error("GetFileMut returned nil file")
	}

	// Verify write access was logged
	accessLog := ac.GetAccessLog()
	if perm, ok := accessLog[fileID]; !ok || perm != transaction.Write {
		t.Errorf("Expected write access to be logged for file %s", fileID.String())
	}
}

// TestExecutionContext_GetFileMut_ReadOnlyPermission tests write access with read-only permission
func TestExecutionContext_GetFileMut_ReadOnlyPermission(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with READ permission only
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Read},
		},
		Data: []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Attempt to get file with write permission when only read is declared
	_, err = ctx.GetFileMut(fileID)
	if err == nil {
		t.Error("Expected error for write access with read-only permission, got nil")
	}
}

// TestExecutionContext_UpdateFile tests updating a file
func TestExecutionContext_UpdateFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("original data"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with write permission
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Write},
		},
		Data: []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Get file for modification
	file, err := ctx.GetFileMut(fileID)
	if err != nil {
		t.Fatalf("GetFileMut failed: %v", err)
	}

	// Modify file data
	file.Data = []byte("updated data")

	// Test: Update file
	err = ctx.UpdateFile(file)
	if err != nil {
		t.Errorf("UpdateFile failed: %v", err)
	}

	// Verify file was updated
	updatedFile, err := fs.GetFile(fileID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated file: %v", err)
	}
	if string(updatedFile.Data) != "updated data" {
		t.Errorf("Expected data 'updated data', got '%s'", string(updatedFile.Data))
	}
}

// TestExecutionContext_GetFileBalance tests balance retrieval
func TestExecutionContext_GetFileBalance(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    12345,
		TxManager:  filestore.FileID{},
		Data:       []byte("test"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with read permission
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Read},
		},
		Data: []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Get file balance
	balance, err := ctx.GetFileBalance(fileID)
	if err != nil {
		t.Errorf("GetFileBalance failed: %v", err)
	}
	if balance != 12345 {
		t.Errorf("Expected balance 12345, got %d", balance)
	}
}

// TestExecutionContext_UpdateFileBalance tests balance updates
func TestExecutionContext_UpdateFileBalance(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a test file
	testFile := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test"),
		Executable: false,
	}
	fileID, err := fs.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create instruction with write permission
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"test": {FileID: fileID, Permission: transaction.Write},
		},
		Data: []byte{},
	}

	// Create access controller
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)

	// Create execution context
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Update balance
	err = ctx.UpdateFileBalance(fileID, 5000)
	if err != nil {
		t.Errorf("UpdateFileBalance failed: %v", err)
	}

	// Verify balance was updated
	newBalance, err := fs.GetFileBalance(fileID)
	if err != nil {
		t.Fatalf("Failed to get updated balance: %v", err)
	}
	if newBalance != 15000 {
		t.Errorf("Expected balance 15000, got %d", newBalance)
	}
}

// TestExecutionContext_HasSigner tests signer verification
func TestExecutionContext_HasSigner(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	signer1 := transaction.PublicKey{1, 2, 3}
	signer2 := transaction.PublicKey{4, 5, 6}
	nonSigner := transaction.PublicKey{7, 8, 9}

	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs:    map[string]transaction.FileAccess{},
		Data:      []byte{},
	}

	ac := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{signer1, signer2}, fs, ac)

	// Test: Check for existing signers
	if !ctx.HasSigner(signer1) {
		t.Error("Expected signer1 to be present")
	}
	if !ctx.HasSigner(signer2) {
		t.Error("Expected signer2 to be present")
	}

	// Test: Check for non-signer
	if ctx.HasSigner(nonSigner) {
		t.Error("Expected nonSigner to not be present")
	}
}

// TestExecutionContext_GetInstructionData tests instruction data retrieval
func TestExecutionContext_GetInstructionData(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	testData := []byte("instruction data")
	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs:    map[string]transaction.FileAccess{},
		Data:      testData,
	}

	ac := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Get instruction data
	data := ctx.GetInstructionData()
	if string(data) != string(testData) {
		t.Errorf("Expected data '%s', got '%s'", string(testData), string(data))
	}
}

// TestExecutionContext_GetInputFileID tests retrieving file IDs from inputs
func TestExecutionContext_GetInputFileID(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	fileID1 := filestore.FileID{1, 2, 3}
	fileID2 := filestore.FileID{4, 5, 6}

	instruction := &transaction.Instruction{
		ProgramID: filestore.FileID{1},
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: fileID1, Permission: transaction.Write},
			"to":   {FileID: fileID2, Permission: transaction.Write},
		},
		Data: []byte{},
	}

	ac := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Get existing input file IDs
	fromID, err := ctx.GetInputFileID("from")
	if err != nil {
		t.Errorf("GetInputFileID('from') failed: %v", err)
	}
	if fromID != fileID1 {
		t.Errorf("Expected fileID1, got %v", fromID)
	}

	toID, err := ctx.GetInputFileID("to")
	if err != nil {
		t.Errorf("GetInputFileID('to') failed: %v", err)
	}
	if toID != fileID2 {
		t.Errorf("Expected fileID2, got %v", toID)
	}

	// Test: Get non-existent input
	_, err = ctx.GetInputFileID("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent input key, got nil")
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// TestNewRuntime tests Runtime initialization
func TestNewRuntime(t *testing.T) {
	runtime := NewRuntime()
	if runtime == nil {
		t.Fatal("NewRuntime returned nil")
	}
	if runtime.builtinPrograms == nil {
		t.Error("builtinPrograms map not initialized")
	}
	if runtime.executionLimit != 1000000 {
		t.Errorf("Expected default execution limit 1000000, got %d", runtime.executionLimit)
	}
}

// mockBuiltinProgram is a test implementation of BuiltinProgram
type mockBuiltinProgram struct {
	id          filestore.FileID
	executeFunc func(ctx *ExecutionContext) error
}

func (m *mockBuiltinProgram) Execute(ctx *ExecutionContext) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil
}

func (m *mockBuiltinProgram) GetProgramID() filestore.FileID {
	return m.id
}

// TestRuntime_RegisterBuiltinProgram tests registering builtin programs
func TestRuntime_RegisterBuiltinProgram(t *testing.T) {
	runtime := NewRuntime()
	programID := filestore.FileID{1, 2, 3}

	mockProgram := &mockBuiltinProgram{
		id: programID,
	}

	// Test: Register builtin program
	err := runtime.RegisterBuiltinProgram(mockProgram)
	if err != nil {
		t.Errorf("RegisterBuiltinProgram failed: %v", err)
	}

	// Verify program was registered
	if !runtime.IsBuiltinProgram(programID) {
		t.Error("Program not registered")
	}

	// Test: Register duplicate program
	err = runtime.RegisterBuiltinProgram(mockProgram)
	if err == nil {
		t.Error("Expected error when registering duplicate program")
	}
}

// TestRuntime_ValidateProgram tests program validation
func TestRuntime_ValidateProgram(t *testing.T) {
	runtime := NewRuntime()

	// Test: Valid executable program
	validProgram := &filestore.File{
		ID:         filestore.FileID{1},
		Executable: true,
		Data:       []byte("bytecode"),
	}
	err := runtime.ValidateProgram(validProgram)
	if err != nil {
		t.Errorf("ValidateProgram failed for valid program: %v", err)
	}

	// Test: Non-executable file
	nonExecutable := &filestore.File{
		ID:         filestore.FileID{2},
		Executable: false,
		Data:       []byte("data"),
	}
	err = runtime.ValidateProgram(nonExecutable)
	if err == nil {
		t.Error("Expected error for non-executable file")
	}

	// Test: Nil program
	err = runtime.ValidateProgram(nil)
	if err == nil {
		t.Error("Expected error for nil program")
	}
}

// TestRuntime_ExecuteProgram_Builtin tests executing a builtin program
func TestRuntime_ExecuteProgram_Builtin(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	runtime := NewRuntime()
	programID := filestore.FileID{1, 2, 3}

	// Create mock builtin program
	executed := false
	mockProgram := &mockBuiltinProgram{
		id: programID,
		executeFunc: func(ctx *ExecutionContext) error {
			executed = true
			return nil
		},
	}

	// Register builtin program
	err = runtime.RegisterBuiltinProgram(mockProgram)
	if err != nil {
		t.Fatalf("Failed to register builtin program: %v", err)
	}

	// Create program file
	programFile := &filestore.File{
		ID:         programID,
		Executable: true,
		Data:       []byte{},
	}

	// Create execution context
	instruction := &transaction.Instruction{
		ProgramID: programID,
		Inputs:    map[string]transaction.FileAccess{},
		Data:      []byte{},
	}
	ac := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Execute builtin program
	err = runtime.ExecuteProgram(programFile, ctx)
	if err != nil {
		t.Errorf("ExecuteProgram failed: %v", err)
	}
	if !executed {
		t.Error("Builtin program was not executed")
	}
}

// TestRuntime_ExecuteProgram_NonBuiltin tests executing a non-builtin program
func TestRuntime_ExecuteProgram_NonBuiltin(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	fs, err := filestore.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	runtime := NewRuntime()
	programID := filestore.FileID{4, 5, 6}

	// Create non-builtin program file
	programFile := &filestore.File{
		ID:         programID,
		Executable: true,
		Data:       []byte("bytecode"),
	}

	// Create execution context
	instruction := &transaction.Instruction{
		ProgramID: programID,
		Inputs:    map[string]transaction.FileAccess{},
		Data:      []byte{},
	}
	ac := access.NewAccessController()
	ctx := NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Test: Execute non-builtin program (should fail with not implemented error)
	err = runtime.ExecuteProgram(programFile, ctx)
	if err == nil {
		t.Error("Expected error for non-builtin program execution")
	}
}

// TestRuntime_ExecutionLimit tests execution limit management
func TestRuntime_ExecutionLimit(t *testing.T) {
	runtime := NewRuntime()

	// Test: Default execution limit
	if runtime.GetExecutionLimit() != 1000000 {
		t.Errorf("Expected default limit 1000000, got %d", runtime.GetExecutionLimit())
	}

	// Test: Set custom execution limit
	runtime.SetExecutionLimit(500000)
	if runtime.GetExecutionLimit() != 500000 {
		t.Errorf("Expected limit 500000, got %d", runtime.GetExecutionLimit())
	}
}

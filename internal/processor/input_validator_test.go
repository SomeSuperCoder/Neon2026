package processor

import (
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/transaction"
)

// TestValidateInstructionInputs_ValidInputs tests validation of properly declared inputs
func TestValidateInstructionInputs_ValidInputs(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03}, // dummy bytecode
		Executable: true,
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create sender and receiver files
	senderID := filestore.GenerateFileID([]byte("sender"))
	receiverID := filestore.GenerateFileID([]byte("receiver"))

	sender := &filestore.File{
		ID:         senderID,
		Balance:    50000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(sender)
	if err != nil {
		t.Fatalf("Failed to create sender: %v", err)
	}

	receiver := &filestore.File{
		ID:         receiverID,
		Balance:    10000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(receiver)
	if err != nil {
		t.Fatalf("Failed to create receiver: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create instruction with proper input declarations
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"program": {
				FileID:     genesis.SystemProgramID,
				Permission: transaction.Read,
			},
			"sender": {
				FileID:     senderID,
				Permission: transaction.Write,
			},
			"receiver": {
				FileID:     receiverID,
				Permission: transaction.Write,
			},
		},
		Data: []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

// TestValidateInstructionInputs_MissingProgramDeclaration tests detection of missing program declaration
func TestValidateInstructionInputs_MissingProgramDeclaration(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: true,
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create instruction WITHOUT program in inputs
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs:    map[string]transaction.FileAccess{}, // Empty!
		Data:      []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err == nil {
		t.Error("Expected validation to fail for missing program declaration")
	}

	// Check error message contains program ID
	if err != nil && !contains(err.Error(), "not declared in instruction inputs") {
		t.Errorf("Expected error about missing program declaration, got: %v", err)
	}
}

// TestValidateInstructionInputs_NonExecutableProgram tests detection of non-executable program
func TestValidateInstructionInputs_NonExecutableProgram(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file WITHOUT executable flag
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: false, // NOT EXECUTABLE
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create instruction with program declared
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"program": {
				FileID:     genesis.SystemProgramID,
				Permission: transaction.Read,
			},
		},
		Data: []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err == nil {
		t.Error("Expected validation to fail for non-executable program")
	}

	// Check error message
	if err != nil && !contains(err.Error(), "not marked as executable") {
		t.Errorf("Expected error about non-executable program, got: %v", err)
	}
}

// TestValidateInstructionInputs_MissingFile tests detection of missing files
func TestValidateInstructionInputs_MissingFile(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: true,
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create a non-existent file ID
	nonExistentID := filestore.GenerateFileID([]byte("does-not-exist"))

	// Create instruction with non-existent file
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"program": {
				FileID:     genesis.SystemProgramID,
				Permission: transaction.Read,
			},
			"sender": {
				FileID:     nonExistentID,
				Permission: transaction.Write,
			},
		},
		Data: []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err == nil {
		t.Error("Expected validation to fail for missing file")
	}

	// Check error message
	if err != nil && !contains(err.Error(), "not found") {
		t.Errorf("Expected error about file not found, got: %v", err)
	}
}

// TestValidateInstructionInputs_InvalidPermission tests detection of incorrect permissions
func TestValidateInstructionInputs_InvalidPermission(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: true,
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create instruction with invalid permission (0 or > 2)
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"program": {
				FileID:     genesis.SystemProgramID,
				Permission: transaction.AccessPermission(99), // Invalid
			},
		},
		Data: []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err == nil {
		t.Error("Expected validation to fail for invalid permission")
	}

	// Check error message
	if err != nil && !contains(err.Error(), "invalid permission") {
		t.Errorf("Expected error about invalid permission, got: %v", err)
	}
}

// TestValidateInstructionInputs_EmptyInputsMap tests validation of empty inputs map
func TestValidateInstructionInputs_EmptyInputsMap(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create System Program file
	systemProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: true,
	}
	_, err = fs.CreateFile(systemProgram)
	if err != nil {
		t.Fatalf("Failed to create system program: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Create instruction with empty inputs map
	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs:    map[string]transaction.FileAccess{},
		Data:      []byte{1, 2, 3},
	}

	// Validate
	err = validator.ValidateInstructionInputs(instr)
	if err == nil {
		t.Error("Expected validation to fail for empty inputs map")
	}
}

// TestValidateExecutableProgram tests program executable validation
func TestValidateExecutableProgram(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create executable program
	executableProgram := &filestore.File{
		ID:         genesis.SystemProgramID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: true,
	}
	_, err = fs.CreateFile(executableProgram)
	if err != nil {
		t.Fatalf("Failed to create executable program: %v", err)
	}

	// Create non-executable file
	nonExecID := filestore.GenerateFileID([]byte("non-exec"))
	nonExecutable := &filestore.File{
		ID:         nonExecID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{0x01, 0x02, 0x03},
		Executable: false,
	}
	_, err = fs.CreateFile(nonExecutable)
	if err != nil {
		t.Fatalf("Failed to create non-executable file: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Test executable program - should pass
	err = validator.ValidateExecutableProgram(genesis.SystemProgramID)
	if err != nil {
		t.Errorf("Expected executable program validation to pass, got: %v", err)
	}

	// Test non-executable file - should fail
	err = validator.ValidateExecutableProgram(nonExecID)
	if err == nil {
		t.Error("Expected non-executable program validation to fail")
	}
}

// TestValidateFileExists tests file existence validation
func TestValidateFileExists(t *testing.T) {
	// Setup
	fs, err := filestore.NewFileStore(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	// Create a file
	existingID := filestore.GenerateFileID([]byte("existing"))
	existingFile := &filestore.File{
		ID:         existingID,
		Balance:    10000,
		TxManager:  genesis.RuntimeProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create validator
	validator := NewInputValidator(fs)

	// Test existing file - should pass
	err = validator.ValidateFileExists(existingID)
	if err != nil {
		t.Errorf("Expected file existence validation to pass, got: %v", err)
	}

	// Test non-existent file - should fail
	nonExistentID := filestore.GenerateFileID([]byte("does-not-exist"))
	err = validator.ValidateFileExists(nonExistentID)
	if err == nil {
		t.Error("Expected file existence validation to fail for non-existent file")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

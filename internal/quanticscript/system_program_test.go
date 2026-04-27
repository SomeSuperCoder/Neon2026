package quanticscript

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// SystemProgramID is the built-in system program (0x00...01)
// Defined here to avoid import cycle with genesis package
var SystemProgramID = filestore.FileID{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 1,
}

// TestSystemProgramCompilation tests that the System_Program compiles successfully
func TestSystemProgramCompilation(t *testing.T) {
	// Read the compiled bytecode
	bytecode, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	// Verify bytecode is not empty
	if len(bytecode) == 0 {
		t.Fatal("System_Program bytecode is empty")
	}

	// Verify bytecode has valid header (8 bytes minimum)
	if len(bytecode) < 8 {
		t.Fatal("System_Program bytecode too short")
	}

	// Check magic number (0x5153 = "QS")
	magic := uint16(bytecode[0]) | (uint16(bytecode[1]) << 8)
	if magic != 0x5153 {
		t.Errorf("Expected magic 0x5153, got 0x%04x", magic)
	}

	// Check version (0x0100)
	version := uint16(bytecode[2]) | (uint16(bytecode[3]) << 8)
	if version != 0x0100 {
		t.Errorf("Expected version 0x0100, got 0x%04x", version)
	}

	t.Logf("System_Program bytecode size: %d bytes", len(bytecode))
	t.Logf("Magic: 0x%04x, Version: 0x%04x", magic, version)

	// Verify bytecode can be loaded by interpreter
	ctx := NewMockExecutionContext()
	interp := NewBytecodeInterpreter(bytecode, ctx, 1000000)
	if interp == nil {
		t.Fatal("Failed to create interpreter for System_Program bytecode")
	}
}

// TestSystemProgramStructure tests the structure of the System_Program
func TestSystemProgramStructure(t *testing.T) {
	// Read the source code
	source, err := os.ReadFile("../../programs/system/system.qs")
	if err != nil {
		t.Fatalf("Failed to read System_Program source: %v", err)
	}

	// Verify source contains required functions
	sourceStr := string(source)

	requiredFunctions := []string{
		"entry",
		"handleCreateFile",
		"handleTransfer",
		"handleBurn",
		"handleCloseFile",
	}

	for _, fn := range requiredFunctions {
		if !contains(sourceStr, fn) {
			t.Errorf("System_Program missing required function: %s", fn)
		}
	}

	// Verify error codes are defined
	requiredConstants := []string{
		"ERROR_INSUFFICIENT_BALANCE",
		"ERROR_UNAUTHORIZED",
		"ERROR_INVALID_AMOUNT",
		"ERROR_FILE_HAS_DATA",
		"ERROR_INVALID_INSTRUCTION",
		"SUCCESS",
	}

	for _, constant := range requiredConstants {
		if !contains(sourceStr, constant) {
			t.Errorf("System_Program missing required constant: %s", constant)
		}
	}

	t.Log("System_Program structure verified")
}

// setupSystemProgramTest creates a test environment with the System Program loaded
func setupSystemProgramTest(t *testing.T) (*filestore.FileStore, []byte) {
	t.Helper()

	dbPath := t.TempDir() + "/system_test.db"
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	t.Cleanup(func() { fs.Close() })

	// Load System Program bytecode
	bytecode, err := os.ReadFile("../../programs/system/system.qsb")
	if err != nil {
		t.Fatalf("Failed to read System_Program bytecode: %v", err)
	}

	// Load the System Program into the FileStore
	storageCost := filestore.CalculateStorageCost(int64(len(bytecode)))
	balance := storageCost // System Program only needs enough for its own storage

	file := &filestore.File{
		ID:         SystemProgramID,
		Balance:    balance,
		TxManager:  SystemProgramID, // System Program owns itself
		Data:       bytecode,
		Executable: true,
	}

	if _, err := fs.CreateFile(file); err != nil {
		t.Fatalf("Failed to load System Program: %v", err)
	}

	return fs, bytecode
}

// encodeCreateFileInstruction encodes a CREATE_FILE instruction
// Format: [opcode(1), fileID(32), payer(32), balance(8), owner(32)]
func encodeCreateFileInstruction(fileID, payerID filestore.FileID, balance int64, owner transaction.PublicKey) []byte {
	data := make([]byte, 105)
	data[0] = 0 // CREATE_FILE opcode
	copy(data[1:33], fileID[:])
	copy(data[33:65], payerID[:])
	binary.LittleEndian.PutUint64(data[65:73], uint64(balance))
	copy(data[73:105], owner[:])
	return data
}

// encodeTransferInstruction encodes a TRANSFER instruction
// Format: [opcode(1), from(32), to(32), amount(8)]
func encodeTransferInstruction(from, to filestore.FileID, amount int64) []byte {
	data := make([]byte, 73)
	data[0] = 1 // TRANSFER opcode
	copy(data[1:33], from[:])
	copy(data[33:65], to[:])
	binary.LittleEndian.PutUint64(data[65:73], uint64(amount))
	return data
}

// encodeBurnInstruction encodes a BURN instruction
// Format: [opcode(1), fileID(32), amount(8)]
func encodeBurnInstruction(fileID filestore.FileID, amount int64) []byte {
	data := make([]byte, 41)
	data[0] = 2 // BURN opcode
	copy(data[1:33], fileID[:])
	binary.LittleEndian.PutUint64(data[33:41], uint64(amount))
	return data
}

// encodeCloseFileInstruction encodes a CLOSE_FILE instruction
// Format: [opcode(1), fileID(32), destination(32)]
func encodeCloseFileInstruction(fileID, destination filestore.FileID) []byte {
	data := make([]byte, 65)
	data[0] = 3 // CLOSE_FILE opcode
	copy(data[1:33], fileID[:])
	copy(data[33:65], destination[:])
	return data
}

// executeSystemProgram executes the System Program with the given instruction data
func executeSystemProgram(t *testing.T, fs *filestore.FileStore, bytecode []byte, instrData []byte, signers []transaction.PublicKey, inputs map[string]transaction.FileAccess) error {
	t.Helper()

	t.Logf("DEBUG: executeSystemProgram called with instrData length=%d", len(instrData))
	t.Logf("DEBUG: inputs=%+v", inputs)
	t.Logf("DEBUG: bytecode length=%d", len(bytecode))

	instr := &transaction.Instruction{
		ProgramID: SystemProgramID,
		Inputs:    inputs,
		Data:      instrData,
	}

	ac := access.NewAccessController()
	ac.SetDeclaredAccess(inputs)

	ctx := &mockExecutionContext{
		fs:      fs,
		ac:      ac,
		instr:   instr,
		signers: signers,
	}

	// Parse bytecode header and extract body (like runtime.go does)
	header, err := ParseBytecodeHeader(bytecode)
	if err != nil {
		t.Fatalf("Failed to parse bytecode header: %v", err)
	}
	t.Logf("DEBUG: Bytecode header: entry offset=%d", header.EntryOffset)

	body, err := GetBytecodeBody(bytecode)
	if err != nil {
		t.Fatalf("Failed to get bytecode body: %v", err)
	}
	t.Logf("DEBUG: Bytecode body length=%d", len(body))

	t.Logf("DEBUG: Creating interpreter...")
	interp := NewBytecodeInterpreter(body, ctx, 1000000)
	if interp == nil {
		t.Fatal("Failed to create interpreter")
	}
	t.Logf("DEBUG: Interpreter created, calling Execute()...")
	err = interp.Execute()
	t.Logf("DEBUG: interp.Execute() returned err=%v", err)
	t.Logf("DEBUG: Stack length after execution: %d", len(interp.stack))

	// Check the return value on the stack
	if err == nil && len(interp.stack) > 0 {
		retVal := interp.stack[len(interp.stack)-1]
		t.Logf("DEBUG: Return value on stack: type=%v, data=%v", retVal.Type, retVal.Data)
		if retVal.Type == TypeI64 {
			val, _ := retVal.AsI64()
			t.Logf("DEBUG: Return value as i64: %d (0x%x)", val, val)
		}
	}

	return err
}

// mockExecutionContext implements ExecutionContext for testing
type mockExecutionContext struct {
	fs      *filestore.FileStore
	ac      *access.AccessController
	instr   *transaction.Instruction
	signers []transaction.PublicKey
}

func (m *mockExecutionContext) GetFile(fileID filestore.FileID) (*filestore.File, error) {
	if err := m.ac.ValidateAccess(fileID, transaction.Read); err != nil {
		return nil, err
	}
	return m.fs.GetFile(fileID)
}

func (m *mockExecutionContext) GetFileMut(fileID filestore.FileID) (*filestore.File, error) {
	if err := m.ac.ValidateAccess(fileID, transaction.Write); err != nil {
		return nil, err
	}
	return m.fs.GetFile(fileID)
}

func (m *mockExecutionContext) UpdateFile(file *filestore.File) error {
	if err := m.ac.ValidateAccess(file.ID, transaction.Write); err != nil {
		return err
	}
	return m.fs.UpdateFile(file.ID, file)
}

func (m *mockExecutionContext) CreateFile(file *filestore.File) error {
	fmt.Printf("DEBUG: mockExecutionContext.CreateFile called for fileID=%v\n", file.ID)
	if err := m.ac.ValidateAccess(file.ID, transaction.Write); err != nil {
		// Debug: log access control failure
		fmt.Printf("DEBUG: CreateFile access control failed for %v: %v\n", file.ID, err)
		return err
	}
	fmt.Printf("DEBUG: CreateFile access control passed for %v\n", file.ID)
	_, err := m.fs.CreateFile(file)
	if err != nil {
		fmt.Printf("DEBUG: FileStore.CreateFile failed for %v: %v\n", file.ID, err)
	} else {
		fmt.Printf("DEBUG: FileStore.CreateFile succeeded for %v\n", file.ID)
	}
	return err
}

func (m *mockExecutionContext) DeleteFile(fileID filestore.FileID) error {
	if err := m.ac.ValidateAccess(fileID, transaction.Write); err != nil {
		return err
	}
	return m.fs.DeleteFile(fileID)
}

func (m *mockExecutionContext) GetFileBalance(fileID filestore.FileID) (int64, error) {
	if err := m.ac.ValidateAccess(fileID, transaction.Read); err != nil {
		return 0, err
	}
	file, err := m.fs.GetFile(fileID)
	if err != nil {
		return 0, err
	}
	return file.Balance, nil
}

func (m *mockExecutionContext) UpdateFileBalance(fileID filestore.FileID, delta int64) error {
	if err := m.ac.ValidateAccess(fileID, transaction.Write); err != nil {
		return err
	}
	return m.fs.UpdateFileBalance(fileID, delta)
}

func (m *mockExecutionContext) HasSigner(pubkey transaction.PublicKey) bool {
	for _, s := range m.signers {
		if s == pubkey {
			return true
		}
	}
	return false
}

func (m *mockExecutionContext) GetInstructionData() []byte {
	return m.instr.Data
}

func (m *mockExecutionContext) GetProgramID() filestore.FileID {
	return m.instr.ProgramID
}

func (m *mockExecutionContext) GetSigners() []transaction.PublicKey {
	return m.signers
}

func (m *mockExecutionContext) QueryBlock(blockHash []byte) ([]byte, error) {
	return nil, nil
}

func (m *mockExecutionContext) QueryTransaction(txID transaction.TxID) ([]byte, error) {
	return nil, nil
}

func (m *mockExecutionContext) QueryInstruction(txID transaction.TxID, instrIndex uint32) ([]byte, error) {
	return nil, nil
}

func (m *mockExecutionContext) InvokeProgram(programID filestore.FileID, invokeData []byte, computeBudget int64, depth int) ([]byte, error) {
	return nil, nil
}

func (m *mockExecutionContext) GetDeclaredPrograms() []filestore.FileID {
	return nil
}

// TestSystemCreateFile tests the CREATE_FILE operation
func TestSystemCreateFile(t *testing.T) {
	t.Run("DirectCreateFileTest", func(t *testing.T) {
		fs, _ := setupSystemProgramTest(t)

		testFileID := filestore.GenerateFileID([]byte("direct-test"))
		testFile := &filestore.File{
			ID:         testFileID,
			Balance:    1000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(testFile)
		if err != nil {
			t.Fatalf("Direct CreateFile failed: %v", err)
		}

		retrieved, err := fs.GetFile(testFileID)
		if err != nil {
			t.Fatalf("Failed to retrieve directly created file: %v", err)
		}

		if retrieved.Balance != 1000 {
			t.Errorf("Balance = %d, want 1000", retrieved.Balance)
		}
		t.Log("Direct CreateFile works!")
	})

	t.Run("OpcodeCreateFileTest", func(t *testing.T) {
		fs, _ := setupSystemProgramTest(t)

		testFileID := filestore.GenerateFileID([]byte("opcode-test"))

		// Create a simple bytecode that calls CREATEFILE
		// PUSH fileID, PUSH emptyData, PUSH balance, CREATEFILE, RET
		bytecode := []byte{
			byte(OpPush), byte(TypeFileID),
		}
		bytecode = append(bytecode, testFileID[:]...)

		// PUSH empty bytes (length 0)
		bytecode = append(bytecode, byte(OpPush), byte(TypeBytes))
		bytecode = append(bytecode, 0, 0, 0, 0, 0, 0, 0, 0) // length = 0 (8 bytes)

		// PUSH balance 1000
		bytecode = append(bytecode, byte(OpPush), byte(TypeI64))
		bytecode = append(bytecode, 232, 3, 0, 0, 0, 0, 0, 0) // 1000 in little-endian

		bytecode = append(bytecode, byte(OpCreateFile), byte(OpRet))

		inputs := map[string]transaction.FileAccess{
			"test_file": {FileID: testFileID, Permission: transaction.Write},
		}

		instr := &transaction.Instruction{
			ProgramID: SystemProgramID,
			Inputs:    inputs,
			Data:      []byte{},
		}

		ac := access.NewAccessController()
		ac.SetDeclaredAccess(inputs)

		ctx := &mockExecutionContext{
			fs:      fs,
			ac:      ac,
			instr:   instr,
			signers: []transaction.PublicKey{},
		}

		interp := NewBytecodeInterpreter(bytecode, ctx, 1000000)
		err := interp.Execute()
		if err != nil {
			t.Fatalf("Opcode CREATEFILE failed: %v", err)
		}

		retrieved, err := fs.GetFile(testFileID)
		if err != nil {
			t.Fatalf("Failed to retrieve opcode-created file: %v", err)
		}

		if retrieved.Balance != 1000 {
			t.Errorf("Balance = %d, want 1000", retrieved.Balance)
		}
		t.Log("Opcode CREATEFILE works!")
	})

	t.Run("ValidCreateFile", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("test-file"))
		payerID := filestore.GenerateFileID([]byte("payer-account"))
		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create payer account with sufficient balance
		payerFile := &filestore.File{
			ID:         payerID,
			Balance:    100000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		_, err := fs.CreateFile(payerFile)
		if err != nil {
			t.Fatalf("Failed to create payer: %v", err)
		}

		instrData := encodeCreateFileInstruction(fileID, payerID, 10000, owner)

		inputs := map[string]transaction.FileAccess{
			"program":  {FileID: SystemProgramID, Permission: transaction.Read},
			"new_file": {FileID: fileID, Permission: transaction.Write},
			"payer":    {FileID: payerID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		t.Logf("Execution error: %v", err)
		if err != nil {
			t.Fatalf("CREATE_FILE failed: %v", err)
		}

		// Verify file was created
		file, err := fs.GetFile(fileID)
		if err != nil {
			t.Fatalf("File not created: %v", err)
		}

		if file.Balance != 10000 {
			t.Errorf("Balance = %d, want 10000", file.Balance)
		}

		if file.TxManager != SystemProgramID {
			t.Errorf("TxManager = %v, want SystemProgramID", file.TxManager)
		}

		// Verify payer balance decreased
		payerAfter, _ := fs.GetFile(payerID)
		if payerAfter.Balance != 90000 {
			t.Errorf("Payer balance = %d, want 90000", payerAfter.Balance)
		}
	})

	t.Run("CreateFileWithZeroBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("zero-balance-file"))
		payerID := filestore.GenerateFileID([]byte("payer-account"))
		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create payer account
		payerFile := &filestore.File{
			ID:         payerID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		_, err := fs.CreateFile(payerFile)
		if err != nil {
			t.Fatalf("Failed to create payer: %v", err)
		}

		instrData := encodeCreateFileInstruction(fileID, payerID, 0, owner)

		inputs := map[string]transaction.FileAccess{
			"program":  {FileID: SystemProgramID, Permission: transaction.Read},
			"new_file": {FileID: fileID, Permission: transaction.Write},
			"payer":    {FileID: payerID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("CREATE_FILE with zero balance failed: %v", err)
		}

		file, err := fs.GetFile(fileID)
		if err != nil {
			t.Fatalf("File not created: %v", err)
		}

		if file.Balance != 0 {
			t.Errorf("Balance = %d, want 0", file.Balance)
		}
	})
}

// TestSystemTransfer tests the TRANSFER operation
func TestSystemTransfer(t *testing.T) {
	t.Run("ValidTransfer", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fromID := filestore.GenerateFileID([]byte("from-account"))
		toID := filestore.GenerateFileID([]byte("to-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create source and destination files
		fromFile := &filestore.File{
			ID:         fromID,
			Balance:    20000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		toFile := &filestore.File{
			ID:         toID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(fromFile)
		if err != nil {
			t.Fatalf("Failed to create from file: %v", err)
		}
		_, err = fs.CreateFile(toFile)
		if err != nil {
			t.Fatalf("Failed to create to file: %v", err)
		}

		instrData := encodeTransferInstruction(fromID, toID, 3000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"from":    {FileID: fromID, Permission: transaction.Write},
			"to":      {FileID: toID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("TRANSFER failed: %v", err)
		}

		// Verify balances
		fromAfter, _ := fs.GetFile(fromID)
		toAfter, _ := fs.GetFile(toID)

		if fromAfter.Balance != 17000 {
			t.Errorf("From balance = %d, want 17000", fromAfter.Balance)
		}

		if toAfter.Balance != 8000 {
			t.Errorf("To balance = %d, want 8000", toAfter.Balance)
		}
	})

	t.Run("TransferInsufficientBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fromID := filestore.GenerateFileID([]byte("poor-account"))
		toID := filestore.GenerateFileID([]byte("rich-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		fromFile := &filestore.File{
			ID:         fromID,
			Balance:    1000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		toFile := &filestore.File{
			ID:         toID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(fromFile)
		if err != nil {
			t.Fatalf("Failed to create from file: %v", err)
		}
		_, err = fs.CreateFile(toFile)
		if err != nil {
			t.Fatalf("Failed to create to file: %v", err)
		}

		instrData := encodeTransferInstruction(fromID, toID, 5000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"from":    {FileID: fromID, Permission: transaction.Write},
			"to":      {FileID: toID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected TRANSFER to fail with insufficient balance")
		}

		// Verify balances unchanged
		fromAfter, _ := fs.GetFile(fromID)
		toAfter, _ := fs.GetFile(toID)

		if fromAfter.Balance != 1000 {
			t.Errorf("From balance changed: %d, want 1000", fromAfter.Balance)
		}

		if toAfter.Balance != 5000 {
			t.Errorf("To balance changed: %d, want 5000", toAfter.Balance)
		}
	})
}

// TestSystemBurn tests the BURN operation
func TestSystemBurn(t *testing.T) {
	t.Run("ValidBurn", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("burn-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		instrData := encodeBurnInstruction(fileID, 3000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"file":    {FileID: fileID, Permission: transaction.Write},
			"system":  {FileID: SystemProgramID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("BURN failed: %v", err)
		}

		// Verify balance decreased
		fileAfter, _ := fs.GetFile(fileID)

		if fileAfter.Balance != 7000 {
			t.Errorf("Balance = %d, want 7000", fileAfter.Balance)
		}
	})

	t.Run("BurnInsufficientBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("poor-burn-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    1000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		instrData := encodeBurnInstruction(fileID, 5000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"file":    {FileID: fileID, Permission: transaction.Write},
			"system":  {FileID: SystemProgramID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected BURN to fail with insufficient balance")
		}

		// Verify balance unchanged
		fileAfter, _ := fs.GetFile(fileID)

		if fileAfter.Balance != 1000 {
			t.Errorf("Balance changed: %d, want 1000", fileAfter.Balance)
		}
	})

	t.Run("BurnWithStorageCost", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("storage-burn-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create file with data that requires storage cost
		data := make([]byte, 1024) // 1KB of data
		storageCost := filestore.CalculateStorageCost(int64(len(data)))

		file := &filestore.File{
			ID:         fileID,
			Balance:    storageCost + 5000,
			TxManager:  SystemProgramID,
			Data:       data,
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Try to burn more than available (leaving less than storage cost)
		instrData := encodeBurnInstruction(fileID, 6000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"file":    {FileID: fileID, Permission: transaction.Write},
			"system":  {FileID: SystemProgramID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected BURN to fail when violating storage cost")
		}
	})
}

// TestSystemCloseFile tests the CLOSE_FILE operation
func TestSystemCloseFile(t *testing.T) {
	t.Run("ValidCloseFile", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("close-account"))
		destID := filestore.GenerateFileID([]byte("destination-account"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{}, // Empty data
			Executable: false,
		}
		destFile := &filestore.File{
			ID:         destID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		_, err = fs.CreateFile(destFile)
		if err != nil {
			t.Fatalf("Failed to create dest file: %v", err)
		}

		instrData := encodeCloseFileInstruction(fileID, destID)

		inputs := map[string]transaction.FileAccess{
			"program":     {FileID: SystemProgramID, Permission: transaction.Read},
			"file":        {FileID: fileID, Permission: transaction.Write},
			"destination": {FileID: destID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("CLOSE_FILE failed: %v", err)
		}

		// Verify file is deleted
		_, err = fs.GetFile(fileID)
		if err == nil {
			t.Error("File should be deleted")
		}

		// Verify destination received the balance
		destAfter, _ := fs.GetFile(destID)
		if destAfter.Balance != 15000 {
			t.Errorf("Destination balance = %d, want 15000", destAfter.Balance)
		}
	})

	t.Run("CloseFileWithData", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("close-with-data"))
		destID := filestore.GenerateFileID([]byte("destination"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create file with non-empty data
		data := []byte("some data")
		storageCost := filestore.CalculateStorageCost(int64(len(data)))

		file := &filestore.File{
			ID:         fileID,
			Balance:    storageCost + 5000,
			TxManager:  SystemProgramID,
			Data:       data,
			Executable: false,
		}
		destFile := &filestore.File{
			ID:         destID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		_, err = fs.CreateFile(destFile)
		if err != nil {
			t.Fatalf("Failed to create dest file: %v", err)
		}

		instrData := encodeCloseFileInstruction(fileID, destID)

		inputs := map[string]transaction.FileAccess{
			"program":     {FileID: SystemProgramID, Permission: transaction.Read},
			"file":        {FileID: fileID, Permission: transaction.Write},
			"destination": {FileID: destID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected CLOSE_FILE to fail with non-empty data")
		}

		// Verify file still exists
		fileAfter, err := fs.GetFile(fileID)
		if err != nil {
			t.Error("File should still exist")
		}
		if fileAfter.Balance != storageCost+5000 {
			t.Errorf("File balance changed unexpectedly")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestSystemCreateFileEdgeCases tests edge cases for CREATE_FILE
func TestSystemCreateFileEdgeCases(t *testing.T) {
	t.Run("CreateFileWithLargeBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("large-balance-file"))
		payerID := filestore.GenerateFileID([]byte("rich-payer"))
		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create payer with large balance
		payerFile := &filestore.File{
			ID:         payerID,
			Balance:    2000000000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		_, err := fs.CreateFile(payerFile)
		if err != nil {
			t.Fatalf("Failed to create payer: %v", err)
		}

		// Create with very large balance
		instrData := encodeCreateFileInstruction(fileID, payerID, 1000000000, owner)

		inputs := map[string]transaction.FileAccess{
			"program":  {FileID: SystemProgramID, Permission: transaction.Read},
			"new_file": {FileID: fileID, Permission: transaction.Write},
			"payer":    {FileID: payerID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("CREATE_FILE with large balance failed: %v", err)
		}

		file, err := fs.GetFile(fileID)
		if err != nil {
			t.Fatalf("File not created: %v", err)
		}

		if file.Balance != 1000000000 {
			t.Errorf("Balance = %d, want 1000000000", file.Balance)
		}
	})
}

// TestSystemTransferEdgeCases tests edge cases for TRANSFER
func TestSystemTransferEdgeCases(t *testing.T) {
	t.Run("TransferZeroAmount", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fromID := filestore.GenerateFileID([]byte("from-zero"))
		toID := filestore.GenerateFileID([]byte("to-zero"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		fromFile := &filestore.File{
			ID:         fromID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		toFile := &filestore.File{
			ID:         toID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(fromFile)
		if err != nil {
			t.Fatalf("Failed to create from file: %v", err)
		}
		_, err = fs.CreateFile(toFile)
		if err != nil {
			t.Fatalf("Failed to create to file: %v", err)
		}

		instrData := encodeTransferInstruction(fromID, toID, 0)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"from":    {FileID: fromID, Permission: transaction.Write},
			"to":      {FileID: toID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected TRANSFER with zero amount to fail")
		}
	})

	t.Run("TransferExactBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fromID := filestore.GenerateFileID([]byte("from-exact"))
		toID := filestore.GenerateFileID([]byte("to-exact"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		fromFile := &filestore.File{
			ID:         fromID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		toFile := &filestore.File{
			ID:         toID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(fromFile)
		if err != nil {
			t.Fatalf("Failed to create from file: %v", err)
		}
		_, err = fs.CreateFile(toFile)
		if err != nil {
			t.Fatalf("Failed to create to file: %v", err)
		}

		// Transfer exact balance
		instrData := encodeTransferInstruction(fromID, toID, 10000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"from":    {FileID: fromID, Permission: transaction.Write},
			"to":      {FileID: toID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("TRANSFER exact balance failed: %v", err)
		}

		fromAfter, _ := fs.GetFile(fromID)
		toAfter, _ := fs.GetFile(toID)

		if fromAfter.Balance != 0 {
			t.Errorf("From balance = %d, want 0", fromAfter.Balance)
		}

		if toAfter.Balance != 15000 {
			t.Errorf("To balance = %d, want 15000", toAfter.Balance)
		}
	})

	t.Run("TransferWithStorageCostConstraint", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fromID := filestore.GenerateFileID([]byte("from-storage"))
		toID := filestore.GenerateFileID([]byte("to-storage"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		// Create from file with data requiring storage cost
		data := make([]byte, 2048) // 2KB
		storageCost := filestore.CalculateStorageCost(int64(len(data)))

		fromFile := &filestore.File{
			ID:         fromID,
			Balance:    storageCost + 5000,
			TxManager:  SystemProgramID,
			Data:       data,
			Executable: false,
		}
		toFile := &filestore.File{
			ID:         toID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(fromFile)
		if err != nil {
			t.Fatalf("Failed to create from file: %v", err)
		}
		_, err = fs.CreateFile(toFile)
		if err != nil {
			t.Fatalf("Failed to create to file: %v", err)
		}

		// Try to transfer more than available (would violate storage cost)
		instrData := encodeTransferInstruction(fromID, toID, 6000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"from":    {FileID: fromID, Permission: transaction.Write},
			"to":      {FileID: toID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected TRANSFER to fail when violating storage cost")
		}
	})
}

// TestSystemBurnEdgeCases tests edge cases for BURN
func TestSystemBurnEdgeCases(t *testing.T) {
	t.Run("BurnZeroAmount", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("burn-zero"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		instrData := encodeBurnInstruction(fileID, 0)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"file":    {FileID: fileID, Permission: transaction.Write},
			"system":  {FileID: SystemProgramID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected BURN with zero amount to fail")
		}
	})

	t.Run("BurnExactBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("burn-exact"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		instrData := encodeBurnInstruction(fileID, 10000)

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
			"file":    {FileID: fileID, Permission: transaction.Write},
			"system":  {FileID: SystemProgramID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("BURN exact balance failed: %v", err)
		}

		fileAfter, _ := fs.GetFile(fileID)

		if fileAfter.Balance != 0 {
			t.Errorf("Balance = %d, want 0", fileAfter.Balance)
		}
	})
}

// TestSystemCloseFileEdgeCases tests edge cases for CLOSE_FILE
func TestSystemCloseFileEdgeCases(t *testing.T) {
	t.Run("CloseFileToSelf", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("self-close"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    5000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Try to close file to itself
		instrData := encodeCloseFileInstruction(fileID, fileID)

		inputs := map[string]transaction.FileAccess{
			"program":     {FileID: SystemProgramID, Permission: transaction.Read},
			"file":        {FileID: fileID, Permission: transaction.Write},
			"destination": {FileID: fileID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		// This may succeed or fail depending on implementation
		// The test documents the behavior
		t.Logf("Close to self result: %v", err)
	})

	t.Run("CloseFileWithZeroBalance", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		fileID := filestore.GenerateFileID([]byte("zero-close"))
		destID := filestore.GenerateFileID([]byte("dest-zero"))

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		file := &filestore.File{
			ID:         fileID,
			Balance:    0,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}
		destFile := &filestore.File{
			ID:         destID,
			Balance:    10000,
			TxManager:  SystemProgramID,
			Data:       []byte{},
			Executable: false,
		}

		_, err := fs.CreateFile(file)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		_, err = fs.CreateFile(destFile)
		if err != nil {
			t.Fatalf("Failed to create dest file: %v", err)
		}

		instrData := encodeCloseFileInstruction(fileID, destID)

		inputs := map[string]transaction.FileAccess{
			"program":     {FileID: SystemProgramID, Permission: transaction.Read},
			"file":        {FileID: fileID, Permission: transaction.Write},
			"destination": {FileID: destID, Permission: transaction.Write},
		}

		err = executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err != nil {
			t.Fatalf("CLOSE_FILE with zero balance failed: %v", err)
		}

		// Verify file is deleted
		_, err = fs.GetFile(fileID)
		if err == nil {
			t.Error("File should be deleted")
		}

		// Verify destination balance unchanged
		destAfter, _ := fs.GetFile(destID)
		if destAfter.Balance != 10000 {
			t.Errorf("Destination balance = %d, want 10000", destAfter.Balance)
		}
	})
}

// TestSystemProgramErrorCodes verifies error code behavior
func TestSystemProgramErrorCodes(t *testing.T) {
	t.Run("InvalidInstructionOpcode", func(t *testing.T) {
		fs, bytecode := setupSystemProgramTest(t)

		// Create instruction with invalid opcode (99)
		instrData := []byte{99, 0, 0, 0}

		inputs := map[string]transaction.FileAccess{
			"program": {FileID: SystemProgramID, Permission: transaction.Read},
		}

		var owner transaction.PublicKey
		copy(owner[:], []byte("owner-pubkey-32-bytes-long!!"))

		err := executeSystemProgram(t, fs, bytecode, instrData, []transaction.PublicKey{owner}, inputs)
		if err == nil {
			t.Fatal("Expected invalid instruction to fail")
		}
		t.Logf("Invalid instruction error: %v", err)
	})
}

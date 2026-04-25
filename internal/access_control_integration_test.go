package internal

import (
	"crypto/ed25519"
	"encoding/binary"
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/programs"
)

// encodeTransferInstructionAC encodes a transfer instruction payload.
// Format: [type(1), amount_le(8), from_fileID(32), to_fileID(32)] = 73 bytes
func encodeTransferInstructionAC(amount int64, from, to filestore.FileID) []byte {
	data := make([]byte, 73)
	data[0] = 1 // Transfer opcode
	binary.LittleEndian.PutUint64(data[1:9], uint64(amount))
	copy(data[9:41], from[:])
	copy(data[41:73], to[:])
	return data
}

// TestReadPermissionEnforcement verifies that read-only permissions are enforced
// Requirements: 4.2, 8.2
func TestReadPermissionEnforcement(t *testing.T) {
	dbPath := t.TempDir() + "/read_perm_test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Load built-in programs via genesis
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
		t.Fatalf("Failed to load builtin programs: %v", err)
	}

	// Register a test program that attempts to write to a read-only file
	testProg := &readOnlyViolatorProgram{}
	err = rt.RegisterBuiltinProgram(testProg)
	if err != nil {
		t.Fatalf("Failed to register test program: %v", err)
	}

	testProgFile := &filestore.File{
		ID:         testProg.GetProgramID(),
		Balance:    0,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(testProgFile)
	if err != nil {
		t.Fatalf("Failed to create test program file: %v", err)
	}

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create test account
	testAccountID := filestore.GenerateFileID([]byte("test-account"))
	testAccountFile := &filestore.File{
		ID:         testAccountID,
		Balance:    10000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(testAccountFile)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	originalBalance := testAccountFile.Balance

	// Create instruction with READ-ONLY permission
	instr := transaction.Instruction{
		ProgramID: testProg.GetProgramID(),
		Inputs: map[string]transaction.FileAccess{
			"account": {FileID: testAccountID, Permission: transaction.Read},
		},
		Data: []byte{},
	}

	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{instr},
	}

	// Sign transaction
	txData, _ := tx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should fail due to write attempt on read-only file
	result, err := txProcessor.ProcessTransaction(tx)
	if err == nil {
		t.Error("Expected transaction to fail due to read permission violation")
	}

	if result != nil && result.Success {
		t.Error("Transaction should not succeed when violating read permission")
	}

	// Verify account balance is unchanged
	accountAfter, _ := fs.GetFile(testAccountID)
	if accountAfter.Balance != originalBalance {
		t.Errorf("Account balance changed despite permission violation: got %d, want %d",
			accountAfter.Balance, originalBalance)
	}

	t.Log("Read permission enforcement test passed: write attempt on read-only file was rejected")
}

// TestWritePermissionEnforcement verifies that write permissions allow both read and write
// Requirements: 4.2, 8.1
func TestWritePermissionEnforcement(t *testing.T) {
	dbPath := t.TempDir() + "/write_perm_test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Load built-in programs via genesis
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
		t.Fatalf("Failed to load builtin programs: %v", err)
	}

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create two accounts for transfer
	fromID := filestore.GenerateFileID([]byte("from-account"))
	toID := filestore.GenerateFileID([]byte("to-account"))

	fromFile := &filestore.File{
		ID:         fromID,
		Balance:    10000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	toFile := &filestore.File{
		ID:         toID,
		Balance:    5000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(fromFile)
	if err != nil {
		t.Fatalf("Failed to create from account: %v", err)
	}
	_, err = fs.CreateFile(toFile)
	if err != nil {
		t.Fatalf("Failed to create to account: %v", err)
	}

	// Create transfer instruction with WRITE permission on both accounts
	transferData := encodeTransferInstructionAC(1000, fromID, toID)
	instr := transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: fromID, Permission: transaction.Write},
			"to":   {FileID: toID, Permission: transaction.Write},
		},
		Data: transferData,
	}

	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{instr},
		Signatures:   []transaction.Signature{}, // Empty for signing
	}

	// Sign transaction
	txData, _ := tx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should succeed with write permissions
	result, err := txProcessor.ProcessTransaction(tx)
	if err != nil {
		t.Fatalf("Transaction failed with write permissions: %v", err)
	}

	if !result.Success {
		t.Fatalf("Transaction should succeed with write permissions: %v", result.Error)
	}

	// Verify balances were updated
	fromAfter, _ := fs.GetFile(fromID)
	toAfter, _ := fs.GetFile(toID)

	if fromAfter.Balance != 9000 {
		t.Errorf("From account balance = %d, want 9000", fromAfter.Balance)
	}

	if toAfter.Balance != 6000 {
		t.Errorf("To account balance = %d, want 6000", toAfter.Balance)
	}

	t.Log("Write permission enforcement test passed: write permissions allow both read and write")
}

// TestUndeclaredFileAccessDetection verifies that accessing undeclared files is rejected
// Requirements: 4.5, 8.1, 8.5
func TestUndeclaredFileAccessDetection(t *testing.T) {
	dbPath := t.TempDir() + "/undeclared_test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Load built-in programs via genesis
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
		t.Fatalf("Failed to load builtin programs: %v", err)
	}

	// Register a test program that accesses an undeclared file
	testProg := &undeclaredAccessProgram{}
	err = rt.RegisterBuiltinProgram(testProg)
	if err != nil {
		t.Fatalf("Failed to register test program: %v", err)
	}

	testProgFile := &filestore.File{
		ID:         testProg.GetProgramID(),
		Balance:    0,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(testProgFile)
	if err != nil {
		t.Fatalf("Failed to create test program file: %v", err)
	}

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create two accounts
	declaredID := filestore.GenerateFileID([]byte("declared-account"))
	undeclaredID := filestore.GenerateFileID([]byte("undeclared-account"))

	declaredFile := &filestore.File{
		ID:         declaredID,
		Balance:    10000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	undeclaredFile := &filestore.File{
		ID:         undeclaredID,
		Balance:    5000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(declaredFile)
	if err != nil {
		t.Fatalf("Failed to create declared account: %v", err)
	}
	_, err = fs.CreateFile(undeclaredFile)
	if err != nil {
		t.Fatalf("Failed to create undeclared account: %v", err)
	}

	originalDeclaredBalance := declaredFile.Balance
	originalUndeclaredBalance := undeclaredFile.Balance

	// Create instruction that only declares one account but program will try to access both
	instr := transaction.Instruction{
		ProgramID: testProg.GetProgramID(),
		Inputs: map[string]transaction.FileAccess{
			"declared": {FileID: declaredID, Permission: transaction.Write},
			// undeclaredID is NOT declared
		},
		Data: undeclaredID[:], // Pass undeclared ID in data
	}

	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{instr},
	}

	// Sign transaction
	txData, _ := tx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should fail due to undeclared file access
	result, err := txProcessor.ProcessTransaction(tx)
	if err == nil {
		t.Error("Expected transaction to fail due to undeclared file access")
	}

	if result != nil && result.Success {
		t.Error("Transaction should not succeed when accessing undeclared file")
	}

	// Verify both accounts are unchanged
	declaredAfter, _ := fs.GetFile(declaredID)
	undeclaredAfter, _ := fs.GetFile(undeclaredID)

	if declaredAfter.Balance != originalDeclaredBalance {
		t.Errorf("Declared account balance changed: got %d, want %d",
			declaredAfter.Balance, originalDeclaredBalance)
	}

	if undeclaredAfter.Balance != originalUndeclaredBalance {
		t.Errorf("Undeclared account balance changed: got %d, want %d",
			undeclaredAfter.Balance, originalUndeclaredBalance)
	}

	t.Log("Undeclared file access detection test passed: access to undeclared file was rejected")
}

// TestPermissionViolationHandling verifies that permission violations halt execution and revert state
// Requirements: 4.2, 8.2, 8.5
func TestPermissionViolationHandling(t *testing.T) {
	dbPath := t.TempDir() + "/perm_violation_test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Load built-in programs via genesis
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram); err != nil {
		t.Fatalf("Failed to load builtin programs: %v", err)
	}

	// Register a test program that violates permissions mid-execution
	testProg := &midExecutionViolatorProgram{}
	err = rt.RegisterBuiltinProgram(testProg)
	if err != nil {
		t.Fatalf("Failed to register test program: %v", err)
	}

	testProgFile := &filestore.File{
		ID:         testProg.GetProgramID(),
		Balance:    0,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(testProgFile)
	if err != nil {
		t.Fatalf("Failed to create test program file: %v", err)
	}

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create two test accounts
	account1ID := filestore.GenerateFileID([]byte("account1"))
	account2ID := filestore.GenerateFileID([]byte("account2"))

	account1File := &filestore.File{
		ID:         account1ID,
		Balance:    10000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	account2File := &filestore.File{
		ID:         account2ID,
		Balance:    5000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(account1File)
	if err != nil {
		t.Fatalf("Failed to create account1: %v", err)
	}
	_, err = fs.CreateFile(account2File)
	if err != nil {
		t.Fatalf("Failed to create account2: %v", err)
	}

	originalAccount1Balance := account1File.Balance
	originalAccount2Balance := account2File.Balance

	// Create instruction with write on account1, read on account2
	// Program will try to write to account2 (violation)
	instr := transaction.Instruction{
		ProgramID: testProg.GetProgramID(),
		Inputs: map[string]transaction.FileAccess{
			"account1": {FileID: account1ID, Permission: transaction.Write},
			"account2": {FileID: account2ID, Permission: transaction.Read}, // Read only
		},
		Data: []byte{},
	}

	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{instr},
	}

	// Sign transaction
	txData, _ := tx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should fail and revert all changes
	result, err := txProcessor.ProcessTransaction(tx)
	if err == nil {
		t.Error("Expected transaction to fail due to permission violation")
	}

	if result != nil && result.Success {
		t.Error("Transaction should not succeed when violating permissions")
	}

	// Verify all state changes are reverted
	account1After, _ := fs.GetFile(account1ID)
	account2After, _ := fs.GetFile(account2ID)

	if account1After.Balance != originalAccount1Balance {
		t.Errorf("Account1 balance changed despite violation: got %d, want %d",
			account1After.Balance, originalAccount1Balance)
	}

	if account2After.Balance != originalAccount2Balance {
		t.Errorf("Account2 balance changed despite violation: got %d, want %d",
			account2After.Balance, originalAccount2Balance)
	}

	t.Log("Permission violation handling test passed: state reverted on permission violation")
}

// Test helper programs

// readOnlyViolatorProgram attempts to write to a file declared as read-only
type readOnlyViolatorProgram struct{}

func (p *readOnlyViolatorProgram) Execute(ctx *runtime.ExecutionContext) error {
	accountID, err := ctx.GetInputFileID("account")
	if err != nil {
		return err
	}

	// Attempt to get mutable access (write) - should fail if declared as read-only
	_, err = ctx.GetFileMut(accountID)
	return err
}

func (p *readOnlyViolatorProgram) GetProgramID() filestore.FileID {
	return filestore.GenerateFileID([]byte("read-only-violator"))
}

// undeclaredAccessProgram attempts to access a file not declared in inputs
type undeclaredAccessProgram struct{}

func (p *undeclaredAccessProgram) Execute(ctx *runtime.ExecutionContext) error {
	// Access declared file first (should succeed)
	declaredID, err := ctx.GetInputFileID("declared")
	if err != nil {
		return err
	}

	_, err = ctx.GetFile(declaredID)
	if err != nil {
		return err
	}

	// Now try to access undeclared file (should fail)
	var undeclaredID filestore.FileID
	copy(undeclaredID[:], ctx.Instruction.Data)

	// This should fail because undeclaredID was not declared in inputs
	_, err = ctx.GetFile(undeclaredID)
	return err
}

func (p *undeclaredAccessProgram) GetProgramID() filestore.FileID {
	return filestore.GenerateFileID([]byte("undeclared-access"))
}

// midExecutionViolatorProgram modifies one file successfully then violates permission on another
type midExecutionViolatorProgram struct{}

func (p *midExecutionViolatorProgram) Execute(ctx *runtime.ExecutionContext) error {
	// First, modify account1 (should succeed - has write permission)
	account1ID, err := ctx.GetInputFileID("account1")
	if err != nil {
		return err
	}

	account1, err := ctx.GetFileMut(account1ID)
	if err != nil {
		return err
	}

	account1.Balance += 1000
	err = ctx.UpdateFile(account1)
	if err != nil {
		return err
	}

	// Now try to modify account2 (should fail - only has read permission)
	account2ID, err := ctx.GetInputFileID("account2")
	if err != nil {
		return err
	}

	// This should fail because account2 is declared as read-only
	_, err = ctx.GetFileMut(account2ID)
	return err
}

func (p *midExecutionViolatorProgram) GetProgramID() filestore.FileID {
	return filestore.GenerateFileID([]byte("mid-execution-violator"))
}

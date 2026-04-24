package processor

import (
	"crypto/ed25519"
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/system"
	"github.com/poh-blockchain/internal/transaction"
)

// TestNewTxProcessor verifies TxProcessor initialization
func TestNewTxProcessor(t *testing.T) {
	// Create temporary database
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	processor := NewTxProcessor(fs, rt)
	if processor == nil {
		t.Fatal("NewTxProcessor returned nil")
	}

	if processor.fileStore != fs {
		t.Error("FileStore not set correctly")
	}

	if processor.runtime != rt {
		t.Error("Runtime not set correctly")
	}

	config := processor.GetFeeConfig()
	if config.BaseFee != 5000 {
		t.Errorf("Expected default base fee 5000, got %d", config.BaseFee)
	}
}

// TestValidateTransaction verifies transaction validation logic
func TestValidateTransaction(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()
	processor := NewTxProcessor(fs, rt)

	// Generate test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := processor.publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create a valid transaction
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs:    map[string]transaction.FileAccess{},
				Data:      []byte{},
			},
		},
		Signatures: []transaction.Signature{}, // Initialize empty signatures array
	}

	// Marshal transaction for signing (with empty signatures)
	txData, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Sign transaction
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{
			PublicKey: pk,
			Signature: sigBytes,
		},
	}

	// Test: Valid transaction should pass
	err = processor.ValidateTransaction(tx)
	if err != nil {
		t.Errorf("Valid transaction failed validation: %v", err)
	}

	// Test: Transaction with no signatures should fail
	txNoSig := &transaction.Transaction{
		Instructions: tx.Instructions,
		Signatures:   []transaction.Signature{},
	}
	err = processor.ValidateTransaction(txNoSig)
	if err == nil {
		t.Error("Transaction with no signatures should fail validation")
	}

	// Test: Transaction with no instructions should fail
	txNoInstr := &transaction.Transaction{
		Instructions: []transaction.Instruction{},
		Signatures:   tx.Signatures,
	}
	err = processor.ValidateTransaction(txNoInstr)
	if err == nil {
		t.Error("Transaction with no instructions should fail validation")
	}
}

// TestDeductFee verifies fee deduction logic
func TestDeductFee(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()
	processor := NewTxProcessor(fs, rt)

	// Create account with balance
	accountID := filestore.GenerateFileID([]byte("test-account"))
	accountFile := &filestore.File{
		ID:         accountID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Test: Deduct valid fee
	err = processor.DeductFee(accountID, 5000)
	if err != nil {
		t.Errorf("Failed to deduct fee: %v", err)
	}

	// Verify balance in store
	updated, err := fs.GetFile(accountID)
	if err != nil {
		t.Fatal("Account not found in store after fee deduction")
	}
	if updated.Balance != 5000 {
		t.Errorf("Expected balance 5000 after fee deduction, got %d", updated.Balance)
	}

	// Test: Deduct fee exceeding balance should fail
	err = processor.DeductFee(accountID, 10000)
	if err == nil {
		t.Error("Deducting fee exceeding balance should fail")
	}

	// Test: Negative fee should fail
	err = processor.DeductFee(accountID, -100)
	if err == nil {
		t.Error("Negative fee should fail")
	}
}

// TestExecuteInstruction verifies instruction execution with access control
func TestExecuteInstruction(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Register system program
	sysProg := system.NewSystemProgram()
	err = rt.RegisterBuiltinProgram(sysProg)
	if err != nil {
		t.Fatalf("Failed to register system program: %v", err)
	}

	// Create system program file
	sysProgramFile := &filestore.File{
		ID:         system.SystemProgramID,
		Balance:    0,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(sysProgramFile)
	if err != nil {
		t.Fatalf("Failed to create system program file: %v", err)
	}

	processor := NewTxProcessor(fs, rt)

	// Generate test keypair
	pubKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create test accounts
	fromID := filestore.GenerateFileID([]byte("from-account"))
	fromFile := &filestore.File{
		ID:         fromID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(fromFile)
	if err != nil {
		t.Fatalf("Failed to create from account: %v", err)
	}

	toID := filestore.GenerateFileID([]byte("to-account"))
	toFile := &filestore.File{
		ID:         toID,
		Balance:    5000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(toFile)
	if err != nil {
		t.Fatalf("Failed to create to account: %v", err)
	}

	// Create transfer instruction
	transferData := system.EncodeTransferInstruction(1000)
	instr := &transaction.Instruction{
		ProgramID: system.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: fromID, Permission: transaction.Write},
			"to":   {FileID: toID, Permission: transaction.Write},
		},
		Data: transferData,
	}

	// Execute instruction
	signers := []transaction.PublicKey{pk}
	err = processor.ExecuteInstruction(instr, signers)
	if err != nil {
		t.Errorf("Failed to execute instruction: %v", err)
	}

	// Verify balances in store
	fromUpdated, err := fs.GetFile(fromID)
	if err != nil {
		t.Fatal("From account not found in store")
	}
	if fromUpdated.Balance != 9000 {
		t.Errorf("Expected from balance 9000, got %d", fromUpdated.Balance)
	}

	toUpdated, err := fs.GetFile(toID)
	if err != nil {
		t.Fatal("To account not found in store")
	}
	if toUpdated.Balance != 6000 {
		t.Errorf("Expected to balance 6000, got %d", toUpdated.Balance)
	}
}

// TestProcessTransaction verifies end-to-end transaction processing
func TestProcessTransaction(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Register system program
	sysProg := system.NewSystemProgram()
	err = rt.RegisterBuiltinProgram(sysProg)
	if err != nil {
		t.Fatalf("Failed to register system program: %v", err)
	}

	// Create system program file
	sysProgramFile := &filestore.File{
		ID:         system.SystemProgramID,
		Balance:    0,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(sysProgramFile)
	if err != nil {
		t.Fatalf("Failed to create system program file: %v", err)
	}

	processor := NewTxProcessor(fs, rt)

	// Generate test keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create fee payer account
	feePayerID := processor.publicKeyToFileID(pk)
	feePayerFile := &filestore.File{
		ID:         feePayerID,
		Balance:    100000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(feePayerFile)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}

	// Create recipient account
	recipientID := filestore.GenerateFileID([]byte("recipient"))
	recipientFile := &filestore.File{
		ID:         recipientID,
		Balance:    5000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(recipientFile)
	if err != nil {
		t.Fatalf("Failed to create recipient account: %v", err)
	}

	// Create transfer transaction
	transferData := system.EncodeTransferInstruction(1000)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: feePayerID, Permission: transaction.Write},
					"to":   {FileID: recipientID, Permission: transaction.Write},
				},
				Data: transferData,
			},
		},
		Signatures: []transaction.Signature{}, // Initialize empty signatures array
	}

	// Sign transaction (with empty signatures)
	txData, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	tx.Signatures = []transaction.Signature{
		{
			PublicKey: pk,
			Signature: sigBytes,
		},
	}

	// Process transaction
	result, err := processor.ProcessTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to process transaction: %v", err)
	}

	if !result.Success {
		t.Errorf("Transaction should succeed, got error: %v", result.Error)
	}

	// Verify balances were updated in store
	feePayerUpdated, err := fs.GetFile(feePayerID)
	if err != nil {
		t.Fatalf("Failed to get fee payer account: %v", err)
	}

	// Fee payer should have: 100000 - fee - 1000 (transfer)
	expectedFee := int64(5000 + 1000 + 500) // base + 1 instruction + 1 signature
	expectedBalance := 100000 - expectedFee - 1000
	if feePayerUpdated.Balance != expectedBalance {
		t.Errorf("Expected fee payer balance %d, got %d", expectedBalance, feePayerUpdated.Balance)
	}

	recipientUpdated, err := fs.GetFile(recipientID)
	if err != nil {
		t.Fatalf("Failed to get recipient account: %v", err)
	}

	if recipientUpdated.Balance != 6000 {
		t.Errorf("Expected recipient balance 6000, got %d", recipientUpdated.Balance)
	}
}

// TestRevertState verifies state rollback on errors
func TestRevertState(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()
	processor := NewTxProcessor(fs, rt)

	// Create test account
	accountID := filestore.GenerateFileID([]byte("test-account"))
	accountFile := &filestore.File{
		ID:         accountID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Initialize state cache and modify balance
	processor.stateCache = make(map[filestore.FileID]*filestore.File)
	modifiedFile := *accountFile
	modifiedFile.Balance = 5000
	processor.stateCache[accountID] = &modifiedFile

	// Verify cache has modified balance
	if processor.stateCache[accountID].Balance != 5000 {
		t.Error("Cache should have modified balance")
	}

	// Revert state
	processor.RevertState()

	// Verify cache is cleared
	if len(processor.stateCache) != 0 {
		t.Error("State cache should be empty after revert")
	}

	// Verify original balance in store is unchanged
	originalFile, err := fs.GetFile(accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if originalFile.Balance != 10000 {
		t.Errorf("Original balance should be unchanged, got %d", originalFile.Balance)
	}
}

// TestAccessControlValidation verifies access control enforcement
func TestAccessControlValidation(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	rt := runtime.NewRuntime()

	// Create a test builtin program that violates access control
	violatingProgram := &testViolatingProgram{}
	err = rt.RegisterBuiltinProgram(violatingProgram)
	if err != nil {
		t.Fatalf("Failed to register test program: %v", err)
	}

	// Create program file
	programFile := &filestore.File{
		ID:         violatingProgram.GetProgramID(),
		Balance:    0,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: true,
	}
	_, err = fs.CreateFile(programFile)
	if err != nil {
		t.Fatalf("Failed to create program file: %v", err)
	}

	processor := NewTxProcessor(fs, rt)
	processor.stateCache = make(map[filestore.FileID]*filestore.File)

	// Create test account
	accountID := filestore.GenerateFileID([]byte("test-account"))
	accountFile := &filestore.File{
		ID:         accountID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(accountFile)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create instruction with read-only permission
	instr := &transaction.Instruction{
		ProgramID: violatingProgram.GetProgramID(),
		Inputs: map[string]transaction.FileAccess{
			"account": {FileID: accountID, Permission: transaction.Read},
		},
		Data: []byte{},
	}

	// Execute instruction - should fail due to write attempt with read-only permission
	err = processor.ExecuteInstruction(instr, []transaction.PublicKey{})
	if err == nil {
		t.Error("Instruction should fail due to access control violation")
	}
}

// testViolatingProgram is a test program that attempts to write to a read-only file
type testViolatingProgram struct{}

func (p *testViolatingProgram) Execute(ctx *runtime.ExecutionContext) error {
	// Attempt to get file with write permission (should fail if declared as read-only)
	accountID, err := ctx.GetInputFileID("account")
	if err != nil {
		return err
	}

	// This should fail if the file was declared with read-only permission
	_, err = ctx.GetFileMut(accountID)
	return err
}

func (p *testViolatingProgram) GetProgramID() filestore.FileID {
	return filestore.GenerateFileID([]byte("test-violating-program"))
}

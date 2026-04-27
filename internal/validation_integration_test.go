package internal

import (
	"crypto/ed25519"
	"encoding/binary"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/programs"
)

// encodeTransferInstructionValidation encodes a transfer instruction payload.
// Format: [type(1), from_fileID(32), to_fileID(32), amount_le(8)] = 73 bytes
func encodeTransferInstructionValidation(amount int64, from, to filestore.FileID) []byte {
	data := make([]byte, 73)
	data[0] = 1 // Transfer opcode
	copy(data[1:33], from[:])
	copy(data[33:65], to[:])
	binary.LittleEndian.PutUint64(data[65:73], uint64(amount))
	return data
}

// setupValidationTestEnv creates a test environment with FileStore, Runtime, and TxProcessor
func setupValidationTestEnv(t *testing.T) (*filestore.FileStore, *runtime.Runtime, *processor.TxProcessor) {
	t.Helper()
	dbPath := t.TempDir() + "/validation_test.db"
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	t.Cleanup(func() { fs.Close() })

	// Load built-in programs via genesis
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram, nil); err != nil {
		t.Fatalf("Failed to load builtin programs: %v", err)
	}

	rt := runtime.NewRuntime()
	tp := processor.NewTxProcessor(fs, rt)
	return fs, rt, tp
}

// createTestAccount creates a test account with the given balance
func createTestAccount(t *testing.T, fs *filestore.FileStore, id filestore.FileID, balance int64) {
	t.Helper()
	account := &filestore.File{
		ID:         id,
		Balance:    balance,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	if _, err := fs.CreateFile(account); err != nil {
		t.Fatalf("Failed to create account %s: %v", id.String(), err)
	}
}

// signTransaction signs a transaction with the given private key
func signTransaction(t *testing.T, tx *transaction.Transaction, pk transaction.PublicKey, priv ed25519.PrivateKey) {
	t.Helper()
	saved := tx.Signatures
	tx.Signatures = []transaction.Signature{}
	data, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction for signing: %v", err)
	}
	tx.Signatures = saved
	raw := ed25519.Sign(priv, data)
	var sig [64]byte
	copy(sig[:], raw)
	tx.Signatures = append(tx.Signatures, transaction.Signature{PublicKey: pk, Signature: sig})
}

// TestEndToEndTransferWithProperInputDeclarations tests a complete transfer flow
// with all inputs properly declared (Requirements 1.1-1.5, 2.1-2.5, 3.1-3.5)
func TestEndToEndTransferWithProperInputDeclarations(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypairs
	alicePub, alicePriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate Alice's keypair: %v", err)
	}
	bobPub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate Bob's keypair: %v", err)
	}

	var alicePK, bobPK transaction.PublicKey
	copy(alicePK[:], alicePub)
	copy(bobPK[:], bobPub)

	// Create accounts
	aliceID := validationPKToFileID(alicePK)
	bobID := validationPKToFileID(bobPK)

	createTestAccount(t, fs, aliceID, 100_000)
	createTestAccount(t, fs, bobID, 5_000)

	// Build transfer transaction with proper input declarations
	transferData := encodeTransferInstructionValidation(10_000, aliceID, bobID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
				"from":    {FileID: aliceID, Permission: transaction.Write},
				"to":      {FileID: bobID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, alicePK, alicePriv)

	// Get balances before transfer
	aliceBefore, _ := fs.GetFile(aliceID)
	bobBefore, _ := fs.GetFile(bobID)

	// Process transaction
	result, err := tp.ProcessTransaction(tx)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("Transfer should succeed, got error: %v", result.Error)
	}

	// Verify balances after transfer
	aliceAfter, _ := fs.GetFile(aliceID)
	bobAfter, _ := fs.GetFile(bobID)

	fee := transaction.CalculateFeeWithDefaults(tx)
	expectedAlice := aliceBefore.Balance - fee - 10_000
	if aliceAfter.Balance != expectedAlice {
		t.Errorf("Alice balance = %d, want %d", aliceAfter.Balance, expectedAlice)
	}
	if bobAfter.Balance != bobBefore.Balance+10_000 {
		t.Errorf("Bob balance = %d, want %d", bobAfter.Balance, bobBefore.Balance+10_000)
	}

	t.Logf("End-to-end transfer succeeded: Alice=%d, Bob=%d, Fee=%d",
		aliceAfter.Balance, bobAfter.Balance, fee)
}

// TestTransactionRejectionMissingProgramDeclaration verifies that transactions
// without program declaration are rejected (Requirements 1.1, 1.2, 5.1, 5.2, 7.1, 7.2)
func TestTransactionRejectionMissingProgramDeclaration(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create accounts
	senderID := filestore.GenerateFileID([]byte("sender-missing-program"))
	recipientID := filestore.GenerateFileID([]byte("recipient-missing-program"))

	createTestAccount(t, fs, senderID, 50_000)
	createTestAccount(t, fs, recipientID, 10_000)

	// Build transaction WITHOUT program declaration
	transferData := encodeTransferInstructionValidation(5_000, senderID, recipientID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				// Missing "program" declaration!
				"from": {FileID: senderID, Permission: transaction.Write},
				"to":   {FileID: recipientID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, pk, priv)

	// Get balances before transaction
	senderBefore, _ := fs.GetFile(senderID)
	recipientBefore, _ := fs.GetFile(recipientID)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail due to missing program declaration")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify balances are unchanged (except fee payer)
	senderAfter, _ := fs.GetFile(senderID)
	recipientAfter, _ := fs.GetFile(recipientID)

	// Recipient should be completely unchanged
	if recipientAfter.Balance != recipientBefore.Balance {
		t.Errorf("Recipient balance changed on failed tx: before=%d after=%d",
			recipientBefore.Balance, recipientAfter.Balance)
	}

	// Sender may have fee deducted (current implementation behavior)
	// but transfer amount should not be deducted
	if senderAfter.Balance < senderBefore.Balance-10_000 {
		t.Errorf("Sender balance decreased more than fee: before=%d after=%d",
			senderBefore.Balance, senderAfter.Balance)
	}

	t.Log("Transaction correctly rejected due to missing program declaration")
}

// TestTransactionRejectionIncorrectPermissions verifies that transactions
// with incorrect permissions are rejected (Requirements 3.1-3.5, 5.3, 7.3)
func TestTransactionRejectionIncorrectPermissions(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create accounts
	senderID := filestore.GenerateFileID([]byte("sender-wrong-perm"))
	recipientID := filestore.GenerateFileID([]byte("recipient-wrong-perm"))

	createTestAccount(t, fs, senderID, 50_000)
	createTestAccount(t, fs, recipientID, 10_000)

	// Build transaction with sender declared as Read-only (should be Write)
	transferData := encodeTransferInstructionValidation(5_000, senderID, recipientID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
				"from":    {FileID: senderID, Permission: transaction.Read}, // Wrong! Should be Write
				"to":      {FileID: recipientID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, pk, priv)

	// Get balances before transaction
	senderBefore, _ := fs.GetFile(senderID)
	recipientBefore, _ := fs.GetFile(recipientID)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail due to incorrect permissions")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify balances are unchanged (except possible fee)
	senderAfter, _ := fs.GetFile(senderID)
	recipientAfter, _ := fs.GetFile(recipientID)

	// Recipient should be completely unchanged
	if recipientAfter.Balance != recipientBefore.Balance {
		t.Errorf("Recipient balance changed on failed tx: before=%d after=%d",
			recipientBefore.Balance, recipientAfter.Balance)
	}

	// Sender may have fee deducted but transfer should not happen
	if senderAfter.Balance < senderBefore.Balance-10_000 {
		t.Errorf("Sender balance decreased more than fee: before=%d after=%d",
			senderBefore.Balance, senderAfter.Balance)
	}

	t.Log("Transaction correctly rejected due to incorrect permissions")
}

// TestTransactionRejectionNonExistentFiles verifies that transactions
// referencing non-existent files are rejected (Requirements 1.3, 5.4, 7.4)
func TestTransactionRejectionNonExistentFiles(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create only sender account
	senderID := filestore.GenerateFileID([]byte("sender-nonexistent"))
	nonExistentID := filestore.GenerateFileID([]byte("non-existent-recipient"))

	createTestAccount(t, fs, senderID, 50_000)
	// Don't create recipient account

	// Build transaction with non-existent recipient
	transferData := encodeTransferInstructionValidation(5_000, senderID, nonExistentID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
				"from":    {FileID: senderID, Permission: transaction.Write},
				"to":      {FileID: nonExistentID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, pk, priv)

	// Get sender balance before transaction
	senderBefore, _ := fs.GetFile(senderID)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail due to non-existent file")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify sender balance (may have fee deducted)
	senderAfter, _ := fs.GetFile(senderID)

	// Transfer should not happen
	if senderAfter.Balance < senderBefore.Balance-10_000 {
		t.Errorf("Sender balance decreased more than fee: before=%d after=%d",
			senderBefore.Balance, senderAfter.Balance)
	}

	// Verify non-existent file still doesn't exist
	_, err = fs.GetFile(nonExistentID)
	if err == nil {
		t.Error("Non-existent file should not be created")
	}

	t.Log("Transaction correctly rejected due to non-existent file")
}

// TestTransactionRejectionNonExecutableProgram verifies that transactions
// invoking non-executable programs are rejected (Requirements 2.3, 5.4, 7.5)
func TestTransactionRejectionNonExecutableProgram(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create a NON-EXECUTABLE program file
	nonExecProgramID := filestore.GenerateFileID([]byte("non-executable-program"))
	nonExecProgram := &filestore.File{
		ID:         nonExecProgramID,
		Balance:    10_000,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{1, 2, 3}, // Some data
		Executable: false,           // NOT executable
	}
	if _, err := fs.CreateFile(nonExecProgram); err != nil {
		t.Fatalf("Failed to create non-executable program: %v", err)
	}

	// Create test accounts
	senderID := filestore.GenerateFileID([]byte("sender-nonexec"))
	recipientID := filestore.GenerateFileID([]byte("recipient-nonexec"))

	createTestAccount(t, fs, senderID, 50_000)
	createTestAccount(t, fs, recipientID, 10_000)

	// Build transaction invoking non-executable program
	transferData := encodeTransferInstructionValidation(5_000, senderID, recipientID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: nonExecProgramID, // Non-executable program
			Inputs: map[string]transaction.FileAccess{
				"program": {FileID: nonExecProgramID, Permission: transaction.Read},
				"from":    {FileID: senderID, Permission: transaction.Write},
				"to":      {FileID: recipientID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, pk, priv)

	// Get balances before transaction
	senderBefore, _ := fs.GetFile(senderID)
	recipientBefore, _ := fs.GetFile(recipientID)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail due to non-executable program")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify balances are unchanged (except possible fee)
	senderAfter, _ := fs.GetFile(senderID)
	recipientAfter, _ := fs.GetFile(recipientID)

	// Recipient should be completely unchanged
	if recipientAfter.Balance != recipientBefore.Balance {
		t.Errorf("Recipient balance changed on failed tx: before=%d after=%d",
			recipientBefore.Balance, recipientAfter.Balance)
	}

	// Sender may have fee deducted but transfer should not happen
	if senderAfter.Balance < senderBefore.Balance-10_000 {
		t.Errorf("Sender balance decreased more than fee: before=%d after=%d",
			senderBefore.Balance, senderAfter.Balance)
	}

	t.Log("Transaction correctly rejected due to non-executable program")
}

// TestNoStateChangesOnValidationFailure verifies that no state changes occur
// when validation fails (Requirements 5.1-5.5, 7.1-7.5)
func TestNoStateChangesOnValidationFailure(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create accounts
	account1ID := filestore.GenerateFileID([]byte("account1-nochange"))
	account2ID := filestore.GenerateFileID([]byte("account2-nochange"))
	account3ID := filestore.GenerateFileID([]byte("account3-nochange"))

	createTestAccount(t, fs, account1ID, 30_000)
	createTestAccount(t, fs, account2ID, 20_000)
	createTestAccount(t, fs, account3ID, 10_000)

	// Get initial balances
	acc1Before, _ := fs.GetFile(account1ID)
	acc2Before, _ := fs.GetFile(account2ID)
	acc3Before, _ := fs.GetFile(account3ID)

	// Build transaction with missing program declaration (will fail validation)
	transferData := encodeTransferInstructionValidation(5_000, account1ID, account2ID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				// Missing program declaration!
				"from": {FileID: account1ID, Permission: transaction.Write},
				"to":   {FileID: account2ID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTransaction(t, tx, pk, priv)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail validation")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify all accounts (except possible fee payer)
	acc1After, _ := fs.GetFile(account1ID)
	acc2After, _ := fs.GetFile(account2ID)
	acc3After, _ := fs.GetFile(account3ID)

	// Account2 and Account3 should be completely unchanged
	if acc2After.Balance != acc2Before.Balance {
		t.Errorf("Account2 balance changed: before=%d after=%d",
			acc2Before.Balance, acc2After.Balance)
	}
	if acc3After.Balance != acc3Before.Balance {
		t.Errorf("Account3 balance changed: before=%d after=%d",
			acc3Before.Balance, acc3After.Balance)
	}

	// Account1 may have fee deducted but transfer should not happen
	if acc1After.Balance < acc1Before.Balance-10_000 {
		t.Errorf("Account1 balance decreased more than fee: before=%d after=%d",
			acc1Before.Balance, acc1After.Balance)
	}

	t.Log("No state changes occurred on validation failure (except fee)")
}

// TestMultiInstructionTransactionValidation verifies that multi-instruction
// transactions are validated correctly (Requirements 1.1-1.5, 5.1-5.5)
func TestMultiInstructionTransactionValidation(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair for fee payer
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create fee payer account
	feePayerID := validationPKToFileID(pk)
	createTestAccount(t, fs, feePayerID, 100_000)

	// Create accounts
	account1ID := filestore.GenerateFileID([]byte("multi1"))
	account2ID := filestore.GenerateFileID([]byte("multi2"))
	account3ID := filestore.GenerateFileID([]byte("multi3"))

	createTestAccount(t, fs, account1ID, 100_000)
	createTestAccount(t, fs, account2ID, 50_000)
	createTestAccount(t, fs, account3ID, 25_000)

	// Build multi-instruction transaction with proper declarations
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{
			{
				ProgramID: genesis.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
					"from":    {FileID: account1ID, Permission: transaction.Write},
					"to":      {FileID: account2ID, Permission: transaction.Write},
				},
				Data: encodeTransferInstructionValidation(10_000, account1ID, account2ID),
			},
			{
				ProgramID: genesis.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
					"from":    {FileID: account2ID, Permission: transaction.Write},
					"to":      {FileID: account3ID, Permission: transaction.Write},
				},
				Data: encodeTransferInstructionValidation(5_000, account2ID, account3ID),
			},
		},
	}
	signTransaction(t, tx, pk, priv)

	// Get balances before transaction
	feePayerBefore, _ := fs.GetFile(feePayerID)
	acc1Before, _ := fs.GetFile(account1ID)
	acc2Before, _ := fs.GetFile(account2ID)
	acc3Before, _ := fs.GetFile(account3ID)

	// Process transaction - should succeed
	result, err := tp.ProcessTransaction(tx)
	if err != nil {
		t.Fatalf("Multi-instruction transaction failed: %v", err)
	}
	if !result.Success {
		t.Fatalf("Multi-instruction transaction should succeed, got error: %v", result.Error)
	}

	// Verify balances after transaction
	feePayerAfter, _ := fs.GetFile(feePayerID)
	acc1After, _ := fs.GetFile(account1ID)
	acc2After, _ := fs.GetFile(account2ID)
	acc3After, _ := fs.GetFile(account3ID)

	fee := transaction.CalculateFeeWithDefaults(tx)

	// FeePayer: -fee
	expectedFeePayer := feePayerBefore.Balance - fee
	if feePayerAfter.Balance != expectedFeePayer {
		t.Errorf("FeePayer balance = %d, want %d", feePayerAfter.Balance, expectedFeePayer)
	}

	// Account1: -10000
	expectedAcc1 := acc1Before.Balance - 10_000
	if acc1After.Balance != expectedAcc1 {
		t.Errorf("Account1 balance = %d, want %d", acc1After.Balance, expectedAcc1)
	}

	// Account2: +10000 -5000 = +5000
	expectedAcc2 := acc2Before.Balance + 5_000
	if acc2After.Balance != expectedAcc2 {
		t.Errorf("Account2 balance = %d, want %d", acc2After.Balance, expectedAcc2)
	}

	// Account3: +5000
	expectedAcc3 := acc3Before.Balance + 5_000
	if acc3After.Balance != expectedAcc3 {
		t.Errorf("Account3 balance = %d, want %d", acc3After.Balance, expectedAcc3)
	}

	t.Logf("Multi-instruction transaction succeeded: FeePayer=%d, Acc1=%d, Acc2=%d, Acc3=%d, Fee=%d",
		feePayerAfter.Balance, acc1After.Balance, acc2After.Balance, acc3After.Balance, fee)
}

// TestMultiInstructionTransactionPartialValidationFailure verifies that if one
// instruction fails validation, the entire transaction is rejected
// (Requirements 5.1-5.5, 7.1-7.5)
func TestMultiInstructionTransactionPartialValidationFailure(t *testing.T) {
	fs, _, tp := setupValidationTestEnv(t)

	// Generate keypair for fee payer
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Create fee payer account
	feePayerID := validationPKToFileID(pk)
	createTestAccount(t, fs, feePayerID, 100_000)

	// Create accounts
	account1ID := filestore.GenerateFileID([]byte("partial1"))
	account2ID := filestore.GenerateFileID([]byte("partial2"))
	account3ID := filestore.GenerateFileID([]byte("partial3"))

	createTestAccount(t, fs, account1ID, 100_000)
	createTestAccount(t, fs, account2ID, 50_000)
	createTestAccount(t, fs, account3ID, 25_000)

	// Get balances before transaction
	feePayerBefore, _ := fs.GetFile(feePayerID)
	acc1Before, _ := fs.GetFile(account1ID)
	acc2Before, _ := fs.GetFile(account2ID)
	acc3Before, _ := fs.GetFile(account3ID)

	// Build multi-instruction transaction where second instruction has missing program
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{
			{
				ProgramID: genesis.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"program": {FileID: genesis.SystemProgramID, Permission: transaction.Read},
					"from":    {FileID: account1ID, Permission: transaction.Write},
					"to":      {FileID: account2ID, Permission: transaction.Write},
				},
				Data: encodeTransferInstructionValidation(10_000, account1ID, account2ID),
			},
			{
				ProgramID: genesis.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					// Missing program declaration!
					"from": {FileID: account2ID, Permission: transaction.Write},
					"to":   {FileID: account3ID, Permission: transaction.Write},
				},
				Data: encodeTransferInstructionValidation(5_000, account2ID, account3ID),
			},
		},
	}
	signTransaction(t, tx, pk, priv)

	// Process transaction - should fail
	result, err := tp.ProcessTransaction(tx)
	if err == nil {
		t.Error("Transaction should fail due to validation error in second instruction")
	}
	if result != nil && result.Success {
		t.Error("Transaction result should indicate failure")
	}

	// Verify balances (except possible fee payer)
	feePayerAfter, _ := fs.GetFile(feePayerID)
	acc1After, _ := fs.GetFile(account1ID)
	acc2After, _ := fs.GetFile(account2ID)
	acc3After, _ := fs.GetFile(account3ID)

	// All accounts should be unchanged
	if acc1After.Balance != acc1Before.Balance {
		t.Errorf("Account1 balance changed: before=%d after=%d",
			acc1Before.Balance, acc1After.Balance)
	}
	if acc2After.Balance != acc2Before.Balance {
		t.Errorf("Account2 balance changed: before=%d after=%d",
			acc2Before.Balance, acc2After.Balance)
	}
	if acc3After.Balance != acc3Before.Balance {
		t.Errorf("Account3 balance changed: before=%d after=%d",
			acc3Before.Balance, acc3After.Balance)
	}

	// Fee payer may have fee deducted
	if feePayerAfter.Balance < feePayerBefore.Balance-20_000 {
		t.Errorf("FeePayer balance decreased more than fee: before=%d after=%d",
			feePayerBefore.Balance, feePayerAfter.Balance)
	}

	t.Log("Multi-instruction transaction correctly rejected on partial validation failure")
}

// validationPKToFileID converts a public key to a file ID
func validationPKToFileID(pk transaction.PublicKey) filestore.FileID {
	var id filestore.FileID
	copy(id[:], pk[:])
	return id
}

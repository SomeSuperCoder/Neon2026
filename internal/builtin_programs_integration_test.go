package internal

import (
	"crypto/ed25519"
	"encoding/binary"
	"testing"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
	"github.com/poh-blockchain/programs"
)

// encodeTransferInstruction encodes a transfer instruction payload.
// Format: [type(1), from_fileID(32), to_fileID(32), amount_le(8)] = 73 bytes
func encodeTransferInstruction(amount int64, from, to filestore.FileID) []byte {
	data := make([]byte, 73)
	data[0] = 1 // Transfer opcode
	copy(data[1:33], from[:])
	copy(data[33:65], to[:])
	binary.LittleEndian.PutUint64(data[65:73], uint64(amount))
	return data
}

// setupBuiltinEnv creates a fresh FileStore with both builtin programs loaded.
func setupBuiltinEnv(t *testing.T) (*filestore.FileStore, *runtime.Runtime, *processor.TxProcessor) {
	t.Helper()
	dbPath := t.TempDir() + "/builtin_test.db"
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	t.Cleanup(func() { fs.Close() })

	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram, nil); err != nil {
		t.Fatalf("LoadBuiltinPrograms: %v", err)
	}

	rt := runtime.NewRuntime()
	tp := processor.NewTxProcessor(fs, rt)
	return fs, rt, tp
}

// makeAccount creates a user account file directly in the FileStore.
func makeAccount(t *testing.T, fs *filestore.FileStore, id filestore.FileID, balance int64) {
	t.Helper()
	f := &filestore.File{
		ID:         id,
		Balance:    balance,
		TxManager:  genesis.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	if _, err := fs.CreateFile(f); err != nil {
		t.Fatalf("makeAccount %s: %v", id.String(), err)
	}
}

// signTx signs a transaction with the given private key and appends the signature.
// It follows the same pattern as the TxProcessor's ValidateTransaction: clear
// signatures, marshal, sign, then restore + append.
func signTx(t *testing.T, tx *transaction.Transaction, pk transaction.PublicKey, priv ed25519.PrivateKey) {
	t.Helper()
	saved := tx.Signatures
	tx.Signatures = []transaction.Signature{}
	data, err := tx.Marshal()
	if err != nil {
		t.Fatalf("marshal for signing: %v", err)
	}
	tx.Signatures = saved
	raw := ed25519.Sign(priv, data)
	var sig [64]byte
	copy(sig[:], raw)
	tx.Signatures = append(tx.Signatures, transaction.Signature{PublicKey: pk, Signature: sig})
}

// TestGenesisLoadsBothPrograms verifies that LoadBuiltinPrograms stores both
// programs as executable files at their well-known IDs (Requirements 3.6, 3.7).
func TestGenesisLoadsBothPrograms(t *testing.T) {
	fs, _, _ := setupBuiltinEnv(t)

	for name, id := range map[string]filestore.FileID{
		"System_Program": genesis.SystemProgramID,
		"Token_Program":  genesis.TokenProgramID,
	} {
		f, err := fs.GetFile(id)
		if err != nil {
			t.Errorf("%s not found after genesis: %v", name, err)
			continue
		}
		if !f.Executable {
			t.Errorf("%s should be executable", name)
		}
		if len(f.Data) == 0 {
			t.Errorf("%s bytecode is empty", name)
		}
		if f.TxManager != genesis.RuntimeProgramID {
			t.Errorf("%s TxManager should be RuntimeProgramID", name)
		}
		if f.Balance <= 0 {
			t.Errorf("%s balance should cover storage rent, got %d", name, f.Balance)
		}
	}
}

// TestGenesisIdempotent verifies that calling LoadBuiltinPrograms twice does not
// fail or duplicate the programs (Requirements 3.6, 3.7).
func TestGenesisIdempotent(t *testing.T) {
	fs, _, _ := setupBuiltinEnv(t)

	// Second call must not error
	if err := genesis.LoadBuiltinPrograms(fs, programs.SystemProgram, programs.TokenProgram, nil); err != nil {
		t.Fatalf("second LoadBuiltinPrograms call failed: %v", err)
	}

	// Programs should still be present and unchanged
	f, err := fs.GetFile(genesis.SystemProgramID)
	if err != nil {
		t.Fatalf("System_Program missing after second load: %v", err)
	}
	if !f.Executable {
		t.Error("System_Program should still be executable")
	}
}

// TestSystemProgramCreateAccountAndTransfer tests the System_Program CreateAccount
// and Transfer instructions end-to-end through the TxProcessor
// (Requirements 1.1, 1.2, 1.4).
func TestSystemProgramCreateAccountAndTransfer(t *testing.T) {
	fs, _, tp := setupBuiltinEnv(t)

	// Generate keypairs
	alicePub, alicePriv, _ := ed25519.GenerateKey(nil)
	bobPub, _, _ := ed25519.GenerateKey(nil)

	var alicePK, bobPK transaction.PublicKey
	copy(alicePK[:], alicePub)
	copy(bobPK[:], bobPub)

	aliceID := builtinPKToFileID(alicePK)
	bobID := builtinPKToFileID(bobPK)

	// Pre-fund Alice and Bob directly (simulating prior account creation)
	makeAccount(t, fs, aliceID, 100_000)
	makeAccount(t, fs, bobID, 0)

	// Build a Transfer transaction: Alice → Bob, 10 000 Neon
	transferData := encodeTransferInstruction(10_000, aliceID, bobID)
	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				"from": {FileID: aliceID, Permission: transaction.Write},
				"to":   {FileID: bobID, Permission: transaction.Write},
			},
			Data: transferData,
		}},
	}
	signTx(t, tx, alicePK, alicePriv)

	aliceBefore, _ := fs.GetFile(aliceID)
	result, err := tp.ProcessTransaction(tx)
	if err != nil || !result.Success {
		t.Fatalf("Transfer failed: err=%v result=%+v", err, result)
	}

	aliceAfter, _ := fs.GetFile(aliceID)
	bobAfter, _ := fs.GetFile(bobID)

	fee := transaction.CalculateFeeWithDefaults(tx)
	expectedAlice := aliceBefore.Balance - fee - 10_000
	if aliceAfter.Balance != expectedAlice {
		t.Errorf("Alice balance = %d, want %d", aliceAfter.Balance, expectedAlice)
	}
	if bobAfter.Balance != 10_000 {
		t.Errorf("Bob balance = %d, want 10000", bobAfter.Balance)
	}
}

// TestSystemProgramTransferInsufficientBalance verifies that a transfer with
// insufficient funds is rejected (Requirement 1.4).
func TestSystemProgramTransferInsufficientBalance(t *testing.T) {
	fs, _, tp := setupBuiltinEnv(t)

	pub, priv, _ := ed25519.GenerateKey(nil)
	var pk transaction.PublicKey
	copy(pk[:], pub)

	// Use a separate fee-payer so the sender state is cleanly tracked
	feePub, feePriv, _ := ed25519.GenerateKey(nil)
	var feePK transaction.PublicKey
	copy(feePK[:], feePub)

	feePayerID := builtinPKToFileID(feePK)
	senderID := filestore.GenerateFileID([]byte("sender-insufficient"))
	recipientID := filestore.GenerateFileID([]byte("recipient-insufficient"))

	makeAccount(t, fs, feePayerID, 50_000)
	makeAccount(t, fs, senderID, 1_000)
	makeAccount(t, fs, recipientID, 0)

	tx := &transaction.Transaction{
		Instructions: []transaction.Instruction{{
			ProgramID: genesis.SystemProgramID,
			Inputs: map[string]transaction.FileAccess{
				"from": {FileID: senderID, Permission: transaction.Write},
				"to":   {FileID: recipientID, Permission: transaction.Write},
			},
			Data: encodeTransferInstruction(999_999_999, senderID, recipientID),
		}},
	}
	// Fee payer signs first, sender signs second
	signTx(t, tx, feePK, feePriv)
	signTx(t, tx, pk, priv)

	senderBefore, _ := fs.GetFile(senderID)
	recipientBefore, _ := fs.GetFile(recipientID)

	result, _ := tp.ProcessTransaction(tx)
	if result != nil && result.Success {
		t.Fatal("Expected transfer to fail due to insufficient balance")
	}

	// Sender and recipient state must be unchanged
	senderAfter, _ := fs.GetFile(senderID)
	recipientAfter, _ := fs.GetFile(recipientID)
	if senderAfter.Balance != senderBefore.Balance {
		t.Errorf("Sender balance changed on failed tx: before=%d after=%d",
			senderBefore.Balance, senderAfter.Balance)
	}
	if recipientAfter.Balance != recipientBefore.Balance {
		t.Errorf("Recipient balance changed on failed tx: before=%d after=%d",
			recipientBefore.Balance, recipientAfter.Balance)
	}
}

// TestStorageRentEnforcement verifies that a file cannot be created with a
// balance that does not cover its data storage cost (Requirements 2.1, 2.2, 2.5).
func TestStorageRentEnforcement(t *testing.T) {
	dbPath := t.TempDir() + "/rent_test.db"
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	defer fs.Close()

	// 10 KB of data requires more than 0 balance
	largeData := make([]byte, 10*1024)
	f := &filestore.File{
		ID:         filestore.GenerateFileID([]byte("large-file")),
		Balance:    0, // Intentionally insufficient
		TxManager:  filestore.FileID{},
		Data:       largeData,
		Executable: false,
	}

	_, err = fs.CreateFile(f)
	if err == nil {
		t.Error("Expected CreateFile to fail when balance is insufficient for storage rent")
	}

	// With sufficient balance it should succeed
	requiredCost := filestore.CalculateStorageCost(int64(len(largeData)))
	f.Balance = requiredCost + 1
	_, err = fs.CreateFile(f)
	if err != nil {
		t.Errorf("CreateFile with sufficient balance failed: %v", err)
	}
}

// TestStorageRentBalanceDecrease verifies that decreasing a file's balance below
// the storage cost is rejected (Requirement 2.3).
func TestStorageRentBalanceDecrease(t *testing.T) {
	dbPath := t.TempDir() + "/rent_decrease_test.db"
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	defer fs.Close()

	data := make([]byte, 2*1024) // 2 KB
	requiredCost := filestore.CalculateStorageCost(int64(len(data)))

	id := filestore.GenerateFileID([]byte("rent-decrease-file"))
	startBalance := requiredCost + 10_000
	f := &filestore.File{
		ID:         id,
		Balance:    startBalance,
		TxManager:  filestore.FileID{},
		Data:       data,
		Executable: false,
	}
	if _, err := fs.CreateFile(f); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}

	// Decrease balance to 1 (well below requiredCost) — must fail
	delta := -(startBalance - 1)
	err = fs.UpdateFileBalance(id, delta)
	if err == nil {
		t.Error("Expected UpdateFileBalance to fail when new balance is below storage rent")
	}

	// Verify balance is unchanged
	after, _ := fs.GetFile(id)
	if after.Balance != startBalance {
		t.Errorf("Balance changed on failed update: got %d, want %d", after.Balance, startBalance)
	}
}

// TestFileStoreStateConsistency verifies that the FileStore state root changes
// after modifications and remains stable when no changes occur (Requirements 3.1-3.5).
func TestFileStoreStateConsistency(t *testing.T) {
	fs, _, _ := setupBuiltinEnv(t)

	root1, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("CalculateStateRoot: %v", err)
	}

	// Add a new file — root must change
	newID := filestore.GenerateFileID([]byte("consistency-test"))
	makeAccount(t, fs, newID, 5_000)

	root2, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("CalculateStateRoot after add: %v", err)
	}
	if string(root1) == string(root2) {
		t.Error("State root should change after adding a file")
	}

	// Calling again without changes — root must be stable
	root3, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("CalculateStateRoot stable: %v", err)
	}
	if string(root2) != string(root3) {
		t.Error("State root should be stable when no changes occur")
	}
}

// TestTransactionProcessingThroughRuntime verifies that the Runtime correctly
// dispatches to the QuanticScript System_Program and that the TxProcessor handles
// the full lifecycle (Requirements 3.3, 3.4, 3.5).
func TestTransactionProcessingThroughRuntime(t *testing.T) {
	fs, rt, _ := setupBuiltinEnv(t)

	// Verify the runtime does NOT have a Go builtin for System_Program
	// (it now runs as QuanticScript bytecode)
	if rt.IsBuiltinProgram(genesis.SystemProgramID) {
		t.Error("System_Program should not be a Go builtin after migration")
	}

	// Verify the program file is executable
	progFile, err := fs.GetFile(genesis.SystemProgramID)
	if err != nil {
		t.Fatalf("System_Program not in FileStore: %v", err)
	}
	if err := rt.ValidateProgram(progFile); err != nil {
		t.Errorf("ValidateProgram failed: %v", err)
	}

	// Execute a simple transfer via the runtime directly (no TxProcessor)
	senderID := filestore.GenerateFileID([]byte("rt-sender"))
	recipientID := filestore.GenerateFileID([]byte("rt-recipient"))
	makeAccount(t, fs, senderID, 20_000)
	makeAccount(t, fs, recipientID, 0)

	instr := &transaction.Instruction{
		ProgramID: genesis.SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: senderID, Permission: transaction.Write},
			"to":   {FileID: recipientID, Permission: transaction.Write},
		},
		Data: encodeTransferInstruction(5_000, senderID, recipientID),
	}

	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instr.Inputs)

	ctx := runtime.NewExecutionContext(instr, nil, fs, ac)
	if err := rt.ExecuteProgram(progFile, ctx); err != nil {
		t.Fatalf("ExecuteProgram: %v", err)
	}

	sender, _ := fs.GetFile(senderID)
	recipient, _ := fs.GetFile(recipientID)
	if sender.Balance != 15_000 {
		t.Errorf("Sender balance = %d, want 15000", sender.Balance)
	}
	if recipient.Balance != 5_000 {
		t.Errorf("Recipient balance = %d, want 5000", recipient.Balance)
	}
}

// builtinPKToFileID mirrors the TxProcessor's publicKeyToFileID mapping.
func builtinPKToFileID(pk transaction.PublicKey) filestore.FileID {
	var id filestore.FileID
	copy(id[:], pk[:])
	return id
}

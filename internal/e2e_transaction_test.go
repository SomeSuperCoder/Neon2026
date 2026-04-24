package internal

import (
	"crypto/ed25519"
	"os"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/processor"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/system"
	"github.com/poh-blockchain/internal/transaction"
)

// TestEndToEndAccountCreationAndTransfer tests the complete flow:
// 1. Create two accounts
// 2. Transfer balance between them
// 3. Verify final balances
func TestEndToEndAccountCreationAndTransfer(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/e2e_test.db"
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

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypairs for two users
	alicePubKey, alicePrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate Alice's keypair: %v", err)
	}

	bobPubKey, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate Bob's keypair: %v", err)
	}

	var alicePK, bobPK transaction.PublicKey
	copy(alicePK[:], alicePubKey)
	copy(bobPK[:], bobPubKey)

	// Create fee payer account (Alice will be the fee payer)
	aliceAccountID := publicKeyToFileID(alicePK)
	aliceFeePayerFile := &filestore.File{
		ID:         aliceAccountID,
		Balance:    100000, // Sufficient for fees and transfers
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(aliceFeePayerFile)
	if err != nil {
		t.Fatalf("Failed to create Alice's account: %v", err)
	}

	// Step 1: Create Bob's account directly (not via transaction for simplicity)
	bobAccountID := publicKeyToFileID(bobPK)
	bobAccountFile := &filestore.File{
		ID:         bobAccountID,
		Balance:    5000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	_, err = fs.CreateFile(bobAccountFile)
	if err != nil {
		t.Fatalf("Failed to create Bob's account: %v", err)
	}

	// Verify Bob's account was created
	bobAccountCheck, err := fs.GetFile(bobAccountID)
	if err != nil {
		t.Fatalf("Bob's account not found after creation: %v", err)
	}

	if bobAccountCheck.Balance != 5000 {
		t.Errorf("Bob's initial balance = %d, want 5000", bobAccountCheck.Balance)
	}

	// Step 2: Transfer 2000 from Alice to Bob
	transferData := system.EncodeTransferInstruction(2000)

	transferTx := &transaction.Transaction{
		LastSeen: transaction.TxID{}, // Zero TxID
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: aliceAccountID, Permission: transaction.Write},
					"to":   {FileID: bobAccountID, Permission: transaction.Write},
				},
				Data: transferData,
			},
		},
		Signatures: []transaction.Signature{}, // Empty initially for signing
	}

	// Sign transfer transaction
	transferTxData, _ := transferTx.Marshal()
	aliceTransferSig := ed25519.Sign(alicePrivKey, transferTxData)
	var aliceTransferSigBytes [64]byte
	copy(aliceTransferSigBytes[:], aliceTransferSig)

	transferTx.Signatures = []transaction.Signature{
		{PublicKey: alicePK, Signature: aliceTransferSigBytes}, // Alice pays fee and authorizes transfer
	}

	// Get Alice's balance before transfer
	aliceBeforeTransfer, _ := fs.GetFile(aliceAccountID)
	aliceBalanceBefore := aliceBeforeTransfer.Balance

	// Process transfer transaction
	result, err := txProcessor.ProcessTransaction(transferTx)
	if err != nil {
		t.Fatalf("Failed to process transfer transaction: %v", err)
	}

	if !result.Success {
		t.Fatalf("Transfer transaction failed: %v", result.Error)
	}

	// Step 3: Verify final balances
	aliceAfter, err := fs.GetFile(aliceAccountID)
	if err != nil {
		t.Fatalf("Failed to get Alice's account: %v", err)
	}

	bobAfter, err := fs.GetFile(bobAccountID)
	if err != nil {
		t.Fatalf("Failed to get Bob's account: %v", err)
	}

	// Calculate expected fee for transfer transaction (1 instruction, 1 signature)
	expectedFee := int64(5000 + 1000 + 500) // base + instruction + signature

	// Alice should have: previous balance - fee - transfer amount
	expectedAliceBalance := aliceBalanceBefore - expectedFee - 2000
	if aliceAfter.Balance != expectedAliceBalance {
		t.Errorf("Alice's final balance = %d, want %d", aliceAfter.Balance, expectedAliceBalance)
	}

	// Bob should have: 5000 (initial) + 2000 (transfer) = 7000
	if bobAfter.Balance != 7000 {
		t.Errorf("Bob's final balance = %d, want 7000", bobAfter.Balance)
	}

	t.Logf("End-to-end test passed: Alice balance=%d, Bob balance=%d", aliceAfter.Balance, bobAfter.Balance)
}

// TestMultiInstructionTransactionAtomicity tests that multi-instruction transactions
// are atomic - either all instructions succeed or all are reverted
func TestMultiInstructionTransactionAtomicity(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/atomicity_test.db"
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

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create three accounts
	account1ID := filestore.GenerateFileID([]byte("account1"))
	account2ID := filestore.GenerateFileID([]byte("account2"))
	account3ID := filestore.GenerateFileID([]byte("account3"))

	account1 := &filestore.File{
		ID:         account1ID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	account2 := &filestore.File{
		ID:         account2ID,
		Balance:    5000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	account3 := &filestore.File{
		ID:         account3ID,
		Balance:    3000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(account1)
	if err != nil {
		t.Fatalf("Failed to create account1: %v", err)
	}
	_, err = fs.CreateFile(account2)
	if err != nil {
		t.Fatalf("Failed to create account2: %v", err)
	}
	_, err = fs.CreateFile(account3)
	if err != nil {
		t.Fatalf("Failed to create account3: %v", err)
	}

	// Save original balances
	originalBalance1 := account1.Balance
	originalBalance2 := account2.Balance
	originalBalance3 := account3.Balance

	// Create a multi-instruction transaction:
	// 1. Transfer 1000 from account1 to account2 (should succeed)
	// 2. Transfer 10000 from account2 to account3 (should fail - insufficient balance)
	multiInstrTx := &transaction.Transaction{
		LastSeen: transaction.TxID{}, // Zero TxID
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: account1ID, Permission: transaction.Write},
					"to":   {FileID: account2ID, Permission: transaction.Write},
				},
				Data: system.EncodeTransferInstruction(1000),
			},
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: account2ID, Permission: transaction.Write},
					"to":   {FileID: account3ID, Permission: transaction.Write},
				},
				Data: system.EncodeTransferInstruction(10000), // This will fail
			},
		},
		Signatures: []transaction.Signature{}, // Empty initially for signing
	}

	// Sign transaction
	txData, _ := multiInstrTx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	multiInstrTx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should fail
	result, err := txProcessor.ProcessTransaction(multiInstrTx)
	if err == nil {
		t.Error("Expected transaction to fail due to insufficient balance in second instruction")
	}

	if result != nil && result.Success {
		t.Error("Transaction should not succeed")
	}

	// Verify all balances are unchanged (atomicity - including fee revert)
	account1After, _ := fs.GetFile(account1ID)
	account2After, _ := fs.GetFile(account2ID)
	account3After, _ := fs.GetFile(account3ID)

	// All accounts should have original balances (fee is also reverted on failure in current implementation)
	if account1After.Balance != originalBalance1 {
		t.Errorf("Account1 balance = %d, want %d (original, fee reverted)", account1After.Balance, originalBalance1)
	}

	if account2After.Balance != originalBalance2 {
		t.Errorf("Account2 balance = %d, want %d (unchanged)", account2After.Balance, originalBalance2)
	}

	if account3After.Balance != originalBalance3 {
		t.Errorf("Account3 balance = %d, want %d (unchanged)", account3After.Balance, originalBalance3)
	}

	t.Log("Multi-instruction atomicity test passed: all state changes reverted on failure")
}

// TestTransactionRevertOnInstructionFailure tests that transaction state is reverted
// when an instruction fails during execution
func TestTransactionRevertOnInstructionFailure(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/revert_test.db"
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

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypair
	pubKey, privKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pk transaction.PublicKey
	copy(pk[:], pubKey)

	// Create two accounts
	fromID := filestore.GenerateFileID([]byte("from_account"))
	toID := filestore.GenerateFileID([]byte("to_account"))

	fromAccount := &filestore.File{
		ID:         fromID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	toAccount := &filestore.File{
		ID:         toID,
		Balance:    5000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(fromAccount)
	if err != nil {
		t.Fatalf("Failed to create from account: %v", err)
	}
	_, err = fs.CreateFile(toAccount)
	if err != nil {
		t.Fatalf("Failed to create to account: %v", err)
	}

	originalFromBalance := fromAccount.Balance
	originalToBalance := toAccount.Balance

	// Create a transaction that will fail (transfer more than available)
	failingTx := &transaction.Transaction{
		LastSeen: transaction.TxID{}, // Zero TxID
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: fromID, Permission: transaction.Write},
					"to":   {FileID: toID, Permission: transaction.Write},
				},
				Data: system.EncodeTransferInstruction(20000), // More than available
			},
		},
		Signatures: []transaction.Signature{}, // Empty initially for signing
	}

	// Sign transaction
	txData, _ := failingTx.Marshal()
	sig := ed25519.Sign(privKey, txData)
	var sigBytes [64]byte
	copy(sigBytes[:], sig)

	failingTx.Signatures = []transaction.Signature{
		{PublicKey: pk, Signature: sigBytes},
	}

	// Process transaction - should fail
	result, err := txProcessor.ProcessTransaction(failingTx)
	if err == nil {
		t.Error("Expected transaction to fail due to insufficient balance")
	}

	if result != nil && result.Success {
		t.Error("Transaction should not succeed")
	}

	// Verify balances are reverted (including fee)
	fromAfter, _ := fs.GetFile(fromID)
	toAfter, _ := fs.GetFile(toID)

	// Both accounts should have original balances (fee is also reverted on failure in current implementation)
	if fromAfter.Balance != originalFromBalance {
		t.Errorf("From account balance = %d, want %d (original, fee reverted)", fromAfter.Balance, originalFromBalance)
	}

	if toAfter.Balance != originalToBalance {
		t.Errorf("To account balance = %d, want %d (unchanged)", toAfter.Balance, originalToBalance)
	}

	t.Log("Transaction revert test passed: state correctly reverted on instruction failure")
}

// TestFeePaymentAndBalanceUpdates tests that fees are correctly deducted
// and balances are properly updated throughout transaction execution
func TestFeePaymentAndBalanceUpdates(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/fee_test.db"
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

	txProcessor := processor.NewTxProcessor(fs, rt)

	// Generate keypairs
	feePayerPubKey, feePayerPrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate fee payer keypair: %v", err)
	}

	senderPubKey, senderPrivKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Failed to generate sender keypair: %v", err)
	}

	var feePayerPK, senderPK transaction.PublicKey
	copy(feePayerPK[:], feePayerPubKey)
	copy(senderPK[:], senderPubKey)

	// Create accounts
	feePayerID := publicKeyToFileID(feePayerPK)
	senderID := filestore.GenerateFileID([]byte("sender"))
	recipientID := filestore.GenerateFileID([]byte("recipient"))

	feePayerAccount := &filestore.File{
		ID:         feePayerID,
		Balance:    50000, // Sufficient for fees
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	senderAccount := &filestore.File{
		ID:         senderID,
		Balance:    20000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	recipientAccount := &filestore.File{
		ID:         recipientID,
		Balance:    10000,
		TxManager:  system.SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(feePayerAccount)
	if err != nil {
		t.Fatalf("Failed to create fee payer account: %v", err)
	}
	_, err = fs.CreateFile(senderAccount)
	if err != nil {
		t.Fatalf("Failed to create sender account: %v", err)
	}
	_, err = fs.CreateFile(recipientAccount)
	if err != nil {
		t.Fatalf("Failed to create recipient account: %v", err)
	}

	originalFeePayerBalance := feePayerAccount.Balance
	originalSenderBalance := senderAccount.Balance
	originalRecipientBalance := recipientAccount.Balance

	// Create a transaction with multiple instructions
	// Fee payer is different from the sender
	multiTx := &transaction.Transaction{
		LastSeen: transaction.TxID{}, // Zero TxID
		Instructions: []transaction.Instruction{
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: senderID, Permission: transaction.Write},
					"to":   {FileID: recipientID, Permission: transaction.Write},
				},
				Data: system.EncodeTransferInstruction(3000),
			},
			{
				ProgramID: system.SystemProgramID,
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: senderID, Permission: transaction.Write},
					"to":   {FileID: recipientID, Permission: transaction.Write},
				},
				Data: system.EncodeTransferInstruction(2000),
			},
		},
		Signatures: []transaction.Signature{}, // Empty initially for signing
	}

	// Sign transaction (fee payer signs first, sender signs second)
	txData, _ := multiTx.Marshal()
	feePayerSig := ed25519.Sign(feePayerPrivKey, txData)
	senderSig := ed25519.Sign(senderPrivKey, txData)

	var feePayerSigBytes, senderSigBytes [64]byte
	copy(feePayerSigBytes[:], feePayerSig)
	copy(senderSigBytes[:], senderSig)

	multiTx.Signatures = []transaction.Signature{
		{PublicKey: feePayerPK, Signature: feePayerSigBytes}, // Fee payer (first signature)
		{PublicKey: senderPK, Signature: senderSigBytes},     // Sender
	}

	// Process transaction
	result, err := txProcessor.ProcessTransaction(multiTx)
	if err != nil {
		t.Fatalf("Failed to process transaction: %v", err)
	}

	if !result.Success {
		t.Fatalf("Transaction failed: %v", result.Error)
	}

	// Verify balances
	feePayerAfter, _ := fs.GetFile(feePayerID)
	senderAfter, _ := fs.GetFile(senderID)
	recipientAfter, _ := fs.GetFile(recipientID)

	// Calculate expected fee (2 instructions, 2 signatures)
	expectedFee := int64(5000 + 2000 + 1000) // base + 2*instruction + 2*signature

	// Fee payer should have: original - fee
	expectedFeePayerBalance := originalFeePayerBalance - expectedFee
	if feePayerAfter.Balance != expectedFeePayerBalance {
		t.Errorf("Fee payer balance = %d, want %d", feePayerAfter.Balance, expectedFeePayerBalance)
	}

	// Sender should have: original - 3000 - 2000 = original - 5000
	expectedSenderBalance := originalSenderBalance - 5000
	if senderAfter.Balance != expectedSenderBalance {
		t.Errorf("Sender balance = %d, want %d", senderAfter.Balance, expectedSenderBalance)
	}

	// Recipient should have: original + 3000 + 2000 = original + 5000
	expectedRecipientBalance := originalRecipientBalance + 5000
	if recipientAfter.Balance != expectedRecipientBalance {
		t.Errorf("Recipient balance = %d, want %d", recipientAfter.Balance, expectedRecipientBalance)
	}

	// Verify gas used matches fee
	if result.GasUsed != expectedFee {
		t.Errorf("Gas used = %d, want %d", result.GasUsed, expectedFee)
	}

	t.Logf("Fee payment test passed: fee=%d, feePayer=%d, sender=%d, recipient=%d",
		expectedFee, feePayerAfter.Balance, senderAfter.Balance, recipientAfter.Balance)
}

// Helper function to convert public key to file ID
func publicKeyToFileID(pubkey transaction.PublicKey) filestore.FileID {
	var fileID filestore.FileID
	copy(fileID[:], pubkey[:])
	return fileID
}

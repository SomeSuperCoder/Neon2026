package transaction

import (
	"encoding/binary"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// SystemProgramID is the built-in system program (0x00...01)
// Defined here to avoid import cycle with genesis package
var testSystemProgramID = filestore.FileID{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 1,
}

// TestNewTransactionBuilder tests the constructor
func TestNewTransactionBuilder(t *testing.T) {
	lastSeen := TxID{1, 2, 3}

	builder := NewTransactionBuilder(lastSeen)

	if builder == nil {
		t.Fatal("NewTransactionBuilder returned nil")
	}

	if builder.lastSeen != lastSeen {
		t.Errorf("lastSeen mismatch: expected %v, got %v", lastSeen, builder.lastSeen)
	}

	if builder.instructions == nil {
		t.Error("instructions slice should be initialized")
	}

	if builder.signatures == nil {
		t.Error("signatures slice should be initialized")
	}
}

// TestAddCreateFileInstruction tests adding a CREATE_FILE instruction
func TestAddCreateFileInstruction(t *testing.T) {
	lastSeen := TxID{}
	builder := NewTransactionBuilder(lastSeen)

	systemID := testSystemProgramID
	newFileID := filestore.FileID{0x01}
	payerID := filestore.FileID{0x02}
	balance := int64(1000000)
	owner := PublicKey{0x03}

	err := builder.AddCreateFileInstruction(systemID, newFileID, payerID, balance, owner)
	if err != nil {
		t.Fatalf("AddCreateFileInstruction failed: %v", err)
	}

	if len(builder.instructions) != 1 {
		t.Fatalf("Expected 1 instruction, got %d", len(builder.instructions))
	}

	instruction := builder.instructions[0]

	// Check program ID
	if instruction.ProgramID != systemID {
		t.Errorf("ProgramID mismatch: expected %v, got %v", systemID, instruction.ProgramID)
	}

	// Check inputs
	expectedInputs := map[string]FileAccess{
		"program": {
			FileID:     systemID,
			Permission: Read,
		},
		"payer": {
			FileID:     payerID,
			Permission: Write,
		},
		"new_file": {
			FileID:     newFileID,
			Permission: Write,
		},
	}

	if len(instruction.Inputs) != len(expectedInputs) {
		t.Fatalf("Expected %d inputs, got %d", len(expectedInputs), len(instruction.Inputs))
	}

	for key, expected := range expectedInputs {
		actual, exists := instruction.Inputs[key]
		if !exists {
			t.Errorf("Missing input key: %s", key)
			continue
		}

		if actual.FileID != expected.FileID {
			t.Errorf("Input %s FileID mismatch: expected %v, got %v", key, expected.FileID, actual.FileID)
		}

		if actual.Permission != expected.Permission {
			t.Errorf("Input %s Permission mismatch: expected %v, got %v", key, expected.Permission, actual.Permission)
		}
	}

	// Check instruction data format
	expectedDataLen := 105 // type(1) + fileID(32) + payer(32) + balance(8) + owner(32)
	if len(instruction.Data) != expectedDataLen {
		t.Fatalf("Expected data length %d, got %d", expectedDataLen, len(instruction.Data))
	}

	// Check instruction type (CREATE_FILE = 0)
	if instruction.Data[0] != 0 {
		t.Errorf("Expected instruction type 0, got %d", instruction.Data[0])
	}

	// Check new file ID
	var actualNewFileID filestore.FileID
	copy(actualNewFileID[:], instruction.Data[1:33])
	if actualNewFileID != newFileID {
		t.Errorf("New file ID mismatch: expected %v, got %v", newFileID, actualNewFileID)
	}

	// Check payer ID
	var actualPayerID filestore.FileID
	copy(actualPayerID[:], instruction.Data[33:65])
	if actualPayerID != payerID {
		t.Errorf("Payer ID mismatch: expected %v, got %v", payerID, actualPayerID)
	}

	// Check balance
	actualBalance := int64(binary.LittleEndian.Uint64(instruction.Data[65:73]))
	if actualBalance != balance {
		t.Errorf("Balance mismatch: expected %d, got %d", balance, actualBalance)
	}

	// Check owner
	var actualOwner PublicKey
	copy(actualOwner[:], instruction.Data[73:105])
	if actualOwner != owner {
		t.Errorf("Owner mismatch: expected %v, got %v", owner, actualOwner)
	}
}

// TestAddCreateFileInstructionNegativeBalance tests validation of negative balance
func TestAddCreateFileInstructionNegativeBalance(t *testing.T) {
	lastSeen := TxID{}
	builder := NewTransactionBuilder(lastSeen)

	systemID := testSystemProgramID
	newFileID := filestore.FileID{0x01}
	payerID := filestore.FileID{0x02}
	balance := int64(-1000) // Negative balance
	owner := PublicKey{0x03}

	err := builder.AddCreateFileInstruction(systemID, newFileID, payerID, balance, owner)
	if err == nil {
		t.Fatal("Expected error for negative balance, got nil")
	}

	expectedError := "balance must be non-negative, got -1000"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Should not add instruction on error
	if len(builder.instructions) != 0 {
		t.Errorf("Expected 0 instructions after error, got %d", len(builder.instructions))
	}
}

// TestAddCreateFileInstructionZeroBalance tests that zero balance is allowed
func TestAddCreateFileInstructionZeroBalance(t *testing.T) {
	lastSeen := TxID{}
	builder := NewTransactionBuilder(lastSeen)

	systemID := testSystemProgramID
	newFileID := filestore.FileID{0x01}
	payerID := filestore.FileID{0x02}
	balance := int64(0) // Zero balance should be allowed
	owner := PublicKey{0x03}

	err := builder.AddCreateFileInstruction(systemID, newFileID, payerID, balance, owner)
	if err != nil {
		t.Fatalf("AddCreateFileInstruction with zero balance failed: %v", err)
	}

	if len(builder.instructions) != 1 {
		t.Fatalf("Expected 1 instruction, got %d", len(builder.instructions))
	}
}

// TestAddTransferInstruction tests adding a transfer instruction
func TestAddTransferInstruction(t *testing.T) {
	lastSeen := TxID{}
	builder := NewTransactionBuilder(lastSeen)

	systemID := testSystemProgramID
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}
	amount := int64(1000)

	err := builder.AddTransferInstruction(systemID, senderID, receiverID, amount)
	if err != nil {
		t.Fatalf("AddTransferInstruction failed: %v", err)
	}

	if len(builder.instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(builder.instructions))
	}

	instr := builder.instructions[0]

	// Verify ProgramID
	if instr.ProgramID != systemID {
		t.Errorf("ProgramID mismatch: expected %v, got %v", systemID, instr.ProgramID)
	}

	// Verify Inputs map has 3 entries
	if len(instr.Inputs) != 3 {
		t.Fatalf("expected 3 inputs, got %d", len(instr.Inputs))
	}

	// Verify program input with Read permission
	programAccess, ok := instr.Inputs["program"]
	if !ok {
		t.Fatal("program input not found")
	}
	if programAccess.FileID != systemID {
		t.Errorf("program FileID mismatch: expected %v, got %v", systemID, programAccess.FileID)
	}
	if programAccess.Permission != Read {
		t.Errorf("program permission mismatch: expected Read, got %v", programAccess.Permission)
	}

	// Verify sender input with Write permission
	senderAccess, ok := instr.Inputs["sender"]
	if !ok {
		t.Fatal("sender input not found")
	}
	if senderAccess.FileID != senderID {
		t.Errorf("sender FileID mismatch: expected %v, got %v", senderID, senderAccess.FileID)
	}
	if senderAccess.Permission != Write {
		t.Errorf("sender permission mismatch: expected Write, got %v", senderAccess.Permission)
	}

	// Verify receiver input with Write permission
	receiverAccess, ok := instr.Inputs["receiver"]
	if !ok {
		t.Fatal("receiver input not found")
	}
	if receiverAccess.FileID != receiverID {
		t.Errorf("receiver FileID mismatch: expected %v, got %v", receiverID, receiverAccess.FileID)
	}
	if receiverAccess.Permission != Write {
		t.Errorf("receiver permission mismatch: expected Write, got %v", receiverAccess.Permission)
	}

	// Verify instruction data format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
	if len(instr.Data) != 73 {
		t.Fatalf("expected data length 73, got %d", len(instr.Data))
	}

	// Verify instruction type (TRANSFER = 1)
	if instr.Data[0] != 1 {
		t.Errorf("expected instruction type 1, got %d", instr.Data[0])
	}

	// Verify from FileID
	var fromID filestore.FileID
	copy(fromID[:], instr.Data[1:33])
	if fromID != senderID {
		t.Errorf("from FileID mismatch: expected %v, got %v", senderID, fromID)
	}

	// Verify to FileID
	var toID filestore.FileID
	copy(toID[:], instr.Data[33:65])
	if toID != receiverID {
		t.Errorf("to FileID mismatch: expected %v, got %v", receiverID, toID)
	}

	// Verify amount (little-endian)
	decodedAmount := int64(binary.LittleEndian.Uint64(instr.Data[65:73]))
	if decodedAmount != amount {
		t.Errorf("amount mismatch: expected %d, got %d", amount, decodedAmount)
	}
}

// TestAddTransferInstructionZeroAmount tests rejection of zero amount
func TestAddTransferInstructionZeroAmount(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	err := builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x01},
		filestore.FileID{0x02},
		0, // zero amount
	)

	if err == nil {
		t.Fatal("expected error for zero amount, got nil")
	}
}

// TestAddTransferInstructionNegativeAmount tests rejection of negative amount
func TestAddTransferInstructionNegativeAmount(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	err := builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x01},
		filestore.FileID{0x02},
		-100, // negative amount
	)

	if err == nil {
		t.Fatal("expected error for negative amount, got nil")
	}
}

// TestAddSignature tests adding signatures
func TestAddSignature(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	sig := Signature{
		PublicKey: PublicKey{0x01},
		Signature: [64]byte{0x02},
	}

	err := builder.AddSignature(sig)
	if err != nil {
		t.Fatalf("AddSignature failed: %v", err)
	}

	if len(builder.signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(builder.signatures))
	}

	if builder.signatures[0] != sig {
		t.Error("signature mismatch")
	}
}

// TestAddMultipleSignatures tests adding multiple signatures
func TestAddMultipleSignatures(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	sig1 := Signature{PublicKey: PublicKey{0x01}, Signature: [64]byte{0x02}}
	sig2 := Signature{PublicKey: PublicKey{0x03}, Signature: [64]byte{0x04}}

	builder.AddSignature(sig1)
	builder.AddSignature(sig2)

	if len(builder.signatures) != 2 {
		t.Fatalf("expected 2 signatures, got %d", len(builder.signatures))
	}
}

// TestBuild tests building a complete transaction
func TestBuild(t *testing.T) {
	lastSeen := TxID{0x99}
	builder := NewTransactionBuilder(lastSeen)

	// Add transfer instruction
	err := builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x01},
		filestore.FileID{0x02},
		1000,
	)
	if err != nil {
		t.Fatalf("AddTransferInstruction failed: %v", err)
	}

	// Add signature
	sig := Signature{
		PublicKey: PublicKey{0x01},
		Signature: [64]byte{0x02},
	}
	builder.AddSignature(sig)

	// Build transaction
	tx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if tx == nil {
		t.Fatal("Build returned nil transaction")
	}

	// Verify LastSeen
	if tx.LastSeen != lastSeen {
		t.Errorf("LastSeen mismatch: expected %v, got %v", lastSeen, tx.LastSeen)
	}

	// Verify Instructions
	if len(tx.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(tx.Instructions))
	}

	// Verify Signatures
	if len(tx.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(tx.Signatures))
	}
	if tx.Signatures[0] != sig {
		t.Error("signature mismatch in built transaction")
	}
}

// TestBuildWithoutInstructions tests building with no instructions
func TestBuildWithoutInstructions(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})
	builder.AddSignature(Signature{PublicKey: PublicKey{0x01}})

	tx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(tx.Instructions) != 0 {
		t.Errorf("expected 0 instructions, got %d", len(tx.Instructions))
	}
}

// TestBuildWithoutSignatures tests building with no signatures
func TestBuildWithoutSignatures(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	err := builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x01},
		filestore.FileID{0x02},
		1000,
	)
	if err != nil {
		t.Fatalf("AddTransferInstruction failed: %v", err)
	}

	tx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(tx.Signatures) != 0 {
		t.Errorf("expected 0 signatures, got %d", len(tx.Signatures))
	}
}

// TestMultipleInstructions tests adding multiple instructions
func TestMultipleInstructions(t *testing.T) {
	builder := NewTransactionBuilder(TxID{})

	// Add first transfer
	err := builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x01},
		filestore.FileID{0x02},
		1000,
	)
	if err != nil {
		t.Fatalf("First AddTransferInstruction failed: %v", err)
	}

	// Add second transfer
	err = builder.AddTransferInstruction(
		testSystemProgramID,
		filestore.FileID{0x02},
		filestore.FileID{0x03},
		500,
	)
	if err != nil {
		t.Fatalf("Second AddTransferInstruction failed: %v", err)
	}

	if len(builder.instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(builder.instructions))
	}

	// Build and verify
	builder.AddSignature(Signature{PublicKey: PublicKey{0x01}})
	tx, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(tx.Instructions) != 2 {
		t.Fatalf("expected 2 instructions in transaction, got %d", len(tx.Instructions))
	}
}

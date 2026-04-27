package transaction

import (
	"encoding/binary"
	"testing"

	"github.com/poh-blockchain/internal/filestore"
)

// Test SystemProgramID for testing purposes
// Defined here to avoid import cycle with genesis package
var testSystemID = filestore.FileID{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 1,
}

// TestCreateTransferInstruction tests creating a transfer instruction with valid parameters
// Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6
func TestCreateTransferInstruction(t *testing.T) {
	systemID := testSystemID
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}
	amount := int64(1000)

	instr, err := CreateTransferInstruction(systemID, senderID, receiverID, amount)
	if err != nil {
		t.Fatalf("CreateTransferInstruction failed: %v", err)
	}

	// Verify ProgramID
	if instr.ProgramID != systemID {
		t.Errorf("ProgramID = %v, want %v", instr.ProgramID, systemID)
	}

	// Verify inputs map has 3 entries
	if len(instr.Inputs) != 3 {
		t.Errorf("Inputs length = %d, want 3", len(instr.Inputs))
	}

	// Verify program declared with Read permission
	programAccess, ok := instr.Inputs[InputKeyProgram]
	if !ok {
		t.Fatalf("Program not declared in inputs")
	}
	if programAccess.FileID != systemID {
		t.Errorf("Program FileID = %v, want %v", programAccess.FileID, systemID)
	}
	if programAccess.Permission != Read {
		t.Errorf("Program Permission = %v, want Read", programAccess.Permission)
	}

	// Verify sender declared with Write permission
	senderAccess, ok := instr.Inputs[InputKeySender]
	if !ok {
		t.Fatalf("Sender not declared in inputs")
	}
	if senderAccess.FileID != senderID {
		t.Errorf("Sender FileID = %v, want %v", senderAccess.FileID, senderID)
	}
	if senderAccess.Permission != Write {
		t.Errorf("Sender Permission = %v, want Write", senderAccess.Permission)
	}

	// Verify receiver declared with Write permission
	receiverAccess, ok := instr.Inputs[InputKeyReceiver]
	if !ok {
		t.Fatalf("Receiver not declared in inputs")
	}
	if receiverAccess.FileID != receiverID {
		t.Errorf("Receiver FileID = %v, want %v", receiverAccess.FileID, receiverID)
	}
	if receiverAccess.Permission != Write {
		t.Errorf("Receiver Permission = %v, want Write", receiverAccess.Permission)
	}

	// Verify instruction data is correctly encoded
	if len(instr.Data) != 73 {
		t.Errorf("Data length = %d, want 73", len(instr.Data))
	}

	// Verify instruction type (TRANSFER = 1)
	if instr.Data[0] != 1 {
		t.Errorf("Instruction type = %d, want 1", instr.Data[0])
	}

	// Verify sender FileID in data
	dataSenderID := filestore.FileID{}
	copy(dataSenderID[:], instr.Data[1:33])
	if dataSenderID != senderID {
		t.Errorf("Data sender ID = %v, want %v", dataSenderID, senderID)
	}

	// Verify receiver FileID in data
	dataReceiverID := filestore.FileID{}
	copy(dataReceiverID[:], instr.Data[33:65])
	if dataReceiverID != receiverID {
		t.Errorf("Data receiver ID = %v, want %v", dataReceiverID, receiverID)
	}

	// Verify amount in data
	dataAmount := int64(binary.LittleEndian.Uint64(instr.Data[65:73]))
	if dataAmount != amount {
		t.Errorf("Data amount = %d, want %d", dataAmount, amount)
	}
}

// TestCreateTransferInstructionZeroAmount tests rejection of zero amount
// Requirements: 2.4
func TestCreateTransferInstructionZeroAmount(t *testing.T) {
	systemID := testSystemID
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}

	_, err := CreateTransferInstruction(systemID, senderID, receiverID, 0)
	if err == nil {
		t.Fatal("Expected error for zero amount, got nil")
	}
}

// TestCreateTransferInstructionNegativeAmount tests rejection of negative amount
// Requirements: 2.4
func TestCreateTransferInstructionNegativeAmount(t *testing.T) {
	systemID := testSystemID
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}

	_, err := CreateTransferInstruction(systemID, senderID, receiverID, -100)
	if err == nil {
		t.Fatal("Expected error for negative amount, got nil")
	}
}

// TestEncodeTransferData tests correct encoding of transfer data
// Requirements: 2.1, 6.4
func TestEncodeTransferData(t *testing.T) {
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}
	amount := int64(5000)

	data := EncodeTransferData(senderID, receiverID, amount)

	// Verify data length
	if len(data) != 73 {
		t.Errorf("Data length = %d, want 73", len(data))
	}

	// Verify instruction type (TRANSFER = 1)
	if data[0] != 1 {
		t.Errorf("Instruction type = %d, want 1", data[0])
	}

	// Verify sender FileID
	dataSenderID := filestore.FileID{}
	copy(dataSenderID[:], data[1:33])
	if dataSenderID != senderID {
		t.Errorf("Sender ID = %v, want %v", dataSenderID, senderID)
	}

	// Verify receiver FileID
	dataReceiverID := filestore.FileID{}
	copy(dataReceiverID[:], data[33:65])
	if dataReceiverID != receiverID {
		t.Errorf("Receiver ID = %v, want %v", dataReceiverID, receiverID)
	}

	// Verify amount (little-endian)
	dataAmount := int64(binary.LittleEndian.Uint64(data[65:73]))
	if dataAmount != amount {
		t.Errorf("Amount = %d, want %d", dataAmount, amount)
	}
}

// TestEncodeTransferDataLargeAmount tests encoding with large amount
// Requirements: 2.1, 6.4
func TestEncodeTransferDataLargeAmount(t *testing.T) {
	senderID := filestore.FileID{0xAA}
	receiverID := filestore.FileID{0xBB}
	amount := int64(9223372036854775807) // Max int64

	data := EncodeTransferData(senderID, receiverID, amount)

	// Verify amount is correctly encoded
	dataAmount := int64(binary.LittleEndian.Uint64(data[65:73]))
	if dataAmount != amount {
		t.Errorf("Amount = %d, want %d", dataAmount, amount)
	}
}

// TestInputKeyConstants tests that standard input key constants are defined
// Requirements: 2.1
func TestInputKeyConstants(t *testing.T) {
	// Verify constants are defined and have expected values
	if InputKeyProgram != "program" {
		t.Errorf("InputKeyProgram = %q, want %q", InputKeyProgram, "program")
	}
	if InputKeySender != "sender" {
		t.Errorf("InputKeySender = %q, want %q", InputKeySender, "sender")
	}
	if InputKeyReceiver != "receiver" {
		t.Errorf("InputKeyReceiver = %q, want %q", InputKeyReceiver, "receiver")
	}
	if InputKeyPayer != "payer" {
		t.Errorf("InputKeyPayer = %q, want %q", InputKeyPayer, "payer")
	}
	if InputKeyOwner != "owner" {
		t.Errorf("InputKeyOwner = %q, want %q", InputKeyOwner, "owner")
	}
	if InputKeyTarget != "target" {
		t.Errorf("InputKeyTarget = %q, want %q", InputKeyTarget, "target")
	}
}

// TestCreateTransferInstructionPermissions tests correct permission declarations
// Requirements: 2.2, 2.3, 3.1, 3.2, 3.3, 3.4, 3.5
func TestCreateTransferInstructionPermissions(t *testing.T) {
	systemID := testSystemID
	senderID := filestore.FileID{0x01}
	receiverID := filestore.FileID{0x02}
	amount := int64(1000)

	instr, err := CreateTransferInstruction(systemID, senderID, receiverID, amount)
	if err != nil {
		t.Fatalf("CreateTransferInstruction failed: %v", err)
	}

	tests := []struct {
		key        string
		wantFileID filestore.FileID
		wantPerm   AccessPermission
	}{
		{InputKeyProgram, systemID, Read},
		{InputKeySender, senderID, Write},
		{InputKeyReceiver, receiverID, Write},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			access, ok := instr.Inputs[tt.key]
			if !ok {
				t.Fatalf("Input %q not found", tt.key)
			}
			if access.FileID != tt.wantFileID {
				t.Errorf("FileID = %v, want %v", access.FileID, tt.wantFileID)
			}
			if access.Permission != tt.wantPerm {
				t.Errorf("Permission = %v, want %v", access.Permission, tt.wantPerm)
			}
		})
	}
}

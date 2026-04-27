package transaction

import (
	"encoding/binary"
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
)

// Standard input key constants for consistent file reference naming across the codebase.
// These constants define the standard keys used in the Inputs map of instructions.
// This implements Requirement 2.1
const (
	// InputKeyProgram is the key for the program being invoked
	InputKeyProgram = "program"

	// InputKeySender is the key for the source account in transfers
	InputKeySender = "sender"

	// InputKeyReceiver is the key for the destination account in transfers
	InputKeyReceiver = "receiver"

	// InputKeyPayer is the key for the fee payer or balance source
	InputKeyPayer = "payer"

	// InputKeyOwner is the key for file owner in ownership checks
	InputKeyOwner = "owner"

	// InputKeyTarget is the key for a generic target file
	InputKeyTarget = "target"
)

// CreateTransferInstruction creates a properly formatted System Program transfer instruction
// with all required input declarations.
//
// The instruction will declare:
// - System Program with Read permission (program can be executed)
// - Sender with Write permission (balance will be modified)
// - Receiver with Write permission (balance will be modified)
//
// This implements Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6
func CreateTransferInstruction(
	systemProgramID filestore.FileID,
	senderID filestore.FileID,
	receiverID filestore.FileID,
	amount int64,
) (*Instruction, error) {
	// Validate amount is positive (Requirement 2.4)
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}

	// Create inputs map with proper declarations (Requirements 2.1, 2.2, 2.3, 3.1, 3.2, 6.3)
	inputs := map[string]FileAccess{
		InputKeyProgram: {
			FileID:     systemProgramID,
			Permission: Read, // Program is read-only (Requirement 2.2)
		},
		InputKeySender: {
			FileID:     senderID,
			Permission: Write, // Sender balance will be modified (Requirement 3.1)
		},
		InputKeyReceiver: {
			FileID:     receiverID,
			Permission: Write, // Receiver balance will be modified (Requirement 3.2)
		},
	}

	// Encode instruction data (Requirements 2.1, 6.4)
	data := EncodeTransferData(senderID, receiverID, amount)

	// Create and return instruction (Requirements 1.1, 1.2, 1.3, 1.4, 1.5)
	return &Instruction{
		ProgramID: systemProgramID,
		Inputs:    inputs,
		Data:      data,
	}, nil
}

// EncodeTransferData encodes transfer parameters into the System Program instruction data format.
//
// Format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
// Total: 73 bytes
//
// - Byte 0: Instruction type (TRANSFER = 1)
// - Bytes 1-32: Sender FileID (32 bytes)
// - Bytes 33-64: Receiver FileID (32 bytes)
// - Bytes 65-72: Amount as little-endian int64 (8 bytes)
//
// This implements Requirements 2.1, 6.4
func EncodeTransferData(
	fromID filestore.FileID,
	toID filestore.FileID,
	amount int64,
) []byte {
	// Allocate data buffer (Requirement 6.4)
	// Format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
	data := make([]byte, 73)

	// Instruction type (TRANSFER = 1)
	data[0] = 1

	// From FileID (bytes 1-32)
	copy(data[1:33], fromID[:])

	// To FileID (bytes 33-64)
	copy(data[33:65], toID[:])

	// Amount (bytes 65-72, little-endian)
	binary.LittleEndian.PutUint64(data[65:73], uint64(amount))

	return data
}

// CreateFileInstruction creates a properly formatted System Program CREATE_FILE instruction
// with all required input declarations.
//
// The instruction will declare:
// - System Program with Read permission (program can be executed)
// - Payer with Write permission (balance will be deducted)
// - New file with Write permission (file will be created)
//
// This implements Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5
func CreateFileInstruction(
	systemProgramID filestore.FileID,
	newFileID filestore.FileID,
	payerID filestore.FileID,
	balance int64,
	owner PublicKey,
) (*Instruction, error) {
	// Validate balance is non-negative
	if balance < 0 {
		return nil, fmt.Errorf("balance must be non-negative, got %d", balance)
	}

	// Create inputs map with proper declarations
	inputs := map[string]FileAccess{
		InputKeyProgram: {
			FileID:     systemProgramID,
			Permission: Read, // Program is read-only
		},
		InputKeyPayer: {
			FileID:     payerID,
			Permission: Write, // Payer balance will be modified
		},
		"new_file": {
			FileID:     newFileID,
			Permission: Write, // New file will be created
		},
	}

	// Encode instruction data
	// Format: [type:u8(0)][fileID:FileID(32)][payer:FileID(32)][balance:i64(8)][owner:PublicKey(32)]
	data := make([]byte, 105)

	// Instruction type (CREATE_FILE = 0)
	data[0] = 0

	// New file ID (bytes 1-32)
	copy(data[1:33], newFileID[:])

	// Payer ID (bytes 33-64)
	copy(data[33:65], payerID[:])

	// Balance (bytes 65-72, little-endian)
	binary.LittleEndian.PutUint64(data[65:73], uint64(balance))

	// Owner public key (bytes 73-104)
	copy(data[73:105], owner[:])

	// Create and return instruction
	return &Instruction{
		ProgramID: systemProgramID,
		Inputs:    inputs,
		Data:      data,
	}, nil
}

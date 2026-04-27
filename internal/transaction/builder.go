package transaction

import (
	"encoding/binary"
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
)

// TransactionBuilder provides a convenient way to construct transactions
// with proper input declarations for instructions.
// This implements Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6
type TransactionBuilder struct {
	lastSeen     TxID
	instructions []Instruction
	signatures   []Signature
}

// NewTransactionBuilder creates a new transaction builder with the specified lastSeen TxID.
// This implements Requirement 6.1
func NewTransactionBuilder(lastSeen TxID) *TransactionBuilder {
	return &TransactionBuilder{
		lastSeen:     lastSeen,
		instructions: make([]Instruction, 0),
		signatures:   make([]Signature, 0),
	}
}

// AddTransferInstruction adds a System Program transfer instruction with proper input declarations.
// The instruction will declare:
// - System Program with Read permission
// - Sender with Write permission
// - Receiver with Write permission
//
// This implements Requirements 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 3.1, 3.2, 6.2, 6.3, 6.4
func (tb *TransactionBuilder) AddTransferInstruction(
	systemProgramID filestore.FileID,
	senderID filestore.FileID,
	receiverID filestore.FileID,
	amount int64,
) error {
	// Validate amount (Requirement 2.4)
	if amount <= 0 {
		return fmt.Errorf("amount must be positive, got %d", amount)
	}

	// Create inputs map with proper declarations (Requirements 1.1, 2.1, 2.2, 3.1, 3.2, 6.3)
	inputs := map[string]FileAccess{
		"program": {
			FileID:     systemProgramID,
			Permission: Read,
		},
		"sender": {
			FileID:     senderID,
			Permission: Write,
		},
		"receiver": {
			FileID:     receiverID,
			Permission: Write,
		},
	}

	// Encode instruction data (Requirement 6.4)
	// Format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
	data := make([]byte, 73)

	// Instruction type (TRANSFER = 1)
	data[0] = 1

	// From FileID
	copy(data[1:33], senderID[:])

	// To FileID
	copy(data[33:65], receiverID[:])

	// Amount (little-endian)
	binary.LittleEndian.PutUint64(data[65:73], uint64(amount))

	// Create instruction (Requirements 1.1, 1.2, 1.3, 1.4, 1.5)
	instruction := Instruction{
		ProgramID: systemProgramID,
		Inputs:    inputs,
		Data:      data,
	}

	tb.instructions = append(tb.instructions, instruction)
	return nil
}

// AddSignature adds a signature to the transaction.
// The first signature is treated as the fee payer.
// This implements Requirements 1.4, 4.1, 4.2, 4.3
func (tb *TransactionBuilder) AddSignature(sig Signature) error {
	tb.signatures = append(tb.signatures, sig)
	return nil
}

// Build constructs the final Transaction from the builder's state.
// This implements Requirements 1.5, 6.1, 6.2, 6.5, 6.6
func (tb *TransactionBuilder) Build() (*Transaction, error) {
	tx := &Transaction{
		LastSeen:     tb.lastSeen,
		Instructions: tb.instructions,
		Signatures:   tb.signatures,
	}

	return tx, nil
}

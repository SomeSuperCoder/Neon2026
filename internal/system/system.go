package system

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
)

// Package system implements the System Program, which provides
// built-in operations for account management and basic file operations.
// This implements Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3, 10.4, 10.5

// SystemProgramID is the unique identifier for the System Program
// This is a well-known constant that identifies the built-in system program
var SystemProgramID = filestore.FileID{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
}

// Instruction type constants
const (
	InstructionCreateAccount uint8 = 0
	InstructionTransfer      uint8 = 1
	InstructionCloseAccount  uint8 = 2
	InstructionAllocateData  uint8 = 3
)

// SystemProgram implements the built-in system program for account management
// This implements Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.2, 10.3, 10.4, 10.5
type SystemProgram struct{}

// NewSystemProgram creates a new instance of the System Program
func NewSystemProgram() *SystemProgram {
	return &SystemProgram{}
}

// GetProgramID returns the unique identifier for the System Program
// This implements the BuiltinProgram interface
func (sp *SystemProgram) GetProgramID() filestore.FileID {
	return SystemProgramID
}

// Execute processes a system program instruction
// This implements the BuiltinProgram interface and dispatches to specific instruction handlers
func (sp *SystemProgram) Execute(ctx *runtime.ExecutionContext) error {
	if ctx == nil {
		return fmt.Errorf("execution context cannot be nil")
	}

	// Get instruction data
	data := ctx.GetInstructionData()
	if len(data) == 0 {
		return fmt.Errorf("instruction data is empty")
	}

	// First byte is the instruction type
	instructionType := data[0]
	instructionData := data[1:]

	// Dispatch to appropriate handler
	switch instructionType {
	case InstructionCreateAccount:
		return sp.CreateAccount(ctx, instructionData)
	case InstructionTransfer:
		return sp.Transfer(ctx, instructionData)
	case InstructionCloseAccount:
		return sp.CloseAccount(ctx, instructionData)
	case InstructionAllocateData:
		return sp.AllocateData(ctx, instructionData)
	default:
		return fmt.Errorf("unknown system instruction type: %d", instructionType)
	}
}

// CreateAccount creates a new user account file
// This implements Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 10.1, 10.4
//
// Instruction data format:
// - bytes 0-31: owner public key (32 bytes)
// - bytes 32-39: initial balance (8 bytes, int64, little-endian)
//
// Expected inputs:
// - "new_account": Write permission to the new account file (must be pre-generated)
func (sp *SystemProgram) CreateAccount(ctx *runtime.ExecutionContext, data []byte) error {
	// Decode instruction data
	if len(data) < 40 {
		return fmt.Errorf("invalid CreateAccount data: expected at least 40 bytes, got %d", len(data))
	}

	// Extract owner public key
	var ownerPubKey transaction.PublicKey
	copy(ownerPubKey[:], data[0:32])

	// Extract initial balance
	initialBalance := int64(binary.LittleEndian.Uint64(data[32:40]))

	if initialBalance < 0 {
		return fmt.Errorf("initial balance cannot be negative: %d", initialBalance)
	}

	// Get the new account file ID from inputs
	newAccountID, err := ctx.GetInputFileID("new_account")
	if err != nil {
		return fmt.Errorf("failed to get new_account input: %w", err)
	}

	// Verify the signer is authorized (owner must sign)
	if !ctx.HasSigner(ownerPubKey) {
		return fmt.Errorf("account creation not authorized: owner must sign transaction")
	}

	// Create the account file
	// User accounts have empty data, executable=false, and tx_manager=SystemProgramID
	now := time.Now()
	accountFile := &filestore.File{
		ID:         newAccountID,
		Balance:    initialBalance,
		TxManager:  SystemProgramID,
		Data:       []byte{}, // User accounts have empty data (Requirement 5.3)
		Executable: false,    // User accounts are not executable (Requirement 5.4)
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Validate write access for the new account
	if err := ctx.AccessController.ValidateAndRecord(newAccountID, transaction.Write); err != nil {
		return fmt.Errorf("access validation failed for new account: %w", err)
	}

	// Store the file using CreateFile (this is a new file)
	_, err = ctx.FileStore.CreateFile(accountFile)
	if err != nil {
		return fmt.Errorf("failed to create account file: %w", err)
	}

	return nil
}

// Transfer transfers balance between two accounts
// This implements Requirements 10.2, 10.4
//
// Instruction data format:
// - bytes 0-7: amount to transfer (8 bytes, int64, little-endian)
//
// Expected inputs:
// - "from": Write permission to the source account
// - "to": Write permission to the destination account
func (sp *SystemProgram) Transfer(ctx *runtime.ExecutionContext, data []byte) error {
	// Decode instruction data
	if len(data) < 8 {
		return fmt.Errorf("invalid Transfer data: expected at least 8 bytes, got %d", len(data))
	}

	// Extract amount
	amount := int64(binary.LittleEndian.Uint64(data[0:8]))

	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive: %d", amount)
	}

	// Get source and destination account IDs from inputs
	fromID, err := ctx.GetInputFileID("from")
	if err != nil {
		return fmt.Errorf("failed to get from input: %w", err)
	}

	toID, err := ctx.GetInputFileID("to")
	if err != nil {
		return fmt.Errorf("failed to get to input: %w", err)
	}

	// Load source account (with write permission)
	fromAccount, err := ctx.GetFileMut(fromID)
	if err != nil {
		return fmt.Errorf("failed to load source account: %w", err)
	}

	// Verify source account has sufficient balance
	if fromAccount.Balance < amount {
		return fmt.Errorf("insufficient balance: have %d, need %d", fromAccount.Balance, amount)
	}

	// Load destination account (with write permission)
	toAccount, err := ctx.GetFileMut(toID)
	if err != nil {
		return fmt.Errorf("failed to load destination account: %w", err)
	}

	// Perform the transfer
	fromAccount.Balance -= amount
	toAccount.Balance += amount

	// Update both accounts
	if err := ctx.UpdateFile(fromAccount); err != nil {
		return fmt.Errorf("failed to update source account: %w", err)
	}

	if err := ctx.UpdateFile(toAccount); err != nil {
		return fmt.Errorf("failed to update destination account: %w", err)
	}

	return nil
}

// CloseAccount closes an account and transfers remaining balance to a destination
// This implements Requirements 10.3, 10.4
//
// Instruction data format:
// - No additional data required (all info from inputs)
//
// Expected inputs:
// - "account": Write permission to the account being closed
// - "destination": Write permission to the account receiving the balance
func (sp *SystemProgram) CloseAccount(ctx *runtime.ExecutionContext, data []byte) error {
	// Get account and destination IDs from inputs
	accountID, err := ctx.GetInputFileID("account")
	if err != nil {
		return fmt.Errorf("failed to get account input: %w", err)
	}

	destinationID, err := ctx.GetInputFileID("destination")
	if err != nil {
		return fmt.Errorf("failed to get destination input: %w", err)
	}

	// Cannot close account to itself
	if accountID == destinationID {
		return fmt.Errorf("cannot close account to itself")
	}

	// Load account being closed (with write permission)
	account, err := ctx.GetFileMut(accountID)
	if err != nil {
		return fmt.Errorf("failed to load account: %w", err)
	}

	// Load destination account (with write permission)
	destination, err := ctx.GetFileMut(destinationID)
	if err != nil {
		return fmt.Errorf("failed to load destination account: %w", err)
	}

	// Transfer remaining balance to destination
	remainingBalance := account.Balance
	if remainingBalance > 0 {
		destination.Balance += remainingBalance
		if err := ctx.UpdateFile(destination); err != nil {
			return fmt.Errorf("failed to update destination account: %w", err)
		}
	}

	// Delete the account
	// Note: We use the FileStore directly here since deletion is not a standard update
	if err := ctx.FileStore.DeleteFile(accountID); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// AllocateData allocates data space for an account
// This implements Requirements 6.1, 6.2, 6.3, 6.4, 6.5, 10.4
//
// Instruction data format:
// - bytes 0-7: size of data to allocate (8 bytes, int64, little-endian)
//
// Expected inputs:
// - "account": Write permission to the account
func (sp *SystemProgram) AllocateData(ctx *runtime.ExecutionContext, data []byte) error {
	// Decode instruction data
	if len(data) < 8 {
		return fmt.Errorf("invalid AllocateData data: expected at least 8 bytes, got %d", len(data))
	}

	// Extract size
	size := int64(binary.LittleEndian.Uint64(data[0:8]))

	if size < 0 {
		return fmt.Errorf("data size cannot be negative: %d", size)
	}

	// Get account ID from inputs
	accountID, err := ctx.GetInputFileID("account")
	if err != nil {
		return fmt.Errorf("failed to get account input: %w", err)
	}

	// Load account (with write permission)
	account, err := ctx.GetFileMut(accountID)
	if err != nil {
		return fmt.Errorf("failed to load account: %w", err)
	}

	// Allocate data space (resize the data field)
	newData := make([]byte, size)
	// Copy existing data if any
	if len(account.Data) > 0 {
		copyLen := len(account.Data)
		if copyLen > int(size) {
			copyLen = int(size)
		}
		copy(newData, account.Data[:copyLen])
	}
	account.Data = newData

	// Validate storage cost (Requirement 6.1, 6.2, 6.3, 6.4)
	// UpdateFile will validate that the balance covers the new storage cost
	if err := ctx.UpdateFile(account); err != nil {
		return fmt.Errorf("failed to allocate data: %w", err)
	}

	return nil
}

// EncodeCreateAccountInstruction encodes a CreateAccount instruction
// This is a helper function for creating properly formatted instruction data
func EncodeCreateAccountInstruction(owner transaction.PublicKey, initialBalance int64) []byte {
	data := make([]byte, 41) // 1 byte type + 32 bytes pubkey + 8 bytes balance
	data[0] = InstructionCreateAccount
	copy(data[1:33], owner[:])
	binary.LittleEndian.PutUint64(data[33:41], uint64(initialBalance))
	return data
}

// EncodeTransferInstruction encodes a Transfer instruction
// This is a helper function for creating properly formatted instruction data
func EncodeTransferInstruction(amount int64) []byte {
	data := make([]byte, 9) // 1 byte type + 8 bytes amount
	data[0] = InstructionTransfer
	binary.LittleEndian.PutUint64(data[1:9], uint64(amount))
	return data
}

// EncodeCloseAccountInstruction encodes a CloseAccount instruction
// This is a helper function for creating properly formatted instruction data
func EncodeCloseAccountInstruction() []byte {
	return []byte{InstructionCloseAccount}
}

// EncodeAllocateDataInstruction encodes an AllocateData instruction
// This is a helper function for creating properly formatted instruction data
func EncodeAllocateDataInstruction(size int64) []byte {
	data := make([]byte, 9) // 1 byte type + 8 bytes size
	data[0] = InstructionAllocateData
	binary.LittleEndian.PutUint64(data[1:9], uint64(size))
	return data
}

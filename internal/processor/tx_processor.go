package processor

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
)

// Package processor implements the TxProcessor for transaction execution,
// orchestrating the full transaction flow including validation, fee payment,
// instruction execution, and state management.

// TxResult represents the result of transaction execution
type TxResult struct {
	TxID      transaction.TxID
	Success   bool
	Error     error
	GasUsed   int64
	StateRoot []byte
}

// TxProcessor executes transactions and manages state transitions
// This implements Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 7.1, 7.2, 7.3, 7.4, 7.5, 8.1, 8.2, 8.3, 8.4, 8.5
type TxProcessor struct {
	// fileStore provides access to file state
	fileStore *filestore.FileStore
	// runtime executes program bytecode
	runtime *runtime.Runtime
	// feeConfig holds fee calculation parameters
	feeConfig transaction.FeeConfig
	// stateCache holds transaction-local state changes for atomicity
	stateCache map[filestore.FileID]*filestore.File
	// inputValidator validates instruction inputs before execution
	inputValidator *InputValidator
	// mu protects concurrent access during transaction execution
	mu sync.Mutex
}

// NewTxProcessor creates a new transaction processor
func NewTxProcessor(fileStore *filestore.FileStore, rt *runtime.Runtime) *TxProcessor {
	return &TxProcessor{
		fileStore:      fileStore,
		runtime:        rt,
		feeConfig:      transaction.DefaultFeeConfig(),
		stateCache:     make(map[filestore.FileID]*filestore.File),
		inputValidator: NewInputValidator(fileStore),
	}
}

// SetFeeConfig updates the fee configuration
func (tp *TxProcessor) SetFeeConfig(config transaction.FeeConfig) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.feeConfig = config
}

// GetFeeConfig returns the current fee configuration
func (tp *TxProcessor) GetFeeConfig() transaction.FeeConfig {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	return tp.feeConfig
}

// ProcessTransaction orchestrates the full transaction execution flow
// This implements Requirements 3.5, 7.1, 7.5, 8.5
func (tp *TxProcessor) ProcessTransaction(tx *transaction.Transaction) (*TxResult, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Initialize transaction-local state cache
	tp.stateCache = make(map[filestore.FileID]*filestore.File)
	originalState := make(map[filestore.FileID]*filestore.File)

	// Generate transaction ID
	txData, err := tx.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}
	txID := sha256.Sum256(txData)

	result := &TxResult{
		TxID:    txID,
		Success: false,
	}

	// Step 1: Validate transaction (Requirement 3.5)
	if err := tp.ValidateTransaction(tx); err != nil {
		result.Error = fmt.Errorf("transaction validation failed: %w", err)
		return result, result.Error
	}

	// Step 2: Calculate and deduct fee (Requirements 7.1, 7.2, 7.3, 7.4, 7.5)
	fee := transaction.CalculateFee(tx, tp.feeConfig)
	feePayer, err := tx.GetFeePayer()
	if err != nil {
		result.Error = fmt.Errorf("failed to get fee payer: %w", err)
		return result, result.Error
	}

	feePayerID := tp.publicKeyToFileID(feePayer)

	// Save original state for rollback
	if original, err := tp.fileStore.GetFile(feePayerID); err == nil {
		originalCopy := *original
		originalState[feePayerID] = &originalCopy
	}

	if err := tp.DeductFee(feePayerID, fee); err != nil {
		result.Error = fmt.Errorf("fee deduction failed: %w", err)
		tp.revertToState(originalState)
		return result, result.Error
	}

	// Step 3: Collect all files that will be accessed
	accessedFiles := make(map[filestore.FileID]bool)
	for _, instr := range tx.Instructions {
		for _, fileAccess := range instr.Inputs {
			accessedFiles[fileAccess.FileID] = true
		}
	}

	// Save original state of all accessed files
	for fileID := range accessedFiles {
		if original, err := tp.fileStore.GetFile(fileID); err == nil {
			originalCopy := *original
			originalState[fileID] = &originalCopy
		}
	}

	// Step 4: Execute instructions sequentially (Requirement 3.5)
	signers := tx.GetSigners()
	for i, instr := range tx.Instructions {
		if err := tp.ExecuteInstruction(&instr, signers); err != nil {
			result.Error = fmt.Errorf("instruction %d execution failed: %w", i, err)
			tp.revertToState(originalState)
			return result, result.Error
		}
	}

	// Success - state is already committed to FileStore during execution
	result.Success = true
	result.GasUsed = fee
	result.StateRoot = tp.calculateStateRoot()

	// Clear cache
	tp.stateCache = make(map[filestore.FileID]*filestore.File)

	return result, nil
}

// ValidateTransaction performs pre-execution validation checks
// This implements Requirements 3.3, 7.4
func (tp *TxProcessor) ValidateTransaction(tx *transaction.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	// Validate that transaction has at least one signature (Requirement 3.3)
	if len(tx.Signatures) == 0 {
		return fmt.Errorf("transaction must have at least one signature")
	}

	// Validate that transaction has at least one instruction
	if len(tx.Instructions) == 0 {
		return fmt.Errorf("transaction must have at least one instruction")
	}

	// Validate all signatures (Requirement 3.3)
	// Temporarily remove signatures for verification
	savedSignatures := tx.Signatures
	tx.Signatures = []transaction.Signature{}

	txData, err := tx.Marshal()
	if err != nil {
		tx.Signatures = savedSignatures // Restore
		return fmt.Errorf("failed to marshal transaction for signature verification: %w", err)
	}

	// Restore signatures
	tx.Signatures = savedSignatures

	for i, sig := range tx.Signatures {
		if !sig.Verify(txData) {
			return fmt.Errorf("invalid signature at index %d", i)
		}
	}

	// Validate fee payer has sufficient balance (Requirement 7.4)
	feePayer, err := tx.GetFeePayer()
	if err != nil {
		return fmt.Errorf("failed to get fee payer: %w", err)
	}

	feePayerID := tp.publicKeyToFileID(feePayer)
	feePayerFile, err := tp.fileStore.GetFile(feePayerID)
	if err != nil {
		return fmt.Errorf("fee payer account not found: %w", err)
	}

	fee := transaction.CalculateFee(tx, tp.feeConfig)
	if feePayerFile.Balance < fee {
		return fmt.Errorf("insufficient balance for fee: required %d, have %d", fee, feePayerFile.Balance)
	}

	return nil
}

// DeductFee charges the fee payer before instruction execution
// This implements Requirements 7.1, 7.5
func (tp *TxProcessor) DeductFee(feePayerID filestore.FileID, fee int64) error {
	if fee < 0 {
		return fmt.Errorf("fee cannot be negative")
	}

	// Get fee payer file from store
	feePayerFile, err := tp.fileStore.GetFile(feePayerID)
	if err != nil {
		return fmt.Errorf("failed to load fee payer account: %w", err)
	}

	// Check sufficient balance
	if feePayerFile.Balance < fee {
		return fmt.Errorf("insufficient balance for fee: required %d, have %d", fee, feePayerFile.Balance)
	}

	// Deduct fee and update in store immediately
	feePayerFile.Balance -= fee
	if err := tp.fileStore.UpdateFile(feePayerID, feePayerFile); err != nil {
		return fmt.Errorf("failed to update fee payer balance: %w", err)
	}

	return nil
}

// ExecuteInstruction processes a single instruction with access control
// This implements Requirements 3.5, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5
func (tp *TxProcessor) ExecuteInstruction(instr *transaction.Instruction, signers []transaction.PublicKey) error {
	if instr == nil {
		return fmt.Errorf("instruction cannot be nil")
	}

	// Validate instruction inputs before execution (Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 5.1, 5.2, 5.3, 5.4, 5.5)
	if err := tp.inputValidator.ValidateInstructionInputs(instr); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	// Load program file
	programFile, err := tp.fileStore.GetFile(instr.ProgramID)
	if err != nil {
		return fmt.Errorf("failed to load program %s: %w", instr.ProgramID.String(), err)
	}

	// Validate program is executable
	if !programFile.Executable {
		return fmt.Errorf("file %s is not executable", instr.ProgramID.String())
	}

	// Validate all input files exist (Requirement 4.5)
	for key, fileAccess := range instr.Inputs {
		if _, err := tp.fileStore.GetFile(fileAccess.FileID); err != nil {
			return fmt.Errorf("input file %s (key: %s) not found: %w", fileAccess.FileID.String(), key, err)
		}
	}

	// Create access controller and set declared access (Requirement 8.1)
	accessController := access.NewAccessController()
	accessController.SetDeclaredAccess(instr.Inputs)

	// Create a transaction-aware execution context
	ctx := tp.newTransactionExecutionContext(instr, signers, accessController)

	// Execute program via runtime (Requirement 3.5)
	if err := tp.runtime.ExecuteProgram(programFile, ctx); err != nil {
		return fmt.Errorf("program execution failed: %w", err)
	}

	// Validate that all accesses matched declared permissions (Requirement 8.1, 8.5)
	if err := accessController.ValidateFinalAccess(); err != nil {
		return fmt.Errorf("access validation failed: %w", err)
	}

	return nil
}

// RevertState rolls back all state changes by restoring from saved state
func (tp *TxProcessor) RevertState() {
	// This is now handled by revertToState in ProcessTransaction
	// Clear the cache
	tp.stateCache = make(map[filestore.FileID]*filestore.File)
}

// revertToState restores files to their original state
func (tp *TxProcessor) revertToState(originalState map[filestore.FileID]*filestore.File) {
	// Restore all files to their original state
	for fileID, originalFile := range originalState {
		// Ignore errors during rollback - best effort
		_ = tp.fileStore.UpdateFile(fileID, originalFile)
	}

	// Clear cache
	tp.stateCache = make(map[filestore.FileID]*filestore.File)
}

// newTransactionExecutionContext creates an execution context that uses the transaction cache
func (tp *TxProcessor) newTransactionExecutionContext(
	instr *transaction.Instruction,
	signers []transaction.PublicKey,
	accessController *access.AccessController,
) *runtime.ExecutionContext {
	// For now, we pass the real FileStore
	// The transaction cache is managed at the processor level
	// Programs will modify files in the store, and we'll track changes for rollback
	return &runtime.ExecutionContext{
		Instruction:      instr,
		Signers:          signers,
		FileStore:        tp.fileStore,
		AccessController: accessController,
	}
}

// publicKeyToFileID converts a public key to a file ID
// This is a placeholder implementation - actual mapping may differ
func (tp *TxProcessor) publicKeyToFileID(pubkey transaction.PublicKey) filestore.FileID {
	// Use the public key bytes directly as the file ID
	var fileID filestore.FileID
	copy(fileID[:], pubkey[:])
	return fileID
}

// calculateStateRoot computes a state root hash from current state
func (tp *TxProcessor) calculateStateRoot() []byte {
	// Simple implementation: hash all file IDs in cache
	// A production implementation would use a Merkle tree
	hasher := sha256.New()
	for fileID := range tp.stateCache {
		hasher.Write(fileID[:])
	}
	root := hasher.Sum(nil)
	return root
}

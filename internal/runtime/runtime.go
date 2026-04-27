package runtime

import (
	"fmt"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/transaction"
)

// Package runtime provides the execution environment for programs,
// including bytecode interpretation and built-in program registry.

// ExecutionContext provides the execution environment for program instructions
// This implements Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5
type ExecutionContext struct {
	// Instruction being executed
	Instruction *transaction.Instruction
	// Signers who signed the transaction
	Signers []transaction.PublicKey
	// FileStore for accessing and modifying files
	FileStore *filestore.FileStore
	// AccessController for validating file access permissions
	AccessController *access.AccessController
}

// NewExecutionContext creates a new execution context for an instruction
func NewExecutionContext(
	instruction *transaction.Instruction,
	signers []transaction.PublicKey,
	fileStore *filestore.FileStore,
	accessController *access.AccessController,
) *ExecutionContext {
	return &ExecutionContext{
		Instruction:      instruction,
		Signers:          signers,
		FileStore:        fileStore,
		AccessController: accessController,
	}
}

// GetFile loads a file with read permission validation
// This implements Requirements 4.1, 4.2, 8.1, 8.2
func (ctx *ExecutionContext) GetFile(fileID filestore.FileID) (*filestore.File, error) {
	// Validate and record read access (Requirement 8.1, 8.2)
	if err := ctx.AccessController.ValidateAndRecord(fileID, transaction.Read); err != nil {
		return nil, fmt.Errorf("access validation failed for file %s: %w", fileID.String(), err)
	}

	// Load file from store
	file, err := ctx.FileStore.GetFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileID.String(), err)
	}

	return file, nil
}

// GetFileMut loads a file with write permission validation
// This implements Requirements 4.1, 4.2, 8.1, 8.2
func (ctx *ExecutionContext) GetFileMut(fileID filestore.FileID) (*filestore.File, error) {
	// Validate and record write access (Requirement 8.1, 8.2)
	if err := ctx.AccessController.ValidateAndRecord(fileID, transaction.Write); err != nil {
		return nil, fmt.Errorf("access validation failed for file %s: %w", fileID.String(), err)
	}

	// Load file from store
	file, err := ctx.FileStore.GetFile(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileID.String(), err)
	}

	return file, nil
}

// UpdateFile updates a file with write permission validation
// This implements Requirements 4.2, 8.1, 8.2, 8.3
func (ctx *ExecutionContext) UpdateFile(file *filestore.File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	// Validate and record write access (Requirement 8.1, 8.2)
	if err := ctx.AccessController.ValidateAndRecord(file.ID, transaction.Write); err != nil {
		return fmt.Errorf("access validation failed for file %s: %w", file.ID.String(), err)
	}

	// Update file in store
	if err := ctx.FileStore.UpdateFile(file.ID, file); err != nil {
		return fmt.Errorf("failed to update file %s: %w", file.ID.String(), err)
	}

	return nil
}

// CreateFile creates a new file with write permission validation
// This implements Requirements 4.2, 8.1, 8.2, 8.3
func (ctx *ExecutionContext) CreateFile(file *filestore.File) error {
	if file == nil {
		return fmt.Errorf("file cannot be nil")
	}

	// Validate and record write access (Requirement 8.1, 8.2)
	if err := ctx.AccessController.ValidateAndRecord(file.ID, transaction.Write); err != nil {
		return fmt.Errorf("access validation failed for file %s: %w", file.ID.String(), err)
	}

	// Create file in store
	if _, err := ctx.FileStore.CreateFile(file); err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.ID.String(), err)
	}

	return nil
}

// DeleteFile deletes a file with write permission validation
// This implements Requirements 4.2, 8.1, 8.2, 8.3
func (ctx *ExecutionContext) DeleteFile(fileID filestore.FileID) error {
	// Validate and record write access (Requirement 8.1, 8.2)
	if err := ctx.AccessController.ValidateAndRecord(fileID, transaction.Write); err != nil {
		return fmt.Errorf("access validation failed for file %s: %w", fileID.String(), err)
	}

	// Delete file from store
	if err := ctx.FileStore.DeleteFile(fileID); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileID.String(), err)
	}

	return nil
}

// GetFileBalance retrieves a file's balance with read permission validation
// This is a convenience method for balance queries
func (ctx *ExecutionContext) GetFileBalance(fileID filestore.FileID) (int64, error) {
	// Validate and record read access
	if err := ctx.AccessController.ValidateAndRecord(fileID, transaction.Read); err != nil {
		return 0, fmt.Errorf("access validation failed for file %s: %w", fileID.String(), err)
	}

	// Get balance from store
	balance, err := ctx.FileStore.GetFileBalance(fileID)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance for file %s: %w", fileID.String(), err)
	}

	return balance, nil
}

// UpdateFileBalance updates a file's balance with write permission validation
// This is a convenience method for balance modifications
func (ctx *ExecutionContext) UpdateFileBalance(fileID filestore.FileID, delta int64) error {
	// Validate and record write access
	if err := ctx.AccessController.ValidateAndRecord(fileID, transaction.Write); err != nil {
		return fmt.Errorf("access validation failed for file %s: %w", fileID.String(), err)
	}

	// Update balance in store
	if err := ctx.FileStore.UpdateFileBalance(fileID, delta); err != nil {
		return fmt.Errorf("failed to update balance for file %s: %w", fileID.String(), err)
	}

	return nil
}

// HasSigner checks if a specific public key is among the transaction signers
// This is useful for authorization checks within programs
func (ctx *ExecutionContext) HasSigner(pubkey transaction.PublicKey) bool {
	for _, signer := range ctx.Signers {
		if signer == pubkey {
			return true
		}
	}
	return false
}

// GetInstructionData returns the instruction data for the current execution
// This is a convenience method for programs to access instruction parameters
func (ctx *ExecutionContext) GetInstructionData() []byte {
	if ctx.Instruction == nil {
		return nil
	}
	return ctx.Instruction.Data
}

// GetProgramID returns the program ID being executed
// This is a convenience method for programs to identify themselves
func (ctx *ExecutionContext) GetProgramID() filestore.FileID {
	if ctx.Instruction == nil {
		return filestore.FileID{}
	}
	return ctx.Instruction.ProgramID
}

// GetDeclaredInputs returns the declared file inputs for the instruction
// This is useful for programs that need to iterate over their inputs
func (ctx *ExecutionContext) GetDeclaredInputs() map[string]transaction.FileAccess {
	if ctx.Instruction == nil {
		return nil
	}
	return ctx.Instruction.Inputs
}

// GetInputFileID retrieves a file ID from the instruction inputs by key
// This is a convenience method for accessing specific input files
func (ctx *ExecutionContext) GetInputFileID(key string) (filestore.FileID, error) {
	if ctx.Instruction == nil {
		return filestore.FileID{}, fmt.Errorf("no instruction in context")
	}

	fileAccess, ok := ctx.Instruction.Inputs[key]
	if !ok {
		return filestore.FileID{}, fmt.Errorf("input key %s not found in instruction inputs", key)
	}

	return fileAccess.FileID, nil
}

// BuiltinProgram defines the interface for built-in programs
// This implements Requirements 2.1, 2.2, 2.3, 2.4, 2.5
type BuiltinProgram interface {
	// Execute runs the program with the given execution context
	Execute(ctx *ExecutionContext) error
	// GetProgramID returns the unique identifier for this builtin program
	GetProgramID() filestore.FileID
}

// Runtime manages program execution and the builtin program registry
// This implements Requirements 2.1, 2.2, 2.3, 2.4, 2.5
type Runtime struct {
	// builtinPrograms maps program IDs to their implementations
	builtinPrograms map[filestore.FileID]BuiltinProgram
	// executionLimit is the maximum computation units per instruction
	executionLimit int64
}

// NewRuntime creates a new Runtime instance with an empty builtin program registry
// This implements Requirement 2.4
func NewRuntime() *Runtime {
	return &Runtime{
		builtinPrograms: make(map[filestore.FileID]BuiltinProgram),
		executionLimit:  1000000, // Default: 1 million compute units
	}
}

// RegisterBuiltinProgram adds a builtin program to the runtime registry
// This allows the runtime to dispatch to native Go implementations
func (r *Runtime) RegisterBuiltinProgram(program BuiltinProgram) error {
	if program == nil {
		return fmt.Errorf("program cannot be nil")
	}

	programID := program.GetProgramID()
	if _, exists := r.builtinPrograms[programID]; exists {
		return fmt.Errorf("builtin program %s already registered", programID.String())
	}

	r.builtinPrograms[programID] = program
	return nil
}

// ExecuteProgram executes a program with the given execution context
// This implements Requirements 2.2, 2.3, 2.5
func (r *Runtime) ExecuteProgram(program *filestore.File, ctx *ExecutionContext) error {
	if program == nil {
		return fmt.Errorf("program cannot be nil")
	}

	if ctx == nil {
		return fmt.Errorf("execution context cannot be nil")
	}

	// Validate the program before execution (Requirement 2.1)
	if err := r.ValidateProgram(program); err != nil {
		return fmt.Errorf("program validation failed: %w", err)
	}

	// Check if this is a builtin program (Requirement 2.5)
	if builtinProgram, isBuiltin := r.builtinPrograms[program.ID]; isBuiltin {
		// Execute builtin program using native Go implementation
		if err := builtinProgram.Execute(ctx); err != nil {
			return fmt.Errorf("builtin program execution failed: %w", err)
		}
		return nil
	}

	// Check if this is QuanticScript bytecode
	if r.isQuanticScriptBytecode(program.Data) {
		// Execute QuanticScript bytecode
		if err := r.executeBytecode(program.Data, ctx); err != nil {
			return fmt.Errorf("bytecode execution failed: %w", err)
		}
		return nil
	}

	// Unknown program format
	return fmt.Errorf("unsupported program format for program %s", program.ID.String())
}

// ValidateProgram performs basic validation on a program file
// This implements Requirement 2.1
func (r *Runtime) ValidateProgram(program *filestore.File) error {
	if program == nil {
		return fmt.Errorf("program cannot be nil")
	}

	// Verify the file is marked as executable (Requirement 2.1)
	if !program.Executable {
		return fmt.Errorf("file %s is not executable", program.ID.String())
	}

	// Verify the program has a valid transaction manager
	// For programs, the TxManager should be the Runtime program ID (Requirement 2.2)
	// Note: We don't validate the specific TxManager here as it's set during program creation

	// For builtin programs, no bytecode validation needed
	if _, isBuiltin := r.builtinPrograms[program.ID]; isBuiltin {
		return nil
	}

	// For non-builtin programs, validate bytecode exists
	if len(program.Data) == 0 {
		return fmt.Errorf("program %s has no bytecode", program.ID.String())
	}

	// Future: Add bytecode format validation here
	// For now, we just check that data exists

	return nil
}

// GetBuiltinProgram retrieves a builtin program by its ID
// Returns nil if the program is not a builtin
func (r *Runtime) GetBuiltinProgram(programID filestore.FileID) BuiltinProgram {
	return r.builtinPrograms[programID]
}

// IsBuiltinProgram checks if a program ID corresponds to a builtin program
func (r *Runtime) IsBuiltinProgram(programID filestore.FileID) bool {
	_, exists := r.builtinPrograms[programID]
	return exists
}

// SetExecutionLimit sets the maximum computation units per instruction
func (r *Runtime) SetExecutionLimit(limit int64) {
	r.executionLimit = limit
}

// GetExecutionLimit returns the current execution limit
func (r *Runtime) GetExecutionLimit() int64 {
	return r.executionLimit
}

// isQuanticScriptBytecode checks if the program data is QuanticScript bytecode
func (r *Runtime) isQuanticScriptBytecode(data []byte) bool {
	return quanticscript.IsQuanticScriptBytecode(data)
}

// executeBytecode executes QuanticScript bytecode with the interpreter
func (r *Runtime) executeBytecode(bytecode []byte, ctx *ExecutionContext) error {
	// Parse bytecode header
	header, err := quanticscript.ParseBytecodeHeader(bytecode)
	if err != nil {
		return fmt.Errorf("failed to parse bytecode header: %w", err)
	}

	// Get bytecode body (without header)
	body, err := quanticscript.GetBytecodeBody(bytecode)
	if err != nil {
		return fmt.Errorf("failed to get bytecode body: %w", err)
	}

	// Validate entry offset
	if int(header.EntryOffset) >= len(body) {
		return fmt.Errorf("invalid entry offset: %d", header.EntryOffset)
	}

	// Create interpreter with compute budget
	interpreter := quanticscript.NewBytecodeInterpreter(body, ctx, r.executionLimit)

	// Execute bytecode starting from entry point
	if err := interpreter.Execute(); err != nil {
		return fmt.Errorf("bytecode execution error: %w", err)
	}

	return nil
}

// GetSigners returns the list of transaction signers
func (ctx *ExecutionContext) GetSigners() []transaction.PublicKey {
	return ctx.Signers
}

// QueryBlock queries a finalized block by hash
// Returns serialized block data or error if not found/not finalized
func (ctx *ExecutionContext) QueryBlock(blockHash []byte) ([]byte, error) {
	// TODO: Implement block query from finalized blockchain state
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("block query not yet implemented")
}

// QueryTransaction queries a finalized transaction by ID
// Returns serialized transaction data or error if not found/not finalized
func (ctx *ExecutionContext) QueryTransaction(txID transaction.TxID) ([]byte, error) {
	// TODO: Implement transaction query from finalized blockchain state
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("transaction query not yet implemented")
}

// QueryInstruction queries a finalized instruction by reference
// Returns serialized instruction data or error if not found/not finalized
func (ctx *ExecutionContext) QueryInstruction(txID transaction.TxID, instrIndex uint32) ([]byte, error) {
	// TODO: Implement instruction query from finalized blockchain state
	// For now, return error indicating not implemented
	return nil, fmt.Errorf("instruction query not yet implemented")
}

// InvokeProgram invokes another program with the given data and compute budget
// This implements Requirements 9.1, 9.2, 9.3, 9.4, 9.5
func (ctx *ExecutionContext) InvokeProgram(programID filestore.FileID, invokeData []byte, computeBudget int64, depth int) ([]byte, error) {
	// Check invocation depth limit (Requirement 9.4)
	if depth > 4 {
		return nil, fmt.Errorf("maximum cross-program invocation depth exceeded: %d", depth)
	}

	// Validate that the target program is in the declared program list (Requirement 9.2)
	declaredPrograms := ctx.GetDeclaredPrograms()
	isProgramDeclared := false
	for _, declaredProgramID := range declaredPrograms {
		if declaredProgramID == programID {
			isProgramDeclared = true
			break
		}
	}

	if !isProgramDeclared {
		return nil, fmt.Errorf("program %s not in declared program list", programID.String())
	}

	// Load the target program
	program, err := ctx.FileStore.GetFile(programID)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoked program %s: %w", programID.String(), err)
	}

	// Validate the program is executable
	if !program.Executable {
		return nil, fmt.Errorf("program %s is not executable", programID.String())
	}

	// Create a new execution context for the invoked program
	// The invoked program gets the same signers and file store, but different instruction data
	invokedInstruction := &transaction.Instruction{
		ProgramID: programID,
		Data:      invokeData,
		Inputs:    ctx.Instruction.Inputs, // Share the same inputs
	}

	invokedCtx := &ExecutionContext{
		Instruction:      invokedInstruction,
		Signers:          ctx.Signers,
		FileStore:        ctx.FileStore,
		AccessController: ctx.AccessController,
	}

	// Check if this is QuanticScript bytecode
	if !quanticscript.IsQuanticScriptBytecode(program.Data) {
		return nil, fmt.Errorf("invoked program %s is not QuanticScript bytecode", programID.String())
	}

	// Parse bytecode header
	header, err := quanticscript.ParseBytecodeHeader(program.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bytecode header: %w", err)
	}

	// Get bytecode body (without header)
	body, err := quanticscript.GetBytecodeBody(program.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytecode body: %w", err)
	}

	// Validate entry offset
	if int(header.EntryOffset) >= len(body) {
		return nil, fmt.Errorf("invalid entry offset: %d", header.EntryOffset)
	}

	// Create interpreter with the allocated compute budget and increased depth
	interpreter := quanticscript.NewBytecodeInterpreterWithDepth(body, invokedCtx, computeBudget, depth)

	// Execute bytecode starting from entry point
	if err := interpreter.Execute(); err != nil {
		// Invocation failed, rollback is handled by the caller (Requirement 9.5)
		return nil, fmt.Errorf("invoked program execution error: %w", err)
	}

	// For now, return empty result data
	// In a full implementation, the invoked program would push result data onto the stack
	// and we would extract it here
	return []byte{}, nil
}

// GetDeclaredPrograms returns the list of programs declared in the instruction
// This is used to validate cross-program invocations (Requirement 9.2)
func (ctx *ExecutionContext) GetDeclaredPrograms() []filestore.FileID {
	if ctx.Instruction == nil {
		return nil
	}

	// For now, we'll extract program IDs from the instruction inputs
	// In a full implementation, there would be a separate field for declared programs
	programs := make([]filestore.FileID, 0)

	// Add the current program ID
	programs = append(programs, ctx.Instruction.ProgramID)

	// Add all file IDs from inputs (they could be programs)
	for _, fileAccess := range ctx.Instruction.Inputs {
		programs = append(programs, fileAccess.FileID)
	}

	return programs
}

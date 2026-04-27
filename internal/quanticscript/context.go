package quanticscript

import (
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// ExecutionContext defines the interface for program execution context
// This breaks the import cycle between quanticscript and runtime packages
type ExecutionContext interface {
	// GetFile loads a file with read permission validation
	GetFile(fileID filestore.FileID) (*filestore.File, error)

	// GetFileMut loads a file with write permission validation
	GetFileMut(fileID filestore.FileID) (*filestore.File, error)

	// UpdateFile updates a file with write permission validation
	UpdateFile(file *filestore.File) error

	// CreateFile creates a new file with write permission validation
	CreateFile(file *filestore.File) error

	// DeleteFile deletes a file with write permission validation
	DeleteFile(fileID filestore.FileID) error

	// GetFileBalance retrieves a file's balance with read permission validation
	GetFileBalance(fileID filestore.FileID) (int64, error)

	// UpdateFileBalance updates a file's balance with write permission validation
	UpdateFileBalance(fileID filestore.FileID, delta int64) error

	// HasSigner checks if a specific public key is among the transaction signers
	HasSigner(pubkey transaction.PublicKey) bool

	// GetInstructionData returns the instruction data for the current execution
	GetInstructionData() []byte

	// GetProgramID returns the program ID being executed
	GetProgramID() filestore.FileID

	// GetSigners returns the list of transaction signers
	GetSigners() []transaction.PublicKey

	// QueryBlock queries a finalized block by hash
	QueryBlock(blockHash []byte) ([]byte, error)

	// QueryTransaction queries a finalized transaction by ID
	QueryTransaction(txID transaction.TxID) ([]byte, error)

	// QueryInstruction queries a finalized instruction by reference
	QueryInstruction(txID transaction.TxID, instrIndex uint32) ([]byte, error)

	// InvokeProgram invokes another program with the given data and compute budget
	// Returns the result data or error
	InvokeProgram(programID filestore.FileID, invokeData []byte, computeBudget int64, depth int) ([]byte, error)

	// GetDeclaredPrograms returns the list of programs declared in the instruction
	GetDeclaredPrograms() []filestore.FileID
}

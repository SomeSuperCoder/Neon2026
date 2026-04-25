package genesis

import (
	"fmt"
	"log"

	"github.com/poh-blockchain/internal/filestore"
)

// Well-known program IDs as defined in the design document.
var (
	// RuntimeProgramID is the runtime itself (0x00...00)
	RuntimeProgramID = filestore.FileID{}

	// SystemProgramID is the built-in system program (0x00...01)
	SystemProgramID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 1,
	}

	// TokenProgramID is the built-in token program (0x00...02)
	TokenProgramID = filestore.FileID{
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 2,
	}
)

// LoadBuiltinPrograms creates the System_Program and Token_Program files in the
// FileStore if they do not already exist. systemBytecode and tokenBytecode are
// the compiled .qsb contents, typically embedded by the caller via //go:embed.
//
// The function is idempotent — programs that are already present are skipped.
//
// Requirements: 3.6, 3.7
func LoadBuiltinPrograms(fs *filestore.FileStore, systemBytecode, tokenBytecode []byte) error {
	if err := loadProgram(fs, SystemProgramID, systemBytecode, "System_Program"); err != nil {
		return fmt.Errorf("failed to load System_Program: %w", err)
	}

	if err := loadProgram(fs, TokenProgramID, tokenBytecode, "Token_Program"); err != nil {
		return fmt.Errorf("failed to load Token_Program: %w", err)
	}

	return nil
}

// loadProgram creates a single built-in program file in the FileStore.
func loadProgram(fs *filestore.FileStore, id filestore.FileID, bytecode []byte, name string) error {
	// Idempotent: skip if already present.
	if _, err := fs.GetFile(id); err == nil {
		log.Printf("genesis: %s already loaded at %s", name, id.String())
		return nil
	}

	// Calculate the minimum balance required to cover storage rent.
	storageCost := filestore.CalculateStorageCost(int64(len(bytecode)))
	balance := storageCost + 1000 // small buffer above the minimum

	file := &filestore.File{
		ID:         id,
		Balance:    balance,
		TxManager:  RuntimeProgramID,
		Data:       bytecode,
		Executable: true,
	}

	if _, err := fs.CreateFile(file); err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}

	log.Printf("genesis: loaded %s at %s (bytecode=%d bytes, balance=%d)",
		name, id.String(), len(bytecode), balance)
	return nil
}

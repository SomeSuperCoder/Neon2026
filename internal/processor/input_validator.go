package processor

import (
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/transaction"
)

// InputValidator validates instruction inputs before execution
type InputValidator struct {
	fileStore *filestore.FileStore
}

// NewInputValidator creates a new input validator
func NewInputValidator(fs *filestore.FileStore) *InputValidator {
	return &InputValidator{
		fileStore: fs,
	}
}

// ErrMissingInput indicates a required file is not declared in inputs
type ErrMissingInput struct {
	FileID   filestore.FileID
	Required string
}

func (e *ErrMissingInput) Error() string {
	return fmt.Sprintf("program %s not declared in instruction inputs; %s", e.FileID.String(), e.Required)
}

// ErrInvalidPermission indicates incorrect permission declaration
type ErrInvalidPermission struct {
	FileID   filestore.FileID
	Declared transaction.AccessPermission
	Required transaction.AccessPermission
}

func (e *ErrInvalidPermission) Error() string {
	return fmt.Sprintf("file %s declared with invalid permission: %d (expected %d)",
		e.FileID.String(), e.Declared, e.Required)
}

// ErrProgramNotExecutable indicates program file lacks executable flag
type ErrProgramNotExecutable struct {
	ProgramID filestore.FileID
}

func (e *ErrProgramNotExecutable) Error() string {
	return fmt.Sprintf("program %s is not marked as executable; ensure Executable flag is true",
		e.ProgramID.String())
}

// ErrFileNotFound indicates input file doesn't exist
type ErrFileNotFound struct {
	FileID filestore.FileID
}

func (e *ErrFileNotFound) Error() string {
	return fmt.Sprintf("input file %s not found in file store; ensure file exists before transaction",
		e.FileID.String())
}

// ValidateInstructionInputs validates all input declarations
// This implements Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 5.1, 5.2, 5.3, 5.4, 5.5
func (iv *InputValidator) ValidateInstructionInputs(instr *transaction.Instruction) error {
	if instr == nil {
		return fmt.Errorf("instruction cannot be nil")
	}

	// Validate inputs map is not empty (Requirement 1.1)
	if len(instr.Inputs) == 0 {
		return &ErrMissingInput{
			FileID:   instr.ProgramID,
			Required: "add to Inputs map with Read permission",
		}
	}

	// Validate program is declared in inputs (Requirement 2.1, 2.4)
	programDeclared := false
	for _, fileAccess := range instr.Inputs {
		if fileAccess.FileID == instr.ProgramID {
			programDeclared = true
			break
		}
	}

	if !programDeclared {
		return &ErrMissingInput{
			FileID:   instr.ProgramID,
			Required: "add to Inputs map with Read permission",
		}
	}

	// Validate all input files exist and have valid permissions (Requirements 1.2, 1.5, 5.1, 5.2)
	for key, fileAccess := range instr.Inputs {
		// Validate permission is valid (Read=1 or Write=2)
		if fileAccess.Permission != transaction.Read && fileAccess.Permission != transaction.Write {
			return &ErrInvalidPermission{
				FileID:   fileAccess.FileID,
				Declared: fileAccess.Permission,
				Required: transaction.Read, // or Write
			}
		}

		// Validate file exists (Requirement 1.2, 5.3)
		if err := iv.ValidateFileExists(fileAccess.FileID); err != nil {
			return fmt.Errorf("input file %s (key: %s) not found: %w",
				fileAccess.FileID.String(), key, err)
		}
	}

	// Validate program is executable (Requirements 2.3, 5.4)
	if err := iv.ValidateExecutableProgram(instr.ProgramID); err != nil {
		return err
	}

	return nil
}

// ValidateExecutableProgram validates program has Executable flag
// This implements Requirements 2.3, 5.4
func (iv *InputValidator) ValidateExecutableProgram(programID filestore.FileID) error {
	// Get program file
	programFile, err := iv.fileStore.GetFile(programID)
	if err != nil {
		return &ErrFileNotFound{FileID: programID}
	}

	// Check executable flag
	if !programFile.Executable {
		return &ErrProgramNotExecutable{ProgramID: programID}
	}

	return nil
}

// ValidateFileExists validates file exists in store
// This implements Requirements 1.2, 5.3
func (iv *InputValidator) ValidateFileExists(fileID filestore.FileID) error {
	_, err := iv.fileStore.GetFile(fileID)
	if err != nil {
		return &ErrFileNotFound{FileID: fileID}
	}
	return nil
}

package system

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/access"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/runtime"
	"github.com/poh-blockchain/internal/transaction"
)

// TestSystemProgramID verifies the System Program ID constant
func TestSystemProgramID(t *testing.T) {
	expected := "0000000000000000000000000000000000000000000000000000000000000001"
	if SystemProgramID.String() != expected {
		t.Errorf("SystemProgramID = %s, want %s", SystemProgramID.String(), expected)
	}
}

// TestGetProgramID verifies the SystemProgram implements BuiltinProgram interface
func TestGetProgramID(t *testing.T) {
	sp := NewSystemProgram()
	if sp.GetProgramID() != SystemProgramID {
		t.Errorf("GetProgramID() = %s, want %s", sp.GetProgramID().String(), SystemProgramID.String())
	}
}

// TestCreateAccount verifies account creation functionality
func TestCreateAccount(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	sp := NewSystemProgram()

	// Create owner public key
	var ownerPubKey transaction.PublicKey
	copy(ownerPubKey[:], []byte("test_owner_public_key_123456"))

	// Generate new account ID
	newAccountID := filestore.GenerateFileID([]byte("new_account_123"))

	// Create instruction
	instructionData := EncodeCreateAccountInstruction(ownerPubKey, 10000)
	instruction := &transaction.Instruction{
		ProgramID: SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"new_account": {FileID: newAccountID, Permission: transaction.Write},
		},
		Data: instructionData,
	}

	// Create execution context
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)
	ctx := runtime.NewExecutionContext(instruction, []transaction.PublicKey{ownerPubKey}, fs, ac)

	// Execute CreateAccount
	err = sp.Execute(ctx)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// Verify account was created
	account, err := fs.GetFile(newAccountID)
	if err != nil {
		t.Fatalf("Failed to retrieve created account: %v", err)
	}

	if account.Balance != 10000 {
		t.Errorf("Account balance = %d, want 10000", account.Balance)
	}
	if account.Executable {
		t.Error("Account should not be executable")
	}
	if account.TxManager != SystemProgramID {
		t.Errorf("Account TxManager = %s, want %s", account.TxManager.String(), SystemProgramID.String())
	}
	if len(account.Data) != 0 {
		t.Errorf("Account data length = %d, want 0", len(account.Data))
	}
}

// TestTransfer verifies balance transfer functionality
func TestTransfer(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	sp := NewSystemProgram()

	// Create two accounts
	fromID := filestore.GenerateFileID([]byte("from_account"))
	toID := filestore.GenerateFileID([]byte("to_account"))

	fromAccount := &filestore.File{
		ID:         fromID,
		Balance:    5000,
		TxManager:  SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	toAccount := &filestore.File{
		ID:         toID,
		Balance:    2000,
		TxManager:  SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(fromAccount)
	if err != nil {
		t.Fatalf("Failed to create from account: %v", err)
	}
	_, err = fs.CreateFile(toAccount)
	if err != nil {
		t.Fatalf("Failed to create to account: %v", err)
	}

	// Create transfer instruction
	instructionData := EncodeTransferInstruction(1500)
	instruction := &transaction.Instruction{
		ProgramID: SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"from": {FileID: fromID, Permission: transaction.Write},
			"to":   {FileID: toID, Permission: transaction.Write},
		},
		Data: instructionData,
	}

	// Create execution context
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)
	ctx := runtime.NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Execute Transfer
	err = sp.Execute(ctx)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Verify balances
	fromAccountAfter, _ := fs.GetFile(fromID)
	toAccountAfter, _ := fs.GetFile(toID)

	if fromAccountAfter.Balance != 3500 {
		t.Errorf("From account balance = %d, want 3500", fromAccountAfter.Balance)
	}
	if toAccountAfter.Balance != 3500 {
		t.Errorf("To account balance = %d, want 3500", toAccountAfter.Balance)
	}
}

// TestCloseAccount verifies account closure functionality
func TestCloseAccount(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	sp := NewSystemProgram()

	// Create two accounts
	accountID := filestore.GenerateFileID([]byte("account_to_close"))
	destinationID := filestore.GenerateFileID([]byte("destination_account"))

	account := &filestore.File{
		ID:         accountID,
		Balance:    3000,
		TxManager:  SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}
	destination := &filestore.File{
		ID:         destinationID,
		Balance:    1000,
		TxManager:  SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(account)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}
	_, err = fs.CreateFile(destination)
	if err != nil {
		t.Fatalf("Failed to create destination: %v", err)
	}

	// Create close account instruction
	instructionData := EncodeCloseAccountInstruction()
	instruction := &transaction.Instruction{
		ProgramID: SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"account":     {FileID: accountID, Permission: transaction.Write},
			"destination": {FileID: destinationID, Permission: transaction.Write},
		},
		Data: instructionData,
	}

	// Create execution context
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)
	ctx := runtime.NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Execute CloseAccount
	err = sp.Execute(ctx)
	if err != nil {
		t.Fatalf("CloseAccount failed: %v", err)
	}

	// Verify account was deleted
	_, err = fs.GetFile(accountID)
	if err == nil {
		t.Error("Account should have been deleted")
	}

	// Verify destination received the balance
	destinationAfter, _ := fs.GetFile(destinationID)
	if destinationAfter.Balance != 4000 {
		t.Errorf("Destination balance = %d, want 4000", destinationAfter.Balance)
	}
}

// TestAllocateData verifies data allocation functionality
func TestAllocateData(t *testing.T) {
	// Setup
	dbPath := t.TempDir() + "/test.db"
	defer os.RemoveAll(dbPath)

	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file store: %v", err)
	}
	defer fs.Close()

	sp := NewSystemProgram()

	// Create account with sufficient balance for storage
	accountID := filestore.GenerateFileID([]byte("account_with_data"))
	account := &filestore.File{
		ID:         accountID,
		Balance:    100000, // Sufficient for storage cost
		TxManager:  SystemProgramID,
		Data:       []byte{},
		Executable: false,
	}

	_, err = fs.CreateFile(account)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create allocate data instruction (allocate 1024 bytes)
	instructionData := EncodeAllocateDataInstruction(1024)
	instruction := &transaction.Instruction{
		ProgramID: SystemProgramID,
		Inputs: map[string]transaction.FileAccess{
			"account": {FileID: accountID, Permission: transaction.Write},
		},
		Data: instructionData,
	}

	// Create execution context
	ac := access.NewAccessController()
	ac.SetDeclaredAccess(instruction.Inputs)
	ctx := runtime.NewExecutionContext(instruction, []transaction.PublicKey{}, fs, ac)

	// Execute AllocateData
	err = sp.Execute(ctx)
	if err != nil {
		t.Fatalf("AllocateData failed: %v", err)
	}

	// Verify data was allocated
	accountAfter, _ := fs.GetFile(accountID)
	if len(accountAfter.Data) != 1024 {
		t.Errorf("Account data length = %d, want 1024", len(accountAfter.Data))
	}
}

package internal

import (
	"os"
	"testing"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/poh"
	"github.com/poh-blockchain/internal/verification"
)

// TestStateRootIntegration tests the integration of file-based state with blockchain
func TestStateRootIntegration(t *testing.T) {
	// Create temporary database for file store
	dbPath := "test_state_integration.db"
	defer os.Remove(dbPath)

	// Initialize FileStore
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fs.Close()

	// Create some test files
	file1 := &filestore.File{
		Balance:    10000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data 1"),
		Executable: false,
	}

	file1ID, err := fs.CreateFile(file1)
	if err != nil {
		t.Fatalf("Failed to create file 1: %v", err)
	}

	file2 := &filestore.File{
		Balance:    20000,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data 2"),
		Executable: true,
	}

	file2ID, err := fs.CreateFile(file2)
	if err != nil {
		t.Fatalf("Failed to create file 2: %v", err)
	}

	// Calculate state root
	stateRoot, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("Failed to calculate state root: %v", err)
	}

	if len(stateRoot) != 32 {
		t.Errorf("Expected state root length 32, got %d", len(stateRoot))
	}

	// Initialize PoH Clock and Block Producer
	pohClock := poh.NewPohClock([]byte("test-state-integration"))
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Produce a block with the state root
	block, err := blockProducer.ProduceBlock(1, stateRoot)
	if err != nil {
		t.Fatalf("Failed to produce block: %v", err)
	}

	// Verify the block contains the state root
	if len(block.Header.StateRoot) == 0 {
		t.Error("Block header state root is empty")
	}

	if string(block.Header.StateRoot) != string(stateRoot) {
		t.Error("Block header state root does not match calculated state root")
	}

	// Verify state root using verifier
	verifier := verification.NewVerifier()
	if err := verifier.VerifyStateRoot(block, fs); err != nil {
		t.Fatalf("State root verification failed: %v", err)
	}

	// Modify file state
	file1.Balance = 15000
	if err := fs.UpdateFile(file1ID, file1); err != nil {
		t.Fatalf("Failed to update file 1: %v", err)
	}

	// Calculate new state root
	newStateRoot, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("Failed to calculate new state root: %v", err)
	}

	// Verify state root changed
	if string(newStateRoot) == string(stateRoot) {
		t.Error("State root should have changed after file modification")
	}

	// Verify old block state root no longer matches
	if err := verifier.VerifyStateRoot(block, fs); err == nil {
		t.Error("Expected state root verification to fail after state change")
	}

	t.Logf("Created files: %s, %s", file1ID.String(), file2ID.String())
	t.Logf("Initial state root: %x", stateRoot)
	t.Logf("New state root: %x", newStateRoot)
	t.Log("State root integration test passed")
}

// TestFileTransactionIntegration tests file-based transaction integration with entries
func TestFileTransactionIntegration(t *testing.T) {
	// Initialize PoH Clock and Block Producer
	pohClock := poh.NewPohClock([]byte("test-file-tx-integration"))
	blockProducer := blockchain.NewBlockProducer(pohClock)

	// Create a mock file transaction (serialized)
	fileTx := []byte(`{"last_seen":"0000000000000000000000000000000000000000000000000000000000000000","instructions":[],"signatures":[]}`)

	// Add file transaction to block producer
	blockProducer.AddFileTransaction(fileTx)

	// Produce an entry
	entry, err := blockProducer.ProduceEntry()
	if err != nil {
		t.Fatalf("Failed to produce entry: %v", err)
	}

	// Verify entry contains file transaction
	if len(entry.FileTransactions) != 1 {
		t.Errorf("Expected 1 file transaction in entry, got %d", len(entry.FileTransactions))
	}

	if string(entry.FileTransactions[0]) != string(fileTx) {
		t.Error("File transaction in entry does not match original")
	}

	// Produce a block with empty state root
	block, err := blockProducer.ProduceBlock(1, []byte{})
	if err != nil {
		t.Fatalf("Failed to produce block: %v", err)
	}

	// Verify block contains the entry with file transaction
	foundFileTx := false
	for _, e := range block.Entries {
		if len(e.FileTransactions) > 0 {
			foundFileTx = true
			break
		}
	}

	if !foundFileTx {
		t.Error("Block does not contain entry with file transaction")
	}

	t.Log("File transaction integration test passed")
}

// TestEmptyStateRoot tests state root calculation with no files
func TestEmptyStateRoot(t *testing.T) {
	// Create temporary database for file store
	dbPath := "test_empty_state.db"
	defer os.Remove(dbPath)

	// Initialize FileStore
	fs, err := filestore.NewFileStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fs.Close()

	// Calculate state root with no files
	stateRoot, err := fs.CalculateStateRoot()
	if err != nil {
		t.Fatalf("Failed to calculate empty state root: %v", err)
	}

	if len(stateRoot) != 32 {
		t.Errorf("Expected state root length 32, got %d", len(stateRoot))
	}

	// Produce a block with empty state root
	pohClock := poh.NewPohClock([]byte("test-empty-state"))
	blockProducer := blockchain.NewBlockProducer(pohClock)

	block, err := blockProducer.ProduceBlock(1, stateRoot)
	if err != nil {
		t.Fatalf("Failed to produce block: %v", err)
	}

	// Verify state root
	verifier := verification.NewVerifier()
	if err := verifier.VerifyStateRoot(block, fs); err != nil {
		t.Fatalf("Empty state root verification failed: %v", err)
	}

	t.Log("Empty state root test passed")
}

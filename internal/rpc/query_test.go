package rpc

import (
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/poh-blockchain/internal/blockchain"
	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/storage"
	"github.com/poh-blockchain/internal/transaction"
)

// TestQueryEngineCreation tests creating a new query engine
func TestQueryEngineCreation(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_query_ledger.db"
	storePath := "test_query_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)
	if qe == nil {
		t.Fatal("Expected non-nil query engine")
	}

	if qe.ledger != ledger {
		t.Error("Query engine ledger not set correctly")
	}

	if qe.fileStore != store {
		t.Error("Query engine filestore not set correctly")
	}

	if qe.cache == nil {
		t.Error("Query engine cache not initialized")
	}
}

// TestGetBalance tests retrieving account balance
func TestGetBalance(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_balance_ledger.db"
	storePath := "test_balance_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create a test file with balance
	testBalance := int64(1000000)
	testFile := &filestore.File{
		Balance:    testBalance,
		TxManager:  filestore.FileID{},
		Data:       []byte("test data"),
		Executable: false,
	}

	fileID, err := store.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting balance
	balance, err := qe.GetBalance(fileID.String())
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	if balance != testBalance {
		t.Errorf("Expected balance %d, got %d", testBalance, balance)
	}
}

// TestGetBalanceNonExistent tests getting balance for non-existent account
func TestGetBalanceNonExistent(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_balance_nonexist_ledger.db"
	storePath := "test_balance_nonexist_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting balance for non-existent account
	nonExistentID := "0000000000000000000000000000000000000000000000000000000000000000"
	balance, err := qe.GetBalance(nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent account")
	}

	if balance != 0 {
		t.Errorf("Expected balance 0 for non-existent account, got %d", balance)
	}
}

// TestGetBalanceInvalidAddress tests getting balance with invalid address
func TestGetBalanceInvalidAddress(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_balance_invalid_ledger.db"
	storePath := "test_balance_invalid_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with invalid hex string
	_, err = qe.GetBalance("invalid_hex")
	if err == nil {
		t.Error("Expected error for invalid address")
	}

	// Test with wrong length
	_, err = qe.GetBalance("1234")
	if err == nil {
		t.Error("Expected error for wrong length address")
	}
}

// TestGetBlockHeight tests retrieving blockchain height
func TestGetBlockHeight(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_height_ledger.db"
	storePath := "test_height_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with empty ledger
	height, err := qe.GetBlockHeight()
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}

	if height != 0 {
		t.Errorf("Expected height 0 for empty ledger, got %d", height)
	}

	// Add a block
	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			BlockHeight:       1,
			Slot:              1,
			Timestamp:         time.Now(),
			PreviousBlockHash: make([]byte, 32),
			MerkleRoot:        make([]byte, 32),
		},
		Entries: []blockchain.Entry{},
	}

	err = ledger.StoreBlock(block)
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}

	// Clear cache to force fresh query
	qe.cache.mu.Lock()
	qe.cache.blockHeight = nil
	qe.cache.mu.Unlock()

	// Test with one block
	height, err = qe.GetBlockHeight()
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}

	if height != 1 {
		t.Errorf("Expected height 1, got %d", height)
	}
}

// TestGetRecentBlockhash tests retrieving recent blockhash
func TestGetRecentBlockhash(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_blockhash_ledger.db"
	storePath := "test_blockhash_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with empty ledger
	_, err = qe.GetRecentBlockhash()
	if err == nil {
		t.Error("Expected error for empty ledger")
	}

	// Add a block with known hash
	expectedHash := make([]byte, 32)
	for i := range expectedHash {
		expectedHash[i] = byte(i)
	}

	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			BlockHeight:       1,
			Slot:              1,
			Timestamp:         time.Now(),
			PreviousBlockHash: make([]byte, 32),
			MerkleRoot:        expectedHash,
		},
		Entries: []blockchain.Entry{},
	}

	err = ledger.StoreBlock(block)
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}

	// Test getting recent blockhash
	blockhash, err := qe.GetRecentBlockhash()
	if err != nil {
		t.Fatalf("Failed to get recent blockhash: %v", err)
	}

	expectedHashStr := hex.EncodeToString(expectedHash)
	if blockhash != expectedHashStr {
		t.Errorf("Expected blockhash %s, got %s", expectedHashStr, blockhash)
	}
}

// TestQueryCaching tests that query results are cached
func TestQueryCaching(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_cache_ledger.db"
	storePath := "test_cache_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Add a block
	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			BlockHeight:       1,
			Slot:              1,
			Timestamp:         time.Now(),
			PreviousBlockHash: make([]byte, 32),
			MerkleRoot:        make([]byte, 32),
		},
		Entries: []blockchain.Entry{},
	}

	err = ledger.StoreBlock(block)
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// First call - should cache
	height1, err := qe.GetBlockHeight()
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}

	// Check cache was populated
	if qe.cache.blockHeight == nil {
		t.Error("Expected block height to be cached")
	}

	// Second call - should use cache
	height2, err := qe.GetBlockHeight()
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}

	if height1 != height2 {
		t.Errorf("Cached height mismatch: %d vs %d", height1, height2)
	}
}

// TestGetAccountInfo tests retrieving full account information
func TestGetAccountInfo(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_account_info_ledger.db"
	storePath := "test_account_info_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create a test file
	txManager := filestore.GenerateFileID([]byte("tx_manager"))
	testFile := &filestore.File{
		Balance:    500000,
		TxManager:  txManager,
		Data:       []byte("test account data"),
		Executable: true,
	}

	fileID, err := store.CreateFile(testFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting account info
	info, err := qe.GetAccountInfo(fileID.String())
	if err != nil {
		t.Fatalf("Failed to get account info: %v", err)
	}

	if info.Address != fileID.String() {
		t.Errorf("Expected address %s, got %s", fileID.String(), info.Address)
	}

	if info.Balance != 500000 {
		t.Errorf("Expected balance 500000, got %d", info.Balance)
	}

	if info.Owner != txManager.String() {
		t.Errorf("Expected owner %s, got %s", txManager.String(), info.Owner)
	}

	if info.DataLength != len(testFile.Data) {
		t.Errorf("Expected data length %d, got %d", len(testFile.Data), info.DataLength)
	}

	if !info.Executable {
		t.Error("Expected executable to be true")
	}
}

// TestGetAccountInfoNonExistent tests getting info for non-existent account
func TestGetAccountInfoNonExistent(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_account_info_nonexist_ledger.db"
	storePath := "test_account_info_nonexist_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting info for non-existent account - should return nil, nil
	nonExistentID := "0000000000000000000000000000000000000000000000000000000000000000"
	info, err := qe.GetAccountInfo(nonExistentID)
	if err != nil {
		t.Errorf("Expected no error for non-existent account, got: %v", err)
	}

	if info != nil {
		t.Error("Expected nil info for non-existent account")
	}
}

// TestGetTransactionHistory tests retrieving transaction history
func TestGetTransactionHistory(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_history_ledger.db"
	storePath := "test_tx_history_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create test accounts
	account1 := &filestore.File{
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       []byte("account1"),
		Executable: false,
	}
	account1ID, err := store.CreateFile(account1)
	if err != nil {
		t.Fatalf("Failed to create account1: %v", err)
	}

	account2 := &filestore.File{
		Balance:    500000,
		TxManager:  filestore.FileID{},
		Data:       []byte("account2"),
		Executable: false,
	}
	account2ID, err := store.CreateFile(account2)
	if err != nil {
		t.Fatalf("Failed to create account2: %v", err)
	}

	// Create a transaction
	var pk1, pk2 transaction.PublicKey
	copy(pk1[:], account1ID[:])
	copy(pk2[:], account2ID[:])

	tx := &transaction.Transaction{
		LastSeen: transaction.TxID{},
		Instructions: []transaction.Instruction{
			{
				ProgramID: filestore.GenerateFileID([]byte("system")),
				Inputs: map[string]transaction.FileAccess{
					"from": {FileID: account1ID, Permission: transaction.Write},
					"to":   {FileID: account2ID, Permission: transaction.Write},
				},
				Data: []byte{0}, // Transfer instruction
			},
		},
		Signatures: []transaction.Signature{
			{PublicKey: pk1, Signature: [64]byte{}},
		},
	}

	txBytes, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Create a block with the transaction
	entry := blockchain.Entry{
		Hash:             make([]byte, 32),
		NumHashes:        1,
		Transactions:     []blockchain.Transaction{},
		FileTransactions: [][]byte{txBytes},
		PreviousHash:     make([]byte, 32),
		Timestamp:        time.Now(),
	}

	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			BlockHeight:       1,
			Slot:              1,
			Timestamp:         time.Now(),
			PreviousBlockHash: make([]byte, 32),
			MerkleRoot:        make([]byte, 32),
		},
		Entries: []blockchain.Entry{entry},
	}

	err = ledger.StoreBlock(block)
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting transaction history for account1
	history, err := qe.GetTransactionHistory(account1ID.String(), 10)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(history))
	}

	if len(history) > 0 {
		record := history[0]
		if record.BlockHeight != 1 {
			t.Errorf("Expected block height 1, got %d", record.BlockHeight)
		}

		if !record.Success {
			t.Error("Expected transaction to be successful")
		}

		if len(record.Instructions) != 1 {
			t.Errorf("Expected 1 instruction, got %d", len(record.Instructions))
		}
	}
}

// TestGetTransactionHistoryPagination tests pagination support
func TestGetTransactionHistoryPagination(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_pagination_ledger.db"
	storePath := "test_tx_pagination_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create test account
	account := &filestore.File{
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       []byte("account"),
		Executable: false,
	}
	accountID, err := store.CreateFile(account)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create multiple blocks with transactions
	for i := 1; i <= 5; i++ {
		var pk transaction.PublicKey
		copy(pk[:], accountID[:])

		tx := &transaction.Transaction{
			LastSeen: transaction.TxID{},
			Instructions: []transaction.Instruction{
				{
					ProgramID: filestore.GenerateFileID([]byte("system")),
					Inputs: map[string]transaction.FileAccess{
						"account": {FileID: accountID, Permission: transaction.Write},
					},
					Data: []byte{0},
				},
			},
			Signatures: []transaction.Signature{
				{PublicKey: pk, Signature: [64]byte{}},
			},
		}

		txBytes, err := tx.Marshal()
		if err != nil {
			t.Fatalf("Failed to marshal transaction: %v", err)
		}

		entry := blockchain.Entry{
			Hash:             make([]byte, 32),
			NumHashes:        1,
			Transactions:     []blockchain.Transaction{},
			FileTransactions: [][]byte{txBytes},
			PreviousHash:     make([]byte, 32),
			Timestamp:        time.Now(),
		}

		block := blockchain.Block{
			Header: blockchain.BlockHeader{
				BlockHeight:       int64(i),
				Slot:              int64(i),
				Timestamp:         time.Now(),
				PreviousBlockHash: make([]byte, 32),
				MerkleRoot:        make([]byte, 32),
			},
			Entries: []blockchain.Entry{entry},
		}

		err = ledger.StoreBlock(block)
		if err != nil {
			t.Fatalf("Failed to store block %d: %v", i, err)
		}
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with limit of 3
	history, err := qe.GetTransactionHistory(accountID.String(), 3)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 transactions with limit, got %d", len(history))
	}

	// Verify reverse chronological order (most recent first)
	if len(history) >= 2 {
		if history[0].BlockHeight < history[1].BlockHeight {
			t.Error("Expected transactions in reverse chronological order")
		}
	}
}

// TestGetTransactionHistoryDefaultLimit tests default limit of 20
func TestGetTransactionHistoryDefaultLimit(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_default_limit_ledger.db"
	storePath := "test_tx_default_limit_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create test account
	account := &filestore.File{
		Balance:    1000000,
		TxManager:  filestore.FileID{},
		Data:       []byte("account"),
		Executable: false,
	}
	accountID, err := store.CreateFile(account)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with limit 0 (should use default of 20)
	history, err := qe.GetTransactionHistory(accountID.String(), 0)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	// Should return empty array for account with no transactions
	if len(history) != 0 {
		t.Errorf("Expected 0 transactions for new account, got %d", len(history))
	}
}

// TestGetTransactionHistoryInvalidAddress tests error handling for invalid address
func TestGetTransactionHistoryInvalidAddress(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_invalid_ledger.db"
	storePath := "test_tx_invalid_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test with invalid address
	_, err = qe.GetTransactionHistory("invalid", 10)
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

// TestGetTransactionStatus tests retrieving transaction status
func TestGetTransactionStatus(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_status_ledger.db"
	storePath := "test_tx_status_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create a transaction with known signature
	var sig [64]byte
	for i := range sig {
		sig[i] = byte(i)
	}

	tx := &transaction.Transaction{
		LastSeen:     transaction.TxID{},
		Instructions: []transaction.Instruction{},
		Signatures: []transaction.Signature{
			{PublicKey: transaction.PublicKey{}, Signature: sig},
		},
	}

	txBytes, err := tx.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	// Create a block with the transaction
	entry := blockchain.Entry{
		Hash:             make([]byte, 32),
		NumHashes:        1,
		Transactions:     []blockchain.Transaction{},
		FileTransactions: [][]byte{txBytes},
		PreviousHash:     make([]byte, 32),
		Timestamp:        time.Now(),
	}

	block := blockchain.Block{
		Header: blockchain.BlockHeader{
			BlockHeight:       1,
			Slot:              1,
			Timestamp:         time.Now(),
			PreviousBlockHash: make([]byte, 32),
			MerkleRoot:        make([]byte, 32),
		},
		Entries: []blockchain.Entry{entry},
	}

	err = ledger.StoreBlock(block)
	if err != nil {
		t.Fatalf("Failed to store block: %v", err)
	}

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting status for confirmed transaction
	sigStr := hex.EncodeToString(sig[:])
	status, err := qe.GetTransactionStatus(sigStr)
	if err != nil {
		t.Fatalf("Failed to get transaction status: %v", err)
	}

	if !status.Confirmed {
		t.Error("Expected transaction to be confirmed")
	}

	if status.BlockHeight != 1 {
		t.Errorf("Expected block height 1, got %d", status.BlockHeight)
	}

	if status.Slot != 1 {
		t.Errorf("Expected slot 1, got %d", status.Slot)
	}
}

// TestGetTransactionStatusNotFound tests status for non-existent transaction
func TestGetTransactionStatusNotFound(t *testing.T) {
	// Create temporary databases
	ledgerPath := "test_tx_status_notfound_ledger.db"
	storePath := "test_tx_status_notfound_store.db"
	defer os.RemoveAll(ledgerPath)
	defer os.RemoveAll(storePath)

	ledger, err := storage.NewLedger(ledgerPath)
	if err != nil {
		t.Fatalf("Failed to create ledger: %v", err)
	}
	defer ledger.Close()

	store, err := filestore.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("Failed to create filestore: %v", err)
	}
	defer store.Close()

	// Create query engine
	qe := NewQueryEngine(ledger, store)

	// Test getting status for non-existent transaction
	nonExistentSig := hex.EncodeToString(make([]byte, 64))
	status, err := qe.GetTransactionStatus(nonExistentSig)
	if err != nil {
		t.Fatalf("Failed to get transaction status: %v", err)
	}

	if status.Confirmed {
		t.Error("Expected transaction to not be confirmed")
	}
}

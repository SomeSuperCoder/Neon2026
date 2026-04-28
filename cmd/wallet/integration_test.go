package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
	"github.com/poh-blockchain/internal/rpc"
	"github.com/poh-blockchain/internal/transaction"
)

// TestWalletCreationFlow tests the complete wallet creation process
func TestWalletCreationFlow(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	walletPath := filepath.Join(tmpDir, "test-wallet.dat")

	// Test wallet creation with 12-word seed phrase
	t.Run("CreateWalletWith12Words", func(t *testing.T) {
		config := &core.WalletConfig{
			RPCEndpoint: "http://localhost:8899",
			WalletPath:  walletPath,
		}

		// Generate seed phrase
		seedPhrase, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		// Validate seed phrase format (12 words)
		words := len(strings.Fields(seedPhrase))
		if words != 12 {
			t.Errorf("Expected 12 words, got %d", words)
		}

		// Create wallet with seed phrase
		wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}

		// Verify wallet has one account
		if wallet.AccountCount() != 1 {
			t.Errorf("Expected 1 account, got %d", wallet.AccountCount())
		}

		// Verify active account is set
		activeAccount := wallet.GetActiveAccount()
		if activeAccount == nil {
			t.Fatal("No active account set")
		}

		// Verify account has valid address
		if len(activeAccount.Address) != 64 { // 32 bytes hex encoded
			t.Errorf("Invalid address length: %d", len(activeAccount.Address))
		}

		// Save wallet with password
		password := "test-password-123"
		if err := wallet.Save(password); err != nil {
			t.Fatalf("Failed to save wallet: %v", err)
		}

		// Verify wallet file exists
		if !fileExists(walletPath) {
			t.Error("Wallet file was not created")
		}

		// Verify file permissions
		info, err := os.Stat(walletPath)
		if err != nil {
			t.Fatalf("Failed to stat wallet file: %v", err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
		}
	})

	// Test wallet creation with 24-word seed phrase
	t.Run("CreateWalletWith24Words", func(t *testing.T) {
		walletPath24 := filepath.Join(tmpDir, "test-wallet-24.dat")
		config := &core.WalletConfig{
			RPCEndpoint: "http://localhost:8899",
			WalletPath:  walletPath24,
		}

		// Generate 24-word seed phrase
		seedPhrase, err := core.GenerateSeedPhrase(24)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		// Validate seed phrase format (24 words)
		words := len(strings.Fields(seedPhrase))
		if words != 24 {
			t.Errorf("Expected 24 words, got %d", words)
		}

		// Create wallet
		wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}

		// Save and verify
		password := "test-password-456"
		if err := wallet.Save(password); err != nil {
			t.Fatalf("Failed to save wallet: %v", err)
		}

		if !fileExists(walletPath24) {
			t.Error("24-word wallet file was not created")
		}
	})
}

// TestWalletRestorationFlow tests wallet restoration from seed phrase
func TestWalletRestorationFlow(t *testing.T) {
	tmpDir := t.TempDir()
	originalWalletPath := filepath.Join(tmpDir, "original-wallet.dat")
	restoredWalletPath := filepath.Join(tmpDir, "restored-wallet.dat")

	// Create original wallet
	seedPhrase, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	originalConfig := &core.WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  originalWalletPath,
	}

	originalWallet, err := core.NewWalletWithSeedPhrase(seedPhrase, originalConfig)
	if err != nil {
		t.Fatalf("Failed to create original wallet: %v", err)
	}

	// Import additional seed phrases to test multi-account restoration
	for i := 0; i < 3; i++ {
		additionalSeed, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate additional seed phrase: %v", err)
		}
		_, err = originalWallet.ImportSeedPhrase(additionalSeed, fmt.Sprintf("Account %d", i+2))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Set active account to test restoration
	if err := originalWallet.SetActiveAccount(2); err != nil {
		t.Fatalf("Failed to set active account: %v", err)
	}

	// Save original wallet
	password := "restoration-test-password"
	if err := originalWallet.Save(password); err != nil {
		t.Fatalf("Failed to save original wallet: %v", err)
	}

	// Test restoration
	t.Run("RestoreFromSeedPhrase", func(t *testing.T) {
		restoredConfig := &core.WalletConfig{
			RPCEndpoint: "http://localhost:8899",
			WalletPath:  restoredWalletPath,
		}

		// Restore wallet using the first seed phrase
		restoredWallet, err := core.NewWalletWithSeedPhrase(seedPhrase, restoredConfig)
		if err != nil {
			t.Fatalf("Failed to restore wallet: %v", err)
		}

		// Verify restored account matches original
		originalAccount := originalWallet.GetAccount(0)
		restoredAccount := restoredWallet.GetActiveAccount()

		if originalAccount.Address != restoredAccount.Address {
			t.Errorf("Restored account address mismatch: original=%s, restored=%s",
				originalAccount.Address, restoredAccount.Address)
		}

		// Verify private keys match by signing the same data
		testData := []byte("test signature data")
		originalSig := ed25519.Sign(originalAccount.PrivateKey, testData)
		restoredSig := ed25519.Sign(restoredAccount.PrivateKey, testData)

		if !bytes.Equal(originalSig, restoredSig) {
			t.Error("Restored private key does not match original")
		}
	})

	// Test restoration with invalid seed phrase
	t.Run("RestoreWithInvalidSeedPhrase", func(t *testing.T) {
		invalidConfig := &core.WalletConfig{
			RPCEndpoint: "http://localhost:8899",
			WalletPath:  filepath.Join(tmpDir, "invalid-wallet.dat"),
		}

		invalidSeed := "invalid seed phrase with wrong words count format"
		_, err := core.NewWalletWithSeedPhrase(invalidSeed, invalidConfig)
		if err == nil {
			t.Error("Expected error for invalid seed phrase")
		}

		walletErr, ok := err.(*core.WalletError)
		if !ok {
			t.Error("Expected WalletError")
		} else if walletErr.Code != core.ErrInvalidSeedPhrase {
			t.Errorf("Expected error code %s, got %s", core.ErrInvalidSeedPhrase, walletErr.Code)
		}
	})

	// Test loading existing wallet
	t.Run("LoadExistingWallet", func(t *testing.T) {
		// Load the original wallet
		loadedWallet, err := core.LoadWallet(originalWalletPath, password)
		if err != nil {
			t.Fatalf("Failed to load existing wallet: %v", err)
		}

		// Verify all accounts are loaded
		if loadedWallet.AccountCount() != originalWallet.AccountCount() {
			t.Errorf("Account count mismatch: expected %d, got %d",
				originalWallet.AccountCount(), loadedWallet.AccountCount())
		}

		// Verify active account is preserved
		if loadedWallet.GetActiveAccountIndex() != originalWallet.GetActiveAccountIndex() {
			t.Errorf("Active account index mismatch: expected %d, got %d",
				originalWallet.GetActiveAccountIndex(), loadedWallet.GetActiveAccountIndex())
		}

		// Verify all account addresses match
		for i := 0; i < originalWallet.AccountCount(); i++ {
			originalAcc := originalWallet.GetAccount(i)
			loadedAcc := loadedWallet.GetAccount(i)
			if originalAcc.Address != loadedAcc.Address {
				t.Errorf("Account %d address mismatch", i)
			}
		}
	})
}

// TestTransactionBuildingAndSigning tests transaction creation and signing
func TestTransactionBuildingAndSigning(t *testing.T) {
	// Create test wallet
	seedPhrase, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &core.WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  filepath.Join(t.TempDir(), "tx-test-wallet.dat"),
	}

	wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	activeAccount := wallet.GetActiveAccount()
	if activeAccount == nil {
		t.Fatal("No active account")
	}

	// Create recipient address
	recipientSeed, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate recipient seed: %v", err)
	}
	recipientWallet, err := core.NewWalletWithSeedPhrase(recipientSeed, config)
	if err != nil {
		t.Fatalf("Failed to create recipient wallet: %v", err)
	}
	recipientAccount := recipientWallet.GetActiveAccount()

	t.Run("BuildValidTransferTransaction", func(t *testing.T) {
		transferReq := &core.TransferRequest{
			From:   activeAccount.Address,
			To:     recipientAccount.Address,
			Amount: 1000,
			Memo:   "Integration test transfer",
		}

		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			t.Fatalf("Failed to build transfer transaction: %v", err)
		}

		// Verify transaction structure
		if len(tx.Instructions) != 1 {
			t.Errorf("Expected 1 instruction, got %d", len(tx.Instructions))
		}

		// Verify transaction is signed
		if len(tx.Signatures) != 1 {
			t.Errorf("Expected 1 signature, got %d", len(tx.Signatures))
		}

		// Verify signature is from active account
		if tx.Signatures[0].PublicKey != activeAccount.PublicKey {
			t.Error("Transaction not signed by active account")
		}

		// Verify signature is valid
		// Remove signatures temporarily for verification
		savedSigs := tx.Signatures
		tx.Signatures = []transaction.Signature{}
		txData, err := tx.Marshal()
		if err != nil {
			t.Fatalf("Failed to marshal transaction for verification: %v", err)
		}
		tx.Signatures = savedSigs

		if !savedSigs[0].Verify(txData) {
			t.Error("Transaction signature verification failed")
		}
	})

	t.Run("BuildTransactionWithInvalidAmount", func(t *testing.T) {
		transferReq := &core.TransferRequest{
			From:   activeAccount.Address,
			To:     recipientAccount.Address,
			Amount: 0, // Invalid amount
			Memo:   "Invalid transfer",
		}

		_, err := wallet.BuildTransferTransaction(transferReq)
		if err == nil {
			t.Error("Expected error for zero amount")
		}

		walletErr, ok := err.(*core.WalletError)
		if !ok {
			t.Error("Expected WalletError")
		} else if walletErr.Code != core.ErrInvalidAmount {
			t.Errorf("Expected error code %s, got %s", core.ErrInvalidAmount, walletErr.Code)
		}
	})

	t.Run("SerializeAndDeserializeTransaction", func(t *testing.T) {
		transferReq := &core.TransferRequest{
			From:   activeAccount.Address,
			To:     recipientAccount.Address,
			Amount: 2500,
			Memo:   "Serialization test",
		}

		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			t.Fatalf("Failed to build transaction: %v", err)
		}

		// Serialize transaction
		txData, err := core.SerializeTransaction(tx)
		if err != nil {
			t.Fatalf("Failed to serialize transaction: %v", err)
		}

		// Verify serialized data is not empty
		if len(txData) == 0 {
			t.Error("Serialized transaction data is empty")
		}

		// Deserialize and verify
		deserializedTx, err := transaction.UnmarshalTransaction(txData)
		if err != nil {
			t.Fatalf("Failed to deserialize transaction: %v", err)
		}

		// Verify structure matches
		if len(deserializedTx.Instructions) != len(tx.Instructions) {
			t.Errorf("Instruction count mismatch after deserialization")
		}

		if len(deserializedTx.Signatures) != len(tx.Signatures) {
			t.Errorf("Signature count mismatch after deserialization")
		}
	})
}

// MockRPCServer creates a mock RPC server for testing
type MockRPCServer struct {
	server   *httptest.Server
	accounts map[string]*MockAccount
	txs      map[string]*MockTransaction
}

type MockAccount struct {
	Address    string
	Balance    int64
	Owner      string
	DataLength int
	Executable bool
}

type MockTransaction struct {
	Signature   string
	Confirmed   bool
	BlockHeight int64
	Slot        int64
}

func NewMockRPCServer() *MockRPCServer {
	mock := &MockRPCServer{
		accounts: make(map[string]*MockAccount),
		txs:      make(map[string]*MockTransaction),
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", mock.handleRPCRequest)
	mock.server = httptest.NewServer(mux)

	return mock
}

func (m *MockRPCServer) Close() {
	m.server.Close()
}

func (m *MockRPCServer) URL() string {
	return m.server.URL
}

func (m *MockRPCServer) AddAccount(address string, balance int64) {
	m.accounts[address] = &MockAccount{
		Address:    address,
		Balance:    balance,
		Owner:      "system",
		DataLength: 0,
		Executable: false,
	}
}

func (m *MockRPCServer) AddTransaction(signature string, confirmed bool, blockHeight int64) {
	m.txs[signature] = &MockTransaction{
		Signature:   signature,
		Confirmed:   confirmed,
		BlockHeight: blockHeight,
		Slot:        blockHeight * 10,
	}
}

func (m *MockRPCServer) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	var req rpc.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var response rpc.JSONRPCResponse
	response.JSONRPC = "2.0"
	response.ID = req.ID

	switch req.Method {
	case "getBalance":
		result := m.handleGetBalance(req.Params)
		if err, ok := result.(*rpc.RPCError); ok {
			response.Error = err
		} else {
			response.Result = result
		}
	case "getAccountInfo":
		result := m.handleGetAccountInfo(req.Params)
		if err, ok := result.(*rpc.RPCError); ok {
			response.Error = err
		} else {
			response.Result = result
		}
	case "getTransactionHistory":
		response.Result = m.handleGetTransactionHistory(req.Params)
	case "sendTransaction":
		response.Result = m.handleSendTransaction(req.Params)
	case "getTransactionStatus":
		response.Result = m.handleGetTransactionStatus(req.Params)
	case "getBlockHeight":
		response.Result = int64(12345)
	default:
		response.Error = &rpc.RPCError{
			Code:    rpc.MethodNotFound,
			Message: "Method not found",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockRPCServer) handleGetBalance(params json.RawMessage) interface{} {
	var p struct {
		Address string `json:"address"`
	}
	json.Unmarshal(params, &p)

	// Validate address format (should be 64 hex characters)
	if len(p.Address) != 64 {
		return &rpc.RPCError{
			Code:    rpc.InvalidParams,
			Message: "Invalid address format",
		}
	}

	if account, exists := m.accounts[p.Address]; exists {
		return account.Balance
	}
	return int64(0)
}

func (m *MockRPCServer) handleGetAccountInfo(params json.RawMessage) interface{} {
	var p struct {
		Address string `json:"address"`
	}
	json.Unmarshal(params, &p)

	if account, exists := m.accounts[p.Address]; exists {
		return rpc.AccountInfo{
			Address:    account.Address,
			Balance:    account.Balance,
			Owner:      account.Owner,
			DataLength: account.DataLength,
			Executable: account.Executable,
		}
	}
	return nil
}

func (m *MockRPCServer) handleGetTransactionHistory(params json.RawMessage) interface{} {
	var p struct {
		Address string `json:"address"`
		Limit   int    `json:"limit"`
	}
	json.Unmarshal(params, &p)

	// Return empty history for simplicity
	return []rpc.TransactionRecord{}
}

func (m *MockRPCServer) handleSendTransaction(params json.RawMessage) interface{} {
	// Generate mock signature
	signature := fmt.Sprintf("mock-signature-%d", time.Now().UnixNano())

	// Add to mock transactions as confirmed
	m.AddTransaction(signature, true, 12346)

	return signature
}

func (m *MockRPCServer) handleGetTransactionStatus(params json.RawMessage) interface{} {
	var p struct {
		Signature string `json:"signature"`
	}
	json.Unmarshal(params, &p)

	if tx, exists := m.txs[p.Signature]; exists {
		return rpc.TransactionStatus{
			Signature:   tx.Signature,
			Confirmed:   tx.Confirmed,
			BlockHeight: tx.BlockHeight,
			Slot:        tx.Slot,
		}
	}

	return rpc.TransactionStatus{
		Signature: p.Signature,
		Confirmed: false,
	}
}

// TestRPCClientWithMockServer tests RPC client functionality with mock server
func TestRPCClientWithMockServer(t *testing.T) {
	// Create mock server
	mockServer := NewMockRPCServer()
	defer mockServer.Close()

	// Create RPC client
	client := walletrpc.NewRPCClient(mockServer.URL())

	// Create test wallet for addresses
	seedPhrase, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &core.WalletConfig{
		RPCEndpoint: mockServer.URL(),
		WalletPath:  filepath.Join(t.TempDir(), "rpc-test-wallet.dat"),
	}

	wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	testAccount := wallet.GetActiveAccount()
	testAddress := testAccount.Address

	t.Run("GetBalance", func(t *testing.T) {
		// Add account to mock server
		expectedBalance := int64(1000000)
		mockServer.AddAccount(testAddress, expectedBalance)

		// Test GetBalance
		balance, err := client.GetBalance(testAddress)
		if err != nil {
			t.Fatalf("GetBalance failed: %v", err)
		}

		if balance != expectedBalance {
			t.Errorf("Expected balance %d, got %d", expectedBalance, balance)
		}
	})

	t.Run("GetAccountInfo", func(t *testing.T) {
		// Test GetAccountInfo
		accountInfo, err := client.GetAccountInfo(testAddress)
		if err != nil {
			t.Fatalf("GetAccountInfo failed: %v", err)
		}

		if accountInfo == nil {
			t.Fatal("Expected account info, got nil")
		}

		if accountInfo.Address != testAddress {
			t.Errorf("Address mismatch: expected %s, got %s", testAddress, accountInfo.Address)
		}

		if accountInfo.Balance != 1000000 {
			t.Errorf("Balance mismatch: expected 1000000, got %d", accountInfo.Balance)
		}
	})

	t.Run("GetTransactionHistory", func(t *testing.T) {
		// Test GetTransactionHistory
		history, err := client.GetTransactionHistory(testAddress, 10)
		if err != nil {
			t.Fatalf("GetTransactionHistory failed: %v", err)
		}

		// Should return empty array for new account
		if len(history) != 0 {
			t.Errorf("Expected empty history, got %d transactions", len(history))
		}
	})

	t.Run("SendTransaction", func(t *testing.T) {
		// Create a test transaction
		recipientSeed, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate recipient seed: %v", err)
		}
		recipientWallet, err := core.NewWalletWithSeedPhrase(recipientSeed, config)
		if err != nil {
			t.Fatalf("Failed to create recipient wallet: %v", err)
		}

		transferReq := &core.TransferRequest{
			From:   testAccount.Address,
			To:     recipientWallet.GetActiveAccount().Address,
			Amount: 500,
			Memo:   "RPC test transfer",
		}

		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			t.Fatalf("Failed to build transaction: %v", err)
		}

		// Serialize transaction
		txData, err := core.SerializeTransaction(tx)
		if err != nil {
			t.Fatalf("Failed to serialize transaction: %v", err)
		}

		// Send transaction
		signature, err := client.SendTransaction(txData)
		if err != nil {
			t.Fatalf("SendTransaction failed: %v", err)
		}

		if signature == "" {
			t.Error("Expected non-empty signature")
		}

		// Verify transaction was added to mock server
		status, err := client.GetTransactionStatus(signature)
		if err != nil {
			t.Fatalf("GetTransactionStatus failed: %v", err)
		}

		if !status.Confirmed {
			t.Error("Expected transaction to be confirmed")
		}

		if status.Signature != signature {
			t.Errorf("Signature mismatch: expected %s, got %s", signature, status.Signature)
		}
	})

	t.Run("GetBlockHeight", func(t *testing.T) {
		height, err := client.GetBlockHeight()
		if err != nil {
			t.Fatalf("GetBlockHeight failed: %v", err)
		}

		if height != 12345 {
			t.Errorf("Expected block height 12345, got %d", height)
		}
	})

	t.Run("RPCErrorHandling", func(t *testing.T) {
		// Test with invalid address
		_, err := client.GetBalance("invalid-address")
		if err == nil {
			t.Error("Expected error for invalid address")
		}

		// Should contain an RPC error (may be wrapped)
		var rpcErr *walletrpc.RPCError
		if !errors.As(err, &rpcErr) {
			t.Errorf("Expected RPCError in error chain, got %T: %v", err, err)
		}
	})

	t.Run("NetworkErrorHandling", func(t *testing.T) {
		// Create client with invalid endpoint
		invalidClient := walletrpc.NewRPCClient("http://invalid-endpoint:9999")

		_, err := invalidClient.GetBalance(testAddress)
		if err == nil {
			t.Error("Expected network error")
		}

		// Should be a network error, not RPC error
		if _, ok := err.(*walletrpc.RPCError); ok {
			t.Error("Expected network error, got RPCError")
		}
	})
}

// TestFullWalletWorkflow tests the complete wallet workflow
func TestFullWalletWorkflow(t *testing.T) {
	// Create mock server
	mockServer := NewMockRPCServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	walletPath := filepath.Join(tmpDir, "workflow-wallet.dat")

	config := &core.WalletConfig{
		RPCEndpoint: mockServer.URL(),
		WalletPath:  walletPath,
	}

	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Step 1: Create new wallet
		seedPhrase, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}

		// Step 2: Import additional accounts
		for i := 0; i < 3; i++ {
			additionalSeed, err := core.GenerateSeedPhrase(12)
			if err != nil {
				t.Fatalf("Failed to generate additional seed: %v", err)
			}
			_, err = wallet.ImportSeedPhrase(additionalSeed, fmt.Sprintf("Account %d", i+2))
			if err != nil {
				t.Fatalf("Failed to import seed phrase: %v", err)
			}
		}

		// Step 3: Set up mock balances
		for i := 0; i < wallet.AccountCount(); i++ {
			account := wallet.GetAccount(i)
			mockServer.AddAccount(account.Address, int64((i+1)*1000000))
		}

		// Step 4: Create RPC client and refresh balances
		rpcClient := walletrpc.NewRPCClient(mockServer.URL())

		// Simulate balance refresh
		for i := 0; i < wallet.AccountCount(); i++ {
			account := wallet.GetAccount(i)
			balance, err := rpcClient.GetBalance(account.Address)
			if err != nil {
				t.Fatalf("Failed to get balance for account %d: %v", i, err)
			}
			account.Balance = balance
			account.LastUpdate = time.Now()
		}

		// Step 5: Perform transfer between accounts
		wallet.SetActiveAccount(0)
		senderAccount := wallet.GetActiveAccount()
		recipientAccount := wallet.GetAccount(1)

		transferReq := &core.TransferRequest{
			From:   senderAccount.Address,
			To:     recipientAccount.Address,
			Amount: 50000,
			Memo:   "Workflow test transfer",
		}

		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			t.Fatalf("Failed to build transfer: %v", err)
		}

		// Step 6: Submit transaction
		result, err := wallet.SubmitTransaction(tx, rpcClient)
		if err != nil {
			t.Fatalf("Failed to submit transaction: %v", err)
		}

		if !result.Success {
			t.Errorf("Transaction failed: %s", result.Error)
		}

		if result.Signature == "" {
			t.Error("Expected transaction signature")
		}

		// Step 7: Save wallet
		password := "workflow-password"
		if err := wallet.Save(password); err != nil {
			t.Fatalf("Failed to save wallet: %v", err)
		}

		// Step 8: Load wallet and verify state
		loadedWallet, err := core.LoadWallet(walletPath, password)
		if err != nil {
			t.Fatalf("Failed to load wallet: %v", err)
		}

		if loadedWallet.AccountCount() != wallet.AccountCount() {
			t.Errorf("Account count mismatch after load")
		}

		// Verify all accounts are preserved
		for i := 0; i < wallet.AccountCount(); i++ {
			original := wallet.GetAccount(i)
			loaded := loadedWallet.GetAccount(i)

			if original.Address != loaded.Address {
				t.Errorf("Account %d address mismatch after load", i)
			}

			if original.Label != loaded.Label {
				t.Errorf("Account %d label mismatch after load", i)
			}
		}
	})
}

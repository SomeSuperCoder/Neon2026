package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
)

// E2ETestSuite manages the end-to-end test environment
type E2ETestSuite struct {
	t           *testing.T
	tmpDir      string
	devnetCmd   *exec.Cmd
	rpcEndpoint string
	walletPath  string
	cleanup     func()
}

// NewE2ETestSuite creates a new end-to-end test suite
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
	tmpDir := t.TempDir()

	suite := &E2ETestSuite{
		t:           t,
		tmpDir:      tmpDir,
		rpcEndpoint: "http://127.0.0.1:8899",
		walletPath:  filepath.Join(tmpDir, "e2e-wallet.dat"),
	}

	return suite
}

// StartDevnet starts the devnet with RPC node for testing
func (s *E2ETestSuite) StartDevnet() error {
	s.t.Helper()

	// Check if binary exists (use absolute path)
	binaryPath := "bin/poh-node"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try relative path from test directory
		binaryPath = "../../bin/poh-node"
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			return fmt.Errorf("poh-node binary not found. Please run 'go build -o bin/poh-node cmd/main.go' first")
		}
	}

	// Start devnet with 3 validators and RPC node
	devnetPath := "./devnet.sh"
	// Try different paths for devnet script
	if _, err := os.Stat(devnetPath); os.IsNotExist(err) {
		devnetPath = "../../devnet.sh"
		if _, err := os.Stat(devnetPath); os.IsNotExist(err) {
			return fmt.Errorf("devnet.sh script not found")
		}
	}

	dbDir := filepath.Join(s.tmpDir, "devnet-data")
	logDir := filepath.Join(s.tmpDir, "logs")

	s.devnetCmd = exec.Command(devnetPath, "start", "3",
		"--db-dir", dbDir,
		"--log-dir", logDir,
		"--rpc-port", "8899")

	// Run from project root (no need to change directory)
	s.devnetCmd.Stdout = os.Stdout
	s.devnetCmd.Stderr = os.Stderr

	if err := s.devnetCmd.Start(); err != nil {
		return fmt.Errorf("failed to start devnet: %v", err)
	}

	// Wait for devnet to be ready
	if err := s.waitForDevnet(); err != nil {
		s.StopDevnet()
		return fmt.Errorf("devnet failed to start properly: %v", err)
	}

	s.t.Logf("Devnet started successfully with RPC endpoint: %s", s.rpcEndpoint)
	return nil
}

// StopDevnet stops the running devnet
func (s *E2ETestSuite) StopDevnet() {
	s.t.Helper()

	if s.devnetCmd != nil && s.devnetCmd.Process != nil {
		// Send interrupt signal to devnet script
		s.devnetCmd.Process.Signal(syscall.SIGINT)

		// Wait for graceful shutdown with timeout
		done := make(chan error, 1)
		go func() {
			done <- s.devnetCmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown completed
		case <-time.After(10 * time.Second):
			// Force kill if timeout
			s.devnetCmd.Process.Kill()
		}

		// Also run devnet.sh stop to ensure cleanup
		stopCmd := exec.Command("./devnet.sh", "stop")
		if _, err := os.Stat("./devnet.sh"); os.IsNotExist(err) {
			stopCmd = exec.Command("../../devnet.sh", "stop")
		}
		stopCmd.Run() // Ignore errors
	}
}

// waitForDevnet waits for the devnet to be ready
func (s *E2ETestSuite) waitForDevnet() error {
	client := walletrpc.NewRPCClient(s.rpcEndpoint)

	// Wait up to 30 seconds for RPC to be ready
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for devnet to be ready")
		case <-ticker.C:
			// Try to get block height to test if RPC is ready
			if _, err := client.GetBlockHeight(); err == nil {
				// RPC is responding, wait a bit more for stability
				time.Sleep(2 * time.Second)
				return nil
			}
		}
	}
}

// TestEndToEndWalletWorkflow tests the complete wallet workflow with real devnet
func TestEndToEndWalletWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	suite := NewE2ETestSuite(t)

	// Start devnet
	if err := suite.StartDevnet(); err != nil {
		t.Fatalf("Failed to start devnet: %v", err)
	}
	defer suite.StopDevnet()

	t.Run("CompleteWorkflow", func(t *testing.T) {
		suite.testCompleteWorkflow(t)
	})
}

// testCompleteWorkflow runs the complete end-to-end workflow
func (s *E2ETestSuite) testCompleteWorkflow(t *testing.T) {
	// Step 1: Create wallet and generate accounts
	t.Log("Step 1: Creating wallet and generating accounts")

	config := &core.WalletConfig{
		RPCEndpoint: s.rpcEndpoint,
		WalletPath:  s.walletPath,
	}

	// Generate seed phrase and create wallet
	seedPhrase, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import additional seed phrases to create multiple accounts
	for i := 0; i < 2; i++ {
		additionalSeed, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate additional seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(additionalSeed, fmt.Sprintf("Account %d", i+2))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	if wallet.AccountCount() != 3 {
		t.Errorf("Expected 3 accounts, got %d", wallet.AccountCount())
	}

	// Step 2: Save wallet with password
	t.Log("Step 2: Saving wallet with password")

	password := "e2e-test-password-123"
	if err := wallet.Save(password); err != nil {
		t.Fatalf("Failed to save wallet: %v", err)
	}

	// Verify wallet file exists
	if !e2eFileExists(s.walletPath) {
		t.Fatal("Wallet file was not created")
	}

	// Step 3: Test wallet lock/unlock functionality
	t.Log("Step 3: Testing wallet lock/unlock")

	// Load wallet (simulates unlock)
	loadedWallet, err := core.LoadWallet(s.walletPath, password)
	if err != nil {
		t.Fatalf("Failed to load wallet (unlock): %v", err)
	}

	if loadedWallet.AccountCount() != wallet.AccountCount() {
		t.Errorf("Account count mismatch after unlock: expected %d, got %d",
			wallet.AccountCount(), loadedWallet.AccountCount())
	}

	// Test with wrong password (should fail)
	_, err = core.LoadWallet(s.walletPath, "wrong-password")
	if err == nil {
		t.Error("Expected error with wrong password")
	}

	walletErr, ok := err.(*core.WalletError)
	if !ok || walletErr.Code != core.ErrInvalidPassword {
		t.Errorf("Expected invalid password error, got: %v", err)
	}

	// Continue with loaded wallet
	wallet = loadedWallet

	// Step 4: Create RPC client and check connectivity
	t.Log("Step 4: Testing RPC connectivity")

	rpcClient := walletrpc.NewRPCClient(s.rpcEndpoint)

	// Test basic RPC functionality
	blockHeight, err := rpcClient.GetBlockHeight()
	if err != nil {
		t.Fatalf("Failed to get block height: %v", err)
	}

	if blockHeight < 0 {
		t.Errorf("Invalid block height: %d", blockHeight)
	}

	t.Logf("Current block height: %d", blockHeight)

	// Step 5: Check initial account balances
	t.Log("Step 5: Checking initial account balances")

	for i := 0; i < wallet.AccountCount(); i++ {
		account := wallet.GetAccount(i)
		balance, err := rpcClient.GetBalance(account.Address)
		if err != nil {
			t.Fatalf("Failed to get balance for account %d: %v", i, err)
		}

		account.Balance = balance
		account.LastUpdate = time.Now()

		t.Logf("Account %d (%s): balance = %d", i, account.Address[:16]+"...", balance)
	}

	// Step 6: Test account info queries
	t.Log("Step 6: Testing account info queries")

	for i := 0; i < wallet.AccountCount(); i++ {
		account := wallet.GetAccount(i)
		accountInfo, err := rpcClient.GetAccountInfo(account.Address)
		if err != nil {
			t.Fatalf("Failed to get account info for account %d: %v", i, err)
		}

		// Account might not exist yet (balance 0), which is fine
		if accountInfo != nil {
			if accountInfo.Address != account.Address {
				t.Errorf("Address mismatch for account %d", i)
			}

			if accountInfo.Balance != account.Balance {
				t.Errorf("Balance mismatch for account %d: expected %d, got %d",
					i, account.Balance, accountInfo.Balance)
			}
		}
	}

	// Step 7: Test transaction history (should be empty initially)
	t.Log("Step 7: Testing initial transaction history")

	account0 := wallet.GetAccount(0)
	history, err := rpcClient.GetTransactionHistory(account0.Address, 10)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	// Should be empty for new accounts
	if len(history) != 0 {
		t.Logf("Note: Account has %d existing transactions", len(history))
	}

	// Step 8: Execute transfers between accounts (if we have sufficient balance)
	t.Log("Step 8: Testing transfer functionality")

	// Set active account to first account
	if err := wallet.SetActiveAccount(0); err != nil {
		t.Fatalf("Failed to set active account: %v", err)
	}

	senderAccount := wallet.GetActiveAccount()
	recipientAccount := wallet.GetAccount(1)

	// Only attempt transfer if sender has sufficient balance
	transferAmount := int64(1000)
	if senderAccount.Balance >= transferAmount {
		t.Logf("Executing transfer: %d from %s to %s",
			transferAmount, senderAccount.Address[:16]+"...", recipientAccount.Address[:16]+"...")

		transferReq := &core.TransferRequest{
			From:   senderAccount.Address,
			To:     recipientAccount.Address,
			Amount: transferAmount,
			Memo:   "E2E test transfer",
		}

		// Build transaction
		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			t.Fatalf("Failed to build transfer transaction: %v", err)
		}

		// Submit transaction
		result, err := wallet.SubmitTransaction(tx, rpcClient)
		if err != nil {
			t.Fatalf("Failed to submit transaction: %v", err)
		}

		if !result.Success {
			t.Errorf("Transaction failed: %s", result.Error)
		} else {
			t.Logf("Transfer successful! Signature: %s", result.Signature)

			// Wait a moment for transaction to be processed
			time.Sleep(2 * time.Second)

			// Verify transaction status
			status, err := rpcClient.GetTransactionStatus(result.Signature)
			if err != nil {
				t.Errorf("Failed to get transaction status: %v", err)
			} else {
				t.Logf("Transaction status: confirmed=%v, block=%d",
					status.Confirmed, status.BlockHeight)
			}

			// Check updated balances
			newSenderBalance, err := rpcClient.GetBalance(senderAccount.Address)
			if err != nil {
				t.Errorf("Failed to get updated sender balance: %v", err)
			} else {
				t.Logf("Sender balance after transfer: %d (was %d)",
					newSenderBalance, senderAccount.Balance)
			}

			newRecipientBalance, err := rpcClient.GetBalance(recipientAccount.Address)
			if err != nil {
				t.Errorf("Failed to get updated recipient balance: %v", err)
			} else {
				t.Logf("Recipient balance after transfer: %d (was %d)",
					newRecipientBalance, recipientAccount.Balance)
			}
		}
	} else {
		t.Logf("Skipping transfer test - insufficient balance: %d < %d",
			senderAccount.Balance, transferAmount)
	}

	// Step 9: Test transaction history after transfer
	t.Log("Step 9: Testing transaction history after operations")

	// Check transaction history for both accounts
	for i := 0; i < 2; i++ {
		account := wallet.GetAccount(i)
		history, err := rpcClient.GetTransactionHistory(account.Address, 20)
		if err != nil {
			t.Errorf("Failed to get transaction history for account %d: %v", i, err)
			continue
		}

		t.Logf("Account %d transaction history: %d transactions", i, len(history))

		// Log recent transactions
		for j, tx := range history {
			if j >= 3 { // Only show first 3
				break
			}
			t.Logf("  TX %d: %s (block %d, success=%v)",
				j+1, tx.Signature[:16]+"...", tx.BlockHeight, tx.Success)
		}
	}

	// Step 10: Test wallet persistence and reload
	t.Log("Step 10: Testing wallet persistence")

	// Save current state
	if err := wallet.Save(password); err != nil {
		t.Fatalf("Failed to save wallet state: %v", err)
	}

	// Reload wallet
	reloadedWallet, err := core.LoadWallet(s.walletPath, password)
	if err != nil {
		t.Fatalf("Failed to reload wallet: %v", err)
	}

	// Verify all data is preserved
	if reloadedWallet.AccountCount() != wallet.AccountCount() {
		t.Errorf("Account count mismatch after reload")
	}

	for i := 0; i < wallet.AccountCount(); i++ {
		original := wallet.GetAccount(i)
		reloaded := reloadedWallet.GetAccount(i)

		if original.Address != reloaded.Address {
			t.Errorf("Account %d address mismatch after reload", i)
		}

		if original.Label != reloaded.Label {
			t.Errorf("Account %d label mismatch after reload", i)
		}
	}

	// Step 11: Test error handling scenarios
	t.Log("Step 11: Testing error handling")

	// Test invalid address
	_, err = rpcClient.GetBalance("invalid-address")
	if err == nil {
		t.Error("Expected error for invalid address")
	}

	// Test transfer with insufficient funds
	if senderAccount.Balance > 0 {
		excessiveAmount := senderAccount.Balance + 1000000
		transferReq := &core.TransferRequest{
			From:   senderAccount.Address,
			To:     recipientAccount.Address,
			Amount: excessiveAmount,
			Memo:   "Should fail - insufficient funds",
		}

		tx, err := wallet.BuildTransferTransaction(transferReq)
		if err != nil {
			// This might fail at build time due to validation
			t.Logf("Transfer build failed as expected: %v", err)
		} else {
			// If build succeeds, submission should fail
			result, err := wallet.SubmitTransaction(tx, rpcClient)
			if err != nil {
				t.Logf("Transfer submission failed as expected: %v", err)
			} else if !result.Success {
				t.Logf("Transfer failed as expected: %s", result.Error)
			} else {
				t.Error("Expected transfer to fail due to insufficient funds")
			}
		}
	}

	t.Log("End-to-end test completed successfully!")
}

// TestEndToEndWalletCreationFlow tests wallet creation scenarios
func TestEndToEndWalletCreationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	suite := NewE2ETestSuite(t)

	// Start devnet
	if err := suite.StartDevnet(); err != nil {
		t.Fatalf("Failed to start devnet: %v", err)
	}
	defer suite.StopDevnet()

	t.Run("CreateNewWallet", func(t *testing.T) {
		suite.testWalletCreation(t, 12)
	})

	t.Run("CreateWalletWith24Words", func(t *testing.T) {
		suite.testWalletCreation(t, 24)
	})

	t.Run("RestoreWalletFromSeed", func(t *testing.T) {
		suite.testWalletRestoration(t)
	})
}

// testWalletCreation tests wallet creation with different word counts
func (s *E2ETestSuite) testWalletCreation(t *testing.T, wordCount int) {
	walletPath := filepath.Join(s.tmpDir, fmt.Sprintf("wallet-%d-words.dat", wordCount))

	config := &core.WalletConfig{
		RPCEndpoint: s.rpcEndpoint,
		WalletPath:  walletPath,
	}

	// Generate seed phrase
	seedPhrase, err := core.GenerateSeedPhrase(wordCount)
	if err != nil {
		t.Fatalf("Failed to generate %d-word seed phrase: %v", wordCount, err)
	}

	// Verify word count
	words := strings.Fields(seedPhrase)
	if len(words) != wordCount {
		t.Errorf("Expected %d words, got %d", wordCount, len(words))
	}

	// Create wallet
	wallet, err := core.NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Verify initial state
	if wallet.AccountCount() != 1 {
		t.Errorf("Expected 1 account, got %d", wallet.AccountCount())
	}

	activeAccount := wallet.GetActiveAccount()
	if activeAccount == nil {
		t.Fatal("No active account")
	}

	// Verify address format
	if len(activeAccount.Address) != 64 { // 32 bytes hex encoded
		t.Errorf("Invalid address length: %d", len(activeAccount.Address))
	}

	// Save wallet
	password := fmt.Sprintf("test-password-%d", wordCount)
	if err := wallet.Save(password); err != nil {
		t.Fatalf("Failed to save wallet: %v", err)
	}

	// Verify file exists and has correct permissions
	if !e2eFileExists(walletPath) {
		t.Error("Wallet file was not created")
	}

	info, err := os.Stat(walletPath)
	if err != nil {
		t.Fatalf("Failed to stat wallet file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Test RPC connectivity with new account
	rpcClient := walletrpc.NewRPCClient(s.rpcEndpoint)

	balance, err := rpcClient.GetBalance(activeAccount.Address)
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	t.Logf("New account balance: %d", balance)
}

// testWalletRestoration tests wallet restoration from seed phrase
func (s *E2ETestSuite) testWalletRestoration(t *testing.T) {
	originalWalletPath := filepath.Join(s.tmpDir, "original-wallet.dat")
	restoredWalletPath := filepath.Join(s.tmpDir, "restored-wallet.dat")

	// Create original wallet
	seedPhrase, err := core.GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	originalConfig := &core.WalletConfig{
		RPCEndpoint: s.rpcEndpoint,
		WalletPath:  originalWalletPath,
	}

	originalWallet, err := core.NewWalletWithSeedPhrase(seedPhrase, originalConfig)
	if err != nil {
		t.Fatalf("Failed to create original wallet: %v", err)
	}

	// Import additional accounts
	for i := 0; i < 2; i++ {
		additionalSeed, err := core.GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate additional seed: %v", err)
		}

		_, err = originalWallet.ImportSeedPhrase(additionalSeed, fmt.Sprintf("Imported %d", i+1))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Save original wallet
	password := "restoration-test-password"
	if err := originalWallet.Save(password); err != nil {
		t.Fatalf("Failed to save original wallet: %v", err)
	}

	// Restore wallet from first seed phrase
	restoredConfig := &core.WalletConfig{
		RPCEndpoint: s.rpcEndpoint,
		WalletPath:  restoredWalletPath,
	}

	restoredWallet, err := core.NewWalletWithSeedPhrase(seedPhrase, restoredConfig)
	if err != nil {
		t.Fatalf("Failed to restore wallet: %v", err)
	}

	// Verify restored account matches original
	originalAccount := originalWallet.GetAccount(0)
	restoredAccount := restoredWallet.GetActiveAccount()

	if originalAccount.Address != restoredAccount.Address {
		t.Errorf("Restored account address mismatch")
	}

	// Verify keys match by testing RPC calls
	rpcClient := walletrpc.NewRPCClient(s.rpcEndpoint)

	originalBalance, err := rpcClient.GetBalance(originalAccount.Address)
	if err != nil {
		t.Fatalf("Failed to get original account balance: %v", err)
	}

	restoredBalance, err := rpcClient.GetBalance(restoredAccount.Address)
	if err != nil {
		t.Fatalf("Failed to get restored account balance: %v", err)
	}

	if originalBalance != restoredBalance {
		t.Errorf("Balance mismatch: original=%d, restored=%d", originalBalance, restoredBalance)
	}

	t.Logf("Wallet restoration successful - balance: %d", restoredBalance)
}

// TestEndToEndRPCErrorHandling tests RPC error scenarios
func TestEndToEndRPCErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	suite := NewE2ETestSuite(t)

	// Start devnet
	if err := suite.StartDevnet(); err != nil {
		t.Fatalf("Failed to start devnet: %v", err)
	}
	defer suite.StopDevnet()

	rpcClient := walletrpc.NewRPCClient(suite.rpcEndpoint)

	t.Run("InvalidAddress", func(t *testing.T) {
		_, err := rpcClient.GetBalance("invalid-address")
		if err == nil {
			t.Error("Expected error for invalid address")
		}
		t.Logf("Got expected error: %v", err)
	})

	t.Run("NonExistentTransaction", func(t *testing.T) {
		status, err := rpcClient.GetTransactionStatus("nonexistent-signature")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if status.Confirmed {
			t.Error("Expected transaction to not be confirmed")
		}
	})

	t.Run("NetworkConnectivity", func(t *testing.T) {
		// Test with invalid endpoint
		invalidClient := walletrpc.NewRPCClient("http://invalid-endpoint:9999")

		_, err := invalidClient.GetBlockHeight()
		if err == nil {
			t.Error("Expected network error")
		}
		t.Logf("Got expected network error: %v", err)
	})
}

// Helper function to check if file exists
func e2eFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

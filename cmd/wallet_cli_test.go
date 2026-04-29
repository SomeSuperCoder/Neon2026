package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/poh-blockchain/internal/wallet"
)

// TestWalletCLICommands tests all wallet CLI commands end-to-end
func TestWalletCLICommands(t *testing.T) {
	// Create temporary directory for test wallets
	tempDir := t.TempDir()

	// Override wallet directory using environment variable
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	t.Run("wallet create", func(t *testing.T) {
		walletName := "test-wallet-1"
		password := "test-password-123"

		// Simulate password input
		input := password + "\n" + password + "\n"

		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet create failed: %v", err)
		}

		// Verify wallet file exists
		walletPath := filepath.Join(tempDir, walletName+".wallet")
		if _, err := os.Stat(walletPath); os.IsNotExist(err) {
			t.Fatalf("wallet file not created: %s", walletPath)
		}

		// Verify wallet can be opened
		w, err := wallet.Open(walletName, password)
		if err != nil {
			t.Fatalf("failed to open created wallet: %v", err)
		}

		if len(w.Keypairs) != 1 {
			t.Fatalf("expected 1 keypair, got %d", len(w.Keypairs))
		}
	})

	t.Run("wallet create - password mismatch", func(t *testing.T) {
		walletName := "test-wallet-mismatch"

		// Simulate mismatched passwords
		input := "password1\npassword2\n"

		err := runWalletCreate(walletName, strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error for mismatched passwords, got nil")
		}

		if !strings.Contains(err.Error(), "do not match") {
			t.Fatalf("expected 'do not match' error, got: %v", err)
		}
	})

	t.Run("wallet create - already exists", func(t *testing.T) {
		walletName := "test-wallet-exists"
		password := "test-password"

		// Create wallet first time
		input := password + "\n" + password + "\n"
		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("first wallet create failed: %v", err)
		}

		// Try to create again
		input = password + "\n" + password + "\n"
		err = runWalletCreate(walletName, strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error for duplicate wallet, got nil")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("wallet list", func(t *testing.T) {
		// Create multiple wallets
		wallets := []string{"list-wallet-1", "list-wallet-2", "list-wallet-3"}
		password := "test-password"

		for _, name := range wallets {
			input := password + "\n" + password + "\n"
			err := runWalletCreate(name, strings.NewReader(input))
			if err != nil {
				t.Fatalf("failed to create wallet %s: %v", name, err)
			}
		}

		// List wallets
		var output bytes.Buffer
		err := runWalletList(&output)
		if err != nil {
			t.Fatalf("wallet list failed: %v", err)
		}

		outputStr := output.String()
		for _, name := range wallets {
			if !strings.Contains(outputStr, name) {
				t.Errorf("wallet list output missing wallet: %s", name)
			}
		}
	})

	t.Run("wallet show", func(t *testing.T) {
		walletName := "show-wallet"
		password := "test-password"

		// Create wallet
		input := password + "\n" + password + "\n"
		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet create failed: %v", err)
		}

		// Show wallet
		var output bytes.Buffer
		err = runWalletShow(walletName, password, &output)
		if err != nil {
			t.Fatalf("wallet show failed: %v", err)
		}

		outputStr := output.String()

		// Verify output contains wallet name
		if !strings.Contains(outputStr, walletName) {
			t.Errorf("output missing wallet name")
		}

		// Verify output contains public key (truncated to 16 hex chars)
		w, _ := wallet.Open(walletName, password)
		expectedPubKey := hex.EncodeToString(w.Keypairs[0].PublicKey[:])[:16]
		if !strings.Contains(outputStr, expectedPubKey) {
			t.Errorf("output missing public key: %s", expectedPubKey)
		}
	})

	t.Run("wallet show - incorrect password", func(t *testing.T) {
		walletName := "show-wallet-wrong-pass"
		password := "correct-password"

		// Create wallet
		input := password + "\n" + password + "\n"
		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet create failed: %v", err)
		}

		// Try to show with wrong password
		var output bytes.Buffer
		err = runWalletShow(walletName, "wrong-password", &output)
		if err == nil {
			t.Fatal("expected error for incorrect password, got nil")
		}

		if !strings.Contains(err.Error(), "password") {
			t.Fatalf("expected password error, got: %v", err)
		}
	})

	t.Run("wallet show - not found", func(t *testing.T) {
		var output bytes.Buffer
		err := runWalletShow("nonexistent-wallet", "password", &output)
		if err == nil {
			t.Fatal("expected error for nonexistent wallet, got nil")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Fatalf("expected 'not found' error, got: %v", err)
		}
	})

	t.Run("wallet export and import", func(t *testing.T) {
		walletName := "export-wallet"
		password := "test-password"
		exportFile := filepath.Join(tempDir, "exported.json")
		importWalletName := "imported-wallet"

		// Create wallet
		input := password + "\n" + password + "\n"
		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet create failed: %v", err)
		}

		// Export wallet
		err = runWalletExport(walletName, password, exportFile)
		if err != nil {
			t.Fatalf("wallet export failed: %v", err)
		}

		// Verify export file exists and is valid JSON
		exportData, err := os.ReadFile(exportFile)
		if err != nil {
			t.Fatalf("failed to read export file: %v", err)
		}

		var exportedData struct {
			Version  int `json:"version"`
			Keypairs []struct {
				PublicKey  []byte `json:"publicKey"`
				PrivateKey []byte `json:"privateKey"`
			} `json:"keypairs"`
		}
		if err := json.Unmarshal(exportData, &exportedData); err != nil {
			t.Fatalf("export file is not valid JSON: %v", err)
		}

		if exportedData.Version != 1 {
			t.Fatalf("expected version 1, got %d", exportedData.Version)
		}

		if len(exportedData.Keypairs) != 1 {
			t.Fatalf("expected 1 keypair in export, got %d", len(exportedData.Keypairs))
		}

		// Import wallet with new name and password
		newPassword := "new-password"
		input = newPassword + "\n" + newPassword + "\n"
		err = runWalletImport(exportFile, importWalletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet import failed: %v", err)
		}

		// Verify imported wallet
		importedWallet, err := wallet.Open(importWalletName, newPassword)
		if err != nil {
			t.Fatalf("failed to open imported wallet: %v", err)
		}

		if len(importedWallet.Keypairs) != 1 {
			t.Fatalf("expected 1 keypair in imported wallet, got %d", len(importedWallet.Keypairs))
		}

		// Verify keypairs match
		originalWallet, _ := wallet.Open(walletName, password)
		if !bytes.Equal(originalWallet.Keypairs[0].PublicKey[:], importedWallet.Keypairs[0].PublicKey[:]) {
			t.Error("imported public key does not match original")
		}
		if !bytes.Equal(originalWallet.Keypairs[0].PrivateKey[:], importedWallet.Keypairs[0].PrivateKey[:]) {
			t.Error("imported private key does not match original")
		}
	})

	t.Run("wallet export - incorrect password", func(t *testing.T) {
		walletName := "export-wallet-wrong-pass"
		password := "correct-password"
		exportFile := filepath.Join(tempDir, "export-fail.json")

		// Create wallet
		input := password + "\n" + password + "\n"
		err := runWalletCreate(walletName, strings.NewReader(input))
		if err != nil {
			t.Fatalf("wallet create failed: %v", err)
		}

		// Try to export with wrong password
		err = runWalletExport(walletName, "wrong-password", exportFile)
		if err == nil {
			t.Fatal("expected error for incorrect password, got nil")
		}

		if !strings.Contains(err.Error(), "password") {
			t.Fatalf("expected password error, got: %v", err)
		}
	})

	t.Run("wallet import - invalid file", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.json")

		// Create invalid JSON file
		err := os.WriteFile(invalidFile, []byte("not valid json"), 0600)
		if err != nil {
			t.Fatalf("failed to create invalid file: %v", err)
		}

		// Try to import
		password := "test-password"
		input := password + "\n" + password + "\n"
		err = runWalletImport(invalidFile, "import-invalid", strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error for invalid import file, got nil")
		}
	})
}

// TestWalletCLIIntegration tests wallet CLI with actual binary execution
func TestWalletCLIIntegration(t *testing.T) {
	// Skip if binary doesn't exist
	binaryPath := "./poh-blockchain"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not built, skipping integration test")
	}

	tempDir := t.TempDir()

	// Override wallet directory
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	t.Run("full CLI workflow", func(t *testing.T) {
		walletName := "cli-test-wallet"
		password := "cli-test-password"

		// Test wallet create
		cmd := exec.Command(binaryPath, "wallet", "create", "--name", walletName)
		cmd.Stdin = strings.NewReader(password + "\n" + password + "\n")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("wallet create command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "created") {
			t.Errorf("unexpected create output: %s", output)
		}

		// Test wallet list
		cmd = exec.Command(binaryPath, "wallet", "list")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("wallet list command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), walletName) {
			t.Errorf("wallet list missing created wallet: %s", output)
		}

		// Test wallet show
		cmd = exec.Command(binaryPath, "wallet", "show", "--name", walletName)
		cmd.Stdin = strings.NewReader(password + "\n")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("wallet show command failed: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), walletName) {
			t.Errorf("wallet show missing wallet name: %s", output)
		}
	})
}

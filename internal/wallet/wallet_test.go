package wallet

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestGetWalletDir tests platform-specific wallet directory resolution
func TestGetWalletDir(t *testing.T) {
	dir, err := GetWalletDir()
	if err != nil {
		t.Fatalf("GetWalletDir failed: %v", err)
	}

	// Verify directory path is not empty
	if dir == "" {
		t.Error("GetWalletDir returned empty string")
	}

	// Verify platform-specific path
	switch runtime.GOOS {
	case "windows":
		if !filepath.IsAbs(dir) {
			t.Errorf("Expected absolute path on Windows, got: %s", dir)
		}
		// Should contain AppData
		// Note: We can't test exact path without mocking environment
	case "linux", "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}
		expectedDir := filepath.Join(home, ".config", "poh-blockchain", "wallets")
		if dir != expectedDir {
			t.Errorf("Expected %s, got %s", expectedDir, dir)
		}
	default:
		// For other platforms, just verify it's an absolute path
		if !filepath.IsAbs(dir) {
			t.Errorf("Expected absolute path, got: %s", dir)
		}
	}
}

// TestGetWalletPath tests wallet file path generation
func TestGetWalletPath(t *testing.T) {
	tests := []struct {
		name       string
		walletName string
	}{
		{"simple name", "my-wallet"},
		{"with spaces", "my wallet"},
		{"with special chars", "wallet_123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetWalletPath(tt.walletName)
			if err != nil {
				t.Fatalf("GetWalletPath failed: %v", err)
			}

			// Verify path is not empty
			if path == "" {
				t.Error("GetWalletPath returned empty string")
			}

			// Verify path ends with .wallet extension
			if filepath.Ext(path) != ".wallet" {
				t.Errorf("Expected .wallet extension, got: %s", filepath.Ext(path))
			}

			// Verify path contains wallet name
			base := filepath.Base(path)
			expectedBase := tt.walletName + ".wallet"
			if base != expectedBase {
				t.Errorf("Expected base name %s, got %s", expectedBase, base)
			}
		})
	}
}

// TestEncryptDecryptRoundTrip tests encryption and decryption
func TestEncryptDecryptRoundTrip(t *testing.T) {
	password := "test-password-123"
	plaintext := []byte("sensitive wallet data")

	// Encrypt
	ciphertext, err := encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Verify ciphertext is different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("Ciphertext should not equal plaintext")
	}

	// Verify ciphertext has expected structure: salt(32) + nonce(12) + ciphertext + tag(16)
	minLength := 32 + 12 + len(plaintext) + 16
	if len(ciphertext) < minLength {
		t.Errorf("Ciphertext too short: expected at least %d bytes, got %d", minLength, len(ciphertext))
	}

	// Decrypt
	decrypted, err := decrypt(ciphertext, password)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	// Verify decrypted matches original
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted data does not match original.\nExpected: %s\nGot: %s", plaintext, decrypted)
	}
}

// TestEncryptDecryptDifferentPasswords tests that different passwords produce different ciphertexts
func TestEncryptDecryptDifferentPasswords(t *testing.T) {
	plaintext := []byte("test data")
	password1 := "password1"
	password2 := "password2"

	ciphertext1, err := encrypt(plaintext, password1)
	if err != nil {
		t.Fatalf("encrypt with password1 failed: %v", err)
	}

	ciphertext2, err := encrypt(plaintext, password2)
	if err != nil {
		t.Fatalf("encrypt with password2 failed: %v", err)
	}

	// Different passwords should produce different ciphertexts
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Different passwords produced identical ciphertexts")
	}
}

// TestDecryptWithIncorrectPassword tests that decryption fails with wrong password
func TestDecryptWithIncorrectPassword(t *testing.T) {
	plaintext := []byte("secret data")
	correctPassword := "correct-password"
	incorrectPassword := "wrong-password"

	// Encrypt with correct password
	ciphertext, err := encrypt(plaintext, correctPassword)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// Try to decrypt with incorrect password
	_, err = decrypt(ciphertext, incorrectPassword)
	if err == nil {
		t.Error("Expected error when decrypting with incorrect password, got none")
	}
}

// TestDecryptCorruptedData tests that decryption fails with corrupted data
func TestDecryptCorruptedData(t *testing.T) {
	password := "test-password"

	tests := []struct {
		name string
		data []byte
	}{
		{"empty data", []byte{}},
		{"too short", []byte{1, 2, 3}},
		{"invalid salt", make([]byte, 32)}, // Only salt, no nonce/ciphertext/tag
		{"corrupted ciphertext", func() []byte {
			plaintext := []byte("test data")
			ciphertext, _ := encrypt(plaintext, password)
			// Corrupt the middle of the ciphertext
			if len(ciphertext) > 50 {
				ciphertext[50] ^= 0xFF
			}
			return ciphertext
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decrypt(tt.data, password)
			if err == nil {
				t.Error("Expected error when decrypting corrupted data, got none")
			}
		})
	}
}

// TestCreateWallet tests wallet creation
func TestCreateWallet(t *testing.T) {
	// Use temporary directory for test
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	name := "test-wallet"
	password := "test-password-123"

	wallet, err := Create(name, password)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify wallet properties
	if wallet.Name != name {
		t.Errorf("Wallet name mismatch: expected %s, got %s", name, wallet.Name)
	}

	if len(wallet.Keypairs) != 1 {
		t.Errorf("Expected 1 keypair, got %d", len(wallet.Keypairs))
	}

	// Verify keypair has valid keys
	kp := wallet.Keypairs[0]
	if kp.PublicKey == [32]byte{} {
		t.Error("Public key is empty")
	}
	if kp.PrivateKey == [64]byte{} {
		t.Error("Private key is empty")
	}

	// Verify wallet file was created
	walletPath := filepath.Join(tempDir, name+".wallet")
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		t.Errorf("Wallet file was not created at %s", walletPath)
	}
}

// TestCreateWalletAlreadyExists tests that creating a wallet with existing name fails
func TestCreateWalletAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	name := "duplicate-wallet"
	password := "password"

	// Create first wallet
	_, err := Create(name, password)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	// Try to create wallet with same name
	_, err = Create(name, password)
	if err == nil {
		t.Error("Expected error when creating wallet with existing name, got none")
	}
}

// TestOpenWallet tests opening an existing wallet
func TestOpenWallet(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	name := "test-wallet"
	password := "test-password-123"

	// Create wallet
	originalWallet, err := Create(name, password)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Open wallet
	openedWallet, err := Open(name, password)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// Verify wallet properties match
	if openedWallet.Name != originalWallet.Name {
		t.Errorf("Wallet name mismatch: expected %s, got %s", originalWallet.Name, openedWallet.Name)
	}

	if len(openedWallet.Keypairs) != len(originalWallet.Keypairs) {
		t.Errorf("Keypair count mismatch: expected %d, got %d", len(originalWallet.Keypairs), len(openedWallet.Keypairs))
	}

	// Verify keypairs match
	if openedWallet.Keypairs[0].PublicKey != originalWallet.Keypairs[0].PublicKey {
		t.Error("Public key mismatch after opening wallet")
	}
	if openedWallet.Keypairs[0].PrivateKey != originalWallet.Keypairs[0].PrivateKey {
		t.Error("Private key mismatch after opening wallet")
	}
}

// TestOpenWalletIncorrectPassword tests that opening with wrong password fails
func TestOpenWalletIncorrectPassword(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	name := "test-wallet"
	correctPassword := "correct-password"
	incorrectPassword := "wrong-password"

	// Create wallet
	_, err := Create(name, correctPassword)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Try to open with incorrect password
	_, err = Open(name, incorrectPassword)
	if err == nil {
		t.Error("Expected error when opening wallet with incorrect password, got none")
	}
}

// TestOpenWalletNotFound tests that opening non-existent wallet fails
func TestOpenWalletNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	_, err := Open("non-existent-wallet", "password")
	if err == nil {
		t.Error("Expected error when opening non-existent wallet, got none")
	}
}

// TestListWallets tests listing all wallets
func TestListWallets(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	// Initially should be empty
	wallets, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(wallets) != 0 {
		t.Errorf("Expected 0 wallets, got %d", len(wallets))
	}

	// Create some wallets
	names := []string{"wallet1", "wallet2", "wallet3"}
	for _, name := range names {
		_, err := Create(name, "password")
		if err != nil {
			t.Fatalf("Create failed for %s: %v", name, err)
		}
	}

	// List wallets
	wallets, err = List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(wallets) != len(names) {
		t.Errorf("Expected %d wallets, got %d", len(names), len(wallets))
	}

	// Verify all wallet names are present
	walletMap := make(map[string]bool)
	for _, w := range wallets {
		walletMap[w] = true
	}

	for _, name := range names {
		if !walletMap[name] {
			t.Errorf("Wallet %s not found in list", name)
		}
	}
}

// TestExportWallet tests exporting wallet to unencrypted JSON
func TestExportWallet(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	name := "test-wallet"
	password := "password"

	// Create wallet
	wallet, err := Create(name, password)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Export wallet
	exportPath := filepath.Join(tempDir, "export.json")
	err = wallet.Export(exportPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); os.IsNotExist(err) {
		t.Errorf("Export file was not created at %s", exportPath)
	}

	// Verify export file is readable JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Export file is empty")
	}
}

// TestImportWallet tests importing wallet from unencrypted JSON
func TestImportWallet(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	// Create and export a wallet
	originalName := "original-wallet"
	originalPassword := "password1"
	originalWallet, err := Create(originalName, originalPassword)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	exportPath := filepath.Join(tempDir, "export.json")
	err = originalWallet.Export(exportPath)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import wallet with new name and password
	importedName := "imported-wallet"
	importedPassword := "password2"
	importedWallet, err := Import(exportPath, importedName, importedPassword)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify imported wallet has correct name
	if importedWallet.Name != importedName {
		t.Errorf("Imported wallet name mismatch: expected %s, got %s", importedName, importedWallet.Name)
	}

	// Verify keypairs match original
	if len(importedWallet.Keypairs) != len(originalWallet.Keypairs) {
		t.Errorf("Keypair count mismatch: expected %d, got %d", len(originalWallet.Keypairs), len(importedWallet.Keypairs))
	}

	if importedWallet.Keypairs[0].PublicKey != originalWallet.Keypairs[0].PublicKey {
		t.Error("Public key mismatch after import")
	}
	if importedWallet.Keypairs[0].PrivateKey != originalWallet.Keypairs[0].PrivateKey {
		t.Error("Private key mismatch after import")
	}

	// Verify imported wallet can be opened with new password
	reopenedWallet, err := Open(importedName, importedPassword)
	if err != nil {
		t.Fatalf("Failed to open imported wallet: %v", err)
	}

	if reopenedWallet.Keypairs[0].PublicKey != originalWallet.Keypairs[0].PublicKey {
		t.Error("Public key mismatch after reopening imported wallet")
	}
}

// TestImportWalletInvalidFile tests that importing invalid file fails
func TestImportWalletInvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	// Create invalid JSON file
	invalidPath := filepath.Join(tempDir, "invalid.json")
	err := os.WriteFile(invalidPath, []byte("not valid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	// Try to import
	_, err = Import(invalidPath, "test-wallet", "password")
	if err == nil {
		t.Error("Expected error when importing invalid file, got none")
	}
}

// TestWalletPasswordValidation tests password requirements
func TestWalletPasswordValidation(t *testing.T) {
	tempDir := t.TempDir()
	originalGetWalletDir := getWalletDirFunc
	getWalletDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getWalletDirFunc = originalGetWalletDir }()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"valid password", "good-password-123", false},
		{"empty password", "", true},
		{"short password", "abc", true},
		{"long password", "this-is-a-very-long-password-that-should-work-fine", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walletName := "test-wallet-" + tt.name
			_, err := Create(walletName, tt.password)

			if tt.expectError && err == nil {
				t.Error("Expected error for invalid password, got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

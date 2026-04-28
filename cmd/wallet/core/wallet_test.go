package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewWallet(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Verify wallet is empty
	if len(wallet.seedPhrases) != 0 {
		t.Errorf("Expected 0 seed phrases, got %d", len(wallet.seedPhrases))
	}

	if len(wallet.accounts) != 0 {
		t.Errorf("Expected 0 accounts, got %d", len(wallet.accounts))
	}

	if wallet.activeIndex != -1 {
		t.Errorf("Expected active index -1, got %d", wallet.activeIndex)
	}
}

func TestNewWalletWithSeedPhrase(t *testing.T) {
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := DefaultConfig()
	wallet, err := NewWalletWithSeedPhrase(mnemonic, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Verify wallet has one seed phrase and one account
	if len(wallet.seedPhrases) != 1 {
		t.Errorf("Expected 1 seed phrase, got %d", len(wallet.seedPhrases))
	}

	if len(wallet.accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(wallet.accounts))
	}

	// Verify active account is set
	if wallet.activeIndex != 0 {
		t.Errorf("Expected active index 0, got %d", wallet.activeIndex)
	}

	// Verify first account
	account := wallet.GetActiveAccount()
	if account == nil {
		t.Fatal("Active account is nil")
	}

	if account.SeedPhraseIndex != 0 {
		t.Errorf("Expected seed phrase index 0, got %d", account.SeedPhraseIndex)
	}
}

func TestImportSeedPhrase(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import multiple seed phrases
	for i := 0; i < 10; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		account, err := wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase %d: %v", i, err)
		}

		if account.SeedPhraseIndex != i {
			t.Errorf("Expected seed phrase index %d, got %d", i, account.SeedPhraseIndex)
		}
	}

	// Verify total accounts
	if len(wallet.accounts) != 10 {
		t.Errorf("Expected 10 accounts, got %d", len(wallet.accounts))
	}

	if len(wallet.seedPhrases) != 10 {
		t.Errorf("Expected 10 seed phrases, got %d", len(wallet.seedPhrases))
	}

	// Verify all accounts have unique addresses
	addresses := make(map[string]bool)
	for _, acc := range wallet.accounts {
		if addresses[acc.Address] {
			t.Errorf("Duplicate address found: %s", acc.Address)
		}
		addresses[acc.Address] = true
	}
}

func TestImportDuplicateSeedPhrase(t *testing.T) {
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import seed phrase
	_, err = wallet.ImportSeedPhrase(mnemonic, "Account 1")
	if err != nil {
		t.Fatalf("Failed to import seed phrase: %v", err)
	}

	// Try to import same seed phrase again
	_, err = wallet.ImportSeedPhrase(mnemonic, "Account 2")
	if err == nil {
		t.Error("Expected error for duplicate seed phrase")
	}

	walletErr, ok := err.(*WalletError)
	if !ok {
		t.Error("Expected WalletError")
	}

	if walletErr.Code != ErrDuplicateSeedPhrase {
		t.Errorf("Expected error code %s, got %s", ErrDuplicateSeedPhrase, walletErr.Code)
	}

	// Verify only one account exists
	if len(wallet.accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(wallet.accounts))
	}
}

func TestImportInvalidSeedPhrase(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	_, err = wallet.ImportSeedPhrase("invalid seed phrase", "Account 1")
	if err == nil {
		t.Error("Expected error for invalid seed phrase")
	}

	walletErr, ok := err.(*WalletError)
	if !ok {
		t.Error("Expected WalletError")
	}

	if walletErr.Code != ErrInvalidSeedPhrase {
		t.Errorf("Expected error code %s, got %s", ErrInvalidSeedPhrase, walletErr.Code)
	}
}

func TestSetActiveAccount(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import multiple seed phrases
	for i := 0; i < 5; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Set active account
	if err := wallet.SetActiveAccount(3); err != nil {
		t.Fatalf("Failed to set active account: %v", err)
	}

	// Verify active account
	active := wallet.GetActiveAccount()
	if active.SeedPhraseIndex != 3 {
		t.Errorf("Expected active account seed phrase index 3, got %d", active.SeedPhraseIndex)
	}

	// Try to set invalid index
	err = wallet.SetActiveAccount(10)
	if err == nil {
		t.Error("Expected error for invalid account index")
	}
}

func TestSaveAndLoadWallet(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	walletPath := filepath.Join(tmpDir, "test-wallet.dat")

	config := DefaultConfig()
	config.WalletPath = walletPath

	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import multiple seed phrases
	for i := 0; i < 5; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Set active account
	wallet.SetActiveAccount(2)

	// Save wallet
	password := "test-password-123"
	if err := wallet.Save(password); err != nil {
		t.Fatalf("Failed to save wallet: %v", err)
	}

	// Verify file exists with correct permissions
	info, err := os.Stat(walletPath)
	if err != nil {
		t.Fatalf("Wallet file not created: %v", err)
	}

	// Check file permissions (0600)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
	}

	// Load wallet
	loadedWallet, err := LoadWallet(walletPath, password)
	if err != nil {
		t.Fatalf("Failed to load wallet: %v", err)
	}

	// Verify loaded wallet
	if len(loadedWallet.seedPhrases) != len(wallet.seedPhrases) {
		t.Errorf("Expected %d seed phrases, got %d", len(wallet.seedPhrases), len(loadedWallet.seedPhrases))
	}

	if len(loadedWallet.accounts) != len(wallet.accounts) {
		t.Errorf("Expected %d accounts, got %d", len(wallet.accounts), len(loadedWallet.accounts))
	}

	if loadedWallet.activeIndex != wallet.activeIndex {
		t.Errorf("Expected active index %d, got %d", wallet.activeIndex, loadedWallet.activeIndex)
	}

	// Verify seed phrases match
	for i, sp := range wallet.seedPhrases {
		if loadedWallet.seedPhrases[i] != sp {
			t.Errorf("Seed phrase %d mismatch", i)
		}
	}

	// Verify accounts match
	for i, acc := range wallet.accounts {
		loadedAcc := loadedWallet.accounts[i]
		if acc.Address != loadedAcc.Address {
			t.Errorf("Account %d address mismatch", i)
		}
		if acc.Label != loadedAcc.Label {
			t.Errorf("Account %d label mismatch", i)
		}
		if acc.SeedPhraseIndex != loadedAcc.SeedPhraseIndex {
			t.Errorf("Account %d seed phrase index mismatch", i)
		}
	}
}

func TestLoadWalletInvalidPassword(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	walletPath := filepath.Join(tmpDir, "test-wallet.dat")

	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := DefaultConfig()
	config.WalletPath = walletPath

	wallet, err := NewWalletWithSeedPhrase(mnemonic, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	password := "correct-password"
	if err := wallet.Save(password); err != nil {
		t.Fatalf("Failed to save wallet: %v", err)
	}

	// Try to load with wrong password
	_, err = LoadWallet(walletPath, "wrong-password")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	walletErr, ok := err.(*WalletError)
	if !ok {
		t.Error("Expected WalletError")
	}

	if walletErr.Code != ErrInvalidPassword {
		t.Errorf("Expected error code %s, got %s", ErrInvalidPassword, walletErr.Code)
	}
}

func TestGetAccount(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import accounts
	for i := 0; i < 5; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Get existing account
	account := wallet.GetAccount(2)
	if account == nil {
		t.Error("Expected account, got nil")
	}
	if account.SeedPhraseIndex != 2 {
		t.Errorf("Expected seed phrase index 2, got %d", account.SeedPhraseIndex)
	}

	// Get non-existent account
	account = wallet.GetAccount(10)
	if account != nil {
		t.Error("Expected nil for non-existent account")
	}
}

func TestGetAccountBySeedPhraseIndex(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import accounts
	for i := 0; i < 5; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Get account by seed phrase index
	account := wallet.GetAccountBySeedPhraseIndex(3)
	if account == nil {
		t.Error("Expected account, got nil")
	}
	if account.SeedPhraseIndex != 3 {
		t.Errorf("Expected seed phrase index 3, got %d", account.SeedPhraseIndex)
	}

	// Get non-existent
	account = wallet.GetAccountBySeedPhraseIndex(10)
	if account != nil {
		t.Error("Expected nil for non-existent seed phrase index")
	}
}

func TestSupport100SeedPhrases(t *testing.T) {
	// Test requirement: support minimum of 100 imported seed phrases
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import 100 seed phrases
	for i := 0; i < 100; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase %d: %v", i, err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase %d: %v", i, err)
		}
	}

	// Verify we have 100 accounts
	if wallet.AccountCount() != 100 {
		t.Errorf("Expected 100 accounts, got %d", wallet.AccountCount())
	}

	// Verify all addresses are unique
	addresses := make(map[string]bool)
	for _, acc := range wallet.accounts {
		if addresses[acc.Address] {
			t.Errorf("Duplicate address found: %s", acc.Address)
		}
		addresses[acc.Address] = true
	}

	if len(addresses) != 100 {
		t.Errorf("Expected 100 unique addresses, got %d", len(addresses))
	}
}

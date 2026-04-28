package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Wallet represents the main wallet structure supporting multiple seed phrases
type Wallet struct {
	seedPhrases []string   // Multiple imported seed phrases
	accounts    []*Account // One account per seed phrase (at index 0)
	activeIndex int
	config      *WalletConfig
	encrypted   bool
}

// NewWallet creates a new empty wallet
func NewWallet(config *WalletConfig) (*Wallet, error) {
	return &Wallet{
		seedPhrases: make([]string, 0),
		accounts:    make([]*Account, 0),
		activeIndex: -1,
		config:      config,
		encrypted:   false,
	}, nil
}

// NewWalletWithSeedPhrase creates a new wallet and imports the first seed phrase
func NewWalletWithSeedPhrase(seedPhrase string, config *WalletConfig) (*Wallet, error) {
	wallet, err := NewWallet(config)
	if err != nil {
		return nil, err
	}

	// Import the first seed phrase
	_, err = wallet.ImportSeedPhrase(seedPhrase, "Account 1")
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// ImportSeedPhrase imports a new seed phrase and derives account at index 0
func (w *Wallet) ImportSeedPhrase(seedPhrase string, label string) (*Account, error) {
	if !ValidateSeedPhrase(seedPhrase) {
		return nil, &WalletError{
			Code:    ErrInvalidSeedPhrase,
			Message: "invalid seed phrase",
		}
	}

	// Check for duplicate seed phrase
	for _, existing := range w.seedPhrases {
		if existing == seedPhrase {
			return nil, &WalletError{
				Code:    ErrDuplicateSeedPhrase,
				Message: "seed phrase already imported",
			}
		}
	}

	// Convert mnemonic to seed
	seed, err := MnemonicToSeed(seedPhrase, "")
	if err != nil {
		return nil, err
	}

	// Derive key at index 0 using path m/44'/501'/0'/0'/0'
	masterHDKey, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		return nil, err
	}

	// Derive account at index 0
	accountHDKey, err := masterHDKey.DeriveAccount(0)
	if err != nil {
		return nil, err
	}

	// Convert to Ed25519 keypair
	publicKey, privateKey := accountHDKey.ToEd25519()

	// Create account
	var pubKey32 [32]byte
	copy(pubKey32[:], publicKey)

	seedPhraseIndex := len(w.seedPhrases)

	account := &Account{
		SeedPhraseIndex: seedPhraseIndex,
		PublicKey:       pubKey32,
		PrivateKey:      privateKey,
		Address:         hex.EncodeToString(publicKey),
		Label:           label,
		Balance:         0,
		LastUpdate:      time.Now(),
	}

	// Add seed phrase and account
	w.seedPhrases = append(w.seedPhrases, seedPhrase)
	w.accounts = append(w.accounts, account)

	// Set as active if it's the first account
	if w.activeIndex == -1 {
		w.activeIndex = 0
	}

	return account, nil
}

// GetActiveAccount returns the currently active account
func (w *Wallet) GetActiveAccount() *Account {
	if w.activeIndex < 0 || w.activeIndex >= len(w.accounts) {
		return nil
	}
	return w.accounts[w.activeIndex]
}

// GetActiveAccountIndex returns the index of the currently active account
func (w *Wallet) GetActiveAccountIndex() int {
	return w.activeIndex
}

// SetActiveAccount sets the active account by index
func (w *Wallet) SetActiveAccount(index int) error {
	if index < 0 || index >= len(w.accounts) {
		return &WalletError{
			Code:    ErrAccountNotFound,
			Message: fmt.Sprintf("account index %d not found", index),
		}
	}
	w.activeIndex = index
	return nil
}

// GetAccounts returns all accounts
func (w *Wallet) GetAccounts() []*Account {
	return w.accounts
}

// GetAccount returns an account by its index in the accounts array
func (w *Wallet) GetAccount(index int) *Account {
	if index < 0 || index >= len(w.accounts) {
		return nil
	}
	return w.accounts[index]
}

// GetAccountBySeedPhraseIndex returns an account by its seed phrase index
func (w *Wallet) GetAccountBySeedPhraseIndex(seedPhraseIndex int) *Account {
	for _, acc := range w.accounts {
		if acc.SeedPhraseIndex == seedPhraseIndex {
			return acc
		}
	}
	return nil
}

// RefreshBalances refreshes balances for all accounts (placeholder for RPC integration)
func (w *Wallet) RefreshBalances() error {
	// This will be implemented when RPC client is integrated
	// For now, just update the LastUpdate timestamp
	for _, acc := range w.accounts {
		acc.LastUpdate = time.Now()
	}
	return nil
}

// Save encrypts and saves the wallet to a file
func (w *Wallet) Save(password string) error {
	if w.config.WalletPath == "" {
		return &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "wallet path not configured",
		}
	}

	// Encrypt wallet
	encrypted, err := EncryptWallet(w, password)
	if err != nil {
		return err
	}

	// Create wallet file structure
	walletFile := WalletFile{
		Version:   1,
		Encrypted: *encrypted,
		Config:    *w.config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(walletFile, "", "  ")
	if err != nil {
		return &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to marshal wallet file",
			Cause:   err,
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(w.config.WalletPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to create wallet directory",
			Cause:   err,
		}
	}

	// Write to file with restricted permissions (0600)
	if err := os.WriteFile(w.config.WalletPath, data, 0600); err != nil {
		return &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to write wallet file",
			Cause:   err,
		}
	}

	w.encrypted = true
	return nil
}

// LoadWallet loads and decrypts a wallet from a file
func LoadWallet(path string, password string) (*Wallet, error) {
	// Read wallet file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to read wallet file",
			Cause:   err,
		}
	}

	// Unmarshal wallet file
	var walletFile WalletFile
	if err := json.Unmarshal(data, &walletFile); err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to parse wallet file",
			Cause:   err,
		}
	}

	// Set wallet path in config
	walletFile.Config.WalletPath = path

	// Decrypt wallet
	wallet, err := DecryptWallet(&walletFile.Encrypted, password, &walletFile.Config)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetSeedPhrases returns all seed phrases (use with caution)
func (w *Wallet) GetSeedPhrases() []string {
	return w.seedPhrases
}

// GetSeedPhrase returns a specific seed phrase by index (use with caution)
func (w *Wallet) GetSeedPhrase(index int) string {
	if index < 0 || index >= len(w.seedPhrases) {
		return ""
	}
	return w.seedPhrases[index]
}

// GetConfig returns the wallet configuration
func (w *Wallet) GetConfig() *WalletConfig {
	return w.config
}

// IsEncrypted returns whether the wallet is encrypted
func (w *Wallet) IsEncrypted() bool {
	return w.encrypted
}

// AccountCount returns the number of accounts in the wallet
func (w *Wallet) AccountCount() int {
	return len(w.accounts)
}

package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Keypair represents an Ed25519 keypair
type Keypair struct {
	PublicKey  [32]byte `json:"publicKey"`
	PrivateKey [64]byte `json:"privateKey"`
}

// Wallet represents a password-protected wallet containing keypairs
type Wallet struct {
	Name     string    `json:"name"`
	Keypairs []Keypair `json:"keypairs"`
}

// walletFile represents the JSON structure stored in wallet files
type walletFile struct {
	Version  int       `json:"version"`
	Keypairs []Keypair `json:"keypairs"`
}

// Argon2id parameters
const (
	argon2Memory      = 64 * 1024 // 64 MB
	argon2Iterations  = 3
	argon2Parallelism = 4
	argon2KeyLength   = 32 // AES-256

	saltLength  = 32
	nonceLength = 12
	tagLength   = 16

	minPasswordLength = 8
)

// getWalletDirFunc is a variable to allow mocking in tests
var getWalletDirFunc = GetWalletDir

// GetWalletDir returns the platform-specific wallet directory
func GetWalletDir() (string, error) {
	// Check for environment variable override (for testing)
	if envDir := os.Getenv("POH_WALLET_DIR"); envDir != "" {
		return envDir, nil
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "poh-blockchain", "wallets"), nil
	default: // linux, darwin
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, ".config", "poh-blockchain", "wallets"), nil
	}
}

// GetWalletPath returns the full path to a wallet file
func GetWalletPath(name string) (string, error) {
	dir, err := getWalletDirFunc()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".wallet"), nil
}

// deriveKey derives an encryption key from a password using Argon2id
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLength,
	)
}

// encrypt encrypts data using AES-256-GCM with Argon2id key derivation
// Format: [salt(32)][nonce(12)][ciphertext][tag(16)]
func encrypt(plaintext []byte, password string) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Combine: salt + nonce + ciphertext (includes tag)
	result := make([]byte, 0, saltLength+nonceLength+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// decrypt decrypts data using AES-256-GCM with Argon2id key derivation
func decrypt(ciphertext []byte, password string) ([]byte, error) {
	// Verify minimum length: salt + nonce + tag
	minLength := saltLength + nonceLength + tagLength
	if len(ciphertext) < minLength {
		return nil, fmt.Errorf("ciphertext too short: expected at least %d bytes, got %d", minLength, len(ciphertext))
	}

	// Extract salt, nonce, and encrypted data
	salt := ciphertext[:saltLength]
	nonce := ciphertext[saltLength : saltLength+nonceLength]
	encryptedData := ciphertext[saltLength+nonceLength:]

	// Derive key from password
	key := deriveKey(password, salt)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (incorrect password or corrupted data): %w", err)
	}

	return plaintext, nil
}

// validatePassword checks if a password meets minimum requirements
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	return nil
}

// Create creates a new encrypted wallet with a generated Ed25519 keypair
func Create(name string, password string) (*Wallet, error) {
	// Validate password
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// Get wallet directory
	walletDir, err := getWalletDirFunc()
	if err != nil {
		return nil, err
	}

	// Create wallet directory if it doesn't exist
	if err := os.MkdirAll(walletDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}

	// Check if wallet already exists
	walletPath, err := GetWalletPath(name)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(walletPath); err == nil {
		return nil, fmt.Errorf("wallet '%s' already exists", name)
	}

	// Generate Ed25519 keypair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Create keypair struct
	var kp Keypair
	copy(kp.PublicKey[:], pubKey)
	copy(kp.PrivateKey[:], privKey)

	// Create wallet
	wallet := &Wallet{
		Name:     name,
		Keypairs: []Keypair{kp},
	}

	// Save wallet
	if err := saveWallet(wallet, password); err != nil {
		return nil, err
	}

	return wallet, nil
}

// Open opens an existing encrypted wallet
func Open(name string, password string) (*Wallet, error) {
	// Get wallet path
	walletPath, err := GetWalletPath(name)
	if err != nil {
		return nil, err
	}

	// Read encrypted wallet file
	encryptedData, err := os.ReadFile(walletPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("wallet '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	// Decrypt wallet data
	decryptedData, err := decrypt(encryptedData, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt wallet (incorrect password?): %w", err)
	}

	// Parse wallet JSON
	var wf walletFile
	if err := json.Unmarshal(decryptedData, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse wallet data: %w", err)
	}

	// Create wallet struct
	wallet := &Wallet{
		Name:     name,
		Keypairs: wf.Keypairs,
	}

	return wallet, nil
}

// List returns a list of all wallet names in the wallet directory
func List() ([]string, error) {
	// Get wallet directory
	walletDir, err := getWalletDirFunc()
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	if _, err := os.Stat(walletDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(walletDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet directory: %w", err)
	}

	// Collect wallet names
	var wallets []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".wallet") {
			// Remove .wallet extension
			walletName := strings.TrimSuffix(name, ".wallet")
			wallets = append(wallets, walletName)
		}
	}

	return wallets, nil
}

// Export exports the wallet keypairs to an unencrypted JSON file
func (w *Wallet) Export(outputPath string) error {
	// Create wallet file structure
	wf := walletFile{
		Version:  1,
		Keypairs: w.Keypairs,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(wf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// Import imports keypairs from an unencrypted JSON file and creates an encrypted wallet
func Import(inputPath string, name string, password string) (*Wallet, error) {
	// Validate password
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	// Parse JSON
	var wf walletFile
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse input file: %w", err)
	}

	// Validate keypairs
	if len(wf.Keypairs) == 0 {
		return nil, fmt.Errorf("input file contains no keypairs")
	}

	// Create wallet
	wallet := &Wallet{
		Name:     name,
		Keypairs: wf.Keypairs,
	}

	// Get wallet directory
	walletDir, err := getWalletDirFunc()
	if err != nil {
		return nil, err
	}

	// Create wallet directory if it doesn't exist
	if err := os.MkdirAll(walletDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}

	// Check if wallet already exists
	walletPath, err := GetWalletPath(name)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(walletPath); err == nil {
		return nil, fmt.Errorf("wallet '%s' already exists", name)
	}

	// Save wallet
	if err := saveWallet(wallet, password); err != nil {
		return nil, err
	}

	return wallet, nil
}

// saveWallet saves a wallet to disk with encryption
func saveWallet(wallet *Wallet, password string) error {
	// Create wallet file structure
	wf := walletFile{
		Version:  1,
		Keypairs: wallet.Keypairs,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(wf)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	// Encrypt wallet data
	encryptedData, err := encrypt(jsonData, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet: %w", err)
	}

	// Get wallet path
	walletPath, err := GetWalletPath(wallet.Name)
	if err != nil {
		return err
	}

	// Write encrypted data to file
	if err := os.WriteFile(walletPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

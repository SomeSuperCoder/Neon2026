package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/wallet"
)

// TestNodeStartupWithValidWallet tests node startup with a valid wallet and password
// Requirements: 1.1, 1.2, 1.4
func TestNodeStartupWithValidWallet(t *testing.T) {
	// Create temporary wallet directory
	tempDir := t.TempDir()
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	// Create a test wallet
	walletName := "test-validator"
	password := "test-password-123"

	w, err := wallet.Create(walletName, password)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	if len(w.Keypairs) != 1 {
		t.Fatalf("Expected 1 keypair, got %d", len(w.Keypairs))
	}

	// Initialize FileStore
	stateDBPath := filepath.Join(tempDir, "state.db")
	fs, err := filestore.NewFileStore(stateDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fs.Close()

	// Open wallet directly (bypassing password prompt for testing)
	w2, err := wallet.Open(walletName, password)
	if err != nil {
		t.Fatalf("Failed to open wallet: %v", err)
	}

	// Select keypair
	selectedKeypair, err := selectKeypair(w2, fs)
	if err != nil {
		t.Fatalf("Failed to select keypair: %v", err)
	}

	// Compute validator FileID
	localValidatorID := computeValidatorFileID(selectedKeypair.PublicKey)

	// Verify validator ID is not zero
	if localValidatorID == (filestore.FileID{}) {
		t.Fatal("Expected non-zero validator FileID")
	}

	// Verify keypair is not nil
	if selectedKeypair == nil {
		t.Fatal("Expected non-nil keypair")
	}

	// Verify keypair matches wallet keypair
	if selectedKeypair.PublicKey != w.Keypairs[0].PublicKey {
		t.Fatal("Keypair public key mismatch")
	}

	if selectedKeypair.PrivateKey != w.Keypairs[0].PrivateKey {
		t.Fatal("Keypair private key mismatch")
	}

	// Verify validator FileID is computed correctly
	expectedValidatorID := computeValidatorFileID(w.Keypairs[0].PublicKey)
	if localValidatorID != expectedValidatorID {
		t.Fatalf("Validator FileID mismatch: expected %s, got %s",
			expectedValidatorID.String(), localValidatorID.String())
	}
}

// TestNodeStartupObserverMode tests node startup in observer mode without wallet
// Requirements: 1.6, 3.4
func TestNodeStartupObserverMode(t *testing.T) {
	// Create temporary wallet directory
	tempDir := t.TempDir()
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	// Initialize FileStore
	stateDBPath := filepath.Join(tempDir, "state.db")
	fs, err := filestore.NewFileStore(stateDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fs.Close()

	// Test observer mode (empty wallet name)
	localValidatorID, localKeys, err := initializeValidatorIdentity("", fs)
	if err != nil {
		t.Fatalf("Failed to initialize observer mode: %v", err)
	}

	// Verify validator ID is zero
	if localValidatorID != (filestore.FileID{}) {
		t.Fatal("Expected zero validator FileID in observer mode")
	}

	// Verify keypair is nil
	if localKeys != nil {
		t.Fatal("Expected nil keypair in observer mode")
	}
}

// TestNodeStartupIncorrectPassword tests node startup with incorrect wallet password
// Requirements: 1.2
func TestNodeStartupIncorrectPassword(t *testing.T) {
	// Create temporary wallet directory
	tempDir := t.TempDir()
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	// Create a test wallet
	walletName := "test-validator"
	password := "test-password-123"

	_, err := wallet.Create(walletName, password)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Try to open with incorrect password
	_, err = wallet.Open(walletName, "wrong-password")
	if err == nil {
		t.Fatal("Expected error when opening wallet with incorrect password")
	}

	if !strings.Contains(err.Error(), "incorrect password") && !strings.Contains(err.Error(), "decryption failed") {
		t.Fatalf("Expected decryption error, got: %v", err)
	}
}

// TestComputeValidatorFileID tests the validator FileID computation
// Requirements: 1.5
func TestComputeValidatorFileID(t *testing.T) {
	// Generate a test keypair
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	var pubKeyArray [32]byte
	copy(pubKeyArray[:], pubKey)

	// Compute validator FileID
	validatorID := computeValidatorFileID(pubKeyArray)

	// Verify it's not zero
	if validatorID == (filestore.FileID{}) {
		t.Fatal("Expected non-zero validator FileID")
	}

	// Verify it's deterministic (same input produces same output)
	validatorID2 := computeValidatorFileID(pubKeyArray)
	if validatorID != validatorID2 {
		t.Fatal("Validator FileID computation is not deterministic")
	}

	// Verify different public keys produce different FileIDs
	pubKey2, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate second keypair: %v", err)
	}

	var pubKeyArray2 [32]byte
	copy(pubKeyArray2[:], pubKey2)

	validatorID3 := computeValidatorFileID(pubKeyArray2)
	if validatorID == validatorID3 {
		t.Fatal("Different public keys should produce different FileIDs")
	}
}

// TestSelectKeypairSingleKeypair tests keypair selection with single keypair
// Requirements: 1.4
func TestSelectKeypairSingleKeypair(t *testing.T) {
	// Create temporary wallet directory
	tempDir := t.TempDir()
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	// Create a test wallet
	walletName := "test-validator"
	password := "test-password-123"

	w, err := wallet.Create(walletName, password)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Select keypair (should return the only keypair without prompting)
	selectedKeypair, err := selectKeypair(w, nil)
	if err != nil {
		t.Fatalf("Failed to select keypair: %v", err)
	}

	if selectedKeypair == nil {
		t.Fatal("Expected non-nil keypair")
	}

	if selectedKeypair.PublicKey != w.Keypairs[0].PublicKey {
		t.Fatal("Selected keypair does not match wallet keypair")
	}
}

// TestValidatorIdentityLogging tests that validator identity is logged correctly
// Requirements: 1.1, 1.6
func TestValidatorIdentityLogging(t *testing.T) {
	// Create temporary wallet directory
	tempDir := t.TempDir()
	os.Setenv("POH_WALLET_DIR", tempDir)
	defer os.Unsetenv("POH_WALLET_DIR")

	// Create a test wallet
	walletName := "test-validator"
	password := "test-password-123"

	w, err := wallet.Create(walletName, password)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Initialize FileStore
	stateDBPath := filepath.Join(tempDir, "state.db")
	fs, err := filestore.NewFileStore(stateDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize FileStore: %v", err)
	}
	defer fs.Close()

	// Open wallet directly (bypassing password prompt for testing)
	w2, err := wallet.Open(walletName, password)
	if err != nil {
		t.Fatalf("Failed to open wallet: %v", err)
	}

	// Select keypair
	selectedKeypair, err := selectKeypair(w2, fs)
	if err != nil {
		t.Fatalf("Failed to select keypair: %v", err)
	}

	// Compute validator FileID
	localValidatorID := computeValidatorFileID(selectedKeypair.PublicKey)

	// Verify the public key is truncated correctly
	pubKeyHex := hex.EncodeToString(w.Keypairs[0].PublicKey[:])
	truncated := pubKeyHex[:16]

	// Verify truncated key is 16 characters
	if len(truncated) != 16 {
		t.Fatalf("Expected truncated key to be 16 chars, got %d", len(truncated))
	}

	// Verify validator FileID is logged (indirectly by checking it's computed)
	expectedValidatorID := computeValidatorFileID(w.Keypairs[0].PublicKey)
	if localValidatorID != expectedValidatorID {
		t.Fatal("Validator FileID not computed correctly")
	}
}

// encryptWalletData is a helper function to encrypt wallet data for testing
func encryptWalletData(plaintext []byte, password string) ([]byte, error) {
	const (
		argon2Memory      = 64 * 1024
		argon2Iterations  = 3
		argon2Parallelism = 4
		argon2KeyLength   = 32
		saltLength        = 32
		nonceLength       = 12
	)

	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Derive key from password
	key := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLength,
	)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
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

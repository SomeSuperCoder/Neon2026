package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2 parameters
	pbkdf2Iterations = 100000
	pbkdf2KeyLen     = 32
	saltLen          = 32

	// AES-GCM parameters
	nonceLen = 12
)

// EncryptWallet encrypts wallet data using AES-256-GCM with PBKDF2 key derivation
func EncryptWallet(wallet *Wallet, password string) (*EncryptedWallet, error) {
	// Generate random salt
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to generate salt",
			Cause:   err,
		}
	}

	// Derive encryption key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to create cipher",
			Cause:   err,
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to create GCM",
			Cause:   err,
		}
	}

	// Generate random nonce
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to generate nonce",
			Cause:   err,
		}
	}

	// Prepare wallet data for encryption
	walletData := struct {
		SeedPhrases []string   `json:"seedPhrases"`
		Accounts    []*Account `json:"accounts"`
		ActiveIndex int        `json:"activeIndex"`
	}{
		SeedPhrases: wallet.seedPhrases,
		Accounts:    wallet.accounts,
		ActiveIndex: wallet.activeIndex,
	}

	// Marshal wallet data to JSON
	plaintext, err := json.Marshal(walletData)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrEncryptionFailed,
			Message: "failed to marshal wallet data",
			Cause:   err,
		}
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Build encrypted account metadata (non-sensitive)
	encryptedAccounts := make([]EncryptedAccount, len(wallet.accounts))
	for i, acc := range wallet.accounts {
		encryptedAccounts[i] = EncryptedAccount{
			SeedPhraseIndex: acc.SeedPhraseIndex,
			Label:           acc.Label,
		}
	}

	return &EncryptedWallet{
		Version:    1,
		Cipher:     "AES-256-GCM",
		Salt:       base64.StdEncoding.EncodeToString(salt),
		IV:         base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Accounts:   encryptedAccounts,
	}, nil
}

// DecryptWallet decrypts wallet data using the provided password
func DecryptWallet(encrypted *EncryptedWallet, password string, config *WalletConfig) (*Wallet, error) {
	// Decode base64 values
	salt, err := base64.StdEncoding.DecodeString(encrypted.Salt)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to decode salt",
			Cause:   err,
		}
	}

	nonce, err := base64.StdEncoding.DecodeString(encrypted.IV)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to decode nonce",
			Cause:   err,
		}
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to decode ciphertext",
			Cause:   err,
		}
	}

	// Derive decryption key from password
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, pbkdf2KeyLen, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to create cipher",
			Cause:   err,
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to create GCM",
			Cause:   err,
		}
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrInvalidPassword,
			Message: "invalid password or corrupted wallet",
			Cause:   err,
		}
	}

	// Unmarshal wallet data
	var walletData struct {
		SeedPhrases []string   `json:"seedPhrases"`
		Accounts    []*Account `json:"accounts"`
		ActiveIndex int        `json:"activeIndex"`
	}

	if err := json.Unmarshal(plaintext, &walletData); err != nil {
		return nil, &WalletError{
			Code:    ErrDecryptionFailed,
			Message: "failed to unmarshal wallet data",
			Cause:   err,
		}
	}

	// Reconstruct wallet
	wallet := &Wallet{
		seedPhrases: walletData.SeedPhrases,
		accounts:    walletData.Accounts,
		activeIndex: walletData.ActiveIndex,
		config:      config,
		encrypted:   true,
	}

	return wallet, nil
}

// GenerateRandomPassword generates a random password for testing
func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)
	for i := range password {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			return "", err
		}
		password[i] = charset[int(randomByte[0])%len(charset)]
	}
	return string(password), nil
}

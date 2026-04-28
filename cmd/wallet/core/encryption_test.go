package core

import (
	"testing"
)

func TestEncryptDecryptWallet(t *testing.T) {
	// Create a test wallet
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import some seed phrases
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

	wallet.SetActiveAccount(2)

	// Encrypt
	password := "test-password-123"
	encrypted, err := EncryptWallet(wallet, password)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet: %v", err)
	}

	// Decrypt
	decrypted, err := DecryptWallet(encrypted, password, config)
	if err != nil {
		t.Fatalf("Failed to decrypt wallet: %v", err)
	}

	// Verify
	if len(decrypted.seedPhrases) != len(wallet.seedPhrases) {
		t.Errorf("Seed phrase count mismatch: expected %d, got %d", len(wallet.seedPhrases), len(decrypted.seedPhrases))
	}

	for i, sp := range wallet.seedPhrases {
		if decrypted.seedPhrases[i] != sp {
			t.Errorf("Seed phrase %d mismatch", i)
		}
	}

	if len(decrypted.accounts) != len(wallet.accounts) {
		t.Errorf("Account count mismatch: expected %d, got %d", len(wallet.accounts), len(decrypted.accounts))
	}

	if decrypted.activeIndex != wallet.activeIndex {
		t.Errorf("Active index mismatch: expected %d, got %d", wallet.activeIndex, decrypted.activeIndex)
	}

	// Verify each account
	for i, acc := range wallet.accounts {
		decAcc := decrypted.accounts[i]
		if acc.SeedPhraseIndex != decAcc.SeedPhraseIndex {
			t.Errorf("Account %d seed phrase index mismatch", i)
		}
		if acc.Address != decAcc.Address {
			t.Errorf("Account %d address mismatch", i)
		}
		if acc.Label != decAcc.Label {
			t.Errorf("Account %d label mismatch", i)
		}
	}
}

func TestEncryptDecryptWithDifferentPasswords(t *testing.T) {
	// Create a test wallet
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := DefaultConfig()
	wallet, err := NewWalletWithSeedPhrase(mnemonic, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Encrypt with one password
	password1 := "password-one"
	encrypted1, err := EncryptWallet(wallet, password1)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet: %v", err)
	}

	// Encrypt with different password
	password2 := "password-two"
	encrypted2, err := EncryptWallet(wallet, password2)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet: %v", err)
	}

	// Ciphertexts should be different
	if encrypted1.Ciphertext == encrypted2.Ciphertext {
		t.Error("Ciphertexts should be different with different passwords")
	}

	// Decrypt with correct passwords
	decrypted1, err := DecryptWallet(encrypted1, password1, config)
	if err != nil {
		t.Fatalf("Failed to decrypt with password1: %v", err)
	}

	decrypted2, err := DecryptWallet(encrypted2, password2, config)
	if err != nil {
		t.Fatalf("Failed to decrypt with password2: %v", err)
	}

	// Both should decrypt to same wallet
	if len(decrypted1.seedPhrases) != len(decrypted2.seedPhrases) {
		t.Error("Decrypted wallets have different number of seed phrases")
	}

	for i := range decrypted1.seedPhrases {
		if decrypted1.seedPhrases[i] != decrypted2.seedPhrases[i] {
			t.Error("Decrypted wallets have different seed phrases")
		}
	}

	// Try to decrypt with wrong password
	_, err = DecryptWallet(encrypted1, password2, config)
	if err == nil {
		t.Error("Expected error when decrypting with wrong password")
	}
}

func TestEncryptionRandomness(t *testing.T) {
	// Create a test wallet
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := DefaultConfig()
	wallet, err := NewWalletWithSeedPhrase(mnemonic, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	password := "same-password"

	// Encrypt twice with same password
	encrypted1, err := EncryptWallet(wallet, password)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet 1: %v", err)
	}

	encrypted2, err := EncryptWallet(wallet, password)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet 2: %v", err)
	}

	// Salt and IV should be different (random)
	if encrypted1.Salt == encrypted2.Salt {
		t.Error("Salts should be different (random)")
	}

	if encrypted1.IV == encrypted2.IV {
		t.Error("IVs should be different (random)")
	}

	// Ciphertexts should be different due to different IVs
	if encrypted1.Ciphertext == encrypted2.Ciphertext {
		t.Error("Ciphertexts should be different with different IVs")
	}

	// Both should decrypt correctly
	decrypted1, err := DecryptWallet(encrypted1, password, config)
	if err != nil {
		t.Fatalf("Failed to decrypt wallet 1: %v", err)
	}

	decrypted2, err := DecryptWallet(encrypted2, password, config)
	if err != nil {
		t.Fatalf("Failed to decrypt wallet 2: %v", err)
	}

	// Both should have same seed phrases
	if len(decrypted1.seedPhrases) != len(decrypted2.seedPhrases) {
		t.Error("Decrypted wallets have different number of seed phrases")
	}

	for i := range decrypted1.seedPhrases {
		if decrypted1.seedPhrases[i] != decrypted2.seedPhrases[i] {
			t.Error("Decrypted wallets have different seed phrases")
		}
	}
}

func TestDecryptWithInvalidData(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name      string
		encrypted *EncryptedWallet
		password  string
	}{
		{
			name: "invalid base64 salt",
			encrypted: &EncryptedWallet{
				Version:    1,
				Cipher:     "AES-256-GCM",
				Salt:       "invalid!!!base64",
				IV:         "dGVzdA==",
				Ciphertext: "dGVzdA==",
			},
			password: "password",
		},
		{
			name: "invalid base64 IV",
			encrypted: &EncryptedWallet{
				Version:    1,
				Cipher:     "AES-256-GCM",
				Salt:       "dGVzdA==",
				IV:         "invalid!!!base64",
				Ciphertext: "dGVzdA==",
			},
			password: "password",
		},
		{
			name: "invalid base64 ciphertext",
			encrypted: &EncryptedWallet{
				Version:    1,
				Cipher:     "AES-256-GCM",
				Salt:       "dGVzdA==",
				IV:         "dGVzdA==",
				Ciphertext: "invalid!!!base64",
			},
			password: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptWallet(tt.encrypted, tt.password, config)
			if err == nil {
				t.Error("Expected error for invalid data")
			}
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	lengths := []int{8, 16, 32, 64}

	for _, length := range lengths {
		password, err := GenerateRandomPassword(length)
		if err != nil {
			t.Fatalf("Failed to generate password of length %d: %v", length, err)
		}

		if len(password) != length {
			t.Errorf("Expected password length %d, got %d", length, len(password))
		}
	}

	// Generate multiple passwords and verify they're different
	passwords := make(map[string]bool)
	for i := 0; i < 10; i++ {
		password, err := GenerateRandomPassword(16)
		if err != nil {
			t.Fatalf("Failed to generate password: %v", err)
		}

		if passwords[password] {
			t.Error("Generated duplicate password")
		}
		passwords[password] = true
	}
}

func TestWalletEncryption(t *testing.T) {
	config := DefaultConfig()
	wallet, err := NewWallet(config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Import some seed phrases
	for i := 0; i < 3; i++ {
		mnemonic, err := GenerateSeedPhrase(12)
		if err != nil {
			t.Fatalf("Failed to generate seed phrase: %v", err)
		}

		_, err = wallet.ImportSeedPhrase(mnemonic, "Account "+string(rune(i+1)))
		if err != nil {
			t.Fatalf("Failed to import seed phrase: %v", err)
		}
	}

	// Encrypt wallet
	password := "test-password"
	encrypted, err := EncryptWallet(wallet, password)
	if err != nil {
		t.Fatalf("Failed to encrypt wallet: %v", err)
	}

	// Verify encrypted structure
	if encrypted.Cipher != "AES-256-GCM" {
		t.Errorf("Expected cipher AES-256-GCM, got %s", encrypted.Cipher)
	}

	if encrypted.Salt == "" {
		t.Error("Salt is empty")
	}

	if encrypted.IV == "" {
		t.Error("IV is empty")
	}

	if encrypted.Ciphertext == "" {
		t.Error("Ciphertext is empty")
	}

	if len(encrypted.Accounts) != len(wallet.accounts) {
		t.Errorf("Expected %d encrypted accounts, got %d", len(wallet.accounts), len(encrypted.Accounts))
	}

	// Decrypt wallet
	decrypted, err := DecryptWallet(encrypted, password, config)
	if err != nil {
		t.Fatalf("Failed to decrypt wallet: %v", err)
	}

	// Verify decrypted wallet
	if len(decrypted.seedPhrases) != len(wallet.seedPhrases) {
		t.Errorf("Expected %d seed phrases, got %d", len(wallet.seedPhrases), len(decrypted.seedPhrases))
	}

	if len(decrypted.accounts) != len(wallet.accounts) {
		t.Errorf("Expected %d accounts, got %d", len(wallet.accounts), len(decrypted.accounts))
	}
}

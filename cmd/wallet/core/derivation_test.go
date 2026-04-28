package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	// Generate a test seed phrase
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("Failed to convert mnemonic to seed: %v", err)
	}

	// Derive master key
	masterKey, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive key: %v", err)
	}

	// Verify key and chain code are 32 bytes
	if len(masterKey.Key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(masterKey.Key))
	}

	if len(masterKey.ChainCode) != 32 {
		t.Errorf("Expected chain code length 32, got %d", len(masterKey.ChainCode))
	}
}

func TestDeriveAccount(t *testing.T) {
	// Generate a test seed
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("Failed to convert mnemonic to seed: %v", err)
	}

	// Derive master key
	masterKey, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive master key: %v", err)
	}

	// Derive multiple accounts
	accounts := make([]*HDKey, 5)
	for i := uint32(0); i < 5; i++ {
		account, err := masterKey.DeriveAccount(i)
		if err != nil {
			t.Fatalf("Failed to derive account %d: %v", i, err)
		}
		accounts[i] = account
	}

	// Verify all accounts are different
	for i := 0; i < len(accounts); i++ {
		for j := i + 1; j < len(accounts); j++ {
			if hex.EncodeToString(accounts[i].Key) == hex.EncodeToString(accounts[j].Key) {
				t.Errorf("Accounts %d and %d have the same key", i, j)
			}
		}
	}
}

func TestToEd25519(t *testing.T) {
	// Generate a test seed
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("Failed to convert mnemonic to seed: %v", err)
	}

	// Derive master key
	masterKey, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive master key: %v", err)
	}

	// Derive account
	account, err := masterKey.DeriveAccount(0)
	if err != nil {
		t.Fatalf("Failed to derive account: %v", err)
	}

	// Convert to Ed25519
	publicKey, privateKey := account.ToEd25519()

	// Verify key lengths
	if len(publicKey) != ed25519.PublicKeySize {
		t.Errorf("Expected public key length %d, got %d", ed25519.PublicKeySize, len(publicKey))
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key length %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	// Verify the public key matches the private key
	derivedPublicKey := privateKey.Public().(ed25519.PublicKey)
	if hex.EncodeToString(publicKey) != hex.EncodeToString(derivedPublicKey) {
		t.Error("Public key does not match private key")
	}
}

func TestDeterministicDerivation(t *testing.T) {
	// Use a fixed mnemonic for deterministic testing
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("Failed to convert mnemonic to seed: %v", err)
	}

	// Derive the same account twice
	masterKey1, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive master key 1: %v", err)
	}

	account1, err := masterKey1.DeriveAccount(0)
	if err != nil {
		t.Fatalf("Failed to derive account 1: %v", err)
	}

	masterKey2, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive master key 2: %v", err)
	}

	account2, err := masterKey2.DeriveAccount(0)
	if err != nil {
		t.Fatalf("Failed to derive account 2: %v", err)
	}

	// Keys should be identical
	if hex.EncodeToString(account1.Key) != hex.EncodeToString(account2.Key) {
		t.Error("Derived keys are not deterministic")
	}

	if hex.EncodeToString(account1.ChainCode) != hex.EncodeToString(account2.ChainCode) {
		t.Error("Derived chain codes are not deterministic")
	}
}

func TestHardenedIndex(t *testing.T) {
	tests := []struct {
		name  string
		index uint32
		want  uint32
	}{
		{
			name:  "index 0",
			index: 0,
			want:  0x80000000,
		},
		{
			name:  "index 44",
			index: 44,
			want:  0x8000002C,
		},
		{
			name:  "index 501",
			index: 501,
			want:  0x800001F5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hardenedIndex(tt.index); got != tt.want {
				t.Errorf("hardenedIndex() = 0x%X, want 0x%X", got, tt.want)
			}
		})
	}
}

func TestDerivationPath(t *testing.T) {
	path := DefaultDerivationPath()

	expected := "m/44'/501'/0'/0'/0'"
	if got := path.String(); got != expected {
		t.Errorf("DerivationPath.String() = %s, want %s", got, expected)
	}
}

func TestMultipleAccountDerivation(t *testing.T) {
	// Test deriving 100 accounts (requirement: support minimum 100 accounts)
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	seed, err := MnemonicToSeed(mnemonic, "")
	if err != nil {
		t.Fatalf("Failed to convert mnemonic to seed: %v", err)
	}

	masterKey, err := DeriveKey(seed, "m/44'/501'/0'/0'")
	if err != nil {
		t.Fatalf("Failed to derive master key: %v", err)
	}

	// Derive 100 accounts
	publicKeys := make(map[string]bool)
	for i := uint32(0); i < 100; i++ {
		account, err := masterKey.DeriveAccount(i)
		if err != nil {
			t.Fatalf("Failed to derive account %d: %v", i, err)
		}

		publicKey, _ := account.ToEd25519()
		pubKeyHex := hex.EncodeToString(publicKey)

		// Verify uniqueness
		if publicKeys[pubKeyHex] {
			t.Errorf("Duplicate public key found at index %d", i)
		}
		publicKeys[pubKeyHex] = true
	}

	// Verify we have 100 unique keys
	if len(publicKeys) != 100 {
		t.Errorf("Expected 100 unique keys, got %d", len(publicKeys))
	}
}

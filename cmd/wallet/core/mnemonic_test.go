package core

import (
	"strings"
	"testing"
)

func TestGenerateSeedPhrase(t *testing.T) {
	tests := []struct {
		name      string
		wordCount int
		wantErr   bool
	}{
		{
			name:      "12 word seed phrase",
			wordCount: 12,
			wantErr:   false,
		},
		{
			name:      "24 word seed phrase",
			wordCount: 24,
			wantErr:   false,
		},
		{
			name:      "invalid word count",
			wordCount: 18,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mnemonic, err := GenerateSeedPhrase(tt.wordCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSeedPhrase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the mnemonic is valid
				if !ValidateSeedPhrase(mnemonic) {
					t.Errorf("Generated mnemonic is invalid: %s", mnemonic)
				}

				// Verify word count
				words := strings.Fields(mnemonic)
				if len(words) != tt.wordCount {
					t.Errorf("Expected %d words, got %d", tt.wordCount, len(words))
				}
			}
		})
	}
}

func TestValidateSeedPhrase(t *testing.T) {
	// Generate a valid seed phrase
	validMnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	tests := []struct {
		name     string
		mnemonic string
		want     bool
	}{
		{
			name:     "valid mnemonic",
			mnemonic: validMnemonic,
			want:     true,
		},
		{
			name:     "invalid mnemonic - wrong words",
			mnemonic: "invalid test phrase with random words that are not valid",
			want:     false,
		},
		{
			name:     "empty mnemonic",
			mnemonic: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateSeedPhrase(tt.mnemonic); got != tt.want {
				t.Errorf("ValidateSeedPhrase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMnemonicToSeed(t *testing.T) {
	// Generate a valid mnemonic
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	tests := []struct {
		name     string
		mnemonic string
		password string
		wantErr  bool
	}{
		{
			name:     "valid mnemonic without password",
			mnemonic: mnemonic,
			password: "",
			wantErr:  false,
		},
		{
			name:     "valid mnemonic with password",
			mnemonic: mnemonic,
			password: "test-password",
			wantErr:  false,
		},
		{
			name:     "invalid mnemonic",
			mnemonic: "invalid test phrase",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed, err := MnemonicToSeed(tt.mnemonic, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("MnemonicToSeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify seed is 64 bytes (512 bits)
				if len(seed) != 64 {
					t.Errorf("Expected seed length 64, got %d", len(seed))
				}
			}
		})
	}
}

func TestMnemonicToSeedDeterministic(t *testing.T) {
	mnemonic, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	// Generate seed twice with same mnemonic and password
	seed1, err := MnemonicToSeed(mnemonic, "password123")
	if err != nil {
		t.Fatalf("Failed to generate seed: %v", err)
	}

	seed2, err := MnemonicToSeed(mnemonic, "password123")
	if err != nil {
		t.Fatalf("Failed to generate seed: %v", err)
	}

	// Seeds should be identical
	if len(seed1) != len(seed2) {
		t.Errorf("Seed lengths differ: %d vs %d", len(seed1), len(seed2))
	}

	for i := range seed1 {
		if seed1[i] != seed2[i] {
			t.Errorf("Seeds differ at byte %d", i)
			break
		}
	}
}

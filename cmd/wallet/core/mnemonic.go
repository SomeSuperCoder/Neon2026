package core

import (
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// GenerateSeedPhrase generates a BIP39-compliant mnemonic seed phrase
// wordCount must be 12 or 24
func GenerateSeedPhrase(wordCount int) (string, error) {
	var entropyBits int
	switch wordCount {
	case 12:
		entropyBits = 128
	case 24:
		entropyBits = 256
	default:
		return "", &WalletError{
			Code:    ErrInvalidSeedPhrase,
			Message: fmt.Sprintf("invalid word count: %d (must be 12 or 24)", wordCount),
		}
	}

	entropy, err := bip39.NewEntropy(entropyBits)
	if err != nil {
		return "", &WalletError{
			Code:    ErrInvalidSeedPhrase,
			Message: "failed to generate entropy",
			Cause:   err,
		}
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", &WalletError{
			Code:    ErrInvalidSeedPhrase,
			Message: "failed to generate mnemonic",
			Cause:   err,
		}
	}

	return mnemonic, nil
}

// ValidateSeedPhrase validates a BIP39 mnemonic seed phrase
func ValidateSeedPhrase(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// MnemonicToSeed converts a BIP39 mnemonic to a seed
// The password parameter is optional (use empty string if not needed)
func MnemonicToSeed(mnemonic string, password string) ([]byte, error) {
	if !ValidateSeedPhrase(mnemonic) {
		return nil, &WalletError{
			Code:    ErrInvalidSeedPhrase,
			Message: "invalid seed phrase",
		}
	}

	seed := bip39.NewSeed(mnemonic, password)
	return seed, nil
}

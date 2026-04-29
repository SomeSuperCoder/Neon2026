package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/quanticscript"
	"github.com/poh-blockchain/internal/wallet"
	"golang.org/x/term"
)

// computeValidatorFileID computes the validator FileID from a public key
// Uses SHA-256("validator:" || pubkey) as per requirements 1.5
func computeValidatorFileID(pubkey [32]byte) filestore.FileID {
	h := sha256.New()
	h.Write([]byte("validator:"))
	h.Write(pubkey[:])
	var fileID filestore.FileID
	copy(fileID[:], h.Sum(nil))
	return fileID
}

// selectKeypair prompts the user to select a keypair from a wallet
// Displays each keypair's public key (truncated) and validator status
// Requirements: 1.4, 1.5
func selectKeypair(w *wallet.Wallet, fs *filestore.FileStore) (*wallet.Keypair, error) {
	if len(w.Keypairs) == 0 {
		return nil, fmt.Errorf("wallet contains no keypairs")
	}

	// If only one keypair, return it directly
	if len(w.Keypairs) == 1 {
		return &w.Keypairs[0], nil
	}

	// Multiple keypairs - prompt user to select
	fmt.Println("\nSelect a keypair:")

	for i, kp := range w.Keypairs {
		pubkeyHex := hex.EncodeToString(kp.PublicKey[:])
		truncated := pubkeyHex[:16]

		// Check if validator is registered
		validatorID := computeValidatorFileID(kp.PublicKey)
		status := "not registered"

		if fs != nil {
			validatorFile, err := fs.GetFile(validatorID)
			if err == nil && validatorFile != nil {
				// Try to deserialize as validator record
				_, _, _, statusByte, _, _, _, err := quanticscript.DeserializeValidatorRecord(validatorFile.Data)
				if err == nil {
					if statusByte == 1 {
						status = "active validator"
					} else {
						status = "inactive validator"
					}
				}
			}
		}

		fmt.Printf("  %d. %s... (%s)\n", i+1, truncated, status)
	}

	// Read user selection
	fmt.Print("\nEnter selection (1-" + strconv.Itoa(len(w.Keypairs)) + "): ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return nil, fmt.Errorf("failed to read selection: %w", err)
	}

	selection, err := strconv.Atoi(input)
	if err != nil || selection < 1 || selection > len(w.Keypairs) {
		return nil, fmt.Errorf("invalid selection: must be between 1 and %d", len(w.Keypairs))
	}

	return &w.Keypairs[selection-1], nil
}

// promptWalletPassword prompts for wallet password using secure terminal input
// Requirements: 1.2
func promptWalletPassword(prompt string) (string, error) {
	fmt.Print(prompt)

	// Check if stdin is a terminal
	if term.IsTerminal(int(os.Stdin.Fd())) {
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // New line after password input
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return string(passwordBytes), nil
	}

	// For non-terminal input (testing), read from stdin
	var password string
	_, err := fmt.Scanln(&password)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return password, nil
}

// openWalletWithPassword opens a wallet and prompts for password if needed
// Requirements: 1.2, 1.3
func openWalletWithPassword(walletName string) (*wallet.Wallet, error) {
	password, err := promptWalletPassword("Enter wallet password: ")
	if err != nil {
		return nil, err
	}

	w, err := wallet.Open(walletName, password)
	if err != nil {
		return nil, fmt.Errorf("failed to open wallet: %w", err)
	}

	return w, nil
}

// initializeValidatorIdentity initializes the local validator identity from a wallet
// Returns the validator FileID and keypair, or zero FileID and nil keypair for observer mode
// Requirements: 1.1, 1.2, 1.4, 1.5, 1.6
func initializeValidatorIdentity(walletName string, fs *filestore.FileStore) (filestore.FileID, *wallet.Keypair, error) {
	// Observer mode: no wallet specified
	if walletName == "" {
		log.Println("Starting node in observer mode (no wallet specified)")
		return filestore.FileID{}, nil, nil
	}

	// Open wallet with password
	w, err := openWalletWithPassword(walletName)
	if err != nil {
		return filestore.FileID{}, nil, err
	}

	// Select keypair if multiple exist
	selectedKeypair, err := selectKeypair(w, fs)
	if err != nil {
		return filestore.FileID{}, nil, err
	}

	// Compute validator FileID from public key
	validatorID := computeValidatorFileID(selectedKeypair.PublicKey)

	// Log validator identity
	pubkeyHex := hex.EncodeToString(selectedKeypair.PublicKey[:])
	truncated := pubkeyHex[:16]
	log.Printf("Starting node as validator: %s (FileID: %s)", truncated, validatorID.String())

	return validatorID, selectedKeypair, nil
}

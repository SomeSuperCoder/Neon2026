package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/poh-blockchain/internal/wallet"
	"golang.org/x/term"
)

// handleWalletCommand handles wallet management commands
func handleWalletCommand() {
	if len(os.Args) < 3 {
		printWalletHelp()
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "create":
		handleWalletCreateCommand()
	case "list":
		handleWalletListCommand()
	case "show":
		handleWalletShowCommand()
	case "export":
		handleWalletExportCommand()
	case "import":
		handleWalletImportCommand()
	case "help":
		printWalletHelp()
	default:
		fmt.Printf("Unknown wallet subcommand: %s\n", subcommand)
		printWalletHelp()
		os.Exit(1)
	}
}

// printWalletHelp prints help for wallet commands
func printWalletHelp() {
	fmt.Println("Wallet Management")
	fmt.Println("\nUsage:")
	fmt.Println("  poh-blockchain wallet <subcommand> [options]")
	fmt.Println("\nSubcommands:")
	fmt.Println("  create --name <wallet-name>")
	fmt.Println("    Create a new password-protected wallet")
	fmt.Println("    Prompts for password (twice for confirmation)")
	fmt.Println()
	fmt.Println("  list")
	fmt.Println("    List all available wallets")
	fmt.Println()
	fmt.Println("  show --name <wallet-name>")
	fmt.Println("    Display wallet information and public keys")
	fmt.Println("    Prompts for password")
	fmt.Println()
	fmt.Println("  export --name <wallet-name> --output <file>")
	fmt.Println("    Export wallet keypairs as unencrypted JSON (for backup)")
	fmt.Println("    Prompts for password")
	fmt.Println()
	fmt.Println("  import --input <file> --name <wallet-name>")
	fmt.Println("    Import keypairs from unencrypted JSON into encrypted wallet")
	fmt.Println("    Prompts for new password")
	fmt.Println()
	fmt.Println("  help")
	fmt.Println("    Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create a new wallet")
	fmt.Println("  poh-blockchain wallet create --name my-validator")
	fmt.Println()
	fmt.Println("  # List all wallets")
	fmt.Println("  poh-blockchain wallet list")
	fmt.Println()
	fmt.Println("  # Show wallet details")
	fmt.Println("  poh-blockchain wallet show --name my-validator")
	fmt.Println()
	fmt.Println("  # Export wallet for backup")
	fmt.Println("  poh-blockchain wallet export --name my-validator --output backup.json")
	fmt.Println()
	fmt.Println("  # Import wallet from backup")
	fmt.Println("  poh-blockchain wallet import --input backup.json --name restored-wallet")
}

// handleWalletCreateCommand creates a new wallet
func handleWalletCreateCommand() {
	fs := flag.NewFlagSet("wallet create", flag.ExitOnError)
	name := fs.String("name", "", "Wallet name")

	fs.Parse(os.Args[3:])

	if *name == "" {
		fmt.Println("Error: --name is required")
		fmt.Println()
		printWalletHelp()
		os.Exit(1)
	}

	// Prompt for password
	password, err := promptPasswordCreate(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create wallet
	err = runWalletCreate(*name, strings.NewReader(password+"\n"+password+"\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating wallet: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wallet '%s' created successfully!\n", *name)

	// Open wallet to display public key
	w, err := wallet.Open(*name, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to open wallet: %v\n", err)
		return
	}

	if len(w.Keypairs) > 0 {
		pubKeyHex := hex.EncodeToString(w.Keypairs[0].PublicKey[:])
		fmt.Printf("Public key: %s\n", pubKeyHex)
	}
}

// handleWalletListCommand lists all wallets
func handleWalletListCommand() {
	err := runWalletList(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing wallets: %v\n", err)
		os.Exit(1)
	}
}

// handleWalletShowCommand shows wallet details
func handleWalletShowCommand() {
	fs := flag.NewFlagSet("wallet show", flag.ExitOnError)
	name := fs.String("name", "", "Wallet name")

	fs.Parse(os.Args[3:])

	if *name == "" {
		fmt.Println("Error: --name is required")
		fmt.Println()
		printWalletHelp()
		os.Exit(1)
	}

	// Prompt for password
	password, err := promptPassword("Enter wallet password: ", os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = runWalletShow(*name, password, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error showing wallet: %v\n", err)
		os.Exit(1)
	}
}

// handleWalletExportCommand exports wallet to unencrypted JSON
func handleWalletExportCommand() {
	fs := flag.NewFlagSet("wallet export", flag.ExitOnError)
	name := fs.String("name", "", "Wallet name")
	output := fs.String("output", "", "Output file path")

	fs.Parse(os.Args[3:])

	if *name == "" || *output == "" {
		fmt.Println("Error: --name and --output are required")
		fmt.Println()
		printWalletHelp()
		os.Exit(1)
	}

	// Prompt for password
	password, err := promptPassword("Enter wallet password: ", os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = runWalletExport(*name, password, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error exporting wallet: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wallet exported to: %s\n", *output)
	fmt.Println("WARNING: This file contains unencrypted private keys. Keep it secure!")
}

// handleWalletImportCommand imports wallet from unencrypted JSON
func handleWalletImportCommand() {
	fs := flag.NewFlagSet("wallet import", flag.ExitOnError)
	input := fs.String("input", "", "Input file path")
	name := fs.String("name", "", "Wallet name")

	fs.Parse(os.Args[3:])

	if *input == "" || *name == "" {
		fmt.Println("Error: --input and --name are required")
		fmt.Println()
		printWalletHelp()
		os.Exit(1)
	}

	// Prompt for new password
	password, err := promptPasswordCreate(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = runWalletImport(*input, *name, strings.NewReader(password+"\n"+password+"\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing wallet: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Wallet '%s' imported successfully!\n", *name)
}

// runWalletCreate creates a new wallet (testable function)
func runWalletCreate(name string, passwordReader io.Reader) error {
	// Read password twice
	password, confirmPassword, err := readPasswordTwice(passwordReader)
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Check if wallet already exists
	walletPath, err := wallet.GetWalletPath(name)
	if err != nil {
		return fmt.Errorf("failed to get wallet path: %w", err)
	}

	if _, err := os.Stat(walletPath); err == nil {
		return fmt.Errorf("wallet '%s' already exists", name)
	}

	// Create wallet
	_, err = wallet.Create(name, password)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

// runWalletList lists all wallets (testable function)
func runWalletList(output io.Writer) error {
	wallets, err := wallet.List()
	if err != nil {
		return fmt.Errorf("failed to list wallets: %w", err)
	}

	if len(wallets) == 0 {
		fmt.Fprintln(output, "No wallets found")
		return nil
	}

	fmt.Fprintln(output, "Available wallets:")
	for _, name := range wallets {
		fmt.Fprintf(output, "  - %s\n", name)
	}

	return nil
}

// runWalletShow shows wallet details (testable function)
func runWalletShow(name, password string, output io.Writer) error {
	// Open wallet
	w, err := wallet.Open(name, password)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("wallet '%s' not found", name)
		}
		return fmt.Errorf("failed to open wallet (incorrect password?): %w", err)
	}

	fmt.Fprintf(output, "Wallet: %s\n", name)
	fmt.Fprintf(output, "Keypairs: %d\n", len(w.Keypairs))
	fmt.Fprintln(output)

	for i, kp := range w.Keypairs {
		pubKeyHex := hex.EncodeToString(kp.PublicKey[:])
		// Truncate to 16 hex chars as per requirements
		truncated := pubKeyHex[:16]
		fmt.Fprintf(output, "  %d. %s...\n", i+1, truncated)
	}

	return nil
}

// runWalletExport exports wallet to unencrypted JSON (testable function)
func runWalletExport(name, password, outputPath string) error {
	// Open wallet
	w, err := wallet.Open(name, password)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("wallet '%s' not found", name)
		}
		return fmt.Errorf("failed to open wallet (incorrect password?): %w", err)
	}

	// Export wallet
	err = w.Export(outputPath)
	if err != nil {
		return fmt.Errorf("failed to export wallet: %w", err)
	}

	return nil
}

// runWalletImport imports wallet from unencrypted JSON (testable function)
func runWalletImport(inputPath, name string, passwordReader io.Reader) error {
	// Read password twice
	password, confirmPassword, err := readPasswordTwice(passwordReader)
	if err != nil {
		return err
	}

	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	// Check if wallet already exists
	walletPath, err := wallet.GetWalletPath(name)
	if err != nil {
		return fmt.Errorf("failed to get wallet path: %w", err)
	}

	if _, err := os.Stat(walletPath); err == nil {
		return fmt.Errorf("wallet '%s' already exists", name)
	}

	// Import wallet
	_, err = wallet.Import(inputPath, name, password)
	if err != nil {
		return fmt.Errorf("failed to import wallet: %w", err)
	}

	return nil
}

// promptPassword prompts for a password (with terminal support)
func promptPassword(prompt string, reader io.Reader) (string, error) {
	// Check if stdin is a terminal
	if file, ok := reader.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		fmt.Print(prompt)
		passwordBytes, err := term.ReadPassword(int(file.Fd()))
		fmt.Println() // New line after password input
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return string(passwordBytes), nil
	}

	// For non-terminal input (testing), read from reader
	var password string
	_, err := fmt.Fscanln(reader, &password)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return password, nil
}

// promptPasswordCreate prompts for password creation (twice for confirmation)
func promptPasswordCreate(reader io.Reader) (string, error) {
	password, err := promptPassword("Enter password: ", reader)
	if err != nil {
		return "", err
	}

	confirmPassword, err := promptPassword("Confirm password: ", reader)
	if err != nil {
		return "", err
	}

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

// readPasswordTwice reads password twice from reader (for testing)
func readPasswordTwice(reader io.Reader) (string, string, error) {
	var password, confirmPassword string

	_, err := fmt.Fscanln(reader, &password)
	if err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}

	_, err = fmt.Fscanln(reader, &confirmPassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to read confirmation password: %w", err)
	}

	return password, confirmPassword, nil
}

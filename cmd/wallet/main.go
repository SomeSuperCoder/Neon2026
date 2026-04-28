package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/poh-blockchain/cmd/wallet/core"
	walletrpc "github.com/poh-blockchain/cmd/wallet/rpc"
	"github.com/poh-blockchain/cmd/wallet/ui"
)

const (
	defaultWalletPath = "~/.poh-wallet/wallet.dat"
	defaultRPCURL     = "http://localhost:8899"
)

func main() {
	// Parse command-line flags
	walletPath := flag.String("wallet-path", defaultWalletPath, "Path to wallet file")
	rpcURL := flag.String("rpc-url", defaultRPCURL, "RPC endpoint URL")
	flag.Parse()

	// Expand home directory in wallet path
	expandedPath, err := expandPath(*walletPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to expand wallet path: %v\n", err)
		os.Exit(1)
	}

	// Check if wallet exists
	walletExists := fileExists(expandedPath)

	var wallet *core.Wallet
	var password string

	if !walletExists {
		// Wallet doesn't exist - run initialization wizard
		fmt.Println("No wallet found. Starting wallet creation wizard...")
		wallet, password, err = runWizard(expandedPath, *rpcURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: wallet initialization failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Wallet exists - prompt for password and load
		fmt.Print("Enter wallet password: ")
		password, err = readPassword()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to read password: %v\n", err)
			os.Exit(1)
		}

		wallet, err = core.LoadWallet(expandedPath, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load wallet: %v\n", err)
			os.Exit(1)
		}
	}

	// Create RPC client
	rpcClient := walletrpc.NewRPCClient(*rpcURL)

	// Initialize Bubble Tea application
	model := ui.NewModel(wallet, rpcClient)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Start TUI event loop
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// expandPath expands ~ to home directory
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if len(path) == 1 {
		return homeDir, nil
	}

	return filepath.Join(homeDir, path[1:]), nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	// For now, use a simple implementation
	// In production, use golang.org/x/term for proper password input
	var password string
	_, err := fmt.Scanln(&password)
	return password, err
}

// runWizard runs the wallet creation wizard
func runWizard(walletPath string, rpcURL string) (*core.Wallet, string, error) {
	// Create wallet config
	config := core.DefaultConfig()
	config.WalletPath = walletPath
	config.RPCEndpoint = rpcURL

	// Create empty wallet
	wallet, err := core.NewWallet(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create wizard view
	wizardView := ui.NewWizardView(wallet)

	// Create wrapper model
	wizardModel := &wizardModelWrapper{view: wizardView}

	// Run wizard in Bubble Tea
	p := tea.NewProgram(wizardModel, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, "", fmt.Errorf("wizard failed: %w", err)
	}

	// Extract result from wizard
	wizardWrapper, ok := finalModel.(*wizardModelWrapper)
	if !ok {
		return nil, "", fmt.Errorf("unexpected model type")
	}

	if wizardWrapper.view.GetResult() == nil {
		return nil, "", fmt.Errorf("wizard did not complete")
	}

	result := wizardWrapper.view.GetResult()
	if !result.Success {
		return nil, "", fmt.Errorf("wizard failed: %s", result.Error)
	}

	// Reload wallet to ensure it's properly initialized
	wallet, err = core.LoadWallet(walletPath, result.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to reload wallet: %w", err)
	}

	return wallet, result.Password, nil
}

// wizardModelWrapper wraps WizardView to implement tea.Model
type wizardModelWrapper struct {
	view *ui.WizardView
}

func (w *wizardModelWrapper) Init() tea.Cmd {
	return w.view.Init()
}

func (w *wizardModelWrapper) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	view, cmd := w.view.Update(msg)
	w.view = view.(*ui.WizardView)
	return w, cmd
}

func (w *wizardModelWrapper) View() string {
	return w.view.View()
}

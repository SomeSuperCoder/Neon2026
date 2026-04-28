package core

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/transaction"
)

func TestBuildTransferTransaction(t *testing.T) {
	// Create a test wallet with a seed phrase
	seedPhrase, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  "/tmp/test-wallet.dat",
	}

	wallet, err := NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	// Get active account
	account := wallet.GetActiveAccount()
	if account == nil {
		t.Fatal("No active account")
	}

	// Create a recipient address
	recipientKey := make([]byte, 32)
	for i := range recipientKey {
		recipientKey[i] = byte(i + 1)
	}
	recipientAddress := hex.EncodeToString(recipientKey)

	// Test valid transfer
	t.Run("ValidTransfer", func(t *testing.T) {
		req := &TransferRequest{
			From:   account.Address,
			To:     recipientAddress,
			Amount: 1000,
			Memo:   "Test transfer",
		}

		tx, err := wallet.BuildTransferTransaction(req)
		if err != nil {
			t.Fatalf("Failed to build transfer transaction: %v", err)
		}

		// Verify transaction structure
		if len(tx.Instructions) != 1 {
			t.Errorf("Expected 1 instruction, got %d", len(tx.Instructions))
		}

		// Verify instruction program ID is System Program
		if tx.Instructions[0].ProgramID != genesis.SystemProgramID {
			t.Errorf("Expected System Program ID, got %s", tx.Instructions[0].ProgramID.String())
		}

		// Verify instruction has correct inputs
		inputs := tx.Instructions[0].Inputs
		if len(inputs) != 3 {
			t.Errorf("Expected 3 inputs, got %d", len(inputs))
		}

		// Verify sender has Write permission
		sender, ok := inputs["sender"]
		if !ok {
			t.Error("Missing sender input")
		} else if sender.Permission != transaction.Write {
			t.Errorf("Expected Write permission for sender, got %v", sender.Permission)
		}

		// Verify receiver has Write permission
		receiver, ok := inputs["receiver"]
		if !ok {
			t.Error("Missing receiver input")
		} else if receiver.Permission != transaction.Write {
			t.Errorf("Expected Write permission for receiver, got %v", receiver.Permission)
		}

		// Verify program has Read permission
		program, ok := inputs["program"]
		if !ok {
			t.Error("Missing program input")
		} else if program.Permission != transaction.Read {
			t.Errorf("Expected Read permission for program, got %v", program.Permission)
		}

		// Verify transaction is signed
		if len(tx.Signatures) != 1 {
			t.Errorf("Expected 1 signature, got %d", len(tx.Signatures))
		}

		// Verify signature is from the active account
		if tx.Signatures[0].PublicKey != account.PublicKey {
			t.Error("Signature public key does not match active account")
		}

		// Note: Signature verification would require reconstructing the exact message
		// that was signed. In a real blockchain, this would be the serialized transaction
		// without signatures. For now, we just verify the signature exists and has the
		// correct public key.
	})

	// Test invalid amount
	t.Run("InvalidAmount", func(t *testing.T) {
		req := &TransferRequest{
			From:   account.Address,
			To:     recipientAddress,
			Amount: 0,
			Memo:   "Invalid transfer",
		}

		_, err := wallet.BuildTransferTransaction(req)
		if err == nil {
			t.Error("Expected error for zero amount, got nil")
		}
	})

	// Test negative amount
	t.Run("NegativeAmount", func(t *testing.T) {
		req := &TransferRequest{
			From:   account.Address,
			To:     recipientAddress,
			Amount: -100,
			Memo:   "Invalid transfer",
		}

		_, err := wallet.BuildTransferTransaction(req)
		if err == nil {
			t.Error("Expected error for negative amount, got nil")
		}
	})

	// Test invalid recipient address
	t.Run("InvalidRecipientAddress", func(t *testing.T) {
		req := &TransferRequest{
			From:   account.Address,
			To:     "invalid-address",
			Amount: 1000,
			Memo:   "Invalid transfer",
		}

		_, err := wallet.BuildTransferTransaction(req)
		if err == nil {
			t.Error("Expected error for invalid recipient address, got nil")
		}
	})

	// Test no active account
	t.Run("NoActiveAccount", func(t *testing.T) {
		emptyWallet, err := NewWallet(config)
		if err != nil {
			t.Fatalf("Failed to create empty wallet: %v", err)
		}

		req := &TransferRequest{
			From:   account.Address,
			To:     recipientAddress,
			Amount: 1000,
			Memo:   "Test transfer",
		}

		_, err = emptyWallet.BuildTransferTransaction(req)
		if err == nil {
			t.Error("Expected error for no active account, got nil")
		}
	})
}

func TestSerializeTransaction(t *testing.T) {
	// Create a test wallet with a seed phrase
	seedPhrase, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  "/tmp/test-wallet.dat",
	}

	wallet, err := NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetActiveAccount()
	if account == nil {
		t.Fatal("No active account")
	}

	// Create a recipient address
	recipientKey := make([]byte, 32)
	for i := range recipientKey {
		recipientKey[i] = byte(i + 1)
	}
	recipientAddress := hex.EncodeToString(recipientKey)

	// Build a transfer transaction
	req := &TransferRequest{
		From:   account.Address,
		To:     recipientAddress,
		Amount: 1000,
		Memo:   "Test transfer",
	}

	tx, err := wallet.BuildTransferTransaction(req)
	if err != nil {
		t.Fatalf("Failed to build transfer transaction: %v", err)
	}

	// Serialize transaction
	txData, err := SerializeTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to serialize transaction: %v", err)
	}

	// Verify serialized data is not empty
	if len(txData) == 0 {
		t.Error("Serialized transaction data is empty")
	}

	// Deserialize and verify
	deserializedTx, err := transaction.UnmarshalTransaction(txData)
	if err != nil {
		t.Fatalf("Failed to deserialize transaction: %v", err)
	}

	// Verify deserialized transaction matches original
	if len(deserializedTx.Instructions) != len(tx.Instructions) {
		t.Errorf("Instruction count mismatch: expected %d, got %d",
			len(tx.Instructions), len(deserializedTx.Instructions))
	}

	if len(deserializedTx.Signatures) != len(tx.Signatures) {
		t.Errorf("Signature count mismatch: expected %d, got %d",
			len(tx.Signatures), len(deserializedTx.Signatures))
	}
}

// MockRPCClient is a mock RPC client for testing
type MockRPCClient struct {
	SendTransactionFunc func(txData []byte) (string, error)
	GetBalanceFunc      func(address string) (int64, error)
}

func (m *MockRPCClient) SendTransaction(txData []byte) (string, error) {
	if m.SendTransactionFunc != nil {
		return m.SendTransactionFunc(txData)
	}
	return "mock-signature-123", nil
}

func (m *MockRPCClient) GetBalance(address string) (int64, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(address)
	}
	return 1000000, nil
}

func TestSubmitTransaction(t *testing.T) {
	// Create a test wallet with a seed phrase
	seedPhrase, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  "/tmp/test-wallet.dat",
	}

	wallet, err := NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetActiveAccount()
	if account == nil {
		t.Fatal("No active account")
	}

	// Create a recipient address
	recipientKey := make([]byte, 32)
	for i := range recipientKey {
		recipientKey[i] = byte(i + 1)
	}
	recipientAddress := hex.EncodeToString(recipientKey)

	// Build a transfer transaction
	req := &TransferRequest{
		From:   account.Address,
		To:     recipientAddress,
		Amount: 1000,
		Memo:   "Test transfer",
	}

	tx, err := wallet.BuildTransferTransaction(req)
	if err != nil {
		t.Fatalf("Failed to build transfer transaction: %v", err)
	}

	// Test successful submission
	t.Run("SuccessfulSubmission", func(t *testing.T) {
		mockClient := &MockRPCClient{
			SendTransactionFunc: func(txData []byte) (string, error) {
				// Verify transaction data is not empty
				if len(txData) == 0 {
					t.Error("Transaction data is empty")
				}
				return "abc123def456", nil
			},
		}

		result, err := wallet.SubmitTransaction(tx, mockClient)
		if err != nil {
			t.Fatalf("SubmitTransaction failed: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success=true, got false. Error: %s", result.Error)
		}

		if result.Signature != "abc123def456" {
			t.Errorf("Expected signature 'abc123def456', got '%s'", result.Signature)
		}

		if result.Error != "" {
			t.Errorf("Expected empty error, got '%s'", result.Error)
		}
	})

	// Test RPC error
	t.Run("RPCError", func(t *testing.T) {
		mockClient := &MockRPCClient{
			SendTransactionFunc: func(txData []byte) (string, error) {
				return "", &RPCError{
					Code:    -32001,
					Message: "Insufficient balance",
				}
			},
		}

		result, err := wallet.SubmitTransaction(tx, mockClient)
		if err != nil {
			t.Fatalf("SubmitTransaction returned error: %v", err)
		}

		if result.Success {
			t.Error("Expected success=false for RPC error")
		}

		if result.Error == "" {
			t.Error("Expected error message, got empty string")
		}

		if result.Signature != "" {
			t.Errorf("Expected empty signature on error, got '%s'", result.Signature)
		}
	})

	// Test network error
	t.Run("NetworkError", func(t *testing.T) {
		mockClient := &MockRPCClient{
			SendTransactionFunc: func(txData []byte) (string, error) {
				return "", fmt.Errorf("connection refused")
			},
		}

		result, err := wallet.SubmitTransaction(tx, mockClient)
		if err != nil {
			t.Fatalf("SubmitTransaction returned error: %v", err)
		}

		if result.Success {
			t.Error("Expected success=false for network error")
		}

		if result.Error == "" {
			t.Error("Expected error message, got empty string")
		}
	})
}

func TestTransferRequestValidation(t *testing.T) {
	// Create a test wallet
	seedPhrase, err := GenerateSeedPhrase(12)
	if err != nil {
		t.Fatalf("Failed to generate seed phrase: %v", err)
	}

	config := &WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  "/tmp/test-wallet.dat",
	}

	wallet, err := NewWalletWithSeedPhrase(seedPhrase, config)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	account := wallet.GetActiveAccount()
	validRecipient := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

	tests := []struct {
		name        string
		request     *TransferRequest
		expectError bool
		errorCode   string
	}{
		{
			name: "ValidRequest",
			request: &TransferRequest{
				From:   account.Address,
				To:     validRecipient,
				Amount: 1000,
				Memo:   "Test",
			},
			expectError: false,
		},
		{
			name: "ZeroAmount",
			request: &TransferRequest{
				From:   account.Address,
				To:     validRecipient,
				Amount: 0,
			},
			expectError: true,
			errorCode:   ErrInvalidAmount,
		},
		{
			name: "NegativeAmount",
			request: &TransferRequest{
				From:   account.Address,
				To:     validRecipient,
				Amount: -100,
			},
			expectError: true,
			errorCode:   ErrInvalidAmount,
		},
		{
			name: "InvalidRecipientFormat",
			request: &TransferRequest{
				From:   account.Address,
				To:     "not-a-hex-string",
				Amount: 1000,
			},
			expectError: true,
			errorCode:   ErrInvalidAddress,
		},
		{
			name: "InvalidRecipientLength",
			request: &TransferRequest{
				From:   account.Address,
				To:     "0102030405", // Too short
				Amount: 1000,
			},
			expectError: true,
			errorCode:   ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := wallet.BuildTransferTransaction(tt.request)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if walletErr, ok := err.(*WalletError); ok {
					if walletErr.Code != tt.errorCode {
						t.Errorf("Expected error code %s, got %s", tt.errorCode, walletErr.Code)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

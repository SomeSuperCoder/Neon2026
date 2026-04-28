package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	"github.com/poh-blockchain/internal/filestore"
	"github.com/poh-blockchain/internal/genesis"
	"github.com/poh-blockchain/internal/transaction"
)

// TransferRequest represents a request to transfer tokens
type TransferRequest struct {
	From   string // Sender address (hex-encoded public key)
	To     string // Recipient address (hex-encoded public key)
	Amount int64  // Amount to transfer
	Memo   string // Optional memo
}

// TransferResult represents the result of a transfer operation
type TransferResult struct {
	Signature string // Transaction signature (hex-encoded)
	Success   bool   // Whether the transfer was successful
	Error     string // Error message if failed
}

// BuildTransferTransaction creates and signs a transfer transaction
// This implements Requirements 7.1, 7.2, 7.3
func (w *Wallet) BuildTransferTransaction(req *TransferRequest) (*transaction.Transaction, error) {
	// Get active account
	account := w.GetActiveAccount()
	if account == nil {
		return nil, &WalletError{
			Code:    ErrAccountNotFound,
			Message: "no active account",
		}
	}

	// Validate amount
	if req.Amount <= 0 {
		return nil, &WalletError{
			Code:    ErrInvalidAmount,
			Message: fmt.Sprintf("amount must be positive, got %d", req.Amount),
		}
	}

	// Parse sender address (should match active account)
	senderBytes, err := hex.DecodeString(req.From)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrInvalidAddress,
			Message: "invalid sender address",
			Cause:   err,
		}
	}
	if len(senderBytes) != 32 {
		return nil, &WalletError{
			Code:    ErrInvalidAddress,
			Message: fmt.Sprintf("invalid sender address length: expected 32 bytes, got %d", len(senderBytes)),
		}
	}

	// Parse recipient address
	recipientBytes, err := hex.DecodeString(req.To)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrInvalidAddress,
			Message: "invalid recipient address",
			Cause:   err,
		}
	}
	if len(recipientBytes) != 32 {
		return nil, &WalletError{
			Code:    ErrInvalidAddress,
			Message: fmt.Sprintf("invalid recipient address length: expected 32 bytes, got %d", len(recipientBytes)),
		}
	}

	// Create FileIDs from addresses
	var senderID, recipientID filestore.FileID
	copy(senderID[:], senderBytes)
	copy(recipientID[:], recipientBytes)

	// Create transaction builder with empty lastSeen (will be filled by RPC node)
	builder := transaction.NewTransactionBuilder(transaction.TxID{})

	// Add transfer instruction
	err = builder.AddTransferInstruction(
		genesis.SystemProgramID,
		senderID,
		recipientID,
		req.Amount,
	)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrTransactionBuildFailed,
			Message: "failed to add transfer instruction",
			Cause:   err,
		}
	}

	// Build transaction (without signatures)
	tx, err := builder.Build()
	if err != nil {
		return nil, &WalletError{
			Code:    ErrTransactionBuildFailed,
			Message: "failed to build transaction",
			Cause:   err,
		}
	}

	// Sign transaction with active account's private key
	// We need to sign the transaction data without signatures
	txData, err := tx.Marshal()
	if err != nil {
		return nil, &WalletError{
			Code:    ErrTransactionBuildFailed,
			Message: "failed to marshal transaction for signing",
			Cause:   err,
		}
	}

	// Sign the transaction data
	signature := ed25519.Sign(account.PrivateKey, txData)

	// Create signature structure
	var sig [64]byte
	copy(sig[:], signature)

	txSig := transaction.Signature{
		PublicKey: account.PublicKey,
		Signature: sig,
	}

	// Add signature to transaction
	tx.Signatures = []transaction.Signature{txSig}

	return tx, nil
}

// SerializeTransaction serializes a transaction for RPC submission
func SerializeTransaction(tx *transaction.Transaction) ([]byte, error) {
	return tx.Marshal()
}

// SubmitTransaction submits a signed transaction to the RPC node
// This implements Requirements 7.3, 7.4, 7.5
func (w *Wallet) SubmitTransaction(tx *transaction.Transaction, rpcClient RPCClient) (*TransferResult, error) {
	// Serialize transaction
	txData, err := SerializeTransaction(tx)
	if err != nil {
		return nil, &WalletError{
			Code:    ErrTransactionSubmitFailed,
			Message: "failed to serialize transaction",
			Cause:   err,
		}
	}

	// Submit transaction via RPC
	signature, err := rpcClient.SendTransaction(txData)
	if err != nil {
		// Parse RPC error for user-friendly message
		errorMsg := err.Error()
		if rpcErr, ok := err.(*RPCError); ok {
			errorMsg = rpcErr.Message
		}

		return &TransferResult{
			Success: false,
			Error:   errorMsg,
		}, nil
	}

	// Update account balance (will be refreshed from RPC)
	// For now, we just mark that an update is needed
	account := w.GetActiveAccount()
	if account != nil {
		// Balance will be updated by RefreshBalances call
		// We don't update it here to avoid inconsistency
	}

	return &TransferResult{
		Signature: signature,
		Success:   true,
	}, nil
}

// RPCClient interface for transaction submission
// This allows for easier testing with mock clients
type RPCClient interface {
	SendTransaction(txData []byte) (string, error)
	GetBalance(address string) (int64, error)
}

// RPCError represents an RPC error (matches the one in cmd/wallet/rpc package)
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

// Error implements the error interface
func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

package core

import (
	"crypto/ed25519"
	"time"
)

// Account represents an account derived from an imported seed phrase
type Account struct {
	SeedPhraseIndex int                `json:"seedPhraseIndex"` // Index into wallet's seedPhrases array
	PublicKey       [32]byte           `json:"publicKey"`
	PrivateKey      ed25519.PrivateKey `json:"-"` // Never serialize private key
	Address         string             `json:"address"`
	Label           string             `json:"label"`
	Balance         int64              `json:"balance"`
	LastUpdate      time.Time          `json:"lastUpdate"`
}

// WalletFile represents the encrypted wallet file format
type WalletFile struct {
	Version   int             `json:"version"`
	Encrypted EncryptedWallet `json:"encrypted"`
	Config    WalletConfig    `json:"config"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// EncryptedWallet contains the encrypted wallet data
type EncryptedWallet struct {
	Version    int                `json:"version"`
	Cipher     string             `json:"cipher"`     // "AES-256-GCM"
	Salt       string             `json:"salt"`       // Base64 encoded
	IV         string             `json:"iv"`         // Base64 encoded
	Ciphertext string             `json:"ciphertext"` // Base64 encoded (contains all seed phrases)
	Accounts   []EncryptedAccount `json:"accounts"`
}

// EncryptedAccount contains non-sensitive account metadata
type EncryptedAccount struct {
	SeedPhraseIndex int    `json:"seedPhraseIndex"`
	Label           string `json:"label"`
}

// WalletError represents a wallet-specific error
type WalletError struct {
	Code    string
	Message string
	Cause   error
}

func (e *WalletError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Error codes
const (
	ErrInvalidSeedPhrase       = "INVALID_SEED_PHRASE"
	ErrDuplicateSeedPhrase     = "DUPLICATE_SEED_PHRASE"
	ErrWalletLocked            = "WALLET_LOCKED"
	ErrInvalidPassword         = "INVALID_PASSWORD"
	ErrAccountNotFound         = "ACCOUNT_NOT_FOUND"
	ErrInsufficientFunds       = "INSUFFICIENT_FUNDS"
	ErrRPCConnection           = "RPC_CONNECTION_ERROR"
	ErrTransactionFailed       = "TRANSACTION_FAILED"
	ErrEncryptionFailed        = "ENCRYPTION_FAILED"
	ErrDecryptionFailed        = "DECRYPTION_FAILED"
	ErrInvalidAmount           = "INVALID_AMOUNT"
	ErrInvalidAddress          = "INVALID_ADDRESS"
	ErrTransactionBuildFailed  = "TRANSACTION_BUILD_FAILED"
	ErrTransactionSubmitFailed = "TRANSACTION_SUBMIT_FAILED"
)

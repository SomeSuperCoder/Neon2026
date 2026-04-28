package core

import (
	"time"
)

// WalletConfig holds wallet configuration settings
type WalletConfig struct {
	RPCEndpoint string        `json:"rpcEndpoint"`
	WalletPath  string        `json:"walletPath"`
	AutoLock    time.Duration `json:"autoLock"`
	Theme       string        `json:"theme"`
}

// DefaultConfig returns the default wallet configuration
func DefaultConfig() *WalletConfig {
	return &WalletConfig{
		RPCEndpoint: "http://localhost:8899",
		WalletPath:  "", // Will be set to ~/.poh-wallet/wallet.dat
		AutoLock:    5 * time.Minute,
		Theme:       "neon",
	}
}

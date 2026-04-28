package rpc

import (
	"time"
)

// ServerConfig holds configuration for the RPC server
type ServerConfig struct {
	// BindAddress is the IP address to bind to (default: "127.0.0.1")
	BindAddress string

	// Port is the HTTP listening port (default: 8899)
	Port int

	// MaxConns is the maximum number of simultaneous connections (default: 100)
	MaxConns int

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration

	// LedgerPath is the path to the SQLite ledger database
	LedgerPath string

	// StatePath is the path to the FileStore state database
	StatePath string
}

// DefaultServerConfig returns a ServerConfig with default values
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		BindAddress:  "127.0.0.1",
		Port:         8899,
		MaxConns:     100,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

# Wallet Configuration Reference

Complete reference for the PoH Blockchain wallet configuration system.

## Overview

The wallet configuration system provides customizable settings for the wallet application, including RPC endpoint, auto-lock timeout, and theme preferences.

## Configuration Structure

### WalletConfig

The `WalletConfig` struct holds all wallet configuration settings:

```go
type WalletConfig struct {
    RPCEndpoint string        `json:"rpcEndpoint"`
    WalletPath  string        `json:"walletPath"`
    AutoLock    time.Duration `json:"autoLock"`
    Theme       string        `json:"theme"`
}
```

**Fields:**

- `RPCEndpoint` (string): URL of the RPC node to connect to
  - Default: `http://localhost:8899`
  - Format: `http://host:port` or `https://host:port`
  - Used for all blockchain queries and transaction submissions

- `WalletPath` (string): Path to the encrypted wallet file
  - Default: `~/.poh-wallet/wallet.dat` (set at runtime)
  - Can be customized via CLI flag `--wallet-path`
  - File permissions automatically set to 0600 for security

- `AutoLock` (time.Duration): Timeout for automatic wallet locking
  - Default: `5 * time.Minute` (5 minutes)
  - Wallet locks after this period of inactivity
  - Requires password to unlock
  - Set to 0 to disable auto-lock (not recommended)

- `Theme` (string): UI theme for the wallet TUI
  - Default: `"neon"`
  - Available themes: `neon` (more themes coming soon)
  - Affects colors, styling, and visual effects

## Default Configuration

Use `DefaultConfig()` to get a configuration with sensible defaults:

```go
config := core.DefaultConfig()
// Returns:
// {
//     RPCEndpoint: "http://localhost:8899",
//     WalletPath:  "",  // Will be set to ~/.poh-wallet/wallet.dat
//     AutoLock:    5 * time.Minute,
//     Theme:       "neon",
// }
```

## Usage Examples

### Creating a Custom Configuration

```go
import (
    "time"
    "github.com/poh-blockchain/cmd/wallet/core"
)

// Start with defaults
config := core.DefaultConfig()

// Customize settings
config.RPCEndpoint = "http://mainnet.example.com:8899"
config.AutoLock = 10 * time.Minute
config.WalletPath = "/secure/path/my-wallet.dat"
```

### Using Configuration in Wallet

```go
// Create wallet with custom config
wallet, err := core.NewWallet(config)
if err != nil {
    log.Fatalf("Failed to create wallet: %v", err)
}
defer wallet.Close()

// Configuration is accessible via wallet
fmt.Printf("Connected to: %s\n", wallet.Config.RPCEndpoint)
```

### CLI Configuration

Configuration can be set via command-line flags:

```bash
# Custom RPC endpoint
poh-wallet --rpc-url http://mainnet.example.com:8899

# Custom wallet path
poh-wallet --wallet-path /secure/path/my-wallet.dat

# Both
poh-wallet --rpc-url http://localhost:9000 --wallet-path ./my-wallet.dat
```

## Configuration Persistence

The wallet configuration is saved as part of the encrypted wallet file. When you modify settings through the wallet UI, they are automatically persisted.

**Configuration Storage:**
- Stored in the encrypted wallet file
- Encrypted with AES-256-GCM
- Protected by wallet password
- Automatically loaded on wallet unlock

## Security Considerations

### RPC Endpoint Security

- **Use HTTPS in production**: Always use `https://` for remote RPC endpoints
- **Verify certificates**: Ensure SSL/TLS certificates are valid
- **Avoid public endpoints**: Use trusted RPC nodes only
- **Local development**: `http://localhost:8899` is safe for local testing

### Wallet Path Security

- **File permissions**: Wallet files are automatically set to 0600 (owner read/write only)
- **Secure location**: Store wallet files in secure directories
- **Backup strategy**: Keep encrypted backups in secure locations
- **Never share**: Wallet files contain encrypted private keys

### Auto-Lock Settings

- **Balance security vs convenience**: Shorter timeouts are more secure
- **Recommended minimum**: 5 minutes (default)
- **High-security environments**: Consider 1-2 minutes
- **Never disable**: Auto-lock is a critical security feature

## Configuration Validation

The wallet validates configuration on startup:

```go
// Validate RPC endpoint format
if !strings.HasPrefix(config.RPCEndpoint, "http://") && 
   !strings.HasPrefix(config.RPCEndpoint, "https://") {
    return fmt.Errorf("invalid RPC endpoint: must start with http:// or https://")
}

// Validate auto-lock timeout
if config.AutoLock < 0 {
    return fmt.Errorf("invalid auto-lock timeout: must be non-negative")
}

// Validate wallet path
if config.WalletPath == "" {
    config.WalletPath = filepath.Join(os.Getenv("HOME"), ".poh-wallet", "wallet.dat")
}
```

## Environment Variables

Configuration can also be set via environment variables:

```bash
# RPC endpoint
export POH_WALLET_RPC_ENDPOINT="http://localhost:8899"

# Wallet path
export POH_WALLET_PATH="$HOME/.poh-wallet/wallet.dat"

# Auto-lock timeout (in minutes)
export POH_WALLET_AUTO_LOCK_MINUTES=5

# Theme
export POH_WALLET_THEME="neon"
```

**Priority order:**
1. Command-line flags (highest priority)
2. Environment variables
3. Saved configuration in wallet file
4. Default values (lowest priority)

## Best Practices

### Development

```go
// Use default config for local development
config := core.DefaultConfig()
config.RPCEndpoint = "http://localhost:8899"
config.WalletPath = "./dev-wallet.dat"
```

### Production

```go
// Use secure settings for production
config := core.DefaultConfig()
config.RPCEndpoint = "https://mainnet.example.com:8899"
config.AutoLock = 2 * time.Minute  // Shorter timeout
config.WalletPath = "/secure/encrypted/volume/wallet.dat"
```

### Testing

```go
// Use in-memory or temporary paths for testing
config := core.DefaultConfig()
config.RPCEndpoint = "http://localhost:18899"  // Test RPC port
config.WalletPath = filepath.Join(t.TempDir(), "test-wallet.dat")
config.AutoLock = 30 * time.Second  // Faster for tests
```

## Troubleshooting

### Cannot Connect to RPC

**Problem:** Wallet cannot connect to RPC endpoint

**Solutions:**
- Verify RPC node is running: `curl http://localhost:8899`
- Check firewall settings
- Verify endpoint URL format
- Try default endpoint: `http://localhost:8899`

### Wallet File Not Found

**Problem:** Wallet file cannot be found or created

**Solutions:**
- Check wallet path is correct
- Verify directory exists and is writable
- Check file permissions (should be 0600)
- Use absolute path instead of relative path

### Auto-Lock Not Working

**Problem:** Wallet doesn't lock after timeout

**Solutions:**
- Verify auto-lock is not set to 0
- Check for active operations that prevent locking
- Restart wallet application
- Verify configuration was saved

## See Also

- [Wallet User Guide](../guides/wallet-usage.md) - Complete wallet usage guide
- [RPC API Reference](rpc-api.md) - RPC endpoint documentation
- [Security Model](security-model.md) - Security best practices
- [CLI Usage Guide](../guides/cli-usage.md) - Command-line interface

# Validator Wallet Quick Start Guide

## Overview

This guide shows you how to create and manage validator wallets for running PoH blockchain validator nodes.

## What is a Validator Wallet?

A validator wallet is an encrypted file containing Ed25519 keypairs used for:
- Validator node identity
- Block signing
- Consensus participation

**Note**: This is different from the Neon Wallet TUI, which is for end-users managing accounts and transactions.

## Installation

The validator wallet is part of the main blockchain binary:

```bash
go build -o poh-blockchain ./cmd/main.go
```

## Quick Start

### 1. Create a Validator Wallet

```bash
# Create wallet programmatically (CLI commands coming soon)
```

For now, create wallets using the Go API:

```go
package main

import (
    "fmt"
    "log"
    "github.com/poh-blockchain/internal/wallet"
)

func main() {
    // Create new wallet
    w, err := wallet.Create("validator1", "strong-password-123")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✓ Wallet created: %s\n", w.Name)
    fmt.Printf("✓ Public key: %x\n", w.Keypairs[0].PublicKey[:16])
    fmt.Printf("✓ Location: ~/.config/poh-blockchain/wallets/validator1.wallet\n")
}
```

### 2. List Your Wallets

```go
wallets, err := wallet.List()
if err != nil {
    log.Fatal(err)
}

fmt.Println("Available wallets:")
for _, name := range wallets {
    fmt.Printf("  - %s\n", name)
}
```

### 3. Open an Existing Wallet

```go
w, err := wallet.Open("validator1", "strong-password-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Wallet opened: %s\n", w.Name)
fmt.Printf("Keypairs: %d\n", len(w.Keypairs))
```

### 4. Start Validator Node with Wallet

```bash
# Start validator with wallet identity
./poh-blockchain --wallet validator1

# You'll be prompted for password:
# Enter wallet password: ********

# Start in observer mode (no validation)
./poh-blockchain
```

## Backup and Recovery

### Export Wallet for Backup

```go
// Export to unencrypted JSON (SECURE THIS FILE!)
err := w.Export("/secure/location/validator1-backup.json")
if err != nil {
    log.Fatal(err)
}

fmt.Println("✓ Wallet exported")
fmt.Println("⚠ WARNING: Backup file contains unencrypted private keys!")
fmt.Println("⚠ Store securely and delete after copying to safe location")
```

### Import Wallet from Backup

```go
// Import from backup with new name and password
w, err := wallet.Import(
    "/secure/location/validator1-backup.json",
    "validator1-restored",
    "new-strong-password",
)
if err != nil {
    log.Fatal(err)
}

fmt.Println("✓ Wallet restored from backup")
```

## Security Best Practices

### Password Requirements

✓ **Minimum 8 characters** (enforced)  
✓ **Recommended: 16+ characters**  
✓ **Use mixed case, numbers, symbols**  
✓ **Use a password manager**  
✗ **Never reuse passwords**

### Wallet File Security

- Wallet files are created with `0600` permissions (owner-only access)
- Store on encrypted filesystems when possible
- Never commit wallet files to version control
- Keep backups on encrypted, offline storage

### Export/Import Security

⚠ **CRITICAL**: Exported JSON files contain unencrypted private keys!

- Only export for backup or migration
- Delete exported files immediately after use
- Never transmit over unencrypted channels
- Store backups on encrypted, offline media

## Common Tasks

### Check Wallet Location

```go
dir, err := wallet.GetWalletDir()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Wallet directory: %s\n", dir)

// Platform-specific locations:
// Linux/macOS: ~/.config/poh-blockchain/wallets/
// Windows: %APPDATA%\poh-blockchain\wallets\
```

### Get Wallet File Path

```go
path, err := wallet.GetWalletPath("validator1")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Wallet file: %s\n", path)
```

### Verify Wallet Exists

```go
import "os"

path, _ := wallet.GetWalletPath("validator1")
if _, err := os.Stat(path); err == nil {
    fmt.Println("✓ Wallet exists")
} else if os.IsNotExist(err) {
    fmt.Println("✗ Wallet not found")
}
```

## Troubleshooting

### Wallet Already Exists

**Error**: `wallet 'validator1' already exists`

**Solution**: Choose a different name or delete the existing wallet:
```bash
rm ~/.config/poh-blockchain/wallets/validator1.wallet
```

### Incorrect Password

**Error**: `failed to decrypt wallet (incorrect password?)`

**Solutions**:
1. Verify you're using the correct password
2. Restore from backup if password is lost
3. Create new wallet if no backup exists

### Wallet Not Found

**Error**: `wallet 'validator1' not found`

**Solutions**:
1. Check wallet name spelling
2. List available wallets: `wallet.List()`
3. Verify wallet directory exists

### Password Too Short

**Error**: `password must be at least 8 characters`

**Solution**: Use a longer password (16+ characters recommended)

## Multi-Validator Setup

For running multiple validators:

```go
// Create separate wallet for each validator
validators := []string{"validator1", "validator2", "validator3"}

for _, name := range validators {
    w, err := wallet.Create(name, "unique-password-"+name)
    if err != nil {
        log.Printf("Failed to create %s: %v", name, err)
        continue
    }
    fmt.Printf("✓ Created %s\n", name)
}
```

Start each validator with its own wallet:

```bash
# Terminal 1
./poh-blockchain --wallet validator1 --port 8000

# Terminal 2
./poh-blockchain --wallet validator2 --port 8001

# Terminal 3
./poh-blockchain --wallet validator3 --port 8002
```

## Next Steps

- Read the [Validator Wallet Reference](../reference/validator-wallet.md) for complete API documentation
- Learn about [Stake-Weighted Leader Schedule](../../.kiro/specs/stake-weighted-leader-schedule/) for consensus integration
- Review [Security Model](../reference/security-model.md) for overall security architecture

## CLI Commands (Coming Soon)

Future tasks will add CLI commands for wallet management:

```bash
# Create wallet
./poh-blockchain wallet create --name validator1

# List wallets
./poh-blockchain wallet list

# Show wallet info
./poh-blockchain wallet show --name validator1

# Export wallet
./poh-blockchain wallet export --name validator1 --output backup.json

# Import wallet
./poh-blockchain wallet import --input backup.json --name validator2
```

See task #2 in [tasks.md](../../.kiro/specs/stake-weighted-leader-schedule/tasks.md) for implementation status.

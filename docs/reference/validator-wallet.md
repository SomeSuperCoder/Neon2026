# Validator Wallet Reference

## Overview

The validator wallet system (`internal/wallet`) provides secure, encrypted storage for Ed25519 keypairs used by validators for node identity and block signing. This is separate from the user-facing Neon Wallet TUI and is specifically designed for validator node operations.

## Architecture

### Core Components

- **Wallet**: Password-protected container for one or more Ed25519 keypairs
- **Keypair**: Ed25519 public/private key pair (32-byte public key, 64-byte private key)
- **Encryption**: AES-256-GCM with Argon2id key derivation
- **Storage**: Platform-specific wallet directory with `.wallet` file extension

### Security Features

1. **AES-256-GCM Encryption**: Industry-standard authenticated encryption
2. **Argon2id Key Derivation**: Memory-hard password hashing resistant to GPU attacks
   - Memory: 64 MB
   - Iterations: 3
   - Parallelism: 4
   - Output: 32-byte key
3. **Random Salt**: 32-byte unique salt per wallet
4. **Random Nonce**: 12-byte unique nonce per encryption
5. **Authentication Tag**: 16-byte GCM tag for integrity verification
6. **Minimum Password Length**: 8 characters

### File Format

Encrypted wallet files use the following binary format:

```
[salt(32 bytes)][nonce(12 bytes)][ciphertext][tag(16 bytes)]
```

The ciphertext contains JSON-encoded wallet data:

```json
{
  "version": 1,
  "keypairs": [
    {
      "publicKey": "base64-encoded-32-bytes",
      "privateKey": "base64-encoded-64-bytes"
    }
  ]
}
```

### Platform-Specific Paths

Wallet files are stored in platform-specific directories:

- **Linux/macOS**: `~/.config/poh-blockchain/wallets/`
- **Windows**: `%APPDATA%\poh-blockchain\wallets\`

## API Reference

### Creating a Wallet

```go
import "github.com/poh-blockchain/internal/wallet"

// Create new wallet with generated keypair
w, err := wallet.Create("validator1", "strong-password-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Wallet created: %s\n", w.Name)
fmt.Printf("Public key: %x\n", w.Keypairs[0].PublicKey)
```

**Parameters:**
- `name`: Wallet name (used as filename without extension)
- `password`: Password for encryption (minimum 8 characters)

**Returns:**
- `*Wallet`: Created wallet with one generated Ed25519 keypair
- `error`: Error if wallet already exists or password is invalid

### Opening a Wallet

```go
// Open existing wallet
w, err := wallet.Open("validator1", "strong-password-123")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Wallet opened: %s\n", w.Name)
fmt.Printf("Keypairs: %d\n", len(w.Keypairs))
```

**Parameters:**
- `name`: Wallet name
- `password`: Password for decryption

**Returns:**
- `*Wallet`: Opened wallet with all keypairs
- `error`: Error if wallet not found or password incorrect

### Listing Wallets

```go
// List all wallets in wallet directory
wallets, err := wallet.List()
if err != nil {
    log.Fatal(err)
}

for _, name := range wallets {
    fmt.Printf("- %s\n", name)
}
```

**Returns:**
- `[]string`: List of wallet names (without `.wallet` extension)
- `error`: Error if wallet directory cannot be read

### Exporting a Wallet

```go
// Export wallet to unencrypted JSON file
err := w.Export("/path/to/backup.json")
if err != nil {
    log.Fatal(err)
}
```

**Warning**: Exported files contain unencrypted private keys. Store securely and delete after use.

**Parameters:**
- `outputPath`: Path to export file

**Returns:**
- `error`: Error if export fails

### Importing a Wallet

```go
// Import keypairs from JSON file and create encrypted wallet
w, err := wallet.Import("/path/to/backup.json", "validator2", "new-password")
if err != nil {
    log.Fatal(err)
}
```

**Parameters:**
- `inputPath`: Path to JSON file with keypairs
- `name`: Name for new wallet
- `password`: Password for encryption

**Returns:**
- `*Wallet`: Created wallet with imported keypairs
- `error`: Error if import fails or wallet already exists

### Utility Functions

```go
// Get platform-specific wallet directory
dir, err := wallet.GetWalletDir()

// Get full path to wallet file
path, err := wallet.GetWalletPath("validator1")
```

## Usage in Validator Nodes

### Node Startup with Wallet

When starting a validator node, the wallet provides the node's identity:

```bash
# Start validator with wallet
./poh-blockchain --wallet validator1

# Start in observer mode (no wallet)
./poh-blockchain
```

The node will:
1. Prompt for wallet password
2. Load keypairs from wallet
3. If multiple keypairs exist, prompt for selection
4. Compute validator ID from selected public key
5. Use keypair for block signing

### Validator Identity Computation

The validator's FileID is computed from the public key:

```go
import (
    "crypto/sha256"
    "github.com/poh-blockchain/internal/filestore"
)

// Compute validator FileID from public key
func ComputeValidatorID(pubKey [32]byte) filestore.FileID {
    h := sha256.New()
    h.Write([]byte("validator:"))
    h.Write(pubKey[:])
    var id filestore.FileID
    copy(id[:], h.Sum(nil))
    return id
}
```

## Security Best Practices

### Password Requirements

- Minimum 8 characters (enforced)
- Recommended: 16+ characters with mixed case, numbers, and symbols
- Use a password manager for strong, unique passwords
- Never reuse passwords from other services

### Wallet File Protection

- Wallet files are created with `0600` permissions (owner read/write only)
- Store wallet files on encrypted filesystems when possible
- Regular backups to secure, offline storage
- Never commit wallet files to version control

### Export/Import Security

- Only export wallets for backup or migration
- Delete exported JSON files immediately after use
- Never transmit exported files over unencrypted channels
- Store backups on encrypted, offline media

### Key Management

- Generate one wallet per validator node
- Never share private keys between nodes
- Rotate keys periodically (requires validator re-registration)
- Keep backup of wallet files and passwords separately

## Testing

The wallet package includes comprehensive unit tests:

```bash
# Run all wallet tests
go test ./internal/wallet/...

# Run with verbose output
go test -v ./internal/wallet/...

# Run with coverage
go test -cover ./internal/wallet/...
```

### Test Coverage

- Platform-specific path resolution (Linux, macOS, Windows)
- Encryption/decryption round-trips
- Password validation
- Incorrect password handling
- Corrupted data handling
- Wallet creation and opening
- Wallet listing
- Export and import operations
- Duplicate wallet detection

## Error Handling

Common errors and solutions:

### Wallet Already Exists

```
Error: wallet 'validator1' already exists
```

**Solution**: Choose a different name or delete the existing wallet

### Wallet Not Found

```
Error: wallet 'validator1' not found
```

**Solution**: Verify wallet name and check wallet directory

### Incorrect Password

```
Error: failed to decrypt wallet (incorrect password?)
```

**Solution**: Verify password or restore from backup

### Password Too Short

```
Error: password must be at least 8 characters
```

**Solution**: Use a longer password

### Corrupted Wallet File

```
Error: ciphertext too short: expected at least 60 bytes, got 32
```

**Solution**: Restore from backup or import from exported JSON

## Migration from Legacy System

The validator wallet system replaces the previous `--type=leader/replica` flag system:

### Before (Deprecated)

```bash
# Old way - static node types
./poh-blockchain --type=leader --port=8000
./poh-blockchain --type=replica --port=8001
```

### After (Current)

```bash
# New way - wallet-based identity
./poh-blockchain --wallet validator1 --port=8000
./poh-blockchain --wallet validator2 --port=8001

# Observer mode (no validation)
./poh-blockchain --port=8002
```

### Migration Steps

1. Create wallet for each validator:
   ```bash
   # This will be done via CLI in future tasks
   # For now, use programmatic API
   ```

2. Update startup scripts to use `--wallet` flag

3. Remove `--type` flag from all scripts

4. Update genesis configuration with validator public keys

## Related Documentation

- [Stake-Weighted Leader Schedule](../specs/stake-weighted-leader-schedule/) - How wallets integrate with consensus
- [Consensus Manager](consensus-manager.md) - Validator identity in consensus
- [Genesis Configuration](genesis-config.md) - Initial validator setup
- [Security Model](security-model.md) - Overall security architecture

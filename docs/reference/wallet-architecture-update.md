# Wallet Architecture Update: Multi-Seed Phrase Support

## Overview

The wallet architecture has been updated from a hierarchical deterministic (HD) wallet model to a multi-seed phrase import model. This change allows users to import and manage multiple independent seed phrases rather than deriving multiple accounts from a single seed phrase.

## Architecture Changes

### Previous Model (HD Wallet)
- **1 seed phrase** → **Multiple derived accounts** (using BIP44 derivation indices)
- Derivation path: `m/44'/501'/0'/0'/index'` where `index` varied per account
- All accounts shared the same seed phrase

### New Model (Multi-Seed Phrase)
- **Multiple seed phrases** → **1 account per seed phrase**
- Derivation path: `m/44'/501'/0'/0'/0'` (fixed at index 0 for each seed phrase)
- Each account has its own independent seed phrase

## Benefits

1. **Flexibility**: Users can import seed phrases from different sources (hardware wallets, other software wallets, etc.)
2. **Security**: Compromising one seed phrase doesn't affect other accounts
3. **Compatibility**: Users can import existing seed phrases from other wallets
4. **Organization**: Better separation of funds across truly independent accounts

## Updated Requirements

### Requirement 4: Seed Phrase Management
- Changed from "derive master keypair" to "derive keypair at index 0"
- Added support for importing seed phrases
- Encryption now covers all seed phrases

### Requirement 5: Multi-Seed Phrase Management (formerly Multi-Account Management)
- Changed from "derive keypair from Seed_Phrase using next sequential derivation index" to "derive one keypair at index 0 per imported seed phrase"
- Support minimum of 100 imported seed phrases (instead of 100 derived accounts)
- Added duplicate seed phrase detection

## Code Changes

### Types (`cmd/wallet/core/types.go`)

**Account Structure:**
```go
// Before
type Account struct {
    Index      uint32              // Derivation index
    PublicKey  [32]byte
    PrivateKey ed25519.PrivateKey
    Address    string
    Label      string
    Balance    int64
    LastUpdate time.Time
}

// After
type Account struct {
    SeedPhraseIndex int                 // Index into wallet's seedPhrases array
    PublicKey       [32]byte
    PrivateKey      ed25519.PrivateKey
    Address         string
    Label           string
    Balance         int64
    LastUpdate      time.Time
}
```

**EncryptedAccount Structure:**
```go
// Before
type EncryptedAccount struct {
    Index  uint32
    Label  string
}

// After
type EncryptedAccount struct {
    SeedPhraseIndex int
    Label           string
}
```

**New Error Code:**
- Added `ErrDuplicateSeedPhrase` for duplicate seed phrase detection

### Wallet (`cmd/wallet/core/wallet.go`)

**Wallet Structure:**
```go
// Before
type Wallet struct {
    seedPhrase   string      // Single seed phrase
    masterKey    []byte      // Derived master key
    accounts     []*Account
    activeIndex  int
    config       *WalletConfig
    encrypted    bool
}

// After
type Wallet struct {
    seedPhrases  []string    // Multiple imported seed phrases
    accounts     []*Account  // One account per seed phrase
    activeIndex  int
    config       *WalletConfig
    encrypted    bool
}
```

**Key API Changes:**

1. **NewWallet()** - Now creates an empty wallet
   ```go
   func NewWallet(config *WalletConfig) (*Wallet, error)
   ```

2. **NewWalletWithSeedPhrase()** - Creates wallet with first seed phrase
   ```go
   func NewWalletWithSeedPhrase(seedPhrase string, config *WalletConfig) (*Wallet, error)
   ```

3. **ImportSeedPhrase()** - New method to import additional seed phrases
   ```go
   func (w *Wallet) ImportSeedPhrase(seedPhrase string, label string) (*Account, error)
   ```

4. **Removed DeriveAccount()** - No longer needed since we don't derive multiple accounts per seed phrase

5. **New Methods:**
   - `GetAccountBySeedPhraseIndex(seedPhraseIndex int) *Account`
   - `GetSeedPhrases() []string`
   - `GetSeedPhrase(index int) string`
   - `AccountCount() int`

### Encryption (`cmd/wallet/core/encryption.go`)

**Encrypted Data Structure:**
```go
// Before
walletData := struct {
    SeedPhrase  string
    Accounts    []*Account
    ActiveIndex int
}

// After
walletData := struct {
    SeedPhrases []string    // Array of seed phrases
    Accounts    []*Account
    ActiveIndex int
}
```

**Encryption Tests Updated:**
All encryption tests now verify:
- Multiple seed phrases are encrypted and decrypted correctly
- Each seed phrase is preserved independently
- Duplicate seed phrase detection works
- Account-to-seed-phrase mapping is maintained

### Derivation (`cmd/wallet/core/derivation.go`)

No changes to the derivation logic itself, but usage changed:
- Always derive at index 0: `m/44'/501'/0'/0'/0'`
- Each seed phrase gets its own derivation

## Test Coverage

All tests have been updated and pass successfully:

### New Tests
- `TestImportSeedPhrase` - Tests importing multiple seed phrases
- `TestImportDuplicateSeedPhrase` - Tests duplicate detection
- `TestImportInvalidSeedPhrase` - Tests validation
- `TestGetAccountBySeedPhraseIndex` - Tests new getter method
- `TestSupport100SeedPhrases` - Tests requirement for 100+ seed phrases
- `TestWalletEncryption` - Tests encryption with multiple seed phrases

### Updated Tests
- `TestEncryptDecryptWallet` - Now imports 5 seed phrases instead of deriving accounts
- `TestEncryptDecryptWithDifferentPasswords` - Uses `NewWalletWithSeedPhrase`
- `TestEncryptionRandomness` - Uses `NewWalletWithSeedPhrase`
- All wallet tests updated to use `ImportSeedPhrase()` instead of `DeriveAccount()`
- All encryption tests updated to handle multiple seed phrases
- Account structure tests updated for `SeedPhraseIndex` field

### Test Results
```
✓ 7 derivation tests
✓ 7 encryption tests (including new TestWalletEncryption)
✓ 9 mnemonic tests
✓ 13 wallet tests (including new multi-seed phrase tests)
Total: 36 tests passing
```

## Migration Guide

For code using the old API:

### Before (HD Wallet)
```go
// Create wallet with seed phrase
wallet, _ := NewWallet(seedPhrase, config)

// Derive additional accounts
account1, _ := wallet.DeriveAccount(1)
account2, _ := wallet.DeriveAccount(2)
```

### After (Multi-Seed Phrase)
```go
// Create empty wallet
wallet, _ := NewWallet(config)

// Import seed phrases
account0, _ := wallet.ImportSeedPhrase(seedPhrase1, "Account 1")
account1, _ := wallet.ImportSeedPhrase(seedPhrase2, "Account 2")
account2, _ := wallet.ImportSeedPhrase(seedPhrase3, "Account 3")

// Or create with first seed phrase
wallet, _ := NewWalletWithSeedPhrase(seedPhrase, config)
```

## Design Document Updates

Updated sections in `.kiro/specs/rpc-node-and-wallet/design.md`:
- Section 2.2: Wallet Core structure
- Section 3.2: BIP44 key derivation (now fixed at index 0)
- Section 3.3: Wallet encryption (handles multiple seed phrases)

## Requirements Document Updates

Updated sections in `.kiro/specs/rpc-node-and-wallet/requirements.md`:
- Requirement 4: Seed Phrase Management
- Requirement 5: Multi-Seed Phrase Management (renamed from Multi-Account Management)

## Backward Compatibility

This is a **breaking change**. Wallets created with the old format cannot be loaded with the new code. A migration tool would be needed to:
1. Read old wallet format
2. Extract the single seed phrase
3. Create new wallet with that seed phrase imported
4. Save in new format

## Future Enhancements

Possible future additions:
1. **Hybrid Model**: Support both multi-seed phrase AND derivation (e.g., derive 5 accounts per seed phrase)
2. **Seed Phrase Export**: Allow exporting individual seed phrases
3. **Seed Phrase Removal**: Allow removing imported seed phrases (with confirmation)
4. **Seed Phrase Metadata**: Store additional metadata per seed phrase (source, import date, etc.)

## RPC Client Integration

### Overview

The wallet includes a JSON-RPC 2.0 client (`cmd/wallet/rpc/client.go`) for communicating with the blockchain RPC node. This client provides a clean, type-safe interface for all blockchain operations.

### RPC Client Structure

```go
type RPCClient struct {
    endpoint   string
    httpClient *http.Client
    requestID  uint64  // Auto-incrementing request ID
}
```

**Features:**
- 10-second request timeout
- Auto-incrementing request IDs for tracking
- Automatic JSON-RPC 2.0 request/response handling
- Type-safe method interfaces
- Comprehensive error handling

### Available Methods

1. **GetBalance(address string) (int64, error)**
   - Retrieves account balance in electrons
   - Returns 0 for non-existent accounts

2. **GetAccountInfo(address string) (*rpc.AccountInfo, error)**
   - Returns full account information (balance, owner, data length, executable status)
   - Returns nil for non-existent accounts

3. **GetTransactionHistory(address string, limit int) ([]rpc.TransactionRecord, error)**
   - Retrieves transaction history with pagination
   - Returns transactions in reverse chronological order
   - Default limit: 20 transactions

4. **SendTransaction(txData []byte) (string, error)**
   - Submits a signed transaction to the blockchain
   - Returns transaction signature on success
   - Handles signature verification errors

5. **GetTransactionStatus(signature string) (*rpc.TransactionStatus, error)**
   - Checks transaction confirmation status
   - Returns block height and slot if confirmed

6. **GetBlockHeight() (int64, error)**
   - Returns current blockchain height

### Error Handling

The client includes a custom `RPCError` type that wraps RPC errors:

```go
type RPCError struct {
    Code    int
    Message string
    Data    interface{}
}
```

**Common Error Codes:**
- `-32700`: Parse error (invalid JSON)
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32001`: Invalid signature
- `-32002`: Malformed transaction
- `-32003`: Insufficient balance

### Usage Example

```go
// Create RPC client
client := rpc.NewRPCClient("http://localhost:8899")

// Query balance
balance, err := client.GetBalance(address)
if err != nil {
    // Handle error
}

// Get account info
info, err := client.GetAccountInfo(address)
if err != nil {
    // Handle error
}

// Send transaction
signature, err := client.SendTransaction(txBytes)
if err != nil {
    if rpcErr, ok := err.(*rpc.RPCError); ok {
        // Handle specific RPC error
        switch rpcErr.Code {
        case -32001:
            // Invalid signature
        case -32003:
            // Insufficient balance
        }
    }
}
```

### Test Coverage

The RPC client includes comprehensive unit tests (`cmd/wallet/rpc/client_test.go`):

**Test Cases:**
- Client creation and configuration
- Request building with auto-incrementing IDs
- HTTP request/response handling
- Timeout handling (15s delay test)
- Successful RPC calls with mock server
- RPC error handling
- Invalid endpoint handling
- Invalid JSON response handling
- All RPC methods with various parameters

**Test Results:**
```
✓ TestNewRPCClient
✓ TestRPCClient_BuildRequest
✓ TestRPCClient_AutoIncrementingID
✓ TestRPCClient_RequestTimeout
✓ TestRPCClient_Call_Success
✓ TestRPCClient_Call_RPCError
✓ TestRPCClient_Call_InvalidEndpoint
✓ TestRPCClient_Call_InvalidJSON
Total: 8 tests passing
```

### Integration with Wallet

The RPC client will be integrated into the wallet for:
- **Balance Refresh**: `RefreshBalances()` will use `GetBalance()` for all accounts
- **Transaction History**: History view will use `GetTransactionHistory()`
- **Transfer Operations**: Transfer flow will use `SendTransaction()`
- **Transaction Status**: Status tracking will use `GetTransactionStatus()`
- **Block Height Display**: Dashboard will use `GetBlockHeight()`

### Configuration

The RPC endpoint is configurable via:
1. Wallet configuration file (default: `http://localhost:8899`)
2. Command-line flag: `--rpc-url`
3. Environment variable: `POH_WALLET_RPC_ENDPOINT`

### Future Enhancements

Possible improvements:
1. **Connection Pooling**: Reuse HTTP connections for better performance
2. **Retry Logic**: Automatic retry with exponential backoff
3. **Batch Requests**: Support JSON-RPC batch requests for multiple queries
4. **WebSocket Support**: Real-time updates via WebSocket connection
5. **Request Caching**: Cache frequently accessed data (block height, account info)

## Summary

The multi-seed phrase architecture provides greater flexibility and security by allowing users to import and manage multiple independent seed phrases. Each seed phrase generates exactly one account at derivation index 0, making the wallet compatible with seed phrases from other sources while maintaining strong security boundaries between accounts.

The integrated RPC client provides a robust, type-safe interface for all blockchain operations, with comprehensive error handling and test coverage. This foundation enables the wallet to interact seamlessly with the blockchain for balance queries, transaction submission, and history retrieval.

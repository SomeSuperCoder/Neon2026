# Wallet RPC Client Reference

Complete reference for the wallet's JSON-RPC 2.0 client for blockchain communication.

## Overview

The wallet RPC client (`cmd/wallet/rpc/client.go`) provides a type-safe, easy-to-use interface for communicating with the PoH blockchain RPC node. It handles all JSON-RPC 2.0 protocol details, request/response marshaling, and error handling.

## Creating a Client

```go
import "github.com/poh-blockchain/cmd/wallet/rpc"

// Create client with default endpoint
client := rpc.NewRPCClient("http://localhost:8899")

// Create client with custom endpoint
client := rpc.NewRPCClient("https://mainnet.example.com:8899")
```

**Client Configuration:**
- Default timeout: 10 seconds
- Auto-incrementing request IDs
- HTTP/HTTPS support
- Automatic JSON-RPC 2.0 formatting

## Available Methods

### GetBalance

Retrieves the balance for a specific address.

```go
balance, err := client.GetBalance(address)
if err != nil {
    // Handle error
}
fmt.Printf("Balance: %d electrons\n", balance)
```

**Parameters:**
- `address` (string): Account address (64-character hex string)

**Returns:**
- `int64`: Balance in electrons (1 Neon = 1,000,000,000 electrons)
- `error`: Error if request fails or address is invalid

**Example:**
```go
address := "a1b2c3d4e5f6..."
balance, err := client.GetBalance(address)
if err != nil {
    log.Fatalf("Failed to get balance: %v", err)
}
fmt.Printf("Account balance: %d electrons\n", balance)
```

---

### GetAccountInfo

Retrieves full account information including balance, owner, data length, and executable status.

```go
info, err := client.GetAccountInfo(address)
if err != nil {
    // Handle error
}
if info == nil {
    fmt.Println("Account does not exist")
} else {
    fmt.Printf("Balance: %d\n", info.Balance)
    fmt.Printf("Owner: %s\n", info.Owner)
    fmt.Printf("Executable: %v\n", info.Executable)
}
```

**Parameters:**
- `address` (string): Account address

**Returns:**
- `*rpc.AccountInfo`: Account information, or `nil` if account doesn't exist
- `error`: Error if request fails

**AccountInfo Structure:**
```go
type AccountInfo struct {
    Address    string
    Balance    int64
    Owner      string
    DataLength int
    Executable bool
}
```

---

### GetTransactionHistory

Retrieves transaction history for an address with pagination support.

```go
history, err := client.GetTransactionHistory(address, 20)
if err != nil {
    // Handle error
}
for _, tx := range history {
    fmt.Printf("Tx: %s at block %d\n", tx.Signature, tx.BlockHeight)
}
```

**Parameters:**
- `address` (string): Account address
- `limit` (int): Maximum number of transactions to return (default: 20)

**Returns:**
- `[]rpc.TransactionRecord`: Array of transaction records in reverse chronological order
- `error`: Error if request fails

**TransactionRecord Structure:**
```go
type TransactionRecord struct {
    Signature    string
    BlockHeight  int64
    Slot         int64
    Timestamp    time.Time
    Success      bool
    Error        string
    Instructions []InstructionRecord
}
```

---

### SendTransaction

Submits a signed transaction to the blockchain.

```go
// txData is the marshaled transaction bytes
signature, err := client.SendTransaction(txData)
if err != nil {
    if rpcErr, ok := err.(*rpc.RPCError); ok {
        switch rpcErr.Code {
        case -32001:
            fmt.Println("Invalid signature")
        case -32003:
            fmt.Println("Insufficient balance")
        default:
            fmt.Printf("RPC error: %v\n", rpcErr)
        }
    }
    return
}
fmt.Printf("Transaction submitted: %s\n", signature)
```

**Parameters:**
- `txData` ([]byte): Marshaled transaction bytes

**Returns:**
- `string`: Transaction signature (hex-encoded)
- `error`: Error if submission fails

**Common Errors:**
- `-32001`: Invalid signature
- `-32002`: Malformed transaction
- `-32003`: Insufficient balance

---

### GetTransactionStatus

Checks the confirmation status of a transaction.

```go
status, err := client.GetTransactionStatus(signature)
if err != nil {
    // Handle error
}
if status.Confirmed {
    fmt.Printf("Confirmed at block %d\n", status.BlockHeight)
} else {
    fmt.Println("Transaction pending")
}
```

**Parameters:**
- `signature` (string): Transaction signature (hex-encoded)

**Returns:**
- `*rpc.TransactionStatus`: Transaction status
- `error`: Error if request fails

**TransactionStatus Structure:**
```go
type TransactionStatus struct {
    Signature   string
    Confirmed   bool
    BlockHeight int64
    Slot        int64
    Error       string
}
```

---

### GetBlockHeight

Retrieves the current blockchain height.

```go
height, err := client.GetBlockHeight()
if err != nil {
    // Handle error
}
fmt.Printf("Current block height: %d\n", height)
```

**Parameters:** None

**Returns:**
- `int64`: Current blockchain height
- `error`: Error if request fails

---

## Error Handling

The client uses a custom `RPCError` type for RPC-specific errors:

```go
type RPCError struct {
    Code    int
    Message string
    Data    interface{}
}
```

**Error Handling Pattern:**
```go
result, err := client.SomeMethod(params)
if err != nil {
    // Check if it's an RPC error
    if rpcErr, ok := err.(*rpc.RPCError); ok {
        fmt.Printf("RPC Error %d: %s\n", rpcErr.Code, rpcErr.Message)
        
        // Handle specific error codes
        switch rpcErr.Code {
        case -32700:
            // Parse error
        case -32600:
            // Invalid request
        case -32601:
            // Method not found
        case -32602:
            // Invalid params
        case -32001:
            // Invalid signature
        case -32002:
            // Malformed transaction
        case -32003:
            // Insufficient balance
        default:
            // Other error
        }
    } else {
        // Network or other error
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Standard Error Codes

### JSON-RPC 2.0 Standard Errors

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON received |
| -32600 | Invalid Request | Invalid JSON-RPC request |
| -32601 | Method not found | Method does not exist |
| -32602 | Invalid params | Invalid method parameters |
| -32603 | Internal error | Internal JSON-RPC error |

### Application-Specific Errors

| Code | Message | Description |
|------|---------|-------------|
| -32001 | Invalid signature | Transaction signature verification failed |
| -32002 | Malformed transaction | Transaction format is invalid |
| -32003 | Insufficient balance | Account has insufficient balance |
| -32004 | Account not found | Account does not exist |
| -32005 | Transaction not found | Transaction not found in ledger |

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/poh-blockchain/cmd/wallet/rpc"
)

func main() {
    // Create RPC client
    client := rpc.NewRPCClient("http://localhost:8899")
    
    // Get current block height
    height, err := client.GetBlockHeight()
    if err != nil {
        log.Fatalf("Failed to get block height: %v", err)
    }
    fmt.Printf("Current block height: %d\n", height)
    
    // Query account balance
    address := "a1b2c3d4e5f6..."
    balance, err := client.GetBalance(address)
    if err != nil {
        log.Fatalf("Failed to get balance: %v", err)
    }
    fmt.Printf("Account balance: %d electrons\n", balance)
    
    // Get account info
    info, err := client.GetAccountInfo(address)
    if err != nil {
        log.Fatalf("Failed to get account info: %v", err)
    }
    if info == nil {
        fmt.Println("Account does not exist")
    } else {
        fmt.Printf("Owner: %s\n", info.Owner)
        fmt.Printf("Executable: %v\n", info.Executable)
    }
    
    // Get transaction history
    history, err := client.GetTransactionHistory(address, 10)
    if err != nil {
        log.Fatalf("Failed to get transaction history: %v", err)
    }
    fmt.Printf("Found %d transactions\n", len(history))
    for _, tx := range history {
        fmt.Printf("  %s at block %d\n", tx.Signature, tx.BlockHeight)
    }
    
    // Send transaction (example with error handling)
    txData := []byte{...} // Your transaction bytes
    signature, err := client.SendTransaction(txData)
    if err != nil {
        if rpcErr, ok := err.(*rpc.RPCError); ok {
            switch rpcErr.Code {
            case -32001:
                fmt.Println("Transaction has invalid signature")
            case -32003:
                fmt.Println("Insufficient balance for transaction")
            default:
                fmt.Printf("RPC error: %v\n", rpcErr)
            }
        } else {
            fmt.Printf("Network error: %v\n", err)
        }
        return
    }
    fmt.Printf("Transaction submitted: %s\n", signature)
    
    // Check transaction status
    status, err := client.GetTransactionStatus(signature)
    if err != nil {
        log.Fatalf("Failed to get transaction status: %v", err)
    }
    if status.Confirmed {
        fmt.Printf("Transaction confirmed at block %d\n", status.BlockHeight)
    } else {
        fmt.Println("Transaction pending confirmation")
    }
}
```

## Testing

The RPC client includes comprehensive unit tests with mock server:

```bash
# Run RPC client tests
go test ./cmd/wallet/rpc/...

# Run with verbose output
go test -v ./cmd/wallet/rpc/...

# Run with coverage
go test -cover ./cmd/wallet/rpc/...
```

**Test Coverage:**
- Client creation and configuration
- Request building with auto-incrementing IDs
- HTTP request/response handling
- Timeout handling
- Successful RPC calls
- RPC error handling
- Invalid endpoint handling
- Invalid JSON response handling

## Configuration

The RPC endpoint can be configured via:

1. **Direct instantiation:**
   ```go
   client := rpc.NewRPCClient("http://localhost:8899")
   ```

2. **Wallet configuration:**
   ```go
   config := core.DefaultConfig()
   config.RPCEndpoint = "http://mainnet.example.com:8899"
   ```

3. **Command-line flag:**
   ```bash
   poh-wallet --rpc-url http://localhost:9000
   ```

4. **Environment variable:**
   ```bash
   export POH_WALLET_RPC_ENDPOINT="http://localhost:8899"
   ```

## Best Practices

1. **Reuse client instances** - Create one client and reuse it for multiple requests
2. **Handle all errors** - Always check for both RPC errors and network errors
3. **Use appropriate timeouts** - The default 10-second timeout is suitable for most cases
4. **Validate addresses** - Ensure addresses are valid before making requests
5. **Check for nil results** - Some methods return nil for non-existent resources
6. **Log errors appropriately** - Include context when logging errors

## Troubleshooting

### Connection Refused

**Problem:** `connection refused` error when calling RPC methods

**Solution:**
- Verify RPC node is running
- Check endpoint URL is correct
- Ensure firewall allows connections
- Try `curl http://localhost:8899` to test connectivity

### Timeout Errors

**Problem:** Requests timeout after 10 seconds

**Solution:**
- Check network connectivity
- Verify RPC node is responsive
- Consider increasing timeout if needed (requires modifying client)
- Check for slow queries (large transaction history)

### Invalid Signature Errors

**Problem:** `-32001` error when submitting transactions

**Solution:**
- Verify transaction is properly signed
- Check that signer's private key matches public key
- Ensure transaction data hasn't been modified after signing
- Verify signature format is correct

### Malformed Transaction Errors

**Problem:** `-32002` error when submitting transactions

**Solution:**
- Verify transaction is properly serialized
- Check transaction format matches expected structure
- Ensure all required fields are present
- Validate instruction data format

## See Also

- [RPC API Reference](rpc-api.md) - Complete RPC API documentation
- [Wallet Configuration](wallet-config.md) - Wallet configuration options
- [Transaction Builder](transaction-builder.md) - Building transactions
- [CLI Usage Guide](../guides/cli-usage.md) - Command-line interface


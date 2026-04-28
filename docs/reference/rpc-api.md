# RPC API Reference

Complete reference for the PoH Blockchain JSON-RPC 2.0 API.

## Overview

The RPC node provides a JSON-RPC 2.0 interface for querying blockchain data and submitting transactions. All requests and responses follow the JSON-RPC 2.0 specification.

## Starting the RPC Node

Start an RPC node using the CLI:

```bash
# Basic usage with required parameters
poh-blockchain rpc --ledger-path ./validator1.db --state-path ./validator1_state.db

# With custom port and bind address
poh-blockchain rpc \
  --ledger-path ./validator1.db \
  --state-path ./validator1_state.db \
  --rpc-port 9000 \
  --rpc-bind 0.0.0.0
```

**Parameters:**
- `--ledger-path` (required): Path to the blockchain ledger database
- `--state-path` (required): Path to the state database
- `--rpc-port` (optional): HTTP listening port (default: 8899)
- `--rpc-bind` (optional): Bind address (default: 127.0.0.1)

**Output:**
```
Starting RPC node...
  Ledger: ./validator1.db
  State: ./validator1_state.db
  Bind: 127.0.0.1:8899
Initializing ledger...
Initializing FileStore (read-only)...
Initializing transaction processor...
Creating RPC server...
Starting RPC server...
RPC server started successfully on http://127.0.0.1:8899
Press Ctrl+C to stop...
```

The RPC server will:
- Accept JSON-RPC 2.0 requests over HTTP
- Query blockchain data from the ledger
- Query account state from the FileStore (read-only mode)
- Process and submit transactions
- Support CORS for browser access

**Note:** The RPC node opens the FileStore in read-only mode, which allows multiple RPC nodes to safely query the same state database without conflicts. This is ideal for scaling read operations across multiple servers.

Press `Ctrl+C` to gracefully shutdown the server.

## Connection

Default endpoint: `http://localhost:8899`

All requests use HTTP POST with `Content-Type: application/json`.

## JSON-RPC 2.0 Protocol

### Request Format

```json
{
  "jsonrpc": "2.0",
  "method": "methodName",
  "params": { ... },
  "id": 1
}
```

### Response Format

**Success:**
```json
{
  "jsonrpc": "2.0",
  "result": { ... },
  "id": 1
}
```

**Error:**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": "Additional error details"
  },
  "id": 1
}
```

## Error Codes

### Standard JSON-RPC 2.0 Errors

| Code | Message | Description |
|------|---------|-------------|
| -32700 | Parse error | Invalid JSON was received |
| -32600 | Invalid Request | The JSON sent is not a valid Request object |
| -32601 | Method not found | The method does not exist |
| -32602 | Invalid params | Invalid method parameter(s) |
| -32603 | Internal error | Internal JSON-RPC error |

### Application-Specific Errors

| Code | Message | Description |
|------|---------|-------------|
| -32001 | Invalid signature | Transaction signature is invalid |
| -32002 | Malformed transaction | Transaction format is invalid |
| -32003 | Insufficient balance | Account has insufficient balance |
| -32004 | Account not found | Account does not exist |
| -32005 | Transaction not found | Transaction not found in ledger |
| -32006 | Network error | Network communication error |

## Methods

### getBalance

Get the balance of an account.

**Parameters:**
- `address` (string, required): Account address (hex-encoded)

**Returns:**
- `balance` (number): Account balance in electrons (1 Neon = 1,000,000,000 electrons)

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getBalance",
    "params": {
      "address": "a1b2c3d4e5f6..."
    },
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "balance": 1000000000
  },
  "id": 1
}
```

**Error Cases:**
- Returns `null` if account does not exist
- Returns error -32602 if address format is invalid

---

### getAccountInfo

Get full account information including balance, owner, data length, and executable status.

**Parameters:**
- `address` (string, required): Account address (hex-encoded)

**Returns:**
- `address` (string): Account address
- `balance` (number): Account balance in electrons
- `owner` (string): Program ID that owns this account
- `dataLength` (number): Size of account data in bytes
- `executable` (boolean): Whether the account contains executable code

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getAccountInfo",
    "params": {
      "address": "a1b2c3d4e5f6..."
    },
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "address": "a1b2c3d4e5f6...",
    "balance": 1000000000,
    "owner": "System_Program",
    "dataLength": 0,
    "executable": false
  },
  "id": 1
}
```

**Error Cases:**
- Returns `null` if account does not exist
- Returns error -32602 if address format is invalid

---

### getBlockHeight

Get the current block height of the blockchain.

**Parameters:** None

**Returns:**
- `blockHeight` (number): Current block height

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getBlockHeight",
    "params": {},
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "blockHeight": 12345
  },
  "id": 1
}
```

---

### getRecentBlockhash

Get the most recent blockhash for transaction construction.

**Parameters:** None

**Returns:**
- `blockhash` (string): Recent blockhash (hex-encoded)
- `blockHeight` (number): Block height of the blockhash

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getRecentBlockhash",
    "params": {},
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "blockhash": "a1b2c3d4e5f6...",
    "blockHeight": 12345
  },
  "id": 1
}
```

---

### getTransactionHistory

Get transaction history for an account with pagination support.

**Parameters:**
- `address` (string, required): Account address (hex-encoded)
- `limit` (number, optional): Maximum number of transactions to return (default: 20, max: 100)
- `before` (string, optional): Transaction signature to paginate before

**Returns:**
Array of transaction records, each containing:
- `signature` (string): Transaction signature (hex-encoded)
- `blockHeight` (number): Block height where transaction was confirmed
- `slot` (number): Slot number where transaction was confirmed
- `timestamp` (string): ISO 8601 timestamp
- `success` (boolean): Whether transaction succeeded
- `error` (string, optional): Error message if transaction failed
- `instructions` (array): Array of instruction records

**Instruction Record:**
- `programId` (string): Program ID that processed the instruction
- `type` (string): Instruction type (e.g., "Transfer", "CreateAccount")
- `accounts` (array): Array of account addresses involved
- `data` (string): Instruction data (hex-encoded)

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getTransactionHistory",
    "params": {
      "address": "a1b2c3d4e5f6...",
      "limit": 10
    },
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "signature": "tx123abc...",
      "blockHeight": 12345,
      "slot": 30863,
      "timestamp": "2026-04-27T16:30:00Z",
      "success": true,
      "instructions": [
        {
          "programId": "System_Program",
          "type": "Transfer",
          "accounts": ["a1b2c3d4...", "e5f6g7h8..."],
          "data": "0100000000000000..."
        }
      ]
    }
  ],
  "id": 1
}
```

**Error Cases:**
- Returns empty array if no transactions found
- Returns error -32602 if address format is invalid
- Returns error -32602 if limit exceeds maximum

---

### sendTransaction

Submit a signed transaction to the blockchain.

**Parameters:**
- `transaction` (string, required): Serialized and signed transaction (hex-encoded)

**Returns:**
- `signature` (string): Transaction signature (hex-encoded)

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "sendTransaction",
    "params": {
      "transaction": "01000000..."
    },
    "id": 1
  }'
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "signature": "tx123abc..."
  },
  "id": 1
}
```

**Error Cases:**
- Returns error -32001 if signature verification fails
- Returns error -32002 if transaction format is invalid
- Returns error -32003 if sender has insufficient balance
- Returns error -32603 for internal processing errors

---

### getTransactionStatus

Get the confirmation status of a transaction.

**Parameters:**
- `signature` (string, required): Transaction signature (hex-encoded)

**Returns:**
- `signature` (string): Transaction signature
- `confirmed` (boolean): Whether transaction is confirmed
- `blockHeight` (number, optional): Block height if confirmed
- `slot` (number, optional): Slot number if confirmed
- `error` (string, optional): Error message if transaction failed

**Example Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getTransactionStatus",
    "params": {
      "signature": "tx123abc..."
    },
    "id": 1
  }'
```

**Example Response (Confirmed):**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "signature": "tx123abc...",
    "confirmed": true,
    "blockHeight": 12345,
    "slot": 30863
  },
  "id": 1
}
```

**Example Response (Pending):**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "signature": "tx123abc...",
    "confirmed": false
  },
  "id": 1
}
```

**Error Cases:**
- Returns error -32005 if transaction not found
- Returns error -32602 if signature format is invalid

---

## Rate Limiting

The RPC server supports configurable connection limits to prevent resource exhaustion. Default limits:
- Maximum concurrent connections: 100
- Request timeout: 30 seconds
- Maximum request size: 1 MB

## CORS Support

The RPC server includes CORS headers for browser-based access:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: POST, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

## Best Practices

1. **Use batch requests** for multiple queries to reduce network overhead
2. **Cache blockhash** for transaction construction (valid for ~40 seconds)
3. **Poll transaction status** with exponential backoff
4. **Handle errors gracefully** with retry logic for network errors
5. **Validate addresses** before making requests to avoid unnecessary errors

## Example: Complete Transfer Flow

```bash
# 1. Get sender account info
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getAccountInfo",
    "params": {"address": "sender_address"},
    "id": 1
  }'

# 2. Get recent blockhash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getRecentBlockhash",
    "params": {},
    "id": 2
  }'

# 3. Build and sign transaction (client-side)
# ... transaction construction code ...

# 4. Submit transaction
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "sendTransaction",
    "params": {"transaction": "signed_tx_hex"},
    "id": 3
  }'

# 5. Check transaction status
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getTransactionStatus",
    "params": {"signature": "tx_signature"},
    "id": 4
  }'
```

## See Also

- [Transaction Builder API](transaction-builder.md) - Programmatic transaction construction
- [CLI Usage Guide](../guides/cli-usage.md) - Command-line interface
- [Cost Model](cost-model.md) - Understanding transaction fees

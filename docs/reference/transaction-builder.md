# Transaction Builder API Reference

The TransactionBuilder provides a convenient way to construct transactions with proper input declarations for instructions. This is part of the transaction input validation system that ensures all file accesses are explicitly declared before execution.

## Overview

The TransactionBuilder simplifies transaction creation by:
- Automatically declaring required file inputs with correct permissions
- Validating instruction parameters (e.g., positive amounts)
- Encoding instruction data in the correct format
- Managing signatures and transaction assembly

## Basic Usage

```go
import (
    "github.com/poh-blockchain/internal/transaction"
    "github.com/poh-blockchain/internal/filestore"
    "github.com/poh-blockchain/internal/genesis"
)

// Create a new builder
lastSeen := transaction.TxID{} // or get from ledger
builder := transaction.NewTransactionBuilder(lastSeen)

// Add a transfer instruction
err := builder.AddTransferInstruction(
    genesis.SystemProgramID,  // System Program
    senderID,                 // Sender account
    receiverID,               // Receiver account
    1000,                     // Amount to transfer
)
if err != nil {
    // Handle error (e.g., zero/negative amount)
}

// Add signature(s)
sig := transaction.Signature{
    PublicKey: senderPublicKey,
    Signature: signatureBytes,
}
builder.AddSignature(sig)

// Build the final transaction
tx, err := builder.Build()
if err != nil {
    // Handle error
}
```

## API Reference

### NewTransactionBuilder

```go
func NewTransactionBuilder(lastSeen TxID) *TransactionBuilder
```

Creates a new transaction builder with the specified lastSeen transaction ID.

**Parameters:**
- `lastSeen`: The TxID of the last seen transaction (used for transaction ordering)

**Returns:**
- `*TransactionBuilder`: A new builder instance

**Example:**
```go
builder := transaction.NewTransactionBuilder(transaction.TxID{})
```

### AddTransferInstruction

```go
func (tb *TransactionBuilder) AddTransferInstruction(
    systemProgramID filestore.FileID,
    senderID filestore.FileID,
    receiverID filestore.FileID,
    amount int64,
) error
```

Adds a System Program transfer instruction with proper input declarations.

**Input Declarations:**
The instruction automatically declares:
- System Program with `Read` permission
- Sender account with `Write` permission
- Receiver account with `Write` permission

**Parameters:**
- `systemProgramID`: FileID of the System Program (use `genesis.SystemProgramID`)
- `senderID`: FileID of the sender account
- `receiverID`: FileID of the receiver account
- `amount`: Amount to transfer (must be positive)

**Returns:**
- `error`: Error if amount is zero or negative, nil otherwise

**Instruction Data Format:**
```
[type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)]
Total: 73 bytes
```

**Example:**
```go
err := builder.AddTransferInstruction(
    genesis.SystemProgramID,
    filestore.FileID{0x01, 0x02, ...},
    filestore.FileID{0x03, 0x04, ...},
    5000,
)
if err != nil {
    log.Fatalf("Failed to add transfer: %v", err)
}
```

### AddSignature

```go
func (tb *TransactionBuilder) AddSignature(sig Signature) error
```

Adds a signature to the transaction. The first signature is treated as the fee payer.

**Parameters:**
- `sig`: Signature containing PublicKey and signature bytes

**Returns:**
- `error`: Always returns nil (reserved for future validation)

**Example:**
```go
sig := transaction.Signature{
    PublicKey: transaction.PublicKey{0x01, 0x02, ...},
    Signature: [64]byte{0xaa, 0xbb, ...},
}
builder.AddSignature(sig)
```

### Build

```go
func (tb *TransactionBuilder) Build() (*Transaction, error)
```

Constructs the final Transaction from the builder's state.

**Returns:**
- `*Transaction`: The constructed transaction
- `error`: Always returns nil (reserved for future validation)

**Example:**
```go
tx, err := builder.Build()
if err != nil {
    log.Fatalf("Failed to build transaction: %v", err)
}
```

## Complete Example

```go
package main

import (
    "crypto/ed25519"
    "log"
    
    "github.com/poh-blockchain/internal/transaction"
    "github.com/poh-blockchain/internal/filestore"
    "github.com/poh-blockchain/internal/genesis"
)

func createTransferTransaction(
    senderKeypair ed25519.PrivateKey,
    senderID filestore.FileID,
    receiverID filestore.FileID,
    amount int64,
) (*transaction.Transaction, error) {
    // Create builder
    builder := transaction.NewTransactionBuilder(transaction.TxID{})
    
    // Add transfer instruction
    err := builder.AddTransferInstruction(
        genesis.SystemProgramID,
        senderID,
        receiverID,
        amount,
    )
    if err != nil {
        return nil, err
    }
    
    // Build transaction to get bytes for signing
    tx, err := builder.Build()
    if err != nil {
        return nil, err
    }
    
    // Sign transaction
    txBytes := tx.Serialize() // Implement serialization
    signature := ed25519.Sign(senderKeypair, txBytes)
    
    // Create new builder with signature
    builder = transaction.NewTransactionBuilder(transaction.TxID{})
    builder.AddTransferInstruction(
        genesis.SystemProgramID,
        senderID,
        receiverID,
        amount,
    )
    
    // Add signature
    var pubKey transaction.PublicKey
    copy(pubKey[:], senderKeypair.Public().(ed25519.PublicKey))
    
    var sig [64]byte
    copy(sig[:], signature)
    
    builder.AddSignature(transaction.Signature{
        PublicKey: pubKey,
        Signature: sig,
    })
    
    // Build final transaction
    return builder.Build()
}
```

## Multi-Instruction Transactions

The builder supports adding multiple instructions to a single transaction:

```go
builder := transaction.NewTransactionBuilder(transaction.TxID{})

// First transfer
builder.AddTransferInstruction(
    genesis.SystemProgramID,
    account1,
    account2,
    1000,
)

// Second transfer
builder.AddTransferInstruction(
    genesis.SystemProgramID,
    account2,
    account3,
    500,
)

// Add signature and build
builder.AddSignature(sig)
tx, _ := builder.Build()
```

## Input Declaration System

The TransactionBuilder automatically handles input declarations, which are required for the transaction processor to validate file access before execution.

### Permission Types

- `Read`: Allows reading file data (required for programs)
- `Write`: Allows both reading and writing file data (required for accounts being modified)

### Standard Input Keys

For transfer instructions, the builder uses these standard keys:
- `"program"`: The System Program (Read permission)
- `"sender"`: The sender account (Write permission)
- `"receiver"`: The receiver account (Write permission)

### Input Validator

The `InputValidator` (in `internal/processor/input_validator.go`) validates instruction inputs before execution:

```go
import (
    "github.com/poh-blockchain/internal/processor"
    "github.com/poh-blockchain/internal/filestore"
)

// Create validator
fs, _ := filestore.NewFileStore("state.db")
validator := processor.NewInputValidator(fs)

// Validate instruction
err := validator.ValidateInstructionInputs(instruction)
if err != nil {
    // Handle validation error
}
```

**Validation Checks:**
1. Inputs map is not empty
2. Program is declared in inputs
3. All input files exist in the file store
4. All permissions are valid (Read=1 or Write=2)
5. Program has the Executable flag set

**Error Types:**
- `ErrMissingInput`: Required file not declared in inputs
- `ErrInvalidPermission`: Invalid permission value
- `ErrProgramNotExecutable`: Program lacks executable flag
- `ErrFileNotFound`: Input file doesn't exist

The transaction processor automatically calls the validator before executing each instruction, ensuring all file accesses are properly declared and authorized.

## Error Handling

### Amount Validation

```go
err := builder.AddTransferInstruction(
    genesis.SystemProgramID,
    sender,
    receiver,
    0, // Invalid: zero amount
)
// Returns: "amount must be positive, got 0"
```

### Future Validations

The API is designed to support additional validations:
- Duplicate instruction detection
- Signature verification
- Fee calculation
- Transaction size limits

## Testing

The TransactionBuilder includes comprehensive tests in `internal/transaction/builder_test.go`:

```bash
# Run builder tests
go test ./internal/transaction -run TestTransactionBuilder

# Run all transaction tests
go test ./internal/transaction -v
```

## Related Documentation

- [Transaction Input Validation Spec](.kiro/specs/transaction-input-validation/)
- [CLI Usage Guide](../guides/cli-usage.md)
- [File-Based State Model](.kiro/specs/file-based-state/)

## Implementation Status

- ✅ TransactionBuilder struct and constructor
- ✅ AddTransferInstruction with input declarations
- ✅ AddSignature for transaction signing
- ✅ Build method for transaction assembly
- ✅ Comprehensive test coverage
- ✅ System Program helper functions (CreateTransferInstruction, EncodeTransferData)
- ✅ Input validator implementation (InputValidator with comprehensive tests)
- 🚧 Input validator integration into TxProcessor
- 🚧 Update existing tests to use proper input declarations
- 🚧 CLI integration

See [tasks.md](.kiro/specs/transaction-input-validation/tasks.md) for the complete implementation plan.

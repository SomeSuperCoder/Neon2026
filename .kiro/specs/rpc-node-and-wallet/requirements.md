# Requirements Document

## Introduction

This document specifies the requirements for a production-ready RPC (Remote Procedure Call) node and a modern Terminal User Interface (TUI) multi-account wallet for the PoH blockchain. The RPC node will provide a JSON-RPC API for blockchain interaction, while the wallet will offer a feature-rich interface for managing accounts, viewing transaction history, and executing transfers with proper seed phrase management.

## Glossary

- **RPC_Node**: A blockchain node that exposes a JSON-RPC API for external clients to query blockchain state and submit transactions
- **TUI_Wallet**: A Terminal User Interface application that provides interactive wallet functionality in the terminal
- **Seed_Phrase**: A BIP39-compatible mnemonic phrase (12 or 24 words) used to derive cryptographic keypairs
- **Devnet**: The development network instance of the PoH blockchain used for testing
- **Account**: A blockchain entity identified by a public key with an associated balance and state
- **Transaction_History**: A chronological record of all transactions involving a specific account
- **Keypair**: A cryptographic key pair consisting of a public key (address) and private key
- **FileStore**: The blockchain's state storage system for accounts and program data

## Requirements

### Requirement 1: RPC Node Infrastructure

**User Story:** As a blockchain developer, I want a dedicated RPC node that exposes blockchain functionality via JSON-RPC, so that external applications can interact with the blockchain programmatically.

#### Acceptance Criteria

1. WHEN the RPC_Node starts, THE RPC_Node SHALL bind to a configurable HTTP port and accept JSON-RPC 2.0 requests
2. WHEN a client sends a malformed JSON-RPC request, THE RPC_Node SHALL return a JSON-RPC error response with error code -32700 for parse errors or -32600 for invalid requests
3. WHILE the RPC_Node is running, THE RPC_Node SHALL maintain a connection to the blockchain ledger and FileStore for state queries
4. THE RPC_Node SHALL process concurrent requests with a configurable maximum of 100 simultaneous connections
5. WHEN the RPC_Node receives a request, THE RPC_Node SHALL log the request method and response time with millisecond precision

### Requirement 2: RPC Query Methods

**User Story:** As a wallet developer, I want to query blockchain state through RPC methods, so that I can retrieve account information and transaction data.

#### Acceptance Criteria

1. WHEN a client calls "getBalance" with a valid public key, THE RPC_Node SHALL return the account balance in the smallest unit
2. WHEN a client calls "getAccountInfo" with a valid public key, THE RPC_Node SHALL return the account balance, owner program ID, and data length
3. WHEN a client calls "getTransactionHistory" with a valid public key and optional limit parameter, THE RPC_Node SHALL return up to the specified limit of transactions involving that account in reverse chronological order
4. WHEN a client calls "getBlockHeight", THE RPC_Node SHALL return the current blockchain height as an integer
5. WHEN a client calls "getRecentBlockhash", THE RPC_Node SHALL return the hash of the most recent block
6. WHEN a client queries a non-existent account, THE RPC_Node SHALL return a null result without error

### Requirement 3: RPC Transaction Submission

**User Story:** As a wallet application, I want to submit signed transactions through the RPC node, so that users can execute blockchain operations.

#### Acceptance Criteria

1. WHEN a client calls "sendTransaction" with a valid serialized signed transaction, THE RPC_Node SHALL submit the transaction to the blockchain and return the transaction signature
2. WHEN a client submits a transaction with invalid signatures, THE RPC_Node SHALL reject the transaction and return error code -32001 with message "Invalid signature"
3. WHEN a client calls "getTransactionStatus" with a transaction signature, THE RPC_Node SHALL return the confirmation status and block height if confirmed
4. THE RPC_Node SHALL validate transaction format before submission and return error code -32002 for malformed transactions
5. WHEN the RPC_Node successfully processes a transaction, THE RPC_Node SHALL return the transaction signature within 100 milliseconds

### Requirement 4: Seed Phrase Management

**User Story:** As a wallet user, I want to generate and restore accounts using BIP39 seed phrases, so that I can securely backup and recover my accounts.

#### Acceptance Criteria

1. WHEN the TUI_Wallet generates a new wallet, THE TUI_Wallet SHALL create a 12-word or 24-word BIP39-compliant Seed_Phrase
2. WHEN a user provides a valid Seed_Phrase for import, THE TUI_Wallet SHALL derive the keypair at index 0 using BIP44 derivation path m/44'/501'/0'/0'/0'
3. WHEN a user enters an invalid Seed_Phrase, THE TUI_Wallet SHALL display an error message "Invalid seed phrase" and prevent wallet creation or import
4. THE TUI_Wallet SHALL display the Seed_Phrase exactly once during wallet creation with a confirmation prompt
5. WHEN the TUI_Wallet stores wallet data, THE TUI_Wallet SHALL encrypt all seed phrases using AES-256-GCM with a user-provided password

### Requirement 5: Multi-Seed Phrase Management

**User Story:** As a wallet user, I want to import and manage multiple seed phrases in one wallet, so that I can organize my funds across different accounts from different sources.

#### Acceptance Criteria

1. WHEN a user imports a new Seed_Phrase, THE TUI_Wallet SHALL derive one keypair at index 0 using BIP44 derivation path m/44'/501'/0'/0'/0'
2. THE TUI_Wallet SHALL display all imported accounts with their public keys, balances, and optional user-defined labels
3. WHEN a user selects an account, THE TUI_Wallet SHALL set that account as the active account for transactions
4. THE TUI_Wallet SHALL support a minimum of 100 imported seed phrases per wallet
5. WHEN the TUI_Wallet queries account balances, THE TUI_Wallet SHALL refresh all account balances within 2 seconds
6. WHEN a user imports a Seed_Phrase that already exists in the wallet, THE TUI_Wallet SHALL display an error message "Seed phrase already imported" and prevent duplicate import

### Requirement 6: Transaction History Display

**User Story:** As a wallet user, I want to view my transaction history, so that I can track all incoming and outgoing transfers.

#### Acceptance Criteria

1. WHEN a user views transaction history for an account, THE TUI_Wallet SHALL display transactions in reverse chronological order with timestamp, amount, and counterparty address
2. THE TUI_Wallet SHALL indicate transaction direction with visual markers for incoming and outgoing transfers
3. WHEN a transaction involves a program invocation, THE TUI_Wallet SHALL display the program ID and instruction type
4. THE TUI_Wallet SHALL support pagination with a default page size of 20 transactions
5. WHEN the TUI_Wallet displays a transaction, THE TUI_Wallet SHALL show the transaction signature as a truncated hex string with copy functionality

### Requirement 7: Transfer Operations

**User Story:** As a wallet user, I want to send tokens to other addresses, so that I can transfer value on the blockchain.

#### Acceptance Criteria

1. WHEN a user initiates a transfer, THE TUI_Wallet SHALL prompt for recipient address, amount, and optional memo
2. WHEN a user confirms a transfer, THE TUI_Wallet SHALL display a confirmation screen showing sender, recipient, amount, and estimated cost
3. WHEN the TUI_Wallet submits a transfer, THE TUI_Wallet SHALL sign the transaction with the active account's private key and submit via RPC
4. IF a transfer fails due to insufficient balance, THEN THE TUI_Wallet SHALL display error message "Insufficient balance" and return to the transfer form
5. WHEN a transfer succeeds, THE TUI_Wallet SHALL display the transaction signature and update the account balance within 1 second

### Requirement 8: Modern TUI Interface Design

**User Story:** As a wallet user, I want a modern, polished terminal interface with smooth animations and rich visual elements similar to Neovim, tmux, and btop++, so that I have an enjoyable and professional user experience.

#### Acceptance Criteria

1. THE TUI_Wallet SHALL use a layout with bordered panels, a navigation sidebar with icons, main content area, and a status bar with real-time indicators
2. THE TUI_Wallet SHALL support keyboard navigation with arrow keys, tab, vim-style hjkl keys, and custom keybindings displayed in a help overlay
3. WHEN the TUI_Wallet displays data tables, THE TUI_Wallet SHALL render with Unicode box-drawing characters, alternating row colors, and sortable column headers with visual indicators
4. THE TUI_Wallet SHALL use a 256-color or true-color palette with gradient effects for progress bars, smooth color transitions for state changes, and syntax highlighting for addresses and transaction data
5. WHEN the terminal window is resized, THE TUI_Wallet SHALL dynamically adjust the layout with smooth transitions within 100 milliseconds
6. THE TUI_Wallet SHALL display animated spinners during loading operations and smooth fade-in effects for new content
7. THE TUI_Wallet SHALL render charts and graphs using Unicode block characters for balance history and transaction volume visualization
8. WHEN the TUI_Wallet shows notifications, THE TUI_Wallet SHALL display toast-style messages with icons that auto-dismiss after 3 seconds
9. THE TUI_Wallet SHALL use a consistent design system with rounded corners using Unicode characters, drop shadows using color gradients, and hover effects for interactive elements

### Requirement 9: Security and Authentication

**User Story:** As a wallet user, I want my wallet to be protected by a password, so that unauthorized users cannot access my funds.

#### Acceptance Criteria

1. WHEN the TUI_Wallet starts, THE TUI_Wallet SHALL prompt for a password before displaying wallet contents
2. WHEN a user enters an incorrect password three consecutive times, THE TUI_Wallet SHALL lock the wallet for 30 seconds
3. THE TUI_Wallet SHALL automatically lock after 5 minutes of inactivity
4. WHEN the TUI_Wallet stores sensitive data, THE TUI_Wallet SHALL use file permissions 0600 to restrict access to the owner
5. THE TUI_Wallet SHALL never display the Seed_Phrase or private keys in plain text after initial wallet creation

### Requirement 10: RPC Node Configuration and Devnet Integration

**User Story:** As a node operator, I want to configure the RPC node through command-line flags and have it integrated into the devnet script, so that I can easily run a complete development environment.

#### Acceptance Criteria

1. THE RPC_Node SHALL accept a --rpc-port flag to specify the HTTP listening port with default value 8899
2. THE RPC_Node SHALL accept a --rpc-bind flag to specify the bind address with default value 127.0.0.1
3. THE RPC_Node SHALL accept a --ledger-path flag to specify the SQLite database path
4. THE RPC_Node SHALL accept a --state-path flag to specify the FileStore database path
5. WHEN the RPC_Node starts with invalid configuration, THE RPC_Node SHALL log a descriptive error message and exit with code 1
6. WHEN the devnet.sh script is executed, THE devnet.sh script SHALL start an RPC_Node instance alongside the validator nodes
7. THE devnet.sh script SHALL configure the RPC_Node to connect to the leader node's ledger and state databases
8. WHEN the devnet.sh script stops, THE devnet.sh script SHALL terminate the RPC_Node process and clean up resources

### Requirement 11: Wallet Configuration and Persistence

**User Story:** As a wallet user, I want my wallet configuration and account data to persist between sessions, so that I don't lose my setup.

#### Acceptance Criteria

1. THE TUI_Wallet SHALL store wallet data in a file at ~/.poh-wallet/wallet.dat by default
2. THE TUI_Wallet SHALL accept a --wallet-path flag to specify an alternate wallet file location
3. THE TUI_Wallet SHALL accept a --rpc-url flag to specify the RPC node endpoint with default value http://localhost:8899
4. WHEN the TUI_Wallet starts and no wallet file exists, THE TUI_Wallet SHALL prompt the user to create a new wallet or restore from seed phrase
5. THE TUI_Wallet SHALL save account labels and preferences immediately after modification

### Requirement 12: Error Handling and Recovery

**User Story:** As a wallet user, I want clear error messages and recovery options when operations fail, so that I can understand and resolve issues.

#### Acceptance Criteria

1. WHEN the TUI_Wallet cannot connect to the RPC_Node, THE TUI_Wallet SHALL display error message "Cannot connect to RPC node at [URL]" and offer retry option
2. WHEN a transaction fails, THE TUI_Wallet SHALL display the error reason from the RPC response in user-friendly language
3. IF the wallet file is corrupted, THEN THE TUI_Wallet SHALL offer to restore from seed phrase backup
4. THE TUI_Wallet SHALL log all errors to ~/.poh-wallet/wallet.log with timestamp and stack trace
5. WHEN the RPC_Node encounters an internal error, THE RPC_Node SHALL return JSON-RPC error code -32603 with a descriptive message

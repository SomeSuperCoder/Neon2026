# Implementation Plan

- [x] 1. Set up RPC node infrastructure
  - Create `internal/rpc/` package directory structure
  - Define core types and interfaces for RPC server, handler, and query engine
  - Implement configuration structures for server settings
  - _Requirements: 1.1, 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ] 2. Implement JSON-RPC 2.0 protocol handler
  - [x] 2.1 Create JSON-RPC request/response structures
    - Define `JSONRPCRequest`, `JSONRPCResponse`, and `RPCError` types
    - Implement JSON marshaling/unmarshaling with proper error handling
    - Define standard error codes (ParseError, InvalidRequest, etc.)
    - _Requirements: 1.2, 2.6_
  
  - [x] 2.2 Implement RPC handler with method routing
    - Create `RPCHandler` struct with method registry
    - Implement `HandleRequest` method with method dispatch logic
    - Add request validation and error response generation
    - Implement logging for all requests with timing
    - Add comprehensive unit tests for handler functionality
    - _Requirements: 1.2, 1.5, 2.6_

- [x] 3. Implement query engine for blockchain data
  - [x] 3.1 Create query engine with ledger and filestore access
    - Implement `QueryEngine` struct with database connections
    - Add query caching layer for frequently accessed data
    - Implement helper methods for data conversion
    - _Requirements: 1.3, 2.1, 2.2, 2.3_
  
  - [x] 3.2 Implement account query methods
    - Implement `GetBalance` method with address validation
    - Implement `GetAccountInfo` method returning full account details
    - Handle non-existent accounts gracefully (return null)
    - _Requirements: 2.1, 2.2, 2.6_
  
  - [x] 3.3 Implement blockchain query methods
    - Implement `GetBlockHeight` method
    - Implement `GetRecentBlockhash` method
    - Add caching for block height and recent blockhash
    - _Requirements: 2.4, 2.5_
  
  - [x] 3.4 Implement transaction history query
    - Implement `GetTransactionHistory` with pagination support
    - Parse transaction data from ledger blocks
    - Return transactions in reverse chronological order
    - Support configurable limit parameter (default 20)
    - _Requirements: 2.3, 6.1, 6.2, 6.3, 6.4, 6.5_

- [x] 4. Implement transaction submission endpoint
  - [x] 4.1 Create transaction submission handler
    - Implement `sendTransaction` RPC method
    - Deserialize and validate transaction format
    - Verify transaction signatures before submission
    - Return transaction signature on success
    - _Requirements: 3.1, 3.2, 3.4, 3.5_
  
  - [x] 4.2 Implement transaction status query
    - Implement `GetTransactionStatus` method
    - Query ledger for transaction confirmation
    - Return status with block height if confirmed
    - _Requirements: 3.3_

- [x] 5. Implement HTTP server with concurrent request handling
  - [x] 5.1 Create HTTP server with configuration
    - Implement `RPCServer` struct with http.Server
    - Add configurable bind address, port, and connection limits
    - Implement graceful shutdown handling
    - _Requirements: 1.1, 1.4, 10.1, 10.2_
  
  - [x] 5.2 Implement HTTP request handler
    - Implement `ServeHTTP` method for JSON-RPC over HTTP
    - Add request size validation (1MB limit)
    - Handle CORS headers for browser access
    - Add request/response logging with timing
    - _Requirements: 1.1, 1.2, 1.5_

- [x] 6. Add RPC command to main CLI
  - [x] 6.1 Implement RPC subcommand handler
    - Add "rpc" case to main command dispatcher
    - Parse RPC-specific flags (--rpc-port, --rpc-bind, --ledger-path, --state-path)
    - Initialize ledger and filestore connections
    - Start RPC server with configuration
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_
  
  - [x] 6.2 Add graceful shutdown for RPC node
    - Handle SIGINT and SIGTERM signals
    - Close database connections cleanly
    - Stop HTTP server gracefully
    - _Requirements: 1.1_

- [x] 7. Integrate RPC node into devnet script
  - [x] 7.1 Add RPC node startup to devnet.sh
    - Add `start_rpc_node` function after leader startup
    - Configure RPC to use leader's ledger and state databases
    - Add RPC node PID tracking
    - Update network info display to show RPC endpoint
    - _Requirements: 10.6, 10.7_
  
  - [x] 7.2 Add RPC node to devnet stop and status commands
    - Update `stop_devnet` to terminate RPC node
    - Update `show_devnet_status` to display RPC node status
    - Add RPC logs to `show_devnet_logs`
    - _Requirements: 10.8_

- [x] 8. Implement wallet core with BIP39/44 support
  - [x] 8.1 Create wallet package structure
    - Create `cmd/wallet/` directory with core, rpc, and ui subdirectories
    - Define wallet configuration structures
    - Implement wallet file format with versioning
    - _Requirements: 4.1, 4.2, 11.1, 11.2_
  
  - [x] 8.2 Implement BIP39 seed phrase generation
    - Integrate go-bip39 library for mnemonic generation
    - Implement 12-word and 24-word seed phrase generation
    - Add seed phrase validation
    - Implement mnemonic to seed conversion
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [x] 8.3 Implement BIP44 key derivation for Ed25519
    - Implement SLIP-0010 Ed25519 key derivation
    - Use derivation path m/44'/501'/0'/0'/index'
    - Implement HDKey structure with chain code
    - Convert HD keys to Ed25519 keypairs
    - _Requirements: 4.2, 5.1, 5.2_
  
  - [x] 8.4 Implement wallet encryption and persistence
    - Implement AES-256-GCM encryption for wallet data
    - Use PBKDF2 for password-based key derivation (100k iterations)
    - Implement wallet save/load with encryption
    - Set file permissions to 0600 for security
    - _Requirements: 4.5, 9.1, 9.4, 11.1, 11.5_

- [x] 9. Implement multi-seed phrase management
  - [x] 9.1 Implement seed phrase import and storage
    - Implement `ImportSeedPhrase` method with duplicate detection
    - Store seed phrase metadata (index, label, balance)
    - Support minimum of 100 imported seed phrases per wallet
    - Derive one account at index 0 per seed phrase
    - _Requirements: 5.1, 5.2, 5.4, 5.6_
  
  - [x] 9.2 Implement account selection and balance refresh
    - Implement active account tracking
    - Implement `RefreshBalances` method using RPC client (placeholder)
    - Update all account balances within 2 seconds
    - Add `GetAccountBySeedPhraseIndex` method
    - _Requirements: 5.3, 5.5_

- [x] 10. Implement RPC client for wallet
  - [x] 10.1 Create RPC client with HTTP transport
    - Implement `RPCClient` struct with http.Client
    - Implement JSON-RPC request building with auto-incrementing IDs
    - Add request timeout handling (10 seconds)
    - Add comprehensive unit tests for client functionality
    - _Requirements: 11.3, 12.1_
  
  - [x] 10.2 Implement RPC client methods
    - Implement `GetBalance` method
    - Implement `GetAccountInfo` method
    - Implement `GetTransactionHistory` method
    - Implement `SendTransaction` method
    - Implement `GetTransactionStatus` method
    - Implement `GetBlockHeight` method
    - Add unit tests for all RPC methods with mock server
    - Test error handling (timeouts, invalid responses, RPC errors)
    - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.3_

- [x] 11. Implement transaction building and signing in wallet
  - [x] 11.1 Implement transfer transaction builder
    - Create `TransferRequest` structure
    - Implement `BuildTransferTransaction` using transaction.Builder
    - Add transfer instruction with proper input declarations
    - Sign transaction with active account's private key
    - _Requirements: 7.1, 7.2, 7.3_
  
  - [x] 11.2 Implement transaction submission flow
    - Serialize signed transaction for RPC submission
    - Handle submission errors with user-friendly messages
    - Update account balance after successful transfer
    - Display transaction signature to user
    - _Requirements: 7.3, 7.4, 7.5_

- [ ] 12. Implement TUI framework with Bubble Tea
  - [ ] 12.1 Set up Bubble Tea application structure
    - Create main `Model` struct with view management
    - Implement `Init`, `Update`, and `View` methods
    - Add view type enumeration and view registry
    - Implement terminal resize handling
    - _Requirements: 8.1, 8.5_
  
  - [ ] 12.2 Implement navigation and keyboard controls
    - Add keyboard navigation (arrow keys, hjkl, tab)
    - Implement view switching logic
    - Add help overlay with keybindings
    - _Requirements: 8.2_
  
  - [ ] 12.3 Define color scheme and styling
    - Create Lipgloss style definitions for all UI elements
    - Define neon-inspired color palette (cyan, magenta, green)
    - Create reusable style components (borders, panels, buttons)
    - _Requirements: 8.4_

- [ ] 13. Implement dashboard view
  - [ ] 13.1 Create dashboard layout
    - Implement `DashboardView` struct and interface methods
    - Display total balance across all accounts
    - Show current block height
    - Add loading spinner for data refresh
    - _Requirements: 8.1, 8.4_
  
  - [ ] 13.2 Add recent transactions display
    - Fetch and display recent transactions from active account
    - Show transaction direction (incoming/outgoing) with visual markers
    - Display transaction amounts and counterparty addresses
    - _Requirements: 6.1, 6.2_

- [ ] 14. Implement accounts view
  - [ ] 14.1 Create accounts table
    - Implement `AccountsView` with table component
    - Display all accounts with address, balance, and label
    - Add account selection with keyboard navigation
    - Support table sorting by balance
    - _Requirements: 5.2, 5.3, 8.3_
  
  - [ ] 14.2 Add account management actions
    - Implement "Add Account" action to derive new account
    - Implement "Set Label" action for account naming
    - Implement "Set Active" action to change active account
    - _Requirements: 5.1, 5.2, 5.3_

- [ ] 15. Implement transfer view
  - [ ] 15.1 Create transfer form
    - Implement `TransferView` with text input components
    - Add recipient address input with validation
    - Add amount input with balance validation
    - Add optional memo input
    - Implement focus management between inputs
    - _Requirements: 7.1_
  
  - [ ] 15.2 Implement transfer confirmation screen
    - Display confirmation with sender, recipient, amount, and estimated cost
    - Add "Confirm" and "Cancel" actions
    - Show loading state during submission
    - _Requirements: 7.2_
  
  - [ ] 15.3 Implement transfer result display
    - Show success message with transaction signature
    - Display error messages for failed transfers (insufficient balance, etc.)
    - Add "Copy Signature" action
    - Return to transfer form after completion
    - _Requirements: 7.4, 7.5_

- [ ] 16. Implement transaction history view
  - [ ] 16.1 Create transaction history table
    - Implement `HistoryView` with table and paginator components
    - Display transactions in reverse chronological order
    - Show timestamp, amount, counterparty, and direction
    - Add loading spinner for data fetch
    - _Requirements: 6.1, 6.2, 6.4_
  
  - [ ] 16.2 Add transaction details display
    - Show full transaction signature with copy functionality
    - Display program ID and instruction type for program invocations
    - Add pagination controls (20 transactions per page)
    - _Requirements: 6.3, 6.4, 6.5_

- [ ] 17. Implement wallet security features
  - [ ] 17.1 Implement password authentication
    - Add password prompt on wallet startup
    - Implement password verification with encrypted wallet
    - Add failed attempt tracking (lock after 3 failures for 30 seconds)
    - _Requirements: 9.1, 9.2_
  
  - [ ] 17.2 Implement auto-lock functionality
    - Track last user activity timestamp
    - Implement auto-lock after 5 minutes of inactivity
    - Clear sensitive data from memory on lock
    - Require password to unlock
    - _Requirements: 9.3_
  
  - [ ] 17.3 Implement secure seed phrase display
    - Display seed phrase only once during wallet creation
    - Require user confirmation before proceeding
    - Never display seed phrase or private keys after creation
    - _Requirements: 4.4, 9.5_

- [ ] 18. Implement wallet initialization flow
  - [ ] 18.1 Create wallet creation wizard
    - Implement "Create New Wallet" flow
    - Generate seed phrase with word count selection (12 or 24)
    - Display seed phrase with confirmation prompt
    - Collect password for encryption
    - Save encrypted wallet to file
    - _Requirements: 4.1, 4.4, 11.4_
  
  - [ ] 18.2 Create wallet restoration flow
    - Implement "Restore from Seed Phrase" flow
    - Prompt for seed phrase input
    - Validate seed phrase format
    - Collect password for encryption
    - Derive initial account and save wallet
    - _Requirements: 4.2, 4.3, 11.4_

- [ ] 19. Implement settings view
  - [ ] 19.1 Create settings interface
    - Display current RPC endpoint
    - Show auto-lock timeout setting
    - Display wallet file path
    - Add theme selection (if multiple themes)
    - _Requirements: 11.2, 11.3_
  
  - [ ] 19.2 Implement settings persistence
    - Save settings changes to wallet configuration
    - Update wallet file immediately on modification
    - Validate RPC endpoint URL format
    - _Requirements: 11.5_

- [ ] 20. Implement error handling and recovery
  - [ ] 20.1 Add RPC connection error handling
    - Detect RPC connection failures
    - Display user-friendly error messages
    - Implement retry mechanism with exponential backoff
    - Show offline mode indicator in UI
    - _Requirements: 12.1_
  
  - [ ] 20.2 Add transaction error handling
    - Parse RPC error responses
    - Display user-friendly error messages for common errors
    - Handle insufficient balance errors gracefully
    - _Requirements: 12.2_
  
  - [ ] 20.3 Implement wallet recovery
    - Detect corrupted wallet files
    - Offer seed phrase restoration option
    - Implement error logging to ~/.poh-wallet/wallet.log
    - _Requirements: 12.3, 12.4_

- [ ] 21. Add UI polish and animations
  - [ ] 21.1 Implement loading states and spinners
    - Add animated spinners for all async operations
    - Show progress indicators for balance refresh
    - Display loading overlays for transaction submission
    - _Requirements: 8.6_
  
  - [ ] 21.2 Add visual effects and transitions
    - Implement smooth fade-in effects for new content
    - Add color transitions for state changes
    - Implement toast notifications with auto-dismiss
    - Add gradient effects for progress bars
    - _Requirements: 8.4, 8.6, 8.8_
  
  - [ ] 21.3 Implement charts and visualizations
    - Add balance history chart using Unicode block characters
    - Display transaction volume visualization
    - Add network status indicators
    - _Requirements: 8.7_
  
  - [ ] 21.4 Polish UI layout and borders
    - Use Unicode box-drawing characters for borders
    - Implement rounded corners using Unicode characters
    - Add drop shadows using color gradients
    - Implement hover effects for interactive elements
    - _Requirements: 8.1, 8.9_

- [ ] 22. Create wallet main entry point
  - [ ] 22.1 Implement wallet CLI
    - Create `cmd/wallet/main.go` with flag parsing
    - Add --wallet-path flag for custom wallet location
    - Add --rpc-url flag for RPC endpoint configuration
    - Implement wallet initialization check
    - _Requirements: 11.2, 11.3, 11.4_
  
  - [ ] 22.2 Wire up all components
    - Initialize wallet core with configuration
    - Create RPC client with endpoint
    - Initialize Bubble Tea application with all views
    - Start TUI event loop
    - _Requirements: 8.1_

- [ ] 23. Write comprehensive tests
  - [ ] 23.1 Write RPC node unit tests
    - Test JSON-RPC request/response parsing
    - Test method routing and error handling
    - Test query engine with mock databases
    - Test concurrent request handling
    - _Requirements: 1.1, 1.2, 1.3, 1.4_
  
  - [ ] 23.2 Write RPC integration tests
    - Test all RPC methods with real ledger and filestore
    - Test transaction submission flow
    - Test error scenarios (invalid addresses, missing data)
    - Test RPC server lifecycle
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3_
  
  - [ ] 23.3 Write wallet core unit tests
    - Test seed phrase generation and validation
    - Test BIP44 key derivation
    - Test wallet encryption/decryption
    - Test account derivation and management
    - _Requirements: 4.1, 4.2, 4.3, 4.5, 5.1, 5.2_
  
  - [ ] 23.4 Write wallet integration tests
    - Test full wallet creation flow
    - Test wallet restoration from seed phrase
    - Test transaction building and signing
    - Test RPC client with mock server
    - _Requirements: 7.1, 7.2, 7.3, 11.4_
  
  - [ ] 23.5 Write end-to-end tests
    - Start devnet with RPC node
    - Create wallet and generate accounts
    - Execute transfers and verify balances
    - Query transaction history
    - Test wallet lock/unlock
    - _Requirements: All requirements_

- [ ] 24. Create documentation
  - [ ] 24.1 Write RPC API documentation
    - Document all RPC methods with parameters and return types
    - Provide example requests and responses
    - Document error codes and messages
    - Add usage examples with curl
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3_
  
  - [ ] 24.2 Write wallet user guide
    - Document wallet installation and setup
    - Provide wallet creation and restoration instructions
    - Document all wallet features and keyboard shortcuts
    - Add security best practices
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 7.1, 7.2, 8.2, 9.1, 9.3_
  
  - [ ] 24.3 Update devnet documentation
    - Document RPC node integration in devnet.sh
    - Provide examples of using RPC with devnet
    - Document wallet usage with devnet
    - _Requirements: 10.6, 10.7, 10.8_

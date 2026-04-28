# Design Document: RPC Node and TUI Wallet

## Overview

This design specifies a production-ready JSON-RPC node and modern Terminal User Interface (TUI) wallet for the PoH blockchain. The RPC node provides a standardized API for blockchain interaction, while the TUI wallet offers a feature-rich, visually appealing interface for account management, transaction execution, and balance monitoring.

The system consists of two main components:
1. **RPC Node**: A read-only query node that exposes blockchain state via JSON-RPC 2.0
2. **TUI Wallet**: A standalone terminal application that communicates with the RPC node

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        TUI Wallet                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   UI Layer   │  │ Wallet Core  │  │  RPC Client  │      │
│  │  (bubbletea) │◄─┤  (BIP39/44)  │◄─┤  (HTTP/JSON) │      │
│  └──────────────┘  └──────────────┘  └──────┬───────┘      │
└─────────────────────────────────────────────┼──────────────┘
                                               │ JSON-RPC 2.0
                                               │ (HTTP)
┌──────────────────────────────────────────────┼──────────────┐
│                      RPC Node                │               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────▼───────┐      │
│  │ HTTP Server  │◄─┤  RPC Handler │◄─┤ Query Engine │      │
│  │ (net/http)   │  │  (JSON-RPC)  │  │              │      │
│  └──────────────┘  └──────────────┘  └──────┬───────┘      │
│                                              │               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────▼───────┐      │
│  │   Ledger     │  │  FileStore   │  │ Tx Processor │      │
│  │  (SQLite)    │  │  (BadgerDB)  │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                                               │
                                               ▼
                                    ┌──────────────────┐
                                    │  Validator Node  │
                                    │   (Leader/Rep)   │
                                    └──────────────────┘
```

### Component Interaction Flow

1. **Query Flow**: TUI Wallet → RPC Client → HTTP → RPC Handler → Query Engine → Ledger/FileStore → Response
2. **Transaction Flow**: TUI Wallet → Sign Locally → RPC Client → HTTP → RPC Handler → Tx Processor → Validator Network
3. **Devnet Integration**: devnet.sh → Start RPC Node → Connect to Leader's Databases

## Components and Interfaces

### 1. RPC Node

#### 1.1 HTTP Server (`internal/rpc/server.go`)

```go
type RPCServer struct {
    httpServer   *http.Server
    handler      *RPCHandler
    config       *ServerConfig
    logger       *log.Logger
}

type ServerConfig struct {
    BindAddress  string  // Default: "127.0.0.1"
    Port         int     // Default: 8899
    MaxConns     int     // Default: 100
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

func NewRPCServer(config *ServerConfig, ledger *storage.Ledger, fileStore *filestore.FileStore, txProcessor *processor.TxProcessor) *RPCServer
func (s *RPCServer) Start() error
func (s *RPCServer) Stop() error
func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

#### 1.2 JSON-RPC Handler (`internal/rpc/handler.go`)

```go
type RPCHandler struct {
    queryEngine  *QueryEngine
    txProcessor  *processor.TxProcessor
    logger       *log.Logger
}

type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params"`
    ID      interface{}     `json:"id"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
    ID      interface{} `json:"id"`
}

type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
    InvalidSignature = -32001
    MalformedTransaction = -32002
)

func NewRPCHandler(queryEngine *QueryEngine, txProcessor *processor.TxProcessor) *RPCHandler
func (h *RPCHandler) HandleRequest(req *JSONRPCRequest) *JSONRPCResponse
```

#### 1.3 Query Engine (`internal/rpc/query.go`)

```go
type QueryEngine struct {
    ledger    *storage.Ledger
    fileStore *filestore.FileStore
    cache     *QueryCache
}

type AccountInfo struct {
    Address    string `json:"address"`
    Balance    int64  `json:"balance"`
    Owner      string `json:"owner"`
    DataLength int    `json:"dataLength"`
    Executable bool   `json:"executable"`
}

type TransactionRecord struct {
    Signature   string    `json:"signature"`
    BlockHeight int64     `json:"blockHeight"`
    Slot        int64     `json:"slot"`
    Timestamp   time.Time `json:"timestamp"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
    Instructions []InstructionRecord `json:"instructions"`
}

type InstructionRecord struct {
    ProgramID string   `json:"programId"`
    Type      string   `json:"type"`
    Accounts  []string `json:"accounts"`
    Data      string   `json:"data"`
}

func NewQueryEngine(ledger *storage.Ledger, fileStore *filestore.FileStore) *QueryEngine
func (q *QueryEngine) GetBalance(address string) (int64, error)
func (q *QueryEngine) GetAccountInfo(address string) (*AccountInfo, error)
func (q *QueryEngine) GetTransactionHistory(address string, limit int) ([]TransactionRecord, error)
func (q *QueryEngine) GetBlockHeight() (int64, error)
func (q *QueryEngine) GetRecentBlockhash() (string, error)
func (q *QueryEngine) GetTransactionStatus(signature string) (*TransactionStatus, error)
```

#### 1.4 RPC Methods

| Method | Parameters | Returns | Description |
|--------|-----------|---------|-------------|
| `getBalance` | `address: string` | `int64` | Get account balance |
| `getAccountInfo` | `address: string` | `AccountInfo` | Get full account details |
| `getTransactionHistory` | `address: string, limit?: int` | `TransactionRecord[]` | Get transaction history (default limit: 20) |
| `getBlockHeight` | - | `int64` | Get current blockchain height |
| `getRecentBlockhash` | - | `string` | Get most recent block hash |
| `sendTransaction` | `transaction: string` | `string` | Submit signed transaction, return signature |
| `getTransactionStatus` | `signature: string` | `TransactionStatus` | Get transaction confirmation status |

### 2. TUI Wallet

#### 2.1 Technology Stack

- **UI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Modern TUI framework with Elm architecture
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles) - Reusable TUI components
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions and layouts
- **BIP39**: [go-bip39](https://github.com/tyler-smith/go-bip39) - Mnemonic seed phrase generation
- **BIP32/44**: Custom implementation for Ed25519 key derivation

#### 2.2 Wallet Core (`cmd/wallet/core/wallet.go`)

```go
type Wallet struct {
    seedPhrases  []string           // Multiple imported seed phrases
    accounts     []*Account         // One account per seed phrase
    activeIndex  int
    config       *WalletConfig
    encrypted    bool
}

type Account struct {
    SeedPhraseIndex int              // Index into seedPhrases array
    PublicKey       [32]byte
    PrivateKey      ed25519.PrivateKey
    Address         string
    Label           string
    Balance         int64
    LastUpdate      time.Time
}

type WalletConfig struct {
    RPCEndpoint string
    WalletPath  string
    AutoLock    time.Duration
}

func NewWallet(config *WalletConfig) (*Wallet, error)
func GenerateSeedPhrase(wordCount int) (string, error)
func (w *Wallet) ImportSeedPhrase(seedPhrase string, label string) (*Account, error)
func (w *Wallet) GetActiveAccount() *Account
func (w *Wallet) SetActiveAccount(index int) error
func (w *Wallet) RefreshBalances() error
func (w *Wallet) Save(password string) error
func LoadWallet(path string, password string) (*Wallet, error)
```

#### 2.3 RPC Client (`cmd/wallet/rpc/client.go`)

```go
type RPCClient struct {
    endpoint   string
    httpClient *http.Client
    requestID  uint64
}

func NewRPCClient(endpoint string) *RPCClient
func (c *RPCClient) GetBalance(address string) (int64, error)
func (c *RPCClient) GetAccountInfo(address string) (*AccountInfo, error)
func (c *RPCClient) GetTransactionHistory(address string, limit int) ([]TransactionRecord, error)
func (c *RPCClient) SendTransaction(tx *transaction.Transaction) (string, error)
func (c *RPCClient) GetTransactionStatus(signature string) (*TransactionStatus, error)
func (c *RPCClient) GetBlockHeight() (int64, error)
```

#### 2.4 UI Architecture (`cmd/wallet/ui/`)

The TUI uses the Bubble Tea framework with a model-update-view pattern:

```go
// Main application model
type Model struct {
    wallet      *core.Wallet
    rpcClient   *rpc.RPCClient
    currentView ViewType
    views       map[ViewType]ViewModel
    width       int
    height      int
    locked      bool
    lastActivity time.Time
}

type ViewType int
const (
    ViewDashboard ViewType = iota
    ViewAccounts
    ViewTransfer
    ViewHistory
    ViewSettings
)

// View interface
type ViewModel interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (ViewModel, tea.Cmd)
    View() string
}

// Dashboard view
type DashboardView struct {
    wallet      *core.Wallet
    totalBalance int64
    recentTxs   []TransactionRecord
    blockHeight int64
    spinner     spinner.Model
}

// Accounts view
type AccountsView struct {
    wallet   *core.Wallet
    table    table.Model
    selected int
}

// Transfer view
type TransferView struct {
    wallet       *core.Wallet
    rpcClient    *rpc.RPCClient
    recipientInput textinput.Model
    amountInput    textinput.Model
    memoInput      textinput.Model
    focusIndex     int
    confirmScreen  bool
    result         *TransferResult
}

// History view
type HistoryView struct {
    wallet    *core.Wallet
    rpcClient *rpc.RPCClient
    table     table.Model
    paginator paginator.Model
    loading   bool
}
```

#### 2.5 UI Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ ⚡ Neon Wallet                                    🔒 Unlocked │ 16:42 │
├──────────────┬──────────────────────────────────────────────────────┤
│              │                                                       │
│  📊 Dashboard│  Total Balance: 1,234,567 Neon                       │
│  💼 Accounts │                                                       │
│  📤 Transfer │  ┌─────────────────────────────────────────────┐    │
│  📜 History  │  │ Recent Transactions                          │    │
│  ⚙️  Settings │  ├─────────────────────────────────────────────┤    │
│              │  │ ↓ Received 1000 from 0x1234...abcd          │    │
│              │  │ ↑ Sent 500 to 0x5678...ef01                 │    │
│              │  │ ↓ Received 2500 from 0x9abc...def2          │    │
│              │  └─────────────────────────────────────────────┘    │
│              │                                                       │
│              │  Block Height: 12,345                                │
│              │  Network: Devnet                                     │
│              │                                                       │
├──────────────┴──────────────────────────────────────────────────────┤
│ ↑/k up • ↓/j down • ←/h back • →/l select • q quit • ? help        │
└─────────────────────────────────────────────────────────────────────┘
```

#### 2.6 Color Scheme

Using a modern, professional color palette:

```go
var (
    // Primary colors
    ColorPrimary   = lipgloss.Color("#00D9FF")  // Neon cyan
    ColorSecondary = lipgloss.Color("#FF00FF")  // Neon magenta
    ColorAccent    = lipgloss.Color("#00FF9F")  // Neon green
    
    // Status colors
    ColorSuccess   = lipgloss.Color("#00FF9F")  // Green
    ColorError     = lipgloss.Color("#FF0055")  // Red
    ColorWarning   = lipgloss.Color("#FFD700")  // Gold
    ColorInfo      = lipgloss.Color("#00D9FF")  // Cyan
    
    // UI colors
    ColorBorder    = lipgloss.Color("#3A3A3A")  // Dark gray
    ColorText      = lipgloss.Color("#E0E0E0")  // Light gray
    ColorTextDim   = lipgloss.Color("#808080")  // Medium gray
    ColorBg        = lipgloss.Color("#1A1A1A")  // Very dark gray
    ColorBgAlt     = lipgloss.Color("#2A2A2A")  // Dark gray
)
```

### 3. Cryptography and Key Derivation

#### 3.1 BIP39 Seed Phrase Generation

```go
// Generate 12 or 24 word mnemonic
func GenerateMnemonic(wordCount int) (string, error) {
    entropy, err := bip39.NewEntropy(wordCount * 11 / 3 * 4)
    if err != nil {
        return "", err
    }
    return bip39.NewMnemonic(entropy)
}

// Validate mnemonic
func ValidateMnemonic(mnemonic string) bool {
    return bip39.IsMnemonicValid(mnemonic)
}

// Convert mnemonic to seed
func MnemonicToSeed(mnemonic string, password string) []byte {
    return bip39.NewSeed(mnemonic, password)
}
```

#### 3.2 BIP44 Key Derivation for Ed25519

Since Ed25519 doesn't have standard BIP32 support, we'll use SLIP-0010 for Ed25519 key derivation:

```go
// Derivation path: m/44'/501'/0'/0'/0' (fixed at index 0 for each seed phrase)
// 501 is Solana's coin type (we'll use the same for compatibility)

type HDKey struct {
    Key       []byte
    ChainCode []byte
}

func DeriveKey(seed []byte, path string) (*HDKey, error) {
    // Implement SLIP-0010 Ed25519 derivation
    // Path format: m/44'/501'/0'/0'/0' (always index 0)
}

func (k *HDKey) ToEd25519() (ed25519.PublicKey, ed25519.PrivateKey) {
    // Convert HD key to Ed25519 keypair
}
```

#### 3.3 Wallet Encryption

```go
type EncryptedWallet struct {
    Version    int    `json:"version"`
    Cipher     string `json:"cipher"`  // "AES-256-GCM"
    Salt       string `json:"salt"`    // Base64 encoded
    IV         string `json:"iv"`      // Base64 encoded
type EncryptedWallet struct {
    Version      int    `json:"version"`
    Cipher       string `json:"cipher"`  // "AES-256-GCM"
    Salt         string `json:"salt"`    // Base64 encoded
    IV           string `json:"iv"`      // Base64 encoded
    Ciphertext   string `json:"ciphertext"` // Base64 encoded (contains all seed phrases)
    Accounts     []EncryptedAccount `json:"accounts"`
}

type EncryptedAccount struct {
    SeedPhraseIndex int    `json:"seedPhraseIndex"`
    Label           string `json:"label"`
}

func EncryptWallet(wallet *Wallet, password string) (*EncryptedWallet, error)
func DecryptWallet(encrypted *EncryptedWallet, password string) (*Wallet, error)
```

### 4. Transaction Building and Signing

#### 4.1 Transaction Builder Integration

```go
type TransferRequest struct {
    From   string
    To     string
    Amount int64
    Memo   string
}

func (w *Wallet) BuildTransferTransaction(req *TransferRequest) (*transaction.Transaction, error) {
    account := w.GetActiveAccount()
    
    fromID, _ := filestore.FileIDFromString(account.Address)
    toID, _ := filestore.FileIDFromString(req.To)
    
    builder := transaction.NewTransactionBuilder(transaction.TxID{})
    
    err := builder.AddTransferInstruction(
        genesis.SystemProgramID,
        fromID,
        toID,
        req.Amount,
    )
    if err != nil {
        return nil, err
    }
    
    tx, err := builder.Build()
    if err != nil {
        return nil, err
    }
    
    // Sign transaction
    txData, _ := tx.Marshal()
    signature := ed25519.Sign(account.PrivateKey, txData)
    
    var sig [64]byte
    copy(sig[:], signature)
    
    tx.Signatures = []transaction.Signature{
        {PublicKey: account.PublicKey, Signature: sig},
    }
    
    return tx, nil
}
```

## Data Models

### 1. RPC Data Models

```go
// Account information
type AccountInfo struct {
    Address    string `json:"address"`
    Balance    int64  `json:"balance"`
    Owner      string `json:"owner"`
    DataLength int    `json:"dataLength"`
    Executable bool   `json:"executable"`
}

// Transaction status
type TransactionStatus struct {
    Signature   string `json:"signature"`
    Confirmed   bool   `json:"confirmed"`
    BlockHeight int64  `json:"blockHeight,omitempty"`
    Slot        int64  `json:"slot,omitempty"`
    Error       string `json:"error,omitempty"`
}

// Transaction record with full details
type TransactionRecord struct {
    Signature    string              `json:"signature"`
    BlockHeight  int64               `json:"blockHeight"`
    Slot         int64               `json:"slot"`
    Timestamp    time.Time           `json:"timestamp"`
    Success      bool                `json:"success"`
    Error        string              `json:"error,omitempty"`
    Instructions []InstructionRecord `json:"instructions"`
}
```

### 2. Wallet Data Models

```go
// Wallet configuration
type WalletConfig struct {
    RPCEndpoint string        `json:"rpcEndpoint"`
    AutoLock    time.Duration `json:"autoLock"`
    Theme       string        `json:"theme"`
}

// Wallet file format
type WalletFile struct {
    Version   int              `json:"version"`
    Encrypted EncryptedWallet  `json:"encrypted"`
    Config    WalletConfig     `json:"config"`
    CreatedAt time.Time        `json:"createdAt"`
    UpdatedAt time.Time        `json:"updatedAt"`
}
```

## Error Handling

### RPC Error Codes

```go
const (
    // JSON-RPC standard errors
    ParseError     = -32700  // Invalid JSON
    InvalidRequest = -32600  // Invalid Request object
    MethodNotFound = -32601  // Method does not exist
    InvalidParams  = -32602  // Invalid method parameters
    InternalError  = -32603  // Internal JSON-RPC error
    
    // Custom application errors
    InvalidSignature      = -32001  // Transaction signature invalid
    MalformedTransaction  = -32002  // Transaction format invalid
    InsufficientBalance   = -32003  // Account has insufficient balance
    AccountNotFound       = -32004  // Account does not exist
    TransactionNotFound   = -32005  // Transaction not found
    NetworkError          = -32006  // Network communication error
)
```

### Wallet Error Handling

```go
type WalletError struct {
    Code    string
    Message string
    Cause   error
}

const (
    ErrInvalidSeedPhrase  = "INVALID_SEED_PHRASE"
    ErrWalletLocked       = "WALLET_LOCKED"
    ErrInvalidPassword    = "INVALID_PASSWORD"
    ErrAccountNotFound    = "ACCOUNT_NOT_FOUND"
    ErrInsufficientFunds  = "INSUFFICIENT_FUNDS"
    ErrRPCConnection      = "RPC_CONNECTION_ERROR"
    ErrTransactionFailed  = "TRANSACTION_FAILED"
)
```

## Testing Strategy

### 1. RPC Node Tests

#### Unit Tests
- `internal/rpc/server_test.go`: HTTP server lifecycle, request handling
- `internal/rpc/handler_test.go`: JSON-RPC request parsing, method routing, error responses
- `internal/rpc/query_test.go`: Query engine operations, data retrieval

#### Integration Tests
- `internal/rpc/integration_test.go`: Full RPC flow with real ledger and filestore
- Test all RPC methods with various inputs
- Test concurrent request handling
- Test error scenarios (invalid addresses, missing data)

### 2. Wallet Tests

#### Unit Tests
- `cmd/wallet/core/wallet_test.go`: Seed phrase generation, key derivation, account management
- `cmd/wallet/core/crypto_test.go`: Encryption/decryption, signing
- `cmd/wallet/rpc/client_test.go`: RPC client request formatting, response parsing

#### Integration Tests
- `cmd/wallet/integration_test.go`: Full wallet flow with mock RPC server
- Test wallet creation, account derivation, transaction building
- Test wallet save/load with encryption

### 3. End-to-End Tests

- Start devnet with RPC node
- Create wallet, generate accounts
- Execute transfers
- Query transaction history
- Verify balances

## Devnet Integration

### Modified devnet.sh

Add RPC node startup to the devnet script:

```bash
# Start RPC node after leader
start_rpc_node() {
  local rpc_port=8899
  local leader_db="$DB_DIR/validator1.db"
  local leader_state="$DB_DIR/validator1_state.db"
  
  echo -e "${BLUE}Starting RPC node on port $rpc_port...${NC}"
  ./bin/poh-node rpc \
    --rpc-port=$rpc_port \
    --rpc-bind=127.0.0.1 \
    --ledger-path="$leader_db" \
    --state-path="$leader_state" \
    > "$LOG_DIR/devnet-rpc.log" 2>&1 &
  echo $! > "$PID_DIR/rpc.pid"
  
  echo -e "${GREEN}✓ RPC node started${NC}"
  echo "  Endpoint: http://127.0.0.1:$rpc_port"
}
```

### RPC Node Command

Add new subcommand to `cmd/main.go`:

```go
case "rpc":
    handleRPCCommand()
    return
```

## Security Considerations

### 1. RPC Node Security

- Bind to localhost by default (127.0.0.1)
- Rate limiting on requests (100 req/sec per IP)
- Request size limits (1MB max)
- No authentication required for read operations
- Transaction submission requires valid signatures

### 2. Wallet Security

- Seed phrase displayed only once during creation
- Wallet file encrypted with AES-256-GCM
- Password-based key derivation using PBKDF2 (100,000 iterations)
- Auto-lock after 5 minutes of inactivity
- Private keys never leave the wallet application
- File permissions set to 0600 (owner read/write only)

### 3. Network Security

- RPC communication over HTTP (localhost only)
- For production: Add TLS support
- Transaction signatures verified before submission
- No sensitive data in RPC responses (only public information)

## Performance Considerations

### 1. RPC Node

- Query caching for frequently accessed data (block height, recent blockhash)
- Connection pooling for database access
- Concurrent request handling with goroutines
- Response streaming for large result sets

### 2. Wallet

- Lazy loading of transaction history
- Background balance refresh
- Debounced UI updates
- Efficient terminal rendering with Bubble Tea

## Future Enhancements

### Phase 2 Features

1. **RPC Node**
   - WebSocket support for real-time updates
   - Subscription API for account changes
   - Transaction simulation endpoint
   - Historical data archival

2. **Wallet**
   - Hardware wallet support
   - Multi-signature accounts
   - Address book
   - Transaction templates
   - CSV export for transaction history
   - QR code generation for addresses

3. **Advanced Features**
   - Program deployment via wallet
   - Token management (SPL-like tokens)
   - Staking interface
   - Governance participation

## Dependencies

### Go Modules

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/tyler-smith/go-bip39 v1.1.0
    github.com/mattn/go-sqlite3 v1.14.18
    github.com/dgraph-io/badger/v4 v4.2.0
)
```

## Deployment

### Building

```bash
# Build RPC node (included in main binary)
go build -o bin/poh-node cmd/main.go

# Build wallet
go build -o bin/neon-wallet cmd/wallet/main.go
```

### Running

```bash
# Start devnet with RPC node
./devnet.sh start 3

# Run wallet
./bin/neon-wallet --rpc-url http://localhost:8899
```

## Conclusion

This design provides a comprehensive, production-ready RPC node and TUI wallet for the PoH blockchain. The RPC node offers a standardized JSON-RPC interface for blockchain interaction, while the TUI wallet provides a modern, secure, and user-friendly interface for account management and transaction execution. The integration with the devnet script ensures seamless development and testing workflows.

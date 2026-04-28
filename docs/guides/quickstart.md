# Quick Start Guide

## 30-Second Blockchain + Wallet Demo

```bash
# Start blockchain network with RPC node
./devnet.sh start 3

# In another terminal, start the wallet
./bin/neon-wallet
```

That's it! You now have:
- A 3-validator blockchain network running
- An RPC node providing API access
- A modern TUI wallet for account management

Use `./devnet.sh stop` to stop everything.

---

## 2-Minute Complete Setup

```bash
# 1. Build everything
go build -o bin/poh-node cmd/main.go
go build -o bin/neon-wallet cmd/wallet/main.go

# 2. Start the blockchain network
./devnet.sh start 3

# 3. Test RPC connectivity
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}'

# 4. Start the wallet
./bin/neon-wallet
```

---

## What You'll See

### Devnet Startup
```
Building poh-node binary...
Starting devnet with 3 validators...

Network Configuration:
  Validator 1 (Leader): localhost:8000 -> devnet-data/validator1.db
  Validator 2:          localhost:8001 -> devnet-data/validator2.db
  Validator 3:          localhost:8002 -> devnet-data/validator3.db

Starting RPC node on port 8899...
✓ RPC node started (PID: 12345)
  Endpoint: http://127.0.0.1:8899

✓ Devnet started successfully!
```

### Wallet Interface
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
│              │  └─────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Understanding the Components

**Validators**: Create and validate blocks every ~400ms
**RPC Node**: Provides JSON-RPC API for external access
**Wallet**: Modern TUI for account management and transfers
**Block Height**: Sequential number of blocks in the chain
**Neon**: The native token (1 Neon = 1,000,000,000 electrons)

---

## Try These Commands

### Different Network Sizes
```bash
./devnet.sh start 5    # 5 validators + RPC
./devnet.sh start 7 --rpc-port 9899    # Custom RPC port
```

### Network Management
```bash
./devnet.sh status     # Check all components
./devnet.sh logs       # View all logs
./devnet.sh logs rpc   # RPC logs only
./devnet.sh restart 3  # Restart with 3 validators
```

### RPC Testing
```bash
# Get blockchain height
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}'

# Get recent blockhash
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getRecentBlockhash","params":{},"id":1}'
```

### Wallet Operations
```bash
# Start wallet with custom settings
./bin/neon-wallet --rpc-url http://localhost:8899
./bin/neon-wallet --wallet-path ./my-wallet.dat

# Create multiple wallets
./bin/neon-wallet --wallet-path ~/.poh-wallet/personal.dat
./bin/neon-wallet --wallet-path ~/.poh-wallet/business.dat
```

### Legacy BFT Testing
```bash
# Test Byzantine Fault Tolerance (legacy tmux demos)
./demo-bft.sh 3 1    # 3 honest + 1 malicious
./demo-bft.sh 6 2    # 6 honest + 2 malicious (strong BFT)
./demo-bft.sh 2 2    # 2 honest + 2 malicious (NO BFT - will fail)
```

---

## Wallet First-Time Setup

When you run the wallet for the first time:

1. **Choose setup method**:
   - Create New Wallet (generates new seed phrase)
   - Restore from Seed Phrase (import existing)

2. **For new wallet**:
   - Select seed phrase length (12 or 24 words)
   - Write down your seed phrase (ONLY shown once!)
   - Set a strong password

3. **For restore**:
   - Enter your seed phrase
   - Set a password for encryption

4. **Start using**:
   - Navigate with number keys (1-5) or arrows
   - View accounts, send transfers, check history
   - Auto-locks after 5 minutes of inactivity

---

## Network Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Validator 1   │    │   Validator 2   │    │   Validator 3   │
│    (Leader)     │◄──►│   (Replica)     │◄──►│   (Replica)     │
│  localhost:8000 │    │  localhost:8001 │    │  localhost:8002 │
└─────────┬───────┘    └─────────────────┘    └─────────────────┘
          │
          │ (reads ledger + state)
          ▼
┌─────────────────┐    ┌─────────────────┐
│    RPC Node     │◄──►│   Neon Wallet   │
│ localhost:8899  │    │  (Terminal UI)  │
└─────────────────┘    └─────────────────┘
```

---

## tmux Basics

| Action | Keys |
|--------|------|
| Switch between panes | `Ctrl+B` then arrow keys |
| Zoom a pane | `Ctrl+B` then `Z` |
| Detach (keep running) | `Ctrl+B` then `D` |
| Reattach later | `tmux attach -t poh-demo` |

---

## Check the Results

After running for 30 seconds, inspect the blockchain:

```bash
# Check devnet status
./devnet.sh status

# How many blocks were created?
sqlite3 devnet-data/validator1.db "SELECT COUNT(*) FROM blocks;"

# View recent blocks
sqlite3 devnet-data/validator1.db "SELECT block_height, slot FROM blocks ORDER BY block_height DESC LIMIT 10;"

# Test RPC responses
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}'
```

---

## Stop Everything

```bash
# Stop devnet (includes RPC node)
./devnet.sh stop

# Or clean everything (removes databases)
./devnet.sh clean

# Legacy cleanup for tmux demos
./stop-demo.sh
```

---

## What's Next?

### Immediate Next Steps
1. **Explore the wallet** - Create accounts, view history, send transfers
2. **Test RPC API** - Try different endpoints (see [RPC API Reference](../reference/rpc-api.md))
3. **Check network status** - `./devnet.sh status` and `./devnet.sh logs`

### Learn More
1. **[Wallet User Guide](wallet-usage.md)** - Complete wallet documentation
2. **[Demo Guide](demo.md)** - Detailed devnet and testing guide
3. **[RPC API Reference](../reference/rpc-api.md)** - Complete API documentation
4. **[CLI Usage Guide](cli-usage.md)** - Command-line tools
5. **[QuanticScript Guide](quanticscript.md)** - Smart contract language

### Development
1. **Run tests** - `go test -v ./internal`
2. **Explore code** - Check out `internal/` and `cmd/` directories
3. **Build features** - Extend the blockchain or wallet
4. **Write smart contracts** - Learn QuanticScript programming

---

## Need Help?

### Common Issues
- **Build failures?** Run `go mod download` and ensure Go 1.23+ is installed
- **Port conflicts?** Use `./devnet.sh stop` or try `--port 9000 --rpc-port 9899`
- **RPC connection refused?** Check `./devnet.sh status` and verify RPC is running
- **Wallet won't connect?** Verify RPC URL with `--rpc-url http://localhost:8899`
- **Database locked?** Run `./devnet.sh clean` to reset everything

### Getting Support
- **Check logs** - `./devnet.sh logs` or `./devnet.sh logs rpc`
- **Review documentation** - All guides are in `docs/guides/`
- **Run diagnostics** - `./devnet.sh status` shows component health
- **Test RPC manually** - Use curl commands shown above

---

## The Big Picture

This is a **Proof of History (PoH)** blockchain inspired by Solana:

- **PoH Clock**: Creates a verifiable timeline using sequential SHA-256 hashing
- **Leader-Based Consensus**: Leader produces blocks, replicas validate
- **JSON-RPC API**: Standard interface for external applications
- **Modern Wallet**: Feature-rich TUI with BIP39/44 support
- **BFT Tolerance**: Network handles malicious nodes if honest > 2×malicious
- **File-Based State**: Uniform abstraction for accounts and programs

The demo shows a complete blockchain ecosystem working in real-time!

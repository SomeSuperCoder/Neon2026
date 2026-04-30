# DPoS Demo Script - Quick Start Guide

## Overview

The `demo-dpos.sh` script demonstrates the stake-weighted leader schedule with wallet-based validators. It automatically creates wallets, generates genesis configuration, and starts a multi-validator network.

## Quick Start

### 1. Start a 3-validator demo
```bash
./demo-dpos.sh start 3
```

This will:
- Create 3 password-protected wallets (`dpos-validator-1`, `dpos-validator-2`, `dpos-validator-3`)
- Generate genesis configuration with stake-weighted distribution
- Start 3 validator nodes on ports 8000, 8001, 8002
- Start an RPC node on port 8899
- Run for 30 seconds (default)
- Collect and display statistics

### 2. Start a 5-validator demo for 60 seconds
```bash
./demo-dpos.sh start 5 --duration 60
```

### 3. Check demo status
```bash
./demo-dpos.sh status
```

Output:
```
DPoS Demo Status:
================

✓ validator1 (PID: 12345, Port: 8000) - RUNNING
    Blocks: 150

✓ validator2 (PID: 12346, Port: 8001) - RUNNING
    Blocks: 75

✓ validator3 (PID: 12347, Port: 8002) - RUNNING
    Blocks: 75

✓ RPC Node (PID: 12348, Port: 8899) - RUNNING

Summary: 4 running, 0 stopped
```

### 4. View logs
```bash
# View logs for validator 1
./demo-dpos.sh logs 1

# View logs for all validators
./demo-dpos.sh logs

# View RPC logs
./demo-dpos.sh logs rpc
```

### 5. Stop the demo
```bash
./demo-dpos.sh stop
```

### 6. Clean all data
```bash
./demo-dpos.sh clean
```

## Stake-Weighted Distribution

The demo script creates validators with the following stake distribution:

### 2 Validators
- Validator 1: 10 Neon (2x stake)
- Validator 2: 5 Neon

**Expected slot distribution:**
- Validator 1: 66.7% of slots (288,000 slots)
- Validator 2: 33.3% of slots (144,000 slots)

### 3 Validators
- Validator 1: 10 Neon (2x stake)
- Validator 2: 5 Neon
- Validator 3: 5 Neon

**Expected slot distribution:**
- Validator 1: 50.0% of slots (216,000 slots)
- Validator 2: 25.0% of slots (108,000 slots)
- Validator 3: 25.0% of slots (108,000 slots)

### 5 Validators
- Validator 1: 10 Neon (2x stake)
- Validators 2-5: 5 Neon each

**Expected slot distribution:**
- Validator 1: 33.3% of slots (144,000 slots)
- Validators 2-5: 16.7% of slots each (72,000 slots)

## Wallet Management

Wallets are created automatically in `~/.config/poh-blockchain/wallets/`:

```
~/.config/poh-blockchain/wallets/
├── dpos-validator-1.wallet
├── dpos-validator-2.wallet
├── dpos-validator-3.wallet
└── ...
```

**Wallet passwords (for demo):**
- Validator 1: `demo-password-1`
- Validator 2: `demo-password-2`
- Validator 3: `demo-password-3`
- etc.

To manually inspect a wallet:
```bash
go run cmd/main.go wallet show --name dpos-validator-1
# Enter password: demo-password-1
```

## Genesis Configuration

The demo script generates a `genesis-dpos.json` file:

```json
{
  "epochLength": 432000,
  "validators": [
    {
      "publicKey": "0x0000000000000000000000000000000000000001",
      "stake": 10000000
    },
    {
      "publicKey": "0x0000000000000000000000000000000000000002",
      "stake": 5000000
    },
    {
      "publicKey": "0x0000000000000000000000000000000000000003",
      "stake": 5000000
    }
  ]
}
```

**Key parameters:**
- `epochLength`: 432,000 slots (~2 days at 400ms/slot)
- `stake`: Minimum 1,000,000 electrons (1 Neon)

## Network Configuration

### Validator Ports
- Validator 1: 8000
- Validator 2: 8001
- Validator 3: 8002
- etc.

### RPC Node
- Port: 8899
- Endpoint: `http://127.0.0.1:8899`

### Databases
- Validator 1: `./devnet-data/validator1.db`
- Validator 2: `./devnet-data/validator2.db`
- etc.

### Logs
- Validator 1: `./logs/devnet-validator-1.log`
- Validator 2: `./logs/devnet-validator-2.log`
- RPC: `./logs/devnet-rpc.log`

## Command Reference

```bash
# Start demo with N validators
./demo-dpos.sh start [N]

# Stop running demo
./demo-dpos.sh stop

# Restart demo with N validators
./demo-dpos.sh restart [N]

# Show demo status
./demo-dpos.sh status

# Show logs for specific validator or all
./demo-dpos.sh logs [VALIDATOR_ID]

# Clean all data
./demo-dpos.sh clean

# Show help
./demo-dpos.sh help
```

## Options

```bash
--validators N        Number of validators (default: 3)
--port PORT          Starting port number (default: 8000)
--rpc-port PORT      RPC node port (default: 8899)
--db-dir DIR         Database directory (default: ./devnet-data)
--log-dir DIR        Log directory (default: ./logs)
--duration SECONDS   Demo duration in seconds (default: 30)
--help               Show help message
```

## Examples

### Start 5 validators for 2 minutes
```bash
./demo-dpos.sh start 5 --duration 120
```

### Start demo on custom ports
```bash
./demo-dpos.sh start 3 --port 9000 --rpc-port 9899
```

### Start demo with custom database directory
```bash
./demo-dpos.sh start 3 --db-dir /tmp/dpos-demo
```

### Monitor demo in real-time
```bash
# Terminal 1: Start demo
./demo-dpos.sh start 3

# Terminal 2: Watch status
watch -n 1 './demo-dpos.sh status'

# Terminal 3: View logs
./demo-dpos.sh logs
```

## Troubleshooting

### Port already in use
```bash
# Stop any running demo
./demo-dpos.sh stop

# Or use custom ports
./demo-dpos.sh start 3 --port 9000 --rpc-port 9899
```

### Database locked
```bash
# Clean and restart
./demo-dpos.sh clean
./demo-dpos.sh start 3
```

### Wallet creation failed
```bash
# Check wallet directory
ls -la ~/.config/poh-blockchain/wallets/

# Manually create wallet
go run cmd/main.go wallet create --name dpos-validator-1
```

### Build failed
```bash
# Rebuild binary
go build -o bin/poh-node cmd/main.go

# Then restart demo
./demo-dpos.sh start 3
```

## Testing

Run comprehensive tests:
```bash
# All DPoS tests
go test -v ./cmd -run DemoDP

# Specific test
go test -v ./cmd -run TestStakeWeightedSlotDistribution

# With coverage
go test -cover ./cmd -run DemoDP
```

## Next Steps

1. **Verify stake-weighted distribution:**
   - Start demo with 3 validators
   - Monitor logs to see block production
   - Verify Validator 1 produces ~2x blocks as Validators 2 and 3

2. **Test epoch boundary:**
   - Run demo for longer duration
   - Observe schedule recalculation at epoch boundaries
   - Verify new schedule is computed correctly

3. **Test missed block handling:**
   - Stop a validator mid-demo
   - Observe missed block recording
   - Verify network continues producing blocks

4. **Integrate with Validator TUI:**
   - Run demo in one terminal
   - Launch Validator TUI in another
   - Monitor real-time validator metrics

## Related Documentation

- [DPoS Implementation Details](DEMO_DPOS_IMPLEMENTATION.md)
- [Stake-Weighted Leader Schedule Requirements](.kiro/specs/stake-weighted-leader-schedule/requirements.md)
- [Stake-Weighted Leader Schedule Design](.kiro/specs/stake-weighted-leader-schedule/design.md)
- [Wallet Management Guide](docs/guides/wallet-usage.md)
- [CLI Usage Guide](docs/guides/cli-usage.md)

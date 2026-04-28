# PoH Blockchain Demo Guide

## Quick Start

### Local Development Network

Run a local development network with devnet.sh:
```bash
# Start with default 3 validators + RPC node
./devnet.sh start

# Start with custom number of validators + RPC node
./devnet.sh start 5

# Start with custom RPC port
./devnet.sh start 3 --rpc-port 9899

# Check network status (includes RPC node)
./devnet.sh status

# View logs for all validators and RPC node
./devnet.sh logs

# View logs for specific validator or RPC node
./devnet.sh logs 1    # Validator 1
./devnet.sh logs rpc  # RPC node only

# Stop the network (includes RPC node)
./devnet.sh stop
```

### Comprehensive Testing

Run the audit script to validate all blockchain functionality:
```bash
# Run with default settings (30s per phase, 3 validators)
./audit.sh

# Run with custom duration and validator count
./audit.sh --duration 60 --validators 5

# Run in CI mode (no colors, no prompts)
./audit.sh --ci
```

The audit script tests:
- Basic consensus with honest nodes
- BFT with tolerance (network can handle malicious nodes)
- BFT without tolerance (insufficient honest nodes)
- DPoS lifecycle (delegation, epochs, rewards, slashing)

## Script Features

### devnet.sh - Local Development Network
- **Persistent state**: Maintains blockchain state across restarts
- **Background processes**: Nodes run as background processes (no tmux required)
- **Multiple commands**: start, stop, restart, status, logs, clean
- **Configurable**: Custom port, database directory, log directory, and RPC port
- **Easy monitoring**: View logs and status for all validators and RPC node
- **Integrated RPC**: Automatically starts RPC node alongside validators
- **Wallet ready**: RPC endpoint ready for wallet connections

### audit.sh - Comprehensive Testing
- **Four test phases**: Basic consensus, BFT with/without tolerance, DPoS lifecycle
- **JSON reports**: Machine-parseable output for CI/CD integration
- **Configurable**: Custom test duration and validator count
- **CI mode**: Automated testing without interactive prompts
- **Exit codes**: 0=pass, 1=fail, 2=config error

## devnet.sh Commands

### start
Start a local development network with N validators:
```bash
./devnet.sh start [N]    # Default: 3 validators
```

Options:
- `--port PORT`: Starting port number (default: 8000)
- `--rpc-port PORT`: RPC node port (default: 8899)
- `--db-dir DIR`: Database directory (default: ./devnet-data)
- `--log-dir DIR`: Log directory (default: ./logs)

### stop
Stop the running devnet:
```bash
./devnet.sh stop
```

### restart
Restart the devnet with the same or different number of validators:
```bash
./devnet.sh restart [N]
```

### status
Show the status of all validators:
```bash
./devnet.sh status
```

Displays:
- Running/stopped status for each validator and RPC node
- Block counts from databases
- Process IDs
- RPC endpoint URL

### logs
View logs for specific validator, RPC node, or all components:
```bash
./devnet.sh logs           # All validators and RPC node
./devnet.sh logs 1         # Validator 1 only
./devnet.sh logs rpc       # RPC node only
```

### clean
Stop devnet and remove all data:
```bash
./devnet.sh clean
```

Prompts for confirmation before deleting databases and logs.

## Stopping the Network

### devnet.sh
```bash
./devnet.sh stop
```

### audit.sh
The audit script automatically cleans up after each test phase. Press `Ctrl+C` to interrupt if needed.

### Legacy cleanup
```bash
./stop-demo.sh
```
This will stop both legacy tmux sessions and any running devnet.

## What You'll See

### devnet.sh Output

When starting the network:
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
  Note: RPC uses a snapshot of the state database

Logs:
  logs/devnet-validator-1.log
  logs/devnet-validator-2.log
  logs/devnet-validator-3.log
  logs/devnet-rpc.log

✓ Devnet started successfully!

Network Configuration:
=====================
  Validator 1 (leader):
    Address:  localhost:8000
    Database: devnet-data/validator1.db
    Logs:     logs/devnet-validator-1.log
    PID:      12340

  Validator 2 (replica):
    Address:  localhost:8001
    Database: devnet-data/validator2.db
    Logs:     logs/devnet-validator-2.log
    PID:      12341

  Validator 3 (replica):
    Address:  localhost:8002
    Database: devnet-data/validator3.db
    Logs:     logs/devnet-validator-3.log
    PID:      12342

  RPC Node:
    Endpoint: http://127.0.0.1:8899
    Logs:     logs/devnet-rpc.log
    PID:      12345

Management Commands:
  Status:  ./devnet.sh status
  Logs:    ./devnet.sh logs [validator_id]
  Stop:    ./devnet.sh stop
  Clean:   ./devnet.sh clean
```

### audit.sh Output

The audit script shows progress through four phases:

**Phase 1: Basic Consensus**
```
=== Phase 1: Basic Consensus ===
Starting 1 leader + 3 honest replicas...
Running for 30 seconds...
✓ Phase complete: 150 blocks produced
```

**Phase 2: BFT With Tolerance**
```
=== Phase 2: BFT With Tolerance ===
Starting 1 leader + 4 honest + 1 malicious...
Running for 30 seconds...
✓ Phase complete: Honest nodes rejected 15 invalid blocks
```

**Phase 3: BFT Without Tolerance**
```
=== Phase 3: BFT Without Tolerance ===
Starting 1 leader + 2 honest + 2 malicious...
Running for 30 seconds...
✓ Phase complete: Network behavior under insufficient BFT observed
```

**Phase 4: DPoS Lifecycle**
```
=== Phase 4: DPoS Lifecycle ===
Starting 3 validators from genesis...
Running for 30 seconds...
✓ Phase complete: DPoS lifecycle simulated
```

**Final Summary**
```
=== Audit Summary ===
Total phases: 4
Passed: 4
Failed: 0

Overall status: PASSED

Report saved to: logs/audit-20260427-120000.json
```

## Network Configuration

### devnet.sh Network

The devnet creates the following network:

- **Validator 1 (Leader)**: `localhost:8000` → `devnet-data/validator1.db`
- **Validator 2**: `localhost:8001` → `devnet-data/validator2.db`
- **Validator 3**: `localhost:8002` → `devnet-data/validator3.db`
- **Validator N**: `localhost:800(N-1)` → `devnet-data/validatorN.db`
- **RPC Node**: `http://127.0.0.1:8899` → Uses leader's ledger + state snapshot

All replicas connect to the leader and receive blocks via TCP.
The RPC node provides JSON-RPC API access to blockchain data.

State is persistent across restarts. Use `./devnet.sh clean` to reset.

### audit.sh Network

The audit script creates temporary networks for each test phase:

- **Phase 1**: 1 leader + 3 honest replicas
- **Phase 2**: 1 leader + 4 honest + 1 malicious
- **Phase 3**: 1 leader + 2 honest + 2 malicious
- **Phase 4**: N validators (configurable, default 3)

All test artifacts are automatically cleaned up after each phase.

## BFT Testing

The audit script includes comprehensive BFT testing in phases 2 and 3.

### Phase 2: BFT With Tolerance

Tests network behavior when honest nodes outnumber malicious nodes:
- **Configuration**: 1 leader + 4 honest + 1 malicious
- **BFT Status**: ✓ Has BFT (4 > 2×1)
- **Expected**: Honest nodes reject invalid blocks, maintain integrity

### Phase 3: BFT Without Tolerance

Tests network behavior when there are insufficient honest nodes:
- **Configuration**: 1 leader + 2 honest + 2 malicious
- **BFT Status**: ✗ No BFT (2 ≤ 2×2)
- **Expected**: Network may accept invalid blocks, demonstrates need for BFT threshold

## Database Files

### devnet.sh Databases

Each validator maintains its own SQLite database in `devnet-data/`:
- `devnet-data/validator1.db` - Leader's blockchain state
- `devnet-data/validator2.db` - Replica 1's blockchain state
- `devnet-data/validator3.db` - Replica 2's blockchain state
- etc.

State persists across restarts. Use `./devnet.sh clean` to remove all data.

### audit.sh Databases

The audit script creates temporary databases for each test phase:
- `audit-leader.db`, `audit-honest-1.db`, etc.
- Automatically cleaned up after each phase
- Final report saved to `logs/audit-TIMESTAMP.json`

## Troubleshooting

### "address already in use"
Another process is using the ports. Either:
1. Stop the existing network: `./devnet.sh stop` or `./stop-demo.sh`
2. Or change the ports: `./devnet.sh start --port 9000`

### Nodes not connecting
- Ensure the leader starts first (scripts handle this automatically)
- Check firewall settings if running across machines
- Verify ports 8000-8009 are available

### Build failures
```bash
go mod download
go build -o bin/poh-node cmd/main.go
```

### Database locked errors
```bash
./devnet.sh clean
./devnet.sh start
```

## Advanced Usage

### Running validators on different machines

Edit the devnet.sh script or use manual node startup:
```bash
# On machine 1 (leader)
go run cmd/main.go --type=leader --port=8000 --db=./leader.db

# On machine 2 (replica)
go run cmd/main.go --type=replica --port=8001 --peers=192.168.1.100:8000 --db=./replica.db
```

### Inspecting databases

```bash
# devnet databases
sqlite3 devnet-data/validator1.db "SELECT block_height, slot, timestamp FROM blocks;"

# audit results
cat logs/audit-20260427-120000.json | jq '.summary'
```

## RPC Node Integration

### Automatic RPC Startup

The devnet script automatically starts an RPC node alongside the validators:

```bash
# RPC node starts automatically with validators
./devnet.sh start 3

# Check RPC node status
./devnet.sh status

# View RPC logs specifically
./devnet.sh logs rpc
```

### RPC Configuration

The RPC node is configured to:
- **Port**: 8899 (configurable with `--rpc-port`)
- **Bind Address**: 127.0.0.1 (localhost only for security)
- **Ledger**: Uses leader's blockchain database (read-only)
- **State**: Uses a snapshot of leader's state database
- **CORS**: Enabled for browser access

### RPC Endpoints

Once devnet is running, the RPC node provides these endpoints:

```bash
# Test RPC connectivity
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}'

# Get account balance (replace with actual address)
curl -X POST http://localhost:8899 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getBalance","params":{"address":"your_address_here"},"id":1}'
```

### Using with Neon Wallet

Start the wallet after devnet is running:

```bash
# Start devnet with RPC
./devnet.sh start 3

# In another terminal, start the wallet
./bin/neon-wallet

# Or specify custom RPC endpoint
./bin/neon-wallet --rpc-url http://localhost:8899
```

The wallet will automatically connect to the RPC node and provide:
- Account management and balance viewing
- Transaction history
- Transfer functionality
- Real-time blockchain data

### RPC Database Architecture

The RPC node uses a unique architecture to avoid database locks:

- **Ledger Access**: Direct read-only access to leader's SQLite database
- **State Access**: Uses a snapshot copy of the leader's state database
- **No Conflicts**: Multiple RPC nodes can run simultaneously
- **Real-time Data**: Ledger data is always current
- **Snapshot State**: State data reflects the snapshot time

### Troubleshooting RPC

**RPC node fails to start:**
```bash
# Check if leader database exists
ls -la devnet-data/validator1.db

# Check RPC logs
./devnet.sh logs rpc

# Restart devnet if needed
./devnet.sh restart 3
```

**Connection refused:**
```bash
# Verify RPC is running
./devnet.sh status

# Check if port is in use
netstat -an | grep 8899

# Try different port
./devnet.sh start 3 --rpc-port 9899
```

**Wallet cannot connect:**
```bash
# Test RPC manually
curl http://localhost:8899

# Check wallet RPC URL
./bin/neon-wallet --rpc-url http://localhost:8899

# Verify firewall settings
```

### Monitoring block production

```bash
# Watch devnet block production
watch -n 1 'sqlite3 devnet-data/validator1.db "SELECT COUNT(*) as total_blocks FROM blocks;"'

# Check devnet status
./devnet.sh status
```

### Analyzing audit results

```bash
# View JSON report
cat logs/audit-TIMESTAMP.json | jq '.'

# Extract specific metrics
cat logs/audit-TIMESTAMP.json | jq '.phases.basic_consensus.metrics'

# Check overall status
cat logs/audit-TIMESTAMP.json | jq '.summary.overall_status'
```

## Performance Notes

- Each block is produced every ~400ms (2.5 blocks/second)
- Each block contains minimum 64 ticks (800,000 hash operations)
- Replicas validate and store blocks in real-time
- Network latency is typically <10ms on localhost

## Next Steps

After running the scripts:

### With devnet.sh
1. Observe block production: `./devnet.sh logs`
2. Check network status: `./devnet.sh status`
3. Test RPC connectivity: `curl -X POST http://localhost:8899 -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"getBlockHeight","params":{},"id":1}'`
4. Start the wallet: `./bin/neon-wallet`
5. Submit transactions via CLI (see [CLI Usage Guide](cli-usage.md))
6. Inspect databases: `sqlite3 devnet-data/validator1.db`
7. Test state persistence: `./devnet.sh restart`

### With audit.sh
1. Review the JSON report: `cat logs/audit-TIMESTAMP.json`
2. Verify all phases passed
3. Integrate into CI/CD pipeline
4. Run with custom parameters: `./audit.sh --duration 60 --validators 5`

### General
1. Run integration tests: `go test -v ./internal`
2. Explore the codebase and modify features
3. Read the [Wallet User Guide](wallet-usage.md)
4. Read the [RPC API Reference](../reference/rpc-api.md)
5. Read the [QuanticScript Guide](quanticscript.md)
6. Check out the [Testing Guide](../testing/automated-testing.md)

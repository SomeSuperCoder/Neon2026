# PoH Blockchain Demo Guide

## Quick Start

### Local Development Network

Run a local development network with devnet.sh:
```bash
# Start with default 3 validators
./devnet.sh start

# Start with custom number of validators
./devnet.sh start 5

# Check network status
./devnet.sh status

# View logs
./devnet.sh logs

# Stop the network
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
- **Configurable**: Custom port, database directory, and log directory
- **Easy monitoring**: View logs and status for all validators

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
- Running/stopped status for each validator
- Block counts from databases
- Process IDs

### logs
View logs for specific validator or all validators:
```bash
./devnet.sh logs           # All validators
./devnet.sh logs 1         # Validator 1 only
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
Building poh-node...
Starting devnet with 3 validators...

Network Configuration:
  Validator 1 (Leader): localhost:8000 -> devnet-data/validator1.db
  Validator 2:          localhost:8001 -> devnet-data/validator2.db
  Validator 3:          localhost:8002 -> devnet-data/validator3.db

Logs:
  logs/devnet-validator-1.log
  logs/devnet-validator-2.log
  logs/devnet-validator-3.log

Devnet started successfully!
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

All replicas connect to the leader and receive blocks via TCP.

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
3. Submit transactions via CLI (see [CLI Usage Guide](cli-usage.md))
4. Inspect databases: `sqlite3 devnet-data/validator1.db`
5. Test state persistence: `./devnet.sh restart`

### With audit.sh
1. Review the JSON report: `cat logs/audit-TIMESTAMP.json`
2. Verify all phases passed
3. Integrate into CI/CD pipeline
4. Run with custom parameters: `./audit.sh --duration 60 --validators 5`

### General
1. Run integration tests: `go test -v ./internal`
2. Explore the codebase and modify features
3. Read the [QuanticScript Guide](quanticscript.md)
4. Check out the [Testing Guide](../testing/automated-testing.md)

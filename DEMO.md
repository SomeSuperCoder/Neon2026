# PoH Blockchain Demo Guide

## Quick Start

### Regular Demo

Run the demo with default settings (1 leader + 2 replicas):
```bash
./demo.sh
```

Run with a custom number of replicas (1-9):
```bash
./demo.sh 5    # 1 leader + 5 replicas
```

### BFT Testing Demo

Run the BFT demo to test Byzantine Fault Tolerance with malicious nodes:
```bash
./demo-bft.sh 3 1    # 3 honest + 1 malicious replica
./demo-bft.sh 4 2    # 4 honest + 2 malicious replicas
```

The script will automatically calculate and display whether the network has BFT:
- **BFT Requirement**: Honest nodes > 2 × Malicious nodes
- **Example**: 4 honest + 1 malicious = BFT ✓ (4 > 2×1)
- **Example**: 2 honest + 2 malicious = NO BFT ✗ (2 ≤ 2×2)

## Demo Script Features

### Regular Demo (demo.sh)
- **Automatic build**: Compiles the project before starting
- **Clean slate**: Removes old database files
- **Configurable replicas**: Specify 1-9 replica nodes
- **Tiled layout**: Automatically arranges panes in tmux
- **Color-coded output**: Leader in blue, replicas in green

### BFT Testing Demo (demo-bft.sh)
- **Malicious nodes**: Simulates Byzantine faults for testing
- **BFT calculation**: Automatically checks if network has fault tolerance
- **Separate databases**: Malicious nodes use separate DB files
- **Color-coded output**: Honest nodes in green, malicious in red
- **Multiple attack vectors**: Tests various Byzantine behaviors

## tmux Navigation

| Action | Command |
|--------|---------|
| Switch panes | `Ctrl+B` then arrow keys |
| Zoom pane | `Ctrl+B` then `Z` (toggle) |
| Detach session | `Ctrl+B` then `D` |
| Reattach | `tmux attach -t poh-demo` |
| Scroll in pane | `Ctrl+B` then `[`, then arrow keys (press `q` to exit) |

## Stopping the Demo

### Option 1: Stop script
```bash
./stop-demo.sh
```
This will prompt you to optionally clean up database files.

### Option 2: Manual stop
1. Press `Ctrl+C` in each pane to stop the nodes
2. Type `exit` in each pane to close them
3. Or kill the session: `tmux kill-session -t poh-demo`

## What You'll See

### Leader Node (Top/First Pane)
```
Leader node: Producing block for slot 42
Leader node: Block produced - height=15, slot=42, entries=64
Leader node: Block stored to ledger - height=15
Leader node: Block broadcasted to peers - height=15
```

### Honest Replica Nodes
```
Replica node: Received block - height=15, slot=42, entries=64
Replica node: Block passed consensus validation - height=15
Replica node: Block passed verification - height=15
Replica node: Block linkage verified - height=15
Replica node: Block stored successfully - height=15, slot=42
```

### Malicious Replica Nodes (BFT Demo Only)
```
MALICIOUS: Skipping validation for block 12
MALICIOUS: Stored block 12 without validation
Replica node: Block verification failed: entry hash does not match
MALICIOUS: Ignoring verification failure, storing anyway
MALICIOUS: Corrupted block 15 by setting invalid hash count
```

### Malicious Leader Behavior (when using --malicious flag)
```
MALICIOUS LEADER: Starting node as MALICIOUS LEADER
MALICIOUS: Corrupted block 9 by setting invalid hash count
MALICIOUS: Corrupted block 10 with wrong previous hash
MALICIOUS: Skipping storage of corrupted block 12
```

Note: The `--malicious` flag enables Byzantine fault behaviors for testing. Use `demo-bft.sh` for automated BFT testing.

## Network Configuration

The demo creates the following network:

- **Leader**: `localhost:8000` (produces blocks)
- **Replica 1**: `localhost:8001` (validates and stores)
- **Replica 2**: `localhost:8002` (validates and stores)
- **Replica N**: `localhost:800N` (validates and stores)

All replicas connect to the leader and receive blocks via TCP.

## Malicious Node Behaviors

When running with the `--malicious` flag or using `demo-bft.sh`, nodes exhibit Byzantine faults:

### Malicious Leader Behaviors
1. **Invalid Hash Counts**: Every 3rd block has corrupted NumHashes (set to 1)
2. **Wrong Previous Hash**: Every 5th block has corrupted PreviousBlockHash
3. **Skip Storage**: Doesn't store corrupted blocks locally
4. **Broadcast Corruption**: Still broadcasts corrupted blocks to replicas

### Malicious Replica Behaviors
1. **Skip Validation**: Every 4th block is stored without validation
2. **Ignore Consensus Failures**: Every 6th block ignores consensus validation errors
3. **Ignore Verification Failures**: Every 6th block ignores PoH verification errors
4. **Ignore Linkage Failures**: Every 6th block ignores block linkage errors

### Expected Outcomes

**With BFT (Honest > 2×Malicious)**:
- Honest nodes reject invalid blocks
- Network maintains valid chain state
- Malicious nodes have corrupted local state
- Consensus is preserved

**Without BFT (Honest ≤ 2×Malicious)**:
- Invalid blocks may be accepted
- Network state becomes inconsistent
- Chain integrity is compromised
- Demonstrates need for BFT threshold

## BFT Testing Scenarios

### Scenario 1: Network Has BFT
```bash
./demo-bft.sh 4 1    # 4 honest, 1 malicious
```
**Expected**: Honest nodes reject corrupted blocks, maintain valid chain

### Scenario 2: Network Lacks BFT
```bash
./demo-bft.sh 2 2    # 2 honest, 2 malicious
```
**Expected**: Network may accept invalid blocks, chain integrity at risk

### Scenario 3: Extreme Byzantine Fault
```bash
./demo-bft.sh 6 2    # 6 honest, 2 malicious
```
**Expected**: Strong BFT, network easily handles Byzantine faults

## Database Files

Each node maintains its own SQLite database:
- `leader.db` - Leader's blockchain state
- `replica1.db` - Replica 1's blockchain state
- `replica2.db` - Replica 2's blockchain state
- etc.

After stopping the demo, you can inspect these databases or restart nodes with the existing state.

## Troubleshooting

### "tmux: command not found"
Install tmux:
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux
```

### "address already in use"
Another process is using the ports. Either:
1. Stop the existing demo: `./stop-demo.sh`
2. Or change the ports in the script

### Nodes not connecting
- Ensure the leader starts first (the script handles this)
- Check firewall settings if running across machines
- Verify ports 8000-8009 are available

## Advanced Usage

### Running nodes on different machines

Edit `demo.sh` and change `localhost` to the leader's IP address:
```bash
--peers=192.168.1.100:8000
```

### Inspecting the database

```bash
sqlite3 leader.db "SELECT block_height, slot, timestamp FROM blocks;"
```

### Monitoring block production rate

```bash
watch -n 1 'sqlite3 leader.db "SELECT COUNT(*) as total_blocks FROM blocks;"'
```

## Performance Notes

- Each block is produced every ~400ms (2.5 blocks/second)
- Each block contains minimum 64 ticks (800,000 hash operations)
- Replicas validate and store blocks in real-time
- Network latency is typically <10ms on localhost

## Next Steps

After running the demo:
1. Observe the block production and propagation
2. Try stopping a replica and restarting it (it will sync)
3. Inspect the database files to see stored blocks
4. Run the integration tests: `go test -v ./internal`
5. Modify the code and rebuild to experiment

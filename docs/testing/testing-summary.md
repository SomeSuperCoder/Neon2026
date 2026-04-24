# Testing & Demo Summary

## Available Demo Scripts

### 1. Regular Demo (`demo.sh`)
**Purpose**: Run a basic blockchain network with honest nodes

**Usage**:
```bash
./demo.sh           # 1 leader + 2 replicas (default)
./demo.sh 5         # 1 leader + 5 replicas
```

**Features**:
- All nodes are honest
- Tests basic blockchain functionality
- Verifies block production and propagation
- Demonstrates P2P networking

---

### 2. BFT Testing Demo (`demo-bft.sh`)
**Purpose**: Test Byzantine Fault Tolerance with malicious nodes

**Usage**:
```bash
./demo-bft.sh 3 1   # 3 honest + 1 malicious (has BFT)
./demo-bft.sh 4 2   # 4 honest + 2 malicious (has BFT)
./demo-bft.sh 2 2   # 2 honest + 2 malicious (NO BFT)
```

**Features**:
- Mix of honest and malicious nodes
- Automatic BFT calculation
- Color-coded output (green=honest, red=malicious)
- Tests network resilience

---

### 3. Stop Script (`stop-demo.sh`)
**Purpose**: Stop any running demo and optionally clean up

**Usage**:
```bash
./stop-demo.sh
```

**Features**:
- Stops both regular and BFT demos
- Optional database cleanup
- Graceful shutdown

---

## Command-Line Flags

### Node Configuration
```bash
--type=leader|replica    # Node type (default: replica)
--port=8000             # Port to listen on (default: 8080)
--peers=host:port       # Comma-separated peer list
--db=path/to/db         # Database file path
--malicious             # Enable malicious mode (default: false)
```

### Examples
```bash
# Honest leader
./bin/poh-node --type=leader --port=8000 --db=leader.db

# Honest replica
./bin/poh-node --type=replica --port=8001 --peers=localhost:8000 --db=replica.db

# Malicious replica
./bin/poh-node --type=replica --port=8002 --peers=localhost:8000 --db=malicious.db --malicious
```

---

## Malicious Behaviors

### Leader Malicious Behaviors
| Frequency | Behavior | Impact |
|-----------|----------|--------|
| Every 3rd block | Invalid hash count (NumHashes=1) | PoH verification fails |
| Every 5th block | Wrong previous hash | Chain linkage breaks |
| Corrupted blocks | Skip local storage | Inconsistent state |
| All blocks | Broadcast corrupted data | Tests replica validation |

### Replica Malicious Behaviors
| Frequency | Behavior | Impact |
|-----------|----------|--------|
| Every 4th block | Skip all validation | Accept any block |
| Every 6th block | Ignore consensus failures | Store invalid blocks |
| Every 6th block | Ignore verification failures | Store unverified blocks |
| Every 6th block | Ignore linkage failures | Break chain continuity |

---

## BFT Requirements

### Formula
```
Honest Nodes > 2 × Malicious Nodes
```

### Examples
| Honest | Malicious | Total | BFT Status | Reason |
|--------|-----------|-------|------------|--------|
| 4 | 1 | 5 | ✓ Has BFT | 4 > 2×1 |
| 3 | 1 | 4 | ✓ Has BFT | 3 > 2×1 |
| 5 | 2 | 7 | ✓ Has BFT | 5 > 2×2 |
| 2 | 1 | 3 | ✗ No BFT | 2 = 2×1 |
| 2 | 2 | 4 | ✗ No BFT | 2 ≤ 2×2 |
| 1 | 2 | 3 | ✗ No BFT | 1 < 2×2 |

---

## Test Scenarios

### Scenario 1: Verify Basic Functionality
```bash
./demo.sh 3
```
**Goal**: Ensure blockchain works correctly with honest nodes
**Expected**: All nodes maintain identical chain state

### Scenario 2: Test BFT Resilience
```bash
./demo-bft.sh 4 1
```
**Goal**: Verify network handles Byzantine faults
**Expected**: Honest nodes reject invalid blocks, maintain consensus

### Scenario 3: Demonstrate BFT Failure
```bash
./demo-bft.sh 2 2
```
**Goal**: Show what happens without BFT threshold
**Expected**: Network may accept invalid blocks, lose consensus

### Scenario 4: Stress Test BFT
```bash
./demo-bft.sh 6 2
```
**Goal**: Test strong BFT with multiple malicious nodes
**Expected**: Network easily handles Byzantine faults

---

## Integration Tests

Run comprehensive integration tests:

```bash
# All tests
go test -v ./internal

# Specific test
go test -v ./internal -run TestFullNodeBlockProductionAndVerification
go test -v ./internal -run TestLeaderReplicaCommunication
go test -v ./internal -run TestLedgerPersistenceAndRecovery
```

**Test Coverage**:
1. Full node initialization and block production
2. Leader-replica communication and block propagation
3. Ledger persistence and recovery after restart

---

## Verification Commands

### Check Chain Heights
```bash
# Honest nodes (should match)
sqlite3 replica1.db "SELECT COUNT(*) FROM blocks;"
sqlite3 replica2.db "SELECT COUNT(*) FROM blocks;"

# Malicious nodes (may differ)
sqlite3 malicious1.db "SELECT COUNT(*) FROM blocks;"
```

### View Block Details
```bash
sqlite3 leader.db "SELECT block_height, slot, timestamp FROM blocks ORDER BY block_height;"
```

### Check Chain Integrity
```bash
sqlite3 replica1.db "SELECT block_height, hex(merkle_root), hex(previous_hash) FROM blocks ORDER BY block_height;"
```

### Monitor Block Production Rate
```bash
watch -n 1 'sqlite3 leader.db "SELECT COUNT(*) as total_blocks FROM blocks;"'
```

---

## tmux Navigation

| Action | Command |
|--------|---------|
| Switch panes | `Ctrl+B` then arrow keys |
| Zoom pane | `Ctrl+B` then `Z` |
| Detach session | `Ctrl+B` then `D` |
| Reattach | `tmux attach -t poh-demo` or `tmux attach -t poh-bft-demo` |
| Scroll in pane | `Ctrl+B` then `[`, use arrows, press `q` to exit |
| Kill pane | `Ctrl+B` then `X` |

---

## Log Patterns to Watch

### Honest Node Success
```
✓ Block passed consensus validation
✓ Block passed verification
✓ Block linkage verified
✓ Block stored successfully
```

### Honest Node Rejection
```
✗ Block validation failed (consensus)
✗ Block verification failed
✗ Block linkage verification failed
```

### Malicious Node Activity
```
⚠ MALICIOUS: Corrupted block X
⚠ MALICIOUS: Skipping validation
⚠ MALICIOUS: Ignoring verification failure
⚠ MALICIOUS: Stored block without validation
```

---

## Performance Metrics

### Block Production
- **Rate**: ~2.5 blocks/second (400ms slots)
- **Size**: Minimum 64 ticks per block (800,000 hashes)
- **Latency**: <10ms on localhost

### Network
- **Protocol**: TCP with length-prefixed messages
- **Serialization**: JSON
- **Propagation**: Broadcast to all connected peers

### Storage
- **Database**: SQLite
- **Tables**: blocks, entries, transactions
- **Indexes**: block_height, block_hash, slot

---

## Documentation

- **README.md**: Project overview and quick start
- **docs/guides/demo.md**: Detailed demo guide with tmux commands
- **docs/testing/bft-testing.md**: In-depth BFT theory and testing guide
- **docs/testing/testing-summary.md**: This file - quick reference

---

## Quick Reference

```bash
# Start basic demo
./demo.sh 3

# Start BFT demo with BFT
./demo-bft.sh 4 1

# Start BFT demo without BFT
./demo-bft.sh 2 2

# Stop any demo
./stop-demo.sh

# Run tests
go test -v ./internal

# Build manually
go build -o bin/poh-node cmd/main.go

# Check database
sqlite3 replica1.db "SELECT * FROM blocks;"
```

---

## Troubleshooting

### "tmux: command not found"
```bash
# Ubuntu/Debian
sudo apt-get install tmux

# macOS
brew install tmux
```

### "address already in use"
```bash
./stop-demo.sh
# Or manually kill processes using ports 8000-8009
```

### Nodes not connecting
- Ensure leader starts first (scripts handle this)
- Check firewall settings
- Verify ports 8000-8009 are available

### Database locked
```bash
# Stop all nodes first
./stop-demo.sh
# Then clean up
rm -f *.db
```

---

## Next Steps

1. Run basic demo to understand normal operation
2. Run BFT demo with BFT to see fault tolerance
3. Run BFT demo without BFT to see failure modes
4. Inspect databases to verify chain state
5. Review logs to understand validation process
6. Experiment with different node configurations
7. Read BFT-TESTING.md for deeper understanding

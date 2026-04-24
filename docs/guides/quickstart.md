# Quick Start Guide

## 30-Second Demo

```bash
# Run a basic blockchain network
./demo.sh
```

That's it! You'll see a leader producing blocks and 2 replicas validating them in real-time.

Press `Ctrl+C` in each pane to stop, or run `./stop-demo.sh`.

---

## 2-Minute BFT Test

```bash
# Test Byzantine Fault Tolerance
./demo-bft.sh 3 1
```

Watch as:
- 3 honest nodes maintain consensus
- 1 malicious node tries to corrupt blocks
- Honest nodes reject invalid blocks
- Network preserves integrity

---

## What You'll See

### Leader Pane (Top)
```
Leader node: Producing block for slot 42
Leader node: Block produced - height=15, slot=42, entries=64
Leader node: Block broadcasted to peers - height=15
```

### Honest Replica Panes (Green)
```
Replica node: Received block - height=15, slot=42, entries=64
Replica node: Block passed verification - height=15
Replica node: Block stored successfully - height=15
```

### Malicious Replica Panes (Red) - BFT Demo Only
```
MALICIOUS: Corrupted block 9 by setting invalid hash count
Replica node: Block verification failed: entry hash does not match
MALICIOUS: Ignoring verification failure, storing anyway
```

---

## Understanding the Output

**Block Production**: Leader creates a new block every ~400ms
**Block Height**: Sequential number of blocks in the chain
**Slot**: Time window for block production
**Entries**: Number of PoH entries in the block (minimum 64)

---

## Try These Commands

### Basic Demo with More Nodes
```bash
./demo.sh 5    # 1 leader + 5 replicas
```

### BFT Demo with Strong Tolerance
```bash
./demo-bft.sh 6 2    # 6 honest + 2 malicious (strong BFT)
```

### BFT Demo without Tolerance
```bash
./demo-bft.sh 2 2    # 2 honest + 2 malicious (NO BFT - will fail)
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

After running for 30 seconds, check the blockchain:

```bash
# How many blocks were created?
sqlite3 leader.db "SELECT COUNT(*) FROM blocks;"

# View the blocks
sqlite3 leader.db "SELECT block_height, slot FROM blocks ORDER BY block_height LIMIT 10;"
```

---

## Stop Everything

```bash
./stop-demo.sh
```

Choose `y` to clean up database files, or `n` to keep them for inspection.

---

## What's Next?

1. **Read DEMO.md** - Detailed demo guide
2. **Read BFT-TESTING.md** - Understand Byzantine Fault Tolerance
3. **Read TESTING-SUMMARY.md** - Complete testing reference
4. **Run tests** - `go test -v ./internal`
5. **Explore code** - Check out `internal/` directory

---

## Need Help?

- **Demo not starting?** Make sure tmux is installed: `sudo apt-get install tmux`
- **Port conflicts?** Run `./stop-demo.sh` first
- **Want to learn more?** Check the documentation files listed above

---

## The Big Picture

This is a **Proof of History (PoH)** blockchain inspired by Solana:

- **PoH Clock**: Creates a verifiable timeline using sequential SHA-256 hashing
- **Leader**: Produces blocks with PoH proofs
- **Replicas**: Validate and store blocks independently
- **BFT**: Network tolerates malicious nodes if honest > 2×malicious

The demo shows all of this working in real-time!

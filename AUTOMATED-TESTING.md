# Automated Testing Guide

This guide covers the automated testing tools designed for AI agents and developers who want to run tests without tmux.

## Overview

Three automated scripts are available:

1. **demo-automated.sh** - Run a single BFT test scenario
2. **test-launcher.sh** - Run multiple test scenarios sequentially
3. **analyze-results.sh** - Analyze results from completed tests

## demo-automated.sh

### Purpose
Run a blockchain network with configurable honest and malicious nodes, with clear log prefixes for easy analysis.

### Usage
```bash
./demo-automated.sh [num_honest] [num_malicious] [duration_seconds]
```

### Examples
```bash
# 2 honest + 1 malicious, run for 10 seconds
./demo-automated.sh 2 1 10

# 4 honest + 1 malicious, run for 30 seconds
./demo-automated.sh 4 1 30

# 3 honest + 2 malicious, run for 20 seconds
./demo-automated.sh 3 2 20
```

### Features
- **No tmux required** - Runs in background with clear output
- **Prefixed logs** - Each node has a clear prefix:
  - `[LEADER]` - Leader node (blue)
  - `[HONEST-1]`, `[HONEST-2]` - Honest replicas (green)
  - `[MALICIOUS-1]`, `[MALICIOUS-2]` - Malicious replicas (red)
- **Automatic analysis** - Shows results when test completes
- **Log files** - Saves logs to `logs/` directory
- **Database files** - Saves databases for inspection

### Output Example
```
[LEADER] Leader node: Producing block for slot 15
[LEADER] Leader node: Block produced - height=10, slot=15, entries=64
[LEADER] Leader node: Block stored to ledger - height=10
[HONEST-1] Replica node: Received block - height=10, slot=15, entries=64
[HONEST-1] Replica node: Block passed verification - height=10
[HONEST-1] Replica node: Block stored successfully - height=10
[MALICIOUS-1] MALICIOUS: Skipping validation for block 10
[MALICIOUS-1] MALICIOUS: Stored block 10 without validation
```

### Results Analysis
After the test completes, you'll see:
- Block counts for each node
- Validation failures for honest nodes
- Malicious actions detected
- BFT verdict (whether network maintained integrity)

---

## test-launcher.sh

### Purpose
Run a comprehensive test suite with multiple BFT scenarios automatically.

### Usage
```bash
./test-launcher.sh [duration_per_test]
```

### Example
```bash
# Run all tests, 20 seconds each
./test-launcher.sh 20

# Run all tests, 30 seconds each (default)
./test-launcher.sh 30
```

### Test Scenarios

The launcher runs 8 different scenarios:

1. **Baseline** - 4 honest + 0 malicious (all honest)
2. **Strong BFT** - 4 honest + 1 malicious
3. **Minimal BFT** - 3 honest + 1 malicious
4. **Strong BFT** - 5 honest + 2 malicious
5. **Edge Case** - 2 honest + 1 malicious (no BFT)
6. **No BFT** - 2 honest + 2 malicious
7. **Robust BFT** - 6 honest + 2 malicious
8. **Compromised** - 1 honest + 2 malicious

### Features
- **Sequential execution** - Runs all tests one after another
- **Automatic cleanup** - Cleans up between tests
- **Markdown report** - Generates detailed report
- **Archived results** - Saves logs and databases for each test
- **Summary statistics** - Shows quick overview at the end

### Output
- **Console output** - Real-time progress and results
- **Report file** - `test-report-YYYYMMDD-HHMMSS.md`
- **Test artifacts** - `test-results/test_N/` directories

### Report Contents
- Configuration for each test
- BFT status calculation
- Block counts for all nodes
- Consistency analysis
- Verdict for each scenario
- Summary of key findings

---

## analyze-results.sh

### Purpose
Analyze blockchain state and logs after running a demo.

### Usage
```bash
# Run after demo-automated.sh completes
./analyze-results.sh
```

### What It Analyzes

#### Database Analysis
- Block counts for each node
- Chain height
- Last 5 blocks produced
- Chain continuity (gaps detection)

#### Consistency Check
- Compares heights across honest nodes
- Identifies inconsistencies
- Flags unexpected behavior

#### Log Analysis
- Counts blocks stored by each node
- Counts blocks rejected by honest nodes
- Counts malicious actions
- Shows sample log entries

#### BFT Verdict
- Calculates if network has BFT
- Determines if honest nodes maintained consistency
- Provides final verdict

### Output Example
```
╔════════════════════════════════════════╗
║  PoH Blockchain Results Analyzer      ║
╚════════════════════════════════════════╝

=== Database Analysis ===

[LEADER]
  Blocks: 25
  Height: 25
  Last 5 blocks:
    25: slot=32, entries=64
    24: slot=31, entries=64

[HONEST-1]
  Blocks: 20
  Height: 20
  ✓ Chain continuous

[MALICIOUS-1]
  Blocks: 23
  Height: 23
  ⚠ Chain gaps: 3 (expected for malicious)

=== Consistency Check ===

✓ All honest nodes have consistent height: 20

=== BFT Verdict ===

Network composition:
  Honest nodes: 2
  Malicious nodes: 1

✗ Network lacks BFT (2 ≤ 2×1)
⚠ VERDICT: NO BFT - network compromised
```

---

## Workflow for AI Agents

### Quick Test
```bash
# Run a single test
./demo-automated.sh 3 1 15

# Analyze results
./analyze-results.sh
```

### Comprehensive Testing
```bash
# Run full test suite
./test-launcher.sh 20

# View report
cat test-report-*.md

# Inspect specific test
ls test-results/test_3/
cat test-results/test_3/honest_1.log
```

### Custom Analysis
```bash
# Check specific database
sqlite3 leader.db "SELECT * FROM blocks ORDER BY block_height DESC LIMIT 5;"

# Search logs for patterns
grep "MALICIOUS:" logs/malicious_1.log | wc -l
grep "verification failed" logs/honest_1.log | wc -l

# Compare block counts
for db in replica*.db; do
    echo "$db: $(sqlite3 $db 'SELECT COUNT(*) FROM blocks;')"
done
```

---

## Understanding Results

### Successful BFT Test
**Indicators:**
- Honest nodes have identical block counts
- Honest nodes rejected invalid blocks
- Malicious nodes have different/higher block counts
- Network maintained consensus

**Example:**
```
✓ Network has BFT (4 > 2×1)
✓ Honest nodes maintained consistency
✓ VERDICT: BFT SUCCESSFUL
```

### Failed BFT Test
**Indicators:**
- Honest nodes have inconsistent block counts
- Some invalid blocks were accepted
- Network state diverged

**Example:**
```
✗ Network lacks BFT (2 ≤ 2×2)
✗ Honest nodes inconsistent
✗ VERDICT: NO BFT - network compromised
```

### Expected Malicious Behavior
- **Skipping validation** - "MALICIOUS: Skipping validation for block X"
- **Storing invalid blocks** - "MALICIOUS: Stored block X without validation"
- **Ignoring failures** - "MALICIOUS: Ignoring verification failure"
- **Chain gaps** - Malicious nodes may have gaps in their chain

### Expected Honest Behavior
- **Rejecting invalid blocks** - "Block verification failed"
- **Consistent state** - All honest nodes have same height
- **Proper validation** - "Block passed verification"

---

## Troubleshooting

### Nodes Not Starting
```bash
# Check if ports are in use
lsof -i :8000-8009

# Kill stuck processes
pkill -f poh-node

# Clean up and retry
rm -f *.db logs/*.log
./demo-automated.sh 2 1 10
```

### No Blocks Stored
**Possible causes:**
- Test duration too short (increase to 15-30 seconds)
- Timing issues (replicas receiving blocks before genesis)
- Database errors (check logs for errors)

### Script Hangs
```bash
# Force kill everything
pkill -9 -f poh-node
pkill -9 -f demo-automated

# Clean up
rm -f *.db logs/*.log
```

### Inconsistent Results
- Run tests for longer duration (30+ seconds)
- Ensure clean state between tests
- Check system load (high load affects timing)

---

## Performance Considerations

### Test Duration
- **Short (5-10s)**: Quick validation, may have timing issues
- **Medium (15-30s)**: Recommended for most tests
- **Long (60s+)**: Comprehensive testing, more stable results

### System Resources
- Each node uses ~10-20MB RAM
- CPU usage depends on PoH hashing
- Disk I/O for SQLite writes
- Network bandwidth minimal (localhost)

### Recommended Limits
- **Max nodes**: 10 total (1 leader + 9 replicas)
- **Max concurrent tests**: 1 (sequential only)
- **Max test duration**: 120 seconds

---

## Integration with CI/CD

### Example GitHub Actions
```yaml
- name: Run BFT Tests
  run: |
    ./test-launcher.sh 15
    
- name: Upload Test Report
  uses: actions/upload-artifact@v2
  with:
    name: bft-test-report
    path: test-report-*.md
```

### Example GitLab CI
```yaml
test:bft:
  script:
    - ./test-launcher.sh 20
  artifacts:
    paths:
      - test-report-*.md
      - test-results/
```

---

## Advanced Usage

### Custom Test Scenarios
```bash
# Test extreme Byzantine fault
./demo-automated.sh 1 3 30

# Test large honest majority
./demo-automated.sh 8 1 30

# Test equal split (no BFT)
./demo-automated.sh 3 3 30
```

### Parallel Analysis
```bash
# Run test in background
./demo-automated.sh 4 2 60 &
TEST_PID=$!

# Do other work...

# Wait for completion
wait $TEST_PID

# Analyze
./analyze-results.sh
```

### Custom Reporting
```bash
# Extract specific metrics
LEADER_BLOCKS=$(sqlite3 leader.db "SELECT COUNT(*) FROM blocks;")
HONEST_BLOCKS=$(sqlite3 replica1.db "SELECT COUNT(*) FROM blocks;")
MALICIOUS_ACTIONS=$(grep -c "MALICIOUS:" logs/malicious_1.log)

echo "Leader: $LEADER_BLOCKS, Honest: $HONEST_BLOCKS, Malicious: $MALICIOUS_ACTIONS"
```

---

## Summary

These automated tools provide:
- **Easy testing** without manual tmux management
- **Clear output** with prefixed logs
- **Comprehensive analysis** of BFT behavior
- **Automated reporting** for multiple scenarios
- **AI-friendly** output for programmatic analysis

Perfect for CI/CD, automated testing, and AI agent analysis of blockchain behavior.

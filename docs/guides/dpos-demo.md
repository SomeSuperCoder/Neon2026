# DPoS Demo Guide

## Overview

The audit script includes a comprehensive DPoS (Delegated Proof of Stake) lifecycle test as Phase 4. This phase exercises the full DPoS lifecycle including genesis initialization, block production, stake delegation, epoch boundaries, reward distribution, and slashing.

## Usage

```bash
# Run full audit including DPoS phase
./audit.sh

# Run with custom settings
./audit.sh --duration 60 --validators 5

# Run in CI mode
./audit.sh --ci
```

### Parameters

- `--duration SECONDS`: Duration for each test phase including DPoS (default: 30)
- `--validators N`: Number of validators for DPoS tests (default: 3)
- `--ci`: CI mode with no colors or interactive prompts
- `--help`: Show help message

### Examples

```bash
# Run full audit with default settings (30s per phase, 3 validators)
./audit.sh

# Run with longer duration and more validators
./audit.sh --duration 60 --validators 5

# Run in CI mode for automated testing
./audit.sh --ci
```

## DPoS Test Phase

The audit script's Phase 4 tests the complete DPoS lifecycle:

### Phase 4: DPoS Lifecycle

The DPoS phase includes four sub-phases:

### Sub-Phase 1: Genesis & Block Production

- Starts N validator nodes from genesis configuration
- First node runs as leader, others as replicas
- Monitors block production in real-time
- Logs each block with slot number, validator ID, and block hash

### Sub-Phase 2: Stake Delegation

- Simulates delegation of stake to each validator
- Calculates total delegated stake across the network
- Logs delegation amounts in electrons and Neon

**Note:** This phase is currently simulated. Full implementation requires transaction submission via CLI.

### Sub-Phase 3: Epoch Boundary & Reward Distribution

- Calculates current epoch and slot numbers
- Simulates reward distribution based on blocks produced
- Applies validator commission (10%)
- Distributes remaining rewards to delegators proportionally

**Note:** Epoch boundaries occur every 432,000 slots (~2 days at 400ms/slot). This phase is simulated for demo purposes.

### Sub-Phase 4: Slashing

- Simulates a double-sign slashing event
- Reduces validator stake by 5%
- Checks if validator should be deactivated (stake below 1 Neon threshold)
- Transfers slashed amount to reward pool

**Note:** This phase is currently simulated. Full implementation requires double-sign proof submission.

## Output Format

### Console Output

The audit script provides structured output for the DPoS phase:

```
=== Phase 4: DPoS Lifecycle ===
Starting 3 validators from genesis...
Running for 30 seconds...

Sub-phase 1: Genesis & Block Production
  ✓ Validators started
  ✓ Blocks produced: 150

Sub-phase 2: Stake Delegation
  ✓ Delegation simulated
  ✓ Total stake: 15 Neon

Sub-phase 3: Epoch & Rewards
  ✓ Epoch boundary simulated
  ✓ Rewards distributed

Sub-phase 4: Slashing
  ✓ Slashing event simulated
  ✓ Validator stake reduced

✓ Phase 4 complete
```

### JSON Report

The audit script generates a comprehensive JSON report at `logs/audit-TIMESTAMP.json`:

```json
{
  "audit_timestamp": "20260427-120000",
  "configuration": {
    "duration_seconds": 30,
    "num_validators": 3,
    "ci_mode": false
  },
  "phases": {
    "dpos_lifecycle": {
      "status": "passed",
      "duration_seconds": 30,
      "validators": 3,
      "total_blocks": 150,
      "delegation_simulated": true,
      "epoch_simulated": true,
      "slashing_simulated": true,
      "metrics": {
        "blocks_per_validator": 50,
        "total_stake_electrons": 15000000000,
        "reward_pool_electrons": 1000000000,
        "slashed_amount_electrons": 750000000
      }
    }
  },
  "summary": {
    "total_phases": 4,
    "passed_phases": 4,
    "failed_phases": 0,
    "overall_status": "passed",
    "total_duration_seconds": 120
  }
}
```

### Database Files

Temporary databases are created for the DPoS phase:
- `audit-validator-1.db`, `audit-validator-2.db`, etc.
- Automatically cleaned up after the phase completes

### Logs

Individual validator logs are saved to:
- `logs/audit-validator-1.log`
- `logs/audit-validator-2.log`
- etc.

## Exit Codes

The audit script uses the following exit codes:

- `0`: All phases passed including DPoS
- `1`: One or more phases failed
- `2`: Configuration error or build failure

## Success Criteria

The DPoS phase is considered successful when:

1. ✅ All validators produce blocks
2. ✅ Total blocks > 0
3. ✅ All validators have consistent chain height
4. ✅ All sub-phases complete without errors

## Troubleshooting

### Build Failures

If the build fails:
```bash
go mod download
go build -o bin/poh-node cmd/main.go
```

### Port Conflicts

If ports 8000-8009 are in use:
```bash
pkill -f "poh-node"
./audit.sh
```

### Database Locked

If you see "database is locked" errors:
```bash
pkill -f "poh-node"
rm -f audit-*.db
./audit.sh
```

### Inconsistent Chain Heights

Minor height differences (1-2 blocks) are normal due to timing. Large differences indicate a problem with block propagation.

## Cleanup

The audit script automatically cleans up after each phase:
- Stops all validator nodes
- Closes database connections
- Removes temporary databases
- Saves final JSON report

To manually clean up if needed:
```bash
pkill -f "poh-node"
rm -f audit-*.db
```

## Integration with CI/CD

The JSON report format is designed for CI/CD integration:

```bash
# Run audit in CI mode
./audit.sh --ci --duration 30 --validators 3

# Check exit code
if [ $? -eq 0 ]; then
  echo "All tests passed including DPoS"
else
  echo "Tests failed"
  exit 1
fi

# Parse JSON report
cat logs/audit-*.json | jq '.phases.dpos_lifecycle.status'
```

### Example GitHub Actions

```yaml
- name: Run DPoS Tests
  run: ./audit.sh --ci --duration 30 --validators 3
  
- name: Upload Test Report
  uses: actions/upload-artifact@v2
  with:
    name: audit-report
    path: logs/audit-*.json
```

### Example GitLab CI

```yaml
test:dpos:
  script:
    - ./audit.sh --ci --duration 30 --validators 3
  artifacts:
    paths:
      - logs/audit-*.json
```

## Future Enhancements

The following features are planned for future releases:

1. **Real Transaction Submission**: Replace simulated delegation with actual `DelegateStake` transactions
2. **Epoch Boundary Triggers**: Wait for actual epoch boundaries instead of simulating
3. **Real Slashing**: Submit actual `ReportDoubleSign` instructions with valid proofs
4. **Validator TUI Integration**: Launch the Validator TUI alongside the demo
5. **Network Partitioning**: Test DPoS behavior under network failures
6. **Dynamic Validator Set**: Add/remove validators during the demo

## Analyzing Results

### View JSON Report

```bash
# View full report
cat logs/audit-TIMESTAMP.json | jq '.'

# View DPoS phase only
cat logs/audit-TIMESTAMP.json | jq '.phases.dpos_lifecycle'

# Check overall status
cat logs/audit-TIMESTAMP.json | jq '.summary.overall_status'

# Extract metrics
cat logs/audit-TIMESTAMP.json | jq '.phases.dpos_lifecycle.metrics'
```

### Inspect Databases (during test)

If you need to inspect databases while the test is running:

```bash
# Check block count
sqlite3 audit-validator-1.db "SELECT COUNT(*) FROM blocks;"

# View recent blocks
sqlite3 audit-validator-1.db "SELECT block_height, slot FROM blocks ORDER BY block_height DESC LIMIT 5;"
```

## Related Documentation

- [Demo Guide](demo.md) - General demo and testing guide
- [Validator TUI Guide](validator-tui.md) - Real-time validator monitoring
- [DPoS Requirements](../../.kiro/specs/delegated-proof-of-stake/requirements.md)
- [DPoS Design](../../.kiro/specs/delegated-proof-of-stake/design.md)
- [CLI Usage Guide](cli-usage.md)
- [Automated Testing Guide](../testing/automated-testing.md)

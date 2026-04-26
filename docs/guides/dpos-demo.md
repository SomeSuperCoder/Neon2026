# DPoS Demo Guide

## Overview

The `demo-dpos.sh` script provides a comprehensive automated demonstration of the Delegated Proof of Stake (DPoS) consensus mechanism. It exercises the full DPoS lifecycle including genesis initialization, block production, stake delegation, epoch boundaries, reward distribution, and slashing.

## Usage

```bash
./demo-dpos.sh <num_validators> <duration_seconds>
```

### Parameters

- `num_validators`: Number of validator nodes to start (2-9, default: 3)
- `duration_seconds`: How long to run the demo in seconds (minimum: 10, default: 30)

### Examples

```bash
# Run with 3 validators for 30 seconds
./demo-dpos.sh 3 30

# Run with 5 validators for 60 seconds
./demo-dpos.sh 5 60

# Quick test with 2 validators for 10 seconds
./demo-dpos.sh 2 10
```

## Demo Phases

The demo executes four distinct phases:

### Phase 1: Genesis & Block Production

- Starts N validator nodes from genesis configuration
- First node runs as leader, others as replicas
- Monitors block production in real-time
- Logs each block with slot number, validator ID, and block hash

### Phase 2: Stake Delegation

- Simulates delegation of stake to each validator
- Calculates total delegated stake across the network
- Logs delegation amounts in electrons and Neon

**Note:** This phase is currently simulated. Full implementation requires transaction submission via CLI.

### Phase 3: Epoch Boundary & Reward Distribution

- Calculates current epoch and slot numbers
- Simulates reward distribution based on blocks produced
- Applies validator commission (10%)
- Distributes remaining rewards to delegators proportionally

**Note:** Epoch boundaries occur every 432,000 slots (~2 days at 400ms/slot). This phase is simulated for demo purposes.

### Phase 4: Slashing

- Simulates a double-sign slashing event
- Reduces validator stake by 5%
- Checks if validator should be deactivated (stake below 1 Neon threshold)
- Transfers slashed amount to reward pool

**Note:** This phase is currently simulated. Full implementation requires double-sign proof submission.

## Output

### Console Output

The demo provides rich console output with:

- Color-coded validator logs
- Real-time block production updates
- Phase completion summaries
- Human-readable summary tables showing:
  - Validator statistics (blocks produced, chain height, status)
  - Lifecycle phase results
  - Network metrics (total blocks, consistency, duration)
  - Overall demo status (success/failure)

### JSON Log

A structured JSON log is saved to `logs/dpos-demo-<timestamp>.json` containing:

```json
{
  "demo_type": "dpos",
  "timestamp": "20260426-120000",
  "num_validators": 3,
  "duration_seconds": 30,
  "phases": {
    "genesis_and_blocks": { ... },
    "delegation": { ... },
    "epoch_and_rewards": { ... },
    "slashing": { ... }
  },
  "summary": {
    "total_blocks": 150,
    "average_blocks_per_validator": 50,
    "consistency": "consistent",
    "phase_failures": 0,
    "overall_status": "success"
  }
}
```

### Database Files

Each validator's blockchain state is persisted to:
- `validator1.db` (leader)
- `validator2.db`, `validator3.db`, etc. (replicas)

### Node Logs

Individual validator logs are saved to:
- `logs/validator_1.log`
- `logs/validator_2.log`
- etc.

## Exit Codes

- `0`: Demo completed successfully (all phases passed, chain consistent)
- `1`: Demo failed (block production failed or chain inconsistent)

## Success Criteria

The demo is considered successful when:

1. ✅ All validators produce blocks
2. ✅ Total blocks > 0
3. ✅ All validators have consistent chain height
4. ✅ No phase failures

## Troubleshooting

### Build Failures

If the build fails:
```bash
go mod download
go build -o bin/poh-node cmd/main.go
```

### Port Conflicts

If ports 8000-8009 are in use, stop conflicting processes:
```bash
pkill -f "poh-node"
```

### Database Locked

If you see "database is locked" errors:
```bash
rm -f validator*.db
./demo-dpos.sh 3 30
```

### Inconsistent Chain Heights

Minor height differences (1-2 blocks) are normal due to timing. Large differences indicate a problem with block propagation.

## Cleanup

The demo automatically cleans up on exit (Ctrl+C or completion):
- Stops all validator nodes
- Closes database connections
- Saves final logs

To manually clean up:
```bash
pkill -f "poh-node"
rm -f validator*.db
```

## Integration with analyze-results.sh

The JSON log format is compatible with the existing `analyze-results.sh` script for automated analysis. The structured format enables:

- Machine parsing of demo results
- Automated testing and CI/CD integration
- Historical comparison of demo runs
- Performance metrics tracking

## Future Enhancements

The following features are planned for future releases:

1. **Real Transaction Submission**: Replace simulated delegation with actual `DelegateStake` transactions
2. **Epoch Boundary Triggers**: Wait for actual epoch boundaries instead of simulating
3. **Real Slashing**: Submit actual `ReportDoubleSign` instructions with valid proofs
4. **Validator TUI Integration**: Launch the Validator TUI alongside the demo
5. **Network Partitioning**: Test DPoS behavior under network failures
6. **Dynamic Validator Set**: Add/remove validators during the demo

## Related Documentation

- [Validator TUI Guide](validator-tui.md)
- [DPoS Requirements](.kiro/specs/delegated-proof-of-stake/requirements.md)
- [DPoS Design](.kiro/specs/delegated-proof-of-stake/design.md)
- [CLI Usage Guide](cli-usage.md)

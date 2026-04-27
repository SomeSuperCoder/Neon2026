# Automated Testing Guide

This guide covers the automated testing tools designed for AI agents, CI/CD pipelines, and developers who want to run comprehensive tests without manual intervention.

## Overview

Two main automated scripts are available:

1. **audit.sh** - Comprehensive test suite with four phases (consensus, BFT, DPoS)
2. **analyze-results.sh** - Analyze blockchain state and logs from devnet or audit runs

## audit.sh

### Purpose
Run a comprehensive blockchain test suite covering consensus, Byzantine Fault Tolerance, and DPoS lifecycle in a single automated execution.

### Usage
```bash
./audit.sh [OPTIONS]
```

### Options
- `--duration SECONDS`: Duration for each test phase (default: 30)
- `--validators N`: Number of validators for DPoS tests (default: 3)
- `--ci`: CI mode with no colors or interactive prompts
- `--help`: Show help message

### Examples
```bash
# Run with default settings (30s per phase, 3 validators)
./audit.sh

# Run with longer duration and more validators
./audit.sh --duration 60 --validators 5

# Run in CI mode for automated pipelines
./audit.sh --ci
```

### Test Phases

The audit script executes four sequential test phases:

#### Phase 1: Basic Consensus
- **Configuration**: 1 leader + 3 honest replicas
- **Duration**: Configurable (default 30s)
- **Tests**: Block production, chain consistency, consensus validation
- **Success Criteria**: All nodes produce blocks, consistent chain height

#### Phase 2: BFT With Tolerance
- **Configuration**: 1 leader + 4 honest + 1 malicious replica
- **Duration**: Configurable (default 30s)
- **Tests**: Byzantine fault handling, invalid block rejection
- **Success Criteria**: Honest nodes reject invalid blocks, maintain integrity

#### Phase 3: BFT Without Tolerance
- **Configuration**: 1 leader + 2 honest + 2 malicious replicas
- **Duration**: Configurable (default 30s)
- **Tests**: Network behavior under insufficient BFT
- **Success Criteria**: Demonstrates need for BFT threshold

#### Phase 4: DPoS Lifecycle
- **Configuration**: N validators from genesis (configurable, default 3)
- **Duration**: Configurable (default 30s)
- **Tests**: Delegation, epoch boundaries, rewards, slashing
- **Success Criteria**: All DPoS sub-phases complete successfully

### Features
- **No manual intervention** - Fully automated execution
- **Automatic cleanup** - Cleans up after each phase
- **JSON reports** - Machine-parseable output at `logs/audit-TIMESTAMP.json`
- **Exit codes** - 0=pass, 1=fail, 2=config error
- **CI-friendly** - `--ci` mode disables colors and prompts

### Output Format

#### Console Output
```
╔════════════════════════════════════════╗
║  PoH Blockchain Audit Suite           ║
╚════════════════════════════════════════╝

Configuration:
  Duration per phase: 30 seconds
  Validators (DPoS): 3
  CI mode: false

=== Phase 1: Basic Consensus ===
Starting 1 leader + 3 honest replicas...
Running for 30 seconds...
✓ Phase complete: 150 blocks produced

=== Phase 2: BFT With Tolerance ===
Starting 1 leader + 4 honest + 1 malicious...
Running for 30 seconds...
✓ Phase complete: Honest nodes rejected 15 invalid blocks

=== Phase 3: BFT Without Tolerance ===
Starting 1 leader + 2 honest + 2 malicious...
Running for 30 seconds...
✓ Phase complete: Network behavior observed

=== Phase 4: DPoS Lifecycle ===
Starting 3 validators from genesis...
Running for 30 seconds...
✓ Phase complete: DPoS lifecycle simulated

=== Audit Summary ===
Total phases: 4
Passed: 4
Failed: 0

Overall status: PASSED

Report saved to: logs/audit-20260427-120000.json
```

#### JSON Report Format

The audit script generates a comprehensive JSON report:

```json
{
  "audit_timestamp": "20260427-120000",
  "configuration": {
    "duration_seconds": 30,
    "num_validators": 3,
    "ci_mode": false
  },
  "phases": {
    "basic_consensus": {
      "status": "passed",
      "duration_seconds": 30,
      "total_blocks": 150,
      "consistency": "consistent",
      "metrics": {
        "blocks_per_validator": 50,
        "leader_blocks": 150,
        "replica_blocks": [50, 50, 50]
      }
    },
    "bft_with_tolerance": {
      "status": "passed",
      "duration_seconds": 30,
      "honest_nodes": 4,
      "malicious_nodes": 1,
      "blocks_rejected": 15,
      "malicious_actions_detected": 20,
      "metrics": {
        "honest_consistency": "consistent",
        "bft_status": "has_bft"
      }
    },
    "bft_without_tolerance": {
      "status": "passed",
      "duration_seconds": 30,
      "honest_nodes": 2,
      "malicious_nodes": 2,
      "consistency": "inconsistent",
      "metrics": {
        "bft_status": "no_bft",
        "compromise_indicators": 5
      }
    },
    "dpos_lifecycle": {
      "status": "passed",
      "duration_seconds": 30,
      "validators": 3,
      "total_blocks": 120,
      "delegation_simulated": true,
      "epoch_simulated": true,
      "slashing_simulated": true,
      "metrics": {
        "blocks_per_validator": 40,
        "total_stake_electrons": 15000000000
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

### Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed
- `2`: Configuration error or build failure

Use these exit codes in CI/CD pipelines to determine test success.

---

## analyze-results.sh

### Purpose
Analyze blockchain state and logs from devnet or audit runs.

### Usage
```bash
# Analyze audit results (default)
./analyze-results.sh

# Analyze devnet databases
./analyze-results.sh --db-dir devnet-data
```

### What It Analyzes

#### Database Analysis
- Block counts for each validator
- Chain height
- Last 5 blocks produced
- Chain continuity (gaps detection)

#### Consistency Check
- Compares heights across all nodes
- Identifies inconsistencies
- Flags unexpected behavior

#### Summary
- Total blocks across network
- Average blocks per validator
- Overall consistency status

### Output Example
```
╔════════════════════════════════════════╗
║  PoH Blockchain Results Analyzer      ║
╚════════════════════════════════════════╝

=== Database Analysis ===

[VALIDATOR-1]
  Blocks: 150
  Height: 150
  Last 5 blocks:
    150: slot=149, entries=64
    149: slot=148, entries=64
  ✓ Chain continuous

[VALIDATOR-2]
  Blocks: 150
  Height: 150
  ✓ Chain continuous

[VALIDATOR-3]
  Blocks: 150
  Height: 150
  ✓ Chain continuous

=== Consistency Check ===

✓ All validators have consistent height: 150

=== Summary ===

Total blocks: 150
Average per validator: 50
Consistency: CONSISTENT
```

---

## Workflow for CI/CD

### Quick Audit Test
```bash
# Run audit in CI mode
./audit.sh --ci --duration 30 --validators 3

# Check exit code
if [ $? -eq 0 ]; then
  echo "All tests passed"
else
  echo "Tests failed"
  exit 1
fi
```

### Parse JSON Results
```bash
# Extract overall status
cat logs/audit-*.json | jq '.summary.overall_status'

# Extract phase results
cat logs/audit-*.json | jq '.phases | to_entries[] | {phase: .key, status: .value.status}'

# Check for failures
FAILED=$(cat logs/audit-*.json | jq '.summary.failed_phases')
if [ "$FAILED" -gt 0 ]; then
  echo "Failed phases: $FAILED"
  exit 1
fi
```

### Analyze Devnet State
```bash
# Start devnet
./devnet.sh start 3

# Run for some time...
sleep 60

# Analyze state
./analyze-results.sh --db-dir devnet-data

# Stop devnet
./devnet.sh stop
```
---

## Understanding Audit Results

### Successful Audit
**Indicators:**
- All four phases pass
- Consistent block production across phases
- BFT phases demonstrate expected behavior
- DPoS lifecycle completes successfully

**Example:**
```json
{
  "summary": {
    "total_phases": 4,
    "passed_phases": 4,
    "failed_phases": 0,
    "overall_status": "passed"
  }
}
```

### Failed Audit
**Indicators:**
- One or more phases fail
- Inconsistent block production
- Unexpected network behavior

**Example:**
```json
{
  "summary": {
    "total_phases": 4,
    "passed_phases": 3,
    "failed_phases": 1,
    "overall_status": "failed"
  }
}
```

### Phase-Specific Results

#### Basic Consensus
- All nodes should produce blocks
- Chain heights should be consistent
- No validation failures

#### BFT With Tolerance
- Honest nodes reject invalid blocks
- Malicious blocks are not stored by honest nodes
- Network maintains integrity

#### BFT Without Tolerance
- Demonstrates insufficient BFT threshold
- May show inconsistent state (expected)
- Validates BFT requirement calculation

#### DPoS Lifecycle
- All sub-phases complete
- Validators produce blocks
- Lifecycle events simulated successfully

---

## Troubleshooting

### Build Failures
```bash
# Update dependencies
go mod download

# Rebuild
go build -o bin/poh-node cmd/main.go

# Retry audit
./audit.sh
```

### Port Conflicts
```bash
# Check ports in use
lsof -i :8000-8009

# Kill stuck processes
pkill -f poh-node

# Clean up and retry
./audit.sh
```

### Database Errors
```bash
# Clean up old databases
rm -f audit-*.db devnet-data/*.db

# Retry
./audit.sh
```

### Timing Issues
- Increase test duration: `./audit.sh --duration 60`
- Reduce system load
- Check for resource constraints

---

## Performance Considerations

### Test Duration
- **Short (15-30s)**: Quick validation, suitable for CI
- **Medium (30-60s)**: Recommended for comprehensive testing
- **Long (60s+)**: Thorough testing, more stable results

### System Resources
- Each validator uses ~10-20MB RAM
- CPU usage depends on PoH hashing
- Disk I/O for SQLite writes
- Network bandwidth minimal (localhost)

### Recommended Settings
- **CI/CD**: `--duration 30 --validators 3 --ci`
- **Development**: `--duration 60 --validators 5`
- **Comprehensive**: `--duration 120 --validators 7`

---

## Integration with CI/CD

### Example GitHub Actions
```yaml
name: Blockchain Audit

on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23
      
      - name: Run Audit
        run: ./audit.sh --ci --duration 30 --validators 3
      
      - name: Upload Report
        if: always()
        uses: actions/upload-artifact@v2
        with:
          name: audit-report
          path: logs/audit-*.json
      
      - name: Check Results
        run: |
          STATUS=$(cat logs/audit-*.json | jq -r '.summary.overall_status')
          if [ "$STATUS" != "passed" ]; then
            echo "Audit failed"
            exit 1
          fi
```

### Example GitLab CI
```yaml
audit:
  stage: test
  script:
    - ./audit.sh --ci --duration 30 --validators 3
  artifacts:
    when: always
    paths:
      - logs/audit-*.json
    reports:
      junit: logs/audit-*.json
  rules:
    - if: '$CI_PIPELINE_SOURCE == "merge_request_event"'
    - if: '$CI_COMMIT_BRANCH == "main"'
```

### Example Jenkins Pipeline
```groovy
pipeline {
  agent any
  stages {
    stage('Audit') {
      steps {
        sh './audit.sh --ci --duration 30 --validators 3'
      }
    }
    stage('Analyze') {
      steps {
        sh '''
          STATUS=$(cat logs/audit-*.json | jq -r '.summary.overall_status')
          if [ "$STATUS" != "passed" ]; then
            exit 1
          fi
        '''
      }
    }
  }
  post {
    always {
      archiveArtifacts artifacts: 'logs/audit-*.json', fingerprint: true
    }
  }
}
```

---

## Advanced Usage

### Custom Test Scenarios

Run audit with specific configurations:
```bash
# Long-running comprehensive test
./audit.sh --duration 120 --validators 7

# Quick smoke test
./audit.sh --duration 15 --validators 2

# CI-optimized test
./audit.sh --ci --duration 20 --validators 3
```

### Analyzing Specific Phases

Extract phase-specific results from JSON:
```bash
# Check basic consensus
cat logs/audit-*.json | jq '.phases.basic_consensus'

# Check BFT with tolerance
cat logs/audit-*.json | jq '.phases.bft_with_tolerance'

# Check DPoS lifecycle
cat logs/audit-*.json | jq '.phases.dpos_lifecycle'

# Get all metrics
cat logs/audit-*.json | jq '.phases | to_entries[] | {phase: .key, metrics: .value.metrics}'
```

### Comparing Multiple Runs

```bash
# Run multiple audits
for i in {1..5}; do
  ./audit.sh --ci --duration 30 --validators 3
  sleep 5
done

# Compare results
for report in logs/audit-*.json; do
  echo "Report: $report"
  jq '.summary' "$report"
done
```

### Integration with Monitoring

```bash
# Export metrics to Prometheus format
cat logs/audit-*.json | jq -r '
  "audit_total_phases \(.summary.total_phases)",
  "audit_passed_phases \(.summary.passed_phases)",
  "audit_failed_phases \(.summary.failed_phases)",
  "audit_duration_seconds \(.summary.total_duration_seconds)"
'

# Send to monitoring system
curl -X POST http://monitoring-system/metrics \
  -d @logs/audit-latest.json
```

---

## Summary

The audit script provides:
- **Comprehensive testing** of all blockchain functionality
- **Automated execution** without manual intervention
- **Machine-parseable output** for CI/CD integration
- **Four test phases** covering consensus, BFT, and DPoS
- **Exit codes** for pipeline integration
- **JSON reports** for analysis and monitoring

Perfect for CI/CD pipelines, automated testing, and continuous validation of blockchain functionality.

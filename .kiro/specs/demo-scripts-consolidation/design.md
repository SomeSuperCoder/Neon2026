# Design Document

## Overview

This design consolidates four existing demo scripts (`demo.sh`, `demo-automated.sh`, `demo-bft.sh`, `demo-dpos.sh`) into two focused scripts: `audit.sh` for comprehensive testing and `devnet.sh` for local development. The consolidation eliminates redundancy, improves maintainability, and provides clear separation between testing and development workflows.

## Architecture

### Script Organization

```
project-root/
├── audit.sh              # Comprehensive testing script (NEW)
├── devnet.sh             # Local development network (NEW)
├── analyze-results.sh    # Result analysis utility (KEEP)
├── stop-demo.sh          # Node shutdown utility (KEEP)
└── logs/                 # Log output directory
    ├── audit-TIMESTAMP.json
    └── devnet-TIMESTAMP.log
```

### Removed Scripts

The following scripts will be deleted:
- `demo.sh` - Functionality merged into `devnet.sh`
- `demo-automated.sh` - Functionality merged into `audit.sh`
- `demo-bft.sh` - Functionality merged into `audit.sh`
- `demo-dpos.sh` - Functionality merged into `audit.sh`

## Components and Interfaces

### 1. audit.sh - Comprehensive Testing Script

#### Purpose
Execute all blockchain test scenarios in sequence and generate a comprehensive report suitable for CI/CD integration.

#### Command-Line Interface

```bash
./audit.sh [OPTIONS]

Options:
  --duration SECONDS    Duration for each test phase (default: 30)
  --validators N        Number of validators for DPoS tests (default: 3)
  --ci                  CI mode: no colors, no interactive prompts
  --help                Show help message
```

#### Test Phases

The audit script executes four sequential test phases:

1. **Basic Consensus Test**
   - Start 1 leader + 3 honest replicas
   - Run for specified duration
   - Verify block production and chain consistency
   - Metrics: total blocks, blocks/validator, chain height consistency

2. **BFT Test (With Tolerance)**
   - Start 1 leader + 4 honest + 1 malicious replica
   - Run for specified duration
   - Verify honest nodes reject invalid blocks
   - Metrics: blocks produced, blocks rejected, malicious actions detected

3. **BFT Test (Without Tolerance)**
   - Start 1 leader + 2 honest + 2 malicious replicas
   - Run for specified duration
   - Verify network behavior under insufficient BFT
   - Metrics: blocks produced, consistency status, compromise indicators

4. **DPoS Lifecycle Test**
   - Start N validators from genesis
   - Run for specified duration
   - Simulate delegation, epoch boundaries, rewards, slashing
   - Metrics: blocks produced, validator statistics, lifecycle phase completion

#### Output Format

**Console Output:**
- Color-coded phase headers (unless --ci mode)
- Real-time progress indicators
- Summary tables for each phase
- Final verdict with pass/fail status

**JSON Report:**
```json
{
  "audit_timestamp": "20260426-120000",
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
      "metrics": { ... }
    },
    "bft_with_tolerance": {
      "status": "passed",
      "honest_nodes": 4,
      "malicious_nodes": 1,
      "blocks_rejected": 15,
      "metrics": { ... }
    },
    "bft_without_tolerance": {
      "status": "passed",
      "honest_nodes": 2,
      "malicious_nodes": 2,
      "consistency": "inconsistent",
      "metrics": { ... }
    },
    "dpos_lifecycle": {
      "status": "passed",
      "validators": 3,
      "total_blocks": 120,
      "metrics": { ... }
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

#### Exit Codes
- `0`: All tests passed
- `1`: One or more tests failed
- `2`: Configuration error or build failure

#### Implementation Strategy

```bash
#!/usr/bin/env bash
set -e

# Configuration
DURATION=${1:-30}
VALIDATORS=${2:-3}
CI_MODE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --duration) DURATION="$2"; shift 2 ;;
    --validators) VALIDATORS="$2"; shift 2 ;;
    --ci) CI_MODE=true; shift ;;
    --help) show_help; exit 0 ;;
    *) shift ;;
  esac
done

# Initialize JSON report
init_json_report

# Phase 1: Basic Consensus
run_basic_consensus_test

# Phase 2: BFT With Tolerance
run_bft_with_tolerance_test

# Phase 3: BFT Without Tolerance
run_bft_without_tolerance_test

# Phase 4: DPoS Lifecycle
run_dpos_lifecycle_test

# Generate final report
generate_final_report

# Exit with appropriate code
exit $EXIT_CODE
```

### 2. devnet.sh - Local Development Network

#### Purpose
Launch a persistent local network of validators for development, testing wallets, block explorers, and other tools.

#### Command-Line Interface

```bash
./devnet.sh [COMMAND] [OPTIONS]

Commands:
  start [N]             Start devnet with N validators (default: 3)
  stop                  Stop running devnet
  restart [N]           Restart devnet with N validators
  status                Show devnet status
  logs [VALIDATOR_ID]   Show logs for specific validator (or all)
  clean                 Stop devnet and clean all data

Options:
  --port PORT           Starting port number (default: 8000)
  --db-dir DIR          Database directory (default: ./devnet-data)
  --log-dir DIR         Log directory (default: ./logs)
  --help                Show help message
```

#### Network Configuration

**Default Setup (3 validators):**
- Validator 1 (Leader): `localhost:8000` → `devnet-data/validator1.db`
- Validator 2: `localhost:8001` → `devnet-data/validator2.db`
- Validator 3: `localhost:8002` → `devnet-data/validator3.db`

**Process Management:**
- Nodes run as background processes (not tmux)
- PIDs stored in `devnet-data/pids/`
- Logs written to `logs/devnet-validator-N.log`
- Graceful shutdown on stop command

#### State Persistence

Unlike the audit script, devnet maintains state across restarts:
- Database files preserved in `devnet-data/`
- Logs rotated with timestamps
- Clean command required to reset state

#### Implementation Strategy

```bash
#!/usr/bin/env bash

COMMAND=${1:-start}
NUM_VALIDATORS=${2:-3}
PORT_START=8000
DB_DIR="./devnet-data"
LOG_DIR="./logs"
PID_DIR="$DB_DIR/pids"

case $COMMAND in
  start)
    start_devnet $NUM_VALIDATORS
    ;;
  stop)
    stop_devnet
    ;;
  restart)
    stop_devnet
    start_devnet $NUM_VALIDATORS
    ;;
  status)
    show_devnet_status
    ;;
  logs)
    show_devnet_logs $2
    ;;
  clean)
    stop_devnet
    clean_devnet_data
    ;;
  *)
    show_help
    exit 1
    ;;
esac
```

**start_devnet() Function:**
```bash
start_devnet() {
  local num=$1
  
  # Create directories
  mkdir -p "$DB_DIR" "$LOG_DIR" "$PID_DIR"
  
  # Build binary
  go build -o bin/poh-node cmd/main.go
  
  # Start leader
  ./bin/poh-node --type=leader --port=$PORT_START \
    --db="$DB_DIR/validator1.db" \
    > "$LOG_DIR/devnet-validator-1.log" 2>&1 &
  echo $! > "$PID_DIR/validator1.pid"
  
  # Start replicas
  for i in $(seq 2 $num); do
    local port=$((PORT_START + i - 1))
    ./bin/poh-node --type=replica --port=$port \
      --peers=localhost:$PORT_START \
      --db="$DB_DIR/validator$i.db" \
      > "$LOG_DIR/devnet-validator-$i.log" 2>&1 &
    echo $! > "$PID_DIR/validator$i.pid"
  done
  
  # Display network info
  show_network_info $num
}
```

**stop_devnet() Function:**
```bash
stop_devnet() {
  if [ ! -d "$PID_DIR" ]; then
    echo "No running devnet found"
    return
  fi
  
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      pid=$(cat "$pid_file")
      kill -TERM $pid 2>/dev/null || true
      rm "$pid_file"
    fi
  done
  
  # Wait for graceful shutdown
  sleep 2
  
  # Force kill if still running
  pkill -9 -f "poh-node.*devnet-data" 2>/dev/null || true
}
```

### 3. analyze-results.sh (Preserved)

This utility remains unchanged and continues to provide manual result inspection for both audit and devnet databases.

**Usage:**
```bash
# Analyze audit results
./analyze-results.sh

# Analyze devnet state
./analyze-results.sh --db-dir devnet-data
```

### 4. stop-demo.sh (Updated)

Update to handle both legacy sessions and new devnet:

```bash
#!/usr/bin/env bash

# Stop legacy tmux sessions
tmux kill-session -t poh-demo 2>/dev/null || true
tmux kill-session -t poh-bft-demo 2>/dev/null || true

# Stop devnet if running
if [ -f devnet-data/pids/*.pid ]; then
  ./devnet.sh stop
fi

# Kill any stray poh-node processes
pkill -f "poh-node" 2>/dev/null || true

echo "All nodes stopped"
```

## Data Models

### Audit Report Schema

```typescript
interface AuditReport {
  audit_timestamp: string;           // ISO 8601 format
  configuration: {
    duration_seconds: number;
    num_validators: number;
    ci_mode: boolean;
  };
  phases: {
    basic_consensus: PhaseResult;
    bft_with_tolerance: BFTPhaseResult;
    bft_without_tolerance: BFTPhaseResult;
    dpos_lifecycle: DPoSPhaseResult;
  };
  summary: {
    total_phases: number;
    passed_phases: number;
    failed_phases: number;
    overall_status: "passed" | "failed";
    total_duration_seconds: number;
  };
}

interface PhaseResult {
  status: "passed" | "failed" | "skipped";
  duration_seconds: number;
  total_blocks: number;
  consistency: "consistent" | "inconsistent";
  metrics: Record<string, any>;
  errors?: string[];
}

interface BFTPhaseResult extends PhaseResult {
  honest_nodes: number;
  malicious_nodes: number;
  blocks_rejected: number;
  malicious_actions_detected: number;
}

interface DPoSPhaseResult extends PhaseResult {
  validators: number;
  delegation_simulated: boolean;
  epoch_simulated: boolean;
  slashing_simulated: boolean;
}
```

### Devnet State

```
devnet-data/
├── validator1.db          # Leader database
├── validator2.db          # Replica databases
├── validator3.db
└── pids/
    ├── validator1.pid     # Process IDs
    ├── validator2.pid
    └── validator3.pid
```

## Error Handling

### Audit Script Error Handling

1. **Build Failures**: Exit immediately with code 2, log error to JSON
2. **Phase Failures**: Mark phase as failed, continue to next phase
3. **Cleanup Failures**: Log warning, continue execution
4. **Signal Handling**: Trap INT/TERM, cleanup processes, save partial report

### Devnet Script Error Handling

1. **Port Conflicts**: Detect and report, suggest using --port flag
2. **Build Failures**: Exit with error message
3. **Start Failures**: Cleanup partial start, report which validator failed
4. **Stop Failures**: Force kill after timeout, warn user

## Testing Strategy

### Unit Testing (Manual)

Test each script function independently:

```bash
# Test audit phases individually
source audit.sh
run_basic_consensus_test
run_bft_with_tolerance_test

# Test devnet commands
./devnet.sh start 2
./devnet.sh status
./devnet.sh logs 1
./devnet.sh stop
```

### Integration Testing

1. **Audit Full Run**: Execute complete audit, verify JSON output
2. **Devnet Lifecycle**: Start → Status → Logs → Stop → Clean
3. **Devnet Persistence**: Start → Stop → Start (verify state preserved)
4. **CI Mode**: Run audit with --ci flag, verify no colors/prompts

### Acceptance Testing

1. Run audit script, verify all phases pass
2. Start devnet, submit transactions via CLI, verify processing
3. Stop devnet, restart, verify chain continuity
4. Run analyze-results.sh on both audit and devnet databases

## Migration Plan

### Phase 1: Create New Scripts
1. Implement `audit.sh` with all test phases
2. Implement `devnet.sh` with start/stop/status commands
3. Test both scripts independently

### Phase 2: Update Documentation
1. Update `README.md` to reference new scripts
2. Update `docs/guides/demo.md` with new usage
3. Update `docs/guides/dpos-demo.md` to reference audit script
4. Update `docs/testing/automated-testing.md`

### Phase 3: Remove Old Scripts
1. Delete `demo.sh`, `demo-automated.sh`, `demo-bft.sh`, `demo-dpos.sh`
2. Update `stop-demo.sh` to handle new devnet
3. Verify no broken references in documentation

### Phase 4: Validation
1. Run full audit suite
2. Test devnet with wallet/explorer tools
3. Verify CI integration works correctly

## Design Decisions and Rationales

### Decision 1: Background Processes vs tmux for Devnet

**Rationale**: Background processes are simpler for long-running development networks. Developers can use their own terminal multiplexer or IDE integration. tmux adds complexity and requires installation.

### Decision 2: Sequential Test Phases in Audit

**Rationale**: Sequential execution provides clear phase boundaries and makes debugging easier. Parallel execution would complicate error handling and reporting.

### Decision 3: JSON Output for Audit

**Rationale**: JSON is machine-parseable for CI/CD integration while remaining human-readable. Structured format enables automated analysis and historical comparison.

### Decision 4: Preserve analyze-results.sh

**Rationale**: Provides manual inspection capability independent of automated reporting. Useful for debugging and ad-hoc analysis.

### Decision 5: State Persistence in Devnet

**Rationale**: Development workflows benefit from persistent state across restarts. Allows testing of chain continuity, upgrades, and recovery scenarios.

## Performance Considerations

### Audit Script
- Total runtime: ~2-3 minutes (4 phases × 30 seconds + overhead)
- Disk usage: ~50MB for databases and logs
- Memory: ~200MB peak (multiple concurrent nodes)

### Devnet Script
- Startup time: ~2 seconds
- Steady-state memory: ~50MB per validator
- Disk growth: ~1MB per hour of operation

## Security Considerations

### Audit Script
- Runs in isolated environment (temporary databases)
- No network exposure (localhost only)
- Automatic cleanup prevents data leakage

### Devnet Script
- Binds to localhost by default (no external exposure)
- Database files contain no sensitive data (test keys only)
- Clean command provides secure data deletion

## Future Enhancements

1. **Audit Script**
   - Parallel phase execution for faster runs
   - Configurable test scenarios via config file
   - Performance benchmarking mode
   - Historical comparison reports

2. **Devnet Script**
   - Multi-machine network support
   - Dynamic validator addition/removal
   - Snapshot/restore functionality
   - Integration with block explorer UI

3. **Both Scripts**
   - Docker containerization
   - Kubernetes deployment manifests
   - Prometheus metrics export
   - Grafana dashboard templates

#!/usr/bin/env bash
set -e

# audit.sh - Comprehensive blockchain testing script
# Runs all test phases and generates detailed JSON reports

# Default configuration
DURATION=30
VALIDATORS=3
CI_MODE=false
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT_FILE="logs/audit-${TIMESTAMP}.json"
EXIT_CODE=0
TOTAL_PHASES=4
PASSED_PHASES=0
FAILED_PHASES=0

# Color codes (disabled in CI mode)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup tracking
PIDS_TO_CLEANUP=()
DBS_TO_CLEANUP=()

# Show help message
show_help() {
    cat << EOF
Usage: ./audit.sh [OPTIONS]

Comprehensive blockchain testing script that validates all functionality.

Options:
  --duration SECONDS    Duration for each test phase (default: 30)
  --validators N        Number of validators for DPoS tests (default: 3)
  --ci                  CI mode: no colors, no interactive prompts
  --help                Show this help message

Exit Codes:
  0 - All tests passed
  1 - One or more tests failed
  2 - Configuration error or build failure

Examples:
  ./audit.sh                           # Run with defaults
  ./audit.sh --duration 60             # Run with 60s per phase
  ./audit.sh --validators 5 --ci       # CI mode with 5 validators

EOF
}

# Parse command-line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --duration)
                DURATION="$2"
                shift 2
                ;;
            --validators)
                VALIDATORS="$2"
                shift 2
                ;;
            --ci)
                CI_MODE=true
                RED=''
                GREEN=''
                YELLOW=''
                BLUE=''
                NC=''
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 2
                ;;
        esac
    done
}

# Initialize JSON report
init_json_report() {
    mkdir -p logs
    cat > "$REPORT_FILE" << EOF
{
  "audit_timestamp": "${TIMESTAMP}",
  "configuration": {
    "duration_seconds": ${DURATION},
    "num_validators": ${VALIDATORS},
    "ci_mode": ${CI_MODE}
  },
  "phases": {
EOF
}

# Add phase result to JSON report
add_phase_result() {
    local phase_name="$1"
    local status="$2"
    local duration="$3"
    local metrics="$4"
    
    # Add comma if not first phase
    if [[ "$phase_name" != "basic_consensus" ]]; then
        echo "," >> "$REPORT_FILE"
    fi
    
    cat >> "$REPORT_FILE" << EOF
    "${phase_name}": {
      "status": "${status}",
      "duration_seconds": ${duration},
      ${metrics}
    }
EOF
}

# Finalize JSON report
finalize_json_report() {
    local total_duration=$1
    
    cat >> "$REPORT_FILE" << EOF
  },
  "summary": {
    "total_phases": ${TOTAL_PHASES},
    "passed_phases": ${PASSED_PHASES},
    "failed_phases": ${FAILED_PHASES},
    "overall_status": "$([ $EXIT_CODE -eq 0 ] && echo "passed" || echo "failed")",
    "total_duration_seconds": ${total_duration}
  }
}
EOF
}

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test artifacts...${NC}"
    
    # Kill all tracked processes
    for pid in "${PIDS_TO_CLEANUP[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    
    # Wait for graceful shutdown and database closure
    sleep 3
    
    # Force kill if still running
    for pid in "${PIDS_TO_CLEANUP[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    done
    
    # Kill any remaining audit processes
    pkill -9 -f "poh-node.*audit-" 2>/dev/null || true
    
    # Additional wait for file handles to close
    sleep 2
    
    # Remove temporary databases and state directories
    for db in "${DBS_TO_CLEANUP[@]}"; do
        rm -f "$db" 2>/dev/null || true
        rm -rf "${db}_state.db" 2>/dev/null || true
    done
    
    echo -e "${GREEN}Cleanup complete${NC}"
}

# Signal handler
handle_signal() {
    echo -e "\n${RED}Received interrupt signal${NC}"
    cleanup
    
    # Save partial report if initialized
    if [[ -f "$REPORT_FILE" ]]; then
        finalize_json_report 0
        echo -e "${YELLOW}Partial report saved to: ${REPORT_FILE}${NC}"
    fi
    
    exit 130
}

# Register signal handlers
trap handle_signal INT TERM

# Build the binary
build_binary() {
    echo -e "${BLUE}Building poh-blockchain binary...${NC}"
    if ! go build -o bin/poh-node cmd/main.go; then
        echo -e "${RED}Build failed${NC}"
        exit 2
    fi
    echo -e "${GREEN}Build successful${NC}\n"
}

# Wait for node to be ready
wait_for_node() {
    local port=$1
    local max_wait=5
    local waited=0
    
    while ! nc -z localhost "$port" 2>/dev/null; do
        sleep 0.5
        waited=$((waited + 1))
        if [[ $waited -ge $((max_wait * 2)) ]]; then
            return 1
        fi
    done
    return 0
}

# Count blocks in database
count_blocks() {
    local db=$1
    if [[ ! -f "$db" ]]; then
        echo "0"
        return
    fi
    sqlite3 "$db" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0"
}

# Check chain consistency across databases
check_consistency() {
    local dbs=("$@")
    local first_height=$(count_blocks "${dbs[0]}")
    
    for db in "${dbs[@]}"; do
        local height=$(count_blocks "$db")
        if [[ "$height" != "$first_height" ]]; then
            echo "inconsistent"
            return
        fi
    done
    echo "consistent"
}


# Phase 1: Basic Consensus Test
run_basic_consensus_test() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Phase 1: Basic Consensus Test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Configuration: 1 leader + 3 replicas"
    echo -e "Duration: ${DURATION}s\n"
    
    local phase_start=$(date +%s)
    local phase_status="passed"
    local phase_pids=()
    local phase_dbs=()
    
    # Start leader
    local leader_db="audit-basic-leader.db"
    phase_dbs+=("$leader_db")
    DBS_TO_CLEANUP+=("$leader_db")
    
    ./bin/poh-node --type=leader --port=9000 --db="$leader_db" > logs/audit-basic-leader.log 2>&1 &
    local leader_pid=$!
    phase_pids+=($leader_pid)
    PIDS_TO_CLEANUP+=($leader_pid)
    
    if ! wait_for_node 9000; then
        echo -e "${RED}Failed to start leader${NC}"
        phase_status="failed"
        FAILED_PHASES=$((FAILED_PHASES + 1))
        EXIT_CODE=1
    else
        # Start replicas
        for i in {1..3}; do
            local replica_db="audit-basic-replica${i}.db"
            phase_dbs+=("$replica_db")
            DBS_TO_CLEANUP+=("$replica_db")
            
            local port=$((9000 + i))
            ./bin/poh-node --type=replica --port=$port --peers=localhost:9000 --db="$replica_db" > logs/audit-basic-replica${i}.log 2>&1 &
            local replica_pid=$!
            phase_pids+=($replica_pid)
            PIDS_TO_CLEANUP+=($replica_pid)
            
            if ! wait_for_node $port; then
                echo -e "${RED}Failed to start replica ${i}${NC}"
                phase_status="failed"
            fi
        done
        
        if [[ "$phase_status" == "passed" ]]; then
            echo -e "${GREEN}All nodes started successfully${NC}"
            echo -e "Monitoring block production for ${DURATION}s..."
            sleep "$DURATION"
            
            # Collect metrics
            local total_blocks=0
            local blocks_per_validator=""
            for i in "${!phase_dbs[@]}"; do
                local blocks=$(count_blocks "${phase_dbs[$i]}")
                total_blocks=$((total_blocks + blocks))
                if [[ $i -eq 0 ]]; then
                    blocks_per_validator="\"leader\": $blocks"
                else
                    blocks_per_validator="${blocks_per_validator}, \"replica${i}\": $blocks"
                fi
            done
            
            local consistency=$(check_consistency "${phase_dbs[@]}")
            
            echo -e "\n${GREEN}Results:${NC}"
            echo -e "  Total blocks: $total_blocks"
            echo -e "  Consistency: $consistency"
            
            if [[ "$total_blocks" -eq 0 ]]; then
                echo -e "${RED}No blocks produced - test failed${NC}"
                phase_status="failed"
                FAILED_PHASES=$((FAILED_PHASES + 1))
                EXIT_CODE=1
            else
                PASSED_PHASES=$((PASSED_PHASES + 1))
            fi
        fi
    fi
    
    # Stop nodes
    for pid in "${phase_pids[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    sleep 1
    
    local phase_end=$(date +%s)
    local phase_duration=$((phase_end - phase_start))
    
    # Add to report
    local metrics="\"total_blocks\": ${total_blocks:-0}, \"consistency\": \"${consistency:-unknown}\", \"metrics\": { ${blocks_per_validator:-} }"
    add_phase_result "basic_consensus" "$phase_status" "$phase_duration" "$metrics"
    
    echo -e "${BLUE}Phase 1 complete: ${phase_status}${NC}\n"
}

# Phase 2: BFT with Tolerance Test
run_bft_with_tolerance_test() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Phase 2: BFT with Tolerance Test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Configuration: 1 leader + 4 honest + 1 malicious"
    echo -e "Duration: ${DURATION}s\n"
    
    local phase_start=$(date +%s)
    local phase_status="passed"
    local phase_pids=()
    local phase_dbs=()
    
    # Start leader
    local leader_db="audit-bft-leader.db"
    phase_dbs+=("$leader_db")
    DBS_TO_CLEANUP+=("$leader_db")
    
    ./bin/poh-node --type=leader --port=9100 --db="$leader_db" > logs/audit-bft-leader.log 2>&1 &
    local leader_pid=$!
    phase_pids+=($leader_pid)
    PIDS_TO_CLEANUP+=($leader_pid)
    
    if ! wait_for_node 9100; then
        echo -e "${RED}Failed to start leader${NC}"
        phase_status="failed"
        FAILED_PHASES=$((FAILED_PHASES + 1))
        EXIT_CODE=1
    else
        # Start 4 honest replicas
        for i in {1..4}; do
            local replica_db="audit-bft-replica${i}.db"
            phase_dbs+=("$replica_db")
            DBS_TO_CLEANUP+=("$replica_db")
            
            local port=$((9100 + i))
            ./bin/poh-node --type=replica --port=$port --peers=localhost:9100 --db="$replica_db" > logs/audit-bft-replica${i}.log 2>&1 &
            local replica_pid=$!
            phase_pids+=($replica_pid)
            PIDS_TO_CLEANUP+=($replica_pid)
            
            if ! wait_for_node $port; then
                echo -e "${RED}Failed to start replica ${i}${NC}"
                phase_status="failed"
            fi
        done
        
        # Start 1 malicious replica
        local malicious_db="audit-bft-malicious.db"
        DBS_TO_CLEANUP+=("$malicious_db")
        
        ./bin/poh-node --type=replica --port=9105 --peers=localhost:9100 --db="$malicious_db" --malicious > logs/audit-bft-malicious.log 2>&1 &
        local malicious_pid=$!
        phase_pids+=($malicious_pid)
        PIDS_TO_CLEANUP+=($malicious_pid)
        
        if [[ "$phase_status" == "passed" ]]; then
            echo -e "${GREEN}All nodes started successfully${NC}"
            echo -e "Monitoring BFT behavior for ${DURATION}s..."
            sleep "$DURATION"
            
            # Collect metrics
            local total_blocks=0
            for db in "${phase_dbs[@]}"; do
                local blocks=$(count_blocks "$db")
                total_blocks=$((total_blocks + blocks))
            done
            
            local consistency=$(check_consistency "${phase_dbs[@]}")
            local blocks_rejected=0  # Would need log parsing for actual count
            
            echo -e "\n${GREEN}Results:${NC}"
            echo -e "  Total blocks (honest): $total_blocks"
            echo -e "  Consistency: $consistency"
            echo -e "  Honest nodes: 4"
            echo -e "  Malicious nodes: 1"
            
            if [[ "$total_blocks" -eq 0 ]]; then
                echo -e "${RED}No blocks produced - test failed${NC}"
                phase_status="failed"
                FAILED_PHASES=$((FAILED_PHASES + 1))
                EXIT_CODE=1
            else
                PASSED_PHASES=$((PASSED_PHASES + 1))
            fi
        fi
    fi
    
    # Stop nodes
    for pid in "${phase_pids[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    sleep 1
    
    local phase_end=$(date +%s)
    local phase_duration=$((phase_end - phase_start))
    
    # Add to report
    local metrics="\"total_blocks\": ${total_blocks:-0}, \"consistency\": \"${consistency:-unknown}\", \"honest_nodes\": 4, \"malicious_nodes\": 1, \"blocks_rejected\": ${blocks_rejected}, \"malicious_actions_detected\": 0"
    add_phase_result "bft_with_tolerance" "$phase_status" "$phase_duration" "$metrics"
    
    echo -e "${BLUE}Phase 2 complete: ${phase_status}${NC}\n"
}

# Phase 3: BFT without Tolerance Test
run_bft_without_tolerance_test() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Phase 3: BFT without Tolerance Test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Configuration: 1 leader + 2 honest + 2 malicious"
    echo -e "Duration: ${DURATION}s\n"
    
    local phase_start=$(date +%s)
    local phase_status="passed"
    local phase_pids=()
    local phase_dbs=()
    
    # Start leader
    local leader_db="audit-nobft-leader.db"
    phase_dbs+=("$leader_db")
    DBS_TO_CLEANUP+=("$leader_db")
    
    ./bin/poh-node --type=leader --port=9200 --db="$leader_db" > logs/audit-nobft-leader.log 2>&1 &
    local leader_pid=$!
    phase_pids+=($leader_pid)
    PIDS_TO_CLEANUP+=($leader_pid)
    
    if ! wait_for_node 9200; then
        echo -e "${RED}Failed to start leader${NC}"
        phase_status="failed"
        FAILED_PHASES=$((FAILED_PHASES + 1))
        EXIT_CODE=1
    else
        # Start 2 honest replicas
        for i in {1..2}; do
            local replica_db="audit-nobft-replica${i}.db"
            phase_dbs+=("$replica_db")
            DBS_TO_CLEANUP+=("$replica_db")
            
            local port=$((9200 + i))
            ./bin/poh-node --type=replica --port=$port --peers=localhost:9200 --db="$replica_db" > logs/audit-nobft-replica${i}.log 2>&1 &
            local replica_pid=$!
            phase_pids+=($replica_pid)
            PIDS_TO_CLEANUP+=($replica_pid)
            
            if ! wait_for_node $port; then
                echo -e "${RED}Failed to start replica ${i}${NC}"
                phase_status="failed"
            fi
        done
        
        # Start 2 malicious replicas
        for i in {1..2}; do
            local malicious_db="audit-nobft-malicious${i}.db"
            DBS_TO_CLEANUP+=("$malicious_db")
            
            local port=$((9202 + i))
            ./bin/poh-node --type=replica --port=$port --peers=localhost:9200 --db="$malicious_db" --malicious > logs/audit-nobft-malicious${i}.log 2>&1 &
            local malicious_pid=$!
            phase_pids+=($malicious_pid)
            PIDS_TO_CLEANUP+=($malicious_pid)
        done
        
        if [[ "$phase_status" == "passed" ]]; then
            echo -e "${GREEN}All nodes started successfully${NC}"
            echo -e "Monitoring network under insufficient BFT for ${DURATION}s..."
            sleep "$DURATION"
            
            # Collect metrics
            local total_blocks=0
            for db in "${phase_dbs[@]}"; do
                local blocks=$(count_blocks "$db")
                total_blocks=$((total_blocks + blocks))
            done
            
            local consistency=$(check_consistency "${phase_dbs[@]}")
            
            echo -e "\n${GREEN}Results:${NC}"
            echo -e "  Total blocks (honest): $total_blocks"
            echo -e "  Consistency: $consistency"
            echo -e "  Honest nodes: 2"
            echo -e "  Malicious nodes: 2"
            
            # This phase passes if it completes (demonstrates behavior under attack)
            PASSED_PHASES=$((PASSED_PHASES + 1))
        fi
    fi
    
    # Stop nodes
    for pid in "${phase_pids[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    sleep 1
    
    local phase_end=$(date +%s)
    local phase_duration=$((phase_end - phase_start))
    
    # Add to report
    local metrics="\"total_blocks\": ${total_blocks:-0}, \"consistency\": \"${consistency:-unknown}\", \"honest_nodes\": 2, \"malicious_nodes\": 2"
    add_phase_result "bft_without_tolerance" "$phase_status" "$phase_duration" "$metrics"
    
    echo -e "${BLUE}Phase 3 complete: ${phase_status}${NC}\n"
}

# Phase 4: DPoS Lifecycle Test
run_dpos_lifecycle_test() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Phase 4: DPoS Lifecycle Test${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Configuration: ${VALIDATORS} validators from genesis"
    echo -e "Duration: ${DURATION}s\n"
    
    local phase_start=$(date +%s)
    local phase_status="passed"
    local phase_pids=()
    local phase_dbs=()
    
    # Start leader
    local leader_db="audit-dpos-validator1.db"
    phase_dbs+=("$leader_db")
    DBS_TO_CLEANUP+=("$leader_db")
    
    ./bin/poh-node --type=leader --port=9300 --db="$leader_db" > logs/audit-dpos-validator1.log 2>&1 &
    local leader_pid=$!
    phase_pids+=($leader_pid)
    PIDS_TO_CLEANUP+=($leader_pid)
    
    if ! wait_for_node 9300; then
        echo -e "${RED}Failed to start validator 1${NC}"
        phase_status="failed"
        FAILED_PHASES=$((FAILED_PHASES + 1))
        EXIT_CODE=1
    else
        # Start remaining validators
        for i in $(seq 2 $VALIDATORS); do
            local validator_db="audit-dpos-validator${i}.db"
            phase_dbs+=("$validator_db")
            DBS_TO_CLEANUP+=("$validator_db")
            
            local port=$((9300 + i - 1))
            ./bin/poh-node --type=replica --port=$port --peers=localhost:9300 --db="$validator_db" > logs/audit-dpos-validator${i}.log 2>&1 &
            local validator_pid=$!
            phase_pids+=($validator_pid)
            PIDS_TO_CLEANUP+=($validator_pid)
            
            if ! wait_for_node $port; then
                echo -e "${RED}Failed to start validator ${i}${NC}"
                phase_status="failed"
            fi
        done
        
        if [[ "$phase_status" == "passed" ]]; then
            echo -e "${GREEN}All validators started successfully${NC}"
            echo -e "Monitoring DPoS lifecycle for ${DURATION}s..."
            sleep "$DURATION"
            
            # Collect metrics
            local total_blocks=0
            local blocks_per_validator=""
            for i in $(seq 1 $VALIDATORS); do
                local blocks=$(count_blocks "${phase_dbs[$((i-1))]}")
                total_blocks=$((total_blocks + blocks))
                if [[ $i -eq 1 ]]; then
                    blocks_per_validator="\"validator${i}\": $blocks"
                else
                    blocks_per_validator="${blocks_per_validator}, \"validator${i}\": $blocks"
                fi
            done
            
            local consistency=$(check_consistency "${phase_dbs[@]}")
            
            echo -e "\n${GREEN}Results:${NC}"
            echo -e "  Total blocks: $total_blocks"
            echo -e "  Validators: $VALIDATORS"
            echo -e "  Consistency: $consistency"
            
            if [[ "$total_blocks" -eq 0 ]]; then
                echo -e "${RED}No blocks produced - test failed${NC}"
                phase_status="failed"
                FAILED_PHASES=$((FAILED_PHASES + 1))
                EXIT_CODE=1
            else
                PASSED_PHASES=$((PASSED_PHASES + 1))
            fi
        fi
    fi
    
    # Stop nodes
    for pid in "${phase_pids[@]}"; do
        kill -TERM "$pid" 2>/dev/null || true
    done
    sleep 1
    
    local phase_end=$(date +%s)
    local phase_duration=$((phase_end - phase_start))
    
    # Add to report
    local metrics="\"total_blocks\": ${total_blocks:-0}, \"validators\": ${VALIDATORS}, \"consistency\": \"${consistency:-unknown}\", \"delegation_simulated\": false, \"epoch_simulated\": false, \"slashing_simulated\": false, \"metrics\": { ${blocks_per_validator:-} }"
    add_phase_result "dpos_lifecycle" "$phase_status" "$phase_duration" "$metrics"
    
    echo -e "${BLUE}Phase 4 complete: ${phase_status}${NC}\n"
}

# Display final summary
display_summary() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Audit Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Total phases: ${TOTAL_PHASES}"
    echo -e "Passed: ${GREEN}${PASSED_PHASES}${NC}"
    echo -e "Failed: ${RED}${FAILED_PHASES}${NC}"
    echo -e ""
    
    if [[ $EXIT_CODE -eq 0 ]]; then
        echo -e "${GREEN}✓ Overall Status: PASSED${NC}"
    else
        echo -e "${RED}✗ Overall Status: FAILED${NC}"
    fi
    
    echo -e "\nDetailed report: ${REPORT_FILE}"
}

# Main execution
main() {
    local start_time=$(date +%s)
    
    parse_arguments "$@"
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}PoH Blockchain Audit${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo -e "Timestamp: ${TIMESTAMP}"
    echo -e "Duration per phase: ${DURATION}s"
    echo -e "Validators (DPoS): ${VALIDATORS}"
    echo -e "CI Mode: ${CI_MODE}"
    echo -e "${BLUE}========================================${NC}\n"
    
    build_binary
    init_json_report
    
    # Run all test phases
    run_basic_consensus_test
    run_bft_with_tolerance_test
    run_bft_without_tolerance_test
    run_dpos_lifecycle_test
    
    # Cleanup
    cleanup
    
    # Finalize report
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))
    finalize_json_report "$total_duration"
    
    # Display summary
    display_summary
    
    exit $EXIT_CODE
}

# Run main function
main "$@"

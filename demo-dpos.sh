#!/usr/bin/env bash

# DPoS Blockchain Demo Script
# Comprehensive automated demo exercising the full DPoS lifecycle
# Compatible with analyze-results.sh for machine parsing

set -e

# Configuration
NUM_VALIDATORS=${1:-3}
DURATION=${2:-30}

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Validate input
if ! [[ "$NUM_VALIDATORS" =~ ^[0-9]+$ ]] || [ "$NUM_VALIDATORS" -lt 2 ] || [ "$NUM_VALIDATORS" -gt 9 ]; then
    echo "Error: Number of validators must be between 2 and 9"
    echo "Usage: ./demo-dpos.sh <num_validators> <duration_seconds>"
    echo "Example: ./demo-dpos.sh 3 30  # 3 validators, 30 seconds"
    exit 1
fi

if ! [[ "$DURATION" =~ ^[0-9]+$ ]] || [ "$DURATION" -lt 10 ]; then
    echo "Error: Duration must be at least 10 seconds"
    exit 1
fi

# Generate timestamp for log file
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="logs/dpos-demo-${TIMESTAMP}.json"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  DPoS Blockchain Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Validators: $NUM_VALIDATORS nodes${NC}"
echo -e "${CYAN}  Duration: ${DURATION}s${NC}"
echo -e "${CYAN}  Log file: ${LOG_FILE}${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Clean up old files
echo -e "${CYAN}Cleaning up old files...${NC}"
rm -f validator*.db dpos-state.db
rm -f logs/*.log
mkdir -p logs

# Initialize JSON log
echo "{" > "$LOG_FILE"
echo "  \"demo_type\": \"dpos\"," >> "$LOG_FILE"
echo "  \"timestamp\": \"${TIMESTAMP}\"," >> "$LOG_FILE"
echo "  \"num_validators\": ${NUM_VALIDATORS}," >> "$LOG_FILE"
echo "  \"duration_seconds\": ${DURATION}," >> "$LOG_FILE"
echo "  \"phases\": {" >> "$LOG_FILE"

# Build
echo -e "${CYAN}Building...${NC}"
go build -o bin/poh-node cmd/main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
    echo "  \"build\": {\"status\": \"failed\"}" >> "$LOG_FILE"
    echo "  }" >> "$LOG_FILE"
    echo "}" >> "$LOG_FILE"
    exit 1
fi
echo -e "${GREEN}Build successful${NC}"
echo ""

# Array to store PIDs
declare -a PIDS

# Function to add color prefix to logs
add_prefix() {
    local prefix=$1
    local color=$2
    while IFS= read -r line; do
        echo -e "${color}[${prefix}]${NC} $line"
    done
}

# Function to cleanup on exit
cleanup() {
    echo ""
    echo -e "${CYAN}Stopping all nodes...${NC}"
    
    # Kill all poh-node processes
    pkill -f "poh-node" 2>/dev/null || true
    
    # Also try to kill by PID
    for pid in "${PIDS[@]}"; do
        kill -9 $pid 2>/dev/null || true
    done
    
    # Give processes a moment to die
    sleep 1
    
    echo -e "${GREEN}All nodes stopped${NC}"
}

trap cleanup EXIT INT TERM

# Phase 1: Genesis Start and Block Production
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Phase 1: Genesis & Block Production${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "    \"genesis_and_blocks\": {" >> "$LOG_FILE"
echo "      \"status\": \"running\"," >> "$LOG_FILE"
echo "      \"validators\": [" >> "$LOG_FILE"

# Start validator nodes
for i in $(seq 1 $NUM_VALIDATORS); do
    PORT=$((8000 + i - 1))
    
    # Determine if this is the leader (first node)
    if [ $i -eq 1 ]; then
        NODE_TYPE="leader"
        echo -e "${CYAN}Starting Validator $i (LEADER) on port $PORT...${NC}"
    else
        NODE_TYPE="replica"
        PEERS="localhost:8000"
        echo -e "${CYAN}Starting Validator $i on port $PORT...${NC}"
    fi
    
    # Start the node
    if [ $i -eq 1 ]; then
        ./bin/poh-node --type=$NODE_TYPE --port=$PORT --db=./validator$i.db 2>&1 | \
            add_prefix "VAL-$i" "${GREEN}" | \
            tee logs/validator_$i.log &
    else
        ./bin/poh-node --type=$NODE_TYPE --port=$PORT --peers=$PEERS --db=./validator$i.db 2>&1 | \
            add_prefix "VAL-$i" "${GREEN}" | \
            tee logs/validator_$i.log &
    fi
    
    PIDS+=($!)
    
    # Add to JSON log
    if [ $i -lt $NUM_VALIDATORS ]; then
        echo "        {\"id\": $i, \"port\": $PORT, \"type\": \"$NODE_TYPE\"}," >> "$LOG_FILE"
    else
        echo "        {\"id\": $i, \"port\": $PORT, \"type\": \"$NODE_TYPE\"}" >> "$LOG_FILE"
    fi
    
    sleep 0.5
done

echo "      ]," >> "$LOG_FILE"
echo "      \"blocks\": [" >> "$LOG_FILE"

echo ""
echo -e "${GREEN}All validators started successfully!${NC}"
echo -e "${CYAN}Running block production for ${DURATION} seconds...${NC}"
echo ""

# Monitor block production
BLOCK_COUNT=0
START_TIME=$(date +%s)

# Sample blocks periodically
while [ $(($(date +%s) - START_TIME)) -lt $DURATION ]; do
    sleep 2
    
    # Check if any validator has produced blocks
    for i in $(seq 1 $NUM_VALIDATORS); do
        if [ -f validator$i.db ]; then
            BLOCKS=$(sqlite3 validator$i.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
            if [ "$BLOCKS" -gt "$BLOCK_COUNT" ]; then
                # Get the latest block info
                LATEST_SLOT=$(sqlite3 validator$i.db "SELECT slot FROM blocks ORDER BY block_height DESC LIMIT 1;" 2>/dev/null || echo "0")
                LATEST_HEIGHT=$(sqlite3 validator$i.db "SELECT block_height FROM blocks ORDER BY block_height DESC LIMIT 1;" 2>/dev/null || echo "0")
                
                echo -e "${GREEN}Block produced: height=$LATEST_HEIGHT, slot=$LATEST_SLOT, validator=$i${NC}"
                
                # Log to JSON (simplified - just count)
                BLOCK_COUNT=$BLOCKS
            fi
        fi
    done
done

echo "      ]," >> "$LOG_FILE"
echo "      \"total_blocks\": $BLOCK_COUNT," >> "$LOG_FILE"
echo "      \"status\": \"completed\"" >> "$LOG_FILE"
echo "    }," >> "$LOG_FILE"

echo ""
echo -e "${GREEN}Phase 1 completed: $BLOCK_COUNT blocks produced${NC}"
echo ""

# Phase 2: Delegation
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Phase 2: Stake Delegation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "    \"delegation\": {" >> "$LOG_FILE"
echo "      \"status\": \"running\"," >> "$LOG_FILE"
echo "      \"delegations\": [" >> "$LOG_FILE"

# For now, log that delegation phase would happen here
# Full implementation requires transaction submission via CLI
echo -e "${YELLOW}Note: Delegation requires transaction submission (not yet implemented in demo)${NC}"
echo -e "${CYAN}Simulating delegation phase...${NC}"
echo ""

TOTAL_DELEGATED=0
for i in $(seq 1 $NUM_VALIDATORS); do
    DELEGATION_AMOUNT=$((1000000 + i * 100000))
    TOTAL_DELEGATED=$((TOTAL_DELEGATED + DELEGATION_AMOUNT))
    
    echo -e "${CYAN}Delegation $i:${NC}"
    echo -e "  Validator: $i"
    echo -e "  Amount: $DELEGATION_AMOUNT electrons ($((DELEGATION_AMOUNT / 1000000)) Neon)"
    echo -e "  Status: Would submit DelegateStake instruction"
    echo ""
    
    if [ $i -lt $NUM_VALIDATORS ]; then
        echo "        {\"validator_id\": $i, \"amount\": $DELEGATION_AMOUNT, \"status\": \"simulated\"}," >> "$LOG_FILE"
    else
        echo "        {\"validator_id\": $i, \"amount\": $DELEGATION_AMOUNT, \"status\": \"simulated\"}" >> "$LOG_FILE"
    fi
done

echo "      ]," >> "$LOG_FILE"
echo "      \"total_delegated\": $TOTAL_DELEGATED," >> "$LOG_FILE"
echo "      \"status\": \"simulated\"" >> "$LOG_FILE"
echo "    }," >> "$LOG_FILE"

echo -e "${GREEN}Phase 2 completed (simulated)${NC}"
echo -e "${CYAN}Total delegated: $TOTAL_DELEGATED electrons${NC}"
echo ""

# Phase 3: Epoch Boundary & Reward Distribution
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Phase 3: Epoch & Rewards${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "    \"epoch_and_rewards\": {" >> "$LOG_FILE"
echo "      \"status\": \"running\"," >> "$LOG_FILE"

echo -e "${YELLOW}Note: Epoch boundaries occur every 432,000 slots (~2 days at 400ms/slot)${NC}"
echo -e "${CYAN}Simulating epoch boundary processing...${NC}"
echo ""

# Calculate current slot (approximate based on blocks produced)
CURRENT_SLOT=$((BLOCK_COUNT * 64))  # Approximate: each block covers ~64 slots
EPOCH_LENGTH=432000
CURRENT_EPOCH=$((CURRENT_SLOT / EPOCH_LENGTH))

echo -e "${CYAN}Current epoch: $CURRENT_EPOCH${NC}"
echo -e "${CYAN}Current slot: ~$CURRENT_SLOT${NC}"
echo -e "${CYAN}Next epoch boundary at slot: $((($CURRENT_EPOCH + 1) * EPOCH_LENGTH))${NC}"
echo ""

# Simulate reward distribution
echo -e "${CYAN}Reward Distribution (simulated):${NC}"
TOTAL_REWARD_POOL=1000000  # 1 Neon in electrons

echo "      \"epoch_number\": $CURRENT_EPOCH," >> "$LOG_FILE"
echo "      \"current_slot\": $CURRENT_SLOT," >> "$LOG_FILE"
echo "      \"reward_pool\": $TOTAL_REWARD_POOL," >> "$LOG_FILE"
echo "      \"validator_rewards\": [" >> "$LOG_FILE"

for i in $(seq 1 $NUM_VALIDATORS); do
    # Calculate proportional reward based on blocks produced
    if [ -f validator$i.db ]; then
        VAL_BLOCKS=$(sqlite3 validator$i.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
    else
        VAL_BLOCKS=0
    fi
    
    if [ $BLOCK_COUNT -gt 0 ]; then
        VAL_REWARD=$(( (VAL_BLOCKS * TOTAL_REWARD_POOL) / BLOCK_COUNT ))
    else
        VAL_REWARD=0
    fi
    
    # Simulate commission (10%)
    COMMISSION=$(( VAL_REWARD * 10 / 100 ))
    DELEGATOR_REWARD=$(( VAL_REWARD - COMMISSION ))
    
    echo -e "  Validator $i:"
    echo -e "    Blocks produced: $VAL_BLOCKS"
    echo -e "    Total reward: $VAL_REWARD electrons"
    echo -e "    Commission (10%): $COMMISSION electrons"
    echo -e "    Delegator share: $DELEGATOR_REWARD electrons"
    echo ""
    
    if [ $i -lt $NUM_VALIDATORS ]; then
        echo "        {\"validator_id\": $i, \"blocks\": $VAL_BLOCKS, \"reward\": $VAL_REWARD, \"commission\": $COMMISSION, \"delegator_share\": $DELEGATOR_REWARD}," >> "$LOG_FILE"
    else
        echo "        {\"validator_id\": $i, \"blocks\": $VAL_BLOCKS, \"reward\": $VAL_REWARD, \"commission\": $COMMISSION, \"delegator_share\": $DELEGATOR_REWARD}" >> "$LOG_FILE"
    fi
done

echo "      ]," >> "$LOG_FILE"
echo "      \"status\": \"simulated\"" >> "$LOG_FILE"
echo "    }," >> "$LOG_FILE"

echo -e "${GREEN}Phase 3 completed (simulated)${NC}"
echo ""

# Phase 4: Slashing
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Phase 4: Slashing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "    \"slashing\": {" >> "$LOG_FILE"
echo "      \"status\": \"running\"," >> "$LOG_FILE"

echo -e "${YELLOW}Note: Slashing requires double-sign proof submission (not yet implemented in demo)${NC}"
echo -e "${CYAN}Simulating slashing event...${NC}"
echo ""

TARGET_VALIDATOR=1
VALIDATOR_STAKE=1000000  # 1 Neon
SLASHING_PERCENTAGE=5
SLASHED_AMOUNT=$(( VALIDATOR_STAKE * SLASHING_PERCENTAGE / 100 ))
NEW_STAKE=$(( VALIDATOR_STAKE - SLASHED_AMOUNT ))
MIN_ACTIVATION_STAKE=1000000

echo -e "${CYAN}Slashing Event:${NC}"
echo -e "  Target validator: $TARGET_VALIDATOR"
echo -e "  Original stake: $VALIDATOR_STAKE electrons ($((VALIDATOR_STAKE / 1000000)) Neon)"
echo -e "  Slashing percentage: ${SLASHING_PERCENTAGE}%"
echo -e "  Slashed amount: $SLASHED_AMOUNT electrons"
echo -e "  New stake: $NEW_STAKE electrons"
echo ""

# Check if validator should be deactivated
if [ $NEW_STAKE -lt $MIN_ACTIVATION_STAKE ]; then
    DEACTIVATED=true
    echo -e "${RED}  Status: Validator DEACTIVATED (stake below minimum threshold)${NC}"
else
    DEACTIVATED=false
    echo -e "${YELLOW}  Status: Validator remains ACTIVE${NC}"
fi

echo ""
echo -e "${CYAN}Slashed amount transferred to Reward Pool${NC}"
echo ""

echo "      \"target_validator\": $TARGET_VALIDATOR," >> "$LOG_FILE"
echo "      \"original_stake\": $VALIDATOR_STAKE," >> "$LOG_FILE"
echo "      \"slashed_amount\": $SLASHED_AMOUNT," >> "$LOG_FILE"
echo "      \"new_stake\": $NEW_STAKE," >> "$LOG_FILE"
echo "      \"deactivated\": $DEACTIVATED," >> "$LOG_FILE"
echo "      \"status\": \"simulated\"" >> "$LOG_FILE"
echo "    }" >> "$LOG_FILE"

echo -e "${GREEN}Phase 4 completed (simulated)${NC}"
echo ""

# Close JSON log
echo "  }," >> "$LOG_FILE"

# Final Analysis
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Results Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "  \"summary\": {" >> "$LOG_FILE"

# Collect final statistics
TOTAL_BLOCKS=0
declare -a VALIDATOR_BLOCKS
declare -a VALIDATOR_HEIGHTS

for i in $(seq 1 $NUM_VALIDATORS); do
    if [ -f validator$i.db ]; then
        BLOCKS=$(sqlite3 validator$i.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 validator$i.db "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        VALIDATOR_BLOCKS[$i]=$BLOCKS
        VALIDATOR_HEIGHTS[$i]=$HEIGHT
        TOTAL_BLOCKS=$((TOTAL_BLOCKS + BLOCKS))
    else
        VALIDATOR_BLOCKS[$i]=0
        VALIDATOR_HEIGHTS[$i]=0
    fi
done

# Print human-readable summary table
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                    VALIDATOR SUMMARY                       ║${NC}"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
printf "${CYAN}║${NC} %-12s %-15s %-15s %-12s ${CYAN}║${NC}\n" "Validator" "Blocks" "Height" "Status"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"

for i in $(seq 1 $NUM_VALIDATORS); do
    BLOCKS=${VALIDATOR_BLOCKS[$i]}
    HEIGHT=${VALIDATOR_HEIGHTS[$i]}
    
    if [ $BLOCKS -gt 0 ]; then
        STATUS="${GREEN}ACTIVE${NC}"
    else
        STATUS="${YELLOW}INACTIVE${NC}"
    fi
    
    printf "${CYAN}║${NC} %-12s %-15s %-15s %-12s ${CYAN}║${NC}\n" "Validator $i" "$BLOCKS" "$HEIGHT" "$(echo -e $STATUS)"
done

echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check consistency
FIRST_HEIGHT=${VALIDATOR_HEIGHTS[1]}
ALL_CONSISTENT=true

for i in $(seq 2 $NUM_VALIDATORS); do
    HEIGHT=${VALIDATOR_HEIGHTS[$i]}
    if [ "$HEIGHT" != "$FIRST_HEIGHT" ]; then
        ALL_CONSISTENT=false
        break
    fi
done

# Print lifecycle phase results
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                    LIFECYCLE PHASES                        ║${NC}"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
printf "${CYAN}║${NC} %-30s %-25s ${CYAN}║${NC}\n" "Phase" "Status"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
printf "${CYAN}║${NC} %-30s ${GREEN}%-25s${NC} ${CYAN}║${NC}\n" "Genesis & Block Production" "COMPLETED"
printf "${CYAN}║${NC} %-30s ${YELLOW}%-25s${NC} ${CYAN}║${NC}\n" "Stake Delegation" "SIMULATED"
printf "${CYAN}║${NC} %-30s ${YELLOW}%-25s${NC} ${CYAN}║${NC}\n" "Epoch & Reward Distribution" "SIMULATED"
printf "${CYAN}║${NC} %-30s ${YELLOW}%-25s${NC} ${CYAN}║${NC}\n" "Slashing" "SIMULATED"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Print metrics
echo -e "${CYAN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                         METRICS                            ║${NC}"
echo -e "${CYAN}╠════════════════════════════════════════════════════════════╣${NC}"
printf "${CYAN}║${NC} %-35s %-20s ${CYAN}║${NC}\n" "Total Blocks Produced:" "$TOTAL_BLOCKS"
printf "${CYAN}║${NC} %-35s %-20s ${CYAN}║${NC}\n" "Average Blocks/Validator:" "$((TOTAL_BLOCKS / NUM_VALIDATORS))"
printf "${CYAN}║${NC} %-35s %-20s ${CYAN}║${NC}\n" "Total Validators:" "$NUM_VALIDATORS"
printf "${CYAN}║${NC} %-35s %-20s ${CYAN}║${NC}\n" "Duration:" "${DURATION}s"

if [ "$ALL_CONSISTENT" = true ]; then
    printf "${CYAN}║${NC} %-35s ${GREEN}%-20s${NC} ${CYAN}║${NC}\n" "Chain Consistency:" "CONSISTENT"
    CONSISTENCY="consistent"
else
    printf "${CYAN}║${NC} %-35s ${RED}%-20s${NC} ${CYAN}║${NC}\n" "Chain Consistency:" "INCONSISTENT"
    CONSISTENCY="inconsistent"
fi

echo -e "${CYAN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Determine overall status
PHASE_FAILURES=0

# Check if block production succeeded
if [ $TOTAL_BLOCKS -eq 0 ]; then
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
fi

# Check consistency
if [ "$ALL_CONSISTENT" = false ]; then
    PHASE_FAILURES=$((PHASE_FAILURES + 1))
fi

if [ $PHASE_FAILURES -eq 0 ]; then
    echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                    ✓ DEMO SUCCESSFUL                       ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
    OVERALL_STATUS="success"
    EXIT_CODE=0
else
    echo -e "${RED}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                    ✗ DEMO FAILED                           ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════════════════════════╝${NC}"
    OVERALL_STATUS="failed"
    EXIT_CODE=1
fi

echo ""

echo "    \"total_blocks\": $TOTAL_BLOCKS," >> "$LOG_FILE"
echo "    \"average_blocks_per_validator\": $((TOTAL_BLOCKS / NUM_VALIDATORS))," >> "$LOG_FILE"
echo "    \"consistency\": \"$CONSISTENCY\"," >> "$LOG_FILE"
echo "    \"phase_failures\": $PHASE_FAILURES," >> "$LOG_FILE"
echo "    \"overall_status\": \"$OVERALL_STATUS\"" >> "$LOG_FILE"
echo "  }" >> "$LOG_FILE"
echo "}" >> "$LOG_FILE"

echo -e "${CYAN}Logs saved to: ${LOG_FILE}${NC}"
echo -e "${CYAN}Database files: validator*.db${NC}"
echo -e "${CYAN}Node logs: logs/validator_*.log${NC}"
echo ""

# Cleanup will be called by trap
exit $EXIT_CODE

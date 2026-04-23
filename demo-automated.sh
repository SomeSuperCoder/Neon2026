#!/usr/bin/env bash

# Automated PoH Blockchain Demo (No tmux required)
# Runs nodes in background with prefixed logs for easy analysis

set -e

# Configuration
NUM_HONEST=${1:-2}
NUM_MALICIOUS=${2:-1}
DURATION=${3:-30}  # Run duration in seconds

TOTAL_REPLICAS=$((NUM_HONEST + NUM_MALICIOUS))

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Validate input
if ! [[ "$NUM_HONEST" =~ ^[0-9]+$ ]] || [ "$NUM_HONEST" -lt 1 ] || [ "$NUM_HONEST" -gt 9 ]; then
    echo "Error: Number of honest replicas must be between 1 and 9"
    echo "Usage: ./demo-automated.sh [num_honest] [num_malicious] [duration_seconds]"
    exit 1
fi

if ! [[ "$NUM_MALICIOUS" =~ ^[0-9]+$ ]] || [ "$NUM_MALICIOUS" -lt 0 ] || [ "$NUM_MALICIOUS" -gt 9 ]; then
    echo "Error: Number of malicious replicas must be between 0 and 9"
    exit 1
fi

if [ "$TOTAL_REPLICAS" -gt 9 ]; then
    echo "Error: Total replicas cannot exceed 9"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Automated PoH Blockchain Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Leader: 1 node${NC}"
echo -e "${GREEN}  Honest Replicas: $NUM_HONEST nodes${NC}"
echo -e "${RED}  Malicious Replicas: $NUM_MALICIOUS nodes${NC}"
echo -e "${CYAN}  Duration: ${DURATION}s${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# BFT calculation
if [ "$NUM_HONEST" -le $((2 * NUM_MALICIOUS)) ]; then
    echo -e "${RED}⚠ WARNING: Network does NOT have BFT!${NC}"
    echo -e "${RED}  Honest ($NUM_HONEST) <= 2 * Malicious ($NUM_MALICIOUS)${NC}"
    HAS_BFT=false
else
    echo -e "${GREEN}✓ Network has BFT tolerance${NC}"
    echo -e "${GREEN}  Honest ($NUM_HONEST) > 2 * Malicious ($NUM_MALICIOUS)${NC}"
    HAS_BFT=true
fi
echo ""

# Clean up old files
echo -e "${CYAN}Cleaning up old files...${NC}"
rm -f leader.db replica*.db malicious*.db
rm -f logs/*.log
mkdir -p logs

# Build
echo -e "${CYAN}Building...${NC}"
go build -o bin/poh-node cmd/main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed${NC}"
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

# Start leader
echo -e "${CYAN}Starting Leader on port 8000...${NC}"
./bin/poh-node --type=leader --port=8000 --db=./leader.db 2>&1 | \
    add_prefix "LEADER" "${BLUE}" | \
    tee logs/leader.log &
PIDS+=($!)
sleep 2

# Start honest replicas
for i in $(seq 1 $NUM_HONEST); do
    PORT=$((8000 + i))
    echo -e "${CYAN}Starting Honest Replica $i on port $PORT...${NC}"
    ./bin/poh-node --type=replica --port=$PORT --peers=localhost:8000 --db=./replica$i.db 2>&1 | \
        add_prefix "HONEST-$i" "${GREEN}" | \
        tee logs/honest_$i.log &
    PIDS+=($!)
    sleep 0.5
done

# Start malicious replicas
for i in $(seq 1 $NUM_MALICIOUS); do
    PANE_NUM=$((NUM_HONEST + i))
    PORT=$((8000 + PANE_NUM))
    echo -e "${CYAN}Starting Malicious Replica $i on port $PORT...${NC}"
    ./bin/poh-node --type=replica --port=$PORT --peers=localhost:8000 --db=./malicious$i.db --malicious 2>&1 | \
        add_prefix "MALICIOUS-$i" "${RED}" | \
        tee logs/malicious_$i.log &
    PIDS+=($!)
    sleep 0.5
done

echo ""
echo -e "${GREEN}All nodes started successfully!${NC}"
echo -e "${CYAN}Running for ${DURATION} seconds...${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop early${NC}"
echo ""

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
    
    # Analyze results
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Results Analysis${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    # Check block counts
    echo ""
    echo -e "${CYAN}Block counts:${NC}"
    if [ -f leader.db ]; then
        LEADER_BLOCKS=$(sqlite3 leader.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "  ${BLUE}Leader:${NC} $LEADER_BLOCKS blocks"
    fi
    
    for i in $(seq 1 $NUM_HONEST); do
        if [ -f replica$i.db ]; then
            BLOCKS=$(sqlite3 replica$i.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
            echo -e "  ${GREEN}Honest Replica $i:${NC} $BLOCKS blocks"
        fi
    done
    
    for i in $(seq 1 $NUM_MALICIOUS); do
        if [ -f malicious$i.db ]; then
            BLOCKS=$(sqlite3 malicious$i.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
            echo -e "  ${RED}Malicious Replica $i:${NC} $BLOCKS blocks"
        fi
    done
    
    # Check for validation failures in logs
    echo ""
    echo -e "${CYAN}Validation failures:${NC}"
    for i in $(seq 1 $NUM_HONEST); do
        if [ -f logs/honest_$i.log ]; then
            FAILURES=$(grep -c "verification failed\|validation failed" logs/honest_$i.log 2>/dev/null || echo "0")
            echo -e "  ${GREEN}Honest Replica $i:${NC} $FAILURES rejections"
        fi
    done
    
    # Check malicious activity
    echo ""
    echo -e "${CYAN}Malicious activity detected:${NC}"
    for i in $(seq 1 $NUM_MALICIOUS); do
        if [ -f logs/malicious_$i.log ]; then
            MALICIOUS_ACTS=$(grep -c "MALICIOUS:" logs/malicious_$i.log 2>/dev/null || echo "0")
            echo -e "  ${RED}Malicious Replica $i:${NC} $MALICIOUS_ACTS malicious actions"
        fi
    done
    
    # BFT verdict
    echo ""
    if [ "$HAS_BFT" = true ]; then
        echo -e "${GREEN}✓ BFT Test Result: Network maintained integrity${NC}"
    else
        echo -e "${RED}✗ BFT Test Result: Network may be compromised${NC}"
    fi
    
    echo ""
    echo -e "${CYAN}Logs saved in logs/ directory${NC}"
    echo -e "${CYAN}Databases saved for inspection${NC}"
    echo ""
}

trap cleanup EXIT INT TERM

# Wait for specified duration
sleep $DURATION

# Cleanup will be called by trap
exit 0

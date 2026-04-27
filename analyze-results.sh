#!/usr/bin/env bash

# Quick Results Analyzer
# Analyzes blockchain state and logs without running nodes

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Default configuration
DB_DIR=""
LOG_DIR="logs"

# Parse command-line arguments
show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Analyze blockchain databases and logs from audit or devnet runs."
    echo ""
    echo "Options:"
    echo "  --db-dir DIR    Directory containing database files (default: current directory)"
    echo "  --log-dir DIR   Directory containing log files (default: logs)"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                          # Analyze databases in current directory"
    echo "  $0 --db-dir devnet-data     # Analyze devnet databases"
    echo "  $0 --db-dir audit-data      # Analyze audit databases"
    exit 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --db-dir)
            DB_DIR="$2"
            shift 2
            ;;
        --log-dir)
            LOG_DIR="$2"
            shift 2
            ;;
        --help)
            show_help
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# If no db-dir specified, check both current directory and devnet-data
if [ -z "$DB_DIR" ]; then
    if ls *.db 1> /dev/null 2>&1; then
        DB_DIR="."
    elif [ -d "devnet-data" ] && ls devnet-data/*.db 1> /dev/null 2>&1; then
        DB_DIR="devnet-data"
        echo -e "${CYAN}Found databases in devnet-data/${NC}"
        echo ""
    else
        echo -e "${RED}No database files found.${NC}"
        echo "Run audit.sh or devnet.sh first, or specify --db-dir"
        exit 1
    fi
fi

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  PoH Blockchain Results Analyzer      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${CYAN}Database directory: $DB_DIR${NC}"
echo -e "${CYAN}Log directory: $LOG_DIR${NC}"
echo ""

# Check if databases exist in specified directory
if ! ls "$DB_DIR"/*.db 1> /dev/null 2>&1; then
    echo -e "${RED}No database files found in $DB_DIR${NC}"
    exit 1
fi

# Analyze databases
echo -e "${CYAN}=== Database Analysis ===${NC}"
echo ""

# Leader
if [ -f "$DB_DIR/leader.db" ]; then
    LEADER_BLOCKS=$(sqlite3 "$DB_DIR/leader.db" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
    LEADER_HEIGHT=$(sqlite3 "$DB_DIR/leader.db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
    echo -e "${BLUE}[LEADER]${NC}"
    echo "  Blocks: $LEADER_BLOCKS"
    echo "  Height: $LEADER_HEIGHT"
    
    # Show last 5 blocks
    echo "  Last 5 blocks:"
    sqlite3 "$DB_DIR/leader.db" "SELECT '    ' || block_height || ': slot=' || slot || ', entries=' || (SELECT COUNT(*) FROM json_each(json_extract(data, '$.entries'))) FROM blocks ORDER BY block_height DESC LIMIT 5;" 2>/dev/null || echo "    (unable to read)"
    echo ""
fi

# Check for validator databases (devnet naming: validator1.db, validator2.db, etc.)
for db in "$DB_DIR"/validator*.db; do
    if [ -f "$db" ]; then
        NUM=$(basename "$db" | grep -o '[0-9]\+')
        BLOCKS=$(sqlite3 "$db" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 "$db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "${GREEN}[VALIDATOR-$NUM]${NC}"
        echo "  Blocks: $BLOCKS"
        echo "  Height: $HEIGHT"
        
        # Check for gaps
        GAPS=$(sqlite3 "$db" "SELECT COUNT(*) FROM (SELECT block_height, LAG(block_height) OVER (ORDER BY block_height) as prev FROM blocks) WHERE block_height != prev + 1 AND prev IS NOT NULL;" 2>/dev/null || echo "0")
        if [ "$GAPS" -gt 0 ]; then
            echo -e "  ${RED}⚠ Chain gaps detected: $GAPS${NC}"
        else
            echo -e "  ${GREEN}✓ Chain continuous${NC}"
        fi
        echo ""
    fi
done

# Honest replicas (audit naming: replica1.db, replica2.db, etc.)
for db in "$DB_DIR"/replica*.db; do
    if [ -f "$db" ]; then
        NUM=$(basename "$db" | grep -o '[0-9]\+')
        BLOCKS=$(sqlite3 "$db" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 "$db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "${GREEN}[HONEST-$NUM]${NC}"
        echo "  Blocks: $BLOCKS"
        echo "  Height: $HEIGHT"
        
        # Check for gaps
        GAPS=$(sqlite3 "$db" "SELECT COUNT(*) FROM (SELECT block_height, LAG(block_height) OVER (ORDER BY block_height) as prev FROM blocks) WHERE block_height != prev + 1 AND prev IS NOT NULL;" 2>/dev/null || echo "0")
        if [ "$GAPS" -gt 0 ]; then
            echo -e "  ${RED}⚠ Chain gaps detected: $GAPS${NC}"
        else
            echo -e "  ${GREEN}✓ Chain continuous${NC}"
        fi
        echo ""
    fi
done

# Malicious replicas
for db in "$DB_DIR"/malicious*.db; do
    if [ -f "$db" ]; then
        NUM=$(basename "$db" | grep -o '[0-9]\+')
        BLOCKS=$(sqlite3 "$db" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 "$db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "${RED}[MALICIOUS-$NUM]${NC}"
        echo "  Blocks: $BLOCKS"
        echo "  Height: $HEIGHT"
        
        # Check for gaps
        GAPS=$(sqlite3 "$db" "SELECT COUNT(*) FROM (SELECT block_height, LAG(block_height) OVER (ORDER BY block_height) as prev FROM blocks) WHERE block_height != prev + 1 AND prev IS NOT NULL;" 2>/dev/null || echo "0")
        if [ "$GAPS" -gt 0 ]; then
            echo -e "  ${YELLOW}⚠ Chain gaps: $GAPS (expected for malicious)${NC}"
        fi
        echo ""
    fi
done

# Consistency check
echo -e "${CYAN}=== Consistency Check ===${NC}"
echo ""

HONEST_HEIGHTS=()

# Check validator databases (devnet)
for db in "$DB_DIR"/validator*.db; do
    if [ -f "$db" ]; then
        HEIGHT=$(sqlite3 "$db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        HONEST_HEIGHTS+=($HEIGHT)
    fi
done

# Check replica databases (audit)
for db in "$DB_DIR"/replica*.db; do
    if [ -f "$db" ]; then
        HEIGHT=$(sqlite3 "$db" "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        HONEST_HEIGHTS+=($HEIGHT)
    fi
done

if [ ${#HONEST_HEIGHTS[@]} -gt 0 ]; then
    FIRST_HEIGHT=${HONEST_HEIGHTS[0]}
    ALL_SAME=true
    for height in "${HONEST_HEIGHTS[@]}"; do
        if [ "$height" != "$FIRST_HEIGHT" ]; then
            ALL_SAME=false
            break
        fi
    done
    
    if [ "$ALL_SAME" = true ]; then
        echo -e "${GREEN}✓ All honest nodes have consistent height: $FIRST_HEIGHT${NC}"
    else
        echo -e "${RED}✗ Honest nodes have inconsistent heights:${NC}"
        for i in "${!HONEST_HEIGHTS[@]}"; do
            echo "  Node $((i+1)): ${HONEST_HEIGHTS[$i]}"
        done
    fi
else
    echo -e "${YELLOW}No honest node databases found${NC}"
fi
echo ""

# Log analysis
if [ -d "$LOG_DIR" ]; then
    echo -e "${CYAN}=== Log Analysis ===${NC}"
    echo ""
    
    # Count validation failures for honest nodes
    for log in "$LOG_DIR"/honest_*.log "$LOG_DIR"/devnet-validator-*.log; do
        if [ -f "$log" ]; then
            BASENAME=$(basename "$log")
            NUM=$(echo "$BASENAME" | grep -o '[0-9]\+')
            FAILURES=$(grep -c "verification failed\|validation failed\|linkage verification failed" "$log" 2>/dev/null || echo "0")
            SUCCESSES=$(grep -c "Block stored successfully" "$log" 2>/dev/null || echo "0")
            
            if [[ "$BASENAME" == devnet-validator-* ]]; then
                echo -e "${GREEN}[VALIDATOR-$NUM]${NC}"
            else
                echo -e "${GREEN}[HONEST-$NUM]${NC}"
            fi
            
            echo "  Blocks stored: $SUCCESSES"
            echo "  Blocks rejected: $FAILURES"
            
            if [ "$FAILURES" -gt 0 ]; then
                echo "  Sample rejections:"
                grep "verification failed\|validation failed" "$log" 2>/dev/null | head -3 | sed 's/^/    /'
            fi
            echo ""
        fi
    done
    
    # Count malicious actions
    for log in "$LOG_DIR"/malicious_*.log; do
        if [ -f "$log" ]; then
            NUM=$(basename "$log" | grep -o '[0-9]\+')
            ACTIONS=$(grep -c "MALICIOUS:" "$log" 2>/dev/null || echo "0")
            STORED=$(grep -c "Block stored" "$log" 2>/dev/null || echo "0")
            echo -e "${RED}[MALICIOUS-$NUM]${NC}"
            echo "  Blocks stored: $STORED"
            echo "  Malicious actions: $ACTIONS"
            
            if [ "$ACTIONS" -gt 0 ]; then
                echo "  Sample actions:"
                grep "MALICIOUS:" "$log" 2>/dev/null | head -3 | sed 's/^/    /'
            fi
            echo ""
        fi
    done
else
    echo -e "${YELLOW}No logs directory found at $LOG_DIR${NC}"
    echo ""
fi

# BFT verdict
echo -e "${CYAN}=== BFT Verdict ===${NC}"
echo ""

# Count honest nodes (both validator and replica databases)
HONEST_COUNT=$(ls "$DB_DIR"/replica*.db "$DB_DIR"/validator*.db 2>/dev/null | wc -l)
MALICIOUS_COUNT=$(ls "$DB_DIR"/malicious*.db 2>/dev/null | wc -l)

echo "Network composition:"
echo "  Honest nodes: $HONEST_COUNT"
echo "  Malicious nodes: $MALICIOUS_COUNT"
echo ""

if [ $MALICIOUS_COUNT -eq 0 ]; then
    echo -e "${CYAN}ℹ No malicious nodes detected (devnet or basic consensus test)${NC}"
    if [ "$ALL_SAME" = true ]; then
        echo -e "${GREEN}✓ All nodes maintained consistency${NC}"
        echo -e "${GREEN}✓ VERDICT: NETWORK HEALTHY${NC}"
    else
        echo -e "${YELLOW}⚠ Nodes have inconsistent heights${NC}"
        echo -e "${YELLOW}⚠ VERDICT: SYNCHRONIZATION ISSUE${NC}"
    fi
elif [ $HONEST_COUNT -gt $((2 * MALICIOUS_COUNT)) ]; then
    echo -e "${GREEN}✓ Network has BFT ($HONEST_COUNT > 2×$MALICIOUS_COUNT)${NC}"
    if [ "$ALL_SAME" = true ]; then
        echo -e "${GREEN}✓ Honest nodes maintained consistency${NC}"
        echo -e "${GREEN}✓ VERDICT: BFT SUCCESSFUL${NC}"
    else
        echo -e "${RED}✗ Honest nodes inconsistent despite BFT${NC}"
        echo -e "${RED}✗ VERDICT: BFT FAILED (unexpected)${NC}"
    fi
else
    echo -e "${RED}✗ Network lacks BFT ($HONEST_COUNT ≤ 2×$MALICIOUS_COUNT)${NC}"
    if [ "$ALL_SAME" = true ]; then
        echo -e "${YELLOW}⚠ Honest nodes consistent without BFT (lucky)${NC}"
        echo -e "${YELLOW}⚠ VERDICT: NO BFT but survived${NC}"
    else
        echo -e "${YELLOW}⚠ Honest nodes inconsistent as expected${NC}"
        echo -e "${YELLOW}⚠ VERDICT: NO BFT - network compromised${NC}"
    fi
fi

echo ""

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

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  PoH Blockchain Results Analyzer      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Check if databases exist
if ! ls *.db 1> /dev/null 2>&1; then
    echo -e "${RED}No database files found. Run a demo first.${NC}"
    exit 1
fi

# Analyze databases
echo -e "${CYAN}=== Database Analysis ===${NC}"
echo ""

# Leader
if [ -f leader.db ]; then
    LEADER_BLOCKS=$(sqlite3 leader.db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
    LEADER_HEIGHT=$(sqlite3 leader.db "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
    echo -e "${BLUE}[LEADER]${NC}"
    echo "  Blocks: $LEADER_BLOCKS"
    echo "  Height: $LEADER_HEIGHT"
    
    # Show last 5 blocks
    echo "  Last 5 blocks:"
    sqlite3 leader.db "SELECT '    ' || block_height || ': slot=' || slot || ', entries=' || (SELECT COUNT(*) FROM json_each(json_extract(data, '$.entries'))) FROM blocks ORDER BY block_height DESC LIMIT 5;" 2>/dev/null || echo "    (unable to read)"
    echo ""
fi

# Honest replicas
for db in replica*.db; do
    if [ -f "$db" ]; then
        NUM=$(echo $db | grep -o '[0-9]\+')
        BLOCKS=$(sqlite3 $db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 $db "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "${GREEN}[HONEST-$NUM]${NC}"
        echo "  Blocks: $BLOCKS"
        echo "  Height: $HEIGHT"
        
        # Check for gaps
        GAPS=$(sqlite3 $db "SELECT COUNT(*) FROM (SELECT block_height, LAG(block_height) OVER (ORDER BY block_height) as prev FROM blocks) WHERE block_height != prev + 1 AND prev IS NOT NULL;" 2>/dev/null || echo "0")
        if [ "$GAPS" -gt 0 ]; then
            echo -e "  ${RED}⚠ Chain gaps detected: $GAPS${NC}"
        else
            echo -e "  ${GREEN}✓ Chain continuous${NC}"
        fi
        echo ""
    fi
done

# Malicious replicas
for db in malicious*.db; do
    if [ -f "$db" ]; then
        NUM=$(echo $db | grep -o '[0-9]\+')
        BLOCKS=$(sqlite3 $db "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "0")
        HEIGHT=$(sqlite3 $db "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
        echo -e "${RED}[MALICIOUS-$NUM]${NC}"
        echo "  Blocks: $BLOCKS"
        echo "  Height: $HEIGHT"
        
        # Check for gaps
        GAPS=$(sqlite3 $db "SELECT COUNT(*) FROM (SELECT block_height, LAG(block_height) OVER (ORDER BY block_height) as prev FROM blocks) WHERE block_height != prev + 1 AND prev IS NOT NULL;" 2>/dev/null || echo "0")
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
for db in replica*.db; do
    if [ -f "$db" ]; then
        HEIGHT=$(sqlite3 $db "SELECT MAX(block_height) FROM blocks;" 2>/dev/null || echo "0")
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
            echo "  Replica $((i+1)): ${HONEST_HEIGHTS[$i]}"
        done
    fi
else
    echo -e "${YELLOW}No honest replica databases found${NC}"
fi
echo ""

# Log analysis
if [ -d logs ]; then
    echo -e "${CYAN}=== Log Analysis ===${NC}"
    echo ""
    
    # Count validation failures
    for log in logs/honest_*.log; do
        if [ -f "$log" ]; then
            NUM=$(echo $log | grep -o '[0-9]\+')
            FAILURES=$(grep -c "verification failed\|validation failed\|linkage verification failed" $log 2>/dev/null || echo "0")
            SUCCESSES=$(grep -c "Block stored successfully" $log 2>/dev/null || echo "0")
            echo -e "${GREEN}[HONEST-$NUM]${NC}"
            echo "  Blocks stored: $SUCCESSES"
            echo "  Blocks rejected: $FAILURES"
            
            if [ "$FAILURES" -gt 0 ]; then
                echo "  Sample rejections:"
                grep "verification failed\|validation failed" $log 2>/dev/null | head -3 | sed 's/^/    /'
            fi
            echo ""
        fi
    done
    
    # Count malicious actions
    for log in logs/malicious_*.log; do
        if [ -f "$log" ]; then
            NUM=$(echo $log | grep -o '[0-9]\+')
            ACTIONS=$(grep -c "MALICIOUS:" $log 2>/dev/null || echo "0")
            STORED=$(grep -c "Block stored" $log 2>/dev/null || echo "0")
            echo -e "${RED}[MALICIOUS-$NUM]${NC}"
            echo "  Blocks stored: $STORED"
            echo "  Malicious actions: $ACTIONS"
            
            if [ "$ACTIONS" -gt 0 ]; then
                echo "  Sample actions:"
                grep "MALICIOUS:" $log 2>/dev/null | head -3 | sed 's/^/    /'
            fi
            echo ""
        fi
    done
else
    echo -e "${YELLOW}No logs directory found${NC}"
    echo ""
fi

# BFT verdict
echo -e "${CYAN}=== BFT Verdict ===${NC}"
echo ""

HONEST_COUNT=$(ls replica*.db 2>/dev/null | wc -l)
MALICIOUS_COUNT=$(ls malicious*.db 2>/dev/null | wc -l)

echo "Network composition:"
echo "  Honest nodes: $HONEST_COUNT"
echo "  Malicious nodes: $MALICIOUS_COUNT"
echo ""

if [ $HONEST_COUNT -gt $((2 * MALICIOUS_COUNT)) ]; then
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

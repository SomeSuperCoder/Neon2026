#!env bash

# PoH Blockchain BFT Demo Script
# This script starts a leader and multiple replica nodes (honest and malicious) for BFT testing

set -e

# Default values
NUM_HONEST=${1:-2}
NUM_MALICIOUS=${2:-1}

# Validate input
if ! [[ "$NUM_HONEST" =~ ^[0-9]+$ ]] || [ "$NUM_HONEST" -lt 1 ] || [ "$NUM_HONEST" -gt 9 ]; then
    echo "Error: Number of honest replicas must be between 1 and 9"
    echo "Usage: ./demo-bft.sh [num_honest] [num_malicious]"
    echo "Example: ./demo-bft.sh 3 2  # 3 honest + 2 malicious replicas"
    exit 1
fi

if ! [[ "$NUM_MALICIOUS" =~ ^[0-9]+$ ]] || [ "$NUM_MALICIOUS" -lt 0 ] || [ "$NUM_MALICIOUS" -gt 9 ]; then
    echo "Error: Number of malicious replicas must be between 0 and 9"
    echo "Usage: ./demo-bft.sh [num_honest] [num_malicious]"
    echo "Example: ./demo-bft.sh 3 2  # 3 honest + 2 malicious replicas"
    exit 1
fi

TOTAL_REPLICAS=$((NUM_HONEST + NUM_MALICIOUS))

if [ "$TOTAL_REPLICAS" -gt 9 ]; then
    echo "Error: Total replicas (honest + malicious) cannot exceed 9"
    exit 1
fi

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PoH Blockchain BFT Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Leader: 1 node${NC}"
echo -e "${GREEN}  Honest Replicas: $NUM_HONEST nodes${NC}"
echo -e "${RED}  Malicious Replicas: $NUM_MALICIOUS nodes${NC}"
echo -e "${BLUE}  Total Replicas: $TOTAL_REPLICAS nodes${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# BFT tolerance calculation
# BFT requires: honest > 2 * malicious (i.e., > 2/3 honest)
if [ "$NUM_HONEST" -le $((2 * NUM_MALICIOUS)) ]; then
    echo -e "${RED}WARNING: Network does NOT have BFT!${NC}"
    echo -e "${RED}For BFT, honest nodes must be > 2 * malicious nodes${NC}"
    echo -e "${RED}Current: $NUM_HONEST honest <= 2 * $NUM_MALICIOUS malicious${NC}"
    echo -e "${YELLOW}The network may accept invalid blocks!${NC}"
    echo ""
else
    echo -e "${GREEN}✓ Network has BFT tolerance${NC}"
    echo -e "${GREEN}Honest nodes ($NUM_HONEST) > 2 * Malicious nodes ($NUM_MALICIOUS)${NC}"
    echo ""
fi

# Check if tmux is installed
if ! command -v tmux &> /dev/null; then
    echo -e "${YELLOW}Error: tmux is not installed${NC}"
    echo "Please install tmux first:"
    echo "  Ubuntu/Debian: sudo apt-get install tmux"
    echo "  macOS: brew install tmux"
    exit 1
fi

# Clean up old database files
echo -e "${GREEN}Cleaning up old database files...${NC}"
rm -f leader.db replica*.db malicious*.db

# Build the binary
echo -e "${GREEN}Building the PoH blockchain node...${NC}"
go build -o bin/poh-node cmd/main.go

if [ $? -ne 0 ]; then
    echo -e "${YELLOW}Build failed. Please fix compilation errors.${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful!${NC}"
echo ""

# Create a new tmux session
SESSION_NAME="poh-bft-demo"

# Kill existing session if it exists
tmux kill-session -t $SESSION_NAME 2>/dev/null || true

echo -e "${GREEN}Starting tmux session: $SESSION_NAME${NC}"
echo ""

# Create new session with first window
tmux new-session -d -s $SESSION_NAME -n "PoH-BFT"

# Start leader node in the first pane
tmux send-keys -t $SESSION_NAME:0.0 "echo -e '${BLUE}=== LEADER NODE (Port 8000) ===${NC}'" C-m
tmux send-keys -t $SESSION_NAME:0.0 "./bin/poh-node --type=leader --port=8000 --db=./leader.db" C-m

# Wait a moment for leader to start
sleep 1

# Create honest replica panes and start honest replica nodes
for i in $(seq 1 $NUM_HONEST); do
    PORT=$((8000 + i))
    
    # Split the window to create a new pane
    if [ $i -eq 1 ]; then
        tmux split-window -v -t $SESSION_NAME:0
    else
        if [ $((i % 2)) -eq 0 ]; then
            tmux split-window -h -t $SESSION_NAME:0
        else
            tmux split-window -v -t $SESSION_NAME:0
        fi
    fi
    
    # Start honest replica node
    tmux send-keys -t $SESSION_NAME:0.$i "echo -e '${GREEN}=== HONEST REPLICA $i (Port $PORT) ===${NC}'" C-m
    tmux send-keys -t $SESSION_NAME:0.$i "./bin/poh-node --type=replica --port=$PORT --peers=localhost:8000 --db=./replica$i.db" C-m
    
    sleep 0.2
done

# Create malicious replica panes and start malicious replica nodes
for i in $(seq 1 $NUM_MALICIOUS); do
    PANE_NUM=$((NUM_HONEST + i))
    PORT=$((8000 + PANE_NUM))
    
    # Split the window to create a new pane
    if [ $((PANE_NUM % 2)) -eq 0 ]; then
        tmux split-window -h -t $SESSION_NAME:0
    else
        tmux split-window -v -t $SESSION_NAME:0
    fi
    
    # Start malicious replica node
    tmux send-keys -t $SESSION_NAME:0.$PANE_NUM "echo -e '${RED}=== MALICIOUS REPLICA $i (Port $PORT) ===${NC}'" C-m
    tmux send-keys -t $SESSION_NAME:0.$PANE_NUM "./bin/poh-node --type=replica --port=$PORT --peers=localhost:8000 --db=./malicious$i.db --malicious" C-m
    
    sleep 0.2
done

# Balance the layout
tmux select-layout -t $SESSION_NAME:0 tiled

echo -e "${GREEN}BFT Demo started successfully!${NC}"
echo ""
echo -e "${BLUE}Network Configuration:${NC}"
echo "  - Leader: localhost:8000"
for i in $(seq 1 $NUM_HONEST); do
    PORT=$((8000 + i))
    echo -e "  - ${GREEN}Honest Replica $i: localhost:$PORT${NC}"
done
for i in $(seq 1 $NUM_MALICIOUS); do
    PANE_NUM=$((NUM_HONEST + i))
    PORT=$((8000 + PANE_NUM))
    echo -e "  - ${RED}Malicious Replica $i: localhost:$PORT${NC}"
done
echo ""
echo -e "${BLUE}Malicious Behaviors:${NC}"
echo "  - Corrupt blocks (invalid hash counts)"
echo "  - Send blocks with wrong previous hash"
echo "  - Skip validation and accept invalid blocks"
echo "  - Store unvalidated blocks"
echo ""
echo -e "${YELLOW}Watch the logs to see:${NC}"
echo "  - Honest nodes rejecting invalid blocks"
echo "  - Malicious nodes accepting/sending corrupted data"
echo "  - Network consensus despite Byzantine faults"
echo ""
echo -e "${BLUE}Commands:${NC}"
echo "  - Attach to session: ${YELLOW}tmux attach -t $SESSION_NAME${NC}"
echo "  - Detach from session: ${YELLOW}Ctrl+B, then D${NC}"
echo "  - Navigate panes: ${YELLOW}Ctrl+B, then arrow keys${NC}"
echo "  - Stop demo: ${YELLOW}Ctrl+C in each pane, or run: ./stop-demo.sh${NC}"
echo ""
echo -e "${GREEN}Attaching to tmux session...${NC}"
echo ""

# Attach to the session
tmux attach -t $SESSION_NAME

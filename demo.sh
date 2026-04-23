#!env bash

# PoH Blockchain Demo Script
# This script starts a leader node and multiple replica nodes in tmux panes

set -e

# Default number of replicas
NUM_REPLICAS=${1:-2}

# Validate input
if ! [[ "$NUM_REPLICAS" =~ ^[0-9]+$ ]] || [ "$NUM_REPLICAS" -lt 1 ] || [ "$NUM_REPLICAS" -gt 9 ]; then
    echo "Error: Number of replicas must be between 1 and 9"
    echo "Usage: ./demo.sh [num_replicas]"
    echo "Example: ./demo.sh 3"
    exit 1
fi

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PoH Blockchain Demo${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Leader: 1 node${NC}"
echo -e "${BLUE}  Replicas: $NUM_REPLICAS nodes${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

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
rm -f leader.db replica*.db

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
SESSION_NAME="poh-demo"

# Kill existing session if it exists
tmux kill-session -t $SESSION_NAME 2>/dev/null || true

echo -e "${GREEN}Starting tmux session: $SESSION_NAME${NC}"
echo ""

# Create new session with first window
tmux new-session -d -s $SESSION_NAME -n "PoH-Demo"

# Start leader node in the first pane
tmux send-keys -t $SESSION_NAME:0.0 "echo -e '${BLUE}=== LEADER NODE (Port 8000) ===${NC}'" C-m
tmux send-keys -t $SESSION_NAME:0.0 "./bin/poh-node --type=leader --port=8000 --db=./leader.db" C-m

# Wait a moment for leader to start
sleep 1

# Create replica panes and start replica nodes
for i in $(seq 1 $NUM_REPLICAS); do
    PORT=$((8000 + i))
    
    # Split the window to create a new pane
    if [ $i -eq 1 ]; then
        # First split: horizontal split (top and bottom)
        tmux split-window -v -t $SESSION_NAME:0
    else
        # Subsequent splits: split the last pane vertically or horizontally
        if [ $((i % 2)) -eq 0 ]; then
            tmux split-window -h -t $SESSION_NAME:0
        else
            tmux split-window -v -t $SESSION_NAME:0
        fi
    fi
    
    # Start replica node in the new pane
    tmux send-keys -t $SESSION_NAME:0.$i "echo -e '${GREEN}=== REPLICA $i (Port $PORT) ===${NC}'" C-m
    tmux send-keys -t $SESSION_NAME:0.$i "./bin/poh-node --type=replica --port=$PORT --peers=localhost:8000 --db=./replica$i.db" C-m
    
    # Small delay between starting replicas
    sleep 0.2
done

# Balance the layout
tmux select-layout -t $SESSION_NAME:0 tiled

echo -e "${GREEN}Demo started successfully!${NC}"
echo ""
echo -e "${BLUE}Network Configuration:${NC}"
echo "  - Leader: localhost:8000"
for i in $(seq 1 $NUM_REPLICAS); do
    PORT=$((8000 + i))
    echo "  - Replica $i: localhost:$PORT"
done
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

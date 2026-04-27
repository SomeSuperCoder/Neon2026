#!/usr/bin/env bash

# Stop the PoH Blockchain Demo
# This script stops all running blockchain nodes including:
# - Legacy tmux sessions (poh-demo, poh-bft-demo)
# - New devnet instances
# - Any stray poh-node processes

SESSION_NAME="poh-demo"
BFT_SESSION_NAME="poh-bft-demo"

echo "Stopping PoH Blockchain demo..."

# Stop legacy tmux sessions
STOPPED=0

tmux kill-session -t $SESSION_NAME 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ Regular demo stopped successfully!"
    STOPPED=1
fi

tmux kill-session -t $BFT_SESSION_NAME 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✓ BFT demo stopped successfully!"
    STOPPED=1
fi

# Stop devnet if running
if [ -f "./devnet.sh" ]; then
    if [ -d "./devnet-data/pids" ] && [ "$(ls -A ./devnet-data/pids 2>/dev/null)" ]; then
        echo "Stopping devnet..."
        ./devnet.sh stop
        STOPPED=1
    fi
fi

# Kill any stray poh-node processes
STRAY_PIDS=$(pgrep -f "poh-node" 2>/dev/null)
if [ -n "$STRAY_PIDS" ]; then
    echo "Killing stray poh-node processes..."
    pkill -TERM -f "poh-node" 2>/dev/null || true
    sleep 1
    # Force kill if still running
    pkill -9 -f "poh-node" 2>/dev/null || true
    STOPPED=1
fi

if [ $STOPPED -eq 0 ]; then
    echo "No running demo session found."
fi

# Clean up database files (optional)
read -p "Do you want to clean up database files? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f leader.db replica*.db malicious*.db validator*.db
    echo "Database files cleaned up."
fi

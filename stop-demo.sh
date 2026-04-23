#!env bash

# Stop the PoH Blockchain Demo

SESSION_NAME="poh-demo"
BFT_SESSION_NAME="poh-bft-demo"

echo "Stopping PoH Blockchain demo..."

# Try to kill both possible sessions
STOPPED=0

tmux kill-session -t $SESSION_NAME 2>/dev/null
if [ $? -eq 0 ]; then
    echo "Regular demo stopped successfully!"
    STOPPED=1
fi

tmux kill-session -t $BFT_SESSION_NAME 2>/dev/null
if [ $? -eq 0 ]; then
    echo "BFT demo stopped successfully!"
    STOPPED=1
fi

if [ $STOPPED -eq 0 ]; then
    echo "No running demo session found."
fi

# Clean up database files (optional)
read -p "Do you want to clean up database files? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -f leader.db replica*.db malicious*.db
    echo "Database files cleaned up."
fi

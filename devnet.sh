#!/usr/bin/env bash
set -e

# Default configuration
COMMAND=${1:-start}
NUM_VALIDATORS=3
PORT_START=8000
RPC_PORT=8899
DB_DIR="./devnet-data"
LOG_DIR="./logs"
PID_DIR="$DB_DIR/pids"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command-line options
parse_options() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --port)
        PORT_START="$2"
        shift 2
        ;;
      --rpc-port)
        RPC_PORT="$2"
        shift 2
        ;;
      --db-dir)
        DB_DIR="$2"
        PID_DIR="$DB_DIR/pids"
        shift 2
        ;;
      --log-dir)
        LOG_DIR="$2"
        shift 2
        ;;
      --help)
        show_help
        exit 0
        ;;
      *)
        shift
        ;;
    esac
  done
}

# Show help message
show_help() {
  cat << EOF
Usage: ./devnet.sh [COMMAND] [N] [OPTIONS]

Commands:
  start [N]             Start devnet with N validators (default: 3)
  stop                  Stop running devnet
  restart [N]           Restart devnet with N validators
  status                Show devnet status
  logs [VALIDATOR_ID]   Show logs for specific validator, RPC node, or all
  clean                 Stop devnet and clean all data

Options:
  --port PORT           Starting port number (default: 8000)
  --rpc-port PORT       RPC node port (default: 8899)
  --db-dir DIR          Database directory (default: ./devnet-data)
  --log-dir DIR         Log directory (default: ./logs)
  --help                Show this help message

Examples:
  ./devnet.sh start 3                    # Start 3 validators
  ./devnet.sh start 5 --port 9000        # Start 5 validators on port 9000
  ./devnet.sh start 3 --rpc-port 9899    # Start with RPC on port 9899
  ./devnet.sh status                     # Check devnet status
  ./devnet.sh logs 1                     # Show logs for validator 1
  ./devnet.sh logs rpc                   # Show logs for RPC node
  ./devnet.sh logs                       # Show logs for all validators and RPC
  ./devnet.sh stop                       # Stop devnet
  ./devnet.sh clean                      # Stop and clean all data

EOF
}

# Start devnet
start_devnet() {
  local num=$1
  
  echo -e "${BLUE}Starting devnet with $num validators...${NC}"
  
  # Check if devnet is already running
  if [ -d "$PID_DIR" ] && [ "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}Warning: Devnet appears to be already running${NC}"
    echo "Use './devnet.sh stop' first or './devnet.sh restart' to restart"
    exit 1
  fi
  
  # Create directories
  mkdir -p "$DB_DIR" "$LOG_DIR" "$PID_DIR"
  
  # Build binary
  echo -e "${BLUE}Building poh-node binary...${NC}"
  if ! go build -o bin/poh-node ./cmd/main.go; then
    echo -e "${RED}Failed to build poh-node binary${NC}"
    exit 1
  fi
  
  # Create wallets for validators if they don't exist
  echo -e "${BLUE}Creating wallets for validators...${NC}"
  for i in $(seq 1 $num); do
    local wallet_name="validator$i"
    if ! ./bin/poh-node wallet list 2>/dev/null | grep -q "^$wallet_name$"; then
      echo "Creating wallet: $wallet_name"
      echo "development-password" | ./bin/poh-node wallet create --name "$wallet_name" --password "development-password" 2>/dev/null || true
    else
      echo "Wallet already exists: $wallet_name"
    fi
  done
  
  # Start leader
  echo -e "${BLUE}Starting leader (validator 1) on port $PORT_START...${NC}"
  WALLET_PASSWORD="development-password" ./bin/poh-node --wallet=validator1 --port=$PORT_START \
    --db="$DB_DIR/validator1.db" \
    > "$LOG_DIR/devnet-validator-1.log" 2>&1 &
  echo $! > "$PID_DIR/validator1.pid"
  
  # Give leader time to start
  sleep 1
  
  # Start replicas
  for i in $(seq 2 $num); do
    local port=$((PORT_START + i - 1))
    echo -e "${BLUE}Starting replica (validator $i) on port $port...${NC}"
    WALLET_PASSWORD="development-password" ./bin/poh-node --wallet=validator$i --port=$port \
      --peers=localhost:$PORT_START \
      --db="$DB_DIR/validator$i.db" \
      > "$LOG_DIR/devnet-validator-$i.log" 2>&1 &
    echo $! > "$PID_DIR/validator$i.pid"
    sleep 0.5
  done
  
  # Wait a moment for nodes to initialize
  sleep 3
  
  # Start RPC node
  start_rpc_node
  
  # Display network info
  show_network_info $num
}

# Start RPC node
start_rpc_node() {
  echo -e "${BLUE}Starting RPC node on port $RPC_PORT...${NC}"
  
  # Use leader's databases
  local leader_db="$DB_DIR/validator1.db"
  local leader_state="$DB_DIR/validator1_state.db"
  local rpc_state="$DB_DIR/rpc_state.db"
  
  # Wait for leader databases to be created and available
  local max_wait=10
  local wait_count=0
  while [ ! -f "$leader_db" ] && [ $wait_count -lt $max_wait ]; do
    echo "Waiting for leader database to be created..."
    sleep 1
    wait_count=$((wait_count + 1))
  done
  
  if [ ! -f "$leader_db" ]; then
    echo -e "${RED}Error: Leader database not found after waiting: $leader_db${NC}"
    return 1
  fi
  
  # Create a copy of the state database for RPC node to avoid lock conflicts
  echo "Creating RPC state database copy..."
  if [ -d "$leader_state" ]; then
    # Remove existing RPC state database if it exists
    rm -rf "$rpc_state" 2>/dev/null || true
    # Copy the leader's state database
    cp -r "$leader_state" "$rpc_state"
  else
    # If state database doesn't exist yet, create an empty one
    mkdir -p "$rpc_state"
  fi
  
  # Start RPC node with its own state database copy
  ./bin/poh-node rpc \
    --rpc-port=$RPC_PORT \
    --rpc-bind=127.0.0.1 \
    --ledger-path="$leader_db" \
    --state-path="$rpc_state" \
    > "$LOG_DIR/devnet-rpc.log" 2>&1 &
  
  local rpc_pid=$!
  echo $rpc_pid > "$PID_DIR/rpc.pid"
  
  # Give RPC node time to start
  sleep 1
  
  # Check if RPC node started successfully
  if kill -0 $rpc_pid 2>/dev/null; then
    echo -e "${GREEN}✓ RPC node started (PID: $rpc_pid)${NC}"
    echo "  Endpoint: http://127.0.0.1:$RPC_PORT"
    echo "  Note: RPC uses a snapshot of the state database"
  else
    echo -e "${RED}✗ RPC node failed to start${NC}"
    echo "  Check logs: $LOG_DIR/devnet-rpc.log"
    return 1
  fi
}

# Show network information
show_network_info() {
  local num=$1
  
  echo ""
  echo -e "${GREEN}✓ Devnet started successfully!${NC}"
  echo ""
  echo "Network Configuration:"
  echo "====================="
  
  for i in $(seq 1 $num); do
    local port=$((PORT_START + i - 1))
    local role="replica"
    if [ $i -eq 1 ]; then
      role="leader"
    fi
    
    echo -e "  Validator $i ($role):"
    echo "    Address:  localhost:$port"
    echo "    Database: $DB_DIR/validator$i.db"
    echo "    Logs:     $LOG_DIR/devnet-validator-$i.log"
    echo "    PID:      $(cat $PID_DIR/validator$i.pid 2>/dev/null || echo 'unknown')"
    echo ""
  done
  
  # Show RPC node info if running
  if [ -f "$PID_DIR/rpc.pid" ]; then
    local rpc_pid=$(cat "$PID_DIR/rpc.pid" 2>/dev/null)
    if [ -n "$rpc_pid" ] && kill -0 $rpc_pid 2>/dev/null; then
      echo -e "  RPC Node:"
      echo "    Endpoint: http://127.0.0.1:$RPC_PORT"
      echo "    Logs:     $LOG_DIR/devnet-rpc.log"
      echo "    PID:      $rpc_pid"
      echo ""
    fi
  fi
  
  echo "Management Commands:"
  echo "  Status:  ./devnet.sh status"
  echo "  Logs:    ./devnet.sh logs [validator_id]"
  echo "  Stop:    ./devnet.sh stop"
  echo "  Clean:   ./devnet.sh clean"
  echo ""
}

# Stop devnet
stop_devnet() {
  echo -e "${BLUE}Stopping devnet...${NC}"
  
  if [ ! -d "$PID_DIR" ] || [ ! "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}No running devnet found${NC}"
    return 0
  fi
  
  local stopped=0
  
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      local pid=$(cat "$pid_file")
      local process_name=$(basename "$pid_file" .pid)
      
      if kill -0 $pid 2>/dev/null; then
        if [ "$process_name" = "rpc" ]; then
          echo "Stopping RPC node (PID: $pid)..."
        else
          echo "Stopping $process_name (PID: $pid)..."
        fi
        kill -TERM $pid 2>/dev/null || true
        stopped=$((stopped + 1))
      fi
      
      rm "$pid_file"
    fi
  done
  
  if [ $stopped -gt 0 ]; then
    # Wait for graceful shutdown
    echo "Waiting for graceful shutdown..."
    sleep 2
    
    # Force kill if still running
    for pid_file in "$PID_DIR"/*.pid; do
      if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 $pid 2>/dev/null; then
          echo "Force killing PID $pid..."
          kill -9 $pid 2>/dev/null || true
        fi
      fi
    done
    
    # Additional cleanup: kill any stray poh-node processes using devnet-data
    pkill -9 -f "poh-node.*devnet-data" 2>/dev/null || true
    
    echo -e "${GREEN}✓ Devnet stopped${NC}"
  else
    echo -e "${YELLOW}No running processes found${NC}"
  fi
}

# Show devnet status
show_devnet_status() {
  echo "Devnet Status:"
  echo "=============="
  echo ""
  
  if [ ! -d "$PID_DIR" ] || [ ! "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}Devnet is not running${NC}"
    echo ""
    echo "Start devnet with: ./devnet.sh start [N]"
    return 0
  fi
  
  local running=0
  local stopped=0
  
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      local pid=$(cat "$pid_file")
      local process_name=$(basename "$pid_file" .pid)
      
      if [ "$process_name" = "rpc" ]; then
        # Handle RPC node
        if kill -0 $pid 2>/dev/null; then
          echo -e "${GREEN}✓${NC} RPC Node (PID: $pid, Port: $RPC_PORT) - RUNNING"
          running=$((running + 1))
          echo "    Endpoint: http://127.0.0.1:$RPC_PORT"
        else
          echo -e "${RED}✗${NC} RPC Node (PID: $pid) - STOPPED"
          stopped=$((stopped + 1))
        fi
      else
        # Handle validator nodes
        local validator_num=$(echo "$process_name" | grep -o '[0-9]\+')
        local port=$((PORT_START + validator_num - 1))
        
        if kill -0 $pid 2>/dev/null; then
          echo -e "${GREEN}✓${NC} $process_name (PID: $pid, Port: $port) - RUNNING"
          running=$((running + 1))
          
          # Try to get block count from database
          local db_file="$DB_DIR/$process_name.db"
          if [ -f "$db_file" ]; then
            local block_count=$(sqlite3 "$db_file" "SELECT COUNT(*) FROM blocks;" 2>/dev/null || echo "N/A")
            echo "    Blocks: $block_count"
          fi
        else
          echo -e "${RED}✗${NC} $process_name (PID: $pid) - STOPPED"
          stopped=$((stopped + 1))
        fi
      fi
      echo ""
    fi
  done
  
  echo "Summary: $running running, $stopped stopped"
  echo ""
}

# Show logs
show_devnet_logs() {
  local validator_id=$1
  
  if [ -z "$validator_id" ]; then
    # Show all logs
    echo "Showing logs for all validators and RPC node (Ctrl+C to exit)..."
    echo ""
    
    local log_files=()
    for log_file in "$LOG_DIR"/devnet-validator-*.log; do
      if [ -f "$log_file" ]; then
        log_files+=("$log_file")
      fi
    done
    
    # Add RPC log if it exists
    if [ -f "$LOG_DIR/devnet-rpc.log" ]; then
      log_files+=("$LOG_DIR/devnet-rpc.log")
    fi
    
    if [ ${#log_files[@]} -eq 0 ]; then
      echo -e "${YELLOW}No log files found${NC}"
      return 1
    fi
    
    tail -f "${log_files[@]}"
  elif [ "$validator_id" = "rpc" ]; then
    # Show RPC log
    local log_file="$LOG_DIR/devnet-rpc.log"
    
    if [ ! -f "$log_file" ]; then
      echo -e "${RED}RPC log file not found: $log_file${NC}"
      echo "Make sure the RPC node is running or has been started"
      return 1
    fi
    
    echo "Showing logs for RPC node (Ctrl+C to exit)..."
    echo ""
    tail -f "$log_file"
  else
    # Show specific validator log
    local log_file="$LOG_DIR/devnet-validator-$validator_id.log"
    
    if [ ! -f "$log_file" ]; then
      echo -e "${RED}Log file not found: $log_file${NC}"
      echo "Available options:"
      ls -1 "$LOG_DIR"/devnet-validator-*.log 2>/dev/null | sed 's/.*validator-\([0-9]\+\).*/  \1/' || echo "  None"
      if [ -f "$LOG_DIR/devnet-rpc.log" ]; then
        echo "  rpc"
      fi
      return 1
    fi
    
    echo "Showing logs for validator $validator_id (Ctrl+C to exit)..."
    echo ""
    tail -f "$log_file"
  fi
}

# Restart devnet
restart_devnet() {
  local num=$1
  
  echo -e "${BLUE}Restarting devnet...${NC}"
  stop_devnet
  sleep 1
  start_devnet $num
}

# Clean devnet data
clean_devnet() {
  echo -e "${YELLOW}This will stop the devnet and remove all data and logs.${NC}"
  read -p "Are you sure? (y/N): " -n 1 -r
  echo
  
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    return 0
  fi
  
  echo -e "${BLUE}Cleaning devnet...${NC}"
  
  # Stop devnet first
  stop_devnet
  
  # Remove data directories
  if [ -d "$DB_DIR" ]; then
    echo "Removing $DB_DIR..."
    rm -rf "$DB_DIR"
  fi
  
  # Remove log files
  if [ -d "$LOG_DIR" ]; then
    echo "Removing devnet logs from $LOG_DIR..."
    rm -f "$LOG_DIR"/devnet-validator-*.log
    rm -f "$LOG_DIR"/devnet-rpc.log
  fi
  
  echo -e "${GREEN}✓ Devnet cleaned${NC}"
}

# Main command dispatcher
main() {
  # Extract command and number of validators
  local cmd=$1
  shift || true
  
  local num=$NUM_VALIDATORS
  local logs_param=""
  
  # Special handling for logs command
  if [ "$cmd" = "logs" ]; then
    logs_param=$1
    shift || true
  elif [[ $1 =~ ^[0-9]+$ ]]; then
    num=$1
    shift || true
  fi
  
  # Parse remaining options
  parse_options "$@"
  
  case $cmd in
    start)
      start_devnet $num
      ;;
    stop)
      stop_devnet
      ;;
    restart)
      restart_devnet $num
      ;;
    status)
      show_devnet_status
      ;;
    logs)
      show_devnet_logs $logs_param
      ;;
    clean)
      clean_devnet
      ;;
    --help|-h|help)
      show_help
      ;;
    *)
      echo -e "${RED}Unknown command: $cmd${NC}"
      echo ""
      show_help
      exit 1
      ;;
  esac
}

# Run main function
main "$@"

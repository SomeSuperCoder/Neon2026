#!/usr/bin/env bash
set -e

# DPoS Demo Script - Stake-Weighted Leader Schedule
# This script demonstrates the stake-weighted leader schedule with wallet-based validators

# Default configuration
COMMAND=${1:-start}
NUM_VALIDATORS=3
PORT_START=8000
RPC_PORT=8899
DB_DIR="./devnet-data"
LOG_DIR="./logs"
PID_DIR="$DB_DIR/pids"
WALLET_DIR="${HOME}/.config/poh-blockchain/wallets"
GENESIS_CONFIG="./genesis-dpos.json"
DEMO_DURATION=30

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
      --validators)
        NUM_VALIDATORS="$2"
        shift 2
        ;;
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
      --duration)
        DEMO_DURATION="$2"
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
Usage: ./demo-dpos.sh [COMMAND] [OPTIONS]

Commands:
  start [N]             Start DPoS demo with N validators (default: 3)
  stop                  Stop running demo
  restart [N]           Restart demo with N validators
  status                Show demo status
  logs [VALIDATOR_ID]   Show logs for specific validator or all
  clean                 Stop demo and clean all data

Options:
  --validators N        Number of validators (default: 3)
  --port PORT           Starting port number (default: 8000)
  --rpc-port PORT       RPC node port (default: 8899)
  --db-dir DIR          Database directory (default: ./devnet-data)
  --log-dir DIR         Log directory (default: ./logs)
  --duration SECONDS    Demo duration in seconds (default: 30)
  --help                Show this help message

Examples:
  ./demo-dpos.sh start 3                    # Start 3 validators
  ./demo-dpos.sh start 5 --duration 60      # Start 5 validators for 60 seconds
  ./demo-dpos.sh status                     # Check demo status
  ./demo-dpos.sh logs 1                     # Show logs for validator 1
  ./demo-dpos.sh stop                       # Stop demo
  ./demo-dpos.sh clean                      # Stop and clean all data

EOF
}

# Create genesis configuration
create_genesis_config() {
  local num=$1
  
  echo -e "${BLUE}Creating genesis configuration for $num validators...${NC}"
  
  # Calculate stake distribution: first validator gets 2x stake
  # This allows us to verify stake-weighted slot distribution
  local base_stake=5000000  # 5 Neon
  
  # Start building the JSON
  local json='{"epochLength": 432000, "validators": ['
  
  for i in $(seq 1 $num); do
    local stake=$base_stake
    
    # First validator gets 2x stake for testing stake-weighted distribution
    if [ $i -eq 1 ]; then
      stake=$((base_stake * 2))
    fi
    
    # Get the public key from the wallet
    local wallet_name="dpos-validator-$i"
    local wallet_path="$WALLET_DIR/$wallet_name.wallet"
    
    if [ ! -f "$wallet_path" ]; then
      echo -e "${RED}Error: Wallet not found: $wallet_path${NC}"
      return 1
    fi
    
    # Extract public key from wallet (this is a simplified approach)
    # In production, we'd use a proper wallet inspection tool
    # For now, we'll use a placeholder that will be replaced by the node
    local pubkey_placeholder="0x$(printf '%064d' $i)"
    
    if [ $i -gt 1 ]; then
      json="$json,"
    fi
    
    json="$json{\"publicKey\": \"$pubkey_placeholder\", \"stake\": $stake}"
  done
  
  json="$json]}"
  
  # Write genesis config
  echo "$json" > "$GENESIS_CONFIG"
  echo -e "${GREEN}✓ Genesis configuration created: $GENESIS_CONFIG${NC}"
}

# Create wallets for validators
create_wallets() {
  local num=$1
  
  echo -e "${BLUE}Creating wallets for $num validators...${NC}"
  
  # Create wallet directory if it doesn't exist
  mkdir -p "$WALLET_DIR"
  
  for i in $(seq 1 $num); do
    local wallet_name="dpos-validator-$i"
    local wallet_path="$WALLET_DIR/$wallet_name.wallet"
    
    # Check if wallet already exists
    if [ -f "$wallet_path" ]; then
      echo -e "${YELLOW}Wallet already exists: $wallet_name${NC}"
    else
      echo "Creating wallet: $wallet_name"
      
      # Create wallet with non-interactive password
      # Use a simple password for demo purposes
      local password="demo-password-$i"
      
      # Use the wallet CLI to create the wallet
      # We need to pipe the password twice (for confirmation)
      echo -e "$password\n$password" | go run cmd/main.go wallet create --name "$wallet_name" > /dev/null 2>&1
      
      if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Wallet created: $wallet_name${NC}"
      else
        echo -e "${RED}✗ Failed to create wallet: $wallet_name${NC}"
        return 1
      fi
    fi
  done
}

# Start DPoS demo
start_demo() {
  local num=$1
  
  echo -e "${BLUE}Starting DPoS demo with $num validators...${NC}"
  
  # Check if demo is already running
  if [ -d "$PID_DIR" ] && [ "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}Warning: Demo appears to be already running${NC}"
    echo "Use './demo-dpos.sh stop' first or './demo-dpos.sh restart' to restart"
    exit 1
  fi
  
  # Create directories
  mkdir -p "$DB_DIR" "$LOG_DIR" "$PID_DIR"
  
  # Build binary
  echo -e "${BLUE}Building poh-node binary...${NC}"
  if ! go build -o bin/poh-node cmd/main.go; then
    echo -e "${RED}Failed to build poh-node binary${NC}"
    exit 1
  fi
  
  # Create wallets
  if ! create_wallets $num; then
    echo -e "${RED}Failed to create wallets${NC}"
    exit 1
  fi
  
  # Create genesis configuration
  if ! create_genesis_config $num; then
    echo -e "${RED}Failed to create genesis configuration${NC}"
    exit 1
  fi
  
  # Start validators
  for i in $(seq 1 $num); do
    local port=$((PORT_START + i - 1))
    local wallet_name="dpos-validator-$i"
    local password="demo-password-$i"
    
    echo -e "${BLUE}Starting validator $i on port $port...${NC}"
    
    # Build the command with wallet and genesis config (only for first validator)
    local cmd="./bin/poh-node --wallet=$wallet_name --port=$port --db=$DB_DIR/validator$i.db"
    
    if [ $i -eq 1 ]; then
      cmd="$cmd --genesis-config=$GENESIS_CONFIG"
    fi
    
    # Start the validator with password piped in
    echo "$password" | $cmd > "$LOG_DIR/devnet-validator-$i.log" 2>&1 &
    local pid=$!
    echo $pid > "$PID_DIR/validator$i.pid"
    
    echo -e "${GREEN}✓ Validator $i started (PID: $pid)${NC}"
    sleep 0.5
  done
  
  # Wait for validators to initialize
  sleep 3
  
  # Start RPC node
  start_rpc_node
  
  # Display network info
  show_network_info $num
  
  # Run demo for specified duration
  echo ""
  echo -e "${BLUE}Running demo for $DEMO_DURATION seconds...${NC}"
  echo "Monitoring block production and stake-weighted slot distribution..."
  echo ""
  
  sleep $DEMO_DURATION
  
  # Collect statistics
  collect_statistics $num
}

# Start RPC node
start_rpc_node() {
  echo -e "${BLUE}Starting RPC node on port $RPC_PORT...${NC}"
  
  local leader_db="$DB_DIR/validator1.db"
  local leader_state="$DB_DIR/validator1_state.db"
  local rpc_state="$DB_DIR/rpc_state.db"
  
  # Wait for leader database to be created
  local max_wait=10
  local wait_count=0
  while [ ! -f "$leader_db" ] && [ $wait_count -lt $max_wait ]; do
    echo "Waiting for leader database..."
    sleep 1
    wait_count=$((wait_count + 1))
  done
  
  if [ ! -f "$leader_db" ]; then
    echo -e "${RED}Error: Leader database not found${NC}"
    return 1
  fi
  
  # Create RPC state database copy
  if [ -d "$leader_state" ]; then
    rm -rf "$rpc_state" 2>/dev/null || true
    cp -r "$leader_state" "$rpc_state"
  else
    mkdir -p "$rpc_state"
  fi
  
  # Start RPC node
  ./bin/poh-node rpc \
    --rpc-port=$RPC_PORT \
    --rpc-bind=127.0.0.1 \
    --ledger-path="$leader_db" \
    --state-path="$rpc_state" \
    > "$LOG_DIR/devnet-rpc.log" 2>&1 &
  
  local rpc_pid=$!
  echo $rpc_pid > "$PID_DIR/rpc.pid"
  
  sleep 1
  
  if kill -0 $rpc_pid 2>/dev/null; then
    echo -e "${GREEN}✓ RPC node started (PID: $rpc_pid)${NC}"
  else
    echo -e "${RED}✗ RPC node failed to start${NC}"
    return 1
  fi
}

# Collect statistics from validator logs
collect_statistics() {
  local num=$1
  
  echo ""
  echo -e "${BLUE}=== DPoS Demo Statistics ===${NC}"
  echo ""
  
  # Count blocks produced by each validator
  for i in $(seq 1 $num); do
    local log_file="$LOG_DIR/devnet-validator-$i.log"
    
    if [ -f "$log_file" ]; then
      # Count "Producing block" messages
      local block_count=$(grep -c "Producing block" "$log_file" 2>/dev/null || echo "0")
      
      # Count "Scheduled leader" messages
      local leader_count=$(grep -c "Scheduled leader" "$log_file" 2>/dev/null || echo "0")
      
      echo "Validator $i:"
      echo "  Blocks produced: $block_count"
      echo "  Times scheduled as leader: $leader_count"
      echo ""
    fi
  done
  
  # Verify stake-weighted distribution
  echo -e "${BLUE}Stake-Weighted Distribution Analysis:${NC}"
  echo "Expected: Validator 1 (2x stake) should produce ~2x blocks as others"
  echo ""
}

# Show network information
show_network_info() {
  local num=$1
  
  echo ""
  echo -e "${GREEN}✓ DPoS demo started successfully!${NC}"
  echo ""
  echo "Network Configuration:"
  echo "====================="
  
  for i in $(seq 1 $num); do
    local port=$((PORT_START + i - 1))
    local wallet_name="dpos-validator-$i"
    
    echo -e "  Validator $i:"
    echo "    Wallet:   $wallet_name"
    echo "    Address:  localhost:$port"
    echo "    Database: $DB_DIR/validator$i.db"
    echo "    Logs:     $LOG_DIR/devnet-validator-$i.log"
    echo "    PID:      $(cat $PID_DIR/validator$i.pid 2>/dev/null || echo 'unknown')"
    echo ""
  done
  
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
  
  echo "Genesis Configuration:"
  echo "  File: $GENESIS_CONFIG"
  echo ""
}

# Stop demo
stop_demo() {
  echo -e "${BLUE}Stopping DPoS demo...${NC}"
  
  if [ ! -d "$PID_DIR" ] || [ ! "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}No running demo found${NC}"
    return 0
  fi
  
  local stopped=0
  
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      local pid=$(cat "$pid_file")
      local process_name=$(basename "$pid_file" .pid)
      
      if kill -0 $pid 2>/dev/null; then
        echo "Stopping $process_name (PID: $pid)..."
        kill -TERM $pid 2>/dev/null || true
        stopped=$((stopped + 1))
      fi
      
      rm "$pid_file"
    fi
  done
  
  if [ $stopped -gt 0 ]; then
    sleep 2
    
    # Force kill if still running
    for pid_file in "$PID_DIR"/*.pid; do
      if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 $pid 2>/dev/null; then
          kill -9 $pid 2>/dev/null || true
        fi
      fi
    done
    
    pkill -9 -f "poh-node.*devnet-data" 2>/dev/null || true
    
    echo -e "${GREEN}✓ DPoS demo stopped${NC}"
  else
    echo -e "${YELLOW}No running processes found${NC}"
  fi
}

# Show demo status
show_demo_status() {
  echo "DPoS Demo Status:"
  echo "================"
  echo ""
  
  if [ ! -d "$PID_DIR" ] || [ ! "$(ls -A $PID_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}Demo is not running${NC}"
    echo ""
    echo "Start demo with: ./demo-dpos.sh start [N]"
    return 0
  fi
  
  local running=0
  local stopped=0
  
  for pid_file in "$PID_DIR"/*.pid; do
    if [ -f "$pid_file" ]; then
      local pid=$(cat "$pid_file")
      local process_name=$(basename "$pid_file" .pid)
      
      if [ "$process_name" = "rpc" ]; then
        if kill -0 $pid 2>/dev/null; then
          echo -e "${GREEN}✓${NC} RPC Node (PID: $pid, Port: $RPC_PORT) - RUNNING"
          running=$((running + 1))
        else
          echo -e "${RED}✗${NC} RPC Node (PID: $pid) - STOPPED"
          stopped=$((stopped + 1))
        fi
      else
        local validator_num=$(echo "$process_name" | grep -o '[0-9]\+')
        local port=$((PORT_START + validator_num - 1))
        
        if kill -0 $pid 2>/dev/null; then
          echo -e "${GREEN}✓${NC} $process_name (PID: $pid, Port: $port) - RUNNING"
          running=$((running + 1))
          
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
show_demo_logs() {
  local validator_id=$1
  
  if [ -z "$validator_id" ]; then
    echo "Showing logs for all validators and RPC node (Ctrl+C to exit)..."
    echo ""
    
    local log_files=()
    for log_file in "$LOG_DIR"/devnet-validator-*.log; do
      if [ -f "$log_file" ]; then
        log_files+=("$log_file")
      fi
    done
    
    if [ -f "$LOG_DIR/devnet-rpc.log" ]; then
      log_files+=("$LOG_DIR/devnet-rpc.log")
    fi
    
    if [ ${#log_files[@]} -eq 0 ]; then
      echo -e "${YELLOW}No log files found${NC}"
      return 1
    fi
    
    tail -f "${log_files[@]}"
  elif [ "$validator_id" = "rpc" ]; then
    local log_file="$LOG_DIR/devnet-rpc.log"
    
    if [ ! -f "$log_file" ]; then
      echo -e "${RED}RPC log file not found: $log_file${NC}"
      return 1
    fi
    
    echo "Showing logs for RPC node (Ctrl+C to exit)..."
    echo ""
    tail -f "$log_file"
  else
    local log_file="$LOG_DIR/devnet-validator-$validator_id.log"
    
    if [ ! -f "$log_file" ]; then
      echo -e "${RED}Log file not found: $log_file${NC}"
      return 1
    fi
    
    echo "Showing logs for validator $validator_id (Ctrl+C to exit)..."
    echo ""
    tail -f "$log_file"
  fi
}

# Restart demo
restart_demo() {
  local num=$1
  
  echo -e "${BLUE}Restarting DPoS demo...${NC}"
  stop_demo
  sleep 1
  start_demo $num
}

# Clean demo data
clean_demo() {
  echo -e "${YELLOW}This will stop the demo and remove all data and logs.${NC}"
  read -p "Are you sure? (y/N): " -n 1 -r
  echo
  
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    return 0
  fi
  
  echo -e "${BLUE}Cleaning DPoS demo...${NC}"
  
  stop_demo
  
  if [ -d "$DB_DIR" ]; then
    echo "Removing $DB_DIR..."
    rm -rf "$DB_DIR"
  fi
  
  if [ -d "$LOG_DIR" ]; then
    echo "Removing devnet logs from $LOG_DIR..."
    rm -f "$LOG_DIR"/devnet-validator-*.log
    rm -f "$LOG_DIR"/devnet-rpc.log
  fi
  
  if [ -f "$GENESIS_CONFIG" ]; then
    echo "Removing genesis configuration..."
    rm -f "$GENESIS_CONFIG"
  fi
  
  echo -e "${GREEN}✓ DPoS demo cleaned${NC}"
}

# Main command dispatcher
main() {
  local cmd=$1
  shift || true
  
  local num=$NUM_VALIDATORS
  local logs_param=""
  
  if [ "$cmd" = "logs" ]; then
    logs_param=$1
    shift || true
  elif [[ $1 =~ ^[0-9]+$ ]]; then
    num=$1
    shift || true
  fi
  
  parse_options "$@"
  
  case $cmd in
    start)
      start_demo $num
      ;;
    stop)
      stop_demo
      ;;
    restart)
      restart_demo $num
      ;;
    status)
      show_demo_status
      ;;
    logs)
      show_demo_logs $logs_param
      ;;
    clean)
      clean_demo
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

# Validator TUI Guide

The Validator TUI is a real-time terminal dashboard for monitoring validator and staking information on the PoH blockchain. It provides a live view of epoch state, validator records, stake accounts, and reward pool information.

Built with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework and [lipgloss](https://github.com/charmbracelet/lipgloss) for rich terminal styling.

## Building the Validator TUI

```bash
# Build the validator TUI binary
go build -o validator-tui ./cmd/validator-tui/main.go
```

## Running the Validator TUI

```bash
# Point to a FileStore database
./validator-tui --state ./state.db

# Or with a full path
./validator-tui --state /path/to/blockchain/state.db
```

**Required Flags:**
- `--state <path>` - Path to the FileStore database (required)

## Dashboard Overview

The validator TUI displays four main sections:

### 1. Header Panel

Shows current epoch, slot, local validator status, active validator count, and local delegated stake:

```
Epoch: 5 | Slot: 2160 | Local Validator: active | Active Validators: 10 | Local Stake: 5000000 electrons
```

**Fields:**
- **Epoch**: Current epoch number
- **Slot**: Current slot within the epoch
- **Local Validator**: Status of the local validator (not_set, inactive, active, deregistered, not_found)
- **Active Validators**: Count of validators with active status
- **Local Stake**: Total delegated stake for the local validator

### 2. Validator Records Table

Displays all validator records with their key metrics:

```
┌──────────────────┬──────────┬──────────────────┬────────┬────────┬──────┬────────┐
│ Pubkey (16 hex)  │ Status   │ Total Stake      │ Comm % │ Blocks │ Miss │ Slashed│
├──────────────────┼──────────┼──────────────────┼────────┼────────┼──────┼────────┤
│ a1b2c3d4e5f6g7h8 │ [active] │        5000000   │     10 │    150 │   2  │   0    │
│ i9j0k1l2m3n4o5p6 │ [inactive]│       1000000   │      5 │      0 │   0  │   0    │
│ q7r8s9t0u1v2w3x4 │ [slashed]│       2500000   │     15 │     75 │   5  │   1    │
└──────────────────┴──────────┴──────────────────┴────────┴────────┴──────┴────────┘
```

**Columns:**
- **Pubkey (16 hex)**: First 8 bytes of validator public key in hex
- **Status**: Validator status
  - `[active]` - Validator is active and producing blocks
  - `[inactive]` - Validator is registered but not active
  - `[deregistered]` - Validator has been deregistered
  - `[slashed]` - Validator was slashed this epoch
- **Total Stake**: Total delegated stake in electrons
- **Comm %**: Commission percentage (0-100)
- **Blocks**: Number of blocks produced this epoch
- **Miss**: Number of missed blocks this epoch
- **Slashed**: Slashing status (0 = not slashed, 1 = slashed)

### 3. Summary Footer

Shows aggregate statistics:

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ Total Staked: 50000000 electrons | Reward Pool: 1000000 electrons | Est. APY: 5.00% │
└──────────────────────────────────────────────────────────────────────────────────┘
```

**Fields:**
- **Total Staked**: Sum of all delegated stakes across all validators
- **Reward Pool**: Current balance of the reward pool
- **Est. APY**: Estimated annual percentage yield (simplified calculation)

### 4. Status Line

Shows last refresh time and exit instructions:

```
Press 'q' or Ctrl+C to exit
Last refresh: 15:04:05
```

## Features

### Real-Time Updates

The dashboard refreshes every 1 second (1000ms) to show the latest state from the FileStore. All data is read-only to prevent accidental modifications.

### Rich Terminal UI

Built with Bubble Tea and lipgloss for:
- Styled components with custom colors
- Smooth rendering and updates
- Responsive keyboard handling
- Alt screen mode (preserves terminal history on exit)

### File Classification

The TUI automatically classifies files in the FileStore by type:

- **Validator Records** (66 bytes) - Validator registration and status
- **Stake Accounts** (89 bytes) - Delegated stake information
- **Epoch State** - Current epoch and validator schedule
- **Reward Pool** - Accumulated rewards for distribution

### Read-Only Access

The TUI opens the FileStore in read-only mode, ensuring no modifications to the blockchain state. This makes it safe to run alongside other blockchain operations.

### Automatic Initialization

If the state directory doesn't exist, the TUI will automatically:
1. Create the directory
2. Initialize a FileStore
3. Load builtin programs (System and Token)
4. Create a minimal DPoS genesis state with one placeholder validator

This allows you to explore the TUI even without a running node.

### Keyboard Controls

- **q** or **Q** - Exit the TUI
- **Ctrl+C** - Exit the TUI (graceful shutdown)

## Usage Examples

### Monitor a Local Node

```bash
# Start a node with state database
go run cmd/main.go --type=leader --port=8000 --db=./state.db

# In another terminal, monitor with TUI
./validator-tui --state ./state.db
```

### Monitor a Remote Node

```bash
# Copy the state database from remote
scp user@remote:/path/to/state.db ./remote-state.db

# Monitor locally
./validator-tui --state ./remote-state.db
```

### Monitor During Demo

```bash
# Start demo in one terminal
./demo-dpos.sh 5 60

# Monitor in another terminal
./validator-tui --state ./state.db
```

## Interpreting the Dashboard

### Validator Status Indicators

- **[active]** - Validator is in the active set and eligible to produce blocks
- **[inactive]** - Validator is registered but not currently active (may be below stake threshold)
- **[deregistered]** - Validator has been deregistered and cannot produce blocks
- **[slashed]** - Validator was slashed this epoch (5% stake reduction)

### Performance Metrics

- **Blocks**: Number of blocks successfully produced by the validator
- **Miss**: Number of slots where the validator was scheduled but didn't produce a block
- **Miss Rate**: Calculated as `Miss / (Blocks + Miss)`

### Stake Information

- **Total Stake**: Sum of all delegated stakes to this validator
- **Comm %**: Commission percentage taken by validator from rewards
- **Est. APY**: Estimated annual return based on current reward rate

## Troubleshooting

### "file not found" Error

The FileStore database doesn't exist at the specified path. Ensure:
1. The path is correct
2. The node has been started and created the database
3. You have read permissions on the database file

### No Validators Displayed

The FileStore may not have any validator records yet. This is normal if:
1. The network is just starting
2. No validators have been registered
3. The genesis initialization hasn't completed

### Stale Data

If the dashboard shows old data:
1. Ensure the node is still running
2. Check that the FileStore is being updated
3. Try restarting the TUI

### Performance Issues

If the TUI is slow or unresponsive:
1. The FileStore may be large; consider archiving old data
2. The system may be under heavy load; try running on a less busy machine
3. Check disk I/O performance

## Architecture

The validator TUI consists of several components:

### Bubble Tea Model

The TUI uses the Bubble Tea framework's Model-View-Update (MVU) architecture:
- **Model**: Holds the current state (validators, epoch info, etc.)
- **Update**: Handles messages (keyboard input, tick events, errors)
- **View**: Renders the UI with lipgloss styling

### ReadOnlyFileStore

A wrapper around BadgerDB that provides read-only access to the FileStore:
- Opens database in read-only mode
- Prevents accidental modifications
- Handles file deserialization

### TUIState

Maintains the current state of the dashboard:
- Stores parsed validator records, stake accounts, epoch state, and reward pool
- Provides query methods for dashboard rendering
- Tracks last refresh time

### File Classification

Automatically determines file types based on:
- Well-known FileID constants (EpochStateFileID, RewardPoolFileID)
- Data size patterns (ValidatorRecord: 66 bytes, StakeAccount: 89 bytes)

### Rendering

Displays the dashboard using lipgloss styles for:
- Title and header panels
- Color-coded validator status (active=green, inactive=red, slashed=bold red)
- Formatted tables with proper alignment
- Footer with summary statistics

## Performance Considerations

- **Refresh Rate**: 1 second (1000ms) - adjustable in code
- **Memory Usage**: Minimal; only stores parsed state in memory
- **Disk I/O**: Read-only access; no writes to FileStore
- **CPU Usage**: Low; simple rendering and parsing

## Future Enhancements

Potential improvements for the validator TUI:

1. **Configurable Refresh Rate** - Allow users to set refresh interval
2. **Filtering and Sorting** - Sort validators by stake, commission, or performance
3. **Historical Data** - Show trends over time
4. **Detailed Views** - Drill down into individual validator or stake account details
5. **Export Functionality** - Export data to CSV or JSON
6. **Network Statistics** - Show network-wide metrics and health
7. **Alerts** - Notify on slashing events or status changes
8. **Multi-Node Monitoring** - Monitor multiple nodes simultaneously

## See Also

- [Delegated Proof of Stake Guide](../reference/dpos-genesis.md)
- [CLI Usage Guide](./cli-usage.md)
- [Demo Guide](./demo.md)

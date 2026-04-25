# Design Document — Delegated Proof of Stake

## Overview

This document describes the technical design for adding Delegated Proof of Stake (DPoS) consensus to the PoH Blockchain. The implementation introduces three major components:

1. **Staking Program** — a native built-in program (registered alongside the System Program) that manages validator registration, stake delegation, reward distribution, and slashing entirely through on-chain file state.
2. **DPoS ConsensusManager** — an extension of the existing `ConsensusManager` that replaces the static `LEADER` node type with a dynamic, stake-weighted validator schedule computed at each epoch boundary.
3. **Validator TUI App** — a standalone terminal dashboard binary (`cmd/validator-tui/main.go`) that reads live state from the FileStore and renders validator metrics.

The design deliberately reuses every existing abstraction: `FileStore`, `Runtime`, `ExecutionContext`, `AccessController`, `TxProcessor`, and `NetworkNode`. No existing interfaces are broken.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Node Process                             │
│                                                                 │
│  ┌──────────────┐    ┌──────────────────────────────────────┐  │
│  │  PoH Clock   │───▶│         DPoS ConsensusManager        │  │
│  └──────────────┘    │  - epoch tracking                    │  │
│                      │  - ValidatorSchedule (slot→pubkey)   │  │
│  ┌──────────────┐    │  - IsLeader(slot) via schedule       │  │
│  │ NetworkNode  │◀───│  - epoch boundary processing         │  │
│  └──────────────┘    └──────────────┬───────────────────────┘  │
│                                     │ reads/writes              │
│  ┌──────────────┐    ┌──────────────▼───────────────────────┐  │
│  │ TxProcessor  │───▶│            FileStore                 │  │
│  └──────┬───────┘    │  ValidatorRecord files               │  │
│         │            │  StakeAccount files                  │  │
│  ┌──────▼───────┐    │  EpochState file                     │  │
│  │   Runtime    │    │  RewardPool file                     │  │
│  │  ┌─────────┐ │    └──────────────────────────────────────┘  │
│  │  │ System  │ │                                               │
│  │  │ Program │ │                                               │
│  │  ├─────────┤ │                                               │
│  │  │ Staking │ │                                               │
│  │  │ Program │ │                                               │
│  │  └─────────┘ │                                               │
│  └──────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   Validator TUI App (separate binary)           │
│   reads FileStore directly (read-only BadgerDB open)            │
│   renders epoch/slot/validator table every 1000ms               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Components and Interfaces

### 1. `internal/staking` — Staking Program

New package. Implements `runtime.BuiltinProgram`.

```go
// StakingProgramID is the well-known ID for the Staking Program (0x...02)
var StakingProgramID = filestore.FileID{..., 0x02}

type StakingProgram struct{}

func (sp *StakingProgram) GetProgramID() filestore.FileID
func (sp *StakingProgram) Execute(ctx *runtime.ExecutionContext) error
```

Instruction dispatch (first byte of `ctx.Instruction.Data`):

| Byte | Instruction          |
|------|----------------------|
| 0    | RegisterValidator    |
| 1    | DeregisterValidator  |
| 2    | DelegateStake        |
| 3    | UndelegateStake      |
| 4    | WithdrawStake        |
| 5    | DistributeRewards    |
| 6    | ReportDoubleSign     |

Each handler is a method on `StakingProgram` that reads/writes exclusively through `ctx.FileStore` and `ctx.AccessController`.

Encoder helpers (mirroring `system.go` pattern):
```go
func EncodeRegisterValidatorInstruction(pubkey transaction.PublicKey, commission uint8) []byte
func EncodeDelegateStakeInstruction(amount int64) []byte
// ... etc
```

### 2. `internal/staking/types.go` — On-Chain Data Structures

All state is stored as JSON in `File.Data`. The `File.Balance` field holds only the storage rent reserve (never the staked amount).

```go
// ValidatorStatus represents the lifecycle state of a validator
type ValidatorStatus uint8
const (
    ValidatorInactive    ValidatorStatus = 0
    ValidatorActive      ValidatorStatus = 1
    ValidatorDeregistered ValidatorStatus = 2
    ValidatorSlashed     ValidatorStatus = 3  // slashed this epoch
)

// ValidatorRecord is stored as File.Data for a Validator Record File
type ValidatorRecord struct {
    PublicKey      transaction.PublicKey `json:"public_key"`
    Commission     uint8                 `json:"commission"`      // 0–100
    TotalStake     int64                 `json:"total_stake"`     // electrons
    Status         ValidatorStatus       `json:"status"`
    BlocksProduced int64                 `json:"blocks_produced"` // current epoch
    BlocksMissed   int64                 `json:"blocks_missed"`   // current epoch
    SlashedEpoch   int64                 `json:"slashed_epoch"`   // -1 if never
}

// StakeStatus represents the lifecycle state of a stake account
type StakeStatus uint8
const (
    StakeActive      StakeStatus = 0
    StakeDeactivating StakeStatus = 1
    StakeWithdrawn   StakeStatus = 2
)

// StakeAccountData is stored as File.Data for a Stake Account File
type StakeAccountData struct {
    DelegatorKey      transaction.PublicKey `json:"delegator_key"`
    ValidatorFileID   filestore.FileID      `json:"validator_file_id"`
    StakedAmount      int64                 `json:"staked_amount"`   // electrons
    ActivationEpoch   int64                 `json:"activation_epoch"`
    DeactivationEpoch int64                 `json:"deactivation_epoch"` // -1 if active
    Status            StakeStatus           `json:"status"`
}

// EpochStateData is stored as File.Data for the Epoch State File
type EpochStateData struct {
    CurrentEpoch      int64              `json:"current_epoch"`
    EpochStartSlot    int64              `json:"epoch_start_slot"`
    SlotsPerEpoch     int64              `json:"slots_per_epoch"`
    Schedule          []ScheduleEntry    `json:"schedule"`        // slot offset → validator pubkey
    LastBlockHash     []byte             `json:"last_block_hash"`
}

// ScheduleEntry maps a slot offset within an epoch to a validator public key
type ScheduleEntry struct {
    SlotOffset int64                 `json:"slot_offset"`
    ValidatorKey transaction.PublicKey `json:"validator_key"`
}
```

### 3. `internal/staking/genesis.go` — Genesis Bootstrap

```go
// GenesisValidator defines a validator to be pre-activated at genesis
type GenesisValidator struct {
    PublicKey  transaction.PublicKey
    Stake      int64  // electrons
    Commission uint8
}

// InitializeGenesisStaking creates the EpochState, RewardPool, and
// ValidatorRecord files for all genesis validators, then computes
// the epoch-0 Validator Schedule. Idempotent if EpochState already exists.
func InitializeGenesisStaking(fs *filestore.FileStore, validators []GenesisValidator) error
```

Well-known file IDs for singleton state files:

```go
var EpochStateFileID = filestore.FileID{..., 0xE0}  // 0x...E0
var RewardPoolFileID = filestore.FileID{..., 0xE1}  // 0x...E1
```

### 4. `internal/consensus` — DPoS ConsensusManager Extension

The existing `ConsensusManager` is extended (not replaced) with:

```go
type ConsensusManager struct {
    // existing fields
    nodeType         network.NodeType
    slotDurationMs   int64
    genesisTimestamp time.Time

    // new fields
    localPubKey      transaction.PublicKey  // this node's identity
    fileStore        *filestore.FileStore   // for reading schedule & writing missed blocks
    epochState       *staking.EpochStateData // cached in memory, refreshed at epoch boundary
    slotsPerEpoch    int64
}

// IsLeader now consults the Validator Schedule instead of node type
func (cm *ConsensusManager) IsLeader(slot int64) bool

// ProcessEpochBoundary triggers schedule recomputation and reward distribution
func (cm *ConsensusManager) ProcessEpochBoundary(lastBlockHash []byte) error

// GetScheduledValidator returns the public key of the validator for a given slot
func (cm *ConsensusManager) GetScheduledValidator(slot int64) (transaction.PublicKey, error)

// RecordMissedBlock increments the missed-block counter for the scheduled validator
func (cm *ConsensusManager) RecordMissedBlock(slot int64) error
```

### 5. `internal/staking/scheduler.go` — Validator Schedule Computation

```go
// ComputeSchedule builds a deterministic weighted-random slot schedule for an epoch.
// seed is the last block hash of the previous epoch.
// validators is the list of active ValidatorRecords with their stakes.
// slotsPerEpoch is the number of slots to assign.
func ComputeSchedule(
    seed []byte,
    validators []ValidatorWithID,
    slotsPerEpoch int64,
) []staking.ScheduleEntry
```

Algorithm: weighted reservoir sampling using a deterministic PRNG seeded with `SHA-256(seed)`. Each validator's probability of being assigned a slot is proportional to `validator.TotalStake / totalStake`.

### 6. `internal/staking/rewards.go` — Reward Distribution

```go
// DistributeEpochRewards reads the RewardPool balance, calculates each validator's
// share based on blocks produced, deducts commission, and distributes the remainder
// to all active Stake Accounts proportionally. Fractional remainders stay in the pool.
func DistributeEpochRewards(fs *filestore.FileStore, epochState *EpochStateData) error
```

### 7. `cmd/validator-tui/main.go` — Validator TUI App

Standalone binary. Uses only the standard library plus a minimal terminal rendering approach (ANSI escape codes, no external TUI framework, keeping with the project's no-external-frameworks rule for blockchain logic — the TUI is UI-only so a lightweight dependency is acceptable if needed, but the design targets stdlib + ANSI).

```
┌─ PoH Blockchain Validator Dashboard ──────────────────────────┐
│ Epoch: 42   Slot: 18,123,456   Local: active   Validators: 7  │
│ Local Stake: 5,000,000 electrons (5.00 Neon)                  │
├───────────────────────────────────────────────────────────────┤
│ Validator          │ Status   │ Stake (e)  │ Comm │ Blocks    │
├────────────────────┼──────────┼────────────┼──────┼───────────┤
│ 0xabcd1234...      │ active   │ 5,000,000  │  5%  │ 1,204     │
│ [inactive] 0xef... │ inactive │   800,000  │ 10%  │     0     │
│ [slashed]  0x12... │ slashed  │ 1,900,000  │  8%  │   312     │
├───────────────────────────────────────────────────────────────┤
│ Total Staked: 12,400,000 e  Pool: 48,200 e  Est. APY: 6.2%   │
│ [q] quit                                                      │
└───────────────────────────────────────────────────────────────┘
```

Refresh loop: `time.Ticker` at 1000ms. Reads FileStore in read-only mode (BadgerDB supports concurrent readers). Exits cleanly on `q` / `Ctrl+C` via `os.Signal` channel.

### 8. `demo-dpos.sh` — Automated Demo Script

Bash script at repo root. Starts N validator nodes from a genesis config, exercises the full DPoS lifecycle, and emits structured JSON logs compatible with `analyze-results.sh`.

---

## Data Models

### File Layout in FileStore

| File Purpose        | FileID                  | Executable | TxManager         | Balance (electrons)         | Data payload              |
|---------------------|-------------------------|------------|-------------------|-----------------------------|---------------------------|
| Staking Program     | `StakingProgramID`      | true       | `RuntimeProgramID`| `CalculateStorageCost(0)`   | empty (builtin)           |
| Epoch State         | `EpochStateFileID`      | false      | `StakingProgramID`| `CalculateStorageCost(data)`| `EpochStateData` JSON     |
| Reward Pool         | `RewardPoolFileID`      | false      | `StakingProgramID`| pool balance + storage cost | `RewardPoolData` JSON     |
| Validator Record    | `SHA256(pubkey+"vr")`   | false      | `StakingProgramID`| `CalculateStorageCost(data)`| `ValidatorRecord` JSON    |
| Stake Account       | `SHA256(delegator+validator+nonce)` | false | `StakingProgramID` | `CalculateStorageCost(data)` | `StakeAccountData` JSON |

**Key design decision**: `File.Balance` on Validator Records and Stake Accounts holds only the storage rent reserve. The actual staked electrons are stored exclusively in the JSON data payload (`ValidatorRecord.TotalStake` and `StakeAccountData.StakedAmount`). This satisfies Requirement 10 — the FileStore's `ValidateStorageCost` check never touches staked funds.

The Reward Pool is the exception: its `File.Balance` holds the actual pool balance, because the pool is not a stake account and has no staked amount to protect.

### Electron Denomination

```go
const (
    ElectronsPerNeon          = int64(1_000_000)
    MinActivationStake        = int64(1_000_000)  // 1 Neon
    MinRentExemptBalance      = int64(1_000)       // 1,000 electrons
    SlashingPenaltyBasisPoints = int64(500)        // 5%
    CooldownEpochs            = int64(1)
    DefaultSlotsPerEpoch      = int64(432_000)
    DefaultSlotTimeoutMs      = int64(200)
)
```

---

## Error Handling

All Staking Program handlers return typed errors using `fmt.Errorf` with `%w` wrapping, consistent with the rest of the codebase. The `TxProcessor` rolls back all state on any instruction error (existing behavior, unchanged).

Specific error conditions:

| Condition                              | Error message pattern                                      |
|----------------------------------------|------------------------------------------------------------|
| Commission out of range                | `"commission rate %d out of range [0,100]"`               |
| Validator already registered           | `"validator %s already registered"`                       |
| Validator not found                    | `"validator record not found: %s"`                        |
| Insufficient balance for delegation    | `"insufficient balance: need %d electrons, have %d"`      |
| Stake account in cooldown              | `"stake in cooldown until epoch %d, current epoch %d"`    |
| Rent reserve violation                 | `"balance %d below storage cost %d for file %s"`          |
| Double-sign signature invalid          | `"double-sign report: signature %d invalid"`              |
| Genesis with zero validators           | `"genesis requires at least one validator"`               |
| Unknown instruction type               | `"unknown staking instruction type: %d"`                  |

---

## Testing Strategy

- **Unit tests** (`internal/staking/*_test.go`): test each instruction handler in isolation using an in-memory FileStore (temp BadgerDB dir). Cover happy path, all error conditions, and the storage-cost/stake separation invariant.
- **Unit tests** (`internal/staking/scheduler_test.go`): verify determinism (same seed → same schedule), proportionality (higher stake → more slots), and edge cases (single validator, equal stakes).
- **Integration tests** (`internal/staking_integration_test.go`): full lifecycle — genesis init → delegate → epoch boundary → rewards → undelegate → withdraw → slash.
- **TUI smoke test**: not automated (terminal rendering); verified manually.
- **Demo script** (`demo-dpos.sh`): serves as the end-to-end automated test, producing machine-parseable JSON output.

---

## Design Decisions and Rationale

1. **Staked amount in data payload, not `File.Balance`**: The FileStore's `ValidateStorageCost` runs on every `UpdateFile` call and rejects files whose `Balance` is below storage cost. If staked electrons lived in `Balance`, any storage cost increase (e.g., from adding data) could silently invalidate a valid stake. Separating the two makes the invariant explicit and testable.

2. **No new external dependencies for the Staking Program**: The staking logic is pure Go using `encoding/json`, `crypto/sha256`, and `math/rand` (seeded deterministically). This matches the project's "no external frameworks for blockchain logic" rule.

3. **Extending `ConsensusManager` rather than replacing it**: The existing slot timing, `WaitForSlotStart`, and `ValidateBlock` logic is correct and reusable. Only `IsLeader` and the epoch boundary hook need to change.

4. **Deterministic schedule via seeded PRNG**: Using `SHA-256(lastBlockHash)` as the PRNG seed ensures all nodes independently compute the same schedule without coordination, which is essential for a leaderless consensus.

5. **`EpochStateFileID` and `RewardPoolFileID` as well-known constants**: Singleton files with fixed IDs avoid the need for a registry or lookup table, consistent with how `SystemProgramID` and `StakingProgramID` work.

6. **TUI reads FileStore directly in read-only mode**: BadgerDB supports multiple concurrent readers. This avoids adding an RPC layer just for the TUI and keeps the binary self-contained.

7. **`demo-dpos.sh` emits JSON compatible with `analyze-results.sh`**: Reusing the existing analysis tooling reduces maintenance burden and lets operators compare DPoS demo results with existing BFT demo results side by side.

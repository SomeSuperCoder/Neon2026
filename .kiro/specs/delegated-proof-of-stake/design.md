# Design Document — Delegated Proof of Stake

## Overview

This document describes the technical design for adding Delegated Proof of Stake (DPoS) consensus to the PoH Blockchain. The implementation introduces three major components:

1. **Staking Program** — a QuanticScript smart contract (`programs/staking/staking.qs`) compiled to bytecode and loaded into the FileStore at genesis, following the exact same pattern as the System and Token programs. All staking logic executes through the existing bytecode interpreter.
2. **DPoS ConsensusManager** — an extension of the existing `ConsensusManager` that replaces the static `LEADER` node type with a dynamic, stake-weighted validator schedule computed at each epoch boundary.
3. **Validator TUI App** — a standalone terminal dashboard binary (`cmd/validator-tui/main.go`) that reads live state from the FileStore and renders validator metrics.

The design deliberately reuses every existing abstraction: `FileStore`, `Runtime`, `ExecutionContext`, `AccessController`, `TxProcessor`, `NetworkNode`, and the QuanticScript compiler pipeline. No existing interfaces are broken.

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

### 1. `programs/staking/staking.qs` — Staking Program (QuanticScript)

The Staking Program is a QuanticScript source file compiled to bytecode, identical in structure to `programs/token/token.qs`. It is **not** a Go `BuiltinProgram` — it executes through the existing `BytecodeInterpreter`.

```
programs/staking/
├── staking.qs    ← source (authored)
├── staking.qsa   ← assembly (compiler output)
└── staking.qsb   ← bytecode (compiler output, loaded at genesis)
```

**Entry point and dispatch** (mirrors `system.qs` / `token.qs` pattern):

```qs
// Instruction type constants
const REGISTER_VALIDATOR:   i64 = 0;
const DEREGISTER_VALIDATOR: i64 = 1;
const DELEGATE_STAKE:       i64 = 2;
const UNDELEGATE_STAKE:     i64 = 3;
const WITHDRAW_STAKE:       i64 = 4;
const DISTRIBUTE_REWARDS:   i64 = 5;
const REPORT_DOUBLE_SIGN:   i64 = 6;

const ERROR_INVALID_INSTRUCTION: i64 = 0x3FFF;

export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    if (len(instrData) < 1) { return ERROR_INVALID_INSTRUCTION; }
    let instrType: i64 = instrData[0];
    if (instrType == REGISTER_VALIDATOR)   { return handleRegisterValidator(instrData); }
    // ... etc
    return ERROR_INVALID_INSTRUCTION;
}
```

**On-chain state encoding**: All structured data (ValidatorRecord, StakeAccountData, EpochStateData) is encoded as packed binary in `File.Data` using the same little-endian `i64` / byte-array conventions used by the Token Program. No JSON — the QuanticScript interpreter works with `bytes`, not structured objects.

**Builtins used** (all already available from Token Program):
- `getInstructionData()`, `getFile()`, `getBalance()`, `hasSigner()`
- `updateFile()`, `createFile()`, `transferBalance()`, `deleteFile()`
- `len()`, `slice()`, `hashBytes()`

### 2. `internal/staking/types.go` — Go-Side Data Structures (for ConsensusManager and TUI)

The Go packages (`internal/staking`, `internal/consensus`, `cmd/validator-tui`) need to read and write the same binary state that the QuanticScript program produces. This package defines the Go structs and binary encode/decode functions that mirror the QuanticScript binary layout.

```go
type ValidatorStatus uint8
const (
    ValidatorInactive      ValidatorStatus = 0
    ValidatorActive        ValidatorStatus = 1
    ValidatorDeregistered  ValidatorStatus = 2
    ValidatorSlashed       ValidatorStatus = 3
)

type ValidatorRecord struct {
    PublicKey      [32]byte
    Commission     uint8
    TotalStake     int64
    Status         ValidatorStatus
    BlocksProduced int64
    BlocksMissed   int64
    SlashedEpoch   int64  // -1 if never slashed
}

type StakeStatus uint8
const (
    StakeActive       StakeStatus = 0
    StakeDeactivating StakeStatus = 1
)

type StakeAccountData struct {
    DelegatorKey      [32]byte
    ValidatorFileID   filestore.FileID
    StakedAmount      int64
    ActivationEpoch   int64
    DeactivationEpoch int64  // -1 if active
    Status            StakeStatus
}

type EpochStateData struct {
    CurrentEpoch   int64
    EpochStartSlot int64
    SlotsPerEpoch  int64
    LastBlockHash  [32]byte
    ScheduleLen    int64
    // Schedule entries follow: each is [32]byte pubkey
}

type ScheduleEntry struct {
    SlotOffset   int64
    ValidatorKey [32]byte
}
```

### 3. `internal/staking/genesis.go` — Genesis Bootstrap

```go
type GenesisValidator struct {
    PublicKey  transaction.PublicKey
    Stake      int64
    Commission uint8
}

// InitializeGenesisStaking creates EpochState, RewardPool, and ValidatorRecord
// files for all genesis validators, then computes the epoch-0 schedule.
// Idempotent if EpochState already exists.
func InitializeGenesisStaking(fs *filestore.FileStore, validators []GenesisValidator) error
```

Well-known singleton file IDs:

```go
var StakingProgramID = filestore.FileID{..., 0x03}  // 0x...03
var EpochStateFileID = filestore.FileID{..., 0xE0}  // 0x...E0
var RewardPoolFileID = filestore.FileID{..., 0xE1}  // 0x...E1
```

### 4. `internal/consensus` — DPoS ConsensusManager Extension

The existing `ConsensusManager` is extended (not replaced):

```go
type ConsensusManager struct {
    // existing fields unchanged
    nodeType         network.NodeType
    slotDurationMs   int64
    genesisTimestamp time.Time

    // new DPoS fields
    localPubKey   transaction.PublicKey
    fileStore     *filestore.FileStore
    epochState    *staking.EpochStateData
    schedule      []staking.ScheduleEntry  // decoded from EpochStateData
    slotsPerEpoch int64
}

func (cm *ConsensusManager) IsLeader(slot int64) bool
func (cm *ConsensusManager) GetScheduledValidator(slot int64) (transaction.PublicKey, error)
func (cm *ConsensusManager) RecordMissedBlock(slot int64) error
func (cm *ConsensusManager) CheckAndProcessEpochBoundary(slot int64, lastBlockHash []byte) error
```

### 5. `internal/staking/scheduler.go` — Schedule Computation

```go
func ComputeSchedule(seed []byte, validators []ValidatorWithID, slotsPerEpoch int64) []ScheduleEntry
```

Weighted reservoir sampling, PRNG seeded with `SHA-256(seed)`.

### 6. `internal/staking/rewards.go` — Reward Distribution

```go
func DistributeEpochRewards(fs *filestore.FileStore, epochState *EpochStateData) error
func ProcessEpochBoundary(fs *filestore.FileStore, lastBlockHash []byte) error
```

### 7. `cmd/validator-tui/main.go` — Validator TUI App

Standalone binary. Reads FileStore in read-only mode. Renders ANSI dashboard at 1000ms intervals.

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

### 8. `demo-dpos.sh` — Automated Demo Script

Bash script at repo root. Starts N validator nodes, exercises the full DPoS lifecycle, emits structured JSON logs compatible with `analyze-results.sh`.

---

## Data Models

### File Layout in FileStore

| File Purpose        | FileID                  | Executable | TxManager         | Balance (electrons)         | Data payload              |
|---------------------|-------------------------|------------|-------------------|-----------------------------|---------------------------|
| Staking Program     | `StakingProgramID`      | true       | `RuntimeProgramID`| `CalculateStorageCost(qsb)` | `staking.qsb` bytecode    |
| Epoch State         | `EpochStateFileID`      | false      | `StakingProgramID`| `CalculateStorageCost(data)`| packed binary (see types) |
| Reward Pool         | `RewardPoolFileID`      | false      | `StakingProgramID`| pool balance + storage cost | packed binary             |
| Validator Record    | `SHA256(pubkey+"vr")`   | false      | `StakingProgramID`| `CalculateStorageCost(data)`| packed binary             |
| Stake Account       | `SHA256(delegator+validator+nonce)` | false | `StakingProgramID` | `CalculateStorageCost(data)` | packed binary |

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

1. **Staking Program as QuanticScript bytecode, not a Go BuiltinProgram**: Writing the Staking Program in QuanticScript means it executes through the same `BytecodeInterpreter` as all user programs, gets the same access control enforcement, and can be audited and upgraded without recompiling the node binary. It also validates the QuanticScript runtime against a complex real-world program.

2. **Packed binary encoding in `File.Data`, not JSON**: The QuanticScript interpreter works with `bytes` and integer operations, not structured objects. Packed little-endian binary (same convention as the Token Program) is the natural format. The Go-side `internal/staking/types.go` provides matching encode/decode functions so the ConsensusManager and TUI can read the same state.

3. **Staked amount in data payload, not `File.Balance`**: The FileStore's `ValidateStorageCost` runs on every `UpdateFile` call and rejects files whose `Balance` is below storage cost. If staked electrons lived in `Balance`, any storage cost increase could silently invalidate a valid stake. Separating the two makes the invariant explicit and testable.

4. **`StakingProgramID = 0x...03`**: `0x...01` is SystemProgram, `0x...02` is TokenProgram. The Staking Program takes the next sequential well-known ID.

5. **Extending `ConsensusManager` rather than replacing it**: The existing slot timing, `WaitForSlotStart`, and `ValidateBlock` logic is correct and reusable. Only `IsLeader` and the epoch boundary hook need to change.

6. **Deterministic schedule via seeded PRNG**: Using `SHA-256(lastBlockHash)` as the PRNG seed ensures all nodes independently compute the same schedule without coordination.

7. **`EpochStateFileID` and `RewardPoolFileID` as well-known constants**: Singleton files with fixed IDs avoid the need for a registry or lookup table, consistent with how `SystemProgramID` and `StakingProgramID` work.

8. **TUI reads FileStore directly in read-only mode**: BadgerDB supports multiple concurrent readers. This avoids adding an RPC layer just for the TUI and keeps the binary self-contained.

9. **`demo-dpos.sh` emits JSON compatible with `analyze-results.sh`**: Reusing the existing analysis tooling reduces maintenance burden and lets operators compare DPoS demo results with existing BFT demo results side by side.

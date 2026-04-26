# DPoS Genesis Initialization

## Overview

The DPoS (Delegated Proof of Stake) genesis initialization system bootstraps the blockchain with validator records, epoch state, and reward pool files. This is handled by the `InitializeDPoSGenesis` function in `internal/genesis/programs.go`.

## Well-Known FileIDs

The DPoS system uses reserved FileIDs for critical state files:

| Component | FileID | Purpose |
|-----------|--------|---------|
| Staking_Program | `0x0000...0003` | Executable program managing DPoS operations |
| Epoch State | `0x0000...0004` | Current epoch number, slot, validator schedule, missed blocks |
| Reward Pool | `0x0000...0005` | Accumulated rewards for distribution |
| Validator Records | Derived from pubkey | Individual validator state (stake, commission, status) |

## Genesis Configuration

The `GenesisConfig` struct defines the initial DPoS state:

```go
type GenesisConfig struct {
    EpochLength       int64              // Slots per epoch (e.g., 432,000)
    GenesisValidators []GenesisValidator // Initial validator set
}

type GenesisValidator struct {
    PublicKey   [32]byte // Validator's Ed25519 public key
    StakeAmount int64    // Initial stake in electrons
}
```

## Initialization Process

### 1. Validation
- Rejects genesis config with zero validators (Requirement 9.6)
- Idempotent: skips if Epoch State File already exists

### 2. Validator Record Creation
For each genesis validator:
- Generates deterministic FileID from public key: `GenerateFileID(append([]byte("validator:"), pubkey[:]))`
- Serializes validator record with:
  - Public key (32 bytes)
  - Commission (i64, little-endian)
  - Total stake (i64, little-endian)
  - Status: 1 (active)
  - Blocks produced: 0
  - Missed blocks: 0
  - Slashed this epoch: 0
- Creates File with TxManager = StakingProgramID
- Allocates balance to cover storage cost + 1000 electron buffer

### 3. Epoch State File Creation
- Serializes epoch state with:
  - Epoch number: 0
  - Epoch start slot: 0
  - Total slots in epoch: from config
  - Validator schedule: array of validator FileIDs
  - Missed block counters: initialized to 0 for each validator
- Creates File at EpochStateFileID with TxManager = StakingProgramID
- Allocates balance to cover storage cost + 1000 electron buffer

### 4. Reward Pool File Creation
- Serializes reward pool with:
  - Balance: 0 (no initial rewards)
  - Last distributed epoch: 0
- Creates File at RewardPoolFileID with TxManager = StakingProgramID
- Allocates balance to cover storage cost + 1000 electron buffer

## Data Serialization

### Validator Record Format (66 bytes)
```
Offset | Size | Field
-------|------|-------
0      | 32   | Public key
32     | 8    | Commission (i64 LE)
40     | 8    | Total stake (i64 LE)
48     | 1    | Status (0=inactive, 1=active, 2=deregistered)
49     | 8    | Blocks produced (i64 LE)
57     | 8    | Missed blocks (i64 LE)
65     | 1    | Slashed this epoch (0=false, 1=true)
```

### Epoch State Format (variable)
```
Offset | Size | Field
-------|------|-------
0      | 8    | Epoch number (i64 LE)
8      | 8    | Epoch start slot (i64 LE)
16     | 8    | Total slots in epoch (i64 LE)
24     | 8    | Validator count (i64 LE)
32     | 32*N | Validator schedule (N FileIDs)
32+32*N| 8*N  | Missed block counters (N i64 LE values)
```

### Reward Pool Format (16 bytes)
```
Offset | Size | Field
-------|------|-------
0      | 8    | Balance (i64 LE)
8      | 8    | Last distributed epoch (i64 LE)
```

## Testing

The `internal/genesis/programs_test.go` file provides comprehensive test coverage:

### Test Functions

1. **TestInitializeDPoSGenesis** (Requirements 9.1-9.6)
   - Creates 3 genesis validators
   - Verifies Epoch State File structure and data
   - Verifies Reward Pool File structure and data
   - Verifies Validator Record Files for each validator
   - Checks file balances cover storage costs
   - Validates TxManager assignments

2. **TestInitializeDPoSGenesisZeroValidators** (Requirement 9.6)
   - Verifies error when genesis config has zero validators

3. **TestInitializeDPoSGenesisIdempotent**
   - Calls InitializeDPoSGenesis twice
   - Verifies second call skips creation (idempotent)
   - Confirms Epoch State File unchanged

4. **TestLoadBuiltinProgramsWithStaking** (Requirement 6.5)
   - Loads System, Token, and Staking programs
   - Verifies all three programs are created and executable

5. **TestLoadBuiltinProgramsWithoutStaking** (Requirement 6.5)
   - Loads only System and Token programs (nil staking bytecode)
   - Verifies Staking Program is NOT created

6. **TestEpochStateFileStructure** (Requirements 9.2, 10.2)
   - Verifies Epoch State File properties (ID, TxManager, not executable)
   - Deserializes and validates all fields
   - Checks missed block counters initialized to 0

7. **TestRewardPoolFileStructure** (Requirements 9.3, 10.3)
   - Verifies Reward Pool File properties (ID, TxManager, not executable)
   - Deserializes and validates balance and epoch fields

### Running Tests

```bash
# Run all genesis tests
go test -v ./internal/genesis

# Run specific test
go test -v ./internal/genesis -run TestInitializeDPoSGenesis

# Run with coverage
go test -cover ./internal/genesis
```

## Integration with Node Startup

The genesis initialization is called during node startup:

```go
// In cmd/main.go or consensus manager initialization
config := GenesisConfig{
    EpochLength: 432000,
    GenesisValidators: []GenesisValidator{
        {PublicKey: validator1Pubkey, StakeAmount: 1000000},
        {PublicKey: validator2Pubkey, StakeAmount: 2000000},
        // ...
    },
}

err := genesis.InitializeDPoSGenesis(fileStore, config)
if err != nil {
    log.Fatalf("Failed to initialize DPoS genesis: %v", err)
}
```

## Storage Costs

Each file's balance is calculated as:
```
balance = CalculateStorageCost(dataLength) + 1000
```

Where `CalculateStorageCost` is defined in `internal/filestore/filestore.go` and implements the exponential storage cost model.

## Error Handling

- **Zero validators**: Returns error "genesis config must have at least one validator"
- **Serialization failures**: Returns error with context (e.g., "failed to serialize validator record 0")
- **File creation failures**: Returns error with context (e.g., "failed to create epoch state file")

## Requirements Coverage

| Requirement | Test | Status |
|-------------|------|--------|
| 9.1 | TestInitializeDPoSGenesis | ✓ |
| 9.2 | TestEpochStateFileStructure | ✓ |
| 9.3 | TestRewardPoolFileStructure | ✓ |
| 9.4 | TestInitializeDPoSGenesis | ✓ |
| 9.6 | TestInitializeDPoSGenesisZeroValidators | ✓ |
| 10.2 | TestEpochStateFileStructure | ✓ |
| 10.3 | TestRewardPoolFileStructure | ✓ |
| 6.5 | TestLoadBuiltinProgramsWithStaking | ✓ |

## Next Steps

1. Implement ConsensusManager epoch processing (Task 4)
2. Implement validator schedule computation (Task 4.3)
3. Implement reward distribution (Task 4.5)
4. Implement slashing logic (Task 4.5)
5. Build validator TUI app (Task 5)
6. Create DPoS demo script (Task 6)

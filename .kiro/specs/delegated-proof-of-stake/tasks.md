# Implementation Plan

- [ ] 1. Define staking constants, token denomination, and on-chain data types
  - Create `internal/staking/types.go` with `ValidatorRecord`, `StakeAccountData`, `EpochStateData`, `ScheduleEntry`, `ValidatorStatus`, `StakeStatus`, and all electron/epoch constants (`ElectronsPerNeon`, `MinActivationStake`, `MinRentExemptBalance`, `SlashingPenaltyBasisPoints`, `CooldownEpochs`, `DefaultSlotsPerEpoch`, `DefaultSlotTimeoutMs`)
  - Add JSON marshal/unmarshal helpers for each data type
  - Define well-known `EpochStateFileID` and `RewardPoolFileID` constants
  - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 10.1, 10.2_

- [ ] 2. Implement the Staking Program core and instruction dispatch
  - Create `internal/staking/staking.go` with `StakingProgram` struct, `StakingProgramID` constant, `GetProgramID()`, and `Execute()` dispatch on the first instruction byte
  - Implement `RegisterValidator` handler: validate commission 0–100, check no duplicate, create Validator Record File with rent-only `Balance`, status `inactive`, zero stake
  - Implement `DeregisterValidator` handler: verify signer, check zero delegated stake, delete Validator Record File
  - Add all `Encode*Instruction` helper functions
  - _Requirements: 1.1, 1.2, 1.3, 1.5, 6.1, 6.2, 6.3, 6.4, 10.2, 10.3_

- [ ] 3. Implement stake delegation and withdrawal instructions
  - Implement `DelegateStake` handler: validate positive amount, check delegator balance ≥ amount + `MinRentExemptBalance`, verify validator is not deregistered, create Stake Account File with staked amount in data payload and rent-only `Balance`, increment `ValidatorRecord.TotalStake`
  - Implement `UndelegateStake` handler: verify delegator ownership, set stake status to `deactivating`, record deactivation epoch
  - Implement `WithdrawStake` handler: verify cooldown elapsed, transfer staked amount back to delegator account, delete Stake Account File, decrement `ValidatorRecord.TotalStake`
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 10.1, 10.2, 10.4, 10.5_

- [ ] 4. Implement the validator schedule computation
  - Create `internal/staking/scheduler.go` with `ComputeSchedule(seed []byte, validators []ValidatorWithID, slotsPerEpoch int64) []ScheduleEntry`
  - Use `SHA-256(seed)` as the deterministic PRNG seed for weighted reservoir sampling
  - Assign slots proportionally to `validator.TotalStake / totalStake`
  - _Requirements: 3.3, 9.4_

- [ ] 5. Implement genesis bootstrap
  - Create `internal/staking/genesis.go` with `InitializeGenesisStaking(fs *filestore.FileStore, validators []GenesisValidator) error`
  - Create Validator Record Files for all genesis validators with status `active` and pre-assigned stake, bypassing the normal activation threshold
  - Create the Epoch State File (epoch 0, slot 0) and Reward Pool File (zero balance)
  - Compute and persist the epoch-0 Validator Schedule
  - Make the function idempotent (skip if Epoch State File already exists)
  - Validate that at least one genesis validator is provided
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [ ] 6. Implement epoch boundary processing and reward distribution
  - Create `internal/staking/rewards.go` with `DistributeEpochRewards(fs *filestore.FileStore, epochState *EpochStateData) error`
  - Calculate each active validator's reward share proportional to blocks produced
  - Deduct commission, distribute remainder to Stake Accounts proportionally
  - Retain fractional electron remainders in the Reward Pool
  - Implement `ProcessEpochBoundary(fs *filestore.FileStore, lastBlockHash []byte) error` that activates/deactivates validators based on stake threshold, calls `DistributeEpochRewards`, computes the new schedule, and updates the Epoch State File
  - _Requirements: 3.1, 3.2, 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.4_

- [ ] 7. Implement slashing
  - Implement `ReportDoubleSign` handler in `staking.go`: decode two block headers from instruction data, verify both Ed25519 signatures against the same validator public key and same slot, apply 5% slash to `ValidatorRecord.TotalStake`, transfer slashed amount to Reward Pool, deactivate validator if stake falls below threshold
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 8. Extend ConsensusManager with DPoS schedule and epoch tracking
  - Add `localPubKey transaction.PublicKey`, `fileStore *filestore.FileStore`, `epochState *staking.EpochStateData`, and `slotsPerEpoch int64` fields to `ConsensusManager`
  - Update `NewConsensusManager` to accept these new parameters
  - Rewrite `IsLeader(slot int64) bool` to look up the slot in the cached `epochState.Schedule`
  - Add `GetScheduledValidator(slot int64) (transaction.PublicKey, error)`
  - Add `RecordMissedBlock(slot int64) error` that increments the validator's `BlocksMissed` counter in the FileStore
  - Add `CheckAndProcessEpochBoundary(slot int64, lastBlockHash []byte) error` that calls `staking.ProcessEpochBoundary` when `slot % slotsPerEpoch == 0` and refreshes the in-memory `epochState`
  - _Requirements: 3.3, 3.4, 3.5, 3.6, 8.2, 8.3, 8.4_

- [ ] 9. Register the Staking Program in genesis and the Runtime
  - Update `internal/genesis/programs.go`: add `StakingProgramID` constant (`0x...02`), add `LoadStakingProgram(fs *filestore.FileStore) error` that creates the Staking Program File (`Executable=true`, `TxManager=RuntimeProgramID`, empty data, rent-only balance)
  - Update node startup (in `cmd/main.go` or wherever `Runtime` is initialized) to call `runtime.RegisterBuiltinProgram(staking.NewStakingProgram())` before processing any transactions
  - _Requirements: 6.1, 6.2, 6.5_

- [ ] 10. Build the Validator TUI application
  - Create `cmd/validator-tui/main.go` as a standalone binary
  - Accept `--state <path>` flag for the FileStore path and `--pubkey <hex>` for the local validator identity
  - Open the FileStore in read-only mode, read `EpochStateData`, `RewardPoolData`, and all Validator Record Files
  - Render the dashboard using ANSI escape codes: header panel (epoch, slot, local status, validator count, local stake), validator table (pubkey truncated to 16 chars, status with `[inactive]`/`[slashed]` prefix, stake, commission, blocks produced), summary panel (total staked, pool balance, estimated APY)
  - Refresh every 1000ms via `time.Ticker`
  - Handle `q` keypress and `os.Interrupt` signal for clean exit, closing the FileStore
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

- [ ] 11. Write the DPoS automated demo script
  - Create `demo-dpos.sh` at the repo root
  - Accept `<num_validators> <duration_seconds>` arguments, validate minimum 2 validators
  - Generate a genesis config with N validators, start N node processes, wait for epoch 0 blocks
  - Submit `DelegateStake` transactions, wait for epoch boundary, log the new schedule
  - Wait for reward distribution, log validator and delegator reward amounts in electrons
  - Submit a `ReportDoubleSign` instruction against validator 0, log the slashing event
  - Emit structured JSON log entries to `logs/dpos-demo-<timestamp>.json` in the same format as `demo-automated.sh`
  - Print a human-readable summary table to stdout; exit 0 on success, non-zero on failure
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8_

- [ ] 12. Write unit tests for staking types and instruction handlers
  - Test `ValidatorRecord` and `StakeAccountData` JSON round-trip serialization
  - Test each instruction handler: happy path, all error conditions, storage-cost/stake separation invariant
  - Test `ComputeSchedule` determinism and proportionality
  - _Requirements: 1.1–1.5, 2.1–2.6, 5.1–5.4, 10.1–10.5_

- [ ] 13. Write integration test for the full DPoS lifecycle
  - Create `internal/staking_integration_test.go`
  - Cover: genesis init → register validator → delegate → epoch boundary → rewards → undelegate → cooldown → withdraw → slash → deactivation
  - _Requirements: 3.1, 3.2, 4.1–4.5, 5.1–5.4, 8.1–8.4, 9.1–9.6_

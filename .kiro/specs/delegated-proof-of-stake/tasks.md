# Implementation Plan

- [ ] 1. Define Go-side staking types and binary encoding
  - Create `internal/staking/types.go` with `ValidatorRecord`, `StakeAccountData`, `EpochStateData`, `ScheduleEntry`, `ValidatorStatus`, `StakeStatus`, and all electron/epoch constants (`ElectronsPerNeon`, `MinActivationStake`, `MinRentExemptBalance`, `SlashingPenaltyBasisPoints`, `CooldownEpochs`, `DefaultSlotsPerEpoch`, `DefaultSlotTimeoutMs`)
  - Implement packed little-endian binary `Encode` / `Decode` functions for each type (matching the format the QuanticScript program will write)
  - Define well-known file ID constants: `StakingProgramID` (`0x...03`), `EpochStateFileID` (`0x...E0`), `RewardPoolFileID` (`0x...E1`)
  - _Requirements: 1.1, 2.1, 3.1, 4.1, 5.1, 10.1, 10.2_

- [ ] 2. Write the Staking Program in QuanticScript — validator registration and deregistration
  - Create `programs/staking/staking.qs` with the `entry()` export function, instruction type constants (`REGISTER_VALIDATOR=0`, `DEREGISTER_VALIDATOR=1`, …, `REPORT_DOUBLE_SIGN=6`), and error code constants (base `0x3000`)
  - Implement `handleRegisterValidator`: parse commission byte, validate 0–100, derive Validator Record FileID via `hashBytes`, check no duplicate, call `createFile` with rent-only balance and packed ValidatorRecord data (status=inactive, zero stake)
  - Implement `handleDeregisterValidator`: verify signer, decode ValidatorRecord, check zero total stake, call `deleteFile`
  - Follow the same helper-function patterns (`parseI64LE`, `slice`, `append`) established in `system.qs` and `token.qs`
  - _Requirements: 1.1, 1.2, 1.3, 1.5, 6.1, 6.3, 6.4, 10.2, 10.3_

- [ ] 3. Write the Staking Program in QuanticScript — stake delegation and withdrawal
  - Implement `handleDelegateStake`: validate positive amount, check delegator balance ≥ amount + `MIN_RENT_EXEMPT`, verify validator not deregistered, call `createFile` for Stake Account with staked amount in data payload and rent-only balance, update ValidatorRecord TotalStake via `updateFile`
  - Implement `handleUndelegateStake`: verify delegator ownership via `hasSigner`, decode StakeAccountData, set status to deactivating, record deactivation epoch, call `updateFile`
  - Implement `handleWithdrawStake`: verify cooldown elapsed (current epoch ≥ deactivation epoch + 1), call `transferBalance` to return staked amount to delegator, call `deleteFile` on Stake Account, decrement ValidatorRecord TotalStake
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 10.1, 10.2, 10.4, 10.5_

- [ ] 4. Write the Staking Program in QuanticScript — reward distribution and slashing
  - Implement `handleDistributeRewards`: read RewardPool balance, iterate over all active ValidatorRecord files, calculate reward shares proportional to blocks produced, deduct commission, distribute remainder to Stake Accounts via `transferBalance`, retain fractional remainder in pool
  - Implement `handleReportDoubleSign`: decode two 64-byte signatures and one 32-byte validator pubkey from instruction data, verify both signatures using `verifySig` builtin, apply 5% slash to ValidatorRecord TotalStake, transfer slashed amount to RewardPool via `transferBalance`, set status to inactive if stake falls below threshold
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 5.1, 5.2, 5.3, 5.4_

- [ ] 5. Compile the Staking Program to bytecode
  - Run `go run cmd/main.go qsc compile -i programs/staking/staking.qs -o programs/staking/staking.qsb` and resolve any compiler errors
  - Run `go run cmd/main.go qsc disassemble -i programs/staking/staking.qsb -o programs/staking/staking.qsa` to produce the assembly artifact
  - Verify the bytecode is valid by running the existing bytecode verification tests against the output
  - _Requirements: 6.2_

- [ ] 6. Register the Staking Program in genesis
  - Update `internal/genesis/programs.go`: add `StakingProgramID` constant (`0x...03`), extend `LoadBuiltinPrograms` to accept and load `stakingBytecode []byte` alongside the existing system and token bytecodes
  - Update `cmd/main.go` to embed `programs/staking/staking.qsb` via `//go:embed` and pass it to `LoadBuiltinPrograms` at node startup
  - _Requirements: 6.5, 9.1_

- [ ] 7. Implement the validator schedule computation
  - Create `internal/staking/scheduler.go` with `ComputeSchedule(seed []byte, validators []ValidatorWithID, slotsPerEpoch int64) []ScheduleEntry`
  - Use `SHA-256(seed)` as the deterministic PRNG seed for weighted reservoir sampling; assign slots proportionally to `validator.TotalStake / totalStake`
  - _Requirements: 3.3, 9.4_

- [ ] 8. Implement genesis bootstrap
  - Create `internal/staking/genesis.go` with `InitializeGenesisStaking(fs *filestore.FileStore, validators []GenesisValidator) error`
  - Create ValidatorRecord Files for all genesis validators with status `active` and pre-assigned stake, bypassing the normal activation threshold
  - Create the EpochState File (epoch 0, slot 0) and RewardPool File (zero balance)
  - Compute and persist the epoch-0 Validator Schedule using `ComputeSchedule`
  - Make the function idempotent (skip if EpochState File already exists); reject empty validator list
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [ ] 9. Implement epoch boundary processing and reward distribution (Go side)
  - Create `internal/staking/rewards.go` with `DistributeEpochRewards(fs *filestore.FileStore, epochState *EpochStateData) error` that reads ValidatorRecord and StakeAccount files, computes shares, and updates balances directly in the FileStore
  - Implement `ProcessEpochBoundary(fs *filestore.FileStore, lastBlockHash []byte) error` that activates/deactivates validators based on stake threshold, calls `DistributeEpochRewards`, computes the new schedule, and atomically updates the EpochState File
  - _Requirements: 3.1, 3.2, 4.1, 4.2, 4.3, 4.4, 4.5, 8.1, 8.4_

- [ ] 10. Extend ConsensusManager with DPoS schedule and epoch tracking
  - Add `localPubKey`, `fileStore`, `epochState`, `schedule`, and `slotsPerEpoch` fields to `ConsensusManager`; update `NewConsensusManager` signature
  - Rewrite `IsLeader(slot int64) bool` to look up the slot offset in the cached schedule
  - Add `GetScheduledValidator`, `RecordMissedBlock`, and `CheckAndProcessEpochBoundary` methods
  - Update node startup in `cmd/main.go` to pass the new parameters and call `InitializeGenesisStaking` before the main loop
  - _Requirements: 3.3, 3.4, 3.5, 3.6, 8.2, 8.3, 8.4_

- [ ] 11. Build the Validator TUI application
  - Create `cmd/validator-tui/main.go` as a standalone binary accepting `--state <path>` and `--pubkey <hex>` flags
  - Open the FileStore in read-only mode; decode EpochStateData, RewardPool balance, and all ValidatorRecord files using the binary decoders from `internal/staking/types.go`
  - Render the ANSI dashboard: header panel, validator table with `[inactive]`/`[slashed]` row prefixes, summary panel with total staked, pool balance, and estimated APY
  - Refresh every 1000ms via `time.Ticker`; handle `q` keypress and `os.Interrupt` for clean exit
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

- [ ] 12. Write the DPoS automated demo script
  - Create `demo-dpos.sh` at the repo root accepting `<num_validators> <duration_seconds>` (min 2, max 9 validators)
  - Generate a genesis config, start N node processes, wait for epoch 0 blocks, submit `DelegateStake` transactions, wait for epoch boundary, trigger reward distribution, submit `ReportDoubleSign`
  - Emit structured JSON log entries to `logs/dpos-demo-<timestamp>.json` in the same format as `demo-automated.sh`; print human-readable summary to stdout; exit 0 on success
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8_

- [ ] 13. Write unit tests for Go staking types and schedule computation
  - Test `ValidatorRecord`, `StakeAccountData`, and `EpochStateData` binary encode/decode round-trips
  - Test `ComputeSchedule` determinism (same seed → same schedule) and proportionality (higher stake → more slots)
  - Test `InitializeGenesisStaking` idempotency and zero-validator rejection
  - _Requirements: 1.1–1.5, 2.1–2.6, 9.1–9.6, 10.1–10.5_

- [ ] 14. Write integration test for the full DPoS lifecycle
  - Create `internal/staking_integration_test.go`
  - Cover: genesis init → compile and load staking.qsb → submit RegisterValidator tx → DelegateStake tx → epoch boundary → DistributeRewards → UndelegateStake → cooldown → WithdrawStake → ReportDoubleSign → validator deactivation
  - _Requirements: 3.1, 3.2, 4.1–4.5, 5.1–5.4, 8.1–8.4, 9.1–9.6_

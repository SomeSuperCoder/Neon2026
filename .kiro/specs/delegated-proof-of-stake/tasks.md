# Implementation Plan

- [ ] 1. Add Go serialization helpers for DPoS data models
  - Create `internal/quanticscript/stdlib_staking.go` with `SerializeValidatorRecord`, `DeserializeValidatorRecord`, `SerializeStakeAccount`, `DeserializeStakeAccount`, `SerializeEpochState`, `DeserializeEpochState`, `SerializeRewardPool`, `DeserializeRewardPool`
  - Define well-known FileID constants: `StakingProgramID` (`0x...03`), `EpochStateFileID` (`0x...04`), `RewardPoolFileID` (`0x...05`) in `internal/genesis/programs.go`
  - _Requirements: 6.1, 10.1, 10.2, 10.3_

- [ ] 1.1 Write unit tests for serialization round-trips
  - Create `internal/quanticscript/stdlib_staking_test.go` covering all four data model types
  - _Requirements: 6.1, 10.1_

- [ ] 2. Implement the Staking Program in QuanticScript
- [ ] 2.1 Write `programs/staking/staking.qs` — entry dispatch and validator registration handlers
  - Implement `entry()` with byte-dispatch to all 7 instruction handlers
  - Implement `handleRegisterValidator` and `handleDeregisterValidator`
  - Use `createFile`/`deleteFile` builtins; set `File.Balance` to storage cost only (Req 10.2, 10.3)
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.4, 10.2, 10.3_

- [ ] 2.2 Add stake delegation handlers to `programs/staking/staking.qs`
  - Implement `handleDelegateStake`, `handleUndelegateStake`, `handleWithdrawStake`
  - Enforce cooldown period (1 epoch), rent-reserve separation, and balance checks
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 10.1, 10.4, 10.5_

- [ ] 2.3 Add reward distribution handler to `programs/staking/staking.qs`
  - Implement `handleDistributeRewards` — proportional reward split, commission deduction, fractional remainder retention
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 2.4 Add slashing handler to `programs/staking/staking.qs`
  - Implement `handleReportDoubleSign` — verify both signatures, reduce stake by 5%, deactivate if below threshold, transfer slashed amount to Reward Pool
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 2.5 Compile `staking.qs` to `staking.qsa` and `staking.qsb`
  - Run `go run cmd/main.go qsc compile -i programs/staking/staking.qs -o programs/staking/staking.qsb`
  - Verify no compiler errors; keep `.qs`, `.qsa`, `.qsb` in sync
  - _Requirements: 6.2, 6.3_

- [ ] 2.6 Write QuanticScript interpreter tests for each instruction type
  - Create `internal/quanticscript/staking_program_test.go` exercising all 7 handlers against a mock ExecutionContext
  - _Requirements: 6.3, 6.4, 6.5, 6.6_

- [ ] 3. Extend genesis bootstrap to load the Staking Program and initialize DPoS state
- [ ] 3.1 Update `internal/genesis/programs.go` — add `stakingBytecode` parameter to `LoadBuiltinPrograms`
  - Load `staking.qsb` at `StakingProgramID` using the existing `loadProgram` helper
  - _Requirements: 6.5, 9.1_

- [ ] 3.2 Implement `InitializeDPoSGenesis` in `internal/genesis/programs.go`
  - Create Epoch State File at `EpochStateFileID` with epoch 0 data
  - Create Reward Pool File at `RewardPoolFileID` with zero balance
  - Create one Validator Record File per genesis validator (status=active, pre-assigned stake, commission=0)
  - Reject startup if genesis config has zero validators
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.6_

- [ ] 3.3 Update `programs/embed.go` to embed `programs/staking/staking.qsb`
  - Add `//go:embed programs/staking/staking.qsb` directive alongside existing system/token embeds
  - Pass staking bytecode through to `LoadBuiltinPrograms` in `cmd/main.go`
  - _Requirements: 6.5_

- [ ] 3.4 Write genesis integration tests
  - Extend `internal/genesis/programs_test.go` to verify Staking Program, Epoch State, and Reward Pool files are created with correct balances and data
  - _Requirements: 9.2, 9.3, 10.2, 10.3_

- [ ] 4. Extend ConsensusManager with DPoS scheduling and epoch processing
- [ ] 4.1 Add epoch fields and `GenesisConfig` to `internal/consensus/consensus_manager.go`
  - Add `epochLength`, `fileStore`, `runtime`, `currentEpoch`, `validatorSchedule`, `genesisValidators` fields
  - Define `GenesisConfig` and `GenesisValidator` structs
  - Update `NewConsensusManager` to accept a `GenesisConfig`
  - _Requirements: 3.3, 9.1_

- [ ] 4.2 Implement `InitializeGenesis` on `ConsensusManager`
  - Call `InitializeDPoSGenesis` if Epoch State File is absent; restore from file if present
  - Log warning and re-initialize from slot 0 if Epoch State File is corrupted
  - _Requirements: 9.2, 9.3, 9.4_

- [ ] 4.3 Implement stake-weighted validator schedule computation
  - Implement `computeValidatorSchedule(epochSeed []byte, validators []ValidatorEntry) []filestore.FileID`
  - Use deterministic weighted-random algorithm seeded with last block hash of previous epoch
  - Persist compact schedule to Epoch State File
  - _Requirements: 3.3, 9.4_

- [ ] 4.4 Implement `IsLeader`, `GetScheduledValidator`, and slot-skip logic
  - Replace static `IsLeader` with stake-weighted schedule lookup
  - Implement 200 ms wait for scheduled validator; skip slot and record missed block on timeout
  - _Requirements: 3.4, 3.5, 3.6_

- [ ] 4.5 Implement `ProcessEpochBoundary` — activate/deactivate validators and trigger reward distribution
  - At each epoch boundary: update Validator Record statuses based on stake threshold (1,000,000 electrons)
  - Submit synthetic `DistributeRewards` instruction to Staking Program via Runtime
  - Reset per-epoch block counters in Validator Records
  - _Requirements: 3.1, 3.2, 4.1, 4.2_

- [ ] 4.6 Implement `RecordMissedBlock` — update missed-block counter in Validator Record and Epoch State File
  - _Requirements: 3.6, 9.4_

- [ ] 4.7 Write ConsensusManager unit tests
  - Test epoch boundary detection, schedule computation determinism, missed-block recording
  - _Requirements: 3.3, 3.6, 9.2_

- [ ] 5. Implement the Validator TUI App
- [ ] 5.1 Create `cmd/validator-tui/main.go` with `--state` flag and FileStore read-only open
  - Parse `--state` flag; open BadgerDB in read-only mode
  - Enumerate all FileIDs; classify files by prefix pattern (validator records, stake accounts, epoch state, reward pool)
  - _Requirements: 7.2_

- [ ] 5.2 Implement terminal dashboard rendering with 1000 ms refresh loop
  - Render header panel: epoch, slot, local validator status, active validator count, local delegated stake
  - Render validator table with `[inactive]`/`[slashed]` row prefixes
  - Render summary footer: total staked electrons, Reward Pool balance, estimated APY
  - Handle `q` / `Ctrl+C` for clean exit
  - _Requirements: 7.1, 7.3, 7.4, 7.5, 7.6, 7.7_

- [ ] 6. Implement the DPoS demo script
- [ ] 6.1 Create `demo-dpos.sh` with genesis start and block production phase
  - Accept `<num_validators> <duration_seconds>` args; start N validator nodes from genesis config
  - Log each block with slot, validator pubkey, block hash to `logs/dpos-demo-<timestamp>.json`
  - _Requirements: 11.1, 11.2_

- [ ] 6.2 Add delegation, epoch boundary, and reward distribution phases to `demo-dpos.sh`
  - Submit one `DelegateStake` tx per validator; wait for epoch boundary; log Validator Schedule and stakes
  - Log each validator's and delegator's reward amount in electrons
  - _Requirements: 11.3, 11.4_

- [ ] 6.3 Add slashing phase and summary output to `demo-dpos.sh`
  - Submit `ReportDoubleSign` against one validator; log slashing event, reduced stake, deactivation status
  - Print human-readable summary table to stdout; exit 0 on full success, non-zero on failure
  - Ensure log format is compatible with `analyze-results.sh`
  - _Requirements: 11.5, 11.6, 11.7, 11.8_

- [ ] 7. Wire everything together in `cmd/main.go`
  - Pass staking bytecode to `LoadBuiltinPrograms`
  - Pass `GenesisConfig` to `NewConsensusManager`
  - Register `validator-tui` subcommand (or build as separate binary)
  - Ensure node startup calls `InitializeGenesis` before processing slots
  - _Requirements: 6.5, 9.1, 9.2_

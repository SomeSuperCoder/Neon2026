# Requirements Document

## Introduction

This document specifies the requirements for implementing Delegated Proof of Stake (DPoS) consensus on the PoH Blockchain. The feature introduces a native Staking Program that manages validator registration, token delegation, reward distribution, and slashing. It also introduces a Validator TUI (terminal user interface) application that allows operators to monitor validator status, stake metrics, and network health in real time. The DPoS mechanism replaces the current static leader assignment with a dynamic, stake-weighted validator schedule that integrates with the existing PoH clock and slot-based consensus.

## Glossary

- **DPoS**: Delegated Proof of Stake — a consensus variant where token holders delegate stake to validators who produce blocks on their behalf.
- **Neon**: The native token of the blockchain. 1 Neon = 1,000,000 electrons.
- **Electron**: The smallest indivisible unit of the native token. All on-chain balances are stored and computed in electrons.
- **Validator**: A network node registered in the Staking Program that is eligible to be selected as a slot leader based on its total delegated stake.
- **Delegator**: A token holder who assigns (delegates) a portion of their balance to a Validator to increase that Validator's voting weight.
- **Stake Account**: A File in the FileStore that records a Delegator's delegation to a specific Validator, including the staked amount and activation epoch.
- **Validator Record**: A File in the FileStore that stores a Validator's public key, commission rate, total delegated stake, and status.
- **Epoch**: A fixed number of consecutive slots (default: 432,000 slots ≈ 2 days at 400 ms/slot) after which the validator schedule is recalculated and rewards are distributed.
- **Staking Program**: The native built-in program (registered in the Runtime) that processes all staking-related instructions.
- **Validator Schedule**: An ordered list of Validators assigned to produce blocks for each slot within an epoch, weighted by stake.
- **Commission**: The percentage of epoch rewards retained by a Validator before distributing the remainder to Delegators.
- **Slashing**: The reduction of a Validator's staked balance as a penalty for provable misbehavior (e.g., double-signing).
- **Reward Pool**: A File that accumulates transaction fees and inflation-based issuance to be distributed at epoch boundaries.
- **TUI**: Terminal User Interface — a text-based interactive dashboard rendered in the terminal.
- **Validator TUI App**: A standalone Go binary that connects to a running node and displays live validator and staking metrics.
- **PoH Clock**: The sequential SHA-256 hash chain used as the blockchain's cryptographic clock.
- **Slot**: A 400 ms time window during which a single Validator is scheduled to produce a block.
- **ConsensusManager**: The existing Go struct in `internal/consensus` that manages slot timing and leader selection.
- **FileStore**: The BadgerDB-backed state store in `internal/filestore` that persists all on-chain Files.
- **Runtime**: The program execution environment in `internal/runtime` that dispatches instructions to built-in and bytecode programs.
- **SystemProgramID**: The well-known FileID `0x...01` identifying the built-in System Program.
- **StakingProgramID**: The well-known FileID `0x...03` identifying the Staking Program bytecode file in the FileStore.

---

## Requirements

### Requirement 1 — Validator Registration

**User Story:** As a node operator, I want to register my node as a Validator on-chain, so that I can be eligible to produce blocks and earn rewards.

#### Acceptance Criteria

1. WHEN a `RegisterValidator` instruction is submitted to the Staking Program with a valid public key, commission rate between 0 and 100, and a signed transaction, THE Staking Program SHALL create a Validator Record File in the FileStore with status `inactive`, zero total delegated stake, and the provided commission rate.
2. WHEN a `RegisterValidator` instruction is submitted with a commission rate outside the range 0–100, THE Staking Program SHALL reject the instruction and return an error indicating the commission rate is invalid.
3. WHEN a `RegisterValidator` instruction is submitted for a public key that already has an existing Validator Record, THE Staking Program SHALL reject the instruction and return an error indicating the validator is already registered.
4. WHILE a Validator Record exists with status `inactive` and total delegated stake below the minimum activation threshold of 1,000,000 electrons (1 Neon), THE Staking Program SHALL keep the Validator Record status as `inactive` and exclude the Validator from the Validator Schedule.
5. WHEN a `DeregisterValidator` instruction is signed by the Validator's public key and the Validator Record has zero delegated stake, THE Staking Program SHALL delete the Validator Record File and return all remaining balance to the Validator's account.

---

### Requirement 2 — Stake Delegation

**User Story:** As a token holder, I want to delegate my tokens to a Validator, so that I can earn staking rewards and contribute to network security.

#### Acceptance Criteria

1. WHEN a `DelegateStake` instruction is submitted with a valid Validator Record FileID, a positive amount, and a signed transaction from the Delegator, THE Staking Program SHALL create a Stake Account File recording the Delegator's public key, the target Validator FileID, the staked amount, and the current epoch as the activation epoch.
2. WHEN a `DelegateStake` instruction is submitted and the Delegator's account balance is less than the requested delegation amount plus the minimum rent-exempt balance of 1,000 electrons, THE Staking Program SHALL reject the instruction and return an insufficient-balance error.
3. WHEN a `DelegateStake` instruction references a Validator Record with status `deregistered`, THE Staking Program SHALL reject the instruction and return an error indicating the target Validator is not accepting delegations.
4. WHEN an `UndelegateStake` instruction is submitted by the Delegator who owns a Stake Account, THE Staking Program SHALL mark the Stake Account as `deactivating` and record the current epoch as the deactivation epoch, without immediately returning funds.
5. WHILE a Stake Account has status `deactivating` and the current epoch is less than the deactivation epoch plus the cooldown period of 1 epoch, THE Staking Program SHALL prevent withdrawal of the staked amount.
6. WHEN a `WithdrawStake` instruction is submitted for a Stake Account with status `deactivating` and the cooldown period has elapsed, THE Staking Program SHALL transfer the staked amount back to the Delegator's account and delete the Stake Account File.

---

### Requirement 3 — Validator Activation and Scheduling

**User Story:** As a network participant, I want validators to be selected as block producers based on their stake weight, so that the network remains decentralized and economically secured.

#### Acceptance Criteria

1. WHEN an epoch boundary is reached and a Validator Record has total delegated stake greater than or equal to 1,000,000 electrons (1 Neon), THE ConsensusManager SHALL set the Validator Record status to `active` and include the Validator in the next epoch's Validator Schedule.
2. WHEN an epoch boundary is reached and a Validator Record has total delegated stake below 1,000,000 electrons (1 Neon), THE ConsensusManager SHALL set the Validator Record status to `inactive` and exclude the Validator from the next epoch's Validator Schedule.
3. WHEN the Validator Schedule for an epoch is computed, THE ConsensusManager SHALL assign slots to Validators proportionally to their total delegated stake using a deterministic weighted-random algorithm seeded with the last block hash of the previous epoch.
4. WHEN a slot is reached and the scheduled Validator for that slot is the local node, THE ConsensusManager SHALL produce a block and broadcast it to peers.
5. WHEN a slot is reached and the scheduled Validator for that slot is not the local node, THE ConsensusManager SHALL wait for a block from the scheduled Validator for up to 200 ms before advancing to the next slot.
6. IF the scheduled Validator for a slot fails to produce a block within 200 ms, THEN THE ConsensusManager SHALL skip the slot and record a missed-block event against that Validator's Record.

---

### Requirement 4 — Reward Distribution

**User Story:** As a Validator and Delegator, I want to receive staking rewards at the end of each epoch, so that I am economically incentivized to participate honestly.

#### Acceptance Criteria

1. WHEN an epoch boundary is reached, THE Staking Program SHALL calculate each active Validator's reward share as proportional to the number of blocks produced by that Validator during the epoch divided by the total blocks produced by all active Validators.
2. WHEN an epoch boundary is reached and a Validator's reward share is calculated, THE Staking Program SHALL transfer the Validator's commission portion (commission rate × reward share) to the Validator's account and distribute the remainder proportionally across all Stake Accounts delegated to that Validator.
3. WHEN a Validator produced zero blocks during an epoch, THE Staking Program SHALL assign that Validator a reward share of zero and distribute no rewards to its Delegators.
4. WHILE the Reward Pool balance is zero, THE Staking Program SHALL skip reward distribution and emit a zero-reward event for the epoch.
5. IF a reward distribution calculation results in a fractional electron remainder, THEN THE Staking Program SHALL retain the remainder in the Reward Pool for the next epoch.

---

### Requirement 5 — Slashing

**User Story:** As a network participant, I want misbehaving validators to be penalized, so that honest behavior is economically enforced.

#### Acceptance Criteria

1. WHEN a `ReportDoubleSign` instruction is submitted with two valid block headers from the same slot signed by the same Validator public key, THE Staking Program SHALL verify both signatures and, if valid, reduce the Validator's total delegated stake by 5% of its current value.
2. WHEN a slashing event reduces a Validator's total delegated stake below the activation threshold of 1,000,000 electrons (1 Neon), THE Staking Program SHALL set the Validator Record status to `inactive` and remove the Validator from the current epoch's Validator Schedule.
3. WHEN a slashing event is applied, THE Staking Program SHALL distribute the slashed amount to the Reward Pool.
4. IF a `ReportDoubleSign` instruction references a Validator Record that does not exist, THEN THE Staking Program SHALL reject the instruction and return a not-found error.

---

### Requirement 6 — Staking Program as QuanticScript Contract

**User Story:** As a developer, I want the Staking Program to be written in QuanticScript and compiled to bytecode, so that it executes through the same bytecode interpreter as all other on-chain programs and can be audited, upgraded, and tested like any smart contract.

#### Acceptance Criteria

1. THE Staking Program SHALL be authored as a QuanticScript source file at `programs/staking/staking.qs`, following the same structure and builtin conventions used by `programs/system/system.qs` and `programs/token/token.qs`.
2. WHEN the QuanticScript compiler is invoked on `programs/staking/staking.qs`, THE compiler SHALL produce a valid assembly file at `programs/staking/staking.qsa` and a bytecode file at `programs/staking/staking.qsb` without errors.
3. THE Staking Program source SHALL export a single `entry()` function that reads the first byte of instruction data to dispatch to the appropriate handler function for each of the seven instruction types: `RegisterValidator`, `DeregisterValidator`, `DelegateStake`, `UndelegateStake`, `WithdrawStake`, `DistributeRewards`, and `ReportDoubleSign`.
4. WHEN the Staking Program is invoked with an unrecognized instruction type byte, THE Staking Program SHALL return the error code `ERROR_INVALID_INSTRUCTION` (value `0x3FFF`).
5. WHEN a node starts, THE node SHALL load the compiled `staking.qsb` bytecode into the FileStore under the well-known `StakingProgramID` (`0x...03`) with `Executable = true`, using the same `LoadBuiltinPrograms` mechanism in `internal/genesis/programs.go` used for the System and Token programs.
6. THE Staking Program SHALL use only QuanticScript builtins available to all programs (`getInstructionData`, `getFile`, `getBalance`, `hasSigner`, `updateFile`, `createFile`, `transferBalance`, `deleteFile`, `len`, `slice`, `hashBytes`) and SHALL NOT require any new runtime builtins beyond those already used by the Token Program.

---

### Requirement 7 — Validator TUI Application

**User Story:** As a node operator, I want a terminal dashboard that shows live validator and staking metrics, so that I can monitor my node's performance and the network's health without writing custom tooling.

#### Acceptance Criteria

1. THE Validator TUI App SHALL render a terminal dashboard that refreshes every 1,000 ms and displays: current epoch number, current slot number, local validator status, total active validators count, and local node's total delegated stake.
2. WHEN the Validator TUI App is started with a `--state` flag pointing to a valid FileStore path, THE Validator TUI App SHALL read validator and staking state directly from the FileStore without requiring a running node.
3. THE Validator TUI App SHALL display a sortable table of all Validator Records showing: validator public key (truncated to 16 hex chars), status, total delegated stake, commission rate, and blocks produced in the current epoch.
4. WHEN a Validator Record has status `inactive`, THE Validator TUI App SHALL render that row with a distinct visual indicator (e.g., dimmed or prefixed with `[inactive]`).
5. WHEN a Validator Record has been slashed in the current epoch, THE Validator TUI App SHALL render that row with a warning indicator (e.g., prefixed with `[slashed]`).
6. THE Validator TUI App SHALL display a summary panel showing: total staked electrons across all active validators, Reward Pool balance, and estimated APY based on current epoch reward rate.
7. WHEN the user presses `q` or `Ctrl+C` in the Validator TUI App, THE Validator TUI App SHALL exit cleanly, releasing all FileStore resources.

---

### Requirement 9 — Genesis Bootstrap

**User Story:** As a network founder, I want the blockchain to start producing blocks at genesis without requiring prior stake delegation, so that the network can bootstrap from zero before any token holders exist.

#### Acceptance Criteria

1. THE ConsensusManager SHALL accept a genesis configuration at startup that specifies an ordered list of genesis Validator public keys and their pre-assigned stake amounts in electrons.
2. WHEN a node starts and the Epoch State File does not exist in the FileStore, THE ConsensusManager SHALL initialize epoch 0 by creating a Validator Record File for each genesis Validator with status `active`, the pre-assigned stake amount, and commission rate 0, without requiring the normal activation threshold.
3. WHEN a node starts and the Epoch State File does not exist in the FileStore, THE ConsensusManager SHALL create the Reward Pool File with an initial balance of zero electrons.
4. WHEN a node starts and the Epoch State File does not exist in the FileStore, THE ConsensusManager SHALL compute the epoch 0 Validator Schedule from the genesis validator set and persist it in the Epoch State File before processing any slots.
5. WHILE the network is in epoch 0 and no non-genesis stake delegations exist, THE Staking Program SHALL distribute epoch rewards exclusively among genesis Validators proportional to their pre-assigned stake amounts.
6. IF the genesis configuration specifies zero validators, THEN THE ConsensusManager SHALL reject startup and return an error indicating at least one genesis Validator is required.


**User Story:** As a node operator, I want epoch state to survive node restarts, so that staking and reward data is not lost on crashes or planned maintenance.

#### Acceptance Criteria

1. THE Staking Program SHALL persist the current epoch number, epoch start slot, and Validator Schedule as a dedicated Epoch State File in the FileStore, updated atomically at each epoch boundary.
2. WHEN a node restarts, THE ConsensusManager SHALL read the Epoch State File from the FileStore to restore the current epoch number and Validator Schedule before processing new slots.
3. IF the Epoch State File is missing or corrupted on startup, THEN THE ConsensusManager SHALL initialize a new epoch starting at slot 0 and log a warning.
4. WHILE the node is running, THE Staking Program SHALL update the Epoch State File's missed-block counters after each slot in which the scheduled Validator did not produce a block.

---

### Requirement 10 — Storage Cost and Stake Balance Separation

**User Story:** As a developer, I want stake account and validator record balances to be protected from storage cost deductions, so that staked funds are never silently consumed by file rent and the two accounting systems do not conflict.

#### Acceptance Criteria

1. THE Staking Program SHALL store the staked amount and the rent reserve as separate fields within the Stake Account File's data payload, rather than relying on the File's top-level `Balance` field to represent both.
2. WHEN a Stake Account File is created, THE Staking Program SHALL set the File's top-level `Balance` field to a value equal to the storage cost of the Stake Account File's data size, ensuring the FileStore's `ValidateStorageCost` check passes without consuming any staked electrons.
3. WHEN a Validator Record File is created, THE Staking Program SHALL set the File's top-level `Balance` field to a value equal to the storage cost of the Validator Record File's data size, keeping the validator's staked total exclusively within the data payload.
4. WHEN a `WithdrawStake` instruction returns funds to a Delegator, THE Staking Program SHALL return only the staked amount from the data payload and SHALL NOT reduce the File's top-level `Balance` below the storage cost required for the Stake Account File's data size.
5. IF a Staking Program instruction would result in a File's top-level `Balance` falling below the storage cost for that File's current data size, THEN THE Staking Program SHALL reject the instruction and return an insufficient-rent-reserve error.

---

### Requirement 11 — Developer and AI-Friendly Demo

**User Story:** As a developer or AI agent, I want a comprehensive automated demo script that exercises the full DPoS lifecycle, so that I can verify correct behavior, reproduce issues, and onboard quickly without manual setup.

#### Acceptance Criteria

1. THE demo script SHALL be executable as a single command (`./demo-dpos.sh <num_validators> <duration_seconds>`) with no interactive prompts, producing structured JSON log output to `logs/dpos-demo-<timestamp>.json` for machine parsing.
2. WHEN the demo script runs, THE demo script SHALL start a configurable number of validator nodes (minimum 2, maximum 9) from a genesis configuration, wait for epoch 0 to begin producing blocks, and log each block produced with its slot number, validator public key, and block hash.
3. WHEN the demo script runs, THE demo script SHALL submit at least one `DelegateStake` transaction per validator, wait for the next epoch boundary, and log the resulting Validator Schedule and each validator's delegated stake.
4. WHEN the demo script runs, THE demo script SHALL trigger at least one epoch reward distribution and log each validator's reward amount and each delegator's reward amount in electrons.
5. WHEN the demo script runs, THE demo script SHALL submit a `ReportDoubleSign` instruction against one validator and log the slashing event, the reduced stake, and whether the validator was deactivated.
6. WHEN the demo script completes, THE demo script SHALL print a human-readable summary table to stdout showing: total blocks produced, total rewards distributed in electrons, number of slashing events, and pass/fail status for each lifecycle phase.
7. THE demo script SHALL exit with code 0 if all lifecycle phases completed without error, and exit with a non-zero code if any phase failed, logging the failure reason to the JSON output.
8. THE demo script SHALL be compatible with the existing `analyze-results.sh` script by producing log entries in the same structured format used by `demo-automated.sh`.

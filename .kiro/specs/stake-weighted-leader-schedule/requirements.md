# Requirements Document

## Introduction

This document specifies the requirements for implementing a deterministic stake-weighted leader schedule for the PoH Blockchain consensus mechanism. The current implementation uses a static leader assignment based on node type (LEADER vs REPLICA), which is unacceptable for a real blockchain. This feature replaces the static system with a dynamic, stake-based validator rotation that determines which validator produces blocks for each slot based on their delegated stake weight.

The leader schedule is computed deterministically at epoch boundaries using a weighted-random algorithm seeded with the last block hash of the previous epoch. All nodes in the network compute the same schedule independently, ensuring consensus without coordination.

## Glossary

- **Validator**: A network node that has registered in the Staking Program and has sufficient delegated stake to participate in block production.
- **Leader Schedule**: A deterministic mapping of slot numbers to validator FileIDs, computed at each epoch boundary based on stake weights.
- **Slot**: A 400 ms time window during which exactly one validator is scheduled to produce a block.
- **Epoch**: A fixed number of consecutive slots (default: 432,000 slots ≈ 2 days at 400 ms/slot) after which the leader schedule is recalculated.
- **Stake Weight**: The total amount of delegated stake (in electrons) assigned to a validator, which determines their probability of being selected as leader for any given slot.
- **Epoch Seed**: The hash of the last block in the previous epoch, used as the random seed for computing the next epoch's leader schedule.
- **ConsensusManager**: The Go struct in `internal/consensus` that manages slot timing and leader selection.
- **Validator Record**: A File in the FileStore that stores a validator's public key, total delegated stake, and status.
- **Epoch State File**: A File in the FileStore (FileID `0x...04`) that persists the current epoch number, epoch start slot, and the leader schedule.
- **Node Identity**: The validator public key that uniquely identifies a node in the network, replacing the old LEADER/REPLICA node type distinction.
- **Local Validator FileID**: The FileID corresponding to the local node's validator public key, used to determine if the node should produce a block for a given slot.

---

## Requirements

### Requirement 1 — Node Identity Based on Validator Public Key with Secure Wallet

**User Story:** As a node operator, I want my node to be identified by its validator public key from a password-protected wallet, so that the network can dynamically assign leadership based on stake while keeping my private keys secure.

#### Acceptance Criteria

1. WHEN a node starts, THE node SHALL accept a `--wallet` flag pointing to a wallet name (not a file path), and SHALL prompt the user for the wallet password via stdin.
2. THE wallet files SHALL be stored in a platform-specific configuration directory: `~/.config/poh-blockchain/wallets/` on Linux/macOS, and `%APPDATA%\poh-blockchain\wallets\` on Windows.
3. WHEN a wallet is opened with the correct password, THE wallet SHALL decrypt and load the Ed25519 keypair (public key and private key) into memory.
4. WHEN a wallet contains multiple keypairs, THE node SHALL prompt the user to select which keypair to use for validation, displaying each keypair's public key (truncated to 16 hex chars) and associated validator status (if registered).
5. WHEN a node starts with a valid `--wallet` flag and correct password, THE ConsensusManager SHALL compute the Validator Record FileID from the selected public key using `SHA-256("validator:" || pubkey)` and store it as the local validator identity.
6. WHEN a node starts without a `--wallet` flag, THE node SHALL operate in observer mode (no block production) and log a warning indicating the node is not configured as a validator.
7. THE ConsensusManager SHALL remove all references to the `network.NodeType` enum (LEADER, REPLICA, OBSERVER) from leader selection logic.
8. THE ConsensusManager SHALL replace the `nodeType` field with a `localValidatorID` field of type `filestore.FileID`.
9. THE wallet password SHALL NOT be stored on disk or logged; it SHALL only exist in memory during the decryption process.
10. THE wallet file format SHALL use AES-256-GCM encryption with a key derived from the password using Argon2id (memory=64MB, iterations=3, parallelism=4).

---

### Requirement 2 — Deterministic Stake-Weighted Leader Schedule Computation

**User Story:** As a network participant, I want the leader schedule to be computed deterministically based on validator stake weights, so that all nodes agree on which validator should produce each block without coordination.

#### Acceptance Criteria

1. WHEN an epoch boundary is reached (slot % epochLength == 0), THE ConsensusManager SHALL enumerate all active Validator Records from the FileStore and extract their FileIDs and total delegated stake amounts.
2. WHEN the ConsensusManager computes a leader schedule, THE ConsensusManager SHALL use a deterministic weighted-random algorithm seeded with the last block hash of the previous epoch to assign validators to slots proportionally to their stake.
3. WHEN the ConsensusManager computes a leader schedule, THE ConsensusManager SHALL ensure that a validator with twice the stake of another validator is assigned approximately twice as many slots over the course of the epoch.
4. WHEN the ConsensusManager computes a leader schedule for epoch N, THE ConsensusManager SHALL use the block hash from the last slot of epoch N-1 as the random seed, ensuring all nodes compute the same schedule.
5. WHEN the ConsensusManager computes a leader schedule and no active validators exist, THE ConsensusManager SHALL reject the computation and return an error indicating at least one active validator is required.
6. THE leader schedule computation algorithm SHALL be deterministic: given the same set of validators, stakes, and epoch seed, all nodes SHALL produce identical schedules.

---

### Requirement 3 — Leader Selection Based on Schedule Lookup

**User Story:** As a validator node, I want to produce blocks only when I am the scheduled leader for the current slot, so that block production follows the deterministic stake-weighted schedule.

#### Acceptance Criteria

1. WHEN a slot begins, THE ConsensusManager SHALL call `GetScheduledValidator(slot)` to retrieve the FileID of the validator scheduled to produce a block for that slot.
2. WHEN the scheduled validator FileID for a slot matches the local node's validator FileID, THE ConsensusManager SHALL return `true` from `IsLeader(slot)` and the node SHALL produce a block.
3. WHEN the scheduled validator FileID for a slot does NOT match the local node's validator FileID, THE ConsensusManager SHALL return `false` from `IsLeader(slot)` and the node SHALL wait to receive a block from the scheduled validator.
4. WHEN a node is in observer mode (no validator keypair configured), THE ConsensusManager SHALL always return `false` from `IsLeader(slot)`.
5. THE `IsLeader(slot)` method SHALL NOT check the old `nodeType` field; it SHALL ONLY compare the scheduled validator FileID with the local validator FileID.

---

### Requirement 4 — Epoch Boundary Processing and Schedule Persistence

**User Story:** As a node operator, I want the leader schedule to be recalculated and persisted at each epoch boundary, so that the schedule adapts to changes in validator stakes and survives node restarts.

#### Acceptance Criteria

1. WHEN an epoch boundary is reached, THE ConsensusManager SHALL call `ProcessEpochBoundary(slot)` to trigger schedule recalculation.
2. WHEN `ProcessEpochBoundary(slot)` is called, THE ConsensusManager SHALL enumerate all Validator Records with status `active` and total delegated stake >= 1,000,000 electrons (1 Neon).
3. WHEN `ProcessEpochBoundary(slot)` is called, THE ConsensusManager SHALL compute the new leader schedule using `ComputeValidatorSchedule(epochSeed, validators)` with the last block hash as the seed.
4. WHEN `ProcessEpochBoundary(slot)` is called, THE ConsensusManager SHALL serialize the new schedule and persist it to the Epoch State File in the FileStore.
5. WHEN a node restarts, THE ConsensusManager SHALL restore the current epoch number and leader schedule from the Epoch State File before processing any slots.
6. IF the Epoch State File is missing or corrupted on startup, THEN THE ConsensusManager SHALL log a warning and re-initialize from the genesis configuration.

---

### Requirement 5 — Slot Skip and Missed Block Handling

**User Story:** As a network participant, I want the network to continue producing blocks even when a scheduled validator fails to produce a block, so that the chain does not stall due to validator downtime.

#### Acceptance Criteria

1. WHEN a slot begins and the local node is not the scheduled leader, THE ConsensusManager SHALL wait for a block from the scheduled validator for up to 200 ms.
2. IF no block is received from the scheduled validator within 200 ms, THEN THE ConsensusManager SHALL skip the slot, record a missed-block event against the scheduled validator's Validator Record, and advance to the next slot.
3. WHEN a missed-block event is recorded, THE ConsensusManager SHALL increment the `missedBlocksThisEpoch` counter in the scheduled validator's Validator Record.
4. WHEN a missed-block event is recorded, THE ConsensusManager SHALL also update the missed-block counter in the Epoch State File for the scheduled validator.
5. THE ConsensusManager SHALL NOT produce a block for a slot if the local node is not the scheduled leader, even if the scheduled leader fails to produce a block.

---

### Requirement 6 — Genesis Bootstrap with Initial Validator Set

**User Story:** As a network founder, I want the blockchain to start with a predefined set of genesis validators and their stake weights, so that the network can begin producing blocks immediately without requiring prior stake delegation.

#### Acceptance Criteria

1. WHEN a node starts and the Epoch State File does not exist, THE ConsensusManager SHALL read the genesis configuration from a `--genesis-config` flag pointing to a JSON file.
2. THE genesis configuration JSON file SHALL specify an array of genesis validators, each with a public key (32-byte hex string) and a pre-assigned stake amount (int64 in electrons).
3. WHEN the ConsensusManager initializes genesis, THE ConsensusManager SHALL create a Validator Record File for each genesis validator with status `active`, the pre-assigned stake amount, and commission rate 0.
4. WHEN the ConsensusManager initializes genesis, THE ConsensusManager SHALL compute the epoch 0 leader schedule from the genesis validator set using a default seed (e.g., all zeros) and persist it to the Epoch State File.
5. IF the genesis configuration specifies zero validators, THEN THE ConsensusManager SHALL reject startup and return an error indicating at least one genesis validator is required.
6. WHILE the network is in epoch 0, THE leader schedule SHALL be based exclusively on the genesis validator set and their pre-assigned stakes.

---

### Requirement 7 — Remove All Static Leader/Replica Logic

**User Story:** As a developer, I want all traces of the old static leader/replica system removed from the codebase, so that the implementation is clean and only supports stake-weighted leader scheduling.

#### Acceptance Criteria

1. THE ConsensusManager SHALL NOT contain any logic that checks `nodeType == network.LEADER` or `nodeType == network.REPLICA`.
2. THE `NewConsensusManager` constructor SHALL NOT accept a `nodeType` parameter; it SHALL accept a `localValidatorID` parameter instead.
3. THE `cmd/main.go` node startup code SHALL NOT accept `--type=leader` or `--type=replica` flags; it SHALL accept `--wallet` instead.
4. THE `internal/network/network_node.go` SHALL remove the `NodeType` enum and all associated logic.
5. ALL integration tests and demo scripts SHALL be updated to use `--wallet` flags instead of `--type` flags.
6. THE codebase SHALL NOT contain any comments, variable names, or function names referencing "leader node" or "replica node" in the context of static node types.

---

### Requirement 8 — Integration with Existing DPoS Staking Program

**User Story:** As a developer, I want the leader schedule to integrate seamlessly with the existing DPoS Staking Program, so that validator registration, stake delegation, and reward distribution work correctly with the new leader selection mechanism.

#### Acceptance Criteria

1. WHEN a validator registers via the Staking Program's `RegisterValidator` instruction, THE validator SHALL be eligible for inclusion in the leader schedule once their total delegated stake reaches 1,000,000 electrons (1 Neon).
2. WHEN a validator's total delegated stake falls below 1,000,000 electrons, THE validator SHALL be excluded from the next epoch's leader schedule.
3. WHEN the Staking Program's `DistributeRewards` instruction is executed at an epoch boundary, THE reward distribution SHALL be based on the number of blocks each validator produced during the epoch, as recorded in their Validator Record's `blocksProducedThisEpoch` field.
4. WHEN a validator produces a block, THE ConsensusManager SHALL increment the `blocksProducedThisEpoch` counter in the validator's Validator Record.
5. THE leader schedule computation SHALL read validator stakes directly from Validator Record Files in the FileStore, ensuring consistency with the Staking Program's state.

---

### Requirement 9 — Logging and Observability

**User Story:** As a node operator, I want detailed logs about leader schedule computation and slot assignments, so that I can debug issues and verify the network is operating correctly.

#### Acceptance Criteria

1. WHEN an epoch boundary is reached and a new leader schedule is computed, THE ConsensusManager SHALL log the epoch number, the number of active validators, and the total stake across all validators.
2. WHEN a node produces a block, THE ConsensusManager SHALL log the slot number, the local validator FileID, and a confirmation that the node was the scheduled leader.
3. WHEN a node skips a slot due to a missed block, THE ConsensusManager SHALL log the slot number, the scheduled validator FileID, and the reason for the skip (timeout).
4. WHEN a node starts and restores the leader schedule from the Epoch State File, THE ConsensusManager SHALL log the restored epoch number and the number of validators in the schedule.
5. WHEN a node starts in observer mode (no validator keypair), THE ConsensusManager SHALL log a warning indicating the node will not produce blocks.

---

### Requirement 10 — Backward Compatibility and Migration

**User Story:** As a network operator, I want existing nodes to gracefully handle the transition from the old static leader system to the new stake-weighted system, so that the network can upgrade without downtime.

#### Acceptance Criteria

1. WHEN a node starts with the old `--type=leader` flag, THE node SHALL reject startup and print an error message instructing the operator to use `--wallet` instead.
2. WHEN a node starts with the old `--type=replica` flag, THE node SHALL reject startup and print an error message instructing the operator to use `--wallet` or run in observer mode.
3. THE error message SHALL include an example command showing how to create a new wallet using the CLI: `poh-blockchain wallet create --name <wallet_name>`.
4. THE migration SHALL NOT require changes to the FileStore schema or existing Validator Record Files.
5. THE migration SHALL NOT require a chain reset; existing Validator Records and Epoch State Files SHALL remain valid.

---

### Requirement 11 — Wallet Management CLI

**User Story:** As a node operator, I want a CLI to create, list, and manage password-protected wallets, so that I can securely store validator keypairs without manual file editing.

#### Acceptance Criteria

1. THE CLI SHALL provide a `wallet create --name <wallet_name>` command that prompts for a password (twice for confirmation) and generates a new Ed25519 keypair encrypted with the password.
2. THE CLI SHALL provide a `wallet list` command that displays all wallet names in the platform-specific wallet directory, without requiring passwords.
3. THE CLI SHALL provide a `wallet export --name <wallet_name> --output <file.json>` command that prompts for the wallet password and exports the keypair as unencrypted JSON (for backup purposes).
4. THE CLI SHALL provide a `wallet import --input <file.json> --name <wallet_name>` command that prompts for a new password and imports an unencrypted JSON keypair into an encrypted wallet.
5. THE CLI SHALL provide a `wallet show --name <wallet_name>` command that prompts for the wallet password and displays the public key(s) in the wallet (truncated to 16 hex chars).
6. WHEN a wallet is created, THE CLI SHALL store the encrypted wallet file at `~/.config/poh-blockchain/wallets/<wallet_name>.wallet` on Linux/macOS, or `%APPDATA%\poh-blockchain\wallets\<wallet_name>.wallet` on Windows.
7. THE wallet file extension SHALL be `.wallet` to distinguish it from other configuration files.

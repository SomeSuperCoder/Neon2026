# Implementation Plan

- [x] 1. Implement secure wallet management system
  - Create `internal/wallet/wallet.go` with `Wallet`, `Keypair` structs and core functions: `Create`, `Open`, `List`, `Export`, `Import`
  - Implement AES-256-GCM encryption with Argon2id key derivation (memory=64MB, iterations=3, parallelism=4)
  - Implement platform-specific wallet directory resolution: `~/.config/poh-blockchain/wallets/` on Linux/macOS, `%APPDATA%\poh-blockchain\wallets\` on Windows
  - Implement wallet file format: [salt(32)][nonce(12)][ciphertext][tag(16)]
  - _Requirements: 1.2, 1.3, 1.9, 1.10, 11.6_

- [x] 1.1 Write unit tests for wallet encryption and platform paths
  - Create `internal/wallet/wallet_test.go` covering encryption/decryption round-trips, password validation, incorrect password handling, corrupted file handling
  - Test platform-specific path resolution on Linux, macOS, Windows
  - _Requirements: 1.2, 1.9, 1.10, 11.6_

- [x] 2. Implement wallet CLI commands
  - Add `wallet create --name <name>` subcommand with password prompt (twice for confirmation)
  - Add `wallet list` subcommand displaying all wallet names
  - Add `wallet show --name <name>` subcommand displaying public keys (truncated to 16 hex chars)
  - Add `wallet export --name <name> --output <file>` subcommand exporting unencrypted JSON
  - Add `wallet import --input <file> --name <name>` subcommand importing and encrypting keypair
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.7_

- [x] 2.1 Write integration tests for wallet CLI commands
  - Test wallet creation, listing, showing, exporting, importing
  - Test error cases: wallet already exists, wallet not found, incorrect password
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

- [x] 3. Refactor ConsensusManager to remove NodeType and add validator identity
  - Remove `nodeType network.NodeType` field from `ConsensusManager` struct
  - Add `localValidatorID filestore.FileID` field (zero value if observer mode)
  - Add `localValidatorKeys *wallet.Keypair` field (nil if observer mode)
  - Update `NewConsensusManager` constructor to accept `localValidatorID` and `localKeys` parameters instead of `nodeType`
  - Update all callers of `NewConsensusManager` in tests and main.go
  - _Requirements: 1.5, 1.8, 7.1, 7.2_

- [x] 3.1 Implement stake-weighted IsLeader logic
  - Refactor `IsLeader(slot int64)` to return false if `localValidatorID` is zero (observer mode)
  - Refactor `IsLeader(slot int64)` to call `GetScheduledValidator(slot)` and compare with `localValidatorID`
  - Remove all logic checking `nodeType == network.LEADER` or `nodeType == network.REPLICA`
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 7.1_

- [x] 3.2 Write unit tests for stake-weighted IsLeader logic
  - Test IsLeader returns true when local validator is scheduled
  - Test IsLeader returns false when local validator is not scheduled
  - Test IsLeader returns false in observer mode (zero localValidatorID)
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 4. Implement deterministic leader schedule computation
  - Implement `enumerateActiveValidators()` method to read all Validator Records from FileStore and filter by status=active and stake>=1,000,000 electrons
  - Refactor `ComputeValidatorSchedule(epochSeed []byte, validators []ValidatorEntry)` to use LCG-based weighted random selection
  - Use LCG constants: multiplier=6364136223846793005, increment=1442695040888963407
  - Seed LCG with first 8 bytes of epochSeed (last block hash)
  - Assign slots proportionally to stake: validator with 2x stake gets ~2x slots
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [x] 4.1 Write unit tests for schedule computation determinism
  - Test same inputs (validators, stakes, seed) produce identical schedules
  - Test different seeds produce different schedules
  - Test stake-weighted distribution: validator with 2x stake gets ~2x slots (within 5% tolerance)
  - Test edge cases: single validator, zero total stake, empty validator list
  - _Requirements: 2.1, 2.2, 2.3, 2.6_

- [x] 5. Implement epoch boundary processing with schedule recalculation
  - Refactor `ProcessEpochBoundary(slot int64, lastBlockHash [32]byte)` to enumerate active validators
  - Compute new leader schedule using `ComputeValidatorSchedule(lastBlockHash[:], validators)`
  - Update `validatorSchedule` and `currentEpoch` fields
  - Persist new schedule to Epoch State File using compact format
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 5.1 Implement compact schedule serialization and in-memory expansion
  - Implement `compactSchedule(fullSchedule []filestore.FileID)` to convert full schedule to compact format (validator FileID + slot count)
  - Implement `expandSchedule(compactSchedule []ScheduleEntry)` to convert compact format to full slot-indexed array
  - Update `SerializeEpochState` and `DeserializeEpochState` in `internal/quanticscript/stdlib_staking.go` to use compact format
  - _Requirements: 4.4, 4.5_

- [x] 5.2 Write unit tests for epoch boundary processing
  - Test schedule recalculation at epoch boundary
  - Test schedule persistence to Epoch State File
  - Test schedule restoration from Epoch State File on node restart
  - Test compact schedule serialization and expansion round-trip
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

- [x] 6. Implement slot skip and missed block handling
  - Update block production loop to wait 200 ms for scheduled validator's block
  - If no block received within 200 ms, skip slot and call `RecordMissedBlock(slot, scheduledValidatorID)`
  - Ensure local node does NOT produce a block if not the scheduled leader, even if scheduled leader misses
  - _Requirements: 5.1, 5.2, 5.5_

- [x] 6.1 Update RecordMissedBlock to use scheduled validator
  - Refactor `RecordMissedBlock(slot int64, validatorID filestore.FileID)` to increment `missedBlocksThisEpoch` in Validator Record
  - Update missed-block counter in Epoch State File
  - _Requirements: 5.3, 5.4_

- [x] 6.2 Write unit tests for missed block handling
  - Test missed block counter increments in Validator Record
  - Test missed block counter increments in Epoch State File
  - Test slot skip does not produce block from non-scheduled validator
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 7. Implement genesis bootstrap with initial validator set
  - Update `InitializeGenesis(config GenesisConfig)` to read genesis validators from config
  - Create Validator Record File for each genesis validator with status=active, pre-assigned stake, commission=0
  - Compute epoch 0 leader schedule using default seed (all zeros) and persist to Epoch State File
  - Reject startup if genesis config has zero validators
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 7.1 Implement genesis configuration file loading
  - Define `GenesisConfig` struct with `EpochLength` and `Validators` fields
  - Implement `loadGenesisConfig(path string)` to parse JSON file
  - Add `--genesis-config <path>` flag to node startup
  - _Requirements: 6.1, 6.2_

- [x] 7.2 Write unit tests for genesis bootstrap
  - Test genesis initialization creates Validator Records with correct stakes
  - Test genesis initialization computes epoch 0 schedule
  - Test genesis initialization rejects zero validators
  - _Requirements: 6.2, 6.3, 6.4, 6.5_

- [x] 8. Refactor node startup to use wallet-based identity
  - Add `--wallet <name>` flag to node startup
  - Implement password prompt using `golang.org/x/term` for secure input
  - Open wallet with password and load keypairs
  - Implement `selectKeypair(wallet, fileStore)` to prompt user if multiple keypairs exist
  - Display each keypair's public key (truncated) and validator status (active/inactive/not registered)
  - Compute `localValidatorID` from selected public key using `SHA-256("validator:" || pubkey)`
  - Pass `localValidatorID` and `localKeys` to `NewConsensusManager`
  - Log validator identity or observer mode status
  - _Requirements: 1.1, 1.2, 1.4, 1.5, 1.6_

- [x] 8.1 Implement observer mode for nodes without wallet
  - When `--wallet` flag is omitted, set `localValidatorID` to zero value
  - Log warning: "Starting node in observer mode (no wallet specified)"
  - Ensure `IsLeader` always returns false in observer mode
  - _Requirements: 1.6, 3.4_

- [x] 8.2 Write integration tests for node startup with wallet
  - Test node starts with valid wallet and password
  - Test node starts in observer mode without wallet
  - Test node rejects incorrect wallet password
  - Test keypair selection prompt with multiple keypairs
  - _Requirements: 1.1, 1.2, 1.4, 1.6_

- [ ] 9. Remove all static leader/replica logic from codebase
  - Remove `network.NodeType` enum from `internal/network/network_node.go`
  - Remove all references to `LEADER`, `REPLICA`, `OBSERVER` constants
  - Remove `--type` flag from `cmd/main.go`
  - Search codebase for "leader node", "replica node", "node type" and remove/refactor
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.6_

- [ ] 9.1 Add migration error messages for deprecated flags
  - When `--type=leader` is used, print error: "Error: --type flag is deprecated. Use --wallet <name> instead. Create a wallet with: poh-blockchain wallet create --name <name>"
  - When `--type=replica` is used, print error: "Error: --type flag is deprecated. Use --wallet <name> for validation, or omit for observer mode."
  - Exit with non-zero code
  - _Requirements: 10.1, 10.2, 10.3_

- [ ] 9.2 Write tests for migration error messages
  - Test `--type=leader` flag is rejected with correct error message
  - Test `--type=replica` flag is rejected with correct error message
  - _Requirements: 10.1, 10.2_

- [ ] 10. Update all integration tests to use wallet-based identity
  - Update `internal/builtin_programs_integration_test.go` to create test wallets instead of using node types
  - Update `internal/staking_integration_test.go` to use wallet-based validators
  - Update all other integration tests that reference `network.LEADER` or `network.REPLICA`
  - _Requirements: 7.5_

- [ ] 11. Implement block production counter increment
  - When a validator produces a block, increment `blocksProducedThisEpoch` in the validator's Validator Record
  - Ensure counter is reset to zero at epoch boundaries
  - _Requirements: 8.4_

- [ ] 11.1 Write unit tests for block production counter
  - Test counter increments when validator produces block
  - Test counter resets at epoch boundary
  - _Requirements: 8.4_

- [ ] 12. Implement logging and observability
  - Log epoch number, active validator count, total stake when computing new schedule
  - Log slot number, local validator FileID when producing block
  - Log slot number, scheduled validator FileID when skipping slot due to missed block
  - Log epoch number, validator count when restoring schedule from Epoch State File
  - Log warning when starting in observer mode
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 12.1 Write tests for logging output
  - Verify correct log messages are emitted for each scenario
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 13. Update demo scripts to use wallet-based validators
  - Update `demo-dpos.sh` to create genesis wallets automatically
  - Replace `--type=leader` flags with `--wallet <name>` flags
  - Add `--wallet-keypair-index` flag for non-interactive keypair selection in automation
  - Update genesis configuration JSON to match new format
  - _Requirements: 7.5_

- [ ] 13.1 Test demo script with 2, 3, and 5 validators
  - Verify stake-weighted slot distribution (validator with 2x stake gets ~2x slots)
  - Verify epoch boundary schedule recalculation
  - Verify missed block handling
  - _Requirements: 2.3, 4.1, 5.1_

- [ ] 14. Write comprehensive integration test for full stake-weighted consensus lifecycle
  - Create `internal/stake_weighted_consensus_test.go`
  - Test: genesis → block production → epoch boundary → schedule recalculation → continued block production
  - Test with 3 validators: stakes 10 Neon, 5 Neon, 2 Neon
  - Verify validator with 10 Neon produces ~58% of blocks, 5 Neon produces ~29%, 2 Neon produces ~12% (within 5% tolerance)
  - Verify schedule changes at epoch boundary when stakes change
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 4.1, 4.2, 4.3, 8.1, 8.2, 8.3, 8.5_

- [ ] 15. Update documentation and README
  - Update README.md to remove references to `--type` flag
  - Add wallet management section to README
  - Add stake-weighted leader schedule explanation
  - Update demo script instructions
  - _Requirements: 7.5, 10.3_

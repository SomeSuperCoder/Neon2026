# Implementation Plan

- [X] 1. Create consolidated binary structure and remove poh-node
 - Remove poh-node binary and migrate its functionality to appropriate binaries
 - Update build system to build new binaries instead of poh-node
 - Update documentation to reflect new binary structure
 - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [-] 2. Implement audit binary with full functionality
- [ ] 2.1 Migrate audit.sh functionality to audit binary
  - Implement network simulation and security verification
  - Add BFT testing scenarios with honest/malicious nodes
  - Add DPoS mechanics and leader election testing
  - Generate comprehensive JSON reports
  - _Requirements: 1.1, 4.1_

- [ ] 2.2 Migrate test-launcher.sh functionality to audit binary
  - Add BFT test suite orchestration
  - Implement test result analysis (analyze-results.sh)
  - Add test execution framework (test_runner.sh)
  - _Requirements: 1.1, 4.1_

- [ ] 2.3 Write unit tests for audit binary
  - Create unit tests for network simulation
  - Write unit tests for BFT verification
  - Test report generation functionality
  - _Requirements: 1.1, 4.1_

- [ ] 3. Enhance neon-wallet with dedicated storage path
- [ ] 3.1 Implement dedicated wallet storage location
  - Store accounts in `~/.config/poh-blockchain/wallets/`
  - Add secure, encrypted storage for account data
  - Create wallet management utilities
  - _Requirements: 2.1, 2.3_

- [ ] 3.2 Update validator to read from neon-wallet storage
  - Modify validator to load accounts from neon-wallet storage location
  - Add account discovery and loading mechanisms
  - Ensure backward compatibility with existing wallets
  - _Requirements: 2.2_

- [ ] 3.3 Write unit tests for wallet storage integration
  - Test wallet storage path enforcement
  - Test account loading from dedicated location
  - Test encryption and security features
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 4. Implement devnet binary with developer account creation
- [ ] 4.1 Migrate devnet.sh functionality to devnet binary
  - Create local validator networks with leadership switching
  - Implement validator startup and management
  - Add network status monitoring and logging
  - _Requirements: 1.4, 4.4_

- [ ] 4.2 Migrate demo-dpos.sh functionality to devnet binary
  - Add DPoS demonstration features
  - Implement stake-weighted scheduling demonstrations
  - Add demo cleanup functionality (stop-demo.sh)
  - _Requirements: 1.4, 4.4_

- [ ] 4.3 Implement developer/tester account creation
  - Create developer account with initial coins on devnet initialization
  - Store seed phrase in `devnet_tester.txt`
  - Allow login with seed phrase for testing
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 4.4 Write unit tests for devnet functionality
  - Test local network creation
  - Test developer account creation
  - Test DPoS demonstration features
  - _Requirements: 1.4, 3.1, 3.2, 3.3, 3.4, 4.4_

- [ ] 5. Implement validator binary with consolidated functionality
- [ ] 5.1 Migrate poh-node validator features to validator binary
  - Extract staking and unstaking operations from poh-node
  - Add delegation to other validators
  - Implement node operations for connecting to devnet
  - _Requirements: 1.3, 4.3_

- [ ] 5.2 Integrate validator-tui into validator binary
  - Add TUI mode to validator binary
  - Migrate terminal UI functionality from validator-tui
  - Create unified command structure for validator operations
  - _Requirements: 1.3, 4.3_

- [ ] 5.3 Remove non-validator features from poh-node
  - Identify and extract non-validator functionality
  - Ensure all validator features are migrated
  - Deprecate poh-node binary
  - _Requirements: 1.6, 4.3_

- [ ] 5.4 Write unit tests for validator operations
  - Test staking and unstaking functionality
  - Test delegation mechanisms
  - Test node operations and devnet connectivity
  - _Requirements: 1.3, 4.3_

- [ ] 6. Implement build binary for project compilation
- [ ] 6.1 Migrate build.sh functionality to build binary
  - Implement compilation of all project binaries
  - Add dependency downloading and management
  - Create minimal, optimized builds
  - _Requirements: 1.5, 4.5_

- [ ] 6.2 Add binary-specific build targets
  - Implement `build audit` command
  - Implement `build neon-wallet` command
  - Implement `build validator` command
  - Implement `build devnet` command
  - Implement `build all` command
  - Implement `build clean` command
  - _Requirements: 1.5, 4.5_

- [ ] 6.3 Write unit tests for build system
  - Test binary compilation
  - Test dependency resolution
  - Test cleanup functionality
  - _Requirements: 1.5, 4.5_

- [ ] 7. Update main.go to use new binary structure
- [ ] 7.1 Remove consolidated functionality from main.go
  - Remove audit-related commands from main.go
  - Remove devnet-related commands from main.go
  - Remove validator-specific commands from main.go
  - Keep only essential blockchain node operations
  - _Requirements: 1.6, 4.5_

- [ ] 7.2 Update command-line interface documentation
  - Update help text to reflect new binary structure
  - Add references to new binaries (audit, neon-wallet, validator, devnet, build)
  - Ensure clear separation of responsibilities
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [ ] 7.3 Write integration tests for binary interactions
  - Test cross-binary workflows (wallet → validator → devnet)
  - Test backward compatibility where applicable
  - Test error handling between binaries
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [ ] 8. Clean up deprecated scripts and update documentation
- [ ] 8.1 Remove deprecated shell scripts
  - Remove audit.sh (functionality migrated to audit binary)
  - Remove devnet.sh (functionality migrated to devnet binary)
  - Remove demo-dpos.sh (functionality migrated to devnet binary)
  - Remove test-launcher.sh (functionality migrated to audit binary)
  - Remove analyze-results.sh (functionality migrated to audit binary)
  - Remove test_runner.sh (functionality migrated to audit binary)
  - Remove run_fibonacci.sh (functionality migrated to QuanticScript examples)
  - Remove stop-demo.sh (functionality migrated to devnet binary)
  - Remove build.sh (functionality migrated to build binary)
  - _Requirements: 1.6_

- [ ] 8.2 Update project documentation
  - Update README.md with new binary structure
  - Update CLI usage guides
  - Update developer documentation
  - Update CI/CD pipeline documentation
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6_

- [ ] 8.3 Update CI/CD pipelines
  - Update build scripts to use new binaries
  - Update test scripts to use audit binary
  - Update deployment scripts for new binary structure
  - _Requirements: 1.6_

- [ ] 8.4 Write acceptance tests for binary consolidation
  - Verify no functionality loss during consolidation
  - Test backward compatibility where applicable
  - Test performance of new binaries vs old scripts
  - _Requirements: 1.6_

# Implementation Plan

- [x] 1. Create audit.sh script with comprehensive testing
  - Implement command-line argument parsing (--duration, --validators, --ci, --help)
  - Create JSON report initialization and finalization functions
  - Implement cleanup and signal handling (trap INT/TERM)
  - _Requirements: 1.1, 1.3, 1.4, 4.1, 4.2, 4.3, 4.4_

- [x] 1.1 Implement basic consensus test phase
  - Start 1 leader + 3 honest replicas
  - Monitor block production for specified duration
  - Collect metrics (total blocks, blocks/validator, consistency)
  - Add results to JSON report
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [x] 1.2 Implement BFT with tolerance test phase
  - Start 1 leader + 4 honest + 1 malicious replica
  - Monitor block production and rejection
  - Collect BFT metrics (blocks rejected, malicious actions)
  - Add results to JSON report
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [x] 1.3 Implement BFT without tolerance test phase
  - Start 1 leader + 2 honest + 2 malicious replicas
  - Monitor network behavior under insufficient BFT
  - Collect consistency metrics
  - Add results to JSON report
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [x] 1.4 Implement DPoS lifecycle test phase
  - Start N validators from genesis
  - Simulate delegation, epoch boundaries, rewards, slashing
  - Collect DPoS metrics (validator stats, lifecycle completion)
  - Add results to JSON report
  - _Requirements: 1.1, 1.2, 4.1, 4.2_

- [x] 1.5 Implement audit summary and exit code logic
  - Calculate overall pass/fail status
  - Generate final summary section in JSON
  - Display console summary with pass/fail verdict
  - Exit with appropriate code (0=pass, 1=fail, 2=config error)
  - _Requirements: 1.2, 1.3, 4.3_

- [x] 2. Create devnet.sh script for local development network
  - Implement command parsing (start, stop, restart, status, logs, clean)
  - Implement command-line option parsing (--port, --db-dir, --log-dir, --help)
  - Create directory structure management (devnet-data/, logs/, pids/)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 2.1 Implement devnet start command
  - Build poh-node binary
  - Start leader node as background process
  - Start N-1 replica nodes as background processes
  - Store PIDs in devnet-data/pids/
  - Display network configuration (addresses, ports, log locations)
  - _Requirements: 2.1, 2.2, 2.4, 2.5_

- [x] 2.2 Implement devnet stop command
  - Read PIDs from devnet-data/pids/
  - Send TERM signal to all processes
  - Wait for graceful shutdown (2 seconds)
  - Force kill if still running
  - Clean up PID files
  - _Requirements: 2.3_

- [x] 2.3 Implement devnet status command
  - Check if PID files exist
  - Verify processes are running
  - Display validator status (running/stopped)
  - Show block counts from databases
  - _Requirements: 2.4_

- [x] 2.4 Implement devnet logs command
  - Display logs for specific validator or all validators
  - Support tail -f style following
  - Handle missing log files gracefully
  - _Requirements: 2.4_

- [x] 2.5 Implement devnet restart and clean commands
  - Restart: stop + start with same configuration
  - Clean: stop + remove all devnet-data and logs
  - Prompt for confirmation on clean command
  - _Requirements: 2.3_

- [x] 3. Update stop-demo.sh to handle new scripts
  - Add logic to stop legacy tmux sessions (poh-demo, poh-bft-demo)
  - Add logic to call devnet.sh stop if devnet is running
  - Kill any stray poh-node processes
  - _Requirements: 3.3_

- [x] 4. Update analyze-results.sh for new structure
  - Add --db-dir option to analyze devnet databases
  - Update default behavior to check both audit and devnet locations
  - Preserve existing analysis functionality
  - _Requirements: 3.2_

- [x] 5. Remove old demo scripts
  - Delete demo.sh
  - Delete demo-automated.sh
  - Delete demo-bft.sh
  - Delete demo-dpos.sh
  - _Requirements: 3.1, 3.4_

- [x] 6. Update documentation
  - Update README.md Quick Start section with new scripts
  - Update README.md Usage section to reference audit.sh and devnet.sh
  - _Requirements: 3.5_

- [x] 6.1 Update docs/guides/demo.md
  - Replace references to old demo scripts with devnet.sh
  - Add audit.sh usage examples
  - Update tmux navigation section (remove if not applicable)
  - Update network configuration section
  - _Requirements: 3.5_

- [x] 6.2 Update docs/guides/dpos-demo.md
  - Replace demo-dpos.sh references with audit.sh
  - Update usage examples
  - Update output format section
  - _Requirements: 3.5_

- [x] 6.3 Update docs/testing/automated-testing.md
  - Replace demo-automated.sh references with audit.sh
  - Update CI/CD integration examples
  - Update JSON output format documentation
  - _Requirements: 3.5_

- [ ] 7. Validate implementation
  - Run audit.sh and verify all phases pass
  - Test devnet.sh start/stop/restart/status/logs/clean commands
  - Verify state persistence across devnet restarts
  - Test analyze-results.sh with both audit and devnet databases
  - Verify no broken documentation links
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.1, 2.2, 2.3, 2.4, 2.5, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5_

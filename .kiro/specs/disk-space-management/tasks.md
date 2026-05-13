# Implementation Plan

- [ ] 1. Create disk space management package structure
 - Create `internal/diskmonitor/` package directory
 - Define core interfaces for resource monitoring and cleanup
 - Set up configuration structures for resource limits
 - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 2. Implement Resource Monitor component
- [ ] 2.1 Create resource monitoring types and interfaces
  - Define `ResourceMonitor` struct with tracking capabilities
  - Implement `ProcessTracker` for per-process resource monitoring
  - Create `DiskUsage` struct to track file system usage
  - _Requirements: 1.1, 1.3, 3.1, 3.2_

- [ ] 2.2 Implement disk usage tracking
  - Write functions to monitor SQLite database file sizes
  - Track BadgerDB directory sizes
  - Monitor log file growth patterns
  - _Requirements: 1.1, 1.4, 3.2_

- [ ] 2.3 Implement process registration and monitoring
  - Create `RegisterProcess()` method to track new processes
  - Implement `CheckThresholds()` for limit validation
  - Add polling mechanism with configurable intervals
  - _Requirements: 1.1, 1.3, 3.3_

- [ ]* 2.4 Write unit tests for resource monitoring
  - Create tests for disk usage tracking functions
  - Write tests for process registration and threshold checking
  - Test polling interval configuration
  - _Requirements: 1.1, 1.3, 3.1, 3.2, 3.3_

- [ ] 3. Implement Log Rotator component
- [ ] 3.1 Create log rotation types and configuration
  - Define `LogRotationPolicy` struct with size/time limits
  - Implement `LogRotator` struct with compression support
  - Create configuration loading from JSON/YAML
  - _Requirements: 2.2, 4.2_

- [ ] 3.2 Implement log file rotation logic
  - Write `RotateLog()` function for size-based rotation
  - Implement `CompressOldLogs()` with gzip compression
  - Create `CleanupExpiredLogs()` for age-based cleanup
  - _Requirements: 2.2, 4.2, 5.3_

- [ ] 3.3 Integrate log rotation with existing logging system
  - Modify existing loggers to use rotation capabilities
  - Add log rotation to devnet and demo scripts
  - Implement log statistics reporting
  - _Requirements: 2.2, 4.2, 5.3_

- [ ]* 3.4 Write unit tests for log rotation
  - Test size-based rotation with mock files
  - Test compression functionality
  - Test age-based cleanup logic
  - _Requirements: 2.2, 4.2, 5.3_

- [ ] 4. Implement Limit Enforcer component
- [ ] 4.1 Create resource limit configuration system
  - Define `ResourceLimit` struct with warning/enforcement actions
  - Implement `LimitEnforcer` struct for limit management
  - Create configuration validation and default values
  - _Requirements: 1.2, 1.5, 5.1, 5.2_

- [ ] 4.2 Implement limit checking and enforcement
  - Write `CheckLimit()` function for resource validation
  - Implement `EnforceLimit()` with graceful degradation
  - Create warning, throttle, and stop phases
  - _Requirements: 1.2, 1.5, 5.4, 5.5_

- [ ] 4.3 Integrate limit enforcement with devnet scripts
  - Modify devnet startup to check disk space limits
  - Add limit checking to node startup process
  - Implement emergency cleanup procedures
  - _Requirements: 1.2, 1.5, 5.4, 5.5_

- [ ]* 4.4 Write unit tests for limit enforcement
  - Test warning threshold logic
  - Test throttle and stop enforcement
  - Test configuration validation
  - _Requirements: 1.2, 1.5, 5.4, 5.5_

- [ ] 5. Implement Cleanup Manager component
- [ ] 5.1 Create cleanup rule system
  - Define `CleanupRule` struct with pattern/age/size matching
  - Implement `CleanupManager` struct with rule processing
  - Create safe cleanup with file usage verification
  - _Requirements: 2.1, 2.3, 4.1, 4.3_

- [ ] 5.2 Implement temporary file cleanup
  - Write `CleanupTemporaryFiles()` function
  - Implement age-based cleanup for test artifacts
  - Create safe deletion with verification
  - _Requirements: 2.1, 2.3, 4.3, 5.3_

- [ ] 5.3 Implement database pruning
  - Create `PruneOldDatabases()` function
  - Implement size-based cleanup (oldest first)
  - Add dry-run mode for safety verification
  - _Requirements: 2.1, 4.1, 4.4_

- [ ] 5.4 Integrate cleanup with test framework
  - Modify test cleanup to use CleanupManager
  - Add automatic cleanup between test runs
  - Implement cleanup statistics reporting
  - _Requirements: 2.4, 4.4, 5.3_

- [ ]* 5.5 Write unit tests for cleanup manager
  - Test temporary file cleanup with mock files
  - Test database pruning logic
  - Test safe deletion verification
  - _Requirements: 2.1, 2.3, 4.1, 4.3, 4.4_

- [ ] 6. Implement Foreground Process Management
- [ ] 6.1 Create foreground process runner
  - Implement `RunInForeground()` function with terminal output
  - Create process grouping for coordinated startup/shutdown
  - Add terminal attachment/detachment support
  - _Requirements: 6.1, 6.2, 6.3_

- [ ] 6.2 Modify devnet scripts for foreground operation
  - Update devnet.sh to run nodes in foreground by default
  - Add `--foreground` flag support
  - Implement terminal log display for multiple nodes
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ] 6.3 Implement terminal log display system
  - Create color-coded output per node type
  - Implement log level filtering (INFO, DEBUG, ERROR, WARN)
  - Add real-time log streaming with configurable buffer
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [ ] 6.4 Integrate foreground management with resource monitoring
  - Connect process monitoring to foreground process runner
  - Add resource tracking to foreground processes
  - Implement cleanup hooks for foreground processes
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ]* 6.5 Write unit tests for foreground process management
  - Test foreground process startup and shutdown
  - Test terminal output piping
  - Test process grouping functionality
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ] 7. Integrate disk space management with existing systems
- [ ] 7.1 Update devnet scripts with resource limits
  - Add disk space checking to devnet.sh startup
  - Implement cleanup hooks in devnet stop/clean commands
  - Add log rotation to devnet logging
  - _Requirements: 1.1, 1.2, 2.1, 2.2_

- [ ] 7.2 Update demo scripts with cleanup integration
  - Modify demo-dpos.sh to use CleanupManager
  - Add resource limit warnings to demo startup
  - Implement automatic cleanup after demo completion
  - _Requirements: 2.1, 2.3, 5.3, 5.4_

- [ ] 7.3 Update test framework with automatic cleanup
  - Modify test helpers to use CleanupManager
  - Add resource monitoring to integration tests
  - Implement test state cleanup between runs
  - _Requirements: 2.4, 4.4, 5.3_

- [ ] 7.4 Create CLI commands for manual management
  - Implement `poh-node disk-stats` command
  - Create `poh-node cleanup` command with options
  - Add `poh-node log-rotate` command for manual rotation
  - _Requirements: 3.5, 4.5, 5.5_

- [ ]* 7.5 Write integration tests for full system
  - Test devnet with resource limits enabled
  - Test demo scripts with automatic cleanup
  - Test CLI commands functionality
  - _Requirements: All requirements_

- [ ] 8. Documentation and configuration
- [ ] 8.1 Create configuration file templates
  - Create default `disk-monitor.yaml` configuration
  - Implement configuration loading from multiple sources
  - Add environment variable support for configuration
  - _Requirements: 5.1, 5.2_

- [ ] 8.2 Update project documentation
  - Add disk space management section to README
  - Create usage guide for resource limits
  - Document cleanup policies and configuration
  - _Requirements: 3.5, 4.5, 5.5_

- [ ] 8.3 Add examples and demos
  - Create example configuration files
  - Add demo script for disk space management features
  - Create troubleshooting guide for common issues
  - _Requirements: 3.5, 4.5, 5.5_

- [ ]* 8.4 Write documentation tests
  - Test configuration file examples
  - Verify documentation commands work
  - Test example scripts
  - _Requirements: 3.5, 4.5, 5.5_
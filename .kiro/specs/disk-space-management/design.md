# Design Document: Disk Space Management

## Overview

This design addresses the critical issue of uncontrolled disk space consumption by development tools, particularly devnet scripts and demo utilities. The recent incident where a devnet script consumed 62GB overnight demonstrates the need for systematic resource management. The solution focuses on configurable limits, automated cleanup, and proactive monitoring to prevent storage exhaustion.

## Architecture

The disk space management system will be integrated into the existing PoH blockchain infrastructure with these components:

1. **Resource Monitor**: Background service that tracks disk usage across all blockchain components
2. **Limit Enforcer**: Applies configurable thresholds and triggers cleanup actions
3. **Log Rotator**: Manages log file lifecycle with compression and archival
4. **Cleanup Manager**: Handles temporary file removal and database pruning
5. **Configuration System**: Centralized settings for resource limits and retention policies

### System Integration Points

- **Devnet Scripts**: Enhanced with resource monitoring and cleanup hooks
- **Node Processes**: Report disk usage and respond to cleanup signals
- **Test Framework**: Automatic cleanup between test runs
- **CLI Tools**: Commands for manual cleanup and monitoring

## Components and Interfaces

### 1. Resource Monitor Component

**Purpose**: Continuously monitor disk usage across all blockchain processes

**Interfaces**:
- `StartMonitoring()`: Begin tracking resource usage
- `GetCurrentUsage()`: Return current disk, memory, and CPU usage
- `RegisterProcess(pid, expectedUsage)`: Track a new process
- `CheckThresholds()`: Verify all processes are within limits

**Implementation**:
- Uses Go's `os/exec` and system calls to monitor processes
- Tracks SQLite database files, BadgerDB directories, and log files
- Configurable polling interval (default: 30 seconds)

### 2. Log Rotator Component

**Purpose**: Manage log file lifecycle with configurable retention

**Interfaces**:
- `RotateLog(filePath)`: Rotate log file if it exceeds size limit
- `CompressOldLogs()`: Compress logs older than retention period
- `CleanupExpiredLogs()`: Remove logs beyond maximum retention
- `GetLogStats()`: Report current log storage usage

**Implementation**:
- Size-based rotation (default: 100MB per log file)
- Time-based retention (default: 7 days)
- Compression using gzip for old logs
- 1GB total limit across all logs with LRU eviction

### 3. Limit Enforcer Component

**Purpose**: Enforce configurable resource limits

**Interfaces**:
- `CheckLimit(process, resource, usage)`: Verify usage against limit
- `EnforceLimit(process, resource)`: Apply limit enforcement
- `GetLimit(resource)`: Return current limit configuration
- `SetLimit(resource, value)`: Update limit configuration

**Implementation**:
- Configurable limits per resource type:
  - Total disk usage: 20GB default (devnet data threshold)
  - Per-process disk: 2GB default  
  - Log storage: 1GB default
  - Temporary files: 5GB default
- Graceful degradation when limits approached
- Hard stop when critical limits exceeded

### 4. Cleanup Manager Component

**Purpose**: Automated cleanup of temporary and expired files

**Interfaces**:
- `CleanupTemporaryFiles()`: Remove temp files older than threshold
- `PruneOldDatabases()`: Remove old database files
- `CleanupTestState()`: Clean test artifacts between runs
- `GetCleanupStats()`: Report cleanup effectiveness

**Implementation**:
- Age-based cleanup (default: 24 hours for temp files)
- Size-based cleanup (oldest first when storage limits exceeded)
- Safe cleanup with verification (don't delete active files)
- Dry-run mode for safety verification

## Data Models

### ResourceLimit Configuration
```go
type ResourceLimit struct {
    ResourceType string  // "disk", "memory", "cpu"
    LimitValue   int64   // Limit in bytes or percentage
    WarningThreshold int64 // Warning level (e.g., 80% of limit)
    EnforcementAction string // "warn", "throttle", "stop"
}
```

### LogRotationPolicy
```go
type LogRotationPolicy struct {
    MaxFileSize    int64  // Maximum size per log file (bytes)
    MaxTotalSize   int64  // Maximum total log storage (bytes)
    RetentionDays  int    // Days to keep logs
    CompressionEnabled bool // Whether to compress old logs
    RotationStrategy string // "size", "time", "both"
}
```

### CleanupRule
```go
type CleanupRule struct {
    Pattern        string // File pattern to match
    MaxAgeHours    int    // Maximum age in hours
    MaxSizeBytes   int64  // Maximum total size
    Action         string // "delete", "compress", "archive"
    SafetyCheck    bool   // Verify file is not in use
}
```

## Error Handling

### Graceful Degradation
1. **Warning Phase**: When usage reaches 80% of limit, log warnings
2. **Throttle Phase**: At 90% usage, throttle non-essential operations
3. **Stop Phase**: At 100% usage, stop processes gracefully
4. **Emergency Phase**: At 110% usage, force cleanup and restart

### Recovery Procedures
1. **Automatic Recovery**: Attempt cleanup and restart within limits
2. **Manual Intervention**: Provide clear instructions for manual cleanup
3. **Fallback Mode**: Continue with reduced functionality if cleanup fails
4. **Audit Trail**: Log all cleanup actions for review

### Error Types
- `ResourceLimitExceededError`: When a limit is breached
- `CleanupFailedError`: When automated cleanup cannot proceed
- `MonitoringError`: When resource monitoring fails
- `ConfigurationError`: When limit configuration is invalid

## Testing Strategy

### Unit Tests
- Resource limit validation and enforcement
- Log rotation logic with various file sizes
- Cleanup rule application and safety checks
- Configuration loading and validation

### Integration Tests
- End-to-end devnet script with resource limits
- Log rotation during extended operation
- Cleanup during test suite execution
- Limit enforcement across multiple processes

### System Tests
- Simulated disk exhaustion scenarios
- Recovery from exceeded limits
- Performance under resource constraints
- Configuration changes during runtime

### Test Data Management
- Use temporary directories for all tests
- Clean up test artifacts automatically
- Verify no residual files after tests
- Test with realistic file sizes and patterns

## Additional Design Considerations

### 6. Foreground Process Management

**Purpose**: Run devnet nodes in foreground with terminal log display instead of background processes

**Interfaces**:
- `RunInForeground(processConfig)`: Run process with stdout/stderr piped to terminal
- `AttachToTerminal(process)`: Attach existing process to terminal for log display
- `DetachFromTerminal(process)`: Detach process from terminal (optional)
- `GetProcessOutput()`: Retrieve real-time process output

**Implementation**:
- Modify devnet scripts to run nodes with `os/exec` Command struct
- Use `cmd.Stdout = os.Stdout` and `cmd.Stderr = os.Stderr` for terminal output
- Implement process grouping for coordinated startup/shutdown
- Add terminal multiplexing support for multiple nodes (tmux/screen optional)
- Provide flag `--foreground` to devnet scripts (default: true)

### 7. Terminal Log Display System

**Purpose**: Display real-time logs from multiple nodes in organized terminal view

**Interfaces**:
- `DisplayLogs(nodeID, logLevel)`: Display filtered logs for specific node
- `AggregateLogs()`: Combine and display logs from all nodes
- `ToggleVerbosity()`: Switch between detailed and summary log views
- `PauseLogDisplay()`: Temporarily pause log output

**Implementation**:
- Color-coded output per node type (leader: blue, replica: green, malicious: red)
- Log level filtering (INFO, DEBUG, ERROR, WARN)
- Real-time log streaming with configurable buffer size
- Optional log file persistence alongside terminal display
- Keyboard shortcuts for log control (Ctrl+C to stop, Space to pause)

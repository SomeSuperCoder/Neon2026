# Requirements Document

## Introduction

The project currently has multiple demo scripts (`demo.sh`, `demo-automated.sh`, `demo-bft.sh`, `demo-dpos.sh`) that serve overlapping purposes. This creates confusion and maintenance overhead. The goal is to consolidate these into exactly two scripts: one for comprehensive testing/auditing and one for running a local development network.

## Glossary

- **Demo Script**: A shell script that launches blockchain nodes for demonstration or testing purposes
- **Audit Script**: A comprehensive testing script that validates all blockchain functionality and produces detailed reports
- **Devnet Script**: A script that launches a local development network for testing wallets, block explorers, and other tools
- **Validator Node**: A blockchain node that participates in consensus and block production
- **BFT**: Byzantine Fault Tolerance - the ability to maintain consensus despite malicious nodes
- **DPoS**: Delegated Proof of Stake - the consensus mechanism used by the blockchain

## Requirements

### Requirement 1

**User Story:** As a developer, I want a single comprehensive audit script, so that I can validate all blockchain functionality in one command

#### Acceptance Criteria

1. WHEN the developer executes the audit script, THE System SHALL run all blockchain test scenarios including basic consensus, BFT testing, and DPoS lifecycle
2. WHEN the audit script completes, THE System SHALL generate a detailed JSON report with test results and metrics
3. WHEN the audit script encounters failures, THE System SHALL exit with a non-zero status code
4. WHEN the audit script runs, THE System SHALL clean up all test artifacts automatically after completion
5. WHERE the developer specifies test parameters, THE System SHALL accept command-line arguments for test duration and node counts

### Requirement 2

**User Story:** As a developer, I want a simple devnet launcher script, so that I can quickly start a local network for development and testing

#### Acceptance Criteria

1. WHEN the developer executes the devnet script, THE System SHALL start a configurable number of validator nodes
2. WHEN the devnet is running, THE System SHALL keep nodes running in the background until explicitly stopped
3. WHEN the developer stops the devnet, THE System SHALL cleanly shut down all validator nodes
4. WHERE the developer needs to inspect node logs, THE System SHALL provide clear instructions for accessing individual node logs
5. WHEN the devnet starts, THE System SHALL display network configuration including all node addresses and ports

### Requirement 3

**User Story:** As a developer, I want old demo scripts removed, so that there is no confusion about which scripts to use

#### Acceptance Criteria

1. WHEN the consolidation is complete, THE System SHALL have exactly two demo scripts in the project root
2. WHEN old demo scripts are removed, THE System SHALL preserve the `analyze-results.sh` utility for manual result inspection
3. WHEN old demo scripts are removed, THE System SHALL preserve the `stop-demo.sh` utility for stopping running nodes
4. THE System SHALL remove `demo.sh`, `demo-automated.sh`, `demo-bft.sh`, and `demo-dpos.sh` from the project root
5. THE System SHALL ensure no documentation references the removed scripts

### Requirement 4

**User Story:** As a CI/CD system, I want the audit script to be machine-parseable, so that I can integrate test results into automated pipelines

#### Acceptance Criteria

1. WHEN the audit script completes, THE System SHALL output results in JSON format
2. WHEN the audit script runs, THE System SHALL include timestamps and test metadata in the output
3. WHEN the audit script detects failures, THE System SHALL include detailed failure information in the JSON output
4. THE System SHALL save the JSON report to a predictable location with a timestamped filename
5. WHEN the audit script runs in CI mode, THE System SHALL suppress interactive prompts and colored output

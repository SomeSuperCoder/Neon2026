# Requirements Document

## Introduction

This feature consolidates the numerous scripts and binaries in the PoH Blockchain project into a clean, focused set of executables with clearly defined responsibilities. The goal is to reduce confusion, eliminate redundancy, and ensure each remaining binary has a single, well-defined purpose.

## Glossary

- **audit**: A binary that runs network simulations and verifies security properties, attack resistance, DPoS mechanics, leader elections, and Byzantine Fault Tolerance
- **neon-wallet**: A binary for creating and managing accounts, storing them in a special path, and sending Neon tokens by address
- **validator**: A binary for staking, unstaking (including delegation to others), spinning up nodes that connect to devnet, and loading Neon accounts from the neon-wallet storage location
- **devnet**: A binary that creates a local set of validators with leadership switching according to the existing leader election system
- **build script**: A binary build script for compiling the project

## Requirements

### Requirement 1

**User Story:** As a developer, I want a clear, minimal set of binaries so I can understand what each tool does without confusion

#### Acceptance Criteria

1. WHEN the developer runs the audit binary, THE audit binary SHALL run network simulations and verify security properties
2. WHEN the developer runs the neon-wallet binary, THE neon-wallet binary SHALL create and manage accounts
3. WHEN the developer runs the validator binary, THE validator binary SHALL handle staking, unstaking, and node operations
4. WHEN the developer runs the devnet binary, THE devnet binary SHALL create a local validator network
5. WHEN the developer runs the build script, THE build script SHALL compile the project binaries
6. WHERE any binary exists beyond the five specified, THE system SHALL remove or consolidate it

### Requirement 2

**User Story:** As a developer, I want neon-wallet to store accounts in a special path so I can easily manage and locate them

#### Acceptance Criteria

1. WHEN neon-wallet creates an account, THE neon-wallet SHALL store it in a dedicated, well-documented storage location
2. WHEN validator needs to load an account, THE validator SHALL read from the neon-wallet storage location
3. WHERE neon-wallet stores account data, THE system SHALL use secure, encrypted storage

### Requirement 3

**User Story:** As a developer/tester, I want a devnet that creates a test account with coins so I can experiment with staking and validation

#### Acceptance Criteria

1. WHEN devnet initializes, THE devnet SHALL create a developer/tester account with initial coins
2. WHEN devnet creates the test account, THE system SHALL store the seed phrase in 'devnet_tester.txt'
3. WHERE the developer uses the seed phrase, THE system SHALL allow login to the test account
4. WHERE the test account exists, THE developer SHALL use it for staking and spinning up validators alongside devnet

### Requirement 4

**User Story:** As a developer, I want each binary to have strictly defined capabilities so there's no overlap or confusion

#### Acceptance Criteria

1. WHILE audit runs, THE audit SHALL only perform security verification and network simulation
2. WHILE neon-wallet runs, THE neon-wallet SHALL only handle account creation, management, and token transfers
3. WHILE validator runs, THE validator SHALL only handle staking, unstaking, delegation, and node operations
4. WHILE devnet runs, THE devnet SHALL only create and manage local validator networks
5. WHERE any binary attempts operations outside its defined scope, THE system SHALL prevent or redirect those operations
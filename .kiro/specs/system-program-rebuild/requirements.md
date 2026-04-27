# Requirements Document

## Introduction

The System Program needs a clean rebuild focused on four core operations: creating files owned by the System Program (tx_manager), transferring Neon from files it owns to any other file, burning Neon, and closing files it owns. This spec removes all legacy complexity and implements only what's needed for production-ready account management.

## Glossary

- **System_Program**: The built-in program responsible for account management, identified by FileID `0x00...01`
- **File**: The fundamental unit of on-chain state containing balance, data, TxManager, and metadata
- **FileID**: A unique 32-byte identifier for a file
- **TxManager**: The FileID of the program that owns and manages a file
- **Neon**: The native token unit of the blockchain, stored as i64 balance in files
- **Storage Cost**: The minimum balance required to sustain a file's data, calculated exponentially based on data size
- **Transfer**: An atomic operation that moves balance from one file to another while respecting storage cost constraints
- **Burn**: Permanently removing Neon from circulation by reducing a file's balance
- **Close**: Deleting a file and reclaiming its balance, only possible when the file has no data

## Requirements

### Requirement 1

**User Story:** As a blockchain user, I want the System Program to create new files that it owns, so that I can establish accounts for storing balance and data.

#### Acceptance Criteria

1. WHEN the System_Program receives a CREATE_FILE instruction, THE System_Program SHALL create a new file with TxManager set to the System_Program's FileID
2. WHEN creating a file, THE System_Program SHALL transfer the specified initial balance from itself to the new file
3. WHEN the initial balance is negative or would cause the System_Program's balance to violate storage cost constraints, THE System_Program SHALL reject the instruction with an error
4. WHEN a file is created successfully, THE System_Program SHALL return SUCCESS (0x00)
5. THE CREATE_FILE instruction format SHALL be: [type:u8(1)][owner:FileID(32)][initial_balance:i64(8)] totaling 41 bytes

### Requirement 2

**User Story:** As a blockchain user, I want the System Program to transfer Neon from files it owns to any other file, so that I can move funds between accounts.

#### Acceptance Criteria

1. WHEN the System_Program receives a TRANSFER instruction, THE System_Program SHALL transfer the specified amount from the source file to the destination file
2. WHEN the source file's TxManager is not the System_Program, THE System_Program SHALL reject the instruction with ERROR_UNAUTHORIZED
3. WHEN the transfer amount is zero or negative, THE System_Program SHALL reject the instruction with ERROR_INVALID_AMOUNT
4. WHEN the transfer would violate the source file's storage cost constraint, THE System_Program SHALL reject the instruction with ERROR_INSUFFICIENT_BALANCE
5. THE TRANSFER instruction format SHALL be: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)] totaling 73 bytes

### Requirement 3

**User Story:** As a blockchain operator, I want the System Program to burn Neon from files it owns, so that I can permanently remove tokens from circulation.

#### Acceptance Criteria

1. WHEN the System_Program receives a BURN instruction, THE System_Program SHALL reduce the specified file's balance by the burn amount
2. WHEN the file's TxManager is not the System_Program, THE System_Program SHALL reject the instruction with ERROR_UNAUTHORIZED
3. WHEN the burn amount is zero or negative, THE System_Program SHALL reject the instruction with ERROR_INVALID_AMOUNT
4. WHEN the burn would violate the file's storage cost constraint, THE System_Program SHALL reject the instruction with ERROR_INSUFFICIENT_BALANCE
5. THE BURN instruction format SHALL be: [type:u8(1)][file:FileID(32)][amount:i64(8)] totaling 41 bytes

### Requirement 4

**User Story:** As a blockchain user, I want the System Program to close files it owns, so that I can reclaim balance from accounts I no longer need.

#### Acceptance Criteria

1. WHEN the System_Program receives a CLOSE_FILE instruction, THE System_Program SHALL delete the specified file and transfer its remaining balance to the destination file
2. WHEN the file's TxManager is not the System_Program, THE System_Program SHALL reject the instruction with ERROR_UNAUTHORIZED
3. WHEN the file has non-zero data length, THE System_Program SHALL reject the instruction with ERROR_FILE_HAS_DATA
4. WHEN the file is successfully closed, THE System_Program SHALL transfer all remaining balance to the destination file before deletion
5. THE CLOSE_FILE instruction format SHALL be: [type:u8(1)][file_to_close:FileID(32)][destination:FileID(32)] totaling 65 bytes

### Requirement 5

**User Story:** As a developer, I want clear error codes for all failure scenarios, so that I can handle errors appropriately in client applications.

#### Acceptance Criteria

1. THE System_Program SHALL return SUCCESS (0x00) when an instruction completes successfully
2. THE System_Program SHALL return ERROR_INVALID_INSTRUCTION (0x1FFF) when instruction data is malformed or has incorrect length
3. THE System_Program SHALL return ERROR_INSUFFICIENT_BALANCE (0x1000) when a balance operation would violate storage cost constraints
4. THE System_Program SHALL return ERROR_INVALID_AMOUNT (0x1005) when an amount parameter is zero or negative
5. THE System_Program SHALL return ERROR_UNAUTHORIZED (0x1004) when attempting to operate on a file not owned by the System_Program
6. THE System_Program SHALL return ERROR_FILE_HAS_DATA (0x1006) when attempting to close a file with non-zero data length

### Requirement 6

**User Story:** As a blockchain operator, I want the System Program to validate all ownership before performing operations, so that programs cannot manipulate files they don't own.

#### Acceptance Criteria

1. WHEN processing any instruction that modifies a file, THE System_Program SHALL verify the file's TxManager equals the System_Program's FileID
2. WHEN a file is not owned by the System_Program, THE System_Program SHALL reject the operation before making any state changes
3. WHEN creating a new file, THE System_Program SHALL set the TxManager to its own FileID
4. WHEN transferring balance, THE System_Program SHALL only allow transfers from files it owns
5. WHEN closing a file, THE System_Program SHALL only allow closing files it owns

### Requirement 7

**User Story:** As a blockchain operator, I want all System Program operations to respect storage cost constraints, so that files never become invalid due to insufficient balance.

#### Acceptance Criteria

1. WHEN transferring balance from a file, THE System_Program SHALL ensure the remaining balance covers the file's storage cost
2. WHEN burning Neon from a file, THE System_Program SHALL ensure the remaining balance covers the file's storage cost
3. WHEN creating a file with initial balance, THE System_Program SHALL ensure the System_Program's remaining balance covers its own storage cost
4. WHILE a file has data stored, THE System_Program SHALL never allow balance to drop below CalculateStorageCost(data_size)
5. IF a file has zero data size, THE System_Program SHALL allow balance to be reduced to zero

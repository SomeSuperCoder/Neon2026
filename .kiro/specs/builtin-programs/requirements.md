# Requirements Document

## Introduction

This document specifies the requirements for a token standard on the blockchain that distinguishes between Neon (the native network coin) and custom tokens. Neon is stored in the `balance` field of files and is used for transaction fees and storage rent. Custom tokens store their balance information in the file's `data` field, with each token account referencing an owner's Neon account. The storage rent model ensures that files with higher Neon balances can store more data, creating an economic constraint on state growth.

Both the System_Program and Token_Program are implemented in QuanticScript, the blockchain's smart contract language.

## Glossary

- **Neon**: The native network coin stored in the `balance` field of all files, used for transaction fees and storage rent
- **Token**: A custom fungible asset created by a token program, with balances stored in file `data` fields
- **Token_Program**: An executable file that manages token minting, transfers, and account creation
- **Token_Account**: A file managed by a Token_Program containing token balance and owner information in its `data` field
- **Owner_Account**: A user's Neon account file managed by the System_Program, controlled by their private key
- **Storage_Rent**: The Neon balance required to maintain a file's data storage on-chain
- **Mint_Account**: A special token account that defines a token's properties including supply and decimals
- **Authority**: A public key or account that has permission to perform privileged operations on a token
- **System_Program**: The built-in program that manages Neon accounts and basic operations
- **Associated_Token_Account**: A deterministically derived token account address for a given owner and token mint
- **QuanticScript**: The smart contract programming language used to implement blockchain programs
- **Runtime**: The execution environment that processes QuanticScript bytecode

## Requirements

### Requirement 1

**User Story:** As a blockchain user, I want to hold Neon in my account, so that I can pay transaction fees and rent storage space

#### Acceptance Criteria

1. THE System_Program SHALL create user account files with Neon balance stored in the `balance` field
2. THE System_Program SHALL allow Neon transfers between accounts by modifying the `balance` field
3. THE system SHALL deduct transaction fees from the fee payer's Neon balance
4. THE system SHALL enforce that Neon balance cannot be negative
5. THE user account File SHALL have an empty `data` field when managed by the System_Program

### Requirement 2

**User Story:** As a blockchain developer, I want storage rent to be proportional to file balance, so that state growth is economically constrained

#### Acceptance Criteria

1. THE system SHALL calculate maximum allowed data size based on the file's Neon balance
2. THE system SHALL use exponential cost growth where larger files require disproportionately more Neon
3. WHEN a file's Neon balance decreases, THE system SHALL verify the remaining balance covers current data size
4. WHEN a file's data size increases, THE system SHALL verify the Neon balance covers the new storage cost
5. THE system SHALL reject operations that would violate the storage rent constraint

### Requirement 3

**User Story:** As a blockchain developer, I want the System_Program and Token_Program to be implemented in QuanticScript, so that they are verifiable and consistent with the blockchain's execution model

#### Acceptance Criteria

1. THE System_Program SHALL be implemented in QuanticScript and compiled to bytecode
2. THE Token_Program SHALL be implemented in QuanticScript and compiled to bytecode
3. THE System_Program SHALL execute within the Runtime environment
4. THE Token_Program SHALL execute within the Runtime environment
5. THE QuanticScript implementations SHALL be deterministic and reproducible
6. THE System_Program SHALL be available in the blockchain state at genesis
7. THE Token_Program SHALL be available in the blockchain state at genesis

### Requirement 4

**User Story:** As a token creator, I want to deploy a token program, so that I can create and manage custom fungible tokens

#### Acceptance Criteria

1. THE Token_Program SHALL be an executable file with bytecode in its `data` field
2. THE Token_Program SHALL have its `tx_manager` field set to the Runtime program
3. THE Token_Program SHALL require Neon balance equal to or greater than the storage cost for its bytecode size
4. THE Token_Program SHALL be immutable once deployed
5. THE Token_Program SHALL process instructions for token operations including mint creation, account creation, and transfers

### Requirement 5

**User Story:** As a token creator, I want to create a mint account, so that I can define a new token with specific properties

#### Acceptance Criteria

1. THE Token_Program SHALL create a Mint_Account file with token metadata in its `data` field
2. THE Mint_Account data SHALL include total supply, decimals, and mint authority fields
3. THE Mint_Account data SHALL include a freeze authority field for optional account freezing
4. THE Mint_Account SHALL have its `tx_manager` field set to the Token_Program File_ID
5. THE Mint_Account SHALL require Neon balance equal to or greater than the storage cost for its metadata size

### Requirement 6

**User Story:** As a token holder, I want to create a token account, so that I can hold a specific token

#### Acceptance Criteria

1. THE Token_Program SHALL create Token_Account files with balance data in the `data` field
2. THE Token_Account data SHALL include an `owner` field referencing the Owner_Account File_ID
3. THE Token_Account data SHALL include a `tokenBalance` field storing the token amount
4. THE Token_Account data SHALL include a `mint` field referencing the Mint_Account File_ID
5. THE Token_Account SHALL require Neon balance equal to or greater than the storage cost for its data size

### Requirement 7

**User Story:** As a token holder, I want to transfer tokens to another account, so that I can send tokens to other users

#### Acceptance Criteria

1. THE Token_Program SHALL validate that the transaction signer controls the source Token_Account owner
2. THE Token_Program SHALL decrease the source Token_Account tokenBalance by the transfer amount
3. THE Token_Program SHALL increase the destination Token_Account tokenBalance by the transfer amount
4. THE Token_Program SHALL verify both accounts belong to the same Mint_Account
5. THE Token_Program SHALL reject transfers that would result in negative balances

### Requirement 8

**User Story:** As a token creator, I want to mint new tokens, so that I can increase the token supply

#### Acceptance Criteria

1. THE Token_Program SHALL validate that the transaction signer is the mint authority
2. THE Token_Program SHALL increase the destination Token_Account tokenBalance by the mint amount
3. THE Token_Program SHALL update the Mint_Account total supply field
4. WHERE maximum supply is configured, THE Token_Program SHALL reject mint operations that would exceed the maximum supply
5. WHEN mint authority is set to null, THE Token_Program SHALL reject all mint operations

### Requirement 9

**User Story:** As a token creator, I want to burn tokens, so that I can decrease the token supply

#### Acceptance Criteria

1. THE Token_Program SHALL validate that the transaction signer controls the Token_Account owner
2. THE Token_Program SHALL decrease the Token_Account tokenBalance by the burn amount
3. THE Token_Program SHALL update the Mint_Account total supply field
4. THE Token_Program SHALL reject burns that would result in negative balances
5. THE Token_Program SHALL allow any token holder to burn their own tokens

### Requirement 10

**User Story:** As a wallet developer, I want deterministic token account addresses, so that I can find a user's token account without querying

#### Acceptance Criteria

1. THE Token_Program SHALL support Associated_Token_Account derivation from owner and mint
2. THE Associated_Token_Account address SHALL be derived using a deterministic function of owner File_ID and mint File_ID
3. THE Token_Program SHALL provide an instruction to create Associated_Token_Accounts
4. THE system SHALL ensure only one Associated_Token_Account exists per owner-mint pair
5. THE Token_Program SHALL allow users to create additional non-associated Token_Accounts for the same owner-mint pair

### Requirement 11

**User Story:** As a token creator, I want to close token accounts, so that users can reclaim Neon used for storage rent

#### Acceptance Criteria

1. THE Token_Program SHALL validate that the transaction signer controls the Token_Account owner
2. THE Token_Program SHALL verify the Token_Account tokenBalance is zero before closing
3. THE Token_Program SHALL transfer the Token_Account Neon balance to a specified destination
4. THE Token_Program SHALL delete the Token_Account file from state
5. THE Token_Program SHALL reject close operations on accounts with non-zero token balances

### Requirement 12

**User Story:** As a token creator, I want to freeze and thaw token accounts, so that I can implement compliance features

#### Acceptance Criteria

1. THE Token_Program SHALL validate that the transaction signer is the freeze authority
2. THE Token_Program SHALL add a `frozen` boolean field to Token_Account data
3. WHEN a Token_Account is frozen, THE Token_Program SHALL reject all transfer and burn operations
4. THE Token_Program SHALL allow the freeze authority to thaw frozen accounts
5. WHEN freeze authority is set to null, THE Token_Program SHALL reject all freeze and thaw operations

### Requirement 13

**User Story:** As a DeFi developer, I want to approve delegates for token accounts, so that I can build applications that transfer tokens on behalf of users

#### Acceptance Criteria

1. THE Token_Program SHALL add a `delegate` field and `delegatedAmount` field to Token_Account data
2. THE Token_Program SHALL validate that the account owner approves delegate assignments
3. THE Token_Program SHALL allow delegates to transfer up to the delegatedAmount from the account
4. THE Token_Program SHALL decrease delegatedAmount as the delegate transfers tokens
5. THE Token_Program SHALL allow the owner to revoke delegate approval by setting the delegate field to null

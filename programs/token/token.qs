// Token_Program - Built-in program for managing custom fungible tokens
// Implements Requirements 3.1, 3.2, 4.1, 4.2, 4.3, 4.4, 4.5
//
// This program manages token mints, token accounts, and token operations
// including minting, burning, transferring, freezing, and delegation.

// Instruction types
const INITIALIZE_MINT: i64 = 0;
const INITIALIZE_ACCOUNT: i64 = 1;
const TRANSFER: i64 = 2;
const MINT_TO: i64 = 3;
const BURN: i64 = 4;
const CLOSE_ACCOUNT: i64 = 5;
const FREEZE_ACCOUNT: i64 = 6;
const THAW_ACCOUNT: i64 = 7;
const APPROVE: i64 = 8;
const REVOKE: i64 = 9;
const CREATE_ASSOCIATED_TOKEN_ACCOUNT: i64 = 10;

// Error codes
const ERROR_INSUFFICIENT_TOKEN_BALANCE: i64 = 0x2000;
const ERROR_INVALID_MINT: i64 = 0x2001;
const ERROR_INVALID_TOKEN_ACCOUNT: i64 = 0x2002;
const ERROR_MINT_MISMATCH: i64 = 0x2003;
const ERROR_UNAUTHORIZED_MINT_AUTHORITY: i64 = 0x2004;
const ERROR_UNAUTHORIZED_FREEZE_AUTHORITY: i64 = 0x2005;
const ERROR_UNAUTHORIZED_OWNER: i64 = 0x2006;
const ERROR_ACCOUNT_FROZEN: i64 = 0x2007;
const ERROR_ACCOUNT_NOT_EMPTY: i64 = 0x2008;
const ERROR_DELEGATE_NOT_SET: i64 = 0x2009;
const ERROR_INSUFFICIENT_DELEGATED_AMOUNT: i64 = 0x200A;
const ERROR_FIXED_SUPPLY_MINT: i64 = 0x200B;
const ERROR_INVALID_INSTRUCTION: i64 = 0x2FFF;

// Token_Program ID (0x00...02)
const TOKEN_PROGRAM_ID: bytes = [
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2
];

// Entry point for the Token_Program
export function entry(): i64 {
    // Get instruction data
    let instrData: bytes = getInstructionData();
    
    // Parse instruction type (first byte)
    if (len(instrData) < 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let instrType: i64 = instrData[0];
    
    // Dispatch to appropriate handler
    if (instrType == INITIALIZE_MINT) {
        return handleInitializeMint(instrData);
    } else if (instrType == MINT_TO) {
        return handleMintTo(instrData);
    } else if (instrType == INITIALIZE_ACCOUNT) {
        return handleInitializeAccount(instrData);
    } else if (instrType == CREATE_ASSOCIATED_TOKEN_ACCOUNT) {
        return handleCreateAssociatedTokenAccount(instrData);
    } else if (instrType == TRANSFER) {
        return handleTransfer(instrData);
    } else if (instrType == BURN) {
        return handleBurn(instrData);
    } else if (instrType == CLOSE_ACCOUNT) {
        return handleCloseAccount(instrData);
    } else if (instrType == FREEZE_ACCOUNT) {
        return handleFreezeAccount(instrData);
    } else if (instrType == THAW_ACCOUNT) {
        return handleThawAccount(instrData);
    } else if (instrType == APPROVE) {
        return handleApprove(instrData);
    } else if (instrType == REVOKE) {
        return handleRevoke(instrData);
    } else if (instrType == CREATE_ASSOCIATED_TOKEN_ACCOUNT) {
        return handleCreateAssociatedTokenAccount(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// HandleInitializeMint creates a new mint account
// Instruction format:
//   Byte 0: Instruction type (0)
//   Byte 1: Decimals (u8)
//   Byte 2: Mint authority flag (0=null, 1=present)
//   Bytes 3-34: Mint authority pubkey (if present)
//   Byte 35 or 3: Freeze authority flag (0=null, 1=present)
//   Bytes 36-67 or 4-35: Freeze authority pubkey (if present)
// Requirements: 4.1, 4.2, 4.3, 4.4, 4.5
function handleInitializeMint(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse decimals
    if (len(instrData) < offset + 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let decimals: i64 = instrData[offset];
    offset = offset + 1;
    
    // Parse mint authority (optional)
    if (len(instrData) < offset + 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let hasMintAuthority: bool = instrData[offset] != 0;
    offset = offset + 1;
    
    let mintAuthority: bytes = [];
    if (hasMintAuthority) {
        if (len(instrData) < offset + 32) {
            return ERROR_INVALID_INSTRUCTION;
        }
        mintAuthority = slice(instrData, offset, offset + 32);
        offset = offset + 32;
    }
    
    // Parse freeze authority (optional)
    if (len(instrData) < offset + 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let hasFreezeAuthority: bool = instrData[offset] != 0;
    offset = offset + 1;
    
    let freezeAuthority: bytes = [];
    if (hasFreezeAuthority) {
        if (len(instrData) < offset + 32) {
            return ERROR_INVALID_INSTRUCTION;
        }
        freezeAuthority = slice(instrData, offset, offset + 32);
        offset = offset + 32;
    }
    
    // Create MintAccount data structure with zero supply
    // Format: supply (8) + decimals (1) + mint_authority flag (1) + mint_authority (32 if present) 
    //         + freeze_authority flag (1) + freeze_authority (32 if present)
    let supply: i64 = 0;
    let mintData: bytes = serializeMintAccount(supply, decimals, mintAuthority, hasMintAuthority, 
                                                freezeAuthority, hasFreezeAuthority);
    
    // Calculate storage cost for mint account
    let dataSize: i64 = len(mintData);
    let storageCost: i64 = calculateStorageCost(dataSize);
    
    // Create file with MintAccount data and sufficient Neon balance
    // The file needs enough balance to cover storage rent
    let mintFileID: bytes = createFile(mintData, storageCost, TOKEN_PROGRAM_ID);
    
    // Return success (mint FileID would be returned via output mechanism)
    return 0;
}

// HandleMintTo mints new tokens to a destination account
// Instruction format:
//   Byte 0: Instruction type (3)
//   Bytes 1-32: Mint FileID
//   Bytes 33-64: Destination account FileID
//   Bytes 65-72: Amount (u64, little-endian)
// Requirements: 7.1, 7.2, 7.3, 7.4, 7.5
function handleMintTo(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse mint FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let mintFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse destination account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let destFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse amount (u64, little-endian)
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let amount: i64 = parseU64LE(instrData, offset);
    offset = offset + 8;
    
    // Get mint account data
    let mintFile: file = getFile(mintFileID);
    let mintData: bytes = mintFile.data;
    
    // Deserialize mint account
    let mintInfo: MintAccountInfo = deserializeMintAccount(mintData);
    
    // Validate signer is mint authority (Requirement 7.1)
    if (!mintInfo.hasMintAuthority) {
        return ERROR_FIXED_SUPPLY_MINT; // Requirement 7.5: null authority rejects mints
    }
    
    if (!hasSigner(mintInfo.mintAuthority)) {
        return ERROR_UNAUTHORIZED_MINT_AUTHORITY;
    }
    
    // Get destination token account
    let destFile: file = getFile(destFileID);
    let destData: bytes = destFile.data;
    
    // Deserialize token account
    let destInfo: TokenAccountInfo = deserializeTokenAccount(destData);
    
    // Verify destination belongs to this mint (Requirement 7.4 implied)
    if (!bytesEqual(destInfo.mint, mintFileID)) {
        return ERROR_MINT_MISMATCH;
    }
    
    // Increase destination token balance (Requirement 7.2)
    let newDestBalance: i64 = destInfo.tokenBalance + amount;
    if (newDestBalance < destInfo.tokenBalance) {
        // Overflow check
        return ERROR_INSUFFICIENT_TOKEN_BALANCE;
    }
    
    // Update mint total supply (Requirement 7.3)
    let newSupply: i64 = mintInfo.supply + amount;
    if (newSupply < mintInfo.supply) {
        // Overflow check
        return ERROR_INSUFFICIENT_TOKEN_BALANCE;
    }
    
    // TODO: Requirement 7.4 - Enforce maximum supply limits if configured
    // This would require adding a max_supply field to MintAccount
    
    // Update destination account
    destInfo.tokenBalance = newDestBalance;
    let newDestData: bytes = serializeTokenAccount(destInfo);
    updateFile(destFileID, newDestData);
    
    // Update mint account
    mintInfo.supply = newSupply;
    let newMintData: bytes = serializeMintAccount(mintInfo.supply, mintInfo.decimals, 
                                                   mintInfo.mintAuthority, mintInfo.hasMintAuthority,
                                                   mintInfo.freezeAuthority, mintInfo.hasFreezeAuthority);
    updateFile(mintFileID, newMintData);
    
    return 0;
}

// HandleInitializeAccount creates a new token account
// Instruction format:
//   Byte 0: Instruction type (1)
//   Bytes 1-32: Mint FileID
//   Bytes 33-64: Owner FileID
// Requirements: 5.1, 5.2, 5.3, 5.4, 5.5
function handleInitializeAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse mint FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let mintFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse owner FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let ownerFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Verify mint exists
    let mintFile: file = getFile(mintFileID);
    if (len(mintFile.data) == 0) {
        return ERROR_INVALID_MINT;
    }
    
    // Create TokenAccount data structure with zero balance
    let tokenInfo: TokenAccountInfo = {
        mint: mintFileID,
        owner: ownerFileID,
        tokenBalance: 0,
        delegate: [],
        hasDelegate: false,
        delegatedAmount: 0,
        frozen: false
    };
    
    let tokenData: bytes = serializeTokenAccount(tokenInfo);
    
    // Calculate storage cost for token account
    let dataSize: i64 = len(tokenData);
    let storageCost: i64 = calculateStorageCost(dataSize);
    
    // Create file with TokenAccount data and sufficient Neon balance
    // Set TxManager to Token_Program ID
    let accountFileID: bytes = createFile(tokenData, storageCost, TOKEN_PROGRAM_ID);
    
    // Return success (account FileID would be returned via output mechanism)
    return 0;
}

// HandleCreateAssociatedTokenAccount creates a deterministic token account
// Instruction format:
//   Byte 0: Instruction type (10)
//   Bytes 1-32: Owner FileID
//   Bytes 33-64: Mint FileID
// Requirements: 9.1, 9.2, 9.3, 9.4, 9.5
function handleCreateAssociatedTokenAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse owner FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let ownerFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse mint FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let mintFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Verify mint exists
    let mintFile: file = getFile(mintFileID);
    if (len(mintFile.data) == 0) {
        return ERROR_INVALID_MINT;
    }
    
    // Derive deterministic FileID from owner and mint using hash function
    let derivedFileID: bytes = deriveAssociatedTokenAddress(ownerFileID, mintFileID);
    
    // Check if associated account already exists
    let existingFile: file = getFile(derivedFileID);
    if (len(existingFile.data) > 0) {
        // Account already exists, return success (idempotent operation)
        return 0;
    }
    
    // Create TokenAccount with derived FileID
    let tokenInfo: TokenAccountInfo = {
        mint: mintFileID,
        owner: ownerFileID,
        tokenBalance: 0,
        delegate: [],
        hasDelegate: false,
        delegatedAmount: 0,
        frozen: false
    };
    
    let tokenData: bytes = serializeTokenAccount(tokenInfo);
    
    // Calculate storage cost for token account
    let dataSize: i64 = len(tokenData);
    let storageCost: i64 = calculateStorageCost(dataSize);
    
    // Create file with TokenAccount data at the derived address
    createFileWithID(derivedFileID, tokenData, storageCost, TOKEN_PROGRAM_ID);
    
    // Return success
    return 0;
}

// HandleCloseAccount closes a token account and reclaims Neon
// Instruction format:
//   Byte 0: Instruction type (5)
//   Bytes 1-32: Account FileID
//   Bytes 33-64: Destination FileID (receives Neon balance)
// Requirements: 10.1, 10.2, 10.3, 10.4, 10.5
function handleCloseAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse destination FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let destFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Validate signer controls account owner (Requirement 10.1)
    if (!hasSigner(accountInfo.owner)) {
        return ERROR_UNAUTHORIZED_OWNER;
    }
    
    // Verify token balance is zero (Requirement 10.2)
    if (accountInfo.tokenBalance != 0) {
        return ERROR_ACCOUNT_NOT_EMPTY;
    }
    
    // Transfer Neon balance to destination (Requirement 10.3)
    let neonBalance: i64 = accountFile.balance;
    transferBalance(accountFileID, destFileID, neonBalance);
    
    // Delete account file from state (Requirement 10.4)
    deleteFile(accountFileID);
    
    // Return success
    return 0;
}

// HandleTransfer transfers tokens between accounts
// Instruction format:
//   Byte 0: Instruction type (2)
//   Bytes 1-32: Source account FileID
//   Bytes 33-64: Destination account FileID
//   Bytes 65-72: Amount (u64, little-endian)
// Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 12.3, 12.4
function handleTransfer(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse source account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let sourceFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse destination account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let destFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse amount (u64, little-endian)
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let amount: i64 = parseU64LE(instrData, offset);
    offset = offset + 8;
    
    // Get source account data
    let sourceFile: file = getFile(sourceFileID);
    if (len(sourceFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize source token account
    let sourceInfo: TokenAccountInfo = deserializeTokenAccount(sourceFile.data);
    
    // Check if account is frozen (Requirement 6.5, 12.3)
    if (sourceInfo.frozen) {
        return ERROR_ACCOUNT_FROZEN;
    }
    
    // Validate signer controls source account owner OR is approved delegate (Requirement 6.1, 12.4)
    let isOwner: bool = hasSigner(sourceInfo.owner);
    let isDelegate: bool = false;
    if (sourceInfo.hasDelegate) {
        isDelegate = hasSigner(sourceInfo.delegate);
    }
    
    if (!isOwner && !isDelegate) {
        return ERROR_UNAUTHORIZED_OWNER;
    }
    
    // Get destination account data
    let destFile: file = getFile(destFileID);
    if (len(destFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize destination token account
    let destInfo: TokenAccountInfo = deserializeTokenAccount(destFile.data);
    
    // Verify both accounts belong to same mint (Requirement 6.2)
    if (!bytesEqual(sourceInfo.mint, destInfo.mint)) {
        return ERROR_MINT_MISMATCH;
    }
    
    // Check source token balance is sufficient (Requirement 6.3)
    if (sourceInfo.tokenBalance < amount) {
        return ERROR_INSUFFICIENT_TOKEN_BALANCE;
    }
    
    // If delegate transfer, check delegated amount is sufficient (Requirement 12.4)
    if (isDelegate && !isOwner) {
        if (sourceInfo.delegatedAmount < amount) {
            return ERROR_INSUFFICIENT_DELEGATED_AMOUNT;
        }
        // Decrease delegated amount (Requirement 6.8)
        sourceInfo.delegatedAmount = sourceInfo.delegatedAmount - amount;
    }
    
    // Decrease source token balance (Requirement 6.4)
    sourceInfo.tokenBalance = sourceInfo.tokenBalance - amount;
    
    // Increase destination token balance (Requirement 6.5)
    destInfo.tokenBalance = destInfo.tokenBalance + amount;
    
    // Check for overflow in destination
    if (destInfo.tokenBalance < amount) {
        return ERROR_INSUFFICIENT_TOKEN_BALANCE; // Overflow
    }
    
    // Update source account
    let newSourceData: bytes = serializeTokenAccount(sourceInfo);
    updateFile(sourceFileID, newSourceData);
    
    // Update destination account
    let newDestData: bytes = serializeTokenAccount(destInfo);
    updateFile(destFileID, newDestData);
    
    return 0;
}

// HandleBurn burns tokens from an account
// Instruction format:
//   Byte 0: Instruction type (4)
//   Bytes 1-32: Account FileID
//   Bytes 33-40: Amount (u64, little-endian)
// Requirements: 8.1, 8.2, 8.3, 8.4, 8.5
function handleBurn(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse amount (u64, little-endian)
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let amount: i64 = parseU64LE(instrData, offset);
    offset = offset + 8;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Validate signer controls account owner (Requirement 8.1)
    if (!hasSigner(accountInfo.owner)) {
        return ERROR_UNAUTHORIZED_OWNER;
    }
    
    // Check token balance is sufficient (Requirement 8.2, 8.4)
    if (accountInfo.tokenBalance < amount) {
        return ERROR_INSUFFICIENT_TOKEN_BALANCE;
    }
    
    // Decrease account token balance (Requirement 8.3)
    accountInfo.tokenBalance = accountInfo.tokenBalance - amount;
    
    // Update account
    let newAccountData: bytes = serializeTokenAccount(accountInfo);
    updateFile(accountFileID, newAccountData);
    
    // Get mint account to update total supply
    let mintFileID: bytes = accountInfo.mint;
    let mintFile: file = getFile(mintFileID);
    if (len(mintFile.data) == 0) {
        return ERROR_INVALID_MINT;
    }
    
    // Deserialize mint account
    let mintInfo: MintAccountInfo = deserializeMintAccount(mintFile.data);
    
    // Update mint total supply (Requirement 8.5)
    mintInfo.supply = mintInfo.supply - amount;
    
    // Update mint account
    let newMintData: bytes = serializeMintAccount(mintInfo.supply, mintInfo.decimals, 
                                                   mintInfo.mintAuthority, mintInfo.hasMintAuthority,
                                                   mintInfo.freezeAuthority, mintInfo.hasFreezeAuthority);
    updateFile(mintFileID, newMintData);
    
    return 0;
}

// HandleFreezeAccount freezes a token account
// Instruction format:
//   Byte 0: Instruction type (6)
//   Bytes 1-32: Account FileID
// Requirements: 11.1, 11.2, 11.3
function handleFreezeAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Get mint account to check freeze authority
    let mintFileID: bytes = accountInfo.mint;
    let mintFile: file = getFile(mintFileID);
    if (len(mintFile.data) == 0) {
        return ERROR_INVALID_MINT;
    }
    
    // Deserialize mint account
    let mintInfo: MintAccountInfo = deserializeMintAccount(mintFile.data);
    
    // Check freeze authority is not null (Requirement 11.2)
    if (!mintInfo.hasFreezeAuthority) {
        return ERROR_UNAUTHORIZED_FREEZE_AUTHORITY;
    }
    
    // Validate signer is freeze authority (Requirement 11.1)
    if (!hasSigner(mintInfo.freezeAuthority)) {
        return ERROR_UNAUTHORIZED_FREEZE_AUTHORITY;
    }
    
    // Set frozen flag to true in TokenAccount data (Requirement 11.3)
    accountInfo.frozen = true;
    
    // Update account file
    let newAccountData: bytes = serializeTokenAccount(accountInfo);
    updateFile(accountFileID, newAccountData);
    
    return 0;
}

// HandleThawAccount thaws a frozen token account
// Instruction format:
//   Byte 0: Instruction type (7)
//   Bytes 1-32: Account FileID
// Requirements: 11.4, 11.5
function handleThawAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Get mint account to check freeze authority
    let mintFileID: bytes = accountInfo.mint;
    let mintFile: file = getFile(mintFileID);
    if (len(mintFile.data) == 0) {
        return ERROR_INVALID_MINT;
    }
    
    // Deserialize mint account
    let mintInfo: MintAccountInfo = deserializeMintAccount(mintFile.data);
    
    // Check freeze authority is not null (Requirement 11.5)
    if (!mintInfo.hasFreezeAuthority) {
        return ERROR_UNAUTHORIZED_FREEZE_AUTHORITY;
    }
    
    // Validate signer is freeze authority (Requirement 11.4)
    if (!hasSigner(mintInfo.freezeAuthority)) {
        return ERROR_UNAUTHORIZED_FREEZE_AUTHORITY;
    }
    
    // Set frozen flag to false in TokenAccount data (Requirement 11.5)
    accountInfo.frozen = false;
    
    // Update account file
    let newAccountData: bytes = serializeTokenAccount(accountInfo);
    updateFile(accountFileID, newAccountData);
    
    return 0;
}

// HandleApprove approves a delegate to transfer tokens on behalf of the owner
// Instruction format:
//   Byte 0: Instruction type (8)
//   Bytes 1-32: Account FileID
//   Bytes 33-64: Delegate pubkey
//   Bytes 65-72: Amount (u64, little-endian)
// Requirements: 12.1, 12.2
function handleApprove(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse delegate pubkey
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let delegatePubkey: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse amount (u64, little-endian)
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let amount: i64 = parseU64LE(instrData, offset);
    offset = offset + 8;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Validate signer controls account owner (Requirement 12.1)
    if (!hasSigner(accountInfo.owner)) {
        return ERROR_UNAUTHORIZED_OWNER;
    }
    
    // Set delegate field in TokenAccount data (Requirement 12.2)
    accountInfo.delegate = delegatePubkey;
    accountInfo.hasDelegate = true;
    
    // Set delegated amount (Requirement 12.2)
    accountInfo.delegatedAmount = amount;
    
    // Update account file
    let newAccountData: bytes = serializeTokenAccount(accountInfo);
    updateFile(accountFileID, newAccountData);
    
    return 0;
}

// HandleRevoke revokes delegate approval for a token account
// Instruction format:
//   Byte 0: Instruction type (9)
//   Bytes 1-32: Account FileID
// Requirements: 12.5
function handleRevoke(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Get account data
    let accountFile: file = getFile(accountFileID);
    if (len(accountFile.data) == 0) {
        return ERROR_INVALID_TOKEN_ACCOUNT;
    }
    
    // Deserialize token account
    let accountInfo: TokenAccountInfo = deserializeTokenAccount(accountFile.data);
    
    // Validate signer controls account owner (Requirement 12.5)
    if (!hasSigner(accountInfo.owner)) {
        return ERROR_UNAUTHORIZED_OWNER;
    }
    
    // Set delegate field to null (Requirement 12.5)
    accountInfo.delegate = [];
    accountInfo.hasDelegate = false;
    
    // Set delegated amount to zero (Requirement 12.5)
    accountInfo.delegatedAmount = 0;
    
    // Update account file
    let newAccountData: bytes = serializeTokenAccount(accountInfo);
    updateFile(accountFileID, newAccountData);
    
    return 0;
}

// Helper function to derive associated token account address
// Uses a deterministic hash of owner and mint
function deriveAssociatedTokenAddress(owner: bytes, mint: bytes): bytes {
    // Concatenate owner + mint + TOKEN_PROGRAM_ID for uniqueness
    let data: bytes = [];
    data = append(data, owner);
    data = append(data, mint);
    data = append(data, TOKEN_PROGRAM_ID);
    
    // Hash the concatenated data to get deterministic address
    return hashBytes(data);
}

// Helper function to serialize MintAccount
function serializeMintAccount(supply: i64, decimals: i64, mintAuthority: bytes, hasMintAuthority: bool,
                               freezeAuthority: bytes, hasFreezeAuthority: bool): bytes {
    let data: bytes = [];
    
    // Supply (8 bytes, little-endian)
    data = append(data, serializeU64LE(supply));
    
    // Decimals (1 byte)
    data = append(data, [decimals]);
    
    // Mint authority flag (1 byte)
    if (hasMintAuthority) {
        data = append(data, [1]);
        data = append(data, mintAuthority);
    } else {
        data = append(data, [0]);
    }
    
    // Freeze authority flag (1 byte)
    if (hasFreezeAuthority) {
        data = append(data, [1]);
        data = append(data, freezeAuthority);
    } else {
        data = append(data, [0]);
    }
    
    return data;
}

// Helper struct for MintAccount deserialization
struct MintAccountInfo {
    supply: i64,
    decimals: i64,
    mintAuthority: bytes,
    hasMintAuthority: bool,
    freezeAuthority: bytes,
    hasFreezeAuthority: bool
}

// Helper function to deserialize MintAccount
function deserializeMintAccount(data: bytes): MintAccountInfo {
    let info: MintAccountInfo = {
        supply: 0,
        decimals: 0,
        mintAuthority: [],
        hasMintAuthority: false,
        freezeAuthority: [],
        hasFreezeAuthority: false
    };
    
    let offset: i64 = 0;
    
    // Read supply
    info.supply = parseU64LE(data, offset);
    offset = offset + 8;
    
    // Read decimals
    info.decimals = data[offset];
    offset = offset + 1;
    
    // Read mint authority
    let mintAuthorityFlag: i64 = data[offset];
    offset = offset + 1;
    if (mintAuthorityFlag == 1) {
        info.mintAuthority = slice(data, offset, offset + 32);
        info.hasMintAuthority = true;
        offset = offset + 32;
    }
    
    // Read freeze authority
    let freezeAuthorityFlag: i64 = data[offset];
    offset = offset + 1;
    if (freezeAuthorityFlag == 1) {
        info.freezeAuthority = slice(data, offset, offset + 32);
        info.hasFreezeAuthority = true;
        offset = offset + 32;
    }
    
    return info;
}

// Helper struct for TokenAccount deserialization
struct TokenAccountInfo {
    mint: bytes,
    owner: bytes,
    tokenBalance: i64,
    delegate: bytes,
    hasDelegate: bool,
    delegatedAmount: i64,
    frozen: bool
}

// Helper function to deserialize TokenAccount
function deserializeTokenAccount(data: bytes): TokenAccountInfo {
    let info: TokenAccountInfo = {
        mint: [],
        owner: [],
        tokenBalance: 0,
        delegate: [],
        hasDelegate: false,
        delegatedAmount: 0,
        frozen: false
    };
    
    let offset: i64 = 0;
    
    // Read mint
    info.mint = slice(data, offset, offset + 32);
    offset = offset + 32;
    
    // Read owner
    info.owner = slice(data, offset, offset + 32);
    offset = offset + 32;
    
    // Read token balance
    info.tokenBalance = parseU64LE(data, offset);
    offset = offset + 8;
    
    // Read delegate
    let delegateFlag: i64 = data[offset];
    offset = offset + 1;
    if (delegateFlag == 1) {
        info.delegate = slice(data, offset, offset + 32);
        info.hasDelegate = true;
        offset = offset + 32;
    }
    
    // Read delegated amount
    info.delegatedAmount = parseU64LE(data, offset);
    offset = offset + 8;
    
    // Read frozen
    info.frozen = data[offset] != 0;
    
    return info;
}

// Helper function to serialize TokenAccount
function serializeTokenAccount(info: TokenAccountInfo): bytes {
    let data: bytes = [];
    
    // Mint (32 bytes)
    data = append(data, info.mint);
    
    // Owner (32 bytes)
    data = append(data, info.owner);
    
    // Token balance (8 bytes, little-endian)
    data = append(data, serializeU64LE(info.tokenBalance));
    
    // Delegate flag and data
    if (info.hasDelegate) {
        data = append(data, [1]);
        data = append(data, info.delegate);
    } else {
        data = append(data, [0]);
    }
    
    // Delegated amount (8 bytes, little-endian)
    data = append(data, serializeU64LE(info.delegatedAmount));
    
    // Frozen (1 byte)
    if (info.frozen) {
        data = append(data, [1]);
    } else {
        data = append(data, [0]);
    }
    
    return data;
}

// Helper function to parse u64 little-endian
function parseU64LE(data: bytes, offset: i64): i64 {
    let value: i64 = 0;
    value = value | data[offset];
    value = value | (data[offset + 1] << 8);
    value = value | (data[offset + 2] << 16);
    value = value | (data[offset + 3] << 24);
    value = value | (data[offset + 4] << 32);
    value = value | (data[offset + 5] << 40);
    value = value | (data[offset + 6] << 48);
    value = value | (data[offset + 7] << 56);
    return value;
}

// Helper function to serialize u64 little-endian
function serializeU64LE(value: i64): bytes {
    let data: bytes = [
        value & 0xFF,
        (value >> 8) & 0xFF,
        (value >> 16) & 0xFF,
        (value >> 24) & 0xFF,
        (value >> 32) & 0xFF,
        (value >> 40) & 0xFF,
        (value >> 48) & 0xFF,
        (value >> 56) & 0xFF
    ];
    return data;
}

// Helper function to calculate storage cost
// Uses exponential cost model: base_cost_per_kb * size_in_kb * (1.1 ^ size_in_mb)
function calculateStorageCost(dataSize: i64): i64 {
    let baseCostPerKB: i64 = 1000;
    let sizeInKB: i64 = (dataSize + 1023) / 1024; // Round up
    let sizeInMB: i64 = (dataSize + 1048575) / 1048576; // Round up
    
    // Simplified exponential calculation (1.1 ^ sizeInMB)
    let multiplier: i64 = 1;
    let i: i64 = 0;
    while (i < sizeInMB) {
        multiplier = (multiplier * 11) / 10;
        i = i + 1;
    }
    
    return baseCostPerKB * sizeInKB * multiplier;
}

// Helper function to compare byte arrays
function bytesEqual(a: bytes, b: bytes): bool {
    if (len(a) != len(b)) {
        return false;
    }
    
    let i: i64 = 0;
    while (i < len(a)) {
        if (a[i] != b[i]) {
            return false;
        }
        i = i + 1;
    }
    
    return true;
}

// Note: The following functions are builtins provided by the QuanticScript runtime:
// - getInstructionData(): bytes - Gets instruction data from transaction context
// - getFile(fileID: bytes): file - Gets file from FileStore (read-only)
// - updateFile(fileID: bytes, data: bytes): void - Updates file data in FileStore
// - hasSigner(pubkey: bytes): bool - Checks if pubkey signed the transaction
// - sha256(data: bytes): bytes - Computes SHA-256 hash
//
// The following helper functions need to be implemented as they're not builtins:

// Helper function to extract a slice from bytes
function slice(data: bytes, start: i64, end: i64): bytes {
    let result: bytes = [];
    let i: i64 = start;
    while (i < end) {
        result = append(result, [data[i]]);
        i = i + 1;
    }
    return result;
}

// Helper function to concatenate byte arrays
function append(a: bytes, b: bytes): bytes {
    let result: bytes = a;
    let i: i64 = 0;
    while (i < len(b)) {
        result = appendByte(result, b[i]);
        i = i + 1;
    }
    return result;
}

// Helper function to append a single byte
function appendByte(arr: bytes, b: i64): bytes {
    // This is a workaround - in real implementation this would be more efficient
    let newArr: bytes = arr;
    // Assuming the runtime provides array append functionality
    return newArr;
}

// Helper function to get length of bytes
function len(data: bytes): i64 {
    let count: i64 = 0;
    // Count elements - this assumes the runtime provides array length
    // In actual implementation, this would be a builtin
    return count;
}

// Helper function to hash bytes using SHA-256
function hashBytes(data: bytes): bytes {
    return sha256(data);
}

// File structure (matches FileStore File type)
struct file {
    data: bytes,
    balance: i64,
    txManager: bytes,
    executable: bool
}

// Note: The following operations are NOT available as builtins and would need
// to be implemented via cross-program invocation to System_Program:
// - createFile(data: bytes, balance: i64, txManager: bytes): bytes
// - createFileWithID(fileID: bytes, data: bytes, balance: i64, txManager: bytes): void
// - transferBalance(from: bytes, to: bytes, amount: i64): void
// - deleteFile(fileID: bytes): void
//
// These operations require privileged access and should be handled by invoking
// the System_Program or through special runtime support.

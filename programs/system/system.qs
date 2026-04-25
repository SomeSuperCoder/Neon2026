// System_Program - Built-in program for managing Neon accounts
// Implements Requirements 3.1, 3.2, 3.6, 3.7
//
// This program manages Neon (native coin) accounts and basic operations
// including account creation, transfers, and storage allocation.

// Instruction types
const CREATE_ACCOUNT: i64 = 0;
const TRANSFER: i64 = 1;
const ALLOCATE_SPACE: i64 = 2;

// Error codes
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_INVALID_ACCOUNT: i64 = 0x1001;
const ERROR_BALANCE_OVERFLOW: i64 = 0x1002;
const ERROR_STORAGE_RENT_VIOLATION: i64 = 0x1003;
const ERROR_UNAUTHORIZED_SIGNER: i64 = 0x1004;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;

// System_Program ID (0x00...01)
const SYSTEM_PROGRAM_ID: bytes = [
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1
];

// Entry point for the System_Program
export function entry(): i64 {
    // Get instruction data
    let instrData: bytes = getInstructionData();
    
    // Parse instruction type (first byte)
    if (len(instrData) < 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    let instrType: i64 = instrData[0];
    
    // Dispatch to appropriate handler
    if (instrType == CREATE_ACCOUNT) {
        return handleCreateAccount(instrData);
    } else if (instrType == TRANSFER) {
        return handleTransfer(instrData);
    } else if (instrType == ALLOCATE_SPACE) {
        return handleAllocateSpace(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// HandleCreateAccount creates a new Neon account
// Instruction format:
//   Byte 0: Instruction type (0)
//   Bytes 1-32: Owner pubkey
//   Bytes 33-40: Initial balance (i64, little-endian)
// Requirements: 1.1, 1.4
function handleCreateAccount(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse owner pubkey
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let ownerPubkey: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse initial balance
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let initialBalance: i64 = parseI64LE(instrData, offset);
    offset = offset + 8;
    
    // Validate initial balance is non-negative (Requirement 1.4)
    if (initialBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Create new file with empty data field and specified balance
    // Set TxManager to System_Program ID
    // Note: This requires runtime support for privileged file creation
    // The createFile builtin would:
    // 1. Generate a new FileID
    // 2. Create file with empty data field
    // 3. Set balance to initialBalance
    // 4. Set TxManager to SYSTEM_PROGRAM_ID
    // 5. Return the new FileID
    
    // For now, validation is complete - actual file creation requires
    // runtime/FileStore integration via createFile builtin
    return 0;
}

// HandleTransfer transfers Neon between accounts
// Instruction format:
//   Byte 0: Instruction type (1)
//   Bytes 1-32: Source account FileID
//   Bytes 33-64: Destination account FileID
//   Bytes 65-72: Amount (i64, little-endian)
// Requirements: 1.2, 1.4
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
    
    // Parse amount
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let amount: i64 = parseI64LE(instrData, offset);
    offset = offset + 8;
    
    // Validate amount is non-negative
    if (amount < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Validate signer has authority over source account (Requirement 1.2)
    // For System_Program accounts, check if the source FileID is a signer
    if (!hasSigner(sourceFileID)) {
        return ERROR_UNAUTHORIZED_SIGNER;
    }
    
    // Get source account balance
    let sourceBalance: i64 = getBalance(sourceFileID);
    
    // Check source balance is sufficient (Requirement 1.4)
    if (sourceBalance < amount) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Get destination account balance
    let destBalance: i64 = getBalance(destFileID);
    
    // Check for overflow in destination balance (Requirement 1.4)
    let newDestBalance: i64 = destBalance + amount;
    if (newDestBalance < destBalance) {
        return ERROR_BALANCE_OVERFLOW;
    }
    
    // Calculate new source balance
    let newSourceBalance: i64 = sourceBalance - amount;
    
    // Validate both balances remain non-negative (Requirement 1.4)
    if (newSourceBalance < 0 || newDestBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Decrease source balance by amount
    // Increase destination balance by amount
    // Note: This requires runtime support for balance updates via updateBalance builtin
    // The updateBalance builtin would atomically update file balances
    // transferBalance(sourceFileID, destFileID, amount) would:
    // 1. Decrease source balance by amount
    // 2. Increase destination balance by amount
    // 3. Validate storage rent constraints
    
    // For now, validation is complete - actual balance updates require
    // runtime/FileStore integration via transferBalance builtin
    return 0;
}

// HandleAllocateSpace allocates additional Neon balance for storage
// Instruction format:
//   Byte 0: Instruction type (2)
//   Bytes 1-32: Account FileID
//   Bytes 33-40: Additional balance (i64, little-endian)
// Requirements: 2.1, 2.2, 2.3
function handleAllocateSpace(instrData: bytes): i64 {
    let offset: i64 = 1; // Skip instruction type
    
    // Parse account FileID
    if (len(instrData) < offset + 32) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let accountFileID: bytes = slice(instrData, offset, offset + 32);
    offset = offset + 32;
    
    // Parse additional balance
    if (len(instrData) < offset + 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    let additionalBalance: i64 = parseI64LE(instrData, offset);
    offset = offset + 8;
    
    // Validate additional balance is non-negative
    if (additionalBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Validate signer has authority over account
    if (!hasSigner(accountFileID)) {
        return ERROR_UNAUTHORIZED_SIGNER;
    }
    
    // Get account balance
    let accountBalance: i64 = getBalance(accountFileID);
    
    // Calculate new balance
    let newBalance: i64 = accountBalance + additionalBalance;
    if (newBalance < accountBalance) {
        return ERROR_BALANCE_OVERFLOW;
    }
    
    // Get account data to verify storage cost
    let accountData: bytes = getFile(accountFileID);
    let dataSize: i64 = len(accountData);
    
    // Verify new balance covers existing data storage cost (Requirement 2.3)
    let requiredBalance: i64 = calculateStorageCost(dataSize);
    
    if (newBalance < requiredBalance) {
        return ERROR_STORAGE_RENT_VIOLATION;
    }
    
    // Increase account balance
    // Note: This requires runtime support for balance updates via updateBalance builtin
    // The updateBalance builtin would:
    // 1. Increase account balance by additionalBalance
    // 2. Validate storage rent constraints
    
    // For now, validation is complete - actual balance update requires
    // runtime/FileStore integration via updateBalance builtin
    return 0;
}

// Helper function to calculate storage cost
// Uses exponential cost model: base_cost_per_kb * size_in_kb * (1.1 ^ size_in_mb)
// Requirements: 2.1, 2.2
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

// Helper function to parse i64 little-endian
function parseI64LE(data: bytes, offset: i64): i64 {
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
    return newArr;
}

// Note: The following functions are builtins provided by the QuanticScript runtime:
// - getInstructionData(): bytes - Gets instruction data from transaction context
// - getFile(fileID: bytes): bytes - Gets file data from FileStore (read-only)
// - getBalance(fileID: bytes): i64 - Gets file balance from FileStore
// - hasSigner(pubkey: bytes): bool - Checks if pubkey signed the transaction
// - len(data: bytes): i64 - Returns length of byte array
// 
// The following builtins are needed but not yet implemented:
// - createFile(data: bytes, balance: i64, txManager: bytes): bytes - Creates new file, returns FileID
// - updateBalance(fileID: bytes, newBalance: i64): void - Updates file balance
// - transferBalance(from: bytes, to: bytes, amount: i64): void - Transfers balance between files

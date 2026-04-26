// System_Program: Built-in program for account management
// Uses the new transfer() instruction for all balance operations
// This is the single source of truth - assembly and bytecode are compiler outputs

// Error codes
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_INVALID_ACCOUNT: i64 = 0x1001;
const ERROR_BALANCE_OVERFLOW: i64 = 0x1002;
const ERROR_STORAGE_RENT_VIOLATION: i64 = 0x1003;
const ERROR_UNAUTHORIZED_SIGNER: i64 = 0x1004;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;
const ERROR_INVALID_AMOUNT: i64 = 0x1005;

const SUCCESS: i64 = 0;
const MAX_I64: i64 = 9223372036854775807;

// Instruction type codes
const INSTR_CREATE_ACCOUNT: i64 = 0;
const INSTR_TRANSFER: i64 = 1;
const INSTR_ALLOCATE_SPACE: i64 = 2;

// Entry point with instruction dispatch
export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    
    // Check if instruction data has at least 8 bytes for type parsing
    if (len(instrData) < 8) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Get instruction type from first 8 bytes (first byte is the type, rest are padding/data)
    let typeBytes: bytes = slice(instrData, 0, 8);
    let instrType: i64 = bytesToI64LE(typeBytes);
    // Mask to get only the first byte (0-255)
    instrType = instrType & 255;
    
    // Dispatch based on instruction type
    if (instrType == INSTR_CREATE_ACCOUNT) {
        return handleCreateAccount(instrData);
    }
    
    if (instrType == INSTR_TRANSFER) {
        return handleTransfer(instrData);
    }
    
    if (instrType == INSTR_ALLOCATE_SPACE) {
        return handleAllocateSpace(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// CREATE_ACCOUNT: Creates a new account file
// Format: [type:u8][owner:FileID(32)][balance:i64(8)]
function handleCreateAccount(instrData: bytes): i64 {
    // Validate length (should be 41 bytes: 1 + 32 + 8)
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract owner FileID (bytes 1-33)
    let ownerBytes: bytes = slice(instrData, 1, 33);
    let owner: FileID = bytesToFileID(ownerBytes);
    
    // Extract initial balance (bytes 33-41, little-endian)
    let balanceBytes: bytes = slice(instrData, 33, 41);
    let balance: i64 = bytesToI64LE(balanceBytes);
    
    // Validate balance is non-negative
    if (balance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Get system program's FileID
    let systemFileID: FileID = getProgramID();
    
    // Transfer balance from system program to new account
    transfer(systemFileID, owner, balance);
    
    return SUCCESS;
}

// TRANSFER: Transfers balance between two accounts
// Format: [type:u8][from:FileID(32)][to:FileID(32)][amount:i64(8)]
function handleTransfer(instrData: bytes): i64 {
    // Validate length (should be 73 bytes: 1 + 32 + 32 + 8)
    if (len(instrData) != 73) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract from FileID (bytes 1-33)
    let fromBytes: bytes = slice(instrData, 1, 33);
    let fromFileID: FileID = bytesToFileID(fromBytes);
    
    // Extract to FileID (bytes 33-65)
    let toBytes: bytes = slice(instrData, 33, 65);
    let toFileID: FileID = bytesToFileID(toBytes);
    
    // Extract amount (bytes 65-73, little-endian)
    let amountBytes: bytes = slice(instrData, 65, 73);
    let amount: i64 = bytesToI64LE(amountBytes);
    
    // Validate amount is positive
    if (amount <= 0) {
        return ERROR_INVALID_AMOUNT;
    }
    
    // Use the transfer instruction
    // This will automatically validate ownership and storage costs
    transfer(fromFileID, toFileID, amount);
    
    return SUCCESS;
}

// ALLOCATE_SPACE: Adds balance to an account for storage rent
// Format: [type:u8][account:FileID(32)][extraBalance:i64(8)]
function handleAllocateSpace(instrData: bytes): i64 {
    // Validate length (should be 41 bytes: 1 + 32 + 8)
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract account FileID (bytes 1-33)
    let accountBytes: bytes = slice(instrData, 1, 33);
    let account: FileID = bytesToFileID(accountBytes);
    
    // Extract extra balance (bytes 33-41, little-endian)
    let extraBalanceBytes: bytes = slice(instrData, 33, 41);
    let extraBalance: i64 = bytesToI64LE(extraBalanceBytes);
    
    // Validate extra balance is non-negative
    if (extraBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Transfer from system program to the account
    let systemFileID: FileID = getProgramID();
    transfer(systemFileID, account, extraBalance);
    
    return SUCCESS;
}

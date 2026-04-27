// System_Program: Built-in program for account management
// Implements four core operations: CREATE_FILE, TRANSFER, BURN, CLOSE_FILE

// Error codes
const SUCCESS: i64 = 0x00;
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_UNAUTHORIZED: i64 = 0x1004;
const ERROR_INVALID_AMOUNT: i64 = 0x1005;
const ERROR_FILE_HAS_DATA: i64 = 0x1006;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;

// Instruction type codes
const INSTR_CREATE_FILE: i64 = 0;
const INSTR_TRANSFER: i64 = 1;
const INSTR_BURN: i64 = 2;
const INSTR_CLOSE_FILE: i64 = 3;

// Entry point with instruction dispatch
export function entry(): i64 {
    let instrData: bytes = getInstructionData();
    
    // Check if instruction data has at least 1 byte for type
    if (len(instrData) < 1) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Get instruction type from first byte
    // Extract first 8 bytes and decode as i64, then mask to get first byte
    let typeBytes: bytes = slice(instrData, 0, 8);
    let instrType: i64 = bytesToI64LE(typeBytes);
    instrType = instrType & 255;
    
    // Dispatch based on instruction type
    if (instrType == INSTR_CREATE_FILE) {
        return handleCreateFile(instrData);
    }
    
    if (instrType == INSTR_TRANSFER) {
        return handleTransfer(instrData);
    }
    
    if (instrType == INSTR_BURN) {
        return handleBurn(instrData);
    }
    
    if (instrType == INSTR_CLOSE_FILE) {
        return handleCloseFile(instrData);
    }
    
    return ERROR_INVALID_INSTRUCTION;
}

// CREATE_FILE: Creates a new file owned by the System Program
// Format: [type:u8(1)][owner:FileID(32)][initial_balance:i64(8)] = 41 bytes
function handleCreateFile(instrData: bytes): i64 {
    // Validate length
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
        return ERROR_INVALID_AMOUNT;
    }
    
    // Get system program's FileID
    let systemFileID: FileID = getProgramID();
    
    // Create new file with System Program as TxManager
    // The createFile opcode sets TxManager to the current program (System Program)
    let emptyData: bytes = slice(instrData, 0, 0);  // Empty bytes
    createFile(owner, emptyData, 0);
    
    // Transfer initial balance from system program to new file
    // This will automatically validate that system program has sufficient balance
    if (balance > 0) {
        transfer(systemFileID, owner, balance);
    }
    
    return SUCCESS;
}

// TRANSFER: Transfers balance from a file owned by System Program to any other file
// Format: [type:u8(1)][from:FileID(32)][to:FileID(32)][amount:i64(8)] = 73 bytes
function handleTransfer(instrData: bytes): i64 {
    // Validate length
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
    // OpTransfer automatically validates:
    // - Ownership (from file must be owned by System Program)
    // - Storage cost constraints
    // - Balance sufficiency
    // If any validation fails, the transfer opcode will error and rollback
    transfer(fromFileID, toFileID, amount);
    
    return SUCCESS;
}

// BURN: Permanently removes Neon from circulation by reducing a file's balance
// Format: [type:u8(1)][file:FileID(32)][amount:i64(8)] = 41 bytes
function handleBurn(instrData: bytes): i64 {
    // Validate length
    if (len(instrData) != 41) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract file FileID (bytes 1-33)
    let fileBytes: bytes = slice(instrData, 1, 33);
    let fileID: FileID = bytesToFileID(fileBytes);
    
    // Extract burn amount (bytes 33-41, little-endian)
    let amountBytes: bytes = slice(instrData, 33, 41);
    let amount: i64 = bytesToI64LE(amountBytes);
    
    // Validate amount is positive
    if (amount <= 0) {
        return ERROR_INVALID_AMOUNT;
    }
    
    // Get file data and current balance
    let fileData: bytes = getFile(fileID);
    let dataLen: i64 = len(fileData);
    let currentBalance: i64 = getBalance(fileID);
    
    // Check if burn would make balance negative
    if (currentBalance < amount) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Calculate storage cost based on data size
    // Storage cost formula: base * size_in_kb * (1.1 ^ size_in_mb)
    // For simplicity, we'll use a conservative estimate:
    // - 0 bytes = 0 cost
    // - 1-1024 bytes = 1000 cost
    // - 1025-2048 bytes = 2000 cost, etc.
    let storageCost: i64 = 0;
    if (dataLen > 0) {
        let sizeInKB: i64 = (dataLen + 1023) / 1024;  // Round up
        storageCost = sizeInKB * 1000;
    }
    
    // Check if remaining balance would cover storage cost
    let newBalance: i64 = currentBalance - amount;
    if (newBalance < storageCost) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Perform the burn by transferring to system program
    // The transfer opcode will validate ownership and storage costs
    let systemFileID: FileID = getProgramID();
    transfer(fileID, systemFileID, amount);
    
    return SUCCESS;
}

// CLOSE_FILE: Deletes a file owned by System Program and transfers remaining balance
// Format: [type:u8(1)][file_to_close:FileID(32)][destination:FileID(32)] = 65 bytes
function handleCloseFile(instrData: bytes): i64 {
    // Validate length
    if (len(instrData) != 65) {
        return ERROR_INVALID_INSTRUCTION;
    }
    
    // Extract file_to_close FileID (bytes 1-33)
    let fileToCloseBytes: bytes = slice(instrData, 1, 33);
    let fileToClose: FileID = bytesToFileID(fileToCloseBytes);
    
    // Extract destination FileID (bytes 33-65)
    let destBytes: bytes = slice(instrData, 33, 65);
    let destination: FileID = bytesToFileID(destBytes);
    
    // Get file data to check if it's empty
    let fileData: bytes = getFile(fileToClose);
    
    // Verify file has zero data length
    if (len(fileData) > 0) {
        return ERROR_FILE_HAS_DATA;
    }
    
    // Get remaining balance
    let remainingBalance: i64 = getBalance(fileToClose);
    
    // Transfer all remaining balance to destination
    // This will automatically validate ownership (file must be owned by System Program)
    if (remainingBalance > 0) {
        transfer(fileToClose, destination, remainingBalance);
    }
    
    // Delete the file
    // OpDeleteFile will validate ownership automatically
    deleteFile(fileToClose);
    
    return SUCCESS;
}

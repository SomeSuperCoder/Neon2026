// System_Program: Built-in program for account management
// Handles account creation, balance transfers, and space allocation
// This is the single source of truth - assembly and bytecode are compiler outputs

// Error codes
const ERROR_INSUFFICIENT_BALANCE: i64 = 0x1000;
const ERROR_INVALID_ACCOUNT: i64 = 0x1001;
const ERROR_BALANCE_OVERFLOW: i64 = 0x1002;
const ERROR_STORAGE_RENT_VIOLATION: i64 = 0x1003;
const ERROR_UNAUTHORIZED_SIGNER: i64 = 0x1004;
const ERROR_INVALID_INSTRUCTION: i64 = 0x1FFF;

// Success code
const SUCCESS: i64 = 0;

// Maximum i64 value for overflow checks
const MAX_I64: i64 = 9223372036854775807;

// Entry point
// For MVP, this returns an error code
// The full implementation will use DISPATCH opcode to parse instructions
// and route to the appropriate handler functions
export function entry(): i64 {
    // Get instruction data
    let instrData: bytes = getInstructionData();
    
    // TODO: Use DISPATCH opcode to parse instruction
    // For now, return invalid instruction error
    // The DISPATCH opcode will:
    // 1. Parse instruction type code from data[0]
    // 2. Look up SystemProgramRegistry
    // 3. Parse all args according to schema
    // 4. Push args + handler name onto stack
    // 5. Branch to appropriate handler
    
    return ERROR_INVALID_INSTRUCTION;
}

// Handler for CREATE_ACCOUNT instruction
// Creates a new account with the specified owner and initial balance
// Note: In production, DISPATCH will convert bytes to PublicKey
// and FileID index before calling this function
function handleCreateAccount(ownerIdx: i64, balance: i64): i64 {
    // For now, we work with FileID indices (i64)
    // The actual implementation will need PublicKey type conversion
    
    // Validate: balance must be non-negative
    if (balance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Create the account by updating its balance
    updateBalance(ownerIdx, balance);
    
    return SUCCESS;
}

// Handler for TRANSFER instruction
// Transfers balance from one account to another
function handleTransfer(fromIdx: i64, toIdx: i64, amount: i64): i64 {
    // Validate: amount must be positive
    if (amount <= 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Get current balance of from account
    let fromBalance: i64 = getBalance(fromIdx);
    
    // Validate: from account must have sufficient balance
    if (fromBalance < amount) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Get current balance of to account
    let toBalance: i64 = getBalance(toIdx);
    
    // Validate: transfer must not cause overflow
    if (toBalance > MAX_I64 - amount) {
        return ERROR_BALANCE_OVERFLOW;
    }
    
    // Perform the transfer
    let negAmount: i64 = 0 - amount;
    updateBalance(fromIdx, negAmount);
    updateBalance(toIdx, amount);
    
    return SUCCESS;
}

// Handler for ALLOCATE_SPACE instruction
// Allocates additional space for an account by adding balance for storage rent
function handleAllocateSpace(accountIdx: i64, extraBalance: i64): i64 {
    // Validate: extra_balance must be non-negative
    if (extraBalance < 0) {
        return ERROR_INSUFFICIENT_BALANCE;
    }
    
    // Get current balance
    let currentBalance: i64 = getBalance(accountIdx);
    
    // Validate: allocation must not cause overflow
    if (currentBalance > MAX_I64 - extraBalance) {
        return ERROR_BALANCE_OVERFLOW;
    }
    
    // Add the extra balance for storage rent
    updateBalance(accountIdx, extraBalance);
    
    return SUCCESS;
}

// Advanced Token Program
// Demonstrates comprehensive QuanticScript features including:
// - Blockchain operations (file access, balance management via system program)
// - Cross-program invocation (calling system program for transfers)
// - Control flow (if/else, loops)
// - Type system usage
// - Error handling
//
// NOTE: This example demonstrates the intended standard library API.
// Balance transfers must be done by invoking the system program.
// Only the system program can call updateBalance directly.

export function entry(ctx: InstructionContext): i64 {
    // Get instruction data
    let data: bytes = getInstructionData();
    
    // Parse operation type from first byte
    // 0 = transfer, 1 = mint (requires authority), 2 = burn
    let operation: i64 = 0;  // Simplified - would parse from data
    
    if (operation == 0) {
        return handleTransfer();
    } else if (operation == 1) {
        return handleMint();
    } else if (operation == 2) {
        return handleBurn();
    } else {
        return -1;  // Error: unknown operation
    }
}

// Handle token transfer between accounts
// Transfers must be done by invoking the system program
function handleTransfer(): i64 {
    let sourceId: i64 = 100;
    let destId: i64 = 200;
    let amount: i64 = 1000;
    
    // Verify source has sufficient balance
    let sourceBalance: i64 = getBalance(sourceId);
    if (sourceBalance < amount) {
        return -2;  // Error: insufficient balance
    }
    
    // Check for overflow on destination
    let destBalance: i64 = getBalance(destId);
    let maxBalance: i64 = 9223372036854775807;  // i64 max
    if (destBalance > maxBalance - amount) {
        return -3;  // Error: overflow
    }
    
    // Execute transfer by invoking the system program
    // System program ID is 0x01 (all zeros except last byte)
    let systemProgramId: i64 = 0x01;
    
    // Encode transfer instruction for system program
    // Format: [instruction_type=1, amount (8 bytes)]
    // Note: In real implementation, would properly encode the instruction
    let transferData: bytes = encodeTransferInstruction(amount);
    
    // Invoke system program with 10000 compute units
    let result: i64 = invoke(systemProgramId, transferData, 10000);
    
    if (result != 0) {
        return -5;  // Error: system program invocation failed
    }
    
    return 0;  // Success
}

// Handle minting new tokens (requires authority)
// Note: Minting also requires invoking system program
function handleMint(): i64 {
    let mintAuthority: i64 = 1;  // Authority account
    let targetAccount: i64 = 200;
    let mintAmount: i64 = 5000;
    
    // Verify mint authority signed the transaction
    // In real implementation, would check actual signer
    let isAuthorized: bool = true;
    
    if (!isAuthorized) {
        return -4;  // Error: unauthorized
    }
    
    // Check for overflow
    let currentBalance: i64 = getBalance(targetAccount);
    let maxBalance: i64 = 9223372036854775807;
    if (currentBalance > maxBalance - mintAmount) {
        return -3;  // Error: overflow
    }
    
    // Mint tokens by invoking system program
    // In a real system, minting would require special authority
    // and would be handled through a dedicated mint instruction
    let systemProgramId: i64 = 0x01;
    let mintData: bytes = encodeMintInstruction(targetAccount, mintAmount);
    let result: i64 = invoke(systemProgramId, mintData, 10000);
    
    if (result != 0) {
        return -5;  // Error: system program invocation failed
    }
    
    return 0;  // Success
}

// Handle burning tokens
// Note: Burning also requires invoking system program
function handleBurn(): i64 {
    let sourceId: i64 = 100;
    let burnAmount: i64 = 500;
    
    // Verify sufficient balance
    let balance: i64 = getBalance(sourceId);
    if (balance < burnAmount) {
        return -2;  // Error: insufficient balance
    }
    
    // Burn tokens by invoking system program
    // Burning is essentially a transfer to a null/burn account
    let systemProgramId: i64 = 0x01;
    let burnData: bytes = encodeBurnInstruction(sourceId, burnAmount);
    let result: i64 = invoke(systemProgramId, burnData, 10000);
    
    if (result != 0) {
        return -5;  // Error: system program invocation failed
    }
    
    return 0;  // Success
}

// Helper function to encode transfer instruction
// Format: [instruction_type=1, amount (8 bytes little-endian)]
function encodeTransferInstruction(amount: i64): bytes {
    // Simplified - real implementation would properly encode
    let data: bytes = allocateBytes(9);
    // data[0] = 1;  // Transfer instruction type
    // Encode amount as little-endian i64
    return data;
}

// Helper function to encode mint instruction
function encodeMintInstruction(target: i64, amount: i64): bytes {
    // Simplified - real implementation would properly encode
    let data: bytes = allocateBytes(17);
    return data;
}

// Helper function to encode burn instruction
function encodeBurnInstruction(source: i64, amount: i64): bytes {
    // Simplified - real implementation would properly encode
    let data: bytes = allocateBytes(17);
    return data;
}

// Calculate transfer fee based on amount
function calculateFee(amount: i64): i64 {
    // Fee is 0.1% (1/1000) with minimum of 1
    let fee: i64 = amount / 1000;
    
    if (fee < 1) {
        fee = 1;
    }
    
    return fee;
}

// Validate account ID is within valid range
function isValidAccount(accountId: i64): bool {
    if (accountId < 1) {
        return false;
    }
    
    if (accountId > 1000000) {
        return false;
    }
    
    return true;
}

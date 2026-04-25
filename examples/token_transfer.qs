// Token Transfer Program
// Demonstrates blockchain operations, file access, and balance management
// 
// NOTE: Balance transfers must be done by invoking the system program.
// Only the system program can call updateBalance directly.
// This example shows the INCORRECT approach for educational purposes.

export function entry(ctx: InstructionContext): i64 {
    // Parse instruction data to get transfer parameters
    // Expected format: [source_file_id (32 bytes)][dest_file_id (32 bytes)][amount (8 bytes)]
    let data: bytes = getInstructionData();
    
    // Extract source and destination file IDs (simplified - real implementation would parse bytes)
    let sourceId: i64 = 1;  // In real implementation, parse from data
    let destId: i64 = 2;
    let amount: i64 = 100;
    
    // Verify the transaction is signed by the source account owner
    // In a real implementation, we'd extract the public key from the file
    let hasSigner: bool = true;  // Simplified check
    
    if (!hasSigner) {
        return -1;  // Error: unauthorized
    }
    
    // Get source account balance
    let sourceBalance: i64 = getBalance(sourceId);
    
    // Check sufficient balance
    if (sourceBalance < amount) {
        return -2;  // Error: insufficient balance
    }
    
    // Get destination balance
    let destBalance: i64 = getBalance(destId);
    
    // CORRECT APPROACH: Invoke the system program to perform the transfer
    let systemProgramId: i64 = 0x01;  // System program ID
    
    // Encode transfer instruction: [instruction_type=1, amount (8 bytes)]
    let transferData: bytes = encodeTransfer(amount);
    
    // Invoke system program with compute budget
    let result: i64 = invoke(systemProgramId, transferData, 10000);
    
    if (result != 0) {
        return -3;  // Error: transfer failed
    }
    
    // Return success
    return 0;
}

// Helper to encode transfer instruction for system program
function encodeTransfer(amount: i64): bytes {
    // Simplified - real implementation would properly encode
    // Format: [instruction_type=1][amount as 8-byte little-endian]
    let data: bytes = allocateBytes(9);
    // data[0] = 1;  // Transfer instruction
    // Encode amount...
    return data;
}

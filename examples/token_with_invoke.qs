// Token Program with Cross-Program Invocation
// Demonstrates calling the system program for balance transfers
// and another program for validation logic
//
// NOTE: Balance transfers must be done by invoking the system program.
// Only the system program can call updateBalance directly.

export function entry(ctx: InstructionContext): i64 {
    // Get instruction data
    let data: bytes = getInstructionData();
    
    // Parse transfer parameters
    let sourceId: i64 = 1;
    let destId: i64 = 2;
    let amount: i64 = 100;
    let validatorProgramId: i64 = 999;  // ID of validator program
    
    // Check if we need to invoke a validator program
    // In a real scenario, this would be parsed from instruction data
    let needsValidation: bool = true;
    
    if (needsValidation) {
        // Invoke a validator program to check if transfer is allowed
        // For example: checking daily limits, blacklists, compliance rules, etc.
        let validationData: bytes = encodeValidationRequest(sourceId, destId, amount);
        let validationResult: i64 = invoke(validatorProgramId, validationData, 5000);
        
        if (validationResult != 0) {
            return -3;  // Error: validation failed
        }
    }
    
    // Verify sufficient balance
    let sourceBalance: i64 = getBalance(sourceId);
    
    if (sourceBalance < amount) {
        return -2;  // Error: insufficient balance
    }
    
    // Calculate fee (1% of amount)
    let fee: i64 = amount / 100;
    let netAmount: i64 = amount - fee;
    
    // Perform the transfer by invoking the system program
    let systemProgramId: i64 = 0x01;
    
    // First transfer: source -> destination (net amount)
    let transferData1: bytes = encodeTransfer(netAmount);
    let result1: i64 = invoke(systemProgramId, transferData1, 10000);
    
    if (result1 != 0) {
        return -4;  // Error: transfer to destination failed
    }
    
    // Second transfer: source -> program (fee)
    let programId: i64 = getProgramId();
    let transferData2: bytes = encodeTransfer(fee);
    let result2: i64 = invoke(systemProgramId, transferData2, 10000);
    
    if (result2 != 0) {
        return -5;  // Error: fee transfer failed
    }
    
    // Return success with fee amount
    return fee;
}

// Helper to encode validation request
function encodeValidationRequest(source: i64, dest: i64, amount: i64): bytes {
    // Simplified - real implementation would properly encode
    let data: bytes = allocateBytes(24);
    return data;
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

// Example: Cryptographic Operations
// Level: 4
// Features: sha256, getInstructionData, len function
// Description: Demonstrates cryptographic hashing with instruction data

export function entry(): i64 {
    // Get instruction data (returns bytes)
    let data: bytes = getInstructionData();
    
    // Compute SHA-256 hash
    let hash: bytes = sha256(data);
    
    // Return the length of the hash (should be 32 bytes)
    return len(hash);
}

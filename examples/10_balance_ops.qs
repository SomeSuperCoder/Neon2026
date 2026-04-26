// Example: Balance Operations
// Level: 4
// Features: getBalance blockchain operation
// Description: Demonstrates querying file balances and computing new balance

export function entry(): i64 {
    // Get current balance of file index 0
    let currentBalance: i64 = getBalance(0);
    
    // Compute new balance (add 100)
    let newBalance: i64 = currentBalance + 100;
    
    // Return the new balance
    return newBalance;
}

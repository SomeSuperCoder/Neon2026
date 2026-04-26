// Example: Balance Operations
// Level: 4
// Features: getBalance, updateBalance blockchain operations
// Description: Demonstrates querying and updating file balances

export function entry(): i64 {
    // Get current balance of file index 0
    let currentBalance: i64 = getBalance(0);
    
    // Update balance by adding 100 (delta, not absolute value)
    updateBalance(0, 100);
    
    // Return the new balance
    return getBalance(0);
}

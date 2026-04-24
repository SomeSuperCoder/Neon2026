// Simple QuanticScript program demonstrating basic syntax
// This program performs a simple transfer calculation

export function entry(ctx: InstructionContext): i64 {
    // Declare variables
    let amount: i64 = 1000;
    let fee: i64 = 10;
    
    // Calculate total
    let total: i64 = amount + fee;
    
    // Return the result
    return total;
}

// Arithmetic Operations Demo
// Demonstrates basic arithmetic and variable usage

export function entry(ctx: InstructionContext): i64 {
    // Basic arithmetic
    let a: i64 = 100;
    let b: i64 = 50;
    
    let sum: i64 = a + b;        // 150
    let diff: i64 = a - b;       // 50
    let product: i64 = a * b;    // 5000
    let quotient: i64 = a / b;   // 2
    let remainder: i64 = a % b;  // 0
    
    // Complex expression
    let result: i64 = (a + b) * 2 - quotient;
    
    return result;  // Returns 298
}

// Control Flow Demo
// Demonstrates if statements, loops, and functions

function max(a: i64, b: i64): i64 {
    if (a > b) {
        return a;
    } else {
        return b;
    }
}

function factorial(n: i64): i64 {
    let result: i64 = 1;
    let i: i64 = 1;
    
    while (i <= n) {
        result = result * i;
        i = i + 1;
    }
    
    return result;
}

function sumToN(n: i64): i64 {
    let sum: i64 = 0;
    
    for (let i: i64 = 1; i <= n; i = i + 1) {
        sum = sum + i;
    }
    
    return sum;
}

export function entry(ctx: InstructionContext): i64 {
    let x: i64 = 10;
    let y: i64 = 20;
    
    // Test max function
    let maximum: i64 = max(x, y);  // 20
    
    // Test factorial
    let fact5: i64 = factorial(5);  // 120
    
    // Test sum
    let sum10: i64 = sumToN(10);  // 55
    
    // Return combined result
    return maximum + fact5 + sum10;  // 195
}

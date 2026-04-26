// Example: For Loop
// Level: 2
// Features: for loops, loop initialization, conditions, updates
// Description: Demonstrates for loop to calculate sum of numbers 1 to 10

export function entry(): i64 {
    let sum: i64 = 0;
    for (let i: i64 = 1; i <= 10; i = i + 1) {
        sum = sum + i;
    }
    return sum;
}

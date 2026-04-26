// Example: While Loop
// Level: 2
// Features: while loops, loop conditions, increment operators
// Description: Demonstrates while loop to calculate sum of numbers 1 to 10

export function entry(): i64 {
    let sum: i64 = 0;
    let i: i64 = 1;
    while (i <= 10) {
        sum = sum + i;
        i = i + 1;
    }
    return sum;
}

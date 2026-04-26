// Example: Recursion
// Level: 3
// Features: recursive function calls, base cases
// Description: Demonstrates recursive functions with the factorial example

function factorial(n: i64): i64 {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

export function entry(): i64 {
    return factorial(5);
}

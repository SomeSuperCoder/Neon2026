// Example: Fibonacci
// Level: 3
// Features: recursive function calls, multiple base cases
// Description: Demonstrates recursive Fibonacci calculation

function fibonacci(n: i64): i64 {
    if (n <= 0) {
        return 0;
    }
    if (n == 1) {
        return 1;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

export function entry(): i64 {
    return fibonacci(10);
}

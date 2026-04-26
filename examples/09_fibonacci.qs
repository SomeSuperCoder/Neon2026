// Example: Fibonacci
// Level: 3
// Features: recursive functions, mathematical computation
// Description: Demonstrates recursive Fibonacci calculation

function fibonacci(n: i64): i64 {
    if (n <= 1) {
        return n;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

export function entry(): i64 {
    return fibonacci(10);
}

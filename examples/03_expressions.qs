// Example: Expressions
// Level: 1
// Features: complex expressions, operator precedence, multiple operations
// Description: Demonstrates complex expressions with operator precedence

export function entry(): i64 {
    let a: i64 = 5;
    let b: i64 = 3;
    let c: i64 = 2;
    let result: i64 = (a + b) * c - a / b;
    return result;
}

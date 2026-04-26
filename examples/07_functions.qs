// Example: Functions
// Level: 3
// Features: function declarations, function calls, parameters, return values
// Description: Demonstrates basic function declarations with parameters and return values

function add(a: i64, b: i64): i64 {
    return a + b;
}

export function entry(): i64 {
    let result: i64 = add(10, 20);
    return result;
}

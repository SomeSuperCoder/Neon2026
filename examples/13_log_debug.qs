// Example: Using the log() function for debugging
// The log() function is a high-cost operation (5000 units) designed for debugging
// It consumes a value from the stack and logs it (in production, this would be expensive)

export function entry(): i64 {
    let x: i64 = 42;
    let msg: string = "debug value";
    
    // Log the integer value
    log(x);
    
    // Log the string value
    log(msg);
    
    // Log the result of an expression
    log(x + 8);
    
    return x;
}

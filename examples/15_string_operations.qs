// String Operations Example
// Demonstrates string literals, concatenation, and logging

export function entry(): i64 {
    let greeting: string = "Hello";
    let name: string = "World";
    
    // Log individual strings
    log(greeting);
    log(name);
    
    // Concatenate strings
    let message: string = stringConcat(greeting, ", ");
    let fullMessage: string = stringConcat(message, name);
    let withExclamation: string = stringConcat(fullMessage, "!");
    
    // Log the final message
    log(withExclamation);
    
    // Get string length
    let len: i64 = stringLength(withExclamation);
    log(len);
    
    return 0;
}

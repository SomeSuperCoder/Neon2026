// Hello, World! in QuanticScript
// This example demonstrates string literals and the log() function for debugging

export function entry(): i64 {
    let greeting: string = "Hello, World!";
    
    // Log the greeting message
    log(greeting);
    
    // Return success code
    return 0;
}
